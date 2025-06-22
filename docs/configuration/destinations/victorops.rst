.. _victorops-destination:

VictorOps
=========

This destination sends notifications to VictorOps (Splunk On-Call) using the REST API.

Configuration
-------------

.. code-block:: yaml

    - name: victorops_destination_name
      type: victorops
      params:
        # VictorOps REST endpoint URL.
        # This URL determines which team and routing key will receive the alert.
        # It is a required parameter.
        url: "https://alert.victorops.com/integrations/generic/20131114/alert/4a6a87eb-fca9-4117-931a-c842277ea90a/$routing_key"

Parameter Reference
-------------------

``url``
  *(Required)* The VictorOps REST endpoint URL. This URL includes the integration ID and routing key that determine which team receives the alert and how it is processed. 