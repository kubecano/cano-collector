Webhook Sender
==============

The Webhook Sender is a generic sender that posts the raw `Issue` object to any specified HTTP/S endpoint. It is designed for maximum flexibility and integration with custom-built tools.

Formatting
----------

There is no visual or structural formatting applied by this sender.

- **Payload**: The entire `Issue` object, including all its fields and `Enrichment` data, is serialized into a single **JSON payload**.
- **Method**: The payload is sent as the body of an **HTTP POST** request.
- **Headers**: The sender allows for custom HTTP headers to be included in the request, which is often necessary for authentication (e.g., using an `Authorization: Bearer <token>` header).

Use Cases
---------

- **Custom Integrations**: Connect to in-house applications or scripts that can process JSON a and trigger custom workflows.
- **Serverless Functions**: Trigger cloud functions (e.g., AWS Lambda, Google Cloud Functions) to perform custom actions in response to an issue.
- **Prototyping**: Quickly test and debug issue generation by pointing the webhook to a request-capturing service like webhook.site. 