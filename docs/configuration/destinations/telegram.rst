Telegram
========

Sends notifications to a Telegram chat.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      telegram:
        - name: "my-telegram-chat"
          bot_token: "your-telegram-bot-token"
          chat_id: "-1001234567890"  # Or your personal chat ID

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`bot_token`** (string, required)
    The token for your Telegram bot, obtained from the BotFather.

-   **`chat_id`** (string, required)
    The unique identifier for the target chat. For groups and channels, it's a negative number. 