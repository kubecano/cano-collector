OpsGenie Destination
====================

The OpsGenie Destination acts as a strategic bridge between the collector's internal `Issue` model and the OpsGenie alert API. Its primary responsibilities are managing the alert lifecycle and dynamically routing alerts to the correct teams.

Responsibilities
----------------

-   **Lifecycle Translation**: The destination interprets the `status` of an `Issue` (`FIRING` or `RESOLVED`) and instructs the `OpsGenieSender` to perform the correct action ("create" or "close").

-   **Dynamic Team Routing**: This is a key feature of the OpsGenie destination. It uses Go templates in the `teams` configuration parameter to resolve team names based on an issue's labels and annotations. For example, a team name like `team-{{- labels.team_name }}` allows for decentralized and flexible alert routing without changing central configuration. It also manages a `default_team` for cases where templates cannot be resolved.

-   **Data Preparation**: It gathers and prepares all necessary data for the sender, including the `apiKey`, the `fingerprint` for deduplication, and any extra labels or tags specified in the configuration.

-   **Delegation**: After preparing the context and determining the action, it delegates the final API payload construction and communication to the `OpsGenieSender`. 