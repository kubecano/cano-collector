Slack Destination
=================

The Slack Destination is responsible for managing and dispatching `Issue` objects to specific Slack channels. It orchestrates the process of sending alerts, including logic for grouping and formatting.

Responsibilities
----------------

-   **Configuration Management**: Holds the specific configuration for a Slack endpoint, including the channel name (`slack_channel`) and authentication details (`api_key` or `webhookURL`).

-   **Notification Grouping**: Implements a key feature of alert management: grouping. By setting the `grouping_interval` parameter, multiple related issues can be bundled into a single, clean summary message instead of flooding a channel with individual alerts. This logic resides within the destination, not the sender.

-   **Delegation to Sender**: Once an issue (or a group of issues) is ready to be sent, the Slack Destination delegates the final formatting and API communication to the `SlackSender`. It passes the `Issue` object and any relevant formatting options (like `unfurl_links`) to the sender.

How It Works
------------

1.  The `DestinationFactory` creates a `SlackDestination` instance for each entry in the `destinations.slack` configuration.
2.  The destination is configured with its parameters (`slack_channel`, `api_key`, `grouping_interval`, etc.).
3.  When the routing engine dispatches an `Issue` to this destination, the destination's logic takes over.
4.  If `grouping_interval` is greater than zero, the destination holds the issue in a buffer to see if other related issues arrive within the time window.
5.  Finally, it calls the `SlackSender` to perform the actual sending. 