# cano-collector

[![SonarQube Cloud](https://sonarcloud.io/images/project_badges/sonarcloud-highlight.svg)](https://sonarcloud.io/summary/new_code?id=kubecano_cano-collector)

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=kubecano_cano-collector&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=kubecano_cano-collector)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=kubecano_cano-collector&metric=bugs)](https://sonarcloud.io/summary/new_code?id=kubecano_cano-collector)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=kubecano_cano-collector&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=kubecano_cano-collector)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=kubecano_cano-collector&metric=coverage)](https://sonarcloud.io/summary/new_code?id=kubecano_cano-collector)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=kubecano_cano-collector&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=kubecano_cano-collector)

# Cano Collector

**Cano Collector** is an open-source alert and event ingestion agent for Kubernetes, designed to help developers and DevOps teams better understand incidents in their clusters by enriching raw alerts and events with valuable context.

Cano Collector is part of the broader **[Kubecano](https://github.com/kubecano)** platform. It runs on Kubernetes clusters and connects telemetry data with notifications, enrichment pipelines, and (in future releases) AI-based analysis.

---

## 🚀 Why Cano Collector?

Traditional alerts and crash loops often lack the full story. Cano Collector gives you:
- **Deep context** behind alerts and events
- **Flexible routing** to the right teams
- **Rich formatting** with structured data and attachments
- **Unified view** of Kubernetes health, enriched and sent where it matters

Whether it's an `OOMKilled` pod or a `CrashLoopBackOff`, Cano Collector helps you understand *why* something broke — not just that it did.

---

## 🧩 What It Does

Cano Collector listens for **Kubernetes cluster signals**, including:

- 📣 **Alerts from Alertmanager**
- ⚠️ **Kubernetes Events** such as:
  - Pod restarts / CrashLoops
  - Helm release failures
  - Resource quota violations

For each alert or event, Cano Collector:

1. **Builds a structured `Issue` object** that includes:
   - Type: `alert` or `event`
   - Source: `prometheus`, `k8s`, `helm`, etc.
   - Severity: `HIGH`, `LOW`, `INFO`, `DEBUG`
   - Timestamps: created/started/resolved

2. **Enriches it with context** through `Enrichment` blocks:
   - **Pod logs** as `MarkdownBlock`
   - **Resource configuration** as `TableBlock`
   - **File attachments** as `FileBlock`
   - **Structured data** as `JsonBlock`

3. **Sends enriched data** to configured destinations:
   - 💬 **Slack channels** (MVP - Available Now)
   - 🧭 Kubecano SaaS (Planned)
   - 📟 PagerDuty, OpsGenie (Planned)
   - 🔀 Kafka topics (Planned)

---

## 📦 Architecture Overview

Cano Collector follows a clean architecture pattern with clear separation of concerns:

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Alertmanager  │    │   Kubernetes    │    │   Other Sources │
│   (Prometheus)  │    │     Events      │    │                 │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼─────────────┐
                    │     Cano Collector        │
                    │  (Deployed on K8s)        │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │      Destination          │
                    │   (Strategy Pattern)      │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │       Sender              │
                    │   (Factory Pattern)       │
                    └─────────────┬─────────────┘
                                  │
                    ┌─────────────▼─────────────┐
                    │    External Services      │
                    │   Slack, Teams, etc.      │
                    └───────────────────────────┘
```

### Core Components

- **`Issue`**: The central data structure containing alert/event information
- **`Enrichment`**: Additional context blocks (logs, tables, files, etc.)
- **`Destination`**: Strategy pattern implementation for different notification channels
- **`Sender`**: Factory pattern implementation for API communication

---

## 🎯 Current Status (MVP)

### ✅ Available Now (v0.0.24)
- **Alertmanager Integration**: Full webhook-based alert processing from Prometheus/Alertmanager
- **Slack Integration**: Full-featured Slack destination with:
  - Rich message formatting with blocks and attachments
  - Thread support for related alerts with intelligent grouping
  - File uploads for logs and structured data
  - Color-coded messages based on severity
  - Table formatting for structured data
  - Message deduplication and fingerprinting
- **Workflow System**: Extensible alert processing with:
  - Built-in workflow actions (pod logs, issue enrichment)
  - Configurable triggers and conditions
  - Support for custom TypeScript workflows (planned)
- **Alert Enrichment**: Automatic context gathering from Kubernetes:
  - Pod logs (current and previous containers)
  - Resource metadata and labels
  - Alert annotations and severity mapping
- **Observability**: Production-ready monitoring with:
  - Health checks (liveness and readiness)
  - Prometheus metrics export
  - OpenTelemetry tracing support
  - Structured logging

### 🔮 Next Releases
Additional integrations and features will be added in future versions:
- **MS Teams Integration**: Adaptive Cards support
- **Direct Kubernetes Event Processing**: Watch and process K8s events (BackOff, CrashLoopBackOff, etc.)
- **Team-Based Routing**: Multi-team alert distribution
- **PagerDuty Integration**: Incident lifecycle management
- **Jira Integration**: Ticket creation and management
- **Additional Senders**: DataDog, Kafka, ServiceNow, OpsGenie

---

## 🔧 Installation

### Prerequisites
- Kubernetes cluster (1.19+)
- Helm 3.x
- Alertmanager configured (optional)

### Quick Start

1. **Add the Helm repository**:
   ```bash
   helm repo add kubecano https://kubecano.github.io/helm-charts
   helm repo update
   ```

2. **Create a values file** (`values.yaml`):
   ```yaml
   destinations:
     - name: slack_production
       type: slack
       params:
         webhook_url: "https://hooks.slack.com/services/YOUR/WEBHOOK/URL"
         channel: "#alerts"
         username: "Cano Collector"
   ```

3. **Install Cano Collector**:
   ```bash
   helm install cano-collector kubecano/cano-collector -f values.yaml
   ```

### Configuration

See the [Configuration Guide](docs/configuration/index.rst) for detailed setup instructions.

---

## 📌 Use Cases

### Current Implementation (v0.0.24)
Cano Collector currently processes **Prometheus alerts from Alertmanager** and enriches them with Kubernetes context:

* **Pod CrashLoopBackOff Alerts**:
  - Receive `KubePodCrashLooping` alerts in Slack
  - Automatic pod logs attachment (previous and current container)
  - Container restart count and exit codes
  - Resource configuration and limits
  - Color-coded severity and threaded conversations

* **General Kubernetes Alerts**:
  - Any Prometheus/Alertmanager alert routed to cano-collector
  - Rich Slack formatting with alert metadata
  - Alert annotations and labels as enrichments
  - Resolved alerts posted as thread replies

* **Custom Alert Enrichment**:
  - Configurable workflows for different alert types
  - Pod logs collection for container failures
  - Kubernetes resource metadata extraction
  - Extensible action system for custom enrichments

### Planned Features (Future Releases)
* **Direct Kubernetes Event Processing**: Watch and process K8s events in real-time (BackOff, ImagePull, Eviction, etc.)
* **Multi-Channel Routing**: Send different alerts to different teams and destinations
* **Additional Destinations**: MS Teams, PagerDuty, Jira, DataDog, Kafka, ServiceNow
* **Advanced Enrichments**:
  - OOMKilled analysis with memory graphs
  - Resource usage trends
  - Node health correlations
  - Deployment history and changes
* **Team-Based Alert Distribution**: Route alerts based on namespace, labels, and severity to specific teams
* **Incident Management Integration**: Create tickets in Jira, incidents in PagerDuty with full lifecycle tracking

---

## 🏗️ Development

### Architecture Documentation
- [Architecture Overview](docs/architecture/index.rst)
- [Data Model](docs/architecture/data_model.rst)
- [Design Patterns](docs/architecture/patterns.rst)
- [Slack Integration](docs/architecture/destination/slack.rst)

### Contributing
We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Building
```bash
# Build the binary
go build -o cano-collector ./main.go

# Build the Docker image
docker build -t cano-collector .

# Run tests
go test ./...
```

---

## 🔮 Roadmap

### ✅ Phase 1: MVP (Completed - v0.0.24)
Core alert processing and Slack integration:
- ✅ Alertmanager webhook integration
- ✅ Slack destination with full feature set
- ✅ Workflow system with configurable actions
- ✅ Basic alert enrichment (pod logs, metadata)
- ✅ Health checks and metrics
- ✅ OpenTelemetry tracing

### 🚧 Phase 2: Core Platform Features (In Progress)
Expanding destination support and processing capabilities:
- 🔨 Direct Kubernetes Event Processing (watch K8s API events)
- 🔨 MS Teams destination
- 🔨 PagerDuty integration
- 🔨 Enhanced workflow system
- 🔨 Team-based routing

### 📋 Phase 3: Enterprise Features (Planned)
Advanced integrations and analysis:
- 📅 Jira Service Management integration
- 📅 OpsGenie integration
- 📅 DataDog event correlation
- 📅 Kafka streaming
- 📅 Alert deduplication system
- 📅 Async processing queue

### 🌟 Phase 4: Advanced Capabilities (Future)
Platform ecosystem and intelligence:
- 🌟 Kubecano CLI tool
- 🌟 Official Slack App
- 🌟 Advanced monitoring and observability
- 🌟 Custom TypeScript workflows (runtime)
- 🌟 ServiceNow integration

### 🚀 Phase 5: SaaS Platform (Long-term Vision)
Multi-tenant SaaS offering:
- 🚀 Web dashboard and SaaS platform
- 🚀 Multi-cluster management
- 🚀 AI-powered root cause analysis
- 🚀 Automated remediation suggestions
- 🚀 Advanced correlation and anomaly detection

**Note**: This roadmap is subject to change based on community feedback and priorities. Specific release dates are not provided as development is driven by community needs and contributions.

---

## 👥 Who is this for?

If you're a:
* **DevOps engineer** managing production Kubernetes
* **Developer** tired of vague alerts
* **SRE** building observability tooling
* **Platform team** looking for better incident response

…then Cano Collector is for you.

---

## 📬 Get Involved

Join us in making Kubernetes incidents understandable!

* **GitHub**: [Kubecano organization](https://github.com/kubecano)
* **Documentation**: [docs/](docs/)
* **Issues**: [GitHub Issues](https://github.com/kubecano/cano-collector/issues)

---

## 📝 License

Cano Collector is licensed under the [Apache 2.0 License](./LICENSE).
