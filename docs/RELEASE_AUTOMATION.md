# Release Automation Guide

This document describes the comprehensive release automation system for Pivot CLI.

## Overview

The Pivot CLI project features a fully automated release pipeline that handles:

- **Cross-platform builds** (Windows, macOS, Linux on AMD64/ARM64)
- **Package generation** (DEB, RPM, Homebrew, Chocolatey)
- **Homebrew tap management** (automated formula updates)
- **Installation testing** (verification across multiple methods)
- **GitHub releases** (automated asset publishing)

## Scripts

### `scripts/release.sh`

The main release script that orchestrates the entire release process.

**Usage:**
```bash
./scripts/release.sh v1.1.0
```

**What it does:**
1. Validates version format and prerequisites
2. Runs comprehensive tests and linting
3. Builds all platform targets
4. Creates and pushes a Git tag
5. Triggers GitHub Actions for automated release
6. Optionally updates the Homebrew tap

**Prerequisites:**
- Clean working directory
- All tests passing
- No lint errors
- Valid version format (vX.Y.Z)

### `scripts/setup-homebrew-tap.sh`

Sets up the Homebrew tap repository with automated update capabilities.

**Usage:**
```bash
./scripts/setup-homebrew-tap.sh
```

**What it does:**
1. Creates `rhino11/homebrew-tap` repository (if needed)
2. Sets up initial formula structure
3. Creates GitHub Actions workflow for automated updates
4. Configures repository for release integration

**Prerequisites:**
- GitHub CLI (`gh`) installed and authenticated
- Repository creation permissions

### `scripts/test-installation.sh`

Comprehensive installation testing across multiple methods.

**Usage:**
```bash
# Test latest version
./scripts/test-installation.sh

# Test specific version
./scripts/test-installation.sh v1.1.0

# Cleanup only
./scripts/test-installation.sh --cleanup
```

**What it tests:**
1. **Homebrew installation:** `brew install rhino11/tap/pivot`
2. **Direct download:** Platform-specific binary downloads
3. **Basic functionality:** Version and help commands
4. **Cross-platform support:** Automatic OS/architecture detection

## GitHub Actions Workflow

The CI/CD pipeline (`.github/workflows/ci.yml`) automatically:

### On Tag Push (Release Trigger)

1. **Builds** binaries for all platforms
2. **Packages** for multiple package managers:
   - DEB packages for Debian/Ubuntu
   - RPM packages for RHEL/Fedora/CentOS
   - Homebrew formula for macOS
   - Chocolatey package for Windows
3. **Calculates** SHA256 checksums
4. **Creates** GitHub release with all assets
5. **Updates** Homebrew tap via API call

### Platform Matrix

- **Windows:** AMD64, ARM64
- **macOS:** AMD64 (Intel), ARM64 (Apple Silicon)
- **Linux:** AMD64, ARM64

### Package Formats

- **DEB:** `pivot_<version>_amd64.deb`
- **RPM:** `pivot-<version>-1.x86_64.rpm`
- **Homebrew:** `pivot.rb` formula
- **Chocolatey:** `pivot.nuspec` and install script

## Homebrew Tap Automation

### Repository Structure

```
rhino11/homebrew-tap/
├── Formula/
│   └── pivot.rb           # Homebrew formula
├── .github/workflows/
│   └── update-formula.yml # Automated update workflow
└── README.md              # Installation instructions
```

### Automated Updates

When a release is created:

1. **GitHub Actions** calculates SHA256 hashes for macOS binaries
2. **API call** triggers `repository_dispatch` event on homebrew-tap
3. **Homebrew tap workflow** updates the formula with new version and hashes
4. **Users** can immediately install the new version

### Manual Formula Update

If automation fails, formula can be updated manually:

```bash
# Clone the tap
git clone https://github.com/rhino11/homebrew-tap.git
cd homebrew-tap

# Download new formula from release
curl -o Formula/pivot.rb https://github.com/rhino11/pivot/releases/download/v1.1.0/pivot.rb

# Commit and push
git add Formula/pivot.rb
git commit -m "Update pivot to v1.1.0"
git push origin main
```

## Release Checklist

### Pre-Release

- [ ] All tests passing (`make test`)
- [ ] No lint errors (`make lint`)
- [ ] Working directory clean
- [ ] Version incremented in `VERSION.md`
- [ ] CHANGELOG updated (if maintained)

### Release Process

- [ ] Run `./scripts/release.sh vX.Y.Z`
- [ ] Monitor GitHub Actions: https://github.com/rhino11/pivot/actions
- [ ] Verify release assets: https://github.com/rhino11/pivot/releases
- [ ] Check Homebrew tap update: https://github.com/rhino11/homebrew-tap/actions

### Post-Release Verification

- [ ] Test Homebrew installation: `brew install rhino11/tap/pivot`
- [ ] Test direct download from releases page
- [ ] Run installation tests: `./scripts/test-installation.sh vX.Y.Z`
- [ ] Verify package manager installations work

## Troubleshooting

### Common Issues

**1. Homebrew tap doesn't exist**
```bash
./scripts/setup-homebrew-tap.sh
```

**2. GitHub Actions failing**
- Check workflow logs
- Verify secrets are configured
- Ensure repository permissions

**3. Installation tests failing**
- Check network connectivity
- Verify release assets are published
- Test specific platform/method

**4. Formula update failing**
- Check homebrew-tap repository exists
- Verify GitHub token permissions
- Try manual formula update

### Debug Commands

```bash
# Check release status
gh release view v1.1.0

# Monitor workflow runs
gh run list --repo rhino11/pivot

# Test local build
make build-all

# Verify tap formula
brew audit --strict pivot.rb
```

## Version Management

See `VERSION.md` for version history and guidelines. Only modify `VERSION.md` when incrementing the release tag version.

### Semantic Versioning

- **MAJOR** (X.0.0): Breaking changes
- **MINOR** (0.Y.0): New features, backward compatible  
- **PATCH** (0.0.Z): Bug fixes, backward compatible

## Security Considerations

- GitHub tokens used for API access (read-only for public repositories)
- Package signatures for integrity verification
- Automated vulnerability scanning in CI pipeline
- Dependency security auditing

---

This automation system ensures reliable, consistent releases while maintaining high quality standards and comprehensive testing across all supported platforms and installation methods.
