Slack Sender
============

The Slack Sender is responsible for formatting an `Issue` object into a rich, interactive Slack message and sending it to the configured Slack channel. It leverages the full capabilities of the Slack platform, including the **Block Kit UI framework** and file uploads.

Message Structure and Formatting
--------------------------------

The sender constructs a message payload for the `chat.postMessage` Slack API endpoint. This payload consists of a primary `text` field (for notifications) and a series of `blocks` for rich content.

Block Conversion Details
~~~~~~~~~~~~~~~~~~~~~~~~

Each `BaseBlock` from an `Issue`'s enrichments is converted into a corresponding Slack Block Kit element:

-   **`HeaderBlock`**: Converted to a `header` block for prominent titles.
-   **`MarkdownBlock`**: Converted to a `section` block with `mrkdwn` text. The sender respects Slack's Markdown flavor and enforces character limits to prevent API errors.
-   **`TableBlock`**: Handled intelligently. Simple two-column tables are formatted into a clean, readable list of key-value pairs. Tables with more columns are converted to a Markdown table inside a `section` block. Very wide tables are automatically converted to an attached text file to preserve readability.
-   **`ListBlock`**: Converted to a `section` block with a Markdown-formatted list.
-   **`LinksBlock`**: Rendered as an `actions` block containing interactive `button` elements, providing a better user experience than plain text links.
-   **`FileBlock`**: Handled specially. The file content is uploaded to Slack using the `files_upload_v2` API. A permanent link to the uploaded file is then included in the message body. This is ideal for attaching logs, graphs, or other diagnostic files.

Special Features
----------------

-   **Message Threading**: The sender supports replying within a thread. If a `thread_ts` (timestamp of the parent message) is provided, the message will be posted as a reply, which is essential for grouping related notifications, like updates to an ongoing alert.
-   **Message Updates**: The sender can update existing messages using the `chat.update` API. This is used, for example, to update a summary message when a grouped alert's status changes.
-   **Attachments**: Enrichments can be designated as "attachments" using a specific annotation. These blocks are placed in a separate `attachments` field in the API payload. This allows for a distinct visual presentation, often with a color-coded border (e.g., red for firing alerts, green for resolved) that indicates the status of the issue.
-   **Link Unfurling**: The sender can control whether Slack should show previews for links in the message, based on the `unfurl_links` parameter from the destination configuration. 