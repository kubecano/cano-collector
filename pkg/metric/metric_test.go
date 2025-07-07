package metric

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/metric/interfaces"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/stretchr/testify/assert"
)

func setupTestMetricsCollector(t *testing.T) interfaces.MetricsInterface {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	// Clear existing metrics
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
	prometheus.DefaultGatherer = prometheus.DefaultRegisterer.(*prometheus.Registry)

	metrics := NewMetricsCollector(mockLogger)

	return metrics
}

func TestNewMetricsCollector(t *testing.T) {
	metrics := setupTestMetricsCollector(t)
	assert.NotNil(t, metrics, "MetricsCollector should not be nil")
}

func TestPrometheusMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	metrics := setupTestMetricsCollector(t)

	router := gin.New()
	router.Use(metrics.PrometheusMiddleware())

	router.GET("/ok", func(c *gin.Context) { c.String(http.StatusOK, "OK") })
	router.GET("/not_found", func(c *gin.Context) { c.Status(http.StatusNotFound) })
	router.GET("/server_error", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })

	statusTests := []struct {
		path       string
		statusCode int
	}{
		{"/ok", http.StatusOK},
		{"/not_found", http.StatusNotFound},
		{"/server_error", http.StatusInternalServerError},
	}

	for _, test := range statusTests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, test.path, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, test.statusCode, w.Code, "Expected status code %d for path %s", test.statusCode, test.path)
	}

	time.Sleep(500 * time.Millisecond)

	metricsHandler := promhttp.Handler()
	metricsW := httptest.NewRecorder()
	metricsReq, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	metricsHandler.ServeHTTP(metricsW, metricsReq)

	metricsOutput := metricsW.Body.String()

	t.Logf("Collected metrics:\n%s", metricsOutput)

	assert.Contains(t, metricsOutput, "http_requests_total", "Expected http_requests_total metric")
	expectedMetrics := []string{
		fmt.Sprintf(`method="GET",path="/ok",status="%s"`, http.StatusText(http.StatusOK)),
		fmt.Sprintf(`method="GET",path="/not_found",status="%s"`, http.StatusText(http.StatusNotFound)),
		fmt.Sprintf(`method="GET",path="/server_error",status="%s"`, http.StatusText(http.StatusInternalServerError)),
	}

	for _, expected := range expectedMetrics {
		assert.Contains(t, metricsOutput, expected, "Expected HTTP metric: %s", expected)
	}
}

func TestObserveAlert(t *testing.T) {
	metrics := setupTestMetricsCollector(t)

	metrics.ObserveAlert("test-receiver", "firing")
	metrics.ObserveAlert("test-receiver", "resolved")
	metrics.ObserveAlert("backup-receiver", "firing")
	metrics.ObserveAlert("backup-receiver", "resolved")

	time.Sleep(500 * time.Millisecond)

	metricsHandler := promhttp.Handler()
	metricsW := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	metricsHandler.ServeHTTP(metricsW, req)

	metricsOutput := metricsW.Body.String()

	t.Logf("Collected metrics:\n%s", metricsOutput)

	assert.Contains(t, metricsOutput, "alertmanager_alerts_total", "Expected alertmanager_alerts_total metric")

	expectedMetrics := []string{
		`receiver="test-receiver",status="firing"`,
		`receiver="test-receiver",status="resolved"`,
		`receiver="backup-receiver",status="firing"`,
		`receiver="backup-receiver",status="resolved"`,
	}

	for _, expected := range expectedMetrics {
		assert.Contains(t, metricsOutput, expected, "Expected alert metric: %s", expected)
	}
}
