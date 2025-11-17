Java Workflows
==============

.. note::
   **Status**: 🔨 Planned - Not yet implemented

Java workflows provide comprehensive debugging and analysis capabilities for Java applications running in Kubernetes pods, including JVM profiling, memory analysis, and debugging tools.

Java Process Inspector
----------------------

**Purpose**: Displays all Java debugging options for every Java process in a pod.

**Trigger**: Fires on pod events with Java processes

**Actions**:
- Identifies all Java processes in the pod
- Lists process IDs, executable paths, and command lines
- Provides Java debugging options for each process
- Shows JVM configuration details
- Generates Java process inspection reports

**When it runs**: Triggered when Java process inspection is requested for pods with Java applications

**Output**: Java process inspection report with debugging options

Pod JMap PID
------------

**Purpose**: Runs jmap on a specific Java process to analyze memory usage.

**Trigger**: Fires on jmap requests for specific Java processes

**Actions**:
- Executes jmap command on specified Java process
- Analyzes memory heap usage
- Provides memory allocation patterns
- Shows object distribution
- Generates memory analysis reports

**When it runs**: Triggered when jmap analysis is requested for Java processes

**Output**: JMap analysis report with memory usage details

Pod JStack PID
--------------

**Purpose**: Runs jstack on a specific Java process to capture thread dumps.

**Trigger**: Fires on jstack requests for specific Java processes

**Actions**:
- Executes jstack command on specified Java process
- Captures thread dump information
- Analyzes thread states and stack traces
- Identifies thread contention issues
- Generates thread analysis reports

**When it runs**: Triggered when jstack analysis is requested for Java processes

**Output**: JStack analysis report with thread dump details

Java Memory Analysis
--------------------

**Purpose**: Provides comprehensive memory analysis for Java applications.

**Trigger**: Fires on Java memory analysis requests

**Actions**:
- Analyzes JVM memory usage patterns
- Identifies memory leaks and excessive usage
- Provides garbage collection analysis
- Shows memory allocation trends
- Generates memory optimization recommendations

**When it runs**: Triggered when Java memory analysis is requested

**Output**: Java memory analysis report with optimization recommendations

Java Performance Profiling
--------------------------

**Purpose**: Provides performance profiling for Java applications.

**Trigger**: Fires on Java performance profiling requests

**Actions**:
- Attaches profiler to Java processes
- Collects performance metrics
- Analyzes CPU usage patterns
- Identifies performance bottlenecks
- Generates performance optimization reports

**When it runs**: Triggered when Java performance profiling is requested

**Output**: Java performance profiling report with optimization insights

Java Debugger
-------------

**Purpose**: Attaches debugger to Java processes for interactive debugging.

**Trigger**: Fires on Java debugging requests

**Actions**:
- Attaches Java debugger to processes
- Enables interactive debugging capabilities
- Provides debugging port configuration
- Supports remote debugging connections
- Generates debugging session reports

**When it runs**: Triggered when Java debugging is requested

**Output**: Java debugging session with interactive capabilities

JVM Configuration Analysis
--------------------------

**Purpose**: Analyzes JVM configuration and provides optimization recommendations.

**Trigger**: Fires on JVM configuration analysis requests

**Actions**:
- Analyzes JVM startup parameters
- Reviews memory configuration settings
- Identifies configuration issues
- Provides optimization recommendations
- Generates configuration analysis reports

**When it runs**: Triggered when JVM configuration analysis is requested

**Output**: JVM configuration analysis report with recommendations

Java Application Health Check
-----------------------------

**Purpose**: Performs comprehensive health checks for Java applications.

**Trigger**: Fires on Java application health check requests

**Actions**:
- Checks JVM health status
- Analyzes application responsiveness
- Monitors thread pool status
- Identifies application issues
- Generates health check reports

**When it runs**: Triggered when Java application health checks are requested

**Output**: Java application health check report with status information

Configuration
-------------

Java workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     javaWorkflows:
       javaProcessInspector:
         enabled: true
         includeDebugOptions: true
       podJmapPid:
         enabled: true
         includeHeapAnalysis: true
       podJstackPid:
         enabled: true
         includeThreadAnalysis: true
       javaMemoryAnalysis:
         enabled: true
         includeGCInfo: true
       javaPerformanceProfiling:
         enabled: true
         profilingDuration: 60
       javaDebugger:
         enabled: true
         defaultPort: 5005
       jvmConfigurationAnalysis:
         enabled: true
         includeOptimizationTips: true
       javaApplicationHealthCheck:
         enabled: true
         includeThreadPoolInfo: true 