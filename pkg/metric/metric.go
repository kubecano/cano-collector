package metric

import (
	"errors"
	"net/http"
	"time"

	"github.com/kubecano/cano-collector/pkg/interfaces"
	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsCollector struct {
	httpRequestsTotal            *prometheus.CounterVec
	httpRequestDuration          *prometheus.HistogramVec
	alertManagerAlertsTotal      *prometheus.CounterVec
	alertsProcessedTotal         *prometheus.CounterVec
	alertsProcessingDuration     *prometheus.HistogramVec
	alertsErrorsTotal            *prometheus.CounterVec
	destinationMessagesSentTotal *prometheus.CounterVec
	destinationSendDuration      *prometheus.HistogramVec
	destinationErrorsTotal       *prometheus.CounterVec
	routingDecisionsTotal        *prometheus.CounterVec
	teamsMatchedTotal            *prometheus.CounterVec
	logger                       logger.LoggerInterface
}

func NewMetricsCollector(log logger.LoggerInterface) interfaces.MetricsInterface {
	mc := &MetricsCollector{
		logger: log,
	}

	mc.logger.Debug("Registering Prometheus metrics")

	// Existing metrics
	mc.httpRequestsTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	), "httpRequestsTotal").(*prometheus.CounterVec)

	mc.httpRequestDuration = mc.registerCollector(prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cano_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	), "httpRequestDuration").(*prometheus.HistogramVec)

	mc.alertManagerAlertsTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alertmanager_alerts_total",
			Help: "Total number of alerts received from AlertManager",
		},
		[]string{"receiver", "status"},
	), "alertManagerAlertsTotal").(*prometheus.CounterVec)

	// New alert processing metrics
	mc.alertsProcessedTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_alerts_processed_total",
			Help: "Total number of alerts processed",
		},
		[]string{"alert_name", "severity", "source"},
	), "alertsProcessedTotal").(*prometheus.CounterVec)

	mc.alertsProcessingDuration = mc.registerCollector(prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cano_alerts_processing_duration_seconds",
			Help:    "Time spent processing alerts in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"alert_name", "workflow_count"},
	), "alertsProcessingDuration").(*prometheus.HistogramVec)

	mc.alertsErrorsTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_alerts_errors_total",
			Help: "Total number of alert processing errors",
		},
		[]string{"alert_name", "error_type"},
	), "alertsErrorsTotal").(*prometheus.CounterVec)

	// New destination metrics
	mc.destinationMessagesSentTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_destination_messages_sent_total",
			Help: "Total number of messages sent to destinations",
		},
		[]string{"destination_name", "destination_type", "status"},
	), "destinationMessagesSentTotal").(*prometheus.CounterVec)

	mc.destinationSendDuration = mc.registerCollector(prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cano_destination_send_duration_seconds",
			Help:    "Time spent sending messages to destinations in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"destination_name", "destination_type"},
	), "destinationSendDuration").(*prometheus.HistogramVec)

	mc.destinationErrorsTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_destination_errors_total",
			Help: "Total number of destination send errors",
		},
		[]string{"destination_name", "destination_type", "error_type"},
	), "destinationErrorsTotal").(*prometheus.CounterVec)

	// New routing metrics
	mc.routingDecisionsTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_routing_decisions_total",
			Help: "Total number of routing decisions made",
		},
		[]string{"team_name", "destination_type", "decision"},
	), "routingDecisionsTotal").(*prometheus.CounterVec)

	mc.teamsMatchedTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cano_teams_matched_total",
			Help: "Total number of team matches for alerts",
		},
		[]string{"team_name", "alert_name"},
	), "teamsMatchedTotal").(*prometheus.CounterVec)

	return mc
}

func (mc *MetricsCollector) ClearMetrics() {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	prometheus.DefaultGatherer = prometheus.NewRegistry()
}

func (mc *MetricsCollector) registerCollector(collector prometheus.Collector, name string) prometheus.Collector {
	if err := prometheus.Register(collector); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			mc.logger.Warnf("%s collector already registered: %v", name, are)
			return are.ExistingCollector
		} else {
			mc.logger.Errorf("Failed to register %s collector: %v", name, err)
			return nil
		}
	}
	mc.logger.Debugf("%s collector registered successfully", name)
	return collector
}

func (mc *MetricsCollector) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start)
		status := c.Writer.Status()

		mc.httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
		mc.httpRequestDuration.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Observe(duration.Seconds())

		mc.logger.Debugf("Incremented Prometheus counter for %s %s with status %d, duration: %v", c.Request.Method, c.FullPath(), status, duration)
	}
}

func (mc *MetricsCollector) ObserveAlert(receiver string, status string) {
	mc.alertManagerAlertsTotal.WithLabelValues(receiver, status).Inc()
	mc.logger.Debugf("Incremented AlertManager alert counter for receiver: %s, status: %s", receiver, status)
}

// Alert processing metrics implementations
func (mc *MetricsCollector) IncAlertsProcessed(alertName, severity, source string) {
	mc.alertsProcessedTotal.WithLabelValues(alertName, severity, source).Inc()
	mc.logger.Debugf("Incremented alerts processed counter for alert: %s, severity: %s, source: %s", alertName, severity, source)
}

func (mc *MetricsCollector) ObserveAlertProcessingDuration(alertName string, workflowCount int, duration time.Duration) {
	workflowCountStr := "0"
	if workflowCount > 0 {
		workflowCountStr = "1+"
	}
	mc.alertsProcessingDuration.WithLabelValues(alertName, workflowCountStr).Observe(duration.Seconds())
	mc.logger.Debugf("Observed alert processing duration for alert: %s, workflows: %d, duration: %v", alertName, workflowCount, duration)
}

func (mc *MetricsCollector) IncAlertErrors(alertName, errorType string) {
	mc.alertsErrorsTotal.WithLabelValues(alertName, errorType).Inc()
	mc.logger.Debugf("Incremented alert errors counter for alert: %s, error_type: %s", alertName, errorType)
}

// Destination metrics implementations
func (mc *MetricsCollector) IncDestinationMessagesSent(destinationName, destinationType, status string) {
	mc.destinationMessagesSentTotal.WithLabelValues(destinationName, destinationType, status).Inc()
	mc.logger.Debugf("Incremented destination messages sent counter for destination: %s, type: %s, status: %s", destinationName, destinationType, status)
}

func (mc *MetricsCollector) ObserveDestinationSendDuration(destinationName, destinationType string, duration time.Duration) {
	mc.destinationSendDuration.WithLabelValues(destinationName, destinationType).Observe(duration.Seconds())
	mc.logger.Debugf("Observed destination send duration for destination: %s, type: %s, duration: %v", destinationName, destinationType, duration)
}

func (mc *MetricsCollector) IncDestinationErrors(destinationName, destinationType, errorType string) {
	mc.destinationErrorsTotal.WithLabelValues(destinationName, destinationType, errorType).Inc()
	mc.logger.Debugf("Incremented destination errors counter for destination: %s, type: %s, error_type: %s", destinationName, destinationType, errorType)
}

// HTTP metrics implementations
func (mc *MetricsCollector) ObserveHTTPRequestDuration(method, path, status string, duration time.Duration) {
	mc.httpRequestDuration.WithLabelValues(method, path, status).Observe(duration.Seconds())
	mc.logger.Debugf("Observed HTTP request duration for %s %s %s, duration: %v", method, path, status, duration)
}

// Routing metrics implementations
func (mc *MetricsCollector) IncRoutingDecisions(teamName, destinationType, decision string) {
	mc.routingDecisionsTotal.WithLabelValues(teamName, destinationType, decision).Inc()
	mc.logger.Debugf("Incremented routing decisions counter for team: %s, destination_type: %s, decision: %s", teamName, destinationType, decision)
}

func (mc *MetricsCollector) IncTeamsMatched(teamName, alertName string) {
	mc.teamsMatchedTotal.WithLabelValues(teamName, alertName).Inc()
	mc.logger.Debugf("Incremented teams matched counter for team: %s, alert: %s", teamName, alertName)
}
