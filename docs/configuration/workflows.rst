Workflows Configuration
=======================

Cano-collector supports two types of workflows:

1. **Built-in workflows** - Pre-compiled Go workflows configured via YAML
2. **Custom workflows** - TypeScript workflows loaded dynamically at runtime

This document describes the configuration for built-in workflows. For custom workflows, see :doc:`../architecture/workflow/custom`.

Built-in Workflows Configuration
--------------------------------

Built-in workflows are configured through a YAML file that defines workflow definitions, triggers, and actions.

Event Types
~~~~~~~~~~~

Cano-collector uses internal event types to process data from various sources:

**Base Event Structure:**
All events inherit from a base event structure that provides common fields:

.. code-block:: go

   type BaseEvent struct {
       ID        uuid.UUID
       Timestamp time.Time
       Source    string
       Type      string
   }

**Currently Supported Event Types:**

- **AlertEvent**: Converted from Alertmanager `template.Data`
  - Contains alert-specific fields like `AlertName`, `Severity`, `Labels`, etc.
  - Triggered by `on_alertmanager_alert`
  - Includes optional Kubernetes resource data (Pod, Node, Deployment)

- **KubernetesEvent**: Converted from Kubernetes events
  - Contains K8s-specific fields like `ResourceType`, `Operation`, `Namespace`, etc.
  - Triggered by `on_kubernetes_event`
  - Supports all Kubernetes resource types

- **ScheduledEvent**: For scheduled tasks
  - Contains scheduling information like `Recurrence`, `TaskName`, `Schedule`
  - Triggered by `on_scheduled_event`

**Planned Event Types:**

- **HelmEvent**: For Helm release changes
  - Contains Helm-specific fields like `ReleaseName`, `Status`, `Version`
  - Triggered by `on_helm_event` (future)

- **WebhookEvent**: For external webhook integrations
  - Contains webhook data like `Payload`, `Headers`, `Source`
  - Triggered by `on_webhook_event` (future)

- **CustomEvent**: For custom integrations
  - Flexible structure for custom data sources
  - Triggered by `on_custom_event` (future)

Workflows receive concrete event types and can access all event-specific fields through the event object.

**Event Structure Details:**

.. code-block:: go

   // Base event structure
   type BaseEvent struct {
       ID        uuid.UUID
       Timestamp time.Time
       Source    string
       Type      string
   }

   // Alert event - most commonly used
   type AlertEvent struct {
       BaseEvent
       Alert Alert
       
       // Most commonly used Kubernetes resources
       Pod        *corev1.Pod
       Node       *corev1.Node
       Deployment *appsv1.Deployment
       
       // Other resources as needed
       OtherResources map[string]interface{}
   }

   // Kubernetes event for resource changes
   type KubernetesEvent struct {
       BaseEvent
       ResourceType string
       Operation    string // Create, Update, Delete
       Namespace    string
       ResourceName string
       Labels       map[string]string
       Annotations  map[string]string
       
       // Most commonly used resources
       Pod        *corev1.Pod
       Node       *corev1.Node
       Deployment *appsv1.Deployment
       
       // Other resources as needed
       OtherResources map[string]interface{}
   }

   // Scheduled event for periodic tasks
   type ScheduledEvent struct {
       BaseEvent
       Recurrence int
       TaskName   string
       Schedule   string // cron expression
   }

**Helper Methods:**

.. code-block:: go

   // AlertEvent helper methods
   func (e AlertEvent) HasPod() bool {
       return e.Pod != nil
   }

   func (e AlertEvent) HasNode() bool {
       return e.Node != nil
   }

   func (e AlertEvent) HasDeployment() bool {
       return e.Deployment != nil
   }

   func (e AlertEvent) GetResource(kind string) interface{} {
       switch kind {
       case "Pod":
           return e.Pod
       case "Node":
           return e.Node
       case "Deployment":
           return e.Deployment
       default:
           return e.OtherResources[kind]
       }
   }

**Usage Examples:**

.. code-block:: go

   // Processing alert events
   func processAlert(event AlertEvent) {
       if event.HasPod() {
           podName := event.Pod.Name
           // Process pod-specific logic
       }
       
       if event.HasNode() {
           nodeName := event.Node.Name
           // Process node-specific logic
       }
       
       // Access other resources
       if service := event.GetResource("Service"); service != nil {
           // Process service-specific logic
       }
   }

   // Processing Kubernetes events
   func processK8sEvent(event KubernetesEvent) {
       switch event.Operation {
       case "Create":
           // Handle resource creation
       case "Update":
           // Handle resource updates
       case "Delete":
           // Handle resource deletion
       }
   }

**Key Concept**: Workflows process internal event types and execute actions. Events are converted from external sources (like Alertmanager's `template.Data`) into internal event types. One of the most important actions is `create_issue` which converts event data into the internal Issue model.

Configuration Structure
~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

   workflows:
     - name: "standard-alert-processing"
       triggers:
         - on_alertmanager_alert:
             alert_name: "PodCrashLooping"
             status: "firing"
             severity: "critical"
             namespace: "production"
             instance: "*.example.com"
             pod_name: "api-*"
       actions:
         - name: "create-issue"
           type: "create_issue"
           data:
             title_template: "Pod {{.pod_name}} is crash looping"
             severity_override: "high"
         - name: "enrich-with-logs"
           type: "pod_logs"
           data:
             lines: 100
             follow: false
       stop_on_match: true

Workflow Definition
~~~~~~~~~~~~~~~~~~

Each workflow definition contains:

- **name** (required): Unique identifier for the workflow
- **triggers** (required): List of conditions that activate the workflow
- **actions** (required): List of operations to perform when triggered
- **stop_on_match** (optional): Whether to stop processing other workflows after this one matches

.. code-block:: yaml

   name: "my-workflow"
   triggers: [...]
   actions: [...]
   stop_on_match: false

Trigger Definitions
~~~~~~~~~~~~~~~~~

Currently supported trigger types:

AlertManager Alert Trigger
^^^^^^^^^^^^^^^^^^^^^^^^^^^

Triggers on Prometheus alerts forwarded by AlertManager. These alerts are converted to internal `AlertEvent` type.

Kubernetes Event Trigger
^^^^^^^^^^^^^^^^^^^^^^^^^

Triggers on Kubernetes events. These events are converted to internal `KubernetesEvent` type.

.. code-block:: yaml

   triggers:
     - on_kubernetes_event:
         resource_type: "Pod"              # Optional: resource type
         operation: "Created"              # Optional: operation type
         namespace: "production"           # Optional: namespace
         resource_name: "api-*"           # Optional: resource name pattern

Custom Event Trigger
^^^^^^^^^^^^^^^^^^^

Triggers on custom events from external integrations. These events are converted to internal `CustomEvent` type.

.. code-block:: yaml

   triggers:
     - on_custom_event:
         event_type: "deployment"          # Optional: event type
         source: "jenkins"                 # Optional: event source
         environment: "production"         # Optional: environment

.. code-block:: yaml

   triggers:
     - on_alertmanager_alert:
         alert_name: "PodCrashLooping"     # Optional: exact alert name
         status: "firing"                  # Optional: "firing" or "resolved"
         severity: "critical"              # Optional: alert severity
         namespace: "production"           # Optional: Kubernetes namespace
         instance: "*.example.com"         # Optional: alert instance (supports wildcards)
         pod_name: "api-*"                 # Optional: pod name pattern (supports wildcards)

All trigger fields are optional. When multiple fields are specified, ALL must match for the trigger to fire.

**Wildcard Support:**
- Use ``*`` for any number of characters
- Use ``?`` for single character
- Examples: ``api-*``, ``*.example.com``, ``prod-?-service``

Action Definitions
~~~~~~~~~~~~~~~~

Actions define what operations to perform when a workflow is triggered.

.. code-block:: yaml

   actions:
     - name: "action-name"
       type: "action-type"
       data:
         # Action-specific configuration
         key: "value"

**Action Structure:**
- **name** (required): Unique identifier for the action
- **type** (required): Type of action to perform
- **data** (optional): Action-specific configuration as key-value pairs

**Core Action Types:**

``create_issue``
  Creates an Issue from alert data. This is typically the first action in most workflows.
  
  - **title_template**: Template for issue title (supports Go template syntax)
  - **severity_override**: Override alert severity
  - **enrich_with**: List of enrichment types to add

``enrich_data``
  Adds contextual information to be included in the Issue.
  
  - **type**: Type of enrichment (logs, events, metrics, etc.)
  - **source**: Data source for enrichment
  - **parameters**: Enrichment-specific parameters

``filter_alert``
  Determines whether to continue processing the alert.
  
  - **condition**: Condition to evaluate
  - **action**: "continue" or "stop"

``transform_data``
  Modifies alert data before Issue creation.
  
  - **field**: Field to modify
  - **value**: New value or template
  - **operation**: "set", "append", "replace"

**Current Implementation:**
Actions currently use a flexible ``data`` field that accepts any key-value pairs. Specific action types and their required/optional fields will be defined as the action framework is implemented.

Configuration Loading
~~~~~~~~~~~~~~~~~~~~

Built-in workflows are loaded from:

1. **File path**: ``/etc/cano-collector/workflows/workflows.yaml``
2. **Helm configuration**: Mounted as ConfigMap or Secret
3. **Environment**: ``CANO_WORKFLOWS_CONFIG`` environment variable

Example Helm Configuration
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

   # values.yaml
   workflows:
     enabled: true
     config:
       workflows:
         - name: "critical-alert-processing"
           triggers:
             - on_alertmanager_alert:
                 severity: "critical"
                 status: "firing"
           actions:
             - name: "create-critical-issue"
               type: "create_issue"
               data:
                 title_template: "🚨 CRITICAL: {{.alert_name}} in {{.namespace}}"
                 severity_override: "critical"
             - name: "add-context"
               type: "enrich_data"
               data:
                 type: "kubernetes_context"
                 include_logs: true
                 include_events: true
           stop_on_match: false

         - name: "pod-crash-analysis"
           triggers:
             - on_alertmanager_alert:
                 alert_name: "PodCrashLooping"
                 namespace: "production"
           actions:
             - name: "create-crash-issue"
               type: "create_issue"
               data:
                 title_template: "Pod {{.pod_name}} is crash looping"
                 severity_override: "high"
             - name: "analyze-crash"
               type: "enrich_data"
               data:
                 type: "pod_analysis"
                 crash_history: 10
                 resource_usage: true
           stop_on_match: true

Configuration Validation
~~~~~~~~~~~~~~~~~~~~~~~

The configuration is validated on startup:

- **Required fields**: ``name``, ``triggers``, ``actions``
- **Unique names**: Workflow names must be unique
- **Trigger validation**: Each trigger must have at least one valid field
- **Action validation**: Each action must have valid ``name`` and ``type``

Validation errors will prevent cano-collector from starting.

Best Practices
~~~~~~~~~~~~~

1. **Always include create_issue action** - Most workflows should create an Issue to continue processing
2. **Use descriptive names** for workflows and actions
3. **Be specific with triggers** to avoid unintended matches
4. **Test trigger patterns** with wildcards carefully
5. **Order actions logically** - create_issue first, then enrichment actions
6. **Use stop_on_match** judiciously to control workflow execution order
7. **Group related workflows** by naming convention (e.g., ``alert-*``, ``pod-*``)
8. **Document complex trigger patterns** with comments in your values.yaml
9. **Consider workflow performance** - avoid expensive operations in frequently triggered workflows

Integration with Custom Workflows
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Built-in workflows and custom workflows are functionally equivalent and run in the same execution context:

- **Both workflow types** process the same internal event types (`AlertEvent`, `KubernetesEvent`, etc.)
- **Both workflow types** can create Issues through `create_issue` actions
- **Both workflow types** can enrich existing Issues through enrichment actions
- **Both workflow types** run in parallel when their triggers match
- Use different naming conventions to avoid conflicts

**Processing Flow:**
1. External source (Alertmanager, Kubernetes, etc.) sends data → `template.Data` or other external format
2. Data is converted to internal event types (`AlertEvent`, `KubernetesEvent`, etc.)
3. All workflows (built-in and custom) evaluate triggers against internal event types
4. Matching workflows execute in parallel
5. Issues are created and enriched by all matching workflows
6. Team routing receives all created Issues

**Key Differences:**
- **Implementation language** - Built-in workflows use Go, custom workflows use TypeScript
- **Deployment method** - Built-in workflows are compiled, custom workflows are runtime-loaded
- **Development workflow** - Built-in workflows require code changes and rebuilds, custom workflows can be updated via ConfigMaps

For more information about custom workflows, see :doc:`../architecture/workflow/custom`. 