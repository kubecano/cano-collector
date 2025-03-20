package alerts

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/pkg/logger"
)

type MockMetricsCollector struct{}

func (m *MockMetricsCollector) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {}
}

func (m *MockMetricsCollector) ObserveAlert(receiver string, status string) {}

func (m *MockMetricsCollector) ClearMetrics() {}

func setupTestRouter(alertHandler AlertHandlerInterface) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.POST("/alert", alertHandler.HandleAlert)
	return r
}

func TestAlertHandler_ValidAlert(t *testing.T) {
	log := logger.NewMockLogger()
	mockMetrics := &MockMetricsCollector{}
	alertHandler := NewAlertHandler(log, mockMetrics)

	router := setupTestRouter(alertHandler)

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{
					"alertname": "HighCPUUsage",
					"severity":  "critical",
				},
				Annotations: map[string]string{
					"summary":     "High CPU usage detected",
					"description": "The CPU usage has exceeded the threshold",
				},
				StartsAt: time.Now(),
			},
		},
	}

	jsonAlert, _ := json.Marshal(alert)
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer(jsonAlert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert received")
}

func TestAlertHandler_InvalidJSON(t *testing.T) {
	log := logger.NewMockLogger()
	mockMetrics := &MockMetricsCollector{}
	alertHandler := NewAlertHandler(log, mockMetrics)

	router := setupTestRouter(alertHandler)

	invalidJSON := "{"
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid alert format")
}

func TestAlertHandler_MissingFields(t *testing.T) {
	log := logger.NewMockLogger()
	mockMetrics := &MockMetricsCollector{}
	alertHandler := NewAlertHandler(log, mockMetrics)

	router := setupTestRouter(alertHandler)

	alert := `{"receiver": "test-receiver"}`
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(alert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid alert format: missing required fields")
}

func TestAlertHandler_EmptyBody(t *testing.T) {
	log := logger.NewMockLogger()
	mockMetrics := &MockMetricsCollector{}
	alertHandler := NewAlertHandler(log, mockMetrics)

	router := setupTestRouter(alertHandler)

	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "empty JSON body")
}

func TestAlertHandler_AdditionalFields(t *testing.T) {
	log := logger.NewMockLogger()
	mockMetrics := &MockMetricsCollector{}
	alertHandler := NewAlertHandler(log, mockMetrics)

	router := setupTestRouter(alertHandler)

	alert := `{"receiver": "test-receiver", "status": "firing", "alerts": [{"status": "firing", "labels": {"alertname": "HighCPUUsage"}}], "extra": "field"}`
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(alert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert received")
}

func TestAlertHandler_LargeAlert(t *testing.T) {
	log := logger.NewMockLogger()
	mockMetrics := &MockMetricsCollector{}
	alertHandler := NewAlertHandler(log, mockMetrics)

	router := setupTestRouter(alertHandler)

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   make([]template.Alert, 1000),
	}

	jsonAlert, _ := json.Marshal(alert)
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer(jsonAlert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert received")
}
