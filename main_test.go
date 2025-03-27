package main

import (
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/getsentry/sentry-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/alerts"
	"github.com/kubecano/cano-collector/pkg/health"
	"github.com/kubecano/cano-collector/pkg/logger"
	"github.com/kubecano/cano-collector/pkg/metrics"
	"github.com/kubecano/cano-collector/pkg/router"
	"github.com/kubecano/cano-collector/pkg/tracer"
)

func resetSentryState() {
	sentry.Flush(0)
	_ = sentry.Init(sentry.ClientOptions{Dsn: ""})
}

func TestInitSentry_Success(t *testing.T) {
	defer resetSentryState()

	err := initSentry("https://xxx@yyy.example.com/111")
	assert.NoError(t, err, "Expected no error when DSN is valid")
}

func TestInitSentry_Fail(t *testing.T) {
	defer resetSentryState()

	err := initSentry("invalid-dsn")
	assert.Error(t, err, "Expected an error when DSN is invalid")
}

func TestInitSentry_Disabled(t *testing.T) {
	defer resetSentryState()

	err := initSentry("")
	assert.NoError(t, err, "Expected no error when Sentry DSN is empty")
}

func TestRun_WithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockHealth := mocks.NewMockHealthInterface(ctrl)
	mockTracer := mocks.NewMockTracerInterface(ctrl)
	mockMetrics := mocks.NewMockMetricsInterface(ctrl)
	mockAlerts := mocks.NewMockAlertHandlerInterface(ctrl)
	mockRouter := mocks.NewMockRouterInterface(ctrl)

	// Mock zachowania
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Panicf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Fatalf(gomock.Any(), gomock.Any()).AnyTimes()

	mockHealth.EXPECT().RegisterHealthChecks().Return(nil).Times(1)

	mockTracer.EXPECT().InitTracer(gomock.Any()).Return(nil).Times(1)
	mockTracer.EXPECT().ShutdownTracer(gomock.Any()).Return(nil).Times(1)

	g := gin.New()
	mockRouter.EXPECT().SetupRouter().Return(g).Times(1)
	mockRouter.EXPECT().StartServer(g).Times(1)

	deps := AppDependencies{
		LoggerFactory:        func(_, _ string) logger.LoggerInterface { return mockLogger },
		HealthCheckerFactory: func(cfg config.Config, log logger.LoggerInterface) health.HealthInterface { return mockHealth },
		TracerManagerFactory: func(cfg config.Config, log logger.LoggerInterface) tracer.TracerInterface { return mockTracer },
		MetricsFactory:       func(log logger.LoggerInterface) metrics.MetricsInterface { return mockMetrics },
		AlertHandlerFactory: func(log logger.LoggerInterface, m metrics.MetricsInterface) alerts.AlertHandlerInterface {
			return mockAlerts
		},
		RouterManagerFactory: func(cfg config.Config, log logger.LoggerInterface, t tracer.TracerInterface, m metrics.MetricsInterface, h health.HealthInterface, a alerts.AlertHandlerInterface) router.RouterInterface {
			return mockRouter
		},
	}

	cfg := config.Config{
		AppName:       "cano-collector",
		AppEnv:        "test",
		LogLevel:      "debug",
		SentryEnabled: false,
	}

	err := run(cfg, deps)
	assert.NoError(t, err)
}
