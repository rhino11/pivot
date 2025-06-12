#!/bin/bash

# Script to diagnose and fix dynamic badge setup issues
# This helps identify missing secrets and gist configuration problems

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

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

echo "🏷️  Dynamic Badge Diagnostics"
echo "============================"
echo ""

# Check if required tools are installed
log_step "Checking required tools..."

if ! command -v curl >/dev/null 2>&1; then
    log_error "curl is required but not installed"
    exit 1
fi

if ! command -v jq >/dev/null 2>&1; then
    log_warning "jq is not installed - some features will be limited"
    JQ_AVAILABLE=false
else
    JQ_AVAILABLE=true
fi

log_success "Required tools check complete"
echo ""

# Check if we have a GIST_SECRET to test with
if [ $# -eq 0 ]; then
    log_info "Usage: $0 <gist_secret_token>"
    echo ""
    echo "This script tests badge setup using a Personal Access Token."
    echo "The token should have 'gist' permissions."
    echo ""
    echo "To get your token:"
    echo "1. Go to GitHub → Settings → Developer settings → Personal access tokens"
    echo "2. Generate new token (classic) with 'gist' scope"
    echo "3. Run: $0 <your_token>"
    echo ""
    log_warning "Without a token, only basic checks can be performed."
    echo ""
fi

GIST_SECRET="${1:-}"

# Function to test gist access
test_gist_access() {
    local gist_id="$1"
    local gist_name="$2"
    
    if [ -z "$GIST_SECRET" ]; then
        log_warning "Cannot test $gist_name gist (no token provided)"
        return 1
    fi
    
    log_step "Testing $gist_name gist access ($gist_id)..."
    
    # Test if gist exists and is accessible
    response=$(curl -s -w "%{http_code}" -H "Authorization: token $GIST_SECRET" \
        "https://api.github.com/gists/$gist_id" -o /tmp/gist_response.json)
    
    http_code="${response: -3}"
    
    case "$http_code" in
        200)
            log_success "$gist_name gist: OK"
            if [ "$JQ_AVAILABLE" = true ]; then
                public=$(jq -r '.public' /tmp/gist_response.json 2>/dev/null || echo "unknown")
                files=$(jq -r '.files | keys | length' /tmp/gist_response.json 2>/dev/null || echo "unknown")
                log_info "  Public: $public, Files: $files"
            fi
            return 0
            ;;
        404)
            log_error "$gist_name gist: NOT FOUND"
            log_info "  The gist ID might be incorrect or the gist was deleted"
            return 1
            ;;
        403)
            log_error "$gist_name gist: ACCESS DENIED"
            log_info "  The token doesn't have permission to access this gist"
            return 1
            ;;
        401)
            log_error "$gist_name gist: UNAUTHORIZED"
            log_info "  The token is invalid or doesn't have 'gist' scope"
            return 1
            ;;
        *)
            log_error "$gist_name gist: HTTP $http_code"
            return 1
            ;;
    esac
}

# Check the current workflow file for gist IDs
log_step "Analyzing CI workflow configuration..."

WORKFLOW_FILE=".github/workflows/ci.yml"
if [ ! -f "$WORKFLOW_FILE" ]; then
    log_error "CI workflow file not found: $WORKFLOW_FILE"
    exit 1
fi

# Extract gist ID references from workflow
COVERAGE_GIST_REF=$(grep -o 'COVERAGE_GIST_ID' "$WORKFLOW_FILE" || true)
SECURITY_GIST_REF=$(grep -o 'SECURITY_GIST_ID' "$WORKFLOW_FILE" || true)
BADGES_GIST_REF=$(grep -o 'BADGES_GIST_ID' "$WORKFLOW_FILE" || true)
GIST_SECRET_REF=$(grep -o 'GIST_SECRET' "$WORKFLOW_FILE" || true)

if [ -n "$COVERAGE_GIST_REF" ]; then
    log_success "Workflow references COVERAGE_GIST_ID"
else
    log_error "Workflow missing COVERAGE_GIST_ID reference"
fi

if [ -n "$SECURITY_GIST_REF" ]; then
    log_success "Workflow references SECURITY_GIST_ID"
else
    log_error "Workflow missing SECURITY_GIST_ID reference"
fi

if [ -n "$BADGES_GIST_REF" ]; then
    log_success "Workflow references BADGES_GIST_ID"
else
    log_error "Workflow missing BADGES_GIST_ID reference"
fi

if [ -n "$GIST_SECRET_REF" ]; then
    log_success "Workflow references GIST_SECRET"
else
    log_error "Workflow missing GIST_SECRET reference"
fi

echo ""

# Test token permissions if provided
if [ -n "$GIST_SECRET" ]; then
    log_step "Testing token permissions..."
    
    # Test basic token validity
    response=$(curl -s -w "%{http_code}" -H "Authorization: token $GIST_SECRET" \
        "https://api.github.com/user" -o /tmp/user_response.json)
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Token is valid"
        if [ "$JQ_AVAILABLE" = true ]; then
            username=$(jq -r '.login' /tmp/user_response.json 2>/dev/null || echo "unknown")
            log_info "  Authenticated as: $username"
        fi
    else
        log_error "Token is invalid (HTTP $http_code)"
        exit 1
    fi
    
    # Test gist permissions
    response=$(curl -s -w "%{http_code}" -H "Authorization: token $GIST_SECRET" \
        "https://api.github.com/gists" -o /tmp/gists_response.json)
    
    http_code="${response: -3}"
    
    if [ "$http_code" = "200" ]; then
        log_success "Token has gist permissions"
    else
        log_error "Token lacks gist permissions (HTTP $http_code)"
        log_info "  Make sure the token has 'gist' scope enabled"
        exit 1
    fi
    
    echo ""
    
    # Test specific gists if IDs are provided
    if [ $# -ge 2 ]; then
        log_step "Testing provided gist IDs..."
        for i in $(seq 2 $#); do
            gist_id="${!i}"
            test_gist_access "$gist_id" "Gist $((i-1))"
        done
    fi
fi

echo ""
log_step "Badge Setup Summary"
echo "==================="

echo ""
log_info "Required Repository Secrets:"
echo "  - GIST_SECRET: Personal Access Token with 'gist' scope"
echo "  - COVERAGE_GIST_ID: ID of gist for coverage badge"
echo "  - SECURITY_GIST_ID: ID of gist for security badge"
echo "  - BADGES_GIST_ID: ID of gist for build/go/license badges"

echo ""
log_info "To create gists:"
echo "  1. Go to https://gist.github.com"
echo "  2. Create public gist with filename: placeholder.json"
echo "  3. Content: {}"
echo "  4. Note the gist ID from the URL"
echo "  5. Repeat for all three gists"

echo ""
log_info "To add secrets:"
echo "  1. Go to repository Settings → Secrets and variables → Actions"
echo "  2. Add the four secrets listed above"
echo "  3. Re-run the CI workflow to test"

if [ -n "$GIST_SECRET" ]; then
    echo ""
    log_success "🎉 Token testing complete!"
    echo ""
    echo "Next steps:"
    echo "1. Create three public gists at https://gist.github.com"
    echo "2. Add the gist IDs as repository secrets"
    echo "3. Add this token as GIST_SECRET repository secret"
    echo "4. Re-run the CI workflow"
else
    echo ""
    log_warning "💡 Run with a token to test gist access:"
    echo "  $0 <your_gist_token>"
fi

# Cleanup
rm -f /tmp/gist_response.json /tmp/user_response.json /tmp/gists_response.json
