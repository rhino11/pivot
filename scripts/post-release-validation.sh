#!/bin/bash

# Post-Release Binary Validation E2E Test Suite
# =============================================
# This script validates all binaries from a GitHub release across platforms

# Note: Not using 'set -e' here because we want to handle test failures gracefully
# through the run_test framework

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
GITHUB_REPO="rhino11/pivot"
TEST_VERSION="${1:-latest}"
TEMP_DIR=""
DOCKER_AVAILABLE=false

# Helper functions
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_step() {
    echo -e "${BLUE}🔧 $1${NC}"
}

# Test result tracking
TESTS_PASSED=0
TESTS_FAILED=0
FAILED_TESTS=()

run_test() {
    local test_name="$1"
    local test_function="$2"
    
    log_step "Running test: $test_name"
    
    if $test_function; then
        log_success "$test_name: PASSED"
        ((TESTS_PASSED++))
        return 0
    else
        log_error "$test_name: FAILED"
        ((TESTS_FAILED++))
        FAILED_TESTS+=("$test_name")
        return 1
    fi
}

# Setup function
setup() {
    log_info "🧪 Post-Release Binary Validation E2E Test Suite"
    echo "Target version: $TEST_VERSION"
    echo "Repository: $GITHUB_REPO"
    echo
    
    # Create temporary directory
    TEMP_DIR=$(mktemp -d)
    log_info "Working directory: $TEMP_DIR"
    
    # Check if Docker is available for Linux testing on macOS
    if command -v docker &> /dev/null && docker info &> /dev/null; then
        DOCKER_AVAILABLE=true
        log_info "Docker available for cross-platform testing"
    else
        log_warning "Docker not available - will skip Linux testing on macOS"
    fi
    
    cd "$TEMP_DIR"
}

# Get release information
get_release_info() {
    log_step "Fetching release information..."
    
    if [ "$TEST_VERSION" = "latest" ]; then
        RELEASE_URL="https://api.github.com/repos/$GITHUB_REPO/releases/latest"
    else
        RELEASE_URL="https://api.github.com/repos/$GITHUB_REPO/releases/tags/$TEST_VERSION"
    fi
    
    if ! curl -s -f "$RELEASE_URL" > release.json; then
        log_error "Failed to fetch release information"
        return 1
    fi
    
    VERSION=$(jq -r '.tag_name' release.json)
    if [ "$VERSION" = "null" ]; then
        log_error "Could not extract version from release"
        return 1
    fi
    
    log_success "Found release: $VERSION"
    return 0
}

# Download binary
download_binary() {
    local platform="$1"
    local architecture="$2"
    local extension="$3"
    
    local binary_name="pivot-${platform}-${architecture}${extension}"
    local download_url="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$binary_name"
    
    log_info "Downloading $binary_name..."
    
    if curl -L -f -o "$binary_name" "$download_url"; then
        log_success "Downloaded $binary_name"
        return 0
    else
        log_error "Failed to download $binary_name"
        return 1
    fi
}

# Verify binary integrity
verify_binary_integrity() {
    local binary_name="$1"
    
    log_info "Verifying integrity of $binary_name..."
    
    # Check file exists and has content
    if [ ! -f "$binary_name" ]; then
        log_error "Binary file not found: $binary_name"
        return 1
    fi
    
    local file_size=$(stat -c%s "$binary_name" 2>/dev/null || stat -f%z "$binary_name" 2>/dev/null)
    if [ "$file_size" -lt 1000000 ]; then  # Less than 1MB seems too small
        log_warning "Binary seems unusually small: ${file_size} bytes"
    fi
    
    # Verify SHA256 if checksums are available
    if [ -f "../checksums.txt" ]; then
        expected_hash=$(grep "$binary_name" ../checksums.txt | cut -d' ' -f1)
        if [ -n "$expected_hash" ]; then
            actual_hash=$(sha256sum "$binary_name" | cut -d' ' -f1)
            if [ "$expected_hash" = "$actual_hash" ]; then
                log_success "SHA256 verification passed"
            else
                log_error "SHA256 verification failed"
                log_error "Expected: $expected_hash"
                log_error "Actual: $actual_hash"
                return 1
            fi
        fi
    fi
    
    log_success "Binary integrity verified"
    return 0
}

# Test binary functionality
test_binary_functionality() {
    local binary_name="$1"
    local platform="$2"
    local arch="$3"
    
    log_info "Testing functionality of $binary_name..."
    
    # Make executable
    chmod +x "$binary_name"
    
    # Test version command
    if [ "$platform" = "windows" ]; then
        # Skip execution test on non-Windows platforms
        log_info "Skipping execution test for Windows binary on non-Windows platform"
        return 0
    elif [ "$platform" = "linux" ] && [ "$(uname -s)" = "Darwin" ]; then
        # Use Docker if available
        if [ "$DOCKER_AVAILABLE" = true ]; then
            log_info "Testing Linux binary with Docker..."
            if docker run --rm -v "$(pwd):/test" ubuntu:latest /test/"$binary_name" version; then
                log_success "Linux binary executed successfully in Docker"
            else
                log_error "Linux binary failed to execute in Docker"
                return 1
            fi
        else
            log_info "Skipping Linux binary execution test (no Docker)"
            return 0
        fi
    elif [ "$platform" = "darwin" ] && [ "$(uname -s)" = "Darwin" ]; then
        # Check if we can run this architecture natively
        local current_arch="$(uname -m)"
        if [ "$current_arch" = "x86_64" ] && [ "$arch" = "arm64" ]; then
            log_info "Skipping ARM64 binary execution test on Intel Mac"
            return 0
        elif [ "$current_arch" = "arm64" ] && [ "$arch" = "amd64" ]; then
            # ARM64 Macs can usually run x86_64 binaries via Rosetta
            log_info "Testing x86_64 binary on ARM64 Mac (Rosetta)"
        fi
        
        # Native execution
        log_info "Testing native execution..."
        
        # Test version command
        if ./"$binary_name" version &> /dev/null; then
            version_output=$(./"$binary_name" version 2>&1)
            log_info "Version output: $version_output"
            
            # Verify version output contains expected elements
            if [[ "$version_output" =~ "pivot version" ]] && [[ "$version_output" =~ "commit:" ]]; then
                log_success "Version command works correctly"
            else
                log_error "Version command output is malformed"
                return 1
            fi
        else
            log_error "Binary failed to execute version command"
            return 1
        fi
        
        # Test help command
        if ./"$binary_name" help &> /dev/null; then
            log_success "Help command works correctly"
        else
            log_error "Help command failed"
            return 1
        fi
        
        # Test invalid command (should fail gracefully)
        if ./"$binary_name" nonexistent-command &> /dev/null; then
            log_error "Binary should reject invalid commands"
            return 1
        else
            log_success "Binary correctly rejects invalid commands"
        fi
    fi
    
    return 0
}

# Test macOS binaries
test_macos_binaries() {
    log_step "Testing macOS binaries..."
    
    local architectures=("amd64" "arm64")
    local platform="darwin"
    
    for arch in "${architectures[@]}"; do
        local binary_name="pivot-${platform}-${arch}"
        
        if download_binary "$platform" "$arch" ""; then
            if verify_binary_integrity "$binary_name"; then
                if test_binary_functionality "$binary_name" "$platform" "$arch"; then
                    log_success "macOS $arch binary test passed"
                else
                    log_error "macOS $arch binary functionality test failed"
                    return 1
                fi
            else
                log_error "macOS $arch binary integrity check failed"
                return 1
            fi
        else
            log_error "Failed to download macOS $arch binary"
            return 1
        fi
    done
    
    return 0
}

# Test Linux binaries
test_linux_binaries() {
    log_step "Testing Linux binaries..."
    
    local architectures=("amd64" "arm64")
    local platform="linux"
    
    for arch in "${architectures[@]}"; do
        local binary_name="pivot-${platform}-${arch}"
        
        if download_binary "$platform" "$arch" ""; then
            if verify_binary_integrity "$binary_name"; then
                if test_binary_functionality "$binary_name" "$platform" "$arch"; then
                    log_success "Linux $arch binary test passed"
                else
                    log_error "Linux $arch binary functionality test failed"
                    return 1
                fi
            else
                log_error "Linux $arch binary integrity check failed"
                return 1
            fi
        else
            log_error "Failed to download Linux $arch binary"
            return 1
        fi
    done
    
    return 0
}

# Test Windows binaries
test_windows_binaries() {
    log_step "Testing Windows binaries..."
    
    local architectures=("amd64" "arm64")
    local platform="windows"
    
    for arch in "${architectures[@]}"; do
        local binary_name="pivot-${platform}-${arch}.exe"
        
        if download_binary "$platform" "$arch" ".exe"; then
            if verify_binary_integrity "$binary_name"; then
                if test_binary_functionality "$binary_name" "$platform" "$arch"; then
                    log_success "Windows $arch binary test passed"
                else
                    log_error "Windows $arch binary functionality test failed"
                    return 1
                fi
            else
                log_error "Windows $arch binary integrity check failed"
                return 1
            fi
        else
            log_error "Failed to download Windows $arch binary"
            return 1
        fi
    done
    
    return 0
}

# Test package files
test_package_files() {
    log_step "Testing package files..."
    
    # Extract version number without 'v' prefix for DEB package name
    local version_num="${VERSION#v}"
    local packages=("pivot_${version_num}_amd64.deb" "pivot.rb")
    local issues=0
    
    for package in "${packages[@]}"; do
        local download_url="https://github.com/$GITHUB_REPO/releases/download/$VERSION/$package"
        
        log_info "Testing package: $package"
        
        if curl -L -f -o "$package" "$download_url"; then
            case "$package" in
                *.deb)
                    # Test DEB package structure
                    if command -v dpkg-deb &> /dev/null; then
                        if dpkg-deb --info "$package" &> /dev/null; then
                            log_success "DEB package structure is valid"
                        else
                            log_error "DEB package structure is invalid"
                            issues=$((issues + 1))
                        fi
                    else
                        log_info "dpkg-deb not available, skipping DEB validation"
                    fi
                    ;;
                *.rb)
                    # Test Homebrew formula syntax
                    if command -v ruby &> /dev/null; then
                        if ruby -c "$package" &> /dev/null; then
                            log_success "Homebrew formula syntax is valid"
                        else
                            log_error "Homebrew formula syntax is invalid"
                            issues=$((issues + 1))
                        fi
                    else
                        log_info "Ruby not available, skipping Homebrew formula validation"
                    fi
                    ;;
            esac
        else
            log_warning "Could not download package: $package"
            issues=$((issues + 1))
        fi
    done
    
    if [ $issues -eq 0 ]; then
        log_success "All package files tested successfully"
        return 0
    else
        log_error "$issues package file issues found"
        return 1
    fi
}

# Test checksums file
test_checksums() {
    log_step "Testing checksums file..."
    
    local checksums_url="https://github.com/$GITHUB_REPO/releases/download/$VERSION/checksums.txt"
    
    if curl -L -f -o checksums.txt "$checksums_url"; then
        log_success "Downloaded checksums file"
        
        # Verify checksums file format
        if grep -E "^[a-f0-9]{64}  " checksums.txt &> /dev/null; then
            log_success "Checksums file format is valid"
            
            # Count entries
            local checksum_count=$(wc -l < checksums.txt)
            log_info "Found $checksum_count checksums"
            
            if [ "$checksum_count" -lt 6 ]; then  # Expect at least 6 binaries
                log_warning "Fewer checksums than expected: $checksum_count"
            else
                log_success "Checksum count looks good: $checksum_count"
            fi
        else
            log_error "Checksums file format is invalid"
            return 1
        fi
    else
        log_error "Could not download checksums file"
        return 1
    fi
    
    return 0
}

# Generate test report
generate_test_report() {
    log_step "Generating test report..."
    
    local report_file="binary-validation-report.md"
    
    cat > "$report_file" << EOF
# Binary Validation Report

**Version:** $VERSION  
**Test Date:** $(date)  
**Test Platform:** $(uname -s) $(uname -m)  

## Test Results

### Summary
- Tests Passed: $TESTS_PASSED
- Tests Failed: $TESTS_FAILED
- Docker Available: $DOCKER_AVAILABLE

### Failed Tests
EOF
    
    if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
        echo "None - All tests passed! ✅" >> "$report_file"
    else
        for test in "${FAILED_TESTS[@]}"; do
            echo "- $test" >> "$report_file"
        done
    fi
    
    cat >> "$report_file" << EOF

### Tested Binaries
- macOS AMD64
- macOS ARM64
- Linux AMD64
- Linux ARM64
- Windows AMD64
- Windows ARM64

### Tested Packages
- DEB package
- Homebrew formula

### Test Coverage
- Binary download and integrity
- SHA256 verification
- Basic functionality (version, help)
- Invalid command handling
- Package file validation

## Recommendations

EOF
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo "✅ All validation tests passed. Release is ready for distribution." >> "$report_file"
    else
        echo "⚠️ Some validation tests failed. Please review and fix issues before distribution." >> "$report_file"
        echo "" >> "$report_file"
        echo "Failed tests:" >> "$report_file"
        for test in "${FAILED_TESTS[@]}"; do
            echo "- $test" >> "$report_file"
        done
    fi
    
    log_success "Test report generated: $report_file"
}

# Cleanup function
cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        log_info "Cleaning up temporary directory: $TEMP_DIR"
        rm -rf "$TEMP_DIR"
    fi
}

# Main execution
main() {
    setup
    
    # Get release information
    run_test "Release Information" get_release_info || exit 1
    
    # Download checksums first
    run_test "Checksums File" test_checksums
    
    # Test all platforms
    run_test "macOS Binaries" test_macos_binaries
    run_test "Linux Binaries" test_linux_binaries  
    run_test "Windows Binaries" test_windows_binaries
    run_test "Package Files" test_package_files
    
    # Generate report
    generate_test_report
    
    echo
    log_info "🎯 Post-Release Validation Summary"
    echo "=================================="
    log_info "Version tested: $VERSION"
    log_success "Tests passed: $TESTS_PASSED"
    
    if [ $TESTS_FAILED -eq 0 ]; then
        log_success "All post-release validation tests passed!"
        echo
        log_info "📝 Full report: binary-validation-report.md"
        return 0
    else
        log_error "Tests failed: $TESTS_FAILED"
        echo
        log_error "Failed tests:"
        for test in "${FAILED_TESTS[@]}"; do
            log_error "  - $test"
        done
        echo
        log_info "📝 Full report: binary-validation-report.md"
        return 1
    fi
}

# Set up cleanup trap
trap cleanup EXIT

# Show usage if help requested
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: $0 [version]"
    echo ""
    echo "Post-Release Binary Validation E2E Test Suite"
    echo ""
    echo "Arguments:"
    echo "  version    Version to test (default: latest)"
    echo ""
    echo "Examples:"
    echo "  $0              # Test latest release"
    echo "  $0 v1.0.3       # Test specific version"
    exit 0
fi

# Run main function
main "$@"
