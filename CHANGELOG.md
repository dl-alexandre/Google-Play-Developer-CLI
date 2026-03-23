# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.6.4] - 2026-03-23

### Changes

- fix(ci): update Go version from 1.24.x to 1.25.x in all workflows

## [0.6.3] - 2026-03-15

### Changes

- feat(release): add GoReleaser with homebrew and scoop distribution
- chore(release): prepare v0.5.8

## [0.5.8] - 2026-03-14

### Changes

- fix(auth): status command now loads existing credentials

## [0.5.7] - 2026-03-12

### Changes

- test: add Windows CI skip to TestDetectExtensionType and TestStateManager
- Fix Windows CI failures - skip Unix-specific tests on Windows CI
- test: add Windows CI skip to watcher tests
- fix(workflow): close stopChan on timeout to prevent goroutine leak
- test: skip integration tests on Windows CI
- fix(workflow): resolve goroutine leaks and Windows process cleanup
- test: skip watcher tests on Windows CI
- Revert "ci: allow Windows tests to have non-blocking failures"
- ci: allow Windows tests to have non-blocking failures
- fix(workflow): resolve data race in watcher tests
- test: skip network-dependent tests on Windows CI
- ci: trigger rebuild to check Windows workflow tests
- fix: resolve errcheck linter errors in extension system
- feat(extensions): implement gh-style extension system for gpd
- feat(workflow): add declarative workflow execution system

## [0.5.6] - 2026-03-10

### Changes

## [0.5.5] - 2026-03-10

### Changes

- chore: reorganize top-level files
- chore: clean up duplicate files and test binaries

## [0.5.4] - 2026-03-10

### Changes

- fix: use correct secret name TAP_GITHUB_TOKEN for homebrew tap

## [0.5.3] - 2026-03-10

### Changes

- ci: exclude internal/apidrift from gosec linting
- fix: exclude G122 from gosec and remove nolint directives
- ci: add gosec config and make security scan tolerant
- fix: inline nolint:gosec directives to suppress unused warning
- ci: fix workflow to tolerate CLI test failures
- fix: correct nolint:gosec directives for path traversal
- feat: Complete 100% Google Play Developer API coverage with optimized builds

## [0.5.0] - 2026-03-01

### Added

#### Google Play API v2 Features
- **Edit Commit Review Behavior**: Added `inProgressReviewBehaviour` parameter to edit commit operations
  - Supports `THROW_ERROR_IF_IN_PROGRESS` - fails if changes already in review
  - Supports `CANCEL_IN_PROGRESS_AND_SUBMIT` - cancels in-progress review and submits new changes
  - Available on: `gpd publish upload`, `gpd publish release`, `gpd bulk upload`
  - Useful for CI/CD pipelines that need to force-cancel stuck reviews

### Changed
- Updated dependencies to maintain compatibility with Go 1.24

## [0.4.8] - 2026-02-24

### Added

#### Comprehensive Testing Infrastructure
- **Build Tags**: Added `//go:build unit` and `//go:build integration` tags for test categorization
  - 29 test files now tagged for selective test execution
  - `make test` runs only unit tests (5x faster than full suite)
  - New Makefile targets: `test-unit`, `test-integration`, `test-e2e`, `test-flaky`
  
- **Benchmark Regression Detection**:
  - New `cmd/benchcheck/` CLI tool for comparing benchmark results
  - Statistical analysis with Welch's t-test for significance testing
  - Three-tier severity: NOTICE (10%), WARNING (10-20%), CRITICAL (>20%)
  - Multiple output formats: text, json, GitHub Actions annotations
  - New GitHub Actions workflow: `benchmark-regression.yml`
  - Makefile targets: `benchmark`, `benchmark-compare`, `benchmark-regression`, `benchmark-baseline`

- **Golden File Testing**:
  - New `internal/testutil/golden.go` package for snapshot testing
  - Compare actual output against saved golden files
  - Update with `-update` flag: `go test -update ./...`
  - Pretty diff output for debugging mismatches

- **Fuzz Testing**:
  - `FuzzConfigJSON`: Config JSON parsing and validation
  - `FuzzPackageName`: Package name validation  
  - `FuzzAPIError`: Error creation and method chaining
  - `FuzzParseDirectory`: Fastlane metadata parsing
  - `FuzzNewResult`: Result creation with various data types
  - `FuzzParseFormat`: Output format parsing

- **Coverage & CI**:
  - Coverage threshold set to 40%
  - New `.github/workflows/coverage.yml` with detailed reporting
  - PR comments with coverage summary
  - Makefile: `test-coverage-threshold` target

### Changed
- Enhanced Makefile with comprehensive test categories
- CI workflow now runs tests with race detector and coverage
- All test files use `t.Parallel()` for concurrent execution

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

[Unreleased]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.6.4...HEAD
[0.6.4]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.6.3...v0.6.4
[0.6.3]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.6.2...v0.6.3
[0.5.8]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.5.7...v0.5.8
[0.5.7]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.5.6...v0.5.7
[0.5.6]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.5.5...v0.5.6
[0.5.5]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.5.4...v0.5.5
[0.5.4]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.5.3...v0.5.4
[0.5.3]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.5.2...v0.5.3
[0.4.8]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.4.7...v0.4.8
[0.4.7]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.4.6...v0.4.7
[0.4.6]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.4.5...v0.4.6
[0.4.5]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.2.0...v0.4.5
[0.2.0]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/dl-alexandre/Google-Play-Developer-CLI/releases/tag/v0.1.0
