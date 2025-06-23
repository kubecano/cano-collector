Pod Troubleshooting Workflows
=============================

Pod troubleshooting workflows provide deep diagnostic capabilities for investigating issues with Kubernetes pods, including process analysis, debugging, and performance profiling.

Python Profiler
---------------

**Purpose**: Attaches a Python profiler to running Python processes in pods to analyze performance bottlenecks.

**Trigger**: Fires on pod events with Python processes

**Actions**:
- Identifies Python processes in the pod
- Attaches PySpy profiler to running processes
- Generates profiling reports in SVG format
- Provides performance analysis data

**When it runs**: Triggered when Python profiling is requested for pods with Python applications

**Output**: Profiling reports showing CPU usage patterns and performance bottlenecks

Pod Process List
----------------

**Purpose**: Displays a list of all running processes in a pod for process analysis.

**Trigger**: Fires on pod events

**Actions**:
- Lists all running processes in the pod
- Shows process IDs, executable paths, and command lines
- Provides process hierarchy information

**When it runs**: Triggered when process listing is requested for any pod

**Output**: Table showing all processes with their details

Python Memory Analysis
----------------------

**Purpose**: Monitors Python processes for memory allocation patterns and potential memory leaks.

**Trigger**: Fires on pod events with Python processes

**Actions**:
- Attaches memory profiler to Python processes
- Tracks memory allocations and deallocations
- Identifies memory leaks and allocation patterns
- Provides memory usage statistics

**When it runs**: Triggered when memory analysis is requested for Python applications

**Output**: Memory analysis report with allocation patterns and leak detection

Debugger Stack Trace
--------------------

**Purpose**: Captures stack traces from Python processes for debugging purposes.

**Trigger**: Fires on pod events with Python processes

**Actions**:
- Attaches debugger to Python processes
- Captures stack traces from all threads
- Provides debugging information
- Supports multiple trace captures

**When it runs**: Triggered when stack trace analysis is requested

**Output**: Stack trace information for debugging Python applications

Python Process Inspector
------------------------

**Purpose**: Provides comprehensive inspection of Python processes including loaded modules and debugging information.

**Trigger**: Fires on pod events with Python processes

**Actions**:
- Inspects Python process internals
- Lists loaded modules and their versions
- Provides debugging warnings and information
- Shows process configuration details

**When it runs**: Triggered when detailed Python process inspection is requested

**Output**: Comprehensive Python process inspection report

Python Debugger
---------------

**Purpose**: Attaches a full Python debugger to running processes for interactive debugging.

**Trigger**: Fires on pod events with Python processes

**Actions**:
- Attaches Python debugger to processes
- Enables interactive debugging capabilities
- Provides debugging port configuration
- Supports remote debugging connections

**When it runs**: Triggered when interactive debugging is requested

**Output**: Debugging session with interactive capabilities

Configuration
-------------

Pod troubleshooting workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     podTroubleshooting:
       pythonProfiler:
         enabled: true
         defaultDuration: 2
         includeIdle: false
       podProcessList:
         enabled: true
       pythonMemoryAnalysis:
         enabled: true
         defaultDuration: 60
       debuggerStackTrace:
         enabled: true
         tracesAmount: 1
         sleepDuration: 1
       pythonProcessInspector:
         enabled: true
       pythonDebugger:
         enabled: true
         defaultPort: 5678 