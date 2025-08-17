package actions

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	actions_interfaces "github.com/kubecano/cano-collector/pkg/workflow/actions/interfaces"
)

// PodLogsActionConfig contains configuration for PodLogsAction
type PodLogsActionConfig struct {
	actions_interfaces.ActionConfig `yaml:",inline"`

	// MaxLines maximum number of lines to retrieve
	MaxLines int `yaml:"max_lines" json:"max_lines"`

	// SinceTime retrieve logs since this time (RFC3339 format)
	SinceTime string `yaml:"since_time" json:"since_time"`

	// TailLines number of lines from the end of the logs to show
	TailLines int `yaml:"tail_lines" json:"tail_lines"`

	// Container name to get logs from (empty means all containers)
	Container string `yaml:"container" json:"container"`

	// Previous get logs from previous instance of the container
	Previous bool `yaml:"previous" json:"previous"`

	// Timestamps add timestamps to each log line
	Timestamps bool `yaml:"timestamps" json:"timestamps"`

	// Java-specific configuration
	JavaSpecific bool `yaml:"java_specific" json:"java_specific"`

	// File naming configuration
	IncludeTimestamp bool   `yaml:"include_timestamp" json:"include_timestamp"`
	IncludeContainer bool   `yaml:"include_container" json:"include_container"`
	TimestampFormat  string `yaml:"timestamp_format" json:"timestamp_format"`
}

// NewPodLogsActionConfigWithDefaults creates a new PodLogsActionConfig with defaults from environment variables
func NewPodLogsActionConfigWithDefaults(baseConfig actions_interfaces.ActionConfig) PodLogsActionConfig {
	config := PodLogsActionConfig{
		ActionConfig: baseConfig,
		// Default values using environment variables
		MaxLines:         getEnvInt("WORKFLOW_POD_LOGS_DEFAULT_MAX_LINES", 1000),
		TailLines:        getEnvInt("WORKFLOW_POD_LOGS_DEFAULT_TAIL_LINES", 100),
		Previous:         getEnvBool("WORKFLOW_POD_LOGS_DEFAULT_PREVIOUS", false),
		Timestamps:       getEnvBool("WORKFLOW_POD_LOGS_DEFAULT_TIMESTAMPS", true),
		IncludeTimestamp: getEnvBool("WORKFLOW_POD_LOGS_INCLUDE_TIMESTAMP", true),
		IncludeContainer: getEnvBool("WORKFLOW_POD_LOGS_INCLUDE_CONTAINER", true),
		TimestampFormat:  getEnvString("WORKFLOW_POD_LOGS_TIMESTAMP_FORMAT", "20060102-150405"),
	}

	return config
}

// ApplyJavaDefaults applies Java-specific defaults if container is detected as Java
func (c *PodLogsActionConfig) ApplyJavaDefaults() {
	c.JavaSpecific = true
	c.MaxLines = getEnvInt("WORKFLOW_POD_LOGS_JAVA_DEFAULT_MAX_LINES", 5000)
	c.TailLines = getEnvInt("WORKFLOW_POD_LOGS_JAVA_DEFAULT_TAIL_LINES", 500)
}

// Validate validates the PodLogsAction configuration
func (c *PodLogsActionConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("action name cannot be empty")
	}

	if c.MaxLines < 0 {
		return fmt.Errorf("max_lines must be non-negative")
	}

	if c.TailLines < 0 {
		return fmt.Errorf("tail_lines must be non-negative")
	}

	if c.SinceTime != "" {
		// Basic validation - RFC3339 format validation will be done during execution
		if strings.TrimSpace(c.SinceTime) == "" {
			return fmt.Errorf("since_time cannot be empty string")
		}
	}

	if c.TimestampFormat == "" {
		return fmt.Errorf("timestamp_format cannot be empty")
	}

	return nil
}

// UpdateFromParameters updates configuration from action parameters
func (c *PodLogsActionConfig) UpdateFromParameters(parameters map[string]interface{}) error {
	if parameters == nil {
		return nil
	}

	if maxLines, ok := parameters["max_lines"]; ok {
		if maxLinesInt, ok := maxLines.(int); ok {
			c.MaxLines = maxLinesInt
		} else {
			return fmt.Errorf("max_lines must be an integer")
		}
	}

	if tailLines, ok := parameters["tail_lines"]; ok {
		if tailLinesInt, ok := tailLines.(int); ok {
			c.TailLines = tailLinesInt
		} else {
			return fmt.Errorf("tail_lines must be an integer")
		}
	}

	if previous, ok := parameters["previous"]; ok {
		if previousBool, ok := previous.(bool); ok {
			c.Previous = previousBool
		} else {
			return fmt.Errorf("previous must be a boolean")
		}
	}

	if timestamps, ok := parameters["timestamps"]; ok {
		if timestampsBool, ok := timestamps.(bool); ok {
			c.Timestamps = timestampsBool
		} else {
			return fmt.Errorf("timestamps must be a boolean")
		}
	}

	if container, ok := parameters["container"]; ok {
		if containerStr, ok := container.(string); ok {
			c.Container = containerStr
		} else {
			return fmt.Errorf("container must be a string")
		}
	}

	if sinceTime, ok := parameters["since_time"]; ok {
		if sinceTimeStr, ok := sinceTime.(string); ok {
			c.SinceTime = sinceTimeStr
		} else {
			return fmt.Errorf("since_time must be a string")
		}
	}

	if javaSpecific, ok := parameters["java_specific"]; ok {
		if javaSpecificBool, ok := javaSpecific.(bool); ok {
			c.JavaSpecific = javaSpecificBool
			if javaSpecificBool {
				c.ApplyJavaDefaults()
			}
		} else {
			return fmt.Errorf("java_specific must be a boolean")
		}
	}

	if includeTimestamp, ok := parameters["include_timestamp"]; ok {
		if includeTimestampBool, ok := includeTimestamp.(bool); ok {
			c.IncludeTimestamp = includeTimestampBool
		} else {
			return fmt.Errorf("include_timestamp must be a boolean")
		}
	}

	if includeContainer, ok := parameters["include_container"]; ok {
		if includeContainerBool, ok := includeContainer.(bool); ok {
			c.IncludeContainer = includeContainerBool
		} else {
			return fmt.Errorf("include_container must be a boolean")
		}
	}

	if timestampFormat, ok := parameters["timestamp_format"]; ok {
		if timestampFormatStr, ok := timestampFormat.(string); ok {
			c.TimestampFormat = timestampFormatStr
		} else {
			return fmt.Errorf("timestamp_format must be a string")
		}
	}

	return nil
}

// IsJavaContainer detects if container is a Java application based on image name
func IsJavaContainer(containerName, imageName string) bool {
	javaIndicators := []string{
		"java", "openjdk", "eclipse-temurin", "adoptopenjdk", "amazoncorretto",
		"spring", "tomcat", "jetty", "wildfly", "jboss",
		"maven", "gradle", "kafka", "elasticsearch", "solr",
	}

	imageNameLower := strings.ToLower(imageName)
	containerNameLower := strings.ToLower(containerName)

	for _, indicator := range javaIndicators {
		if strings.Contains(imageNameLower, indicator) || strings.Contains(containerNameLower, indicator) {
			return true
		}
	}

	return false
}

// Helper functions for environment variables
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvString(key string, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
