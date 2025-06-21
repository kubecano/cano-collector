ServiceNow Destination
======================

The ServiceNow Destination manages the creation and lifecycle of incidents in ServiceNow, handling the conversion of `Issue` objects into ServiceNow incident records.

Responsibilities
----------------

-   **Incident Lifecycle Management**: The destination determines whether to create a new incident or update an existing one based on the issue's status and ServiceNow configuration.

-   **ServiceNow Context Preparation**: It prepares the ServiceNow instance URL, authentication credentials, and caller information needed by the `ServiceNowSender`.

-   **Priority Mapping**: It maps Robusta severity levels to ServiceNow's Impact, Urgency, and Priority (IUP) matrix using ServiceNow's specific numerical combinations.

-   **Delegation**: After preparing the ServiceNow context, it delegates the actual incident creation and management to the `ServiceNowSender`.

Key Implementation Details
--------------------------

-   **ServiceNow IUP Matrix**: The destination uses ServiceNow's specific Impact-Urgency-Priority mapping system, which uses numerical combinations to determine incident priority levels.

-   **Caller Management**: It supports configuring a dedicated caller ID (like "robusta_bot") to easily track and manage incidents created by the system.

-   **Incident Categorization**: It automatically categorizes incidents as "Network" type, which can be customized based on organizational needs. 