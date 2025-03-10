package config

import (
	"os"
	"strconv"
)

type Config struct {
	AppName         string
	AppVersion      string
	AppEnv          string
	LogLevel        string
	TracingMode     string
	TracingEndpoint string
	SentryDSN       string
	SentryEnabled   bool
}

var GlobalConfig Config

func LoadConfig() {
	GlobalConfig = Config{
		AppName:         getEnvString("APP_NAME", "cano-collector"),
		AppVersion:      getEnvString("APP_VERSION", "dev"),
		AppEnv:          getEnvString("APP_ENV", "production"),
		LogLevel:        getEnvEnum("LOG_LEVEL", []string{"debug", "info", "warn", "error"}, "info"),
		TracingMode:     getEnvEnum("TRACING_MODE", []string{"disabled", "local", "remote"}, "disabled"),
		TracingEndpoint: getEnvString("TRACING_ENDPOINT", "http://localhost:4317"),
		SentryDSN:       getEnvString("SENTRY_DSN", ""),
		SentryEnabled:   getEnvBool("ENABLE_TELEMETRY", true),
	}
}

func getEnvString(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	parsedValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return parsedValue
}

func getEnvEnum(key string, allowedValues []string, defaultValue string) string {
	value := getEnvString(key, defaultValue)
	for _, allowed := range allowedValues {
		if value == allowed {
			return value
		}
	}
	return defaultValue
}
