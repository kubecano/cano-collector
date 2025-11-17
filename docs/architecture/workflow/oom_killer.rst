OOM Killer Workflows
====================

.. note::
   **Status**: 🔨 Planned - Not yet implemented

OOM (Out of Memory) killer workflows analyze and provide insights into memory-related issues in Kubernetes clusters, helping identify the root causes of OOM events.

Pod OOM Killer Enricher
-----------------------

**Purpose**: Analyzes pods that have been killed by the OOM killer to understand memory usage patterns and resource allocation.

**Trigger**: Fires on pod events where containers have been OOM killed

**Actions**:
- Identifies OOM killed containers in the pod
- Analyzes memory requests and limits
- Provides container memory usage statistics
- Shows container start and finish times
- Generates memory usage graphs

**When it runs**: Triggered when pods are detected with OOM killed containers

**Output**: Detailed analysis of OOM killed containers with memory usage data and recommendations

OOM Killer Enricher
-------------------

**Purpose**: Analyzes node-level OOM killer events to understand cluster-wide memory pressure.

**Trigger**: Fires on node OOM killer events or alerts

**Actions**:
- Identifies all pods killed by OOM killer on the node
- Analyzes node memory usage patterns
- Provides memory pressure analysis
- Shows historical memory usage trends
- Generates node memory graphs

**When it runs**: Triggered when node OOM killer events are detected

**Output**: Node-level OOM analysis with affected pods and memory pressure insights

OOM Killed Container Graph Enricher
-----------------------------------

**Purpose**: Generates detailed memory usage graphs for containers that were killed by the OOM killer.

**Trigger**: Fires on pod events with OOM killed containers

**Actions**:
- Creates memory usage graphs for OOM killed containers
- Shows memory usage over time
- Displays memory limits and requests
- Provides memory pressure indicators
- Supports different time ranges

**When it runs**: Triggered when OOM killed containers are detected and graph generation is enabled

**Output**: Memory usage graphs showing container memory patterns before OOM kill

Memory Analysis
---------------

**Purpose**: Provides comprehensive memory analysis for OOM events including root cause identification.

**Trigger**: Fires on OOM events with analysis configuration

**Actions**:
- Analyzes memory allocation patterns
- Identifies memory leaks and excessive usage
- Provides memory optimization recommendations
- Shows memory usage by container and node
- Generates memory pressure reports

**When it runs**: Triggered when detailed memory analysis is requested for OOM events

**Output**: Comprehensive memory analysis with optimization recommendations

DMESG Log Enricher
------------------

**Purpose**: Captures and analyzes kernel dmesg logs related to OOM events for deeper investigation.

**Trigger**: Fires on OOM events when dmesg analysis is enabled

**Actions**:
- Captures kernel dmesg logs from the node
- Filters logs for OOM-related entries
- Provides kernel-level memory pressure information
- Shows system memory allocation details

**When it runs**: Triggered when dmesg log analysis is enabled for OOM events

**Output**: Kernel-level memory analysis from dmesg logs

Configuration
-------------

OOM killer workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     oomKiller:
       podOomKillerEnricher:
         enabled: true
         attachLogs: true
         containerMemoryGraph: true
         nodeMemoryGraph: true
         dmesgLog: false
       oomKillerEnricher:
         enabled: true
         newOomKillsDurationInSec: 1200
         metricsDurationInSecs: 1200
       oomKilledContainerGraphEnricher:
         enabled: true
         delayGraphS: 0
       memoryAnalysis:
         enabled: true
       dmesgLogEnricher:
         enabled: false 