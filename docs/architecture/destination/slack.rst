Slack Destination
=================

The Slack Destination is responsible for managing and dispatching `Issue` objects to specific Slack channels. It orchestrates the process of sending alerts, including the business logic for grouping notifications.

Responsibilities
----------------

-   **Configuration Management**: Holds the specific configuration for a Slack endpoint, including the channel name (`slack_channel`), authentication details (`api_key` or `webhookURL`), and formatting preferences (`unfurl_links`).

-   **Notification Grouping**: Implements a key feature of alert management: grouping. By setting the `grouping_interval` parameter, multiple related issues can be buffered and then sent as a single, clean summary message. This logic resides entirely within the destination and is its primary strategic responsibility. If an update to a grouped alert arrives, the destination will manage the process of updating the summary message in the thread.

-   **Delegation to Sender**: Once an issue (or a group of issues) is ready to be sent, the Slack Destination delegates the final formatting and API communication to the `SlackSender`. It passes the `Issue` object and any relevant parameters to the sender. For grouped messages, it will also pass the `thread_ts` to ensure replies are correctly threaded.

How It Works
------------

1.  The `DestinationFactory` creates a `SlackDestination` instance for each entry in the `destinations.slack` configuration.
2.  The destination is configured with all its parameters.
3.  When the routing engine dispatches an `Issue` to this destination, its internal logic takes over.
4.  If `grouping_interval` is greater than zero, the destination holds the issue in a buffer, waiting to see if other related issues arrive within the time window. After the interval, it sends a summary and then sends each individual issue as a reply in the thread of the summary message.
5.  If grouping is disabled, the issue is passed to the `SlackSender` for immediate dispatch. 