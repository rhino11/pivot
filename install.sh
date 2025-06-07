#!/bin/bash

# Pivot CLI Installer Script
# This script downloads and installs the latest version of Pivot CLI

set -e

# Configuration
REPO="rhino11/pivot"
BINARY_NAME="pivot"
INSTALL_DIR="/usr/local/bin"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

log_success() {
    echo -e "${GREEN}✓${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

log_error() {
    echo -e "${RED}✗${NC} $1"
    exit 1
}

# Detect OS and architecture
detect_platform() {
    local os=""
    local arch=""
    
    case "$(uname -s)" in
        Linux*)     os="linux";;
        Darwin*)    os="darwin";;
        MINGW*)     os="windows";;
        *)          log_error "Unsupported operating system: $(uname -s)";;
    esac
    
    case "$(uname -m)" in
        x86_64)     arch="amd64";;
        arm64)      arch="arm64";;
        aarch64)    arch="arm64";;
        *)          log_error "Unsupported architecture: $(uname -m)";;
    esac
    
    echo "${os}-${arch}"
}

# Get the latest release version
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        log_error "Failed to get latest version"
    fi
    
    echo "$version"
}

# Download and install
install_pivot() {
    local platform="$1"
    local version="$2"
    local download_url="https://github.com/${REPO}/releases/download/${version}/${BINARY_NAME}-${platform}"
    local tmp_file="/tmp/${BINARY_NAME}"
    
    # Add .exe extension for Windows
    if [[ "$platform" == *"windows"* ]]; then
        download_url="${download_url}.exe"
        tmp_file="${tmp_file}.exe"
    fi
    
    log_info "Downloading Pivot ${version} for ${platform}..."
    
    if ! curl -L -o "$tmp_file" "$download_url"; then
        log_error "Failed to download Pivot"
    fi
    
    # Make executable
    chmod +x "$tmp_file"
    
    # Check if we need sudo for installation
    if [ ! -w "$INSTALL_DIR" ]; then
        log_info "Installing to ${INSTALL_DIR} (requires sudo)..."
        sudo mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    else
        log_info "Installing to ${INSTALL_DIR}..."
        mv "$tmp_file" "${INSTALL_DIR}/${BINARY_NAME}"
    fi
    
    log_success "Pivot installed successfully!"
}

# Verify installation
verify_installation() {
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        local installed_version
        installed_version=$($BINARY_NAME version 2>/dev/null | head -n1 | awk '{print $2}' || echo "unknown")
        log_success "Pivot is installed and ready to use (version: $installed_version)"
        log_info "Run 'pivot --help' to get started"
    else
        log_warning "Pivot was installed but is not in your PATH"
        log_info "You may need to restart your shell or add ${INSTALL_DIR} to your PATH"
    fi
}

# Main installation process
main() {
    echo "Pivot CLI Installer"
    echo "==================="
    echo
    
    # Check if curl is available
    if ! command -v curl >/dev/null 2>&1; then
        log_error "curl is required but not installed"
    fi
    
    # Detect platform
    local platform
    platform=$(detect_platform)
    log_info "Detected platform: $platform"
    
    # Get latest version
    local version
    version=$(get_latest_version)
    log_info "Latest version: $version"
    
    # Install
    install_pivot "$platform" "$version"
    
    # Verify
    verify_installation
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Install the latest version of Pivot CLI"
        echo
        echo "Options:"
        echo "  --help, -h    Show this help message"
        echo
        echo "This script will:"
        echo "  1. Detect your operating system and architecture"
        echo "  2. Download the latest Pivot release"
        echo "  3. Install it to /usr/local/bin (may require sudo)"
        echo
        exit 0
        ;;
    *)
        main "$@"
        ;;
esac
