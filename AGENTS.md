# Agent Guidelines for gpd

This document provides essential guidance for AI agents working on the gpd (Google Play Developer CLI) codebase.

## Build & Test Commands

```bash
# Build
make build              # Build for current platform
make build-all          # Build for all platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)
make run ARGS="cmd"     # Build and run CLI with arguments

# Test
make test               # Run all tests with race detection and coverage
make test-coverage      # Generate HTML coverage report (coverage.html)
go test ./internal/auth              # Run tests for a specific package
go test -v -run TestAuthStatus ./internal/auth  # Run a single test

# Lint
make lint               # Run golangci-lint (must be installed: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)

# Clean
make clean              # Remove build artifacts and coverage reports
```

## Code Style Guidelines

### Imports & Formatting
- **Import order**: Standard library → third-party → internal packages (blank lines between each group)
- Use `gofmt` and `goimports`; both are enforced by CI
- Local package prefix in goimports config: `github.com/dl-alexandre/gpd`

### Naming Conventions
- **Exported**: PascalCase (types, functions, constants, methods)
- **Unexported**: camelCase (variables, functions, methods)
- **Constants**: PascalCase with descriptive names (e.g., `ScopeAndroidPublisher`)
- **Interfaces**: Simple names ending with capability (e.g., `SecureStorage`, `Option`)
- **Test helpers**: Create in `*_test.go` files, use `t.Helper()` wrapper

### Comments
- Package comment: `// Package <name> provides <description>.` (first line of file)
- Exported types/functions must have comments
- Use sentence case with first word capitalized
- Constants may have comments explaining purpose

### Error Handling
```go
// Use structured errors from internal/errors
import "github.com/dl-alexandre/gpd/internal/errors"

// Create errors
err := errors.NewAPIError(errors.CodeValidationError, "message")
    .WithHint("actionable hint")
    .WithDetails(map[string]interface{}{"key": "value"})

// Use predefined common errors
errors.ErrAuthNotConfigured
errors.ErrPermissionDenied
errors.ErrPackageRequired

// NEVER suppress type errors
// Forbidden: as any, @ts-ignore, @ts-expect-error
```

### Function & Type Guidelines
- Max 150 lines per function, 80 statements (enforced by golangci-lint)
- Max cyclomatic complexity: 40 (enforced by golangci-lint)
- Use `sync.Once` for lazy initialization (see `api/client.go`)
- Use options pattern for configuration (see `api/client.go:WithTimeout`)

### Concurrency
- Use `sync.RWMutex` for protecting shared state
- Always `defer mu.Unlock()` after `mu.Lock()`
- Use channels for synchronization
- No package-level variables with complex initialization

### Context Management
- Pass `context.Context` as first parameter
- Respect context cancellation in all I/O operations
- Use `context.Background()` for lazy initialization (see API services)

### Testing Patterns
- Table-driven tests for multiple scenarios
- Test files co-located: `package.go` → `package_test.go`
- Use `t.Helper()` for test helper functions
- Exemptions: Test files are exempt from `funlen`, `goconst`, `gosec` linters

### Linting & Quality
- golangci-lint configured with 30+ linters (`.golangci.yml`)
- Key rules enforced:
  - No empty catch blocks (errcheck)
  - No duplicates except CLI package (intentional command structure)
  - No type assertions without checking
  - No exported functions without comments
- Test files exempt from: dupl, funlen, goconst, gosec

## Architecture Patterns

### API Client Pattern
```go
// Lazy initialization with sync.Once
type Client struct {
  publisherOnce sync.Once
  publisherSvc  *androidpublisher.Service
  publisherErr  error
}

// Method returns service with lazy initialization
func (c *Client) AndroidPublisher() (*androidpublisher.Service, error) {
  c.publisherOnce.Do(func() {
    c.publisherSvc, c.publisherErr = androidpublisher.NewService(...)
  })
  return c.publisherSvc, c.publisherErr
}

// Retry with exponential backoff
func (c *Client) DoWithRetry(ctx context.Context, fn func() error) error
```

### CLI Command Pattern
1. Define `cobra.Command` with `Use`, `Short`, `Long`, `RunE`
2. Add flags specific to the command
3. `RunE` calls a method on the CLI struct (e.g., `c.authStatus(ctx)`)
4. Method implementation:
   - Parse/validate inputs
   - Call API via `c.apiClient`
   - Handle errors with structured error types
   - Return `output.Result` via `c.output.Write(result)`

### Error Response Structure
All CLI commands return `output.Result` with:
- `data`: Response payload
- `error`: Structured error with code, message, hint, details
- `meta`: No-op flag, durationMs, services list, warnings

### Exit Codes
```go
const (
  ExitSuccess = 0
  ExitGeneralError = 1
  ExitAuthFailure = 2
  ExitPermissionDenied = 3
  ExitValidationError = 4
  ExitRateLimited = 5
  ExitNetworkError = 6
  ExitNotFound = 7
  ExitConflict = 8
)
```

## Project Structure

- `cmd/gpd/`: Entry point (minimal, calls `cli.New().Execute()`)
- `internal/api/`: Unified API client wrapper for Google Play APIs
- `internal/auth/`: Authentication manager (service accounts, OAuth, credential sources)
- `internal/cli/`: Cobra-based command definitions (32 command files)
- `internal/edits/`: Edit transaction lifecycle with file locking and idempotency
- `internal/errors/`: Centralized error types with exit codes
- `internal/output/`: Standardized JSON envelope structure
- `internal/storage/`: Platform-specific secure credential storage
- `internal/config/`: Configuration file management
- `internal/logging/`: PII-redacting logger

## Important Notes

- Go version: 1.24.0
- Never commit unless explicitly requested
- AI-agent friendly: JSON-first output, predictable exit codes, explicit flags
- Credentials stored in platform keychains (Keychain/Secret Service/Credential Manager)
- PII automatically redacted from logs
- Service account keys never stored in config files