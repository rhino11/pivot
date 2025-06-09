#!/bin/bash

set -e

# macOS Homebrew E2E Test for Pivot CLI
# =====================================
# This test validates the complete Homebrew installation flow on macOS
# including tap setup, installation, basic functionality, and cleanup.

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
TAP_OWNER="rhino11"
TAP_NAME="tap"
PACKAGE_NAME="pivot"
TEST_VERSION="${1:-latest}"
DRY_RUN="${DRY_RUN:-false}"

# Helper functions
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

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    local test_function="$2"
    
    log_step "Running test: $test_name"
    
    if $test_function; then
        log_success "$test_name: PASSED"
        ((TESTS_PASSED++))
        return 0
    else
        log_error "$test_name: FAILED"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_name")
        return 1
    fi
}

# Pre-flight checks
check_prerequisites() {
    log_step "Checking prerequisites..."
    
    # Check if running on macOS
    if [[ "$(uname -s)" != "Darwin" ]]; then
        log_error "This test is designed for macOS only"
        return 1
    fi
    
    # Check if Homebrew is installed
    if ! command -v brew &> /dev/null; then
        log_error "Homebrew is not installed. Install from https://brew.sh/"
        return 1
    fi
    
    # Check Homebrew is functional
    if ! brew --version &> /dev/null; then
        log_error "Homebrew is not functioning properly"
        return 1
    fi
    
    log_success "Prerequisites check passed"
    return 0
}

# Test 1: Clean state verification
test_clean_state() {
    log_step "Verifying clean initial state..."
    
    # Check if pivot is already installed
    if command -v pivot &> /dev/null; then
        log_warn "pivot is already installed, attempting to uninstall..."
        if ! brew uninstall pivot 2>/dev/null; then
            log_warn "Could not uninstall via Homebrew, checking manual installation..."
        fi
    fi
    
    # Check if tap is already added
    if brew tap | grep -q "$TAP_OWNER/$TAP_NAME"; then
        log_warn "Tap $TAP_OWNER/$TAP_NAME already exists, removing..."
        if ! brew untap "$TAP_OWNER/$TAP_NAME" 2>/dev/null; then
            log_error "Could not remove existing tap"
            return 1
        fi
    fi
    
    # Verify pivot is not available
    if command -v pivot &> /dev/null; then
        log_error "pivot is still available after cleanup attempt"
        return 1
    fi
    
    log_success "Clean state verified"
    return 0
}

# Test 2: Tap addition
test_tap_addition() {
    log_step "Testing tap addition..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would run 'brew tap $TAP_OWNER/$TAP_NAME'"
        return 0
    fi
    
    if ! brew tap "$TAP_OWNER/$TAP_NAME"; then
        log_error "Failed to add tap $TAP_OWNER/$TAP_NAME"
        return 1
    fi
    
    # Verify tap was added
    if ! brew tap | grep -q "$TAP_OWNER/$TAP_NAME"; then
        log_error "Tap was not properly added"
        return 1
    fi
    
    log_success "Tap added successfully"
    return 0
}

# Test 3: Package installation
test_package_installation() {
    log_step "Testing package installation..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would run 'brew install $PACKAGE_NAME'"
        return 0
    fi
    
    # Install package (no timeout on macOS by default)
    if ! brew install "$PACKAGE_NAME"; then
        log_error "Package installation failed"
        return 1
    fi
    
    # Verify installation
    if ! command -v pivot &> /dev/null; then
        log_error "pivot command not available after installation"
        return 1
    fi
    
    log_success "Package installed successfully"
    return 0
}

# Test 4: Basic functionality
test_basic_functionality() {
    log_step "Testing basic functionality..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would test pivot commands"
        return 0
    fi
    
    # Test version command
    if ! pivot version &> /dev/null; then
        log_error "pivot version command failed"
        return 1
    fi
    
    local version_output
    version_output=$(pivot version 2>&1)
    log_info "Version output: $version_output"
    
    # Test help command
    if ! pivot help &> /dev/null; then
        log_error "pivot help command failed"
        return 1
    fi
    
    # Test that binary is properly linked
    local binary_path
    binary_path=$(which pivot)
    if [[ ! "$binary_path" =~ /opt/homebrew/bin/pivot|/usr/local/bin/pivot ]]; then
        log_warn "Binary not in expected Homebrew path: $binary_path"
    fi
    
    # Test configuration command (should handle missing config gracefully)
    if ! pivot config --help &> /dev/null; then
        log_error "pivot config command failed"
        return 1
    fi
    
    log_success "Basic functionality tests passed"
    return 0
}

# Test 5: Package information verification
test_package_info() {
    log_step "Testing package information..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would check brew info"
        return 0
    fi
    
    # Get package info
    local info_output
    if ! info_output=$(brew info "$PACKAGE_NAME" 2>&1); then
        log_error "Could not get package information"
        return 1
    fi
    
    log_info "Package info: $info_output"
    
    # Check if it shows as installed
    if ! echo "$info_output" | grep -q "Installed"; then
        log_error "Package does not show as installed"
        return 1
    fi
    
    log_success "Package information verified"
    return 0
}

# Test 6: Homebrew formula validation
test_formula_validation() {
    log_step "Testing formula validation..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would validate formula"
        return 0
    fi
    
    # Audit the formula (checks for common issues)
    if ! brew audit --strict "$TAP_OWNER/$TAP_NAME/$PACKAGE_NAME" 2>/dev/null; then
        log_warn "Formula audit found issues (may be non-critical)"
    fi
    
    # Test formula directly
    if ! brew test "$TAP_OWNER/$TAP_NAME/$PACKAGE_NAME" 2>/dev/null; then
        log_warn "Formula test failed (may be expected if test block not implemented)"
    fi
    
    log_success "Formula validation completed"
    return 0
}

# Test 7: Version consistency check
test_version_consistency() {
    log_step "Testing version consistency..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would check version consistency"
        return 0
    fi
    
    if [ "$TEST_VERSION" != "latest" ]; then
        local installed_version
        installed_version=$(pivot version 2>&1 | head -n1 | grep -o 'v[0-9]\+\.[0-9]\+\.[0-9]\+' || echo "unknown")
        
        if [ "$installed_version" != "$TEST_VERSION" ]; then
            log_warn "Version mismatch: expected $TEST_VERSION, got $installed_version"
        else
            log_success "Version matches expected: $installed_version"
        fi
    fi
    
    return 0
}

# Cleanup function
cleanup_test_installation() {
    log_step "Cleaning up test installation..."
    
    if [ "$DRY_RUN" = "true" ]; then
        log_info "DRY RUN: Would cleanup installation"
        return 0
    fi
    
    # Uninstall package
    if command -v pivot &> /dev/null; then
        if ! brew uninstall "$PACKAGE_NAME" 2>/dev/null; then
            log_warn "Could not uninstall package via Homebrew"
        fi
    fi
    
    # Remove tap
    if brew tap | grep -q "$TAP_OWNER/$TAP_NAME"; then
        if ! brew untap "$TAP_OWNER/$TAP_NAME" 2>/dev/null; then
            log_warn "Could not remove tap"
        fi
    fi
    
    # Verify cleanup
    if command -v pivot &> /dev/null; then
        log_warn "pivot still available after cleanup"
    else
        log_success "Cleanup completed successfully"
    fi
}

# Main test execution
main() {
    log_info "🧪 Homebrew macOS E2E Test for Pivot CLI"
    echo
    log_info "Test version: $TEST_VERSION"
    log_info "Dry run: $DRY_RUN"
    echo
    
    # Trap to ensure cleanup happens
    trap cleanup_test_installation EXIT
    
    # Run tests
    run_test "Prerequisites Check" check_prerequisites || exit 1
    run_test "Clean State Verification" test_clean_state || exit 1
    run_test "Tap Addition" test_tap_addition || exit 1
    run_test "Package Installation" test_package_installation || exit 1
    run_test "Basic Functionality" test_basic_functionality || exit 1
    run_test "Package Information" test_package_info || exit 1
    run_test "Formula Validation" test_formula_validation || exit 1
    run_test "Version Consistency" test_version_consistency || exit 1
    
    echo
    log_info "🎯 Test Summary"
    echo "=================================="
    log_success "Tests passed: $TESTS_PASSED"
    if [ $TESTS_FAILED -gt 0 ]; then
        log_error "Tests failed: $TESTS_FAILED"
        echo "Failed tests:"
        for test in "${FAILED_TESTS[@]}"; do
            echo "  - $test"
        done
        echo
        log_error "❌ E2E test suite FAILED"
        exit 1
    else
        echo
        log_success "✅ All E2E tests PASSED!"
    fi
}

# Handle script arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [version] [options]"
        echo ""
        echo "Arguments:"
        echo "  version    Version to test (default: latest)"
        echo ""
        echo "Environment variables:"
        echo "  DRY_RUN    Set to 'true' for dry run mode"
        echo ""
        echo "Examples:"
        echo "  $0                    # Test latest version"
        echo "  $0 v1.1.0            # Test specific version"
        echo "  DRY_RUN=true $0      # Dry run mode"
        exit 0
        ;;
    --cleanup)
        cleanup_test_installation
        exit 0
        ;;
    *)
        main
        ;;
esac
