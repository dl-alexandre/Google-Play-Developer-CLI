# gpd – Google Play Developer CLI

**The fast, lightweight, zero-Ruby alternative to Fastlane for Google Play Console.**

[![Go Report Card](https://goreportcard.com/badge/github.com/dl-alexandre/Google-Play-Developer-CLI)](https://goreportcard.com/report/github.com/dl-alexandre/Google-Play-Developer-CLI)
[![Release](https://img.shields.io/github/v/release/dl-alexandre/Google-Play-Developer-CLI?color=green)](https://github.com/dl-alexandre/Google-Play-Developer-CLI/releases)
[![CI](https://github.com/dl-alexandre/Google-Play-Developer-CLI/actions/workflows/ci.yml/badge.svg)](https://github.com/dl-alexandre/Google-Play-Developer-CLI/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24.0-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Platforms](https://img.shields.io/badge/platforms-macOS%20%7C%20Linux%20%7C%20Windows-blue)](https://github.com/dl-alexandre/Google-Play-Developer-CLI/releases)
[![Downloads](https://img.shields.io/github/downloads/dl-alexandre/Google-Play-Developer-CLI/total)](https://github.com/dl-alexandre/Google-Play-Developer-CLI/releases)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Stars](https://img.shields.io/github/stars/dl-alexandre/Google-Play-Developer-CLI?style=social)](https://github.com/dl-alexandre/Google-Play-Developer-CLI)

- ⚡ **Sub-200ms cold start** — no Ruby VM bloat, instant even on slow CI runners
- 🖥️ **JSON-first output** — built for AI agents, scripts, and automation (predictable exit codes, no parsing HTML)
- 🔒 **Secure by default** — platform keystore + PII redaction, never leak service account keys
- ✅ **Full Play Console coverage** — publishing, reviews, analytics, vitals, monetization, integrity + 61+ commands
- 🛠️ **Cross-platform** — macOS, Linux, Windows via Homebrew, curl, or Go install

If Fastlane feels heavy/slow, gpd gives you the same power with a tiny footprint and modern DX.

**[Quick Install →](#quick-install)**

---

## gpd vs Fastlane

| Feature | gpd (Google-Play-Developer-CLI) | Fastlane (supply) |
|---------|----------------------------------|-------------------|
| Cold start time | <200ms | 2–10s+ (Ruby boot) |
| Runtime footprint | ~10–20 MB | 100+ MB + gems |
| Language / Deps | Go – single binary | Ruby + heavy dependencies |
| Output format | JSON-first, machine-readable | Human text + plugins |
| AI/Script friendliness | Designed for agents (exit codes, structured) | Requires parsing plugins |
| Credential security | Platform store + redaction | Manual JSON key handling |
| Play API coverage | Full (publishing, reviews, analytics, monetization, integrity…) | Good, but fragmented actions |
| Install size / ease | Homebrew / curl / go install | gem install + bundle |
| Windows support | Native | WSL or painful |

**Choose gpd when you want speed, security, and scriptability without the Ruby tax.**

---

## Quick Install

```bash
# macOS/Linux via Homebrew (10 seconds)
brew tap dl-alexandre/tap && brew install gpd

# OR macOS/Linux via curl
curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Play-Developer-CLI/main/install.sh | bash

# OR via Go
go install github.com/dl-alexandre/Google-Play-Developer-CLI/cmd/gpd@latest
```

**[View all installation options →](#installation)**

---

## Quick Start

```bash
# 1. Authenticate once (service account JSON)
gpd auth login --key service-account.json

# 2. Verify setup
gpd auth check --package com.example.app

# 3. Upload AAB to internal track – JSON output ready for scripts
gpd publish upload myapp.aab --package com.example.app --track internal --json
```

**[Full auth guide →](#authentication)**

---

## Why gpd Over Fastlane?

<details>
<summary><b>Click to see why developers are switching</b></summary>

### Speed That Matters
- Cold start in <200ms vs Ruby's 2–10s boot time
- CI pipelines finish faster, saving compute costs
- Instant feedback during development

### Built for Modern Workflows
- JSON-first output designed for AI agents and automation
- Predictable exit codes for reliable scripting
- No browser-based auth interruptions

### Security-First Design
- Credentials stored in platform keychain (macOS Keychain, Linux Secret Service, Windows Credential Manager)
- Automatic PII redaction from all logs
- Service account keys never touch disk unencrypted

### Zero Dependencies
- Single static binary—no gemfile nightmares
- Works on Windows natively (no WSL required)
- Same behavior everywhere

### Full API Coverage
61+ commands covering:
- App publishing and release management
- Review monitoring and replies
- Android Vitals (crashes, ANRs)
- Monetization (IAPs, subscriptions)
- Play Integrity API
- Play Games Services
- Custom app publishing
- User permissions management

</details>

---

## AI Agent Integration

gpd is purpose-built for programmatic access by AI agents and automation:

```bash
# Get the AI agent quickstart guide
gpd help agent

# Example: Query reviews and get structured JSON for processing
gpd reviews list --package com.example.app --min-rating 1 --output json --pretty
```

**Key features for automation:**

- **Minified JSON by default** — single-line output for easy parsing
- **Predictable exit codes** — 0=success, 1-8=specific error types
- **Explicit flags** — no interactive prompts to hang scripts
- **No browser auth** — fully headless operation
- **Dry-run mode** — plan operations safely with `--dry-run`

### Sample JSON Output

```json
{
  "data": {
    "reviews": [
      {
        "reviewId": "12345",
        "authorName": "Jane D.",
        "comments": [{"text": "Great app!"}],
        "starRating": 5
      }
    ]
  },
  "error": null,
  "meta": {
    "durationMs": 145,
    "services": ["androidpublisher"],
    "nextPageToken": null,
    "warnings": []
  }
}
```

---

## Features

- **Fast**: Sub-200ms cold start, minimal memory usage
- **AI-Agent Friendly**: JSON-first output, predictable exit codes, explicit flags
- **Secure**: Platform-specific credential storage, comprehensive PII redaction
- **Cross-Platform**: Native macOS, Linux, and Windows support
- **Comprehensive**: Full API coverage for publishing, reviews, analytics, and monetization

---

## Installation

### Quick Install

```bash
# One-liner for macOS/Linux
brew tap dl-alexandre/tap && brew install gpd
```

### Homebrew (macOS/Linux)

```bash
brew tap dl-alexandre/tap
brew install gpd
```

### Install Script (macOS/Linux)

```bash
curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Play-Developer-CLI/main/install.sh | bash
```

### Go Install

```bash
go install github.com/dl-alexandre/Google-Play-Developer-CLI/cmd/gpd@latest
```

### Download Binary

Download the latest release from the [Releases](https://github.com/dl-alexandre/Google-Play-Developer-CLI/releases) page.

### Build from Source

```bash
git clone https://github.com/dl-alexandre/Google-Play-Developer-CLI.git
cd Google-Play-Developer-CLI
make build
```

---

## Authentication

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

### Verify Setup

```bash
# Check authentication status
gpd auth status

# Check permissions for a specific app
gpd auth check --package com.example.app

# Diagnose configuration issues
gpd config doctor
```

---

## Command Reference

### Global Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--package` | App package name | - |
| `--output` | Output format: json, table, markdown, csv (analytics/vitals only) | json |
| `--pretty` | Pretty print JSON output | false |
| `--timeout` | Network timeout | 30s |
| `--key` | Service account key file path | - |
| `--quiet` | Suppress stderr except errors | false |
| `--verbose` | Verbose output | false |
| `--profile` | Authentication profile name | - |
| `--store-tokens` | Token storage: auto, never, secure | auto |
| `--fields` | JSON field projection (comma-separated paths) | - |
| `-v, --version` | Print version information | false |

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

# Release workflow mapping (ASC submit/versions parity)
gpd publish capabilities
docs/examples/release-workflow.md

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

# Manage offers
gpd monetization offers list --package ... --subscription-id sub123
gpd monetization offers get sub123 offer123 --package ...

# Manage base plans
gpd monetization baseplans list --package ... --subscription-id sub123
gpd monetization baseplans activate sub123 base-plan-id --package ...
gpd monetization baseplans deactivate sub123 base-plan-id --package ...
```

#### `gpd customapp` - Custom App Publishing

```bash
# Create a custom app
gpd customapp create --name "My App" --title "My App Title" --category GAMES
```

#### `gpd games` - Play Games Services

```bash
# Manage achievements
gpd games achievements list --package ...
gpd games achievements get achievement-id --package ...

# Manage events
gpd games events list --package ...
gpd games events get event-id --package ...

# Manage scores (leaderboards)
gpd games scores list --package ... --leaderboard-id leaderboard-id

# Manage player visibility
gpd games players get --player-id player123
gpd games players update --player-id player123 --visibility VISIBLE

# Manage applications
gpd games applications list
gpd games applications get application-id

# View capabilities
gpd games capabilities
```

#### `gpd grouping` - Play Grouping API

```bash
# Generate Play Grouping API tokens
gpd grouping token --package ...

# Generate Recall tokens
gpd grouping token-recall --package ...
```

#### `gpd integrity` - Play Integrity API

```bash
# Decode a Play Integrity token
gpd integrity decode --token <integrity-token>
```

#### `gpd migrate` - Metadata Migration

```bash
# Migrate fastlane supply format
gpd migrate fastlane /path/to/metadata --package ...
```

#### `gpd permissions` - Permissions Management

```bash
# Manage developer account users
gpd permissions users list
gpd permissions users get user123
gpd permissions users invite --email user@example.com --role viewer

# Manage app-level permission grants
gpd permissions grants list --package ...
gpd permissions grants get grant123 --package ...
gpd permissions grants create --package ... --user user@example.com --permission release

# View capabilities
gpd permissions capabilities
```

#### `gpd recovery` - App Recovery

```bash
# List recovery actions
gpd recovery list --package ...

# Create a recovery action
gpd recovery create --package ... --region us --version-code 123

# Deploy a recovery action
gpd recovery deploy action123 --package ... --confirm

# Cancel a recovery action
gpd recovery cancel action123 --package ... --confirm

# Add targeting to a recovery action
gpd recovery add-targeting action123 --package ... --country us

# View capabilities
gpd recovery capabilities
```

---

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

---

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

---

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

---

## Security

- Credentials are stored in platform-specific secure storage (Keychain, Secret Service, Credential Manager)
- PII is automatically redacted from logs
- Service account keys are never stored in configuration files
- All API communications use HTTPS with certificate validation

---

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

---

## Support

**⭐ Star this repo** if it saves you time in CI/CD pipelines.

**🐛 Open an issue** for bugs or feature requests.

**💙 Sponsor** if gpd helps your business—sponsors get priority support.

---

## Related Projects

- [App Store Connect CLI](https://github.com/ittybittyapps/appstoreconnect-cli) - Similar tool for iOS/macOS apps
- [fastlane](https://fastlane.tools/) - Automation for iOS and Android
- [gradle-play-publisher](https://github.com/Triple-T/gradle-play-publisher) - Gradle plugin for Android publishing

---

## App Store Connect CLI Parity

See the parity matrix in [docs/asc-parity.md](docs/asc-parity.md).

---

## License

MIT License - see [LICENSE](LICENSE) for details.
