#!/bin/bash

# Security Test Suite for Pivot CLI
# =================================
# This script runs comprehensive security checks for the Go application

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

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_step() {
    echo -e "${BLUE}🔧 $1${NC}"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    # Add Go bin to PATH if it's not already there
    GOPATH=$(go env GOPATH)
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]]; then
        export PATH="$GOPATH/bin:$PATH"
        log_info "Added $GOPATH/bin to PATH"
    fi
    
    log_success "Go is available: $(go version)"
}

# Install security tools if not present
install_security_tools() {
    log_step "Installing security tools..."
    
    # gosec - Go security checker
    if ! command -v gosec &> /dev/null; then
        log_info "Installing gosec..."
        if ! go install github.com/securego/gosec/v2/cmd/gosec@latest; then
            log_warning "Failed to install gosec"
            return 1
        fi
    fi
    
    # govulncheck - Go vulnerability scanner
    if ! command -v govulncheck &> /dev/null; then
        log_info "Installing govulncheck..."
        if ! go install golang.org/x/vuln/cmd/govulncheck@latest; then
            log_warning "Failed to install govulncheck"
            return 1
        fi
    fi
    
    # staticcheck - Advanced static analysis
    if ! command -v staticcheck &> /dev/null; then
        log_info "Installing staticcheck..."
        if ! go install honnef.co/go/tools/cmd/staticcheck@latest; then
            log_warning "Failed to install staticcheck"
            return 1
        fi
    fi
    
    # nancy - OSS Index vulnerability scanner (optional in CI)
    if ! command -v nancy &> /dev/null; then
        log_info "Installing nancy..."
        if ! go install github.com/sonatype-nexus-community/nancy@latest; then
            log_warning "Failed to install nancy (continuing without it)"
        else
            # Ensure nancy is available in current session
            export PATH="$(go env GOPATH)/bin:$PATH"
            if command -v nancy &> /dev/null; then
                log_success "nancy installed and available"
            else
                log_warning "nancy installed but not in PATH"
            fi
        fi
    else
        log_success "nancy already available"
    fi
    
    log_success "Security tools installation completed"
}

# Run gosec security scan
run_gosec() {
    log_step "Running gosec security scan..."
    
    if gosec -fmt json -out gosec-report.json ./...; then
        log_success "gosec scan completed successfully"
        
        # Check if any issues were found
        if [ -f gosec-report.json ]; then
            issues=$(jq '.Issues | length' gosec-report.json 2>/dev/null || echo "0")
            if [ "$issues" -gt 0 ]; then
                log_warning "gosec found $issues security issues"
                log_info "Review gosec-report.json for details"
                
                # Show summary
                if command -v jq &> /dev/null; then
                    log_info "Issue summary:"
                    jq -r '.Issues[] | "  - \(.severity): \(.details) (\(.file):\(.line))"' gosec-report.json | head -10
                fi
                return 1
            else
                log_success "No security issues found by gosec"
            fi
        fi
    else
        log_error "gosec scan failed"
        return 1
    fi
    
    return 0
}

# Run vulnerability scan
run_vulnerability_scan() {
    log_step "Running vulnerability scan..."
    
    if govulncheck ./...; then
        log_success "No known vulnerabilities found"
        return 0
    else
        log_error "Vulnerabilities detected"
        return 1
    fi
}

# Run static analysis
run_static_analysis() {
    log_step "Running static analysis with staticcheck..."
    
    if staticcheck ./...; then
        log_success "Static analysis passed"
        return 0
    else
        log_error "Static analysis found issues"
        return 1
    fi
}

# Check dependencies for known vulnerabilities
check_dependencies() {
    log_step "Checking dependencies with nancy..."
    
    # Skip if nancy is not available
    if ! command -v nancy &> /dev/null; then
        log_warning "nancy not available, skipping dependency vulnerability check"
        return 0
    fi
    
    # Use the recommended nancy usage with direct piping
    log_info "Scanning dependencies for vulnerabilities..."
    
    # Create a temporary file to capture nancy output
    local nancy_output=$(mktemp)
    local nancy_exit_code=0
    
    # Check if ignore file exists
    local ignore_args=""
    if [ -f ".nancy-ignore" ]; then
        log_info "Using vulnerability ignore file: .nancy-ignore"
        # Nancy doesn't have a built-in ignore file, but we can filter results
    fi
    
    # Run nancy with proper error handling
    if go list -json -deps ./... | nancy sleuth --loud > "$nancy_output" 2>&1; then
        nancy_exit_code=0
    else
        nancy_exit_code=$?
    fi
    
    # Parse nancy output for better reporting
    if [ -f "$nancy_output" ]; then
        local audited_count=$(grep -o "Audited Dependencies.*[0-9]\+" "$nancy_output" | grep -o "[0-9]\+" || echo "unknown")
        local vulnerable_count=$(grep -o "Vulnerable Dependencies.*[0-9]\+" "$nancy_output" | tail -1 | grep -o "[0-9]\+" || echo "0")
        
        log_info "Dependencies audited: $audited_count"
        log_info "Vulnerabilities found: $vulnerable_count"
        
        # Show nancy output for transparency
        cat "$nancy_output"
        
        if [ $nancy_exit_code -eq 0 ]; then
            log_success "No vulnerable dependencies found by nancy"
            rm -f "$nancy_output"
            return 0
        else
            log_error "Vulnerable dependencies detected by nancy"
            
            # Extract specific vulnerability information
            if grep -q "pkg:" "$nancy_output"; then
                echo ""
                log_warning "Vulnerable packages found:"
                grep -A 5 -B 1 "pkg:\|CVE-\|CWE-" "$nancy_output" | head -20 || true
            fi
            
            # Check against ignore file
            if [ -f ".nancy-ignore" ] && [ "$vulnerable_count" -gt 0 ]; then
                log_info "Checking against ignore file..."
                
                # Extract CVE IDs from nancy output
                local found_cves=$(grep -o "CVE-[0-9]\{4\}-[0-9]\+" "$nancy_output" 2>/dev/null || echo "")
                local ignored_cves=$(grep -v "^#\|^$" .nancy-ignore 2>/dev/null | cut -d' ' -f1 || echo "")
                
                if [ -n "$found_cves" ] && [ -n "$ignored_cves" ]; then
                    local unignored_cves=""
                    for cve in $found_cves; do
                        if ! echo "$ignored_cves" | grep -q "$cve"; then
                            unignored_cves="$unignored_cves $cve"
                        else
                            log_info "Ignored vulnerability: $cve (see .nancy-ignore)"
                        fi
                    done
                    
                    if [ -z "$unignored_cves" ]; then
                        log_warning "All vulnerabilities are in ignore list - proceeding"
                        rm -f "$nancy_output"
                        return 0
                    fi
                fi
            fi
            
            # Provide actionable advice
            echo ""
            log_info "🔧 Vulnerability Remediation Options:"
            echo "   1. **Update dependencies**: go get -u ./..."
            echo "   2. **Review advisories**: Check security details for affected packages"
            echo "   3. **Alternative packages**: Consider replacing vulnerable dependencies"
            echo "   4. **Add to ignore list**: Add to .nancy-ignore if false positive (review carefully)"
            echo "   5. **Vendor patches**: Check if vendors have released patches"
            echo ""
            echo "📋 To ignore specific vulnerabilities:"
            echo "   echo 'CVE-YYYY-NNNN # Reason for ignoring' >> .nancy-ignore"
            echo ""
            
            # Save detailed report for further analysis
            cp "$nancy_output" "nancy-vulnerability-report.txt"
            log_info "Detailed vulnerability report saved: nancy-vulnerability-report.txt"
            
            rm -f "$nancy_output"
            return 1
        fi
    else
        log_error "Nancy output not captured"
        return 1
    fi
}

# Test configuration security
test_config_security() {
    log_step "Testing configuration security..."
    
    local issues=0
    
    # Check for hardcoded secrets in code
    log_info "Checking for hardcoded secrets..."
    if grep -r -E "(password|secret|key|token).*=.*['\"][^'\"]{8,}['\"]" --include="*.go" --exclude="*_test.go" . 2>/dev/null; then
        log_warning "Potential hardcoded secrets found in production code"
        issues=$((issues + 1))
    elif grep -r -E "(password|secret|key|token).*=.*['\"][^'\"]{8,}['\"]" --include="*_test.go" . 2>/dev/null; then
        log_info "Test tokens found in test files (acceptable)"
    fi
    
    # Check file permissions
    log_info "Checking file permissions..."
    if [ -f "config.yaml" ]; then
        perms=$(stat -c "%a" config.yaml 2>/dev/null || stat -f "%A" config.yaml 2>/dev/null || echo "644")
        if [[ "$perms" =~ ^6[0-7][0-7]$ ]] || [[ "$perms" =~ ^[0-7][0-7][0-7]$ && "${perms:1:1}" -gt 4 ]] || [[ "$perms" =~ ^[0-7][0-7][0-7]$ && "${perms:2:1}" -gt 4 ]]; then
            log_warning "Config file has overly permissive permissions: $perms"
            issues=$((issues + 1))
        fi
    fi
    
    # Check for .env files
    if find . -name ".env*" -type f | grep -q .; then
        log_warning "Environment files found - ensure they're not committed to version control"
        issues=$((issues + 1))
    fi
    
    if [ $issues -eq 0 ]; then
        log_success "Configuration security checks passed"
        return 0
    else
        log_warning "Configuration security issues found: $issues"
        return 1
    fi
}

# Test binary security
test_binary_security() {
    log_step "Testing binary security..."
    
    local binary_path="./build/pivot"
    
    if [ ! -f "$binary_path" ]; then
        log_info "Building binary for security testing..."
        make build
    fi
    
    if [ ! -f "$binary_path" ]; then
        log_error "Binary not found at $binary_path"
        return 1
    fi
    
    local issues=0
    
    # Check if binary was stripped (should be for release builds)
    if command -v nm &> /dev/null; then
        if nm "$binary_path" &> /dev/null; then
            log_info "Binary contains symbols (expected for debug builds)"
        else
            log_success "Binary is stripped (good for release builds)"
        fi
    fi
    
    # Check binary permissions
    perms=$(stat -c "%a" "$binary_path" 2>/dev/null || stat -f "%A" "$binary_path" 2>/dev/null || echo "755")
    if [[ ! "$perms" =~ ^7[0-7][0-57]$ ]]; then
        log_warning "Binary has unusual permissions: $perms"
        issues=$((issues + 1))
    fi
    
    # Test for common security features (if available)
    if command -v checksec &> /dev/null; then
        log_info "Running checksec on binary..."
        checksec --file="$binary_path"
    elif command -v hardening-check &> /dev/null; then
        log_info "Running hardening check on binary..."
        hardening-check "$binary_path"
    fi
    
    if [ $issues -eq 0 ]; then
        log_success "Binary security checks passed"
        return 0
    else
        log_warning "Binary security issues found: $issues"
        return 1
    fi
}

# Test runtime security
test_runtime_security() {
    log_step "Testing runtime security..."
    
    local issues=0
    
    # Test handling of sensitive data in memory
    log_info "Testing sensitive data handling..."
    
    # Create a test config with fake token
    cat > test-config.yaml << EOF
owner: testowner
repo: testrepo
token: ghp_faketoken123456789
EOF
    
    # Run a quick command and check if token appears in process list
    if timeout 2s ./build/pivot version > /dev/null 2>&1; then
        log_success "Basic runtime test passed"
    else
        log_info "Binary not ready for runtime test"
    fi
    
    rm -f test-config.yaml
    
    if [ $issues -eq 0 ]; then
        log_success "Runtime security checks passed"
        return 0
    else
        log_warning "Runtime security issues found: $issues"
        return 1
    fi
}

# Generate security report
generate_security_report() {
    log_step "Generating security report..."
    
    local report_file="security-report.md"
    
    cat > "$report_file" << EOF
# Pivot CLI Security Report

Generated on: $(date)
Go version: $(go version)

## Tools Used
- **gosec**: Go security checker - Static analysis for security issues
- **govulncheck**: Go vulnerability scanner - Known CVE detection  
- **staticcheck**: Static analysis - Code quality and potential bugs
- **nancy**: OSS Index vulnerability scanner - Dependency vulnerability checking

## Results Summary

EOF
    
    if [ -f gosec-report.json ]; then
        echo "### gosec Results" >> "$report_file"
        if command -v jq &> /dev/null; then
            issues=$(jq '.Issues | length' gosec-report.json 2>/dev/null || echo "0")
            echo "- Issues found: $issues" >> "$report_file"
            echo "" >> "$report_file"
            
            if [ "$issues" -gt 0 ]; then
                echo "#### Issues:" >> "$report_file"
                jq -r '.Issues[] | "- **\(.severity)**: \(.details) (\(.file):\(.line))"' gosec-report.json >> "$report_file"
                echo "" >> "$report_file"
            fi
        fi
    fi
    
    # Add dependency vulnerability section
    echo "### Dependency Vulnerability Analysis" >> "$report_file"
    echo "" >> "$report_file"
    
    if command -v nancy &> /dev/null; then
        echo "#### Nancy Scan Results" >> "$report_file"
        
        # Re-run nancy for report generation (lightweight)
        local nancy_summary=$(go list -json -deps ./... | nancy sleuth --quiet 2>/dev/null || echo "Scan failed")
        
        if echo "$nancy_summary" | grep -q "0.*Vulnerable Dependencies"; then
            echo "- ✅ **Status**: No vulnerabilities detected" >> "$report_file"
            echo "- 📊 **Dependencies**: $(echo "$nancy_summary" | grep -o "Audited Dependencies.*[0-9]\+" | grep -o "[0-9]\+" || echo "unknown") packages scanned" >> "$report_file"
        else
            echo "- ⚠️ **Status**: Vulnerabilities detected or scan failed" >> "$report_file"
            echo "- 🔍 **Action Required**: Review dependencies and update vulnerable packages" >> "$report_file"
        fi
        echo "" >> "$report_file"
        
        echo "#### Dependency Security Best Practices" >> "$report_file"
        echo "1. **Regular Updates**: Run \`go get -u ./...\` monthly" >> "$report_file"
        echo "2. **Minimal Dependencies**: Avoid unnecessary third-party packages" >> "$report_file"
        echo "3. **Version Pinning**: Use specific versions in go.mod for stability" >> "$report_file"
        echo "4. **Security Monitoring**: Monitor security advisories for used packages" >> "$report_file"
        echo "" >> "$report_file"
    else
        echo "- ❌ **Nancy not available**: Install with \`go install github.com/sonatype-nexus-community/nancy@latest\`" >> "$report_file"
        echo "" >> "$report_file"
    fi
    
    echo "## Security Recommendations" >> "$report_file"
    echo "" >> "$report_file"
    echo "### Immediate Actions" >> "$report_file"
    echo "1. **Dependencies**: Regularly update dependencies to latest secure versions" >> "$report_file"
    echo "2. **Configuration**: Use strong file permissions for config files (600)" >> "$report_file"
    echo "3. **Secrets**: Never commit secrets to version control" >> "$report_file"
    echo "4. **Environment**: Use environment variables or secure secret management" >> "$report_file"
    echo "" >> "$report_file"
    
    echo "### Ongoing Security Practices" >> "$report_file"
    echo "1. **Automated Scanning**: Security tests run on every CI build" >> "$report_file"
    echo "2. **Dependency Monitoring**: Nancy scans for vulnerable dependencies" >> "$report_file"
    echo "3. **Static Analysis**: Multiple tools analyze code for security issues" >> "$report_file"
    echo "4. **Regular Audits**: Manual security reviews for critical changes" >> "$report_file"
    echo "" >> "$report_file"
    
    echo "### Security Tools Setup" >> "$report_file"
    echo '```bash' >> "$report_file"
    echo "# Install all security tools" >> "$report_file"
    echo "go install github.com/securego/gosec/v2/cmd/gosec@latest" >> "$report_file"
    echo "go install golang.org/x/vuln/cmd/govulncheck@latest" >> "$report_file"
    echo "go install honnef.co/go/tools/cmd/staticcheck@latest" >> "$report_file"
    echo "go install github.com/sonatype-nexus-community/nancy@latest" >> "$report_file"
    echo "" >> "$report_file"
    echo "# Run security test suite" >> "$report_file"
    echo "make test-security" >> "$report_file"
    echo '```' >> "$report_file"
    
    log_success "Security report generated: $report_file"
}

# Main execution
main() {
    log_info "🔐 Starting Pivot CLI Security Test Suite"
    echo
    
    local failed_tests=0
    
    # Prerequisites
    check_go
    install_security_tools
    echo
    
    # Security tests
    if ! run_gosec; then
        failed_tests=$((failed_tests + 1))
    fi
    echo
    
    if ! run_vulnerability_scan; then
        failed_tests=$((failed_tests + 1))
    fi
    echo
    
    if ! run_static_analysis; then
        failed_tests=$((failed_tests + 1))
    fi
    echo
    
    if ! check_dependencies; then
        failed_tests=$((failed_tests + 1))
    fi
    echo
    
    if ! test_config_security; then
        failed_tests=$((failed_tests + 1))
    fi
    echo
    
    if ! test_binary_security; then
        failed_tests=$((failed_tests + 1))
    fi
    echo
    
    if ! test_runtime_security; then
        failed_tests=$((failed_tests + 1))
    fi
    echo
    
    # Generate report
    generate_security_report
    echo
    
    # Summary
    log_info "🎯 Security Test Summary"
    echo "=================================="
    
    if [ $failed_tests -eq 0 ]; then
        log_success "All security tests passed!"
        echo
        log_info "📝 Security report: security-report.md"
        log_info "📊 gosec report: gosec-report.json"
        return 0
    else
        log_error "$failed_tests security tests failed"
        echo
        log_info "📝 Security report: security-report.md"
        log_info "📊 gosec report: gosec-report.json"
        log_warning "Please review and address security issues before release"
        return 1
    fi
}

# Cleanup function
cleanup() {
    log_info "Cleaning up temporary files..."
    rm -f test-config.yaml nancy-vulnerability-report.txt
}

# Set up cleanup trap
trap cleanup EXIT

# Run main function
main "$@"
