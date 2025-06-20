ServiceNow
==========

Creates incidents in a ServiceNow instance using the Table API.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      servicenow:
        - name: "my-servicenow-instance"
          url: "https://my-instance.service-now.com"
          username: "api-user"
          password: "api-user-password"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`url`** (string, required)
    The full URL of your ServiceNow instance.

-   **`username`** (string, required)
    The username for authenticating with the ServiceNow API.

-   **`password`** (string, required)
    The password for the API user. 