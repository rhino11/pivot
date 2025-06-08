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

# Configuration
TAP_OWNER="rhino11"
TAP_REPO="homebrew-tap"
TAP_URL="https://github.com/$TAP_OWNER/$TAP_REPO"
MAIN_REPO="rhino11/pivot"

log_info "🍺 Setting up Homebrew tap for Pivot CLI"
echo

# Check if GitHub CLI is available
if ! command -v gh &> /dev/null; then
    log_warn "GitHub CLI (gh) not found. You'll need to create the repository manually."
    echo
    echo "To install GitHub CLI:"
    echo "  macOS: brew install gh"
    echo "  Linux: Visit https://cli.github.com/"
    echo
    exit 1
fi

# Check if user is authenticated with GitHub CLI
if ! gh auth status &> /dev/null; then
    log_error "Not authenticated with GitHub CLI. Run 'gh auth login' first."
    exit 1
fi

# Check if tap repository already exists
log_step "Checking if tap repository exists..."
if gh repo view "$TAP_OWNER/$TAP_REPO" &> /dev/null; then
    log_info "✅ Tap repository already exists: $TAP_URL"
    read -p "Do you want to update the existing tap? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "Exiting without changes."
        exit 0
    fi
else
    log_step "Creating new tap repository..."
    
    # Create the repository
    gh repo create "$TAP_OWNER/$TAP_REPO" \
        --public \
        --description "Homebrew tap for Pivot CLI" \
        --clone=false
    
    if [ $? -eq 0 ]; then
        log_info "✅ Repository created: $TAP_URL"
    else
        log_error "Failed to create repository"
        exit 1
    fi
fi

# Clone the repository to a temporary directory
TEMP_DIR=$(mktemp -d)
log_step "Cloning tap repository to $TEMP_DIR..."

cd "$TEMP_DIR"
gh repo clone "$TAP_OWNER/$TAP_REPO" .

# Create Formula directory
log_step "Setting up repository structure..."
mkdir -p Formula

# Create README if it doesn't exist
if [ ! -f README.md ]; then
    cat > README.md << EOF
# Homebrew Tap for Pivot CLI

This is the official Homebrew tap for [Pivot CLI](https://github.com/$MAIN_REPO).

## Installation

\`\`\`bash
brew tap $TAP_OWNER/tap
brew install pivot
\`\`\`

## About Pivot CLI

Pivot is a CLI tool for managing GitHub issues locally with offline sync capabilities.

### Features

- Local GitHub issues management
- Offline sync capabilities
- Interactive configuration setup
- Cross-platform support

## Documentation

For more information, visit the [main repository](https://github.com/$MAIN_REPO).
EOF
fi

# Create initial formula if it doesn't exist
if [ ! -f Formula/pivot.rb ]; then
    log_step "Creating initial formula..."
    cat > Formula/pivot.rb << 'EOF'
class Pivot < Formula
  desc "GitHub Issues Management CLI"
  homepage "https://github.com/rhino11/pivot"
  version "1.0.0"
  
  if Hardware::CPU.arm?
    url "https://github.com/rhino11/pivot/releases/download/v1.0.0/pivot-darwin-arm64"
    sha256 "PLACEHOLDER_SHA256_ARM64"
  else
    url "https://github.com/rhino11/pivot/releases/download/v1.0.0/pivot-darwin-amd64"
    sha256 "PLACEHOLDER_SHA256_AMD64"
  end
  
  def install
    bin.install Dir["pivot-darwin-*"].first => "pivot"
  end
  
  test do
    assert_match "pivot", shell_output("#{bin}/pivot version")
  end
end
EOF
fi

# Create GitHub Actions workflow for automated updates
log_step "Creating GitHub Actions workflow..."
mkdir -p .github/workflows

cat > .github/workflows/update-formula.yml << 'EOF'
name: Update Formula

on:
  repository_dispatch:
    types: [update-formula]
  workflow_dispatch:
    inputs:
      version:
        description: 'Version to update to (e.g., v1.1.0)'
        required: true
        type: string
      sha256_amd64:
        description: 'SHA256 for AMD64 binary'
        required: true
        type: string
      sha256_arm64:
        description: 'SHA256 for ARM64 binary'
        required: true
        type: string

jobs:
  update:
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout
      uses: actions/checkout@v4
    
    - name: Update formula
      run: |
        VERSION="${{ github.event.inputs.version || github.event.client_payload.version }}"
        SHA256_AMD64="${{ github.event.inputs.sha256_amd64 || github.event.client_payload.sha256_amd64 }}"
        SHA256_ARM64="${{ github.event.inputs.sha256_arm64 || github.event.client_payload.sha256_arm64 }}"
        
        # Remove 'v' prefix if present
        VERSION_CLEAN=${VERSION#v}
        
        # Update the formula
        cat > Formula/pivot.rb << EOF
        class Pivot < Formula
          desc "GitHub Issues Management CLI"
          homepage "https://github.com/rhino11/pivot"
          version "${VERSION_CLEAN}"
          
          if Hardware::CPU.arm?
            url "https://github.com/rhino11/pivot/releases/download/${VERSION}/pivot-darwin-arm64"
            sha256 "${SHA256_ARM64}"
          else
            url "https://github.com/rhino11/pivot/releases/download/${VERSION}/pivot-darwin-amd64"
            sha256 "${SHA256_AMD64}"
          end
          
          def install
            bin.install Dir["pivot-darwin-*"].first => "pivot"
          end
          
          test do
            assert_match "pivot", shell_output("#{bin}/pivot version")
          end
        end
        EOF
    
    - name: Commit and push
      run: |
        VERSION="${{ github.event.inputs.version || github.event.client_payload.version }}"
        git config --local user.email "action@github.com"
        git config --local user.name "GitHub Action"
        git add Formula/pivot.rb
        git commit -m "Update pivot to ${VERSION}" || exit 0
        git push
EOF

# Commit and push changes
log_step "Committing and pushing changes..."
git add .
git commit -m "Initial tap setup with automated workflow" || log_info "No changes to commit"
git push origin main

log_info "✅ Homebrew tap setup completed successfully!"
echo
echo "Next steps:"
echo "1. The tap is now available at: $TAP_URL"
echo "2. Users can install with: brew install $TAP_OWNER/tap/pivot"
echo "3. The formula will be automatically updated when releases are created"
echo "4. Test the installation: brew install $TAP_OWNER/tap/pivot"
echo

# Cleanup
cd ..
rm -rf "$TEMP_DIR"

log_info "🎉 Setup complete! Your Homebrew tap is ready to use."
