Development Guide
=================

This guide is for developers implementing and extending cano-collector. It covers the codebase structure, development practices, and how to implement new features.

Codebase Structure
------------------

**Current Structure:**
```
cano-collector/
├── main.go                 # Application entry point
├── config/                 # Configuration management
│   ├── config.go          # Main configuration structure
│   ├── destination/       # Destination configuration
│   └── team/             # Team configuration
├── pkg/                   # Core packages
│   ├── alert/            # Alert processing
│   ├── core/             # Core data models
│   │   └── issue/        # Issue and enrichment models
│   ├── destination/      # Destination management
│   ├── sender/           # Message senders
│   ├── router/           # HTTP routing
│   ├── logger/           # Logging
│   ├── metric/           # Metrics collection
│   ├── health/           # Health checks
│   ├── tracer/           # Distributed tracing
│   └── util/             # Utilities
└── mocks/                # Generated mocks for testing
```

**Target Structure (After Implementation):**
```
cano-collector/
├── main.go                           # Application entry point with dependency injection
├── config/                           # Configuration management
│   ├── config.go                    # Main configuration structure
│   ├── destination/                 # Destination configuration
│   │   ├── destinations_config.go   # Destination config loading
│   │   ├── destinations_config_test.go
│   │   └── types.go                 # Destination type definitions
│   ├── team/                        # Team configuration
│   │   ├── teams_config.go          # Team config loading
│   │   ├── teams_config_test.go
│   │   └── types.go                 # Team type definitions
│   └── workflow/                    # Workflow configuration
│       ├── workflow_config.go       # Workflow config loading
│       ├── workflow_config_test.go
│       └── types.go                 # Workflow type definitions
├── pkg/                             # Core packages
│   ├── alert/                       # Alert processing
│   │   ├── alert_handler.go         # HTTP handler for alerts
│   │   ├── alert_handler_test.go
│   │   ├── converter.go             # Alert to Issue conversion
│   │   ├── converter_test.go
│   │   ├── deduplication.go         # Deduplication system
│   │   ├── deduplication_test.go
│   │   ├── queue.go                 # Async processing queue
│   │   ├── queue_test.go
│   │   ├── relabeling.go            # Alert relabeling
│   │   ├── relabeling_test.go
│   │   ├── processor.go             # Alert processor
│   │   ├── processor_test.go
│   │   └── types.go                 # Alert type definitions
│   ├── core/                        # Core data models
│   │   ├── issue/                   # Issue and enrichment models
│   │   │   ├── issue.go             # Issue data structure
│   │   │   ├── issue_test.go
│   │   │   ├── blocks.go            # Enrichment blocks
│   │   │   ├── blocks_test.go
│   │   │   ├── builder.go           # Issue builder pattern
│   │   │   ├── builder_test.go
│   │   │   └── types.go             # Issue type definitions
│   │   └── prometheus/              # Prometheus alert models
│   │       ├── alert.go             # PrometheusAlert structure
│   │       ├── alert_test.go
│   │       └── types.go             # Prometheus type definitions
│   ├── workflow/                    # Workflow processing
│   │   ├── workflow.go              # Workflow definitions
│   │   ├── workflow_test.go
│   │   ├── processor.go             # Workflow processor
│   │   ├── processor_test.go
│   │   ├── actions/                 # Workflow actions
│   │   │   ├── pod_logs.go          # Pod logs enrichment
│   │   │   ├── pod_logs_test.go
│   │   │   ├── resource_status.go   # Resource status enrichment
│   │   │   ├── resource_status_test.go
│   │   │   ├── pod_events.go        # Pod events enrichment
│   │   │   ├── pod_events_test.go
│   │   │   ├── node_metrics.go      # Node metrics enrichment
│   │   │   ├── node_metrics_test.go
│   │   │   ├── custom_script.go     # External TypeScript script execution
│   │   │   ├── custom_script_test.go
│   │   │   └── base.go              # Base action interface
│   │   ├── triggers.go              # Workflow triggers
│   │   ├── triggers_test.go
│   │   └── types.go                 # Workflow type definitions
│   ├── destination/                 # Destination management
│   │   ├── router.go                # Destination routing
│   │   ├── router_test.go
│   │   ├── matcher.go               # Team/destination matching
│   │   ├── matcher_test.go
│   │   └── types.go                 # Destination type definitions
│   ├── sender/                      # Message senders
│   │   ├── sender.go                # Base sender interface
│   │   ├── slack/                   # Slack sender
│   │   │   ├── slack_sender.go
│   │   │   ├── slack_sender_test.go
│   │   │   └── blocks.go            # Slack block rendering
│   │   ├── msteams/                 # MS Teams sender
│   │   │   ├── msteams_sender.go
│   │   │   ├── msteams_sender_test.go
│   │   │   └── cards.go             # Adaptive card rendering
│   │   ├── jira/                    # Jira sender
│   │   │   ├── jira_sender.go
│   │   │   ├── jira_sender_test.go
│   │   │   └── fields.go            # Jira field mapping
│   │   ├── servicenow/              # ServiceNow sender
│   │   │   ├── servicenow_sender.go
│   │   │   ├── servicenow_sender_test.go
│   │   │   └── incidents.go         # Incident creation
│   │   ├── datadog/                 # DataDog sender
│   │   │   ├── datadog_sender.go
│   │   │   ├── datadog_sender_test.go
│   │   │   └── events.go            # Event creation
│   │   ├── kafka/                   # Kafka sender
│   │   │   ├── kafka_sender.go
│   │   │   ├── kafka_sender_test.go
│   │   │   └── messages.go          # Message serialization
│   │   ├── webhook/                 # Generic webhook sender
│   │   │   ├── webhook_sender.go
│   │   │   ├── webhook_sender_test.go
│   │   │   └── templates.go         # Template rendering
│   │   ├── opsgenie/                # OpsGenie sender
│   │   │   ├── opsgenie_sender.go
│   │   │   ├── opsgenie_sender_test.go
│   │   │   └── alerts.go            # Alert creation
│   │   ├── pagerduty/               # PagerDuty sender
│   │   │   ├── pagerduty_sender.go
│   │   │   ├── pagerduty_sender_test.go
│   │   │   └── incidents.go         # Incident creation
│   │   └── common/                  # Common sender utilities
│   │       ├── http_client.go       # HTTP client wrapper
│   │       ├── retry.go             # Retry logic
│   │       ├── rate_limiter.go      # Rate limiting
│   │       └── validation.go        # Payload validation
│   ├── router/                      # HTTP routing
│   │   ├── router.go                # Router setup
│   │   ├── router_test.go
│   │   ├── middleware/              # HTTP middleware
│   │   │   ├── logging.go           # Request logging
│   │   │   ├── metrics.go           # Request metrics
│   │   │   ├── tracing.go           # Request tracing
│   │   │   ├── auth.go              # Authentication
│   │   │   └── cors.go              # CORS handling
│   │   └── handlers/                # HTTP handlers
│   │       ├── alerts.go            # Alert endpoint handler
│   │       ├── health.go            # Health check handler
│   │       ├── metrics.go           # Metrics endpoint handler
│   │       └── config.go            # Config endpoint handler
│   ├── logger/                      # Logging
│   │   ├── logger.go                # Logger interface and implementation
│   │   ├── logger_test.go
│   │   ├── formatters/              # Log formatters
│   │   │   ├── json.go              # JSON formatter
│   │   │   ├── text.go              # Text formatter
│   │   │   └── structured.go        # Structured formatter
│   │   └── levels.go                # Log level definitions
│   ├── metric/                      # Metrics collection
│   │   ├── metric.go                # Metrics interface
│   │   ├── metric_test.go
│   │   ├── prometheus/              # Prometheus metrics
│   │   │   ├── collector.go         # Metrics collector
│   │   │   ├── alerts.go            # Alert metrics
│   │   │   ├── destinations.go      # Destination metrics
│   │   │   └── system.go            # System metrics
│   │   └── types.go                 # Metric type definitions
│   ├── health/                      # Health checks
│   │   ├── health.go                # Health interface
│   │   ├── health_test.go
│   │   ├── checks/                  # Health check implementations
│   │   │   ├── config.go            # Configuration health check
│   │   │   ├── destinations.go      # Destination health check
│   │   │   ├── database.go          # Database health check
│   │   │   └── external.go          # External service health check
│   │   └── types.go                 # Health check type definitions
│   ├── tracer/                      # Distributed tracing
│   │   ├── tracer.go                # Tracer interface
│   │   ├── tracer_test.go
│   │   ├── otel/                    # OpenTelemetry implementation
│   │   │   ├── tracer.go            # OTEL tracer
│   │   │   ├── spans.go             # Span management
│   │   │   └── propagation.go       # Context propagation
│   │   └── types.go                 # Tracing type definitions
│   └── util/                        # Utilities
│       ├── http_client.go           # HTTP client utilities
│       ├── http_client_test.go
│       ├── crypto.go                # Cryptographic utilities
│       ├── crypto_test.go
│       ├── time.go                  # Time utilities
│       ├── time_test.go
│       └── validation.go            # Validation utilities
├── mocks/                           # Generated mocks for testing
│   ├── alert_handler_mock.go
│   ├── destinations_loader_mock.go
│   ├── fullconfig_loader_mock.go
│   ├── health_mock.go
│   ├── http_client_mock.go
│   ├── logger_mock.go
│   ├── metrics_mock.go
│   ├── router_mock.go
│   ├── teams_loader_mock.go
│   └── tracer_mock.go
├── helm/                            # Helm chart
│   └── cano-collector/
│       ├── Chart.yaml
│       ├── values.yaml
│       └── templates/
│           ├── deployment.yaml
│           ├── service.yaml
│           ├── configmap.yaml
│           ├── secret.yaml
│           └── rbac.yaml
├── docs/                            # Documentation
│   ├── architecture/                # Architecture documentation
│   ├── configuration/               # Configuration documentation
│   ├── development_guide.rst        # This file
│   ├── api_reference.rst            # API documentation
│   └── implementation_tasks.rst     # Implementation tasks
├── examples/                        # Example configurations
│   ├── destinations.yaml            # Example destinations config
│   ├── teams.yaml                   # Example teams config
│   ├── workflows.yaml               # Example workflows config
│   └── alertmanager.yaml            # Example Alertmanager config
├── Dockerfile                       # Container build file
├── go.mod                           # Go module file
├── go.sum                           # Go module checksums
├── Makefile                         # Build automation
├── .gitignore                       # Git ignore rules
├── README.md                        # Project README
└── VERSION                          # Version file
```

**Key New Components:**

1. **`pkg/core/prometheus/`** - PrometheusAlert and related structures
2. **`pkg/workflow/`** - Complete workflow processing system (built-in Go workflows + external TypeScript scripts)
3. **`pkg/alert/`** - Enhanced with converter, deduplication, queue, processor
4. **`pkg/destination/`** - Enhanced with router and matcher (static configuration)
5. **`pkg/sender/`** - Organized by destination type with common utilities
6. **`config/workflow/`** - Workflow configuration management (built-in workflows)
7. **`examples/`** - Example configurations for users

**Removed Components:**
- **`pkg/destination/registry.go`** - Destinations configured statically
- **`pkg/destination/factory.go`** - Senders created directly from destination config
- **`pkg/cache/`** - No long-term caching needed
- **`pkg/util/kubernetes/`** - Not needed for basic functionality
- **`scripts/`** - Build automation handled by Makefile

**File Naming Conventions:**
- `*_test.go` - Unit tests for each component
- `types.go` - Type definitions for each package
- `*_mock.go` - Generated mocks for testing
- `README.md` - Documentation for each major component

**Package Organization:**
- Each major feature has its own package
- Common utilities are shared across packages
- Clear separation of concerns
- Consistent naming and structure

Core Data Models
----------------

Issue Model
~~~~~~~~~~~

The `Issue` is the central data structure. See `pkg/core/issue/issue.go`:

.. code-block:: go

    type Issue struct {
        ID             uuid.UUID
        Title          string
        Description    string
        AggregationKey string
        Severity       Severity
        Status         Status
        Source         Source
        Subject        Subject
        Enrichments    []Enrichment
        Links          []Link
        Fingerprint    string
        StartsAt       time.Time
        EndsAt         *time.Time
    }

Key interfaces to implement:

.. code-block:: go

    // IssueProcessor processes alerts into Issues
    type IssueProcessor interface {
        ProcessAlert(alert *PrometheusAlert) (*Issue, error)
    }

    // IssueEnricher adds context to Issues
    type IssueEnricher interface {
        Enrich(ctx context.Context, issue *Issue) error
    }

    // IssueRouter routes Issues to destinations
    type IssueRouter interface {
        Route(issue *Issue) ([]Destination, error)
    }

PrometheusAlert Model
~~~~~~~~~~~~~~~~~~~~~

The `PrometheusAlert` structure for handling Alertmanager webhooks:

.. code-block:: go

    // pkg/core/prometheus/alert.go
    type PrometheusAlert struct {
        EndsAt       time.Time            `json:"endsAt"`
        GeneratorURL string               `json:"generatorURL"`
        StartsAt     time.Time            `json:"startsAt"`
        Fingerprint  string               `json:"fingerprint"`
        Status       string               `json:"status"`           // "firing" or "resolved"
        Labels       map[string]string    `json:"labels"`
        Annotations  map[string]string    `json:"annotations"`
    }

    type AlertManagerEvent struct {
        Alerts             []PrometheusAlert `json:"alerts"`
        ExternalURL        string            `json:"externalURL"`
        GroupKey           string            `json:"groupKey"`
        Version            string            `json:"version"`
        CommonAnnotations  map[string]string `json:"commonAnnotations,omitempty"`
        CommonLabels       map[string]string `json:"commonLabels,omitempty"`
        GroupLabels        map[string]string `json:"groupLabels,omitempty"`
        Receiver           string            `json:"receiver"`
        Status             string            `json:"status"`
    }

Workflow Models
~~~~~~~~~~~~~~

Workflow processing structures:

.. code-block:: go

    // pkg/workflow/workflow.go
    type Workflow struct {
        Name        string           `yaml:"name"`
        Description string           `yaml:"description"`
        Triggers    []WorkflowTrigger `yaml:"triggers"`
        Actions     []WorkflowAction  `yaml:"actions"`
        Enabled     bool             `yaml:"enabled"`
        Priority    int              `yaml:"priority"`
    }

    type WorkflowTrigger struct {
        AlertName    string            `yaml:"alertName,omitempty"`
        Namespace    string            `yaml:"namespace,omitempty"`
        Severity     string            `yaml:"severity,omitempty"`
        Labels       map[string]string `yaml:"labels,omitempty"`
        Annotations  map[string]string `yaml:"annotations,omitempty"`
        Priority     int               `yaml:"priority,omitempty"`
    }

    type WorkflowAction interface {
        Execute(ctx context.Context, alert *PrometheusAlert) (*issue.Enrichment, error)
        GetName() string
        GetType() string
    }

    // pkg/workflow/actions/base.go
    type BaseAction struct {
        Name string
        Type string
    }

    func (ba *BaseAction) GetName() string {
        return ba.Name
    }

    func (ba *BaseAction) GetType() string {
        return ba.Type
    }

Destination Models
~~~~~~~~~~~~~~~~~~

Destination management structures:

.. code-block:: go

    // pkg/destination/types.go
    type Destination struct {
        Name       string                 `yaml:"name"`
        Type       string                 `yaml:"type"`
        Config     map[string]interface{} `yaml:"config"`
        Enabled    bool                   `yaml:"enabled"`
        Priority   int                    `yaml:"priority"`
    }

    type Team struct {
        Name         string   `yaml:"name"`
        Destinations []string `yaml:"destinations"`
        Rules        []Rule   `yaml:"rules,omitempty"`
    }

    type Rule struct {
        Field    string `yaml:"field"`
        Operator string `yaml:"operator"` // "equals", "contains", "regex"
        Value    string `yaml:"value"`
    }

Configuration Models
~~~~~~~~~~~~~~~~~~~

Enhanced configuration structures:

.. code-block:: go

    // config/config.go
    type Config struct {
        AppName         string
        AppVersion      string
        AppEnv          string
        LogLevel        string
        TracingMode     string
        TracingEndpoint string
        SentryDSN       string
        SentryEnabled   bool
        
        // Alert processing configuration
        DeduplicationTTL  time.Duration `yaml:"deduplicationTTL"`
        QueueWorkers      int           `yaml:"queueWorkers"`
        QueueSize         int           `yaml:"queueSize"`
        MaxRetries        int           `yaml:"maxRetries"`
        
        // Relabeling configuration
        RelabelRules      []RelabelRule `yaml:"relabelRules"`
        
        // Workflow configuration
        WorkflowConfig    WorkflowConfig `yaml:"workflows"`
        
        // Destination configuration
        Destinations      destination.DestinationsConfig
        Teams             team.TeamsConfig
    }

    type RelabelRule struct {
        Source    string `yaml:"source"`
        Target    string `yaml:"target"`
        Operation string `yaml:"operation"` // "add" or "replace"
    }

    // config/workflow/workflow_config.go
    type WorkflowConfig struct {
        Workflows []Workflow `yaml:"workflows"`
        Defaults  struct {
            Enabled bool `yaml:"enabled"`
		} `yaml:"defaults"`
    }

Cache Models
~~~~~~~~~~~

Caching layer for performance and deduplication:

.. code-block:: go

    // pkg/cache/cache.go
    type Cache interface {
        Get(key string) (interface{}, bool)
        Set(key string, value interface{}, ttl time.Duration) error
        Delete(key string) error
        Clear() error
        Close() error
    }

    // pkg/cache/memory/memory_cache.go
    type MemoryCache struct {
        cache map[string]cacheItem
        mutex sync.RWMutex
    }

    type cacheItem struct {
        value      interface{}
        expiration time.Time
    }

    func (mc *MemoryCache) Get(key string) (interface{}, bool) {
        mc.mutex.RLock()
        defer mc.mutex.RUnlock()
        
        item, exists := mc.cache[key]
        if !exists {
            return nil, false
        }
        
        if time.Now().After(item.expiration) {
            delete(mc.cache, key)
            return nil, false
        }
        
        return item.value, true
    }

    func (mc *MemoryCache) Set(key string, value interface{}, ttl time.Duration) error {
        mc.mutex.Lock()
        defer mc.mutex.Unlock()
        
        mc.cache[key] = cacheItem{
            value:      value,
            expiration: time.Now().Add(ttl),
        }
        
        return nil
    }

Enrichment Blocks
~~~~~~~~~~~~~~~~~

Enrichment blocks are defined in `pkg/core/issue/blocks.go`. To add a new block type:

1. Define the block structure:

.. code-block:: go

    type CustomBlock struct {
        Data string
        Type string
    }

    func (c CustomBlock) IsBlock() {}

2. Implement rendering in senders:

.. code-block:: go

    func (s *SlackSender) renderCustomBlock(block CustomBlock) slack.Block {
        return slack.NewSectionBlock(
            slack.NewTextBlockObject("mrkdwn", block.Data, false, false),
            nil, nil,
        )
    }

Alert Processing Pipeline
-------------------------

Current Implementation
~~~~~~~~~~~~~~~~~~~~~

The alert processing pipeline is defined in `pkg/alert/alert_handler.go`:

.. code-block:: go

    func (h *AlertHandler) HandleAlert(c *gin.Context) {
        // 1. Parse alert from Alertmanager
        var alert template.Data
        if err := c.ShouldBindJSON(&alert); err != nil {
            // Handle error
        }

        // 2. Validate alert
        if alert.Receiver == "" || alert.Status == "" || len(alert.Alerts) == 0 {
            // Handle validation error
        }

        // 3. Convert to Issue and process
        // This needs to be implemented
    }

Required Implementation
~~~~~~~~~~~~~~~~~~~~~~

The following components need to be implemented:

1. **Alert to Issue Conversion:**

.. code-block:: go

    type AlertConverter struct {
        logger logger.LoggerInterface
    }

    func (ac *AlertConverter) ConvertAlert(alert template.Data) (*issue.Issue, error) {
        if len(alert.Alerts) == 0 {
            return nil, errors.New("no alerts in template data")
        }
        
        promAlert := alert.Alerts[0]
        
        issue := &issue.Issue{
            ID:             uuid.New(),
            Title:          extractTitle(promAlert),
            Description:    extractDescription(promAlert),
            AggregationKey: promAlert.Labels["alertname"],
            Severity:       mapSeverity(promAlert.Labels["severity"]),
            Status:         mapStatus(promAlert.Status),
            Source:         issue.SourcePrometheus,
            Subject:        extractSubject(promAlert),
            Fingerprint:    promAlert.Fingerprint,
            StartsAt:       promAlert.StartsAt,
            EndsAt:         &promAlert.EndsAt,
        }
        
        return issue, nil
    }

2. **Deduplication System:**

.. code-block:: go

    type DeduplicationCache struct {
        cache map[string]time.Time
        mutex sync.RWMutex
        ttl   time.Duration
    }

    func (dc *DeduplicationCache) IsDuplicate(alert *PrometheusAlert) bool {
        hash := dc.generateCompoundHash(alert)
        
        dc.mutex.Lock()
        defer dc.mutex.Unlock()
        
        if lastSeen, exists := dc.cache[hash]; exists {
            if time.Since(lastSeen) < dc.ttl {
                return true
            }
        }
        
        dc.cache[hash] = time.Now()
        return false
    }

3. **Async Processing Queue:**

.. code-block:: go

    type AlertQueue struct {
        queue    chan *AlertTask
        workers  int
        logger   logger.LoggerInterface
        metrics  metric.MetricsInterface
    }

Key Component Implementations
----------------------------

Workflow Processor
~~~~~~~~~~~~~~~~~~

Complete workflow processing implementation:

.. code-block:: go

    // pkg/workflow/processor.go
    type WorkflowProcessor struct {
        workflows []Workflow
        logger    logger.LoggerInterface
        metrics   metric.MetricsInterface
        tracer    tracer.TracerInterface
    }

    func (wp *WorkflowProcessor) ProcessAlert(ctx context.Context, alert *PrometheusAlert) (*issue.Issue, error) {
        span := wp.tracer.StartSpan("workflow.process_alert")
        defer span.End()
        
        // Select applicable workflows
        selectedWorkflows := wp.selectWorkflows(alert)
        if len(selectedWorkflows) == 0 {
            wp.logger.Debugf("No workflows selected for alert %s", alert.Fingerprint)
            return wp.createBasicIssue(alert), nil
        }
        
        // Execute workflows in priority order
        enrichments := []issue.Enrichment{}
        for _, workflow := range selectedWorkflows {
            workflowEnrichments, err := wp.executeWorkflow(ctx, workflow, alert)
            if err != nil {
                wp.logger.Errorf("Workflow %s failed: %v", workflow.Name, err)
                wp.metrics.IncrementCounter("workflow_execution_failures", map[string]string{
                    "workflow": workflow.Name,
                })
                continue
            }
            
            enrichments = append(enrichments, workflowEnrichments...)
            wp.metrics.IncrementCounter("workflow_execution_success", map[string]string{
                "workflow": workflow.Name,
            })
        }
        
        // Create enriched issue
        return wp.createEnrichedIssue(alert, enrichments), nil
    }

    func (wp *WorkflowProcessor) selectWorkflows(alert *PrometheusAlert) []Workflow {
        var selected []Workflow
        
        for _, workflow := range wp.workflows {
            if !workflow.Enabled {
                continue
            }
            
            for _, trigger := range workflow.Triggers {
                if wp.matchesTrigger(alert, trigger) {
                    selected = append(selected, workflow)
                    break
                }
            }
        }
        
        // Sort by priority (higher priority first)
        sort.Slice(selected, func(i, j int) bool {
            return selected[i].Priority > selected[j].Priority
        })
        
        return selected
    }

    func (wp *WorkflowProcessor) executeWorkflow(ctx context.Context, workflow Workflow, alert *PrometheusAlert) ([]issue.Enrichment, error) {
        var enrichments []issue.Enrichment
        
        for _, action := range workflow.Actions {
            enrichment, err := action.Execute(ctx, alert)
            if err != nil {
                wp.logger.Errorf("Workflow %s action %s failed: %v", workflow.Name, action.GetName(), err)
                continue
            }
            
            if enrichment != nil {
                enrichments = append(enrichments, *enrichment)
            }
        }
        
        return enrichments, nil
    }

Destination Router
~~~~~~~~~~~~~~~~~~

Intelligent destination routing based on team rules:

.. code-block:: go

    // pkg/destination/router.go
    type DestinationRouter struct {
        destinations []Destination
        teams        []Team
        logger       logger.LoggerInterface
        metrics      metric.MetricsInterface
    }

    func (dr *DestinationRouter) RouteIssue(issue *issue.Issue) ([]Destination, error) {
        // Find matching teams
        matchingTeams := dr.findMatchingTeams(issue)
        if len(matchingTeams) == 0 {
            dr.logger.Warnf("No teams match issue %s", issue.ID)
            return nil, fmt.Errorf("no matching teams for issue")
        }
        
        // Get destinations for matching teams
        var destinations []Destination
        for _, team := range matchingTeams {
            teamDestinations := dr.getTeamDestinations(team)
            destinations = append(destinations, teamDestinations...)
        }
        
        // Remove duplicates and sort by priority
        destinations = dr.deduplicateAndSort(destinations)
        
        dr.metrics.IncrementCounter("issue_routing_success", map[string]string{
            "teams_count":        strconv.Itoa(len(matchingTeams)),
            "destinations_count": strconv.Itoa(len(destinations)),
        })
        
        return destinations, nil
    }

    func (dr *DestinationRouter) getTeamDestinations(team Team) []Destination {
        var destinations []Destination
        
        for _, destName := range team.Destinations {
            for _, dest := range dr.destinations {
                if dest.Name == destName && dest.Enabled {
                    destinations = append(destinations, dest)
                }
            }
        }
        
        return destinations
    }

Sender Implementation
~~~~~~~~~~~~~~~~~~~~~

Direct sender creation from destination configuration:

.. code-block:: go

    // pkg/sender/sender.go
    type Sender interface {
        Send(ctx context.Context, issue *issue.Issue) error
        GetName() string
    }

    // pkg/sender/slack/slack_sender.go
    type SlackSender struct {
        webhookURL string
        channel    string
        username   string
        iconEmoji  string
        logger     logger.LoggerInterface
        tracer     tracer.TracerInterface
    }

    func NewSlackSender(destination Destination) (Sender, error) {
        webhookURL, ok := destination.Config["webhook_url"].(string)
        if !ok {
            return nil, fmt.Errorf("slack webhook_url not configured")
        }
        
        channel, _ := destination.Config["channel"].(string)
        username, _ := destination.Config["username"].(string)
        iconEmoji, _ := destination.Config["icon_emoji"].(string)
        
        return &SlackSender{
            webhookURL: webhookURL,
            channel:    channel,
            username:   username,
            iconEmoji:  iconEmoji,
            logger:     logger,
            tracer:     tracer,
        }, nil
    }

    func (s *SlackSender) Send(ctx context.Context, issue *issue.Issue) error {
        span := s.tracer.StartSpan("slack.send")
        defer span.End()
        
        blocks := s.buildBlocks(issue)
        
        payload := slack.WebhookMessage{
            Channel:   s.channel,
            Username:  s.username,
            IconEmoji: s.iconEmoji,
            Blocks:    blocks,
        }
        
        jsonPayload, err := json.Marshal(payload)
        if err != nil {
            return fmt.Errorf("failed to marshal slack payload: %w", err)
        }
        
        resp, err := http.Post(s.webhookURL, "application/json", bytes.NewBuffer(jsonPayload))
        if err != nil {
            return fmt.Errorf("failed to send to slack: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 400 {
            return fmt.Errorf("slack returned error: %d", resp.StatusCode)
        }
        
        s.logger.Infof("Successfully sent issue %s to Slack", issue.ID)
        return nil
    }

    // pkg/sender/opsgenie/opsgenie_sender.go
    type OpsGenieSender struct {
        apiKey string
        baseURL string
        logger  logger.LoggerInterface
        tracer  tracer.TracerInterface
    }

    func NewOpsGenieSender(destination Destination) (Sender, error) {
        apiKey, ok := destination.Config["api_key"].(string)
        if !ok {
            return nil, fmt.Errorf("opsgenie api_key not configured")
        }
        
        baseURL, _ := destination.Config["base_url"].(string)
        if baseURL == "" {
            baseURL = "https://api.opsgenie.com"
        }
        
        return &OpsGenieSender{
            apiKey:  apiKey,
            baseURL: baseURL,
            logger:  logger,
            tracer:  tracer,
        }, nil
    }

    func (o *OpsGenieSender) Send(ctx context.Context, issue *issue.Issue) error {
        span := o.tracer.StartSpan("opsgenie.send")
        defer span.End()
        
        alert := o.buildAlert(issue)
        
        jsonPayload, err := json.Marshal(alert)
        if err != nil {
            return fmt.Errorf("failed to marshal opsgenie payload: %w", err)
        }
        
        req, err := http.NewRequestWithContext(ctx, "POST", o.baseURL+"/v2/alerts", bytes.NewBuffer(jsonPayload))
        if err != nil {
            return fmt.Errorf("failed to create request: %w", err)
        }
        
        req.Header.Set("Authorization", "GenieKey "+o.apiKey)
        req.Header.Set("Content-Type", "application/json")
        
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return fmt.Errorf("failed to send to opsgenie: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 400 {
            return fmt.Errorf("opsgenie returned error: %d", resp.StatusCode)
        }
        
        o.logger.Infof("Successfully sent issue %s to OpsGenie", issue.ID)
        return nil
    }

    // pkg/sender/pagerduty/pagerduty_sender.go
    type PagerDutySender struct {
        apiKey string
        baseURL string
        logger  logger.LoggerInterface
        tracer  tracer.TracerInterface
    }

    func NewPagerDutySender(destination Destination) (Sender, error) {
        apiKey, ok := destination.Config["api_key"].(string)
        if !ok {
            return nil, fmt.Errorf("pagerduty api_key not configured")
        }
        
        baseURL, _ := destination.Config["base_url"].(string)
        if baseURL == "" {
            baseURL = "https://api.pagerduty.com"
        }
        
        return &PagerDutySender{
            apiKey:  apiKey,
            baseURL: baseURL,
            logger:  logger,
            tracer:  tracer,
        }, nil
    }

    func (p *PagerDutySender) Send(ctx context.Context, issue *issue.Issue) error {
        span := p.tracer.StartSpan("pagerduty.send")
        defer span.End()
        
        incident := p.buildIncident(issue)
        
        jsonPayload, err := json.Marshal(incident)
        if err != nil {
            return fmt.Errorf("failed to marshal pagerduty payload: %w", err)
        }
        
        req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/incidents", bytes.NewBuffer(jsonPayload))
        if err != nil {
            return fmt.Errorf("failed to create request: %w", err)
        }
        
        req.Header.Set("Authorization", "Token token="+p.apiKey)
        req.Header.Set("Content-Type", "application/json")
        req.Header.Set("Accept", "application/vnd.pagerduty+json;version=2")
        
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            return fmt.Errorf("failed to send to pagerduty: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 400 {
            return fmt.Errorf("pagerduty returned error: %d", resp.StatusCode)
        }
        
        p.logger.Infof("Successfully sent issue %s to PagerDuty", issue.ID)
        return nil
    }

Alert Processor
~~~~~~~~~~~~~~~

Main alert processing orchestrator:

.. code-block:: go

    // pkg/alert/processor.go
    type AlertProcessor struct {
        converter           *AlertConverter
        deduplicationCache  *DeduplicationCache
        workflowProcessor   *workflow.WorkflowProcessor
        destinationRouter   *destination.DestinationRouter
        logger              logger.LoggerInterface
        metrics             metric.MetricsInterface
        tracer              tracer.TracerInterface
    }

    func (ap *AlertProcessor) ProcessAlert(ctx context.Context, alert *PrometheusAlert) error {
        span := ap.tracer.StartSpan("alert.process")
        defer span.End()
        
        // Check deduplication
        if ap.deduplicationCache.IsDuplicate(alert) {
            ap.logger.Debugf("Alert %s is duplicate, skipping", alert.Fingerprint)
            ap.metrics.IncrementCounter("alert_duplicate", nil)
            return nil
        }
        
        // Convert to Issue
        issue, err := ap.converter.ConvertAlert(alert)
        if err != nil {
            ap.logger.Errorf("Failed to convert alert: %v", err)
            ap.metrics.IncrementCounter("alert_conversion_failure", nil)
            return err
        }
        
        // Process workflows
        enrichedIssue, err := ap.workflowProcessor.ProcessAlert(ctx, alert)
        if err != nil {
            ap.logger.Errorf("Failed to process workflows: %v", err)
            ap.metrics.IncrementCounter("workflow_processing_failure", nil)
            return err
        }
        
        // Route to destinations
        destinations, err := ap.destinationRouter.RouteIssue(enrichedIssue)
        if err != nil {
            ap.logger.Errorf("Failed to route issue: %v", err)
            ap.metrics.IncrementCounter("issue_routing_failure", nil)
            return err
        }
        
        // Send to destinations
        for _, dest := range destinations {
            go ap.sendToDestination(ctx, enrichedIssue, dest)
        }
        
        ap.metrics.IncrementCounter("alert_processed_success", map[string]string{
            "destinations_count": strconv.Itoa(len(destinations)),
        })
        
        return nil
    }

    func (ap *AlertProcessor) sendToDestination(ctx context.Context, issue *issue.Issue, destination Destination) {
        span := ap.tracer.StartSpan("alert.send_to_destination")
        defer span.End()
        
        // Create sender based on destination type
        var sender Sender
        var err error
        
        switch destination.Type {
        case "slack":
            sender, err = NewSlackSender(destination)
        case "msteams":
            sender, err = NewMSTeamsSender(destination)
        case "jira":
            sender, err = NewJiraSender(destination)
        case "servicenow":
            sender, err = NewServiceNowSender(destination)
        case "datadog":
            sender, err = NewDataDogSender(destination)
        case "kafka":
            sender, err = NewKafkaSender(destination)
        case "webhook":
            sender, err = NewWebhookSender(destination)
        case "opsgenie":
            sender, err = NewOpsGenieSender(destination)
        case "pagerduty":
            sender, err = NewPagerDutySender(destination)
        default:
            ap.logger.Errorf("Unsupported destination type: %s", destination.Type)
            ap.metrics.IncrementCounter("sender_creation_failure", map[string]string{
                "destination": destination.Name,
                "type":        destination.Type,
            })
            return
        }
        
        if err != nil {
            ap.logger.Errorf("Failed to create sender for destination %s: %v", destination.Name, err)
            ap.metrics.IncrementCounter("sender_creation_failure", map[string]string{
                "destination": destination.Name,
            })
            return
        }
        
        // Send issue
        err = sender.Send(ctx, issue)
        if err != nil {
            ap.logger.Errorf("Failed to send to destination %s: %v", destination.Name, err)
            ap.metrics.IncrementCounter("send_failure", map[string]string{
                "destination": destination.Name,
            })
            return
        }
        
        ap.logger.Infof("Successfully sent issue %s to destination %s", issue.ID, destination.Name)
        ap.metrics.IncrementCounter("send_success", map[string]string{
            "destination": destination.Name,
        })
    }

Testing Guidelines
------------------

Unit Testing
~~~~~~~~~~~~

1. **Use mocks for dependencies:**

.. code-block:: go

    func TestAlertHandler_HandleAlert(t *testing.T) {
        mockLogger := mocks.NewLoggerInterface(t)
        mockMetrics := mocks.NewMetricsInterface(t)
        
        handler := alert.NewAlertHandler(mockLogger, mockMetrics)
        
        // Test implementation
    }

2. **Test error conditions:**

.. code-block:: go

    func TestAlertConverter_ConvertAlert_EmptyAlerts(t *testing.T) {
        converter := &AlertConverter{}
        
        alert := template.Data{
            Alerts: []template.Alert{},
        }
        
        _, err := converter.ConvertAlert(alert)
        assert.Error(t, err)
        assert.Contains(t, err.Error(), "no alerts")
    }

Integration Testing
~~~~~~~~~~~~~~~~~~

1. **Test with real HTTP server:**

.. code-block:: go

    func TestSlackSender_Integration(t *testing.T) {
        server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Verify request
            assert.Equal(t, "POST", r.Method)
            assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
            w.WriteHeader(http.StatusOK)
        }))
        defer server.Close()
        
        sender := sender.NewSlackSender(server.URL, logger, http.DefaultClient)
        err := sender.Send(sender.Alert{Title: "Test", Message: "Test message"})
        assert.NoError(t, err)
    }

Development Workflow
--------------------

1. **Create feature branch:**

.. code-block:: bash

    git checkout -b feature/implement-workflow-processing

2. **Implement feature with tests:**

.. code-block:: go

    // Implement the feature
    // Add comprehensive tests
    // Update documentation

3. **Run tests:**

.. code-block:: bash

    go test ./...
    go vet ./...
    golangci-lint run

4. **Update documentation:**

.. code-block:: bash

    # Update relevant .rst files
    # Add examples
    # Update architecture diagrams

Code Quality Standards
---------------------

1. **Error Handling:**
   - Always check errors
   - Provide meaningful error messages
   - Use wrapped errors with context

2. **Logging:**
   - Use structured logging with zap
   - Include relevant context
   - Use appropriate log levels

3. **Metrics:**
   - Add metrics for all operations
   - Use consistent naming (`cano_*`)
   - Include labels for filtering

4. **Configuration:**
   - Validate configuration at startup
   - Provide sensible defaults
   - Document all configuration options

Performance Considerations
-------------------------

1. **Memory Management:**
   - Reuse objects where possible
   - Use object pools for frequently allocated objects
   - Monitor memory usage

2. **Concurrency:**
   - Use goroutines for I/O operations
   - Implement proper synchronization
   - Avoid blocking operations

3. **Network:**
   - Use connection pooling
   - Implement timeouts
   - Handle retries gracefully

Common Patterns
---------------

1. **Dependency Injection:**

.. code-block:: go

    type Service struct {
        logger   logger.LoggerInterface
        metrics  metric.MetricsInterface
        client   util.HTTPClient
    }

    func NewService(logger logger.LoggerInterface, metrics metric.MetricsInterface, client util.HTTPClient) *Service {
        return &Service{
            logger:  logger,
            metrics: metrics,
            client:  client,
        }
    }

2. **Interface Segregation:**

.. code-block:: go

    type AlertProcessor interface {
        ProcessAlert(alert *PrometheusAlert) (*Issue, error)
    }

    type AlertEnricher interface {
        Enrich(ctx context.Context, issue *Issue) error
    }

    type AlertRouter interface {
        Route(issue *Issue) ([]Destination, error)
    }

3. **Builder Pattern for Complex Objects:**

.. code-block:: go

    type IssueBuilder struct {
        issue *Issue
    }

    func NewIssueBuilder() *IssueBuilder {
        return &IssueBuilder{
            issue: &Issue{},
        }
    }

    func (b *IssueBuilder) WithTitle(title string) *IssueBuilder {
        b.issue.Title = title
        return b
    }

    func (b *IssueBuilder) Build() *Issue {
        return b.issue
    }

Development Practices
---------------------

Testing Guidelines
~~~~~~~~~~~~~~~~~~

1. **Unit Tests**: Every component should have comprehensive unit tests:

.. code-block:: go

    // pkg/alert/converter_test.go
    func TestAlertConverter_ConvertAlert(t *testing.T) {
        converter := &AlertConverter{
            logger: &mocks.LoggerMock{},
        }
        
        alert := template.Data{
            Alerts: []template.Alert{
                {
                    Labels: map[string]string{
                        "alertname": "TestAlert",
                        "severity":  "critical",
                        "namespace": "default",
                    },
                    Annotations: map[string]string{
                        "summary": "Test alert summary",
                        "description": "Test alert description",
                    },
                    Status: "firing",
                    Fingerprint: "test-fingerprint",
                    StartsAt: time.Now(),
                    EndsAt:   time.Now().Add(time.Hour),
                },
            },
        }
        
        issue, err := converter.ConvertAlert(alert)
        assert.NoError(t, err)
        assert.NotNil(t, issue)
        assert.Equal(t, "TestAlert", issue.AggregationKey)
        assert.Equal(t, issue.SeverityCritical, issue.Severity)
        assert.Equal(t, issue.StatusFiring, issue.Status)
    }

2. **Integration Tests**: Test component interactions:

.. code-block:: go

    // pkg/workflow/processor_integration_test.go
    func TestWorkflowProcessor_Integration(t *testing.T) {
        // Setup test dependencies
        logger := &mocks.LoggerMock{}
        metrics := &mocks.MetricsMock{}
        tracer := &mocks.TracerMock{}
        
        // Create test workflow
        workflow := Workflow{
            Name: "test-workflow",
            Triggers: []WorkflowTrigger{
                {
                    AlertName: "TestAlert",
                    Severity:  "critical",
                },
            },
            Actions: []WorkflowAction{
                &TestAction{},
            },
            Enabled:  true,
            Priority: 1,
        }
        
        processor := &WorkflowProcessor{
            workflows: []Workflow{workflow},
            logger:    logger,
            metrics:   metrics,
            tracer:    tracer,
        }
        
        // Create test alert
        alert := &PrometheusAlert{
            Labels: map[string]string{
                "alertname": "TestAlert",
                "severity":  "critical",
            },
            Status: "firing",
        }
        
        // Process alert
        issue, err := processor.ProcessAlert(context.Background(), alert)
        assert.NoError(t, err)
        assert.NotNil(t, issue)
        assert.Len(t, issue.Enrichments, 1)
    }

3. **Mock Generation**: Use mockery for interface mocking:

.. code-block:: bash

    # Generate mocks
    mockery --dir pkg/alert --name AlertProcessor --output mocks
    mockery --dir pkg/workflow --name WorkflowProcessor --output mocks
    mockery --dir pkg/destination --name DestinationRouter --output mocks

Configuration Examples
~~~~~~~~~~~~~~~~~~~~~

1. **Workflow Configuration:**

.. code-block:: yaml

    # config/workflows.yaml
    workflows:
      - name: "pod-crash-enrichment"
        description: "Enrich pod crash alerts with logs and events"
        enabled: true
        priority: 10
        triggers:
          - alertName: "PodCrashLooping"
            severity: "critical"
          - alertName: "PodRestarting"
            severity: "warning"
        actions:
          - name: "pod-logs"
            type: "pod_logs"
            config:
              container: "main"
              lines: 50
              since: "5m"
          - name: "pod-events"
            type: "pod_events"
            config:
              limit: 10
          - name: "resource-status"
            type: "resource_status"
            config:
              include_related: true

      - name: "node-issues-enrichment"
        description: "Enrich node-related alerts with metrics"
        enabled: true
        priority: 5
        triggers:
          - alertName: "NodeHighCPU"
            severity: "warning"
          - alertName: "NodeHighMemory"
            severity: "warning"
        actions:
          - name: "node-metrics"
            type: "node_metrics"
            config:
              metrics: ["cpu", "memory", "disk"]
              duration: "10m"

      - name: "custom-organization-enrichment"
        description: "Organization-specific enrichment using TypeScript script"
        enabled: true
        priority: 15
        triggers:
          - alertName: ".*"
            severity: "critical"
            namespace: "production"
        actions:
          - name: "custom-script"
            type: "custom_script"
            config:
              script_path: "/scripts/custom-enrichment.ts"
              timeout: "30s"
              env:
                API_ENDPOINT: "https://internal-api.company.com"
                API_KEY: "{{CUSTOM_API_KEY}}"

2. **Destination Configuration:**

.. code-block:: yaml

    # config/destinations.yaml
    destinations:
      - name: "slack-dev-team"
        type: "slack"
        enabled: true
        priority: 1
        config:
          webhook_url: "https://hooks.slack.com/services/..."
          channel: "#dev-alerts"
          username: "cano-collector"
          icon_emoji: ":warning:"
          
      - name: "slack-ops-team"
        type: "slack"
        enabled: true
        priority: 1
        config:
          webhook_url: "https://hooks.slack.com/services/..."
          channel: "#ops-alerts"
          username: "cano-collector"
          
      - name: "msteams-alerts"
        type: "msteams"
        enabled: true
        priority: 1
        config:
          webhook_url: "https://company.webhook.office.com/webhookb2/..."
          theme_color: "#FF0000"
          
      - name: "jira-incidents"
        type: "jira"
        enabled: true
        priority: 2
        config:
          url: "https://company.atlassian.net"
          username: "cano-collector"
          api_token: "{{JIRA_API_TOKEN}}"
          project_key: "OPS"
          issue_type: "Incident"
          
      - name: "servicenow-incidents"
        type: "servicenow"
        enabled: true
        priority: 2
        config:
          url: "https://company.service-now.com"
          username: "cano-collector"
          password: "{{SERVICENOW_PASSWORD}}"
          table: "incident"
          
      - name: "datadog-events"
        type: "datadog"
        enabled: true
        priority: 3
        config:
          api_key: "{{DATADOG_API_KEY}}"
          app_key: "{{DATADOG_APP_KEY}}"
          tags: ["env:production", "service:cano-collector"]
          
      - name: "kafka-alerts"
        type: "kafka"
        enabled: true
        priority: 3
        config:
          brokers: ["kafka-1:9092", "kafka-2:9092"]
          topic: "alerts"
          key_serializer: "string"
          value_serializer: "json"
          
      - name: "opsgenie-alerts"
        type: "opsgenie"
        enabled: true
        priority: 2
        config:
          api_key: "{{OPSGENIE_API_KEY}}"
          base_url: "https://api.opsgenie.com"
          team: "ops-team"
          priority_mapping:
            critical: "P1"
            warning: "P2"
            info: "P3"
            
      - name: "pagerduty-incidents"
        type: "pagerduty"
        enabled: true
        priority: 2
        config:
          api_key: "{{PAGERDUTY_API_KEY}}"
          base_url: "https://api.pagerduty.com"
          service_id: "{{PAGERDUTY_SERVICE_ID}}"
          escalation_policy_id: "{{PAGERDUTY_ESCALATION_POLICY_ID}}"
          
      - name: "webhook-generic"
        type: "webhook"
        enabled: true
        priority: 4
        config:
          url: "https://internal-api.company.com/alerts"
          method: "POST"
          headers:
            Authorization: "Bearer {{WEBHOOK_TOKEN}}"
            Content-Type: "application/json"
          timeout: "30s"

3. **Team Configuration:**

.. code-block:: yaml

    # config/teams.yaml
    teams:
      - name: "dev-team"
        destinations: ["slack-dev-team"]
        rules:
          - field: "namespace"
            operator: "equals"
            value: "development"
          - field: "severity"
            operator: "equals"
            value: "critical"
            
      - name: "ops-team"
        destinations: ["slack-ops-team", "jira-incidents", "opsgenie-alerts"]
        rules:
          - field: "namespace"
            operator: "equals"
            value: "production"
          - field: "severity"
            operator: "in"
            value: "critical,warning"
            
      - name: "oncall-team"
        destinations: ["pagerduty-incidents", "opsgenie-alerts"]
        rules:
          - field: "severity"
            operator: "equals"
            value: "critical"
          - field: "alertname"
            operator: "regex"
            value: ".*Down.*|.*Unavailable.*|.*Error.*"
            
      - name: "monitoring-team"
        destinations: ["datadog-events", "kafka-alerts"]
        rules:
          - field: "alertname"
            operator: "regex"
            value: ".*HighCPU.*|.*HighMemory.*|.*HighDisk.*"
            
      - name: "management-team"
        destinations: ["msteams-alerts", "webhook-generic"]
        rules:
          - field: "severity"
            operator: "equals"
            value: "critical"
          - field: "namespace"
            operator: "equals"
            value: "production"

Build and Deployment
~~~~~~~~~~~~~~~~~~~

1. **Makefile Targets:**

.. code-block:: makefile

    # Makefile
    .PHONY: build test lint clean docker-build docker-push

    build:
        go build -o bin/cano-collector main.go

    test:
        go test -v ./...

    test-coverage:
        go test -v -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html

    lint:
        golangci-lint run

    clean:
        rm -rf bin/
        rm -f coverage.out coverage.html

    docker-build:
        docker build -t cano-collector:latest .

    docker-push:
        docker tag cano-collector:latest registry.company.com/cano-collector:latest
        docker push registry.company.com/cano-collector:latest

    generate-mocks:
        mockery --all --output mocks

2. **Dockerfile:**

.. code-block:: dockerfile

    # Dockerfile
    FROM golang:1.21-alpine AS builder
    
    WORKDIR /app
    
    # Install dependencies
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Copy source code
    COPY . .
    
    # Build application
    RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o cano-collector main.go
    
    # Final stage
    FROM alpine:latest
    
    RUN apk --no-cache add ca-certificates
    
    WORKDIR /root/
    
    COPY --from=builder /app/cano-collector .
    COPY --from=builder /app/config ./config
    COPY --from=builder /app/examples ./examples
    
    EXPOSE 8080
    
    CMD ["./cano-collector"]

3. **Helm Values:**

.. code-block:: yaml

    # helm/cano-collector/values.yaml
    replicaCount: 2
    
    image:
      repository: cano-collector
      tag: latest
      pullPolicy: IfNotPresent
    
    service:
      type: ClusterIP
      port: 8080
    
    ingress:
      enabled: true
      className: nginx
      annotations:
        nginx.ingress.kubernetes.io/rewrite-target: /
      hosts:
        - host: cano-collector.company.com
          paths:
            - path: /
              pathType: Prefix
    
    config:
      logLevel: info
      tracingMode: jaeger
      tracingEndpoint: "http://jaeger:14268/api/traces"
      
      alertProcessing:
        deduplicationTTL: 5m
        queueWorkers: 10
        queueSize: 1000
        maxRetries: 3
        
      relabelRules:
        - source: "pod_name"
          target: "pod"
          operation: "add"
        - source: "deployment_name"
          target: "deployment"
          operation: "replace"
    
    resources:
      limits:
        cpu: 500m
        memory: 512Mi
      requests:
        cpu: 100m
        memory: 128Mi
    
    autoscaling:
      enabled: true
      minReplicas: 2
      maxReplicas: 10
      targetCPUUtilizationPercentage: 80

Performance Considerations
~~~~~~~~~~~~~~~~~~~~~~~~~

1. **Memory Management:**
   - Use object pools for frequently allocated structures
   - Implement proper cleanup in caches
   - Monitor memory usage with metrics

2. **Concurrency:**
   - Use worker pools for alert processing
   - Implement rate limiting for external API calls
   - Use context cancellation for timeouts

3. **Caching:**
   - Cache workflow selection results
   - Cache destination routing decisions
   - Use Redis for distributed caching

4. **Monitoring:**
   - Track processing latency
   - Monitor queue depths
   - Alert on processing failures

Security Considerations
~~~~~~~~~~~~~~~~~~~~~~

1. **Authentication:**
   - Implement API key authentication
   - Use mTLS for internal communication
   - Validate webhook signatures

2. **Authorization:**
   - Implement RBAC for team access
   - Validate destination permissions
   - Audit all configuration changes

3. **Data Protection:**
   - Encrypt sensitive configuration
   - Mask sensitive data in logs
   - Implement data retention policies

4. **Network Security:**
   - Use HTTPS for all external communication
   - Implement network policies
   - Monitor for suspicious activity

Custom Script Action
~~~~~~~~~~~~~~~~~~~~

Support for external TypeScript scripts:

.. code-block:: go

    // pkg/workflow/actions/custom_script.go
    type CustomScriptAction struct {
        BaseAction
        scriptPath string
        timeout    time.Duration
        env        map[string]string
        logger     logger.LoggerInterface
    }

    func NewCustomScriptAction(config map[string]interface{}) (*CustomScriptAction, error) {
        scriptPath, ok := config["script_path"].(string)
        if !ok {
            return nil, fmt.Errorf("script_path not configured")
        }
        
        timeoutStr, _ := config["timeout"].(string)
        timeout := 30 * time.Second
        if timeoutStr != "" {
            if parsed, err := time.ParseDuration(timeoutStr); err == nil {
                timeout = parsed
            }
        }
        
        env, _ := config["env"].(map[string]interface{})
        envMap := make(map[string]string)
        for k, v := range env {
            if str, ok := v.(string); ok {
                envMap[k] = str
            }
        }
        
        return &CustomScriptAction{
            BaseAction: BaseAction{
                Name: "custom-script",
                Type: "custom_script",
            },
            scriptPath: scriptPath,
            timeout:    timeout,
            env:        envMap,
            logger:     logger,
        }, nil
    }

    func (csa *CustomScriptAction) Execute(ctx context.Context, alert *PrometheusAlert) (*issue.Enrichment, error) {
        // Create temporary file with alert data
        alertData, err := json.Marshal(alert)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal alert data: %w", err)
        }
        
        tempFile, err := os.CreateTemp("", "alert-*.json")
        if err != nil {
            return nil, fmt.Errorf("failed to create temp file: %w", err)
        }
        defer os.Remove(tempFile.Name())
        
        if _, err := tempFile.Write(alertData); err != nil {
            return nil, fmt.Errorf("failed to write alert data: %w", err)
        }
        tempFile.Close()
        
        // Prepare environment variables
        env := os.Environ()
        for k, v := range csa.env {
            env = append(env, fmt.Sprintf("%s=%s", k, v))
        }
        env = append(env, fmt.Sprintf("ALERT_DATA_FILE=%s", tempFile.Name()))
        
        // Execute TypeScript script
        ctx, cancel := context.WithTimeout(ctx, csa.timeout)
        defer cancel()
        
        cmd := exec.CommandContext(ctx, "node", csa.scriptPath)
        cmd.Env = env
        
        output, err := cmd.CombinedOutput()
        if err != nil {
            csa.logger.Errorf("Custom script failed: %v, output: %s", err, string(output))
            return nil, fmt.Errorf("custom script execution failed: %w", err)
        }
        
        // Parse script output as enrichment
        var enrichment issue.Enrichment
        if err := json.Unmarshal(output, &enrichment); err != nil {
            csa.logger.Errorf("Failed to parse script output as enrichment: %v", err)
            return nil, fmt.Errorf("failed to parse script output: %w", err)
        }
        
        csa.logger.Infof("Custom script %s executed successfully", csa.scriptPath)
        return &enrichment, nil
    }

Example TypeScript Script
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: typescript

    // /scripts/custom-enrichment.ts
    import * as fs from 'fs';
    import * as https from 'https';

    interface AlertData {
        labels: Record<string, string>;
        annotations: Record<string, string>;
        status: string;
        startsAt: string;
        endsAt: string;
    }

    interface Enrichment {
        blocks: any[];
        annotations: Record<string, string>;
    }

    async function main() {
        try {
            // Read alert data from file
            const alertDataFile = process.env.ALERT_DATA_FILE;
            if (!alertDataFile) {
                throw new Error('ALERT_DATA_FILE environment variable not set');
            }
            
            const alertData: AlertData = JSON.parse(fs.readFileSync(alertDataFile, 'utf8'));
            
            // Custom enrichment logic
            const enrichment: Enrichment = {
                blocks: [],
                annotations: {}
            };
            
            // Example: Call internal API for additional context
            if (process.env.API_ENDPOINT && process.env.API_KEY) {
                const apiData = await callInternalAPI(alertData);
                enrichment.annotations['internal_context'] = JSON.stringify(apiData);
            }
            
            // Example: Add custom block based on alert type
            if (alertData.labels.alertname?.includes('Database')) {
                enrichment.blocks.push({
                    type: 'section',
                    text: {
                        type: 'mrkdwn',
                        text: `*Database Alert Detected*\nAlert: ${alertData.labels.alertname}\nStatus: ${alertData.status}`
                    }
                });
            }
            
            // Output enrichment as JSON
            console.log(JSON.stringify(enrichment));
            
        } catch (error) {
            console.error('Script execution failed:', error);
            process.exit(1);
        }
    }

    async function callInternalAPI(alertData: AlertData): Promise<any> {
        return new Promise((resolve, reject) => {
            const data = JSON.stringify({
                alert: alertData.labels.alertname,
                namespace: alertData.labels.namespace,
                severity: alertData.labels.severity
            });
            
            const options = {
                hostname: new URL(process.env.API_ENDPOINT!).hostname,
                port: 443,
                path: '/api/context',
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${process.env.API_KEY}`,
                    'Content-Length': data.length
                }
            };
            
            const req = https.request(options, (res) => {
                let responseData = '';
                res.on('data', (chunk) => {
                    responseData += chunk;
                });
                res.on('end', () => {
                    try {
                        resolve(JSON.parse(responseData));
                    } catch (error) {
                        reject(error);
                    }
                });
            });
            
            req.on('error', reject);
            req.write(data);
            req.end();
        });
    }

    main(); 