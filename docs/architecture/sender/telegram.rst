Telegram Sender
===============

The Telegram Sender delivers notifications to a Telegram chat through a bot. It's a lightweight and fast way to receive alerts on mobile devices.

Formatting
----------

Messages are formatted using **Telegram's Bot API**, which supports a subset of Markdown or HTML for styling.

- **`HeaderBlock`**: Rendered as bold text.
- **`MarkdownBlock`** and **`ListBlock`**: Converted to Telegram-compatible Markdown.
- **`TableBlock`**: Formatted as pre-formatted, fixed-width text to preserve table structure.
- **`FileBlock`**: Sent as a file attachment using the `sendDocument` API method. This is useful for logs.
- **`LinksBlock`**: Rendered as inline URL keyboard buttons for a cleaner user experience.

Key Differences
---------------

- **Simpler than Slack**: The formatting is less complex than Slack's Block Kit, with no concept of attachments or complex layouts. It focuses on delivering clear, concise text-based messages.
- **Strong Mobile Experience**: Telegram is optimized for mobile, making it an excellent choice for on-call engineers who need to receive alerts on the go.
- **Bot-Centric**: All interactions happen through a bot, which must be created via the "BotFather" and invited to the target chat. 