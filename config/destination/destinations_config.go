package destination

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

type DestinationsConfig struct {
	Destinations struct {
		Slack []SlackDestination `yaml:"slack"`
		Teams []TeamsDestination `yaml:"teams"`
	} `yaml:"destinations"`
}

// SlackDestination represents a Slack notification destination
type SlackDestination struct {
	Name             string `yaml:"name"`
	APIKey           string `yaml:"api_key"`
	SlackChannel     string `yaml:"slack_channel"`
	GroupingInterval int    `yaml:"grouping_interval,omitempty"`
	UnfurlLinks      *bool  `yaml:"unfurl_links,omitempty"`
}

// TeamsDestination represents a Microsoft Teams notification destination
type TeamsDestination struct {
	Name       string `yaml:"name"`
	WebhookURL string `yaml:"webhookURL"`
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

	// Validate Teams destinations
	for _, d := range config.Destinations.Teams {
		if err := validateTeamsDestination(d); err != nil {
			return nil, fmt.Errorf("invalid Teams destination '%s': %w", d.Name, err)
		}
	}

	return &config, nil
}

func validateSlackDestination(d SlackDestination) error {
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}

	if d.SlackChannel == "" {
		return fmt.Errorf("slack_channel is required")
	}

	if d.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}

	// Validate grouping_interval if provided
	if d.GroupingInterval < 0 {
		return fmt.Errorf("grouping_interval must be non-negative")
	}

	return nil
}

func validateTeamsDestination(d TeamsDestination) error {
	if d.Name == "" {
		return fmt.Errorf("name is required")
	}
	if d.WebhookURL == "" {
		return fmt.Errorf("webhookURL is required")
	}
	return nil
}
