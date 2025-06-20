PagerDuty Sender
================

The PagerDuty Sender integrates with the PagerDuty incident response platform. Its function is to trigger, acknowledge, or resolve PagerDuty incidents based on the `Issue` object.

Formatting
----------

The sender uses the **PagerDuty Events API v2**. This API is event-driven and expects a specific JSON payload. The focus is on providing structured data for routing and incident management, not on creating a visual message.

Field Mapping
~~~~~~~~~~~~~

The properties of the `Issue` object are mapped to the fields of a PagerDuty event payload:

- **`issue.Status`**: Determines the `event_action`. `FIRING` maps to `trigger`, and `RESOLVED` maps to `resolve`.
- **`issue.AggregationKey`**: Used as the `dedup_key`. This is crucial for PagerDuty's deduplication logic, ensuring that multiple alerts for the same problem are grouped into a single incident and that a `resolve` event closes the correct incident.
- **`issue.Title`**: Mapped to the `payload.summary`.
- **`issue.Source`**: Mapped to the `payload.source`.
- **`issue.Severity`**: Mapped to the `payload.severity` field (e.g., `critical`, `warning`, `info`).

Handling Enrichments
--------------------

`Enrichment` blocks are handled by serializing them into a JSON object. This object is then sent in the `payload.custom_details` field of the PagerDuty event.

This approach preserves the structure of the enrichment data, making it available within the PagerDuty UI's "Details" section. While not rendered as a rich UI, the data is accessible to on-call engineers who need to investigate the incident further.

Key Functionality
-----------------

- **Lifecycle Management**: The sender fully manages the incident lifecycle in PagerDuty by correctly mapping the issue status to `trigger` and `resolve` actions.
- **Structured Data for Triage**: By placing detailed context in `custom_details`, it provides all necessary information for an engineer to begin troubleshooting directly within the PagerDuty incident, without needing to switch to other tools immediately.
- **Deduplication**: Correct use of the `dedup_key` is fundamental to preventing alert storms and ensuring a clean incident timeline. 