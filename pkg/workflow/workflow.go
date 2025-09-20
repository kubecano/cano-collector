package workflow

import (
	"context"
	"fmt"
	"sort"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// WorkflowEngine processes workflows for incoming events
type WorkflowEngine struct {
	config   *workflow.WorkflowConfig
	executor actions_interfaces.ActionExecutor
	logger   logger_interfaces.LoggerInterface
	metrics  metric_interfaces.MetricsInterface
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(config *workflow.WorkflowConfig, executor actions_interfaces.ActionExecutor, logger logger_interfaces.LoggerInterface, metrics metric_interfaces.MetricsInterface) *WorkflowEngine {
	return &WorkflowEngine{
		config:   config,
		executor: executor,
		logger:   logger,
		metrics:  metrics,
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

// ExecuteWorkflowWithEnrichments executes a workflow and returns enrichments from the results
func (we *WorkflowEngine) ExecuteWorkflowWithEnrichments(ctx context.Context, wf *workflow.WorkflowDefinition, event event.WorkflowEvent) ([]issue.Enrichment, error) {
	if wf == nil {
		return nil, fmt.Errorf("workflow definition cannot be nil")
	}

	if event == nil {
		return nil, fmt.Errorf("workflow event cannot be nil")
	}

	if we.executor == nil {
		return nil, fmt.Errorf("action executor is not configured")
	}

	// Convert workflow action definitions to action configs
	actionConfigs, err := we.createActionConfigs(wf)
	if err != nil {
		return nil, err
	}

	// Create workflow actions from configs
	workflowActions, err := we.executor.CreateActionsFromConfig(actionConfigs)
	if err != nil {
		if we.logger != nil {
			we.logger.Error("Failed to create actions for workflow",
				zap.String("workflow", wf.Name),
				zap.Error(err))
		}
		return nil, err
	}

	// Execute actions in sequence using the provided context
	results, err := we.executor.ExecuteActions(ctx, workflowActions, event)
	if err != nil {
		if we.logger != nil {
			we.logger.Error("Failed to execute actions for workflow",
				zap.String("workflow", wf.Name),
				zap.Error(err))
		}
		return nil, err
	}

	// Collect enrichments from all successful action results
	var allEnrichments []issue.Enrichment
	successCount := 0

	for i, result := range results {
		if result.Success {
			successCount++
			// Add enrichments from this action result
			allEnrichments = append(allEnrichments, result.Enrichments...)
		} else if result.Error != nil {
			// Log error but don't fail completely - collect what we can
			// This allows partial enrichment even if some actions fail
			if we.logger != nil {
				we.logger.Warn("Action failed in workflow, continuing with partial enrichment",
					zap.Int("action_index", i),
					zap.String("workflow", wf.Name),
					zap.Error(result.Error))
			}
			if we.metrics != nil {
				we.metrics.IncWorkflowEnrichmentErrors(wf.Name, "action_execution_failed")
			}
		}
	}

	// Record metrics for workflow execution
	if we.metrics != nil {
		if successCount == len(results) {
			we.metrics.IncWorkflowsExecuted(wf.Name, "success")
		} else if successCount > 0 {
			we.metrics.IncWorkflowsExecuted(wf.Name, "partial_success")
		} else {
			we.metrics.IncWorkflowsExecuted(wf.Name, "failed")
		}

		// Record enrichment count
		we.metrics.ObserveWorkflowEnrichments(wf.Name, len(allEnrichments))
	}

	if we.logger != nil {
		we.logger.Info("Workflow execution completed",
			zap.String("workflow", wf.Name),
			zap.Int("successful_actions", successCount),
			zap.Int("total_actions", len(results)),
			zap.Int("enrichments_generated", len(allEnrichments)))
	}

	return allEnrichments, nil
}

// createActionConfigs converts workflow action definitions to action configs
// This method extracts the common logic shared between ExecuteWorkflow and ExecuteWorkflowWithEnrichments
func (we *WorkflowEngine) createActionConfigs(wf *workflow.WorkflowDefinition) ([]actions_interfaces.ActionConfig, error) {
	actionConfigs := make([]actions_interfaces.ActionConfig, 0, len(wf.Actions))

	for i, actionDef := range wf.Actions {
		// Extract action type from RawData
		actionType := actionDef.ActionType
		if actionType == "" {
			// First, check for explicit action_type field in RawData
			if explicitActionType, exists := actionDef.RawData["action_type"]; exists {
				if actionTypeStr, ok := explicitActionType.(string); ok {
					actionType = actionTypeStr
				}
			}
		}
		
		if actionType == "" {
			// Try to infer action type from RawData keys deterministically (backward compatibility)
			var candidateKeys []string
			for key := range actionDef.RawData {
				if key != "action_type" && key != "data" {
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
			return nil, fmt.Errorf("action %d in workflow '%s' has no action type", i, wf.Name)
		}

		// Create action config with proper parameter extraction
		var parameters map[string]interface{}
		
		// Check if we have action_type/data structure
		if _, hasActionType := actionDef.RawData["action_type"]; hasActionType {
			if dataField, hasData := actionDef.RawData["data"]; hasData {
				if dataMap, ok := dataField.(map[string]interface{}); ok {
					// Use data field as parameters for action_type/data structure
					parameters = make(map[string]interface{})
					for key, value := range dataMap {
						parameters[key] = value
					}
				} else {
					// Fallback to full RawData if data field is not a map
					parameters = actionDef.RawData
				}
			} else {
				// No data field, use RawData but exclude action_type
				parameters = make(map[string]interface{})
				for key, value := range actionDef.RawData {
					if key != "action_type" {
						parameters[key] = value
					}
				}
			}
		} else {
			// Backward compatibility: use all RawData as parameters
			parameters = actionDef.RawData
		}

		actionConfig := actions_interfaces.ActionConfig{
			Name:       fmt.Sprintf("%s-action-%d", wf.Name, i),
			Type:       actionType,
			Enabled:    true,
			Timeout:    30, // Default timeout
			Parameters: parameters,
		}

		// For backward compatibility: extract specific parameters if they exist in RawData
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

	return actionConfigs, nil
}

// ExecuteWorkflowsWithEnrichments executes multiple workflows and returns all enrichments
func (we *WorkflowEngine) ExecuteWorkflowsWithEnrichments(ctx context.Context, workflows []*workflow.WorkflowDefinition, event event.WorkflowEvent) ([]issue.Enrichment, error) {
	if len(workflows) == 0 {
		return []issue.Enrichment{}, nil
	}

	var allEnrichments []issue.Enrichment
	var errors []error

	for _, wf := range workflows {
		enrichments, err := we.ExecuteWorkflowWithEnrichments(ctx, wf, event)
		if err != nil {
			we.logger.Error("Workflow execution failed, continuing with others",
				zap.String("workflow", wf.Name),
				zap.Error(err))
			errors = append(errors, err)
			continue
		}
		allEnrichments = append(allEnrichments, enrichments...)
	}

	// Return enrichments even if some workflows failed
	// This allows graceful degradation
	if len(errors) > 0 && len(allEnrichments) == 0 {
		we.logger.Error("All workflows failed, no enrichments collected",
			zap.Int("workflow_count", len(workflows)),
			zap.Int("error_count", len(errors)))
		return nil, fmt.Errorf("all %d workflows failed", len(workflows))
	}

	return allEnrichments, nil
}
