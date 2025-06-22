VictorOps Sender
================

The `VictorOpsSender` communicates with the VictorOps REST API using a simple JSON payload structure. It receives data from the `VictorOpsDestination` and handles the final payload construction and HTTP communication.

Responsibilities
----------------

-   **REST API Communication**: It sends HTTP POST requests to the configured VictorOps REST endpoint with a JSON payload containing all alert information.

-   **Payload Construction**: It builds a flat JSON object where each piece of information becomes a top-level field:
    -   `entity_id`: Uses the issue's `fingerprint` for deduplication
    -   `entity_display_name`: The formatted title with severity emoji
    -   `state_message`: Contains the description and all enrichments converted to plain text
    -   `message_type`: Always set to "CRITICAL"
    -   `monitoring_tool`: Set to "Robusta"
    -   Custom fields for resource details (Resource, Source, Namespace, Node)

-   **Link Handling**: Links from the issue are added as fields with the prefix `vo_annotate.u.`, which VictorOps displays as clickable annotations.

Key Implementation Details
--------------------------

-   **Flattened Structure**: Unlike other senders that use nested objects, VictorOps expects all information in a flat JSON structure where each field represents a piece of data.

-   **No Lifecycle Management**: VictorOps does not support automatic incident resolution. All alerts are sent as "CRITICAL" and require manual intervention to close.

-   **Text Conversion**: All enrichment blocks (tables, markdown, etc.) are converted to plain text and combined into the `state_message` field. 