Senders
=======

Senders are a core component of Cano-Collector, responsible for the final step in the alert pipeline: **delivering the processed and enriched alert to its destination**.

Each sender is tailored for a specific notification service (like Slack, Microsoft Teams, or PagerDuty) and handles two key responsibilities:

1.  **Formatting**: It transforms Cano-Collector's internal alert format into the specific payload required by the destination's API. This could mean creating a JSON structure for a webhook, formatting a message with Markdown, or building a complex card-based notification.
2.  **Dispatching**: It communicates with the destination's API, handling the actual HTTP requests, authentication, and error handling to ensure the notification is sent reliably.

This modular design decouples the core alert processing logic from the specifics of each integration, making it easy to add support for new notification services in the future.

.. toctree::
   :maxdepth: 1

   datadog
   discord
   google_chat
   jira
   kafka
   mattermost
   msteams
   opsgenie
   pagerduty
   servicenow
   slack
   telegram
   victorops
   webhook 