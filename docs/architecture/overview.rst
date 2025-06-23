Architecture Overview
=====================

The collector's architecture is designed to be modular and extensible, allowing for easy integration with various notification systems and internal tools. It is composed of three main concepts: the Data Model (`Issue`), Destinations, and Senders.

Core Components
---------------

- **Data Model (`Issue`)**: The `Issue` is the central data structure that represents any event, alert, or finding within the system. It contains all the necessary information, from the title and severity to detailed contextual `Enrichments`. This standardized object is passed through the processing pipeline. The model is defined in `pkg/core/issue/*`.

- **Destination**: A Destination is a logical target for an `Issue`. It represents a configured instance of a notification channel, such as a specific Slack channel or a PagerDuty service. Destinations are responsible for high-level logic, such as routing and deciding if and when an issue should be sent. Destinations are defined in `pkg/destination/`.

- **Sender**: A Sender is responsible for the final step: formatting the `Issue` object into the specific format required by the target API and sending it. For example, the `SlackSender` converts an `Issue` into a Slack Block Kit message, while the `PagerdutySender` transforms it into a PagerDuty API event payload. Senders contain the implementation details for each integration and are located in `pkg/sender/`.

Flow of an Issue
----------------

1.  **Creation**: An issue is created by a detector component within the collector.
2.  **Processing**: The issue is enriched with additional context (logs, resource status, etc.).
3.  **Routing**: The collector determines which configured `Destination`(s) should receive this issue based on routing rules.
4.  **Dispatching**: Each matched `Destination` uses its corresponding `Sender` to format and deliver the `Issue` to the external system.

This decoupled architecture allows for great flexibility. For instance, to support a new notification service, one only needs to implement a new `Sender` and a corresponding `Destination` configuration, without altering the core issue generation or enrichment logic. 