#!/bin/bash

set -euo pipefail

# Configuration
VERSION="1.0.1"
SHA256_AMD64="bc901bc05bd4be1a1e5a17125b787b94d725fb3e12802b187c16605543c4e5c1"
SHA256_ARM64="c37ccb349a1cf07490b0481a20d58fd554eb62cf6409e2065a3afe8e1bc65174"
TAP_REPO="rhino11/homebrew-tap"

echo "🔧 Fixing Homebrew tap formula for version v${VERSION}"
echo "📦 AMD64 SHA256: ${SHA256_AMD64}"
echo "📦 ARM64 SHA256: ${SHA256_ARM64}"
echo

# Create temporary directory
TEMP_DIR=$(mktemp -d)
echo "📁 Working in: ${TEMP_DIR}"

# Clone the homebrew-tap repository
echo "📥 Cloning homebrew-tap repository..."
git clone "https://github.com/${TAP_REPO}.git" "${TEMP_DIR}/homebrew-tap"
cd "${TEMP_DIR}/homebrew-tap"

# Create the updated formula
echo "✏️  Updating formula..."
cat > Formula/pivot.rb << EOF
class Pivot < Formula
  desc "GitHub Issues Management CLI"
  homepage "https://github.com/rhino11/pivot"
  version "${VERSION}"
  
  if Hardware::CPU.arm?
    url "https://github.com/rhino11/pivot/releases/download/v${VERSION}/pivot-darwin-arm64"
    sha256 "${SHA256_ARM64}"
  else
    url "https://github.com/rhino11/pivot/releases/download/v${VERSION}/pivot-darwin-amd64"
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

# Show the changes
echo "📋 Updated formula content:"
echo "----------------------------------------"
cat Formula/pivot.rb
echo "----------------------------------------"
echo

# Commit and push (this will require manual authentication)
echo "💾 Committing changes..."
git config user.name "GitHub Actions"
git config user.email "action@github.com"
git add Formula/pivot.rb
git commit -m "Update pivot to v${VERSION} with correct SHA256 hashes"

echo "🚀 Ready to push changes. Please authenticate if prompted..."
git push origin main

echo "✅ Homebrew tap updated successfully!"
echo "🍺 Users can now install with: brew install rhino11/tap/pivot"

# Cleanup
cd /
rm -rf "${TEMP_DIR}"

echo "🎉 Done!"
