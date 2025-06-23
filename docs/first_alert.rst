Your First Alert
===============

This guide will help you configure and test your first alert with cano-collector.

Prerequisites
-------------

- Cano-collector installed and running
- Alertmanager configured to send alerts to cano-collector
- At least one destination configured (Slack, MS Teams, etc.)

Step 1: Verify Cano-collector is Running
-----------------------------------------

.. code-block:: bash

    # Check if pods are running
    kubectl get pods -n monitoring -l app=cano-collector

    # Check if service is available
    kubectl get svc -n monitoring cano-collector

    # Test the health endpoint
    kubectl port-forward svc/cano-collector 8080:8080 -n monitoring &
    curl http://localhost:8080/health

Step 2: Create a Test Alert Rule
---------------------------------

Create a simple Prometheus alert rule for testing:

.. code-block:: yaml

    # test-alert.yaml
    apiVersion: monitoring.coreos.com/v1alpha1
    kind: PrometheusRule
    metadata:
      name: test-alert
      namespace: monitoring
    spec:
      groups:
      - name: test
        rules:
        - alert: TestAlert
          expr: vector(1)
          for: 0s
          labels:
            severity: warning
          annotations:
            summary: "Test alert from cano-collector"
            description: "This is a test alert to verify cano-collector is working"

Apply the rule:

.. code-block:: bash

    kubectl apply -f test-alert.yaml

Step 3: Configure Alertmanager
-------------------------------

Ensure Alertmanager is configured to send alerts to cano-collector:

.. code-block:: yaml

    # alertmanager-config.yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: alertmanager-config
      namespace: monitoring
    type: Opaque
    data:
      alertmanager.yaml: |
        global:
          resolve_timeout: 5m
        route:
          receiver: 'cano-collector'
          group_by: ['alertname']
          group_wait: 10s
          group_interval: 10s
          repeat_interval: 1h
        receivers:
        - name: 'cano-collector'
          webhook_configs:
          - url: 'http://cano-collector.monitoring.svc.cluster.local:8080/api/alerts'
            send_resolved: true

Apply the configuration:

.. code-block:: bash

    kubectl apply -f alertmanager-config.yaml
    kubectl rollout restart deployment/alertmanager -n monitoring

Step 4: Test the Alert
----------------------

The test alert should fire immediately since we set `for: 0s`. Check if you received a notification in your configured destination (Slack, MS Teams, etc.).

If you don't see the alert, check the logs:

.. code-block:: bash

    # Check cano-collector logs
    kubectl logs -n monitoring -l app=cano-collector -f

    # Check Alertmanager logs
    kubectl logs -n monitoring -l app=alertmanager -f

Step 5: Create a Real Alert
----------------------------

Now let's create a more realistic alert. Create a pod that will fail:

.. code-block:: bash

    # Create a pod that will crash
    kubectl run test-pod --image=busybox --command -- sleep 1

    # Wait for it to fail
    sleep 10

    # Check pod status
    kubectl get pod test-pod

    # Clean up
    kubectl delete pod test-pod 