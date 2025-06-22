# Makefile for Pivot - Go GitHub Issues Manager
# ==============================================

# Variables
PROJECT_NAME = pivot
BINARY_NAME = pivot
BUILD_DIR = build
DIST_DIR = dist
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

# Go variables
GO_CMD = go
GO_MAIN = ./cmd/main.go
GO_BUILD_FLAGS = -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"
GO_TEST_FLAGS = -v -race -coverprofile=coverage.out

# Platform targets
PLATFORMS = \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64 \
	windows/arm64

# Default target
.PHONY: all
all: clean check-go deps build

# Help target
.PHONY: help
help:
	@echo "Pivot Project Makefile"
	@echo "======================"
	@echo ""
	@echo "Available targets:"
	@echo "  all           - Clean, check Go, install dependencies, and build"
	@echo "  build         - Build binary for current platform"
	@echo "  build-all     - Build binaries for all platforms"
	@echo "  clean         - Remove build artifacts"
	@echo "  check-go      - Check if Go is installed and install if missing"
	@echo "  deps          - Install Go dependencies"
	@echo "  fix-go        - Troubleshoot and fix common Go issues"
	@echo "  check-gosec   - Check if gosec is installed and install if missing"
	@echo "  test          - Run tests"
	@echo "  lint          - Run Go linter"
	@echo "  coverage      - Generate test coverage report"
	@echo "  format        - Format Go code"
	@echo "  run           - Install to user PATH and run pivot"
	@echo "  run-local     - Run pivot from build directory"
	@echo "  install-user  - Install to user's ~/bin (no sudo)"
	@echo "  install       - Install to system PATH (requires sudo)"
	@echo "  uninstall-user- Remove from user's ~/bin"
	@echo "  uninstall     - Remove from system PATH (requires sudo)"
	@echo "  release       - Create release binaries with checksums"
	@echo "  test-e2e      - Run end-to-end tests"
	@echo "  test-homebrew-e2e - Run Homebrew E2E test (macOS only)"
	@echo "  test-homebrew-local - Test local build via Homebrew"
	@echo "  test-cleanup  - Clean up test installations"
	@echo "  test-e2e      - Run end-to-end tests"
	@echo "  test-homebrew-e2e - Run Homebrew end-to-end test (macOS only)"
	@echo "  test-homebrew-local - Test local build via Homebrew"
	@echo "  test-cleanup  - Cleanup test installations"
	@echo "  test-cli      - Run comprehensive CLI test suite"
	@echo "  test-security - Run security test suite"
	@echo "  test-badges   - Test badge system locally"
	@echo "  test-dynamic-badges - Test Schneegans dynamic badges integration"
	@echo "  test-post-release - Run post-release validation tests"
	@echo "  test-all      - Run all test suites (unit + CLI + security + E2E)"
	@echo "  submit-homebrew-core - Submit formula to Homebrew Core"
	@echo "  setup-badges  - Setup dynamic badges for repository"
	@echo "  ci            - Run CI pipeline (matches GitHub Actions: deps + lint + test + cli + security + build)"
	@echo "  ci-local      - Run local CI pipeline (includes formatting)"
	@echo "  ci-full       - Run full CI pipeline with coverage, E2E tests, and multi-platform builds"
	@echo "  ci-quick      - Run quick CI validation for development (format + lint + test + cli)"

# CI Pipeline - Match GitHub Actions workflow exactly
.PHONY: ci
ci: clean check-go deps lint test test-cli test-security build
	@echo ""
	@echo "🎉 CI Pipeline completed successfully!"
	@echo "✅ Dependencies installed"
	@echo "✅ Linting passed"
	@echo "✅ Unit tests passed (with coverage)"
	@echo "✅ CLI tests passed"
	@echo "✅ Security tests passed"
	@echo "✅ Build successful"
	@echo ""
	@echo "Ready for commit and push! 🚀"

# Local CI Pipeline - Includes formatting for local development
.PHONY: ci-local
ci-local: clean check-go deps format lint test test-cli test-security build
	@echo ""
	@echo "🎉 Local CI Pipeline completed successfully!"
	@echo "✅ Dependencies installed"
	@echo "✅ Code formatted"
	@echo "✅ Linting passed"
	@echo "✅ Unit tests passed (with coverage)"
	@echo "✅ CLI tests passed"
	@echo "✅ Security tests passed"
	@echo "✅ Build successful"
	@echo ""
	@echo "Ready for commit and push! 🚀"

# Full CI Pipeline - Includes coverage, E2E tests, and multi-platform builds
.PHONY: ci-full
ci-full: clean check-go deps format lint test coverage test-cli test-security build-all test-e2e
	@echo ""
	@echo "🎉 Full CI Pipeline completed successfully!"
	@echo "✅ Dependencies installed"
	@echo "✅ Code formatted"
	@echo "✅ Linting passed"
	@echo "✅ Unit tests passed"
	@echo "✅ Coverage generated"
	@echo "✅ CLI tests passed"
	@echo "✅ Security tests passed"
	@echo "✅ Multi-platform builds successful"
	@echo "✅ E2E tests passed"
	@echo ""
	@echo "Ready for release! 🚀"

# Quick CI - Fast validation for development
.PHONY: ci-quick
ci-quick: format lint test test-cli
	@echo ""
	@echo "🎉 Quick CI completed successfully!"
	@echo "✅ Code formatted"
	@echo "✅ Linting passed"
	@echo "✅ Unit tests passed"
	@echo "✅ CLI tests passed"
	@echo ""
	@echo "Ready for development! 🚀"

# Dependencies
.PHONY: check-go
check-go:
	@echo "🔍 Checking Go installation..."
	@if ! command -v go >/dev/null 2>&1; then \
		echo "❌ Go is not installed. Installing Go..."; \
		$(MAKE) install-go; \
	else \
		echo "🔧 Validating Go installation..."; \
		if ! go version >/dev/null 2>&1; then \
			echo "❌ Go command exists but is not functioning properly"; \
			echo "🔄 Attempting to reinstall Go..."; \
			$(MAKE) install-go; \
		else \
			GO_VERSION=$$(go version 2>/dev/null | cut -d' ' -f3 | sed 's/go//' || echo "unknown"); \
			REQUIRED_VERSION="1.23"; \
			echo "✅ Go is installed: go$$GO_VERSION"; \
			if ! $(MAKE) check-go-version GO_VERSION=$$GO_VERSION REQUIRED_VERSION=$$REQUIRED_VERSION; then \
				echo "⚠️  Go version $$GO_VERSION may be incompatible with this project (requires ≥$$REQUIRED_VERSION)"; \
				echo "🔄 Attempting to install a compatible version..."; \
				$(MAKE) install-go; \
			fi; \
		fi; \
	fi

.PHONY: install-go
install-go:
	@echo "📦 Installing Go 1.23..."
	@GO_VERSION="1.23.10"; \
	if command -v apt-get >/dev/null 2>&1; then \
		echo "📦 Installing Go $$GO_VERSION from official source..."; \
		sudo rm -rf /usr/local/go; \
		wget -qO- "https://go.dev/dl/go$$GO_VERSION.linux-amd64.tar.gz" | sudo tar -C /usr/local -xzf -; \
		echo "✅ Go $$GO_VERSION installed from official source"; \
	elif command -v yum >/dev/null 2>&1; then \
		echo "📦 Using yum to install Go..."; \
		sudo yum install -y golang; \
	elif command -v dnf >/dev/null 2>&1; then \
		echo "📦 Using dnf to install Go..."; \
		sudo dnf install -y golang; \
	elif command -v pacman >/dev/null 2>&1; then \
		echo "📦 Using pacman to install Go..."; \
		sudo pacman -S --noconfirm go; \
	elif command -v zypper >/dev/null 2>&1; then \
		echo "📦 Using zypper to install Go..."; \
		sudo zypper install -y go; \
	elif command -v brew >/dev/null 2>&1; then \
		echo "📦 Using Homebrew to install Go..."; \
		brew install go; \
	else \
		echo "❌ No supported package manager found."; \
		echo "📋 Manual installation:"; \
		echo "   wget -qO- https://go.dev/dl/go$$GO_VERSION.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -"; \
		echo "   export PATH=\$$PATH:/usr/local/go/bin"; \
		exit 1; \
	fi
	@echo "🔍 Verifying installation..."
	@if command -v go >/dev/null 2>&1 && go version >/dev/null 2>&1; then \
		GO_VERSION=$$(go version 2>/dev/null | cut -d' ' -f3 || echo "unknown"); \
		echo "✅ Go installation completed: $$GO_VERSION"; \
	else \
		echo "❌ Go installation failed or is not working properly"; \
		echo "🔧 Try adding Go to your PATH:"; \
		echo "   export PATH=\$$PATH:/usr/local/go/bin"; \
		echo "   # Add this to your ~/.bashrc or ~/.zshrc for persistence"; \
		exit 1; \
	fi

.PHONY: check-go-version
check-go-version:
	@if [ -z "$(GO_VERSION)" ] || [ -z "$(REQUIRED_VERSION)" ]; then \
		echo "❌ GO_VERSION and REQUIRED_VERSION must be provided"; \
		exit 1; \
	fi; \
	CURRENT_MAJOR=$$(echo "$(GO_VERSION)" | cut -d'.' -f1); \
	CURRENT_MINOR=$$(echo "$(GO_VERSION)" | cut -d'.' -f2); \
	REQUIRED_MAJOR=$$(echo "$(REQUIRED_VERSION)" | cut -d'.' -f1); \
	REQUIRED_MINOR=$$(echo "$(REQUIRED_VERSION)" | cut -d'.' -f2); \
	if [ "$$CURRENT_MAJOR" -gt "$$REQUIRED_MAJOR" ] || \
	   ([ "$$CURRENT_MAJOR" -eq "$$REQUIRED_MAJOR" ] && [ "$$CURRENT_MINOR" -ge "$$REQUIRED_MINOR" ]); then \
		exit 0; \
	else \
		exit 1; \
	fi

.PHONY: deps
deps:
	@echo "📦 Installing Go dependencies..."
	@if ! go mod download 2>/dev/null; then \
		echo "⚠️  Initial mod download failed, attempting fixes..."; \
		if go version | grep -q "go1\.2[0-4]"; then \
			echo "🔄 Detected newer Go version, updating go.mod..."; \
			go mod edit -go=1.21; \
			echo "✅ Updated go.mod to use Go 1.21 for compatibility"; \
		fi; \
		echo "🔄 Retrying dependency download..."; \
		go mod download || { \
			echo "❌ Dependency download failed. Possible solutions:"; \
			echo "   1. Check internet connection"; \
			echo "   2. Verify GOPROXY settings: go env GOPROXY"; \
			echo "   3. Clear module cache: go clean -modcache"; \
			echo "   4. Check for network proxy issues"; \
			exit 1; \
		}; \
	fi
	@echo "🧹 Tidying module dependencies..."
	@go mod tidy || { \
		echo "❌ Module tidy failed"; \
		echo "🔧 Try: go clean -modcache && go mod download"; \
		exit 1; \
	}
	@echo "✅ Dependencies installed and verified"

.PHONY: fix-go
fix-go:
	@echo "🔧 Diagnosing Go installation issues..."
	@echo ""
	@echo "📋 Go Environment:"
	@if command -v go >/dev/null 2>&1; then \
		echo "   Go Version: $$(go version 2>/dev/null || echo 'ERROR: go version failed')"; \
		echo "   Go Root: $$(go env GOROOT 2>/dev/null || echo 'ERROR: GOROOT not set')"; \
		echo "   Go Path: $$(go env GOPATH 2>/dev/null || echo 'ERROR: GOPATH not set')"; \
		echo "   Go Proxy: $$(go env GOPROXY 2>/dev/null || echo 'ERROR: GOPROXY not set')"; \
		echo "   Go Mod Cache: $$(go env GOMODCACHE 2>/dev/null || echo 'ERROR: GOMODCACHE not set')"; \
	else \
		echo "   ❌ Go command not found in PATH"; \
	fi
	@echo ""
	@echo "🔍 Common fixes:"
	@echo "   1. Clear module cache: go clean -modcache"
	@echo "   2. Reset proxy: go env -w GOPROXY=https://proxy.golang.org,direct"
	@echo "   3. Update Go toolchain: go install golang.org/dl/go1.21.6@latest && go1.21.6 download"
	@echo "   4. Check PATH includes Go bin directory"
	@echo ""
	@echo "🛠️  Attempting automatic fixes..."
	@if command -v go >/dev/null 2>&1; then \
		echo "📁 Clearing module cache..."; \
		go clean -modcache 2>/dev/null || echo "   ⚠️  Cache clear failed"; \
		echo "🌐 Resetting GOPROXY..."; \
		go env -w GOPROXY=https://proxy.golang.org,direct 2>/dev/null || echo "   ⚠️  GOPROXY reset failed"; \
		echo "📊 Testing basic go command..."; \
		if go version >/dev/null 2>&1; then \
			echo "   ✅ Go command working"; \
		else \
			echo "   ❌ Go command still failing"; \
		fi; \
	else \
		echo "❌ Cannot run fixes - Go not found in PATH"; \
		echo "🔧 Manual steps:"; \
		echo "   export PATH=\$$PATH:/usr/local/go/bin"; \
		echo "   source ~/.bashrc"; \
	fi

.PHONY: check-gosec
check-gosec:
	@echo "🔍 Checking gosec installation..."
	@if ! command -v gosec >/dev/null 2>&1; then \
		echo "❌ gosec is not installed. Installing gosec..."; \
		$(MAKE) install-gosec; \
	else \
		echo "✅ gosec is already installed: $$(gosec --version 2>/dev/null || echo 'version unknown')"; \
	fi

.PHONY: install-gosec
install-gosec:
	@echo "📦 Installing gosec security scanner..."
	@if ! go install github.com/securego/gosec/v2/cmd/gosec@latest; then \
		echo "❌ Failed to install gosec"; \
		echo "🔧 Troubleshooting:"; \
		echo "   1. Check internet connection"; \
		echo "   2. Verify GOPROXY settings: go env GOPROXY"; \
		echo "   3. Clear module cache: go clean -modcache"; \
		exit 1; \
	fi
	@echo "🔍 Verifying gosec installation..."
	@export PATH="$$(go env GOPATH)/bin:$$PATH"; \
	if command -v gosec >/dev/null 2>&1; then \
		echo "✅ gosec installation completed: $$(gosec --version 2>/dev/null || echo 'installed successfully')"; \
	else \
		echo "❌ gosec installation failed or not in PATH"; \
		echo "🔧 Try adding Go bin to your PATH:"; \
		echo "   export PATH=\"$$(go env GOPATH)/bin:\$$PATH\""; \
		exit 1; \
	fi

# Build for current platform
.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	$(GO_CMD) build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd

# Build for all platforms
.PHONY: build-all
build-all: clean
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		OS=$$(echo $$platform | cut -d'/' -f1); \
		ARCH=$$(echo $$platform | cut -d'/' -f2); \
		BINARY_SUFFIX=""; \
		if [ "$$OS" = "windows" ]; then \
			BINARY_SUFFIX=".exe"; \
		fi; \
		echo "Building $$OS/$$ARCH..."; \
		GOOS=$$OS GOARCH=$$ARCH $(GO_CMD) build $(GO_BUILD_FLAGS) \
			-o $(DIST_DIR)/$(BINARY_NAME)-$$OS-$$ARCH$$BINARY_SUFFIX $(GO_MAIN); \
	done

# Clean build artifacts
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f $(BINARY_NAME) coverage*

# Test
.PHONY: test
test:
	$(GO_CMD) test $(GO_TEST_FLAGS) ./...

# Lint
.PHONY: lint
lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@export PATH=$$PATH:$$(go env GOPATH)/bin && golangci-lint run ./...

# Coverage
.PHONY: coverage
coverage:
	$(GO_CMD) test $(GO_TEST_FLAGS) ./cmd ./internal ./internal/csv
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@echo "To view, open coverage.html in your browser."
	@echo "You can also run 'go tool cover -func=coverage.out' for a summary."

# Format
.PHONY: format
format:
	$(GO_CMD) fmt ./...
	$(GO_CMD) mod tidy

# Run (installs to user's local bin and runs)
.PHONY: run
run: build install-user
	@echo "Running pivot..."
	@pivot

# Run without installing (original behavior)
.PHONY: run-local  
run-local: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install to user's local bin (no sudo required)
.PHONY: install-user
install-user: build
	@echo "Installing pivot to user PATH..."
	@mkdir -p ~/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) ~/bin/$(BINARY_NAME)
	@echo "✅ pivot installed to ~/bin/$(BINARY_NAME)"
	@if ! echo "$$PATH" | grep -q "$$HOME/bin"; then \
		echo "⚠️  Note: ~/bin is not in your PATH. Add this to your shell profile:"; \
		echo "   export PATH=\"\$$HOME/bin:\$$PATH\""; \
	else \
		echo "✅ pivot is now available from any directory"; \
	fi

# Install to system PATH (requires sudo)
.PHONY: install
install: build
	@echo "Installing pivot to system PATH..."
	@sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME)
	@echo "✅ pivot installed to /usr/local/bin/$(BINARY_NAME)"
	@echo "You can now run 'pivot' from any directory"

# Uninstall from user's local bin
.PHONY: uninstall-user
uninstall-user:
	@echo "Removing pivot from user PATH..."
	@rm -f ~/bin/$(BINARY_NAME)
	@echo "✅ pivot removed from ~/bin"

# Uninstall from system PATH
.PHONY: uninstall
uninstall:
	@echo "Removing pivot from system PATH..."
	@sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "✅ pivot removed from system PATH"

# Create release with checksums
.PHONY: release
release: build-all
	@echo "Creating release files..."
	@cd $(DIST_DIR) && for file in *; do \
		if [ -f "$$file" ]; then \
			shasum -a 256 "$$file" >> checksums.txt; \
		fi; \
	done
	@echo "Release files created in $(DIST_DIR)/"

# E2E Testing
.PHONY: test-e2e
test-e2e:
	@echo "Running E2E tests..."
	@./test/e2e/end-user-package-install.sh

# Homebrew E2E Test (macOS only)
.PHONY: test-homebrew-e2e
test-homebrew-e2e:
	@echo "Running Homebrew E2E test..."
	@./scripts/test-homebrew-e2e.sh

# Test with local build via Homebrew
.PHONY: test-homebrew-local
test-homebrew-local:
	@echo "Testing local build via Homebrew..."
	@./scripts/test-homebrew-e2e.sh --local-build

# Cleanup test installations
.PHONY: test-cleanup
test-cleanup:
	@echo "Cleaning up test installations..."
	@./scripts/test-homebrew-e2e.sh --cleanup

# CLI Test Suite
.PHONY: test-cli
test-cli:
	@echo "Running comprehensive CLI test suite..."
	@chmod +x ./scripts/test-cli.sh
	@./scripts/test-cli.sh

# Security Test Suite
.PHONY: test-security
test-security: check-gosec
	@echo "Running security test suite..."
	@chmod +x ./scripts/security-test.sh
	@./scripts/security-test.sh

# Badge System Tests
.PHONY: test-badges
test-badges:
	@echo "Testing badge system locally..."
	@chmod +x ./scripts/test-badges.sh
	@./scripts/test-badges.sh

.PHONY: test-ci-badges
test-ci-badges:
	@echo "Simulating CI badge workflow..."
	@chmod +x ./scripts/test-ci-badges.sh
	@./scripts/test-ci-badges.sh

.PHONY: test-dynamic-badges
test-dynamic-badges:
	@echo "Testing Schneegans dynamic badges integration..."
	@chmod +x ./scripts/test-dynamic-badges.sh
	@./scripts/test-dynamic-badges.sh

# Post-Release Validation Tests
.PHONY: test-post-release
test-post-release:
	@echo "Running post-release validation tests..."
	@chmod +x ./scripts/post-release-validation.sh
	@./scripts/post-release-validation.sh

# All Test Suites
.PHONY: test-all
test-all: test test-cli test-security test-e2e
	@echo "✅ All test suites completed!"

# Submit to Homebrew Core
.PHONY: submit-homebrew-core
submit-homebrew-core:
	@echo "Submitting pivot to Homebrew Core..."
	@./scripts/submit-to-homebrew-core.sh

# Setup Dynamic Badges
.PHONY: setup-badges
setup-badges:
	@echo "Setting up dynamic badges..."
	@chmod +x ./scripts/setup-badges.sh
	@./scripts/setup-badges.sh
