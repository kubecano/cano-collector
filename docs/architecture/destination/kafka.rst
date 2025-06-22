Kafka Destination
=================

The Kafka Destination prepares data for streaming to Kafka topics, focusing on sending structured change events and JSON data to message queues.

Responsibilities
----------------

-   **Topic Configuration**: The destination prepares the Kafka topic name and connection details needed by the `KafkaSender` to publish messages.

-   **Event Filtering**: It identifies which enrichments contain supported block types (primarily `KubernetesDiffBlock` and `JsonBlock`) for Kafka streaming.

-   **Delegation**: After preparing the Kafka context, it delegates the message serialization and publishing to the `KafkaSender`.

Key Implementation Details
--------------------------

-   **Change-Focused**: Unlike other destinations that send all types of alerts, Kafka is primarily designed for streaming configuration changes and structured data.

-   **Limited Block Support**: The destination only processes specific block types (`KubernetesDiffBlock` and `JsonBlock`), making it ideal for change tracking and data pipeline integration.

-   **Streaming Architecture**: Kafka is designed for high-throughput message streaming rather than human-readable notifications, making it suitable for integration with data processing pipelines. 