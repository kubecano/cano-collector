Jira
====

Creates issues in a Jira project.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      jira:
        - name: "my-jira-project"
          url: "https://my-company.atlassian.net"
          username: "automation-user@my-company.com"
          api_token: "your-jira-api-token"
          project_key: "PROJ"
          issue_type: "Task"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`url`** (string, required)
    The URL of your Jira Cloud instance.

-   **`username`** (string, required)
    The email address of the user to create issues as.

-   **`api_token`** (string, required)
    A Jira API token associated with the user.

-   **`project_key`** (string, required)
    The key of the Jira project where issues will be created (e.g., "PROJ").

-   **`issue_type`** (string, required)
    The type of issue to create (e.g., "Bug", "Task", "Story"). This must be a valid issue type in your project. 