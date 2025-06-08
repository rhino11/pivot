#!/bin/bash

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

echo "🧪 Testing Pivot installation methods..."

# Test 1: Direct binary download
echo
echo "Testing direct binary installation..."
if curl -sSL https://raw.githubusercontent.com/rhino11/pivot/main/install.sh | bash -s -- --test; then
    log_info "Direct installation works"
else
    log_error "Direct installation failed"
fi

# Test 2: Homebrew (if available)
if command -v brew &> /dev/null; then
    echo
    echo "Testing Homebrew installation..."
    if brew tap rhino11/tap && brew install pivot; then
        log_info "Homebrew installation works"
        pivot version
        brew uninstall pivot
        brew untap rhino11/tap
    else
        log_error "Homebrew installation failed"
    fi
fi

# Test 3: Manual download and verify
echo
echo "Testing manual download..."
VERSION=$(curl -s https://api.github.com/repos/rhino11/pivot/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
PLATFORM="linux-amd64"

if [[ "$OSTYPE" == "darwin"* ]]; then
    PLATFORM="darwin-amd64"
    if [[ "$(uname -m)" == "arm64" ]]; then
        PLATFORM="darwin-arm64"
    fi
elif [[ "$OSTYPE" == "msys" ]]; then
    PLATFORM="windows-amd64.exe"
fi

DOWNLOAD_URL="https://github.com/rhino11/pivot/releases/download/$VERSION/pivot-$PLATFORM"

if curl -L -o /tmp/pivot-test "$DOWNLOAD_URL" && chmod +x /tmp/pivot-test; then
    if /tmp/pivot-test version; then
        log_info "Manual download works"
    else
        log_error "Downloaded binary doesn't work"
    fi
    rm -f /tmp/pivot-test
else
    log_error "Manual download failed"
fi

echo
echo "🎉 Installation testing complete!"