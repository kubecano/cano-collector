PagerDuty Destination
=====================

The PagerDuty Destination's role is to interpret the nature of an `Issue` and command the `PagerDutySender` with the correct context. It distinguishes between configuration changes and alerts, managing the incident lifecycle accordingly.

Responsibilities
----------------

-   **Event Type Classification**: The destination inspects the `Issue` to determine if it represents a **Configuration Change** or an **Alert**. This is a critical distinction, as each type maps to a different PagerDuty API endpoint (`Change Events` vs. `Alert Events`). It uses the `aggregation_key` for this classification.

-   **Lifecycle Management**: For alerts, the destination determines the `event_action` (`trigger` or `resolve`) based on the issue's `status` or title prefix. This is essential for opening and closing incidents in PagerDuty.

-   **Data Preparation**: It extracts the `fingerprint` to be used as the `dedup_key`, ensuring alerts are correctly grouped. It also prepares the `apiKey` (routing key) for the sender.

-   **Delegation**: After classifying the event and preparing the necessary metadata, it delegates the construction of the final API payload and the HTTP call to the `PagerDutySender`. 