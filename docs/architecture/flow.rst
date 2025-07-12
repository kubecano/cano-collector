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

3. **Workflow Processing (Planned)**
   
   - `template.Data` evaluated against workflow trigger conditions
   - Matching workflows execute their defined actions (both built-in and custom)
   - All workflows run in parallel when their triggers match
   - Workflow actions can:
     - **Create Issue** - Convert alert data to internal Issue model
     - **Enrich data** - Add contextual information from Kubernetes cluster
     - **Filter alerts** - Decide whether to process the alert
     - **Transform data** - Modify alert data before Issue creation
     - **Generate multiple Issues** - Create several Issues from one alert
     - **Execute custom logic** - Organization-specific processing
   - Workflows that don't create Issues result in no further processing

4. **Issue Creation**
   
   - Issues created by workflow actions (typically `create_issue` action)
   - `Issue` contains:
     - Title and description from alert annotations (or workflow-modified data)
     - Severity mapped from alert labels (or workflow-adjusted)
     - Labels and annotations preserved (or workflow-enriched)
     - Timestamps (start/end times)
     - Resource information (namespace, pod, etc.)
     - Any additional enrichment data added by workflows
   - Multiple Issues can be created from a single alert

5. **Team Routing (Planned)**
   
   - Routing engine evaluates `Issue` against team configurations
   - Routing rules (to be implemented) determine which team receives the alert
   - Rules can be based on:
     - Alert name (identifier)
     - Namespace
     - Severity level
     - Custom labels
   - Team selection determines which destinations receive the notification

6. **Destination Resolution**
   
   - Selected team's `destinations` list resolved
   - Each destination name looked up in destinations configuration
   - Destination instances retrieved with their specific configuration
   - Multiple destinations can receive the same alert

7. **Message Formatting and Sending**
   
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
   - Configuration files loaded:
     - Destinations: `/etc/cano-collector/destinations/destinations.yaml`
     - Teams: `/etc/cano-collector/teams/teams.yaml`
     - Workflows: `/etc/cano-collector/workflows/workflows.yaml` (planned)
   - `FileDestinationsLoader` parses destinations configuration
   - `FileTeamsLoader` parses teams configuration
   - `WorkflowConfigLoader` parses workflow configuration (planned)
   - Configuration validated for required fields

2. **Destination Factory Initialization**
   
   - `DestinationFactory` created with logger and HTTP client
   - For each configured destination:
     - Appropriate `Sender` instance created
     - Sender configured with destination-specific parameters
     - `Destination` wrapper created around sender
     - Destination registered in routing system

3. **Workflow Configuration Processing (Planned)**
   
   - Workflows configuration loaded into `WorkflowConfig` structure
   - Each workflow contains:
     - Unique name
     - List of trigger conditions
     - List of actions to execute
   - Trigger conditions and actions validated
   - Workflows registered in workflow engine

4. **Team Configuration Processing**
   
   - Teams configuration loaded into `TeamsConfig` structure
   - Each team contains:
     - Unique name
     - List of destination names
   - Destination names validated against actual destinations
   - Teams made available to routing engine

Data Flow Architecture
----------------------

::

    Alertmanager → AlertHandler → template.Data → Workflow → Issue(s) → Team Router → Destinations → Senders → External APIs
         ↓              ↓             ↓          Engine       ↓           ↓              ↓           ↓
      template.Data  Validation   Alert Data    Triggers   create_issue  Team Match   Destination   Sender    HTTP Request
                                                Actions     Action      Resolution   Resolution    Format

Key Components in Flow
----------------------

1. **AlertHandler**
   - Entry point for alert processing
   - Handles HTTP request/response
   - Performs initial validation
   - Records metrics

2. **Workflow Engine (Planned)**
   - Evaluates template.Data against workflow triggers
   - Executes matching workflow actions
   - Coordinates built-in and custom workflows
   - Responsible for Issue creation through actions
   - Can filter, transform, or enrich alert data

3. **Issue Model**
   - Central data structure created by workflows
   - Contains all alert information (potentially enriched)
   - Supports enrichment blocks
   - Passed through team routing pipeline

4. **Team Router (Planned)**
   - Determines which team receives alert
   - Implements routing rules
   - Maps teams to destinations

5. **Destination**
   - Holds configuration for notification endpoint
   - Delegates to appropriate sender
   - Manages destination-specific logic

6. **Sender**
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

4. **Workflow Errors (Planned)**
   - Workflow trigger evaluation failures: Logged and skipped
   - Action execution failures: Logged with error details
   - Custom workflow runtime errors: Timeout and exception handling
   - Configuration validation errors: Prevent startup

5. **Routing Errors**
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

3. **Workflow Metrics (Planned)**
   - Workflow execution statistics
   - Trigger matching rates
   - Action execution time
   - Built-in vs custom workflow performance
   - Workflow enrichment effectiveness

4. **Routing Metrics**
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

3. **Monitoring and Alerting**
   - Self-monitoring capabilities
   - Alert on processing failures
   - Performance metrics dashboard

This flow architecture provides a clear separation of concerns while maintaining simplicity and extensibility. Each component has a well-defined responsibility, making the system easy to understand, test, and extend. 