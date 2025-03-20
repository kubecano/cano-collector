package metrics

import (
	"errors"
	"net/http"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsInterface interface {
	PrometheusMiddleware() gin.HandlerFunc
	ObserveAlert(receiver string, status string)
	ClearMetrics()
}

type MetricsCollector struct {
	httpRequestsTotal       *prometheus.CounterVec
	alertManagerAlertsTotal *prometheus.CounterVec
	logger                  logger.LoggerInterface
}

func NewMetricsCollector(log logger.LoggerInterface) *MetricsCollector {
	mc := &MetricsCollector{
		logger: log,
	}

	mc.logger.Debug("Registering Prometheus metrics")
	mc.httpRequestsTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	), "httpRequestsTotal").(*prometheus.CounterVec)

	mc.alertManagerAlertsTotal = mc.registerCollector(prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "alertmanager_alerts_total",
			Help: "Total number of alerts received from AlertManager",
		},
		[]string{"receiver", "status"},
	), "alertManagerAlertsTotal").(*prometheus.CounterVec)

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
		c.Next()
		status := c.Writer.Status()
		mc.httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
		mc.logger.Debugf("Incremented Prometheus counter for %s %s with status %d", c.Request.Method, c.FullPath(), status)
	}
}

func (mc *MetricsCollector) ObserveAlert(receiver string, status string) {
	mc.alertManagerAlertsTotal.WithLabelValues(receiver, status).Inc()
	mc.logger.Debugf("Incremented AlertManager alert counter for receiver: %s, status: %s", receiver, status)
}
