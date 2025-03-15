package alerts

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metrics"
	"github.com/prometheus/alertmanager/template"
	"go.uber.org/zap"
	"io"
	"net/http"
)

// AlertHandler handles incoming alerts from Alertmanager
func AlertHandler(c *gin.Context) {
	// Check if the request body is empty
	body, err := io.ReadAll(c.Request.Body)
	if err != nil || len(bytes.TrimSpace(body)) == 0 {
		logger.Error("Empty or invalid request body")
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty or invalid JSON body"})
		return
	}

	// Restore the body for JSON binding
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var alert template.Data
	if err := c.ShouldBindJSON(&alert); err != nil {
		logger.Error("Failed to parse alert", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert format"})
		return
	}

	// Validate the parsed alert
	if alert.Receiver == "" || alert.Status == "" || len(alert.Alerts) == 0 {
		logger.Error("Invalid alert structure: missing required fields")
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid alert format: missing required fields"})
		return
	}

	// Register received alert metric
	metrics.ObserveAlert(alert.Receiver, alert.Status)

	// Wrap the alert in EnrichedAlert for future extension
	enrichedAlert := EnrichedAlert{Original: alert}

	// TODO: Dispatch alert using different strategies (e.g., Slack, Teams, OpsGenie)
	//  This will be implemented in the next tasks

	logger.Info("Received alert", zap.Any("alert", enrichedAlert))
	c.JSON(http.StatusOK, gin.H{"status": "alert received"})
}
