.. _pagerduty-destination:

PagerDuty
=========

This destination sends notifications to PagerDuty, creating and resolving incidents.

Configuration
-------------

.. code-block:: yaml

    - name: pagerduty_destination_name
      type: pagerduty
      params:
        # PagerDuty Integration Key (also known as the routing key).
        # This key determines which service in PagerDuty will receive the event.
        # It is a required parameter.
        apiKey: "f6c6e02a5a1a490ee02e90cde19ee388"

Parameter Reference
-------------------

``apiKey``
  *(Required)* The integration key (or routing key) for your PagerDuty service. This directs the alert to the correct team and escalation policy within PagerDuty. 