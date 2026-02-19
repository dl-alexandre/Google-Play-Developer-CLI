# Contributing to gpd

Thank you for your interest in contributing to gpd! This document provides guidelines and information about contributing.

## Code of Conduct

Please be respectful and considerate in all interactions. We're all here to build something useful together.

## How to Contribute

### Reporting Bugs

Before submitting a bug report:
1. Check existing issues to avoid duplicates
2. Use the latest version of gpd
3. Collect relevant information (version, OS, error messages)

When reporting bugs, include:
- gpd version (`gpd version --pretty`)
- Operating system and version
- Steps to reproduce the issue
- Expected vs actual behavior
- Any error messages (with sensitive data redacted)

### Suggesting Features

Feature requests are welcome! Please:
1. Check if the feature has already been requested
2. Describe the use case and why it would be valuable
3. Consider how it fits with the project's scope (Google Play Developer API automation)

### Pull Requests

1. **Fork and clone** the repository
2. **Create a branch** for your changes: `git checkout -b feature/my-feature`
3. **Make your changes** following the code style guidelines
4. **Add tests** for new functionality
5. **Run tests** locally: `make test`
6. **Commit** with clear, descriptive messages
7. **Push** and create a pull request

#### Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and reasonably sized

#### Commit Messages

Use clear, descriptive commit messages:
- `feat: add support for custom tracks`
- `fix: handle rate limiting correctly`
- `docs: update installation instructions`
- `test: add tests for review filtering`

#### Testing

- Write unit tests for new functionality
- Ensure existing tests pass
- Consider edge cases and error conditions
- Use table-driven tests where appropriate

## Development Setup

### Prerequisites

- Go 1.24 or later
- Make (optional, for convenience)

### Building

```bash
# Clone the repository
git clone https://github.com/dl-alexandre/gpd.git
cd gpd

# Download dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linter
make lint
```

### Project Structure

```
gpd/
├── cmd/gpd/          # Entry point
├── internal/         # Internal packages
│   ├── api/          # Google API clients
│   ├── auth/         # Authentication
│   ├── cli/          # CLI commands
│   ├── config/       # Configuration
│   ├── edits/        # Edit transaction management
│   ├── errors/       # Error types and codes
│   ├── logging/      # Logging with PII redaction
│   ├── output/       # Output formatting
│   └── storage/      # Secure credential storage
├── pkg/version/      # Version information
└── Makefile          # Build automation
```

### Running Locally

```bash
# Build and run
make run ARGS="version"

# Or directly
go run ./cmd/gpd version
```

### Testing with a Service Account

For integration testing, you'll need a Google Cloud service account:

1. Create a service account in Google Cloud Console
2. Enable the Google Play Android Publisher API
3. Add the service account to your Play Console with appropriate permissions
4. Download the JSON key file

```bash
export GPD_SERVICE_ACCOUNT_KEY="$(cat /path/to/key.json)"
gpd auth status
```

## Release Process

Releases are automated via GitHub Actions when tags are pushed:

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser handles:
- Cross-platform builds
- GitHub releases
- Homebrew formula updates
- Docker image publishing

## Questions?

- Open an issue for bugs or feature requests
- Start a discussion for questions or ideas

Thank you for contributing!
