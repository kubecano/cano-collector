Google Chat
===========

Sends notifications to a Google Chat space via an incoming webhook.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      google_chat:
        - name: "my-google-chat-space"
          url: "https://chat.googleapis.com/v1/spaces/..."

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`url`** (string, required)
    The webhook URL provided by Google Chat for your space. 