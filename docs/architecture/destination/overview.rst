Destinations Overview
=====================

A **Destination** is a configured instance of a notification channel. It acts as a bridge between the collector's core logic and a specific `Sender`. While a `Sender` knows *how* to talk to an API (e.g., Slack's API), a `Destination` represents a concrete endpoint, like the `#alerts-prod` Slack channel or a specific PagerDuty service.

Role of a Destination
---------------------

The primary responsibilities of a Destination are:

1.  **Configuration Holding**: A Destination holds the specific configuration needed for its `Sender`. This includes API tokens, webhook URLs, channel IDs, service keys, and other parameters.
2.  **Dispatching**: It implements the logic for dispatching an `Issue`. In most cases, this is a straightforward delegation to its configured `Sender`.
3.  **Abstraction**: It provides a uniform interface for the rest of the application. The routing engine doesn't need to know the details of a `SlackSender` versus an `OpsGenieSender`; it simply tells a `Destination` to handle an `Issue`.

How It Works
------------

When the application starts, a `DestinationFactory` reads the destination configurations provided by the user (e.g., in the Helm `values.yaml`). For each configured destination, it:

1.  Instantiates the correct `Sender` type (e.g., `SlackSender`).
2.  Injects the specific configuration (e.g., the webhook URL for the `#alerts-prod` channel) into the `Sender`.
3.  Wraps the configured `Sender` in a `Destination` object.

This `Destination` object is then registered and made available to the issue routing engine. The core logic is intentionally kept thin within the `Destination` itself; the heavy lifting of API communication and message formatting is handled by the `Sender`. For more details on the specific formatting logic, refer to the documentation for each `Sender`. 