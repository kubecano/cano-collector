Implementation Tasks
====================

This document outlines the specific implementation tasks needed to complete cano-collector according to the architecture documentation.

Current State Analysis
----------------------

**Implemented Components:**
- ✅ Basic AlertHandler structure
- ✅ Issue and enrichment block models
- ✅ Basic sender interfaces (Slack, MS Teams)
- ✅ Configuration loading
- ✅ Logging and metrics infrastructure
- ✅ Health checks and routing

**Missing Components:**
- ❌ Alert to Issue conversion
- ❌ Deduplication system
- ❌ Async processing queue
- ❌ Workflow processing
- ❌ Destination routing
- ❌ Enrichment system
- ❌ Additional senders (Jira, ServiceNow, etc.)

Implementation Tasks
--------------------

Task 1: Alert to Issue Conversion
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** High
**Estimated Time:** 2-3 days

**Description:** Implement conversion from Alertmanager format to internal Issue model.

**Files to Create/Modify:**
- `pkg/alert/converter.go` - Alert to Issue conversion logic
- `pkg/alert/alert_handler.go` - Integrate conversion into handler

**Implementation:**

.. code-block:: go

    // pkg/alert/converter.go
    type AlertConverter struct {
        logger logger.LoggerInterface
    }

    func NewAlertConverter(logger logger.LoggerInterface) *AlertConverter {
        return &AlertConverter{logger: logger}
    }

    func (ac *AlertConverter) ConvertAlert(alert template.Data) (*issue.Issue, error) {
        if len(alert.Alerts) == 0 {
            return nil, errors.New("no alerts in template data")
        }
        
        promAlert := alert.Alerts[0]
        
        issue := &issue.Issue{
            ID:             uuid.New(),
            Title:          ac.extractTitle(promAlert),
            Description:    ac.extractDescription(promAlert),
            AggregationKey: promAlert.Labels["alertname"],
            Severity:       ac.mapSeverity(promAlert.Labels["severity"]),
            Status:         ac.mapStatus(promAlert.Status),
            Source:         issue.SourcePrometheus,
            Subject:        ac.extractSubject(promAlert),
            Fingerprint:    promAlert.Fingerprint,
            StartsAt:       promAlert.StartsAt,
            EndsAt:         &promAlert.EndsAt,
        }
        
        return issue, nil
    }

**Tests to Write:**
- Test conversion with valid alerts
- Test conversion with missing fields
- Test severity mapping
- Test subject extraction

Task 2: Deduplication System
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** High
**Estimated Time:** 1-2 days

**Description:** Implement deduplication to prevent processing the same alert multiple times.

**Files to Create:**
- `pkg/alert/deduplication.go` - Deduplication logic

**Implementation:**

.. code-block:: go

    // pkg/alert/deduplication.go
    type DeduplicationCache struct {
        cache map[string]time.Time
        mutex sync.RWMutex
        ttl   time.Duration
    }

    func NewDeduplicationCache(ttl time.Duration) *DeduplicationCache {
        return &DeduplicationCache{
            cache: make(map[string]time.Time),
            ttl:   ttl,
        }
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

    func (dc *DeduplicationCache) generateCompoundHash(alert *PrometheusAlert) string {
        data := []byte{
            alert.Fingerprint,
            alert.Status,
            strconv.FormatInt(alert.StartsAt.Unix(), 10),
            strconv.FormatInt(alert.EndsAt.Unix(), 10),
        }
        
        hash := sha256.Sum256(data)
        return hex.EncodeToString(hash[:])
    }

**Tests to Write:**
- Test deduplication with same alert
- Test deduplication with different alerts
- Test TTL expiration
- Test concurrent access

Task 3: Async Processing Queue
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** High
**Estimated Time:** 2-3 days

**Description:** Implement asynchronous processing queue for alerts.

**Files to Create:**
- `pkg/alert/queue.go` - Async queue implementation

**Implementation:**

.. code-block:: go

    // pkg/alert/queue.go
    type AlertQueue struct {
        queue    chan *AlertTask
        workers  int
        logger   logger.LoggerInterface
        metrics  metric.MetricsInterface
        processor AlertProcessor
    }

    type AlertTask struct {
        Alert     *PrometheusAlert
        Timestamp time.Time
        Attempts  int
    }

    func NewAlertQueue(workers int, logger logger.LoggerInterface, metrics metric.MetricsInterface, processor AlertProcessor) *AlertQueue {
        return &AlertQueue{
            queue:     make(chan *AlertTask, 1000),
            workers:   workers,
            logger:    logger,
            metrics:   metrics,
            processor: processor,
        }
    }

    func (aq *AlertQueue) Start() {
        for i := 0; i < aq.workers; i++ {
            go aq.worker()
        }
    }

    func (aq *AlertQueue) worker() {
        for task := range aq.queue {
            start := time.Now()
            
            if err := aq.processor.ProcessAlert(task.Alert); err != nil {
                aq.logger.Errorf("Failed to process alert: %v", err)
                aq.metrics.IncAlertProcessingErrors()
                
                if task.Attempts < maxRetries {
                    task.Attempts++
                    aq.queue <- task
                }
            } else {
                aq.metrics.ObserveAlertProcessingDuration(time.Since(start))
            }
        }
    }

**Tests to Write:**
- Test queue enqueue/dequeue
- Test worker processing
- Test retry logic
- Test queue overflow handling

Task 4: Workflow Processing
~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** Medium
**Estimated Time:** 3-4 days

**Description:** Implement workflow processing for alert enrichment.

**Files to Create:**
- `pkg/workflow/workflow.go` - Workflow definitions
- `pkg/workflow/processor.go` - Workflow processing logic
- `pkg/workflow/actions.go` - Workflow actions

**Implementation:**

.. code-block:: go

    // pkg/workflow/workflow.go
    type Workflow struct {
        Name        string           `yaml:"name"`
        Description string           `yaml:"description"`
        Triggers    []WorkflowTrigger `yaml:"triggers"`
        Actions     []WorkflowAction  `yaml:"actions"`
        Enabled     bool             `yaml:"enabled"`
    }

    type WorkflowTrigger struct {
        AlertName    string            `yaml:"alertName,omitempty"`
        Namespace    string            `yaml:"namespace,omitempty"`
        Severity     string            `yaml:"severity,omitempty"`
        Labels       map[string]string `yaml:"labels,omitempty"`
        Annotations  map[string]string `yaml:"annotations,omitempty"`
        Priority     int               `yaml:"priority,omitempty"`
    }

    // pkg/workflow/processor.go
    type WorkflowProcessor struct {
        workflows []Workflow
        logger    logger.LoggerInterface
    }

    func (wp *WorkflowProcessor) ProcessWorkflows(issue *issue.Issue) error {
        for _, workflow := range wp.workflows {
            if !workflow.Enabled {
                continue
            }
            
            for _, trigger := range workflow.Triggers {
                if wp.matchesTrigger(issue, trigger) {
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
                    break
                }
            }
        }
        return nil
    }

**Tests to Write:**
- Test workflow matching
- Test action execution
- Test enrichment creation
- Test error handling

Task 5: Destination Routing
~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** Medium
**Estimated Time:** 2-3 days

**Description:** Implement routing logic to determine which destinations receive alerts.

**Files to Create:**
- `pkg/destination/router.go` - Routing logic
- `pkg/destination/registry.go` - Destination registry

**Implementation:**

.. code-block:: go

    // pkg/destination/router.go
    type DestinationRouter struct {
        teams       []team.Team
        destinations map[string]Destination
        logger      logger.LoggerInterface
    }

    func (dr *DestinationRouter) Route(issue *issue.Issue) ([]Destination, error) {
        var selectedDestinations []Destination
        
        for _, team := range dr.teams {
            if dr.matchesTeam(issue, team) {
                for _, destName := range team.Destinations {
                    if dest, exists := dr.destinations[destName]; exists {
                        selectedDestinations = append(selectedDestinations, dest)
                    }
                }
            }
        }
        
        return selectedDestinations, nil
    }

    // pkg/destination/registry.go
    type DestinationRegistry struct {
        destinations map[string]Destination
        mutex        sync.RWMutex
    }

    func (dr *DestinationRegistry) RegisterDestination(name string, destination Destination) {
        dr.mutex.Lock()
        defer dr.mutex.Unlock()
        dr.destinations[name] = destination
    }

    func (dr *DestinationRegistry) GetDestination(name string) (Destination, bool) {
        dr.mutex.RLock()
        defer dr.mutex.RUnlock()
        destination, exists := dr.destinations[name]
        return destination, exists
    }

**Tests to Write:**
- Test team matching
- Test destination resolution
- Test registry operations
- Test routing logic

Task 6: Additional Senders
~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** Low
**Estimated Time:** 1-2 days per sender

**Description:** Implement additional destination senders.

**Senders to Implement:**
- Jira sender
- ServiceNow sender
- DataDog sender
- Kafka sender
- Webhook sender

**Implementation Pattern:**

.. code-block:: go

    // pkg/sender/jira_sender.go
    type JiraSender struct {
        url        string
        username   string
        apiToken   string
        projectKey string
        issueType  string
        client     util.HTTPClient
        logger     logger.LoggerInterface
    }

    func NewJiraSender(config JiraConfig, logger logger.LoggerInterface, client util.HTTPClient) *JiraSender {
        return &JiraSender{
            url:        config.URL,
            username:   config.Username,
            apiToken:   config.APIToken,
            projectKey: config.ProjectKey,
            issueType:  config.IssueType,
            client:     client,
            logger:     logger,
        }
    }

    func (s *JiraSender) Send(alert sender.Alert) error {
        payload := s.buildPayload(alert)
        
        req, err := http.NewRequest("POST", s.url+"/rest/api/2/issue", bytes.NewBuffer(payload))
        if err != nil {
            return err
        }
        
        req.SetBasicAuth(s.username, s.apiToken)
        req.Header.Set("Content-Type", "application/json")
        
        resp, err := s.client.Do(req)
        if err != nil {
            return err
        }
        defer resp.Body.Close()
        
        if resp.StatusCode >= 400 {
            return fmt.Errorf("Jira API error: %d", resp.StatusCode)
        }
        
        return nil
    }

**Tests to Write:**
- Test sender initialization
- Test payload building
- Test API communication
- Test error handling

Task 7: Configuration Enhancement
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** Medium
**Estimated Time:** 1-2 days

**Description:** Enhance configuration to support new features.

**Files to Modify:**
- `config/config.go` - Add new configuration options
- `config/destination/destinations_config.go` - Add new destination types

**Implementation:**

.. code-block:: go

    // config/config.go
    type Config struct {
        // ... existing fields
        WorkflowConfig    WorkflowConfig    `yaml:"workflows"`
        DeduplicationTTL  time.Duration     `yaml:"deduplicationTTL"`
        QueueWorkers      int               `yaml:"queueWorkers"`
        QueueSize         int               `yaml:"queueSize"`
    }

    // config/destination/destinations_config.go
    type DestinationsConfig struct {
        Destinations struct {
            Slack      []SlackDestination      `yaml:"slack"`
            MSTeams    []MSTeamsDestination    `yaml:"msteams"`
            Jira       []JiraDestination       `yaml:"jira"`
            ServiceNow []ServiceNowDestination `yaml:"servicenow"`
            DataDog    []DataDogDestination    `yaml:"datadog"`
            Kafka      []KafkaDestination      `yaml:"kafka"`
            Webhook    []WebhookDestination    `yaml:"webhook"`
        } `yaml:"destinations"`
    }

Task 8: Integration and Testing
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

**Priority:** High
**Estimated Time:** 2-3 days

**Description:** Integrate all components and add comprehensive testing.

**Tasks:**
1. Update `main.go` to wire all components
2. Add integration tests
3. Add end-to-end tests
4. Performance testing
5. Documentation updates

**Implementation:**

.. code-block:: go

    // main.go - Updated wiring
    func run(cfg config.Config, deps AppDependencies) error {
        log := deps.LoggerFactory(cfg.LogLevel, cfg.AppEnv)
        
        // Initialize components
        healthChecker := deps.HealthCheckerFactory(cfg, log)
        tracerManager := deps.TracerManagerFactory(cfg, log)
        metricsCollector := deps.MetricsFactory(log)
        
        // Initialize alert processing components
        alertConverter := alert.NewAlertConverter(log)
        deduplicationCache := alert.NewDeduplicationCache(cfg.DeduplicationTTL)
        alertProcessor := alert.NewAlertProcessor(log, metricsCollector)
        alertQueue := alert.NewAlertQueue(cfg.QueueWorkers, log, metricsCollector, alertProcessor)
        
        // Initialize workflow processing
        workflowProcessor := workflow.NewWorkflowProcessor(cfg.WorkflowConfig, log)
        
        // Initialize destination system
        destinationRegistry := destination.NewDestinationRegistry()
        destinationFactory := destination.NewDestinationFactory(log, util.NewHTTPClient())
        destinationRouter := destination.NewDestinationRouter(cfg.Teams, destinationRegistry, log)
        
        // Initialize alert handler
        alertHandler := alert.NewAlertHandler(
            log, 
            metricsCollector, 
            alertConverter, 
            deduplicationCache, 
            alertQueue,
            workflowProcessor,
            destinationRouter,
        )
        
        // Start processing
        alertQueue.Start()
        
        // Setup router
        routerManager := deps.RouterManagerFactory(cfg, log, tracerManager, metricsCollector, healthChecker, alertHandler)
        r := routerManager.SetupRouter()
        routerManager.StartServer(r)
        
        return nil
    }

Testing Strategy
----------------

Unit Tests
~~~~~~~~~~

- Test each component in isolation
- Use mocks for dependencies
- Test error conditions
- Test edge cases

Integration Tests
~~~~~~~~~~~~~~~~~

- Test component interactions
- Test with real HTTP servers
- Test configuration loading
- Test end-to-end flows

Performance Tests
~~~~~~~~~~~~~~~~~

- Test alert processing throughput
- Test memory usage
- Test concurrent processing
- Test queue performance

Acceptance Criteria
-------------------

**Functional Requirements:**
- ✅ Alerts are received from Alertmanager
- ✅ Alerts are converted to Issues
- ✅ Deduplication prevents duplicate processing
- ✅ Workflows enrich alerts with context
- ✅ Alerts are routed to correct destinations
- ✅ Notifications are sent successfully

**Non-Functional Requirements:**
- ✅ Processing latency < 1 second
- ✅ Support for 1000+ alerts per minute
- ✅ Memory usage < 512MB
- ✅ 99.9% uptime
- ✅ Comprehensive error handling
- ✅ Detailed metrics and logging

**Quality Requirements:**
- ✅ 90%+ test coverage
- ✅ No critical security vulnerabilities
- ✅ Comprehensive documentation
- ✅ Performance benchmarks
- ✅ Error handling for all failure modes

Implementation Timeline
-----------------------

**Week 1:**
- Task 1: Alert to Issue Conversion
- Task 2: Deduplication System

**Week 2:**
- Task 3: Async Processing Queue
- Task 4: Workflow Processing (part 1)

**Week 3:**
- Task 4: Workflow Processing (part 2)
- Task 5: Destination Routing

**Week 4:**
- Task 6: Additional Senders
- Task 7: Configuration Enhancement

**Week 5:**
- Task 8: Integration and Testing
- Documentation updates
- Performance optimization

**Total Estimated Time:** 5 weeks for complete implementation 