Kafka Sender
============

The `KafkaSender` publishes messages to Kafka topics, converting `Issue` data into structured JSON format suitable for data processing pipelines.

Responsibilities
----------------

-   **Kafka Producer Management**: It manages the Kafka producer connection and handles message publishing to the configured topic.

-   **JSON Serialization**: It converts supported enrichment blocks (`KubernetesDiffBlock` and `JsonBlock`) into structured JSON messages with metadata like cluster name, resource information, and change details.

-   **Message Publishing**: It handles the actual publishing of messages to Kafka topics, including error handling and connection management.

-   **Authentication Support**: It supports various Kafka authentication mechanisms (SASL, SSL) through the configuration parameters.

Key Implementation Details
--------------------------

-   **Structured Data Focus**: The sender is designed for machine-readable data rather than human notifications, converting changes into structured JSON with clear property mappings.

-   **Change Tracking**: For `KubernetesDiffBlock`, it creates detailed change records showing the old and new values for each modified property.

-   **JSON Enhancement**: For `JsonBlock`, it adds cluster metadata to the existing JSON structure before publishing.

-   **High-Throughput**: Kafka is optimized for high-volume message streaming, making it suitable for real-time change tracking and data pipeline integration.

Formatting
----------

There is no visual formatting involved. The entire `Issue` object, including all its fields and `Enrichment` data, is serialized into a **JSON string**.

Functionality
-------------

- **Data Serialization**: The sender's sole responsibility is to convert the `Issue` object into a JSON payload.
- **Message Publishing**: It then connects to a Kafka broker and publishes this JSON payload as a message to a specified topic.

Use Cases
---------

The Kafka Sender is ideal for:

- **Archiving**: Storing all generated issues in a long-term, durable message queue.
- **Data Analysis**: Feeding issues into a data lake or analytics platform (like ELK, Splunk, or a custom data warehouse) for trend analysis and reporting.
- **Custom Workflows**: Triggering complex, custom automation workflows in other backend systems that consume messages from the Kafka topic.

This sender is not meant for human eyes but as a machine-to-machine integration point. 