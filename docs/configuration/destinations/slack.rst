Slack
=====

Sends notifications to a Slack channel. This is the most feature-rich destination, offering interactive messages, file attachments, and notification grouping.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      slack:
        - name: "my-slack-destination"
          api_key: "xoxb-YOUR-SLACK-BOT-TOKEN"
          slack_channel: "my-alerts-channel"
          unfurl_links: true    # Optional: Defaults to true.
          grouping_interval: 60 # Optional: Time in seconds to group notifications. Defaults to 0 (disabled).

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`api_key`** (string, required)
    The Slack Bot User OAuth Token, starting with `xoxb-`. This is required for advanced features like file uploads, message updates, and interactivity.

-   **`slack_channel`** (string, required)
    The name of the Slack channel to send notifications to (e.g., `#my-channel`).

-   **`grouping_interval`** (integer, optional)
    Default: `0`. The time in seconds to wait and group multiple issues into a single summary message, with individual alerts posted in a thread. This helps to reduce channel noise. If set to `0`, each issue is sent as a separate message immediately.

-   **`unfurl_links`** (boolean, optional)
    Default: `true`. If `true`, links in the notification will be unfurled by Slack to show a preview. Set to `false` to disable this.

-   **`webhookURL`** (string, optional - alternative to `api_key`)
    For simple, non-interactive notifications, you can use a traditional Slack Incoming Webhook URL. If you use this, functionality will be limited (e.g., no file uploads, no grouping, no message updates). It is highly recommended to use the `api_key` method for the best experience.

.. note::
    Using the `api_key` method is strongly recommended to enable all features like log uploads, message grouping with threading, and future interactive components. 