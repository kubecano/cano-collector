DataDog Destination
==================

The DataDog destination allows cano-collector to send alerts and findings to DataDog as events.

Features
--------

-   **Event Creation**: Creates DataDog events for alerts and findings
-   **Severity Mapping**: It maps severity levels to DataDog event types (`error` for HIGH severity, `info` for others), which affects how events are displayed and filtered in DataDog.
-   **Tag Management**: Automatically adds relevant tags for filtering and grouping
-   **Alert Correlation**: Links related events and alerts
-   **Custom Attributes**: Supports custom event attributes
-   **Metric Integration**: Can trigger metric-based alerts

Configuration
-------------

.. code-block:: yaml

    destinations:
      datadog:
        - name: "production-events"
          apiKey: "your-datadog-api-key"
          appKey: "your-datadog-app-key"
          site: "datadoghq.com"
          tags:
            - "env:production"
            - "team:platform"
          severityMapping:
            critical: "error"
            warning: "warning"
            info: "info" 