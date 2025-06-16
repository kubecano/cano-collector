package router

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/mocks"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/alertmanager/template"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/metric"
)

func setupTestRouter(t *testing.T) *RouterManager {
	t.Helper()
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	realLogger := zap.NewNop()
	mockLogger.EXPECT().GetLogger().Return(realLogger).AnyTimes()

	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	mockTracer := mocks.NewMockTracerInterface(ctrl)
	mockAlerts := mocks.NewMockAlertHandlerInterface(ctrl)
	mockAlerts.EXPECT().HandleAlert(gomock.Any()).DoAndReturn(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "alert received"})
	}).AnyTimes()

	mockHealth := mocks.NewMockHealthInterface(ctrl)

	mockMetrics := metric.NewMetricsCollector(mockLogger)

	// Stub minimalne wymagane zachowania
	mockTracer.EXPECT().TraceLoggerMiddleware().AnyTimes().Return(func(c *gin.Context) {})
	mockHealth.EXPECT().RegisterHealthChecks().AnyTimes().Return(nil)
	mockHealth.EXPECT().Handler().AnyTimes().Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	cfg := config.Config{
		AppName:    "cano-collector",
		AppVersion: "1.0.0",
	}

	routerManager := NewRouterManager(cfg, mockLogger, mockTracer, mockMetrics, mockHealth, mockAlerts)

	if routerManager.logger == nil {
		panic("RouterManager.logger is nil!")
	}

	return routerManager
}

func TestStartServer(t *testing.T) {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()

	routerManager := setupTestRouter(t)
	router := routerManager.SetupRouter()

	srv := &http.Server{
		Addr:    ":8081",
		Handler: router,
	}

	var wg sync.WaitGroup
	wg.Add(1)

	serverErrChan := make(chan error, 1)

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// Wait for the server to start
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var serverReady bool
	for {
		resp, err := http.Get("http://127.0.0.1:8081/")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				serverReady = true
				break
			}
		}
		if ctx.Err() != nil {
			t.Fatal("Server did not start in time")
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !serverReady {
		t.Fatal("Server did not start correctly")
	}

	select {
	case err := <-serverErrChan:
		if err != nil {
			t.Fatalf("Server encountered an error: %v", err)
		}
	default:
	}

	// Gracefully shut down the server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	shutdownErrChan := make(chan error, 1)
	go func() {
		shutdownErrChan <- srv.Shutdown(shutdownCtx)
	}()

	// Wait for server shutdown
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		t.Fatal("Server did not shut down in time")
	case <-done:
		t.Log("Server shut down successfully")
	case err := <-shutdownErrChan:
		if err != nil {
			t.Fatalf("Server shutdown failed: %v", err)
		}
	}
}

func TestHelloWorld(t *testing.T) {
	routerManager := setupTestRouter(t)
	assert.NotNil(t, routerManager, "RouterManager should not be nil")

	router := routerManager.SetupRouter()
	assert.NotNil(t, router, "Router should not be nil")

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "Hello world!", w.Body.String())
}

func TestHealthEndpoints(t *testing.T) {
	routerManager := setupTestRouter(t)
	router := routerManager.SetupRouter()

	endpoints := []string{"/livez", "/readyz", "/healthz"}

	for _, endpoint := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, endpoint, nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "Expected 200 OK for endpoint %s", endpoint)
	}
}

func TestApiAlertsEndpoint(t *testing.T) {
	routerManager := setupTestRouter(t)
	router := routerManager.SetupRouter()

	w := httptest.NewRecorder()
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
	req, _ := http.NewRequest(http.MethodPost, "/api/alerts", bytes.NewBuffer(jsonAlert))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status": "alert received"}`, w.Body.String())
}

func TestMetricsEndpoint(t *testing.T) {
	routerManager := setupTestRouter(t)
	router := routerManager.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "go_goroutines")
}

func TestMetricsEndpoint_Uninitialized(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)

	realLogger := zap.NewNop()
	mockLogger.EXPECT().GetLogger().Return(realLogger).AnyTimes()

	mockLogger.EXPECT().Error(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debugf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()

	mockTracer := mocks.NewMockTracerInterface(ctrl)
	mockAlerts := mocks.NewMockAlertHandlerInterface(ctrl)
	mockHealth := mocks.NewMockHealthInterface(ctrl)

	mockMetrics := metric.NewMetricsCollector(mockLogger)

	// Stub minimalne wymagane zachowania
	mockTracer.EXPECT().TraceLoggerMiddleware().AnyTimes().Return(func(c *gin.Context) {})
	mockHealth.EXPECT().RegisterHealthChecks().AnyTimes().Return(nil)
	mockHealth.EXPECT().Handler().AnyTimes().Return(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	emptyRegistry := prometheus.NewRegistry()
	prometheus.DefaultRegisterer = emptyRegistry
	prometheus.DefaultGatherer = emptyRegistry

	mockMetrics.ClearMetrics()

	cfg := config.Config{
		AppName:    "cano-collector",
		AppVersion: "1.0.0",
	}

	routerManager := NewRouterManager(cfg, mockLogger, mockTracer, mockMetrics, mockHealth, mockAlerts)
	router := routerManager.SetupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code, "Expected 500 Internal Server Error when Prometheus is not initialized")
	assert.Contains(t, w.Body.String(), "Prometheus collector not initialized", "Expected error message in response body")
}
