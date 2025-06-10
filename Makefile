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
all: clean deps build

# Help target
.PHONY: help
help:
	@echo "Pivot Project Makefile"
	@echo "======================"
	@echo ""
	@echo "Available targets:"
	@echo "  all           - Clean, install dependencies, and build"
	@echo "  build         - Build binary for current platform"
	@echo "  build-all     - Build binaries for all platforms"
	@echo "  clean         - Remove build artifacts"
	@echo "  deps          - Install Go dependencies"
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
	@echo "  test-post-release - Run post-release validation tests"
	@echo "  test-all      - Run all test suites (unit + CLI + security + E2E)"
	@echo "  submit-homebrew-core - Submit formula to Homebrew Core"
	@echo "  setup-badges  - Setup dynamic badges for repository"
	@echo "  ci            - Run complete CI pipeline locally (deps + format + lint + test + security + build)"
	@echo "  ci-full       - Run full CI pipeline with coverage, E2E tests, and multi-platform builds"
	@echo "  ci-quick      - Run quick CI validation for development (format + lint + test + cli)"

# CI Pipeline - Run complete CI pipeline locally
.PHONY: ci
ci: clean deps format lint test test-cli test-security build
	@echo ""
	@echo "🎉 CI Pipeline completed successfully!"
	@echo "✅ Dependencies installed"
	@echo "✅ Code formatted"
	@echo "✅ Linting passed"
	@echo "✅ Unit tests passed"
	@echo "✅ CLI tests passed"
	@echo "✅ Security tests passed"
	@echo "✅ Build successful"
	@echo ""
	@echo "Ready for commit and push! 🚀"

# Full CI Pipeline - Includes coverage, E2E tests, and multi-platform builds
.PHONY: ci-full
ci-full: clean deps format lint test coverage test-cli test-security build-all test-e2e
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
.PHONY: deps
deps:
	$(GO_CMD) mod download
	$(GO_CMD) mod tidy

# Build for current platform
.PHONY: build
build:
	mkdir -p $(BUILD_DIR)
	$(GO_CMD) build $(GO_BUILD_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(GO_MAIN)

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
	$(GO_CMD) test $(GO_TEST_FLAGS) ./...
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
test-security:
	@echo "Running security test suite..."
	@chmod +x ./scripts/security-test.sh
	@./scripts/security-test.sh

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
