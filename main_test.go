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
	alert_interfaces "github.com/kubecano/cano-collector/pkg/alert/interfaces"
	destination_interfaces "github.com/kubecano/cano-collector/pkg/destination/interfaces"
	health_interfaces "github.com/kubecano/cano-collector/pkg/health/interfaces"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"
	metric_interfaces "github.com/kubecano/cano-collector/pkg/metric/interfaces"
	router_interfaces "github.com/kubecano/cano-collector/pkg/router/interfaces"
	tracer_interfaces "github.com/kubecano/cano-collector/pkg/tracer/interfaces"
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
		LoggerFactory: func(_, _ string) logger_interfaces.LoggerInterface { return mockLogger },
		HealthCheckerFactory: func(cfg config.Config, log logger_interfaces.LoggerInterface) health_interfaces.HealthInterface {
			return mockHealth
		},
		TracerManagerFactory: func(cfg config.Config, log logger_interfaces.LoggerInterface) tracer_interfaces.TracerInterface {
			return mockTracer
		},
		MetricsFactory: func(log logger_interfaces.LoggerInterface) metric_interfaces.MetricsInterface { return mockMetrics },
		DestinationFactory: func(log logger_interfaces.LoggerInterface) destination_interfaces.DestinationFactoryInterface {
			return mockDestinationFactory
		},
		DestinationRegistry: func(factory destination_interfaces.DestinationFactoryInterface, log logger_interfaces.LoggerInterface) destination_interfaces.DestinationRegistryInterface {
			return mockDestinationRegistry
		},
		TeamResolverFactory: func(teams config_team.TeamsConfig, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface) alert_interfaces.TeamResolverInterface {
			return mockTeamResolver
		},
		AlertDispatcherFactory: func(registry destination_interfaces.DestinationRegistryInterface, log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface) alert_interfaces.AlertDispatcherInterface {
			return mockAlertDispatcher
		},
		AlertHandlerFactory: func(log logger_interfaces.LoggerInterface, m metric_interfaces.MetricsInterface, tr alert_interfaces.TeamResolverInterface, ad alert_interfaces.AlertDispatcherInterface) alert_interfaces.AlertHandlerInterface {
			return mockAlerts
		},
		RouterManagerFactory: func(cfg config.Config, log logger_interfaces.LoggerInterface, t tracer_interfaces.TracerInterface, m metric_interfaces.MetricsInterface, h health_interfaces.HealthInterface, a alert_interfaces.AlertHandlerInterface) router_interfaces.RouterInterface {
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
