package main

import (
	"context"
	"time"

	"github.com/kubecano/cano-collector/pkg/alert"
	"github.com/kubecano/cano-collector/pkg/metric"

	"github.com/kubecano/cano-collector/pkg/router"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/kubecano/cano-collector/pkg/health"

	"github.com/kubecano/cano-collector/pkg/tracer"

	"github.com/getsentry/sentry-go"

	"github.com/kubecano/cano-collector/config"
)

type AppDependencies struct {
	LoggerFactory        func(level string, env string) logger.LoggerInterface
	HealthCheckerFactory func(cfg config.Config, log logger.LoggerInterface) health.HealthInterface
	TracerManagerFactory func(cfg config.Config, log logger.LoggerInterface) tracer.TracerInterface
	MetricsFactory       func(log logger.LoggerInterface) metric.MetricsInterface
	AlertHandlerFactory  func(log logger.LoggerInterface, m metric.MetricsInterface) alert.AlertHandlerInterface
	RouterManagerFactory func(cfg config.Config, log logger.LoggerInterface, t tracer.TracerInterface, m metric.MetricsInterface, h health.HealthInterface, a alert.AlertHandlerInterface) router.RouterInterface
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Error loading config: " + err.Error())
	}

	deps := AppDependencies{
		LoggerFactory: func(level, env string) logger.LoggerInterface {
			return logger.NewLogger(level, env)
		},
		HealthCheckerFactory: func(cfg config.Config, log logger.LoggerInterface) health.HealthInterface {
			return health.NewHealthChecker(cfg, log)
		},
		TracerManagerFactory: func(cfg config.Config, log logger.LoggerInterface) tracer.TracerInterface {
			return tracer.NewTracerManager(cfg, log)
		},
		MetricsFactory: func(log logger.LoggerInterface) metric.MetricsInterface {
			return metric.NewMetricsCollector(log)
		},
		AlertHandlerFactory: func(log logger.LoggerInterface, m metric.MetricsInterface) alert.AlertHandlerInterface {
			return alert.NewAlertHandler(log, m)
		},
		RouterManagerFactory: func(cfg config.Config, log logger.LoggerInterface, t tracer.TracerInterface, m metric.MetricsInterface, h health.HealthInterface, a alert.AlertHandlerInterface) router.RouterInterface {
			return router.NewRouterManager(cfg, log, t, m, h, a)
		},
	}

	if err := run(cfg, deps); err != nil {
		panic("Error running app: " + err.Error())
	}
}

func run(cfg config.Config, deps AppDependencies) error {
	log := deps.LoggerFactory(cfg.LogLevel, cfg.AppEnv)
	log.Debug("Logger initialized")

	healthChecker := deps.HealthCheckerFactory(cfg, log)
	err := healthChecker.RegisterHealthChecks()
	if err != nil {
		log.Panicf("Failed to register health checks: %v", err)
		return err
	}
	log.Debug("Health checks registered")

	tracerManager := deps.TracerManagerFactory(cfg, log)
	metricsCollector := deps.MetricsFactory(log)
	alertHandler := deps.AlertHandlerFactory(log, metricsCollector)
	routerManager := deps.RouterManagerFactory(cfg, log, tracerManager, metricsCollector, healthChecker, alertHandler)

	if cfg.SentryEnabled {
		if err := initSentry(cfg.SentryDSN); err != nil {
			log.Fatalf("Sentry initialization failed: %v", err)
		}
		log.Debug("Sentry initialized")
	} else {
		log.Debug("Sentry is disabled")
	}

	defer sentry.Flush(2 * time.Second)

	ctx := context.Background()
	err = tracerManager.InitTracer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}

	defer func(tracerManager tracer.TracerInterface, ctx context.Context) {
		err := tracerManager.ShutdownTracer(ctx)
		if err != nil {
			log.Fatalf("Failed to shutdown tracing: %v", err)
		}
	}(tracerManager, ctx)

	r := routerManager.SetupRouter()
	log.Debug("Router setup complete")
	routerManager.StartServer(r)

	return nil
}

func initSentry(sentryDSN string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
}
