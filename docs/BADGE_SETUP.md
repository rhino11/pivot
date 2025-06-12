# Dynamic Badge Setup Guide

This guide explains how to set up dynamic badges that automatically update when your CI pipeline runs.

## Overview

The project uses GitHub Actions to generate dynamic badges for:
- **Coverage**: Shows test coverage percentage
- **Security**: Shows security rating (A-D)
- **Build Status**: Shows CI build status
- **Go Version**: Shows Go version from go.mod
- **License**: Shows detected license type

## Setup Instructions

### 1. Create GitHub Gists

You need to create three GitHub gists to store badge data:

1. **Coverage Gist**: For coverage badge
2. **Security Gist**: For security rating badge  
3. **Badges Gist**: For build status, Go version, and license badges

#### Creating Gists

1. Go to [gist.github.com](https://gist.github.com)
2. Create a new gist with:
   - **Filename**: `placeholder.json`
   - **Content**: `{}`
   - **Visibility**: Public
3. Create the gist and note the gist ID from the URL
4. Repeat for all three gists

### 2. Configure Repository Secrets

Add the following secrets to your GitHub repository:

1. Go to **Settings** → **Secrets and variables** → **Actions**
2. Add these repository secrets:

```
COVERAGE_GIST_ID = your-coverage-gist-id
SECURITY_GIST_ID = your-security-gist-id  
BADGES_GIST_ID = your-badges-gist-id
GIST_SECRET = your-github-personal-access-token
```

**Creating GIST_SECRET:**
1. Go to GitHub → Settings → Developer settings → Personal access tokens
2. Generate new token (classic) with **'gist'** scope
3. Copy the token and add it as `GIST_SECRET` repository secret

> **Note**: The `GIST_SECRET` is required for the Schneegans dynamic-badges-action to update your gists. Without it, the action will fallback to `GITHUB_TOKEN` which may have limited permissions.

### 3. Update README Badge URLs

Replace the placeholder gist IDs in README.md with your actual gist IDs:

```markdown
[![Coverage Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/YOUR_COVERAGE_GIST_ID/raw/pivot-coverage.json)](https://github.com/rhino11/pivot/actions)
[![Security Rating](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/YOUR_SECURITY_GIST_ID/raw/pivot-security.json)](https://github.com/rhino11/pivot/security)
[![Build Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/YOUR_BADGES_GIST_ID/raw/pivot-build.json)](https://github.com/rhino11/pivot/actions)
[![Go Version](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/YOUR_BADGES_GIST_ID/raw/pivot-go-version.json)](https://golang.org)
[![License](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/YOUR_BADGES_GIST_ID/raw/pivot-license.json)](https://opensource.org/licenses/MIT)
```

## How It Works

### Automated Badge Updates (CI/CD)

The badges are now automatically updated using the [Schneegans/dynamic-badges-action](https://github.com/Schneegans/dynamic-badges-action) whenever code is pushed to the main branch:

1. **Coverage Badge**: Runs tests, calculates coverage percentage, and updates with color coding
2. **Security Badge**: Runs security scans (gosec, govulncheck), counts issues, and assigns A-D rating
3. **Build Status Badge**: Shows "passing" (green) or "failing" (red) based on test results
4. **Go Version Badge**: Reads version from go.mod and displays it
5. **License Badge**: Shows "MIT" license type

### Badge Color Coding

**Coverage Badge** (automatic color range):
- Green (≥90%)
- Yellow-Green (70-89%)  
- Orange (50-69%)
- Red (<50%)

**Security Badge** (issue-based rating):
- **A**: 0 issues (bright green)
- **B**: 1-2 issues (green)
- **C**: 3-5 issues (yellow)
- **D**: 6+ issues (red)

### License Badge
- Detects license from LICENSE file
- Color codes by license type:
  - MIT (yellow)
  - Apache-2.0 (blue)
  - BSD (orange)
  - GPL (green)
  - Custom/None (grey)

## Testing

### Local Testing

Test your badge setup before pushing to production:

```bash
# Test the entire dynamic badges integration
make test-dynamic-badges

# Test traditional badge system (legacy)
make test-badges

# Simulate CI badge workflow  
make test-ci-badges
```

The `test-dynamic-badges` script will:
- Calculate current coverage and security metrics
- Generate badge JSON files locally
- Test gist access permissions
- Show what badge data would be generated

### Manual Testing

You can also manually test badge generation:

```bash
# Run the test script directly
./scripts/test-dynamic-badges.sh

# Check generated badge JSON files
ls /tmp/badge-test/
cat /tmp/badge-test/coverage.json
```

## Badge Updates

Badges update automatically when:
- Code is pushed to `main` branch
- Tests run successfully
- Security scans complete

The badges will show the most recent results from the main branch.

## Troubleshooting

### Badges Not Updating
1. Check that gist IDs are correct in secrets
2. Verify gists are public
3. Check GitHub Actions logs for errors
4. Ensure secrets have proper permissions

### Badge Shows "Unknown"
- Usually indicates the gist hasn't been updated yet
- Wait for next CI run or manually trigger workflow

### Permission Errors
- Ensure gists are public
- Verify repository secrets are set correctly
- Check that GITHUB_TOKEN has gist permissions

## Manual Badge Generation

For testing, you can manually create badge JSON:

```bash
# Coverage badge JSON
echo '{"schemaVersion": 1, "label": "coverage", "message": "85%", "color": "brightgreen"}' > coverage.json

# Security badge JSON  
echo '{"schemaVersion": 1, "label": "security", "message": "A", "color": "brightgreen"}' > security.json

# Build badge JSON
echo '{"schemaVersion": 1, "label": "build", "message": "passing", "color": "brightgreen"}' > build.json
```

Then upload these to your gists manually to test the badge rendering.

## Alternative Approaches

If you prefer not to use gists, consider:

1. **Codecov/Coveralls**: Third-party coverage services
2. **SonarCloud**: Comprehensive code quality platform
3. **Shields.io dynamic badges**: Using your own JSON endpoint
4. **GitHub Pages**: Host badge JSON files statically

This setup provides full control over badge appearance and updates while remaining completely free and open source.
