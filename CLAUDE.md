# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**gpd** is a fast, lightweight CLI for the Google Play Developer Console - the Android equivalent to the App Store Connect CLI. It's designed to be AI-agent friendly with JSON-first output, predictable exit codes, and explicit flags.

Key characteristics:
- Sub-200ms cold start, minimal memory usage
- Comprehensive API coverage: publishing, reviews, analytics, monetization, vitals, purchases, permissions
- Cross-platform: macOS, Linux, Windows
- Secure credential storage with PII redaction
- Go 1.24.0

## Build & Test Commands

```bash
# Build
make build                    # Build for current platform
make build-all               # Build for all platforms (linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64)

# Test
make test                    # Run all tests with race detection and coverage
make test-coverage           # Generate HTML coverage report
go test ./internal/auth      # Run tests for specific package
go test -v -run TestAuthStatus ./internal/auth  # Run single test

# Lint
make lint                    # Run golangci-lint (must be installed)

# Development
make run ARGS="version"      # Build and run with arguments
go run ./cmd/gpd auth status # Run directly without building

# Clean
make clean                   # Remove build artifacts and coverage reports
```

## Code Architecture

### Package Structure

The codebase follows a clean architecture with separation of concerns:

- **cmd/gpd/**: Entry point - minimal, just calls `cli.New().Execute()`
- **internal/api/**: Unified API client wrapper for Google Play APIs (Android Publisher, Play Developer Reporting, Games Management)
- **internal/auth/**: Authentication manager handling service accounts, OAuth, and credential sources
- **internal/cli/**: Cobra-based command definitions (32 command files)
- **internal/edits/**: Edit transaction lifecycle management with file locking and idempotency
- **internal/errors/**: Centralized error types with exit codes (0-8) and structured error responses
- **internal/output/**: Standardized JSON envelope structure with metadata
- **internal/storage/**: Platform-specific secure credential storage (Keychain/Secret Service/Credential Manager)
- **internal/config/**: Configuration file management
- **internal/logging/**: PII-redacting logger

### Authentication Flow

Authentication follows a priority chain (auth.Manager.Authenticate):

1. Explicit `--key` flag (keyfile)
2. `GPD_SERVICE_ACCOUNT_KEY` environment variable (JSON content)
3. `GOOGLE_APPLICATION_CREDENTIALS` environment variable (file path)
4. Application Default Credentials (ADC)

All authentication paths converge to `oauth2.TokenSource` with scopes for Android Publisher and Play Developer Reporting APIs.

### API Client Pattern

The `api.Client` uses lazy initialization with `sync.Once` for each API service:
- `AndroidPublisher()` - Publishing, reviews, monetization, purchases
- `PlayReporting()` - Analytics and vitals
- `GamesManagement()` - Games API

The client includes:
- Automatic retry with exponential backoff for 429/5xx errors
- Respect for Retry-After headers
- Semaphore-based concurrency control (default: 3 concurrent calls)
- Exclusive locking for upload operations

### Output Structure

All commands return a consistent JSON envelope (`output.Result`):

```json
{
  "data": { ... },
  "error": {
    "code": "AUTH_FAILURE",
    "message": "...",
    "hint": "...",
    "details": {}
  },
  "meta": {
    "noop": false,
    "durationMs": 150,
    "services": ["androidpublisher"],
    "nextPageToken": "...",
    "warnings": []
  }
}
```

Exit codes (errors/codes.go):
- 0: Success
- 1: General API error
- 2: Authentication failure
- 3: Permission denied
- 4: Validation error
- 5: Rate limited
- 6: Network error
- 7: Not found
- 8: Conflict

### Error Handling Pattern

Use structured errors from `internal/errors`:
- Create errors with `errors.NewAPIError(code, message)`
- Chain with `.WithHint()`, `.WithDetails()`, `.WithHTTPStatus()`
- Convert to output with `output.NewErrorResult(err)`
- Common errors are predefined: `ErrAuthNotConfigured`, `ErrPermissionDenied`, `ErrPackageRequired`, etc.

### Edit Transaction Management

The `edits.Manager` handles Google Play's edit transaction lifecycle:
- Creates/validates/commits edit sessions
- File-based locking with PID tracking (prevents concurrent edits)
- Idempotency store for deduplication
- Automatic cleanup of stale edits (7 day TTL, 1 hour idle TTL)
- Supports `--edit-id` for explicit edit IDs and `--no-auto-commit` for manual commits

### CLI Command Pattern

Commands in `internal/cli/` follow this structure:

1. Define cobra.Command with Use, Short, Long, RunE
2. Add flags specific to the command
3. RunE calls a method on the CLI struct (e.g., `c.authStatus(ctx)`)
4. Method implementation:
   - Parse/validate inputs
   - Call API via `c.apiClient`
   - Handle errors with structured error types
   - Return `output.Result` via `c.output.Write(result)`
5. CLI struct tracks startTime for duration metrics

### Testing Patterns

- Use table-driven tests for multiple scenarios
- Test files are co-located with source: `auth.go` → `auth_test.go`
- Mock external dependencies (API clients, storage)
- Helper functions in `*_test.go` files (e.g., `auth/helpers_test.go`)
- Coverage reports via `make test-coverage`

### Linter Configuration

golangci-lint is configured (.golangci.yml) with:
- 30+ linters enabled (errcheck, gosec, revive, staticcheck, etc.)
- Function length limit: 150 lines, 80 statements
- Cyclomatic complexity limit: 40
- Test files exempt from dupl, funlen, goconst, gosec
- CLI package exempt from dupl (expected repetition in command definitions)

## Key Implementation Details

### Retry Logic
- Retries on 429 (rate limit) and 5xx errors only
- Exponential backoff with jitter: `initialDelay * 2^attempt + random(30%)`
- Respects Retry-After header if present
- Max 3 attempts by default (configurable via `api.WithMaxRetryAttempts`)

### Secure Storage
- Uses platform-specific keychains via `99designs/keyring`
- Service name: "gpd"
- Keys stored under account name derived from package/operation
- Never stores service account keys in config files

### PII Redaction
- Automatic redaction in logs for: emails, tokens, keys, package names, IDs
- Implemented in `internal/logging` package

### Concurrency
- Semaphore controls concurrent API calls (default: 3)
- Upload operations acquire all semaphore slots for exclusive access
- Edit operations use file-based locks with PID tracking

### Version Information
Build-time injection via ldflags in Makefile:
- `pkg/version.Version` - Git tag or "dev"
- `pkg/version.GitCommit` - Short commit hash
- `pkg/version.BuildTime` - RFC3339 timestamp

## Working with This Codebase

### Adding a New Command

1. Create or update file in `internal/cli/` (e.g., `vitals_commands.go`)
2. Define cobra.Command with appropriate flags
3. Add command to parent in `addXXXCommands()` method
4. Implement handler method on CLI struct
5. Use structured errors and output.Result
6. Update tests in corresponding `*_test.go` file

### Adding API Functionality

1. Extend `api.Client` if new service needed
2. Use `DoWithRetry()` wrapper for retryable operations
3. Acquire/Release semaphore for concurrency control
4. Convert API errors to structured `errors.APIError` types
5. Test retry behavior with mock failures

### Modifying Authentication

- Changes go in `internal/auth/auth.go` and `token_source.go`
- Maintain the priority chain (keyfile > env > GOOGLE_APPLICATION_CREDENTIALS > ADC)
- Update token storage in `token_storage.go` if credential format changes
- Test all credential sources in `auth_test.go`

### Configuration Changes

- Config file paths: OS-specific (macOS: ~/Library/Application Support/gpd, Linux: ~/.config/gpd, Windows: %APPDATA%\gpd)
- Environment variables prefixed with `GPD_` (e.g., `GPD_PACKAGE`, `GPD_TIMEOUT`)
- Update `config.go` and corresponding `config_commands.go`

## Notes

- Commands output JSON by default (minified single-line). Use `--pretty` for formatted JSON.
- The `--package` flag is required for most operations (can be set via `GPD_PACKAGE` env var or config).
- Edit transactions prevent concurrent modifications - file locks enforce this.
- OAuth tokens in testing mode expire after 7 days - use production mode or service accounts for automation.
- The API client lazy-loads services - only creates services when first accessed.
