ServiceNow Sender
=================

The `ServiceNowSender` communicates with the ServiceNow REST API to create and manage incidents. It receives data from the `ServiceNowDestination` and handles the conversion of `Issue` data into ServiceNow-compatible format.

Responsibilities
----------------

-   **ServiceNow REST API Communication**: It handles HTTP requests to the ServiceNow REST API for incident creation and management, including authentication and error handling.

-   **HTML Conversion**: It converts `Enrichment` blocks into HTML format using the `HTMLTransformer`, which preserves formatting and structure for display in the ServiceNow interface.

-   **Incident Payload Construction**: It builds the ServiceNow incident payload with fields like `short_description`, `description`, `impact`, `urgency`, `priority`, and `category` according to the ServiceNow API specification.

-   **File Attachment Management**: It handles file uploads to ServiceNow incidents, converting `FileBlock` objects into proper ServiceNow attachments.

Key Implementation Details
--------------------------

-   **HTML-Based Content**: Unlike other senders that use plain text or markdown, the ServiceNow Sender converts enrichments into HTML format, which provides rich formatting in the ServiceNow interface.

-   **ServiceNow IUP Matrix**: The sender implements ServiceNow's specific Impact-Urgency-Priority mapping system, using numerical combinations (like 1,1,1 for critical incidents) that ServiceNow interprets to determine priority levels.

-   **REST API Integration**: It uses ServiceNow's modern REST API rather than the older SOAP interface, providing better performance and easier integration.

-   **Authentication**: It uses HTTP Basic Authentication with the configured username and password to access the ServiceNow instance.

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