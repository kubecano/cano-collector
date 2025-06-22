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

### ✅ Available Now
- **Slack Integration**: Full-featured Slack destination with:
  - Rich message formatting with blocks and attachments
  - Thread support for related alerts
  - File uploads for logs and data
  - Color-coded messages based on severity
  - Table formatting for structured data

### 🚧 Coming Soon
- **MS Teams Integration**: Adaptive Cards support
- **PagerDuty Integration**: Incident lifecycle management
- **OpsGenie Integration**: Dynamic team routing
- **Jira Integration**: Ticket creation and management
- **DataDog Integration**: Event correlation
- **Kafka Integration**: Data streaming
- **ServiceNow Integration**: Incident management

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

### Current (Slack MVP)
* **OOMKilled Pod**: Get a Slack message with pod logs, resource limits, and memory usage
* **CrashLoopBackOff**: Receive container logs, restart count, and exit codes
* **Helm Release Failure**: See release details, failed hooks, and rollback status

### Planned
* **Multi-channel routing**: Send different alerts to different teams
* **Incident management**: Create tickets in Jira, incidents in PagerDuty
* **Data streaming**: Send events to Kafka for downstream processing
* **AI analysis**: Automated root cause analysis and remediation suggestions

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

### v1.0 (MVP) - Current
- ✅ Slack destination
- ✅ Basic alert enrichment
- ✅ Kubernetes event processing

### v1.1 - Q2 2024
- 🚧 MS Teams destination
- 🚧 PagerDuty destination
- 🚧 Enhanced enrichment capabilities

### v1.2 - Q3 2024
- 🚧 Jira integration
- 🚧 OpsGenie integration
- 🚧 Kafka streaming

### v2.0 - Q4 2024
- 🚧 AI-powered analysis
- 🚧 Web dashboard
- 🚧 Advanced routing rules

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
