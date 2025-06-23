Event Enrichment and Finding Creation
=====================================

This document describes how cano-collector converts Kubernetes events into actionable findings and how these findings are enriched by workflows to provide comprehensive context for operators and automated systems.

Event to Finding Conversion
---------------------------

The conversion from raw Kubernetes events to actionable findings is a multi-stage process that transforms basic event data into rich, contextual information.

**1. Raw Event Processing:**
Raw Kubernetes events are received and parsed into structured data:

.. code-block:: go

    type KubernetesEvent struct {
        Operation   string                 // CREATE, UPDATE, DELETE
        Kind        string                 // Pod, Deployment, Service, etc.
        Name        string                 // Resource name
        Namespace   string                 // Resource namespace
        Object      runtime.Object         // Full resource object
        OldObject   runtime.Object         // Previous state (for updates)
        Timestamp   time.Time              // Event timestamp
        Reason      string                 // Event reason
        Message     string                 // Event message
        Severity    Severity               // Calculated severity
        Impact      ImpactAssessment       // Impact assessment
    }

    type ImpactAssessment struct {
        AffectedResources int               // Number of affected resources
        Scope             string            // Cluster, namespace, or resource
        Criticality       string            // Critical, important, or minor
        BusinessImpact    string            // High, medium, or low
    }

**2. Issue Creation Process:**
The Kubernetes event is transformed into an Issue object with enriched context:

.. code-block:: go

    type IssueConverter struct {
        logger     *log.Logger
        config     *Config
    }

    func (c *IssueConverter) ConvertToIssue(event *KubernetesEvent, context *EventContext) (*Issue, error) {
        // Generate unique ID
        id := uuid.New()
        
        // Create title and description
        title := c.generateTitle(event)
        description := c.generateDescription(event)
        
        // Generate aggregation key for grouping similar issues
        aggregationKey := c.generateAggregationKey(event)
        
        // Determine severity
        severity := c.determineSeverity(event)
        
        // Create subject information
        subject := c.createSubject(event)
        
        // Create initial enrichments
        enrichments := c.createInitialEnrichments(event, context)
        
        // Create links
        links := c.createLinks(event)
        
        // Generate fingerprint for deduplication
        fingerprint := c.generateFingerprint(event)
        
        issue := &Issue{
            ID:             id,
            Title:          title,
            Description:    description,
            AggregationKey: aggregationKey,
            Severity:       severity,
            Status:         Status.FIRING,
            Source:         Source.KUBERNETES_API_SERVER,
            Subject:        subject,
            Enrichments:    enrichments,
            Links:          links,
            Fingerprint:    fingerprint,
            StartsAt:       event.Timestamp,
            EndsAt:         nil, // Will be set when issue is resolved
        }
        
        return issue, nil
    }

    func (c *IssueConverter) generateTitle(event *KubernetesEvent) string {
        switch event.Kind {
        case "Pod":
            return c.generatePodTitle(event)
        case "Deployment":
            return c.generateDeploymentTitle(event)
        case "Service":
            return c.generateServiceTitle(event)
        default:
            return fmt.Sprintf("%s %s: %s", event.Operation, event.Kind, event.Name)
        }
    }

    func (c *IssueConverter) generatePodTitle(event *KubernetesEvent) string {
        if pod, ok := event.Object.(*corev1.Pod); ok {
            // Check for specific pod issues
            for _, container := range pod.Status.ContainerStatuses {
                if container.State.Waiting != nil {
                    switch container.State.Waiting.Reason {
                    case "CrashLoopBackOff":
                        return fmt.Sprintf("Pod %s in CrashLoopBackOff", pod.Name)
                    case "ImagePullBackOff":
                        return fmt.Sprintf("Pod %s: Image Pull Failed", pod.Name)
                    }
                }
            }
            
            switch pod.Status.Phase {
            case corev1.PodFailed:
                return fmt.Sprintf("Pod %s Failed", pod.Name)
            case corev1.PodPending:
                return fmt.Sprintf("Pod %s Pending", pod.Name)
            }
        }
        
        return fmt.Sprintf("Pod %s: %s", event.Name, event.Operation)
    }

    func (c *IssueConverter) generateDescription(event *KubernetesEvent) string {
        var description strings.Builder
        
        description.WriteString(fmt.Sprintf("**Resource:** %s/%s\n", event.Namespace, event.Name))
        description.WriteString(fmt.Sprintf("**Operation:** %s\n", event.Operation))
        description.WriteString(fmt.Sprintf("**Timestamp:** %s\n", event.Timestamp.Format(time.RFC3339)))
        
        if event.Message != "" {
            description.WriteString(fmt.Sprintf("**Message:** %s\n", event.Message))
        }
        
        if event.Reason != "" {
            description.WriteString(fmt.Sprintf("**Reason:** %s\n", event.Reason))
        }
        
        // Add impact assessment
        if event.Impact.AffectedResources > 0 {
            description.WriteString(fmt.Sprintf("**Impact:** %s (%d resources affected)\n", 
                event.Impact.Criticality, event.Impact.AffectedResources))
        }
        
        return description.String()
    }

    func (c *IssueConverter) generateAggregationKey(event *KubernetesEvent) string {
        // Create a key that groups similar events together
        return fmt.Sprintf("%s:%s:%s:%s", event.Kind, event.Namespace, event.Operation, event.Reason)
    }

    func (c *IssueConverter) determineSeverity(event *KubernetesEvent) Severity {
        // Base severity on operation type
        baseSeverity := map[string]Severity{
            "CREATE": Severity.INFO,
            "UPDATE": Severity.WARNING,
            "DELETE": Severity.HIGH,
        }[event.Operation]
        
        // Adjust based on resource type and specific conditions
        switch event.Kind {
        case "Pod":
            return c.calculatePodSeverity(event, baseSeverity)
        case "Deployment":
            return c.calculateDeploymentSeverity(event, baseSeverity)
        case "Service":
            return c.calculateServiceSeverity(event, baseSeverity)
        case "Node":
            return Severity.HIGH // Nodes are always critical
        case "PersistentVolume":
            return Severity.HIGH // Storage issues are critical
        default:
            return baseSeverity
        }
    }

    func (c *IssueConverter) calculatePodSeverity(event *KubernetesEvent, baseSeverity Severity) Severity {
        if pod, ok := event.Object.(*corev1.Pod); ok {
            // Check for critical container states
            for _, container := range pod.Status.ContainerStatuses {
                if container.State.Waiting != nil {
                    switch container.State.Waiting.Reason {
                    case "CrashLoopBackOff":
                        return Severity.HIGH
                    case "ImagePullBackOff":
                        return Severity.WARNING
                    case "ContainerCreating":
                        return Severity.INFO
                    }
                }
                
                if container.State.Terminated != nil {
                    if container.State.Terminated.ExitCode != 0 {
                        return Severity.HIGH
                    }
                }
            }
            
            // Check pod phase
            switch pod.Status.Phase {
            case corev1.PodFailed:
                return Severity.HIGH
            case corev1.PodPending:
                return Severity.WARNING
            case corev1.PodRunning:
                return Severity.INFO
            }
        }
        
        return baseSeverity
    }

    func (c *IssueConverter) createSubject(event *KubernetesEvent) Subject {
        return Subject{
            Name:      event.Name,
            Namespace: event.Namespace,
            Kind:      event.Kind,
            Labels:    c.extractLabels(event.Object),
            Node:      c.extractNodeName(event.Object),
        }
    }

    func (c *IssueConverter) createInitialEnrichments(event *KubernetesEvent, context *EventContext) []Enrichment {
        var enrichments []Enrichment
        
        // Add resource metadata
        metadataBlock := c.createMetadataBlock(event)
        enrichments = append(enrichments, Enrichment{
            Blocks: []BaseBlock{metadataBlock},
            Annotations: map[string]string{
                "type": "metadata",
            },
        })
        
        // Add status information
        if statusBlock := c.createStatusBlock(event); statusBlock != nil {
            enrichments = append(enrichments, Enrichment{
                Blocks: []BaseBlock{statusBlock},
                Annotations: map[string]string{
                    "type": "status",
                },
            })
        }
        
        // Add related resources if available
        if context != nil && len(context.Related) > 0 {
            relatedBlock := c.createRelatedResourcesBlock(context.Related)
            enrichments = append(enrichments, Enrichment{
                Blocks: []BaseBlock{relatedBlock},
                Annotations: map[string]string{
                    "type": "related_resources",
                },
            })
        }
        
        return enrichments
    }

    func (c *IssueConverter) createMetadataBlock(event *KubernetesEvent) BaseBlock {
        metadata := map[string]interface{}{
            "kind":      event.Kind,
            "name":      event.Name,
            "namespace": event.Namespace,
            "operation": event.Operation,
            "timestamp": event.Timestamp,
            "reason":    event.Reason,
            "message":   event.Message,
        }
        
        // Add labels and annotations
        if labels := c.extractLabels(event.Object); len(labels) > 0 {
            metadata["labels"] = labels
        }
        
        if annotations := c.extractAnnotations(event.Object); len(annotations) > 0 {
            metadata["annotations"] = annotations
        }
        
        jsonData, _ := json.MarshalIndent(metadata, "", "  ")
        return &JsonBlock{JsonStr: string(jsonData)}
    }

    func (c *IssueConverter) createStatusBlock(event *KubernetesEvent) BaseBlock {
        switch event.Kind {
        case "Pod":
            return c.createPodStatusBlock(event)
        case "Deployment":
            return c.createDeploymentStatusBlock(event)
        case "Service":
            return c.createServiceStatusBlock(event)
        default:
            return nil
        }
    }

    func (c *IssueConverter) createPodStatusBlock(event *KubernetesEvent) BaseBlock {
        if pod, ok := event.Object.(*corev1.Pod); ok {
            var rows [][]string
            rows = append(rows, []string{"Phase", string(pod.Status.Phase)})
            rows = append(rows, []string{"Node", pod.Spec.NodeName})
            rows = append(rows, []string{"QoS Class", string(pod.Status.QOSClass)})
            
            // Add container statuses
            for _, container := range pod.Status.ContainerStatuses {
                status := "Running"
                if container.State.Waiting != nil {
                    status = fmt.Sprintf("Waiting (%s)", container.State.Waiting.Reason)
                } else if container.State.Terminated != nil {
                    status = fmt.Sprintf("Terminated (exit: %d)", container.State.Terminated.ExitCode)
                }
                
                rows = append(rows, []string{container.Name, status})
            }
            
            return &TableBlock{
                Headers: []string{"Field", "Value"},
                Rows:    rows,
                Title:   "Pod Status",
            }
        }
        
        return nil
    }

Workflow Enrichment Process
---------------------------

Once an Issue is created from a Kubernetes event, it enters the workflow enrichment pipeline where it can be enhanced with additional context and actions.

**What Makes an Event "Magically" Enrichable:**

The Issue object contains several key properties that make it automatically processable by workflows:

**1. Rich Resource Context:**
- **Complete Resource Object**: Full Kubernetes resource with all fields and metadata
- **Resource Relationships**: Owner references, labels, annotations, and dependencies
- **Status Information**: Current state, conditions, health checks, and readiness
- **Metadata**: Namespace, creation time, resource version, and UID

**2. Event-Specific Information:**
- **Operation Details**: What changed, how it changed, and when
- **Change History**: Previous and current states for updates and rollbacks
- **Event Metadata**: Reason, message, timestamp, and event source
- **Impact Assessment**: Scope, severity, and business impact of the change

**3. Cluster Context:**
- **Node Information**: Which node hosts the resource and node status
- **Namespace Context**: Namespace-level information and policies
- **Cluster Metadata**: Cluster name, version, configuration, and capabilities
- **Resource Quotas**: Available resources, limits, and usage patterns

**4. Extensible Enrichment Framework:**
- **Block System**: Structured data blocks for different content types
- **Annotation Support**: Custom metadata for workflow processing and routing
- **Link Management**: Related URLs, references, and external resources
- **Fingerprinting**: Unique identification for deduplication and correlation

**Workflow Processing Pipeline:**

.. code-block:: go

    type WorkflowProcessor struct {
        workflows []Workflow
        logger    *log.Logger
    }

    func (p *WorkflowProcessor) ProcessIssue(issue *Issue) error {
        // Find applicable workflows
        applicableWorkflows := p.findApplicableWorkflows(issue)
        
        for _, workflow := range applicableWorkflows {
            // Execute workflow
            if err := p.executeWorkflow(workflow, issue); err != nil {
                p.logger.Printf("Workflow %s failed: %v", workflow.Name, err)
                continue
            }
            
            // Check if workflow should stop processing
            if workflow.StopOnSuccess {
                break
            }
        }
        
        return nil
    }

    func (p *WorkflowProcessor) findApplicableWorkflows(issue *Issue) []Workflow {
        var applicable []Workflow
        
        for _, workflow := range p.workflows {
            if p.isWorkflowApplicable(workflow, issue) {
                applicable = append(applicable, workflow)
            }
        }
        
        // Sort by priority
        sort.Slice(applicable, func(i, j int) bool {
            return applicable[i].Priority > applicable[j].Priority
        })
        
        return applicable
    }

    func (p *WorkflowProcessor) isWorkflowApplicable(workflow Workflow, issue *Issue) bool {
        // Check resource type
        if len(workflow.ResourceTypes) > 0 {
            if !contains(workflow.ResourceTypes, issue.Subject.Kind) {
                return false
            }
        }
        
        // Check namespace
        if len(workflow.Namespaces) > 0 {
            if !contains(workflow.Namespaces, issue.Subject.Namespace) {
                return false
            }
        }
        
        // Check severity
        if len(workflow.Severities) > 0 {
            if !contains(workflow.Severities, issue.Severity) {
                return false
            }
        }
        
        // Check labels
        if len(workflow.LabelSelectors) > 0 {
            if !p.matchesLabelSelectors(workflow.LabelSelectors, issue.Subject.Labels) {
                return false
            }
        }
        
        return true
    }

    func (p *WorkflowProcessor) executeWorkflow(workflow Workflow, issue *Issue) error {
        p.logger.Printf("Executing workflow %s for issue %s", workflow.Name, issue.ID)
        
        // Create workflow context
        context := &WorkflowContext{
            Issue:    issue,
            Workflow: workflow,
        }
        
        // Execute workflow steps
        for _, step := range workflow.Steps {
            if err := p.executeStep(step, context); err != nil {
                return fmt.Errorf("workflow step %s failed: %v", step.Name, err)
            }
        }
        
        return nil
    }

**Example Workflow Enrichment:**

When a Pod enters a CrashLoopBackOff state, the workflow might:

.. code-block:: go

    func (p *WorkflowProcessor) executeCrashLoopBackOffWorkflow(issue *Issue) error {
        pod := issue.Subject.Object.(*corev1.Pod)
        
        // Step 1: Fetch Pod Logs
        logs, err := p.fetchPodLogs(pod)
        if err == nil {
            logBlock := &FileBlock{
                Filename: fmt.Sprintf("%s-logs.txt", pod.Name),
                Contents: []byte(logs),
            }
            issue.AddEnrichment([]BaseBlock{logBlock}, map[string]string{
                "type": "logs",
                "container": "all",
            })
        }
        
        // Step 2: Check Resource Usage
        metrics, err := p.fetchResourceMetrics(pod)
        if err == nil {
            metricsBlock := p.createMetricsBlock(metrics)
            issue.AddEnrichment([]BaseBlock{metricsBlock}, map[string]string{
                "type": "metrics",
            })
        }
        
        // Step 3: Examine Related Events
        events, err := p.fetchRelatedEvents(pod)
        if err == nil {
            eventsBlock := p.createEventsBlock(events)
            issue.AddEnrichment([]BaseBlock{eventsBlock}, map[string]string{
                "type": "events",
            })
        }
        
        // Step 4: Validate Configuration
        configIssues := p.validatePodConfiguration(pod)
        if len(configIssues) > 0 {
            configBlock := p.createConfigurationBlock(configIssues)
            issue.AddEnrichment([]BaseBlock{configBlock}, map[string]string{
                "type": "configuration",
            })
        }
        
        // Step 5: Generate Recommendations
        recommendations := p.generateRecommendations(pod, logs, metrics, events)
        if len(recommendations) > 0 {
            recBlock := p.createRecommendationsBlock(recommendations)
            issue.AddEnrichment([]BaseBlock{recBlock}, map[string]string{
                "type": "recommendations",
            })
        }
        
        return nil
    }

**Enrichment Types:**

Workflows can add various types of enrichments:

**1. Log Blocks:**
- **Container Logs**: Recent application and system logs
- **System Logs**: Node-level system logs and events
- **Audit Logs**: Kubernetes audit trail information

**2. Metric Blocks:**
- **Performance Metrics**: CPU, memory, disk, and network usage
- **Application Metrics**: Custom application metrics and KPIs
- **Trend Analysis**: Historical performance trends and patterns

**3. Table Blocks:**
- **Status Tables**: Current status of related resources
- **Configuration Tables**: Configuration parameters and settings
- **Comparison Tables**: Before/after comparisons for updates

**4. File Blocks:**
- **Diagnostic Files**: System diagnostics and health checks
- **Configuration Files**: Resource configuration dumps
- **Screenshot Files**: Visual representations of issues

**5. Link Blocks:**
- **Related URLs**: Links to monitoring dashboards, logs, and tools
- **Documentation Links**: Relevant documentation and runbooks
- **Action Links**: Direct links to remediation actions

**6. Markdown Blocks:**
- **Explanations**: Detailed explanations of issues and causes
- **Instructions**: Step-by-step remediation instructions
- **Context**: Additional context and background information

**Workflow Configuration:**

Workflows are configured to trigger based on:

.. code-block:: yaml

    workflows:
      - name: "pod-crashloopbackoff"
        description: "Enrich Pod CrashLoopBackOff issues"
        resourceTypes: ["Pod"]
        severities: ["HIGH"]
        priority: 100
        stopOnSuccess: false
        steps:
          - name: "fetch-logs"
            action: "fetchPodLogs"
            parameters:
              maxLines: 100
              containers: ["all"]
          - name: "check-metrics"
            action: "fetchResourceMetrics"
            parameters:
              duration: "5m"
              metrics: ["cpu", "memory", "disk"]
          - name: "analyze-events"
            action: "fetchRelatedEvents"
            parameters:
              maxEvents: 20
              timeWindow: "1h"
          - name: "generate-recommendations"
            action: "generateRecommendations"
            parameters:
              includeFixes: true
              includePrevention: true
      
      - name: "deployment-rollout-failed"
        description: "Enrich failed deployment rollouts"
        resourceTypes: ["Deployment"]
        severities: ["HIGH"]
        priority: 90
        stopOnSuccess: false
        steps:
          - name: "check-replicasets"
            action: "fetchRelatedReplicaSets"
          - name: "analyze-pods"
            action: "analyzePodIssues"
          - name: "check-resources"
            action: "checkResourceAvailability"
          - name: "generate-rollback-plan"
            action: "generateRollbackPlan"

**Automatic Enrichment Triggers:**

Certain enrichments are automatically triggered based on event characteristics:

**1. Resource-Specific Enrichments:**
- **Pods**: Logs, metrics, events, configuration validation
- **Deployments**: ReplicaSet analysis, rollout status, scaling history
- **Services**: Endpoint analysis, load balancer status, connectivity tests
- **Nodes**: Resource usage, capacity analysis, scheduling issues

**2. Severity-Based Enrichments:**
- **HIGH**: Comprehensive analysis with all available data
- **WARNING**: Standard analysis with key metrics and logs
- **INFO**: Basic analysis with essential information

**3. Operation-Based Enrichments:**
- **CREATE**: Resource validation and initial health checks
- **UPDATE**: Change analysis and impact assessment
- **DELETE**: Cleanup verification and dependency analysis

This enrichment process transforms basic Kubernetes events into rich, actionable findings that provide comprehensive context for operators and automated systems. The modular workflow system allows for flexible and extensible enrichment capabilities that can be tailored to specific environments and requirements. 