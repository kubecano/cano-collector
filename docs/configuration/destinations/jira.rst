.. _jira-destination:

Jira Destination Configuration
=============================

The Jira destination allows cano-collector to create and manage Jira issues for alerts and findings.

Configuration
-------------

.. code-block:: yaml

    destinations:
      jira:
        - name: "production-issues"
          url: "https://your-org.atlassian.net"
          username: "jira-bot@your-org.com"
          apiToken: "your-api-token"
          projectKey: "OPS"
          issueType: "Incident"
          priorityMapping:
            critical: "Highest"
            warning: "High"
            info: "Medium"
          customFields:
            environment: "Production"
            team: "Platform"
          labels:
            - "automated"
            - "kubernetes"

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
     - Unique name for this Jira destination
   * - url
     - string
     - Yes
     - Jira instance URL
   * - username
     - string
     - Yes
     - Jira username or email
   * - apiToken
     - string
     - Yes
     - Jira API token
   * - projectKey
     - string
     - Yes
     - Jira project key
   * - issueType
     - string
     - Yes
     - Type of issue to create
   * - priorityMapping
     - map
     - No
     - Maps severity levels to Jira priority levels
   * - customFields
     - map
     - No
     - Custom field values to set
   * - labels
     - list
     - No
     - Labels to add to issues 