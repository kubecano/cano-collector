package config

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	_ = os.Setenv("APP_NAME", "test-app")
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("SENTRY_DSN", "https://example@sentry.io/123")
	_ = os.Setenv("ENABLE_TELEMETRY", "true")

	LoadConfig()

	if GlobalConfig.AppName != "test-app" {
		t.Errorf("Expected APP_NAME to be 'test-app', got '%s'", GlobalConfig.AppName)
	}
	if GlobalConfig.LogLevel != "debug" {
		t.Errorf("Expected LOG_LEVEL to be 'debug', got '%s'", GlobalConfig.LogLevel)
	}
	if GlobalConfig.SentryDSN != "https://example@sentry.io/123" {
		t.Errorf("Expected SENTRY_DSN to be 'https://example@sentry.io/123', got '%s'", GlobalConfig.SentryDSN)
	}
	if !GlobalConfig.SentryEnabled {
		t.Errorf("Expected SENTRY_ENABLED to be true, got false")
	}
}
