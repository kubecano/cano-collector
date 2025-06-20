Slack
=====

Sends notifications to a Slack channel. This is the most feature-rich destination, offering interactive messages and detailed formatting.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      slack:
        - name: "my-slack-destination"
          api_key: "xoxb-YOUR-SLACK-BOT-TOKEN"
          slack_channel: "my-alerts-channel"
          grouping_interval: 0  # Optional: Time in seconds to group notifications. 0 means no grouping.
          unfurl_links: true    # Optional: Whether to unfurl links in messages.

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`api_key`** (string, required)
    The Slack Bot User OAuth Token. It should start with `xoxb-`. This is required for advanced features like file uploads (for logs) and interactivity.

-   **`slack_channel`** (string, required)
    The name of the Slack channel to send notifications to (e.g., `#my-channel`).

-   **`grouping_interval`** (integer, optional)
    The time in seconds to wait and group multiple issues into a single summary message. If set to `0` (the default), each issue is sent as a separate message immediately.

-   **`unfurl_links`** (boolean, optional)
    If `true` (the default), links in the notification will be unfurled by Slack to show a preview. Set to `false` to disable this.

-   **`webhookURL`** (string, optional - alternative to `api_key`)
    For simple, non-interactive notifications, you can use a traditional Slack Incoming Webhook URL instead of setting up a bot with an `api_key`. Functionality will be limited (e.g., no log uploads).

.. note::
    It is highly recommended to use the `api_key` method for the best experience, as it enables all formatting and interactive features. 