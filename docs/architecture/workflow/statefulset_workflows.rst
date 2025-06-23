StatefulSet Workflows
=====================

StatefulSet workflows provide comprehensive monitoring and management capabilities for Kubernetes StatefulSets, including status analysis, scaling monitoring, and troubleshooting.

StatefulSet Replicas Enricher
-----------------------------

**Purpose**: Analyzes StatefulSet replica configuration and scaling patterns.

**Trigger**: Fires on StatefulSet events or scaling requests

**Actions**:
- Analyzes current and desired replica counts
- Shows scaling history and patterns
- Identifies scaling issues
- Provides scaling recommendations
- Generates replica analysis reports

**When it runs**: Triggered when StatefulSet replica analysis is requested or scaling events occur

**Output**: StatefulSet replica analysis report with scaling insights

StatefulSet Status Analysis
---------------------------

**Purpose**: Provides detailed analysis of StatefulSet status and health.

**Trigger**: Fires on StatefulSet events or status requests

**Actions**:
- Analyzes StatefulSet status and conditions
- Shows pod status and readiness
- Provides StatefulSet configuration details
- Identifies StatefulSet issues
- Generates status analysis reports

**When it runs**: Triggered when StatefulSet status analysis is requested or StatefulSet events occur

**Output**: StatefulSet status report with detailed analysis

StatefulSet Pod Analysis
------------------------

**Purpose**: Analyzes pods managed by StatefulSets.

**Trigger**: Fires on StatefulSet pod events

**Actions**:
- Lists all pods managed by the StatefulSet
- Analyzes pod status and readiness
- Provides pod resource usage information
- Shows pod configuration details
- Generates pod analysis reports

**When it runs**: Triggered when StatefulSet pod analysis is requested or pod events occur

**Output**: StatefulSet pod analysis report with pod details

StatefulSet Scaling Monitor
---------------------------

**Purpose**: Monitors StatefulSet scaling operations and provides progress updates.

**Trigger**: Fires on StatefulSet scaling events

**Actions**:
- Monitors scaling progress in real-time
- Shows pod status during scaling
- Identifies scaling issues and delays
- Provides scaling timeline
- Generates scaling progress reports

**When it runs**: Triggered when StatefulSet scaling operations are initiated or in progress

**Output**: StatefulSet scaling progress report with status updates

StatefulSet Health Analysis
---------------------------

**Purpose**: Analyzes StatefulSet health and identifies potential issues.

**Trigger**: Fires on StatefulSet health events or analysis requests

**Actions**:
- Analyzes StatefulSet health status
- Identifies health issues and warnings
- Provides health recommendations
- Shows health metrics and trends
- Generates health analysis reports

**When it runs**: Triggered when StatefulSet health analysis is requested or health events occur

**Output**: StatefulSet health analysis report with recommendations

Configuration
-------------

StatefulSet workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     statefulsetWorkflows:
       statefulsetReplicasEnricher:
         enabled: true
         includeHistory: true
       statefulsetStatusAnalysis:
         enabled: true
         showDetails: true
       statefulsetPodAnalysis:
         enabled: true
         includeResourceInfo: true
       statefulsetScalingMonitor:
         enabled: true
         monitorProgress: true
       statefulsetHealthAnalysis:
         enabled: true
         includeMetrics: true 