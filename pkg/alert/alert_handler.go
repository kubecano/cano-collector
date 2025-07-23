package alert

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"

	alert_interfaces "github.com/kubecano/cano-collector/pkg/alert/interfaces"
	"github.com/kubecano/cano-collector/pkg/core/event"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	workflow_interfaces "github.com/kubecano/cano-collector/pkg/workflow/interfaces"
)

// AlertHandler handles incoming alerts from Alertmanager
type AlertHandler struct {
	logger          logger_interfaces.LoggerInterface
	metrics         metric_interfaces.MetricsInterface
	teamResolver    alert_interfaces.TeamResolverInterface
	alertDispatcher alert_interfaces.AlertDispatcherInterface
	converter       alert_interfaces.ConverterInterface
	workflowEngine  workflow_interfaces.WorkflowEngineInterface
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(
	logger logger_interfaces.LoggerInterface,
	metrics metric_interfaces.MetricsInterface,
	teamResolver alert_interfaces.TeamResolverInterface,
	alertDispatcher alert_interfaces.AlertDispatcherInterface,
	converter alert_interfaces.ConverterInterface,
	workflowEngine workflow_interfaces.WorkflowEngineInterface,
) *AlertHandler {
	return &AlertHandler{
		logger:          logger,
		metrics:         metrics,
		teamResolver:    teamResolver,
		alertDispatcher: alertDispatcher,
		converter:       converter,
		workflowEngine:  workflowEngine,
	}
}

// HandleAlert processes alerts
func (h *AlertHandler) HandleAlert(c *gin.Context) {
	start := time.Now()

	// Check if the request body is empty
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	if len(bytes.TrimSpace(body)) == 0 {
		h.logger.Error("Empty request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty JSON body"})
		return
	}

	// Restore the body for JSON binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var templateData template.Data
	if err := c.ShouldBindJSON(&templateData); err != nil {
		h.logger.Error("Failed to parse alert", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert format"})
		return
	}

	// Convert template.Data to AlertManagerEvent
	alertEvent := event.NewAlertManagerEvent(templateData)
	if err := alertEvent.Validate(); err != nil {
		h.logger.Error("Invalid alert structure", zap.Error(err), zap.Any("alert", templateData))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert format: " + err.Error()})
		return
	}

	// Register received alert metric
	h.metrics.ObserveAlert(alertEvent.Receiver, alertEvent.Status)

	// Process workflows if workflow engine is available
	if h.workflowEngine != nil {
		// Convert AlertManagerEvent to WorkflowEvent
		workflowEvent := event.NewAlertManagerWorkflowEvent(alertEvent)

		matchingWorkflows := h.workflowEngine.SelectWorkflows(workflowEvent)
		h.logger.Info("Workflow processing",
			zap.String("alert_name", alertEvent.GetAlertName()),
			zap.Int("matching_workflows", len(matchingWorkflows)))

		// Execute matching workflows
		for _, workflow := range matchingWorkflows {
			if err := h.workflowEngine.ExecuteWorkflow(c.Request.Context(), workflow, workflowEvent); err != nil {
				h.logger.Error("Failed to execute workflow",
					zap.Error(err),
					zap.String("workflow_name", workflow.Name))
			}
		}
	}

	// Resolve which team should handle this alert
	team, err := h.teamResolver.ResolveTeam(alertEvent)
	if err != nil {
		h.logger.Error("Failed to resolve team for alert", zap.Error(err))
		h.metrics.IncAlertErrors(alertEvent.GetAlertName(), "team_resolution_failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve team"})
		return
	}

	// Convert AlertManagerEvent to Issues
	issues, err := h.converter.ConvertAlertManagerEventToIssues(alertEvent)
	if err != nil {
		h.logger.Error("Failed to convert alert to issues", zap.Error(err))
		h.metrics.IncAlertErrors(alertEvent.GetAlertName(), "conversion_failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to convert alert"})
		return
	}

	// Dispatch issues to team destinations
	ctx := c.Request.Context()
	dispatchErr := h.alertDispatcher.DispatchIssues(ctx, issues, team)
	if dispatchErr != nil {
		h.logger.Error("Failed to dispatch issues", zap.Error(dispatchErr))
		h.metrics.IncAlertErrors(alertEvent.GetAlertName(), "dispatch_failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to dispatch issues"})
		return
	}

	// Record processing metrics
	processingDuration := time.Since(start)
	workflowCount := 0
	if team != nil {
		workflowCount = len(team.Destinations)
	}

	h.metrics.ObserveAlertProcessingDuration(alertEvent.GetAlertName(), workflowCount, processingDuration)

	if team == nil {
		h.logger.Warn("Alert received but no team resolved - alert not processed",
			zap.String("receiver", alertEvent.Receiver),
			zap.String("status", alertEvent.Status),
			zap.Int("alerts_count", len(alertEvent.Alerts)),
			zap.Int("issues_count", len(issues)))
		h.metrics.IncAlertsProcessed(alertEvent.GetAlertName(), alertEvent.GetSeverity(), "no_team_resolved")
	} else if len(team.Destinations) == 0 {
		h.logger.Warn("Alert received for team, but team has no destinations - alert not processed",
			zap.String("receiver", alertEvent.Receiver),
			zap.String("status", alertEvent.Status),
			zap.Int("alerts_count", len(alertEvent.Alerts)),
			zap.Int("issues_count", len(issues)),
			zap.String("team", team.Name))
		h.metrics.IncAlertsProcessed(alertEvent.GetAlertName(), alertEvent.GetSeverity(), "no_destinations")
	} else {
		h.logger.Info("Alert processed successfully",
			zap.String("receiver", alertEvent.Receiver),
			zap.String("status", alertEvent.Status),
			zap.Int("alerts_count", len(alertEvent.Alerts)),
			zap.Int("issues_count", len(issues)),
			zap.String("team", team.Name))
		h.metrics.IncAlertsProcessed(alertEvent.GetAlertName(), alertEvent.GetSeverity(), "processed")
	}

	// Log only essential alert information to avoid memory issues with large alerts
	h.logger.Debug("Alert details",
		zap.String("receiver", alertEvent.Receiver),
		zap.String("status", alertEvent.Status),
		zap.Int("alerts_count", len(alertEvent.Alerts)),
		zap.Int("issues_count", len(issues)),
		zap.String("alert_name", alertEvent.GetAlertName()),
		zap.Any("group_labels", alertEvent.GroupLabels))
	c.JSON(http.StatusOK, gin.H{"status": "alert processed"})
}
