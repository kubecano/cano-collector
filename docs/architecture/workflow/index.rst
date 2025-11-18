Workflows
=========

Cano-collector provides a comprehensive set of workflows for Kubernetes monitoring and alerting. These workflows are designed to automatically enrich alerts with relevant context and perform automated actions based on Kubernetes events.

Overview
--------

Workflows in cano-collector are automated processes that:

- **Process incoming alerts** from Alertmanager and other sources
- **Create and enrich Issues** from alert data
- **Perform automated actions** based on specific triggers
- **Provide debugging capabilities** for troubleshooting issues
- **Monitor application health** and performance metrics
- **Support custom logic** through both built-in (Go) and custom (TypeScript) workflows

A workflow in cano-collector consists of three main components:

1. **Trigger** - The condition that activates the workflow (e.g., a Prometheus alert, pod crash, or resource change)
2. **Actions** - The operations performed when the trigger fires (e.g., creating Issues, enriching data, or gathering context)
3. **Output** - The Issues created or enriched, which are then sent to team routing

**Workflow Execution:**
1. Alert data comes into cano-collector and is checked against **all workflow triggers** (both built-in and custom)
2. When there's a match, the workflow runs (regardless of whether it's built-in or custom)
3. Workflow actions execute, creating or enriching Issues
4. All workflows run in the same execution context and can access the same data
5. Issues are sent to team routing

**Key Point:** Built-in and custom workflows are functionally equivalent - they only differ in implementation language (Go vs TypeScript) and deployment method (compiled vs runtime-loaded).

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

Built-in workflows are pre-compiled Go workflows that are configured via YAML and are part of the core cano-collector distribution. These workflows provide comprehensive coverage for common Kubernetes monitoring and troubleshooting scenarios.

**Key Characteristics:**
- **Compiled into the binary** - No runtime loading required
- **YAML-configured** - Defined through configuration files
- **High performance** - Native Go execution
- **Validated at startup** - Configuration errors prevent startup
- **Predictable behavior** - Well-tested and stable

**Configuration:**
Built-in workflows are configured through a ``workflows.yaml`` file that defines:
- **Workflow definitions** with triggers and actions
- **Trigger conditions** (e.g., alert names, severities, namespaces)
- **Action specifications** for Issue creation and enrichment

**Primary Responsibility:**
Both built-in and custom workflows are responsible for **processing alert data and creating/enriching Issues**. The most important action type is ``create_issue``, which transforms ``template.Data`` from Alertmanager into the internal Issue model that can be processed by team routing.

**Available Built-in Workflows:**
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

For configuration details, see :doc:`../../configuration/workflows`.

Custom Workflows
----------------

Custom workflows allow you to extend cano-collector's capabilities by creating tailored automation rules for your specific use cases and requirements. These are TypeScript-based workflows that are loaded dynamically at runtime.

**Key Characteristics:**
- **TypeScript-based** - Written in TypeScript with full type safety
- **Runtime loading** - Loaded dynamically from mounted volumes
- **Organization-specific** - Designed for your specific environment and requirements
- **Flexible deployment** - Can be updated without rebuilding the container
- **Deno runtime** - Executed securely using the Deno JavaScript runtime
- **Issue-focused** - Operate on Issues already created by built-in workflows

**Two main approaches for extending cano-collector:**

1. **Contributing to the codebase** - Adding new built-in workflows directly to the cano-collector source code for inclusion in the main distribution
2. **TypeScript-based custom workflows** - Creating organization-specific workflows using TypeScript that are loaded dynamically

**TypeScript custom workflows are ideal for:**
- **Organization-specific needs** - Workflows tailored to your specific environment and requirements
- **Missing functionality** - Workflows that aren't implemented in the standard package but are needed for your use case
- **Local deployment** - Workflows that remain within your cluster and aren't contributed back to the project
- **Rapid prototyping** - Quick development and testing of workflow ideas
- **Integration with internal systems** - Workflows that connect to your internal tools and APIs

**Built-in vs Custom Workflows:**

+---------------------+------------------+---------------------+
| Aspect              | Built-in         | Custom              |
+=====================+==================+=====================+
| **Language**        | Go               | TypeScript          |
+---------------------+------------------+---------------------+
| **Input Data**      | template.Data    | template.Data       |
+---------------------+------------------+---------------------+
| **Functionality**   | Create/Enrich    | Create/Enrich       |
+---------------------+------------------+---------------------+
| **Performance**     | High             | Moderate            |
+---------------------+------------------+---------------------+
| **Loading**         | Compile-time     | Runtime             |
+---------------------+------------------+---------------------+
| **Configuration**   | YAML             | TypeScript + Helm   |
+---------------------+------------------+---------------------+
| **Deployment**      | Binary rebuild   | Volume mount        |
+---------------------+------------------+---------------------+
| **Use Case**        | Generic          | Organization-specific|
+---------------------+------------------+---------------------+

**Note:** Both workflow types are functionally equivalent - they can both create and enrich Issues. The only differences are implementation language and deployment method.

For detailed information on creating custom workflows, see the :doc:`custom` documentation.

Usage
-----

Workflows are automatically executed when relevant events occur in the cluster. Cano-collector supports running both built-in and custom workflows together:

**Execution Order:**
1. **All workflows** (both built-in and custom) are evaluated against the same alert data (`template.Data`)
2. **Matching workflows** execute in parallel, regardless of type
3. **All workflows** can create and enrich Issues
4. **Issues** are sent to team routing

**Configuration:**
- **Built-in workflows** are configured through ``workflows.yaml`` files
- **Custom workflows** are configured through Helm values and mounted as TypeScript files
- Both types can be enabled/disabled independently

**Best Practices:**
- Use **built-in workflows** for common, well-tested scenarios
- Use **custom workflows** for organization-specific requirements
- Both workflow types can create and enrich Issues
- Use different naming conventions to avoid conflicts
- Test workflows in development environments before production deployment
- Remember that both workflow types operate on the same alert data (`template.Data`)

Built-in Workflows
------------------

.. toctree::
   :maxdepth: 2

   pod_logs_action
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

Implementation Status
~~~~~~~~~~~~~~~~~~~~~

✅ **Implemented**: pod_logs_action

🔨 **In Development/Planned**: All other workflows listed above 