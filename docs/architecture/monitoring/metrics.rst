Application Metrics
==================

This document describes the metrics exposed by cano-collector at the `/metrics` endpoint, including existing metrics and planned business metrics for comprehensive monitoring.

Current Metrics
---------------

Cano-collector currently exposes the following metrics:

HTTP Request Metrics
~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
   * - http_requests_total
     - Counter
     - method, path, status
     - Total number of HTTP requests processed

Alert Processing Metrics
~~~~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
   * - alertmanager_alerts_total
     - Counter
     - receiver, status
     - Total number of alerts received from AlertManager

Go Runtime Metrics
~~~~~~~~~~~~~~~~~~

Standard Go runtime metrics are automatically exposed:

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Description
   * - go_goroutines
     - Gauge
     - Number of goroutines currently running
   * - go_threads
     - Gauge
     - Number of OS threads created
   * - go_heap_alloc_bytes
     - Gauge
     - Current heap memory usage
   * - go_heap_sys_bytes
     - Gauge
     - Total heap memory allocated
   * - go_gc_cycles_total
     - Counter
     - Total number of garbage collection cycles

Planned Business Metrics
------------------------

The following business metrics should be implemented to provide comprehensive monitoring of cano-collector's operations:

Alert Processing Metrics
~~~~~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
     - Implementation Priority
   * - cano_alerts_processed_total
     - Counter
     - alert_name, severity, source
     - Total number of alerts processed
     - High
   * - cano_alerts_processing_duration_seconds
     - Histogram
     - alert_name, workflow_count
     - Time spent processing alerts
     - High
   * - cano_alerts_deduplicated_total
     - Counter
     - alert_name
     - Number of duplicate alerts detected
     - Medium
   * - cano_alerts_relabeled_total
     - Counter
     - source_label, target_label
     - Number of label relabeling operations
     - Medium
   * - cano_alerts_queue_size
     - Gauge
     - queue_name
     - Current size of alert processing queues
     - High
   * - cano_alerts_queue_duration_seconds
     - Histogram
     - queue_name
     - Time alerts spend in processing queues
     - High

Workflow Processing Metrics
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
     - Implementation Priority
   * - cano_workflows_executed_total
     - Counter
     - workflow_name, action_type, status
     - Total number of workflow actions executed
     - High
   * - cano_workflow_execution_duration_seconds
     - Histogram
     - workflow_name, action_type
     - Time spent executing workflow actions
     - High
   * - cano_workflows_selected_total
     - Counter
     - workflow_name, trigger_type
     - Number of times workflows were selected for alerts
     - Medium
   * - cano_enrichments_created_total
     - Counter
     - enrichment_type, block_type
     - Number of enrichment blocks created
     - Medium

Routing Metrics
~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
     - Implementation Priority
   * - cano_routing_decisions_total
     - Counter
     - team_name, destination_type, decision
     - Number of routing decisions made
     - High
   * - cano_routing_duration_seconds
     - Histogram
     - team_name
     - Time spent on routing decisions
     - Medium
   * - cano_teams_matched_total
     - Counter
     - team_name, alert_name
     - Number of team matches for alerts
     - High

Destination/Sender Metrics
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
     - Implementation Priority
   * - cano_destination_messages_sent_total
     - Counter
     - destination_name, destination_type, status
     - Total messages sent to destinations
     - High
   * - cano_destination_send_duration_seconds
     - Histogram
     - destination_name, destination_type
     - Time spent sending messages to destinations
     - High
   * - cano_destination_errors_total
     - Counter
     - destination_name, destination_type, error_type
     - Number of destination send errors
     - High
   * - cano_destination_health_status
     - Gauge
     - destination_name, destination_type
     - Health status of destinations (1=healthy, 0=unhealthy)
     - High
   * - cano_destination_retry_attempts_total
     - Counter
     - destination_name, destination_type
     - Number of retry attempts for failed sends
     - Medium

Resource Usage Metrics
~~~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
     - Implementation Priority
   * - cano_memory_usage_bytes
     - Gauge
     - type (heap, stack, system)
     - Memory usage by type
     - Medium
   * - cano_goroutine_count
     - Gauge
     - N/A
     - Number of active goroutines
     - Medium
   * - cano_cpu_usage_percent
     - Gauge
     - N/A
     - CPU usage percentage
     - Low

Configuration Metrics
~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Metric Name
     - Type
     - Labels
     - Description
     - Implementation Priority
   * - cano_configuration_reloads_total
     - Counter
     - config_type, status
     - Number of configuration reloads
     - Medium
   * - cano_configuration_errors_total
     - Counter
     - config_type, error_type
     - Number of configuration errors
     - High

Implementation Example
---------------------

Example implementation of the metrics collector:

.. code-block:: go

    type MetricsCollector struct {
        alertsProcessedTotal           *prometheus.CounterVec
        alertsProcessingDuration       *prometheus.HistogramVec
        alertsDeduplicatedTotal        *prometheus.CounterVec
        alertsQueueSize                *prometheus.GaugeVec
        workflowsExecutedTotal         *prometheus.CounterVec
        workflowExecutionDuration      *prometheus.HistogramVec
        routingDecisionsTotal          *prometheus.CounterVec
        destinationMessagesSentTotal   *prometheus.CounterVec
        destinationSendDuration        *prometheus.HistogramVec
        destinationErrorsTotal         *prometheus.CounterVec
        destinationHealthStatus        *prometheus.GaugeVec
    }

    func NewMetricsCollector() *MetricsCollector {
        return &MetricsCollector{
            alertsProcessedTotal: prometheus.NewCounterVec(
                prometheus.CounterOpts{
                    Name: "cano_alerts_processed_total",
                    Help: "Total number of alerts processed",
                },
                []string{"alert_name", "severity", "source"},
            ),
            alertsProcessingDuration: prometheus.NewHistogramVec(
                prometheus.HistogramOpts{
                    Name:    "cano_alerts_processing_duration_seconds",
                    Help:    "Time spent processing alerts",
                    Buckets: prometheus.DefBuckets,
                },
                []string{"alert_name", "workflow_count"},
            ),
            // ... other metrics initialization
        }
    }

    func (mc *MetricsCollector) IncAlertsProcessed(alertName, severity, source string) {
        mc.alertsProcessedTotal.WithLabelValues(alertName, severity, source).Inc()
    }

    func (mc *MetricsCollector) ObserveAlertProcessingDuration(alertName string, workflowCount int, duration time.Duration) {
        mc.alertsProcessingDuration.WithLabelValues(alertName, strconv.Itoa(workflowCount)).Observe(duration.Seconds())
    }

    func (mc *MetricsCollector) IncWorkflowsExecuted(workflowName, actionType, status string) {
        mc.workflowsExecutedTotal.WithLabelValues(workflowName, actionType, status).Inc()
    }

    func (mc *MetricsCollector) ObserveWorkflowExecutionDuration(workflowName, actionType string, duration time.Duration) {
        mc.workflowExecutionDuration.WithLabelValues(workflowName, actionType).Observe(duration.Seconds())
    }

    func (mc *MetricsCollector) IncRoutingDecisions(teamName, destinationType, decision string) {
        mc.routingDecisionsTotal.WithLabelValues(teamName, destinationType, decision).Inc()
    }

    func (mc *MetricsCollector) IncDestinationMessagesSent(destinationName, destinationType, status string) {
        mc.destinationMessagesSentTotal.WithLabelValues(destinationName, destinationType, status).Inc()
    }

    func (mc *MetricsCollector) ObserveDestinationSendDuration(destinationName, destinationType string, duration time.Duration) {
        mc.destinationSendDuration.WithLabelValues(destinationName, destinationType).Observe(duration.Seconds())
    }

    func (mc *MetricsCollector) IncDestinationErrors(destinationName, destinationType, errorType string) {
        mc.destinationErrorsTotal.WithLabelValues(destinationName, destinationType, errorType).Inc()
    }

    func (mc *MetricsCollector) SetDestinationHealthStatus(destinationName, destinationType string, healthy bool) {
        status := 0
        if healthy {
            status = 1
        }
        mc.destinationHealthStatus.WithLabelValues(destinationName, destinationType).Set(float64(status))
    }

Alerting Rules
--------------

Recommended Prometheus alerting rules for cano-collector:

.. code-block:: yaml

    groups:
      - name: cano-collector
        rules:
          # High error rate
          - alert: CanoCollectorHighErrorRate
            expr: rate(cano_destination_errors_total[5m]) > 0.1
            for: 2m
            labels:
              severity: warning
            annotations:
              summary: "High error rate in cano-collector"
              description: "Cano-collector is experiencing a high rate of destination errors"

          # Queue backlog
          - alert: CanoCollectorQueueBacklog
            expr: cano_alerts_queue_size > 100
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: "Alert processing queue backlog"
              description: "Alert processing queue has more than 100 items"

          # Destination health
          - alert: CanoCollectorDestinationUnhealthy
            expr: cano_destination_health_status == 0
            for: 2m
            labels:
              severity: warning
            annotations:
              summary: "Destination is unhealthy"
              description: "Destination {{ $labels.destination_name }} is reporting unhealthy status"

          # High processing time
          - alert: CanoCollectorHighProcessingTime
            expr: histogram_quantile(0.95, rate(cano_alerts_processing_duration_seconds_bucket[5m])) > 30
            for: 2m
            labels:
              severity: warning
            annotations:
              summary: "High alert processing time"
              description: "95th percentile of alert processing time is above 30 seconds"

          # High memory usage
          - alert: CanoCollectorHighMemoryUsage
            expr: cano_memory_usage_bytes{type="heap"} > 1e9
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: "High memory usage"
              description: "Cano-collector is using more than 1GB of heap memory"

          # High goroutine count
          - alert: CanoCollectorHighGoroutineCount
            expr: cano_goroutine_count > 1000
            for: 5m
            labels:
              severity: warning
            annotations:
              summary: "High goroutine count"
              description: "Cano-collector has more than 1000 active goroutines"

Grafana Dashboard
----------------

A comprehensive Grafana dashboard should include:

1. **Alert Processing Overview**:
   - Alert processing rate
   - Processing duration percentiles
   - Queue size and backlog
   - Deduplication rate

2. **Workflow Performance**:
   - Workflow execution rate
   - Execution duration by workflow type
   - Workflow selection rate
   - Enrichment block creation rate

3. **Routing and Destination Health**:
   - Routing decision rate
   - Destination message send rate
   - Destination error rate
   - Destination health status

4. **System Resources**:
   - Memory usage by type
   - Goroutine count
   - CPU usage
   - Configuration reload rate

5. **Error Analysis**:
   - Error rate by type
   - Error distribution by destination
   - Retry attempt rate
   - Configuration error rate

This comprehensive metrics approach provides full observability into cano-collector's operations and performance. 