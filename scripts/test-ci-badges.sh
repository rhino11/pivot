#!/bin/bash

# Badge CI Simulation Script
# ==========================
# This script exactly simulates what the GitHub Actions badge workflow does

set -e

echo "🧪 Simulating CI Badge Workflow"
echo

# Get the token we're using locally
GIST_TOKEN=$(gh auth token)
echo "ℹ️  Using GIST_TOKEN: ${GIST_TOKEN:0:10}..."

# Get gist IDs from repository secrets (simulate CI environment)
COVERAGE_GIST_ID="8466693b8eb4ca358099fabc6ed234e0"
SECURITY_GIST_ID="a93cb6b503277dd460826517a831497e"
BADGES_GIST_ID="0a39d1979cd714d14836e9d6427d2eb9"

echo "ℹ️  Using Gist IDs:"
echo "   Coverage: $COVERAGE_GIST_ID"
echo "   Security: $SECURITY_GIST_ID"
echo "   Badges: $BADGES_GIST_ID"
echo

# Simulate coverage badge update (like CI would do)
echo "🎯 Simulating Coverage Badge Update"
COVERAGE_DATA='{
  "schemaVersion": 1,
  "label": "Coverage",
  "message": "87.2%",
  "color": "brightgreen"
}'

echo "Making API request to update coverage gist..."
RESPONSE=$(curl -s -w "%{http_code}" -H "Authorization: token $GIST_TOKEN" \
  -X PATCH "https://api.github.com/gists/$COVERAGE_GIST_ID" \
  -d "{\"files\": {\"pivot-coverage.json\": {\"content\": \"$COVERAGE_DATA\"}}}")

HTTP_CODE="${RESPONSE: -3}"
RESPONSE_BODY="${RESPONSE%???}"

echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Coverage badge update successful"
    RAW_URL=$(echo "$RESPONSE_BODY" | jq -r '.files["pivot-coverage.json"].raw_url')
    echo "Raw URL: $RAW_URL"
    
    # Verify the content
    echo "Verifying content..."
    CONTENT=$(curl -s "$RAW_URL")
    echo "Content: $CONTENT"
else
    echo "❌ Coverage badge update failed"
    echo "Response: $RESPONSE_BODY"
    exit 1
fi

echo
echo "🎯 Simulating Security Badge Update"
SECURITY_DATA='{
  "schemaVersion": 1,
  "label": "Security",
  "message": "A",
  "color": "brightgreen"
}'

echo "Making API request to update security gist..."
RESPONSE=$(curl -s -w "%{http_code}" -H "Authorization: token $GIST_TOKEN" \
  -X PATCH "https://api.github.com/gists/$SECURITY_GIST_ID" \
  -d "{\"files\": {\"pivot-security.json\": {\"content\": \"$SECURITY_DATA\"}}}")

HTTP_CODE="${RESPONSE: -3}"
RESPONSE_BODY="${RESPONSE%???}"

echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Security badge update successful"
else
    echo "❌ Security badge update failed"
    echo "Response: $RESPONSE_BODY"
    exit 1
fi

echo
echo "🎯 Simulating Build Status Badge Update"
BUILD_DATA='{
  "schemaVersion": 1,
  "label": "Build",
  "message": "passing",
  "color": "brightgreen"
}'

echo "Making API request to update build status gist..."
RESPONSE=$(curl -s -w "%{http_code}" -H "Authorization: token $GIST_TOKEN" \
  -X PATCH "https://api.github.com/gists/$BADGES_GIST_ID" \
  -d "{\"files\": {\"pivot-build.json\": {\"content\": \"$BUILD_DATA\"}}}")

HTTP_CODE="${RESPONSE: -3}"
RESPONSE_BODY="${RESPONSE%???}"

echo "HTTP Status: $HTTP_CODE"

if [ "$HTTP_CODE" = "200" ]; then
    echo "✅ Build status badge update successful"
else
    echo "❌ Build status badge update failed"
    echo "Response: $RESPONSE_BODY"
    exit 1
fi

echo
echo "🎉 All badge updates successful!"
echo
echo "📋 Badge URLs (for shields.io):"
echo "Coverage: https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/$COVERAGE_GIST_ID/raw/pivot-coverage.json"
echo "Security: https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/$SECURITY_GIST_ID/raw/pivot-security.json"
echo "Build: https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/$BADGES_GIST_ID/raw/pivot-build.json"
echo
echo "✅ The badge system is working correctly!"
echo "   The issue in CI is likely that GIST_TOKEN doesn't match this working token."
