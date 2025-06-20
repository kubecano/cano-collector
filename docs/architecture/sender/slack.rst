Slack Sender
============

The Slack Sender is responsible for formatting an `Issue` object into a rich, interactive Slack message and sending it to the configured Slack webhook.

Formatting
----------

The sender leverages **Slack Block Kit**, Slack's UI framework for building messages. This allows for detailed and well-structured notifications that are easy to read and act upon.

Block Conversion
~~~~~~~~~~~~~~~~

Each `BaseBlock` from an `Issue`'s enrichments is converted into a corresponding Slack Block Kit element:

- **`HeaderBlock`**: Converted to a `header` block.
- **`MarkdownBlock`**: Converted to a `section` block with `mrkdwn` text. This supports Slack's flavor of Markdown for formatting like bold, italics, and lists.
- **`TableBlock`**: Rendered as a formatted `section` block using a fixed-width font to simulate a table. For very wide tables, it may instead be converted to a `FileBlock` to preserve readability.
- **`ListBlock`**: Converted to a `section` block with a Markdown list.
- **`FileBlock`**: This is handled specially. The file content is uploaded to Slack using the `files.upload` API, and a link to the uploaded file is included in the message. This is ideal for attaching logs or other diagnostic files.
- **`LinksBlock`**: Converted to an `actions` block containing interactive `button` elements.

Special Features
----------------

- **Attachments**: Enrichments can be designated as "attachments" using annotations. These are displayed in a visually distinct, color-coded section of the Slack message, which is useful for separating primary information from secondary details like labels or annotations.
- **Interactivity**: The sender supports interactive components like buttons, which can be used for features like silencing alerts or triggering follow-up actions. 