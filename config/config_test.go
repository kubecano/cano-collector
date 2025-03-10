package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestGetEnvString(t *testing.T) {
	_ = os.Setenv("TEST_STRING", "value1")
	defer func() {
		_ = os.Unsetenv("TEST_STRING")
	}()

	assert.Equal(t, "value1", getEnvString("TEST_STRING", "default"))
	assert.Equal(t, "default", getEnvString("NON_EXISTENT_STRING", "default"))
}

func TestGetEnvBool(t *testing.T) {
	_ = os.Setenv("TEST_BOOL_TRUE", "true")
	_ = os.Setenv("TEST_BOOL_FALSE", "false")
	_ = os.Setenv("TEST_BOOL_INVALID", "invalid")
	defer func() {
		_ = os.Unsetenv("TEST_BOOL_TRUE")
	}()
	defer func() {
		_ = os.Unsetenv("TEST_BOOL_FALSE")
	}()
	defer func() {
		_ = os.Unsetenv("TEST_BOOL_INVALID")
	}()

	assert.True(t, getEnvBool("TEST_BOOL_TRUE", false))
	assert.False(t, getEnvBool("TEST_BOOL_FALSE", true))

	assert.True(t, getEnvBool("NON_EXISTENT_BOOL", true))
	assert.False(t, getEnvBool("NON_EXISTENT_BOOL", false))

	assert.False(t, getEnvBool("TEST_BOOL_INVALID", false))
}

func TestGetEnvEnum(t *testing.T) {
	allowedValues := []string{"disabled", "local", "remote"}

	_ = os.Setenv("TEST_ENUM_VALID", "local")
	_ = os.Setenv("TEST_ENUM_INVALID", "invalid")
	defer func() {
		_ = os.Unsetenv("TEST_ENUM_VALID")
	}()
	defer func() {
		_ = os.Unsetenv("TEST_ENUM_INVALID")
	}()

	assert.Equal(t, "local", getEnvEnum("TEST_ENUM_VALID", allowedValues, "disabled"))
	assert.Equal(t, "disabled", getEnvEnum("TEST_ENUM_INVALID", allowedValues, "disabled"))
	assert.Equal(t, "disabled", getEnvEnum("NON_EXISTENT_ENUM", allowedValues, "disabled"))
}
