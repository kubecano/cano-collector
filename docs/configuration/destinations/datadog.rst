.. _datadog-destination:

DataDog
=======

This destination sends events to the DataDog Events API for correlation with metrics and logs.

Configuration
-------------

.. code-block:: yaml

    - name: datadog_destination_name
      type: datadog
      params:
        # DataDog API key from your DataDog account
        api_key: "your-datadog-api-key"

Parameter Reference
-------------------

``api_key``
  *(Required)* Your DataDog API key. This can be retrieved from your DataDog Account Settings. 