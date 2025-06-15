package config

import (
	"os"
	"strconv"

	"github.com/kubecano/cano-collector/config/destinations"
	"github.com/kubecano/cano-collector/config/teams"
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
	Destinations    destinations.DestinationsConfig
	Teams           teams.TeamsConfig
}

//go:generate mockgen -destination=../mocks/fullconfig_loader_mock.go -package=mocks github.com/kubecano/cano-collector/config FullConfigLoader
type FullConfigLoader interface {
	Load() (destinations.DestinationsConfig, teams.TeamsConfig, error)
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

func (f *fileConfigLoader) Load() (destinations.DestinationsConfig, teams.TeamsConfig, error) {
	destLoader := destinations.NewFileDestinationsLoader(f.destinationsPath)
	teamLoader := teams.NewFileTeamsLoader(f.teamsPath)

	d, err := destLoader.Load()
	if err != nil {
		return destinations.DestinationsConfig{}, teams.TeamsConfig{}, err
	}

	t, err := teamLoader.Load()
	if err != nil {
		return destinations.DestinationsConfig{}, teams.TeamsConfig{}, err
	}

	return *d, *t, nil
}

func LoadConfig() (Config, error) {
	loader := NewFileConfigLoader(
		"/etc/cano-collector/destinations/destinations.yaml",
		"/etc/cano-collector/teams/teams.yaml",
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
