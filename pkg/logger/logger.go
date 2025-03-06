package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

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

	config := zap.Config{
		Encoding:         "json",
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

	l, err := config.Build()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
	Logger = l.Sugar()
}

func GetLogger() *zap.SugaredLogger {
	return Logger
}
