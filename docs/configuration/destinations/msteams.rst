MS Teams
========

Sends notifications to a Microsoft Teams channel via an incoming webhook.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      msteams:
        - name: "my-msteams-destination"
          webhookURL: "https://your-org.webhook.office.com/webhookb2/..."
          webhook_override: ""  # Optional: See note below.

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`webhookURL`** (string, required)
    The default incoming webhook URL provided by the MS Teams channel connector.

-   **`webhook_override`** (string, optional)
    The name of a Kubernetes annotation to check for on a resource. If the annotation exists, its value will be used as the webhook URL, overriding the default `webhookURL`. This allows for dynamic routing. For example, if you set `webhook_override: "msteams.webhook/my-team"`, the system will look for that annotation on the alert's subject.

.. note::
    For dynamic routing, you can add an annotation to your Kubernetes resources. For instance, if `webhook_override` is set to `msteams.webhook/my-team`, adding the following annotation to a Deployment will send its alerts to the specified URL:
    
    .. code-block:: yaml

        apiVersion: apps/v1
        kind: Deployment
        metadata:
          name: my-app
          annotations:
            msteams.webhook/my-team: "https://.../another-webhook-url" 