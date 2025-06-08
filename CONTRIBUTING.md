# Contributing to Pivot CLI

Thank you for your interest in contributing to Pivot CLI! This document provides guidelines and information for contributors.

## GenAI Starting Prompt

To date, I have developed this project largely with Claude 4 Sonnet running in Agent mode in VS Code. If you are also using GenAI, I recommend the following prompt to get your PR started on the right track.

>Hey Claude! We're going to work a PR consisting of a {feature|fix|...|chore} from the pivot backlog. I'm going to set some ground rules for us. Before we start coding, review CONTRIBUTING.md to ensure we have a strong Test Driven Development approach. Also, install the latest binary for the pivot package itself to sync local GitHub issues for this project with upstream issues. You should make small changes and test each change with unit, integration, CLI UI, and E2E tests. You should strive to only work within the context of the issue that corresponds to the branch we've checked out. Before we conclude work on a prompt, you should always ensure all quality and security gates pass. Aim to keep source code files no larger than 500 lines, refactoring files that exceed this limit into file structures that maintain testability. Run all tests at the outset to ensure our contributions maintain or improve code functionality, quality, and security. Only modify VERSION.md upon incrementing the release tag version. Let's go!

## Development Setup

### Prerequisites

- Go 1.22 or later
- Git
- Make (optional, but recommended)

### Getting Started

1. **Fork and clone the repository:**
   ```bash
   git clone https://github.com/your-username/pivot.git
   cd pivot
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Build the project:**
   ```bash
   make build
   # or manually:
   go build -o build/pivot ./cmd/main.go
   ```

4. **Run tests:**
   ```bash
   make test
   # or manually:
   go test ./...
   ```

## Development Workflow

### Making Changes

1. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes and test thoroughly:**
   ```bash
   # Build and test
   make build test
   
   # Format code
   make format
   
   # Run linter (if available)
   make lint
   ```

3. **Commit your changes:**
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

### Commit Message Convention

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification, with an additional modifier to delineate human contributions from AI contributions:

- `{hu|ai}:feat:` - New features
- `{hu|ai}:fix:` - Bug fixes
- `{hu|ai}:docs:` - Documentation changes
- `{hu|ai}:style:` - Code style changes (formatting, etc.)
- `{hu|ai}:refactor:` - Code refactoring
- `{hu|ai}:test:` - Test additions or modifications
- `{hu|ai}:chore:` - Maintenance tasks

### Testing

- Write tests for new functionality
- Ensure all existing tests pass
- Test on multiple platforms when possible
- Test CLI commands manually

### Documentation

- Update README.md if adding new features
- Add inline code comments for complex logic
- Update help text and command descriptions
- Consider adding examples for new functionality

## Code Style

### Go Code Style

- Follow standard Go conventions
- Use `go fmt` for formatting
- Run `go vet` to catch common issues
- Use meaningful variable and function names
- Add comments for exported functions and types

### Project Structure

```
.
├── cmd/           # Application entry point
├── internal/      # Internal packages (not importable)
├── docs/          # Documentation
├── scripts/       # Build and deployment scripts
└── tests/         # Integration tests
```

## Pull Request Process

1. **Ensure your PR:**
   - Has a clear title and description
   - References any related issues
   - Includes tests for new functionality
   - Updates documentation as needed
   - Passes all CI checks

2. **PR Template:**
   ```markdown
   ## Description
   Brief description of changes
   
   ## Type of Change
   - [ ] Bug fix
   - [ ] New feature
   - [ ] Breaking change
   - [ ] Documentation update
   
   ## Testing
   - [ ] Tests pass locally
   - [ ] New tests added (if applicable)
   - [ ] Manual testing completed
   - [ ] Overall coverage remains over 80%
   
   ## Checklist
   - [ ] Code follows project style guidelines
   - [ ] Self-review completed
   - [ ] Documentation updated
   - [ ] Secrets not accidentally added to 'config.example.yml'
   ```

3. **Review Process:**
   - PRs require at least one approval
   - Address review feedback promptly
   - Keep PRs focused and reasonably sized

## Issue Reporting

### Bug Reports

When reporting bugs, please include:

- **Environment:** OS, Go version, Pivot version
- **Steps to reproduce:** Clear, step-by-step instructions
- **Expected behavior:** What should happen
- **Actual behavior:** What actually happens
- **Logs/Output:** Any error messages or relevant output
- **Additional context:** Screenshots, config files, etc.

### Feature Requests

For feature requests, please include:

- **Problem description:** What problem does this solve?
- **Proposed solution:** How should it work?
- **Alternatives considered:** Other approaches you've thought of
- **Additional context:** Use cases, examples, etc.

## Development Guidelines

### Database Changes

- Always provide migration scripts for schema changes
- Test migrations with existing data
- Document any breaking changes to the database schema

### API Changes

- Maintain backward compatibility when possible
- Document breaking changes clearly
- Consider deprecation warnings before removing features

### Configuration Changes

- Provide sensible defaults
- Document new configuration options
- Test with various configuration scenarios

## Release Process

Releases are automated through GitHub Actions when tags are pushed:

1. **Create a release:**
   ```bash
   git tag v1.2.3
   git push origin v1.2.3
   ```

2. **The CI will:**
   - Build binaries for all platforms
   - Create packages for various package managers
   - Generate checksums
   - Create a GitHub release with assets

## Getting Help

- **Documentation:** Check the README and Wiki
- **Discussions:** Use GitHub Discussions for questions
- **Issues:** Create an issue for bugs or feature requests
- **Chat:** Join our community chat (if available)

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

## License

By contributing to Pivot CLI, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to Pivot CLI! 🚀
