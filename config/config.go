package config

import (
	"log"
	"os"
)

type Config struct {
	AppName       string
	LogLevel      string
	SentryDSN     string
	SentryEnabled bool
}

var GlobalConfig Config

func LoadConfig() {
	sentryDSN := getEnv("SENTRY_DSN", "")

	GlobalConfig = Config{
		AppName:       getEnv("APP_NAME", "cano-collector"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
		SentryDSN:     sentryDSN,
		SentryEnabled: sentryDSN != "",
	}

	log.Printf("Configuration loaded: %+v", GlobalConfig)
}

func getEnv(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
