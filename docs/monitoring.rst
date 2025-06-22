Monitoring and Operations
=========================

This guide covers monitoring cano-collector in production and operational best practices.

Metrics and Monitoring
----------------------

Cano-collector exposes Prometheus metrics for monitoring:

Metrics Endpoint
~~~~~~~~~~~~~~~

.. code-block:: bash

    # Access metrics
    curl http://cano-collector.monitoring.svc.cluster.local:8080/metrics

Key Metrics
~~~~~~~~~~~

**Alert Processing Metrics:**
- `cano_alerts_received_total` - Total alerts received
- `cano_alerts_processed_total` - Total alerts processed
- `cano_alerts_processing_duration_seconds` - Alert processing time
- `cano_alerts_errors_total` - Total processing errors

**Destination Metrics:**
- `cano_destination_sent_total` - Messages sent per destination
- `cano_destination_errors_total` - Send errors per destination
- `cano_destination_duration_seconds` - Send duration per destination

**System Metrics:**
- `cano_http_requests_total` - HTTP request count
- `cano_http_request_duration_seconds` - HTTP request duration
- `cano_config_reloads_total` - Configuration reload count

Example Prometheus Queries
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: promql

    # Alert processing rate
    rate(cano_alerts_received_total[5m])

    # Error rate
    rate(cano_alerts_errors_total[5m])

    # 95th percentile processing time
    histogram_quantile(0.95, rate(cano_alerts_processing_duration_seconds_bucket[5m]))

    # Destination success rate
    rate(cano_destination_sent_total[5m]) / (rate(cano_destination_sent_total[5m]) + rate(cano_destination_errors_total[5m]))

Alerting Rules
--------------

Create Prometheus alerting rules for cano-collector:

.. code-block:: yaml

    apiVersion: monitoring.coreos.com/v1alpha1
    kind: PrometheusRule
    metadata:
      name: cano-collector-alerts
      namespace: monitoring
    spec:
      groups:
      - name: cano-collector
        rules:
        - alert: CanoCollectorDown
          expr: up{job="cano-collector"} == 0
          for: 1m
          labels:
            severity: critical
          annotations:
            summary: "Cano-collector is down"
            description: "Cano-collector pod is not running"

        - alert: CanoCollectorHighErrorRate
          expr: rate(cano_alerts_errors_total[5m]) > 0.1
          for: 2m
          labels:
            severity: warning
          annotations:
            summary: "High error rate in cano-collector"
            description: "Cano-collector is experiencing high error rates"

        - alert: CanoCollectorHighLatency
          expr: histogram_quantile(0.95, rate(cano_alerts_processing_duration_seconds_bucket[5m])) > 5
          for: 2m
          labels:
            severity: warning
          annotations:
            summary: "High processing latency in cano-collector"
            description: "Cano-collector is taking too long to process alerts"

        - alert: CanoCollectorDestinationFailure
          expr: rate(cano_destination_errors_total[5m]) > 0
          for: 1m
          labels:
            severity: warning
          annotations:
            summary: "Destination failures in cano-collector"
            description: "Cano-collector is failing to send to destinations"

Health Checks
-------------

Cano-collector provides health check endpoints:

Basic Health Check
~~~~~~~~~~~~~~~~~

.. code-block:: bash

    curl http://cano-collector.monitoring.svc.cluster.local:8080/health

Response: `{"status":"ok"}`

Detailed Health Check
~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    curl http://cano-collector.monitoring.svc.cluster.local:8080/health/detailed

Response:
.. code-block:: json

    {
      "status": "ok",
      "components": {
        "config": "ok",
        "destinations": "ok",
        "database": "ok"
      },
      "uptime": "2h30m15s",
      "version": "1.0.0"
    }

Kubernetes Health Checks
~~~~~~~~~~~~~~~~~~~~~~~

Configure health checks in deployment:

.. code-block:: yaml

    livenessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
      timeoutSeconds: 5
      failureThreshold: 3

    readinessProbe:
      httpGet:
        path: /health
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
      timeoutSeconds: 3
      failureThreshold: 3

Logging
-------

Log Configuration
~~~~~~~~~~~~~~~~

Configure log levels and format:

.. code-block:: yaml

    env:
      - name: LOG_LEVEL
        value: "info"  # debug, info, warn, error
      - name: LOG_FORMAT
        value: "json"  # json, text

Log Aggregation
~~~~~~~~~~~~~~

For production environments, configure log aggregation:

.. code-block:: yaml

    # Fluentd configuration
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: fluentd-config
    data:
      fluent.conf: |
        <source>
          @type tail
          path /var/log/cano-collector/*.log
          pos_file /var/log/fluentd-cano-collector.log.pos
          tag cano-collector
          <parse>
            @type json
          </parse>
        </source>

        <match cano-collector>
          @type elasticsearch
          host elasticsearch
          port 9200
          index_name cano-collector
        </match>

Backup and Recovery
-------------------

Configuration Backup
~~~~~~~~~~~~~~~~~~~

Backup your configuration regularly:

.. code-block:: bash

    # Backup destinations configuration
    kubectl get configmap cano-collector-config -n monitoring -o yaml > backup-destinations.yaml

    # Backup secrets
    kubectl get secret cano-collector-secrets -n monitoring -o yaml > backup-secrets.yaml

    # Backup Helm values
    helm get values cano-collector -n monitoring > backup-values.yaml

Recovery Procedures
~~~~~~~~~~~~~~~~~~

1. **Configuration Recovery:**
   .. code-block:: bash

       kubectl apply -f backup-destinations.yaml
       kubectl apply -f backup-secrets.yaml

2. **Application Recovery:**
   .. code-block:: bash

       helm upgrade cano-collector cano-collector/cano-collector \
         --values backup-values.yaml \
         --namespace monitoring

3. **Data Recovery:**
   - Cano-collector is stateless, no data recovery needed
   - Configuration is stored in ConfigMaps and Secrets

Performance Tuning
------------------

Resource Limits
~~~~~~~~~~~~~~

Adjust resource limits based on load:

.. code-block:: yaml

    resources:
      requests:
        memory: "256Mi"
        cpu: "200m"
      limits:
        memory: "512Mi"
        cpu: "500m"

Scaling
~~~~~~~

Scale horizontally for high load:

.. code-block:: bash

    # Scale to multiple replicas
    kubectl scale deployment cano-collector --replicas=3 -n monitoring

    # Or use HPA
    kubectl autoscale deployment cano-collector \
      --cpu-percent=70 \
      --min=2 \
      --max=10 \
      -n monitoring

Network Optimization
~~~~~~~~~~~~~~~~~~~

Optimize network settings:

.. code-block:: yaml

    # Increase connection pool
    env:
      - name: HTTP_MAX_IDLE_CONNS
        value: "100"
      - name: HTTP_IDLE_CONN_TIMEOUT
        value: "90s"

Security
--------

Network Policies
~~~~~~~~~~~~~~~

Restrict network access:

.. code-block:: yaml

    apiVersion: networking.k8s.io/v1
    kind: NetworkPolicy
    metadata:
      name: cano-collector-network-policy
      namespace: monitoring
    spec:
      podSelector:
        matchLabels:
          app: cano-collector
      policyTypes:
      - Ingress
      - Egress
      ingress:
      - from:
        - namespaceSelector:
            matchLabels:
              name: monitoring
        ports:
        - protocol: TCP
          port: 8080
      egress:
      - to:
        - namespaceSelector:
            matchLabels:
              name: monitoring
        ports:
        - protocol: TCP
          port: 9090

RBAC Configuration
~~~~~~~~~~~~~~~~~~

Minimal RBAC permissions:

.. code-block:: yaml

    apiVersion: rbac.authorization.k8s.io/v1
    kind: ClusterRole
    metadata:
      name: cano-collector
    rules:
    - apiGroups: [""]
      resources: ["pods", "services", "endpoints"]
      verbs: ["get", "list", "watch"]
    - apiGroups: [""]
      resources: ["events"]
      verbs: ["create", "patch"]

Maintenance
-----------

Regular Maintenance Tasks
~~~~~~~~~~~~~~~~~~~~~~~~

1. **Update cano-collector:**
   .. code-block:: bash

       helm upgrade cano-collector cano-collector/cano-collector \
         --namespace monitoring

2. **Rotate secrets:**
   .. code-block:: bash

       # Update webhook URLs and API tokens
       kubectl patch secret cano-collector-secrets -n monitoring \
         --patch='{"data":{"slack-webhook":"new-base64-encoded-value"}}'

3. **Clean up old logs:**
   .. code-block:: bash

       # Configure log rotation in deployment
       kubectl patch deployment cano-collector -n monitoring \
         --patch='{"spec":{"template":{"spec":{"containers":[{"name":"cano-collector","volumeMounts":[{"name":"logs","mountPath":"/var/log"}]}]}}}}'

4. **Monitor resource usage:**
   .. code-block:: bash

       # Check resource usage
       kubectl top pods -n monitoring -l app=cano-collector

       # Check disk usage
       kubectl exec -n monitoring deployment/cano-collector -- df -h 