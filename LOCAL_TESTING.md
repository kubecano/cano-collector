# Local Testing Guide

Quick reference for testing cano-collector locally on macOS using k3d and Makefile.

## Prerequisites

```bash
# Install k3d (lightweight k3s in Docker)
brew install k3d

# Verify Docker Desktop is running
docker ps

# View all available local development commands
make help
```

## Quick Start

### 1. Configure Slack (First Time Only)

Edit `values-local.yaml`:

```yaml
destinations:
  slack:
    - name: "dev-local-slack"
      api_key: "xoxb-YOUR-DEV-TOKEN-HERE"  # ← Replace with your dev token
      slack_channel: "#your-dev-channel"   # ← Replace with your dev channel
```

**Get Slack Bot Token:**
1. Go to https://api.slack.com/apps
2. Select your app (or create new for dev)
3. OAuth & Permissions → Bot User OAuth Token (`xoxb-...`)
4. Ensure bot has scopes: `chat:write`, `files:write`, `chat:write.public`
5. Add bot to your dev Slack channel

### 2. Deploy to Local k3d Cluster

```bash
# Full setup: cluster + tests + build + deploy
make local-dev

# Quick iteration: rebuild + redeploy (skip tests)
make local-dev-quick

# Config-only update: no rebuild, no tests
make local-dev-config
```

**What `make local-dev` does:**
1. ✅ Checks/creates k3d cluster `cano-dev` with local registry on port 5001
2. ✅ Runs Go tests
3. ✅ Builds Docker image
4. ✅ Pushes to local registry
5. ✅ Deploys via Helm with `values-local.yaml`
6. ✅ Shows deployment status

**Note:** Port 5001 is used instead of 5000 to avoid conflict with macOS AirPlay Receiver.

### 3. Monitor Deployment

```bash
# View logs (live tail)
make local-logs

# Check pod status
make local-status

# Port forward for local API access
make local-port-forward

# Or use kubectl directly:
kubectl logs -n kubecano -l app=cano-collector -f
kubectl get pods -n kubecano
kubectl describe pod -n kubecano -l app=cano-collector
```

## Testing Alerts

### Test Webhook Endpoint

**Using Makefile (recommended)**:

```bash
# In terminal 1: Start port-forward
make local-port-forward

# In terminal 2: Send test alert
make local-test-alert
```

**Or manually with curl**:

```bash
# Basic health check
curl http://localhost:8080/api/alerts \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '[]'

# Test with realistic alert (crash looping pod)
curl http://localhost:8080/api/alerts \
  -X POST \
  -H 'Content-Type: application/json' \
  -d '[{
    "status": "firing",
    "labels": {
      "alertname": "KubePodCrashLooping",
      "pod": "test-pod",
      "namespace": "default",
      "severity": "critical"
    },
    "annotations": {
      "summary": "Pod test-pod is crash looping",
      "description": "Test alert for local development"
    }
  }]'
```

**Expected Result:**
- Check logs: `make local-logs` or `kubectl logs -n kubecano -l app=cano-collector | grep -i "alert\|crash"`
- Check Slack channel for notification
- For `KubePodCrashLooping`, should include pod logs attachment

### Test with Real Crash Loop

```bash
# Deploy crash loop test pod
./tests/scripts/run-test.sh crash-loop busybox-crash

# Monitor cano-collector processing
kubectl logs -n kubecano -l app=cano-collector -f | grep -i "alert\|crash"

# Check if kube-state-metrics sees the test pod
kubectl get pod -n test-pods busybox-crash-test

# Cleanup test pods
./tests/scripts/cleanup.sh
```

**Note:** For real alerts, you need Prometheus + Alertmanager running in the cluster. The test pods trigger alerts via kube-state-metrics metrics.

## Development Workflow

```bash
# 1. Make code changes in your IDE

# 2. Quick rebuild + redeploy
make local-dev-quick

# 3. Watch logs
make local-logs

# 4. Test with alert
make local-port-forward  # Terminal 1
make local-test-alert    # Terminal 2

# 5. Verify in Slack channel
```

### Makefile Commands Reference

| Command | Description | When to Use |
|---------|-------------|-------------|
| `make local-dev` | Full setup (tests + build + deploy) | First time, major changes |
| `make local-dev-quick` | Rebuild + redeploy (skip tests) | Code changes |
| `make local-dev-config` | Config-only update (no rebuild) | values-local.yaml changes |
| `make local-logs` | Tail logs | Debugging |
| `make local-status` | Show pod status | Check health |
| `make local-port-forward` | Port forward to localhost:8080 | Testing alerts |
| `make local-test-alert` | Send test alert | Quick validation |
| `make local-clean` | Remove deployment | Reset without cluster delete |
| `make local-clean-all` | Full cleanup | Start fresh |

## Cleanup

```bash
# Remove deployment (keep cluster)
make local-clean

# Remove entire cluster (fresh start)
make local-clean-all

# Remove test pods only
./tests/scripts/cleanup.sh

# Or manually:
helm uninstall cano-collector -n kubecano
k3d cluster delete cano-dev
```

## Troubleshooting

### Cluster Won't Start

```bash
# Check Docker Desktop is running
docker ps

# Delete and recreate cluster
make local-delete-cluster
make local-create-cluster

# Or full redeploy
make local-clean-all
make local-dev
```

### Port 5001 Already in Use

```bash
# Check what's using the port
lsof -i :5001

# Stop existing registry
docker stop cano-registry
docker rm cano-registry

# Or use different port (edit Makefile:LOCAL_REGISTRY_PORT and values-local.yaml:registry)
```

### Image Pull Failures (ImagePullBackOff)

```bash
# Check image exists in local registry
curl http://localhost:5001/v2/_catalog

# Rebuild and push
docker build -t localhost:5001/cano-collector:dev-local .
docker push localhost:5001/cano-collector:dev-local

# Force pod restart
kubectl delete pod -n kubecano -l app=cano-collector
```

### Pod Crashes or CrashLoopBackOff

```bash
# Check logs for errors
kubectl logs -n kubecano -l app=cano-collector --previous

# Check pod events
kubectl describe pod -n kubecano -l app=cano-collector

# Common issues:
# - Missing Slack token in values-local.yaml
# - Invalid config YAML syntax
# - Resource limits too low
```

### No Alerts Reaching Slack

```bash
# Check cano-collector logs
kubectl logs -n kubecano -l app=cano-collector | grep -i "slack\|error\|failed"

# Verify Slack token is correct
kubectl get configmap -n kubecano -o yaml | grep -A5 slack

# Test Slack API directly
curl -X POST https://slack.com/api/chat.postMessage \
  -H "Authorization: Bearer xoxb-YOUR-TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"channel":"#your-channel","text":"Test from cano-collector"}'
```

### Config Changes Not Applied

```bash
# Edit values-local.yaml, then redeploy
make local-dev-config

# Force restart pods
kubectl rollout restart deployment cano-collector -n kubecano

# Verify config was updated
kubectl get configmap -n kubecano cano-collector-config -o yaml
```

## Architecture Differences: Local vs Production

### Local Setup (k3d)
- **No ArgoCD** - Direct Helm install
- **No Prometheus** - Test alerts via webhook only
- **Local registry** - Images built and pushed locally
- **Smaller resources** - 128Mi RAM / 500m CPU limits
- **Debug logging** - More verbose output

### Production Setup (AWS)
- **ArgoCD** - GitOps deployment
- **kube-prometheus-stack** - Full monitoring stack
- **ECR registry** - AWS container registry
- **Production resources** - 512Mi RAM / 1000m CPU
- **Info logging** - Standard output

## Configuration Files

- **`Makefile`** - Local development automation (targets: `local-*`)
- **`values-local.yaml`** - Local Helm values (Slack config, workflows)
  - Compared with production: `/Users/tnowodzinski/projekty/sadsharkdev/devops/argocd-k3s-dev-aws/apps/kubecano/cano-collector/values.yaml`
- **`helm/cano-collector/values.yaml`** - Production defaults
- **`Dockerfile`** - Container image build

## Next Steps

- **Production testing:** Use ArgoCD + AWS cluster
- **Alert integration:** Deploy kube-prometheus-stack locally
- **Custom workflows:** Edit `workflows:` section in `values-local.yaml`
- **New destinations:** Add MS Teams, PagerDuty, etc.

## Useful Commands

```bash
# Check k3d clusters
k3d cluster list

# Check cluster info
kubectl cluster-info

# List all resources in kubecano namespace
kubectl get all -n kubecano

# Check Helm releases
helm list -n kubecano

# Get deployment YAML
helm get values cano-collector -n kubecano

# Exec into pod (if needed)
kubectl exec -it -n kubecano deployment/cano-collector -- /bin/sh
```

## Performance Notes

- **Build time:** ~30-60 seconds (depends on changes)
- **Deploy time:** ~10-20 seconds (Helm + pod startup)
- **Total iteration:** ~1-2 minutes (code → running pod)
- **First deploy:** ~2-3 minutes (includes cluster creation)

## Further Reading

- [CLAUDE.md](./CLAUDE.md) - Project architecture and workflow details
- [tests/README.md](./tests/README.md) - Test pod scenarios
- [docs/](./docs/) - Full Sphinx documentation
- [k3d Documentation](https://k3d.io/) - k3d reference
