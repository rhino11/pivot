# Pivot CLI

**Pivot** is a powerful CLI tool for syncing GitHub issues to a local database, enabling agile, AI-driven project management with offline capabilities.

[![Build Status](https://github.com/rhino11/pivot/workflows/Build%20and%20Release/badge.svg)](https://github.com/rhino11/pivot/actions)
[![Coverage Status](https://img.shields.io/badge/coverage-80.6%25-green.svg)](https://github.com/rhino11/pivot/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/rhino11/pivot)](https://goreportcard.com/report/github.com/rhino11/pivot)
[![Security Rating](https://img.shields.io/badge/security-A-brightgreen)](https://github.com/rhino11/pivot/security)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

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

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- 📚 [Documentation](https://github.com/rhino11/pivot/wiki)
- 🐛 [Issue Tracker](https://github.com/rhino11/pivot/issues)
- 💬 [Discussions](https://github.com/rhino11/pivot/discussions)
- Release binaries for major OSes
