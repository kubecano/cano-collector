Golang Workflows
===============

Golang workflows provide comprehensive debugging and analysis capabilities for Go applications running in Kubernetes pods, including profiling, memory analysis, and debugging tools.

Go Process Inspector
------------------

**Purpose**: Displays all Go debugging options for every Go process in a pod.

**Trigger**: Fires on pod events with Go processes

**Actions**:
- Identifies all Go processes in the pod
- Lists process IDs, executable paths, and command lines
- Provides Go debugging options for each process
- Shows Go runtime configuration details
- Generates Go process inspection reports

**When it runs**: Triggered when Go process inspection is requested for pods with Go applications

**Output**: Go process inspection report with debugging options

Go Profiling
-----------

**Purpose**: Runs Go profiling tools (pprof) on Go processes to analyze performance.

**Trigger**: Fires on Go profiling requests

**Actions**:
- Executes pprof commands on Go processes
- Collects CPU profiling data
- Analyzes memory profiling information
- Provides goroutine profiling
- Generates performance analysis reports

**When it runs**: Triggered when Go profiling is requested for Go processes

**Output**: Go profiling report with performance insights

Go Memory Analysis
-----------------

**Purpose**: Provides comprehensive memory analysis for Go applications.

**Trigger**: Fires on Go memory analysis requests

**Actions**:
- Analyzes Go runtime memory usage
- Identifies memory leaks and excessive allocations
- Provides garbage collection analysis
- Shows memory allocation patterns
- Generates memory optimization recommendations

**When it runs**: Triggered when Go memory analysis is requested

**Output**: Go memory analysis report with optimization recommendations

Go Goroutine Analysis
--------------------

**Purpose**: Analyzes goroutine usage and identifies potential issues.

**Trigger**: Fires on Go goroutine analysis requests

**Actions**:
- Counts active goroutines
- Identifies goroutine leaks
- Analyzes goroutine stack traces
- Detects deadlocks and blocking operations
- Generates goroutine health reports

**When it runs**: Triggered when Go goroutine analysis is requested

**Output**: Go goroutine analysis report with health status

Go Debugger
----------

**Purpose**: Attaches debugger to Go processes for interactive debugging.

**Trigger**: Fires on Go debugging requests

**Actions**:
- Attaches Delve debugger to Go processes
- Enables interactive debugging capabilities
- Provides debugging port configuration
- Supports remote debugging connections
- Generates debugging session reports

**When it runs**: Triggered when Go debugging is requested

**Output**: Go debugging session with interactive capabilities

Go Runtime Analysis
-----------------

**Purpose**: Analyzes Go runtime configuration and provides optimization recommendations.

**Trigger**: Fires on Go runtime analysis requests

**Actions**:
- Analyzes GOMAXPROCS configuration
- Reviews garbage collector settings
- Identifies runtime configuration issues
- Provides optimization recommendations
- Generates runtime analysis reports

**When it runs**: Triggered when Go runtime analysis is requested

**Output**: Go runtime analysis report with recommendations

Go Application Health Check
-------------------------

**Purpose**: Performs comprehensive health checks for Go applications.

**Trigger**: Fires on Go application health check requests

**Actions**:
- Checks Go runtime health status
- Analyzes application responsiveness
- Monitors goroutine pool status
- Identifies application issues
- Generates health check reports

**When it runs**: Triggered when Go application health checks are requested

**Output**: Go application health check report with status information

Go Build Information
-------------------

**Purpose**: Extracts and displays Go build information from binaries.

**Trigger**: Fires on Go build information requests

**Actions**:
- Extracts build version information
- Shows build flags and settings
- Displays module dependencies
- Provides build timestamp
- Generates build information reports

**When it runs**: Triggered when Go build information is requested

**Output**: Go build information report with version details

Go Module Analysis
-----------------

**Purpose**: Analyzes Go module dependencies and versions.

**Trigger**: Fires on Go module analysis requests

**Actions**:
- Analyzes go.mod and go.sum files
- Identifies dependency conflicts
- Shows module version information
- Detects security vulnerabilities
- Generates module analysis reports

**When it runs**: Triggered when Go module analysis is requested

**Output**: Go module analysis report with dependency information

Configuration
-------------

Go workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     golangWorkflows:
       goProcessInspector:
         enabled: true
         includeDebugOptions: true
       goProfiling:
         enabled: true
         includeCPUProfile: true
         includeMemoryProfile: true
         includeGoroutineProfile: true
       goMemoryAnalysis:
         enabled: true
         includeGCInfo: true
       goGoroutineAnalysis:
         enabled: true
         includeStackTraces: true
       goDebugger:
         enabled: true
         defaultPort: 2345
       goRuntimeAnalysis:
         enabled: true
         includeOptimizationTips: true
       goApplicationHealthCheck:
         enabled: true
         includeGoroutineInfo: true
       goBuildInformation:
         enabled: true
         includeModuleInfo: true
       goModuleAnalysis:
         enabled: true
         includeSecurityScan: true 