Incident Management Destinations
================================

Destinations for incident management platforms like **PagerDuty**, **OpsGenie**, and **VictorOps** share a common set of responsibilities focused on managing the alert lifecycle.

Shared Responsibilities
-----------------------

-   **Lifecycle Translation**: These destinations are responsible for interpreting the `status` of an `Issue` (`FIRING` or `RESOLVED`) and ensuring the sender is called with the correct corresponding action (`trigger`/`critical` or `resolve`/`recovery`). This is a core piece of logic that lives at the destination level.

-   **Deduplication Strategy**: They ensure that the `aggregation_key` from the `Issue` is correctly passed to the sender so it can be used as the `dedup_key`, `alias`, or `entity_id`. This guarantees that alerts are correctly grouped in the external platform, preventing alert storms.

-   **Configuration Holding**: Each destination holds the necessary authentication details for its platform, such as an `integrationKey` (PagerDuty), `apiKey` (OpsGenie), or `routing_key` (VictorOps).

-   **Delegation**: After translating the intent (e.g., "resolve this issue"), the destination delegates the task of constructing the specific API payload to its corresponding sender (`PagerdutySender`, `OpsGenieSender`, etc.).

While the senders handle the API-specific field names and formats, the destinations handle the "business logic" of incident management: translating status into actions and ensuring proper deduplication. 