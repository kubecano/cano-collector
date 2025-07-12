Destination Registration Architecture
=====================================

This document explains cano-collector's static destination registration approach and its benefits for reliability and simplicity.

Cano-collector's Static Destination Registration
------------------------------------------------

Cano-collector uses a static, configuration-driven approach for destination registration:

Configuration Loading
~~~~~~~~~~~~~~~~~~~~~

Destinations are configured in Helm values and loaded at startup:

.. code-block:: yaml

    # values.yaml
    destinations:
      slack:
        - name: "alerts-prod"
          webhookURL: "https://hooks.slack.com/services/..."
        - name: "alerts-dev"
          webhookURL: "https://hooks.slack.com/services/..."
      msteams:
        - name: "ops-team"
          webhookURL: "https://your-org.webhook.office.com/..."

Loading Process
~~~~~~~~~~~~~~~

.. code-block:: go

    type DestinationsLoader interface {
        Load() (*DestinationsConfig, error)
    }

    type FileDestinationsLoader struct {
        Path string
    }

    func (f *FileDestinationsLoader) Load() (*DestinationsConfig, error) {
        file, err := os.Open(f.Path)
        if err != nil {
            return nil, fmt.Errorf("cannot open destination config: %w", err)
        }
        defer file.Close()

        return parseDestinationsYAML(file)
    }

    func LoadConfig() (Config, error) {
        loader := NewFileConfigLoader(
            "/etc/cano-collector/destinations/destinations.yaml",
            "/etc/cano-collector/teams/teams.yaml",
            "/etc/cano-collector/workflows/workflows.yaml",
        )
        return LoadConfigWithLoader(loader)
    }

Characteristics
~~~~~~~~~~~~~~~

1. **Static Configuration**: Destinations defined in Helm values
2. **Startup Loading**: Configuration loaded once at application startup
3. **No Runtime Changes**: Requires configuration update and restart
4. **Simple Structure**: Direct mapping from configuration to destinations
5. **Type-based Grouping**: Destinations grouped by type (slack, teams, etc.)

Advantages
~~~~~~~~~~

- **Simplicity**: Easy to understand and configure
- **Predictability**: Clear, static configuration
- **Reliability**: No runtime configuration complexity
- **Performance**: No dynamic discovery overhead
- **Security**: Configuration controlled through GitOps

Disadvantages
~~~~~~~~~~~~~

- **No Runtime Updates**: Requires restart for configuration changes
- **Limited Flexibility**: Cannot add destinations dynamically
- **No Health Checking**: No built-in destination health monitoring
- **No Discovery**: Cannot discover destinations automatically

Future Considerations for Cano-collector
----------------------------------------

To enhance destination management capabilities, cano-collector could consider:

1. **Dynamic Configuration Loading**:
   - Implement configuration file watching
   - Support for runtime configuration updates
   - Graceful configuration reloading

2. **Destination Health Monitoring**:
   - Add health check endpoints for destinations
   - Implement destination availability monitoring
   - Add metrics for destination health status

3. **Enhanced Registration**:
   - Implement destination registry pattern
   - Support for dynamic destination discovery
   - Add destination lifecycle management

4. **Configuration Validation**:
   - Real-time configuration validation
   - Better error reporting for configuration issues
   - Configuration schema validation

This approach prioritizes simplicity and reliability over dynamic flexibility, making it well-suited for environments where configuration stability is valued over runtime adaptability. 