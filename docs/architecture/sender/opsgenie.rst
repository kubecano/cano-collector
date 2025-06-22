OpsGenie Sender
===============

The OpsGenie Sender is responsible for communicating with the OpsGenie Alert API to manage the incident lifecycle. Unlike UI-focused senders, its goal is not to format a rich message but to create and resolve alerts in a structured way.

Formatting and Field Mapping
----------------------------

The sender translates the `Issue` object into a payload for the OpsGenie `CreateAlertPayload`.

-   **`issue.Title`**: Mapped directly to the `message` field of the OpsGenie alert.
-   **`issue.Fingerprint`**: This is the most critical field, mapped to the `alias`. It is the key used by OpsGenie for alert deduplication and to correlate a `resolve` action with the correct open alert.
-   **`issue.Severity`**: Mapped to OpsGenie's `priority` levels (e.g., `HIGH` becomes `P1`, `LOW` becomes `P4`).
-   **`issue.Subject` and `Enrichments`**:
    -   The `description` of the OpsGenie alert is constructed by converting all enrichment blocks into a single **HTML** string. This provides context for the on-call engineer directly within the incident.
    -   Key details from the `issue.subject` (like resource name, namespace, node) are placed in the `details` field of the alert.
    -   Other labels from the issue can be added to the `details` as well, based on the destination configuration.
-   **`Links`**: Links for investigating the issue are converted to HTML `<a>` tags and prepended to the alert's description.

Key Functionality
-----------------

-   **Lifecycle Management**: The sender's logic differentiates between opening and closing an alert. It sends a "create" request for new `FIRING` issues and a "close" request (identified by the `alias`) for `RESOLVED` issues. It also supports acknowledging alerts.
-   **Deduplication**: The correct and consistent use of the `alias` field is fundamental to how this sender works, preventing alert storms and ensuring a clean incident timeline.
-   **Team Routing**: The `responders` field in the API payload is populated with team names, which are determined at the destination level, allowing for dynamic routing of alerts. 