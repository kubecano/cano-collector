Development Guide
=================

This guide is for developers implementing and extending cano-collector. It covers the codebase structure, development practices, and how to implement new features.

Codebase Structure
------------------

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

        // 3. TODO: Convert to Issue and process
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

    func (aq *AlertQueue) Enqueue(alert *PrometheusAlert) {
        task := &AlertTask{
            Alert:     alert,
            Timestamp: time.Now(),
            Attempts:  0,
        }
        
        select {
        case aq.queue <- task:
            aq.metrics.IncAlertQueued()
        default:
            aq.metrics.IncAlertDropped()
        }
    }

4. **Workflow Processing:**

.. code-block:: go

    type WorkflowProcessor struct {
        workflows []Workflow
        logger    logger.LoggerInterface
    }

    func (wp *WorkflowProcessor) ProcessWorkflows(issue *issue.Issue) error {
        for _, workflow := range wp.workflows {
            if wp.matchesWorkflow(issue, workflow) {
                for _, action := range workflow.Actions {
                    enrichment, err := action.Execute(context.Background(), issue)
                    if err != nil {
                        wp.logger.Errorf("Workflow action failed: %v", err)
                        continue
                    }
                    
                    if enrichment != nil {
                        issue.Enrichments = append(issue.Enrichments, *enrichment)
                    }
                }
            }
        }
        return nil
    }

Destination System
------------------

Current Implementation
~~~~~~~~~~~~~~~~~~~~~

Destinations are configured in `config/destination/destinations_config.go`:

.. code-block:: go

    type DestinationsConfig struct {
        Destinations struct {
            Slack []Destination `yaml:"slack"`
            Teams []Destination `yaml:"teams"`
        } `yaml:"destinations"`
    }

Required Implementation
~~~~~~~~~~~~~~~~~~~~~~

1. **Destination Registry:**

.. code-block:: go

    type DestinationRegistry struct {
        destinations map[string]Destination
        mutex        sync.RWMutex
    }

    func (dr *DestinationRegistry) GetDestination(name string) (Destination, bool) {
        dr.mutex.RLock()
        defer dr.mutex.RUnlock()
        
        destination, exists := dr.destinations[name]
        return destination, exists
    }

2. **Destination Factory:**

.. code-block:: go

    type DestinationFactory struct {
        logger logger.LoggerInterface
        client util.HTTPClient
    }

    func (df *DestinationFactory) CreateDestination(config Destination) (sender.DestinationSender, error) {
        switch {
        case strings.HasPrefix(config.WebhookURL, "https://hooks.slack.com"):
            return sender.NewSlackSender(config.WebhookURL, df.logger, df.client), nil
        case strings.Contains(config.WebhookURL, "webhook.office.com"):
            return sender.NewMSTeamsSender(config.WebhookURL, df.logger, df.client), nil
        default:
            return nil, fmt.Errorf("unknown destination type")
        }
    }

Sender Implementation
---------------------

To implement a new sender, follow this pattern:

1. **Define the sender structure:**

.. code-block:: go

    type CustomSender struct {
        webhookURL string
        client     util.HTTPClient
        logger     logger.LoggerInterface
    }

    func NewCustomSender(webhookURL string, logger logger.LoggerInterface, client util.HTTPClient) *CustomSender {
        return &CustomSender{
            webhookURL: webhookURL,
            client:     client,
            logger:     logger,
        }
    }

2. **Implement the DestinationSender interface:**

.. code-block:: go

    func (s *CustomSender) Send(alert sender.Alert) error {
        payload := s.buildPayload(alert)
        
        resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewBuffer(payload))
        if err != nil {
            return fmt.Errorf("failed to send alert: %w", err)
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 400 {
            return fmt.Errorf("destination returned error: %d", resp.StatusCode)
        }
        
        return nil
    }

3. **Add to the factory:**

.. code-block:: go

    func (df *DestinationFactory) CreateDestination(config Destination) (sender.DestinationSender, error) {
        switch {
        case strings.HasPrefix(config.WebhookURL, "https://api.custom.com"):
            return sender.NewCustomSender(config.WebhookURL, df.logger, df.client), nil
        // ... other cases
        }
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