ServiceNow Sender
=================

The ServiceNow Sender integrates with the ServiceNow platform to create incidents based on `Issue` objects. It is designed for enterprise ITSM workflows.

Formatting
----------

Similar to other incident management senders (PagerDuty, OpsGenie), the ServiceNow Sender does not focus on rich visual formatting. It maps `Issue` data to the fields of a **ServiceNow Incident record** via the Table API.

Field Mapping
~~~~~~~~~~~~~

- **`issue.Title`**: Mapped to the `short_description` field.
- **`issue.Description`**: Mapped to the `description` field.
- **`issue.Severity`**: Converted to ServiceNow's `impact` and `urgency` fields, which in turn determine the `priority`.
- **`issue.AggregationKey`**: Used for deduplication. The sender may query ServiceNow for existing incidents with a matching correlation ID before creating a new one.
- **`Enrichments`**: All enrichment data is serialized into a text format and appended to the `work_notes` or `comments` of the incident, providing context for the service desk agent.

Key Functionality
-----------------

- **ITSM Integration**: The primary purpose is to bridge cloud-native monitoring with traditional IT Service Management processes.
- **Incident Creation**: It automates the creation of incident tickets, reducing manual effort and response times.
- **Structured Data for Triage**: Provides the necessary information within the ServiceNow incident record for the support team to begin their investigation. 