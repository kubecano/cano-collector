Deployment Workflows
====================

Deployment workflows provide comprehensive monitoring and management capabilities for Kubernetes deployments, including status analysis, rollout monitoring, and troubleshooting.

Deployment Status Enricher
--------------------------

**Purpose**: Provides detailed analysis of deployment status and rollout progress.

**Trigger**: Fires on deployment events or status requests

**Actions**:
- Analyzes deployment status and conditions
- Shows rollout progress and history
- Provides replica set information
- Identifies deployment issues
- Generates status reports

**When it runs**: Triggered when deployment status analysis is requested or deployment events occur

**Output**: Deployment status report with detailed analysis

Deployment Events Enricher
--------------------------

**Purpose**: Captures and analyzes events related to Kubernetes deployments.

**Trigger**: Fires on deployment events or event analysis requests

**Actions**:
- Collects deployment-related events
- Filters events by type and severity
- Provides event timeline and analysis
- Shows event details and messages
- Generates deployment event reports

**When it runs**: Triggered when deployment event analysis is requested or deployment events occur

**Output**: Deployment event report with timeline and analysis

Deployment Replicas Analysis
----------------------------

**Purpose**: Analyzes deployment replica configuration and scaling patterns.

**Trigger**: Fires on deployment scaling events or analysis requests

**Actions**:
- Analyzes current and desired replica counts
- Shows scaling history and patterns
- Identifies scaling issues
- Provides scaling recommendations
- Generates replica analysis reports

**When it runs**: Triggered when deployment replica analysis is requested or scaling events occur

**Output**: Deployment replica analysis report with scaling insights

Deployment Rollout Monitor
--------------------------

**Purpose**: Monitors deployment rollouts and provides progress updates.

**Trigger**: Fires on deployment rollout events

**Actions**:
- Monitors rollout progress in real-time
- Shows pod status during rollout
- Identifies rollout issues and delays
- Provides rollout timeline
- Generates rollout progress reports

**When it runs**: Triggered when deployment rollouts are initiated or in progress

**Output**: Deployment rollout progress report with status updates

Deployment Rollback
-------------------

**Purpose**: Provides capability to rollback deployments to previous versions.

**Trigger**: Fires on deployment rollback requests

**Actions**:
- Validates rollback permissions
- Shows available rollback versions
- Executes deployment rollback
- Monitors rollback progress
- Provides rollback status updates

**When it runs**: Triggered when deployment rollback is requested

**Output**: Deployment rollback status and progress information

Deployment Restart
------------------

**Purpose**: Provides capability to restart deployments by updating annotations.

**Trigger**: Fires on deployment restart requests

**Actions**:
- Validates restart permissions
- Updates deployment restart annotation
- Triggers deployment restart
- Monitors restart progress
- Provides restart status updates

**When it runs**: Triggered when deployment restart is requested

**Output**: Deployment restart status and progress information

Deployment Health Analysis
--------------------------

**Purpose**: Analyzes deployment health and identifies potential issues.

**Trigger**: Fires on deployment health events or analysis requests

**Actions**:
- Analyzes deployment health status
- Identifies health issues and warnings
- Provides health recommendations
- Shows health metrics and trends
- Generates health analysis reports

**When it runs**: Triggered when deployment health analysis is requested or health events occur

**Output**: Deployment health analysis report with recommendations

Configuration
-------------

Deployment workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     deploymentWorkflows:
       deploymentStatusEnricher:
         enabled: true
         showDetails: true
       deploymentEventsEnricher:
         enabled: true
         eventTypes: ["Warning", "Normal"]
         dependentPodMode: true
       deploymentReplicasAnalysis:
         enabled: true
         includeHistory: true
       deploymentRolloutMonitor:
         enabled: true
         monitorProgress: true
       deploymentRollback:
         enabled: true
         requireConfirmation: true
       deploymentRestart:
         enabled: true
         requireConfirmation: true
       deploymentHealthAnalysis:
         enabled: true
         includeMetrics: true 