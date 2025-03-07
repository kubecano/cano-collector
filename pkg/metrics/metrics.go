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

func RegisterMetrics() {
	logger.Debug("Registering Prometheus metrics")
	if err := prometheus.Register(httpRequestsTotal); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			logger.Warnf("Prometheus collector already registered: %v", are)
			httpRequestsTotal = are.ExistingCollector.(*prometheus.CounterVec)
		} else {
			logger.Errorf("Failed to register Prometheus collector: %v", err)
		}
	} else {
		logger.Debug("Prometheus metrics registered successfully")
	}
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		status := c.Writer.Status()
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
		logger.Debugf("Incremented Prometheus counter for %s %s with status %d", c.Request.Method, c.FullPath(), status)
	}
}
