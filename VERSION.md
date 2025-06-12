# Version History

This file tracks the version history and release notes for Pivot CLI.

## v1.1.0

- **🎉 NEW FEATURE**: Multi-Project Support
- Enhanced pivot CLI to support managing multiple GitHub projects from a single installation
- New multi-project database structure with central storage in `~/.pivot/` directory
- Enhanced CLI commands with multi-project capabilities:
  - `pivot init --multi-project` - Set up multi-project configuration
  - `pivot config setup` - Interactive multi-project setup
  - `pivot config add-project` - Add new projects to existing configuration
  - `pivot config import` - Import configuration from external files
  - `pivot sync --project owner/repo` - Sync specific projects
- Git repository auto-detection for seamless project setup
- Backward compatibility with existing single-project configurations
- Database migration system for upgrading existing installations
- Enhanced test coverage: 52.4% overall (61.6% cmd, 40.3% internal, 90.1% CSV)

### Multi-Project Features
- **Central Database Storage**: All project data stored in `~/.pivot/` with automatic directory creation
- **Project-Specific Tokens**: Support for per-project GitHub tokens with global fallback
- **Git Integration**: Automatic project detection from Git remotes for streamlined setup
- **Configuration Import/Export**: Import existing configurations and merge with current setup
- **Database Migration**: Seamless upgrade from single-project to multi-project structure
- **Project Filtering**: Sync specific projects using `--project` flag

### Technical Improvements
- New multi-project database schema with `projects` table and foreign key relationships
- Enhanced error handling and validation for multi-project scenarios
- Comprehensive test suite with Test-Driven Development approach
- Security compliance with proper `#nosec` annotations for code scanning
- Full CI/CD pipeline validation with all quality gates passing

## v1.0.5

- **🎉 NEW FEATURE**: CSV Import/Export for GitHub Issues
- Added comprehensive CSV import/export functionality for seamless GitHub issues interoperability
- New commands: `pivot import csv` and `pivot export csv`
- Enhanced CLI with preview (`--preview`) and dry-run (`--dry-run`) modes for safe testing
- Real GitHub API integration for creating issues directly from CSV files
- Comprehensive test coverage (90.1% for CSV package)
- Production-ready with full error handling and validation

### CSV Import/Export Features
- **CSV Import**: `pivot import csv <file>` - Import GitHub issues from CSV files
  - Preview mode: `--preview` to see what would be imported without creating issues
  - Dry-run mode: `--dry-run` to test import logic without API calls
  - Repository targeting: `--repository owner/repo` to specify target repository
  - Duplicate detection: `--skip-duplicates` to avoid creating duplicate issues
  - Real GitHub API integration with proper authentication and error handling

- **CSV Export**: `pivot export csv` - Export local issues to CSV format
  - Custom output: `--output filename.csv` to specify export file
  - Field filtering: `--fields title,state,labels` to export specific columns
  - Repository filtering: `--repository owner/repo` to export specific repository issues

### Supported CSV Fields
- **Core Fields**: `id`, `title`, `state`, `priority`, `labels`, `assignee`, `milestone`
- **Content Fields**: `body`, `acceptance_criteria`, `epic`
- **Metadata Fields**: `created_at`, `updated_at`, `estimated_hours`, `story_points`
- **Relationships**: `dependencies` (comma-separated issue IDs)

### Usage Examples
```bash
# Import issues with preview
pivot import csv --preview backlog.csv

# Import to specific repository with dry-run
pivot import csv --dry-run --repository myorg/myrepo issues.csv

# Export issues to custom file
pivot export csv --output exported-issues.csv

# Export specific fields only
pivot export csv --fields title,state,labels --output minimal.csv
```

### Technical Implementation
- Robust CSV validation with required field checking
- Hierarchical CLI command structure with proper flag handling
- GitHub API integration using existing authentication system
- Comprehensive unit tests (90.1% coverage), CLI tests, and E2E workflow tests
- Security-hardened with proper input validation and error handling
- CI/CD integration with automated testing pipeline

## v1.0.4

- Minor CI fix discovered in the final pipeline stage post release.

## v1.0.3

- Fixed release.sh relative path to setup-homebrew-tap.sh

## v1.0.2

- Fixed issues with Homebrew formula publication
- Improved release automation for Homebrew compatibility
- Updated documentation for Homebrew installation steps
- Minor bug fixes related to package distribution

## v1.0.0 (Current)

- Initial release of Pivot CLI
- Core functionality for GitHub issues management
- Local database with offline sync capabilities
- Interactive configuration setup with `pivot init`
- Support for multiple platforms (Windows, macOS, Linux)
- Comprehensive test coverage (80%+)
- CI/CD pipeline with automated releases

### Features
- **Issue Management**: Create, update, sync, and manage GitHub issues locally
- **Offline Support**: Work with issues without internet connectivity
- **Interactive Setup**: Guided configuration with `pivot init` command
- **Cross-Platform**: Native binaries for Windows, macOS, and Linux
- **Package Manager Support**: Homebrew, APT, YUM/DNF, and Chocolatey packages

### Installation Methods
- Homebrew: `brew install rhino11/tap/pivot`
- Direct download from GitHub releases
- Package managers (DEB, RPM, Chocolatey)

### Technical Details
- Go 1.22+ compatibility
- SQLite database backend
- GitHub API integration
- Comprehensive error handling
- Extensive unit and integration tests

---

## Version Guidelines

This file should only be modified when incrementing the release tag version. Follow semantic versioning (MAJOR.MINOR.PATCH):

- **MAJOR**: Breaking changes
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible

When creating a new release:
1. Update this file with the new version and changes
2. Commit the changes
3. Create and push the version tag
4. The CI/CD pipeline will handle the rest
