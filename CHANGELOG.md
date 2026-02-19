# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.4.5] - 2026-02-19

### Fixed
- Fixed `TestValidatePathExtended` to use platform-appropriate paths for Windows (was using Unix paths causing CI failure)

## [0.4.4] - 2026-02-19

### Fixed
- Fixed race condition in `TestCryptoRandFloat64Error` by removing `t.Parallel()` from test that modifies global state

## [0.2.0] - 2026-01-24

### Added

#### Phase 1: Foundation & Workflows
- Edit lifecycle management commands:
  - `publish edit create` - Create new edits
  - `publish edit list` - List all edits
  - `publish edit get` - Get edit details
  - `publish edit commit` - Commit edits
  - `publish edit validate` - Validate edits
  - `publish edit delete` - Delete edits
- `--edit-id` flag support for all publish commands
- `--no-auto-commit` flag for all publish commands
- Error search & reporting commands:
  - `vitals errors issues search` - Search error issues
  - `vitals errors reports search` - Search error reports
  - `errors counts get` - Get error counts
  - `errors counts query` - Query error counts
- Deobfuscation file upload: `publish deobfuscation upload`

#### Phase 2: Monetization
- Full subscriptions CRUD operations:
  - `monetization subscriptions create` - Create subscriptions
  - `monetization subscriptions update` - Update subscriptions
  - `monetization subscriptions patch` - Patch subscriptions
  - `monetization subscriptions delete` - Delete subscriptions
  - `monetization subscriptions archive` - Archive subscriptions
- Batch operations for subscriptions:
  - `monetization subscriptions batchGet` - Batch get subscriptions
  - `monetization subscriptions batchUpdate` - Batch update subscriptions
- Base plans management:
  - `monetization base-plans activate` - Activate base plans
  - `monetization base-plans deactivate` - Deactivate base plans
  - `monetization base-plans delete` - Delete base plans
  - `monetization base-plans migrate-prices` - Migrate base plan prices
  - Batch operations for base plans (batchGet, batchUpdate)
- Offers management with full CRUD:
  - Create, update, patch, delete offers
  - Batch operations for offers (batchGet, batchUpdate)
- One-time products commands for managing one-time purchases
- Regional pricing conversion: `monetization convert-region-prices`

#### Phase 3: Purchase Management
- Voided purchases listing: `purchases voided list`
- Product purchase actions:
  - `purchases products acknowledge` - Acknowledge product purchases
  - `purchases products consume` - Consume product purchases
- Subscription purchase actions:
  - `purchases subscriptions acknowledge` - Acknowledge subscription purchases
  - `purchases subscriptions cancel` - Cancel subscriptions
  - `purchases subscriptions defer` - Defer subscription renewals
  - `purchases subscriptions refund` - Refund subscriptions
  - `purchases subscriptions revoke` - Revoke subscriptions

#### Phase 4: Advanced Vitals
- Additional vitals metrics:
  - Excessive wakeups tracking
  - LMK (Low Memory Killer) rate monitoring
  - Slow rendering detection
  - Slow start detection
  - Stuck wakelocks monitoring
- Anomalies detection: `vitals anomalies list` - List detected anomalies

#### Phase 5: Publishing Enhancements
- Images API commands:
  - `publish images upload` - Upload app images
  - `publish images list` - List uploaded images
  - `publish images delete` - Delete specific images
  - `publish images deleteall` - Delete all images
- Internal app sharing upload: `publish internalappsharing upload`
- App details management:
  - `publish app-details get` - Get app details
  - `publish app-details update` - Update app details
  - `publish app-details patch` - Patch app details

#### Phase 6: Access Control
- Users management:
  - `permissions users create` - Create users
  - `permissions users list` - List users
  - `permissions users get` - Get user details
  - `permissions users patch` - Update users
  - `permissions users delete` - Delete users
- Grants management:
  - `permissions grants create` - Create grants
  - `permissions grants patch` - Update grants
  - `permissions grants delete` - Delete grants
- Permission validation: `permissions validate`
- Permission listing: `permissions list`

#### Phase 7: Optional Features
- App recovery commands:
  - `recovery create` - Create recovery actions
  - `recovery list` - List recovery actions
  - `recovery deploy` - Deploy recovery actions
  - `recovery cancel` - Cancel recovery actions
  - `recovery add-targeting` - Add targeting to recovery actions
- Games management:
  - Achievements management commands
  - Scores management commands
  - Events management commands
  - Players reset and hide commands

## [0.1.0] - 2024-01-23

### Added
- Initial release of gpd (Google Play Developer CLI)
- Authentication via service account key files
- Publishing commands for app management
- Reviews and ratings commands
- Purchases and subscriptions commands
- Analytics commands
- Android vitals commands
- Monetization commands
- Multiple output formats: JSON, table, markdown
- Shell completion support for bash, zsh, and fish
- Cross-platform support: Linux, macOS, Windows
- Docker image support
- Homebrew formula for macOS/Linux

[Unreleased]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/releases/tag/v0.1.0
