Data & Ticketing Destinations
==============================

This category includes destinations designed for data integration, workflow automation, and ticketing rather than real-time user notification. This includes **Jira**, **DataDog**, **Kafka**, **ServiceNow**, and the generic **Webhook**.

Shared Responsibilities
-----------------------

-   **Data Transformation Logic**: Unlike simple chat destinations, these destinations may contain logic to prepare the `Issue` for the target system. For example, a `JiraDestination` would be responsible for mapping the issue's severity to a Jira priority field or adding specific labels before passing the data to the `JiraSender`.

-   **Configuration Holding**: Each destination holds the endpoint and authentication configuration required by its sender (e.g., Jira project key, Kafka broker addresses, DataDog API key).

-   **One-Way Dispatch**: These destinations typically represent a one-way data flow. An `Issue` is received and immediately dispatched to the sender for processing. There is usually no concept of "resolving" a ticket or an event in the same way as in an incident management platform, so the `status` field is often just another piece of data to be recorded.

-   **Delegation**: As with all destinations, the final task of API communication and payload formatting is delegated to the corresponding sender. For instance, the `KafkaDestination` tells the `KafkaSender` to serialize the issue to JSON and publish it, while the `JiraDestination` tells the `JiraSender` to construct and send the API request to create an issue.

The primary role of these destinations is to act as a configured, strategic bridge between the collector's internal `Issue` format and an external system's API, with the destination itself handling any necessary pre-transformation logic. 