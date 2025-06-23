ServiceNow Destination
======================

The ServiceNow destination allows cano-collector to create and manage incidents in ServiceNow.

Features
--------

-   **Incident Creation**: Automatically creates ServiceNow incidents for alerts
-   **Incident Updates**: Updates existing incidents when alerts are resolved
-   **Priority Mapping**: It maps severity levels to ServiceNow's Impact, Urgency, and Priority (IUP) matrix using ServiceNow's specific numerical combinations.
-   **Assignment Rules**: Supports automatic assignment based on configuration
-   **Category Mapping**: Maps alert types to ServiceNow categories
-   **Caller Management**: It supports configuring a dedicated caller ID (like "cano_bot") to easily track and manage incidents created by the system.
-   **Custom Fields**: Supports mapping to custom ServiceNow fields

Configuration
-------------

.. code-block:: yaml

    destinations:
      servicenow:
        - name: "production-incidents"
          url: "https://your-org.service-now.com"
          username: "cano-bot"
          password: "your-password"
          caller_id: "cano_bot"
          category: "Infrastructure"
          subcategory: "Kubernetes"
          priorityMapping:
            critical: "1"
            warning: "2"
            info: "3" 