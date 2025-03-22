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
	Destinations    DestinationsConfig
	Teams           TeamsConfig
}

type FullConfigLoader interface {
	Load() (DestinationsConfig, TeamsConfig, error)
}

// LoadConfigWithLoader reads the Config from the provided loader
func LoadConfigWithLoader(loader FullConfigLoader) (Config, error) {
	destinations, teams, err := loader.Load()
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppName:         getEnvString("APP_NAME", "cano-collector"),
		AppVersion:      getEnvString("APP_VERSION", "dev"),
		AppEnv:          getEnvString("APP_ENV", "production"),
		LogLevel:        getEnvEnum("LOG_LEVEL", []string{"debug", "info", "warn", "error"}, "info"),
		TracingMode:     getEnvEnum("TRACING_MODE", []string{"disabled", "local", "remote"}, "disabled"),
		TracingEndpoint: getEnvString("TRACING_ENDPOINT", "http://localhost:4317"),
		SentryDSN:       getEnvString("SENTRY_DSN", ""),
		SentryEnabled:   getEnvBool("ENABLE_TELEMETRY", true),
		Destinations:    destinations,
		Teams:           teams,
	}, nil
}

type fileConfigLoader struct {
	destinationsPath string
	teamsPath        string
}

func NewFileConfigLoader(destinationsPath, teamsPath string) FullConfigLoader {
	return &fileConfigLoader{destinationsPath: destinationsPath, teamsPath: teamsPath}
}

func (f *fileConfigLoader) Load() (DestinationsConfig, TeamsConfig, error) {
	destLoader := NewFileDestinationsLoader(f.destinationsPath)
	teamLoader := NewFileTeamsLoader(f.teamsPath)

	dest, err := destLoader.Load()
	if err != nil {
		return DestinationsConfig{}, TeamsConfig{}, err
	}

	teams, err := teamLoader.Load()
	if err != nil {
		return DestinationsConfig{}, TeamsConfig{}, err
	}

	return *dest, *teams, nil
}

func LoadConfig() (Config, error) {
	loader := NewFileConfigLoader(
		"/etc/cano-collector/destinations.yaml",
		"/etc/cano-collector/teams.yaml",
	)
	return LoadConfigWithLoader(loader)
}

// Helpers
func getEnvString(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	if parsed, err := strconv.ParseBool(value); err == nil {
		return parsed
	}
	return defaultValue
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
