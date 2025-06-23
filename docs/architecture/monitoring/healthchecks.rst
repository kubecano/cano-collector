Health Checks
=============

This document describes the health check architecture in cano-collector, including application health checks and destination health monitoring that doesn't affect Kubernetes readiness/liveness probes.

Current Health Check Implementation
-----------------------------------

Cano-collector currently implements basic health checks with the following endpoints:

Health Check Endpoints
~~~~~~~~~~~~~~~~~~~~~~

.. list-table::
   :header-rows: 1

   * - Endpoint
     - Method
     - Description
     - Kubernetes Probe Type
   * - /livez
     - GET
     - Liveness check - application is running
     - Liveness Probe
   * - /readyz
     - GET
     - Readiness check - application is ready to serve
     - Readiness Probe
   * - /healthz
     - GET
     - General health status
     - N/A

Current Implementation
~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    type HealthChecker struct {
        logger logger.LoggerInterface
        checks map[string]health.Check
    }

    func (hc *HealthChecker) RegisterHealthChecks() error {
        // Register basic application health checks
        hc.checks["application"] = health.Check{
            Name:     "application",
            Timeout:  time.Second * 5,
            SkipOnErr: false,
            Check:    hc.checkApplicationHealth,
        }
        
        return nil
    }

    func (hc *HealthChecker) checkApplicationHealth(ctx context.Context) error {
        // Basic application health check
        // Should always pass if the application is running
        return nil
    }

    func (hc *HealthChecker) Handler() http.Handler {
        return health.Handler()
    }

Planned Health Check Enhancements
---------------------------------

The following health check enhancements should be implemented to provide comprehensive health monitoring:

Application Health Checks
~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    type ApplicationHealthChecker struct {
        logger    logger.LoggerInterface
        metrics   metric.MetricsInterface
        config    config.Config
        checks    map[string]health.Check
    }

    func (ahc *ApplicationHealthChecker) RegisterHealthChecks() error {
        // Core application checks (affect Kubernetes probes)
        ahc.checks["application"] = health.Check{
            Name:      "application",
            Timeout:   time.Second * 5,
            SkipOnErr: false,
            Check:     ahc.checkApplicationHealth,
        }
        
        ahc.checks["configuration"] = health.Check{
            Name:      "configuration",
            Timeout:   time.Second * 10,
            SkipOnErr: false,
            Check:     ahc.checkConfigurationHealth,
        }
        
        ahc.checks["kubernetes_client"] = health.Check{
            Name:      "kubernetes_client",
            Timeout:   time.Second * 15,
            SkipOnErr: false,
            Check:     ahc.checkKubernetesClientHealth,
        }
        
        // Memory and resource checks
        ahc.checks["memory_usage"] = health.Check{
            Name:      "memory_usage",
            Timeout:   time.Second * 5,
            SkipOnErr: true, // Don't fail probes for memory warnings
            Check:     ahc.checkMemoryUsage,
        }
        
        ahc.checks["goroutine_count"] = health.Check{
            Name:      "goroutine_count",
            Timeout:   time.Second * 5,
            SkipOnErr: true, // Don't fail probes for goroutine warnings
            Check:     ahc.checkGoroutineCount,
        }
        
        return nil
    }

    func (ahc *ApplicationHealthChecker) checkApplicationHealth(ctx context.Context) error {
        // Basic application health - should always pass if running
        return nil
    }

    func (ahc *ApplicationHealthChecker) checkConfigurationHealth(ctx context.Context) error {
        // Verify configuration is valid and loaded
        if ahc.config.AppName == "" {
            return errors.New("application name not configured")
        }
        
        // Check if required configuration files exist
        if err := ahc.verifyConfigFiles(); err != nil {
            return fmt.Errorf("configuration files error: %w", err)
        }
        
        return nil
    }

    func (ahc *ApplicationHealthChecker) checkKubernetesClientHealth(ctx context.Context) error {
        // Test Kubernetes API connectivity
        client, err := kubernetes.NewForConfig(ahc.config.KubeConfig)
        if err != nil {
            return fmt.Errorf("kubernetes client creation failed: %w", err)
        }
        
        // Simple API call to verify connectivity
        _, err = client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{Limit: 1})
        if err != nil {
            return fmt.Errorf("kubernetes API call failed: %w", err)
        }
        
        return nil
    }

    func (ahc *ApplicationHealthChecker) checkMemoryUsage(ctx context.Context) error {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        // Warn if memory usage is high (but don't fail health check)
        memoryUsageMB := m.Alloc / 1024 / 1024
        if memoryUsageMB > 1000 { // 1GB threshold
            ahc.logger.Warnf("High memory usage: %d MB", memoryUsageMB)
        }
        
        return nil
    }

    func (ahc *ApplicationHealthChecker) checkGoroutineCount(ctx context.Context) error {
        goroutineCount := runtime.NumGoroutine()
        
        // Warn if goroutine count is high (but don't fail health check)
        if goroutineCount > 1000 {
            ahc.logger.Warnf("High goroutine count: %d", goroutineCount)
        }
        
        return nil
    }

Destination Health Monitoring
-----------------------------

Destination health checks should be implemented separately from Kubernetes probes to avoid affecting application availability:

Destination Health Checker
~~~~~~~~~~~~~~~~~~~~~~~~~~

.. code-block:: go

    type DestinationHealthChecker struct {
        logger       logger.LoggerInterface
        metrics      metric.MetricsInterface
        destinations map[string]Destination
        healthStatus map[string]DestinationHealth
        mutex        sync.RWMutex
    }

    type DestinationHealth struct {
        Name         string
        Type         string
        Status       HealthStatus
        LastCheck    time.Time
        LastError    string
        ResponseTime time.Duration
        RetryCount   int
    }

    type HealthStatus string

    const (
        HealthStatusHealthy   HealthStatus = "healthy"
        HealthStatusDegraded  HealthStatus = "degraded"
        HealthStatusUnhealthy HealthStatus = "unhealthy"
        HealthStatusUnknown   HealthStatus = "unknown"
    )

    func (dhc *DestinationHealthChecker) StartHealthMonitoring() {
        // Start background health monitoring
        go dhc.monitorDestinations()
    }

    func (dhc *DestinationHealthChecker) monitorDestinations() {
        ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
        defer ticker.Stop()

        for range ticker.C {
            dhc.checkAllDestinations()
        }
    }

    func (dhc *DestinationHealthChecker) checkAllDestinations() {
        dhc.mutex.Lock()
        defer dhc.mutex.Unlock()

        for name, destination := range dhc.destinations {
            health := dhc.checkDestinationHealth(destination)
            dhc.healthStatus[name] = health
            
            // Update metrics
            dhc.updateDestinationHealthMetrics(health)
            
            // Log status changes
            if dhc.isStatusChanged(name, health) {
                dhc.logger.Infof("Destination %s health status changed to %s", name, health.Status)
            }
        }
    }

    func (dhc *DestinationHealthChecker) checkDestinationHealth(destination Destination) DestinationHealth {
        start := time.Now()
        
        health := DestinationHealth{
            Name:      destination.Name,
            Type:      destination.Type,
            LastCheck: time.Now(),
        }

        // Perform health check based on destination type
        switch destination.Type {
        case "slack":
            health = dhc.checkSlackHealth(destination)
        case "msteams":
            health = dhc.checkMSTeamsHealth(destination)
        case "webhook":
            health = dhc.checkWebhookHealth(destination)
        default:
            health.Status = HealthStatusUnknown
            health.LastError = "unknown destination type"
        }

        health.ResponseTime = time.Since(start)
        return health
    }

    func (dhc *DestinationHealthChecker) checkSlackHealth(destination Destination) DestinationHealth {
        health := DestinationHealth{
            Name:      destination.Name,
            Type:      destination.Type,
            LastCheck: time.Now(),
        }

        // Test Slack webhook with a simple message
        testMessage := map[string]string{
            "text": "Health check - " + time.Now().Format(time.RFC3339),
        }

        resp, err := http.PostJSON(destination.WebhookURL, testMessage)
        if err != nil {
            health.Status = HealthStatusUnhealthy
            health.LastError = err.Error()
            return health
        }
        defer resp.Body.Close()

        if resp.StatusCode == 200 {
            health.Status = HealthStatusHealthy
        } else {
            health.Status = HealthStatusDegraded
            health.LastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
        }

        return health
    }

    func (dhc *DestinationHealthChecker) checkMSTeamsHealth(destination Destination) DestinationHealth {
        health := DestinationHealth{
            Name:      destination.Name,
            Type:      destination.Type,
            LastCheck: time.Now(),
        }

        // Test MS Teams webhook with a simple card
        testCard := map[string]interface{}{
            "@type": "MessageCard",
            "text":  "Health check - " + time.Now().Format(time.RFC3339),
        }

        resp, err := http.PostJSON(destination.WebhookURL, testCard)
        if err != nil {
            health.Status = HealthStatusUnhealthy
            health.LastError = err.Error()
            return health
        }
        defer resp.Body.Close()

        if resp.StatusCode == 200 {
            health.Status = HealthStatusHealthy
        } else {
            health.Status = HealthStatusDegraded
            health.LastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
        }

        return health
    }

    func (dhc *DestinationHealthChecker) checkWebhookHealth(destination Destination) DestinationHealth {
        health := DestinationHealth{
            Name:      destination.Name,
            Type:      destination.Type,
            LastCheck: time.Now(),
        }

        // Test webhook with a simple GET request
        resp, err := http.Get(destination.WebhookURL)
        if err != nil {
            health.Status = HealthStatusUnhealthy
            health.LastError = err.Error()
            return health
        }
        defer resp.Body.Close()

        if resp.StatusCode >= 200 && resp.StatusCode < 300 {
            health.Status = HealthStatusHealthy
        } else {
            health.Status = HealthStatusDegraded
            health.LastError = fmt.Sprintf("HTTP %d", resp.StatusCode)
        }

        return health
    }

    func (dhc *DestinationHealthChecker) isStatusChanged(name string, health DestinationHealth) bool {
        if existing, exists := dhc.healthStatus[name]; exists {
            return existing.Status != health.Status
        }
        return true // First time seeing this destination
    }

    func (dhc *DestinationHealthChecker) GetHealthStatus() map[string]DestinationHealth {
        dhc.mutex.RLock()
        defer dhc.mutex.RUnlock()
        
        result := make(map[string]DestinationHealth)
        for k, v := range dhc.healthStatus {
            result[k] = v
        }
        return result
    }

    func (dhc *DestinationHealthChecker) HealthStatusHandler(c *gin.Context) {
        dhc.mutex.RLock()
        defer dhc.mutex.RUnlock()
        
        status := map[string]interface{}{
            "timestamp": time.Now(),
            "destinations": dhc.healthStatus,
            "summary": dhc.calculateHealthSummary(),
        }
        
        c.JSON(http.StatusOK, status)
    }

    func (dhc *DestinationHealthChecker) calculateHealthSummary() map[string]interface{} {
        healthy := 0
        degraded := 0
        unhealthy := 0
        unknown := 0
        
        for _, health := range dhc.healthStatus {
            switch health.Status {
            case HealthStatusHealthy:
                healthy++
            case HealthStatusDegraded:
                degraded++
            case HealthStatusUnhealthy:
                unhealthy++
            case HealthStatusUnknown:
                unknown++
            }
        }
        
        total := len(dhc.healthStatus)
        
        return map[string]interface{}{
            "total":     total,
            "healthy":   healthy,
            "degraded":  degraded,
            "unhealthy": unhealthy,
            "unknown":   unknown,
        }
    }

Metrics Integration
~~~~~~~~~~~~~~~~~~~

Update metrics based on destination health:

.. code-block:: go

    func (dhc *DestinationHealthChecker) updateDestinationHealthMetrics(health DestinationHealth) {
        status := 0
        switch health.Status {
        case HealthStatusHealthy:
            status = 1
        case HealthStatusDegraded:
            status = 0.5
        case HealthStatusUnhealthy:
            status = 0
        case HealthStatusUnknown:
            status = 0
        }
        
        dhc.metrics.SetDestinationHealthStatus(health.Name, health.Type, status)
    }

Configuration
-------------

Health check configuration:

.. code-block:: yaml

    healthChecks:
      # Application health checks (affect Kubernetes probes)
      application:
        enabled: true
        timeout: 5s
        skipOnError: false
        
      configuration:
        enabled: true
        timeout: 10s
        skipOnError: false
        
      kubernetesClient:
        enabled: true
        timeout: 15s
        skipOnError: false
        
      # Resource monitoring (don't affect probes)
      memoryUsage:
        enabled: true
        timeout: 5s
        skipOnError: true
        warningThreshold: 1000MB
        
      goroutineCount:
        enabled: true
        timeout: 5s
        skipOnError: true
        warningThreshold: 1000
        
      # Destination health monitoring (separate from probes)
      destinations:
        enabled: true
        checkInterval: 30s
        timeout: 10s
        retryCount: 3
        
        # Health check endpoints
        slack:
          testMessage: "Health check - {{timestamp}}"
          
        msteams:
          testCard: true
          
        webhook:
          method: "GET"
          expectedStatus: [200, 201, 202]

Kubernetes Probe Configuration
------------------------------

Kubernetes probe configuration that separates application health from destination health:

.. code-block:: yaml

    # values.yaml
    collector:
      livenessProbe:
        httpGet:
          path: /livez
          port: 8080
        initialDelaySeconds: 30
        periodSeconds: 10
        timeoutSeconds: 5
        failureThreshold: 3
        
      readinessProbe:
        httpGet:
          path: /readyz
          port: 8080
        initialDelaySeconds: 5
        periodSeconds: 5
        timeoutSeconds: 3
        failureThreshold: 3

    # Separate endpoint for destination health (doesn't affect probes)
    service:
      ports:
        - name: http
          port: 8080
          targetPort: 8080
        - name: health
          port: 8081
          targetPort: 8081

Key Principles
--------------

1. **Separation of Concerns**:
   - Application health checks affect Kubernetes probes
   - Destination health checks are separate and don't cause restarts
   - Resource monitoring provides warnings but doesn't fail probes

2. **Probe Independence**:
   - `/livez` and `/readyz` only check core application health
   - Destination failures don't cause pod restarts
   - External service dependencies are isolated

3. **Graceful Degradation**:
   - Application continues running even if destinations are unhealthy
   - Failed destinations are logged and monitored
   - Retry mechanisms handle temporary failures

4. **Comprehensive Monitoring**:
   - All health aspects are monitored and reported
   - Metrics provide operational visibility
   - Health status is available via API endpoints

This approach ensures that cano-collector remains stable and available even when external notification services are experiencing issues.
