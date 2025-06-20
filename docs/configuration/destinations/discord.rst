Discord
=======

Sends notifications to a Discord channel via a webhook.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      discord:
        - name: "my-discord-channel"
          url: "https://discord.com/api/webhooks/..."

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`url`** (string, required)
    The webhook URL provided by Discord for your channel. 