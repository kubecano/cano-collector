DataDog
=======

Sends events to DataDog.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      datadog:
        - name: "my-datadog-destination"
          apiKey: "your-datadog-api-key"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`apiKey`** (string, required)
    Your DataDog API key. 