Discord Sender
==============

The `DiscordSender` communicates with Discord's webhook API to create rich, embed-based messages. It receives data from the `DiscordDestination` and handles the conversion of `Issue` data into Discord's embed format.

Responsibilities
----------------

-   **Discord Webhook Communication**: It sends HTTP POST requests to Discord webhooks to create embed messages with rich formatting, colors, and fields.

-   **Embed Construction**: It builds Discord embeds with proper structure including title, description, fields, and color coding based on issue severity.

-   **Block Conversion**: It converts `Enrichment` blocks into Discord-compatible format:
    -   `MarkdownBlock` becomes embed description or fields
    -   `TableBlock` is converted to code blocks or file attachments for wide tables
    -   `FileBlock` is sent as separate file attachments
    -   `HeaderBlock` becomes embed title
    -   `ListBlock` is converted to markdown lists

-   **File Attachment Management**: It handles file uploads by sending separate webhook requests with file attachments alongside the main embed message.

Key Implementation Details
--------------------------

-   **Embed-Based Architecture**: Discord uses embeds for rich message formatting, with support for fields, colors, and structured content.

-   **Color Coding**: The sender maps issue severity to Discord embed colors (red for HIGH, yellow for LOW, green for INFO) for immediate visual identification.

-   **Table Handling**: Wide tables that exceed Discord's field limits are automatically converted to file attachments to maintain readability.

-   **Length Limits**: The sender respects Discord's character limits (2048 for descriptions, 1024 for fields) and truncates content accordingly.

-   **Dual Webhook Requests**: For messages with file attachments, the sender makes two webhook requests: one for the embed and one for the files.

Formatting
----------

The sender constructs **Discord Embeds** to present notifications in a structured and visually appealing way. An embed can have a title, description, color-coded border, and fields.

Block Conversion
~~~~~~~~~~~~~~~~

- **`HeaderBlock`**: Becomes the `title` of the Discord embed.
- **`MarkdownBlock`**: The content is used as the `description` of the embed or as a `field` if a title is present in the markdown. Discord supports standard Markdown.
- **`TableBlock`**: Formatted as a code block within a field to maintain alignment.
- **`ListBlock`**: Converted to a Markdown list within the description or a field.
- **`FileBlock`**: Sent as a separate file attachment alongside the embed message.
- **`LinksBlock`**: Rendered as Markdown links in the embed's description.

Special Features
----------------

- **Color Coding**: The left border of the embed is color-coded based on the `Issue`'s severity (e.g., red for high, yellow for low), providing an immediate visual cue.
- **Fields**: Enrichments are often broken down into individual `fields` within the embed, creating a clean key-value layout, for example, for displaying labels or annotations. 