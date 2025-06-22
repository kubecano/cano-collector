Jira Destination
===============

The Jira destination allows cano-collector to create and update Jira issues for alerts and findings.

Features
--------

-   **Issue Creation**: Automatically creates Jira issues for new alerts
-   **Issue Updates**: Updates existing issues when alerts are resolved
-   **Custom Fields**: Supports mapping to custom Jira fields
-   **Priority Mapping**: It supports custom mapping between severity levels and Jira priority levels, allowing for organization-specific priority schemes.
-   **Label Management**: Automatically adds and removes labels based on alert status
-   **Comment Updates**: Adds comments with enrichment information
-   **Attachment Support**: Attaches files and logs to issues

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