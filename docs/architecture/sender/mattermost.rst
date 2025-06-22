Mattermost Sender
=================

The Mattermost Sender sends notifications to a Mattermost channel using an incoming webhook.

Formatting
----------

The sender formats messages using **GitHub-flavored Markdown**. It constructs a single text payload that is sent to the Mattermost webhook.

Block Conversion
~~~~~~~~~~~~~~~~

- **`HeaderBlock`**: Rendered as a bolded line of text.
- **`MarkdownBlock`**: The Markdown text is passed through directly.
- **`TableBlock`**: Converted into a Markdown table.
- **`ListBlock`**: Converted into a Markdown list.
- **`FileBlock`**: Sent as a file attachment to the post. This is handled by making a separate API call to upload the file and then referencing it.
- **`LinksBlock`**: Rendered as standard Markdown links.

Key Features
------------

- **Markdown-Native**: The formatting is simple and familiar to anyone who has used GitHub or other Markdown-based platforms.
- **File Support**: Like Slack, it fully supports attaching log files and other outputs, which is a key advantage over simpler webhook-based systems.
- **Simplicity**: Compared to the block-based or card-based UI of Slack and Google Chat, Mattermost messages are simpler and less structured, which can be either a pro or a con depending on the use case. 