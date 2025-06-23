Alert Enrichment Workflows
==========================

Alert enrichment workflows enhance Prometheus alerts with additional context, making them more actionable and informative.

Default Alert Enricher
----------------------

**Purpose**: Provides basic enrichment for all Prometheus alerts with essential information.

**Trigger**: Fires on any Prometheus alert

**Actions**:
- Adds alert annotations and labels
- Includes alert generator link (if configured)
- Provides basic alert metadata

**When it runs**: Automatically triggered for all Prometheus alerts unless disabled

**Output**: Enhanced alert with basic context and metadata

Graph Enricher
--------------

**Purpose**: Adds resource graphs to alerts showing historical metrics data.

**Trigger**: Fires on Prometheus alerts with resource information

**Actions**:
- Queries Prometheus for historical metrics
- Generates graphs showing resource usage over time
- Supports CPU, memory, disk, and network metrics

**When it runs**: Triggered when alerts contain resource information and graph enrichment is enabled

**Output**: Alert with embedded graphs showing resource trends

Custom Graph Enricher
---------------------

**Purpose**: Adds custom Prometheus queries as graphs to alerts.

**Trigger**: Fires on Prometheus alerts with custom graph configuration

**Actions**:
- Executes custom Prometheus queries
- Generates graphs from query results
- Supports multiple graph types and time ranges

**When it runs**: Triggered when custom graph queries are configured for specific alerts

**Output**: Alert with custom metric graphs

Alert Explanation Enricher
--------------------------

**Purpose**: Adds human-readable explanations and recommended resolutions to alerts.

**Trigger**: Fires on Prometheus alerts with explanation configuration

**Actions**:
- Adds alert explanation text
- Provides recommended resolution steps
- Includes troubleshooting guidance

**When it runs**: Triggered when alert explanations are configured

**Output**: Alert with explanation and resolution guidance

Stack Overflow Enricher
-----------------------

**Purpose**: Automatically searches Stack Overflow for solutions related to the alert.

**Trigger**: Fires on Prometheus alerts

**Actions**:
- Searches Stack Overflow using alert keywords
- Finds relevant solutions and discussions
- Provides links to helpful resources

**When it runs**: Triggered for alerts where Stack Overflow search is enabled

**Output**: Alert with links to relevant Stack Overflow discussions

Template Enricher
-----------------

**Purpose**: Adds custom templated content to alerts using variables.

**Trigger**: Fires on any Kubernetes resource event

**Actions**:
- Processes template strings with event variables
- Supports custom markdown formatting
- Includes dynamic content based on event data

**When it runs**: Triggered when template enrichment is configured

**Output**: Alert with custom templated content

Mention Enricher
----------------

**Purpose**: Adds user mentions to notifications based on alert labels or static configuration.

**Trigger**: Fires on Kubernetes resource events

**Actions**:
- Extracts mention information from labels
- Adds static mentions from configuration
- Formats mentions for different platforms (Slack, Teams, etc.)

**When it runs**: Triggered when mention configuration is present

**Output**: Notification with appropriate user mentions

Severity Silencer
-----------------

**Purpose**: Silences alerts based on severity level to reduce noise.

**Trigger**: Fires on Prometheus alerts

**Actions**:
- Checks alert severity against configured levels
- Stops processing for matching alerts
- Logs silence actions if enabled

**When it runs**: Triggered for all alerts when severity silencing is configured

**Output**: Silenced alerts (no further processing)

Name Silencer
-------------

**Purpose**: Silences specific alerts by name to reduce noise.

**Trigger**: Fires on Prometheus alerts

**Actions**:
- Checks alert name against configured list
- Stops processing for matching alerts
- Logs silence actions

**When it runs**: Triggered for all alerts when name-based silencing is configured

**Output**: Silenced alerts (no further processing)

Configuration
-------------

Alert enrichment workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     alertEnrichment:
       defaultEnricher:
         enabled: true
         alertAnnotationsEnrichment: true
         alertGeneratorLink: true
       graphEnricher:
         enabled: true
         defaultDuration: "1h"
       customGraphEnricher:
         enabled: true
       alertExplanationEnricher:
         enabled: true
       stackOverflowEnricher:
         enabled: false
       templateEnricher:
         enabled: true
       mentionEnricher:
         enabled: true
         staticMentions: []
         mentionsLabel: "mention_users"
       severitySilencer:
         enabled: true
         severity: "none"
       nameSilencer:
         enabled: true
         names: [] 