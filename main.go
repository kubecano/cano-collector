package main

import (
	"context"
	"github.com/kubecano/cano-collector/pkg/alerts"
	"github.com/kubecano/cano-collector/pkg/metrics"
	"time"

	"github.com/kubecano/cano-collector/pkg/router"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/kubecano/cano-collector/pkg/health"

	"github.com/kubecano/cano-collector/pkg/tracer"

	"github.com/getsentry/sentry-go"

	"github.com/kubecano/cano-collector/config"
)

func main() {
	cfg := config.LoadConfig()

	log, err := logger.NewLogger(cfg.LogLevel, cfg.AppEnv)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	log.Debug("Logger initialized")

	if cfg.SentryEnabled {
		if err := initSentry(cfg.SentryDSN); err != nil {
			log.Fatalf("Sentry initialization failed: %v", err)
		}
		log.Debug("Sentry initialized")
	} else {
		log.Debug("Sentry is disabled")
	}

	defer sentry.Flush(2 * time.Second)

	healthChecker := health.NewHealthChecker(cfg, log)
	h, err := healthChecker.RegisterHealthChecks()
	if err != nil {
		log.PanicF("Failed to register health checks: %v", err)
	}
	log.Debug("Health checks registered")

	ctx := context.Background()
	tracerManager := tracer.NewTracerManager(cfg, log)
	tp, err := tracerManager.InitTracer(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize tracing: %v", err)
	}
	defer func() { _ = tp.Shutdown(ctx) }()

	metricsCollector := metrics.NewMetricsCollector(log)

	alertHandler := alerts.NewAlertHandler(log, metricsCollector)
	routerManager := router.NewRouterManager(cfg, log, tracerManager, metricsCollector, h, alertHandler)
	r := routerManager.SetupRouter()
	log.Debug("Router setup complete")
	routerManager.StartServer(r)
}

func initSentry(sentryDSN string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
}
