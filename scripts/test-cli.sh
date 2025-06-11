#!/bin/bash

# CLI Test Runner
# ===============
# This script runs the comprehensive CLI test suite

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_step() {
    echo -e "${BLUE}🔧 $1${NC}"
}

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

main() {
    log_info "🧪 Running Pivot CLI Test Suite"
    echo
    
    cd "$PROJECT_ROOT"
    
    # Run CLI-specific tests
    log_step "Running CLI command tests..."
    if TESTING=true go test -v -timeout=30s ./cmd/main_cli_test.go ./cmd/main.go; then
        log_success "CLI tests passed"
    else
        log_error "CLI tests failed"
        return 1
    fi
    
    # Run CSV CLI tests
    log_step "Running CSV CLI tests..."
    if TESTING=true go test -v -timeout=30s ./cmd/csv_cli_test.go ./cmd/main.go; then
        log_success "CSV CLI tests passed"
    else
        log_error "CSV CLI tests failed"
        return 1
    fi
    
    # Test binary compilation
    log_step "Testing binary compilation..."
    if go build -o /tmp/pivot-test ./cmd/main.go; then
        log_success "Binary compilation successful"
        
        # Test basic commands
        log_step "Testing basic command functionality..."
        if /tmp/pivot-test version &> /dev/null && /tmp/pivot-test help &> /dev/null; then
            log_success "Basic commands work"
        else
            log_error "Basic commands failed"
            return 1
        fi
        
        # Cleanup
        rm -f /tmp/pivot-test
    else
        log_error "Binary compilation failed"
        return 1
    fi
    
    log_success "All CLI tests passed!"
    return 0
}

# Run tests
main "$@"
