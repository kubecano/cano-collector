package config_destination

import (
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type DestinationsConfig struct {
	Destinations struct {
		Slack []DestinationSlack `yaml:"slack"`
	} `yaml:"destinations"`
}

// SlackDestination represents a Slack notification destination
type DestinationSlack struct {
	Name             string `yaml:"name"`
	APIKey           string `yaml:"api_key"`
	SlackChannel     string `yaml:"slack_channel"`
	GroupingInterval int    `yaml:"grouping_interval,omitempty"`
	UnfurlLinks      *bool  `yaml:"unfurl_links,omitempty"`
}

//go:generate mockgen -destination=../../mocks/destinations_loader_mock.go -package=mocks github.com/kubecano/cano-collector/config/destination DestinationsLoader
type DestinationsLoader interface {
	Load() (*DestinationsConfig, error)
}

// FileDestinationsLoader loads destinations from a file or secret (ConfigMap/Secret mount)
type FileDestinationsLoader struct {
	Path string
}

func NewFileDestinationsLoader(path string) *FileDestinationsLoader {
	return &FileDestinationsLoader{Path: path}
}

func (f *FileDestinationsLoader) Load() (*DestinationsConfig, error) {
	file, err := os.Open(f.Path)
	if err != nil {
		return nil, fmt.Errorf("cannot open destination config: %w", err)
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	return parseDestinationsYAML(file)
}

func parseDestinationsYAML(r io.Reader) (*DestinationsConfig, error) {
	var config DestinationsConfig
	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&config); err != nil {
		return nil, fmt.Errorf("failed to decode destinations YAML: %w", err)
	}

	// Validate Slack destinations
	for _, d := range config.Destinations.Slack {
		if err := validateSlackDestination(d); err != nil {
			return nil, fmt.Errorf("invalid Slack destination '%s': %w", d.Name, err)
		}
	}

	// Process environment variable placeholders for API keys
	for i, d := range config.Destinations.Slack {
		if strings.HasPrefix(d.APIKey, "${") && strings.HasSuffix(d.APIKey, "}") {
			envVar := strings.TrimSuffix(strings.TrimPrefix(d.APIKey, "${"), "}")
			val, ok := os.LookupEnv(envVar)
			if !ok {
				return nil, fmt.Errorf("missing required env %s for slack destination %s", envVar, d.Name)
			}
			config.Destinations.Slack[i].APIKey = val
		}
	}

	return &config, nil
}

func validateSlackDestination(d DestinationSlack) error {
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}

	if d.SlackChannel == "" {
		return fmt.Errorf("slack_channel is required")
	}

	if d.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	// Allow placeholder format ${ENV_VAR}
	if strings.HasPrefix(d.APIKey, "${") && strings.HasSuffix(d.APIKey, "}") {
		return nil
	}

	// Validate grouping_interval if provided
	if d.GroupingInterval < 0 {
		return fmt.Errorf("grouping_interval must be non-negative")
	}

	return nil
}
