.. _servicenow-destination:

ServiceNow
==========

This destination creates incidents in ServiceNow based on issues in your Kubernetes cluster.

Configuration
-------------

.. code-block:: yaml

    - name: servicenow_destination_name
      type: servicenow
      params:
        # ServiceNow instance identifier
        instance: "your-instance"
        # ServiceNow username
        username: "admin"
        # ServiceNow password
        password: "SecurePassword@123"
        # Optional: Caller ID for incidents (default: empty)
        caller_id: "robusta_bot"

Parameter Reference
-------------------

``instance``
  *(Required)* Your ServiceNow instance identifier (e.g., "mycompany" for mycompany.service-now.com).

``username``
  *(Required)* The ServiceNow username for authentication.

``password``
  *(Required)* The ServiceNow password for authentication.

``caller_id``
  *(Optional)* Used to specify a user for the "Caller" field in ServiceNow incidents. It's advisable to create a dedicated user like "robusta_bot" to easily track incidents from the system. 