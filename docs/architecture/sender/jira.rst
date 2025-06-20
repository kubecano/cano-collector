Jira Sender
===========

The Jira Sender integrates with Atlassian Jira to automatically create issues in a specified project, streamlining the process of tracking and managing work generated from cluster events.

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