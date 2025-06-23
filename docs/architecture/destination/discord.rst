Discord Destination
===================

The Discord Destination prepares data for Discord webhook communication, focusing on creating rich, embed-based messages for Discord channels.

Responsibilities
----------------

-   **Webhook Configuration**: The destination prepares the Discord webhook URL and any additional configuration needed by the `DiscordSender` to communicate with Discord's webhook API.

-   **Data Preparation**: It extracts basic information from the `Issue` and prepares the context needed for Discord's embed-based message format.

-   **Delegation**: After preparing the webhook context, it delegates the construction of Discord embeds and the webhook communication to the `DiscordSender`.

Key Implementation Details
--------------------------

-   **Webhook-Based**: Unlike other destinations that use dedicated APIs, Discord uses webhooks for simple, stateless communication.

-   **Embed-Focused**: Discord is designed around rich embeds rather than simple text messages, making it ideal for structured data presentation.

-   **File Support**: The destination supports file attachments, which are sent as separate webhook requests alongside the main embed message. 