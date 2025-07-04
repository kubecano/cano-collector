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

type alertHandlerTestDeps struct {
	ctrl            *gomock.Controller
	logger          *mocks.MockLoggerInterface
	teamResolver    *mocks.MockTeamResolverInterface
	alertDispatcher *mocks.MockAlertDispatcherInterface
	handler         *AlertHandler
	router          *gin.Engine
}

func setupTestRouter(t *testing.T) alertHandlerTestDeps {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)

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

	// Happy path - team resolved and alert dispatched
	mockTeam := &config_team.Team{
		Name:         "test-team",
		Destinations: []string{"test-destination"},
	}
	mockTeamResolver.EXPECT().ResolveTeam(gomock.Any()).Return(mockTeam, nil).AnyTimes()
	mockAlertDispatcher.EXPECT().DispatchAlert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	alertHandler := NewAlertHandler(mockLogger, mockMetrics, mockTeamResolver, mockAlertDispatcher)

	r := gin.Default()
	r.POST("/alert", alertHandler.HandleAlert)

	return alertHandlerTestDeps{
		ctrl:            ctrl,
		logger:          mockLogger,
		teamResolver:    mockTeamResolver,
		alertDispatcher: mockAlertDispatcher,
		handler:         alertHandler,
		router:          r,
	}
}

func TestAlertHandler_ValidAlert_HappyPath(t *testing.T) {
	deps := setupTestRouter(t)
	defer deps.ctrl.Finish()

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
	deps.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert processed")
}

func TestAlertHandler_ValidAlert(t *testing.T) {
	deps := setupTestRouter(t)
	defer deps.ctrl.Finish()

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
	deps.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert processed")
}

func TestAlertHandler_InvalidJSON(t *testing.T) {
	deps := setupTestRouter(t)
	defer deps.ctrl.Finish()

	invalidJSON := "{"
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid alert format")
}

func TestAlertHandler_MissingFields(t *testing.T) {
	deps := setupTestRouter(t)
	defer deps.ctrl.Finish()

	alert := `{"receiver": "test-receiver"}`
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(alert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid alert format")
	assert.Contains(t, w.Body.String(), "alert status is required")
}

func TestAlertHandler_EmptyBody(t *testing.T) {
	deps := setupTestRouter(t)
	defer deps.ctrl.Finish()

	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer([]byte{}))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "empty JSON body")
}

func TestAlertHandler_AdditionalFields(t *testing.T) {
	deps := setupTestRouter(t)
	defer deps.ctrl.Finish()

	alert := `{"receiver": "test-receiver", "status": "firing", "alerts": [{"status": "firing", "labels": {"alertname": "HighCPUUsage"}, "startsAt": "2023-01-01T00:00:00Z"}], "extra": "field"}`
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBufferString(alert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert processed")
}

func TestAlertHandler_LargeAlert(t *testing.T) {
	deps := setupTestRouter(t)
	defer deps.ctrl.Finish()

	alerts := make([]template.Alert, 1000)
	for i := range alerts {
		alerts[i] = template.Alert{
			Status: "firing",
			Labels: map[string]string{
				"alertname": "HighCPUUsage",
			},
			StartsAt: time.Now(),
		}
	}

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts:   alerts,
	}

	jsonAlert, _ := json.Marshal(alert)
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer(jsonAlert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	deps.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert processed")
}

func TestAlertHandler_NoTeamResolved(t *testing.T) {
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

	// Edge case - no team resolved
	mockTeamResolver.EXPECT().ResolveTeam(gomock.Any()).Return(nil, nil).AnyTimes()
	mockAlertDispatcher.EXPECT().DispatchAlert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	alertHandler := NewAlertHandler(mockLogger, mockMetrics, mockTeamResolver, mockAlertDispatcher)

	r := gin.Default()
	r.POST("/alert", alertHandler.HandleAlert)

	alert := template.Data{
		Receiver: "test-receiver",
		Status:   "firing",
		Alerts: []template.Alert{
			{
				Status:   "firing",
				Labels:   map[string]string{"alertname": "HighCPUUsage"},
				StartsAt: time.Now(),
			},
		},
	}
	jsonAlert, _ := json.Marshal(alert)
	req, _ := http.NewRequest(http.MethodPost, "/alert", bytes.NewBuffer(jsonAlert))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "alert processed")
}
