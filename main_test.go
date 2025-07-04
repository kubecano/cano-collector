package main

import (
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/getsentry/sentry-go"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/kubecano/cano-collector/config"
	config_team "github.com/kubecano/cano-collector/config/team"
	"github.com/kubecano/cano-collector/mocks"
	"github.com/kubecano/cano-collector/pkg/alert"
	"github.com/kubecano/cano-collector/pkg/health"
	"github.com/kubecano/cano-collector/pkg/interfaces"
	"github.com/kubecano/cano-collector/pkg/logger"
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
	mockDestinationFactory := mocks.NewMockDestinationFactoryInterface(ctrl)
	mockDestinationRegistry := mocks.NewMockDestinationRegistryInterface(ctrl)
	mockTeamResolver := mocks.NewMockTeamResolverInterface(ctrl)
	mockAlertDispatcher := mocks.NewMockAlertDispatcherInterface(ctrl)

	// Mock zachowania
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Panicf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Fatalf(gomock.Any(), gomock.Any()).AnyTimes()

	mockHealth.EXPECT().RegisterHealthChecks().Return(nil).Times(1)

	mockTracer.EXPECT().InitTracer(gomock.Any()).Return(nil).Times(1)
	mockTracer.EXPECT().ShutdownTracer(gomock.Any()).Return(nil).Times(1)

	mockDestinationRegistry.EXPECT().LoadFromConfig(gomock.Any()).Return(nil).Times(1)

	mockTeamResolver.EXPECT().ValidateTeamDestinations(gomock.Any()).Return(nil).Times(1)

	// Mock DestinationFactory - oczekujemy, że może być wywołany podczas ładowania konfiguracji
	mockDestinationFactory.EXPECT().CreateDestination(gomock.Any()).AnyTimes()

	g := gin.New()
	mockRouter.EXPECT().SetupRouter().Return(g).Times(1)
	mockRouter.EXPECT().StartServer(g).Times(1)

	deps := AppDependencies{
		LoggerFactory:        func(_, _ string) logger.LoggerInterface { return mockLogger },
		HealthCheckerFactory: func(cfg config.Config, log logger.LoggerInterface) health.HealthInterface { return mockHealth },
		TracerManagerFactory: func(cfg config.Config, log logger.LoggerInterface) tracer.TracerInterface { return mockTracer },
		MetricsFactory:       func(log logger.LoggerInterface) interfaces.MetricsInterface { return mockMetrics },
		DestinationFactory:   func(log logger.LoggerInterface) interfaces.DestinationFactoryInterface { return mockDestinationFactory },
		DestinationRegistry: func(factory interfaces.DestinationFactoryInterface, log logger.LoggerInterface) interfaces.DestinationRegistryInterface {
			return mockDestinationRegistry
		},
		TeamResolverFactory: func(teams config_team.TeamsConfig, log logger.LoggerInterface, m interfaces.MetricsInterface) interfaces.TeamResolverInterface {
			return mockTeamResolver
		},
		AlertDispatcherFactory: func(registry interfaces.DestinationRegistryInterface, log logger.LoggerInterface, m interfaces.MetricsInterface) interfaces.AlertDispatcherInterface {
			return mockAlertDispatcher
		},
		AlertHandlerFactory: func(log logger.LoggerInterface, m interfaces.MetricsInterface, tr interfaces.TeamResolverInterface, ad interfaces.AlertDispatcherInterface) alert.AlertHandlerInterface {
			return mockAlerts
		},
		RouterManagerFactory: func(cfg config.Config, log logger.LoggerInterface, t tracer.TracerInterface, m interfaces.MetricsInterface, h health.HealthInterface, a alert.AlertHandlerInterface) router.RouterInterface {
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
