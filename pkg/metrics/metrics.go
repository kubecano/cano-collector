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
	if err := prometheus.Register(httpRequestsTotal); err != nil {
		var are prometheus.AlreadyRegisteredError
		if errors.As(err, &are) {
			logger.Errorf("Prometheus collector already registered: %v", are)
			httpRequestsTotal = are.ExistingCollector.(*prometheus.CounterVec)
		}
	}
}

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		status := c.Writer.Status()
		httpRequestsTotal.WithLabelValues(c.Request.Method, c.FullPath(), http.StatusText(status)).Inc()
	}
}
