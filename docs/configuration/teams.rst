Configuring Teams
===================

Teams are a core concept for routing issues to the correct destinations. A team is a named entity that groups together one or more destinations. This allows you to manage notification endpoints centrally.

Defining a Team
---------------

A team is defined by its `name` and a list of `destinations` it should send issues to. The names in the `destinations` list must correspond to the `name` of a configured destination.

Configuration is provided in the `teams` section of the Helm `values.yaml` file.

**Example `values.yaml`:**

.. code-block:: yaml

    # values.yaml

    # First, define the destinations that will be used.
    destinations:
      slack:
        - name: "alerts-prod-channel"
          webhookURL: "https://hooks.slack.com/services/..."
        - name: "alerts-staging-channel"
          webhookURL: "https://hooks.slack.com/services/..."
      pagerduty:
        - name: "on-call-critical"
          integrationKey: "your-pagerduty-integration-key"

    # Next, define the teams and map them to the destinations.
    teams:
      - name: "backend-devs"
        destinations:
          - "alerts-staging-channel"

      - name: "on-call-team"
        destinations:
          - "alerts-prod-channel"
          - "on-call-critical"

How it Works
------------

1.  **Team Definition**: You define a list of teams. Each team has a unique name.
2.  **Destination Mapping**: For each team, you specify a list of destination names. These names must match the `name` field of a destination defined in the `destinations` configuration block.
3.  **Routing (Future)**: The collector will use routing rules (to be documented separately) to match an incoming issue to a specific team. Once a team is matched, the issue is sent to all destinations associated with that team.

This structure decouples routing logic from endpoint configuration, making it easy to change where a team's alerts are sent without modifying the routing rules themselves. 