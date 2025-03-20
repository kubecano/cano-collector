package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DestinationsConfig struct {
	Destinations struct {
		Slack []Destination `yaml:"slack"`
		Teams []Destination `yaml:"teams"`
	} `yaml:"destinations"`
}

type Destination struct {
	Name       string `yaml:"name"`
	WebhookURL string `yaml:"webhookURL"`
}

// LoadDestinationsConfig loads the Destinations configuration from the given file
func LoadDestinationsConfig(secretPath string) (*DestinationsConfig, error) {
	var config DestinationsConfig

	configData, err := os.ReadFile(secretPath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
