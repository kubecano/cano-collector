Event Enrichment Workflows
=========================

Event enrichment workflows provide comprehensive analysis and enrichment capabilities for Kubernetes events, helping to understand the context and impact of various cluster events.

Event Enrichments
----------------

**Purpose**: Provides comprehensive analysis of Kubernetes events.

**Trigger**: Fires on Kubernetes events

**Actions**:
- Collects and analyzes Kubernetes events
- Provides event timeline and correlation
- Shows event details and messages
- Identifies event patterns and trends
- Generates event analysis reports

**When it runs**: Triggered when Kubernetes events occur

**Output**: Event analysis report with timeline and correlation

Resource Events Enricher
-----------------------

**Purpose**: Enriches events with related resource information.

**Trigger**: Fires on resource-related events

**Actions**:
- Links events to related resources
- Provides resource status information
- Shows resource configuration details
- Analyzes event impact on resources
- Generates resource event reports

**When it runs**: Triggered when resource-related events occur

**Output**: Resource event report with related resource information

Event Resource Events
--------------------

**Purpose**: Provides comprehensive event analysis for specific resources.

**Trigger**: Fires on resource events

**Actions**:
- Analyzes events for specific resources
- Shows event history and patterns
- Provides resource status correlation
- Identifies event impact
- Generates resource-specific event reports

**When it runs**: Triggered when events occur for monitored resources

**Output**: Resource-specific event analysis report

Event Timeline Analysis
----------------------

**Purpose**: Provides timeline analysis of related events.

**Trigger**: Fires on event timeline requests

**Actions**:
- Creates event timelines
- Correlates related events
- Shows event sequences
- Identifies event patterns
- Generates timeline reports

**When it runs**: Triggered when event timeline analysis is requested

**Output**: Event timeline report with correlation analysis

Event Pattern Recognition
------------------------

**Purpose**: Identifies patterns in Kubernetes events.

**Trigger**: Fires on event pattern analysis requests

**Actions**:
- Analyzes event patterns
- Identifies recurring issues
- Provides pattern recognition
- Shows event correlations
- Generates pattern analysis reports

**When it runs**: Triggered when event pattern analysis is requested

**Output**: Event pattern analysis report with insights

Configuration
-------------

Event enrichment workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     eventEnrichments:
       eventEnrichments:
         enabled: true
         includeDetails: true
       resourceEventsEnricher:
         enabled: true
         includeResourceInfo: true
       eventResourceEvents:
         enabled: true
         includeTimeline: true
       eventTimelineAnalysis:
         enabled: true
         timelineDuration: "1h"
       eventPatternRecognition:
         enabled: true
         patternThreshold: 3 