PACKAGE=github.com/kubecano/cano-collector
CURRENT_DIR=$(shell pwd)
DIST_DIR=${CURRENT_DIR}/dist
CLI_NAME=cano-collector
BIN_NAME=cano-collector
CGO_FLAG=0

HOST_OS:=$(shell go env GOOS)
HOST_ARCH:=$(shell go env GOARCH)

TARGET_ARCH?=linux/amd64

VERSION=$(shell cat ${CURRENT_DIR}/VERSION)

CANO_LINT_GOGC?=20

# Local development configuration
LOCAL_IMAGE_NAME=localhost:5001/cano-collector
LOCAL_IMAGE_TAG=latest
LOCAL_NAMESPACE=kubecano
LOCAL_RELEASE_NAME=cano-collector
LOCAL_CLUSTER_NAME=cano-dev
LOCAL_REGISTRY_NAME=cano-registry
LOCAL_REGISTRY_PORT=5001

.PHONY: gogen
gogen:
	export GO111MODULE=off
	go generate ./...

.PHONY: mod-download
mod-download:
	go mod download && go mod tidy # go mod download changes go.sum https://github.com/golang/go/issues/42970

.PHONY: mod-vendor
mod-vendor: mod-download
	go mod vendor

# Run linter on the code (local version)
.PHONY: lint
lint:
	golangci-lint --version
	# NOTE: If you get a "Killed" OOM message, try reducing the value of GOGC
	# See https://github.com/golangci/golangci-lint#memory-usage-of-golangci-lint
	GOGC=$(CANO_LINT_GOGC) GOMAXPROCS=2 golangci-lint run --fix --verbose

.PHONY: vet
vet:
	go vet -json ./...

# Build all Go code (local version)
.PHONY: build
build:
	go build -v `go list ./...`

# Run all unit tests (local version)
.PHONY: test
test:
	go test -v `go list ./...`

.PHONY: help
help:
	@echo 'Common targets'
	@echo
	@echo 'build:'
	@echo '  build                     -- compile go'
	@echo
	@echo 'local development:'
	@echo '  local-dev                 -- full local setup (cluster + build + deploy)'
	@echo '  local-dev-quick           -- quick rebuild and redeploy (skip tests)'
	@echo '  local-dev-config          -- config-only redeploy (skip build + tests)'
	@echo '  local-create-cluster      -- create k3d cluster with local registry'
	@echo '  local-delete-cluster      -- delete k3d cluster'
	@echo '  local-build-image         -- build and push Docker image to local registry'
	@echo '  local-deploy              -- deploy with Helm using values-local.yaml'
	@echo '  local-logs                -- tail cano-collector logs'
	@echo '  local-status              -- show deployment status'
	@echo '  local-test-alert          -- send basic test alert'
	@echo '  local-test-alert-with-logs -- send alert that triggers pod logs workflow'
	@echo '  local-test-alert-resolved -- send resolved alert (test threading)'
	@echo '  local-clean               -- delete Helm release (keep cluster)'

# ============================================================================
# Local Development Targets (k3d + Docker + Helm)
# ============================================================================

# Full local development setup: cluster + build + test + deploy
.PHONY: local-dev
local-dev: local-check-cluster test local-build-image local-deploy local-status
	@echo ""
	@echo "🎉 Local deployment successful!"
	@echo ""
	@echo "Useful commands:"
	@echo "  make local-logs                -- View logs"
	@echo "  make local-status              -- Check deployment status"
	@echo "  make local-test-alert          -- Send basic test alert"
	@echo "  make local-test-alert-with-logs -- Test pod logs workflow (uses live cano-collector pod)"
	@echo "  make local-clean               -- Remove deployment"
	@echo ""

# Quick iteration: rebuild + redeploy (skip tests)
.PHONY: local-dev-quick
local-dev-quick: local-check-cluster local-build-image local-deploy local-status
	@echo ""
	@echo "🚀 Quick redeploy complete!"
	@echo ""

# Config-only redeploy: no build, no tests
.PHONY: local-dev-config
local-dev-config: local-check-cluster local-deploy local-status
	@echo ""
	@echo "⚙️  Config update complete!"
	@echo ""

# Check if cluster exists, create if not
.PHONY: local-check-cluster
local-check-cluster:
	@echo "🔍 Checking k3d cluster..."
	@if ! k3d cluster list | grep -q "$(LOCAL_CLUSTER_NAME)"; then \
		echo "⚠️  Cluster '$(LOCAL_CLUSTER_NAME)' not found. Creating..."; \
		$(MAKE) local-create-cluster; \
	else \
		echo "✅ Cluster '$(LOCAL_CLUSTER_NAME)' is running"; \
	fi

# Create k3d cluster with local registry
.PHONY: local-create-cluster
local-create-cluster:
	@echo "🚀 Creating k3d cluster '$(LOCAL_CLUSTER_NAME)'..."
	k3d cluster create $(LOCAL_CLUSTER_NAME) \
		--port "8080:80@loadbalancer" \
		--registry-create $(LOCAL_REGISTRY_NAME):$(LOCAL_REGISTRY_PORT)
	@echo "✅ Cluster created successfully"

# Delete k3d cluster
.PHONY: local-delete-cluster
local-delete-cluster:
	@echo "🗑️  Deleting k3d cluster '$(LOCAL_CLUSTER_NAME)'..."
	k3d cluster delete $(LOCAL_CLUSTER_NAME)
	@echo "✅ Cluster deleted"

# Build and push Docker image to local registry
.PHONY: local-build-image
local-build-image:
	@echo "🐳 Building Docker image..."
	docker build -t $(LOCAL_IMAGE_NAME):$(LOCAL_IMAGE_TAG) .
	@echo "✅ Image built: $(LOCAL_IMAGE_NAME):$(LOCAL_IMAGE_TAG)"
	@echo "📤 Pushing to local registry..."
	docker push $(LOCAL_IMAGE_NAME):$(LOCAL_IMAGE_TAG)
	@echo "✅ Image pushed"

# Create namespace if not exists
.PHONY: local-create-namespace
local-create-namespace:
	@echo "📁 Creating namespace '$(LOCAL_NAMESPACE)'..."
	@kubectl create namespace $(LOCAL_NAMESPACE) --dry-run=client -o yaml | kubectl apply -f - > /dev/null 2>&1
	@echo "✅ Namespace ready"

# Deploy with Helm using values-local.yaml
.PHONY: local-deploy
local-deploy: local-create-namespace
	@echo "🚢 Deploying with Helm..."
	helm upgrade --install $(LOCAL_RELEASE_NAME) ./helm/cano-collector \
		--namespace $(LOCAL_NAMESPACE) \
		-f values-local.yaml \
		--set collector.image.tag=$(LOCAL_IMAGE_TAG) \
		--wait \
		--timeout 5m
	@echo "🔄 Restarting deployment to pull new image..."
	kubectl rollout restart deployment/$(LOCAL_RELEASE_NAME) -n $(LOCAL_NAMESPACE)
	kubectl rollout status deployment/$(LOCAL_RELEASE_NAME) -n $(LOCAL_NAMESPACE) --timeout=2m
	@echo "✅ Deployment complete"

# Show deployment status
.PHONY: local-status
local-status:
	@echo "📊 Deployment status:"
	@kubectl get pods -n $(LOCAL_NAMESPACE) -l app.kubernetes.io/name=cano-collector

# Tail logs
.PHONY: local-logs
local-logs:
	@echo "📋 Tailing logs (Ctrl+C to exit)..."
	kubectl logs -n $(LOCAL_NAMESPACE) -l app.kubernetes.io/name=cano-collector -f

# Port forward for local testing
.PHONY: local-port-forward
local-port-forward:
	@echo "🌐 Port forwarding to localhost:8080..."
	@echo "Press Ctrl+C to stop"
	kubectl port-forward -n $(LOCAL_NAMESPACE) svc/cano-collector 8080:8080

# Send test alert
.PHONY: local-test-alert
local-test-alert:
	@echo "🧪 Sending test alert..."
	@echo "Sending alert directly to service in cluster..."
	kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -n kubecano -- \
		curl -X POST http://cano-collector.kubecano.svc.cluster.local:80/api/alerts \
		-H 'Content-Type: application/json' \
		-d '{"receiver":"cano-collector","status":"firing","alerts":[{"status":"firing","labels":{"alertname":"KubePodCrashLooping","container":"busybox","namespace":"test-pods","pod":"busybox-crash-test","severity":"warning","uid":"test-uid-123"},"annotations":{"description":"Pod test-pods/busybox-crash-test (busybox) is in waiting state (reason: CrashLoopBackOff).","summary":"Pod is crash looping.","runbook_url":"https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping/"},"startsAt":"2025-11-10T19:00:00.000Z","endsAt":"0001-01-01T00:00:00Z","generatorURL":"http://prometheus:9090/graph","fingerprint":"test123456"}],"groupLabels":{"alertname":"KubePodCrashLooping"},"commonLabels":{"alertname":"KubePodCrashLooping","severity":"warning"},"commonAnnotations":{"summary":"Pod is crash looping."},"externalURL":"http://alertmanager:9093","version":"4","groupKey":"{}:{alertname=\"KubePodCrashLooping\"}"}'
	@echo ""
	@echo "✅ Alert sent! Check Slack channel and logs."

# Send test alert for cano-collector pod (with logs workflow)
.PHONY: local-test-alert-with-logs
local-test-alert-with-logs:
	@echo "🧪 Sending test alert with pod logs workflow..."
	@echo "Getting cano-collector pod name..."
	@POD_NAME=$$(kubectl get pods -n $(LOCAL_NAMESPACE) -l app.kubernetes.io/name=cano-collector -o jsonpath='{.items[0].metadata.name}' 2>/dev/null); \
	if [ -z "$$POD_NAME" ]; then \
		echo "❌ Error: No cano-collector pod found in namespace $(LOCAL_NAMESPACE)"; \
		echo "Run 'make local-deploy' first to create the pod."; \
		exit 1; \
	fi; \
	echo "📦 Found pod: $$POD_NAME"; \
	echo ""; \
	echo "Sending KubePodCrashLooping alert for cano-collector pod..."; \
	kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -n $(LOCAL_NAMESPACE) -- \
		curl -X POST http://cano-collector.$(LOCAL_NAMESPACE).svc.cluster.local:80/api/alerts \
		-H 'Content-Type: application/json' \
		-d "{\"receiver\":\"cano-collector\",\"status\":\"firing\",\"alerts\":[{\"status\":\"firing\",\"labels\":{\"alertname\":\"KubePodCrashLooping\",\"container\":\"cano-collector\",\"namespace\":\"$(LOCAL_NAMESPACE)\",\"pod\":\"$$POD_NAME\",\"severity\":\"warning\",\"uid\":\"test-uid-collector-123\"},\"annotations\":{\"description\":\"Pod $(LOCAL_NAMESPACE)/$$POD_NAME (cano-collector) is being tested for log collection.\",\"summary\":\"Testing pod log collection workflow.\",\"runbook_url\":\"https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping/\"},\"startsAt\":\"$$(date -u +%Y-%m-%dT%H:%M:%S.000Z)\",\"endsAt\":\"0001-01-01T00:00:00Z\",\"generatorURL\":\"http://prometheus:9090/graph\",\"fingerprint\":\"test-collector-logs-$$(date +%s)\"}],\"groupLabels\":{\"alertname\":\"KubePodCrashLooping\"},\"commonLabels\":{\"alertname\":\"KubePodCrashLooping\",\"severity\":\"warning\"},\"commonAnnotations\":{\"summary\":\"Testing pod log collection workflow.\"},\"externalURL\":\"http://alertmanager:9093\",\"version\":\"4\",\"groupKey\":\"{}:{alertname=\\\"KubePodCrashLooping\\\"}\"}"
	@echo ""
	@echo "✅ Alert sent! This will trigger pod_logs workflow."
	@echo ""
	@echo "Expected workflow actions:"
	@echo "  1. pod_logs action will fetch logs from cano-collector pod"
	@echo "  2. Logs will be uploaded to Slack as file attachment"
	@echo "  3. Check Slack channel for message with log file"
	@echo ""
	@echo "To verify workflow execution, check logs:"
	@echo "  make local-logs | grep -E 'pod_logs|upload'"

# Send resolved test alert (for testing threading)
.PHONY: local-test-alert-resolved
local-test-alert-resolved:
	@echo "🧪 Sending resolved test alert (for threading test)..."
	@echo "Sending alert directly to service in cluster..."
	@END_TIME=$$(date -u +%Y-%m-%dT%H:%M:%S.000Z); \
	kubectl run curl-test --image=curlimages/curl:latest --rm -i --restart=Never -n kubecano -- \
		curl -X POST http://cano-collector.kubecano.svc.cluster.local:80/api/alerts \
		-H 'Content-Type: application/json' \
		-d "{\"receiver\":\"cano-collector\",\"status\":\"resolved\",\"alerts\":[{\"status\":\"resolved\",\"labels\":{\"alertname\":\"KubePodCrashLooping\",\"container\":\"busybox\",\"namespace\":\"test-pods\",\"pod\":\"busybox-crash-test\",\"severity\":\"warning\",\"uid\":\"test-uid-123\"},\"annotations\":{\"description\":\"Pod test-pods/busybox-crash-test (busybox) is no longer crash looping.\",\"summary\":\"Pod is crash looping.\",\"runbook_url\":\"https://runbooks.prometheus-operator.dev/runbooks/kubernetes/kubepodcrashlooping/\"},\"startsAt\":\"2025-11-10T19:00:00.000Z\",\"endsAt\":\"$$END_TIME\",\"generatorURL\":\"http://prometheus:9090/graph\",\"fingerprint\":\"test123456\"}],\"groupLabels\":{\"alertname\":\"KubePodCrashLooping\"},\"commonLabels\":{\"alertname\":\"KubePodCrashLooping\",\"severity\":\"warning\"},\"commonAnnotations\":{\"summary\":\"Pod is crash looping.\"},\"externalURL\":\"http://alertmanager:9093\",\"version\":\"4\",\"groupKey\":\"{}:{alertname=\\\"KubePodCrashLooping\\\"}\"}"
	@echo ""
	@echo "✅ Resolved alert sent! Should reply in thread of firing alert."

# Delete Helm release (keep cluster)
.PHONY: local-clean
local-clean:
	@echo "🧹 Removing Helm release..."
	helm uninstall $(LOCAL_RELEASE_NAME) -n $(LOCAL_NAMESPACE) || true
	@echo "✅ Release removed"

# Full cleanup (cluster + images)
.PHONY: local-clean-all
local-clean-all: local-clean local-delete-cluster
	@echo "🧹 Removing local Docker images..."
	docker rmi $(LOCAL_IMAGE_NAME):$(LOCAL_IMAGE_TAG) || true
	@echo "✅ Full cleanup complete"
