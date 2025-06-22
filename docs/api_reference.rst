API Reference
============

This document provides detailed API documentation for cano-collector's internal interfaces and data structures.

Core Interfaces
---------------

AlertHandlerInterface
~~~~~~~~~~~~~~~~~~~~~

Handles incoming alerts from Alertmanager.

.. code-block:: go

    type AlertHandlerInterface interface {
        HandleAlert(c *gin.Context)
    }

**Location:** `pkg/alert/alert_handler.go`

**Methods:**
- `HandleAlert(c *gin.Context)` - Processes HTTP requests from Alertmanager

**Example Usage:**
.. code-block:: go

    handler := alert.NewAlertHandler(logger, metrics)
    // Used by router to handle /api/alerts endpoint

DestinationSender Interface
~~~~~~~~~~~~~~~~~~~~~~~~~~~

Sends alerts to various notification destinations.

.. code-block:: go

    type DestinationSender interface {
        Send(alert Alert) error
    }

**Location:** `pkg/sender/sender.go`

**Methods:**
- `Send(alert Alert) error` - Sends an alert to the destination

**Example Usage:**
.. code-block:: go

    sender := sender.NewSlackSender(webhookURL, logger, client)
    err := sender.Send(sender.Alert{
        Title:   "Pod CrashLooping",
        Message: "Pod is in CrashLoopBackOff state",
    })

LoggerInterface
~~~~~~~~~~~~~~

Provides structured logging capabilities.

.. code-block:: go

    type LoggerInterface interface {
        Debug(msg string, fields ...zap.Field)
        Info(msg string, fields ...zap.Field)
        Warn(msg string, fields ...zap.Field)
        Error(msg string, fields ...zap.Field)
        Fatal(msg string, fields ...zap.Field)
        Panic(msg string, fields ...zap.Field)
    }

**Location:** `pkg/logger/logger.go`

**Methods:**
- `Debug/Info/Warn/Error/Fatal/Panic(msg string, fields ...zap.Field)` - Log at different levels

**Example Usage:**
.. code-block:: go

    logger.Info("Alert processed", 
        zap.String("alert_name", alert.Labels["alertname"]),
        zap.String("severity", alert.Labels["severity"]),
    )

MetricsInterface
~~~~~~~~~~~~~~~~

Collects and exposes application metrics.

.. code-block:: go

    type MetricsInterface interface {
        IncAlertReceived(receiver, status string)
        IncAlertProcessed(receiver, status string)
        IncAlertError(receiver, errorType string)
        ObserveAlertProcessingDuration(duration time.Duration)
        IncDestinationSent(destination string)
        IncDestinationError(destination string)
        ObserveDestinationDuration(destination string, duration time.Duration)
    }

**Location:** `pkg/metric/metric.go`

**Methods:**
- `IncAlertReceived(receiver, status string)` - Increment alert received counter
- `IncAlertProcessed(receiver, status string)` - Increment alert processed counter
- `IncAlertError(receiver, errorType string)` - Increment error counter
- `ObserveAlertProcessingDuration(duration time.Duration)` - Record processing time
- `IncDestinationSent(destination string)` - Increment destination sent counter
- `IncDestinationError(destination string)` - Increment destination error counter
- `ObserveDestinationDuration(destination string, duration time.Duration)` - Record destination send time

**Example Usage:**
.. code-block:: go

    metrics.IncAlertReceived(alert.Receiver, alert.Status)
    start := time.Now()
    // Process alert
    metrics.ObserveAlertProcessingDuration(time.Since(start))

Data Structures
---------------

Issue
~~~~~

Central data structure for all alerts and events.

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

**Location:** `pkg/core/issue/issue.go`

**Fields:**
- `ID` - Unique identifier for the issue
- `Title` - Human-readable title
- `Description` - Detailed description
- `AggregationKey` - Key for grouping similar issues
- `Severity` - Issue severity level (DEBUG, INFO, LOW, HIGH)
- `Status` - Current status (FIRING, RESOLVED)
- `Source` - Origin of the issue (PROMETHEUS, KUBERNETES_API_SERVER, etc.)
- `Subject` - Information about the affected resource
- `Enrichments` - Additional context blocks
- `Links` - Related URLs
- `Fingerprint` - Unique hash for deduplication
- `StartsAt` - When the issue started
- `EndsAt` - When the issue ended (nil if ongoing)

**Example Usage:**
.. code-block:: go

    issue := &issue.Issue{
        ID:             uuid.New(),
        Title:          "Pod CrashLooping",
        Description:    "Pod is in CrashLoopBackOff state",
        AggregationKey: "PodCrashLooping",
        Severity:       issue.SeverityHigh,
        Status:         issue.StatusFiring,
        Source:         issue.SourcePrometheus,
        Subject: issue.Subject{
            Name:        "my-app-pod",
            SubjectType: issue.SubjectTypePod,
            Namespace:   "default",
        },
        Fingerprint: "abc123",
        StartsAt:    time.Now(),
    }

Subject
~~~~~~~

Information about the Kubernetes resource related to the issue.

.. code-block:: go

    type Subject struct {
        Name        string
        SubjectType SubjectType
        Namespace   string
        Node        string
        Container   string
        Labels      map[string]string
        Annotations map[string]string
    }

**Location:** `pkg/core/issue/issue.go`

**Fields:**
- `Name` - Resource name
- `SubjectType` - Type of resource (pod, deployment, node, etc.)
- `Namespace` - Kubernetes namespace
- `Node` - Node name (for pod-related issues)
- `Container` - Container name (for pod-related issues)
- `Labels` - Kubernetes labels
- `Annotations` - Kubernetes annotations

Enrichment
~~~~~~~~~~

Additional context data for an issue.

.. code-block:: go

    type Enrichment struct {
        Blocks []BaseBlock
        Annotations map[string]string
    }

**Location:** `pkg/core/issue/issue.go`

**Fields:**
- `Blocks` - Content blocks for rendering
- `Annotations` - Metadata for senders

BaseBlock Interface
~~~~~~~~~~~~~~~~~~

Interface for all content blocks.

.. code-block:: go

    type BaseBlock interface {
        IsBlock()
    }

**Location:** `pkg/core/issue/blocks.go`

**Implementations:**
- `MarkdownBlock` - Markdown text
- `TableBlock` - Tabular data
- `FileBlock` - File attachments
- `ListBlock` - Lists
- `HeaderBlock` - Headers
- `DividerBlock` - Visual separators
- `LinksBlock` - Clickable links

**Example Usage:**
.. code-block:: go

    enrichment := issue.Enrichment{
        Blocks: []issue.BaseBlock{
            issue.MarkdownBlock{Text: "**Pod Logs:**\n```\nError: connection refused\n```"},
            issue.TableBlock{
                Name:    "Resource Status",
                Headers: []string{"Field", "Value"},
                Rows:    [][]string{{"Status", "CrashLoopBackOff"}},
            },
        },
    }

Configuration Structures
------------------------

Config
~~~~~~

Main application configuration.

.. code-block:: go

    type Config struct {
        AppName         string
        AppVersion      string
        AppEnv          string
        LogLevel        string
        TracingMode     string
        TracingEndpoint string
        SentryDSN       string
        SentryEnabled   bool
        Destinations    destination.DestinationsConfig
        Teams           team.TeamsConfig
    }

**Location:** `config/config.go`

**Fields:**
- `AppName` - Application name
- `AppVersion` - Application version
- `AppEnv` - Environment (production, development, etc.)
- `LogLevel` - Logging level (debug, info, warn, error)
- `TracingMode` - Tracing mode (disabled, local, remote)
- `TracingEndpoint` - Tracing endpoint URL
- `SentryDSN` - Sentry DSN for error reporting
- `SentryEnabled` - Whether Sentry is enabled
- `Destinations` - Destination configurations
- `Teams` - Team configurations

DestinationsConfig
~~~~~~~~~~~~~~~~~

Configuration for notification destinations.

.. code-block:: go

    type DestinationsConfig struct {
        Destinations struct {
            Slack []Destination `yaml:"slack"`
            Teams []Destination `yaml:"teams"`
        } `yaml:"destinations"`
    }

**Location:** `config/destination/destinations_config.go`

**Fields:**
- `Destinations.Slack` - Slack webhook configurations
- `Destinations.Teams` - MS Teams webhook configurations

Destination
~~~~~~~~~~~

Individual destination configuration.

.. code-block:: go

    type Destination struct {
        Name       string `yaml:"name"`
        WebhookURL string `yaml:"webhookURL"`
    }

**Location:** `config/destination/destinations_config.go`

**Fields:**
- `Name` - Unique destination name
- `WebhookURL` - Webhook URL for the destination

HTTP Endpoints
--------------

Alert Endpoint
~~~~~~~~~~~~~

Receives alerts from Alertmanager.

**Endpoint:** `POST /api/alerts`

**Request Body:** Alertmanager webhook format

.. code-block:: json

    {
      "receiver": "cano-collector",
      "status": "firing",
      "alerts": [
        {
          "status": "firing",
          "labels": {
            "alertname": "PodCrashLooping",
            "severity": "warning",
            "pod": "my-app-pod",
            "namespace": "default"
          },
          "annotations": {
            "summary": "Pod is crash looping",
            "description": "Pod my-app-pod is in CrashLoopBackOff state"
          },
          "startsAt": "2024-01-15T10:30:00Z",
          "endsAt": "2024-01-15T10:35:00Z",
          "fingerprint": "abc123"
        }
      ]
    }

**Response:**
- `200 OK` - Alert received successfully
- `400 Bad Request` - Invalid alert format
- `500 Internal Server Error` - Processing error

Health Endpoint
~~~~~~~~~~~~~~

Provides health status information.

**Endpoint:** `GET /health`

**Response:**
.. code-block:: json

    {
      "status": "ok"
    }

**Endpoint:** `GET /health/detailed`

**Response:**
.. code-block:: json

    {
      "status": "ok",
      "components": {
        "config": "ok",
        "destinations": "ok"
      },
      "uptime": "2h30m15s",
      "version": "1.0.0"
    }

Metrics Endpoint
~~~~~~~~~~~~~~~

Exposes Prometheus metrics.

**Endpoint:** `GET /metrics`

**Response:** Prometheus metrics in text format

.. code-block:: text

    # HELP cano_alerts_received_total Total alerts received
    # TYPE cano_alerts_received_total counter
    cano_alerts_received_total{receiver="cano-collector",status="firing"} 42

    # HELP cano_alerts_processing_duration_seconds Alert processing duration
    # TYPE cano_alerts_processing_duration_seconds histogram
    cano_alerts_processing_duration_seconds_bucket{le="0.1"} 35
    cano_alerts_processing_duration_seconds_bucket{le="0.5"} 40
    cano_alerts_processing_duration_seconds_bucket{le="1"} 42

Error Handling
--------------

Error Types
~~~~~~~~~~

Common error types and their handling:

1. **Configuration Errors:**
   - Invalid YAML syntax
   - Missing required fields
   - Invalid webhook URLs

2. **Network Errors:**
   - Destination unreachable
   - Timeout errors
   - Authentication failures

3. **Processing Errors:**
   - Invalid alert format
   - Enrichment failures
   - Destination send failures

Error Response Format
~~~~~~~~~~~~~~~~~~~

All error responses follow this format:

.. code-block:: json

    {
      "error": "Error description",
      "details": "Additional error details"
    }

**HTTP Status Codes:**
- `400 Bad Request` - Client error (invalid input)
- `500 Internal Server Error` - Server error (processing failure)

Logging
-------

Log Levels
~~~~~~~~~

- `DEBUG` - Detailed debugging information
- `INFO` - General information about application flow
- `WARN` - Warning messages for potentially harmful situations
- `ERROR` - Error messages for error conditions
- `FATAL` - Fatal errors that cause application termination
- `PANIC` - Panic messages for unrecoverable errors

Structured Logging
~~~~~~~~~~~~~~~~~

All logs use structured logging with zap fields:

.. code-block:: go

    logger.Info("Alert processed",
        zap.String("alert_name", alert.Labels["alertname"]),
        zap.String("severity", alert.Labels["severity"]),
        zap.String("namespace", alert.Labels["namespace"]),
        zap.String("pod", alert.Labels["pod"]),
        zap.Duration("processing_time", processingTime),
    )

Common Log Fields
~~~~~~~~~~~~~~~~

- `alert_name` - Name of the alert
- `severity` - Alert severity level
- `namespace` - Kubernetes namespace
- `pod` - Pod name
- `destination` - Destination name
- `processing_time` - Time taken to process
- `error` - Error details
- `status` - Processing status

Metrics
-------

Alert Metrics
~~~~~~~~~~~~

- `cano_alerts_received_total` - Total alerts received
- `cano_alerts_processed_total` - Total alerts processed
- `cano_alerts_errors_total` - Total processing errors
- `cano_alerts_processing_duration_seconds` - Processing time histogram

Destination Metrics
~~~~~~~~~~~~~~~~~~

- `cano_destination_sent_total` - Messages sent per destination
- `cano_destination_errors_total` - Send errors per destination
- `cano_destination_duration_seconds` - Send duration per destination

System Metrics
~~~~~~~~~~~~~

- `cano_http_requests_total` - HTTP request count
- `cano_http_request_duration_seconds` - HTTP request duration
- `cano_config_reloads_total` - Configuration reload count

Metric Labels
~~~~~~~~~~~~

- `receiver` - Alertmanager receiver name
- `status` - Alert status (firing, resolved)
- `destination` - Destination name
- `method` - HTTP method
- `endpoint` - HTTP endpoint
- `status_code` - HTTP status code 