Kafka Sender
============

The Kafka Sender is a specialized sender used for data pipelining rather than direct user notification. It sends the full `Issue` object to an Apache Kafka topic.

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