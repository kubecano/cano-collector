Workflows
=========

Cano-collector provides a comprehensive set of workflows for Kubernetes monitoring and alerting. These workflows are designed to automatically enrich alerts with relevant context and perform automated actions based on Kubernetes events.

Built-in Workflows
-----------------

.. toctree::
   :maxdepth: 2

   alert_enrichment
   custom_workflows
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
   statefulset_workflows

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