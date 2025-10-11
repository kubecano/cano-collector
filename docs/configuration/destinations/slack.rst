Slack
=====

Sends notifications to a Slack channel. This is the most feature-rich destination, offering interactive messages, file attachments, and notification grouping.

Creating a Slack App
--------------------

To use Slack integration, you need to create a Slack App and obtain an API key. Follow these steps:

1. **Create a New Slack App**
   
   Go to `https://api.slack.com/apps?new_app=1` and click "Create New App".
   
   Choose one of the following methods:

2. **Choose App Creation Method**
   
   a. **Using Manifest (Recommended)**
   
      - Choose "From manifest"
      - Select your workspace
      - In the next step, use the following manifest:

      .. code-block:: json

          {
              "display_information": {
                  "name": "Kubecano Alerts",
                  "description": "Kubernetes alerting and monitoring integration",
                  "background_color": "#2eb67d"
              },
              "features": {
                  "bot_user": {
                      "display_name": "Kubecano Bot",
                      "always_online": false
                  }
              },
              "oauth_config": {
                  "scopes": {
                      "bot": [
                          "chat:write",
                          "chat:write.public",
                          "files:write"
                      ]
                  }
              },
              "settings": {
                  "org_deploy_enabled": false,
                  "socket_mode_enabled": false,
                  "is_hosted": false,
                  "token_rotation_enabled": false
              }
          }

   b. **Using From Scratch**
   
      - Choose "From scratch"
      - Enter an app name (e.g., "Kubecano Alerts")
      - Select your workspace
      - Go to "OAuth & Permissions" and add the following scopes:
        - `chat:write` - Send messages to channels
        - `chat:write.public` - Send messages to public channels
        - `files:write` - Upload files (for log attachments)

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

.. important::

   The external Kubernetes Secret **must be in the same namespace** where you install the Helm chart. 
   If the secret is in a different namespace, Helm will fail to resolve the API key during template rendering.

.. note::

   The destinations configuration is stored in a Kubernetes Secret and mounted as a YAML file inside the cano-collector pod. This ensures secure handling of sensitive configuration data.

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`api_key`** (string, required - mutually exclusive with `api_key_value_from`)
    The Slack Bot User OAuth Token, starting with `xoxb-`. This is required for all Slack features including file uploads, message updates, and interactivity. You must provide either `api_key` or `api_key_value_from`.

-   **`api_key_value_from`** (object, required - mutually exclusive with `api_key`)
    Reference to a Kubernetes Secret containing the Slack API key. Use this instead of `api_key` when you want to store the token in a separate secret. You must provide either `api_key` or `api_key_value_from`.
    
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

File Upload Configuration
~~~~~~~~~~~~~~~~~~~~~~~~~

Configure file upload behavior for enrichments:

.. code-block:: yaml

   destinations:
     - name: "default-slack"
       type: "slack"
       channel: "alerts"
       api_key: "${SLACK_BOT_TOKEN}"

       enrichments:
         max_table_rows: 20      # Tables larger than this → CSV files

**Required Bot Permissions**:

The Slack app must have these OAuth scopes:

- ``chat:write`` - Send messages to channels (required)
- ``files:write`` - Upload files to Slack workspace (required for logs/CSV files)
- ``files:read`` - Read file info for permalinks (required for file links)

.. note::
   Files are uploaded to workspace storage (not channel-specific). This avoids requiring
   the ``conversations:list`` permission. Permalinks work across all channels.

**File Upload Behavior**:

- **Pod Logs**: Uploaded as ``.log`` files with timestamps in filename
- **Large Tables**: Tables with >``max_table_rows`` rows are converted to CSV files
- **Deduplication**: Identical enrichments are automatically removed (always enabled)
- **Error Fallback**: Upload failures result in text display with error explanation
- **Empty Logs**: Graceful handling with helpful diagnostic message

**Default Values**:

- ``max_table_rows``: 20 rows

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