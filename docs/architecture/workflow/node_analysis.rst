Node Analysis Workflows
=======================


.. note::
   **Status**: 🔨 Planned - Not yet implemented
Node analysis workflows provide comprehensive insights into Kubernetes node health, performance, and resource utilization.

Node CPU Analysis
-----------------

**Purpose**: Analyzes CPU usage patterns and performance on Kubernetes nodes.

**Trigger**: Fires on node CPU events or alerts

**Actions**:
- Collects CPU usage metrics from nodes
- Analyzes CPU throttling and performance
- Identifies CPU bottlenecks
- Provides CPU optimization recommendations
- Generates CPU usage graphs

**When it runs**: Triggered when node CPU analysis is requested or CPU alerts are received

**Output**: CPU analysis report with performance insights and optimization recommendations

Node Disk Analysis
------------------

**Purpose**: Analyzes disk usage and performance on Kubernetes nodes.

**Trigger**: Fires on node disk events or alerts

**Actions**:
- Monitors disk space usage
- Analyzes disk I/O performance
- Identifies disk bottlenecks
- Provides disk optimization recommendations
- Generates disk usage graphs

**When it runs**: Triggered when node disk analysis is requested or disk alerts are received

**Output**: Disk analysis report with usage patterns and optimization recommendations

Node Memory Analysis
--------------------

**Purpose**: Analyzes memory usage and patterns on Kubernetes nodes.

**Trigger**: Fires on node memory events or alerts

**Actions**:
- Collects memory usage metrics
- Analyzes memory pressure patterns
- Identifies memory bottlenecks
- Provides memory optimization recommendations
- Generates memory usage graphs

**When it runs**: Triggered when node memory analysis is requested or memory alerts are received

**Output**: Memory analysis report with usage patterns and optimization recommendations

Node Status Enricher
--------------------

**Purpose**: Provides comprehensive node status information including conditions and capacity.

**Trigger**: Fires on node status changes or events

**Actions**:
- Collects node status information
- Analyzes node conditions
- Provides capacity and allocatable resource information
- Shows node labels and annotations
- Generates node status reports

**When it runs**: Triggered when node status analysis is requested or node events occur

**Output**: Node status report with comprehensive node information

Node Running Pods Enricher
--------------------------

**Purpose**: Lists and analyzes pods running on specific nodes.

**Trigger**: Fires on node events or pod scheduling events

**Actions**:
- Lists all pods running on the node
- Provides pod resource usage information
- Analyzes pod distribution patterns
- Shows pod status and health
- Generates node pod reports

**When it runs**: Triggered when node pod analysis is requested or pod scheduling events occur

**Output**: Node pod report with running pods and their status

Node Allocatable Resources Enricher
-----------------------------------

**Purpose**: Analyzes allocatable resources on nodes for capacity planning.

**Trigger**: Fires on node resource events or capacity planning requests

**Actions**:
- Collects allocatable resource information
- Analyzes resource allocation patterns
- Provides capacity planning insights
- Shows resource reservation information
- Generates resource allocation reports

**When it runs**: Triggered when node resource analysis is requested or capacity planning is needed

**Output**: Node resource allocation report with capacity insights

Node DMESG Enricher
-------------------

**Purpose**: Captures and analyzes kernel dmesg logs from nodes for system-level issues.

**Trigger**: Fires on node system events or debugging requests

**Actions**:
- Captures kernel dmesg logs
- Filters logs for relevant system events
- Provides system-level debugging information
- Shows kernel-level issues and warnings
- Generates system analysis reports

**When it runs**: Triggered when node system analysis is requested or system events occur

**Output**: Node system analysis report with kernel-level insights

Configuration
-------------

Node analysis workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     nodeAnalysis:
       nodeCpuAnalysis:
         enabled: true
         defaultDuration: "1h"
       nodeDiskAnalysis:
         enabled: true
         defaultDuration: "1h"
       nodeMemoryAnalysis:
         enabled: true
         defaultDuration: "1h"
       nodeStatusEnricher:
         enabled: true
         showDetails: true
       nodeRunningPodsEnricher:
         enabled: true
       nodeAllocatableResourcesEnricher:
         enabled: true
       nodeDmesgEnricher:
         enabled: false 