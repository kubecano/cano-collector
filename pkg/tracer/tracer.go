package tracer

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/kubecano/cano-collector/config"
	"github.com/kubecano/cano-collector/pkg/logger"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type TracerManager struct {
	cfg    config.Config
	logger *logger.Logger
}

func NewTracerManager(cfg config.Config, logger *logger.Logger) *TracerManager {
	return &TracerManager{cfg: cfg, logger: logger}
}

func (tm *TracerManager) InitTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	tm.logger.Debug("Initializing Tracer...")

	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.NeverSample())))

	var tp *sdktrace.TracerProvider

	switch tm.cfg.TracingMode {
	case "disabled":
		tm.logger.Info("Tracing is disabled. No traces will be collected.")
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)

	case "local":
		tm.logger.Info("Tracing is enabled in local mode. Traces will not be exported.")
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(tm.cfg.AppName),
			)),
		)

	case "remote":
		endpoint := tm.cfg.TracingEndpoint
		tm.logger.Infof("Tracing is enabled in remote mode. Exporting traces to: %s", endpoint)

		exporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpointURL(endpoint),
		)
		if err != nil {
			tm.logger.Errorf("Failed to initialize OTLP exporter: %v", err)
			return nil, err
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
			sdktrace.WithResource(sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(tm.cfg.AppName),
			)),
		)
	}

	if tp == nil {
		tm.logger.Warn("TracerProvider was nil. Creating a new default TracerProvider.")
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithSampler(sdktrace.NeverSample()),
		)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	tm.logger.Debug("Tracer initialized successfully.")
	return tp, nil
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

			tm.logger.Debug("Request Trace Info:", "trace_id:", traceID, "span_id:", spanID)
		} else {
			tm.logger.Debug("No active trace found for request.")
		}

		c.Next()
	}
}
