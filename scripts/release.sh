#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Check if version argument provided
if [ $# -eq 0 ]; then
    log_error "Usage: $0 <version>"
    echo "Example: $0 v1.1.0"
    exit 1
fi

VERSION=$1

# Validate version format
if [[ ! $VERSION =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    log_error "Version must be in format vX.Y.Z (e.g., v1.1.0)"
    exit 1
fi

log_info "🚀 Starting release process for Pivot $VERSION"

# Pre-release checks
log_info "Running pre-release checks..."

# Check if we're on main branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [ "$CURRENT_BRANCH" != "main" ]; then
    log_warn "You're not on the main branch (currently on: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_error "Release cancelled"
        exit 1
    fi
fi

# Check if working directory is clean
if ! git diff-index --quiet HEAD --; then
    log_error "Working directory is not clean. Commit or stash changes first."
    exit 1
fi

# Check if tag already exists
if git rev-parse "$VERSION" >/dev/null 2>&1; then
    log_error "Tag $VERSION already exists"
    exit 1
fi

# Run tests
log_info "Running tests..."
if ! make test; then
    log_error "Tests failed"
    exit 1
fi

# Run linter
log_info "Running linter..."
if ! make lint; then
    log_error "Linting failed"
    exit 1
fi

# Build all platforms to verify
log_info "Building all platforms..."
if ! make build-all; then
    log_error "Build failed"
    exit 1
fi

# Create GitHub release
log_info "Creating GitHub release..."

# Generate release notes
PREVIOUS_TAG=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
if [ -n "$PREVIOUS_TAG" ]; then
    RELEASE_NOTES=$(git log --pretty=format:"- %s" $PREVIOUS_TAG..HEAD)
else
    RELEASE_NOTES="Initial release"
fi

# Create annotated tag
git tag -a $VERSION -m "Release $VERSION

$RELEASE_NOTES"

# Push the tag
log_info "Pushing tag to trigger automated release..."
git push origin $VERSION

# Wait for GitHub Actions to complete
log_info "GitHub Actions will now build and create the release"
log_info "Monitor progress: https://github.com/rhino11/pivot/actions"

# Check if homebrew-tap repository exists
log_info "Checking Homebrew tap repository..."

TAP_REPO="rhino11/homebrew-tap"
TAP_URL="https://github.com/$TAP_REPO"

if ! curl -s -f -I "$TAP_URL" >/dev/null 2>&1; then
    log_warn "Homebrew tap repository doesn't exist: $TAP_URL"
    echo
    log_info "To set up the Homebrew tap, run:"
    echo "  ./scripts/setup-homebrew-tap.sh"
    echo
    log_info "This will:"
    echo "- Create the homebrew-tap repository"
    echo "- Set up the initial formula"
    echo "- Configure automated updates"
    echo
else
    log_info "✅ Homebrew tap repository exists"
    
    # Check if we can auto-update the tap
    read -p "Would you like to auto-update the Homebrew tap? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        chmod +x ./scripts/setup-homebrew-tap.sh
        log_info "Setting up Homebrew tap for version $VERSION..."
        ./scripts/setup-homebrew-tap.sh $VERSION
    fi
fi

log_info "✅ Release $VERSION initiated successfully!"
echo
echo "Next steps:"
echo "- Monitor GitHub Actions: https://github.com/rhino11/pivot/actions"
echo "- Check release: https://github.com/rhino11/pivot/releases/tag/$VERSION"
echo "- Monitor Homebrew tap update: https://github.com/rhino11/homebrew-tap/actions"
echo
echo "Available scripts:"
echo "- Test installation methods: ./scripts/test-installation.sh $VERSION"
echo "- Set up Homebrew tap (if needed): ./scripts/setup-homebrew-tap.sh"
echo
echo "Installation commands for users:"
echo "- Homebrew: brew install rhino11/tap/pivot"
echo "- Direct download: Available from GitHub releases"

# Function to setup Homebrew tap
setup_homebrew_tap() {
    local version=$1
    
    log_info "Setting up Homebrew tap via GitHub API..."
    
    # Check if GitHub CLI is available
    if ! command -v gh &> /dev/null; then
        log_warn "GitHub CLI not available. Using manual method..."
        setup_homebrew_tap_manual "$version"
        return
    fi
    
    # Check if authenticated
    if ! gh auth status &> /dev/null; then
        log_warn "Not authenticated with GitHub CLI. Using manual method..."
        setup_homebrew_tap_manual "$version"
        return
    fi
    
    # Wait for GitHub release to be available and get SHA256 hashes
    log_info "Waiting for GitHub release to be available..."
    local max_attempts=30
    local attempt=1
    local sha256_amd64=""
    local sha256_arm64=""
    
    while [ $attempt -le $max_attempts ]; do
        log_step "Attempt $attempt/$max_attempts: Checking release assets..."
        
        # Try to get the SHA256 hashes from the release assets
        if curl -s -f "https://github.com/rhino11/pivot/releases/download/$version/pivot-darwin-amd64" >/dev/null 2>&1; then
            # Download and calculate SHA256 hashes
            local temp_dir=$(mktemp -d)
            cd "$temp_dir"
            
            if curl -L -o "pivot-darwin-amd64" "https://github.com/rhino11/pivot/releases/download/$version/pivot-darwin-amd64" && \
               curl -L -o "pivot-darwin-arm64" "https://github.com/rhino11/pivot/releases/download/$version/pivot-darwin-arm64"; then
                
                sha256_amd64=$(sha256sum pivot-darwin-amd64 | cut -d' ' -f1)
                sha256_arm64=$(sha256sum pivot-darwin-arm64 | cut -d' ' -f1)
                
                log_success "Got SHA256 hashes:"
                log_info "AMD64: $sha256_amd64"
                log_info "ARM64: $sha256_arm64"
                
                cd - >/dev/null
                rm -rf "$temp_dir"
                break
            fi
            
            cd - >/dev/null
            rm -rf "$temp_dir"
        fi
        
        sleep 10
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        log_warn "GitHub release not available yet. Using manual method..."
        setup_homebrew_tap_manual "$version"
        return
    fi
    
    # Trigger the tap update via GitHub Actions workflow dispatch
    log_step "Triggering Homebrew tap update..."
    
    if gh workflow run update-formula.yml \
        --repo "rhino11/homebrew-tap" \
        --field "version=$version" \
        --field "sha256_amd64=$sha256_amd64" \
        --field "sha256_arm64=$sha256_arm64"; then
        
        log_success "✅ Homebrew tap update triggered successfully"
        log_info "Monitor progress: https://github.com/rhino11/homebrew-tap/actions"
    else
        log_warn "Failed to trigger workflow. Using manual method..."
        setup_homebrew_tap_manual "$version"
    fi
}

# Function to setup Homebrew tap manually (fallback)
setup_homebrew_tap_manual() {
    local version=$1
    local tap_dir="/tmp/homebrew-tap-$$"
    
    log_info "Setting up Homebrew tap manually..."
    
    # Clone the tap repository
    if ! git clone "https://github.com/rhino11/homebrew-tap.git" "$tap_dir"; then
        log_error "Failed to clone tap repository"
        return 1
    fi
    
    cd "$tap_dir"
    
    # Create Formula directory if it doesn't exist
    mkdir -p Formula
    
    # Wait for GitHub release to be available
    log_info "Waiting for GitHub release to be available..."
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "https://github.com/rhino11/pivot/releases/download/$version/pivot.rb" >/dev/null 2>&1; then
            break
        fi
        echo "Attempt $attempt/$max_attempts: Waiting for release..."
        sleep 10
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        log_warn "GitHub release not available yet. You'll need to update the tap manually later."
        cd - >/dev/null
        rm -rf "$tap_dir"
        return
    fi
    
    # Download the generated formula
    if curl -o "Formula/pivot.rb" "https://github.com/rhino11/pivot/releases/download/$version/pivot.rb"; then
        log_success "Downloaded formula"
    else
        log_error "Failed to download formula"
        cd - >/dev/null
        rm -rf "$tap_dir"
        return 1
    fi
    
    # Commit and push
    git add Formula/pivot.rb
    git commit -m "Update pivot to $version"
    
    if git push origin main; then
        log_success "✅ Homebrew tap updated successfully"
    else
        log_error "Failed to push changes"
    fi
    
    # Cleanup
    cd - >/dev/null
    rm -rf "$tap_dir"
}