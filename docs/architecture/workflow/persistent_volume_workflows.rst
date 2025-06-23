Persistent Volume Workflows
===========================

Persistent volume workflows provide comprehensive monitoring and management capabilities for Kubernetes persistent volumes and persistent volume claims, including backup, analysis, and troubleshooting.

Persistent Volume Actions
-------------------------

**Purpose**: Provides management capabilities for persistent volumes.

**Trigger**: Fires on persistent volume events or action requests

**Actions**:
- Analyzes persistent volume status
- Provides volume capacity information
- Shows volume access modes
- Identifies volume issues
- Generates volume analysis reports

**When it runs**: Triggered when persistent volume analysis is requested or volume events occur

**Output**: Persistent volume analysis report with status and configuration details

PVC Snapshots
-------------

**Purpose**: Creates and manages snapshots of persistent volume claims.

**Trigger**: Fires on PVC snapshot requests

**Actions**:
- Creates PVC snapshots
- Manages snapshot lifecycle
- Provides snapshot status information
- Shows snapshot history
- Generates snapshot reports

**When it runs**: Triggered when PVC snapshot operations are requested

**Output**: PVC snapshot report with status and management information

Persistent Data Analysis
------------------------

**Purpose**: Analyzes persistent data usage and patterns.

**Trigger**: Fires on persistent data analysis requests

**Actions**:
- Analyzes data usage patterns
- Provides storage capacity information
- Shows data growth trends
- Identifies storage issues
- Generates data analysis reports

**When it runs**: Triggered when persistent data analysis is requested

**Output**: Persistent data analysis report with usage insights

Volume Backup Management
------------------------

**Purpose**: Manages backup operations for persistent volumes.

**Trigger**: Fires on backup requests or scheduled backups

**Actions**:
- Creates volume backups
- Manages backup schedules
- Provides backup status information
- Shows backup history
- Generates backup reports

**When it runs**: Triggered when backup operations are requested or scheduled

**Output**: Volume backup report with status and management information

Volume Performance Analysis
---------------------------

**Purpose**: Analyzes persistent volume performance metrics.

**Trigger**: Fires on volume performance analysis requests

**Actions**:
- Analyzes I/O performance metrics
- Provides throughput information
- Shows latency patterns
- Identifies performance bottlenecks
- Generates performance analysis reports

**When it runs**: Triggered when volume performance analysis is requested

**Output**: Volume performance analysis report with optimization recommendations

Configuration
-------------

Persistent volume workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     persistentVolumeWorkflows:
       persistentVolumeActions:
         enabled: true
         includeDetails: true
       pvcSnapshots:
         enabled: true
         retentionDays: 30
       persistentDataAnalysis:
         enabled: true
         includeMetrics: true
       volumeBackupManagement:
         enabled: true
         backupSchedule: "0 2 * * *"
       volumePerformanceAnalysis:
         enabled: true
         includeIOMetrics: true 