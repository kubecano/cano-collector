Maintenance Guide
================

This guide covers routine maintenance tasks and best practices for keeping cano-collector running smoothly in production.

Daily Maintenance
-----------------

Health Checks
~~~~~~~~~~~~

Perform daily health checks:

.. code-block:: bash

    # Check pod status
    kubectl get pods -n monitoring -l app=cano-collector

    # Check service endpoints
    kubectl get endpoints -n monitoring cano-collector

    # Verify health endpoint
    kubectl port-forward svc/cano-collector 8080:8080 -n monitoring &
    curl http://localhost:8080/health

    # Check recent logs for errors
    kubectl logs -n monitoring -l app=cano-collector --tail=100 | grep -i error

Metrics Review
~~~~~~~~~~~~~

Review key metrics daily:

.. code-block:: bash

    # Check alert processing rate
    curl -s http://cano-collector.monitoring.svc.cluster.local:8080/metrics | grep cano_alerts_received_total

    # Check error rate
    curl -s http://cano-collector.monitoring.svc.cluster.local:8080/metrics | grep cano_alerts_errors_total

    # Check destination success rates
    curl -s http://cano-collector.monitoring.svc.cluster.local:8080/metrics | grep cano_destination

Weekly Maintenance
------------------

Configuration Review
~~~~~~~~~~~~~~~~~~~

Review and validate configuration:

.. code-block:: bash

    # Backup current configuration
    kubectl get configmap cano-collector-config -n monitoring -o yaml > backup-$(date +%Y%m%d).yaml

    # Validate YAML syntax
    kubectl get configmap cano-collector-config -n monitoring -o jsonpath='{.data.destinations\.yaml}' | yq eval .

    # Check for unused destinations
    kubectl get configmap cano-collector-config -n monitoring -o jsonpath='{.data.destinations\.yaml}' | yq eval '.destinations | keys'

Log Analysis
~~~~~~~~~~~~

Analyze logs for patterns and issues:

.. code-block:: bash

    # Check for repeated errors
    kubectl logs -n monitoring -l app=cano-collector --since=7d | grep -i error | sort | uniq -c | sort -nr

    # Check for slow processing
    kubectl logs -n monitoring -l app=cano-collector --since=7d | grep "processing duration"

    # Check for destination failures
    kubectl logs -n monitoring -l app=cano-collector --since=7d | grep "failed to send"

Resource Usage
~~~~~~~~~~~~~

Monitor resource consumption:

.. code-block:: bash

    # Check CPU and memory usage
    kubectl top pods -n monitoring -l app=cano-collector

    # Check disk usage
    kubectl exec -n monitoring deployment/cano-collector -- df -h

    # Check network connections
    kubectl exec -n monitoring deployment/cano-collector -- netstat -an | wc -l

Monthly Maintenance
-------------------

Security Review
~~~~~~~~~~~~~~

Review and rotate secrets:

.. code-block:: bash

    # List all secrets
    kubectl get secrets -n monitoring | grep cano-collector

    # Check secret age
    kubectl get secrets -n monitoring cano-collector-secrets -o yaml | grep creationTimestamp

    # Rotate webhook URLs and API tokens
    # Update secrets with new values
    kubectl patch secret cano-collector-secrets -n monitoring \
      --patch='{"data":{"slack-webhook":"new-base64-encoded-value"}}'

Performance Analysis
~~~~~~~~~~~~~~~~~~~

Analyze performance trends:

.. code-block:: bash

    # Export metrics for analysis
    curl -s http://cano-collector.monitoring.svc.cluster.local:8080/metrics > metrics-$(date +%Y%m%d).txt

    # Check processing latency trends
    # Review Grafana dashboards for trends

    # Analyze destination performance
    curl -s http://cano-collector.monitoring.svc.cluster.local:8080/metrics | grep cano_destination_duration

Backup Verification
~~~~~~~~~~~~~~~~~~~

Verify backup integrity:

.. code-block:: bash

    # Test configuration restore
    kubectl apply -f backup-$(date +%Y%m%d).yaml --dry-run=client

    # Verify backup completeness
    ls -la backup-*.yaml

    # Test recovery procedure in staging environment

Quarterly Maintenance
---------------------

Comprehensive Review
~~~~~~~~~~~~~~~~~~~

Perform comprehensive system review:

1. **Architecture Review:**
   - Review current configuration
   - Identify optimization opportunities
   - Plan for scaling needs

2. **Security Audit:**
   - Review RBAC permissions
   - Check network policies
   - Audit secret management

3. **Performance Optimization:**
   - Analyze resource usage patterns
   - Optimize resource limits
   - Review scaling policies

4. **Documentation Update:**
   - Update runbooks
   - Review procedures
   - Update configuration examples

Upgrade Planning
~~~~~~~~~~~~~~~~

Plan for upgrades:

.. code-block:: bash

    # Check current version
    kubectl exec -n monitoring deployment/cano-collector -- cano-collector --version

    # Check for new versions
    helm search repo cano-collector/cano-collector

    # Test upgrade in staging
    helm upgrade cano-collector cano-collector/cano-collector \
      --namespace monitoring \
      --dry-run

Capacity Planning
~~~~~~~~~~~~~~~~~

Assess capacity needs:

.. code-block:: bash

    # Analyze alert volume trends
    # Review processing capacity
    # Plan for growth

    # Check current limits
    kubectl describe deployment cano-collector -n monitoring | grep -A 5 Resources

Emergency Procedures
--------------------

Service Outage Response
~~~~~~~~~~~~~~~~~~~~~~

If cano-collector is down:

1. **Immediate Actions:**
   .. code-block:: bash

       # Check pod status
       kubectl get pods -n monitoring -l app=cano-collector

       # Check events
       kubectl get events -n monitoring --sort-by='.lastTimestamp'

       # Check logs
       kubectl logs -n monitoring -l app=cano-collector --previous

2. **Quick Recovery:**
   .. code-block:: bash

       # Restart deployment
       kubectl rollout restart deployment/cano-collector -n monitoring

       # Scale up if needed
       kubectl scale deployment cano-collector --replicas=2 -n monitoring

3. **Fallback Plan:**
   - Configure Alertmanager to send to backup notification system
   - Use direct webhook URLs as temporary solution

Configuration Emergency
~~~~~~~~~~~~~~~~~~~~~~~

If configuration is corrupted:

1. **Restore from Backup:**
   .. code-block:: bash

       # Restore configuration
       kubectl apply -f backup-$(date +%Y%m%d).yaml

       # Restart to reload configuration
       kubectl rollout restart deployment/cano-collector -n monitoring

2. **Emergency Configuration:**
   - Create minimal working configuration
   - Restore full configuration after service is stable

Performance Emergency
~~~~~~~~~~~~~~~~~~~~~

If performance is degraded:

1. **Immediate Actions:**
   .. code-block:: bash

       # Scale up
       kubectl scale deployment cano-collector --replicas=3 -n monitoring

       # Increase resource limits
       kubectl patch deployment cano-collector -n monitoring \
         --patch='{"spec":{"template":{"spec":{"containers":[{"name":"cano-collector","resources":{"limits":{"memory":"1Gi","cpu":"1000m"}}}]}}}}'

2. **Investigation:**
   - Check for resource constraints
   - Analyze processing bottlenecks
   - Review destination performance

Automated Maintenance
---------------------

Cron Jobs
~~~~~~~~~

Set up automated maintenance tasks:

.. code-block:: yaml

    apiVersion: batch/v1
    kind: CronJob
    metadata:
      name: cano-collector-backup
      namespace: monitoring
    spec:
      schedule: "0 2 * * *"  # Daily at 2 AM
      jobTemplate:
        spec:
          template:
            spec:
              containers:
              - name: backup
                image: bitnami/kubectl
                command:
                - /bin/sh
                - -c
                - |
                  kubectl get configmap cano-collector-config -n monitoring -o yaml > /backup/backup-$(date +%Y%m%d).yaml
                  kubectl get secret cano-collector-secrets -n monitoring -o yaml > /backup/secrets-$(date +%Y%m%d).yaml
              volumes:
              - name: backup
                persistentVolumeClaim:
                  claimName: backup-pvc
              restartPolicy: OnFailure

Monitoring Alerts
~~~~~~~~~~~~~~~~~

Set up alerts for maintenance tasks:

.. code-block:: yaml

    - alert: CanoCollectorBackupFailed
      expr: time() - cano_backup_last_success_timestamp > 86400
      for: 1h
      labels:
        severity: warning
      annotations:
        summary: "Cano-collector backup failed"
        description: "Backup has not completed successfully in 24 hours"

    - alert: CanoCollectorConfigOld
      expr: time() - cano_config_last_modified_timestamp > 2592000
      for: 1h
      labels:
        severity: info
      annotations:
        summary: "Cano-collector configuration is old"
        description: "Configuration has not been updated in 30 days"

Maintenance Checklist
---------------------

Daily Checklist
~~~~~~~~~~~~~~~

- [ ] Check pod status
- [ ] Verify health endpoint
- [ ] Review error logs
- [ ] Check metrics

Weekly Checklist
~~~~~~~~~~~~~~~

- [ ] Backup configuration
- [ ] Analyze logs
- [ ] Review resource usage
- [ ] Validate configuration

Monthly Checklist
~~~~~~~~~~~~~~~~~

- [ ] Security review
- [ ] Performance analysis
- [ ] Backup verification
- [ ] Documentation update

Quarterly Checklist
~~~~~~~~~~~~~~~~~~~

- [ ] Comprehensive review
- [ ] Upgrade planning
- [ ] Capacity planning
- [ ] Emergency procedure review 