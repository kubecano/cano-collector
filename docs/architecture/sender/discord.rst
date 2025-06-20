Discord Sender
==============

The Discord Sender formats `Issue` objects for delivery to a Discord channel via a webhook.

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