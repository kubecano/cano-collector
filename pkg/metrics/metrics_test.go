package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/stretchr/testify/assert"
)

func TestRegisterMetrics(t *testing.T) {
	l, _ := zap.NewDevelopment()
	logger.SetLogger(l)
	assert.NotPanics(t, func() {
		RegisterMetrics()
	})
}

func TestPrometheusMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	l, _ := zap.NewDevelopment()
	logger.SetLogger(l)
	router := gin.New()

	// Reset the default Prometheus registry
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	RegisterMetrics()

	router.Use(PrometheusMiddleware())

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "test response")
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	metricsHandler := promhttp.Handler()
	metricsW := httptest.NewRecorder()
	metricsReq, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	metricsHandler.ServeHTTP(metricsW, metricsReq)

	assert.Contains(t, metricsW.Body.String(), "http_requests_total")
}
