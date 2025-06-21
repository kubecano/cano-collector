MS Teams Sender
===============

The MS Teams Sender formats an `Issue` object for delivery to a Microsoft Teams channel by constructing an **Adaptive Card**. This is a JSON-based UI framework used by Microsoft for creating rich, interactive content.

Formatting and Block Conversion
-------------------------------

The sender translates each `BaseBlock` from an `Issue`'s enrichments into corresponding elements within a single Adaptive Card payload.

-   **`HeaderBlock`**: Converted to a `TextBlock` with a large font size and bold weight, serving as a title.
-   **`MarkdownBlock`**: Converted to a `TextBlock` that supports the subset of Markdown compatible with Adaptive Cards.
-   **`TableBlock`**: Handled in two ways:
    -   Simple two-column tables are converted into a `FactSet`, which is the ideal format for key-value data in MS Teams.
    -   Tables with more columns are formatted as pre-formatted text within a `TextBlock`, preserving the tabular layout.
-   **`ListBlock`**: Rendered as a Markdown-formatted list within a `TextBlock`.
-   **`LinksBlock`**: Creates `Action.OpenUrl` buttons in the card's action section, providing clickable links.
-   **`FileBlock`**: This is a key limitation. The content of a `FileBlock` is not sent as a file attachment. Instead, it is embedded directly into the card as pre-formatted text inside a `TextBlock`. This is suitable for short snippets but impractical for large log files.
-   **`CallbackBlock`** (Interactivity): Not supported. Buttons for triggering actions are not available in this sender.

Key Differences from Other Senders
----------------------------------

-   **No True File Attachments**: Unlike the Slack sender, log files or other artifacts are embedded directly in the message body, which can be cumbersome for large files.
-   **No Grouping or Threading**: The sender processes and sends each issue individually. It does not support grouping multiple alerts into a summary message or replying in threads.
-   **Limited Interactivity**: Only supports buttons that open URLs. More complex, backend-driven callbacks are not available.

Comparison to Slack Sender
--------------------------

- **Formatting Richness**: While Adaptive Cards are powerful, Slack's Block Kit generally offers more fine-grained control and a wider variety of interactive components out-of-the-box.
- **File Handling**: The most significant difference is in file handling. The MS Teams Sender embeds file content directly, whereas the Slack Sender uploads it as a distinct file, which is often a better user experience for logs. 