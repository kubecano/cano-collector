package workflow

import (
	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/workflow/model"
)

// WorkflowEngine - podstawowa implementacja silnika workflow
type WorkflowEngine struct {
	Workflows []workflow.WorkflowDefinition
}

// NewWorkflowEngine - tworzy nowy silnik workflow
func NewWorkflowEngine(workflows []workflow.WorkflowDefinition) *WorkflowEngine {
	return &WorkflowEngine{Workflows: workflows}
}

// SelectWorkflows - wybiera workflow pasujące do alertu (TODO: implementacja matching)
func (e *WorkflowEngine) SelectWorkflows(input model.AlertInput) []workflow.WorkflowDefinition {
	// TODO: Matching logic (na razie zwraca wszystkie)
	return e.Workflows
}

// ExecuteWorkflow - wykonuje akcje workflow (TODO: implementacja akcji)
func (e *WorkflowEngine) ExecuteWorkflow(wf workflow.WorkflowDefinition, input model.AlertInput) error {
	// TODO: Wykonanie akcji workflow (create_issue, resolve_issue, dispatch_issue)
	return nil
}
