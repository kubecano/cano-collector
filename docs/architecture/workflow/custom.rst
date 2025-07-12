Custom Workflows
================

Custom workflows allow you to extend cano-collector's capabilities by creating tailored automation rules for your specific use cases and requirements.

Overview
--------

Custom workflows provide two main approaches for extending cano-collector:

1. **Contributing to the codebase** - Adding new workflows directly to the cano-collector source code for inclusion in the main distribution
2. **TypeScript-based custom workflows** - Creating organization-specific workflows using TypeScript that are loaded dynamically

TypeScript custom workflows are designed for:
- **Organization-specific needs** - Workflows tailored to your specific environment and requirements
- **Missing functionality** - Workflows that aren't implemented in the standard package but are needed for your use case
- **Local deployment** - Workflows that remain within your cluster and aren't contributed back to the project

Contributing Workflows
----------------------

You can contribute new workflows to the cano-collector project by:

1. **Creating a pull request** with your workflow implementation in Go
2. **Following the contribution guidelines** for code quality and testing
3. **Adding documentation** for your workflow
4. **Including configuration examples** and usage instructions

This approach is ideal for:
- Workflows that would benefit the broader community
- Complex workflows requiring deep integration with cano-collector internals
- Workflows that need to be part of the core distribution

TypeScript Custom Workflows
---------------------------

TypeScript custom workflows provide a flexible way to create organization-specific workflows without modifying the core codebase. These workflows are:

- **Written in TypeScript** with full type safety
- **Loaded dynamically** from mounted volumes
- **Executed securely** using the Deno runtime
- **Configurable** through Helm values
- **Organization-specific** - designed for your specific needs and environment
- **Local deployment** - remain within your cluster and aren't contributed back

### Use Cases for TypeScript Custom Workflows

TypeScript custom workflows are ideal for:

- **Organization-specific alerting** - Custom alert enrichment for your specific applications
- **Integration with internal tools** - Workflows that integrate with your internal monitoring or ticketing systems
- **Custom resource monitoring** - Monitoring of custom resources specific to your environment
- **Business logic automation** - Workflows that implement your specific business rules
- **Missing functionality** - Workflows that aren't available in the standard package but are needed for your use case

### Creating TypeScript Workflows

TypeScript workflows are defined as functions that receive alert data (`template.Data`) and return actions:

.. code-block:: typescript

   interface AlertData {
     alerts: Alert[];
     status: string;
     receiver: string;
     groupLabels: { [key: string]: string };
     commonLabels: { [key: string]: string };
     commonAnnotations: { [key: string]: string };
     externalURL: string;
   }

   interface Alert {
     status: string;
     labels: { [key: string]: string };
     annotations: { [key: string]: string };
     startsAt: string;
     endsAt: string;
     generatorURL: string;
     fingerprint: string;
   }

   interface WorkflowAction {
     type: string;
     data: any;
     priority?: number;
   }

   export function customWorkflow(alertData: AlertData): WorkflowAction[] {
     const actions: WorkflowAction[] = [];
     
     // Custom logic here - processing alert data
     if (alertData.alerts.some(alert => alert.labels.severity === 'critical')) {
       actions.push({
         type: 'create_issue',
         data: {
           title: 'Critical Alert Detected',
           severity: 'high',
           description: 'Custom analysis of critical alert',
           aggregation_key: 'critical_alert_custom'
         },
         priority: 1
       });
     }
     
     return actions;
   }

### Available Action Types

Custom workflows can use various action types for Issue creation and enrichment:

- **`create_issue`** - Create a new Issue from alert data
- **`add_enrichment`** - Add data to an existing Issue
- **`log`** - Add log entries to the Issue
- **`file`** - Create file attachments for the Issue
- **`table`** - Create tabular data for the Issue
- **`markdown`** - Add markdown content to the Issue
- **`graph`** - Create metric graphs for the Issue
- **`callback_block`** - Add interactive buttons to the Issue
- **`modify_severity`** - Modify the Issue's severity level
- **`add_labels`** - Add additional labels to the Issue

**Note:** Custom workflows can create new Issues just like built-in workflows. Both workflow types are functionally equivalent and can use the same action types.

### Configuration

TypeScript workflows are configured through Helm values:

.. code-block:: yaml

   customWorkflows:
     typescript:
       enabled: true
       volume:
         mountPath: "/workflows"
         configMap:
           name: "custom-workflows-config"
       deno:
         enabled: true
         image: "denoland/deno:latest"
         resources:
           requests:
             memory: "256Mi"
             cpu: "200m"
       workflows:
         - name: "custom-alert-processing"
           file: "custom-alert-processing.ts"
           triggers:
             - on_alertmanager_alert:
                 severity: "critical"
                 namespace: "production"

### Deployment

TypeScript workflows are deployed by:

1. **Creating a ConfigMap** with your TypeScript files
2. **Mounting the ConfigMap** as a volume in the cano-collector pod
3. **Configuring the workflow** in Helm values
4. **Restarting cano-collector** to load the new workflows

Example ConfigMap:

.. code-block:: yaml

   apiVersion: v1
   kind: ConfigMap
   metadata:
     name: custom-workflows-config
   data:
     custom-workflow.ts: |
       export function customWorkflow(alertData) {
         // Your workflow logic here - processing alert data
         return [{
           type: 'create_issue',
           data: {
             title: 'Custom Alert Processing',
             severity: 'medium',
             aggregation_key: 'custom_alert'
           }
         }];
       }

### Security Considerations

TypeScript workflows run in a sandboxed environment with:

- **Limited file system access** - Only to specified directories
- **Network restrictions** - Controlled network access
- **Resource limits** - CPU and memory constraints
- **Timeout limits** - Maximum execution time

### Best Practices

When creating custom workflows:

1. **Use TypeScript** for type safety and better development experience
2. **Handle errors gracefully** with proper error handling
3. **Keep workflows focused** on specific use cases
4. **Add logging** for debugging and monitoring
5. **Test thoroughly** before deployment
6. **Document your workflows** with clear descriptions and examples
7. **Keep workflows organization-specific** - don't try to create generic solutions

### Examples

#### Organization-Specific Alert Processing

.. code-block:: typescript

   export function processCriticalAlerts(alertData: AlertData): WorkflowAction[] {
     if (alertData.alerts.some(alert => alert.labels.severity === 'critical')) {
       return [{
         type: 'create_issue',
         data: {
           title: 'Critical Alert Detected',
           severity: 'high',
           description: 'This is a critical alert requiring immediate attention according to our organization\'s procedures.',
           aggregation_key: 'critical_alert_org'
         }
       }];
     }
     return [];
   }

#### Internal Tool Integration

.. code-block:: typescript

   export function createInternalTicket(alertData: AlertData): WorkflowAction[] {
     if (alertData.alerts.some(alert => alert.labels.component === 'custom-resource')) {
       return [{
         type: 'create_issue',
         data: {
           title: 'Internal Ticket Created',
           description: 'Ticket created in our internal system for custom resource',
           severity: 'medium',
           aggregation_key: 'internal_ticket'
         }
       }];
     }
     return [];
   }

#### Business Logic Workflow

.. code-block:: typescript

   export function businessHoursAlert(alertData: AlertData): WorkflowAction[] {
     const now = new Date();
     const hour = now.getHours();
     
     // Only process during business hours (9 AM - 5 PM)
     if (hour >= 9 && hour <= 17 && alertData.alerts.some(alert => alert.labels.severity === 'critical')) {
       return [{
         type: 'create_issue',
         data: {
           title: 'Business Hours Critical Alert',
           description: 'Critical alert escalated during business hours - immediate attention required.',
           severity: 'urgent',
           aggregation_key: 'business_hours_critical'
         }
       }];
     }
     return [];
   }

Integration with Built-in Workflows
-----------------------------------

Custom workflows and built-in workflows are functionally equivalent and run in the same execution context:

**Shared Execution Environment:**
- Both workflow types process the same alert data (`template.Data`)
- Both can create Issues through `create_issue` actions
- Both can enrich existing Issues through enrichment actions
- Both run in parallel when their triggers match

**Key Differences:**
- **Implementation language** - Built-in workflows use Go, custom workflows use TypeScript
- **Deployment method** - Built-in workflows are compiled, custom workflows are runtime-loaded
- **Development workflow** - Built-in workflows require code changes and rebuilds, custom workflows can be updated via ConfigMaps

**Integration Benefits:**
- **Unified interface** - Both workflow types use the same action types and data structures
- **Parallel execution** - Multiple workflows can run simultaneously
- **Flexible deployment** - Choose the right tool for your use case
- **Consistent behavior** - Same capabilities regardless of implementation language

This design provides maximum flexibility while maintaining consistency across all workflow types.

Configuration
-------------

Custom workflows can be configured through Helm values:

.. code-block:: yaml

   customWorkflows:
     typescript:
       enabled: true
       volume:
         mountPath: "/workflows"
         configMap:
           name: "custom-workflows-config"
       deno:
         enabled: true
         image: "denoland/deno:latest"
         resources:
           requests:
             memory: "256Mi"
             cpu: "200m"
         limits:
           memory: "512Mi"
           cpu: "500m"
       security:
         allowNetwork: false
         allowFileSystem: true
         timeout: 30
       workflows:
         - name: "custom-workflow-1"
           file: "workflow1.ts"
           triggers:
             - on_alertmanager_alert:
                 severity: "critical"
                 labels:
                   component: "api"
         - name: "custom-workflow-2"
           file: "workflow2.ts"
           triggers:
             - on_alertmanager_alert:
                 namespace: "production"
                 annotations:
                   "cano.io/enrichment": "pod-analysis" 