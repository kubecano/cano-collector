# Cano-Collector Test Pod Suite

This test suite provides comprehensive validation scenarios for cano-collector functionality. Each test triggers specific Kubernetes events that should be monitored and processed by cano-collector.

## Quick Start

```bash
# 1. Validate cano-collector is running
./scripts/validate-collector.sh

# 2. Run a simple crash test
./scripts/run-test.sh crash-loop busybox-crash

# 3. Monitor the events
kubectl get events -n test-pods --sort-by=.metadata.creationTimestamp -w

# 4. Clean up when done
./scripts/cleanup.sh
```

## Test Categories

### 🚨 Pod Crash Scenarios (`pods/crash-loop/`)

Tests that create CrashLoopBackOff conditions:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `busybox-crash` | Simple container that exits after 1 second | CrashLoopBackOff, pod restart events |
| `jdk-crash` | Java app that throws RuntimeException | Java crash events, application error logs |
| `nginx-crash` | Nginx with invalid configuration | Configuration error events, service startup failures |

```bash
./scripts/run-test.sh crash-loop busybox-crash
./scripts/run-test.sh crash-loop jdk-crash
./scripts/run-test.sh crash-loop nginx-crash
```

### 💾 Memory Issues (`pods/oom/`)

Tests that trigger Out of Memory conditions:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `memory-bomb` | Fast memory allocation (200MB in 128MB limit) | OOMKilled events, memory limit exceeded |
| `gradual-oom` | Slow memory leak simulation | Gradual memory increase, eventual OOM |

```bash
./scripts/run-test.sh oom memory-bomb
./scripts/run-test.sh oom gradual-oom
```

### 📦 Image Pull Failures (`pods/image-pull/`)

Tests that create ImagePullBackOff conditions:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `nonexistent-image` | Tries to pull non-existent image | ImagePullBackOff, image not found errors |
| `private-registry` | Tries to pull from private registry without credentials | Authentication failures, access denied |

```bash
./scripts/run-test.sh image-pull nonexistent-image
./scripts/run-test.sh image-pull private-registry
```

### ⚡ Resource Constraints (`pods/resource-limits/`)

Tests that create resource scheduling issues:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `cpu-starved` | Requests more CPU than available | Pod pending, insufficient CPU |
| `impossible-resources` | Requests impossible amounts (500 CPU, 500GB RAM) | Permanent pending state, unschedulable |

```bash
./scripts/run-test.sh resource-limits cpu-starved
./scripts/run-test.sh resource-limits impossible-resources
```

### 🌐 Network Failures (`pods/network/`)

Tests that simulate network connectivity issues:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `dns-failure` | Tries to resolve non-existent DNS names | DNS resolution failures, service unreachable |
| `service-unreachable` | Tries to connect to non-existent services | Connection timeouts, network errors |

```bash
./scripts/run-test.sh network dns-failure
./scripts/run-test.sh network service-unreachable
```

### 🚀 Deployment Issues (`deployments/`)

Tests that create deployment-level failures:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `replica-failure` | Deployment with non-existent image | ReplicaFailure, deployment stuck |
| `rollout-failure` | Invalid rolling update configuration | ProgressDeadlineExceeded, rollout stuck |
| `scaling-test` | Manual scaling test (instructions included) | ScalingUp/Down events |

```bash
./scripts/run-test.sh deployments replica-failure
./scripts/run-test.sh deployments rollout-failure
./scripts/run-test.sh deployments scaling-test

# For scaling test, manually scale:
kubectl scale deployment scaling-test --replicas=5 -n test-pods
kubectl scale deployment scaling-test --replicas=1 -n test-pods
```

### ⚙️ Job Failures (`jobs/`)

Tests that create job execution failures:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `job-failure` | Job that always fails (exits with code 1) | JobFailed, BackoffLimitExceeded |
| `job-timeout` | Job that exceeds activeDeadlineSeconds | JobTimeout, DeadlineExceeded |
| `cronjob-failure` | CronJob that fails every 2 minutes | Recurring job failures |

```bash
./scripts/run-test.sh jobs job-failure
./scripts/run-test.sh jobs job-timeout
./scripts/run-test.sh jobs cronjob-failure
```

### 🔗 Service Issues (`services/`)

Tests that create service connectivity problems:

| Test | Description | Expected Events |
|------|-------------|----------------|
| `no-endpoints` | Service with no matching pods | Service without endpoints, no backend available |
| `loadbalancer-pending` | LoadBalancer service pending external IP | LoadBalancer provisioning issues |

```bash
./scripts/run-test.sh services no-endpoints
./scripts/run-test.sh services loadbalancer-pending
```

## Expected Cano-Collector Events

Each test scenario should trigger specific events that cano-collector monitors:

### Pod Events
- **CrashLoopBackOff**: `pod_crash_looping`, `container_restart`
- **ImagePullBackOff**: `image_pull_failed`, `pod_pending`
- **OOMKilled**: `pod_oom_killed`, `container_memory_exceeded`
- **Pending**: `pod_pending`, `insufficient_resources`

### Deployment Events  
- **ReplicaFailure**: `deployment_replica_failure`
- **ProgressDeadlineExceeded**: `deployment_progress_failed`
- **Scaling**: `deployment_scaling_up`, `deployment_scaling_down`

### Job Events
- **JobFailed**: `job_failed`, `pod_failed`
- **JobTimeout**: `job_timeout`, `deadline_exceeded`

### Service Events
- **NoEndpoints**: `service_no_endpoints`, `endpoint_missing`
- **LoadBalancerPending**: `loadbalancer_pending`

## Monitoring Commands

### Real-time Event Monitoring
```bash
# Watch events in test namespace
kubectl get events -n test-pods --sort-by=.metadata.creationTimestamp -w

# Monitor cano-collector logs
kubectl logs -n monitoring -l app=cano-collector -f

# Watch pod status changes
kubectl get pods -n test-pods -w
```

### Check Specific Resource Types
```bash
# Deployment status
kubectl get deployments -n test-pods -w

# Job status  
kubectl get jobs -n test-pods -w

# Service and endpoints
kubectl get svc,endpoints -n test-pods
```

### Cano-Collector Validation
```bash
# Check if cano-collector is processing events
./scripts/validate-collector.sh

# Check metrics
kubectl port-forward svc/cano-collector 8080:8080 -n monitoring &
curl http://localhost:8080/metrics | grep event_processed_total
```

## Cleanup

### Clean Specific Namespace
```bash
./scripts/cleanup.sh test-pods
```

### Clean All Test Resources  
```bash
./scripts/cleanup.sh --all
```

### Force Cleanup (No Confirmation)
```bash
./scripts/cleanup.sh test-pods --force
```

## Troubleshooting

### Test Not Triggering Expected Events

1. **Check cano-collector status**:
   ```bash
   ./scripts/validate-collector.sh
   ```

2. **Verify RBAC permissions**:
   ```bash
   kubectl auth can-i watch events --as=system:serviceaccount:monitoring:cano-collector
   ```

3. **Check cano-collector configuration**:
   ```bash
   kubectl get configmap -n monitoring
   kubectl describe configmap cano-collector-config -n monitoring
   ```

### Pod Stuck in Pending State

1. **Check node resources**:
   ```bash
   kubectl describe nodes
   kubectl top nodes
   ```

2. **Check pod events**:
   ```bash
   kubectl describe pod <pod-name> -n test-pods
   ```

### ImagePullBackOff Not Occurring

Some Kubernetes clusters cache image pull failures. To force re-pull:
```bash
kubectl delete pod <pod-name> -n test-pods
# Wait for recreation
```

### LoadBalancer Test Not Working

LoadBalancer behavior depends on cluster type:
- **Minikube**: Use `minikube tunnel` for LoadBalancer support
- **Kind**: LoadBalancer will remain pending (expected)
- **Cloud**: Should provision external IP (unless quotas exceeded)

## Advanced Usage

### Custom Namespace
```bash
./scripts/run-test.sh crash-loop busybox-crash my-custom-namespace
```

### Run Multiple Tests
```bash
# Run several tests in parallel
./scripts/run-test.sh crash-loop busybox-crash ns1 &
./scripts/run-test.sh oom memory-bomb ns2 &
./scripts/run-test.sh image-pull nonexistent-image ns3 &
wait
```

### Integration with CI/CD
```bash
#!/bin/bash
# Example CI test script

# Validate cano-collector is ready
./scripts/validate-collector.sh || exit 1

# Run core test scenarios
./scripts/run-test.sh crash-loop busybox-crash ci-test
sleep 30  # Wait for events to be processed

# Check if events were processed
if kubectl logs -n monitoring -l app=cano-collector --tail=50 | grep -q "busybox-crash-test"; then
  echo "✅ Test events processed successfully"
else
  echo "❌ Test events not found in logs"
  exit 1
fi

# Cleanup
./scripts/cleanup.sh ci-test --force
```

## Development

### Adding New Test Scenarios

1. Create YAML file in appropriate category directory
2. Add test description to this README
3. Include expected events and validation steps
4. Test the scenario manually
5. Update `run-test.sh` if new category is added

### Test File Requirements

- Include `test-type` label for cleanup identification
- Include `scenario` label for specific test identification  
- Add resource limits to prevent resource exhaustion
- Include comments explaining expected behavior
- Use ConfigMaps for complex configurations (Java code, configs)

## Integration with Alertmanager

If Alertmanager is configured to receive cano-collector events:

```bash
# Check Alertmanager status
kubectl get pod -n monitoring -l app=alertmanager

# Check active alerts
kubectl port-forward svc/alertmanager 9093:9093 -n monitoring &
curl http://localhost:9093/api/v1/alerts
```

## Validation Checklist

For each test scenario, verify:

- [ ] **Pod reaches expected state** (CrashLoopBackOff, Pending, etc.)
- [ ] **Kubernetes events generated** (`kubectl get events`)
- [ ] **Cano-collector processes events** (check logs/metrics)
- [ ] **Alerts triggered** (if Alertmanager configured)
- [ ] **Cleanup successful** (no leftover resources)

This test suite provides comprehensive validation of cano-collector's ability to monitor and respond to various Kubernetes failure scenarios.