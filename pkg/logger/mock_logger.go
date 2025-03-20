package logger

import (
	"context"

	"go.uber.org/zap"
)

type MockLogger struct {
	zapLogger *zap.Logger
}

func (m *MockLogger) Debug(...interface{})                          {}
func (m *MockLogger) Debugf(string, ...interface{})                 {}
func (m *MockLogger) Info(...interface{})                           {}
func (m *MockLogger) Infof(string, ...interface{})                  {}
func (m *MockLogger) Warn(...interface{})                           {}
func (m *MockLogger) Warnf(string, ...interface{})                  {}
func (m *MockLogger) Error(...interface{})                          {}
func (m *MockLogger) Errorf(string, ...interface{})                 {}
func (m *MockLogger) Fatal(...interface{})                          {}
func (m *MockLogger) Fatalf(string, ...interface{})                 {}
func (m *MockLogger) Panic(...interface{})                          {}
func (m *MockLogger) Panicf(string, ...interface{})                 {}
func (m *MockLogger) WithContextLogger(context.Context) *zap.Logger { return m.zapLogger }
func (m *MockLogger) GetLogger() *zap.Logger                        { return m.zapLogger }

func NewMockLogger() LoggerInterface {
	logger, _ := zap.NewDevelopment()
	return &MockLogger{zapLogger: logger}
}
