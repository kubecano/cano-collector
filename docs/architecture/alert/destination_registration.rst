Destination Registration Comparison
==================================

This document explains the differences between cano-collector's static destination registration and Robusta's dynamic sink registration approach.

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
~~~~~~~~~~~~

- **No Runtime Updates**: Requires restart for configuration changes
- **Limited Flexibility**: Cannot add destinations dynamically
- **No Health Checking**: No built-in destination health monitoring
- **No Discovery**: Cannot discover destinations automatically

Robusta's Dynamic Sink Registration
-----------------------------------

Robusta uses a dynamic, registry-based approach for sink management:

SinkRegistry Implementation
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: python

    class SinksRegistry:
        def __init__(self, sinks: Dict[str, SinkBase]):
            self.sinks = sinks
            self.default_sinks = [sink.sink_name for sink in sinks.values() if sink.default]
            if not self.default_sinks:
                logging.warning("No default sinks defined. By default, actions results are ignored.")
            platform_sinks = [sink for sink in sinks.values() if isinstance(sink.params, RobustaSinkParams)]
            self.platform_enabled = len(platform_sinks) > 0

        def get_sink_by_name(self, sink_name: str) -> Optional[SinkBase]:
            return self.sinks.get(sink_name)

        def get_all(self) -> Dict[str, SinkBase]:
            return self.sinks

Dynamic Construction
~~~~~~~~~~~~~~~~~~~

.. code-block:: python

    @classmethod
    def construct_new_sinks(
        cls,
        new_sinks_config: List[SinkConfigBase],
        existing_sinks: Dict[str, SinkBase],
        registry,
    ) -> Dict[str, SinkBase]:
        new_sink_names = [sink_config.get_name() for sink_config in new_sinks_config]
        
        # remove deleted sinks
        deleted_sink_names = [sink_name for sink_name in existing_sinks.keys() if sink_name not in new_sink_names]
        for deleted_sink in deleted_sink_names:
            logging.info(f"Deleting sink {deleted_sink}")
            existing_sinks[deleted_sink].stop()
            del existing_sinks[deleted_sink]

        new_sinks: Dict[str, SinkBase] = dict()
        
        # Reload sinks, order does matter and should be loaded & added to the dict by config order.
        for sink_config in new_sinks_config:
            sink_name = sink_config.get_name()
            exists_sink = existing_sinks.get(sink_name, None)
            
            if not exists_sink:
                logging.info(f"Adding {type(sink_config)} sink named {sink_name}")
                new_sinks[sink_name] = SinkFactory.create_sink(sink_config, registry)
                continue

            is_global_config_changed = exists_sink.is_global_config_changed()
            is_sink_changed = sink_config.get_params() != exists_sink.params or is_global_config_changed
            
            if is_sink_changed:
                config_change_msg = "due to global config change" if is_global_config_changed else "due to param change"
                logging.info(f"Updating {type(sink_config)} sink named {sink_config.get_name()} {config_change_msg}")
                exists_sink.stop()
                new_sinks[sink_name] = SinkFactory.create_sink(sink_config, registry)
                continue

            logging.info("Sink %s not changed", sink_name)
            new_sinks[sink_name] = exists_sink

        return new_sinks

SinkFactory Pattern
~~~~~~~~~~~~~~~~~~

.. code-block:: python

    class SinkFactory:
        __sink_config_mapping: Dict[Type[SinkConfigBase], Type[SinkBase]] = {
            SlackSinkConfigWrapper: SlackSink,
            MsTeamsSinkConfigWrapper: MsTeamsSink,
            KafkaSinkConfigWrapper: KafkaSink,
            DataDogSinkConfigWrapper: DataDogSink,
            # ... more sink types
        }

        @classmethod
        def create_sink(cls, sink_config: SinkConfigBase, registry) -> SinkBase:
            SinkClass = cls.__sink_config_mapping.get(type(sink_config))
            if SinkClass is None:
                raise Exception(f"Sink not supported {type(sink_config)}")
            return SinkClass(sink_config, registry)

Characteristics
~~~~~~~~~~~~~~~

1. **Dynamic Registration**: Sinks can be added/removed at runtime
2. **Configuration Hot-Reload**: Supports configuration updates without restart
3. **Health Checking**: Built-in sink health monitoring
4. **Discovery**: Can discover and register sinks automatically
5. **Factory Pattern**: Uses factory pattern for sink creation

Advantages
~~~~~~~~~~

- **Runtime Flexibility**: Can add/remove sinks without restart
- **Configuration Hot-Reload**: Dynamic configuration updates
- **Health Monitoring**: Built-in health checking capabilities
- **Extensibility**: Easy to add new sink types
- **Discovery**: Can discover sinks automatically

Disadvantages
~~~~~~~~~~~~

- **Complexity**: More complex implementation and configuration
- **Runtime Overhead**: Dynamic discovery and health checking overhead
- **Error Handling**: More complex error handling for dynamic operations
- **Debugging**: Harder to debug dynamic configuration issues

Key Differences Summary
-----------------------

.. list-table::
   :header-rows: 1

   * - Aspect
     - Cano-collector
     - Robusta
   * - Registration Method
     - Static configuration
     - Dynamic registry
   * - Configuration Updates
     - Requires restart
     - Hot-reload supported
   * - Health Checking
     - Manual implementation
     - Built-in health checks
   * - Discovery
     - Manual configuration
     - Automatic discovery
   * - Complexity
     - Simple, predictable
     - Complex, flexible
   * - Runtime Overhead
     - Minimal
     - Moderate
   * - Error Handling
     - Simple
     - Complex

Future Considerations for Cano-collector
----------------------------------------

To bridge the gap with Robusta's capabilities, cano-collector could consider:

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

Example Implementation for Dynamic Loading
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    type DynamicDestinationsLoader struct {
        configPath string
        watcher    *fsnotify.Watcher
        logger     logger.LoggerInterface
        callbacks  []func(*DestinationsConfig)
    }

    func (d *DynamicDestinationsLoader) Watch() error {
        watcher, err := fsnotify.NewWatcher()
        if err != nil {
            return err
        }
        d.watcher = watcher

        go func() {
            for event := range watcher.Events {
                if event.Op&fsnotify.Write == fsnotify.Write {
                    if strings.HasSuffix(event.Name, "destinations.yaml") {
                        d.reloadConfiguration()
                    }
                }
            }
        }()

        return watcher.Add(d.configPath)
    }

    func (d *DynamicDestinationsLoader) reloadConfiguration() {
        config, err := d.Load()
        if err != nil {
            d.logger.Errorf("Failed to reload configuration: %v", err)
            return
        }

        // Notify all registered callbacks
        for _, callback := range d.callbacks {
            callback(config)
        }
    }

This comparison highlights the trade-offs between simplicity and flexibility in destination management approaches. 