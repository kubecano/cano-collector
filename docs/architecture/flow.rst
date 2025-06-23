Cano-collector Flow Architecture
================================

This document describes the detailed flow of how cano-collector processes alerts from reception to notification delivery, providing a comprehensive understanding of the internal architecture and data flow.

Alert Processing Flow
---------------------

1. **Alert Reception**
   
   **Endpoint:** `/api/alerts`
   
   - Alertmanager sends POST request with Prometheus alert data
   - `AlertHandler.HandleAlert()` receives the request
   - Request body validated for non-empty content
   - JSON parsed into `template.Data` structure (Alertmanager format)
   - Basic validation ensures required fields are present
   - Alert converted to internal `Issue` model
   - Metrics recorded for alert reception

2. **Alert Parsing and Validation**
   
   - Alert data parsed from Alertmanager format
   - Validation checks:
     - Receiver field present
     - Status field present (firing/resolved)
     - At least one alert in the alerts array
   - Invalid alerts return HTTP 400 with error details
   - Valid alerts proceed to processing pipeline

3. **Issue Creation**
   
   - `template.Data` converted to internal `Issue` model
   - `Issue` contains:
     - Title and description from alert annotations
     - Severity mapped from alert labels
     - Labels and annotations preserved
     - Timestamps (start/end times)
     - Resource information (namespace, pod, etc.)
   - Enrichment blocks applied to add context

4. **Team Routing (Planned)**
   
   - Routing engine evaluates `Issue` against team configurations
   - Routing rules (to be implemented) determine which team receives the alert
   - Rules can be based on:
     - Alert name (identifier)
     - Namespace
     - Severity level
     - Custom labels
   - Team selection determines which destinations receive the notification

5. **Destination Resolution**
   
   - Selected team's `destinations` list resolved
   - Each destination name looked up in destinations configuration
   - Destination instances retrieved with their specific configuration
   - Multiple destinations can receive the same alert

6. **Message Formatting and Sending**
   
   - Each destination processes the `Issue` independently
   - Destination delegates formatting to appropriate `Sender`
   - `SenderFactory` creates correct sender type based on destination
   - Sender formats `Issue` into target API format:
     - Slack: Block Kit message structure
     - MS Teams: Adaptive Card format
     - OpsGenie: Alert API payload
     - etc.
   - HTTP request sent to destination endpoint
   - Response handling and error management

Configuration Flow
------------------

1. **Startup Configuration Loading**
   
   - `config.LoadConfig()` called during application startup
   - Two configuration files loaded:
     - Destinations: `/etc/cano-collector/destinations/destinations.yaml`
     - Teams: `/etc/cano-collector/teams/teams.yaml`
   - `FileDestinationsLoader` parses destinations configuration
   - `FileTeamsLoader` parses teams configuration
   - Configuration validated for required fields

2. **Destination Factory Initialization**
   
   - `DestinationFactory` created with logger and HTTP client
   - For each configured destination:
     - Appropriate `Sender` instance created
     - Sender configured with destination-specific parameters
     - `Destination` wrapper created around sender
     - Destination registered in routing system

3. **Team Configuration Processing**
   
   - Teams configuration loaded into `TeamsConfig` structure
   - Each team contains:
     - Unique name
     - List of destination names
   - Destination names validated against actual destinations
   - Teams made available to routing engine

Data Flow Architecture
----------------------

```
Alertmanager → AlertHandler → Issue → Team Router → Destinations → Senders → External APIs
     ↓              ↓          ↓         ↓              ↓           ↓
  template.Data  Validation  Issue    Team Match   Destination   Sender    HTTP Request
                                      Resolution   Resolution    Format
```

Key Components in Flow
----------------------

1. **AlertHandler**
   - Entry point for alert processing
   - Handles HTTP request/response
   - Performs initial validation
   - Records metrics

2. **Issue Model**
   - Central data structure
   - Contains all alert information
   - Supports enrichment blocks
   - Passed through entire pipeline

3. **Team Router (Planned)**
   - Determines which team receives alert
   - Implements routing rules
   - Maps teams to destinations

4. **Destination**
   - Holds configuration for notification endpoint
   - Delegates to appropriate sender
   - Manages destination-specific logic

5. **Sender**
   - Formats Issue for target API
   - Handles HTTP communication
   - Manages API-specific requirements

Error Handling
--------------

1. **Alert Reception Errors**
   - Invalid JSON: HTTP 400 with parsing error
   - Missing required fields: HTTP 400 with validation error
   - Empty body: HTTP 400 with appropriate message

2. **Configuration Errors**
   - Missing configuration files: Application startup failure
   - Invalid YAML: Configuration loading failure
   - Missing destinations: Validation error during startup

3. **Sending Errors**
   - HTTP failures: Logged with retry logic (planned)
   - API errors: Error responses logged
   - Timeout errors: Configurable timeout handling

4. **Routing Errors**
   - No matching team: Fallback handling (planned)
   - Invalid destination references: Configuration validation error

Metrics and Observability
-------------------------

1. **Alert Metrics**
   - Alerts received per receiver
   - Alert status distribution
   - Processing time measurements

2. **Destination Metrics**
   - Messages sent per destination
   - Success/failure rates
   - Response time tracking

3. **Routing Metrics**
   - Team matching statistics
   - Routing decision tracking
   - Fallback usage metrics

Future Enhancements
-------------------

1. **Async Processing**
   - Implement message queue for alert processing
   - Background processing of alerts
   - Improved throughput and reliability

2. **Advanced Routing**
   - Complex matching rules
   - Dynamic routing based on alert content
   - Fallback routing mechanisms

3. **Enrichment Pipeline**
   - Automatic context gathering
   - Resource status enrichment
   - Custom enrichment actions

4. **Monitoring and Alerting**
   - Self-monitoring capabilities
   - Alert on processing failures
   - Performance metrics dashboard

This flow architecture provides a clear separation of concerns while maintaining simplicity and extensibility. Each component has a well-defined responsibility, making the system easy to understand, test, and extend. 