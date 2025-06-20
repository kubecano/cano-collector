MS Teams Destination
====================

The MS Teams Destination manages the dispatch of `Issue` objects to a Microsoft Teams channel. Its role is straightforward: to hold configuration and delegate to the appropriate sender.

Responsibilities
----------------

-   **Configuration Management**: Its primary role is to hold the configuration for a specific MS Teams channel, which consists of the `webhookURL`.

-   **Delegation to Sender**: The destination receives an `Issue` from the routing engine and immediately delegates the task of formatting and sending the message to the `MsTeamsSender`.

Unlike the Slack destination, the MS Teams destination typically does not implement complex logic like grouping, as the underlying sender and API have fewer features for aggregation. Its main purpose is to act as a configured "pipe" to the sender. 