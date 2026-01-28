# gpd - Google Play Developer CLI

A fast, lightweight command-line interface for the Google Play Developer Console. The Google Play equivalent to the App Store Connect CLI.

[![CI](https://github.com/dl-alexandre/gpd/actions/workflows/ci.yml/badge.svg)](https://github.com/dl-alexandre/gpd/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/dl-alexandre/gpd)](https://goreportcard.com/report/github.com/dl-alexandre/gpd)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## Features

- **Fast**: Sub-200ms cold start, minimal memory usage
- **AI-Agent Friendly**: JSON-first output, predictable exit codes, explicit flags
- **Secure**: Platform-specific credential storage, comprehensive PII redaction
- **Cross-Platform**: macOS, Linux, and Windows support
- **Comprehensive**: Full API coverage for publishing, reviews, analytics, monetization, vitals, purchases, and permissions
- **Edit Lifecycle Management**: Create, manage, validate, and commit edit transactions with `--edit-id` and `--no-auto-commit` flags
- **Advanced Vitals**: Error search, reporting, anomalies detection, and performance metrics (excessive wakeups, slow rendering, slow start, stuck wakelocks)
- **Monetization**: Complete subscriptions management with base plans, offers, batch operations, and regional pricing conversion
- **Purchase Management**: Voided purchases tracking, product/subscription acknowledge, consume, cancel, defer, refund, and revoke
- **Access Control**: Users and grants management for developer accounts and apps
- **Play Integrity**: Decode integrity tokens for mobile and Play Games on PC
- **Play Grouping**: Generate Play Grouping API tokens via Play Games Services
- **Custom App Publishing**: Create and publish custom apps for managed Play distribution

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

**Note:** gpd uses service account authentication only. OAuth user sign-in is not supported.

### OAuth Testing-Mode Limits
If you use OAuth credentials in testing mode, refresh tokens expire after 7 days and Google enforces a 100 refresh-token issuance cap per client. If you see repeated `invalid_grant` errors, re-authenticate and revoke unused tokens in Google Cloud Console or move the app to production to avoid the testing-mode limits.

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
| Flag | Description | Default |
|------|-------------|---------|
| `--package` | App package name | - |
| `--output` | Output format: json, table, markdown | json |
| `--pretty` | Pretty print JSON output | false |
| `--timeout` | Network timeout | 30s |
| `--key` | Service account key file path | - |
| `--quiet` | Suppress stderr except errors | false |
| `--verbose` | Verbose output | false |
| `--edit-id` | Explicit edit transaction ID (publish commands) | - |
| `--no-auto-commit` | Keep edit open for manual commit (publish commands) | false |

### Command Namespaces

#### `gpd auth` - Authentication

```bash
gpd auth status              # Check authentication status
gpd auth check --package ... # Validate permissions
gpd auth logout              # Clear stored credentials
```

#### `gpd config` - Configuration

```bash
gpd config init              # Initialize project configuration
gpd config doctor            # Diagnose configuration issues
gpd config path              # Show configuration file locations
gpd config get <key>         # Get a configuration value
gpd config set <key> <value> # Set a configuration value
gpd config completion bash   # Generate shell completions
```

#### `gpd publish` - App Publishing

```bash
# Edit lifecycle management
gpd publish edit create --package ...
gpd publish edit list --package ...
gpd publish edit get <edit-id> --package ...
gpd publish edit commit <edit-id> --package ...
gpd publish edit validate <edit-id> --package ...
gpd publish edit delete <edit-id> --package ...

# Upload artifacts (with edit support)
gpd publish upload app.aab --package com.example.app
gpd publish upload app.aab --package ... --edit-id <edit-id> --no-auto-commit

# Create/update releases (with edit support)
gpd publish release --package ... --track internal --status draft
gpd publish release --package ... --track production --status inProgress --version-code 123
gpd publish release --package ... --track beta --edit-id <edit-id> --no-auto-commit

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

# App details
gpd publish details get --package ...
gpd publish details update --package ... --contact-email support@example.com
gpd publish details patch --package ... --contact-phone "+1234567890" --update-mask contactPhone

# Images API
gpd publish images upload icon icon.png --package ... --locale en-US
gpd publish images list phoneScreenshots --package ... --locale en-US
gpd publish images delete phoneScreenshots <image-id> --package ... --locale en-US
gpd publish images deleteall featureGraphic --package ... --locale en-US

# Assets
gpd publish assets upload ./assets --package ...
gpd publish assets spec

# Deobfuscation files
gpd publish deobfuscation upload mapping.txt --package ... --type proguard --version-code 123

# Internal app sharing
gpd publish internal-share upload app.aab --package ...

# Testers
gpd publish testers list --package ... --track internal
gpd publish testers add --package ... --track internal --group testers@example.com
```

#### `gpd customapp` - Custom App Publishing

```bash
# Create a custom app (APK required)
gpd customapp create --account 1234567890 --title "My Custom App" --language en-US --apk app.apk

# Restrict access to specific organizations
gpd customapp create --account 1234567890 --title "My App" --language en-US --apk app.apk --org-id 0123456789
```

#### `gpd migrate` - Metadata Migration

```bash
# Validate fastlane metadata locally
gpd migrate fastlane validate --dir fastlane/metadata/android

# Export Google Play listings to fastlane format
gpd migrate fastlane export --package ... --output fastlane/metadata/android
gpd migrate fastlane export --package ... --output fastlane/metadata/android --include-images
gpd migrate fastlane export --package ... --output fastlane/metadata/android --locales en-US,ja-JP

# Import fastlane metadata into Google Play
gpd migrate fastlane import --package ... --dir fastlane/metadata/android --dry-run
gpd migrate fastlane import --package ... --dir fastlane/metadata/android --replace-images
gpd migrate fastlane import --package ... --dir fastlane/metadata/android --skip-images
gpd migrate fastlane import --package ... --dir fastlane/metadata/android --edit-id <edit-id> --no-auto-commit
```

#### `gpd reviews` - Review Management

```bash
# List reviews with filtering
gpd reviews list --package ... --min-rating 1 --max-rating 3
gpd reviews list --package ... --include-review-text --scan-limit 200

# Reply to reviews
gpd reviews reply --package ... --review-id abc123 --text "Thank you!"
gpd reviews reply --package ... --review-id abc123 --template-file reply.txt

# View capabilities
gpd reviews capabilities
```

#### `gpd purchases` - Purchase Verification

```bash
# Verify a purchase token
gpd purchases verify --package ... --token <token> --product-id sku123

# Voided purchases
gpd purchases voided list --package ... --start-time 2024-01-01T00:00:00Z --type product

# Product purchase actions
gpd purchases products acknowledge --package ... --product-id sku123 --token <token>
gpd purchases products consume --package ... --product-id sku123 --token <token>

# Subscription purchase actions
gpd purchases subscriptions acknowledge --package ... --subscription-id sub123 --token <token>
gpd purchases subscriptions cancel --package ... --subscription-id sub123 --token <token>
gpd purchases subscriptions defer --package ... --subscription-id sub123 --token <token> \
  --expected-expiry-time 2024-12-31T23:59:59Z --desired-expiry-time 2025-01-31T23:59:59Z
gpd purchases subscriptions refund --package ... --subscription-id sub123 --token <token>
gpd purchases subscriptions revoke --package ... --token <token> --revoke-type fullRefund

# View capabilities
gpd purchases capabilities
```

#### `gpd analytics` - App Analytics

```bash
# Query analytics data
gpd analytics query --package ... --start-date 2024-01-01 --end-date 2024-01-31

# View capabilities
gpd analytics capabilities
```

#### `gpd integrity` - Play Integrity

```bash
# Decode a standard integrity token
gpd integrity decode --package ... --token <token>
```

#### `gpd grouping` - Play Grouping

```bash
# Generate a Play Grouping API token
gpd grouping token --package ... --persona user-123

# Generate a Play Grouping API token using Recall
gpd grouping token-recall --package ... --persona user-123 --recall-session-id <session-id>
```

#### `gpd permissions` - Access Control

```bash
# Users management
gpd permissions users create --developer-id <id> --email user@example.com --developer-permissions CAN_VIEW_FINANCIAL_DATA_GLOBAL
gpd permissions users list --developer-id <id>
gpd permissions users get developers/<id>/users/user@example.com
gpd permissions users patch developers/<id>/users/user@example.com --developer-permissions CAN_MANAGE_PERMISSIONS_GLOBAL
gpd permissions users delete developers/<id>/users/user@example.com

# Grants management
gpd permissions grants create --package ... --email user@example.com --app-permissions CAN_REPLY_TO_REVIEWS
gpd permissions grants patch developers/<id>/users/user@example.com/grants/com.example.app --app-permissions CAN_MANAGE_PUBLIC_APKS
gpd permissions grants delete developers/<id>/users/user@example.com/grants/com.example.app

# List available permissions
gpd permissions grants create --list-permissions

# View capabilities
gpd permissions capabilities
```

#### `gpd vitals` - Android Vitals

```bash
# Query crash data
gpd vitals crashes --package ... --start-date 2024-01-01 --end-date 2024-01-31

# Query ANR data
gpd vitals anrs --package ... --start-date 2024-01-01 --end-date 2024-01-31

# Additional performance metrics
gpd vitals excessive-wakeups --package ... --start-date 2024-01-01 --end-date 2024-01-31
gpd vitals slow-rendering --package ... --start-date 2024-01-01 --end-date 2024-01-31
gpd vitals slow-start --package ... --start-date 2024-01-01 --end-date 2024-01-31
gpd vitals stuck-wakelocks --package ... --start-date 2024-01-01 --end-date 2024-01-31

# Error search and reporting
gpd vitals errors issues search --package ... --query "NullPointerException" --interval last30Days
gpd vitals errors reports search --package ... --query "crash" --interval last7Days --deobfuscate
gpd vitals errors counts get --package ...
gpd vitals errors counts query --package ... --start-date 2024-01-01 --end-date 2024-01-31

# Anomalies detection
gpd vitals anomalies list --package ... --metric crashRate --time-period last30Days

# View capabilities
gpd vitals capabilities
```

#### `gpd monetization` - In-App Products & Subscriptions

```bash
# One-time products (managed/consumable)
gpd monetization products list --package ...
gpd monetization products get sku123 --package ...
gpd monetization products create --package ... --product-id sku123 --type managed --default-price 990000
gpd monetization products update --package ... sku123 --status inactive
gpd monetization products delete --package ... sku123

# Subscriptions CRUD
gpd monetization subscriptions list --package ...
gpd monetization subscriptions get sub123 --package ...
gpd monetization subscriptions create --package ... --product-id sub123 --file subscription.json
gpd monetization subscriptions update --package ... sub123 --file subscription.json
gpd monetization subscriptions patch --package ... sub123 --file subscription.json --update-mask basePlans
gpd monetization subscriptions delete --package ... sub123 --confirm
gpd monetization subscriptions archive --package ... sub123

# Batch operations
gpd monetization subscriptions batchGet --package ... --ids sub1,sub2,sub3
gpd monetization subscriptions batchUpdate --package ... --file batch-update.json

# Base plans management
gpd monetization baseplans activate --package ... sub123 plan456
gpd monetization baseplans deactivate --package ... sub123 plan456
gpd monetization baseplans delete --package ... sub123 plan456 --confirm
gpd monetization baseplans migrate-prices --package ... sub123 plan456 --region-code US --price-micros 999000
gpd monetization baseplans batch-migrate-prices --package ... sub123 --file migrate.json
gpd monetization baseplans batch-update-states --package ... sub123 --file states.json

# Offers management
gpd monetization offers create --package ... sub123 plan456 --offer-id offer789 --file offer.json
gpd monetization offers get --package ... sub123 plan456 offer789
gpd monetization offers list --package ... sub123 plan456
gpd monetization offers delete --package ... sub123 plan456 offer789 --confirm
gpd monetization offers activate --package ... sub123 plan456 offer789
gpd monetization offers deactivate --package ... sub123 plan456 offer789
gpd monetization offers batchGet --package ... sub123 plan456 --offer-ids offer1,offer2
gpd monetization offers batchUpdate --package ... sub123 plan456 --file offers.json
gpd monetization offers batchUpdateStates --package ... sub123 plan456 --file states.json

# One-time products (alias)
gpd monetization onetimeproducts list --package ...
gpd monetization onetimeproducts get sku123 --package ...
gpd monetization onetimeproducts create --package ... --product-id sku123 --type consumable
gpd monetization onetimeproducts update --package ... sku123 --default-price 1990000
gpd monetization onetimeproducts delete --package ... sku123

# Regional pricing conversion
gpd monetization convert-region-prices --package ... --price-micros 990000 --currency USD --to-regions US,GB,JP
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
