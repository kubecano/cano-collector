Data Model
==========

The core of the collector's reporting pipeline is its data model. All events, alerts, and notifications are standardized into an `Issue` object, which is then enriched and passed to various destinations.

The Issue Object
----------------

The `Issue` is the canonical data structure for any reportable event. It is defined in ``pkg/core/issue/issue.go``.

.. code-block:: go

    type Issue struct {
        ID             uuid.UUID
        Title          string
        Description    string
        AggregationKey string
        Severity       Severity
        Status         Status
        Source         Source
        Subject        Subject
        Enrichments    []Enrichment
        Links          []Link
        Fingerprint    string
        StartsAt       time.Time
        EndsAt         *time.Time
    }

- **Title**: A brief, human-readable title for the issue.
- **Description**: A more detailed description of the issue.
- **AggregationKey**: A string used to group similar issues together.
- **Severity**: The priority of the issue (e.g., `HIGH`, `LOW`, `INFO`).
- **Status**: The current state of the issue (`FIRING` or `RESOLVED`).
- **Source**: The origin of the issue (e.g., `PROMETHEUS`, `KUBERNETES_API_SERVER`).
- **Subject**: The Kubernetes resource the issue pertains to.
- **Enrichments**: A list of contextual data blocks that add more information.
- **Links**: A list of relevant URLs.
- **Fingerprint**: A unique hash to identify an instance of an issue.

Enrichments and Blocks
----------------------

An `Enrichment` provides additional context to an `Issue`. It is a container for one or more `BaseBlock` objects. This structure allows for rich, detailed reports.

.. code-block:: go

    type Enrichment struct {
        Blocks []BaseBlock
        Annotations map[string]string
    }

The `Blocks` are the visual components of the report. The primary block types are defined in ``pkg/core/issue/blocks.go``:

- **MarkdownBlock**: A block of text formatted with Markdown.
- **TableBlock**: A block for displaying tabular data with headers and rows.
- **FileBlock**: A block for attaching files, such as log files or diagnostic outputs.
- **ListBlock**: A simple list of string items.
- **HeaderBlock**: A simple text header.
- **DividerBlock**: A visual separator.
- **LinksBlock**: A block dedicated to displaying a set of hyperlinks.

This block-based system allows senders to intelligently format reports for their target platform. For example, a `SlackSender` can translate these blocks into Slack's Block Kit UI components, while a `PagerdutySender` might serialize them into a JSON payload in the `custom_details` field. 