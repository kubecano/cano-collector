package metrics

import (
	"errors"
	"net/http"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var httpRequestsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Total number of HTTP requests",
	},
	[]string{"method", "path", "status"},
)

var alertManagerAlertsTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "alertmanager_alerts_total",
		Help: "Total number of alerts received from AlertManager",
	},
	[]string{"receiver", "status"},
)

func registerCollector(collector prometheus.Collector, name string) {
	if err := prometheus.Register(collector); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			logger.Warnf("%s collector already registered: %v", name, are)
			switch v := collector.(type) {
			case *prometheus.CounterVec:
				*v = *are.ExistingCollector.(*prometheus.CounterVec)
			case *prometheus.GaugeVec:
				*v = *are.ExistingCollector.(*prometheus.GaugeVec)
			default:
				logger.Warnf("Unknown collector type for %s: %T", name, v)
			}
		} else {
			logger.Errorf("Failed to register %s collector: %v", name, err)
		}
	} else {
		logger.Debugf("%s collector registered successfully", name)
	}
}

func RegisterMetrics() {
	logger.Debug("Registering Prometheus metrics")
	registerCollector(httpRequestsTotal, "httpRequestsTotal")
	registerCollector(alertManagerAlertsTotal, "alertManagerAlertsTotal")
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		status := c.Writer.Status()
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
		logger.Debugf("Incremented Prometheus counter for %s %s with status %d", c.Request.Method, c.FullPath(), status)
	}
}

func ObserveAlert(receiver string, status string) {
	alertManagerAlertsTotal.WithLabelValues(receiver, status).Inc()
	logger.Debugf("Incremented AlertManager alert counter for receiver: %s, status: %s", receiver, status)
}
