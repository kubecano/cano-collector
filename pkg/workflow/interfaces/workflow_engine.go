package interfaces

import (
	"context"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

//go:generate mockgen -source=workflow_engine.go -destination=../../../mocks/workflow_engine_mock.go -package=mocks

// WorkflowEngineInterface defines the interface for workflow processing
type WorkflowEngineInterface interface {
	// SelectWorkflows returns workflows that match the given event
	SelectWorkflows(event event.WorkflowEvent) []*workflow.WorkflowDefinition

	// ExecuteWorkflow executes a single workflow for the given event
	ExecuteWorkflow(ctx context.Context, workflow *workflow.WorkflowDefinition, event event.WorkflowEvent) error

	// ExecuteWorkflowWithEnrichments executes a workflow and returns enrichments from the results
	ExecuteWorkflowWithEnrichments(ctx context.Context, workflow *workflow.WorkflowDefinition, event event.WorkflowEvent) ([]issue.Enrichment, error)

	// ExecuteWorkflowsWithEnrichments executes multiple workflows and returns all enrichments
	ExecuteWorkflowsWithEnrichments(ctx context.Context, workflows []*workflow.WorkflowDefinition, event event.WorkflowEvent) ([]issue.Enrichment, error)
}

// WorkflowExecutionResult contains the results of workflow execution
type WorkflowExecutionResult struct {
	WorkflowName  string
	Success       bool
	Error         error
	ActionResults []*actions_interfaces.ActionResult
	Enrichments   []issue.Enrichment
}
