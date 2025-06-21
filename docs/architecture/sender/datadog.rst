DataDog Sender
==============

The `DataDogSender` communicates with the DataDog Events API to create events that appear in the DataDog dashboard. It receives data from the `DataDogDestination` and handles the final payload construction and API communication.

Responsibilities
----------------

-   **DataDog Events API Communication**: It sends HTTP requests to the DataDog Events API to create events that appear in the DataDog dashboard timeline.

-   **Text Conversion**: It converts all `Enrichment` blocks into a single text field using DataDog's specific formatting conventions, including special markers like `%%%` for sections and `presto` table format.

-   **Event Payload Construction**: It builds the DataDog event payload with fields like `title`, `text`, `aggregation_key`, `alert_type`, and `tags` according to the DataDog Events API specification.

-   **Content Truncation**: It handles DataDog's content length limits by truncating long text fields while preserving the most important information.

Key Implementation Details
--------------------------

-   **DataDog-Specific Formatting**: The sender uses DataDog's unique text formatting conventions, including `%%%` markers for section headers and the `presto` table format for structured data.

-   **Length Limits**: DataDog has strict limits on event content (97 characters for title, 3997 for text), so the sender includes intelligent truncation to ensure events are created successfully.

-   **Tag-Based Organization**: It adds cluster information as tags to help organize and filter events in the DataDog dashboard.

-   **Event Timeline**: Unlike other destinations that create persistent incidents, DataDog events appear in a timeline view, making them ideal for monitoring and alerting rather than incident management.

-   **Severity Mapping**: It maps Robusta severity levels to DataDog event types (`error` for HIGH severity, `info` for others), which affects how events are displayed and filtered in DataDog.

Payload Structure
-----------------

The sender transforms an `Issue` into a **DataDog Event** JSON payload with the following key fields:

-   **`title`**: The issue title (truncated to 97 characters)
-   **`text`**: Combined description and enrichments (truncated to 3997 characters)
-   **`aggregation_key`**: Used for event grouping and deduplication
-   **`alert_type`**: Mapped from issue severity (`error` or `info`)
-   **`tags`**: Cluster and resource information for filtering
-   **`source_type_name`**: Set to "Robusta" for identification
-   **`host`**: Resource information in format `namespace/resource_type/name`