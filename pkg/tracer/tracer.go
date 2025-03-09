package tracer

import (
	"context"

	"github.com/gin-gonic/gin"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/kubecano/cano-collector/config"
)

func InitTracer(ctx context.Context) (*sdktrace.TracerProvider, error) {
	var tp *sdktrace.TracerProvider

	if config.GlobalConfig.TracingEnabled {
		endpoint := config.GlobalConfig.TracingEndpoint
		exporter, err := otlptracegrpc.New(
			ctx,
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpointURL(endpoint),
		)
		if err != nil {
			return nil, err
		}

		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(config.GlobalConfig.AppName),
			)),
		)
	} else {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithResource(sdkresource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(config.GlobalConfig.AppName),
			)),
		)
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}

func TraceLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context())
		spanCtx := span.SpanContext()

		if spanCtx.HasTraceID() {
			c.Set("trace_id", spanCtx.TraceID().String())
			c.Set("span_id", spanCtx.SpanID().String())
		}

		c.Next()
	}
}
