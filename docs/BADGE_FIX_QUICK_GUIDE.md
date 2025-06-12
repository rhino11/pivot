# 🏷️ Badge Setup Issue - Quick Fix Guide

## Current Problem

The dynamic badges are failing with 403 errors because the `GIST_SECRET` token either:
- Doesn't exist as a repository secret
- Has expired 
- Doesn't have the correct `gist` permissions
- The workflow is falling back to `github.token` which can't access gists

## Quick Fix Steps

### Step 1: Create a New Personal Access Token

1. Go to [GitHub Personal Access Tokens](https://github.com/settings/tokens)
2. Click "Generate new token (classic)"
3. Set the name: `Pivot Badge Updates`
4. Set expiration: `90 days` (or longer)
5. **Important**: Select the `gist` scope (this is crucial!)
6. Generate and copy the token

### Step 2: Update Repository Secret

1. Go to [Pivot Repository Secrets](https://github.com/rhino11/pivot/settings/secrets/actions)
2. Find `GIST_SECRET` or create it if it doesn't exist
3. Update/create with the new token from Step 1

### Step 3: Test the Fix

1. Go to [GitHub Actions](https://github.com/rhino11/pivot/actions)
2. Re-run a failed workflow or push a new commit to main
3. Check that the "Update Coverage Badge" step succeeds

## Current Gist Configuration

The badges use these gists (already configured correctly):

- **Coverage**: `8466693b8eb4ca358099fabc6ed234e0` → `pivot-coverage.json`
- **Security**: `a93cb6b503277dd460826517a831497e` → `pivot-security.json`  
- **Build/Go/License**: `0a39d1979cd714d14836e9d6427d2eb9` → `pivot-build.json`, `pivot-go-version.json`, `pivot-license.json`

## Verification

After fixing the secret, the badges should update automatically and show:
- ✅ Coverage: 80.1%
- ✅ Security: A rating
- ✅ Build: passing
- ✅ Go: 1.24
- ✅ License: MIT

## Troubleshooting

If badges still fail:

1. **Check token permissions**: The token MUST have `gist` scope
2. **Check token expiration**: Tokens expire and need renewal
3. **Check repository secret**: Ensure `GIST_SECRET` exists and has the right value
4. **Check workflow logs**: Look for specific error messages in the Actions tab

## Alternative: Disable Dynamic Badges Temporarily

If you need to disable dynamic badges temporarily, you can:

1. Comment out the badge update steps in `.github/workflows/ci.yml`
2. Use static badges in the README instead
3. Re-enable when the token issue is resolved

The static badge URLs are:
```markdown
[![Build Status](https://img.shields.io/badge/build-passing-brightgreen)](https://github.com/rhino11/pivot/actions)
[![Coverage](https://img.shields.io/badge/coverage-80.1%25-green)](https://github.com/rhino11/pivot/actions)
[![Security](https://img.shields.io/badge/security-A-brightgreen)](https://github.com/rhino11/pivot/security)
```
