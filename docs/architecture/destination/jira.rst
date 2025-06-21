Jira Destination
================

The Jira Destination manages the lifecycle of Jira tickets, handling creation, resolution, and reopening based on the status of `Issue` objects.

Responsibilities
----------------

-   **Ticket Lifecycle Management**: The destination determines whether to create a new ticket, resolve an existing one, or reopen a closed ticket based on the issue's status and configuration settings like `sendResolved` and `reopenIssues`.

-   **Deduplication Strategy**: It uses the `fingerprint` (or other configured deduplication fields) to identify existing tickets and prevent duplicate creation. This is crucial for maintaining clean ticket management.

-   **Project and Issue Type Configuration**: It prepares the project context and issue type information needed by the `JiraSender` to create tickets in the correct Jira project.

-   **Delegation**: After determining the appropriate action, it delegates the actual Jira API communication and ticket creation/management to the `JiraSender`.

Key Implementation Details
--------------------------

-   **Flexible Deduplication**: Unlike other destinations that use only `fingerprint`, Jira supports multiple deduplication strategies through the `dedups` parameter, allowing for more sophisticated ticket management.

-   **Status Mapping**: The destination maps issue statuses to Jira workflow transitions, supporting custom status names through configuration parameters like `doneStatusName` and `reopenStatusName`.

-   **Priority Mapping**: It supports custom mapping between Robusta severity levels and Jira priority levels, allowing for organization-specific priority schemes. 