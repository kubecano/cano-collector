Troubleshooting Guide
====================

This guide helps you diagnose and fix common issues with cano-collector.

Common Issues
-------------

Alerts Not Being Received
-------------------------

**Symptoms:**
- No notifications in Slack/MS Teams/Jira
- Cano-collector logs show no incoming requests
- Alertmanager shows failed webhook deliveries

**Diagnosis:**

1. Check if cano-collector is receiving alerts:

.. code-block:: bash

    # Check cano-collector logs
    kubectl logs -n monitoring -l app=cano-collector -f

    # Look for lines like:
    # "Received alert from Alertmanager"
    # "Processing alert: TestAlert"

2. Check Alertmanager configuration:

.. code-block:: bash

    # Get Alertmanager config
    kubectl get secret alertmanager-config -n monitoring -o jsonpath='{.data.alertmanager\.yaml}' | base64 -d

    # Verify the webhook URL is correct
    # Should be: http://cano-collector.monitoring.svc.cluster.local:8080/api/alerts

3. Test connectivity:

.. code-block:: bash

    # Test from Alertmanager pod to cano-collector
    kubectl exec -n monitoring deployment/alertmanager -- curl -v http://cano-collector:8080/health

**Solutions:**

1. **Incorrect webhook URL**: Update Alertmanager configuration with correct service name
2. **Network issues**: Check if pods can communicate within the cluster
3. **Service not exposed**: Verify cano-collector service is running

Notifications Not Being Sent
----------------------------

**Symptoms:**
- Cano-collector receives alerts but no notifications sent
- Error messages in logs about failed HTTP requests
- Destination configuration issues

**Diagnosis:**

1. Check destination configuration:

.. code-block:: bash

    # Get current configuration
    kubectl get configmap cano-collector-config -n monitoring -o yaml

2. Check logs for specific errors:

.. code-block:: bash

    # Look for error messages
    kubectl logs -n monitoring -l app=cano-collector | grep -i error

**Common Solutions:**

1. **Invalid webhook URL**: Verify Slack/MS Teams webhook URLs are correct
2. **Authentication issues**: Check API tokens and credentials
3. **Rate limiting**: Some services have rate limits

Pod Not Starting
----------------

**Symptoms:**
- Pod stuck in Pending or CrashLoopBackOff
- Application errors during startup

**Diagnosis:**

1. Check pod status:

.. code-block:: bash

    kubectl describe pod -n monitoring -l app=cano-collector

2. Check logs:

.. code-block:: bash

    kubectl logs -n monitoring -l app=cano-collector --previous

**Common Solutions:**

1. **Resource constraints**: Increase CPU/memory limits
2. **Configuration errors**: Fix YAML syntax in ConfigMap
3. **Image pull issues**: Check image repository and credentials

Configuration Issues
--------------------

**Symptoms:**
- Cano-collector starts but doesn't load configuration
- Invalid YAML syntax errors
- Missing required fields

**Diagnosis:**

1. Validate YAML syntax:

.. code-block:: bash

    # Test YAML syntax
    kubectl get configmap cano-collector-config -n monitoring -o jsonpath='{.data.destinations\.yaml}' | yq eval .

2. Check configuration structure:

.. code-block:: yaml

    # Valid destinations.yaml structure
    destinations:
      slack:
        - name: "alerts-prod"
          webhookURL: "https://hooks.slack.com/services/YOUR/WEBHOOK"
      msteams:
        - name: "ops-team"
          webhookURL: "https://your-org.webhook.office.com/webhookb2/YOUR/WEBHOOK"

**Common Issues:**

1. **Missing required fields**: Ensure all required parameters are provided
2. **Invalid YAML**: Check indentation and syntax
3. **Wrong file paths**: Verify ConfigMap mounts correctly

Debug Commands
--------------

Useful commands for debugging:

.. code-block:: bash

    # Check pod status and events
    kubectl describe pod -n monitoring -l app=cano-collector

    # Follow logs in real-time
    kubectl logs -n monitoring -l app=cano-collector -f

    # Check service endpoints
    kubectl get endpoints -n monitoring cano-collector

    # Test service connectivity
    kubectl run test-pod --image=busybox --rm -it --restart=Never -- \
      wget -qO- http://cano-collector:8080/health

    # Check configuration
    kubectl get configmap -n monitoring cano-collector-config -o yaml

    # Check secrets
    kubectl get secret -n monitoring cano-collector-secrets -o yaml

    # Port forward for local testing
    kubectl port-forward svc/cano-collector 8080:8080 -n monitoring

Health Checks
-------------

Cano-collector provides health check endpoints:

.. code-block:: bash

    # Basic health check
    curl http://cano-collector.monitoring.svc.cluster.local:8080/health

    # Detailed health check
    curl http://cano-collector.monitoring.svc.cluster.local:8080/health/detailed

    # Metrics endpoint
    curl http://cano-collector.monitoring.svc.cluster.local:8080/metrics

Expected Responses:

- **Health**: `{"status":"ok"}`
- **Detailed**: `{"status":"ok","components":{"config":"ok","destinations":"ok"}}`
- **Metrics**: Prometheus metrics in text format

Log Levels
----------

Adjust log verbosity for debugging:

.. code-block:: yaml

    # In values.yaml
    env:
      - name: LOG_LEVEL
        value: "debug"  # Options: debug, info, warn, error

Common Log Messages
------------------

**Normal Operation:**
- `"Starting cano-collector"`
- `"Configuration loaded successfully"`
- `"Received alert from Alertmanager"`
- `"Alert processed successfully"`

**Warning Messages:**
- `"Destination not found"`
- `"Failed to send notification"`
- `"Configuration validation warning"`

**Error Messages:**
- `"Failed to load configuration"`
- `"Invalid webhook URL"`
- `"Authentication failed"`

Getting Help
------------

If you're still experiencing issues:

1. **Check the logs**: Use the debug commands above
2. **Verify configuration**: Ensure all required fields are set
3. **Test connectivity**: Verify network connectivity between components
4. **Check documentation**: Review the configuration guides
5. **Open an issue**: Create a GitHub issue with logs and configuration 