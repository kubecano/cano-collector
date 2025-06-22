Workflows
=========

Cano-collector provides a comprehensive set of workflows for Kubernetes monitoring and alerting. These workflows are designed to automatically enrich alerts with relevant context and perform automated actions based on Kubernetes events.

Built-in Workflows
-----------------

.. toctree::
   :maxdepth: 2

   workflows/alert_enrichment
   workflows/custom_workflows
   workflows/daemonset_workflows
   workflows/deployment_workflows
   workflows/event_enrichments
   workflows/golang_workflows
   workflows/java_workflows
   workflows/job_workflows
   workflows/monitoring_tracking
   workflows/node_analysis
   workflows/oom_killer
   workflows/persistent_volume_workflows
   workflows/pod_enrichments
   workflows/pod_troubleshooting
   workflows/prometheus_enrichments
   workflows/python_workflows
   workflows/statefulset_workflows

Overview
--------

Workflows in cano-collector are automated processes that:

- **Enrich alerts** with contextual information from Kubernetes resources
- **Perform automated actions** based on specific triggers
- **Provide debugging capabilities** for troubleshooting issues
- **Monitor application health** and performance metrics
- **Support custom logic** through TypeScript-based custom workflows

Key Features
-----------

- **Event-driven**: Workflows are triggered by Kubernetes events and alerts
- **Context-aware**: Automatically gather relevant information from the cluster
- **Extensible**: Support for custom workflows written in TypeScript
- **Configurable**: Each workflow can be enabled/disabled and configured via Helm
- **Multi-resource**: Support for pods, deployments, statefulsets, daemonsets, and more
- **Language-specific**: Specialized workflows for Java, Go, and Python applications

Usage
-----

Workflows are automatically executed when relevant events occur in the cluster. They can be configured through Helm values and custom workflows can be added by mounting TypeScript files as volumes.

For more information about specific workflows, see the individual documentation pages above.

Workflows are the core automation mechanism in cano-collector that define how the system responds to various events in your Kubernetes cluster. They provide a powerful and flexible way to automate monitoring, troubleshooting, and remediation tasks.

Overview
--------

A workflow in cano-collector consists of three main components:

1. **Trigger** - The condition that activates the workflow (e.g., a Prometheus alert, pod crash, or resource change)
2. **Actions** - The operations performed when the trigger fires (e.g., enriching data, creating findings, or sending notifications)
3. **Output** - Where the results are sent (e.g., Slack, Teams, or other destinations)

Workflows follow a pipeline pattern:

1. Events come into cano-collector and are checked against triggers
2. When there's a match, the trigger fires
3. The relevant workflow runs
4. All workflow actions execute, receiving the event as context
5. Results are sent to configured destinations

Key Features
-----------

- **Event-driven automation** - Respond to Kubernetes events, Prometheus alerts, and custom triggers
- **Rich data enrichment** - Add context, logs, metrics, and analysis to alerts and events
- **Flexible output routing** - Send results to multiple destinations (Slack, Teams, PagerDuty, etc.)
- **Custom workflows** - Create tailored workflows using TypeScript
- **Built-in workflows** - Comprehensive set of pre-built workflows for common scenarios
- **Configuration management** - Easy configuration through Helm values

Built-in Workflows
-----------------

cano-collector includes a comprehensive set of built-in workflows organized into categories:

.. toctree::
   :maxdepth: 1

   workflows/alert_enrichment
   workflows/pod_troubleshooting
   workflows/pod_enrichments
   workflows/oom_killer
   workflows/resource_analysis
   workflows/node_analysis
   workflows/job_workflows
   workflows/deployment_workflows
   workflows/statefulset_workflows
   workflows/daemonset_workflows
   workflows/persistent_volume_workflows
   workflows/event_enrichments
   workflows/prometheus_enrichments
   workflows/monitoring_tracking

### Alert Enrichment Workflows

Enhance Prometheus alerts with additional context and information:

- **Default Alert Enricher** - Basic enrichment for all alerts
- **Graph Enricher** - Add resource usage graphs
- **Alert Explanation Enricher** - Add human-readable explanations
- **Stack Overflow Enricher** - Search for solutions automatically
- **Template Enricher** - Add custom templated content
- **Mention Enricher** - Add user mentions to notifications
- **Silencers** - Reduce noise by silencing specific alerts

### Pod Troubleshooting Workflows

Deep diagnostic capabilities for investigating pod issues:

- **Python Profiler** - Performance analysis for Python applications
- **Pod Process List** - List all processes in a pod
- **Python Memory Analysis** - Memory leak detection for Python apps
- **Debugger Stack Trace** - Capture stack traces for debugging
- **Python Process Inspector** - Comprehensive Python process analysis
- **Python Debugger** - Interactive debugging capabilities

### Pod Enrichment Workflows

Comprehensive analysis and enrichment for Kubernetes pods:

- **Pod Investigator Enricher** - Comprehensive pod investigation
- **Pod Enrichments** - Basic pod information and status
- **Pod Evicted Enrichments** - Analysis of evicted pods
- **Image Pull Backoff Enricher** - Analysis of image pull issues
- **Restart Loop Reporter** - Monitoring of restart loops
- **Pod Actions** - Basic pod management capabilities

### OOM Killer Workflows

Analyze memory-related issues and OOM events:

- **Pod OOM Killer Enricher** - Analyze OOM killed containers
- **OOM Killer Enricher** - Node-level OOM analysis
- **OOM Killed Container Graph Enricher** - Memory usage graphs
- **Memory Analysis** - Comprehensive memory analysis
- **DMESG Log Enricher** - Kernel-level memory analysis

### Resource Analysis Workflows

Comprehensive cluster resource analysis and optimization:

- **KRR (Kubernetes Resource Recommender) Scan** - Resource optimization recommendations
- **Popeye Scan** - Cluster health analysis
- **Resource Usage Analysis** - Usage pattern analysis
- **Capacity Planning** - Scaling and capacity insights
- **Resource Optimization** - Specific optimization recommendations

### Node Analysis Workflows

Monitor and analyze Kubernetes node health and performance:

- **Node CPU Analysis** - CPU usage and performance analysis
- **Node Disk Analysis** - Disk usage and I/O analysis
- **Node Memory Analysis** - Memory usage and pressure analysis
- **Node Status Enricher** - Comprehensive node status information
- **Node Running Pods Enricher** - Pod distribution analysis
- **Node Allocatable Resources Enricher** - Resource allocation analysis
- **Node DMESG Enricher** - Kernel-level system analysis

### Job Workflows

Monitor and manage Kubernetes jobs:

- **Job Failure Reporter** - Detailed failure analysis
- **Job Info Enricher** - Comprehensive job information
- **Job Events Enricher** - Job event analysis
- **Job Pod Enricher** - Pod analysis for jobs
- **Job Restart** - Restart failed jobs
- **Job Delete** - Delete jobs and cleanup
- **Job Performance Analysis** - Performance optimization

### Deployment Workflows

Monitor and manage Kubernetes deployments:

- **Deployment Status Enricher** - Status and rollout analysis
- **Deployment Events Enricher** - Event analysis
- **Deployment Replicas Analysis** - Scaling pattern analysis
- **Deployment Rollout Monitor** - Real-time rollout monitoring
- **Deployment Rollback** - Rollback capabilities
- **Deployment Restart** - Restart deployments
- **Deployment Health Analysis** - Health monitoring

### StatefulSet Workflows

Monitor and manage Kubernetes StatefulSets:

- **StatefulSet Replicas Enricher** - Replica configuration analysis
- **StatefulSet Status Analysis** - Status and health analysis
- **StatefulSet Pod Analysis** - Pod analysis for StatefulSets
- **StatefulSet Scaling Monitor** - Real-time scaling monitoring
- **StatefulSet Health Analysis** - Health monitoring

### DaemonSet Workflows

Monitor and manage Kubernetes DaemonSets:

- **DaemonSet Status Enricher** - Status and scheduling analysis
- **DaemonSet Misscheduled Analysis** - Misscheduling analysis
- **DaemonSet Pod Analysis** - Pod analysis for DaemonSets
- **DaemonSet Rollout Monitor** - Real-time rollout monitoring
- **DaemonSet Health Analysis** - Health monitoring

### Persistent Volume Workflows

Monitor and manage persistent volumes and PVCs:

- **Persistent Volume Actions** - Volume management capabilities
- **PVC Snapshots** - Snapshot creation and management
- **Persistent Data Analysis** - Data usage analysis
- **Volume Backup Management** - Backup operations
- **Volume Performance Analysis** - Performance metrics analysis

### Event Enrichment Workflows

Comprehensive analysis of Kubernetes events:

- **Event Enrichments** - Basic event analysis
- **Resource Events Enricher** - Resource-related event analysis
- **Event Resource Events** - Resource-specific event analysis
- **Event Timeline Analysis** - Timeline correlation analysis
- **Event Pattern Recognition** - Pattern identification

### Prometheus Enrichment Workflows

Advanced Prometheus metric enrichment:

- **Prometheus Enrichments** - Comprehensive metric enrichment
- **Prometheus Simulation** - Query simulation and testing
- **Target Down Enrichment** - Target down analysis
- **CPU Throttling Analysis** - CPU throttling analysis
- **Overcommit Enrichments** - Resource overcommit analysis

### Monitoring and Tracking Workflows

Track changes and maintain audit trails:

- **Resource Babysitter** - Change tracking and analysis
- **Resource Events Diff** - Detailed change diffs
- **Change Tracking** - Comprehensive audit trails
- **Resource Monitoring** - Continuous health monitoring
- **Event Tracking** - Event pattern analysis
- **Audit Logging** - Compliance and security logging
- **Configuration Tracking** - Configuration drift analysis

Custom Workflows
---------------

In addition to built-in workflows, you can create custom workflows tailored to your specific needs:

.. toctree::
   :maxdepth: 1

   workflows/custom_workflows

Custom workflows provide two approaches:

1. **Contributing to the codebase** - Add workflows directly to cano-collector
2. **TypeScript-based workflows** - Create workflows using TypeScript loaded dynamically

TypeScript custom workflows offer:
- **Type safety** with full TypeScript support
- **Dynamic loading** from mounted volumes
- **Secure execution** using Deno runtime
- **Easy configuration** through Helm values

Architecture
-----------

Workflows in cano-collector follow a modular architecture:

.. image:: /images/workflow-architecture.png
   :alt: Workflow Architecture
   :align: center

### Components

- **Event Sources** - Kubernetes API, Prometheus, custom events
- **Trigger Engine** - Matches events to workflow triggers
- **Action Engine** - Executes workflow actions
- **Output Engine** - Routes results to destinations
- **Custom Workflow Engine** - Executes TypeScript workflows

### Event Flow

1. **Event Reception** - Events are received from various sources
2. **Trigger Matching** - Events are matched against workflow triggers
3. **Workflow Execution** - Matching workflows are executed
4. **Action Processing** - Workflow actions are processed
5. **Output Generation** - Results are generated and formatted
6. **Destination Routing** - Outputs are sent to configured destinations

Configuration
-------------

Workflows are configured through Helm values, allowing you to:

- Enable or disable specific workflows
- Customize workflow parameters
- Route workflow outputs to different destinations
- Create custom workflows using TypeScript

Example configuration:

.. code-block:: yaml

   workflows:
     alertEnrichment:
       defaultEnricher:
         enabled: true
         alertAnnotationsEnrichment: true
       graphEnricher:
         enabled: true
         defaultDuration: "1h"
     
     podTroubleshooting:
       pythonProfiler:
         enabled: true
         defaultDuration: 2
     
     oomKiller:
       podOomKillerEnricher:
         enabled: true
         attachLogs: true
         containerMemoryGraph: true

   customWorkflows:
     typescript:
       enabled: true
       volume:
         mountPath: "/workflows"
         configMap:
           name: "custom-workflows-config"

Best Practices
-------------

When working with workflows:

1. **Start with built-in workflows** - Use existing workflows before creating custom ones
2. **Configure appropriately** - Adjust parameters for your environment
3. **Monitor performance** - Watch for workflow execution times and resource usage
4. **Test thoroughly** - Test workflows in non-production environments first
5. **Document custom workflows** - Document any custom workflows you create
6. **Use TypeScript for custom workflows** - Leverage type safety and better tooling

Next Steps
----------

- :ref:`Explore built-in workflows <workflows-index>`
- :ref:`Learn about custom workflows <workflows-custom_workflows>`
- :ref:`Configure workflows <workflow-configuration>`
- :ref:`Deploy cano-collector <installation>` 