# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2024-01-23

### Added

- Initial release of gpd - Google Play Developer CLI
- **Authentication**
  - Service account key authentication
  - Application Default Credentials (ADC) support
  - Environment variable authentication (`GPD_SERVICE_ACCOUNT_KEY`)
  - Platform-specific secure credential storage (Keychain, Secret Service, Credential Manager)
  - `gpd auth status` - Check authentication status
  - `gpd auth check` - Validate service account permissions
  - `gpd auth logout` - Clear stored credentials

- **Configuration**
  - OS-appropriate configuration directories
  - `gpd config init` - Initialize project configuration
  - `gpd config doctor` - Diagnose configuration and credential issues
  - `gpd config path` - Show configuration file locations
  - `gpd config get/set` - Manage configuration values
  - `gpd config completion` - Generate shell completions (bash, zsh, fish)

- **Publishing**
  - `gpd publish upload` - Upload AAB/APK artifacts with SHA256 caching
  - `gpd publish release` - Create releases on tracks (internal, alpha, beta, production)
  - `gpd publish rollout` - Update staged rollout percentage
  - `gpd publish promote` - Promote releases between tracks
  - `gpd publish halt` - Halt production rollouts
  - `gpd publish rollback` - Rollback to previous versions
  - `gpd publish status` - Get track status
  - `gpd publish tracks` - List all tracks
  - `gpd publish capabilities` - List publishing capabilities
  - `gpd publish listing` - Manage store listing (title, description, localization)
  - `gpd publish assets` - Upload and manage screenshots and graphics
  - `gpd publish testers` - Manage tester groups

- **Reviews**
  - `gpd reviews list` - List reviews with client-side filtering
  - `gpd reviews reply` - Reply to reviews with rate limiting and templating
  - `gpd reviews capabilities` - List review API capabilities

- **Purchases**
  - `gpd purchases verify` - Verify purchase tokens (products and subscriptions v2)
  - `gpd purchases capabilities` - List purchase verification capabilities

- **Analytics**
  - `gpd analytics query` - Query app analytics data
  - `gpd analytics capabilities` - List available metrics and dimensions

- **Vitals**
  - `gpd vitals crashes` - Query crash rate data
  - `gpd vitals anrs` - Query ANR rate data
  - `gpd vitals query` - Generic vitals query
  - `gpd vitals capabilities` - List vitals capabilities and thresholds

- **Monetization**
  - `gpd monetization products list/get/create/update/delete` - Manage in-app products
  - `gpd monetization subscriptions list/get` - View subscriptions (read-only)
  - `gpd monetization capabilities` - List monetization capabilities

- **Core Features**
  - JSON-first output with consistent envelope structure (`{data, error, meta}`)
  - Minified JSON by default, `--pretty` for human-readable output
  - Table and markdown output formats
  - Standardized exit codes (0-8) for scripting
  - Edit transaction management with file locking
  - Idempotency support with SHA256 artifact caching
  - PII redaction in logs
  - Cross-platform support (macOS, Linux, Windows)
  - `--dry-run` flag for safe operation planning
  - AI agent quickstart guide (`gpd help agent`)

### Security

- Platform-specific secure credential storage
- PII automatically redacted from logs
- Service account keys never stored in config files
- HTTPS-only API communications

[Unreleased]: https://github.com/google-play-cli/gpd/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/google-play-cli/gpd/releases/tag/v0.1.0
