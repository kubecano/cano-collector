Job Workflows
=============

Job workflows provide comprehensive monitoring and management capabilities for Kubernetes jobs, including failure analysis, restart capabilities, and performance monitoring.

Job Failure Reporter
-------------------

**Purpose**: Monitors and reports on failed Kubernetes jobs with detailed analysis.

**Trigger**: Fires on job failure events

**Actions**:
- Detects job failures and captures failure details
- Analyzes job status and completion conditions
- Provides job configuration information
- Shows job events and logs
- Generates failure analysis reports

**When it runs**: Triggered when jobs enter failed state

**Output**: Job failure report with analysis and troubleshooting information

Job Info Enricher
----------------

**Purpose**: Provides comprehensive information about Kubernetes jobs including status and configuration.

**Trigger**: Fires on job events or information requests

**Actions**:
- Collects job status and configuration information
- Shows job completion conditions and deadlines
- Provides job template and spec details
- Analyzes job parallelism and backoff settings
- Generates job information reports

**When it runs**: Triggered when job information is requested or job events occur

**Output**: Job information report with comprehensive job details

Job Events Enricher
------------------

**Purpose**: Captures and analyzes events related to Kubernetes jobs.

**Trigger**: Fires on job events or event analysis requests

**Actions**:
- Collects job-related events from the Kubernetes API
- Filters events by type and severity
- Provides event timeline and analysis
- Shows event details and messages
- Generates job event reports

**When it runs**: Triggered when job event analysis is requested or job events occur

**Output**: Job event report with timeline and analysis

Job Pod Enricher
---------------

**Purpose**: Analyzes pods created by Kubernetes jobs to understand job execution.

**Trigger**: Fires on job pod events or analysis requests

**Actions**:
- Lists all pods created by the job
- Analyzes pod status and completion
- Provides pod logs and events
- Shows pod resource usage
- Generates job pod analysis reports

**When it runs**: Triggered when job pod analysis is requested or job pod events occur

**Output**: Job pod analysis report with pod details and status

Job Restart
----------

**Purpose**: Provides capability to restart failed or completed jobs.

**Trigger**: Fires on job restart requests

**Actions**:
- Validates job restart permissions
- Deletes existing job pods
- Restarts job with original configuration
- Monitors restart progress
- Provides restart status updates

**When it runs**: Triggered when job restart is requested

**Output**: Job restart status and progress information

Job Delete
---------

**Purpose**: Provides capability to delete Kubernetes jobs and their associated resources.

**Trigger**: Fires on job deletion requests

**Actions**:
- Validates job deletion permissions
- Deletes job and associated pods
- Cleans up job-related resources
- Provides deletion confirmation
- Shows deletion status

**When it runs**: Triggered when job deletion is requested

**Output**: Job deletion status and confirmation

Job Performance Analysis
-----------------------

**Purpose**: Analyzes job performance and execution patterns.

**Trigger**: Fires on job completion or performance analysis requests

**Actions**:
- Analyzes job execution time and patterns
- Identifies performance bottlenecks
- Provides optimization recommendations
- Shows resource usage patterns
- Generates performance reports

**When it runs**: Triggered when job performance analysis is requested or jobs complete

**Output**: Job performance report with optimization recommendations

Configuration
-------------

Job workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     jobWorkflows:
       jobFailureReporter:
         enabled: true
         includeLogs: true
         includeEvents: true
       jobInfoEnricher:
         enabled: true
         showDetails: true
       jobEventsEnricher:
         enabled: true
         eventTypes: ["Warning", "Normal"]
       jobPodEnricher:
         enabled: true
         includeLogs: true
       jobRestart:
         enabled: true
         requireConfirmation: true
       jobDelete:
         enabled: true
         requireConfirmation: true
       jobPerformanceAnalysis:
         enabled: true
         defaultDuration: "24h" 