DataDog Destination
===================

The DataDog Destination prepares data for the DataDog Events API, focusing on creating structured events that can be viewed in the DataDog dashboard.

Responsibilities
----------------

-   **Event Preparation**: The destination extracts key information from the `Issue` and prepares it for the DataDog Events API, including the `aggregation_key` for event grouping and severity mapping.

-   **DataDog API Context**: It prepares the API key and configuration needed by the `DataDogSender` to communicate with the DataDog Events API.

-   **Delegation**: After preparing the basic context, it delegates the construction of the DataDog event payload and the API communication to the `DataDogSender`.

Key Implementation Details
--------------------------

-   **Event-Based Architecture**: Unlike incident management systems, DataDog uses an event-based model where each alert becomes an event in the DataDog timeline.

-   **Aggregation Key**: The destination uses the issue's `aggregation_key` to group related events in DataDog, ensuring that similar alerts are grouped together in the dashboard.

-   **Severity Mapping**: It maps Robusta severity levels to DataDog event types (`error` for HIGH severity, `info` for others), which affects how events are displayed and filtered in DataDog. 