# Agent Guidelines for gpd

Quick reference for AI agents. See linked docs for details.

## Quick Reference

| Topic | Reference |
|-------|-----------|
| Commands | [README.md](./README.md#command-reference) |
| API Coverage | [docs/api-coverage-matrix.md](./docs/api-coverage-matrix.md) |
| Code Style | [CONTRIBUTING.md](./CONTRIBUTING.md#code-style) |
| Workflows | [docs/examples/](./docs/examples/) |

---

## Table of Contents

1. [Build & Test](#build--test)
2. [Code Style](#code-style)
3. [Architecture](#architecture)
4. [Project Structure](#project-structure)
5. [Key Packages](#key-packages)
6. [Important Notes](#important-notes)

---

## Build & Test

See: [CONTRIBUTING.md - Development Setup](./CONTRIBUTING.md#development-setup)

```bash
make build              # Build current platform
make build-all          # All platforms
make test               # Tests with race + coverage
make test-coverage      # HTML coverage report
make lint               # golangci-lint
go test ./internal/auth # Specific package
go test -v -run TestX ./pkg # Single test
```

---

## Code Style

See: [CONTRIBUTING.md - Code Style](./CONTRIBUTING.md#code-style)

- **Imports**: stdlib → third-party → internal (blank lines between)
- **Naming**: PascalCase (exported), camelCase (unexported)
- **Errors**: Use `internal/errors` with `WithHint()`, `WithDetails()`
- **Never suppress errors**: No `as any`, `@ts-ignore`
- **Max function**: 150 lines, 80 statements, complexity 40
- **Testing**: Table-driven, co-located `*_test.go`, use `t.Helper()`

---

## Architecture

See: [internal/api/client.go](./internal/api/client.go), [internal/cli/*.go](./internal/cli/)

### API Client
- Lazy init with `sync.Once`
- Retry with exponential backoff via `DoWithRetry()`
- Options pattern for config

### CLI Commands
1. Define `cobra.Command` with `Use`, `Short`, `Long`, `RunE`
2. `RunE` calls handler method on CLI struct
3. Return `output.Result` via `c.output.Write()`

### Output Structure
```json
{"data": {...}, "error": {...}, "meta": {"durationMs": 150, "services": [...]}}
```

### Exit Codes
0=Success, 1=API, 2=Auth, 3=Permission, 4=Validation, 5=RateLimit, 6=Network, 7=NotFound, 8=Conflict

---

## Project Structure

```
cmd/gpd/           # Entry point
internal/
  api/             # Google Play API client (lazy init, retry)
  auth/            # Service account, OAuth, ADC
  cli/             # 41 command files (Cobra)
  edits/           # Edit transaction lifecycle + file locking
  errors/          # Structured errors + exit codes
  output/          # JSON envelope
  storage/         # Platform keychain (Keychain/Secret Service)
  config/          # Config file management
  logging/         # PII redaction
docs/examples/     # Workflow guides
```

---

## Key Packages

| Package | Purpose |
|---------|---------|
| `internal/api` | API client, lazy init, retry |
| `internal/auth` | Multi-source auth |
| `internal/cli` | All CLI commands |
| `internal/edits` | Transactional edits |
| `internal/errors` | Structured errors |
| `internal/output` | JSON envelope |
| `internal/storage` | Secure credential storage |

---

## Important Notes

- Go version: 1.24.0
- **Never commit unless explicitly requested**
- JSON-first output, predictable exit codes
- Credentials in platform keychains only
- PII auto-redacted from logs

---

## Example Workflows

| Guide | Description |
|-------|-------------|
| [edit-workflow.md](./docs/examples/edit-workflow.md) | Edit transactions |
| [subscription-management.md](./docs/examples/subscription-management.md) | Monetization |
| [ci-cd-integration.md](./docs/examples/ci-cd-integration.md) | CI/CD pipelines |
| [error-debugging.md](./docs/examples/error-debugging.md) | Android Vitals |

---

## Common Commands

```bash
gpd auth status                    # Check auth
gpd auth check --package <pkg>     # Validate permissions
gpd publish upload <file> --package <pkg>
gpd publish edit create --package <pkg>
gpd publish release --package <pkg> --track internal
gpd reviews list --package <pkg> --min-rating 1
gpd config doctor                  # Diagnose issues
```