Quick Start Guide
================

This guide will help you get cano-collector up and running in under 10 minutes.

Prerequisites
------------

- Kubernetes cluster (local or remote)
- Helm 3.x installed
- kubectl configured
- Alertmanager already configured and running

Step 1: Add the Helm Repository
-------------------------------

.. code-block:: bash

    helm repo add cano-collector https://your-org.github.io/cano-collector
    helm repo update

Step 2: Create Basic Configuration
---------------------------------

Create a basic configuration file `values.yaml`:

.. code-block:: yaml

    # Basic configuration for testing
    destinations:
      slack:
        - name: "test-channel"
          webhookURL: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
      
      msteams:
        - name: "test-team"
          webhookURL: "https://your-org.webhook.office.com/webhookb2/YOUR/WEBHOOK"

    teams:
      - name: "default"
        destinations:
          - "test-channel"
          - "test-team"

Step 3: Install cano-collector
------------------------------

.. code-block:: bash

    helm install cano-collector cano-collector/cano-collector \
      --values values.yaml \
      --namespace monitoring \
      --create-namespace

Step 4: Verify Installation
---------------------------

Check if the pod is running:

.. code-block:: bash

    kubectl get pods -n monitoring -l app=cano-collector

You should see output like:
::

    NAME                              READY   STATUS    RESTARTS   AGE
    cano-collector-7d8f9c4b5-abc12   1/1     Running   0          2m

Step 5: Configure Alertmanager
------------------------------

Add cano-collector as a receiver in your Alertmanager configuration:

.. code-block:: yaml

    receivers:
      - name: 'cano-collector'
        webhook_configs:
          - url: 'http://cano-collector.monitoring.svc.cluster.local:8080/api/alerts'
            send_resolved: true

    route:
      receiver: 'cano-collector'

Step 6: Test with a Sample Alert
--------------------------------

Create a test alert to verify everything works:

.. code-block:: bash

    # Create a test pod that will fail
    kubectl run test-pod --image=busybox --command -- sleep 1

    # Wait for the pod to fail
    sleep 10

    # Check if you received notifications in Slack/MS Teams
    kubectl delete pod test-pod 