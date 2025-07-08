Slack Sender
============

The Slack Sender is responsible for formatting an `Issue` object into a rich, interactive Slack message and sending it to the configured Slack channel. It leverages the full capabilities of the Slack platform, including the **Block Kit UI framework**, **message threading**, and **enhanced enrichment formatting**.

Message Structure and Formatting
--------------------------------

The sender constructs a message payload for the `chat.postMessage` Slack API endpoint. This payload consists of a primary `text` field (for notifications) and a series of `blocks` for rich content.

Enhanced Enrichment Formatting
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

Each `BaseBlock` from an `Issue`'s enrichments is converted into a corresponding Slack Block Kit element with enhanced formatting:

-   **`HeaderBlock`**: Converted to a `header` block for prominent titles with proper styling.
-   **`MarkdownBlock`**: Converted to a `section` block with `mrkdwn` text. The sender respects Slack's Markdown flavor and enforces character limits to prevent API errors.
-   **`TableBlock`**: Enhanced intelligent handling:
    
    - **Two-column tables**: Formatted as clean, readable key-value pairs with consistent spacing
    - **Multi-column tables**: Converted to properly formatted Markdown tables
    - **Large tables**: Automatically converted to attached text files to preserve readability
    - **Color-coded attachments**: Tables are displayed with color borders based on enrichment type
    
-   **`JsonBlock`**: Enhanced JSON formatting with proper syntax highlighting and code block presentation.
-   **`ListBlock`**: Converted to well-formatted `section` blocks with proper bullet points and indentation.
-   **`LinksBlock`**: Rendered as interactive `actions` blocks containing button elements for better user experience.
-   **`FileBlock`**: Files are uploaded using the `files_upload_v2` API with permanent links included in message body.

Enhanced enrichments include:

-   **Alert Labels**: Displayed as color-coded tables with blue borders (#17A2B8)
-   **Alert Annotations**: Displayed as color-coded tables with purple borders (#6610F2)  
-   **Graphs and Charts**: Displayed with green borders (#28A745)
-   **AI Analysis**: Displayed with orange borders (#FD7E14)

Thread Management and Related Alerts
------------------------------------

The Slack Sender implements intelligent **stateless thread management** to group related alerts while maintaining application scalability.

Threading Strategy
~~~~~~~~~~~~~~~~~

The threading system uses the following approach:

1. **Firing Alerts**: Create new messages in the main channel
2. **Resolved Alerts**: Reply to the original firing alert thread when possible
3. **Related Alerts**: Alerts with the same fingerprint are grouped in threads

Stateless Thread Discovery
~~~~~~~~~~~~~~~~~~~~~~~~~

Since cano-collector is designed as a stateless, scalable application, thread relationships are discovered through:

**Primary Method - Cache-First Lookup:**

1. **In-Memory Cache**: Short-term cache (5-10 minutes TTL) stores recent fingerprint→thread_ts mappings
2. **Slack API Search**: If cache miss, search recent channel history (last 100 messages, 24-hour window) for matching fingerprints
3. **Fingerprint Embedding**: Messages include hidden metadata blocks containing issue fingerprints for reliable matching

**Thread Lifecycle:**

.. code-block:: text

    Firing Alert → New Channel Message → Cache thread_ts
         ↓
    Resolved Alert → Search Cache → Found: Reply in Thread
                            ↓
                    Cache Miss → Search Slack History → Found: Reply in Thread
                                                  ↓
                            Not Found → New Channel Message

**Performance Optimizations:**

-   **Intelligent Caching**: Frequently accessed threads are cached longer
-   **Batch API Calls**: Multiple thread searches are batched when possible  
-   **Circuit Breaker**: Falls back to new messages if Slack API is slow/unavailable
-   **Configurable Windows**: Search window and cache TTL are configurable

Thread Manager Architecture
~~~~~~~~~~~~~~~~~~~~~~~~~~

The threading functionality is implemented through a dedicated `ThreadManager` component:

.. code-block:: go

    type SlackThreadManager struct {
        slackClient   SlackClientInterface
        cache         map[string]*ThreadCacheEntry
        channelID     string
        cacheTTL      time.Duration
        searchLimit   int
        searchWindow  time.Duration
    }

**Key Methods:**

-   `GetThreadTS(fingerprint)`: Returns thread timestamp for existing alerts
-   `SetThreadTS(fingerprint, threadTS)`: Caches new thread relationships  
-   `SearchSlackHistory(fingerprint)`: Searches channel history for existing threads
-   `InvalidateThread(fingerprint)`: Removes expired or invalid thread mappings

Configuration Options
--------------------

Thread Management Settings
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

    destinations:
      slack:
        threading:
          enabled: true
          cache_ttl: "10m"              # Cache duration for thread relationships
          search_limit: 100             # Max messages to search in history
          search_window: "24h"          # Time window for history search
          fingerprint_in_metadata: true # Include fingerprint in message metadata

Enrichment Display Settings
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: yaml

    destinations:
      slack:
        enrichments:
          format_as_blocks: true        # Use Slack blocks instead of plain text
          color_coding: true            # Color-code enrichments by type
          table_formatting: "enhanced"  # "simple", "enhanced", or "attachment"
          max_table_rows: 20           # Convert large tables to files
          attachment_threshold: 1000    # Characters threshold for file conversion

Special Features
----------------

**Message Threading**: 
    Advanced stateless threading system that groups related alerts while supporting horizontal scaling.

**Message Updates**: 
    Support for updating existing messages using the `chat.update` API for summary messages and grouped alerts.

**Smart Attachments**: 
    Enrichments can be designated as attachments with color-coded borders indicating status and enrichment type.

**Adaptive Formatting**: 
    Automatic format selection based on content size and type - small tables as inline text, large tables as files.

**Link Unfurling**: 
    Configurable link preview behavior based on destination configuration.

**Fallback Handling**: 
    Graceful degradation when Slack API is unavailable - falls back to simple text formatting.

Performance Considerations
-------------------------

**API Rate Limiting**: 
    Built-in rate limiting and retry logic for Slack API calls with exponential backoff.

**Memory Management**: 
    Thread cache with TTL-based cleanup to prevent memory leaks in long-running deployments.

**Concurrent Safety**: 
    Thread-safe cache operations supporting multiple concurrent senders.

**Circuit Breaker**: 
    Automatic fallback to non-threaded messages when thread discovery fails repeatedly.

Error Handling and Reliability
------------------------------

**Graceful Degradation**: 
    If thread discovery fails, messages are sent as new channel messages rather than failing completely.

**Retry Logic**: 
    Intelligent retry for transient Slack API failures with exponential backoff.

**Monitoring Integration**: 
    Metrics and logging for thread cache hit rates, API call latency, and error rates.

**Configuration Validation**: 
    Startup validation of Slack configuration and connectivity before processing alerts. 