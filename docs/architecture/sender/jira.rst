Jira Sender
===========

The `JiraSender` communicates with the Jira REST API to create, update, and manage tickets. It receives instructions from the `JiraDestination` and handles the conversion of `Issue` data into Jira-compatible format.

Responsibilities
----------------

-   **Jira REST API Communication**: It handles all HTTP requests to the Jira REST API, including authentication, ticket creation, status transitions, and updates.

-   **Rich Text Conversion**: It converts `Enrichment` blocks into Jira's Atlassian Document Format (ADF), preserving formatting like markdown, tables, and lists while maintaining readability in the Jira interface.

-   **File Attachment Handling**: It manages file uploads to Jira tickets, converting `FileBlock` objects into proper Jira attachments.

-   **Ticket Management**: It handles the complete lifecycle of Jira tickets, including creation, status transitions, and updates based on the commands from the destination.

Key Implementation Details
--------------------------

-   **Atlassian Document Format**: Unlike other senders that use simple text or HTML, the Jira Sender converts enrichments into Jira's native ADF format, which supports rich formatting, tables, and structured content.

-   **Markdown Processing**: It includes sophisticated markdown parsing to convert common markdown elements (bold, italic, code blocks) into their ADF equivalents.

-   **Status Transition Management**: The sender handles complex Jira workflow transitions, mapping issue statuses to appropriate Jira status changes through the REST API.

-   **Authentication**: It uses HTTP Basic Authentication with the configured username and API key to access the Jira instance.

Formatting
----------

The sender maps `Issue` data to the fields of a **Jira Issue** using the Jira Cloud Platform API.

- **`issue.Title`**: Becomes the `summary` of the Jira issue.
- **`issue.Description`**: Used as the main `description`.
- **`issue.Severity`**: Can be used to set the `priority` of the Jira issue.
- **`issue.Subject` and labels**: Mapped to Jira `labels` for categorization.
- **`Enrichments`**: All enrichment blocks are serialized into a text format and appended to the `description` or added as a `comment` on the newly created issue.

Key Functionality
-----------------

- **Automated Ticket Creation**: Eliminates the manual work of creating bug reports or tasks for operations teams. When a critical issue is detected, a Jira ticket can be created and assigned automatically.
- **Workflow Integration**: Bridges the gap between automated monitoring and manual project management workflows in Jira.
- **Traceability**: Provides a clear link between a cluster event and the work item tracking its resolution. 