package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/kubecano/cano-collector/config"
)

var logger *zap.Logger

func InitLogger(level string) {
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
		Development:      config.GlobalConfig.AppEnv != "production",
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

	var err error
	logger, err = zapConfig.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

func GetLogger() *zap.Logger {
	return logger
}

// Debug logs a message at DebugLevel. The message includes any fields passed at the log site.
func Debug(args ...interface{}) {
	logger.Sugar().Debug(args...)
}

// Debugf logs a formatted message at DebugLevel. The message includes any fields passed at the log site.
func Debugf(template string, args ...interface{}) {
	logger.Sugar().Debugf(template, args...)
}

// Info logs a message at InfoLevel. The message includes any fields passed at the log site.
func Info(args ...interface{}) {
	logger.Sugar().Info(args...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed at the log site.
func Warn(args ...interface{}) {
	logger.Sugar().Warn(args...)
}

// Warnf logs a formatted message at WarnLevel. The message includes any fields passed at the log site.
func Warnf(template string, args ...interface{}) {
	logger.Sugar().Warnf(template, args...)
}

// Errorf logs a message at ErrorLevel. The message includes any fields passed at the log site.
func Errorf(template string, args ...interface{}) {
	logger.Sugar().Errorf(template, args...)
}

// Fatalf logs a message at FatalLevel and calls os.Exit. The message includes any fields passed at the log site.
func Fatalf(template string, args ...interface{}) {
	logger.Sugar().Fatalf(template, args...)
}

// Fatal logs a message at FatalLevel and calls os.Exit. The message includes any fields passed at the log site.
func Fatal(args ...interface{}) {
	logger.Sugar().Fatal(args...)
}

// PanicF logs a message at PanicLevel and panics. The message includes any fields passed at the log site.
func PanicF(template string, args ...interface{}) {
	logger.Sugar().Panicf(template, args...)
}
