package alert

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metric"
)

//go:generate mockgen -destination=../../mocks/alert_handler_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/alert AlertHandlerInterface
type AlertHandlerInterface interface {
	HandleAlert(c *gin.Context)
}

// AlertHandler handles incoming alerts from Alertmanager
type AlertHandler struct {
	logger          logger.LoggerInterface
	metrics         metric.MetricsInterface
	teamResolver    TeamResolverInterface
	alertDispatcher AlertDispatcherInterface
}

// NewAlertHandler creates a new handler with dependencies
func NewAlertHandler(
	logger logger.LoggerInterface,
	metrics metric.MetricsInterface,
	teamResolver TeamResolverInterface,
	alertDispatcher AlertDispatcherInterface,
) *AlertHandler {
	return &AlertHandler{
		logger:          logger,
		metrics:         metrics,
		teamResolver:    teamResolver,
		alertDispatcher: alertDispatcher,
	}
}

// HandleAlert processes alerts
func (h *AlertHandler) HandleAlert(c *gin.Context) {
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

	var alert template.Data
	if err := c.ShouldBindJSON(&alert); err != nil {
		h.logger.Error("Failed to parse alert", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert format"})
		return
	}

	// Validate the parsed alert
	if alert.Receiver == "" || alert.Status == "" || len(alert.Alerts) == 0 {
		h.logger.Error("Invalid alert structure: missing required fields", zap.Any("alert", alert))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert format: missing required fields"})
		return
	}

	// Register received alert metric
	h.metrics.ObserveAlert(alert.Receiver, alert.Status)

	// Resolve which team should handle this alert
	team, err := h.teamResolver.ResolveTeam(alert)
	if err != nil {
		h.logger.Error("Failed to resolve team for alert", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve team"})
		return
	}

	// Dispatch alert to team destinations
	ctx := c.Request.Context()
	dispatchErr := h.alertDispatcher.DispatchAlert(ctx, alert, team)
	if dispatchErr != nil {
		h.logger.Error("Failed to dispatch alert", zap.Error(dispatchErr))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to dispatch alert"})
		return
	}

	if team == nil {
		h.logger.Warn("Alert received but no team resolved - alert not processed",
			zap.String("receiver", alert.Receiver),
			zap.String("status", alert.Status),
			zap.Int("alerts_count", len(alert.Alerts)))
	} else if len(team.Destinations) == 0 {
		h.logger.Warn("Alert received for team, but team has no destinations - alert not processed",
			zap.String("receiver", alert.Receiver),
			zap.String("status", alert.Status),
			zap.Int("alerts_count", len(alert.Alerts)),
			zap.String("team", team.Name))
	} else {
		h.logger.Info("Alert processed successfully",
			zap.String("receiver", alert.Receiver),
			zap.String("status", alert.Status),
			zap.Int("alerts_count", len(alert.Alerts)),
			zap.String("team", team.Name))
	}

	h.logger.Info("Processed alert: ", zap.Any("alert", alert))
	c.JSON(http.StatusOK, gin.H{"status": "alert processed"})
}
