OpsGenie
========

Creates, updates, and resolves alerts in Atlassian OpsGenie.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      opsgenie:
        - name: "my-opsgenie-service"
          apiKey: "your-opsgenie-api-key"
          region: "eu"  # Optional: "eu" or "us". Defaults to global.

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`apiKey`** (string, required)
    The API key from your OpsGenie integration. It must have permissions to create and close alerts.

-   **`region`** (string, optional)
    The OpsGenie region to send API requests to. Can be `eu` or `us`. If omitted, the default global API endpoint (`api.opsgenie.com`) is used. 