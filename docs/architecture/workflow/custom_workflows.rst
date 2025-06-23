Custom Workflows
===============

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
---------------------

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
-------------------------

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

TypeScript workflows are defined as functions that receive events and return actions:

.. code-block:: typescript

   interface CanoEvent {
     type: string;
     alert?: Alert;
     pod?: Pod;
     namespace: string;
     timestamp: string;
   }

   interface CanoAction {
     type: string;
     data: any;
     priority?: number;
   }

   export function customWorkflow(event: CanoEvent): CanoAction[] {
     const actions: CanoAction[] = [];
     
     // Custom logic here
     if (event.alert && event.alert.severity === 'critical') {
       actions.push({
         type: 'create_finding',
         data: {
           title: 'Critical Alert Detected',
           severity: 'high',
           description: 'Custom analysis of critical alert'
         },
         priority: 1
       });
     }
     
     return actions;
   }

### Available Action Types

Custom workflows can use various action types:

- **`create_finding`** - Create a new finding/issue
- **`add_enrichment`** - Add data to an existing finding
- **`log`** - Add log entries
- **`file`** - Create file attachments
- **`table`** - Create tabular data
- **`markdown`** - Add markdown content
- **`graph`** - Create metric graphs
- **`callback_block`** - Add interactive buttons

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
         - name: "custom-alert-enrichment"
           file: "custom-alert-enrichment.ts"
           triggers:
             - on_alert:
                 alert_name: "CustomAlert"

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
       export function customWorkflow(event) {
         // Your workflow logic here
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

#### Organization-Specific Alert Enrichment

.. code-block:: typescript

   export function enrichCriticalAlerts(event: CanoEvent): CanoAction[] {
     if (event.alert?.severity === 'critical') {
       return [{
         type: 'add_enrichment',
         data: {
           type: 'markdown',
           content: 'This is a critical alert requiring immediate attention according to our organization\'s procedures.'
         }
       }];
     }
     return [];
   }

#### Internal Tool Integration

.. code-block:: typescript

   export function createInternalTicket(event: CanoEvent): CanoAction[] {
     if (event.type === 'resource_change' && event.resource?.kind === 'CustomResource') {
       return [{
         type: 'create_finding',
         data: {
           title: 'Internal Ticket Created',
           description: `Ticket created in our internal system for resource ${event.resource.name}`,
           severity: 'medium'
         }
       }];
     }
     return [];
   }

#### Business Logic Workflow

.. code-block:: typescript

   export function businessHoursAlert(event: CanoEvent): CanoAction[] {
     const now = new Date();
     const hour = now.getHours();
     
     // Only escalate during business hours (9 AM - 5 PM)
     if (hour >= 9 && hour <= 17) {
       return [{
         type: 'add_enrichment',
         data: {
           type: 'markdown',
           content: 'Alert escalated during business hours.'
         }
       }];
     }
     return [];
   }

Integration with Built-in Workflows
----------------------------------

Custom workflows can integrate with built-in workflows by:

- **Extending built-in functionality** with custom logic
- **Adding custom enrichments** to existing workflows
- **Creating custom triggers** for specific events
- **Providing custom outputs** for different destinations

This allows you to build upon the existing workflow ecosystem while adding your organization-specific requirements.

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
             - on_alert:
                 alert_name: "CustomAlert1"
         - name: "custom-workflow-2"
           file: "workflow2.ts"
           triggers:
             - on_pod_event:
                 type: "custom_pod_analysis" 