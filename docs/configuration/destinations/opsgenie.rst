OpsGenie
========

Creates, updates, and resolves alerts in Atlassian OpsGenie.

Configuration
-------------

.. code-block:: yaml

    # values.yaml
    destinations:
      opsgenie:
        - name: "my-opsgenie-destination"
          apiKey: "your-opsgenie-api-key"
          region: "eu"  # Optional: "eu" or "us". Defaults to global.
          teams:  # Optional: List of teams to be assigned as responders.
            - "my-static-team-name"
            - "team-{{- labels.team_name_label }}" # Template example
          default_team: "fallback-team" # Optional: Used if template resolution fails.
          tags: # Optional: Extra static tags to add to all alerts.
            - "k8s"
            - "production"
          extra_details_labels: # Optional: List of issue labels to add to alert details.
            - "pod_name"
            - "namespace"

Parameters
----------

-   **`name`** (string, required)
    A unique name for this destination instance.

-   **`apiKey`** (string, required)
    The API key from your OpsGenie integration (type: API). It must have permissions to create and close alerts.

-   **`region`** (string, optional)
    The OpsGenie region to send API requests to. Can be `eu` or `us`. If omitted, the default global API endpoint (`api.opsgenie.com`) is used.

-   **`teams`** (list of strings, optional)
    A list of team names that will be assigned as responders to the alert. Team names can be static or contain Go templates that reference the issue's labels or annotations (e.g., `{{- labels.team }}`).

-   **`default_team`** (string, optional)
    A fallback team name to use if a template in the `teams` list cannot be resolved from the issue's metadata. This is highly recommended when using templated team names.

-   **`tags`** (list of strings, optional)
    A list of static tags to be added to every alert sent to this destination. The cluster name is always added as a tag by default.

-   **`extra_details_labels`** (list of strings, optional)
    A list of label keys from the `issue.subject.labels`. If a listed label exists on an issue, its key-value pair will be added to the `details` section of the OpsGenie alert. 