MS Teams Sender
===============

The MS Teams Sender formats `Issue` objects for delivery to a Microsoft Teams channel via an incoming webhook.

Formatting
----------

The sender uses the **Adaptive Cards** format, which is Microsoft's open framework for building rich, interactive UI content that can be used across different apps and services.

Block Conversion
~~~~~~~~~~~~~~~~

The `BaseBlock` elements from an `Issue` are translated into elements within an Adaptive Card JSON payload:

- **`HeaderBlock`**: Becomes a `TextBlock` with a larger font size and weight.
- **`MarkdownBlock`**: Converted to a `TextBlock` that supports a subset of Markdown.
- **`TableBlock`**: Rendered as a `FactSet` for simple key-value pairs or as a formatted `TextBlock` with a monospace font for more complex tables.
- **`ListBlock`**: Translated into a `TextBlock` with a Markdown list.
- **`FileBlock`**: MS Teams webhooks do not directly support file attachments in the same way as Slack. The content of a `FileBlock` is typically embedded directly into the card as a pre-formatted text block. This is suitable for short logs but less ideal for large files.
- **`LinksBlock`**: Rendered as a set of `Action.OpenUrl` buttons within the card's `actions` section.

Comparison to Slack Sender
--------------------------

- **Formatting Richness**: While Adaptive Cards are powerful, Slack's Block Kit generally offers more fine-grained control and a wider variety of interactive components out-of-the-box.
- **File Handling**: The most significant difference is in file handling. The MS Teams Sender embeds file content directly, whereas the Slack Sender uploads it as a distinct file, which is often a better user experience for logs. 