package tracer

import (
	"context"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/semconv/v1.17.0"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func InitTracer(ctx context.Context) (*trace.TracerProvider, error) {
	// OTLP Exporter (wysyłanie trace’ów do backendu)
	exporter, err := otlptrace.New(ctx)
	if err != nil {
		return nil, err
	}

	// Konfiguracja Tracer Providera
	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String("my-gin-app"),
		)),
	)

	// Ustawienie globalnego Tracer Providera
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	return tp, nil
}
