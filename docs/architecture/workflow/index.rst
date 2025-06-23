Workflows
=========

Cano-collector provides a comprehensive set of workflows for Kubernetes monitoring and alerting. These workflows are designed to automatically enrich alerts with relevant context and perform automated actions based on Kubernetes events.

Overview
--------

Workflows in cano-collector are automated processes that:

- **Enrich alerts** with contextual information from Kubernetes resources
- **Perform automated actions** based on specific triggers
- **Provide debugging capabilities** for troubleshooting issues
- **Monitor application health** and performance metrics
- **Support custom logic** through TypeScript-based custom workflows

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
------------

- **Event-driven**: Workflows are triggered by Kubernetes events and alerts
- **Context-aware**: Automatically gather relevant information from the cluster
- **Extensible**: Support for custom workflows written in TypeScript
- **Configurable**: Each workflow can be enabled/disabled and configured via Helm
- **Multi-resource**: Support for pods, deployments, statefulsets, daemonsets, and more
- **Language-specific**: Specialized workflows for Java, Go, and Python applications

Built-in Workflows
------------------

Built-in workflows are pre-configured automation rules that come with cano-collector. These workflows provide comprehensive coverage for common Kubernetes monitoring and troubleshooting scenarios:

- **Alert Enrichment** - Enhance Prometheus alerts with additional context and information
- **Pod Troubleshooting** - Deep diagnostic capabilities for investigating pod issues
- **Resource Analysis** - Comprehensive cluster resource analysis and optimization
- **Node Analysis** - Monitor and analyze Kubernetes node health and performance
- **Job Workflows** - Monitor and manage Kubernetes jobs
- **Deployment Workflows** - Monitor and manage Kubernetes deployments
- **StatefulSet Workflows** - Monitor and manage Kubernetes StatefulSets
- **DaemonSet Workflows** - Monitor and manage Kubernetes DaemonSets
- **Persistent Volume Workflows** - Monitor and manage persistent volumes
- **Event Enrichments** - Enrich various Kubernetes events with additional context
- **Prometheus Enrichments** - Enhance Prometheus metrics and alerts
- **Monitoring Tracking** - Track and monitor application health and performance
- **Language-specific Workflows** - Specialized workflows for Java, Go, and Python applications

Built-in workflows are automatically available and can be enabled/disabled through Helm configuration.

Custom Workflows
----------------

Custom workflows allow you to extend cano-collector's capabilities by creating tailored automation rules for your specific use cases and requirements. There are two main approaches:

1. **Contributing to the codebase** - Adding new workflows directly to the cano-collector source code for inclusion in the main distribution
2. **TypeScript-based custom workflows** - Creating organization-specific workflows using TypeScript that are loaded dynamically

TypeScript custom workflows are ideal for:
- **Organization-specific needs** - Workflows tailored to your specific environment and requirements
- **Missing functionality** - Workflows that aren't implemented in the standard package but are needed for your use case
- **Local deployment** - Workflows that remain within your cluster and aren't contributed back to the project

For detailed information on creating custom workflows, see the :doc:`custom` documentation.

Usage
-----

Workflows are automatically executed when relevant events occur in the cluster. They can be configured through Helm values and custom workflows can be added by mounting TypeScript files as volumes.

Built-in Workflows
------------------

.. toctree::
   :maxdepth: 2

   alert_enrichment
   custom
   daemonset_workflows
   deployment_workflows
   event_enrichments
   golang_workflows
   java_workflows
   job_workflows
   monitoring_tracking
   node_analysis
   oom_killer
   persistent_volume_workflows
   pod_enrichments
   pod_troubleshooting
   prometheus_enrichments
   python_workflows
   resource_analysis
   statefulset_workflows 