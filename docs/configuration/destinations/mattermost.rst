Mattermost
==========

Sends notifications to a Mattermost channel via an incoming webhook.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      mattermost:
        - name: "my-mattermost-channel"
          url: "https://my-mattermost-instance.com/hooks/..."

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`url`** (string, required)
    The incoming webhook URL from your Mattermost instance. 