package workflow

import (
	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
)

// WorkflowEngine processes workflows for incoming events
type WorkflowEngine struct {
	config *workflow.WorkflowConfig
}

// NewWorkflowEngine creates a new workflow engine
func NewWorkflowEngine(config *workflow.WorkflowConfig) *WorkflowEngine {
	return &WorkflowEngine{
		config: config,
	}
}

// SelectWorkflows returns workflows that match the given event
func (we *WorkflowEngine) SelectWorkflows(event *event.AlertManagerEvent) []*workflow.WorkflowDefinition {
	var matchingWorkflows []*workflow.WorkflowDefinition

	for _, wf := range we.config.ActiveWorkflows {
		if we.matchesWorkflow(&wf, event) {
			matchingWorkflows = append(matchingWorkflows, &wf)
		}
	}

	return matchingWorkflows
}

// matchesWorkflow checks if a workflow should be triggered for the given event
func (we *WorkflowEngine) matchesWorkflow(wf *workflow.WorkflowDefinition, event *event.AlertManagerEvent) bool {
	// Check all triggers in the workflow
	for _, trigger := range wf.Triggers {
		if we.matchesTrigger(&trigger, event) {
			return true
		}
	}
	return false
}

// matchesTrigger checks if a single trigger matches the event
func (we *WorkflowEngine) matchesTrigger(trigger *workflow.TriggerDefinition, event *event.AlertManagerEvent) bool {
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
func (we *WorkflowEngine) matchesAlertmanagerAlertTrigger(trigger *workflow.AlertmanagerAlertTrigger, event *event.AlertManagerEvent) bool {
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
func (we *WorkflowEngine) ExecuteWorkflow(wf *workflow.WorkflowDefinition, event *event.AlertManagerEvent) error {
	// TODO: Implement workflow execution logic
	// This will include:
	// 1. Processing each action in the workflow
	// 2. Handling action types (create_issue, resolve_issue, dispatch_issue)
	// 3. Template rendering for action data
	// 4. Error handling and logging

	// For now, return nil (no-op)
	return nil
}
