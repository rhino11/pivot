#!/bin/bash

# Script to manually update the homebrew-tap repository
# Usage: ./scripts/update-homebrew-tap.sh [version] [sha256_amd64] [sha256_arm64]
#   OR:  ./scripts/update-homebrew-tap.sh (uses latest release and downloads checksums)

set -euo pipefail

# Configuration
TAP_REPO="rhino11/homebrew-tap"
MAIN_REPO="rhino11/pivot"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

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

# Parse arguments
if [ $# -eq 3 ]; then
    # All parameters provided
    VERSION="$1"
    SHA256_AMD64="$2"
    SHA256_ARM64="$3"
    VERSION_CLEAN="${VERSION#v}"
    
    if [[ ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format. Expected: v1.0.1 or 1.0.1"
        exit 1
    fi
    
    log_info "🍺 Using provided values:"
    echo "   📦 Version: ${VERSION}"
    echo "   🔐 AMD64 SHA256: ${SHA256_AMD64}"
    echo "   🔐 ARM64 SHA256: ${SHA256_ARM64}"
    
elif [ $# -eq 1 ]; then
    # Only version provided, fetch checksums
    VERSION="$1"
    VERSION_CLEAN="${VERSION#v}"
    
    if [[ ! "$VERSION" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        log_error "Invalid version format. Expected: v1.0.1 or 1.0.1"
        exit 1
    fi
    
    log_step "Fetching checksums for version ${VERSION}..."
    CHECKSUMS_URL="https://github.com/${MAIN_REPO}/releases/download/${VERSION}/checksums.txt"
    
    if ! CHECKSUMS=$(curl -s -f "$CHECKSUMS_URL"); then
        log_error "Failed to fetch checksums from ${CHECKSUMS_URL}"
        log_info "💡 Make sure the release exists and includes checksums.txt"
        exit 1
    fi
    
    SHA256_AMD64=$(echo "$CHECKSUMS" | grep "pivot-darwin-amd64" | cut -d' ' -f1)
    SHA256_ARM64=$(echo "$CHECKSUMS" | grep "pivot-darwin-arm64" | cut -d' ' -f1)
    
    if [ -z "$SHA256_AMD64" ] || [ -z "$SHA256_ARM64" ]; then
        log_error "Could not extract SHA256 values from checksums"
        log_info "Available checksums:"
        echo "$CHECKSUMS"
        exit 1
    fi
    
    log_info "🍺 Fetched checksums:"
    echo "   📦 Version: ${VERSION}"
    echo "   🔐 AMD64 SHA256: ${SHA256_AMD64}"
    echo "   🔐 ARM64 SHA256: ${SHA256_ARM64}"
    
elif [ $# -eq 0 ]; then
    # No parameters, use latest release
    log_info "No version specified, fetching latest release..."
    VERSION=$(curl -s https://api.github.com/repos/$MAIN_REPO/releases/latest | jq -r '.tag_name')
    VERSION_CLEAN="${VERSION#v}"
    
    log_step "Fetching checksums for latest version ${VERSION}..."
    CHECKSUMS_URL="https://github.com/${MAIN_REPO}/releases/download/${VERSION}/checksums.txt"
    
    if ! CHECKSUMS=$(curl -s -f "$CHECKSUMS_URL"); then
        log_error "Failed to fetch checksums from ${CHECKSUMS_URL}"
        exit 1
    fi
    
    SHA256_AMD64=$(echo "$CHECKSUMS" | grep "pivot-darwin-amd64" | cut -d' ' -f1)
    SHA256_ARM64=$(echo "$CHECKSUMS" | grep "pivot-darwin-arm64" | cut -d' ' -f1)
    
    log_info "🍺 Using latest release:"
    echo "   📦 Version: ${VERSION}"
    echo "   🔐 AMD64 SHA256: ${SHA256_AMD64}"
    echo "   🔐 ARM64 SHA256: ${SHA256_ARM64}"
    
else
    echo "Usage: $0 [version] [sha256_amd64] [sha256_arm64]"
    echo "   OR: $0 [version]  (fetches checksums automatically)"
    echo "   OR: $0           (uses latest release)"
    echo
    echo "Examples:"
    echo "  $0 v1.0.1 abc123... def456...  # Specify all values"
    echo "  $0 v1.0.1                     # Fetch checksums automatically"
    echo "  $0                            # Use latest release"
    exit 1
fi

echo

# Create temporary directory
TEMP_DIR=$(mktemp -d)
log_step "Working in temporary directory: ${TEMP_DIR}"

# Clone the homebrew-tap repository
log_step "Cloning homebrew-tap repository..."
if ! git clone "https://github.com/$TAP_REPO.git" "${TEMP_DIR}/homebrew-tap"; then
    log_error "Failed to clone homebrew-tap repository"
    log_info "Make sure the repository exists: https://github.com/$TAP_REPO"
    rm -rf "$TEMP_DIR"
    exit 1
fi

cd "${TEMP_DIR}/homebrew-tap"

# Verify we're on the right branch
git checkout main

# Create the updated formula
log_step "Creating updated formula for version ${VERSION_CLEAN}..."
cat > Formula/pivot.rb << EOF
class Pivot < Formula
  desc "GitHub Issues Management CLI"
  homepage "https://github.com/$MAIN_REPO"
  version "$VERSION_CLEAN"
  
  if Hardware::CPU.arm?
    url "https://github.com/$MAIN_REPO/releases/download/$VERSION/pivot-darwin-arm64"
    sha256 "$SHA256_ARM64"
  else
    url "https://github.com/$MAIN_REPO/releases/download/$VERSION/pivot-darwin-amd64"
    sha256 "$SHA256_AMD64"
  end
  
  def install
    bin.install Dir["pivot-darwin-*"].first => "pivot"
  end
  
  test do
    assert_match "pivot", shell_output("#{bin}/pivot version")
  end
end
EOF

# Check if there are changes
if git diff --quiet Formula/pivot.rb; then
    log_warning "No changes detected in the formula. Already up to date?"
    cd /
    rm -rf "$TEMP_DIR"
    exit 0
fi

# Show the changes
log_step "Formula updated. Changes:"
echo "----------------------------------------"
git diff Formula/pivot.rb
echo "----------------------------------------"
echo

# Commit the changes
log_step "Committing changes..."
git config --local user.email "action@github.com"
git config --local user.name "GitHub Action"
git add Formula/pivot.rb
git commit -m "Update pivot to $VERSION

- Updated version to $VERSION_CLEAN
- Updated SHA256 hashes for macOS binaries
- AMD64: $SHA256_AMD64
- ARM64: $SHA256_ARM64"

# Push the changes
log_step "Pushing changes to GitHub..."
if git push origin main; then
    log_success "Successfully updated homebrew-tap repository!"
    echo
    log_info "🍺 Users can now install the updated version with:"
    echo "   brew tap rhino11/tap"
    echo "   brew install pivot"
    echo
    log_info "🔄 Or upgrade existing installations with:"
    echo "   brew upgrade pivot"
else
    log_error "Failed to push changes to GitHub"
    log_warning "You may need to authenticate or check repository permissions"
    cd /
    rm -rf "$TEMP_DIR"
    exit 1
fi

# Cleanup
cd /
rm -rf "$TEMP_DIR"

log_success "🎉 Homebrew tap update completed successfully!"
