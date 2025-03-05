package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterMetrics(t *testing.T) {
	assert.NotPanics(t, func() {
		RegisterMetrics()
	})
}

//func TestPrometheusMiddleware(t *testing.T) {
//	gin.SetMode(gin.TestMode)
//	router := gin.New()
//	RegisterMetrics()
//
//	router.Use(PrometheusMiddleware())
//
//	router.GET("/test", func(c *gin.Context) {
//		c.String(http.StatusOK, "test response")
//	})
//
//	w := httptest.NewRecorder()
//	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
//	router.ServeHTTP(w, req)
//
//	assert.Equal(t, http.StatusOK, w.Code)
//
//	metricsHandler := promhttp.Handler()
//	metricsW := httptest.NewRecorder()
//	metricsReq, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
//	metricsHandler.ServeHTTP(metricsW, metricsReq)
//
//	assert.Contains(t, metricsW.Body.String(), "http_requests_total")
//}
