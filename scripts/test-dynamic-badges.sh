#!/bin/bash

# Test Dynamic Badges Integration
# ===============================
# This script tests the Schneegans dynamic-badges-action integration locally

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) is required but not installed."
        log_info "Install it from: https://cli.github.com/"
        exit 1
    fi
    
    if ! gh auth status &> /dev/null; then
        log_error "You must be logged in to GitHub CLI."
        log_info "Run: gh auth login"
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed."
        log_info "Install it with: brew install jq (macOS) or apt-get install jq (Ubuntu)"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Get current coverage
get_coverage() {
    if ! go test -v -coverprofile=coverage.out -covermode=atomic ./... >/dev/null 2>&1; then
        echo "0.0"
        return 1
    fi
    
    local coverage_percent=$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')
    echo "${coverage_percent}"
}

# Get security rating
get_security_rating() {
    # Install security tools if needed
    if ! command -v gosec &> /dev/null; then
        go install github.com/securego/gosec/v2/cmd/gosec@latest >/dev/null 2>&1
    fi
    
    if ! command -v govulncheck &> /dev/null; then
        go install golang.org/x/vuln/cmd/govulncheck@latest >/dev/null 2>&1
    fi
    
    local security_issues=0
    
    # Run gosec
    if gosec -fmt json -out gosec-report.json ./... >/dev/null 2>&1; then
        local gosec_issues=$(jq '.Issues | length' gosec-report.json 2>/dev/null || echo "0")
        security_issues=$((security_issues + gosec_issues))
    fi
    
    # Run govulncheck
    if ! govulncheck ./... >/dev/null 2>&1; then
        security_issues=$((security_issues + 1))
    fi
    
    # Determine security rating
    local rating color
    if [ $security_issues -eq 0 ]; then
        rating="A"
        color="brightgreen"
    elif [ $security_issues -le 2 ]; then
        rating="B"
        color="green"
    elif [ $security_issues -le 5 ]; then
        rating="C"
        color="yellow"
    else
        rating="D"
        color="red"
    fi
    
    echo "${rating},${color},${security_issues}"
}

# Get Go version
get_go_version() {
    grep "^go " go.mod | cut -d' ' -f2
}

# Test badge JSON generation (simulate what Schneegans action does)
test_badge_json_generation() {
    log_info "Testing badge JSON generation..."
    
    log_info "Calculating test coverage..."
    local coverage=$(get_coverage)
    if [ $? -ne 0 ]; then
        log_warning "Tests failed, using placeholder coverage value"
        coverage="0.0"
    fi
    
    log_info "Running security analysis..."
    local security_info=$(get_security_rating)
    local security_rating=$(echo "$security_info" | cut -d',' -f1)
    local security_color=$(echo "$security_info" | cut -d',' -f2)
    local security_issues=$(echo "$security_info" | cut -d',' -f3)
    
    local go_version=$(get_go_version)
    
    echo "📊 Badge Data Summary:"
    echo "   Coverage: ${coverage}%"
    echo "   Security: ${security_rating} (${security_issues} issues)"
    echo "   Go Version: ${go_version}"
    echo
    
    # Generate badge JSON files locally for testing
    mkdir -p /tmp/badge-test
    
    # Coverage badge with color range
    local coverage_color
    if (( $(echo "$coverage >= 90" | bc -l 2>/dev/null || echo "0") )); then
        coverage_color="brightgreen"
    elif (( $(echo "$coverage >= 70" | bc -l 2>/dev/null || echo "0") )); then
        coverage_color="green"
    elif (( $(echo "$coverage >= 50" | bc -l 2>/dev/null || echo "0") )); then
        coverage_color="yellow"
    else
        coverage_color="red"
    fi
    
    jq -n \
        --arg coverage "${coverage}%" \
        --arg color "$coverage_color" \
        '{
            schemaVersion: 1,
            label: "Coverage",
            message: $coverage,
            color: $color
        }' > /tmp/badge-test/coverage.json
    
    # Security badge
    jq -n \
        --arg rating "$security_rating" \
        --arg color "$security_color" \
        '{
            schemaVersion: 1,
            label: "Security",
            message: $rating,
            color: $color
        }' > /tmp/badge-test/security.json
    
    # Build status badge
    jq -n '{
        schemaVersion: 1,
        label: "Build",
        message: "passing",
        color: "brightgreen"
    }' > /tmp/badge-test/build.json
    
    # Go version badge
    jq -n \
        --arg version "$go_version" \
        '{
            schemaVersion: 1,
            label: "Go",
            message: $version,
            color: "00ADD8"
        }' > /tmp/badge-test/go-version.json
    
    # License badge
    jq -n '{
        schemaVersion: 1,
        label: "License",
        message: "MIT",
        color: "yellow"
    }' > /tmp/badge-test/license.json
    
    log_success "Badge JSON files generated in /tmp/badge-test/"
    
    # Show the JSON files
    echo
    log_info "Generated Badge JSON Files:"
    for file in /tmp/badge-test/*.json; do
        echo "📄 $(basename "$file"):"
        cat "$file" | jq .
        echo
    done
}

# Test gist access (simulate what the action would do)
test_gist_access() {
    log_info "Testing GitHub gist access..."
    
    # Get current gist IDs from environment or prompt
    local coverage_gist_id="${COVERAGE_GIST_ID:-8466693b8eb4ca358099fabc6ed234e0}"
    local security_gist_id="${SECURITY_GIST_ID:-a93cb6b503277dd460826517a831497e}"
    local badges_gist_id="${BADGES_GIST_ID:-0a39d1979cd714d14836e9d6427d2eb9}"
    
    echo "📋 Testing with Gist IDs:"
    echo "   Coverage: $coverage_gist_id"
    echo "   Security: $security_gist_id"
    echo "   Badges: $badges_gist_id"
    echo
    
    # Test gist access
    for gist_id in "$coverage_gist_id" "$security_gist_id" "$badges_gist_id"; do
        log_info "Testing access to gist $gist_id..."
        if gh api "/gists/$gist_id" >/dev/null 2>&1; then
            log_success "Gist $gist_id is accessible"
        else
            log_warning "Gist $gist_id is not accessible or doesn't exist"
        fi
    done
}

# Main function
main() {
    echo "🧪 Testing Schneegans Dynamic Badges Integration"
    echo "================================================="
    echo
    
    check_prerequisites
    echo
    
    test_badge_json_generation
    echo
    
    test_gist_access
    echo
    
    log_info "🎯 Next Steps:"
    echo "1. Ensure your repository has the required secrets:"
    echo "   - COVERAGE_GIST_ID"
    echo "   - SECURITY_GIST_ID"
    echo "   - BADGES_GIST_ID"
    echo "   - GIST_SECRET (GitHub Personal Access Token with 'gist' scope)"
    echo
    echo "2. Push code to main branch to trigger automated badge updates"
    echo
    echo "3. Check your badges in README.md - they should update automatically!"
    echo
    
    log_success "✨ Dynamic badges integration test completed!"
}

# Run main function
main "$@"
