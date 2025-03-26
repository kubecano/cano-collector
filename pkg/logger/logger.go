package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate mockgen -destination=../../mocks/logger_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/logger LoggerInterface
type LoggerInterface interface {
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Panic(args ...interface{})
	Panicf(template string, args ...interface{})
	WithContextLogger(ctx context.Context) *zap.Logger
	GetLogger() *zap.Logger
}

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

// Debug logs a message at DebugLevel. The message includes any fields passed at the log site.
func (l *Logger) Debug(args ...interface{}) {
	l.zapLogger.Sugar().Debug(args...)
}

// Debugf logs a formatted message at DebugLevel. The message includes any fields passed at the log site.
func (l *Logger) Debugf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Debugf(template, args...)
}

// Info logs a message at InfoLevel. The message includes any fields passed at the log site.
func (l *Logger) Info(args ...interface{}) {
	l.zapLogger.Sugar().Info(args...)
}

// Infof logs a formatted message at InfoLevel. The message includes any fields passed at the log site.
func (l *Logger) Infof(template string, args ...interface{}) {
	l.zapLogger.Sugar().Infof(template, args...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed at the log site.
func (l *Logger) Warn(args ...interface{}) {
	l.zapLogger.Sugar().Warn(args...)
}

// Warnf logs a formatted message at WarnLevel. The message includes any fields passed at the log site.
func (l *Logger) Warnf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Warnf(template, args...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed at the log site.
func (l *Logger) Error(args ...interface{}) {
	l.zapLogger.Sugar().Error(args...)
}

// Errorf logs a message at ErrorLevel. The message includes any fields passed at the log site.
func (l *Logger) Errorf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Errorf(template, args...)
}

// Fatal logs a message at FatalLevel and calls os.Exit. The message includes any fields passed at the log site.
func (l *Logger) Fatal(args ...interface{}) {
	l.zapLogger.Sugar().Fatal(args...)
}

// Fatalf logs a message at FatalLevel and calls os.Exit. The message includes any fields passed at the log site.
func (l *Logger) Fatalf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Fatalf(template, args...)
}

// Panic logs a message at PanicLevel and panics. The message includes any fields passed at the log site.
func (l *Logger) Panic(args ...interface{}) {
	l.zapLogger.Sugar().Panic(args...)
}

// Panicf logs a message at PanicLevel and panics. The message includes any fields passed at the log site.
func (l *Logger) Panicf(template string, args ...interface{}) {
	l.zapLogger.Sugar().Panicf(template, args...)
}
