Python Workflows
================

Python workflows provide comprehensive debugging and analysis capabilities for Python applications running in Kubernetes pods, including profiling, memory analysis, and debugging tools.

Python Process Inspector
------------------------

**Purpose**: Displays all Python debugging options for every Python process in a pod.

**Trigger**: Fires on pod events with Python processes

**Actions**:
- Identifies all Python processes in the pod
- Lists process IDs, executable paths, and command lines
- Provides Python debugging options for each process
- Shows Python interpreter configuration details
- Generates Python process inspection reports

**When it runs**: Triggered when Python process inspection is requested for pods with Python applications

**Output**: Python process inspection report with debugging options

Python Profiling
----------------

**Purpose**: Runs Python profiling tools (cProfile, line_profiler) on Python processes.

**Trigger**: Fires on Python profiling requests

**Actions**:
- Executes cProfile on Python processes
- Collects line-by-line profiling data
- Analyzes function call statistics
- Provides memory profiling with memory_profiler
- Generates performance analysis reports

**When it runs**: Triggered when Python profiling is requested for Python processes

**Output**: Python profiling report with performance insights

Python Memory Analysis
----------------------

**Purpose**: Provides comprehensive memory analysis for Python applications.

**Trigger**: Fires on Python memory analysis requests

**Actions**:
- Analyzes Python memory usage with tracemalloc
- Identifies memory leaks and excessive allocations
- Provides garbage collection analysis
- Shows object reference counts
- Generates memory optimization recommendations

**When it runs**: Triggered when Python memory analysis is requested

**Output**: Python memory analysis report with optimization recommendations

Python Thread Analysis
----------------------

**Purpose**: Analyzes Python thread usage and identifies potential issues.

**Trigger**: Fires on Python thread analysis requests

**Actions**:
- Counts active threads
- Identifies thread leaks
- Analyzes thread stack traces
- Detects deadlocks and blocking operations
- Generates thread health reports

**When it runs**: Triggered when Python thread analysis is requested

**Output**: Python thread analysis report with health status

Python Debugger
---------------

**Purpose**: Attaches debugger to Python processes for interactive debugging.

**Trigger**: Fires on Python debugging requests

**Actions**:
- Attaches pdb/ipdb debugger to Python processes
- Enables interactive debugging capabilities
- Provides debugging port configuration
- Supports remote debugging connections
- Generates debugging session reports

**When it runs**: Triggered when Python debugging is requested

**Output**: Python debugging session with interactive capabilities

Python Interpreter Analysis
---------------------------

**Purpose**: Analyzes Python interpreter configuration and provides optimization recommendations.

**Trigger**: Fires on Python interpreter analysis requests

**Actions**:
- Analyzes Python version and implementation
- Reviews interpreter flags and settings
- Identifies configuration issues
- Provides optimization recommendations
- Generates interpreter analysis reports

**When it runs**: Triggered when Python interpreter analysis is requested

**Output**: Python interpreter analysis report with recommendations

Python Application Health Check
-------------------------------

**Purpose**: Performs comprehensive health checks for Python applications.

**Trigger**: Fires on Python application health check requests

**Actions**:
- Checks Python interpreter health status
- Analyzes application responsiveness
- Monitors thread pool status
- Identifies application issues
- Generates health check reports

**When it runs**: Triggered when Python application health checks are requested

**Output**: Python application health check report with status information

Python Package Analysis
-----------------------

**Purpose**: Analyzes Python package dependencies and versions.

**Trigger**: Fires on Python package analysis requests

**Actions**:
- Analyzes requirements.txt and setup.py files
- Identifies dependency conflicts
- Shows package version information
- Detects security vulnerabilities
- Generates package analysis reports

**When it runs**: Triggered when Python package analysis is requested

**Output**: Python package analysis report with dependency information

Python Virtual Environment Analysis
-----------------------------------

**Purpose**: Analyzes Python virtual environment configuration.

**Trigger**: Fires on Python virtual environment analysis requests

**Actions**:
- Identifies virtual environment location
- Shows Python interpreter path
- Analyzes installed packages
- Provides environment configuration details
- Generates virtual environment reports

**When it runs**: Triggered when Python virtual environment analysis is requested

**Output**: Python virtual environment analysis report with configuration details

Python Async Analysis
---------------------

**Purpose**: Analyzes Python async/await code and event loops.

**Trigger**: Fires on Python async analysis requests

**Actions**:
- Analyzes event loop configuration
- Identifies async task issues
- Shows coroutine statistics
- Detects blocking operations in async code
- Generates async analysis reports

**When it runs**: Triggered when Python async analysis is requested

**Output**: Python async analysis report with async code insights

Configuration
-------------

Python workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     pythonWorkflows:
       pythonProcessInspector:
         enabled: true
         includeDebugOptions: true
       pythonProfiling:
         enabled: true
         includeCPUProfile: true
         includeLineProfile: true
         includeMemoryProfile: true
       pythonMemoryAnalysis:
         enabled: true
         includeGCInfo: true
       pythonThreadAnalysis:
         enabled: true
         includeStackTraces: true
       pythonDebugger:
         enabled: true
         defaultPort: 5678
       pythonInterpreterAnalysis:
         enabled: true
         includeOptimizationTips: true
       pythonApplicationHealthCheck:
         enabled: true
         includeThreadInfo: true
       pythonPackageAnalysis:
         enabled: true
         includeSecurityScan: true
       pythonVirtualEnvironmentAnalysis:
         enabled: true
         includePackageList: true
       pythonAsyncAnalysis:
         enabled: true
         includeEventLoopInfo: true 