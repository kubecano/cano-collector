package workflow

import (
	"context"

	"go.uber.org/zap"

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

	// Execute workflows and collect enrichments using the new enrichment-aware method
	enrichments, err := workflowEngine.ExecuteWorkflowsWithEnrichments(ctx, matchingWorkflows, workflowEvent)
	if err != nil {
		ep.logger.Error("Failed to execute workflows for enrichments", zap.Error(err))
		return err
	}

	// Group enrichments by workflow name for application to issues
	allEnrichments := make(map[string][]issue.Enrichment)
	if len(enrichments) > 0 {
		// For now, we'll apply all enrichments from all workflows
		// Future enhancement could track which workflow generated which enrichments
		allEnrichments["all_workflows"] = enrichments
		ep.logger.Info("Collected enrichments from workflows",
			zap.Int("total_enrichments", len(enrichments)),
		)
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
