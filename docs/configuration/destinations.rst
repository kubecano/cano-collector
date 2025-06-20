Configuring Destinations
==========================

Destinations are the configured endpoints where the collector sends `Issue` notifications. Each destination has a `type` (e.g., `slack`, `msteams`), a unique `name`, and parameters specific to that type.

All destinations are configured under the `destinations` key in your Helm `values.yaml` file.

Slack
-----

Sends notifications to a Slack channel via an incoming webhook.

**Parameters:**

-   `name` (string, required): A unique name for this destination instance, e.g., "alerts-prod-channel".
-   `webhookURL` (string, required): The incoming webhook URL provided by Slack.

**Example:**

.. code-block:: yaml

    # values.yaml
    destinations:
      slack:
        - name: "alerts-prod"
          webhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
        - name: "alerts-dev"
          webhookURL: "https://hooks.slack.com/services/T00000001/B00000001/YYYYYYYYYYYYYYYYYYYYYYYY"

MS Teams
--------

Sends notifications to an MS Teams channel via an incoming webhook.

**Parameters:**

-   `name` (string, required): A unique name for this destination, e.g., "qa-team-channel".
-   `webhookURL` (string, required): The incoming webhook URL from your MS Teams channel connector.

**Example:**

.. code-block:: yaml

    # values.yaml
    destinations:
      msteams:
        - name: "devops-notifications"
          webhookURL: "https://your-org.webhook.office.com/webhookb2/..."

OpsGenie
--------

Creates alerts in OpsGenie. This integration is designed for incident management.

**Parameters:**

-   `name` (string, required): A unique name for this destination.
-   `apiKey` (string, required): Your OpsGenie API key with alert creation permissions.
-   `region` (string, optional): The OpsGenie region (e.g., `eu`, `us`). Defaults to the global API endpoint if not specified.

**Example:**

.. code-block:: yaml

    # values.yaml
    destinations:
      opsgenie:
        - name: "opsgenie-critical-alerts"
          apiKey: "your-opsgenie-api-key"
          region: "eu"

PagerDuty
---------

Triggers and resolves incidents in PagerDuty.

**Parameters:**

-   `name` (string, required): A unique name for this destination.
-   `integrationKey` (string, required): The integration key (also known as the routing key) from your PagerDuty service integration (Events API v2).

**Example:**

.. code-block:: yaml

    # values.yaml
    destinations:
      pagerduty:
        - name: "pagerduty-on-call"
          integrationKey: "your-32-character-pagerduty-integration-key" 