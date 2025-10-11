# Robusta Slack Message Structure Analysis

Analysis of Robusta's Slack notification format patterns for comparison with cano-collector implementation.

---

## Message Type Schemas

### Schema 1: K8s Event with Inline Logs

**Use Case**: Kubernetes events (CrashLoopBackOff, etc.) with embedded log content

**Header Pattern**:
- Event Type Icon: 👀 + "K8s event detected"
- Severity Badge: 🔴/🟡/🟢 + severity text (High/Medium/Low)
- Subject Line: Clickable link describing the event
  - Format: "{Event description} {resource} in namespace {namespace}"

**Action Buttons**:
- None in cano-collector (Robusta buttons link to SaaS features)

**Content Structure**:
1. **Source Information**: Cluster name in orange badge
2. **Crash Info Section** (optional, for crash events):
   - Container name (badge)
   - Restart count (badge)
   - Status (badge)
   - Reason (badge)
3. **Previous Container Section** (optional, for crashes):
   - Status (badge)
   - Reason (badge)
   - Started at timestamp (badge)
   - Finished at timestamp (badge)
4. **Inline Log Viewer** (collapsible):
   - Dropdown header: "{resource-name}.log ▼"
   - Line numbers (left column)
   - Monospace font
   - Syntax highlighting for log levels
   - Full log content embedded in message
   - NO external file links

**Key Characteristics**:
- Logs embedded directly in Slack message (not as files)
- Collapsible UI with dropdown arrow
- Line numbers for reference
- Syntax highlighting appropriate to content
- Clean visual hierarchy

**Visual Elements**:
- Orange/brown badges for all technical values
- Monospace font for technical content
- Collapsible sections to save space
- No red accent borders (not Prometheus alerts)

---

### Schema 2: Prometheus Alert Firing (Basic)

**Use Case**: Prometheus alerts without enrichments (node alerts, API alerts, etc.)

**Header Pattern**:
- Event Type: 🔥 + "Prometheus Alert Firing" + 🔥
- Severity Badge: 🔴/🟡/🟢 + severity text
- Subject Line: Clickable link from alert annotations/summary

**Action Buttons**:
- None in cano-collector (Robusta buttons link to SaaS features)

**Content Structure**:
1. **Source Information**: Cluster name in orange badge
2. **Alert Description Section**:
   - Contextual icon (⏰ for time, 💽 for storage/disk, 🚨 for generic alerts)
   - Alert text from annotations (plain text or multi-paragraph)
   - Optional: Inline code badges for technical terms
3. **Alert Labels Section** (collapsible):
   - Section header: "Alert labels" with RED left border/accent line
   - Bullet-point list with badges
   - Collapsed by default (shows first 4 labels)
   - "See more" / "See less" toggle link
   - All labels displayed as orange badges

**Key Characteristics**:
- No logs or file attachments
- Metric-based alerts (not event-based)
- Collapsible metadata
- Contextual icons for alert types
- RED left accent border on labels section

**Visual Elements**:
- 🔥 flames emoji for firing state
- RED accent border on labels section
- Orange badges for all label values
- Clean separation between description and metadata

---

### Schema 3: Prometheus Alert Firing with File Attachments

**Use Case**: Pod-related Prometheus alerts (KubePodCrashLooping, etc.) with log enrichments

**Header Pattern**:
- Same as Schema 2 (Prometheus Alert Firing)

**Action Buttons**:
- None in cano-collector (Robusta buttons link to SaaS features)

**Content Structure**:
1. **Source Information**: Cluster name in orange badge
2. **Alert Description Section**:
   - Icon (💽 or 🚨)
   - Alert text with pod/container details
3. **File Attachments Section** (NEW):
   - Section header: "{N} files ▼ | 📥 Download all"
   - File cards with dark background:
     - File icon (📄 "T" for text files)
     - Filename (truncated if long)
     - File type ("Plain Text")
   - Collapsible file list
   - Individual download + bulk download option
4. **Alert Labels Section** (collapsed):
   - Same as Schema 2

**Key Characteristics**:
- Logs as downloadable file cards (NOT inline)
- Multiple files possible (logs, events, etc.)
- Card-based UI for files
- Files embedded in Slack message (not external links)
- Collapsible file list with count indicator

**Visual Elements**:
- File cards with dark background
- File type icons
- Download controls visible
- Clean spacing between alert and files

---

### Schema 4: Prometheus Alert Firing with Inline Content

**Use Case**: Alerts with YAML/JSON enrichments (API definitions, config resources, etc.)

**Header Pattern**:
- Same as Schema 2 (Prometheus Alert Firing)

**Action Buttons**:
- None in cano-collector (Robusta buttons link to SaaS features)

**Content Structure**:
1. **Source Information**: Cluster name in orange badge
2. **Alert Description Section**:
   - Icon (💽, 🚨, or contextual)
   - Multi-paragraph alert text
   - Inline code badges for technical terms
   - Explanatory text introducing the content
3. **Inline Content Viewer** (collapsible):
   - Dropdown header: "{resource-name}.yaml ▼"
   - Line numbers (left column)
   - Syntax highlighting for YAML/JSON
   - Color-coded keys and values
   - Monospace font
   - Properly formatted and indented
4. **Alert Labels Section** (collapsed):
   - Same as Schema 2

**Key Characteristics**:
- YAML/JSON embedded inline (like logs in Schema 1)
- Syntax highlighting appropriate to content type
- Line numbers for reference
- Contextual explanation before content
- Same collapsible pattern as log viewer

**Visual Elements**:
- Same as Schema 1 but for YAML/JSON
- Color-coded syntax highlighting
- Line numbers
- Collapsible dropdown

---

### Schema 5: Prometheus Alert Firing with Extended Enrichments

**Use Case**: Deployment/ReplicaSet alerts with pod status and crash information

**Header Pattern**:
- Same as Schema 2 (Prometheus Alert Firing)

**Action Buttons**:
- None in cano-collector (Robusta buttons link to SaaS features)

**Content Structure**:
1. **Source Information**: Cluster name in orange badge
2. **Alert Description Section**:
   - Icon (🚨 or contextual)
   - Alert text with deployment/resource details
3. **Pod Status Summary** (optional):
   - Plain text: "{N} pod(s) are available. {M} pod(s) are not ready due to {reason}"
4. **Technical Details Block** (optional):
   - Single-line technical description with IDs
   - Format: "{Reason}: {details} container={name} pod={name}_{namespace}({uuid})"
5. **Crash Info Section** (optional, for pod crashes):
   - Bullet-point list with badges:
     - Container name
     - Restart count
     - Status
     - Reason
6. **Previous Container Section** (optional, for crashes):
   - Same format as Crash Info
7. **Resource Status Summary** (optional):
   - Plain text: "Replicas: Desired ({N}) --> Running ({M})"
8. **Alert Labels Section** (collapsed):
   - Same as Schema 2

**Key Characteristics**:
- Highest information density
- Multiple structured sections
- Pod availability metrics
- Technical IDs in plain text (UUIDs)
- Deployment/replica-specific metrics
- No file attachments or inline content

**Visual Elements**:
- Same badge styling throughout
- Hierarchical information flow
- Multiple status summaries
- Plain text for summaries, badges for technical values

---

### Schema 6: Prometheus Alert Resolved

**Use Case**: Resolution notification for previously firing alerts

**Header Pattern**:
- Event Type: ✅ + "Prometheus resolved" (no flames)
- Severity Badge: Same as original firing alert
- Subject Line: **Same as firing alert** (for threading)

**Action Buttons**:
- None in cano-collector (Robusta buttons link to SaaS features)

**Content Structure**:
1. **Source Information**: Cluster name in orange badge
2. **Alert Description Section**:
   - Icon (same as firing alert)
   - Alert text (same as firing alert)
3. **Alert Labels Section** (collapsed):
   - Section header: "Alert labels" with **GREEN left border/accent line**
   - Same format as firing alert

**Key Characteristics**:
- **GREEN border** instead of RED
- ✅ checkmark instead of 🔥 flames
- **No enrichments** (no files, logs, YAML)
- Same structure as firing alert for consistency
- Subject line matches for threading

**Visual Elements**:
- ✅ green checkmark
- GREEN accent border (vs RED for firing)
- "Prometheus resolved" text
- Same alert content for context

---

## Common Design Patterns Across All Schemas

### Badge Styling
- **Orange/brown color** for all technical values
- Monospace font inside badges
- Consistent size and spacing
- Used for: container names, labels, status values, timestamps, numbers

### Collapsible Sections
- **Dropdown arrow** (▼) indicator
- **Toggle text**: "See more" / "See less"
- **Close button** (×) for expanded sections
- Applied to: logs, YAML, files, labels

### Action Buttons
- **Cano-collector**: No action buttons (Robusta buttons are SaaS-specific)
- **Future consideration**: Could add custom buttons for internal tools/dashboards

### Severity Indicators
- **Icons**: 🔴 High (red), 🟡 Medium/Low (yellow), 🟢 Low (green)
- **Text**: "High", "Medium", "Low" next to icon
- Consistent placement in header

### Source Identification
- **Label**: "Source:"
- **Value**: Cluster name in orange badge
- Placed immediately after action buttons

### State-Based Visual Cues
- **Firing**: 🔥 flames, RED accent border
- **Resolved**: ✅ checkmark, GREEN accent border
- Border color on labels section indicates state

### Content Presentation
- **Inline embedding**: Logs, YAML (Schema 1, 4)
- **File cards**: Downloadable attachments (Schema 3)
- **Structured sections**: Crash Info, Previous Container (Schema 1, 5)
- **Plain text summaries**: Pod availability, replica status (Schema 5)

### Typography
- **Monospace font**: Technical values, logs, code, YAML
- **Plain text**: Alert descriptions, summaries
- **Inline code badges**: Technical terms in descriptions
- **Line numbers**: Left column for logs/YAML

---

## Message Type Selection Logic

### When to Use Each Schema

**Schema 1 (K8s Event + Inline Logs)**:
- Source: Kubernetes API events
- Condition: Event has associated pod logs
- Example alerts: BackOff, CrashLoopBackOff events
- Enrichment: Pod logs embedded inline

**Schema 2 (Basic Prometheus Alert)**:
- Source: Prometheus AlertManager
- Condition: No pod-related enrichments
- Example alerts: NodeClockNotSynchronising, NodeRAIDDegraded
- Enrichment: None (just alert description + labels)

**Schema 3 (Prometheus + File Attachments)**:
- Source: Prometheus AlertManager
- Condition: Pod-related alert with logs
- Example alerts: KubePodCrashLooping, KubePodOOMKilled
- Enrichment: Pod logs as downloadable files (1+ files)

**Schema 4 (Prometheus + Inline Content)**:
- Source: Prometheus AlertManager
- Condition: Alert with YAML/JSON resource context
- Example alerts: KubeAggregatedAPIDown, ConfigMapChanged
- Enrichment: Kubernetes resource definition inline

**Schema 5 (Prometheus + Extended Enrichments)**:
- Source: Prometheus AlertManager
- Condition: Deployment/ReplicaSet alerts with pod details
- Example alerts: KubeDeploymentReplicasMismatch, KubeStatefulSetReplicasMismatch
- Enrichment: Pod status, crash info, replica metrics (no files)

**Schema 6 (Resolved)**:
- Source: Prometheus AlertManager
- Condition: Alert status = "resolved"
- Example: Any resolved alert
- Enrichment: None (stripped from firing alert)

---

## Content Enrichment Decision Tree

```
Alert Received
│
├─ Is it a K8s Event?
│  └─ YES → Use Schema 1 (inline logs with line numbers)
│
└─ Is it a Prometheus Alert?
   │
   ├─ Is status "resolved"?
   │  └─ YES → Use Schema 6 (green border, no enrichments)
   │
   └─ Is status "firing"?
      │
      ├─ Is it pod-related? (has pod label)
      │  │
      │  ├─ Is it a deployment alert?
      │  │  └─ YES → Use Schema 5 (extended enrichments: pod status, crash info, replicas)
      │  │
      │  └─ Does it have logs?
      │     └─ YES → Use Schema 3 (file attachments as cards)
      │
      ├─ Does it have YAML/JSON context?
      │  └─ YES → Use Schema 4 (inline YAML/JSON viewer)
      │
      └─ Basic alert
         └─ Use Schema 2 (description + labels only)
```

---

## Robusta Code Analysis Summary

### Architecture Overview

**Key Components**:
1. **SlackSink** (`slack_sink.py`): Main sink handling Finding dispatch
2. **SlackSender** (`integrations/slack/sender.py`): Converts Findings to Slack blocks
3. **Template System** (`templates/`): Jinja2 templates for message headers
4. **Block Converters**: Convert enrichment blocks to Slack blocks

### Alert-to-Message Mapping Logic

**Header Creation** (`__create_finding_header_preview` lines 408-493):
```python
# Determines alert type based on source
if finding.source == FindingSource.PROMETHEUS:
    alert_type = "Alert"
elif finding.source == FindingSource.KUBERNETES_API_SERVER:
    alert_type = "K8s Event"
else:
    alert_type = "Notification"

# Status indicators
status_text = "Firing" if status == FindingStatus.FIRING else "Resolved"
status_emoji = "⚠️" if status == FindingStatus.FIRING else "✅"
```

**Template Context** (lines 465-487):
- `title`: Alert title
- `description`: Alert description
- `status_text`/`status_emoji`: Firing/Resolved indicators
- `severity`: Alert severity (High/Medium/Low)
- `alert_type`: Prometheus/K8s Event/Notification
- `cluster_name`: Source cluster
- `resource_text`: Kind/Namespace/Name format
- `labels`: Alert labels
- `annotations`: Alert annotations
- `fingerprint`: Deduplication key

**Template Rendering** (`header.j2`):
- Uses Jinja2 templates for flexible message formatting
- Renders to Slack Block Kit JSON structures
- Creates section blocks + context blocks
- Includes emojis for visual indicators

### Enrichment Processing

**Block Conversion Pattern**:
1. **Finding** contains list of **Enrichments**
2. Each **Enrichment** contains list of **Blocks** (MarkdownBlock, FileBlock, TableBlock, etc.)
3. **SlackSender** converts blocks to Slack format via `__to_slack()` method

**Key Block Types**:
- `MarkdownBlock` → Slack section block with mrkdwn
- `FileBlock` → Uploaded as Slack file
- `TableBlock` → Converted to markdown list format
- `ListBlock` → Bullet-point markdown
- `DividerBlock` → Slack divider block

**File Upload Logic** (implicit from code):
- Uses `slack_client.files_upload_v2()`
- Files get workspace storage permalinks
- Referenced in message with `<permalink|filename>` syntax

### Critical Implementation Details

**No Inline Log Viewer in Code**:
- Robusta code doesn't show inline log viewer implementation
- All logs appear to be uploaded as files via `FileBlock`
- The screenshots show inline viewers, but code uses file uploads
- **This suggests Robusta may use custom Slack app features or different rendering**

**Badge Styling**:
- Achieved via markdown code blocks: `` `value` ``
- Orange/brown color is Slack's default for inline code
- No custom styling applied - uses Slack defaults

**Collapsible Sections**:
- NOT implemented in core Robusta code
- Slack doesn't natively support collapsible blocks
- "See more/See less" likely handled by Robusta SaaS platform
- Uses message updates and interactive components

**Thread Management**:
- `thread_ts` parameter for threading
- Fingerprint-based deduplication for grouping
- Summary messages as thread parents

---

## Cano-Collector Message Format Specification

### Design Principles

1. **Clean and Readable**: Inspired by Robusta but simpler (no SaaS features)
2. **Badge Styling**: Use Slack's inline code (`` `value` ``) for technical values
3. **File-Based Enrichments**: Upload logs as files (no inline viewers initially)
4. **Threading Support**: Use fingerprints for proper threading
5. **State Indicators**: Visual cues for firing vs resolved

### Message Schema for Cano-Collector

**Schema A: Prometheus Alert (Firing)**

```
Header Section:
- Type: 🔥 Alert Firing 🔥  🔴 High
- Title: <alert title> (clickable if link available)

Context Bar:
- 📊 Type: Prometheus Alert
- Cluster: `cluster-name`
- Namespace: `namespace` (if available)
- Pod: `pod-name` (if pod-related)

Alert Description:
- Text from annotations.summary or annotations.description
- Format as markdown

Enrichments (if present):
- Crash Info (for pod alerts):
  • Container: `container-name`
  • Restarts: `count`
  • Status: `WAITING`
  • Reason: `CrashLoopBackOff`

- Files (logs, events):
  [Uploaded as Slack files with permalinks]

Alert Labels (expandable in future):
- Show key labels as badges:
  • alertname: `KubePodCrashLooping`
  • severity: `critical`
  • pod: `pod-name`
```

**Schema B: Prometheus Alert (Resolved)**

```
Header Section:
- Type: ✅ Alert Resolved  🟡 Medium
- Title: <same as firing alert>

Context Bar:
- Same as firing

Alert Description:
- Same as firing (for context)

NO Enrichments:
- Resolved alerts don't re-attach files
```

### Slack Block Kit Structure

**Header Block** (section):
```json
{
  "type": "section",
  "text": {
    "type": "mrkdwn",
    "text": "🔥 *Alert Firing* 🔥  🔴 *High*\n<url|*Alert Title*>"
  }
}
```

**Context Block** (context):
```json
{
  "type": "context",
  "elements": [
    {"type": "mrkdwn", "text": "📊 Type: Prometheus Alert"},
    {"type": "mrkdwn", "text": ":globe_with_meridians: Cluster: `cluster-name`"},
    {"type": "mrkdwn", "text": ":package: Pod: `pod-name`"}
  ]
}
```

**Description Block** (section):
```json
{
  "type": "section",
  "text": {
    "type": "mrkdwn",
    "text": "Pod is crash looping due to application error"
  }
}
```

**Enrichment Block - Crash Info** (section):
```json
{
  "type": "section",
  "text": {
    "type": "mrkdwn",
    "text": "*Crash Info*\n• Container: `busybox`\n• Restarts: `5`\n• Status: `WAITING`\n• Reason: `CrashLoopBackOff`"
  }
}
```

**File Reference Block** (section with permalink):
```json
{
  "type": "section",
  "text": {
    "type": "mrkdwn",
    "text": "📄 Pod Logs: <file_permalink|busybox-logs.txt>"
  }
}
```

### Enrichment Decision Matrix

| Alert Type | Has Pod Label | Status | Enrichments |
|-----------|--------------|---------|-------------|
| KubePodCrashLooping | Yes | Firing | Crash Info + Pod Logs File |
| KubePodOOMKilled | Yes | Firing | Crash Info + Pod Logs File |
| KubeDeploymentReplicasMismatch | Yes | Firing | Pod Status + Crash Info |
| NodeClockNotSynchronising | No | Firing | None (basic alert) |
| Any Alert | - | Resolved | None (no enrichments) |

### Badge Styling Guide

**Use Inline Code** (`` `value` ``) for:
- Container names: `busybox`
- Pod names: `my-app-7d8f9b-xyz`
- Namespace names: `default`
- Status values: `WAITING`, `TERMINATED`
- Reason values: `CrashLoopBackOff`, `Error`
- Numeric values: `5`, `10`
- Timestamps: `2025-01-10T12:00:00Z`
- Label values: `critical`, `high`

**Use Plain Text** for:
- Alert descriptions
- Explanatory text
- Section headers (with bold: `*Header*`)

**Use Emojis** for:
- Status indicators: 🔥 (firing), ✅ (resolved)
- Severity: 🔴 (high), 🟡 (medium), 🟢 (low)
- Resource types: 📦 (pod), 🖥️ (node), 🔗 (service)
- Alert type: 📊 (prometheus), 👀 (k8s event)

### Threading Strategy

**Fingerprint Generation**:
```go
fingerprint := fmt.Sprintf("%s-%s-%s-%s",
    issue.Title,
    issue.Source,
    issue.Subject.Name,
    issue.Subject.Namespace,
)
hash := sha256.Sum256([]byte(fingerprint))
return hex.EncodeToString(hash[:8])
```

**Thread Lookup**:
1. Generate fingerprint for new issue
2. Search recent messages (last 24h) for matching fingerprint
3. If found: Post as reply to that thread
4. If not found: Create new parent message
5. Store thread_ts with fingerprint in cache (10min TTL)

**Resolved Alert Threading**:
- Use SAME fingerprint as firing alert
- Will automatically thread under firing message
- Include ✅ emoji to indicate resolution

### 3. Cano-Collector Slack Sender Analysis

#### Current Implementation Review

**File**: `pkg/sender/slack/sender_slack.go`

**Current Structure**:
```go
type SenderSlack struct {
    apiKey      string
    channel     string
    channelID   string
    logger      logger_interfaces.LoggerInterface
    unfurlLinks bool
    slackClient sender_interfaces.SlackClientInterface
    threadManager sender_interfaces.SlackThreadManagerInterface
    tableFormat  string
    maxTableRows int
    // Prometheus metrics
    fileUploadsTotal          *prometheus.CounterVec
    fileUploadSizeBytes       *prometheus.HistogramVec
    tableConversionsTotal     *prometheus.CounterVec
    channelResolutionDuration prometheus.Histogram
}
```

**Current Features** ✅:
- Threading support with fingerprint-based deduplication (`pkg/sender/slack/sender_slack.go:150-194`)
- File upload support for logs (`pkg/sender/slack/sender_slack.go:597-650`)
- Enrichment deduplication (`pkg/sender/slack/sender_slack.go:450-470`)
- Resolved alert threading (posts as replies)
- Status indicators (🔥 firing, ✅ resolved)
- Table-to-CSV conversion for large datasets
- Thread timestamp caching with TTL

**Current Message Format** (needs improvement):

**Header** (`pkg/sender/slack/sender_slack.go:1022-1044`):
```go
func (s *SenderSlack) formatHeader(issue *issuepkg.Issue) string {
    var statusText string
    var statusEmoji string
    if issue.IsResolved() {
        statusText = "Alert resolved"
        statusEmoji = "✅"
    } else {
        statusText = "Alert firing"
        statusEmoji = "🔥"
    }

    severityText := s.getSeverityText(issue.Severity)

    return fmt.Sprintf("%s %s %s\n%s",
        statusEmoji, statusText, severityText, issue.Title)
}
```

**Current Structure**:
- Single-line header with emoji + status + severity + title
- Uses attachments for metadata (source, cluster, namespace) - older Slack API pattern
- Enrichments added as separate blocks

**Gap Analysis**:

| Feature | Current Status | Robusta Pattern | Gap |
|---------|---------------|-----------------|-----|
| Header Format | Single line | Multi-block with context bar | ❌ Needs restructure |
| Metadata Display | Attachment fields | Context block elements | ❌ Needs context blocks |
| Badge Styling | Plain text | Inline code `` `value` `` | ❌ Needs markdown |
| Severity Display | Text "High" | Emoji 🔴 High | ⚠️ Partially implemented |
| Crash Info Section | Generic enrichment | Structured section | ❌ Needs specific formatting |
| File Links | Basic upload | Permalink with title | ⚠️ Has upload, needs better display |
| Alert Type Indicator | In attachment | Context bar with emoji | ❌ Needs context element |
| Threading | ✅ Implemented | Fingerprint-based | ✅ Already working |
| Deduplication | ✅ Implemented | Hash-based | ✅ Already working |

#### Template-Based Approach (Recommended)

**Why Templates?**

Following Robusta's architecture pattern, cano-collector should use Go templates for Slack message formatting instead of hardcoding block construction in Go code. This provides:

1. **Separation of Concerns** - Message structure separate from business logic
2. **Easier Maintenance** - Non-developers can modify message format
3. **Template Reusability** - Share common patterns across message types
4. **Testing** - Test templates independently of sender logic
5. **Flexibility** - Easy A/B testing of different formats

**Recommended Architecture**:

```
pkg/sender/slack/
├── sender_slack.go           # Main sender logic
├── templates/
│   ├── loader.go             # Template loading and rendering
│   ├── header.tmpl           # Header block template
│   ├── context_bar.tmpl      # Context bar template
│   ├── crash_info.tmpl       # Crash info section template
│   ├── description.tmpl      # Description block template
│   └── enrichment.tmpl       # Enrichment blocks template
└── template_context.go       # Template context structs
```

**Template System Design**:

```go
// pkg/sender/slack/template_context.go
package slack

type MessageContext struct {
    Status          string  // "firing" or "resolved"
    StatusEmoji     string  // "🔥" or "✅"
    StatusText      string  // "Alert firing" or "Alert resolved"
    Severity        string  // "High", "Medium", "Low"
    SeverityEmoji   string  // "🔴", "🟡", "🟢"
    Title           string
    Description     string
    AlertType       string  // "Prometheus Alert" or "K8s Event"
    AlertTypeEmoji  string  // "📊" or "👀"
    Cluster         string
    Namespace       string
    PodName         string
    Source          string
    Links           []Link
    CrashInfo       *CrashInfo
    Enrichments     []EnrichmentData
}

type CrashInfo struct {
    Container   string
    Restarts    string
    Status      string
    Reason      string
}

type Link struct {
    Text string
    URL  string
}

type EnrichmentData struct {
    Type      string  // "file", "table", "markdown"
    Title     string
    Content   string
    FileLink  string
}
```

**Template Loader**:

```go
// pkg/sender/slack/templates/loader.go
package templates

import (
    "bytes"
    "encoding/json"
    "text/template"
    "embed"

    slackapi "github.com/slack-go/slack"
)

//go:embed *.tmpl
var templateFS embed.FS

type TemplateLoader struct {
    templates map[string]*template.Template
}

func NewTemplateLoader() (*TemplateLoader, error) {
    loader := &TemplateLoader{
        templates: make(map[string]*template.Template),
    }

    // Load all templates
    files := []string{
        "header.tmpl",
        "context_bar.tmpl",
        "crash_info.tmpl",
        "description.tmpl",
        "enrichment.tmpl",
    }

    for _, file := range files {
        tmpl, err := template.ParseFS(templateFS, file)
        if err != nil {
            return nil, err
        }
        loader.templates[file] = tmpl
    }

    return loader, nil
}

func (l *TemplateLoader) RenderToBlocks(templateName string, context interface{}) ([]slackapi.Block, error) {
    tmpl, exists := l.templates[templateName]
    if !exists {
        return nil, fmt.Errorf("template not found: %s", templateName)
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, context); err != nil {
        return nil, err
    }

    // Parse JSON blocks from template output
    var blocks []map[string]interface{}
    if err := json.Unmarshal(buf.Bytes(), &blocks); err != nil {
        return nil, err
    }

    // Convert to slack.Block
    return convertToSlackBlocks(blocks)
}
```

**Template Example - Header**:

```json
{{/* pkg/sender/slack/templates/header.tmpl */}}
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "{{.StatusEmoji}} {{.StatusText}} {{.StatusEmoji}}  {{.SeverityEmoji}} {{.Severity}}\n*{{.Title}}*"
    }
  }
]
```

**Template Example - Context Bar**:

```json
{{/* pkg/sender/slack/templates/context_bar.tmpl */}}
[
  {
    "type": "context",
    "elements": [
      {
        "type": "mrkdwn",
        "text": "{{.AlertTypeEmoji}} Type: {{.AlertType}}"
      }
      {{if .Cluster}},
      {
        "type": "mrkdwn",
        "text": "Cluster: `{{.Cluster}}`"
      }
      {{end}}
      {{if .Namespace}},
      {
        "type": "mrkdwn",
        "text": "Namespace: `{{.Namespace}}`"
      }
      {{end}}
      {{if .PodName}},
      {
        "type": "mrkdwn",
        "text": "📦 Pod: `{{.PodName}}`"
      }
      {{end}}
    ]
  }
]
```

**Template Example - Crash Info**:

```json
{{/* pkg/sender/slack/templates/crash_info.tmpl */}}
{{if .CrashInfo}}
[
  {
    "type": "section",
    "text": {
      "type": "mrkdwn",
      "text": "*Crash Info*\n• Container: `{{.CrashInfo.Container}}`\n• Restarts: `{{.CrashInfo.Restarts}}`\n• Status: `{{.CrashInfo.Status}}`\n• Reason: `{{.CrashInfo.Reason}}`"
    }
  }
]
{{end}}
```

**Updated Sender Implementation**:

```go
// pkg/sender/slack/sender_slack.go
type SenderSlack struct {
    apiKey         string
    channel        string
    channelID      string
    logger         logger_interfaces.LoggerInterface
    unfurlLinks    bool
    slackClient    sender_interfaces.SlackClientInterface
    threadManager  sender_interfaces.SlackThreadManagerInterface
    templateLoader *templates.TemplateLoader  // NEW: Template system
    // ... rest of fields
}

func NewSenderSlack(apiKey, channel string, unfurlLinks bool, logger logger_interfaces.LoggerInterface, client util.HTTPClient) *SenderSlack {
    // ... existing initialization

    // Initialize template loader
    templateLoader, err := templates.NewTemplateLoader()
    if err != nil {
        logger.Error("Failed to load Slack message templates", zap.Error(err))
        // Fallback to hardcoded format or panic
    }

    return &SenderSlack{
        // ... existing fields
        templateLoader: templateLoader,
    }
}

func (s *SenderSlack) buildSlackBlocks(issue *issuepkg.Issue) []slackapi.Block {
    // Build template context from issue
    context := s.buildMessageContext(issue)

    blocks := []slackapi.Block{}

    // Render header from template
    headerBlocks, err := s.templateLoader.RenderToBlocks("header.tmpl", context)
    if err != nil {
        s.logger.Error("Failed to render header template", zap.Error(err))
        // Fallback to hardcoded header
        headerBlocks = s.buildHeaderBlockFallback(issue)
    }
    blocks = append(blocks, headerBlocks...)

    // Render context bar from template
    contextBlocks, err := s.templateLoader.RenderToBlocks("context_bar.tmpl", context)
    if err != nil {
        s.logger.Error("Failed to render context bar template", zap.Error(err))
    } else {
        blocks = append(blocks, contextBlocks...)
    }

    // Divider
    blocks = append(blocks, slackapi.NewDividerBlock())

    // Render description if present
    if context.Description != "" {
        descBlocks, _ := s.templateLoader.RenderToBlocks("description.tmpl", context)
        blocks = append(blocks, descBlocks...)
    }

    // Render crash info if present
    if context.CrashInfo != nil {
        crashBlocks, _ := s.templateLoader.RenderToBlocks("crash_info.tmpl", context)
        blocks = append(blocks, crashBlocks...)
    }

    // Render enrichments
    for _, enrichment := range context.Enrichments {
        enrichBlocks, _ := s.templateLoader.RenderToBlocks("enrichment.tmpl", enrichment)
        blocks = append(blocks, enrichBlocks...)
    }

    return blocks
}

func (s *SenderSlack) buildMessageContext(issue *issuepkg.Issue) *MessageContext {
    context := &MessageContext{
        Title:       issue.Title,
        Description: issue.Description,
        Cluster:     issue.Cluster,
        Namespace:   issue.Subject.Namespace,
        PodName:     issue.Subject.Name,
        Source:      issue.Source,
    }

    // Status
    if issue.IsResolved() {
        context.Status = "resolved"
        context.StatusEmoji = "✅"
        context.StatusText = "Alert resolved"
    } else {
        context.Status = "firing"
        context.StatusEmoji = "🔥"
        context.StatusText = "Alert firing"
    }

    // Severity
    context.Severity = s.getSeverityText(issue.Severity)
    context.SeverityEmoji = s.getSeverityEmoji(issue.Severity)

    // Alert type
    if issue.Source == "prometheus" {
        context.AlertType = "Prometheus Alert"
        context.AlertTypeEmoji = "📊"
    } else {
        context.AlertType = "K8s Event"
        context.AlertTypeEmoji = "👀"
    }

    // Crash info (if pod alert)
    if containerName := issue.GetLabel("container"); containerName != "" {
        context.CrashInfo = &CrashInfo{
            Container: containerName,
            Restarts:  issue.GetLabel("restarts"),
            Status:    issue.GetLabel("status"),
            Reason:    issue.GetLabel("reason"),
        }
    }

    // Links
    for _, link := range issue.GeneratedLinks {
        context.Links = append(context.Links, Link{
            Text: link.Text,
            URL:  link.URL,
        })
    }

    // Enrichments
    for _, enrichment := range issue.Enrichments {
        context.Enrichments = append(context.Enrichments, EnrichmentData{
            Type:     string(enrichment.Type),
            Title:    enrichment.Title,
            Content:  enrichment.Content,
            FileLink: enrichment.FileInfo.Permalink,
        })
    }

    return context
}
```

**Benefits of Template Approach**:

1. **Cleaner Code** - Sender logic focuses on data preparation, not formatting
2. **Easy Updates** - Change message format without recompiling (if using external files)
3. **Testability** - Test templates with sample data independently
4. **Consistency** - Ensure all messages follow same structure
5. **Documentation** - Templates serve as message format documentation
6. **Future-Proofing** - Easy to add new message types or variations

**Migration Path**:

1. Create template system alongside existing code
2. Add feature flag `use_templates: true/false`
3. Test templates with subset of alerts
4. Gradually roll out to all destinations
5. Remove old hardcoded formatting once stable

#### Alternative: Hardcoded Changes (Not Recommended)

If template system is too complex for initial implementation, here are the hardcoded changes:

**Change 1: Restructure Header Block** (`pkg/sender/slack/sender_slack.go:1022-1044`)

**Current**:
```go
func (s *SenderSlack) formatHeader(issue *issuepkg.Issue) string {
    return fmt.Sprintf("%s %s %s\n%s",
        statusEmoji, statusText, severityText, issue.Title)
}
```

**New**:
```go
func (s *SenderSlack) buildHeaderBlock(issue *issuepkg.Issue) slackapi.Block {
    var statusText string
    var statusEmoji string
    if issue.IsResolved() {
        statusText = "Alert resolved"
        statusEmoji = "✅"
    } else {
        statusText = "Alert firing"
        statusEmoji = "🔥"
    }

    severityEmoji := s.getSeverityEmoji(issue.Severity)
    severityText := s.getSeverityText(issue.Severity)

    // Format: 🔥 Alert firing 🔥  🔴 High\n*Title*
    headerText := fmt.Sprintf("%s %s %s  %s %s\n*%s*",
        statusEmoji, statusText, statusEmoji,
        severityEmoji, severityText,
        issue.Title)

    return slackapi.NewSectionBlock(
        slackapi.NewTextBlockObject("mrkdwn", headerText, false, false),
        nil, nil,
    )
}
```

**Change 2: Add Context Bar Block** (new method)

```go
func (s *SenderSlack) buildContextBarBlock(issue *issuepkg.Issue) slackapi.Block {
    elements := []slackapi.MixedElement{}

    // Alert type indicator
    alertTypeEmoji := "📊"
    if issue.Source == "kubernetes" {
        alertTypeEmoji = "👀"
    }
    elements = append(elements,
        slackapi.NewTextBlockObject("mrkdwn",
            fmt.Sprintf("%s Type: %s", alertTypeEmoji, issue.Source),
            false, false))

    // Cluster
    if issue.Cluster != "" {
        elements = append(elements,
            slackapi.NewTextBlockObject("mrkdwn",
                fmt.Sprintf("Cluster: `%s`", issue.Cluster),
                false, false))
    }

    // Namespace (pod-related alerts)
    if issue.Subject.Namespace != "" {
        elements = append(elements,
            slackapi.NewTextBlockObject("mrkdwn",
                fmt.Sprintf("Namespace: `%s`", issue.Subject.Namespace),
                false, false))
    }

    // Pod name (if pod-related)
    if issue.Subject.Kind == "Pod" && issue.Subject.Name != "" {
        elements = append(elements,
            slackapi.NewTextBlockObject("mrkdwn",
                fmt.Sprintf("📦 Pod: `%s`", issue.Subject.Name),
                false, false))
    }

    return slackapi.NewContextBlock("", elements...)
}
```

**Change 3: Add Crash Info Enrichment Block** (new method)

```go
func (s *SenderSlack) buildCrashInfoBlock(issue *issuepkg.Issue) slackapi.Block {
    // Extract crash-related metadata from issue labels/annotations
    containerName := issue.GetLabel("container")
    restartCount := issue.GetLabel("restarts")
    status := issue.GetLabel("status")
    reason := issue.GetLabel("reason")

    if containerName == "" {
        return nil // No crash info available
    }

    crashInfoText := fmt.Sprintf("*Crash Info*\n"+
        "• Container: `%s`\n"+
        "• Restarts: `%s`\n"+
        "• Status: `%s`\n"+
        "• Reason: `%s`",
        containerName, restartCount, status, reason)

    return slackapi.NewSectionBlock(
        slackapi.NewTextBlockObject("mrkdwn", crashInfoText, false, false),
        nil, nil,
    )
}
```

**Change 4: Update `buildSlackBlocks()` Method** (`pkg/sender/slack/sender_slack.go:272-363`)

**Current Structure**:
```go
func (s *SenderSlack) buildSlackBlocks(issue *issuepkg.Issue) []slackapi.Block {
    blocks := []slackapi.Block{}

    // Header
    headerText := s.formatHeader(issue)
    blocks = append(blocks, slackapi.NewSectionBlock(
        slackapi.NewTextBlockObject("mrkdwn", headerText, false, false),
        nil, nil,
    ))

    // Add links, description, enrichments...
    return blocks
}
```

**New Structure**:
```go
func (s *SenderSlack) buildSlackBlocks(issue *issuepkg.Issue) []slackapi.Block {
    blocks := []slackapi.Block{}

    // 1. Header block (multi-line with emojis)
    blocks = append(blocks, s.buildHeaderBlock(issue))

    // 2. Context bar (type, cluster, namespace, pod)
    blocks = append(blocks, s.buildContextBarBlock(issue))

    // 3. Divider
    blocks = append(blocks, slackapi.NewDividerBlock())

    // 4. Description (if available)
    if issue.Description != "" {
        blocks = append(blocks, slackapi.NewSectionBlock(
            slackapi.NewTextBlockObject("mrkdwn", issue.Description, false, false),
            nil, nil,
        ))
    }

    // 5. Links (if available)
    if len(issue.GeneratedLinks) > 0 {
        blocks = append(blocks, s.buildLinksBlock(issue))
    }

    // 6. Crash Info (for pod alerts)
    if crashBlock := s.buildCrashInfoBlock(issue); crashBlock != nil {
        blocks = append(blocks, crashBlock)
    }

    // 7. Enrichments (deduplicated)
    enrichmentBlocks := s.buildEnrichmentBlocks(issue)
    blocks = append(blocks, enrichmentBlocks...)

    return blocks
}
```

**Change 5: Remove Attachment-Based Metadata** (`pkg/sender/slack/sender_slack.go:365-399`)

**Current**: Uses colored attachments for source/cluster/namespace
**New**: These should be in context bar block instead

**Action**: Remove or simplify `buildSlackAttachments()` method since metadata is now in context blocks.

**Change 6: Improve File Link Display** (`pkg/sender/slack/sender_slack.go:597-650`)

**Current**: Uploads files and gets permalink
**New**: Format file enrichment blocks with clear titles and links

```go
func (s *SenderSlack) buildFileEnrichmentBlock(enrichment *issuepkg.Enrichment) slackapi.Block {
    // Assuming enrichment.FileInfo contains permalink and metadata
    fileTitle := enrichment.Title
    fileLink := enrichment.FileInfo.Permalink

    fileText := fmt.Sprintf("*%s*\n<%s|View File>", fileTitle, fileLink)

    return slackapi.NewSectionBlock(
        slackapi.NewTextBlockObject("mrkdwn", fileText, false, false),
        nil, nil,
    )
}
```

#### Implementation Plan

**Phase 1: Header and Context Bar** (Priority: High)
1. ✅ Create `buildHeaderBlock()` method
2. ✅ Create `buildContextBarBlock()` method
3. ✅ Add helper method `getSeverityEmoji()`
4. ✅ Update `buildSlackBlocks()` to use new blocks
5. ✅ Remove old single-line header format
6. Test with basic firing/resolved alerts

**Phase 2: Crash Info Enrichment** (Priority: High)
1. ✅ Create `buildCrashInfoBlock()` method
2. ✅ Add label extraction helpers
3. ✅ Integrate into `buildSlackBlocks()`
4. Test with CrashLoopBackOff alerts

**Phase 3: Badge Styling** (Priority: Medium)
1. ✅ Update all metadata fields to use inline code formatting
2. ✅ Apply to: container names, pod names, namespaces, status values, reasons
3. Test rendering in Slack

**Phase 4: Cleanup and Optimization** (Priority: Low)
1. ✅ Remove/simplify `buildSlackAttachments()` method
2. ✅ Consolidate metadata display in context blocks
3. ✅ Update file enrichment display format
4. Performance testing with high alert volume

**Phase 5: Documentation** (Priority: Medium)
1. Update `docs/configuration/destinations/slack.rst` with new format examples
2. Add screenshots of new message format
3. Document configuration options for enrichments

#### Testing Strategy

**Test Cases**:
1. **Basic Alert Firing** - Verify header + context bar render correctly
2. **Pod Crash Alert** - Verify crash info section appears
3. **Alert Resolution** - Verify threading and ✅ emoji
4. **File Enrichment** - Verify logs upload and display properly
5. **Multiple Enrichments** - Verify deduplication works
6. **Long Descriptions** - Verify text truncation/formatting
7. **Missing Metadata** - Verify graceful degradation (no panic)

**Validation Commands**:
```bash
# Deploy updated cano-collector
helm upgrade cano-collector ./helm/cano-collector

# Trigger test alert
./tests/scripts/run-test.sh crash-loop busybox-crash

# Check Slack channel for new format
# Verify: Multi-line header, context bar, crash info section

# Check logs for any formatting errors
kubectl logs -n kubecano -l app=cano-collector | grep -i "slack\|error"
```

#### Migration Notes

**Backward Compatibility**:
- Existing threading will continue to work (fingerprint logic unchanged)
- Existing file uploads will continue to work (no API changes)
- Old messages won't be reformatted (only new alerts use new format)

**Configuration Changes** (if needed):
- No breaking changes to `values.yaml` configuration
- New enrichment parameters are optional
- Default behavior: Show crash info for pod alerts automatically

**Rollback Plan**:
- Keep old `formatHeader()` method as `formatHeaderLegacy()`
- Add feature flag `slack_new_format: true/false` in config
- Allow gradual rollout per destination

### 4. Implementation
- [ ] Implement new Slack message formatting
- [ ] Add badge styling with inline code
- [ ] Create crash info enrichment blocks
- [ ] Test with various alert types
- [ ] Update documentation with examples
