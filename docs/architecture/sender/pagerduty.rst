PagerDuty Sender
================

The `PagerDutySender` is responsible for communicating with the PagerDuty Events API. It receives instructions and a prepared data context from the `PagerDutyDestination` and handles the final payload construction and API interaction.

Responsibilities
----------------

-   **Endpoint Selection**: Based on the command from the destination, it sends requests to either the `Alert Events API` (`/v2/enqueue`) or the `Change Events API` (`/v2/change/enqueue`).

-   **Payload Construction**: It builds the specific JSON payload required by the selected PagerDuty API endpoint.
    -   **For Alerts**:
        -   It maps the `Issue` severity to PagerDuty's severity levels (`critical`, `error`, `warning`, `info`).
        -   It converts all `Enrichment` blocks into plain text and combines them into a single `state_message` field within `custom_details`. This means structured data like tables or diffs are presented as simple text.
        -   It extracts links from `LinksBlock` and places them in the top-level `links` array of the PagerDuty payload.
        -   Key `Issue` attributes like `source`, `component`, and `fingerprint` are mapped to their corresponding fields (`source`, `component`, `dedup_key`) in the PagerDuty event.
    -   **For Changes**:
        -   It constructs a simpler payload focused on a summary of the change, including which resource was affected and in what way.

-   **API Communication**: It handles the final HTTP POST request to PagerDuty, including setting the correct headers and the `routing_key`.

Key Implementation Details
--------------------------

-   **Flattening Enrichments**: Unlike Senders for chat platforms (Slack, MSTeams), the PagerDuty Sender does not preserve the rich formatting of blocks like tables. All enrichments are converted to a single text block, optimized for readability within the PagerDuty UI's "Custom Details" section.

-   **Deduplication**: Correct use of the `dedup_key` is fundamental to preventing alert storms and ensuring a clean incident timeline. 