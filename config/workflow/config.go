package workflow

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

// WorkflowConfig represents the complete workflow configuration structure
type WorkflowConfig struct {
	ActiveWorkflows []WorkflowDefinition `yaml:"active_workflows" json:"active_workflows"`
}

// ConfigLoader handles loading workflow configuration from various sources
type ConfigLoader struct {
	configPath string
}

// NewConfigLoader creates a new workflow config loader
func NewConfigLoader(configPath string) *ConfigLoader {
	return &ConfigLoader{
		configPath: configPath,
	}
}

// LoadConfig loads workflow configuration from the configured path
func (c *ConfigLoader) LoadConfig() (*WorkflowConfig, error) {
	return LoadWorkflowConfig(c.configPath)
}

// LoadWorkflowConfig loads workflow configuration from a YAML file
func LoadWorkflowConfig(configPath string) (*WorkflowConfig, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path cannot be empty")
	}

	if !fileExists(configPath) {
		return nil, fmt.Errorf("workflow config file not found at: %s", configPath)
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open workflow config file '%s': %w", configPath, err)
	}
	defer file.Close()

	return LoadWorkflowConfigFromReader(file)
}

// LoadWorkflowConfigFromReader loads workflow configuration from an io.Reader
func LoadWorkflowConfigFromReader(reader io.Reader) (*WorkflowConfig, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow config: %w", err)
	}

	var config WorkflowConfig
	if err := yaml.Unmarshal(content, &config); err != nil {
		return nil, fmt.Errorf("failed to parse workflow config YAML: %w", err)
	}

	// Validate configuration
	if err := ValidateWorkflowConfig(&config); err != nil {
		return nil, fmt.Errorf("workflow config validation failed: %w", err)
	}

	return &config, nil
}

// LoadWorkflowConfigFromString loads workflow configuration from a YAML string
func LoadWorkflowConfigFromString(yamlContent string) (*WorkflowConfig, error) {
	var config WorkflowConfig
	if err := yaml.Unmarshal([]byte(yamlContent), &config); err != nil {
		return nil, fmt.Errorf("failed to parse workflow config YAML: %w", err)
	}

	// Validate configuration
	if err := ValidateWorkflowConfig(&config); err != nil {
		return nil, fmt.Errorf("workflow config validation failed: %w", err)
	}

	return &config, nil
}

// ValidateWorkflowConfig validates the entire workflow configuration
func ValidateWorkflowConfig(config *WorkflowConfig) error {
	if config == nil {
		return fmt.Errorf("workflow config cannot be nil")
	}

	if len(config.ActiveWorkflows) == 0 {
		return fmt.Errorf("no active workflows configured")
	}

	// Check for duplicate workflow names
	workflowNames := make(map[string]bool)
	for i, workflow := range config.ActiveWorkflows {
		if err := workflow.Validate(); err != nil {
			return fmt.Errorf("workflow %d validation failed: %w", i, err)
		}

		if workflowNames[workflow.Name] {
			return fmt.Errorf("duplicate workflow name '%s' found", workflow.Name)
		}
		workflowNames[workflow.Name] = true
	}

	return nil
}

// GetWorkflowsByTriggerType returns workflows that have triggers of the specified type
func (c *WorkflowConfig) GetWorkflowsByTriggerType(triggerType string) []WorkflowDefinition {
	var matchingWorkflows []WorkflowDefinition

	for _, workflow := range c.ActiveWorkflows {
		if workflow.HasTriggerType(triggerType) {
			matchingWorkflows = append(matchingWorkflows, workflow)
		}
	}

	return matchingWorkflows
}

// GetWorkflowByName returns a workflow by its name
func (c *WorkflowConfig) GetWorkflowByName(name string) (*WorkflowDefinition, error) {
	for _, workflow := range c.ActiveWorkflows {
		if workflow.Name == name {
			return &workflow, nil
		}
	}
	return nil, fmt.Errorf("workflow '%s' not found", name)
}

// GetWorkflowCount returns the number of active workflows
func (c *WorkflowConfig) GetWorkflowCount() int {
	return len(c.ActiveWorkflows)
}

// GetTriggerTypeCounts returns a map of trigger types to their counts
func (c *WorkflowConfig) GetTriggerTypeCounts() map[string]int {
	counts := make(map[string]int)

	for _, workflow := range c.ActiveWorkflows {
		for _, trigger := range workflow.Triggers {
			triggerType := trigger.GetTriggerType()
			counts[triggerType]++
		}
	}

	return counts
}

// DefaultWorkflowConfig returns a default workflow configuration for testing/development
func DefaultWorkflowConfig() *WorkflowConfig {
	return &WorkflowConfig{
		ActiveWorkflows: []WorkflowDefinition{
			{
				Name: "AlertManagerAlertEnrichment",
				Triggers: []TriggerDefinition{
					{
						OnAlertmanagerAlert: &AlertmanagerAlertTrigger{
							AlertName: "", // Match all alerts
							Status:    "firing",
						},
					},
				},
				Actions: []ActionDefinition{
					{
						ActionType: "placeholder_action",
						RawData: map[string]interface{}{
							"placeholder_action": map[string]interface{}{},
						},
					},
				},
				Stop: false,
			},
		},
	}
}

// fileExists checks if a file exists and is not a directory
func fileExists(path string) bool {
	if info, err := os.Stat(path); err == nil {
		return !info.IsDir()
	}
	return false
}
