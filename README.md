[![Build Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/9eb5e7a714008e7e2c3a9ce48aeb7cd2/raw/pivot-build.json)](https://github.com/rhino11/pivot/actions)
[![Coverage Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/6b163e8929917383d59754315852f901/raw/pivot-coverage.json)](https://github.com/rhino11/pivot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/rhino11/pivot)](https://goreportcard.com/report/github.com/rhino11/pivot)
[![Security Rating](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/02cf4d8ae9f5bb1b9911eb75beafeaf9/raw/pivot-security.json)](https://github.com/rhino11/pivot/security)
[![Go Version](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/9eb5e7a714008e7e2c3a9ce48aeb7cd2/raw/pivot-go-version.json)](https://golang.org)
[![License](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/9eb5e7a714008e7e2c3a9ce48aeb7cd2/raw/pivot-license.json)](https://opensource.org/licenses/MIT)

# Pivot CLI

**Pivot** is a CLI tool for syncing GitHub issues to a local database, enabling agile, AI-driven project management with offline capabilities.

As I've worked with GenAI coding assistants, I've noticed that they tend to create their own backlog without documentation. This leads to feature drift which, while exciting, tends to not be grounded the reality of a rigorous, curated, and prioritized backlog. I've noticed my own tendency to exercise weak project management while working with GenAI, relying solely on recommendations, "Pinky and the Brain"-style prompts from the AI to continue coding ("Gee Brain, what do you want to do next, Brain"?), and Markdown roadmaps.

I created `pivot` to help developers like me keep their GenAI assistant focused on the human-owned backlog. My hypothesis that led to creating pivot was that routine use of project management tools can help the AI work in small, high-quality batches just like us humans try to do. GenAI can, of course, help inform and maintain this backlog for or alongside us humans. The `pivot` tool just helps us work in an agile fashion with the robots, which is a layer higher than "vibe coding" currently allows.

The first `pivot` release validates issue synchronization, local configuration management, and multi-platform binary and package distribution. To fully integrate with GenAI coding assistants, expect a `pivot` Model Context Protocol (MCP) background service to follow soon.

## Features

- 🔄 **Bidirectional Sync**: Sync GitHub issues to local SQLite database
- 🛠️ **Offline Support**: Work with issues even without internet connectivity
- 🚀 **AI-Ready**: Designed for future AI/GenAI integration
- 📦 **Multi-Platform**: Available for Windows, macOS, and Linux
- 🔧 **Simple Configuration**: Easy setup with YAML config

## Installation

### Quick Install (Recommended)

**Unix/Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/rhino11/pivot/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/rhino11/pivot/main/install.ps1 | iex
```

### Package Managers

#### Homebrew (macOS)
```bash
# Option 1: Official Homebrew (when available)
brew install pivot

# Option 2: Custom tap (available now)
brew tap rhino11/tap
brew install pivot
```

#### Chocolatey (Windows)
```powershell
choco install pivot
```

#### APT (Ubuntu/Debian)
```bash
wget https://github.com/rhino11/pivot/releases/latest/download/pivot_amd64.deb
sudo dpkg -i pivot_amd64.deb
```

#### YUM/DNF (RHEL/Fedora/CentOS)
```bash
wget https://github.com/rhino11/pivot/releases/latest/download/pivot-1.0.0-1.x86_64.rpm
sudo rpm -i pivot-1.0.0-1.x86_64.rpm
```

#### Snap (Linux)
```bash
sudo snap install pivot
```

#### Docker
```bash
docker run --rm -v $(pwd):/workspace ghcr.io/rhino11/pivot:latest --help
```

### Manual Installation

Download the appropriate binary for your platform from the [releases page](https://github.com/rhino11/pivot/releases) and place it in your PATH.

## Usage

### Initial Setup

1. Create a `config.yaml` file in your project directory:
```yaml
owner: your-github-username-or-org
repo: your-repo-name
token: your-github-personal-access-token
```

2. Initialize the local database:
```bash
pivot init
```

3. Sync issues from GitHub:
```bash
pivot sync
```

### Commands

- `pivot init` - Initialize the local issues database
- `pivot sync` - Sync issues between GitHub and local database
- `pivot version` - Show version information
- `pivot help` - Show help information

### Configuration

The `config.yaml` file contains your GitHub repository data, including secrets like the token `pivot` uses to access various API endpoints. The command `pivot init` ensures the config is populated, and subsequent commands won't function unless there is a valid `config.yaml`.

There is a template `config.example.yaml` included in the versioned code. When you run `pivot init`, the actual `config.yaml` is generated. The `config.yaml` file is excluded from `git` tracking in `.gitignore`.

The `config.yaml` file supports the following options:

```yaml
# GitHub repository details
owner: your-username-or-org
repo: your-repository-name

# GitHub Personal Access Token
# Required scopes: repo (for private repos) or public_repo (for public repos)
token: ghp_your_token_here

# Optional: Database file path (default: ./pivot.db)
database: ./pivot.db

# Optional: Sync options
sync:
  include_closed: true    # Include closed issues (default: true)
  batch_size: 100        # Number of issues to fetch per request (default: 100)
```

## Building from Source

### Prerequisites
- Go 1.22 or later
- Make (optional, for using Makefile)

### Build
```bash
git clone https://github.com/rhino11/pivot.git
cd pivot
go build -o pivot ./cmd/main.go
```

### Build for all platforms
```bash
make build-all
```

## Development

### Running Tests
```bash
go test ./...
```

### Linting
```bash
golangci-lint run ./...
```

### Formatting
```bash
go fmt ./...
```

## Database Schema

Issues are stored in a local SQLite database with the following schema:

```sql
CREATE TABLE issues (
    github_id INTEGER PRIMARY KEY,
    number INTEGER,
    title TEXT,
    body TEXT,
    state TEXT,
    labels TEXT,           -- JSON array of label names
    assignees TEXT,        -- JSON array of assignee usernames
    created_at TEXT,
    updated_at TEXT,
    closed_at TEXT
);
```

## Roadmap

- [ ] **Two-way sync** - Push local changes back to GitHub
- [ ] **AI Integration** - GenAI-powered project planning and insights
- [ ] **Advanced Filtering** - Query and filter issues with SQL-like syntax
- [ ] **Team Collaboration** - Multi-user support with conflict resolution
- [ ] **Custom Fields** - Add custom metadata to issues
- [ ] **Reporting** - Generate project reports and analytics
- [ ] **Plugin System** - Extensible architecture for custom integrations

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Scripts

The project includes several automation scripts to streamline development and release processes:

#### Release Management
```bash
# Create a new release (automated CI/CD)
./scripts/release.sh v1.1.0

# Set up Homebrew tap repository
./scripts/setup-homebrew-tap.sh

# Test installation methods
./scripts/test-installation.sh [version]
```

#### Local Development
```bash
# Build and install locally
make install-user

# Run with system-wide access
make run

# Run tests with coverage
make test

# Clean up local installation
make uninstall-user
```

The release process is fully automated:
- **GitHub Actions** handles cross-platform builds, packaging, and releases
- **Homebrew tap** is automatically updated via GitHub API
- **Multiple package formats** (DEB, RPM, Chocolatey) are generated
- **Installation testing** ensures all delivery methods work correctly

## E2E Testing

## Homebrew macOS E2E Testing

The project includes comprehensive end-to-end testing for Homebrew installation on macOS, ensuring the complete user experience works correctly from tap addition through package installation and basic functionality verification.

### Local Development

```bash
# Run E2E test with latest version
make test-homebrew-e2e

# Test with local build (no release required)
make test-homebrew-local

# Dry run (no actual installation)
./scripts/test-homebrew-e2e.sh --dry-run

# Test specific version
./scripts/test-homebrew-e2e.sh v1.1.0

# Clean up test installations
make test-cleanup
```

### What the E2E Test Covers

1. **Prerequisites validation** - macOS and Homebrew availability
2. **Clean state verification** - ensures no conflicting installations
3. **Tap addition** - `brew tap rhino11/tap`
4. **Package installation** - `brew install pivot`
5. **Basic functionality** - version, help, and config commands
6. **Package information** - validates installation metadata
7. **Formula validation** - homebrew formula audit
8. **Version consistency** - ensures correct version installed
9. **Automatic cleanup** - removes test installations

### CI Integration

The E2E test runs automatically in GitHub Actions on every tagged release:

- **Triggers**: After successful release creation
- **Platform**: macOS runner (latest)
- **Timeout**: 10 minutes with proper error handling
- **Artifacts**: Test logs uploaded for debugging

## Testing

The project includes comprehensive testing infrastructure covering unit tests, CLI tests, security scanning, E2E tests, and post-release validation.

### Quick Test Commands
```bash
make test-all           # Run all test suites
make test              # Unit tests
make test-cli          # CLI command testing
make test-security     # Security scanning
make test-homebrew-e2e # Homebrew E2E testing
```

### Documentation
- 📋 **[Comprehensive Testing Guide](docs/COMPREHENSIVE_TESTING.md)** - Complete testing documentation
- 🧪 **[Homebrew E2E Testing](docs/E2E_TESTING.md)** - Homebrew-specific E2E testing
- 🚀 **[Release Automation](docs/RELEASE_AUTOMATION.md)** - Release process and testing

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📚 [Documentation](https://github.com/rhino11/pivot/wiki)
- 🐛 [Issue Tracker](https://github.com/rhino11/pivot/issues)
- 💬 [Discussions](https://github.com/rhino11/pivot/discussions)
# Testing badge fixes
