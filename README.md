[![Build Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/0a39d1979cd714d14836e9d6427d2eb9/raw/pivot-build.json)](https://github.com/rhino11/pivot/actions)
[![Coverage Status](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/8466693b8eb4ca358099fabc6ed234e0/raw/pivot-coverage.json)](https://github.com/rhino11/pivot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/rhino11/pivot)](https://goreportcard.com/report/github.com/rhino11/pivot)
[![Security Rating](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/a93cb6b503277dd460826517a831497e/raw/pivot-security.json)](https://github.com/rhino11/pivot/security)
[![Go Version](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/0a39d1979cd714d14836e9d6427d2eb9/raw/pivot-go-version.json)](https://golang.org)
[![License](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/rhino11/0a39d1979cd714d14836e9d6427d2eb9/raw/pivot-license.json)](https://opensource.org/licenses/MIT)

# Pivot CLI

**Pivot** is a CLI tool for syncing GitHub issues to a local database, enabling agile, AI-driven project management with offline capabilities.

As I've worked with GenAI coding assistants, I've noticed that they tend to create their own backlog without documentation. This leads to feature drift which, while exciting, tends to not be grounded the reality of a rigorous, curated, and prioritized backlog. I've noticed my own tendency to exercise weak project management while working with GenAI, relying solely on recommendations, "Pinky and the Brain"-style prompts from the AI to continue coding ("Gee Brain, what do you want to do next, Brain"?), and Markdown roadmaps.

I created `pivot` to help developers like me keep their GenAI assistant focused on the human-owned backlog. My hypothesis that led to creating pivot was that routine use of project management tools can help the AI work in small, high-quality batches just like us humans try to do. GenAI can, of course, help inform and maintain this backlog for or alongside us humans. The `pivot` tool just helps us work in an agile fashion with the robots, which is a layer higher than "vibe coding" currently allows.

The first `pivot` release validates issue synchronization, local configuration management, and multi-platform binary and package distribution. To fully integrate with GenAI coding assistants, expect a `pivot` Model Context Protocol (MCP) background service to follow soon.

## Features

- 🔄 **Bidirectional Sync**: Sync GitHub issues to local SQLite database
- 🗂️ **Multi-Project Support**: Manage multiple GitHub repositories from a single installation
- 🛠️ **Offline Support**: Work with issues even without internet connectivity
- 📊 **CSV Import/Export**: Import/export issues to/from CSV files with GitHub API integration
- 🚀 **AI-Ready**: Designed for future AI/GenAI integration
- 📦 **Multi-Platform**: Available for Windows, macOS, and Linux
- 🔧 **Flexible Configuration**: Support both single-project and multi-project setups
- 🎯 **Git Integration**: Automatic project detection from Git repositories

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                        Pivot CLI Architecture                       │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐             │
│  │   CLI       │    │   GitHub    │    │    CSV      │             │
│  │ Commands    │◄──►│     API     │◄──►│ Import/     │             │
│  │             │    │Integration  │    │  Export     │             │
│  └─────────────┘    └─────────────┘    └─────────────┘             │
│         │                   │                   │                   │
│         ▼                   ▼                   ▼                   │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                Configuration Layer                          │   │
│  │  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │   │
│  │  │Single-Proj  │     │Multi-Project│     │Config Import│   │   │
│  │  │   Config    │◄───►│   Config    │◄───►│  /Export    │   │   │
│  │  │             │     │             │     │             │   │   │
│  │  └─────────────┘     └─────────────┘     └─────────────┘   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│         │                                                           │
│         ▼                                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                  Database Layer                             │   │
│  │                                                             │   │
│  │  Central DB: ~/.pivot/                                      │   │
│  │  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │   │
│  │  │  projects   │────►│   issues    │◄───►│  Migration  │   │   │
│  │  │   table     │     │   table     │     │   System    │   │   │
│  │  │             │     │(project_id) │     │             │   │   │
│  │  └─────────────┘     └─────────────┘     └─────────────┘   │   │
│  └─────────────────────────────────────────────────────────────┘   │
│         │                                                           │
│         ▼                                                           │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                Git Integration Layer                        │   │
│  │  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐   │   │
│  │  │Git Remote   │     │Project Auto │     │  Path       │   │   │
│  │  │ Detection   │────►│ Detection   │────►│ Resolution  │   │   │
│  │  │             │     │             │     │             │   │   │
│  │  └─────────────┘     └─────────────┘     └─────────────┘   │   │
│  └─────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                     Data Flow Diagram                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  GitHub APIs ◄─────────┐                                           │
│      ▲                 │                                           │
│      │                 ▼                                           │
│  ┌───────┐    ┌─────────────┐    ┌─────────────┐                  │
│  │ Sync  │◄──►│    Pivot    │◄──►│    CSV      │                  │
│  │Process│    │  Database   │    │   Files     │                  │
│  └───────┘    │(SQLite)     │    └─────────────┘                  │
│      ▲        └─────────────┘                                      │
│      │               ▲                                             │
│      │               │                                             │
│  ┌───────┐    ┌─────────────┐                                     │
│  │Multi- │    │   Config    │                                     │
│  │Project│◄──►│ Management  │                                     │
│  │Filter │    │             │                                     │
│  └───────┘    └─────────────┘                                     │
└─────────────────────────────────────────────────────────────────────┘
```

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

### Quick Start Examples

#### Single Project Setup (Simple)
```bash
# In your Git repository directory
pivot init                    # Auto-detects project from Git remote
pivot sync                    # Sync all issues
```

#### Multi-Project Setup (Recommended)
```bash
# Set up multi-project configuration
pivot init --multi-project

# Add multiple projects interactively
pivot config add-project

# Sync specific project
pivot sync --project myorg/myrepo

# Sync all configured projects
pivot sync
```

#### CSV Workflow
```bash
# Export current issues to CSV
pivot export csv --output issues.csv

# Preview CSV import without creating issues
pivot import csv --preview new-issues.csv

# Import issues from CSV file
pivot import csv new-issues.csv
```

### Detailed Setup

#### Option 1: Interactive Setup (Recommended)
```bash
pivot config setup           # Guided multi-project setup
pivot init                   # Initialize database
pivot sync                   # Start syncing
```

#### Option 2: Manual Configuration
Create a `config.yaml` file in your project directory:
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

#### Core Commands
- `pivot init` - Initialize the local issues database (single-project mode)
- `pivot init --multi-project` - Initialize with multi-project support
- `pivot init --import <file>` - Initialize by importing configuration from file
- `pivot sync` - Sync issues between GitHub and local database
- `pivot sync --project owner/repo` - Sync specific project only
- `pivot version` - Show version information
- `pivot help` - Show help information

#### Configuration Management
- `pivot config setup` - Interactive configuration setup
- `pivot config show` - Display current configuration
- `pivot config add-project` - Add new project to multi-project setup
- `pivot config import <file>` - Import configuration from external file

#### Data Import/Export
- `pivot import csv <file>` - Import GitHub issues from CSV file
- `pivot import csv --preview <file>` - Preview CSV import without creating issues
- `pivot import csv --dry-run <file>` - Test import logic without API calls
- `pivot export csv` - Export local issues to CSV file
- `pivot export csv --output <file>` - Export to specific file

### Configuration

Pivot supports both single-project and multi-project configurations. The configuration file contains your GitHub repository data, including access tokens for API endpoints.

#### Single-Project Configuration (Legacy)

For single repository management, use the traditional `config.yaml` format:

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

#### Multi-Project Configuration (Recommended)

For managing multiple repositories, use the enhanced multi-project format:

```yaml
global:
  # Central database location (supports ~ expansion)
  database: "~/.pivot/issues.db"
  # Global GitHub token (can be overridden per project)
  token: "ghp_your_global_token_here"

projects:
  - owner: "your-org"
    repo: "first-repo"
    path: "/path/to/local/repo"  # Optional: local filesystem path
    
  - owner: "your-org" 
    repo: "second-repo"
    token: "ghp_project_specific_token"  # Optional: project-specific token
    
  - owner: "other-org"
    repo: "third-repo"
```

#### Setup Methods

1. **Interactive Setup**: Run `pivot config setup` for guided configuration
2. **Import from File**: Use `pivot init --import config.yaml` to import existing config
3. **Auto-Detection**: Run `pivot init` in a Git repository for automatic project detection
4. **Multi-Project Migration**: Existing single-project setups are automatically migrated

**Note**: The `config.yaml` file is excluded from Git tracking for security.

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

### Test Coverage

Pivot maintains comprehensive test coverage across all components:
- **Overall Coverage**: 52.4% (as of v1.1.0)
- **CMD Package**: 61.6% - CLI commands and user interface
- **Internal Package**: 40.3% - Core business logic and database operations
- **CSV Package**: 90.1% - Data import/export functionality

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
make coverage

# Run specific test suites
make test          # Unit tests only
make test-cli      # CLI integration tests
make test-security # Security and vulnerability tests

# Full CI pipeline
make ci            # Complete CI/CD validation
```

### Development Workflow

The project follows Test-Driven Development (TDD) principles:

1. **Write Tests First**: New features start with comprehensive test cases
2. **Implement Functionality**: Code to pass the tests
3. **Refactor**: Improve code quality while maintaining test coverage
4. **Security Validation**: All code passes security scans with `gosec`

### Project Structure

```
pivot/
├── cmd/                    # CLI entry points and command implementations
├── internal/              # Core business logic (not exported)
│   ├── csv/              # CSV import/export functionality
│   ├── *_test.go         # Comprehensive test suites
│   ├── config.go         # Configuration management
│   ├── db.go             # Database operations
│   ├── multiproject*.go  # Multi-project support
│   └── sync.go           # GitHub synchronization
├── scripts/              # Development and release automation
├── docs/                 # Technical documentation
└── test/e2e/            # End-to-end testing suites
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

Pivot uses SQLite for local data storage with different schemas depending on configuration mode:

### Multi-Project Schema (v1.1.0+)

The enhanced multi-project schema supports managing multiple repositories:

```sql
-- Projects table stores repository configurations
CREATE TABLE projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner TEXT NOT NULL,
    repo TEXT NOT NULL,
    path TEXT,                    -- Local filesystem path (optional)
    token TEXT,                   -- Project-specific token (optional)
    database_path TEXT,           -- Database location (optional)
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(owner, repo)
);

-- Issues table with project associations
CREATE TABLE issues (
    github_id INTEGER PRIMARY KEY,
    project_id INTEGER NOT NULL,  -- Foreign key to projects table
    number INTEGER,
    title TEXT,
    body TEXT,
    state TEXT,
    labels TEXT,                  -- JSON array of label names
    assignees TEXT,               -- JSON array of assignee usernames
    created_at TEXT,
    updated_at TEXT,
    closed_at TEXT,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

### Legacy Single-Project Schema

For backward compatibility, single-project configurations use the original schema:

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

**Migration**: Existing single-project databases are automatically migrated to multi-project format when upgrading.

## Roadmap

### ✅ Completed (v1.1.0)
- [x] **Multi-Project Support** - Manage multiple GitHub repositories from single installation
- [x] **CSV Import/Export** - Full CSV integration with GitHub API
- [x] **Database Migration** - Seamless upgrade from single to multi-project
- [x] **Git Integration** - Automatic project detection from Git remotes
- [x] **Enhanced CLI** - Comprehensive command structure with help and validation

### 🚧 In Progress
- [ ] **CI Enhancement Features** - Advanced pipeline management and dogfooding improvements
- [ ] **Model Context Protocol (MCP) Server** - AI assistant integration for project management

### 🔮 Future Features
- [ ] **Two-way sync** - Push local changes back to GitHub
- [ ] **AI Integration** - GenAI-powered project planning and insights  
- [ ] **Advanced Filtering** - Query and filter issues with SQL-like syntax
- [ ] **Team Collaboration** - Multi-user support with conflict resolution
- [ ] **Custom Fields** - Add custom metadata to issues
- [ ] **Reporting** - Generate project reports and analytics
- [ ] **Plugin System** - Extensible architecture for custom integrations
- [ ] **Web Interface** - Browser-based issue management dashboard

## Troubleshooting

### Common Issues

#### Configuration Not Found
```bash
# Error: config.yaml not found
pivot config setup           # Set up new configuration
# OR
pivot init --import config.yaml  # Import existing configuration
```

#### Database Migration Issues
```bash
# Error: database schema mismatch
# Backup your database first, then:
pivot init                   # Re-initialize with migration
```

#### Multi-Project vs Single-Project Mode
```bash
# Check current configuration type
pivot config show

# Migrate from single to multi-project
pivot init --multi-project   # Automatically migrates existing data
```

#### GitHub API Rate Limits
```bash
# Use project-specific tokens to increase rate limits
pivot config add-project     # Add token per project
```

### Getting Help

- **Documentation**: Check the `docs/` directory for detailed guides
- **Issues**: Report bugs on [GitHub Issues](https://github.com/rhino11/pivot/issues)
- **Discussions**: Join conversations on [GitHub Discussions](https://github.com/rhino11/pivot/discussions)

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
