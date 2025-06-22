Application Tracing
==================

This document describes the tracing architecture in cano-collector, including current implementation and planned enhancements for comprehensive distributed tracing.

Current Tracing Implementation
-----------------------------

Cano-collector currently implements basic OpenTelemetry tracing with the following components:

Tracing Configuration
~~~~~~~~~~~~~~~~~~~~~

The tracing system is configured through environment variables:

.. code-block:: yaml

    tracing:
      mode: "disabled" | "local" | "remote"
      endpoint: "http://jaeger:14268/api/traces"  # For remote mode

Tracing Modes
~~~~~~~~~~~~~

1. **Disabled Mode**: No tracing is performed
2. **Local Mode**: Traces are generated but not exported (for development)
3. **Remote Mode**: Traces are exported to a configured endpoint (Jaeger, Zipkin, etc.)

Current Trace Points
~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Component
     - Trace Name
     - Attributes
     - Description
   * - Router
     - root-handler
     - endpoint, version
     - Root endpoint requests
   * - Alert Handler
     - alert-processing
     - alert_name, receiver, status
     - Alert processing pipeline
   * - HTTP Middleware
     - http-request
     - method, path, status_code
     - All HTTP requests

Example Trace Structure
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    func (rm *RouterManager) rootHandler(c *gin.Context) {
        tr := otel.Tracer(rm.cfg.AppName)
        _, span := tr.Start(c.Request.Context(), "root-handler")
        defer span.End()
        
        span.SetAttributes(
            attribute.String("endpoint", "/"),
            attribute.String("version", rm.cfg.AppVersion),
        )
        
        c.String(http.StatusOK, "Hello world!")
    }

Planned Tracing Enhancements
---------------------------

The following tracing enhancements should be implemented to provide comprehensive observability:

Alert Processing Tracing
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    type AlertTracer struct {
        tracer trace.Tracer
        logger logger.LoggerInterface
    }

    func (at *AlertTracer) TraceAlertProcessing(ctx context.Context, alert *PrometheusAlert) (context.Context, trace.Span) {
        ctx, span := at.tracer.Start(ctx, "alert.processing")
        
        span.SetAttributes(
            attribute.String("alert.name", alert.Labels["alertname"]),
            attribute.String("alert.severity", alert.Labels["severity"]),
            attribute.String("alert.status", alert.Status),
            attribute.String("alert.fingerprint", alert.Fingerprint),
            attribute.String("alert.namespace", alert.Labels["namespace"]),
        )
        
        return ctx, span
    }

    func (at *AlertTracer) TraceWorkflowExecution(ctx context.Context, workflowName string, actionType string) (context.Context, trace.Span) {
        ctx, span := at.tracer.Start(ctx, "workflow.execution")
        
        span.SetAttributes(
            attribute.String("workflow.name", workflowName),
            attribute.String("action.type", actionType),
        )
        
        return ctx, span
    }

    func (at *AlertTracer) TraceIssueCreation(ctx context.Context, issue *Issue) (context.Context, trace.Span) {
        ctx, span := at.tracer.Start(ctx, "issue.creation")
        
        span.SetAttributes(
            attribute.String("issue.id", issue.ID.String()),
            attribute.String("issue.severity", string(issue.Severity)),
            attribute.String("issue.source", string(issue.Source)),
            attribute.String("issue.subject.type", string(issue.Subject.SubjectType)),
            attribute.String("issue.subject.name", issue.Subject.Name),
        )
        
        return ctx, span
    }

Routing Tracing
~~~~~~~~~~~~~~~

.. code-block:: go

    func (r *Router) TraceRoutingDecision(ctx context.Context, alert *PrometheusAlert, teams []Team) (context.Context, trace.Span) {
        ctx, span := r.tracer.Start(ctx, "routing.decision")
        
        span.SetAttributes(
            attribute.String("alert.name", alert.Labels["alertname"]),
            attribute.Int("teams.matched", len(teams)),
        )
        
        // Add team details as events
        for _, team := range teams {
            span.AddEvent("team.matched", trace.WithAttributes(
                attribute.String("team.name", team.Name),
                attribute.String("team.destination.type", team.DestinationType),
            ))
        }
        
        return ctx, span
    }

Destination Tracing
~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    func (d *Destination) TraceMessageSend(ctx context.Context, issue *Issue) (context.Context, trace.Span) {
        ctx, span := d.tracer.Start(ctx, "destination.send")
        
        span.SetAttributes(
            attribute.String("destination.name", d.Name),
            attribute.String("destination.type", d.Type),
            attribute.String("issue.id", issue.ID.String()),
        )
        
        return ctx, span
    }

    func (s *Sender) TraceMessageFormat(ctx context.Context, issue *Issue, destinationType string) (context.Context, trace.Span) {
        ctx, span := s.tracer.Start(ctx, "sender.format")
        
        span.SetAttributes(
            attribute.String("sender.type", destinationType),
            attribute.String("issue.id", issue.ID.String()),
            attribute.Int("enrichments.count", len(issue.Enrichments)),
        )
        
        return ctx, span
    }

Queue Tracing
~~~~~~~~~~~~~

.. code-block:: go

    func (aq *AlertQueue) TraceQueueOperation(ctx context.Context, operation string, alert *PrometheusAlert) (context.Context, trace.Span) {
        ctx, span := aq.tracer.Start(ctx, "queue."+operation)
        
        span.SetAttributes(
            attribute.String("queue.name", aq.name),
            attribute.String("operation", operation),
            attribute.String("alert.fingerprint", alert.Fingerprint),
        )
        
        return ctx, span
    }

    func (aq *AlertQueue) TraceWorkerProcessing(ctx context.Context, task *AlertTask) (context.Context, trace.Span) {
        ctx, span := aq.tracer.Start(ctx, "queue.worker.processing")
        
        span.SetAttributes(
            attribute.String("queue.name", aq.name),
            attribute.String("alert.fingerprint", task.Alert.Fingerprint),
            attribute.Int("task.attempts", task.Attempts),
        )
        
        return ctx, span
    }

Workflow Tracing
~~~~~~~~~~~~~~~~

.. code-block:: go

    func (w *Workflow) TraceWorkflowSelection(ctx context.Context, alert *PrometheusAlert) (context.Context, trace.Span) {
        ctx, span := w.tracer.Start(ctx, "workflow.selection")
        
        span.SetAttributes(
            attribute.String("workflow.name", w.Name),
            attribute.String("alert.name", alert.Labels["alertname"]),
            attribute.Bool("workflow.enabled", w.Enabled),
        )
        
        return ctx, span
    }

    func (w *Workflow) TraceActionExecution(ctx context.Context, action WorkflowAction, alert *PrometheusAlert) (context.Context, trace.Span) {
        ctx, span := w.tracer.Start(ctx, "workflow.action.execution")
        
        span.SetAttributes(
            attribute.String("workflow.name", w.Name),
            attribute.String("action.name", action.GetName()),
            attribute.String("alert.fingerprint", alert.Fingerprint),
        )
        
        return ctx, span
    }

Complete Trace Flow
------------------

A complete trace flow for alert processing would look like:

.. code-block:: go

    func (wh *WorkflowHandler) HandleAlert(c *gin.Context) {
        ctx := c.Request.Context()
        
        // Start root span
        tr := otel.Tracer("cano-collector")
        ctx, span := tr.Start(ctx, "alert.handler")
        defer span.End()
        
        // Parse alert
        alert, err := wh.parseAlert(c)
        if err != nil {
            span.SetStatus(codes.Error, err.Error())
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        
        // Trace alert processing
        ctx, alertSpan := wh.alertTracer.TraceAlertProcessing(ctx, alert)
        defer alertSpan.End()
        
        // Check deduplication
        if wh.deduplication.IsDuplicate(alert) {
            alertSpan.AddEvent("alert.duplicate")
            c.JSON(http.StatusOK, gin.H{"status": "duplicate"})
            return
        }
        
        // Apply relabeling
        alert = wh.applyRelabeling(alert)
        alertSpan.AddEvent("alert.relabeled")
        
        // Enqueue for processing
        ctx, queueSpan := wh.alertQueue.TraceQueueOperation(ctx, "enqueue", alert)
        wh.alertQueue.Enqueue(alert)
        queueSpan.End()
        
        c.JSON(http.StatusOK, gin.H{"status": "queued"})
    }

    func (wh *WorkflowHandler) processAlert(alert *PrometheusAlert) error {
        ctx := context.Background()
        tr := otel.Tracer("cano-collector")
        
        ctx, span := tr.Start(ctx, "alert.processing")
        defer span.End()
        
        // Select workflows
        ctx, selectionSpan := wh.workflowRegistry.TraceWorkflowSelection(ctx, alert)
        workflows := wh.workflowRegistry.SelectWorkflows(alert)
        selectionSpan.SetAttributes(attribute.Int("workflows.selected", len(workflows)))
        selectionSpan.End()
        
        // Execute workflows
        ctx, executionSpan := tr.Start(ctx, "workflows.execution")
        issue, err := wh.executeWorkflows(ctx, alert, workflows)
        if err != nil {
            executionSpan.SetStatus(codes.Error, err.Error())
            return err
        }
        executionSpan.End()
        
        // Trace issue creation
        ctx, issueSpan := wh.alertTracer.TraceIssueCreation(ctx, issue)
        issueSpan.End()
        
        // Route to destinations
        ctx, routingSpan := wh.router.TraceRoutingDecision(ctx, alert, wh.teams)
        err = wh.routeIssue(issue)
        if err != nil {
            routingSpan.SetStatus(codes.Error, err.Error())
            return err
        }
        routingSpan.End()
        
        return nil
    }

Trace Attributes and Events
--------------------------

Key attributes to include in traces:

.. list-table::
   :header-rows: 1

   * - Attribute
     - Type
     - Description
     - Example
   * - alert.name
     - String
     - Name of the alert
     - "PodCrashLooping"
   * - alert.severity
     - String
     - Alert severity level
     - "warning"
   * - alert.fingerprint
     - String
     - Unique alert identifier
     - "abc123"
   * - workflow.name
     - String
     - Name of the workflow
     - "pod-crashloop-enrichment"
   * - destination.name
     - String
     - Name of the destination
     - "alerts-prod"
   * - destination.type
     - String
     - Type of destination
     - "slack"

Key events to include in traces:

.. list-table::
   :header-rows: 1

   * - Event
     - Description
     - Attributes
   * - alert.received
     - Alert received from Alertmanager
     - receiver, status
   * - alert.duplicate
     - Alert identified as duplicate
     - fingerprint
   * - alert.relabeled
     - Alert labels modified
     - labels_changed
   * - workflow.selected
     - Workflow selected for alert
     - workflow_name, trigger_type
   * - workflow.executed
     - Workflow execution completed
     - workflow_name, actions_count
   * - destination.matched
     - Destination matched for routing
     - destination_name, team_name
   * - message.sent
     - Message sent to destination
     - destination_name, status

OpenTelemetry Configuration
--------------------------

Complete OpenTelemetry setup:

.. code-block:: go

    func setupTracing(cfg *config.Config) (*trace.TracerProvider, error) {
        if cfg.TracingMode == "disabled" {
            return trace.NewNoopTracerProvider(), nil
        }
        
        // Create resource
        res, err := resource.New(context.Background(),
            resource.WithAttributes(
                attribute.String("service.name", cfg.AppName),
                attribute.String("service.version", cfg.AppVersion),
                attribute.String("service.environment", cfg.AppEnv),
            ),
        )
        if err != nil {
            return nil, err
        }
        
        // Create exporter
        var exp trace.SpanExporter
        if cfg.TracingMode == "remote" {
            exp, err = otlptrace.New(context.Background(), otlptracehttp.NewClient(
                otlptracehttp.WithEndpoint(cfg.TracingEndpoint),
                otlptracehttp.WithInsecure(),
            ))
            if err != nil {
                return nil, err
            }
        } else {
            // Local mode - no export
            exp = &noopSpanExporter{}
        }
        
        // Create tracer provider
        tp := trace.NewTracerProvider(
            trace.WithBatcher(exp),
            trace.WithResource(res),
            trace.WithSampler(trace.AlwaysSample()),
        )
        
        // Set global tracer provider
        otel.SetTracerProvider(tp)
        
        return tp, nil
    }

    type noopSpanExporter struct{}
    
    func (n *noopSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
        return nil
    }
    
    func (n *noopSpanExporter) Shutdown(ctx context.Context) error {
        return nil
    }

Tracing Best Practices
---------------------

1. **Span Naming**: Use descriptive, hierarchical names (e.g., "alert.processing", "workflow.execution")
2. **Attribute Selection**: Include relevant business context without high cardinality
3. **Error Handling**: Always set span status and record errors
4. **Event Usage**: Use events for important state changes
5. **Context Propagation**: Pass context through all function calls
6. **Sampling**: Use appropriate sampling strategies for production

Example Jaeger Query
-------------------

Useful Jaeger queries for debugging:

.. code-block:: sql

    -- Find all traces for a specific alert
    service.name="cano-collector" AND alert.name="PodCrashLooping"
    
    -- Find slow workflow executions
    service.name="cano-collector" AND operation="workflow.execution" AND duration > 5s
    
    -- Find failed destination sends
    service.name="cano-collector" AND operation="destination.send" AND error=true
    
    -- Find traces with specific workflow
    service.name="cano-collector" AND workflow.name="pod-crashloop-enrichment"

This comprehensive tracing approach provides:

- **End-to-End Visibility**: Complete view of alert processing flow
- **Performance Analysis**: Identify bottlenecks and slow operations
- **Error Debugging**: Quickly locate and understand failures
- **Business Context**: Understand alert processing patterns
- **Operational Insights**: Monitor system behavior and health 