VictorOps Sender
================

The VictorOps Sender integrates with Splunk On-Call (formerly VictorOps) to manage the incident lifecycle.

Formatting
----------

The sender uses the **VictorOps REST Endpoint API** to trigger, acknowledge, and resolve incidents.

- **`issue.Status`**: This is mapped to the `message_type` field. `FIRING` becomes `CRITICAL`, and `RESOLVED` becomes `RECOVERY`.
- **`issue.Title`**: Mapped to the `entity_display_name` field, which is the main title of the incident.
- **`issue.AggregationKey`**: Used as the `entity_id`, which is VictorOps's key for alert deduplication and lifecycle management.
- **`issue.Description`** and **`Enrichments`**: This content is serialized into a string and sent in the `state_message`, providing the body of the incident details.

Key Functionality
-----------------

- **Full Incident Lifecycle**: The sender correctly maps issue states to `CRITICAL` and `RECOVERY` message types, ensuring incidents are automatically opened and closed.
- **Deduplication**: Use of the `entity_id` field prevents alert storms by grouping subsequent alerts for the same issue into a single, active incident.
- **On-Call Routing**: Incidents are directed to the correct on-call team based on the `routing_key` provided in the destination configuration. 