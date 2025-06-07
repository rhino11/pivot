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
	@echo "  format        - Format Go code"
	@echo "  run           - Run the CLI"
	@echo "  install       - Install binary to system PATH"
	@echo "  release       - Create release binaries"

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
	rm -f $(BINARY_NAME) coverage.out

# Test
.PHONY: test
test:
	$(GO_CMD) test $(GO_TEST_FLAGS) ./...

# Lint
.PHONY: lint
lint:
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@export PATH=$$PATH:$$(go env GOPATH)/bin && golangci-lint run ./...

# Format
.PHONY: format
format:
	$(GO_CMD) fmt ./...
	$(GO_CMD) mod tidy

# Run
.PHONY: run
run: build
	./$(BUILD_DIR)/$(BINARY_NAME)

# Install to system PATH
.PHONY: install
install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/

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
