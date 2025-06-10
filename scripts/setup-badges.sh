#!/bin/bash

# Dynamic Badge Setup Script
# ===========================
# This script helps set up the gists and secrets needed for dynamic badges.

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
    
    log_success "Prerequisites check passed"
}

# Create a gist for badges
create_gist() {
    local description="$1"
    local filename="$2"
    
    log_info "Creating gist: $description"
    
    # Create temporary file
    local temp_file=$(mktemp)
    echo '{}' > "$temp_file"
    
    # Create gist and extract ID
    local gist_url=$(gh gist create --public --desc "$description" --filename "$filename" "$temp_file")
    local gist_id=$(echo "$gist_url" | sed 's/.*\///')
    
    # Cleanup
    rm "$temp_file"
    
    echo "$gist_id"
}

# Main setup function
main() {
    log_info "🎯 Setting up dynamic badges for Pivot CLI"
    echo
    
    check_prerequisites
    echo
    
    log_info "Creating GitHub gists for badge data..."
    
    # Create gists
    COVERAGE_GIST_ID=$(create_gist "Pivot CLI - Coverage Badge Data" "pivot-coverage.json")
    log_success "Coverage gist created: $COVERAGE_GIST_ID"
    
    SECURITY_GIST_ID=$(create_gist "Pivot CLI - Security Badge Data" "pivot-security.json")
    log_success "Security gist created: $SECURITY_GIST_ID"
    
    BADGES_GIST_ID=$(create_gist "Pivot CLI - General Badge Data" "pivot-build.json")
    log_success "Badges gist created: $BADGES_GIST_ID"
    
    echo
    log_info "📋 Repository Secrets Setup"
    echo "Add these secrets to your GitHub repository:"
    echo "  Settings → Secrets and variables → Actions → New repository secret"
    echo
    echo "COVERAGE_GIST_ID = $COVERAGE_GIST_ID"
    echo "SECURITY_GIST_ID = $SECURITY_GIST_ID"
    echo "BADGES_GIST_ID = $BADGES_GIST_ID"
    echo
    
    log_info "📝 README.md Update"
    echo "Replace the badge URLs in README.md with:"
    echo
    cat << EOF
[![Build Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$(gh api user --jq .login)/$BADGES_GIST_ID/raw/pivot-build.json)](https://github.com/rhino11/pivot/actions)
[![Coverage Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$(gh api user --jq .login)/$COVERAGE_GIST_ID/raw/pivot-coverage.json)](https://github.com/rhino11/pivot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/rhino11/pivot)](https://goreportcard.com/report/github.com/rhino11/pivot)
[![Security Rating](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$(gh api user --jq .login)/$SECURITY_GIST_ID/raw/pivot-security.json)](https://github.com/rhino11/pivot/security)
[![Go Version](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$(gh api user --jq .login)/$BADGES_GIST_ID/raw/pivot-go-version.json)](https://golang.org)
[![License](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/$(gh api user --jq .login)/$BADGES_GIST_ID/raw/pivot-license.json)](https://opensource.org/licenses/MIT)
EOF
    echo
    
    log_info "🔧 Automatic Setup (Optional)"
    echo "Would you like to automatically set the repository secrets? (y/N)"
    read -r response
    
    if [[ "$response" =~ ^[Yy]$ ]]; then
        log_info "Setting repository secrets..."
        
        # Check if we're in a git repository
        if ! git rev-parse --git-dir > /dev/null 2>&1; then
            log_error "Not in a git repository. Please set secrets manually."
            exit 1
        fi
        
        # Get repository info
        local repo_url=$(git remote get-url origin)
        local repo_name
        
        if [[ "$repo_url" =~ github\.com[:/]([^/]+/[^/]+) ]]; then
            repo_name="${BASH_REMATCH[1]}"
            repo_name="${repo_name%.git}"
        else
            log_error "Could not determine repository name. Please set secrets manually."
            exit 1
        fi
        
        # Set secrets
        echo "$COVERAGE_GIST_ID" | gh secret set COVERAGE_GIST_ID --repo "$repo_name"
        echo "$SECURITY_GIST_ID" | gh secret set SECURITY_GIST_ID --repo "$repo_name"
        echo "$BADGES_GIST_ID" | gh secret set BADGES_GIST_ID --repo "$repo_name"
        
        log_success "Repository secrets set successfully!"
    fi
    
    echo
    log_success "🎉 Badge setup complete!"
    log_info "Next steps:"
    echo "  1. Verify repository secrets are set"
    echo "  2. Update README.md with new badge URLs"
    echo "  3. Push changes to main branch"
    echo "  4. Wait for CI to run and badges to update"
    echo
    log_info "📚 For more details, see: docs/BADGE_SETUP.md"
}

# Run main function
main "$@"
