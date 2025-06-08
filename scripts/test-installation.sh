#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Test configuration
TAP_OWNER="rhino11"
TAP_NAME="tap"
PACKAGE_NAME="pivot"
EXPECTED_COMMANDS=("version" "help" "init")

log_info "🧪 Testing Pivot CLI installation methods"
echo

# Check if version argument provided
if [ $# -eq 0 ]; then
    log_warn "No version specified. Testing with latest available version."
    VERSION=""
else
    VERSION=$1
    log_info "Testing version: $VERSION"
fi

echo

# Function to test command availability
test_command() {
    local cmd=$1
    local description=$2
    
    log_step "Testing: $description"
    if command -v "$cmd" &> /dev/null; then
        log_success "$cmd is available"
        
        # Test basic functionality
        log_step "Testing basic functionality..."
        if "$cmd" version &> /dev/null; then
            log_success "Version command works"
            "$cmd" version
        else
            log_warn "Version command failed"
        fi
        
        if "$cmd" help &> /dev/null; then
            log_success "Help command works"
        else
            log_warn "Help command failed"
        fi
        
        return 0
    else
        log_error "$cmd is not available"
        return 1
    fi
}

# Function to test Homebrew installation
test_homebrew() {
    log_step "Testing Homebrew installation..."
    
    if ! command -v brew &> /dev/null; then
        log_warn "Homebrew not available, skipping test"
        return 1
    fi
    
    log_step "Checking if tap is available..."
    if brew tap | grep -q "$TAP_OWNER/$TAP_NAME"; then
        log_success "Tap is already added"
    else
        log_step "Adding tap..."
        if brew tap "$TAP_OWNER/$TAP_NAME"; then
            log_success "Tap added successfully"
        else
            log_error "Failed to add tap"
            return 1
        fi
    fi
    
    log_step "Installing pivot via Homebrew..."
    if brew install "$PACKAGE_NAME"; then
        log_success "Homebrew installation successful"
        test_command "pivot" "Homebrew installed pivot"
        return $?
    else
        log_error "Homebrew installation failed"
        return 1
    fi
}

# Function to test direct download
test_direct_download() {
    log_step "Testing direct download installation..."
    
    # Detect platform
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    ARCH=$(uname -m)
    
    case $ARCH in
        x86_64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            log_warn "Unsupported architecture: $ARCH"
            return 1
            ;;
    esac
    
    case $OS in
        darwin)
            PLATFORM="darwin"
            EXT=""
            ;;
        linux)
            PLATFORM="linux"
            EXT=""
            ;;
        mingw*|msys*|cygwin*)
            PLATFORM="windows"
            EXT=".exe"
            ;;
        *)
            log_warn "Unsupported OS: $OS"
            return 1
            ;;
    esac
    
    BINARY_NAME="pivot-${PLATFORM}-${ARCH}${EXT}"
    
    if [ -z "$VERSION" ]; then
        # Get latest version from GitHub API
        log_step "Getting latest version from GitHub..."
        VERSION=$(curl -s https://api.github.com/repos/rhino11/pivot/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
        if [ -z "$VERSION" ]; then
            log_error "Could not determine latest version"
            return 1
        fi
        log_info "Latest version: $VERSION"
    fi
    
    DOWNLOAD_URL="https://github.com/rhino11/pivot/releases/download/$VERSION/$BINARY_NAME"
    TEMP_DIR=$(mktemp -d)
    LOCAL_BINARY="$TEMP_DIR/pivot${EXT}"
    
    log_step "Downloading from: $DOWNLOAD_URL"
    if curl -L -o "$LOCAL_BINARY" "$DOWNLOAD_URL"; then
        log_success "Download successful"
        chmod +x "$LOCAL_BINARY"
        
        log_step "Testing downloaded binary..."
        if "$LOCAL_BINARY" version &> /dev/null; then
            log_success "Direct download binary works"
            "$LOCAL_BINARY" version
            rm -rf "$TEMP_DIR"
            return 0
        else
            log_error "Downloaded binary failed to run"
            rm -rf "$TEMP_DIR"
            return 1
        fi
    else
        log_error "Download failed"
        rm -rf "$TEMP_DIR"
        return 1
    fi
}

# Function to cleanup test installations
cleanup_installations() {
    log_step "Cleaning up test installations..."
    
    # Remove Homebrew installation
    if command -v brew &> /dev/null; then
        if brew list | grep -q "^pivot$"; then
            log_step "Removing Homebrew installation..."
            brew uninstall pivot || log_warn "Failed to uninstall via Homebrew"
        fi
        
        if brew tap | grep -q "$TAP_OWNER/$TAP_NAME"; then
            log_step "Removing tap..."
            brew untap "$TAP_OWNER/$TAP_NAME" || log_warn "Failed to remove tap"
        fi
    fi
    
    log_success "Cleanup completed"
}

# Main testing flow
main() {
    local test_results=()
    
    echo "=== Testing Installation Methods ==="
    echo
    
    # Test Homebrew
    if test_homebrew; then
        test_results+=("✅ Homebrew: PASS")
    else
        test_results+=("❌ Homebrew: FAIL")
    fi
    
    echo
    echo "=== Testing Direct Download ==="
    echo
    
    # Test direct download
    if test_direct_download; then
        test_results+=("✅ Direct Download: PASS")
    else
        test_results+=("❌ Direct Download: FAIL")
    fi
    
    echo
    echo "=== Test Results Summary ==="
    echo
    
    for result in "${test_results[@]}"; do
        echo "$result"
    done
    
    echo
    
    # Cleanup
    read -p "Do you want to clean up test installations? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cleanup_installations
    fi
    
    log_info "🎉 Testing completed!"
}

# Handle script arguments
case "${1:-}" in
    --cleanup)
        cleanup_installations
        exit 0
        ;;
    --help|-h)
        echo "Usage: $0 [version] [--cleanup] [--help]"
        echo
        echo "Options:"
        echo "  version     Test specific version (default: latest)"
        echo "  --cleanup   Only run cleanup"
        echo "  --help      Show this help"
        exit 0
        ;;
    *)
        main
        ;;
esac
