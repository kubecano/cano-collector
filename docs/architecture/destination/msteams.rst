MS Teams Destination
====================

The MS Teams Destination manages the dispatch of `Issue` objects to a specific Microsoft Teams channel. Its role is primarily to hold configuration and delegate sending to the `MsTeamsSender`.

Responsibilities
----------------

-   **Configuration Management**: Its main responsibility is to hold the configuration for a specific MS Teams channel, which consists of the `webhookURL`.

-   **Webhook Overriding**: A key feature of this destination is the ability to dynamically override the configured `webhookURL` based on an annotation on the Kubernetes resource related to the issue. This allows for flexible, on-the-fly routing of alerts to different Teams channels without changing the central configuration. For example, you can add an annotation like `msteams.webhook.url/alerts: "https://..."` to a Deployment to send its alerts to a specific channel.

-   **Delegation to Sender**: The destination receives an `Issue` from the routing engine and immediately delegates the task of formatting and sending the message to the `MsTeamsSender`.

This destination acts as a strategic "pipe" to the sender, with its most advanced feature being the dynamic routing via annotations. It does not implement complex logic like notification grouping. 