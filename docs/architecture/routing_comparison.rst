Routing Architecture Comparison
===============================

This document provides a detailed comparison between cano-collector's routing architecture and Robusta's approach, highlighting the key differences in how alerts are processed, routed, and delivered to notification endpoints.

Core Architectural Differences
-----------------------------

**Robusta's Approach:**
- Uses **Sinks** as the primary concept for notification endpoints
- Implements a **SinkRegistry** for dynamic sink discovery and management
- Routing is configured at the sink level using **scopes** and **matchers**
- Each sink has its own routing rules (include/exclude patterns)

**Cano-collector's Approach:**
- Uses **Destinations** for notification endpoints (equivalent to Robusta's sinks)
- Uses **Teams** for routing logic and destination grouping
- Routing is configured at the team level, not at individual destinations
- No equivalent to SinkRegistry - routing is static and configuration-driven

Detailed Flow Comparison
-----------------------

1. **Alert Reception and Parsing**

   **Robusta:**
   - Alerts received via `/api/alerts` endpoint
   - Parsed into `PrometheusAlert` objects
   - Deduplication using compound hash of fingerprint, status, and timestamps
   - Relabeling applied based on configuration
   - Added to async task queue for processing

   **Cano-collector:**
   - Alerts received via `/api/alerts` endpoint
   - Parsed into `template.Data` (Alertmanager format)
   - Converted to internal `Issue` model
   - Currently synchronous processing (TODO: async queue implementation)

2. **Alert Processing and Enrichment**

   **Robusta:**
   - `PlaybooksEventHandler` processes alerts
   - Creates `Finding` objects with enriched context
   - Applies playbook actions for additional enrichment
   - Supports custom enrichment through playbooks

   **Cano-collector:**
   - `AlertHandler` processes alerts
   - Creates `Issue` objects (equivalent to Robusta's Finding)
   - Enrichment planned through `Enrichment` blocks
   - Currently basic processing (TODO: enrichment implementation)

3. **Sink/Destination Registration**

   **Robusta:**
   - Sinks registered dynamically through `SinkRegistry`
   - Each sink has its own configuration and routing rules
   - Sinks can be added/removed at runtime
   - Supports sink discovery and health checking

   **Cano-collector:**
   - Destinations configured statically in Helm values
   - Loaded at startup through `DestinationsLoader`
   - No runtime registration - requires configuration change
   - Destinations grouped by type (slack, teams, etc.)

4. **Routing Logic**

   **Robusta:**
   - Routing configured per sink using `scope` blocks
   - Supports include/exclude patterns based on:
     - `identifier` (alert name)
     - `namespace`
     - `severity`
     - `type`
     - `kind`
     - `source`
     - `labels` (Kubernetes selectors)
   - Complex matching with multiple conditions
   - Fallback sinks for unmatched alerts

   **Cano-collector:**
   - Routing configured at team level
   - Teams map to destination names
   - Routing rules planned but not yet implemented
   - Simpler, more centralized approach
   - No equivalent to Robusta's scope-based routing

5. **Destination/Sink Matching**

   **Robusta:**
   - Each sink evaluated independently
   - Alert matched against sink's scope rules
   - Multiple sinks can receive the same alert
   - Sink-specific configuration (channels, webhooks, etc.)

   **Cano-collector:**
   - Alert matched to team based on routing rules (planned)
   - Team maps to multiple destinations
   - All destinations in a team receive the alert
   - Destination configuration separate from routing logic

6. **Message Formatting and Sending**

   **Robusta:**
   - Each sink type has its own formatting logic
   - Sinks handle API communication directly
   - Supports rich formatting (embeds, attachments, etc.)
   - Built-in support for multiple notification types

   **Cano-collector:**
   - Uses Strategy pattern with `Sender` implementations
   - `Destination` delegates to `Sender` for formatting
   - Clean separation between routing and formatting
   - Factory pattern for sender creation

7. **Configuration Management**

   **Robusta:**
   - Sink configuration in `sinksConfig` section
   - Routing rules embedded in sink configuration
   - Dynamic configuration updates
   - Complex configuration with many options

   **Cano-collector:**
   - Destinations configured in `destinations` section
   - Teams configured in `teams` section
   - Static configuration loaded at startup
   - Simpler, more structured approach

Key Advantages of Each Approach
------------------------------

**Robusta's Advantages:**
- More flexible routing with complex matching rules
- Dynamic sink registration and management
- Rich ecosystem of built-in sinks
- Mature implementation with extensive features
- Better support for complex notification scenarios

**Cano-collector's Advantages:**
- Simpler, more maintainable architecture
- Clear separation of concerns (routing vs. formatting)
- Easier to understand and extend
- More predictable behavior
- Better suited for simpler use cases

Future Considerations for Cano-collector
----------------------------------------

To bridge the gap with Robusta's capabilities, cano-collector could consider:

1. **Enhanced Routing Engine:**
   - Implement scope-based routing similar to Robusta
   - Add support for complex matching rules
   - Include fallback routing mechanisms

2. **Dynamic Configuration:**
   - Add support for runtime configuration updates
   - Implement destination health checking
   - Support for dynamic destination registration

3. **Advanced Features:**
   - Implement async processing queue
   - Add support for alert grouping and deduplication
   - Include rich enrichment capabilities

4. **Monitoring and Observability:**
   - Add metrics for routing decisions
   - Implement tracing for alert flow
   - Include health checks for destinations

This comparison highlights that while cano-collector takes a simpler, more focused approach, Robusta provides a more feature-rich but complex solution. The choice between them depends on the specific requirements for alert routing complexity and operational needs. 