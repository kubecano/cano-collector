package interfaces

import (
	"context"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
)

//go:generate mockgen -source=workflow_engine.go -destination=../../../mocks/workflow_engine_mock.go -package=mocks

// WorkflowEngineInterface defines the interface for workflow processing
type WorkflowEngineInterface interface {
	// SelectWorkflows returns workflows that match the given event
	SelectWorkflows(event event.WorkflowEvent) []*workflow.WorkflowDefinition

	// ExecuteWorkflow executes a single workflow for the given event
	ExecuteWorkflow(ctx context.Context, workflow *workflow.WorkflowDefinition, event event.WorkflowEvent) error
}
