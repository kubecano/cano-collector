DataDog Sender
==============

The DataDog Sender forwards `Issue` objects to the DataDog Events API, allowing you to correlate infrastructure issues with your application metrics and logs within the DataDog platform.

Formatting
----------

The sender transforms an `Issue` into a **DataDog Event** JSON payload.

- **`issue.Title`**: Mapped to the `title` of the DataDog event.
- **`issue.Description`** and **`Enrichments`**: The description and all enrichment blocks are serialized into a single Markdown-formatted string and placed in the `text` body of the event.
- **`issue.Severity`**: Converted to the DataDog event `alert_type` (e.g., `error`, `warning`, `info`).
- **`issue.Source`**: Used as the `source_type_name`.
- **`issue.Subject` and labels**: Mapped to `tags` in the event, enabling filtering and correlation within DataDog.

Key Functionality
-----------------

- **Event Correlation**: By sending issues as events tagged with relevant metadata (like pod name, namespace, node), you can see them directly on your DataDog dashboards and metric graphs.
- **Centralized Visibility**: It allows you to see operational events from your cluster alongside your performance monitoring data, providing a unified view for troubleshooting. 