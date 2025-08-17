Pod Logs Workflow Action
========================

Pod logs workflow action provides automated log collection capabilities for Kubernetes pods when alerts are triggered, enhancing alert enrichment with relevant debugging information.

Pod Logs Action
---------------

**Purpose**: Automatically fetches container logs from Kubernetes pods to enrich alert notifications with debugging information.

**Trigger**: Fires on AlertManager events that contain pod information

**Actions**:
- Extracts pod name, namespace, and container from alert labels
- Fetches container logs using Kubernetes API
- Applies Java-specific configuration for Java containers
- Generates descriptive log filenames with timestamps
- Creates FileBlock enrichments with log content

**When it runs**: Triggered when workflows include ``pod_logs`` action and alert contains pod metadata

**Output**: FileBlock enrichment containing pod logs with descriptive filename and title

Java Application Support
------------------------

**Purpose**: Provides enhanced log collection for Java applications with specialized configuration.

**Trigger**: Automatically detects Java containers or when ``java_specific`` parameter is enabled

**Actions**:
- Auto-detects Java containers based on name patterns and images
- Applies increased log line limits for Java stack traces
- Uses Java-specific timeout configurations
- Adds "java-" prefix to log filenames
- Supports common Java frameworks and applications

**When it runs**: Triggered when Java containers are detected or Java-specific mode is explicitly enabled

**Output**: Enhanced log collection optimized for Java application debugging

Java Container Detection
~~~~~~~~~~~~~~~~~~~~~~~~

The action automatically identifies Java containers using these patterns:

- **Container names**: ``java``, ``spring``, ``tomcat``, ``jetty``, ``wildfly``, ``jboss``
- **Image names**: ``openjdk``, ``eclipse-temurin``, ``adoptopenjdk``, ``amazoncorretto``
- **Java applications**: ``kafka``, ``elasticsearch``, ``solr``, ``maven``, ``gradle``

Multi-Container Pod Support
---------------------------

**Purpose**: Handles pods with multiple containers by allowing specific container targeting.

**Trigger**: Fires on pods with multiple containers when container specification is needed

**Actions**:
- Extracts container name from alert labels or configuration
- Targets specific containers within multi-container pods
- Defaults to first container if no specific container is specified
- Includes container name in log filename for identification

**When it runs**: Triggered when processing multi-container pods or when specific container is configured

**Output**: Container-specific logs with container name included in enrichment

Configuration Integration
-------------------------

**Purpose**: Integrates with Helm chart configuration for default values and environment-based settings.

**Trigger**: Loads configuration from environment variables set by Helm deployment

**Actions**:
- Reads default values from Helm chart configuration
- Applies environment-based defaults for log limits and timeouts
- Supports Java-specific default overrides
- Allows workflow-level parameter overrides

**When it runs**: Configuration is loaded during action initialization and can be overridden per workflow

**Output**: Configured action with appropriate defaults and overrides applied

Alert Label Processing
---------------------

**Purpose**: Extracts pod information from Prometheus alert labels for log collection.

**Trigger**: Processes AlertManager events with pod-related labels

**Actions**:
- Extracts pod name from ``pod`` label
- Determines namespace from alert or defaults to "default"
- Identifies container from ``container`` label if available
- Falls back to ``instance`` label parsing if needed

**When it runs**: Triggered during alert processing to identify target pod and container

**Output**: Pod identification information used for log collection

File Naming and Organization
----------------------------

**Purpose**: Generates descriptive and organized log filenames for better log management.

**Trigger**: Creates filenames during log file generation

**Actions**:
- Includes timestamp in configurable format
- Adds namespace and pod name for identification
- Includes container name for multi-container pods
- Uses Java-specific prefix for Java applications
- Supports customizable filename patterns

**When it runs**: Triggered during log file creation and enrichment generation

**Output**: Descriptive log filenames following organizational patterns

Example Usage
-------------

Basic pod logs collection:

.. code-block:: yaml

    actions:
      - action_type: "pod_logs"
        data:
          max_lines: 1000
          tail_lines: 200

Java application logs with auto-detection:

.. code-block:: yaml

    actions:
      - action_type: "pod_logs"
        data:
          java_specific: true

Multi-container pod targeting:

.. code-block:: yaml

    actions:
      - action_type: "pod_logs"
        data:
          container: "app"
          max_lines: 1500

Previous container logs for crash investigation:

.. code-block:: yaml

    actions:
      - action_type: "pod_logs"
        data:
          previous: true
          max_lines: 2000 