#!/bin/bash

# Cleanup Script for Cano-Collector Test Resources
# Usage: ./cleanup.sh [namespace] [--all]
# Examples:
#   ./cleanup.sh                    # Clean default test-pods namespace
#   ./cleanup.sh my-test-ns         # Clean specific namespace  
#   ./cleanup.sh --all              # Clean all test namespaces
#   ./cleanup.sh my-test-ns --force # Force delete without confirmation

set -euo pipefail

# Configuration
NAMESPACE_DEFAULT="test-pods"
TEST_LABEL="test-type"

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
    echo "Usage: $0 [namespace] [--all|--force]"
    echo ""
    echo "Options:"
    echo "  namespace    Specific namespace to clean (default: test-pods)"
    echo "  --all        Clean all namespaces with test resources"
    echo "  --force      Skip confirmation prompts"
    echo ""
    echo "Examples:"
    echo "  $0                    # Clean default test-pods namespace"
    echo "  $0 my-test-ns         # Clean specific namespace"
    echo "  $0 --all              # Clean all test namespaces"
    echo "  $0 my-test-ns --force # Force delete without confirmation"
}

# Function to get confirmation
confirm_action() {
    local message="$1"
    local force="$2"
    
    if [ "$force" = "true" ]; then
        return 0
    fi
    
    echo -e "${YELLOW}$message${NC}"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled"
        exit 0
    fi
}

# Function to clean specific namespace
clean_namespace() {
    local namespace="$1"
    local force="$2"
    
    if ! kubectl get namespace "$namespace" &>/dev/null; then
        print_warning "Namespace '$namespace' does not exist"
        return 0
    fi
    
    print_info "Checking test resources in namespace '$namespace'..."
    
    # Check if namespace has test resources
    local test_resources=$(kubectl get all,configmaps,secrets -n "$namespace" -l "$TEST_LABEL" --no-headers 2>/dev/null | wc -l)
    
    if [ "$test_resources" -eq 0 ]; then
        print_info "No test resources found in namespace '$namespace'"
        
        # Check if namespace is empty and is a test namespace
        local all_resources=$(kubectl get all,configmaps,secrets -n "$namespace" --no-headers 2>/dev/null | wc -l)
        if [ "$all_resources" -eq 0 ] && [[ "$namespace" == *"test"* ]]; then
            confirm_action "Namespace '$namespace' is empty. Delete the namespace itself?" "$force"
            kubectl delete namespace "$namespace"
            print_success "Deleted empty test namespace '$namespace'"
        fi
        return 0
    fi
    
    print_info "Found $test_resources test resources in namespace '$namespace'"
    
    # Show what will be deleted
    echo ""
    print_info "Resources to be deleted:"
    kubectl get all,configmaps -n "$namespace" -l "$TEST_LABEL" 2>/dev/null || true
    echo ""
    
    confirm_action "Delete all test resources in namespace '$namespace'?" "$force"
    
    # Delete test resources
    print_info "Deleting test resources..."
    
    # Delete in specific order to avoid dependency issues
    kubectl delete jobs,cronjobs -n "$namespace" -l "$TEST_LABEL" --ignore-not-found=true
    kubectl delete deployments,replicasets -n "$namespace" -l "$TEST_LABEL" --ignore-not-found=true  
    kubectl delete services -n "$namespace" -l "$TEST_LABEL" --ignore-not-found=true
    kubectl delete pods -n "$namespace" -l "$TEST_LABEL" --ignore-not-found=true
    kubectl delete configmaps -n "$namespace" -l "$TEST_LABEL" --ignore-not-found=true
    
    print_success "Cleaned test resources from namespace '$namespace'"
    
    # Check if namespace is now empty and offer to delete it
    local remaining_resources=$(kubectl get all,configmaps,secrets -n "$namespace" --no-headers 2>/dev/null | wc -l)
    if [ "$remaining_resources" -eq 0 ] && [[ "$namespace" == *"test"* ]]; then
        confirm_action "Namespace '$namespace' is now empty. Delete the namespace itself?" "$force"
        kubectl delete namespace "$namespace"
        print_success "Deleted namespace '$namespace'"
    fi
}

# Function to find and clean all test namespaces
clean_all_namespaces() {
    local force="$1"
    
    print_info "Finding all namespaces with test resources..."
    
    # Find namespaces with test resources
    local test_namespaces=$(kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n' | while read ns; do
        if kubectl get all,configmaps -n "$ns" -l "$TEST_LABEL" --no-headers &>/dev/null; then
            local count=$(kubectl get all,configmaps -n "$ns" -l "$TEST_LABEL" --no-headers 2>/dev/null | wc -l)
            if [ "$count" -gt 0 ]; then
                echo "$ns"
            fi
        fi
    done)
    
    if [ -z "$test_namespaces" ]; then
        print_info "No test resources found in any namespace"
        return 0
    fi
    
    echo ""
    print_info "Found test resources in the following namespaces:"
    echo "$test_namespaces"
    echo ""
    
    confirm_action "Clean test resources from ALL these namespaces?" "$force"
    
    # Clean each namespace
    echo "$test_namespaces" | while read ns; do
        if [ -n "$ns" ]; then
            echo ""
            clean_namespace "$ns" "$force"
        fi
    done
}

# Function to show current test resources
show_test_resources() {
    print_info "Current test resources across all namespaces:"
    echo ""
    
    local found_any=false
    
    # Check each namespace for test resources
    kubectl get namespaces -o jsonpath='{.items[*].metadata.name}' | tr ' ' '\n' | while read ns; do
        local test_resources=$(kubectl get all,configmaps -n "$ns" -l "$TEST_LABEL" --no-headers 2>/dev/null | wc -l)
        if [ "$test_resources" -gt 0 ]; then
            echo "=== Namespace: $ns ==="
            kubectl get all,configmaps -n "$ns" -l "$TEST_LABEL" 2>/dev/null || true
            echo ""
            found_any=true
        fi
    done
    
    if [ "$found_any" = "false" ]; then
        print_info "No test resources found in any namespace"
    fi
}

# Main execution
main() {
    local namespace=""
    local clean_all=false
    local force=false
    
    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --all)
                clean_all=true
                shift
                ;;
            --force)
                force=true
                shift
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            --show)
                show_test_resources
                exit 0
                ;;
            -*)
                print_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                if [ -z "$namespace" ]; then
                    namespace="$1"
                else
                    print_error "Multiple namespaces specified"
                    show_usage
                    exit 1
                fi
                shift
                ;;
        esac
    done
    
    # Set default namespace if none specified and not cleaning all
    if [ "$clean_all" = "false" ] && [ -z "$namespace" ]; then
        namespace="$NAMESPACE_DEFAULT"
    fi
    
    print_info "Starting cleanup process..."
    
    if [ "$clean_all" = "true" ]; then
        clean_all_namespaces "$force"
    else
        clean_namespace "$namespace" "$force"
    fi
    
    print_success "Cleanup completed!"
}

# Execute main function with all arguments
main "$@"