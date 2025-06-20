VictorOps
=========

Triggers and resolves incidents in Splunk On-Call (formerly VictorOps).

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      victorops:
        - name: "my-victorops-team"
          api_key: "your-victorops-api-key"
          routing_key: "my-team-routing-key"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`api_key`** (string, required)
    Your VictorOps API key.

-   **`routing_key`** (string, required)
    The routing key to direct the incident to the correct team or escalation policy. 