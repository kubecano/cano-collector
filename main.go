package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/pkg/logger"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/metrics"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	r := setupRouter()

	StartServer(logg, r)
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

	r.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	}))
	r.Use(func(ctx *gin.Context) {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			hub.Scope().SetTag("endpoint", ctx.FullPath())
			hub.Scope().SetTag("version", config.GlobalConfig.AppVersion)
		}
		ctx.Next()
	})

	// Set up routes
	metrics.RegisterMetrics()
	r.Use(metrics.PrometheusMiddleware())

	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world!")
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return r
}

func StartServer(logg *zap.SugaredLogger, router *gin.Engine) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logg.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be caught, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logg.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logg.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	logg.Info("timeout of 5 seconds.")
	logg.Info("Server exiting")
}
