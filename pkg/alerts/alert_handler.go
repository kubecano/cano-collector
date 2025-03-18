package alerts

import (
	"bytes"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metrics"
)

// AlertHandler handles incoming alerts from Alertmanager
type AlertHandler struct {
	logger  *logger.Logger
	metrics *metrics.MetricsCollector
}

// NewAlertHandler creates a new handler with dependencies
func NewAlertHandler(logger *logger.Logger, metrics *metrics.MetricsCollector) *AlertHandler {
	return &AlertHandler{logger: logger, metrics: metrics}
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

	// Wrap the alert in EnrichedAlert for future extension
	enrichedAlert := EnrichedAlert{Original: alert}

	// TODO: Dispatch alert using different strategies (e.g., Slack, Teams, OpsGenie)
	//  This will be implemented in the next tasks

	h.logger.Info("Received alert: ", zap.Any("alert", enrichedAlert))
	c.JSON(http.StatusOK, gin.H{"status": "alert received"})
}
