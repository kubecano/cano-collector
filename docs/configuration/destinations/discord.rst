.. _discord-destination:

Discord
=======

This destination sends notifications to Discord channels using webhooks with rich embed formatting.

Configuration
-------------

.. code-block:: yaml

    - name: discord_destination_name
      type: discord
      params:
        # Discord webhook URL
        # This URL determines which channel will receive the messages
        # It is a required parameter
        url: "https://discord.com/api/webhooks/YOUR_WEBHOOK_ID/YOUR_WEBHOOK_TOKEN"

Parameter Reference
-------------------

``url``
  *(Required)* The Discord webhook URL. This URL includes the webhook ID and token that determine which channel receives the messages.

Getting a Discord Webhook URL
----------------------------

1. Open the Discord channel where you want to receive notifications
2. Go to **Channel Settings** → **Integrations** → **Webhooks**
3. Click **New Webhook**
4. Configure the webhook (name, avatar, etc.)
5. Copy the **Webhook URL** from the webhook settings 