OpsGenie Sender
===============

The OpsGenie Sender is designed to integrate with Atlassian's OpsGenie incident management platform. Its primary goal is to create, update, or resolve alerts in OpsGenie based on an `Issue` object.

Formatting
----------

Unlike UI-focused senders like Slack or MS Teams, the OpsGenie Sender does not generate a rich visual message. Instead, it maps the `Issue` data to the fields of the **OpsGenie Alert API**.

Field Mapping
~~~~~~~~~~~~~

The `Issue` object's properties are mapped directly to OpsGenie alert fields:

- **`issue.Title`**: Mapped to the `message` field of the OpsGenie alert.
- **`issue.Description`**: Mapped to the `description` field.
- **`issue.Severity`**: Converted to OpsGenie's priority levels (e.g., `P1`, `P2`, `P3`).
- **`issue.AggregationKey`**: Used as the `alias` for alert deduplication. OpsGenie will use this key to group related alerts or to resolve an existing alert when a `RESOLVED` status is received.
- **`issue.Subject` and other labels**: Mapped to OpsGenie `tags` for filtering and routing.

Handling Enrichments
--------------------

All `Enrichment` blocks (`MarkdownBlock`, `TableBlock`, etc.) are serialized into a human-readable text format. This text is then appended to the main `description` of the OpsGenie alert. The goal is to provide context for the on-call engineer directly within the incident, not to render a complex UI.

Key Functionality
-----------------

- **Deduplication and Resolution**: The sender's main strength is its use of the `alias` field. This ensures that a firing alert creates a new incident, subsequent occurrences of the same alert are deduplicated, and a `RESOLVED` issue automatically closes the corresponding incident in OpsGenie.
- **Focus on Incident Management**: The integration is built for operational efficiency. It provides the essential information needed to begin triage within a dedicated incident management tool, rather than for general-purpose messaging. 