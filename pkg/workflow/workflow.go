package workflow

import (
	"context"
	"fmt"
	"sort"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// WorkflowEngine processes workflows for incoming events
type WorkflowEngine struct {
	config   *workflow.WorkflowConfig
	executor actions_interfaces.ActionExecutor
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(config *workflow.WorkflowConfig, executor actions_interfaces.ActionExecutor) *WorkflowEngine {
	return &WorkflowEngine{
		config:   config,
		executor: executor,
	}
}

// SelectWorkflows returns workflows that match the given event
func (we *WorkflowEngine) SelectWorkflows(event event.WorkflowEvent) []*workflow.WorkflowDefinition {
	var matchingWorkflows []*workflow.WorkflowDefinition

	for _, wf := range we.config.ActiveWorkflows {
		if we.matchesWorkflow(&wf, event) {
			matchingWorkflows = append(matchingWorkflows, &wf)
		}
	}

	return matchingWorkflows
}

// matchesWorkflow checks if a workflow should be triggered for the given event
func (we *WorkflowEngine) matchesWorkflow(wf *workflow.WorkflowDefinition, event event.WorkflowEvent) bool {
	// Check all triggers in the workflow
	for _, trigger := range wf.Triggers {
		if we.matchesTrigger(&trigger, event) {
			return true
		}
	}
	return false
}

// matchesTrigger checks if a single trigger matches the event
func (we *WorkflowEngine) matchesTrigger(trigger *workflow.TriggerDefinition, event event.WorkflowEvent) bool {
	// Currently only support AlertManager triggers
	if trigger.OnAlertmanagerAlert != nil {
		return we.matchesAlertmanagerAlertTrigger(trigger.OnAlertmanagerAlert, event)
	}

	// Future: Add support for other trigger types
	// if trigger.OnKubernetesEvent != nil { ... }
	// if trigger.OnScheduledEvent != nil { ... }

	return false
}

// matchesAlertmanagerAlertTrigger checks if an AlertManager trigger matches the event
// Uses WorkflowEvent interface methods to avoid direct coupling to AlertManagerEvent
func (we *WorkflowEngine) matchesAlertmanagerAlertTrigger(trigger *workflow.AlertmanagerAlertTrigger, event event.WorkflowEvent) bool {
	// Check alert_name if specified
	if trigger.AlertName != "" && trigger.AlertName != event.GetAlertName() {
		return false
	}

	// Check status if specified
	if trigger.Status != "" && trigger.Status != event.GetStatus() {
		return false
	}

	// Check severity if specified
	if trigger.Severity != "" && trigger.Severity != event.GetSeverity() {
		return false
	}

	// Check namespace if specified
	if trigger.Namespace != "" && trigger.Namespace != event.GetNamespace() {
		return false
	}

	// All specified conditions match
	return true
}

// ExecuteWorkflow executes a single workflow for the given event
func (we *WorkflowEngine) ExecuteWorkflow(ctx context.Context, wf *workflow.WorkflowDefinition, event event.WorkflowEvent) error {
	if wf == nil {
		return fmt.Errorf("workflow definition cannot be nil")
	}

	if event == nil {
		return fmt.Errorf("workflow event cannot be nil")
	}

	if we.executor == nil {
		return fmt.Errorf("action executor is not configured")
	}

	// Convert workflow action definitions to action configs
	actionConfigs := make([]actions_interfaces.ActionConfig, 0, len(wf.Actions))

	for i, actionDef := range wf.Actions {
		// Extract action type from RawData
		actionType := actionDef.ActionType
		if actionType == "" {
			// Try to infer action type from RawData keys deterministically
			var candidateKeys []string
			for key := range actionDef.RawData {
				if key != "action_type" {
					candidateKeys = append(candidateKeys, key)
				}
			}

			// Sort keys alphabetically to ensure deterministic behavior
			sort.Strings(candidateKeys)

			// Use the first key after sorting
			if len(candidateKeys) > 0 {
				actionType = candidateKeys[0]
			}
		}

		if actionType == "" {
			return fmt.Errorf("action %d in workflow '%s' has no action type", i, wf.Name)
		}

		// Create action config
		actionConfig := actions_interfaces.ActionConfig{
			Name:       fmt.Sprintf("%s-action-%d", wf.Name, i),
			Type:       actionType,
			Enabled:    true,
			Timeout:    30, // Default timeout
			Parameters: actionDef.RawData,
		}

		// Extract specific parameters if they exist in RawData
		if actionData, exists := actionDef.RawData[actionType]; exists {
			if actionDataMap, ok := actionData.(map[string]interface{}); ok {
				// Merge action-specific data into parameters
				for key, value := range actionDataMap {
					actionConfig.Parameters[key] = value
				}
			}
		}

		actionConfigs = append(actionConfigs, actionConfig)
	}

	// Create workflow actions from configs
	workflowActions, err := we.executor.CreateActionsFromConfig(actionConfigs)
	if err != nil {
		return fmt.Errorf("failed to create actions for workflow '%s': %w", wf.Name, err)
	}

	// Execute actions in sequence using the provided context
	results, err := we.executor.ExecuteActions(ctx, workflowActions, event)
	if err != nil {
		return fmt.Errorf("failed to execute actions for workflow '%s': %w", wf.Name, err)
	}

	// Process results - check if any action failed
	var actionErrors []error
	successCount := 0

	for i, result := range results {
		if result.Success {
			successCount++
		} else if result.Error != nil {
			actionErrors = append(actionErrors, fmt.Errorf("action %d failed: %w", i, result.Error))
		}
	}

	// Return error if any actions failed (could be configurable in the future)
	if len(actionErrors) > 0 {
		return fmt.Errorf("workflow '%s' completed with %d/%d actions successful, errors: %v",
			wf.Name, successCount, len(results), actionErrors)
	}

	return nil
}
