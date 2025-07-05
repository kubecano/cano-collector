package router

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/kubecano/cano-collector/pkg/alert"

	"github.com/kubecano/cano-collector/pkg/tracer"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"

	"github.com/kubecano/cano-collector/pkg/logger"

	sentrygin "github.com/getsentry/sentry-go/gin"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/health"
	"github.com/kubecano/cano-collector/pkg/interfaces"
)

//go:generate mockgen -destination=../../mocks/router_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/router RouterInterface
type RouterInterface interface {
	SetupRouter() *gin.Engine
	StartServer(router *gin.Engine)
}

type RouterManager struct {
	cfg     config.Config
	logger  logger.LoggerInterface
	tracer  tracer.TracerInterface
	metrics interfaces.MetricsInterface
	health  health.HealthInterface
	alerts  alert.AlertHandlerInterface
}

func NewRouterManager(
	cfg config.Config,
	log logger.LoggerInterface,
	tracer tracer.TracerInterface,
	metrics interfaces.MetricsInterface,
	health health.HealthInterface,
	alerts alert.AlertHandlerInterface,
) *RouterManager {
	return &RouterManager{
		cfg:     cfg,
		logger:  log,
		tracer:  tracer,
		metrics: metrics,
		health:  health,
		alerts:  alerts,
	}
}

func (rm *RouterManager) SetupRouter() *gin.Engine {
	r := gin.New()

	rm.logger.Debug("Setting up router")

	r.Use(sentrygin.New(sentrygin.Options{
		Repanic:         true,
		WaitForDelivery: false,
		Timeout:         2 * time.Second,
	}))

	r.Use(func(ctx *gin.Context) {
		if hub := sentrygin.GetHubFromContext(ctx); hub != nil {
			hub.Scope().SetTag("endpoint", ctx.FullPath())
			hub.Scope().SetTag("version", rm.cfg.AppVersion)
		}
		ctx.Next()
	})

	if rm.cfg.TracingMode != "disabled" {
		r.Use(otelgin.Middleware(rm.cfg.AppName))
		r.Use(rm.tracer.TraceLoggerMiddleware())
	} else {
		rm.logger.Debug("otelgin middleware is disabled to prevent trace generation.")
	}

	r.Use(ginzap.GinzapWithConfig(rm.logger.GetLogger(), &ginzap.Config{
		TimeFormat: time.RFC3339,
		UTC:        true,
		Context: func(c *gin.Context) []zapcore.Field {
			var fields []zapcore.Field

			if traceID, exists := c.Get("trace_id"); exists {
				fields = append(fields, zap.String("trace_id", traceID.(string)))
			}
			if spanID, exists := c.Get("span_id"); exists {
				fields = append(fields, zap.String("span_id", spanID.(string)))
			}

			return fields
		},
	}))

	// Logs all panic to error log
	//   - stack means whether output the stack info.
	r.Use(ginzap.RecoveryWithZap(rm.logger.GetLogger(), true))

	r.Use(rm.metrics.PrometheusMiddleware())

	r.GET("/", rm.rootHandler)
	r.GET("/metrics", func(c *gin.Context) {
		metricsFamilies, err := prometheus.DefaultGatherer.Gather()
		if err != nil || len(metricsFamilies) == 0 {
			rm.logger.Error("Prometheus collector not initialized or empty registry")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Prometheus collector not initialized"})
			return
		}
		promhttp.Handler().ServeHTTP(c.Writer, c.Request)
	})

	r.GET("/livez", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/readyz", gin.WrapH(rm.health.Handler()))
	r.GET("/healthz", gin.WrapH(rm.health.Handler()))

	api := r.Group("/api")
	{
		api.POST("/alerts", rm.alerts.HandleAlert)
	}

	rm.logger.Debug("Router setup complete")
	return r
}

func (rm *RouterManager) rootHandler(c *gin.Context) {
	tr := otel.Tracer(rm.cfg.AppName)
	_, span := tr.Start(c.Request.Context(), "root-handler")
	defer span.End()
	c.String(http.StatusOK, "Hello world!")
}

func (rm *RouterManager) StartServer(router *gin.Engine) {
	rm.logger.Info("Cano-collector server starting...")
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		// service connections
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			rm.logger.Fatalf("Failed to start server: %v", err)
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
	rm.logger.Info("Cano-collector shutdown server ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		rm.logger.Fatalf("Cano-collector server shutdown: %v", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	<-ctx.Done()
	rm.logger.Info("Cano-collector server exiting")
}
