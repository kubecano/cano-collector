Alert Processing Architecture
=============================

This document describes the alert processing architecture in cano-collector, including the PrometheusAlert structure, parsing, deduplication, relabeling, and asynchronous processing queue.

PrometheusAlert Structure
-------------------------

The PrometheusAlert structure in cano-collector follows the standard Alertmanager format:

.. code-block:: go

    type PrometheusAlert struct {
        EndsAt       time.Time            `json:"endsAt"`
        GeneratorURL string               `json:"generatorURL"`
        StartsAt     time.Time            `json:"startsAt"`
        Fingerprint  string               `json:"fingerprint"`
        Status       string               `json:"status"`           // "firing" or "resolved"
        Labels       map[string]string    `json:"labels"`
        Annotations  map[string]string    `json:"annotations"`
    }

    type AlertManagerEvent struct {
        Alerts             []PrometheusAlert `json:"alerts"`
        ExternalURL        string            `json:"externalURL"`
        GroupKey           string            `json:"groupKey"`
        Version            string            `json:"version"`
        CommonAnnotations  map[string]string `json:"commonAnnotations,omitempty"`
        CommonLabels       map[string]string `json:"commonLabels,omitempty"`
        GroupLabels        map[string]string `json:"groupLabels,omitempty"`
        Receiver           string            `json:"receiver"`
        Status             string            `json:"status"`
    }

Key fields and their purposes:

- **Fingerprint**: Unique identifier for the alert, used for deduplication
- **Status**: Current state ("firing" or "resolved")
- **Labels**: Key-value pairs for routing and filtering
- **Annotations**: Human-readable descriptions and metadata
- **StartsAt/EndsAt**: Timestamps for alert lifecycle
- **GeneratorURL**: Link to the alerting rule or source

This structure ensures compatibility with Alertmanager webhook format and provides all necessary information for alert processing.

Alert Parsing from template.Data
--------------------------------

Alerts are received from Alertmanager in the standard webhook format and parsed into the internal PrometheusAlert structure:

.. code-block:: go

    func parseAlertFromTemplateData(data template.Data) (*PrometheusAlert, error) {
        if len(data.Alerts) == 0 {
            return nil, errors.New("no alerts in template data")
        }
        
        alert := data.Alerts[0] // Process first alert for now
        
        prometheusAlert := &PrometheusAlert{
            EndsAt:       alert.EndsAt,
            GeneratorURL: alert.GeneratorURL,
            StartsAt:     alert.StartsAt,
            Fingerprint:  alert.Fingerprint,
            Status:       alert.Status,
            Labels:       alert.Labels,
            Annotations:  alert.Annotations,
        }
        
        return prometheusAlert, nil
    }

The parsing process includes:

1. **Validation**: Ensure required fields are present
2. **Type Conversion**: Convert template.Data to internal PrometheusAlert
3. **Normalization**: Standardize field formats and values
4. **Enrichment**: Add metadata like cluster information

Alert Deduplication
-------------------

Deduplication prevents processing the same alert multiple times:

.. code-block:: go

    type DeduplicationCache struct {
        cache map[string]time.Time
        mutex sync.RWMutex
        ttl   time.Duration
    }

    func (dc *DeduplicationCache) IsDuplicate(alert *PrometheusAlert) bool {
        hash := dc.generateCompoundHash(alert)
        
        dc.mutex.Lock()
        defer dc.mutex.Unlock()
        
        if lastSeen, exists := dc.cache[hash]; exists {
            if time.Since(lastSeen) < dc.ttl {
                return true
            }
        }
        
        dc.cache[hash] = time.Now()
        return false
    }

    func (dc *DeduplicationCache) generateCompoundHash(alert *PrometheusAlert) string {
        data := []byte{
            alert.Fingerprint,
            alert.Status,
            strconv.FormatInt(alert.StartsAt.Unix(), 10),
            strconv.FormatInt(alert.EndsAt.Unix(), 10),
        }
        
        hash := sha256.Sum256(data)
        return hex.EncodeToString(hash[:])
    }

The deduplication strategy:

1. **Compound Hash**: Combines fingerprint, status, and timestamps
2. **TTL-based Cache**: Prevents reprocessing within a configurable time window
3. **Thread-safe**: Concurrent access protection
4. **Memory Management**: Automatic cleanup of expired entries

Alert Relabeling
----------------

Relabeling allows mapping custom alert labels to standard cano-collector expectations:

.. code-block:: go

    type RelabelRule struct {
        Source    string `yaml:"source"`
        Target    string `yaml:"target"`
        Operation string `yaml:"operation"` // "add" or "replace"
    }

    func relabelAlert(alert *PrometheusAlert, rules []RelabelRule) *PrometheusAlert {
        for _, rule := range rules {
            if sourceValue, exists := alert.Labels[rule.Source]; exists {
                alert.Labels[rule.Target] = sourceValue
                
                if rule.Operation == "replace" {
                    delete(alert.Labels, rule.Source)
                }
            }
        }
        return alert
    }

Common relabeling scenarios:

- **Pod Mapping**: Map custom pod labels to standard `pod` label
- **Namespace Mapping**: Map custom namespace labels to standard `namespace` label
- **Severity Mapping**: Map custom severity levels to standard levels
- **Resource Type Mapping**: Map custom resource type labels to standard types

Example configuration:

.. code-block:: yaml

    alertRelabel:
      - source: "pod_name"
        target: "pod"
        operation: "add"
      - source: "deployment_name"
        target: "deployment"
        operation: "replace"
      - source: "custom_severity"
        target: "severity"
        operation: "add"

Asynchronous Processing Queue
----------------------------

The asynchronous processing queue ensures non-blocking alert reception and reliable processing:

.. code-block:: go

    type AlertQueue struct {
        queue    chan *AlertTask
        workers  int
        logger   logger.LoggerInterface
        metrics  metric.MetricsInterface
    }

    type AlertTask struct {
        Alert     *PrometheusAlert
        Timestamp time.Time
        Attempts  int
    }

    func (aq *AlertQueue) Start() {
        for i := 0; i < aq.workers; i++ {
            go aq.worker()
        }
    }

    func (aq *AlertQueue) worker() {
        for task := range aq.queue {
            start := time.Now()
            
            if err := aq.processAlert(task); err != nil {
                aq.logger.Errorf("Failed to process alert: %v", err)
                aq.metrics.IncAlertProcessingErrors()
                
                // Retry logic for failed alerts
                if task.Attempts < maxRetries {
                    task.Attempts++
                    aq.queue <- task
                }
            } else {
                aq.metrics.ObserveAlertProcessingDuration(time.Since(start))
            }
        }
    }

    func (aq *AlertQueue) Enqueue(alert *PrometheusAlert) {
        task := &AlertTask{
            Alert:     alert,
            Timestamp: time.Now(),
            Attempts:  0,
        }
        
        select {
        case aq.queue <- task:
            aq.metrics.IncAlertQueued()
        default:
            aq.metrics.IncAlertDropped()
            aq.logger.Warn("Alert queue full, dropping alert")
        }
    }

Queue characteristics:

1. **Buffered Channel**: Configurable queue size to handle burst traffic
2. **Multiple Workers**: Parallel processing for improved throughput
3. **Retry Logic**: Automatic retry for failed processing
4. **Metrics**: Comprehensive monitoring of queue performance
5. **Backpressure**: Graceful handling of queue overflow

Processing Flow
---------------

The complete alert processing flow:

1. **Reception**: Alert received via `/api/alerts` endpoint
2. **Parsing**: Convert template.Data to PrometheusAlert
3. **Deduplication**: Check if alert was recently processed
4. **Relabeling**: Apply custom label mappings
5. **Enqueue**: Add to asynchronous processing queue
6. **Processing**: Worker processes alert in background
7. **Enrichment**: Add context and create Issue object
8. **Routing**: Determine target destinations
9. **Delivery**: Send to configured destinations

This architecture ensures:

- **Reliability**: No alert loss through queuing and retries
- **Performance**: Non-blocking reception and parallel processing
- **Scalability**: Configurable worker count and queue size
- **Observability**: Comprehensive metrics and logging
- **Flexibility**: Customizable relabeling and processing rules 