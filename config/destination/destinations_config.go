package destination

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// DestinationsConfig configuration for all destinations
type DestinationsConfig struct {
	Destinations struct {
		Slack []SlackDestinationConfig `yaml:"slack,omitempty"`
		Teams []TeamsDestinationConfig `yaml:"teams,omitempty"`
	} `yaml:"destinations"`
}

// BaseDestinationConfig base configuration used in all destinations
type BaseDestinationConfig struct {
	Name string `yaml:"name"`
}

// SlackDestinationConfig configuration for Slack
type SlackDestinationConfig struct {
	BaseDestinationConfig `yaml:",inline"`
	Token                 string `yaml:"token"`
	Channel               string `yaml:"channel"`
	SigningKey            string `yaml:"signingKey,omitempty"`
	AccountID             string `yaml:"accountId,omitempty"`
	ClusterName           string `yaml:"clusterName,omitempty"`
}

// TeamsDestinationConfig configuration for Microsoft Teams
type TeamsDestinationConfig struct {
	BaseDestinationConfig `yaml:",inline"`
	WebhookURL            string `yaml:"webhookURL"`
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

	// Validate the configuration
	for _, d := range config.Destinations.Slack {
		if d.Name == "" || d.Token == "" || d.Channel == "" {
			return nil, fmt.Errorf("invalid Slack destination entry: name, token and channel must be set")
		}
	}

	for _, d := range config.Destinations.Teams {
		if d.Name == "" || d.WebhookURL == "" {
			return nil, fmt.Errorf("invalid Teams destination entry: name and webhookURL must be set")
		}
	}

	// Similar validation can be added for Email and OpsGenie destinations

	return &config, nil
}
