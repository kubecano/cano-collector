package alert

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/metric"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockTeamResolver := mocks.NewMockTeamResolverInterface(ctrl)
	mockAlertDispatcher := mocks.NewMockAlertDispatcherInterface(ctrl)

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	mockMetrics := metric.NewMetricsCollector(mockLogger)

	// Setup mock expectations for team resolver and alert dispatcher
	mockTeamResolver.EXPECT().ResolveTeam(gomock.Any()).Return(nil, nil).AnyTimes()
	mockAlertDispatcher.EXPECT().DispatchAlert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	alertHandler := NewAlertHandler(mockLogger, mockMetrics, mockTeamResolver, mockAlertDispatcher)

	r := gin.Default()
	r.POST("/alert", alertHandler.HandleAlert)
	return r
}

func TestAlertHandler_ValidAlert(t *testing.T) {
	router := setupTestRouter(t)

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
	assert.Contains(t, w.Body.String(), "alert processed")
}

func TestAlertHandler_InvalidJSON(t *testing.T) {
	router := setupTestRouter(t)

	invalidJSON := "{"
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid alert format")
}

func TestAlertHandler_MissingFields(t *testing.T) {
	router := setupTestRouter(t)

	alert := `{"receiver": "test-receiver"}`
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(alert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid alert format: missing required fields")
}

func TestAlertHandler_EmptyBody(t *testing.T) {
	router := setupTestRouter(t)

	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "empty JSON body")
}

func TestAlertHandler_AdditionalFields(t *testing.T) {
	router := setupTestRouter(t)

	alert := `{"receiver": "test-receiver", "status": "firing", "alerts": [{"status": "firing", "labels": {"alertname": "HighCPUUsage"}}], "extra": "field"}`
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(alert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert processed")
}

func TestAlertHandler_LargeAlert(t *testing.T) {
	router := setupTestRouter(t)

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
	assert.Contains(t, w.Body.String(), "alert processed")
}

func TestAlertHandler_TeamWithoutDestinations(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockTeamResolver := mocks.NewMockTeamResolverInterface(ctrl)
	mockAlertDispatcher := mocks.NewMockAlertDispatcherInterface(ctrl)

	team := &config_team.Team{
		Name:         "team-no-dest",
		Destinations: []string{},
	}

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warn(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	mockMetrics := metric.NewMetricsCollector(mockLogger)

	mockTeamResolver.EXPECT().ResolveTeam(gomock.Any()).Return(team, nil).AnyTimes()
	mockAlertDispatcher.EXPECT().DispatchAlert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	alertHandler := NewAlertHandler(mockLogger, mockMetrics, mockTeamResolver, mockAlertDispatcher)

	r := gin.Default()
	r.POST("/alert", alertHandler.HandleAlert)

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status: "firing",
				Labels: map[string]string{"alertname": "HighCPUUsage"},
			},
		},
	}
	jsonAlert, _ := json.Marshal(alert)
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer(jsonAlert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	// Tu nie sprawdzamy loga bezpośrednio, ale test przejdzie jeśli nie będzie panic i log Warn zostanie wywołany
}
