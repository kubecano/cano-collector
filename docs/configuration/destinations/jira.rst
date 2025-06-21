.. _jira-destination:

Jira
====

This destination creates and manages Jira tickets based on issues in your Kubernetes cluster.

Configuration
-------------

.. code-block:: yaml

    - name: jira_destination_name
      type: jira
      params:
        # Jira instance URL (HTTPS required)
        url: "https://workspace.atlassian.net"
        # Email used to log into Jira
        username: "jira-user@company.com"
        # API key from Atlassian account
        api_key: "your-atlassian-api-key"
        # Project where tickets will be created
        project_name: "KUBERNETES"
        # Optional: Jira issue type (default: Task)
        issue_type: "Task"
        # Optional: Custom priority mapping
        priority_mapping:
          HIGH: "High"
          LOW: "Low"
          INFO: "Lowest"
        # Optional: Deduplication fields (default: fingerprint)
        dedups: ["fingerprint"]
        # Optional: Send resolved events (default: false)
        sendResolved: true
        # Optional: Reopen closed tickets (default: false)
        reopenIssues: true

Parameter Reference
-------------------

``url``
  *(Required)* The URL of your Jira workspace. HTTPS is required.

``username``
  *(Required)* The email address used to log into your Jira account.

``api_key``
  *(Required)* Your Atlassian API key. Follow the `Atlassian documentation <https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/>`_ to create one.

``project_name``
  *(Required)* The name of the Jira project where tickets will be created.

``issue_type``
  *(Optional)* The type of Jira issue to create. Defaults to "Task".

``priority_mapping``
  *(Optional)* Maps Robusta severity levels to Jira priority levels.

``dedups``
  *(Optional)* Fields to use for ticket deduplication. Defaults to `["fingerprint"]`. Can include: `fingerprint`, `cluster_name`, `title`, `node`, `type`, `source`, `namespace`, `creation_date`.

``sendResolved``
  *(Optional)* Whether to resolve Jira tickets when alerts are resolved. Defaults to `false`.

``reopenIssues``
  *(Optional)* Whether to reopen resolved tickets when alerts fire again. Defaults to `false`. 