#!/bin/bash

# Badge Testing Script
# ====================
# Test the dynamic badge system locally without triggering CI.

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

# Test gist authentication
test_gist_auth() {
    log_info "Testing GitHub gist authentication..."
    
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI (gh) is required but not installed."
        return 1
    fi
    
    if ! gh auth status &> /dev/null; then
        log_error "You must be logged in to GitHub CLI."
        return 1
    fi
    
    # Test gist access
    if gh api /gists > /dev/null 2>&1; then
        log_success "GitHub gist authentication working"
        return 0
    else
        log_error "GitHub gist authentication failed"
        return 1
    fi
}

# Test gist creation and update
test_gist_operations() {
    log_info "Testing gist create/update operations..."
    
    # Create a test gist
    local temp_file=$(mktemp)
    echo '{"schemaVersion": 1, "label": "Test", "message": "working", "color": "brightgreen"}' > "$temp_file"
    
    local gist_url=$(gh gist create --public --desc "Badge Test - DELETE ME" --filename "test-badge.json" "$temp_file")
    local gist_id=$(echo "$gist_url" | sed 's/.*\///')
    
    log_success "Created test gist: $gist_id"
    
    # Test gist update
    echo '{"schemaVersion": 1, "label": "Test", "message": "updated", "color": "blue"}' > "$temp_file"
    gh gist edit "$gist_id" -f test-badge.json < "$temp_file"
    
    log_success "Updated test gist successfully"
    
    # Test public access
    local gist_url="https://gist.githubusercontent.com/$(gh api user --jq .login)/$gist_id/raw/test-badge.json"
    if curl -s "$gist_url" | jq -r '.message' | grep -q "updated"; then
        log_success "Gist is publicly accessible"
    else
        log_error "Gist is not publicly accessible"
        rm "$temp_file"
        return 1
    fi
    
    # Cleanup
    gh gist delete "$gist_id" --confirm
    rm "$temp_file"
    log_success "Cleaned up test gist"
    
    return 0
}

# Simulate badge generation like CI would do
simulate_badge_generation() {
    log_info "Simulating CI badge generation..."
    
    # Get current secrets
    local coverage_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="COVERAGE_GIST_ID") | .value' 2>/dev/null || echo "")
    local security_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="SECURITY_GIST_ID") | .value' 2>/dev/null || echo "")
    local badges_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="BADGES_GIST_ID") | .value' 2>/dev/null || echo "")
    
    if [[ -z "$coverage_gist_id" || -z "$security_gist_id" || -z "$badges_gist_id" ]]; then
        log_error "Repository secrets not found. Run setup-badges.sh first."
        return 1
    fi
    
    log_info "Found gist IDs:"
    echo "  Coverage: $coverage_gist_id"
    echo "  Security: $security_gist_id"
    echo "  Badges: $badges_gist_id"
    
    # Simulate coverage badge update
    log_info "Testing coverage badge update..."
    local coverage_data='{"schemaVersion": 1, "label": "Coverage", "message": "85.2%", "color": "brightgreen"}'
    local temp_file=$(mktemp)
    echo "$coverage_data" > "$temp_file"
    
    if gh gist edit "$coverage_gist_id" -f pivot-coverage.json < "$temp_file"; then
        log_success "Coverage badge updated successfully"
    else
        log_error "Failed to update coverage badge"
        rm "$temp_file"
        return 1
    fi
    
    # Simulate security badge update
    log_info "Testing security badge update..."
    local security_data='{"schemaVersion": 1, "label": "Security", "message": "A", "color": "brightgreen"}'
    echo "$security_data" > "$temp_file"
    
    if gh gist edit "$security_gist_id" -f pivot-security.json < "$temp_file"; then
        log_success "Security badge updated successfully"
    else
        log_error "Failed to update security badge"
        rm "$temp_file"
        return 1
    fi
    
    # Simulate build status badge update
    log_info "Testing build status badge update..."
    local build_data='{"schemaVersion": 1, "label": "Build", "message": "passing", "color": "brightgreen"}'
    echo "$build_data" > "$temp_file"
    
    if gh gist edit "$badges_gist_id" -f pivot-build.json < "$temp_file"; then
        log_success "Build status badge updated successfully"
    else
        log_error "Failed to update build status badge"
        rm "$temp_file"
        return 1
    fi
    
    # Test Go version badge
    log_info "Testing Go version badge update..."
    local go_version=$(grep "^go " go.mod | cut -d' ' -f2)
    local go_data='{"schemaVersion": 1, "label": "Go", "message": "'$go_version'", "color": "00ADD8"}'
    echo "$go_data" > "$temp_file"
    
    if gh gist edit "$badges_gist_id" -f pivot-go-version.json < "$temp_file"; then
        log_success "Go version badge updated successfully"
    else
        log_error "Failed to update Go version badge"
        rm "$temp_file"
        return 1
    fi
    
    # Test license badge
    log_info "Testing license badge update..."
    local license_data='{"schemaVersion": 1, "label": "License", "message": "MIT", "color": "yellow"}'
    echo "$license_data" > "$temp_file"
    
    if gh gist edit "$badges_gist_id" -f pivot-license.json < "$temp_file"; then
        log_success "License badge updated successfully"
    else
        log_error "Failed to update license badge"
        rm "$temp_file"
        return 1
    fi
    
    rm "$temp_file"
    return 0
}

# Test badge accessibility
test_badge_accessibility() {
    log_info "Testing badge accessibility..."
    
    local username=$(gh api user --jq .login)
    local coverage_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="COVERAGE_GIST_ID") | .value' 2>/dev/null || echo "")
    local security_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="SECURITY_GIST_ID") | .value' 2>/dev/null || echo "")
    local badges_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="BADGES_GIST_ID") | .value' 2>/dev/null || echo "")
    
    if [[ -z "$coverage_gist_id" || -z "$security_gist_id" || -z "$badges_gist_id" ]]; then
        log_error "Repository secrets not found."
        return 1
    fi
    
    # Test each badge endpoint
    local badges=(
        "Coverage:https://gist.githubusercontent.com/$username/$coverage_gist_id/raw/pivot-coverage.json"
        "Security:https://gist.githubusercontent.com/$username/$security_gist_id/raw/pivot-security.json"
        "Build:https://gist.githubusercontent.com/$username/$badges_gist_id/raw/pivot-build.json"
        "Go Version:https://gist.githubusercontent.com/$username/$badges_gist_id/raw/pivot-go-version.json"
        "License:https://gist.githubusercontent.com/$username/$badges_gist_id/raw/pivot-license.json"
    )
    
    for badge in "${badges[@]}"; do
        local name=$(echo "$badge" | cut -d':' -f1)
        local url=$(echo "$badge" | cut -d':' -f2-)
        
        log_info "Testing $name badge..."
        
        local response=$(curl -s "$url")
        if echo "$response" | jq . > /dev/null 2>&1; then
            local message=$(echo "$response" | jq -r '.message')
            log_success "$name badge accessible: $message"
        else
            log_error "$name badge not accessible or invalid JSON"
            echo "Response: $response"
            return 1
        fi
    done
    
    return 0
}

# Generate badge preview URLs
generate_badge_preview() {
    log_info "Generating badge preview URLs..."
    
    local username=$(gh api user --jq .login)
    local coverage_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="COVERAGE_GIST_ID") | .value' 2>/dev/null || echo "")
    local security_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="SECURITY_GIST_ID") | .value' 2>/dev/null || echo "")
    local badges_gist_id=$(gh secret list --json name,value | jq -r '.[] | select(.name=="BADGES_GIST_ID") | .value' 2>/dev/null || echo "")
    
    echo
    log_info "📋 Badge Preview URLs:"
    echo
    echo "Build Status:"
    echo "https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$username/$badges_gist_id/raw/pivot-build.json"
    echo
    echo "Coverage:"
    echo "https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$username/$coverage_gist_id/raw/pivot-coverage.json"
    echo
    echo "Security:"
    echo "https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$username/$security_gist_id/raw/pivot-security.json"
    echo
    echo "Go Version:"
    echo "https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$username/$badges_gist_id/raw/pivot-go-version.json"
    echo
    echo "License:"
    echo "https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$username/$badges_gist_id/raw/pivot-license.json"
    echo
}

# Main test function
main() {
    log_info "🧪 Testing Badge System Locally"
    echo
    
    # Check prerequisites
    if ! command -v curl &> /dev/null; then
        log_error "curl is required but not installed."
        exit 1
    fi
    
    if ! command -v jq &> /dev/null; then
        log_error "jq is required but not installed."
        log_info "Install with: brew install jq"
        exit 1
    fi
    
    # Run tests
    if test_gist_auth; then
        log_success "✅ Authentication test passed"
    else
        log_error "❌ Authentication test failed"
        exit 1
    fi
    
    echo
    if test_gist_operations; then
        log_success "✅ Gist operations test passed"
    else
        log_error "❌ Gist operations test failed"
        exit 1
    fi
    
    echo
    if simulate_badge_generation; then
        log_success "✅ Badge generation test passed"
    else
        log_error "❌ Badge generation test failed"
        exit 1
    fi
    
    echo
    if test_badge_accessibility; then
        log_success "✅ Badge accessibility test passed"
    else
        log_error "❌ Badge accessibility test failed"
        exit 1
    fi
    
    echo
    generate_badge_preview
    
    echo
    log_success "🎉 All badge tests passed!"
    log_info "The badge system is working correctly. CI should now succeed."
}

# Run main function
main "$@"
