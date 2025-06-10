#!/bin/bash

# Simple Badge Updater Script
# ============================
# Updates badge JSON in GitHub gists using direct API calls

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Update a gist file with badge JSON
update_badge() {
    local gist_id="$1"
    local filename="$2"
    local label="$3"
    local message="$4"
    local color="$5"
    local token="${GIST_TOKEN:-$GITHUB_TOKEN}"
    
    if [[ -z "$token" ]]; then
        log_error "No authentication token found. Set GIST_TOKEN or GITHUB_TOKEN."
        return 1
    fi
    
    # Create badge JSON
    local badge_json=$(jq -n \
        --arg label "$label" \
        --arg message "$message" \
        --arg color "$color" \
        '{
            schemaVersion: 1,
            label: $label,
            message: $message,
            color: $color
        }')
    
    log_info "Updating $label badge in gist $gist_id..."
    
    # Update gist using GitHub API
    local response=$(curl -s -w "%{http_code}" \
        -X PATCH \
        -H "Authorization: token $token" \
        -H "Accept: application/vnd.github.v3+json" \
        "https://api.github.com/gists/$gist_id" \
        -d "{
            \"files\": {
                \"$filename\": {
                    \"content\": $(echo "$badge_json" | jq -c .)
                }
            }
        }")
    
    local http_code="${response: -3}"
    local response_body="${response%???}"
    
    if [[ "$http_code" == "200" ]]; then
        log_success "$label badge updated successfully"
        return 0
    else
        log_error "$label badge update failed (HTTP $http_code)"
        echo "Response: $response_body"
        return 1
    fi
}

# Main function
main() {
    local action="$1"
    
    case "$action" in
        "coverage")
            update_badge "$2" "$3" "Coverage" "$4" "$5"
            ;;
        "security")
            update_badge "$2" "$3" "Security" "$4" "$5"
            ;;
        "build")
            update_badge "$2" "$3" "Build" "$4" "$5"
            ;;
        "go-version")
            update_badge "$2" "$3" "Go" "$4" "$5"
            ;;
        "license")
            update_badge "$2" "$3" "License" "$4" "$5"
            ;;
        *)
            echo "Usage: $0 <action> <gist_id> <filename> <message> <color>"
            echo "Actions: coverage, security, build, go-version, license"
            echo "Example: $0 coverage abc123 pivot-coverage.json '85.2%' brightgreen"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
