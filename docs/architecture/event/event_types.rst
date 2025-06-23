Event Types and Resource Monitoring
===================================

This document provides a comprehensive overview of all Kubernetes resource types that cano-collector monitors, including their event characteristics, severity levels, and monitoring considerations.

Core Resource Types
-------------------

**Pods**
~~~~~~~~

Pods are the fundamental units of deployment in Kubernetes and are critical for application monitoring.

**Monitored Events:**
- **CREATE**: Pod creation and scheduling
- **UPDATE**: Status changes, container state updates, resource allocation
- **DELETE**: Pod termination and cleanup

**Key Status Indicators:**
- **PodPhase**: Pending, Running, Succeeded, Failed, Unknown
- **ContainerStatus**: Waiting, Running, Terminated
- **Conditions**: PodScheduled, Ready, Initialized, ContainersReady

**Severity Mapping:**
- **CrashLoopBackOff**: HIGH - Container repeatedly failing
- **ImagePullBackOff**: WARNING - Image pull issues
- **Pending**: INFO - Pod waiting for resources
- **Failed**: HIGH - Pod failed to start
- **Unknown**: WARNING - Pod status unclear

**Example Event Processing:**
.. code-block:: go

    func processPodEvent(event *KubernetesEvent) *Issue {
        pod := event.Object.(*corev1.Pod)
        
        // Check for critical issues
        for _, container := range pod.Status.ContainerStatuses {
            if container.State.Waiting != nil {
                switch container.State.Waiting.Reason {
                case "CrashLoopBackOff":
                    return createHighSeverityIssue(event, "Pod in CrashLoopBackOff", 
                        "Container is repeatedly crashing")
                case "ImagePullBackOff":
                    return createWarningIssue(event, "Image Pull Failed", 
                        "Unable to pull container image")
                }
            }
        }
        
        // Check pod phase
        switch pod.Status.Phase {
        case corev1.PodFailed:
            return createHighSeverityIssue(event, "Pod Failed", 
                "Pod has failed to start or run")
        case corev1.PodPending:
            return createInfoIssue(event, "Pod Pending", 
                "Pod is waiting for resources")
        }
        
        return nil
    }

**Deployments**
~~~~~~~~~~~~~~~

Deployments manage the desired state for Pods and ReplicaSets, providing declarative updates.

**Monitored Events:**
- **CREATE**: Deployment creation
- **UPDATE**: Scaling, rolling updates, configuration changes
- **DELETE**: Deployment removal

**Key Status Indicators:**
- **Replicas**: Current, desired, updated, available, unavailable
- **Conditions**: Available, Progressing, ReplicaFailure
- **Strategy**: RollingUpdate, Recreate

**Severity Mapping:**
- **ReplicaFailure**: HIGH - Unable to create replicas
- **ProgressDeadlineExceeded**: HIGH - Update stuck
- **ScalingUp**: INFO - Increasing replica count
- **ScalingDown**: WARNING - Decreasing replica count
- **RolloutFailed**: HIGH - Rolling update failed

**Example Event Processing:**
.. code-block:: go

    func processDeploymentEvent(event *KubernetesEvent) *Issue {
        deployment := event.Object.(*appsv1.Deployment)
        
        // Check deployment conditions
        for _, condition := range deployment.Status.Conditions {
            switch condition.Type {
            case appsv1.DeploymentReplicaFailure:
                return createHighSeverityIssue(event, "Deployment Replica Failure", 
                    "Unable to create required replicas")
            case appsv1.DeploymentProgressing:
                if condition.Status == corev1.ConditionFalse {
                    return createHighSeverityIssue(event, "Deployment Progress Failed", 
                        "Deployment is not progressing")
                }
            }
        }
        
        // Check scaling events
        if event.Operation == "UPDATE" && event.OldObject != nil {
            oldDeployment := event.OldObject.(*appsv1.Deployment)
            if deployment.Spec.Replicas != nil && oldDeployment.Spec.Replicas != nil {
                if *deployment.Spec.Replicas > *oldDeployment.Spec.Replicas {
                    return createInfoIssue(event, "Deployment Scaling Up", 
                        fmt.Sprintf("Scaling from %d to %d replicas", 
                            *oldDeployment.Spec.Replicas, *deployment.Spec.Replicas))
                } else if *deployment.Spec.Replicas < *oldDeployment.Spec.Replicas {
                    return createWarningIssue(event, "Deployment Scaling Down", 
                        fmt.Sprintf("Scaling from %d to %d replicas", 
                            *oldDeployment.Spec.Replicas, *deployment.Spec.Replicas))
                }
            }
        }
        
        return nil
    }

**Services**
~~~~~~~~~~~~

Services provide stable endpoints for accessing Pods and enable load balancing.

**Monitored Events:**
- **CREATE**: Service creation
- **UPDATE**: Endpoint changes, configuration updates
- **DELETE**: Service removal

**Key Status Indicators:**
- **Endpoints**: Available endpoints for the service
- **LoadBalancer**: External IP allocation
- **Ports**: Service port configuration

**Severity Mapping:**
- **NoEndpoints**: HIGH - Service has no available endpoints
- **LoadBalancerPending**: WARNING - LoadBalancer IP pending
- **PortConflict**: HIGH - Port already in use
- **EndpointUpdate**: INFO - Endpoints changed

**Example Event Processing:**
.. code-block:: go

    func processServiceEvent(event *KubernetesEvent) *Issue {
        service := event.Object.(*corev1.Service)
        
        // Check for endpoint issues
        if service.Spec.Type == corev1.ServiceTypeClusterIP {
            endpoints, err := getServiceEndpoints(service.Namespace, service.Name)
            if err == nil && len(endpoints.Subsets) == 0 {
                return createHighSeverityIssue(event, "Service Has No Endpoints", 
                    "Service is not connected to any pods")
            }
        }
        
        // Check LoadBalancer status
        if service.Spec.Type == corev1.ServiceTypeLoadBalancer {
            if service.Status.LoadBalancer.Ingress == nil {
                return createWarningIssue(event, "LoadBalancer Pending", 
                    "Waiting for external IP allocation")
            }
        }
        
        return nil
    }

Workload Controllers
--------------------

**ReplicaSets**
~~~~~~~~~~~~~~~

ReplicaSets ensure a specified number of Pod replicas are running at any given time.

**Monitored Events:**
- **CREATE**: ReplicaSet creation
- **UPDATE**: Replica count changes, pod template updates
- **DELETE**: ReplicaSet removal

**Key Status Indicators:**
- **Replicas**: Current, desired, ready, available
- **Selector**: Pod selection criteria
- **Template**: Pod template specification

**Severity Mapping:**
- **ReplicaFailure**: HIGH - Unable to create replicas
- **ScalingEvent**: INFO - Replica count changed
- **TemplateUpdate**: WARNING - Pod template modified

**DaemonSets**
~~~~~~~~~~~~~~

DaemonSets ensure all (or some) nodes run a copy of a Pod.

**Monitored Events:**
- **CREATE**: DaemonSet creation
- **UPDATE**: Rolling updates, node affinity changes
- **DELETE**: DaemonSet removal

**Key Status Indicators:**
- **DesiredNumberScheduled**: Number of nodes that should be running pods
- **CurrentNumberScheduled**: Number of nodes currently running pods
- **NumberReady**: Number of nodes with ready pods
- **UpdatedNumberScheduled**: Number of nodes with updated pods

**Severity Mapping:**
- **NodeFailure**: HIGH - Pod failed to schedule on node
- **RolloutFailed**: HIGH - Rolling update failed
- **NodeAddition**: INFO - New node added to cluster
- **NodeRemoval**: WARNING - Node removed from cluster

**StatefulSets**
~~~~~~~~~~~~~~~~

StatefulSets manage stateful applications with stable network identities and persistent storage.

**Monitored Events:**
- **CREATE**: StatefulSet creation
- **UPDATE**: Scaling, rolling updates, storage changes
- **DELETE**: StatefulSet removal

**Key Status Indicators:**
- **Replicas**: Current, desired, ready, current
- **UpdateStrategy**: RollingUpdate, OnDelete
- **VolumeClaimTemplates**: Persistent volume claims

**Severity Mapping:**
- **StorageFailure**: HIGH - Persistent volume issues
- **ScalingEvent**: WARNING - Replica count changed
- **RolloutFailed**: HIGH - Rolling update failed
- **VolumeBinding**: INFO - Volume binding status

**Jobs and CronJobs**
~~~~~~~~~~~~~~~~~~~~~

Jobs create one or more Pods and ensure they complete successfully.

**Monitored Events:**
- **CREATE**: Job creation
- **UPDATE**: Status updates, completion
- **DELETE**: Job removal

**Key Status Indicators:**
- **Active**: Number of actively running pods
- **Succeeded**: Number of successfully completed pods
- **Failed**: Number of failed pods
- **CompletionTime**: When job completed

**Severity Mapping:**
- **JobFailed**: HIGH - Job execution failed
- **JobTimeout**: HIGH - Job exceeded timeout
- **JobCompleted**: INFO - Job completed successfully
- **JobSuspended**: WARNING - Job suspended

Configuration and Storage
-------------------------

**ConfigMaps**
~~~~~~~~~~~~~~

ConfigMaps store non-confidential configuration data.

**Monitored Events:**
- **CREATE**: ConfigMap creation
- **UPDATE**: Configuration data changes
- **DELETE**: ConfigMap removal

**Key Status Indicators:**
- **Data**: Configuration key-value pairs
- **BinaryData**: Binary configuration data

**Severity Mapping:**
- **ConfigUpdate**: WARNING - Configuration changed
- **ConfigDeletion**: HIGH - Configuration removed
- **ConfigCreation**: INFO - New configuration created

**Secrets**
~~~~~~~~~~~

Secrets store sensitive information like passwords and tokens.

**Monitored Events:**
- **CREATE**: Secret creation
- **UPDATE**: Secret data changes
- **DELETE**: Secret removal

**Key Status Indicators:**
- **Data**: Secret key-value pairs
- **Type**: Secret type (Opaque, kubernetes.io/service-account-token, etc.)

**Severity Mapping:**
- **SecretUpdate**: HIGH - Secret data changed
- **SecretDeletion**: HIGH - Secret removed
- **SecretCreation**: WARNING - New secret created

**PersistentVolumes and PersistentVolumeClaims**
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Persistent storage resources for stateful applications.

**Monitored Events:**
- **CREATE**: Volume creation
- **UPDATE**: Status changes, binding
- **DELETE**: Volume removal

**Key Status Indicators:**
- **Phase**: Available, Bound, Released, Failed
- **AccessModes**: ReadWriteOnce, ReadOnlyMany, ReadWriteMany
- **Capacity**: Storage capacity

**Severity Mapping:**
- **VolumeFailure**: HIGH - Volume provisioning failed
- **VolumeBinding**: WARNING - Volume binding issues
- **VolumeDeletion**: HIGH - Volume removed
- **VolumeExpansion**: INFO - Volume capacity increased

Networking and Security
-----------------------

**Ingress**
~~~~~~~~~~~

Ingress manages external access to services in a cluster.

**Monitored Events:**
- **CREATE**: Ingress creation
- **UPDATE**: Rule changes, TLS configuration
- **DELETE**: Ingress removal

**Key Status Indicators:**
- **Rules**: Ingress rules and paths
- **TLS**: TLS configuration
- **LoadBalancer**: Load balancer status

**Severity Mapping:**
- **IngressFailure**: HIGH - Ingress configuration failed
- **TLSUpdate**: WARNING - TLS configuration changed
- **RuleUpdate**: INFO - Ingress rules modified

**NetworkPolicies**
~~~~~~~~~~~~~~~~~~~

NetworkPolicies specify how Pods communicate with each other.

**Monitored Events:**
- **CREATE**: Policy creation
- **UPDATE**: Rule changes
- **DELETE**: Policy removal

**Key Status Indicators:**
- **PodSelector**: Pod selection criteria
- **PolicyTypes**: Ingress, Egress
- **Rules**: Network policy rules

**Severity Mapping:**
- **PolicyUpdate**: WARNING - Network policy changed
- **PolicyDeletion**: HIGH - Network policy removed
- **PolicyCreation**: INFO - New network policy created

**ServiceAccounts**
~~~~~~~~~~~~~~~~~~~

ServiceAccounts provide identity for Pods.

**Monitored Events:**
- **CREATE**: ServiceAccount creation
- **UPDATE**: Token changes
- **DELETE**: ServiceAccount removal

**Key Status Indicators:**
- **Secrets**: Associated secrets
- **ImagePullSecrets**: Image pull secrets

**Severity Mapping:**
- **AccountUpdate**: WARNING - ServiceAccount modified
- **AccountDeletion**: HIGH - ServiceAccount removed
- **TokenUpdate**: INFO - ServiceAccount token updated

**ClusterRoles and ClusterRoleBindings**
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

RBAC resources for cluster-wide permissions.

**Monitored Events:**
- **CREATE**: Role/Binding creation
- **UPDATE**: Permission changes
- **DELETE**: Role/Binding removal

**Key Status Indicators:**
- **Rules**: Permission rules
- **Subjects**: Users, groups, service accounts
- **RoleRef**: Referenced role

**Severity Mapping:**
- **PermissionChange**: HIGH - Permissions modified
- **RoleDeletion**: HIGH - Role removed
- **BindingUpdate**: WARNING - Role binding changed

Event Filtering and Configuration
---------------------------------

**Resource-Specific Filters:**
Each resource type can have specific filtering rules:

.. code-block:: yaml

    eventTypes:
      Pod:
        enabled: true
        filters:
          namespaces:
            - "production"
            - "staging"
          labels:
            app: ".*"
            tier: "frontend|backend"
          annotations:
            "kubernetes.io/change-cause": ".*"
        severity:
          CrashLoopBackOff: "HIGH"
          ImagePullBackOff: "WARNING"
          Pending: "INFO"
          Running: "INFO"
      
      Deployment:
        enabled: true
        filters:
          namespaces:
            - "production"
          labels:
            app: ".*"
        severity:
          ReplicaFailure: "HIGH"
          ProgressDeadlineExceeded: "HIGH"
          ScalingUp: "INFO"
          ScalingDown: "WARNING"
      
      Service:
        enabled: true
        filters:
          types:
            - "LoadBalancer"
            - "ClusterIP"
        severity:
          NoEndpoints: "HIGH"
          LoadBalancerPending: "WARNING"
          EndpointUpdate: "INFO"

**Global Event Filters:**
Global filters apply to all resource types:

.. code-block:: yaml

    globalFilters:
      # Namespace filters
      namespaces:
        include:
          - "production"
          - "staging"
        exclude:
          - "kube-system"
          - "default"
      
      # Label filters
      labels:
        required:
          app: ".*"
        excluded:
          component: "test"
      
      # Annotation filters
      annotations:
        required:
          "monitoring.kubernetes.io/enabled": "true"
      
      # Operation filters
      operations:
        - "CREATE"
        - "UPDATE"
        - "DELETE"
      
      # Severity filters
      severity:
        - "WARNING"
        - "HIGH"
        - "CRITICAL"

**Custom Event Types:**
Support for custom resource definitions (CRDs):

.. code-block:: yaml

    customResources:
      - apiVersion: "custom.example.com/v1"
        kind: "CustomResource"
        enabled: true
        filters:
          namespaces:
            - "production"
        severity:
          CustomError: "HIGH"
          CustomWarning: "WARNING"
          CustomInfo: "INFO"

Event Processing Configuration
------------------------------

**Resource-Specific Processing:**
Configure how each resource type is processed:

.. code-block:: yaml

    processing:
      Pod:
        contextGathering:
          includeLogs: true
          includeMetrics: true
          maxLogLines: 100
          includeEvents: true
          maxEvents: 20
        enrichment:
          autoEnrich: true
          includeRelatedResources: true
          includeNodeInfo: true
      
      Deployment:
        contextGathering:
          includeMetrics: true
          includeEvents: true
          maxEvents: 10
        enrichment:
          autoEnrich: true
          includeRelatedPods: true
          includeReplicaSetInfo: true
      
      Service:
        contextGathering:
          includeEndpoints: true
          includeEvents: true
        enrichment:
          autoEnrich: true
          includeRelatedPods: true

**Event Aggregation:**
Configure how similar events are aggregated:

.. code-block:: yaml

    aggregation:
      enabled: true
      window: "5m"
      rules:
        - resourceType: "Pod"
          groupBy: ["namespace", "app"]
          maxEvents: 10
        - resourceType: "Deployment"
          groupBy: ["namespace", "app"]
          maxEvents: 5
        - resourceType: "Service"
          groupBy: ["namespace"]
          maxEvents: 3

This comprehensive event type configuration provides fine-grained control over what events are monitored, how they are processed, and what actions are taken based on their severity and context. 