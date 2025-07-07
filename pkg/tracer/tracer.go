package tracer

import (
	"context"
	"fmt"
	"net/url"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/kubecano/cano-collector/config"
	logger_interfaces "github.com/kubecano/cano-collector/pkg/logger/interfaces"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type TracerManager struct {
	cfg      config.Config
	logger   logger_interfaces.LoggerInterface
	provider *sdktrace.TracerProvider
}

func NewTracerManager(cfg config.Config, logger logger_interfaces.LoggerInterface) *TracerManager {
	return &TracerManager{cfg: cfg, logger: logger}
}

func (tm *TracerManager) InitTracer(ctx context.Context) error {
	tm.logger.Debug("Initializing Tracer...")

	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample())))

	switch tm.cfg.TracingMode {
	case "disabled":
		tm.logger.Info("Tracing is disabled. No traces will be collected.")
		tm.provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)

	case "local":
		tm.logger.Info("Tracing is enabled in local mode. Traces will not be exported.")
		tm.provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(tm.cfg.AppName),
			)),
		)

	case "remote":
		endpoint := tm.cfg.TracingEndpoint
		tm.logger.Infof("Tracing is enabled in remote mode. Exporting traces to: %s", endpoint)

		if _, err := url.ParseRequestURI(endpoint); err != nil {
			tm.logger.Errorf("Invalid tracing endpoint: %s", err)
			return fmt.Errorf("invalid tracing endpoint: %w", err)
		}

		exporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpointURL(endpoint),
		)
		if err != nil {
			tm.logger.Errorf("Failed to initialize OTLP exporter: %v", err)
			return err
		}

		tm.provider = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(tm.cfg.AppName),
			)),
		)
	}

	if tm.provider == nil {
		tm.logger.Warn("TracerProvider was nil. Creating a new default TracerProvider.")
		tm.provider = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)
	}

	otel.SetTracerProvider(tm.provider)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tm.logger.Debug("Tracer initialized successfully.")
	return nil
}

func (tm *TracerManager) TraceLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		spanCtx := span.SpanContext()

		if spanCtx.HasTraceID() {
			traceID := spanCtx.TraceID().String()
			spanID := spanCtx.SpanID().String()

			c.Set("trace_id", traceID)
			c.Set("span_id", spanID)

			tm.logger.Debug("Request Trace Info:",
				zap.String("trace_id", traceID),
				zap.String("span_id", spanID),
			)
		} else {
			tm.logger.Debug("No active trace found for request.")
		}

		c.Next()
	}
}

func (tm *TracerManager) ShutdownTracer(ctx context.Context) error {
	if tm.provider != nil {
		return tm.provider.Shutdown(ctx)
	}
	return nil
}
