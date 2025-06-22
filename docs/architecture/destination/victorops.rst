VictorOps Destination
=====================

The VictorOps Destination prepares data for the VictorOps REST API. Unlike other incident management destinations, VictorOps uses a simpler approach without built-in lifecycle management.

Responsibilities
----------------

-   **Data Preparation**: The destination extracts the `fingerprint` to be used as the `entity_id` for deduplication. It also prepares the REST endpoint URL and any additional metadata needed by the sender.

-   **Message Type Assignment**: It sets the `message_type` to "CRITICAL" for all alerts, as VictorOps does not distinguish between different alert states in the same way as PagerDuty or OpsGenie.

-   **Delegation**: After preparing the basic context, it delegates the construction of the JSON payload and the HTTP call to the `VictorOpsSender`.

Key Implementation Details
--------------------------

-   **Simplified Lifecycle**: VictorOps does not have built-in incident lifecycle management like PagerDuty or OpsGenie. All alerts are sent as "CRITICAL" messages, and manual intervention is required to resolve incidents.

-   **REST API**: Unlike other destinations that use dedicated SDKs, VictorOps uses a simple REST endpoint that accepts a JSON payload with all alert information. 