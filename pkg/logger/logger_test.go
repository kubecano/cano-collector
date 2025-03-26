package logger

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	log := NewLogger("debug", "development")
	assert.NotNil(t, log, "Logger should not be nil")

	logLevel := log.zapLogger.Core().Enabled(zap.DebugLevel)
	assert.True(t, logLevel, "Expected logger to be initialized with debug level")
}

func TestLoggingToBuffer(t *testing.T) {
	var buf bytes.Buffer
	writer := zapcore.AddSync(&buf)

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeLevel:  zapcore.LowercaseLevelEncoder,
		EncodeTime:   zapcore.ISO8601TimeEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		writer,
		zap.DebugLevel,
	)

	log := &Logger{zapLogger: zap.New(core)}

	testMessage := "Test log entry"
	log.Info(testMessage)

	logOutput := buf.String()
	assert.Contains(t, logOutput, testMessage, "Expected log output to contain message")
}

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level       string
		expectedMin zapcore.Level
	}{
		{"debug", zap.DebugLevel},
		{"info", zap.InfoLevel},
		{"warn", zap.WarnLevel},
		{"error", zap.ErrorLevel},
		{"invalid", zap.InfoLevel},
	}

	for _, tt := range tests {
		log := NewLogger(tt.level, "production")
		assert.NotNil(t, log, "Logger should not be nil")

		assert.True(t, log.zapLogger.Core().Enabled(tt.expectedMin),
			"Logger with level %s should allow %v level", tt.level, tt.expectedMin)
	}
}

func TestWithContextLogger(t *testing.T) {
	log := NewLogger("debug", "development")

	tracer := noop.NewTracerProvider().Tracer("test-tracer")
	ctx, span := tracer.Start(context.Background(), "test-span")
	defer span.End()

	logWithCtx := log.WithContextLogger(ctx)

	spanCtx := span.SpanContext()
	traceID := spanCtx.TraceID().String()
	spanID := spanCtx.SpanID().String()

	assert.True(t, logWithCtx.Core().Enabled(zap.DebugLevel), "Logger should support debug level")
	assert.NotEmpty(t, traceID, "Trace ID should not be empty")
	assert.NotEmpty(t, spanID, "Span ID should not be empty")
}

func TestNewLogger_InvalidConfig(t *testing.T) {
	assert.NotPanics(t, func() { NewLogger("", "development") }, "Not expected an error when log level is empty")

	assert.NotPanics(t, func() { NewLogger("invalid", "production") }, "Not expected an error when log level is invalid")
}

func TestNewLogger_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("LoggerFactory should not panic, got: %v", r)
		}
	}()

	assert.NotPanics(t, func() { NewLogger("invalid", "production") }, "Expected error but not a panic")
}
