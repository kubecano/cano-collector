#!/bin/bash

# Test Pod Runner Script for Cano-Collector Validation
# Usage: ./run-test.sh <category> <test-name> [namespace]
# Examples:
#   ./run-test.sh crash-loop busybox-crash
#   ./run-test.sh oom memory-bomb test-env
#   ./run-test.sh deployments replica-failure

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TESTS_DIR="$(dirname "$SCRIPT_DIR")"
NAMESPACE_DEFAULT="test-pods"

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
    echo "Usage: $0 <category> <test-name> [namespace]"
    echo ""
    echo "Available test categories and scenarios:"
    echo "  pods/crash-loop:     busybox-crash, jdk-crash, nginx-crash"
    echo "  pods/oom:            memory-bomb, gradual-oom"
    echo "  pods/image-pull:     nonexistent-image, private-registry"
    echo "  pods/resource-limits: cpu-starved, impossible-resources"
    echo "  pods/network:        dns-failure, service-unreachable"
    echo "  deployments:         replica-failure, rollout-failure, scaling-test"
    echo "  jobs:                job-failure, job-timeout, cronjob-failure"
    echo "  services:            no-endpoints, loadbalancer-pending"
    echo ""
    echo "Examples:"
    echo "  $0 crash-loop busybox-crash"
    echo "  $0 oom memory-bomb my-test-ns"
    echo "  $0 deployments replica-failure"
}

# Function to validate inputs
validate_inputs() {
    if [ $# -lt 2 ]; then
        print_error "Missing required arguments"
        show_usage
        exit 1
    fi

    CATEGORY="$1"
    TEST_NAME="$2"
    NAMESPACE="${3:-$NAMESPACE_DEFAULT}"

    # Determine test file path based on category
    case "$CATEGORY" in
        crash-loop|oom|image-pull|resource-limits|network)
            TEST_FILE="$TESTS_DIR/pods/$CATEGORY/$TEST_NAME.yaml"
            ;;
        deployments|jobs|services)
            TEST_FILE="$TESTS_DIR/$CATEGORY/$TEST_NAME.yaml"
            ;;
        *)
            print_error "Unknown test category: $CATEGORY"
            show_usage
            exit 1
            ;;
    esac

    if [ ! -f "$TEST_FILE" ]; then
        print_error "Test file not found: $TEST_FILE"
        show_usage
        exit 1
    fi
}

# Function to create namespace if it doesn't exist
ensure_namespace() {
    print_info "Ensuring namespace '$NAMESPACE' exists..."
    if ! kubectl get namespace "$NAMESPACE" &>/dev/null; then
        kubectl create namespace "$NAMESPACE"
        print_success "Created namespace '$NAMESPACE'"
    else
        print_info "Namespace '$NAMESPACE' already exists"
    fi
}

# Function to apply test resources
apply_test() {
    print_info "Applying test: $CATEGORY/$TEST_NAME"
    print_info "File: $TEST_FILE"
    print_info "Namespace: $NAMESPACE"
    
    kubectl apply -f "$TEST_FILE" -n "$NAMESPACE"
    print_success "Test resources applied successfully"
}

# Function to show monitoring commands
show_monitoring_info() {
    echo ""
    print_info "Test '$TEST_NAME' started in namespace '$NAMESPACE'"
    echo ""
    echo "Monitor the test with these commands:"
    echo ""
    
    case "$CATEGORY" in
        crash-loop|oom|image-pull|resource-limits|network)
            echo "  # Watch pod status:"
            echo "  kubectl get pods -n $NAMESPACE -w"
            echo ""
            echo "  # Check pod details:"
            echo "  kubectl describe pod -n $NAMESPACE -l scenario=$TEST_NAME"
            echo ""
            ;;
        deployments)
            echo "  # Watch deployment status:"
            echo "  kubectl get deployment -n $NAMESPACE -w"
            echo ""
            echo "  # Check deployment details:"
            echo "  kubectl describe deployment -n $NAMESPACE -l scenario=$TEST_NAME"
            echo ""
            echo "  # Watch pods created by deployment:"
            echo "  kubectl get pods -n $NAMESPACE -l scenario=$TEST_NAME -w"
            echo ""
            ;;
        jobs)
            echo "  # Watch job status:"
            echo "  kubectl get jobs -n $NAMESPACE -w"
            echo ""
            echo "  # Check job details:"
            echo "  kubectl describe job -n $NAMESPACE -l scenario=$TEST_NAME"
            echo ""
            ;;
        services)
            echo "  # Check service status:"
            echo "  kubectl get svc -n $NAMESPACE -l scenario=$TEST_NAME"
            echo ""
            echo "  # Check endpoints:"
            echo "  kubectl get endpoints -n $NAMESPACE"
            echo ""
            ;;
    esac
    
    echo "  # View all events in namespace:"
    echo "  kubectl get events -n $NAMESPACE --sort-by=.metadata.creationTimestamp"
    echo ""
    echo "  # Watch events in real-time:"
    echo "  kubectl get events -n $NAMESPACE --sort-by=.metadata.creationTimestamp -w"
    echo ""
    echo "  # Check cano-collector logs:"
    echo "  kubectl logs -n monitoring -l app=cano-collector -f"
    echo ""
    echo "  # Clean up when done:"
    echo "  ./cleanup.sh $NAMESPACE"
    echo ""
}

# Function to wait for initial state
wait_for_initial_state() {
    print_info "Waiting for initial resource state..."
    sleep 3
    
    case "$CATEGORY" in
        crash-loop|oom|image-pull|resource-limits|network)
            kubectl get pods -n "$NAMESPACE" -l scenario="$TEST_NAME"
            ;;
        deployments)
            kubectl get deployment -n "$NAMESPACE" -l scenario="$TEST_NAME"
            ;;
        jobs)
            kubectl get jobs -n "$NAMESPACE" -l scenario="$TEST_NAME"
            ;;
        services)
            kubectl get svc -n "$NAMESPACE" -l scenario="$TEST_NAME"
            ;;
    esac
}

# Main execution
main() {
    print_info "Starting test execution..."
    
    validate_inputs "$@"
    ensure_namespace
    apply_test
    wait_for_initial_state
    show_monitoring_info
    
    print_success "Test setup completed successfully!"
}

# Execute main function with all arguments
main "$@"