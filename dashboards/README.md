# Grafana Dashboards for Cano Collector

This folder contains Grafana dashboards for monitoring the Cano Collector application.

## 📊 Available Dashboards

### `cano-collector-dashboard.json`

Main dashboard monitoring the Cano Collector application with the following sections:

#### 🚀 Application Health & Performance

- **HTTP Requests Rate** - HTTP request frequency grouped by method, path and status
- **HTTP Response Times** - Response time percentiles (95th, 50th percentile)

#### 💻 Go Runtime Metrics

- **Goroutines** - Number of active goroutines
- **Memory Usage** - Memory utilization (allocated, heap)
- **GC Duration** - Garbage Collection duration

#### 📊 Business Metrics - Alert Processing

- **Alerts Processed Rate** - Alert processing frequency by name and severity
- **Alert Processing Duration** - Alert processing time (percentiles)

#### 🎯 Destination & Routing Metrics

- **Messages Sent to Destinations** - Messages sent to destinations by type and status
- **Team Routing Decisions** - Routing decisions by teams

#### 🚨 Error Tracking

- **Alert Processing Errors** - Alert processing errors
- **Destination Send Errors** - Destination sending errors

#### 📈 Summary Stats

- **Total Alerts Processed** - Total number of processed alerts
- **Total Messages Sent** - Total number of sent messages  
- **Active Teams** - Number of active teams
- **Error Rate** - Error frequency per second

### `cano-collector-sla-dashboard.json`

Business dashboard focused on SLA/SLO and key business metrics:

#### 🎯 Service Level Objectives

- **Alert Processing SLO (99.9%)** - Alert processing success rate
- **Destination Delivery SLO (99.5%)** - Message delivery success rate
- **Response Time SLO (< 500ms)** - 95th percentile response times
- **System Availability** - System availability based on HTTP 5xx errors

#### 📊 Business Impact Metrics

- **Alert Processing Volume** - Volume of processed alerts per minute
- **End-to-End Processing Latency** - Full processing time (P50, P95, P99)

#### 🚨 Error Budget Tracking

- **Error Budget Burn Rate** - Error budget burn rate
- **Mean Time to Recovery (MTTR)** - Mean time to recovery

#### 📈 Performance Trends

- **Throughput by Team** - Throughput by teams
- **Destination Performance** - Performance of individual destination types

#### 💼 Business KPIs

- **Daily Alert Volume** - Daily alert volume
- **Customer Satisfaction (SLA)** - Customer satisfaction expressed as SLA
- **Active Integrations** - Number of active integrations
- **Resource Efficiency** - Resource efficiency (alerts/goroutine)

## 🚀 Installation

### Option 1: Import via Grafana UI

1. Open Grafana
2. Go to **Dashboards** → **Import**
3. Paste the JSON file contents
4. Click **Import**

### Option 2: Provisioning

Copy files to the provisioning folder in Grafana:

```bash
cp *.json /etc/grafana/provisioning/dashboards/
```

## 📋 Requirements

- Grafana 8.0+
- Prometheus as data source
- Cano Collector application exporting metrics on `/metrics`

## 🔧 Configuration

### Main Dashboard (`cano-collector-dashboard.json`)

- **Refresh**: 30s
- **Time Range**: Last 1 hour
- **Data Source**: Prometheus (default)

### SLA/SLO Dashboard (`cano-collector-sla-dashboard.json`)

- **Refresh**: 1m
- **Time Range**: Last 24 hours
- **Data Source**: Prometheus (default)

## 📊 Available Metrics

The dashboards use the following metrics from the application:

### HTTP Metrics

- `http_requests_total` - Number of HTTP requests
- `cano_http_request_duration_seconds` - HTTP request duration

### Alert Processing Metrics

- `alertmanager_alerts_total` - Alerts received from AlertManager
- `cano_alerts_processed_total` - Processed alerts
- `cano_alerts_processing_duration_seconds` - Alert processing time
- `cano_alerts_errors_total` - Alert processing errors

### Destination Metrics

- `cano_destination_messages_sent_total` - Messages sent to destinations
- `cano_destination_send_duration_seconds` - Time to send to destinations
- `cano_destination_errors_total` - Destination sending errors

### Routing Metrics

- `cano_routing_decisions_total` - Routing decisions
- `cano_teams_matched_total` - Team matches

### Go Runtime Metrics

- `go_goroutines` - Number of goroutines
- `go_memstats_*` - Go memory statistics
- `go_gc_duration_seconds` - GC duration

## 🎨 Customization

The dashboards can be customized:

1. Change time intervals for individual panels
2. Add alerts for critical metrics
3. Adjust colors and thresholds
4. Add additional panels for specific use cases

## 🔍 Troubleshooting

### No data on dashboard

1. Check if Prometheus is scraping metrics from the application
2. Verify that the application exports metrics on `/metrics`
3. Ensure Prometheus is properly configured as a data source

### Incorrect metric values

1. Check application logs for metric registration errors
2. Verify that the application actually performs operations that should increment metrics
3. Check if labels in PromQL queries are correct

## 📝 Example PromQL Queries

```promql
# Rate of HTTP requests by endpoint
rate(http_requests_total[5m])

# 95th percentile of alert processing time
histogram_quantile(0.95, rate(cano_alerts_processing_duration_seconds_bucket[5m]))

# Total error rate across all components
rate(cano_alerts_errors_total[5m]) + rate(cano_destination_errors_total[5m])

# Memory usage trend
go_memstats_alloc_bytes

# SLA calculation (success rate)
(
  sum(rate(cano_alerts_processed_total[24h])) - 
  sum(rate(cano_alerts_errors_total[24h]))
) / sum(rate(cano_alerts_processed_total[24h])) * 100

# Error budget burn rate
(
  sum(rate(cano_alerts_errors_total[1h])) + 
  sum(rate(cano_destination_errors_total[1h]))
) / sum(rate(cano_alerts_processed_total[1h])) * 100
```
