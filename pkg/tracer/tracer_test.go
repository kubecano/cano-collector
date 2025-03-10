package tracer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/kubecano/cano-collector/config"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestInitTracer_Disabled(t *testing.T) {
	config.GlobalConfig.TracingMode = "disabled"

	tp, err := InitTracer(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, tp, "TraceProvider should not be nil")

	assert.Equal(t, tp, otel.GetTracerProvider())

	tracer := otel.Tracer("test-tracer")
	assert.NotNil(t, tracer, "Tracer should not be nil")

	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	assert.Equal(t, trace.TraceFlags(0x0), span.SpanContext().TraceFlags(), "TraceFlags should be 0x0 for mode=disabled")
}

func TestInitTracer_Local(t *testing.T) {
	config.GlobalConfig.TracingMode = "local"

	tp, err := InitTracer(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, tp)

	assert.Equal(t, tp, otel.GetTracerProvider())

	tracer := otel.Tracer("test-tracer")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	assert.True(t, span.SpanContext().HasTraceID(), "TraceID should be generated for mode=local")
	assert.True(t, span.SpanContext().HasSpanID(), "SpanID should be generated for mode=local")
}

func TestInitTracer_Remote(t *testing.T) {
	config.GlobalConfig.TracingMode = "remote"
	config.GlobalConfig.TracingEndpoint = "localhost:4317"

	tp, err := InitTracer(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, tp)

	assert.Equal(t, tp, otel.GetTracerProvider())

	tracer := otel.Tracer("test-tracer")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	assert.True(t, span.SpanContext().HasTraceID(), "TraceID should be generated for mode=remote")
	assert.True(t, span.SpanContext().HasSpanID(), "SpanID should be generated for mode=remote")
}

func TestTraceLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(TraceLoggerMiddleware())
	router.GET("/test", func(c *gin.Context) {
		traceID, _ := c.Get("trace_id")
		spanID, _ := c.Get("span_id")
		c.JSON(http.StatusOK, gin.H{"trace_id": traceID, "span_id": spanID})
	})

	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "trace_id")
	assert.Contains(t, body, "span_id")
}
