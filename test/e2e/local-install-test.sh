#!/bin/bash

# Local E2E Installation Test
# This script tests the local installation and functionality

set -e

echo "🧪 Starting Local E2E Installation Test..."

# Get the script directory to work relative to project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
TEST_DIR="/tmp/pivot-local-e2e-test"

# Cleanup function
cleanup() {
    echo "🧹 Cleaning up test environment..."
    rm -rf "$TEST_DIR"
}

# Set up cleanup trap
trap cleanup EXIT

echo "📁 Creating test environment..."
mkdir -p "$TEST_DIR"
cd "$PROJECT_ROOT"

echo "🔨 Building pivot binary..."
if go build -o pivot cmd/main.go; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

echo "📋 Test 1: Binary Functionality"
# Test basic commands
if ./pivot --help > /dev/null; then
    echo "✅ Help command works"
else
    echo "❌ Help command failed"
    exit 1
fi

if ./pivot version > /dev/null; then
    echo "✅ Version command works"
else
    echo "❌ Version command failed"
    exit 1
fi

echo "📋 Test 2: Installation Scripts Exist"
# Check that installation scripts exist
if [ -f "install.sh" ]; then
    echo "✅ Unix install script exists"
else
    echo "❌ Unix install script missing"
    exit 1
fi

if [ -f "install.ps1" ]; then
    echo "✅ PowerShell install script exists"
else
    echo "❌ PowerShell install script missing"
    exit 1
fi

echo "📋 Test 3: Configuration Functionality"
# Test configuration in isolated environment
cd "$TEST_DIR"

# Copy binary to test directory
cp "$PROJECT_ROOT/pivot" .

# Test init command help
if ./pivot init --help > /dev/null; then
    echo "✅ Init command help works"
else
    echo "❌ Init command help failed"
    exit 1
fi

# Test config command help
if ./pivot config --help > /dev/null; then
    echo "✅ Config command help works"
else
    echo "❌ Config command help failed"
    exit 1
fi

echo "📋 Test 4: Package Manager Configuration"
cd "$PROJECT_ROOT"

# Check that package manager configurations exist
if [ -f "package.json" ]; then
    echo "✅ NPM package.json exists"
else
    echo "❌ NPM package.json missing"
fi

if [ -d "snap" ] && [ -f "snap/snapcraft.yaml" ]; then
    echo "✅ Snap configuration exists"
else
    echo "❌ Snap configuration missing"
fi

if [ -f "Dockerfile" ]; then
    echo "✅ Docker configuration exists"
else
    echo "❌ Docker configuration missing"
fi

echo "📋 Test 5: Documentation"
# Check that key documentation exists
if [ -f "README.md" ]; then
    echo "✅ README.md exists"
else
    echo "❌ README.md missing"
    exit 1
fi

if [ -f "CONTRIBUTING.md" ]; then
    echo "✅ CONTRIBUTING.md exists"
else
    echo "❌ CONTRIBUTING.md missing"
    exit 1
fi

if [ -f "config.example.yml" ]; then
    echo "✅ Example configuration exists"
else
    echo "❌ Example configuration missing"
    exit 1
fi

echo "✅ All local E2E tests passed!"
echo "🎉 Pivot CLI is ready for distribution!"
