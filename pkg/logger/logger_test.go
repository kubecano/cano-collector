package logger

import (
	"bytes"
	"testing"

	"go.uber.org/zap/zapcore"

	"go.uber.org/zap"
)

func TestInitLogger(t *testing.T) {
	InitLogger("debug")

	if logger == nil {
		t.Fatal("Logger should not be nil after initialization")
	}

	logLevel := logger.Core().Enabled(zap.DebugLevel)
	if !logLevel {
		t.Error("Expected logger to be initialized with debug level")
	}
}

func TestLogging(t *testing.T) {
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

	logger := zap.New(core).Sugar()

	testMessage := "Test log entry"
	logger.Info(testMessage)

	logOutput := buf.String()

	if !bytes.Contains([]byte(logOutput), []byte(testMessage)) {
		t.Errorf("Expected log output to contain message: %s, but got: %s", testMessage, logOutput)
	}
}

func TestGetLogger(t *testing.T) {
	InitLogger("info")
	logger := GetLogger()

	if logger == nil {
		t.Fatal("GetLogger() returned nil, expected an initialized logger")
	}
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
		{"invalid", zap.InfoLevel}, // Domyślny poziom
	}

	for _, tt := range tests {
		InitLogger(tt.level)
		logger := GetLogger()

		if !logger.Core().Enabled(tt.expectedMin) {
			t.Errorf("Logger with level %s should allow %v level", tt.level, tt.expectedMin)
		}
	}
}
