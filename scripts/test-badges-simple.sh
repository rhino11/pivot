#!/bin/bash

echo "🧪 Testing Badge System Authentication"
echo

# Check if gh is available
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is required but not installed."
    exit 1
fi

# Check authentication
echo "ℹ️  Testing GitHub authentication..."
if gh auth status &> /dev/null; then
    echo "✅ GitHub CLI authenticated"
else
    echo "❌ GitHub CLI not authenticated"
    exit 1
fi

# Test gist access
echo "ℹ️  Testing gist access..."
if gh api /gists > /dev/null 2>&1; then
    echo "✅ GitHub gist access working"
else
    echo "❌ GitHub gist access failed"
    exit 1
fi

# Check repository secrets
echo "ℹ️  Checking repository secrets..."
gh secret list

echo
echo "✅ Basic authentication tests passed!"
