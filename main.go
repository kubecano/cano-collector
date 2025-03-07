package main

import (
	"time"

	"github.com/kubecano/cano-collector/pkg/router"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/kubecano/cano-collector/pkg/health"

	"github.com/getsentry/sentry-go"

	"github.com/kubecano/cano-collector/config"
)

func main() {
	config.LoadConfig()

	logger.InitLogger(config.GlobalConfig.LogLevel)
	logger.Debug("Logger initialized")

	if config.GlobalConfig.SentryEnabled {
		if err := initSentry(config.GlobalConfig.SentryDSN); err != nil {
			logger.Fatalf("Sentry initialization failed: %v", err)
		}
		logger.Debug("Sentry initialized")
	} else {
		logger.Debug("Sentry is disabled")
	}

	defer sentry.Flush(2 * time.Second)

	h, err := health.RegisterHealthChecks()
	if err != nil {
		logger.PanicF("Failed to register health checks: %v", err)
	}
	logger.Debug("Health checks registered")

	r := router.SetupRouter(h)
	logger.Debug("Router setup complete")

	router.StartServer(r)
}

func initSentry(sentryDSN string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
}
