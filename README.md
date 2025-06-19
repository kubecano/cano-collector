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
- **Deep context** behind alerts and events.
- **Flexible routing** to the right teams.
- **Optional AI assistance** (coming soon).
- **Unified view** of Kubernetes health, enriched and sent where it matters.

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

1. **Builds a structured object** that includes:
   - Type: `alert` or `event`
   - Source: `prometheus`, `k8s`, `helm`, etc.
   - Timestamps: created/started/resolved

2. **Enriches it with context**, such as:
   - Recent **pod logs**
   - **CPU/memory requests and limits**
   - **OOMKilled reasons** or container exit codes
   - (Planned) `jstack` traces for JVM applications

3. **Sends enriched data** to various destinations:
   - 🧭 Kubecano SaaS (for dashboards & storage)
   - 💬 Slack or Microsoft Teams channels
   - 📟 PagerDuty or OpsGenie alerts
   - 🔀 Kafka topics for downstream processing

---

## 📦 Architecture Overview

```text
             ┌──────────────────┐
             │  Alertmanager    │
             └────────┬─────────┘
                      │
                      ▼
         ┌─────────────────────────┐
         │      Cano Collector     │
         │  (Deployed on K8s)      │
         └─────────────────────────┘
                      │
    ┌─────────────────┼────────────────────────────┐
    ▼                 ▼                            ▼
[SaaS API]     [Notification Channels]      [Kafka Streams]
                  Slack, Teams               For async usage
               OpsGenie, PagerDuty
```

Routing is team-aware: different teams can receive alerts in separate channels, PagerDuty services, or Kafka topics.

## 🔧 Installation
Coming soon – Helm chart and manifests for deploying Cano Collector on your Kubernetes cluster.

## 📌 Use Cases
* Your app is OOMKilled → Cano Collector captures pod logs + resource config + alert, and sends a full incident package.

* Helm release fails → get a Slack message with release name, namespace, and failed hook logs.

* An alert fires → your on-call team receives enriched payloads with context to act faster and smarter.

## 🔮 What’s Next?
We are actively working on:

* 🧠 AI-powered incident analysis via LLMs (e.g., GPT-4, Claude, etc.)

* 🕵️‍♀️ Root cause suggestions and remediation hints

* 🌐 Web dashboard for managing and viewing incidents

* 🧩 Integration with additional notification and observability tools

## 👥 Who is this for?
If you’re a:

* DevOps engineer managing production Kubernetes

* Developer tired of vague alerts

* SRE building observability tooling

…then Cano Collector is for you.

## 📬 Get Involved
Join us in making Kubernetes incidents understandable!

* GitHub: [Kubecano organization](https://github.com/kubecano)

*  Coming soon: Discussions, Slack workspace, contribution guide

## 📝 License
Cano Collector is licensed under the [Apache 2.0 License.](./LICENSE)
