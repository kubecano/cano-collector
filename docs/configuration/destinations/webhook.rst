Webhook
=======

Sends the raw `Issue` object as a JSON payload to a generic webhook endpoint. This is useful for integrating with custom tools or services that are not officially supported.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      webhook:
        - name: "my-custom-webhook"
          url: "https://my-service.com/api/alerts"
          headers:                # Optional
            Authorization: "Bearer my-secret-token"
            X-Custom-Header: "value"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`url`** (string, required)
    The webhook endpoint URL to which the POST request will be sent.

-   **`headers`** (dictionary, optional)
    A key-value map of custom HTTP headers to include in the request (e.g., for authentication). 