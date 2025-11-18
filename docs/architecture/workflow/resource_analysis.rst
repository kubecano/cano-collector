Resource Analysis Workflows
===========================


.. note::
   **Status**: 🔨 Planned - Not yet implemented
Resource analysis workflows provide comprehensive analysis of Kubernetes cluster resources, helping optimize resource usage and identify potential issues.

KRR (Kubernetes Resource Recommender) Scan
------------------------------------------

**Purpose**: Analyzes Kubernetes resource requests and limits to provide optimization recommendations.

**Trigger**: Fires on scheduled scans or manual requests

**Actions**:
- Scans all pods in the cluster for resource configurations
- Analyzes CPU and memory requests and limits
- Provides optimization recommendations
- Generates resource usage reports
- Identifies over-provisioned and under-provisioned resources

**When it runs**: Triggered on scheduled intervals or manual execution

**Output**: Comprehensive resource optimization report with recommendations

Popeye Scan
-----------

**Purpose**: Performs cluster health analysis to identify potential issues and best practice violations.

**Trigger**: Fires on scheduled scans or manual requests

**Actions**:
- Scans cluster resources for health issues
- Identifies best practice violations
- Provides cluster health score
- Generates detailed issue reports
- Supports multiple resource types (pods, services, deployments, etc.)

**When it runs**: Triggered on scheduled intervals or manual execution

**Output**: Cluster health report with issues and recommendations

Resource Usage Analysis
-----------------------

**Purpose**: Analyzes current resource usage patterns across the cluster.

**Trigger**: Fires on resource usage events or scheduled analysis

**Actions**:
- Collects resource usage metrics from Prometheus
- Analyzes usage patterns and trends
- Identifies resource bottlenecks
- Provides capacity planning insights
- Generates usage optimization recommendations

**When it runs**: Triggered when resource usage analysis is requested

**Output**: Resource usage analysis with optimization recommendations

Capacity Planning
-----------------

**Purpose**: Provides insights for cluster capacity planning and resource allocation.

**Trigger**: Fires on capacity planning requests or scheduled analysis

**Actions**:
- Analyzes historical resource usage trends
- Predicts future resource requirements
- Provides scaling recommendations
- Identifies resource allocation inefficiencies
- Generates capacity planning reports

**When it runs**: Triggered when capacity planning analysis is requested

**Output**: Capacity planning report with scaling recommendations

Resource Optimization
---------------------

**Purpose**: Provides specific recommendations for optimizing resource allocation.

**Trigger**: Fires on resource optimization requests

**Actions**:
- Analyzes current resource allocation
- Identifies optimization opportunities
- Provides specific resource adjustment recommendations
- Calculates potential cost savings
- Generates optimization action plans

**When it runs**: Triggered when resource optimization is requested

**Output**: Resource optimization recommendations with action plans

Configuration
-------------

Resource analysis workflows can be configured through Helm values:

.. code-block:: yaml

   workflows:
     resourceAnalysis:
       krrScan:
         enabled: true
         schedule: "0 2 * * *"  # Daily at 2 AM
         strategy: "simple"
         maxWorkers: 3
         timeout: 3600
       popeyeScan:
         enabled: true
         schedule: "0 3 * * *"  # Daily at 3 AM
         timeout: 300
         args: "-s no,ns,po,svc,sa,cm,dp,sts,ds,pv,pvc,hpa,pdb,cr,crb,ro,rb,ing,np,psp"
       resourceUsageAnalysis:
         enabled: true
         schedule: "0 */6 * * *"  # Every 6 hours
       capacityPlanning:
         enabled: true
         schedule: "0 4 * * 1"  # Weekly on Monday at 4 AM
       resourceOptimization:
         enabled: true 