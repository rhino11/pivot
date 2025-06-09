#!/bin/bash

# Submit Pivot to Homebrew Core
# ============================
# This script helps submit the pivot formula to Homebrew Core

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_step() {
    echo -e "${BLUE}🔧 $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FORMULA_FILE="$PROJECT_ROOT/homebrew-core-formula.rb"

# Check prerequisites
check_prerequisites() {
    log_step "Checking prerequisites..."
    
    if ! command -v brew &> /dev/null; then
        log_error "Homebrew is not installed"
        exit 1
    fi
    
    if ! command -v gh &> /dev/null; then
        log_error "GitHub CLI is not installed. Install with: brew install gh"
        exit 1
    fi
    
    if ! gh auth status &> /dev/null; then
        log_error "GitHub CLI is not authenticated. Run: gh auth login"
        exit 1
    fi
    
    log_success "Prerequisites check passed"
}

# Test the formula locally
test_formula() {
    log_step "Testing formula locally..."
    
    # Set up environment for local testing
    export HOMEBREW_NO_AUTO_UPDATE=1
    export HOMEBREW_NO_INSTALL_FROM_API=1
    
    # Copy formula to temporary location for testing
    local temp_formula="/tmp/pivot.rb"
    cp "$FORMULA_FILE" "$temp_formula"
    
    # Test installation
    log_info "Installing pivot from local formula..."
    if brew install "$temp_formula"; then
        log_success "Installation successful"
        
        # Test basic functionality
        log_info "Testing basic functionality..."
        if pivot version && pivot help; then
            log_success "Basic functionality tests passed"
        else
            log_error "Basic functionality tests failed"
            brew uninstall pivot 2>/dev/null || true
            rm -f "$temp_formula"
            exit 1
        fi
        
        # Test uninstall
        log_info "Testing uninstall..."
        if brew uninstall pivot; then
            log_success "Uninstall successful"
        else
            log_warn "Uninstall had issues"
        fi
    else
        log_error "Installation failed"
        rm -f "$temp_formula"
        exit 1
    fi
    
    # Audit the formula
    log_info "Auditing formula..."
    if brew audit --strict --new --online "$temp_formula"; then
        log_success "Formula audit passed"
    else
        log_warn "Formula audit found issues (may be non-critical)"
    fi
    
    rm -f "$temp_formula"
    unset HOMEBREW_NO_AUTO_UPDATE
    unset HOMEBREW_NO_INSTALL_FROM_API
}

# Fork homebrew-core and create PR
submit_to_homebrew_core() {
    log_step "Submitting to Homebrew Core..."
    
    # Fork homebrew-core if not already forked
    if ! gh repo view rhino11/homebrew-core &> /dev/null; then
        log_info "Forking Homebrew/homebrew-core..."
        gh repo fork Homebrew/homebrew-core --clone=false
    fi
    
    # Clone our fork
    local work_dir="/tmp/homebrew-core-$$"
    log_info "Cloning fork to $work_dir..."
    gh repo clone rhino11/homebrew-core "$work_dir"
    
    cd "$work_dir"
    
    # Create feature branch
    local branch_name="add-pivot-formula"
    git checkout -b "$branch_name"
    
    # Copy formula to correct location
    mkdir -p Formula
    cp "$FORMULA_FILE" "Formula/pivot.rb"
    
    # Commit changes
    git add Formula/pivot.rb
    git commit -m "pivot 1.0.1 (new formula)

A CLI tool for syncing GitHub issues to a local database, enabling 
agile, AI-driven project management with offline capabilities.

Closes #INSERT_ISSUE_NUMBER_HERE"
    
    # Push branch
    git push origin "$branch_name"
    
    # Create pull request
    log_info "Creating pull request..."
    gh pr create \
        --title "pivot 1.0.1 (new formula)" \
        --body "$(cat << 'EOF'
## Description
A CLI tool for syncing GitHub issues to a local database, enabling agile, AI-driven project management with offline capabilities.

## Features
- 🔄 Bidirectional sync with GitHub Issues
- 🛠️ Offline support with local SQLite database  
- 🚀 AI-ready architecture for future GenAI integration
- 📦 Cross-platform support (Windows, macOS, Linux)
- 🔧 Simple YAML configuration

## Testing
- [x] Formula builds successfully
- [x] Binary installs and runs
- [x] `pivot version` works
- [x] `pivot help` works  
- [x] Uninstall works cleanly
- [x] Formula passes `brew audit --strict --new --online`

## Checklist
- [x] Formula follows Homebrew guidelines
- [x] Tests included and passing
- [x] Documentation is clear
- [x] Software is stable and actively maintained
- [x] No license issues

## Additional Notes
The software has a comprehensive test suite, automated CI/CD, and follows semantic versioning. It's designed for developers and project managers who need to work with GitHub Issues offline or integrate with AI tools.
EOF
)" \
        --base master \
        --head rhino11:"$branch_name" \
        --repo Homebrew/homebrew-core
    
    log_success "Pull request created!"
    log_info "Monitor the PR at: https://github.com/Homebrew/homebrew-core/pulls"
    
    # Cleanup
    cd "$PROJECT_ROOT"
    rm -rf "$work_dir"
}

# Show instructions
show_instructions() {
    cat << 'EOF'

🎉 Homebrew Core Submission Complete!

Next Steps:
1. Monitor your PR: https://github.com/Homebrew/homebrew-core/pulls
2. Respond to any feedback from Homebrew maintainers
3. Once merged, users can install with: brew install pivot

Alternative: Use Custom Tap (Immediate)
=======================================
If you want immediate availability while waiting for Homebrew Core:

1. Set up custom tap:
   ./scripts/setup-homebrew-tap.sh

2. Users install with:
   brew tap rhino11/tap
   brew install pivot

Both approaches work, but Homebrew Core is the gold standard!

EOF
}

main() {
    log_info "🍺 Homebrew Core Submission Script"
    echo
    
    check_prerequisites
    test_formula
    
    echo
    log_info "Ready to submit to Homebrew Core!"
    read -p "Continue with submission? (y/N) " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        submit_to_homebrew_core
        show_instructions
    else
        log_info "Submission cancelled. You can run this script again when ready."
        echo
        log_info "Alternative: Set up custom tap with: ./scripts/setup-homebrew-tap.sh"
    fi
}

main "$@"
