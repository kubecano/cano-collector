Workflow Processing and Issue Creation
=====================================

This document describes how cano-collector processes alerts through workflows, creates Issue objects with enrichment blocks, and the architectural considerations for workflow handling.

Workflow Selection and Processing
--------------------------------

Workflows in cano-collector are the equivalent of Robusta's playbooks, providing a flexible mechanism for alert enrichment and processing. The workflow selection process determines which workflows should be applied to a specific alert.

Workflow Selection Logic
~~~~~~~~~~~~~~~~~~~~~~~~

Workflows are selected based on matching criteria defined in the workflow configuration:

.. code-block:: go

    type WorkflowTrigger struct {
        AlertName    string            `yaml:"alertName,omitempty"`
        Namespace    string            `yaml:"namespace,omitempty"`
        Severity     string            `yaml:"severity,omitempty"`
        Labels       map[string]string `yaml:"labels,omitempty"`
        Annotations  map[string]string `yaml:"annotations,omitempty"`
        Priority     int               `yaml:"priority,omitempty"`
    }

    type Workflow struct {
        Name        string           `yaml:"name"`
        Description string           `yaml:"description"`
        Triggers    []WorkflowTrigger `yaml:"triggers"`
        Actions     []WorkflowAction  `yaml:"actions"`
        Enabled     bool             `yaml:"enabled"`
    }

    func selectWorkflows(alert *PrometheusAlert, workflows []Workflow) []Workflow {
        var selectedWorkflows []Workflow
        
        for _, workflow := range workflows {
            if !workflow.Enabled {
                continue
            }
            
            for _, trigger := range workflow.Triggers {
                if matchesTrigger(alert, trigger) {
                    selectedWorkflows = append(selectedWorkflows, workflow)
                    break
                }
            }
        }
        
        // Sort by priority (higher priority first)
        sort.Slice(selectedWorkflows, func(i, j int) bool {
            return getWorkflowPriority(selectedWorkflows[i]) > getWorkflowPriority(selectedWorkflows[j])
        })
        
        return selectedWorkflows
    }

    func matchesTrigger(alert *PrometheusAlert, trigger WorkflowTrigger) bool {
        if trigger.AlertName != "" && alert.Labels["alertname"] != trigger.AlertName {
            return false
        }
        
        if trigger.Namespace != "" && alert.Labels["namespace"] != trigger.Namespace {
            return false
        }
        
        if trigger.Severity != "" && alert.Labels["severity"] != trigger.Severity {
            return false
        }
        
        // Check label matches
        for key, value := range trigger.Labels {
            if alert.Labels[key] != value {
                return false
            }
        }
        
        // Check annotation matches
        for key, value := range trigger.Annotations {
            if alert.Annotations[key] != value {
                return false
            }
        }
        
        return true
    }

Comparison with Robusta's Playbook Selection:

Robusta uses a similar trigger-based system but with additional capabilities:
- More complex matching patterns (regex, wildcards)
- Time-based triggers
- Resource type matching
- Custom matcher functions

Workflow Execution
~~~~~~~~~~~~~~~~~~

Selected workflows are executed in priority order, with each workflow potentially enriching the alert:

.. code-block:: go

    type WorkflowAction interface {
        Execute(ctx context.Context, alert *PrometheusAlert) (*Enrichment, error)
        GetName() string
    }

    func executeWorkflows(ctx context.Context, alert *PrometheusAlert, workflows []Workflow) (*Issue, error) {
        enrichments := []Enrichment{}
        
        for _, workflow := range workflows {
            for _, action := range workflow.Actions {
                enrichment, err := action.Execute(ctx, alert)
                if err != nil {
                    logger.Errorf("Workflow %s action %s failed: %v", workflow.Name, action.GetName(), err)
                    continue
                }
                
                if enrichment != nil {
                    enrichments = append(enrichments, *enrichment)
                }
            }
        }
        
        return createIssueFromAlert(alert, enrichments)
    }

Issue Creation with Enrichment Blocks
-------------------------------------

The Issue object is the central data structure in cano-collector, equivalent to Robusta's Finding. It contains all the enriched context and metadata about the alert.

Issue Structure
~~~~~~~~~~~~~~~

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

    type Subject struct {
        Name        string
        SubjectType SubjectType
        Namespace   string
        Node        string
        Container   string
        Labels      map[string]string
        Annotations map[string]string
    }

    type Enrichment struct {
        Blocks []BaseBlock
        Annotations map[string]string
    }

Comparison with Robusta's Finding:

.. code-block:: python

    # Robusta's Finding structure
    class Finding(Filterable):
        title: str
        description: str
        aggregation_key: str
        severity: FindingSeverity
        status: FindingStatus
        source: FindingSource
        subject: FindingSubject
        enrichments: List[Enrichment]
        links: List[Link]
        fingerprint: str
        starts_at: datetime
        ends_at: Optional[datetime]

Both structures serve the same purpose but with language-specific implementations.

Issue Creation Process
~~~~~~~~~~~~~~~~~~~~~~

The Issue creation process transforms a PrometheusAlert into a rich Issue object:

.. code-block:: go

    func createIssueFromAlert(alert *PrometheusAlert, enrichments []Enrichment) (*Issue, error) {
        // Determine subject information from alert labels
        subject := extractSubjectFromAlert(alert)
        
        // Create basic issue
        issue := &Issue{
            ID:             uuid.New(),
            Title:          extractTitle(alert),
            Description:    extractDescription(alert),
            AggregationKey: alert.Labels["alertname"],
            Severity:       mapSeverity(alert.Labels["severity"]),
            Status:         mapStatus(alert.Status),
            Source:         SourcePrometheus,
            Subject:        subject,
            Enrichments:    enrichments,
            Links:          extractLinks(alert),
            Fingerprint:    alert.Fingerprint,
            StartsAt:       alert.StartsAt,
            EndsAt:         &alert.EndsAt,
        }
        
        return issue, nil
    }

    func extractSubjectFromAlert(alert *PrometheusAlert) Subject {
        subject := Subject{
            Labels:      alert.Labels,
            Annotations: alert.Annotations,
        }
        
        // Determine subject type and name from labels
        if pod, exists := alert.Labels["pod"]; exists {
            subject.SubjectType = SubjectTypePod
            subject.Name = pod
            subject.Namespace = alert.Labels["namespace"]
            subject.Container = alert.Labels["container"]
        } else if deployment, exists := alert.Labels["deployment"]; exists {
            subject.SubjectType = SubjectTypeDeployment
            subject.Name = deployment
            subject.Namespace = alert.Labels["namespace"]
        } else if node, exists := alert.Labels["node"]; exists {
            subject.SubjectType = SubjectTypeNode
            subject.Name = node
        }
        
        return subject
    }

Enrichment Blocks
~~~~~~~~~~~~~~~~~

Enrichment blocks provide structured content that can be rendered by different senders:

.. code-block:: go

    type BaseBlock interface {
        IsBlock()
    }

    type MarkdownBlock struct {
        Text string
    }

    type TableBlock struct {
        Rows    [][]string
        Headers []string
        Name    string
    }

    type FileBlock struct {
        Filename string
        Contents []byte
    }

    type ListBlock struct {
        Items []string
    }

    type HeaderBlock struct {
        Text string
    }

    type DividerBlock struct{}

    type LinksBlock struct {
        Links []Link
    }

Comparison with Robusta's BaseBlock:

.. code-block:: python

    # Robusta's BaseBlock structure
    class BaseBlock(BaseModel):
        hidden: bool = False
        html_class: str = None

    class MarkdownBlock(BaseBlock):
        text: str

    class TableBlock(BaseBlock):
        rows: List[List[str]]
        headers: List[str]
        name: str

Both implementations provide similar block types for rich content rendering.

Example workflow actions that create enrichment blocks:

.. code-block:: go

    type PodLogsAction struct {
        Container string `yaml:"container"`
        Lines     int    `yaml:"lines"`
    }

    func (a *PodLogsAction) Execute(ctx context.Context, alert *PrometheusAlert) (*Enrichment, error) {
        podName := alert.Labels["pod"]
        namespace := alert.Labels["namespace"]
        
        logs, err := getPodLogs(ctx, namespace, podName, a.Container, a.Lines)
        if err != nil {
            return nil, err
        }
        
        return &Enrichment{
            Blocks: []BaseBlock{
                MarkdownBlock{Text: fmt.Sprintf("**Pod Logs (%s):**\n```\n%s\n```", a.Container, logs)},
            },
        }, nil
    }

    type ResourceStatusAction struct{}

    func (a *ResourceStatusAction) Execute(ctx context.Context, alert *PrometheusAlert) (*Enrichment, error) {
        subject := extractSubjectFromAlert(alert)
        
        status, err := getResourceStatus(ctx, subject)
        if err != nil {
            return nil, err
        }
        
        return &Enrichment{
            Blocks: []BaseBlock{
                TableBlock{
                    Name:    "Resource Status",
                    Headers: []string{"Field", "Value"},
                    Rows:    status,
                },
            },
        }, nil
    }

Planned Alert Enrichment Features (TODO)
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Based on the routing_comparison.rst TODO items, the following enrichment features are planned:

1. **Pod Logs Enrichment**: Automatically fetch and include relevant pod logs
2. **Resource Status Enrichment**: Add current resource status and conditions
3. **Event History Enrichment**: Include recent Kubernetes events
4. **Metrics Enrichment**: Add relevant Prometheus metrics
5. **Configuration Analysis**: Validate and analyze resource configuration
6. **Recommendation Engine**: Provide actionable recommendations

WorkflowHandler vs AlertHandler
------------------------------

Currently, cano-collector uses `AlertHandler` for processing alerts, but there's a consideration to rename it to `WorkflowHandler` to better reflect its responsibilities.

Current AlertHandler Responsibilities
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

The current `AlertHandler` handles:

1. **Alert Reception**: Receives alerts from Alertmanager
2. **Basic Parsing**: Converts template.Data to internal format
3. **Simple Processing**: Basic alert handling without enrichment
4. **Metrics Recording**: Tracks alert processing metrics

Proposed WorkflowHandler Responsibilities
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

A `WorkflowHandler` would be responsible for:

1. **Workflow Selection**: Determine which workflows apply to the alert
2. **Workflow Execution**: Execute selected workflows in order
3. **Enrichment Management**: Collect and organize enrichment blocks
4. **Issue Creation**: Create the final Issue object
5. **Routing Coordination**: Coordinate with routing engine
6. **Error Handling**: Handle workflow execution failures

Benefits of WorkflowHandler
~~~~~~~~~~~~~~~~~~~~~~~~~~

- **Clearer Naming**: Better reflects the actual functionality
- **Separation of Concerns**: Distinguishes from simple alert handling
- **Extensibility**: Easier to add workflow-specific features
- **Consistency**: Aligns with workflow-centric architecture

Example WorkflowHandler Implementation
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    type WorkflowHandler struct {
        logger           logger.LoggerInterface
        metrics          metric.MetricsInterface
        workflowRegistry WorkflowRegistry
        deduplication    DeduplicationCache
        alertQueue       AlertQueue
    }

    func (wh *WorkflowHandler) HandleAlert(c *gin.Context) {
        // Parse alert from request
        alert, err := wh.parseAlert(c)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        // Check for duplicates
        if wh.deduplication.IsDuplicate(alert) {
            c.JSON(http.StatusOK, gin.H{"status": "duplicate"})
            return
        }
        
        // Apply relabeling
        alert = wh.applyRelabeling(alert)
        
        // Enqueue for processing
        wh.alertQueue.Enqueue(alert)
        
        c.JSON(http.StatusOK, gin.H{"status": "queued"})
    }

    func (wh *WorkflowHandler) processAlert(alert *PrometheusAlert) error {
        // Select applicable workflows
        workflows := wh.workflowRegistry.SelectWorkflows(alert)
        
        // Execute workflows
        issue, err := wh.executeWorkflows(context.Background(), alert, workflows)
        if err != nil {
            return err
        }
        
        // Route to destinations
        return wh.routeIssue(issue)
    }

Configuration Example
--------------------

Workflow configuration example:

.. code-block:: yaml

    workflows:
      - name: "pod-crashloop-enrichment"
        description: "Enrich pod crashloop alerts with logs and status"
        enabled: true
        triggers:
          - alertName: "PodCrashLooping"
            severity: "warning"
            priority: 10
        actions:
          - type: "pod_logs"
            container: "main"
            lines: 50
          - type: "resource_status"
            resource: "pod"
          - type: "pod_events"
            limit: 10

      - name: "node-pressure-enrichment"
        description: "Enrich node pressure alerts with resource usage"
        enabled: true
        triggers:
          - alertName: "NodeHighCpuLoad"
            severity: "warning"
            priority: 5
        actions:
          - type: "node_metrics"
            duration: "5m"
          - type: "resource_status"
            resource: "node"

This architecture provides:

- **Flexible Enrichment**: Customizable workflow actions
- **Priority-based Execution**: Important workflows run first
- **Rich Context**: Comprehensive Issue objects with multiple enrichment blocks
- **Extensible Design**: Easy to add new workflow actions
- **Clear Separation**: Distinct responsibilities for different components
