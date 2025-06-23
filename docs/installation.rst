Installation Guide
=================

This guide provides detailed instructions for installing cano-collector in different environments.

Installation Methods
-------------------

Cano-collector can be installed using:

1. **Helm Chart** (Recommended)
2. **Kubernetes Manifests**
3. **Docker Container**

Helm Installation
-----------------

Prerequisites
~~~~~~~~~~~~~

- Kubernetes 1.19+
- Helm 3.0+
- kubectl configured
- Alertmanager running

Step 1: Add Helm Repository
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    helm repo add cano-collector https://your-org.github.io/cano-collector
    helm repo update

Step 2: Create Configuration
~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Create a `values.yaml` file with your configuration:

.. code-block:: yaml

    # Global settings
    replicaCount: 2
    image:
      repository: your-registry/cano-collector
      tag: "latest"
      pullPolicy: IfNotPresent

    # Service configuration
    service:
      type: ClusterIP
      port: 8080

    # Resource limits
    resources:
      requests:
        memory: "128Mi"
        cpu: "100m"
      limits:
        memory: "256Mi"
        cpu: "200m"

    # Destinations configuration
    destinations:
      slack:
        - name: "alerts-prod"
          webhookURL: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
        - name: "alerts-dev"
          webhookURL: "https://hooks.slack.com/services/YOUR/DEV/WEBHOOK"
      
      msteams:
        - name: "ops-team"
          webhookURL: "https://your-org.webhook.office.com/webhookb2/YOUR/WEBHOOK"

    # Teams configuration
    teams:
      - name: "production"
        destinations:
          - "alerts-prod"
          - "ops-team"
      - name: "development"
        destinations:
          - "alerts-dev"

Step 3: Install the Chart
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    # Create namespace
    kubectl create namespace monitoring

    # Install cano-collector
    helm install cano-collector cano-collector/cano-collector \
      --values values.yaml \
      --namespace monitoring

Step 4: Verify Installation
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    # Check pods
    kubectl get pods -n monitoring -l app=cano-collector

    # Check services
    kubectl get svc -n monitoring -l app=cano-collector

    # Check logs
    kubectl logs -n monitoring -l app=cano-collector

Kubernetes Manifests Installation
---------------------------------

If you prefer to use raw Kubernetes manifests:

Step 1: Download Manifests
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    git clone https://github.com/your-org/cano-collector
    cd cano-collector/helm/cano-collector/templates

Step 2: Customize Configuration
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Edit the ConfigMap and Secret files with your configuration:

.. code-block:: yaml

    # configmap.yaml
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: cano-collector-config
    data:
      destinations.yaml: |
        destinations:
          slack:
            - name: "alerts-prod"
              webhookURL: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

Step 3: Apply Manifests
~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    kubectl apply -f namespace.yaml
    kubectl apply -f configmap.yaml
    kubectl apply -f secret.yaml
    kubectl apply -f deployment.yaml
    kubectl apply -f service.yaml

Docker Installation
-------------------

For local development or testing:

Step 1: Build Image
~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    docker build -t cano-collector:latest .

Step 2: Run Container
~~~~~~~~~~~~~~~~~~~~~

.. code-block:: bash

    docker run -d \
      --name cano-collector \
      -p 8080:8080 \
      -v $(pwd)/config:/etc/cano-collector \
      cano-collector:latest

Configuration Files
-------------------

Cano-collector uses two main configuration files:

1. **destinations.yaml** - Defines notification endpoints
2. **teams.yaml** - Defines team routing rules

Example destinations.yaml:
~~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

    destinations:
      slack:
        - name: "alerts-prod"
          webhookURL: "https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"
          channel: "#alerts"
          username: "Cano Collector"
      
      msteams:
        - name: "ops-team"
          webhookURL: "https://your-org.webhook.office.com/webhookb2/YOUR/WEBHOOK"
          title: "Kubernetes Alert"
      
      jira:
        - name: "production-issues"
          url: "https://your-org.atlassian.net"
          username: "jira-bot@your-org.com"
          apiToken: "your-api-token"
          projectKey: "OPS"
          issueType: "Incident"

Example teams.yaml:
~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

    teams:
      - name: "production"
        destinations:
          - "alerts-prod"
          - "ops-team"
          - "production-issues"
      
      - name: "development"
        destinations:
          - "alerts-dev"

Alertmanager Integration
-----------------------

Configure Alertmanager to send alerts to cano-collector:

.. code-block:: yaml

    receivers:
      - name: 'cano-collector'
        webhook_configs:
          - url: 'http://cano-collector.monitoring.svc.cluster.local:8080/api/alerts'
            send_resolved: true
            timeout: 10s

    route:
      receiver: 'cano-collector'
      group_by: ['alertname', 'namespace']
      group_wait: 30s
      group_interval: 5m
      repeat_interval: 4h

Troubleshooting Installation
---------------------------

Common Issues
~~~~~~~~~~~~

1. **Pod not starting**
   - Check resource limits
   - Verify image pull permissions
   - Check configuration syntax

2. **Configuration not loaded**
   - Verify ConfigMap exists
   - Check file paths in deployment
   - Validate YAML syntax

3. **Alerts not received**
   - Verify Alertmanager configuration
   - Check network connectivity
   - Validate webhook URLs

Debug Commands
~~~~~~~~~~~~~

.. code-block:: bash

    # Check pod status
    kubectl describe pod -n monitoring -l app=cano-collector

    # Check logs
    kubectl logs -n monitoring -l app=cano-collector -f

    # Check configuration
    kubectl get configmap -n monitoring cano-collector-config -o yaml

    # Test webhook endpoint
    curl -X POST http://cano-collector.monitoring.svc.cluster.local:8080/health 