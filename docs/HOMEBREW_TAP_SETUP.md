# Homebrew Tap Automatic Updates Setup

This document explains how to configure automatic updates for the Homebrew tap repository when releases are created.

## Overview

When a new release is tagged (e.g., `v1.1.0`), the GitHub Actions workflow automatically:
1. Builds binaries for all platforms
2. Creates the GitHub release
3. **Attempts to update the Homebrew tap repository**

The last step requires proper authentication to commit changes to the homebrew-tap repository.

## Current Issue

The automatic Homebrew tap update is failing because the workflow doesn't have permission to commit to the external `rhino11/homebrew-tap` repository.

## Solution: Setup HOMEBREW_PAT Secret

### Step 1: Create a Personal Access Token

1. Go to GitHub → Settings → Developer settings → Personal access tokens → Tokens (classic)
2. Click "Generate new token (classic)"
3. Set the following:
   - **Note**: `Homebrew Tap Updates for Pivot CLI`
   - **Expiration**: Set to a reasonable timeframe (90 days, 1 year, or no expiration)
   - **Scopes**: Select `repo` (this includes all repository permissions)

### Step 2: Add the Secret to the Repository

1. Go to the main repository: https://github.com/rhino11/pivot
2. Navigate to Settings → Secrets and variables → Actions
3. Click "New repository secret"
4. Set:
   - **Name**: `HOMEBREW_PAT`
   - **Value**: The Personal Access Token from Step 1

### Step 3: Test the Setup

1. Create a new release (or re-run the existing v1.1.0 release workflow)
2. Monitor the "Update Homebrew tap" step in the release job
3. Verify that the homebrew-tap repository gets updated automatically

## Manual Fallback

If automatic updates fail, you can manually update the tap using:

```bash
# Using the manual script
./scripts/update-homebrew-tap.sh v1.1.0

# Or with specific checksums
./scripts/update-homebrew-tap.sh v1.1.0 [sha256_amd64] [sha256_arm64]
```

## Workflow Details

The updated workflow now:

1. **Checks for HOMEBREW_PAT**: Uses the secret if available
2. **Direct Git Operations**: Clones the tap repo, updates the formula, commits, and pushes
3. **Better Error Handling**: Provides clear messages about what went wrong
4. **Graceful Degradation**: Continues the release even if tap update fails

## Security Notes

- The HOMEBREW_PAT has `repo` access to all repositories the user owns
- Consider using a dedicated service account for production environments
- The token is only used to update the homebrew-tap repository
- All operations are logged in the GitHub Actions workflow

## Verification

After setup, check that:
- [ ] HOMEBREW_PAT secret exists in repository settings
- [ ] Token has `repo` permissions
- [ ] Next release automatically updates the homebrew-tap
- [ ] Users can install with `brew install rhino11/tap/pivot`

## Troubleshooting

### "Authentication failed (401/403)"
- Verify the HOMEBREW_PAT secret exists and is correctly set
- Check that the token hasn't expired
- Ensure the token has `repo` permissions

### "Failed to clone homebrew-tap repository"
- Verify the homebrew-tap repository exists
- Check that the token owner has write access to the repository

### "No changes to commit"
- The formula might already be up to date
- This is not an error - the workflow will continue successfully
