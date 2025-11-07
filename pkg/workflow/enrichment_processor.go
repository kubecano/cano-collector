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

	if len(alertEvent.Alerts) != len(issues) {
		ep.logger.Warn("Mismatch between alerts and issues count",
			zap.Int("alerts_count", len(alertEvent.Alerts)),
			zap.Int("issues_count", len(issues)),
		)
	}

	ep.logger.Info("Processing workflow enrichments individually per issue",
		zap.Int("issues_count", len(issues)),
	)

	// Process each issue with its corresponding alert to prevent cross-contamination
	for i, iss := range issues {
		var singleAlert event.PrometheusAlert
		if i < len(alertEvent.Alerts) {
			singleAlert = alertEvent.Alerts[i]
		} else {
			ep.logger.Warn("Alert index mismatch, using first alert",
				zap.Int("issue_index", i),
				zap.Int("alerts_count", len(alertEvent.Alerts)),
			)
			singleAlert = alertEvent.Alerts[0]
		}
		singleAlertEvent := &event.AlertManagerEvent{
			Receiver:          alertEvent.Receiver,
			Status:            alertEvent.Status,
			Alerts:            []event.PrometheusAlert{singleAlert},
			GroupLabels:       alertEvent.GroupLabels,
			CommonLabels:      alertEvent.CommonLabels,
			CommonAnnotations: alertEvent.CommonAnnotations,
			ExternalURL:       alertEvent.ExternalURL,
		}

		workflowEvent := event.NewAlertManagerWorkflowEvent(singleAlertEvent)
		matchingWorkflows := workflowEngine.SelectWorkflows(workflowEvent)

		if len(matchingWorkflows) == 0 {
			ep.logger.Debug("No matching workflows for this issue",
				zap.String("issue_id", iss.ID.String()),
				zap.String("alert_name", singleAlert.Labels["alertname"]),
			)
			continue
		}

		ep.logger.Debug("Executing workflows for individual issue",
			zap.String("issue_id", iss.ID.String()),
			zap.String("alert_name", singleAlert.Labels["alertname"]),
			zap.String("pod", singleAlert.Labels["pod"]),
			zap.Int("matching_workflows", len(matchingWorkflows)),
		)

		enrichments, err := workflowEngine.ExecuteWorkflowsWithEnrichments(ctx, matchingWorkflows, workflowEvent)
		if err != nil {
			ep.logger.Error("Failed to execute workflows for issue",
				zap.Error(err),
				zap.String("issue_id", iss.ID.String()),
			)
			continue
		}

		for _, enrichment := range enrichments {
			iss.AddEnrichment(enrichment)
		}

		ep.logger.Debug("Applied enrichments to issue",
			zap.String("issue_id", iss.ID.String()),
			zap.String("alert_name", singleAlert.Labels["alertname"]),
			zap.String("pod", singleAlert.Labels["pod"]),
			zap.Int("enrichments_count", len(enrichments)),
		)
	}

	ep.logger.Info("Completed individual workflow enrichment processing",
		zap.Int("processed_issues", len(issues)),
	)

	return nil
}
