package router

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

	"github.com/hellofresh/health-go/v5"

	sentrygin "github.com/getsentry/sentry-go/gin"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/metrics"
)

func SetupRouter(logger *zap.Logger, health *health.Health) *gin.Engine {
	r := gin.Default()

	logger.Debug("Setting up router")

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

	// Add a ginzap middleware, which:
	//   - Logs all requests, like a combined access and error log.
	//   - Logs to stdout.
	//   - RFC3339 with UTC time format.
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	r.Use(ginzap.RecoveryWithZap(logger, true))

	// Set up routes
	metrics.RegisterMetrics()
	r.Use(metrics.PrometheusMiddleware())

	r.GET("/", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "Hello world!")
	})

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.GET("/livez", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", gin.WrapH(health.Handler()))
	r.GET("/healthz", gin.WrapH(health.Handler()))

	logger.Debug("Router setup complete")
	return r
}

func StartServer(router *gin.Engine) {
	logger.Info("Cano-collector server starting...")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("listen: %s\n", err)
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
	logger.Info("Cano-collector shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Cano-collector server shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	logger.Info("Cano-collector server exiting")
}
