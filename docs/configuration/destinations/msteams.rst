MS Teams
========

Sends notifications to a Microsoft Teams channel via an incoming webhook.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      msteams:
        - name: "my-msteams-destination"
          webhookURL: "https://your-org.webhook.office.com/webhookb2/..."

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`webhookURL`** (string, required)
    The incoming webhook URL provided by the MS Teams channel connector. 