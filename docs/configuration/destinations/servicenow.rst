.. _servicenow-destination:

ServiceNow Destination Configuration
====================================

The ServiceNow destination allows cano-collector to create and manage incidents in ServiceNow.

Configuration
-------------

.. code-block:: yaml

    destinations:
      servicenow:
        - name: "production-incidents"
          url: "https://your-org.service-now.com"
          username: "cano-bot"
          password: "your-password"
          caller_id: "cano_bot"
          category: "Infrastructure"
          subcategory: "Kubernetes"
          priorityMapping:
            critical: "1"
            warning: "2"
            info: "3"
          customFields:
            environment: "Production"
            team: "Platform"

Parameters
----------

.. list-table::
   :header-rows: 1

   * - Parameter
     - Type
     - Required
     - Description
   * - name
     - string
     - Yes
     - Unique name for this ServiceNow destination
   * - url
     - string
     - Yes
     - ServiceNow instance URL
   * - username
     - string
     - Yes
     - ServiceNow username
   * - password
     - string
     - Yes
     - ServiceNow password
   * - caller_id
     - string
     - No
     - Used to specify a user for the "Caller" field in ServiceNow incidents. It's advisable to create a dedicated user like "cano_bot" to easily track incidents from the system.
   * - category
     - string
     - No
     - Incident category
   * - subcategory
     - string
     - No
     - Incident subcategory
   * - priorityMapping
     - map
     - No
     - Maps severity levels to ServiceNow priority levels
   * - customFields
     - map
     - No
     - Custom field values to set 