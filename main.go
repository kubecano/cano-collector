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
	logg := logger.GetLogger()

	if config.GlobalConfig.SentryEnabled {
		if err := initSentry(config.GlobalConfig.SentryDSN); err != nil {
			logg.Fatalf("Sentry initialization failed: %v", err)
		}
	}

	defer sentry.Flush(2 * time.Second)

	h := health.RegisterHealthChecks()

	r := router.SetupRouter(h)

	router.StartServer(logg, r)
}

func initSentry(sentryDSN string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
}
