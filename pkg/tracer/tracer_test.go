package tracer

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/kubecano/cano-collector/mocks"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/trace"

	"github.com/kubecano/cano-collector/config"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func setupTestTracerProvider(t *testing.T, cfg config.Config) *TracerManager {
	t.Helper()
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks.NewMockLoggerInterface(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Debug(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Info(gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Infof(gomock.Any(), gomock.Any()).AnyTimes()

	originalProvider := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider())

	defer otel.SetTracerProvider(originalProvider)

	return NewTracerManager(cfg, mockLogger)
}

func TestInitTracer_Disabled(t *testing.T) {
	cfg := config.Config{TracingMode: "disabled"}
	tm := setupTestTracerProvider(t, cfg)

	err := tm.InitTracer(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, tm.provider, "TraceProvider should not be nil")

	assert.Equal(t, tm.provider, otel.GetTracerProvider(), "TracerProvider should be global")

	tracer := otel.Tracer("test-tracer")
	assert.NotNil(t, tracer, "Tracer should not be nil")

	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	assert.False(t, span.IsRecording(), "Span should not be recording for mode=disabled")
	assert.Equal(t, trace.TraceFlags(0x0), span.SpanContext().TraceFlags(), "TraceFlags should be 0x0 for mode=disabled")
}

func TestInitTracer_Local(t *testing.T) {
	cfg := config.Config{TracingMode: "local", AppName: "cano-collector"}

	tm := setupTestTracerProvider(t, cfg)
	err := tm.InitTracer(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, tm.provider)

	assert.Equal(t, tm.provider, otel.GetTracerProvider())

	tracer := otel.Tracer("test-tracer")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	assert.True(t, span.SpanContext().HasTraceID(), "TraceID should be generated for mode=local")
	assert.True(t, span.SpanContext().HasSpanID(), "SpanID should be generated for mode=local")
}

func TestInitTracer_Remote(t *testing.T) {
	cfg := config.Config{
		TracingMode:     "remote",
		TracingEndpoint: "localhost:4317",
		AppName:         "cano-collector",
	}

	tm := setupTestTracerProvider(t, cfg)
	err := tm.InitTracer(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, tm.provider)

	assert.Equal(t, tm.provider, otel.GetTracerProvider())

	tracer := otel.Tracer("test-tracer")
	_, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	assert.True(t, span.SpanContext().HasTraceID(), "TraceID should be generated for mode=remote")
	assert.True(t, span.SpanContext().HasSpanID(), "SpanID should be generated for mode=remote")
}

func TestInitTracer_InvalidEndpoint(t *testing.T) {
	cfg := config.Config{
		TracingMode:     "remote",
		TracingEndpoint: "invalid-endpoint",
	}

	tm := setupTestTracerProvider(t, cfg)
	err := tm.InitTracer(context.Background())

	require.Error(t, err, "Expected error for invalid tracing endpoint")
	assert.Nil(t, tm.provider, "TracerProvider should be nil on failure")
}

func TestTraceLoggerMiddleware(t *testing.T) {
	cfg := config.Config{TracingMode: "local"}

	tm := setupTestTracerProvider(t, cfg)

	router := gin.New()
	router.Use(tm.TraceLoggerMiddleware())
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
