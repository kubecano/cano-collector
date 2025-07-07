package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	zapLogger *zap.Logger
}

func NewLogger(level, env string) *Logger {
	var logLevel zapcore.Level

	switch level {
	case "debug":
		logLevel = zap.DebugLevel
	case "info":
		logLevel = zap.InfoLevel
	case "warn":
		logLevel = zap.WarnLevel
	case "error":
		logLevel = zap.ErrorLevel
	default:
		logLevel = zap.InfoLevel
	}

	zapConfig := zap.Config{
		Encoding:         "json",
		Development:      env != "production",
		Level:            zap.NewAtomicLevelAt(logLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:      "timestamp",
			LevelKey:     "level",
			MessageKey:   "message",
			CallerKey:    "caller",
			EncodeLevel:  zapcore.LowercaseLevelEncoder, // info, debug, error
			EncodeTime:   zapcore.ISO8601TimeEncoder,
			EncodeCaller: zapcore.ShortCallerEncoder,
		},
	}

	zapLogger, err := zapConfig.Build()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	return &Logger{zapLogger: zapLogger}
}

func (l *Logger) GetLogger() *zap.Logger {
	return l.zapLogger
}

func (l *Logger) WithContextLogger(ctx context.Context) *zap.Logger {
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()

	if spanCtx.HasTraceID() {
		return l.zapLogger.With(
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		)
	}
	return l.zapLogger
}

// Debug logs a message at DebugLevel with structured fields
func (l *Logger) Debug(msg string, fields ...zapcore.Field) {
	l.zapLogger.Debug(msg, fields...)
}

// Debugf logs a formatted message at DebugLevel. The message includes any fields passed at the log site.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Debugf(template, args...)
}

// Info logs a message at InfoLevel with structured fields
func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.zapLogger.Info(msg, fields...)
}

// Infof logs a formatted message at InfoLevel. The message includes any fields passed at the log site.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.zapLogger.Sugar().Infof(template, args...)
}

// Warn logs a message at WarnLevel with structured fields
func (l *Logger) Warn(msg string, fields ...zapcore.Field) {
	l.zapLogger.Warn(msg, fields...)
}

// Warnf logs a formatted message at WarnLevel. The message includes any fields passed at the log site.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Warnf(template, args...)
}

// Error logs a message at ErrorLevel with structured fields
func (l *Logger) Error(msg string, fields ...zapcore.Field) {
	l.zapLogger.Error(msg, fields...)
}

// Errorf logs a formatted message at ErrorLevel. The message includes any fields passed at the log site.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Errorf(template, args...)
}

// Fatal logs a message at FatalLevel with structured fields and calls os.Exit
func (l *Logger) Fatal(msg string, fields ...zapcore.Field) {
	l.zapLogger.Fatal(msg, fields...)
}

// Fatalf logs a message at FatalLevel and calls os.Exit. The message includes any fields passed at the log site.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Fatalf(template, args...)
}

// Panic logs a message at PanicLevel with structured fields and panics
func (l *Logger) Panic(msg string, fields ...zapcore.Field) {
	l.zapLogger.Panic(msg, fields...)
}

// Panicf logs a message at PanicLevel and panics. The message includes any fields passed at the log site.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Panicf(template, args...)
}
