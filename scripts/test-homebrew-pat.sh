#!/bin/bash

# Script to test if HOMEBREW_PAT is properly configured
# This should be run locally with the PAT to verify it works

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

# Check if PAT is provided
if [ $# -ne 1 ]; then
    log_error "Usage: $0 <personal_access_token>"
    echo ""
    echo "This script tests if a Personal Access Token can access the homebrew-tap repository."
    echo "Use this to verify your PAT before adding it as a GitHub secret."
    echo ""
    echo "Example:"
    echo "  $0 ghp_your_personal_access_token_here"
    exit 1
fi

PAT="$1"

log_info "Testing Personal Access Token for Homebrew tap access..."
echo ""

# Test 1: Check if we can read the repository
log_info "Test 1: Checking repository read access..."
if curl -s -f -H "Authorization: token $PAT" \
   "https://api.github.com/repos/rhino11/homebrew-tap" >/dev/null; then
    log_success "Repository read access: OK"
else
    log_error "Repository read access: FAILED"
    echo "The token cannot read the homebrew-tap repository."
    exit 1
fi

# Test 2: Check if we can list repository contents
log_info "Test 2: Checking repository contents access..."
if curl -s -f -H "Authorization: token $PAT" \
   "https://api.github.com/repos/rhino11/homebrew-tap/contents" >/dev/null; then
    log_success "Repository contents access: OK"
else
    log_error "Repository contents access: FAILED"
    echo "The token cannot list repository contents."
    exit 1
fi

# Test 3: Check if we can clone the repository (read-only test)
log_info "Test 3: Testing git clone access..."
TEMP_DIR=$(mktemp -d)
if git clone "https://$PAT@github.com/rhino11/homebrew-tap.git" "$TEMP_DIR/test-clone" >/dev/null 2>&1; then
    log_success "Git clone access: OK"
    rm -rf "$TEMP_DIR"
else
    log_error "Git clone access: FAILED"
    echo "The token cannot clone the repository."
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Test 4: Check token scopes
log_info "Test 4: Checking token scopes..."
SCOPES=$(curl -s -H "Authorization: token $PAT" -I "https://api.github.com/user" | grep -i "x-oauth-scopes:" | cut -d' ' -f2- | tr -d '\r\n')

if [[ "$SCOPES" == *"repo"* ]]; then
    log_success "Token scopes: OK (includes 'repo')"
    log_info "Available scopes: $SCOPES"
else
    log_warning "Token scopes: Missing 'repo' scope"
    log_info "Available scopes: $SCOPES"
    echo ""
    echo "The token may not have sufficient permissions for write operations."
    echo "Make sure the token has 'repo' scope selected."
fi

# Test 5: Check if we can read the current formula
log_info "Test 5: Testing formula file access..."
if curl -s -f -H "Authorization: token $PAT" \
   "https://api.github.com/repos/rhino11/homebrew-tap/contents/Formula/pivot.rb" >/dev/null; then
    log_success "Formula file access: OK"
else
    log_warning "Formula file access: File might not exist yet"
    echo "This is normal for new tap repositories."
fi

echo ""
log_success "🎉 All tests passed! The Personal Access Token should work for automatic updates."
echo ""
echo "Next steps:"
echo "1. Go to https://github.com/rhino11/pivot/settings/secrets/actions"
echo "2. Add a new secret named 'HOMEBREW_PAT'"
echo "3. Use this token as the value"
echo "4. Test with a new release"
