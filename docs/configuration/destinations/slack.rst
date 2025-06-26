Slack
=====

Sends notifications to a Slack channel. This is the most feature-rich destination, offering interactive messages, file attachments, and notification grouping.

Creating a Slack App
-------------------

To use Slack integration, you need to create a Slack App and obtain an API key. Follow these steps:

1. **Create a New Slack App**
   
   Go to `https://api.slack.com/apps?new_app=1` and click "Create New App".
   
   - Choose "From scratch"
   - Enter an app name (e.g., "Kubecano Alerts")
   - Select your workspace

2. **Configure App Permissions**
   
   In your app settings, go to "OAuth & Permissions" and add the following scopes:
   
   - `chat:write` - Send messages to channels
   - `chat:write.public` - Send messages to public channels
   - `files:write` - Upload files (for log attachments)
   - `incoming-webhook` - Post messages via webhook (if using webhook method)

3. **Install App to Workspace**
   
   - Go to "Install App" in the left sidebar
   - Click "Install to Workspace"
   - Authorize the app for your workspace

4. **Get Your API Key**
   
   - Return to "OAuth & Permissions"
   - Copy the "Bot User OAuth Token" (starts with `xoxb-`)
   - This is your `api_key` for configuration

5. **Add Bot to Channel**
   
   - Go to the Slack channel where you want to receive alerts
   - Type `/invite @your-bot-name` to add the bot to the channel
   - Or add the bot manually through channel settings

Configuration
-------------

There are several ways to configure the Slack API key:

**Method 1: Direct API Key in Helm Values**

.. code-block:: yaml

    # values.yaml
    destinations:
      slack:
        - name: "my-slack-destination"
          api_key: "xoxb-YOUR-SLACK-BOT-TOKEN"
          slack_channel: "my-alerts-channel"
          unfurl_links: true    # Optional: Defaults to true.
          grouping_interval: 60 # Optional: Time in seconds to group notifications. Defaults to 0 (disabled).

**Method 2: API Key from Helm Install Flag**

.. code-block:: yaml

    # values.yaml
    destinations:
      slack:
        - name: "my-slack-destination"
          api_key: "{{ .Values.slackApiKey }}"
          slack_channel: "my-alerts-channel"

    # Install with:
    # helm install kubecano ./helm/cano-collector --set slackApiKey="xoxb-YOUR-SLACK-BOT-TOKEN"

.. code-block:: bash

    helm install cano-collector ./helm/cano-collector \
      --set destinations.slack[0].api_key="xoxb-your-slack-bot-token" \
      --set destinations.slack[0].slack_channel="#prod-alerts"

**Method 3: API Key from External Kubernetes Secret**

Create a Kubernetes Secret with your Slack API keys:

.. code-block:: bash

    kubectl create secret generic kubecano-slack-api-keys \
      --from-literal=prod-slack="xoxb-PROD-TOKEN" \
      --from-literal=dev-slack="xoxb-DEV-TOKEN" \
      --namespace=monitoring

Then reference it in your Helm values:

.. code-block:: yaml

    destinations:
      slack:
        - name: "prod-slack-destination"
          api_key_value_from:
            secretName: "kubecano-slack-api-keys"
            secretKey: "prod-slack"
          slack_channel: "#prod-alerts"
          grouping_interval: 30

**Method 4: Slack Incoming Webhook (Limited Features)**

For simple notifications without advanced features:

.. code-block:: yaml

    # values.yaml
    destinations:
      slack:
        - name: "simple-slack-destination"
          webhookURL: "https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX"
          slack_channel: "#alerts"

.. note::

   The destinations configuration is stored in a Kubernetes Secret and mounted as a YAML file inside the cano-collector pod. This ensures secure handling of sensitive configuration data.

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`api_key`** (string, required - mutually exclusive with `api_key_value_from` and `webhookURL`)
    The Slack Bot User OAuth Token, starting with `xoxb-`. This is required for advanced features like file uploads, message updates, and interactivity. You must provide either `api_key`, `api_key_value_from`, or `webhookURL` - but only one of them.

-   **`api_key_value_from`** (object, required - mutually exclusive with `api_key` and `webhookURL`)
    Reference to a Kubernetes Secret containing the Slack API key. Use this instead of `api_key` when you want to store the token in a separate secret. You must provide either `api_key`, `api_key_value_from`, or `webhookURL` - but only one of them.
    
    .. code-block:: yaml
    
        api_key_value_from:
          secretName: "kubecano-slack-api-keys"  # Name of the Kubernetes Secret
          secretKey: "prod-slack"                # Key within the secret
    
    The secret should contain the API key as a key-value pair:
    
    .. code-block:: bash
    
        kubectl create secret generic kubecano-slack-api-keys \
          --from-literal=prod-slack="xoxb-PROD-TOKEN" \
          --from-literal=dev-slack="xoxb-DEV-TOKEN" \
          --namespace=monitoring

-   **`slack_channel`** (string, required)
    The name of the Slack channel to send notifications to (e.g., `#my-channel`).

-   **`grouping_interval`** (integer, optional)
    Default: `0`. The time in seconds to wait and group multiple issues into a single summary message, with individual alerts posted in a thread. This helps to reduce channel noise. If set to `0`, each issue is sent as a separate message immediately.

-   **`unfurl_links`** (boolean, optional)
    Default: `true`. If `true`, links in the notification will be unfurled by Slack to show a preview. Set to `false` to disable this.

-   **`webhookURL`** (string, required - mutually exclusive with `api_key` and `api_key_value_from`)
    For simple, non-interactive notifications, you can use a traditional Slack Incoming Webhook URL. If you use this, functionality will be limited (e.g., no file uploads, no grouping, no message updates). You must provide either `api_key`, `api_key_value_from`, or `webhookURL` - but only one of them. It is highly recommended to use the `api_key` method for the best experience.

Security Best Practices
-----------------------

- **Never commit API keys to version control**
- **Use Kubernetes secrets** to store sensitive credentials
- **Rotate API keys regularly** for security
- **Use the minimum required permissions** for your Slack app
- **Consider using environment-specific apps** for different environments (dev, staging, prod)
- **Use separate API keys** for different environments to limit blast radius

.. note::
    Using the `api_key` method is strongly recommended to enable all features like log uploads, message grouping with threading, and future interactive components. 