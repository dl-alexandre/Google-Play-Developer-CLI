# gpd - Google Play Developer CLI

A fast, lightweight command-line interface for the Google Play Developer Console. The Google Play equivalent to the App Store Connect CLI.

[![CI](https://github.com/dl-alexandre/gpd/actions/workflows/ci.yml/badge.svg)](https://github.com/dl-alexandre/gpd/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dl-alexandre/gpd)](https://goreportcard.com/report/github.com/dl-alexandre/gpd)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## Test Coverage

- Current statement coverage: 31.5% (from `coverage.out`)
- Lowest coverage areas:
  - `internal/cli` (~22%)
  - `cmd/gpd` (~50%)
  - `internal/edits` (~64%)
  - `internal/migrate/fastlane` (~82%)
  - `internal/api` (~89%)
- Gaps to fill next: CLI command handlers, edit lock/content flows, API client edge conditions

## Features

- **Fast**: Sub-200ms cold start, minimal memory usage
- **AI-Agent Friendly**: JSON-first output, predictable exit codes, explicit flags
- **Secure**: Platform-specific credential storage, comprehensive PII redaction
- **Cross-Platform**: macOS, Linux, and Windows support
- **Comprehensive**: Full API coverage for publishing, reviews, analytics, and monetization

## Installation

### Homebrew (macOS/Linux)

```bash
brew tap dl-alexandre/tap
brew install gpd
```

### Install Script (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
```

### Go Install

```bash
go install github.com/dl-alexandre/gpd/cmd/gpd@latest
```

### Download Binary

Download the latest release from the [Releases](https://github.com/dl-alexandre/gpd/releases) page.

### Build from Source

```bash
git clone https://github.com/dl-alexandre/gpd.git
cd gpd
make build
```

## Quick Start

### 1. Set Up Authentication

Create a service account in Google Cloud Console with the Google Play Android Publisher API enabled, then:

```bash
# Option 1: Environment variable
export GPD_SERVICE_ACCOUNT_KEY='{"type": "service_account", ...}'

# Option 2: Key file
gpd --key /path/to/service-account.json auth status

# Option 3: Application Default Credentials
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
gpd auth status
```

### 2. Verify Setup

```bash
# Check authentication status
gpd auth status

# Check permissions for a specific app
gpd auth check --package com.example.app

# Diagnose configuration issues
gpd config doctor
```

### 3. Start Using

```bash
# Upload an app bundle
gpd publish upload app.aab --package com.example.app

# Create a release
gpd publish release --package com.example.app --track internal --status draft

# List reviews
gpd reviews list --package com.example.app --min-rating 1 --max-rating 3
```

## Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--package` | App package name | - |
| `--output` | Output format: json, table, markdown | json |
| `--pretty` | Pretty print JSON output | false |
| `--timeout` | Network timeout | 30s |
| `--key` | Service account key file path | - |
| `--quiet` | Suppress stderr except errors | false |
| `--verbose` | Verbose output | false |

### Command Namespaces

#### `gpd auth` - Authentication

```bash
gpd auth login              # OAuth device login (uses GPD_CLIENT_ID)
gpd auth init               # Alias for auth login
gpd auth switch <profile>   # Switch active profile
gpd auth list               # List stored profiles
gpd auth status              # Check authentication status
gpd auth check --package ... # Validate permissions
gpd auth logout              # Clear stored credentials
gpd auth diagnose            # Detailed auth diagnostics
gpd auth doctor              # Diagnose authentication setup
```

If your OAuth consent screen is in testing mode, refresh tokens can expire after 7 days and Google enforces a 100 refresh-token issuance cap per OAuth client. If you encounter repeated `invalid_grant` refresh failures, re-authenticate and revoke unused tokens in Google Cloud Console, or move the app to production.

#### `gpd config` - Configuration

```bash
gpd config init              # Initialize project configuration
gpd config doctor            # Diagnose configuration issues
gpd config path              # Show configuration file locations
gpd config get <key>         # Get a configuration value
gpd config set <key> <value> # Set a configuration value
gpd config completion bash   # Generate shell completions
```

#### `gpd apps` - App Discovery

```bash
# List apps in the developer account
gpd apps list

# Get app details by package
gpd apps get com.example.app
```

#### `gpd publish` - App Publishing

```bash
# Upload artifacts
gpd publish upload app.aab --package com.example.app

# List and inspect builds
gpd publish builds list --package ...
gpd publish builds get 123 --package ...
gpd publish builds expire 123 --package ... --confirm
gpd publish builds expire-all --package ... --confirm

# ASC beta-group compatibility workflow
gpd publish beta-groups list --package ...
gpd publish beta-groups get internal --package ...
gpd publish beta-groups add-testers internal --group qa@example.com --package ...

# Create/update releases
gpd publish release --package ... --track internal --status draft
gpd publish release --package ... --track production --status inProgress --version-code 123

# Manage rollouts
gpd publish rollout --package ... --track production --percentage 10
gpd publish promote --package ... --from-track beta --to-track production
gpd publish halt --package ... --track production --confirm
gpd publish rollback --package ... --track production --confirm

# View status
gpd publish status --package ... --track production
gpd publish tracks --package ...
gpd publish capabilities

# Store listing
gpd publish listing update --package ... --locale en-US --title "My App"
gpd publish listing get --package ...
gpd publish listing delete --package ... --locale en-US --confirm
gpd publish listing delete-all --package ... --confirm

# Assets
gpd publish assets upload ./assets --package ...
gpd publish assets spec

# Testers
gpd publish testers list --package ... --track internal
gpd publish testers add --package ... --track internal --group testers@example.com
gpd publish testers get --package ... --track internal
```

#### `gpd reviews` - Review Management

```bash
# List reviews with filtering
gpd reviews list --package ... --min-rating 1 --max-rating 3
gpd reviews list --package ... --include-review-text --scan-limit 200

# Reply to reviews
gpd reviews reply --package ... --review-id abc123 --text "Thank you!"
gpd reviews reply --package ... --review-id abc123 --template-file reply.txt

# Get a review
gpd reviews get --review-id abc123

# Get the developer response for a review
gpd reviews response get --review-id abc123
gpd reviews response for-review --review-id abc123

# View capabilities
gpd reviews capabilities
```

#### `gpd purchases` - Purchase Verification

```bash
# Verify a purchase token
gpd purchases verify --package ... --token <token> --product-id sku123

# View capabilities
gpd purchases capabilities
```

#### `gpd analytics` - App Analytics

```bash
# Query analytics data (vitals metrics)
gpd analytics query --package ... --metrics crashRate --start-date 2024-01-01 --end-date 2024-01-31

# View capabilities
gpd analytics capabilities
```

#### `gpd vitals` - Android Vitals

```bash
# Query crash data
gpd vitals crashes --package ... --start-date 2024-01-01 --end-date 2024-01-31

# Query ANR data
gpd vitals anrs --package ... --start-date 2024-01-01 --end-date 2024-01-31

# View capabilities
gpd vitals capabilities
```

#### `gpd monetization` - In-App Products

```bash
# List products
gpd monetization products list --package ...

# Get product details
gpd monetization products get sku123 --package ...

# Create/update products
gpd monetization products create --package ... --product-id sku123 --type managed

# List subscriptions (read-only)
gpd monetization subscriptions list --package ...
```

## Output Format

All commands output JSON with a consistent envelope structure:

```json
{
  "data": { ... },
  "error": null,
  "meta": {
    "noop": false,
    "durationMs": 150,
    "services": ["androidpublisher"],
    "nextPageToken": "...",
    "warnings": []
  }
}
```

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General API error |
| 2 | Authentication failure |
| 3 | Permission denied |
| 4 | Validation error |
| 5 | Rate limited |
| 6 | Network error |
| 7 | Not found |
| 8 | Conflict |

## Configuration

Configuration files are stored in OS-appropriate locations:

| OS | Config Directory | Cache Directory |
|----|------------------|-----------------|
| macOS | `~/Library/Application Support/gpd` | `~/Library/Caches/gpd` |
| Linux | `~/.config/gpd` | `~/.cache/gpd` |
| Windows | `%APPDATA%\gpd` | `%LOCALAPPDATA%\gpd` |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `GPD_SERVICE_ACCOUNT_KEY` | Service account JSON key content |
| `GPD_PACKAGE` | Default package name |
| `GPD_TIMEOUT` | Network timeout |
| `GOOGLE_APPLICATION_CREDENTIALS` | Path to service account key file |

## AI Agent Integration

gpd is designed for programmatic access by AI agents:

```bash
# Get the AI agent quickstart guide
gpd help agent
```

Key features for automation:
- Minified JSON output by default (single-line)
- Predictable exit codes for error handling
- Explicit flags over interactive prompts
- No browser-based authentication
- `--dry-run` flag for safe operation planning

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
gpd config completion bash > /etc/bash_completion.d/gpd

# Zsh
gpd config completion zsh > "${fpath[1]}/_gpd"

# Fish
gpd config completion fish > ~/.config/fish/completions/gpd.fish
```

## Security

- Credentials are stored in platform-specific secure storage (Keychain, Secret Service, Credential Manager)
- PII is automatically redacted from logs
- Service account keys are never stored in configuration files
- All API communications use HTTPS with certificate validation

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0 - see [LICENSE](LICENSE) for details.

## Related Projects

- [App Store Connect CLI](https://github.com/ittybittyapps/appstoreconnect-cli) - Similar tool for iOS/macOS apps
- [fastlane](https://fastlane.tools/) - Automation for iOS and Android
- [gradle-play-publisher](https://github.com/Triple-T/gradle-play-publisher) - Gradle plugin for Android publishing

## App Store Connect CLI Parity

See the parity matrix in [docs/asc-parity.md](docs/asc-parity.md).
