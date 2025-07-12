package iface

import (
	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/workflow/model"
)

//go:generate mockgen -destination=../../mocks/workflow_engine_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/workflow/interface WorkflowEngineInterface

type WorkflowEngineInterface interface {
	SelectWorkflows(input model.AlertInput) []workflow.WorkflowDefinition
	ExecuteWorkflow(wf workflow.WorkflowDefinition, input model.AlertInput) error
}
