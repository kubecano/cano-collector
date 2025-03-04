package main

import (
	"log"
	"net/http"
	"time"

	"github.com/kubecano/cano-collector/config"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

func main() {
	config.LoadConfig()

	if config.GlobalConfig.SentryEnabled {
		if err := initSentry(config.GlobalConfig.SentryDSN); err != nil {
			log.Fatalf("Sentry initialization failed: %v", err)
		}
	}

	// Flush buffered events before the program terminates.
	// Set the timeout to the maximum duration the program can afford to wait.
	defer sentry.Flush(2 * time.Second)

	r := setupRouter()

	r.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	}))

	if err := StartServer(r); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}

func initSentry(sentryDSN string) error {
	return sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDSN,
		EnableTracing:    true,
		TracesSampleRate: 1.0,
	})
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	// Set up routes
	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world!")
	})
	r.GET("/ping", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "pong")
	})
	return r
}

func StartServer(r *gin.Engine) error {
	var err error

	if gin.Mode() != gin.TestMode {
		err = r.Run(":3000")
	}

	return err
}
