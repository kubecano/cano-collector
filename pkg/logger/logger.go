package logger

import (
	"github.com/kubecano/cano-collector/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func Info(args ...interface{}) {
	logger.Sugar().Info(args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Sugar().Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	logger.Sugar().Fatalf(template, args...)
}

func Fatal(args ...interface{}) {
	logger.Sugar().Fatal(args...)
}

func PanicF(template string, args ...interface{}) {
	logger.Sugar().Panicf(template, args...)
}
