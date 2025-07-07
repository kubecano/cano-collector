package interfaces

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

//go:generate mockgen -destination=../../../mocks/logger_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/logger/interfaces LoggerInterface
type LoggerInterface interface {
	// Structured logging methods - preferred for application code
	Debug(msg string, fields ...zapcore.Field)
	Info(msg string, fields ...zapcore.Field)
	Warn(msg string, fields ...zapcore.Field)
	Error(msg string, fields ...zapcore.Field)
	Fatal(msg string, fields ...zapcore.Field)
	Panic(msg string, fields ...zapcore.Field)

	// Sugar API methods - for simple logging
	Debugf(template string, args ...interface{})
	Infof(template string, args ...interface{})
	Warnf(template string, args ...interface{})
	Errorf(template string, args ...interface{})
	Fatalf(template string, args ...interface{})
	Panicf(template string, args ...interface{})

	// Zap-specific methods
	WithContextLogger(ctx context.Context) *zap.Logger
	GetLogger() *zap.Logger
}
