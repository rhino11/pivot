#!/bin/bash

# Homebrew E2E Test Runner for Local Development
# ==============================================
# This script helps developers run the Homebrew E2E test locally
# with various options and configurations.

set -e

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_step() {
    echo -e "${BLUE}🔧 $1${NC}"
}

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
E2E_SCRIPT="$PROJECT_ROOT/test/e2e/homebrew-macos-e2e.sh"

# Default values
MODE="test"
VERSION="latest"
DRY_RUN="false"

show_help() {
    cat << EOF
Homebrew E2E Test Runner

Usage: $0 [options] [version]

Options:
    -h, --help          Show this help message
    -d, --dry-run       Run in dry-run mode (no actual installation)
    -c, --cleanup       Only run cleanup (remove test installations)
    -v, --verbose       Enable verbose output
    --local-build       Test with locally built binary instead of released version
    --pre-release       Test with latest pre-release version

Arguments:
    version             Specific version to test (e.g., v1.1.0)
                       Default: latest

Examples:
    $0                           # Test latest release
    $0 v1.1.0                    # Test specific version
    $0 --dry-run                 # Dry run with latest
    $0 --dry-run v1.1.0          # Dry run with specific version
    $0 --cleanup                 # Clean up previous test installations
    $0 --local-build             # Test with locally built binary

Environment Variables:
    HOMEBREW_E2E_VERBOSE=1       Enable detailed logging
    HOMEBREW_E2E_NO_CLEANUP=1    Skip cleanup after test
    HOMEBREW_E2E_FORCE=1         Force test even if prerequisites fail

EOF
}

check_macos() {
    if [[ "$(uname -s)" != "Darwin" ]]; then
        log_warn "This test is designed for macOS only"
        log_info "Current OS: $(uname -s)"
        log_info "For cross-platform testing, use: ./scripts/test-installation.sh"
        exit 1
    fi
}

check_homebrew() {
    if ! command -v brew &> /dev/null; then
        log_warn "Homebrew is not installed"
        log_info "Install Homebrew from: https://brew.sh/"
        log_info "Then run: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        exit 1
    fi
}

test_local_build() {
    log_step "Testing with local build..."
    
    # Build the project
    cd "$PROJECT_ROOT"
    if ! make build; then
        log_warn "Local build failed"
        exit 1
    fi
    
    # Create a temporary tap for testing
    local temp_tap_dir="/tmp/homebrew-test-tap-$$"
    mkdir -p "$temp_tap_dir/Formula"
    
    # Create a test formula pointing to local binary
    cat > "$temp_tap_dir/Formula/pivot.rb" << EOF
class Pivot < Formula
  desc "GitHub Issues Management CLI (Local Test Build)"
  homepage "https://github.com/rhino11/pivot"
  version "dev"
  
  def install
    bin.install "$PROJECT_ROOT/build/pivot"
  end
  
  test do
    assert_match "pivot", shell_output("#{bin}/pivot version")
  end
end
EOF
    
    # Add temporary tap
    brew tap-new local/test-pivot
    local tap_path="$(brew --repository)/Library/Taps/local/homebrew-test-pivot"
    cp "$temp_tap_dir/Formula/pivot.rb" "$tap_path/Formula/"
    
    # Install from local tap
    if brew install local/test-pivot/pivot; then
        log_info "✅ Local build test successful"
        
        # Test basic functionality
        pivot version
        pivot help
        
        # Cleanup
        brew uninstall local/test-pivot/pivot
        brew untap local/test-pivot
    else
        log_warn "❌ Local build test failed"
        brew untap local/test-pivot 2>/dev/null || true
        exit 1
    fi
    
    rm -rf "$temp_tap_dir"
}

run_pre_release_test() {
    log_step "Testing with latest pre-release..."
    
    # Get latest pre-release from GitHub API
    local pre_release_info
    pre_release_info=$(curl -s "https://api.github.com/repos/rhino11/pivot/releases" | \
                       jq -r '.[] | select(.prerelease == true) | .tag_name' | head -n1)
    
    if [ -z "$pre_release_info" ] || [ "$pre_release_info" = "null" ]; then
        log_warn "No pre-release versions found"
        exit 1
    fi
    
    log_info "Testing pre-release version: $pre_release_info"
    VERSION="$pre_release_info"
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -h|--help)
            show_help
            exit 0
            ;;
        -d|--dry-run)
            DRY_RUN="true"
            shift
            ;;
        -c|--cleanup)
            MODE="cleanup"
            shift
            ;;
        -v|--verbose)
            export HOMEBREW_E2E_VERBOSE=1
            shift
            ;;
        --local-build)
            MODE="local-build"
            shift
            ;;
        --pre-release)
            MODE="pre-release"
            shift
            ;;
        -*)
            log_warn "Unknown option: $1"
            show_help
            exit 1
            ;;
        *)
            VERSION="$1"
            shift
            ;;
    esac
done

main() {
    log_info "🧪 Homebrew E2E Test Runner"
    echo
    log_info "Mode: $MODE"
    log_info "Version: $VERSION"
    log_info "Dry run: $DRY_RUN"
    echo
    
    # Pre-flight checks
    check_macos
    
    case $MODE in
        cleanup)
            log_step "Running cleanup only..."
            if [ -f "$E2E_SCRIPT" ]; then
                "$E2E_SCRIPT" --cleanup
            else
                log_warn "E2E script not found: $E2E_SCRIPT"
                exit 1
            fi
            ;;
        local-build)
            check_homebrew
            test_local_build
            ;;
        pre-release)
            check_homebrew
            run_pre_release_test
            export DRY_RUN="$DRY_RUN"
            "$E2E_SCRIPT" "$VERSION"
            ;;
        test)
            check_homebrew
            export DRY_RUN="$DRY_RUN"
            "$E2E_SCRIPT" "$VERSION"
            ;;
        *)
            log_warn "Unknown mode: $MODE"
            exit 1
            ;;
    esac
    
    log_info "✅ Test runner completed"
}

# Check if jq is available for pre-release testing
if [[ "$MODE" == "pre-release" ]] && ! command -v jq &> /dev/null; then
    log_warn "jq is required for pre-release testing"
    log_info "Install with: brew install jq"
    exit 1
fi

main
