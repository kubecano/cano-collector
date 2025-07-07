package interfaces

import (
	"context"

	"go.uber.org/zap"
)

//go:generate mockgen -destination=../../../mocks/logger_mock.go -package=mocks github.com/kubecano/cano-collector/pkg/logger/interfaces LoggerInterface
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
