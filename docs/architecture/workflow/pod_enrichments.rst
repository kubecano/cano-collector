Pod Enrichment Workflows
========================

Pod enrichment workflows provide comprehensive analysis and enrichment capabilities for Kubernetes pods, including status analysis, resource monitoring, and troubleshooting.

Pod Investigator Enricher
-------------------------

**Purpose**: Provides comprehensive investigation capabilities for pods with issues.

**Trigger**: Fires on pod events or investigation requests

**Actions**:
- Analyzes pod status and conditions
- Collects pod events and logs
- Provides resource usage information
- Shows pod configuration details
- Generates comprehensive investigation reports

**When it runs**: Triggered when pod investigation is requested or pod issues are detected

**Output**: Comprehensive pod investigation report with analysis and recommendations

Pod Enrichments
---------------

**Purpose**: Provides basic pod information and status analysis.

**Trigger**: Fires on pod events or enrichment requests

**Actions**:
- Collects pod status information
- Shows pod conditions and readiness
- Provides resource requests and limits
- Analyzes pod configuration
- Generates pod information reports

**When it runs**: Triggered when pod enrichment is requested or pod events occur

**Output**: Pod information report with status and configuration details

Pod Evicted Enrichments
-----------------------

**Purpose**: Analyzes pods that have been evicted from nodes.

**Trigger**: Fires on pod eviction events

**Actions**:
- Identifies evicted pods and reasons
- Analyzes eviction patterns
- Provides node resource information
- Shows eviction history
- Generates eviction analysis reports

**When it runs**: Triggered when pods are evicted from nodes

**Output**: Pod eviction analysis report with root cause information

Image Pull Backoff Enricher
---------------------------

**Purpose**: Analyzes pods with image pull backoff issues.

**Trigger**: Fires on image pull backoff events

**Actions**:
- Identifies image pull backoff issues
- Analyzes image pull errors
- Provides image registry information
- Shows pull attempt history
- Generates image pull analysis reports

**When it runs**: Triggered when image pull backoff issues are detected

**Output**: Image pull backoff analysis report with troubleshooting information

Restart Loop Reporter
---------------------

**Purpose**: Monitors and reports on pods in restart loops.

**Trigger**: Fires on pod restart loop events

**Actions**:
- Detects restart loop patterns
- Analyzes restart reasons
- Provides restart history
- Shows container logs
- Generates restart loop reports

**When it runs**: Triggered when pods enter restart loops

**Output**: Restart loop analysis report with troubleshooting guidance

Pod Actions
-----------

**Purpose**: Provides basic pod management actions.

**Trigger**: Fires on pod action requests

**Actions**:
- Pod deletion capabilities
- Pod restart functionality
- Pod status updates
- Pod configuration changes
- Pod management operations

**When it runs**: Triggered when pod management actions are requested

**Output**: Pod action status and confirmation

Configuration
-------------

Pod enrichment workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     podEnrichments:
       podInvestigatorEnricher:
         enabled: true
         includeLogs: true
         includeEvents: true
       podEnrichments:
         enabled: true
         showDetails: true
       podEvictedEnrichments:
         enabled: true
         includeNodeInfo: true
       imagePullBackoffEnricher:
         enabled: true
         includeRegistryInfo: true
       restartLoopReporter:
         enabled: true
         includeLogs: true
       podActions:
         enabled: true
         requireConfirmation: true 