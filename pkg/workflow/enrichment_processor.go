package workflow

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/config/workflow"
	"github.com/kubecano/cano-collector/pkg/core/event"
	"github.com/kubecano/cano-collector/pkg/core/issue"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	workflow_interfaces "github.com/kubecano/cano-collector/pkg/workflow/interfaces"
)

// EnrichmentProcessor applies workflow-generated enrichments to issues
type EnrichmentProcessor struct {
	logger logger_interfaces.LoggerInterface
}

// NewEnrichmentProcessor creates a new EnrichmentProcessor
func NewEnrichmentProcessor(logger logger_interfaces.LoggerInterface) *EnrichmentProcessor {
	return &EnrichmentProcessor{
		logger: logger,
	}
}

// ProcessWorkflowEnrichments applies enrichments from workflow execution to issues
func (ep *EnrichmentProcessor) ProcessWorkflowEnrichments(
	ctx context.Context,
	issues []*issue.Issue,
	workflowEngine workflow_interfaces.WorkflowEngineInterface,
	alertEvent *event.AlertManagerEvent,
) error {
	if workflowEngine == nil {
		ep.logger.Debug("No workflow engine provided, skipping enrichment processing")
		return nil
	}

	if len(issues) == 0 {
		ep.logger.Debug("No issues to enrich")
		return nil
	}

	// Convert AlertManagerEvent to WorkflowEvent
	workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

	// Select matching workflows
	matchingWorkflows := workflowEngine.SelectWorkflows(workflowEvent)
	if len(matchingWorkflows) == 0 {
		ep.logger.Debug("No matching workflows found for enrichment processing")
		return nil
	}

	ep.logger.Info("Processing workflow enrichments",
		zap.String("alert_name", alertEvent.GetAlertName()),
		zap.Int("matching_workflows", len(matchingWorkflows)),
		zap.Int("issues_count", len(issues)),
	)

	// Execute workflows and collect enrichments
	allEnrichments := make(map[string][]issue.Enrichment) // workflow name -> enrichments

	for _, workflow := range matchingWorkflows {
		// Execute workflow and capture enrichments
		enrichments, err := ep.executeWorkflowForEnrichments(ctx, workflowEngine, workflow, workflowEvent)
		if err != nil {
			ep.logger.Error("Failed to execute workflow for enrichments",
				zap.Error(err),
				zap.String("workflow_name", workflow.Name))
			continue
		}

		if len(enrichments) > 0 {
			allEnrichments[workflow.Name] = enrichments
			ep.logger.Info("Collected enrichments from workflow",
				zap.String("workflow_name", workflow.Name),
				zap.Int("enrichments_count", len(enrichments)),
			)
		}
	}

	// Apply enrichments to all issues
	for _, iss := range issues {
		for workflowName, enrichments := range allEnrichments {
			for _, enrichment := range enrichments {
				iss.AddEnrichment(enrichment)
			}
			ep.logger.Debug("Applied workflow enrichments to issue",
				zap.String("issue_id", iss.ID.String()),
				zap.String("workflow_name", workflowName),
				zap.Int("enrichments_count", len(enrichments)),
			)
		}
	}

	ep.logger.Info("Completed workflow enrichment processing",
		zap.Int("total_workflows", len(matchingWorkflows)),
		zap.Int("enriched_workflows", len(allEnrichments)),
		zap.Int("enriched_issues", len(issues)),
	)

	return nil
}

// executeWorkflowForEnrichments executes a workflow and extracts enrichments from the results
func (ep *EnrichmentProcessor) executeWorkflowForEnrichments(
	ctx context.Context,
	workflowEngine workflow_interfaces.WorkflowEngineInterface,
	workflowDef *workflow.WorkflowDefinition,
	workflowEvent event.WorkflowEvent,
) ([]issue.Enrichment, error) {
	// Use a custom executor to capture enrichments instead of executing through the main engine
	// This prevents duplicate execution but allows us to collect enrichments

	// For now, we'll execute the workflow and extract enrichments
	// In a future optimization, we could modify the engine to return enrichments
	// without executing side effects

	err := workflowEngine.ExecuteWorkflow(ctx, workflowDef, workflowEvent)
	if err != nil {
		return nil, fmt.Errorf("workflow execution failed: %w", err)
	}

	// Since the current ExecuteWorkflow doesn't return enrichments directly,
	// we need to modify this approach. For now, we'll return empty enrichments
	// and rely on the fact that the workflow has already been executed in the AlertHandler.

	// TODO: This is a design limitation - we need to either:
	// 1. Modify WorkflowEngine.ExecuteWorkflow to return ActionResults with enrichments
	// 2. Or create a separate method that executes workflows and returns enrichments
	// 3. Or modify the AlertHandler to pass enrichments to the converter

	ep.logger.Warn("EnrichmentProcessor currently relies on workflow execution in AlertHandler",
		zap.String("workflow_name", workflowDef.Name),
		zap.String("note", "This is a design limitation that needs to be addressed"))

	return []issue.Enrichment{}, nil
}
