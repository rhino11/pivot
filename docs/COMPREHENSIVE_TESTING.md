# Comprehensive Testing Documentation

## Overview

This document provides a complete guide to the comprehensive testing infrastructure for the Pivot CLI project. The testing suite includes unit tests, CLI command tests, security tests, E2E tests, and post-release validation across multiple platforms and package managers.

## Table of Contents

1. [Testing Infrastructure Overview](#testing-infrastructure-overview)
2. [Unit Tests](#unit-tests)
3. [CLI Test Suite](#cli-test-suite)
4. [Security Test Suite](#security-test-suite)
5. [End-to-End (E2E) Tests](#end-to-end-e2e-tests)
6. [Post-Release Validation](#post-release-validation)
7. [CI/CD Integration](#cicd-integration)
8. [Local Development Workflow](#local-development-workflow)
9. [Test Configuration and Environment](#test-configuration-and-environment)
10. [Troubleshooting](#troubleshooting)

## Testing Infrastructure Overview

### Test Categories

The Pivot CLI project includes the following test categories:

| Test Type | Purpose | Scope | Automation |
|-----------|---------|-------|------------|
| **Unit Tests** | Code logic validation | Individual functions/modules | CI + Local |
| **CLI Tests** | Command-line interface validation | CLI commands and interactions | CI + Local |
| **Security Tests** | Security vulnerability scanning | Code, dependencies, binaries | CI + Local |
| **E2E Tests** | End-user installation experience | Package managers, platforms | CI + Local |
| **Post-Release Tests** | Binary validation after release | All release artifacts | CI + Manual |

### Test Execution Methods

1. **Local Development**: Run individual test suites during development
2. **CI/CD Pipeline**: Automated execution on code changes and releases
3. **Manual Validation**: Post-release verification and troubleshooting

### Test Scripts Overview

| Script | Purpose | Location |
|--------|---------|----------|
| `test-cli.sh` | CLI command testing | `scripts/test-cli.sh` |
| `security-test.sh` | Security scanning | `scripts/security-test.sh` |
| `test-homebrew-e2e.sh` | Homebrew E2E testing | `scripts/test-homebrew-e2e.sh` |
| `post-release-validation.sh` | Binary validation | `scripts/post-release-validation.sh` |
| `test-installation.sh` | Cross-platform installation | `scripts/test-installation.sh` |

## Unit Tests

### Overview
Standard Go unit tests that validate individual functions and modules.

### Execution
```bash
# Run all unit tests
make test

# Run with coverage
make coverage

# Run specific package
go test ./cmd/...
```

### Coverage
- Function-level testing
- Error handling validation
- Configuration parsing
- Core business logic

## CLI Test Suite

### Overview
Comprehensive testing of all CLI commands, including happy path scenarios, error cases, and edge conditions.

### Test Script: `scripts/test-cli.sh`

#### What It Tests
1. **Binary Compilation**: Ensures the CLI can be built successfully
2. **Command Execution**: Tests all primary commands (init, config, sync, version, help)
3. **Error Handling**: Validates proper error responses for invalid inputs
4. **Command Interactions**: Tests command combinations and workflows
5. **Configuration Handling**: Tests config creation, validation, and usage

#### Execution
```bash
# Run CLI test suite
make test-cli

# Direct execution
./scripts/test-cli.sh
```

#### Test Coverage
- `pivot version` - Version information display
- `pivot help` - Help system functionality
- `pivot init` - Project initialization
- `pivot config` - Configuration management
- `pivot sync` - Data synchronization
- `pivot completion` - Shell completion
- Error cases and invalid commands
- Concurrent execution scenarios
- Performance benchmarks

#### CI Integration
- Runs on every push and pull request
- Tests against compiled binary
- Validates command-line interface consistency

## Security Test Suite

### Overview
Multi-layered security testing using industry-standard tools to identify vulnerabilities, security issues, and compliance problems.

### Test Script: `scripts/security-test.sh`

#### Security Tools Used
1. **gosec**: Go security checker for code vulnerabilities (https://github.com/securego/gosec)
2. **govulncheck**: Go vulnerability scanner for known CVEs
3. **staticcheck**: Advanced static analysis
4. **nancy**: OSS Index vulnerability scanner for dependencies

**Nancy Dependency Vulnerability Scanner**

Nancy scans Go dependencies for known security vulnerabilities using Sonatype's OSS Index.

```bash
# Nancy installation
go install github.com/sonatype-nexus-community/nancy@latest

# Manual dependency scan
go list -json -deps ./... | nancy sleuth

# Include non-vulnerable packages in output
go list -json -deps ./... | nancy sleuth --loud
```

**Key Features:**
- **OSS Index Integration**: Uses Sonatype's OSS Index vulnerability database
- **Go Module Support**: Scans Go dependencies directly from module information
- **CVE Detection**: Identifies specific CVE and CWE vulnerabilities
- **Detailed Reporting**: Provides package-level vulnerability information

**Vulnerability Management:**
- **Ignore File**: Use `.nancy-ignore` to document and ignore false positives
- **Remediation Guidance**: Provides specific steps for addressing vulnerabilities
- **Version Tracking**: Identifies which package versions contain vulnerabilities
- **Update Recommendations**: Suggests dependency updates to resolve issues

**Integration:**
- Runs automatically in security test suite
- Generates detailed vulnerability reports
- Supports ignore lists for managing false positives
- Provides actionable remediation steps

**For comprehensive vulnerability management procedures, see [Dependency Vulnerability Management Guide](DEPENDENCY_VULNERABILITY_MANAGEMENT.md).**

#### What It Tests
1. **Code Security**: Scans source code for security anti-patterns
2. **Dependency Vulnerabilities**: Checks for known vulnerabilities in dependencies
3. **Configuration Security**: Validates secure configuration practices
4. **Binary Security**: Tests compiled binary security features
5. **Runtime Security**: Tests for information leakage and runtime security

#### Execution
```bash
# Run security test suite
make test-security

# Direct execution
./scripts/security-test.sh
```

#### Test Categories

##### Code Security Tests
- Hardcoded secrets detection
- SQL injection vulnerabilities
- Path traversal issues
- Insecure random number generation
- TLS/SSL configuration issues

##### Configuration Security Tests
- File permission validation
- Environment variable security
- Configuration file protection
- Secret management practices

##### Binary Security Tests
- Binary stripping verification
- Permission validation
- Security feature detection (when available)

##### Runtime Security Tests
- Memory safety checks
- Information leakage prevention
- Secure data handling

#### Reports Generated
- `security-report.md`: Comprehensive security assessment
- `gosec-report.json`: Detailed gosec findings
- Security recommendations and remediation steps

#### CI Integration
- Runs on every push and pull request
- Uploads security reports as artifacts
- Fails build on critical security issues

## End-to-End (E2E) Tests

### Overview
E2E tests validate the complete user experience from package installation through basic functionality verification across different platforms and package managers.

### Homebrew E2E Testing

#### Test Script: `scripts/test-homebrew-e2e.sh`

##### What It Tests
1. **Prerequisites Validation**: macOS and Homebrew availability
2. **Clean State Verification**: No conflicting installations
3. **Tap Addition**: `brew tap rhino11/tap` functionality
4. **Package Installation**: `brew install pivot` success
5. **Basic Functionality**: Core commands (version, help, config)
6. **Package Information**: Installation metadata verification
7. **Formula Validation**: Homebrew formula audit compliance
8. **Version Consistency**: Installed version matches expected
9. **Cleanup**: Complete removal of test installations

##### Execution Modes
```bash
# Standard test with latest release
make test-homebrew-e2e

# Test with local build (no release required)
make test-homebrew-local

# Dry run (no actual installation)
./scripts/test-homebrew-e2e.sh --dry-run

# Test specific version
./scripts/test-homebrew-e2e.sh v1.1.0

# Test pre-release versions
./scripts/test-homebrew-e2e.sh --pre-release

# Cleanup only
make test-cleanup
```

##### Safety Features
- **Dry Run Mode**: Test logic without actual installation
- **Automatic Cleanup**: Removes test installations on failure
- **State Verification**: Ensures clean test environment
- **Error Handling**: Comprehensive error reporting
- **Timeout Protection**: Prevents hanging on network issues

### Cross-Platform Installation Testing

#### Test Script: `scripts/test-installation.sh`

##### What It Tests
1. **Homebrew Installation**: macOS package manager testing
2. **Direct Download**: Platform-specific binary downloads
3. **Basic Functionality**: Version and help commands
4. **Cross-Platform Support**: Automatic OS/architecture detection

##### Execution
```bash
# Test latest version
./scripts/test-installation.sh

# Test specific version
./scripts/test-installation.sh v1.1.0

# Cleanup only
./scripts/test-installation.sh --cleanup
```

### Local Installation Testing

#### Test Script: `test/e2e/local-install-test.sh`

Tests local build functionality and development workflow:
- Binary compilation and execution
- Installation script validation
- Configuration functionality
- Documentation completeness

## Post-Release Validation

### Overview
Comprehensive validation of all release artifacts after a GitHub release is published, ensuring all binaries and packages work correctly across platforms.

### Test Script: `scripts/post-release-validation.sh`

#### What It Tests
1. **Binary Downloads**: All platform binaries (macOS, Linux, Windows)
2. **SHA256 Verification**: Integrity checking using checksums
3. **Cross-Platform Functionality**: Basic command execution
4. **Package Files**: DEB packages and Homebrew formulas
5. **Architecture Coverage**: AMD64 and ARM64 support
6. **Release Completeness**: All expected artifacts present

#### Supported Platforms
- **macOS**: AMD64, ARM64
- **Linux**: AMD64, ARM64 (tested via Docker when available)
- **Windows**: AMD64, ARM64 (download and integrity only)

#### Package Testing
- **DEB Packages**: Structure and metadata validation
- **Homebrew Formula**: Syntax and structure validation

#### Execution
```bash
# Test latest release
make test-post-release

# Test specific version
./scripts/post-release-validation.sh v1.1.0

# Direct execution
./scripts/post-release-validation.sh
```

#### Docker Integration
- Uses Docker for Linux binary testing on macOS
- Gracefully degrades when Docker unavailable
- Cross-platform execution validation

#### Reports Generated
- `binary-validation-report.md`: Comprehensive validation results
- `checksums.txt`: Downloaded checksums for verification
- Platform-specific test results and recommendations

## CI/CD Integration

### GitHub Actions Workflow

The CI/CD pipeline includes all test suites with proper dependencies and artifact management:

#### Test Jobs

1. **test**: Standard Go unit tests
2. **cli-tests**: CLI command testing
3. **security-tests**: Security scanning
4. **e2e-homebrew-macos**: Homebrew E2E testing (on releases)
5. **post-release-validation**: Binary validation (on releases)

#### Workflow Dependencies
```yaml
build:
  needs: [test, cli-tests, security-tests]

release:
  needs: [build]

e2e-homebrew-macos:
  needs: [release]

post-release-validation:
  needs: [release]
```

#### Artifact Management
- **Security Reports**: 30-day retention
- **Test Logs**: 7-day retention
- **Validation Reports**: 30-day retention

#### Environment Variables
- `GO_VERSION`: Go version for testing
- `HOMEBREW_PAT`: GitHub token for Homebrew automation
- `GITHUB_TOKEN`: Standard GitHub Actions token

### Trigger Conditions

| Test Suite | Trigger |
|------------|---------|
| Unit Tests | Every push, PR |
| CLI Tests | Every push, PR |
| Security Tests | Every push, PR |
| Homebrew E2E | Tagged releases |
| Post-Release | Tagged releases |

## Local Development Workflow

### Quick Testing Commands

```bash
# Run all test suites
make test-all

# Individual test suites
make test              # Unit tests
make test-cli          # CLI tests
make test-security     # Security tests
make test-homebrew-e2e # Homebrew E2E
make test-post-release # Post-release validation

# Badge system testing
make test-badges           # Legacy badge system
make test-dynamic-badges   # Schneegans dynamic badges integration

# Development workflow
make build             # Build binary
make test-cli          # Test CLI functionality
make test-homebrew-local # Test local build via Homebrew
```

### Development Best Practices

1. **Before Committing**:
   ```bash
   make test test-cli test-security
   ```

2. **Before Releasing**:
   ```bash
   make test-all
   ./scripts/test-homebrew-e2e.sh --dry-run
   ```

3. **After Release**:
   ```bash
   ./scripts/post-release-validation.sh
   ```

### Continuous Development

```bash
# Watch mode for unit tests
go test -watch ./...

# Quick CLI validation
make build && ./build/pivot version

# Security scan during development
make test-security
```

## Test Configuration and Environment

### Environment Variables

#### Global Test Configuration
- `GO_VERSION`: Go version for testing
- `CI`: Indicates CI environment
- `GITHUB_TOKEN`: GitHub API access

#### Homebrew E2E Configuration
- `DRY_RUN`: Enable dry-run mode
- `HOMEBREW_E2E_VERBOSE`: Enable verbose logging
- `HOMEBREW_E2E_NO_CLEANUP`: Skip cleanup after test
- `HOMEBREW_E2E_FORCE`: Force test even if prerequisites fail

#### Security Test Configuration
- Security tools auto-install in CI
- Graceful degradation for missing tools
- CI-friendly error handling

#### Post-Release Configuration
- `TEST_VERSION`: Version to validate (default: latest)
- `GITHUB_REPO`: Repository for testing (default: rhino11/pivot)
- Docker availability detection

### Prerequisites

#### Local Development
- Go 1.19+ installed
- macOS (for Homebrew testing)
- Homebrew installed (for E2E tests)
- Internet connection for release testing

#### CI Environment
- Ubuntu/macOS runners
- Go environment configured
- Security tools auto-installed
- Docker available for cross-platform testing

### Test Data and Fixtures

#### Test Configurations
- Mock configuration files for testing
- Temporary directories for isolation
- Cleanup mechanisms for all test data

#### Network Dependencies
- GitHub API for release information
- Package manager repositories
- External security databases

## Troubleshooting

### Common Issues

#### Homebrew E2E Test Failures

**Issue**: Test fails with "tap not found"
```bash
# Solution: Check tap status
brew tap | grep rhino11
# Or reset
brew untap rhino11/tap 2>/dev/null || true
```

**Issue**: Formula audit failures
```bash
# Solution: Manual audit
brew audit --strict rhino11/tap/pivot
```

#### Security Test Issues

**Issue**: Tools not found in CI
- Security tools auto-install in CI
- Check CI logs for installation issues
- Tools are optional and degrade gracefully

**Issue**: False positive security findings
- Review `security-report.md` for details
- Check `gosec-report.json` for specific issues
- Security findings may be informational

#### CLI Test Failures

**Issue**: Binary compilation fails
```bash
# Solution: Check Go environment
go version
go env GOPATH
make clean build
```

**Issue**: Command tests fail
```bash
# Solution: Test manually
./build/pivot version
./build/pivot help
```

#### Post-Release Validation Issues

**Issue**: Release not found
- Wait for GitHub release to be fully published
- Check release URL manually
- Verify version format (vX.Y.Z)

**Issue**: Docker unavailable for Linux testing
- Install Docker for cross-platform testing
- Tests will skip Linux execution gracefully

### Debug Commands

```bash
# Check test environment
go version
brew --version
docker --version

# Manual test execution
./scripts/test-cli.sh
./scripts/test-homebrew-e2e.sh --dry-run
./scripts/security-test.sh
./scripts/post-release-validation.sh

# CI debug
# Check GitHub Actions logs
# Review uploaded artifacts
# Verify environment variables
```

### Test Isolation

All tests are designed for isolation:
- Temporary directories for file operations
- Cleanup mechanisms for all installations
- No persistent state between test runs
- Safe concurrent execution

### Recovery Procedures

#### Reset Test Environment
```bash
# Clean all test installations
make test-cleanup

# Reset Homebrew state
brew untap rhino11/tap 2>/dev/null || true
brew uninstall pivot 2>/dev/null || true

# Clean build artifacts
make clean
```

#### Re-run Failed Tests
```bash
# Individual test re-run
make test-cli
make test-security
make test-homebrew-e2e

# Full test suite re-run
make test-all
```

## Best Practices

### Development Testing
1. Run `make test test-cli` before committing
2. Use `--dry-run` modes for safe testing
3. Test locally before pushing to CI
4. Review security reports regularly

### Release Testing
1. Run full test suite before release
2. Monitor CI test results
3. Validate post-release artifacts
4. Test installation methods manually

### CI/CD Testing
1. Keep test execution time reasonable
2. Upload artifacts for debugging
3. Use appropriate test parallelization
4. Handle flaky tests gracefully

### Security Testing
1. Regular security scans
2. Review and address findings
3. Keep security tools updated
4. Monitor for new vulnerabilities

## Conclusion

The Pivot CLI project includes a comprehensive testing infrastructure that covers all aspects of software quality, security, and user experience. This multi-layered approach ensures:

- **Code Quality**: Unit tests and static analysis
- **Security**: Vulnerability scanning and security best practices
- **User Experience**: E2E testing across platforms and package managers
- **Release Quality**: Post-release validation of all artifacts
- **Developer Experience**: Local testing capabilities and CI integration

The testing infrastructure is designed to be:
- **Comprehensive**: Covers all aspects of the application
- **Automated**: Integrated into CI/CD pipeline
- **Developer-Friendly**: Easy to run locally
- **Safe**: Includes cleanup and isolation mechanisms
- **Scalable**: Can be extended for new platforms and features

This testing approach ensures high-quality releases and provides confidence in the reliability and security of the Pivot CLI across all supported platforms and installation methods.
