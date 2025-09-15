#!/bin/bash

# Validation Script for Cano-Collector Event Processing
# Usage: ./validate-collector.sh [--namespace NAMESPACE] [--timeout SECONDS]
# Examples:
#   ./validate-collector.sh                           # Check default setup
#   ./validate-collector.sh --namespace monitoring    # Check specific namespace
#   ./validate-collector.sh --timeout 30              # Custom timeout

set -e

# Configuration
COLLECTOR_NAMESPACE_DEFAULT="monitoring"
COLLECTOR_SERVICE_NAME="cano-collector"
COLLECTOR_PORT="8080"
TIMEOUT_DEFAULT="10"
PORT_FORWARD_PID=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
print_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --namespace NS    Namespace where cano-collector is deployed (default: monitoring)"
    echo "  --service NAME    Service name for cano-collector (default: cano-collector)"
    echo "  --port PORT       Service port (default: 8080)"
    echo "  --timeout SEC     Connection timeout in seconds (default: 10)"
    echo "  --help, -h        Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                                    # Check with defaults"
    echo "  $0 --namespace kube-system            # Custom namespace"
    echo "  $0 --timeout 30                      # Custom timeout"
}

# Function to cleanup port-forward on exit
cleanup() {
    if [ -n "$PORT_FORWARD_PID" ]; then
        print_info "Cleaning up port-forward (PID: $PORT_FORWARD_PID)..."
        kill $PORT_FORWARD_PID 2>/dev/null || true
        wait $PORT_FORWARD_PID 2>/dev/null || true
    fi
}

# Set trap for cleanup
trap cleanup EXIT

# Function to check if cano-collector pods are running
check_pods() {
    local namespace="$1"
    
    print_info "Checking cano-collector pods in namespace '$namespace'..."
    
    local pods=$(kubectl get pods -n "$namespace" -l app=cano-collector --no-headers 2>/dev/null | wc -l)
    
    if [ "$pods" -eq 0 ]; then
        print_error "No cano-collector pods found in namespace '$namespace'"
        print_info "Check if cano-collector is deployed with label 'app=cano-collector'"
        return 1
    fi
    
    print_success "Found $pods cano-collector pod(s)"
    
    # Show pod status
    kubectl get pods -n "$namespace" -l app=cano-collector
    
    # Check if any pods are not ready
    local not_ready=$(kubectl get pods -n "$namespace" -l app=cano-collector --no-headers | grep -v "Running\|Completed" | wc -l)
    
    if [ "$not_ready" -gt 0 ]; then
        print_warning "$not_ready pod(s) are not in Running state"
        return 1
    fi
    
    print_success "All cano-collector pods are running"
    return 0
}

# Function to check if service exists
check_service() {
    local namespace="$1"
    local service_name="$2"
    
    print_info "Checking service '$service_name' in namespace '$namespace'..."
    
    if ! kubectl get svc "$service_name" -n "$namespace" &>/dev/null; then
        print_error "Service '$service_name' not found in namespace '$namespace'"
        print_info "Available services:"
        kubectl get svc -n "$namespace"
        return 1
    fi
    
    print_success "Service '$service_name' found"
    kubectl get svc "$service_name" -n "$namespace"
    return 0
}

# Function to setup port-forward
setup_port_forward() {
    local namespace="$1"
    local service_name="$2"
    local port="$3"
    
    print_info "Setting up port-forward to $service_name:$port..."
    
    # Start port-forward in background
    kubectl port-forward "svc/$service_name" "$port:$port" -n "$namespace" >/dev/null 2>&1 &
    PORT_FORWARD_PID=$!
    
    # Wait for port-forward to be ready
    sleep 3
    
    # Check if port-forward is still running
    if ! kill -0 $PORT_FORWARD_PID 2>/dev/null; then
        print_error "Port-forward failed to start"
        return 1
    fi
    
    print_success "Port-forward established (PID: $PORT_FORWARD_PID)"
    return 0
}

# Function to check health endpoint
check_health() {
    local port="$1"
    local timeout="$2"
    
    print_info "Checking health endpoint at http://localhost:$port/health..."
    
    local health_response
    if health_response=$(curl -s --max-time "$timeout" "http://localhost:$port/health" 2>/dev/null); then
        print_success "Health endpoint responded successfully"
        echo "Response: $health_response"
        return 0
    else
        print_error "Health endpoint is not responding"
        print_info "This could indicate:"
        print_info "  - Cano-collector is not ready"
        print_info "  - Health endpoint is not available"
        print_info "  - Port-forward connection issues"
        return 1
    fi
}

# Function to check metrics endpoint
check_metrics() {
    local port="$1"
    local timeout="$2"
    
    print_info "Checking metrics endpoint at http://localhost:$port/metrics..."
    
    local metrics_response
    if metrics_response=$(curl -s --max-time "$timeout" "http://localhost:$port/metrics" 2>/dev/null); then
        print_success "Metrics endpoint responded successfully"
        
        # Look for cano-collector specific metrics
        local event_metrics=$(echo "$metrics_response" | grep -E "(event|alert|kubernetes)" | head -5)
        if [ -n "$event_metrics" ]; then
            print_success "Found event processing metrics:"
            echo "$event_metrics"
        else
            print_warning "No event processing metrics found"
        fi
        
        return 0
    else
        print_error "Metrics endpoint is not responding"
        return 1
    fi
}

# Function to check recent logs
check_logs() {
    local namespace="$1"
    
    print_info "Checking recent cano-collector logs..."
    
    local log_lines=$(kubectl logs -n "$namespace" -l app=cano-collector --tail=20 --since=5m 2>/dev/null | wc -l)
    
    if [ "$log_lines" -eq 0 ]; then
        print_warning "No recent logs found (last 5 minutes)"
        print_info "Checking older logs..."
        kubectl logs -n "$namespace" -l app=cano-collector --tail=10 2>/dev/null || print_error "No logs available"
    else
        print_success "Found $log_lines recent log lines"
        print_info "Recent log sample:"
        kubectl logs -n "$namespace" -l app=cano-collector --tail=5 --since=5m 2>/dev/null
    fi
}

# Function to test event processing
test_event_processing() {
    local namespace="$1"
    
    print_info "Testing event processing capabilities..."
    
    # Check if there are any events in the cluster
    local recent_events=$(kubectl get events --all-namespaces --field-selector involvedObject.kind=Pod --since-time=$(date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ) --no-headers 2>/dev/null | wc -l)
    
    if [ "$recent_events" -gt 0 ]; then
        print_success "Found $recent_events recent Pod events in the cluster"
        print_info "Sample recent events:"
        kubectl get events --all-namespaces --field-selector involvedObject.kind=Pod --since-time=$(date -u -d '5 minutes ago' +%Y-%m-%dT%H:%M:%SZ) --no-headers 2>/dev/null | head -3
    else
        print_warning "No recent Pod events found in the cluster"
        print_info "Consider running a test pod to generate events:"
        print_info "  ./run-test.sh crash-loop busybox-crash"
    fi
}

# Function to provide recommendations
provide_recommendations() {
    echo ""
    print_info "=== RECOMMENDATIONS ==="
    echo ""
    
    print_info "To test cano-collector event processing:"
    echo "  1. Run a test scenario:"
    echo "     ./run-test.sh crash-loop busybox-crash"
    echo ""
    echo "  2. Monitor cano-collector logs:"
    echo "     kubectl logs -n monitoring -l app=cano-collector -f"
    echo ""
    echo "  3. Check for new events:"
    echo "     kubectl get events -n test-pods --sort-by=.metadata.creationTimestamp"
    echo ""
    echo "  4. Verify alert generation (if configured):"
    echo "     kubectl logs -n monitoring prometheus-alertmanager-0"
    echo ""
    
    print_info "For troubleshooting:"
    echo "  - Check cano-collector configuration"
    echo "  - Verify RBAC permissions for event watching"
    echo "  - Ensure webhook endpoints are reachable"
    echo "  - Check network policies and firewall rules"
}

# Main execution
main() {
    local namespace="$COLLECTOR_NAMESPACE_DEFAULT"
    local service_name="$COLLECTOR_SERVICE_NAME"
    local port="$COLLECTOR_PORT"
    local timeout="$TIMEOUT_DEFAULT"
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --namespace)
                namespace="$2"
                shift 2
                ;;
            --service)
                service_name="$2"
                shift 2
                ;;
            --port)
                port="$2"
                shift 2
                ;;
            --timeout)
                timeout="$2"
                shift 2
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
        esac
    done
    
    print_info "Starting cano-collector validation..."
    print_info "Namespace: $namespace"
    print_info "Service: $service_name"
    print_info "Port: $port"
    echo ""
    
    local validation_failed=false
    
    # Run validation steps
    if ! check_pods "$namespace"; then
        validation_failed=true
    fi
    echo ""
    
    if ! check_service "$namespace" "$service_name"; then
        validation_failed=true
    fi
    echo ""
    
    if ! setup_port_forward "$namespace" "$service_name" "$port"; then
        validation_failed=true
    fi
    echo ""
    
    if ! check_health "$port" "$timeout"; then
        validation_failed=true
    fi
    echo ""
    
    if ! check_metrics "$port" "$timeout"; then
        validation_failed=true
    fi
    echo ""
    
    check_logs "$namespace"
    echo ""
    
    test_event_processing "$namespace"
    
    # Summary
    echo ""
    if [ "$validation_failed" = "true" ]; then
        print_error "=== VALIDATION FAILED ==="
        print_error "Cano-collector is not functioning properly"
        provide_recommendations
        exit 1
    else
        print_success "=== VALIDATION SUCCESSFUL ==="
        print_success "Cano-collector is running and accessible"
        provide_recommendations
        exit 0
    fi
}

# Execute main function with all arguments
main "$@"