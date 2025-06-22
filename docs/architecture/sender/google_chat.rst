Google Chat Sender
==================

The Google Chat Sender sends notifications to a Google Chat space using an incoming webhook.

Formatting
----------

The sender formats messages using **Google Chat Cards (v2)**. A card is a UI element that can contain headers, text sections, images, and interactive widgets like buttons.

Block Conversion
~~~~~~~~~~~~~~~~

- **`HeaderBlock`**: Becomes the `title` in the card's `header`.
- **`MarkdownBlock`**: Converted to a `textParagraph` widget within the card. Google Chat supports a limited set of Markdown-like formatting.
- **`TableBlock`**: Simple two-column tables are rendered as a list of `decoratedText` widgets with top labels. More complex tables are formatted as pre-formatted text in a `textParagraph` widget.
- **`ListBlock`**: Converted to a `textParagraph` with a bulleted list.
- **`FileBlock`**: Google Chat webhooks do not support file attachments. The content of `FileBlock` is ignored.
- **`LinksBlock`**: Rendered as a `buttonList` widget with `onClick` actions that open the provided URLs.

Key Differences
---------------

- **No File Attachments**: Unlike Slack or Discord, this sender cannot attach files like logs.
- **Limited Markdown**: The formatting capabilities are less extensive than those of Slack or Mattermost. The card-based structure is more rigid. 