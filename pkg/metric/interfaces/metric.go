package interfaces

import (
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsInterface defines the interface for collecting and exposing application metrics.
//
//go:generate mockgen -source=metric.go -destination=../../../mocks/metrics_mock.go -package=mocks
type MetricsInterface interface {
	PrometheusMiddleware() gin.HandlerFunc
	ObserveAlert(receiver string, status string)
	ClearMetrics()

	// Alert processing metrics
	IncAlertsProcessed(alertName, severity, source string)
	ObserveAlertProcessingDuration(alertName string, workflowCount int, duration time.Duration)
	IncAlertErrors(alertName, errorType string)

	// Destination metrics
	IncDestinationMessagesSent(destinationName, destinationType, status string)
	ObserveDestinationSendDuration(destinationName, destinationType string, duration time.Duration)
	IncDestinationErrors(destinationName, destinationType, errorType string)

	// HTTP metrics
	ObserveHTTPRequestDuration(method, path, status string, duration time.Duration)

	// Routing metrics
	IncRoutingDecisions(teamName, destinationType, decision string)
	IncTeamsMatched(teamName, alertName string)

	// Workflow metrics
	IncWorkflowsExecuted(workflowName, status string)
	ObserveWorkflowEnrichments(workflowName string, enrichmentCount int)
	IncWorkflowEnrichmentErrors(workflowName, errorType string)
}
