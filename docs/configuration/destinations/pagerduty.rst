PagerDuty
=========

Triggers, acknowledges, and resolves incidents in PagerDuty using the Events API v2.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      pagerduty:
        - name: "my-pagerduty-service"
          integrationKey: "your-pagerduty-integration-key"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`integrationKey`** (string, required)
    The Integration Key (also known as Routing Key) from a PagerDuty service integration using the "Events API v2". 