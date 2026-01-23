# Requirements Document

## Introduction

The Google Play Developer CLI is a fast, lightweight command-line interface that serves as the Google Play equivalent to the App Store Connect CLI. It provides developers, CI/CD systems, and AI agents with programmatic access to Google Play Developer Console functionality for automating Android app publishing and management tasks.

## Glossary

- **CLI**: Command Line Interface - the primary interface for user interaction (binary name: gpd)
- **Android_Publisher_API**: Google Play Android Publisher API v3 - primary API for app publishing, in-app products, and reviews
- **Play_Developer_Reporting_API**: Google Play Developer Reporting API - provides analytics and crash data  
- **Service_Account**: Google Service Account - authentication mechanism for API access using JSON key files
- **Edit_Transaction**: Atomic transaction for making changes to app listings and releases
- **Track**: Release track (internal, alpha, beta, production) for app distribution
- **AAB**: Android App Bundle - the preferred publishing format for Android apps
- **APK**: Android Package - legacy app package format
- **Store_Listing**: App metadata including descriptions, screenshots, and localization
- **In_App_Product**: Subscription or one-time purchase item within an app
- **Review_Response**: Developer response to user reviews
- **ADC**: Application Default Credentials - Google's standard credential discovery mechanism
- **Access_Token**: Short-lived OAuth 2.0 token derived from service account key

## API-to-Namespace Mapping

| Namespace | API Endpoint Family | Primary Operations |
|-----------|-------------------|-------------------|
| publish/* | Android Publisher API v3 - Edits | Upload, release, rollout, promote, rollback |
| reviews/* | Android Publisher API v3 - Reviews | List, reply to user reviews |
| vitals/* | Play Developer Reporting API | Crash rates, ANR rates, performance metrics |
| analytics/* | Play Developer Reporting API | Install statistics, ratings data |
| monetization/* | Android Publisher API v3 - InAppProducts | Basic in-app products (managed/consumable), read-only subscriptions |
| purchases/* | Android Publisher API v3 - Purchases | Purchase token verification |
| auth/* | OAuth 2.0 / Service Account | Authentication, token management |
| config/* | Local configuration | Settings, credentials, validation |

## Requirements

### Requirement 1: API Surface and Command Namespaces

**User Story:** As a developer, I want to understand exactly which Google Play APIs are supported and how commands are organized, so that I can plan my automation workflows accordingly.

#### Acceptance Criteria

1. THE gpd SHALL integrate with Android Publisher API v3 for app publishing using the "edits" workflow
2. THE gpd SHALL integrate with Play Developer Reporting API for analytics and Android vitals data
3. THE gpd SHALL integrate with Android Publisher API v3 for in-app products (managed/consumable) and read-only subscriptions
4. THE gpd SHALL integrate with Android Publisher API v3 for reviews and developer responses
5. THE gpd SHALL organize commands into namespaces: publish/*, reviews/*, vitals/*, analytics/*, monetization/*, purchases/*, auth/*, config/*
6. THE gpd SHALL validate service account permissions for required API scopes before operations
7. THE gpd SHALL NOT support Play Console UI scraping or unofficial API endpoints

### Requirement 2: Edit Transaction Management

**User Story:** As a developer, I want atomic edit transactions for publishing operations, so that my changes are applied consistently or not at all.

#### Acceptance Criteria

1. THE gpd SHALL use auto-edit mode by default, automatically creating and committing edit transactions
2. WHEN explicit edit mode is enabled, THE gpd SHALL support --edit-id flag with user-provided local handles
3. WHEN explicit edit IDs are used, THE gpd SHALL persist edit mappings in {configDir}/edits/{package}.json with file locking
4. WHEN concurrent processes use the same --edit-id, THE gpd SHALL prevent concurrent writes with file locking
5. WHEN re-running commands, THE gpd SHALL maintain optional SHA256 cache in {cacheDir}/ with 24-hour TTL
6. WHEN duplicate operations are detected, THE gpd SHALL no-op successfully and return meta.noop=true with reason
7. THE gpd SHALL support --dry-run mode to show intended API calls without committing changes

### Requirement 3: Authentication and Token Management

**User Story:** As a developer, I want secure and flexible authentication options, so that I can integrate the CLI into different environments safely.

#### Acceptance Criteria

1. WHEN a service account JSON key is provided, THE gpd SHALL mint short-lived access tokens using OAuth 2.0
2. THE gpd SHALL support Application Default Credentials (ADC) and GOOGLE_APPLICATION_CREDENTIALS environment variable
3. THE gpd SHALL support --store-tokens flag with values: auto (secure when available, never when CI=true or GPD_CI=true), never, secure
4. WHEN --store-tokens=never, THE gpd SHALL use in-memory caching only
5. WHEN --store-tokens=secure, THE gpd SHALL require platform-specific secure storage
6. THE gpd SHALL refresh tokens when less than 300 seconds remain, handling clock skew appropriately
7. THE gpd SHALL never emit interactive OAuth flows or open browsers

### Requirement 4: Platform-Specific Credential Storage

**User Story:** As a developer, I want secure credential storage that works across different operating systems, so that I can use the CLI consistently.

#### Acceptance Criteria

1. ON macOS, THE gpd SHALL use Keychain Services for credential storage
2. ON Linux with Secret Service, THE gpd SHALL use Secret Service API for credential storage  
3. ON Linux without Secret Service, THE gpd SHALL fall back to no-store mode rather than insecure file storage
4. ON Windows, THE gpd SHALL use Windows Credential Manager for credential storage
5. THE gpd SHALL validate service account has required Play Console access before operations
6. THE gpd SHALL provide actionable error messages for missing API enablement or permissions

### Requirement 5: Exit Code Taxonomy and Error Handling

**User Story:** As an AI agent or CI/CD system, I want predictable exit codes for different error types, so that I can handle failures appropriately.

#### Acceptance Criteria

1. WHEN commands succeed, THE gpd SHALL exit with code 0
2. WHEN authentication fails, THE gpd SHALL exit with code 2
3. WHEN permission is denied, THE gpd SHALL exit with code 3
4. WHEN input validation fails, THE gpd SHALL exit with code 4
5. WHEN rate limits are exceeded (HTTP 429 or quota errors), THE gpd SHALL exit with code 5
6. WHEN network errors occur (DNS, TLS, timeouts), THE gpd SHALL exit with code 6
7. WHEN resources are not found, THE gpd SHALL exit with code 7
8. WHEN conflicts occur (edit already exists), THE gpd SHALL exit with code 8
9. WHEN other API errors occur, THE gpd SHALL exit with code 1
10. THE gpd SHALL provide descriptive error messages with suggested solutions for each error type

### Requirement 6: Output Format and Data Structure

**User Story:** As a developer, I want consistent, structured output formats, so that I can reliably parse command results.

#### Acceptance Criteria

1. THE gpd SHALL output minified JSON format by default (single-line) with envelope structure: {data, error, meta}
2. THE gpd SHALL support --pretty flag to enable human-readable JSON formatting
3. THE gpd SHALL guarantee error object schema with fields: code, message, hint, details, httpStatus, retryAfterSeconds, service, operation
4. THE gpd SHALL guarantee meta object schema with fields: noop (boolean), durationMs, services (array), and optional requestId, pageToken, nextPageToken, warnings
5. THE gpd SHALL ensure stdout contains only valid JSON in JSON mode with error envelope {data: null, error: {...}, meta: {...}} on failures
6. THE gpd SHALL support --output flag with values: json, table, markdown (csv only for analytics/vitals exports)
7. THE gpd SHALL support --fields flag using comma-separated dotted paths for JSON projection
8. THE gpd SHALL support --quiet flag to suppress stderr except errors (--quiet wins over --verbose)

### Requirement 7: Pagination and Data Limits

**User Story:** As a developer, I want consistent pagination controls, so that I can handle large datasets efficiently.

#### Acceptance Criteria

1. THE gpd SHALL support --page-size flag for controlling API request batch sizes
2. THE gpd SHALL support --page-token flag for continuing paginated requests
3. THE gpd SHALL support --limit flag for client-side result capping
4. WHEN using table or markdown format, THE gpd SHALL display one page unless --all flag is provided
5. THE gpd SHALL support --no-auto-paginate flag to disable automatic pagination
6. THE gpd SHALL include pagination metadata in JSON output meta field

### Requirement 8: App Publishing Operations

**User Story:** As a developer, I want granular control over app publishing operations, so that I can manage my app distribution lifecycle precisely.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd publish upload` command for uploading artifacts to edit (discovers latest artifacts in edit)
2. THE gpd SHALL support `gpd publish release` command with --track, --name, --status (draft/completed/halted/inProgress), and --version-code flags
3. THE gpd SHALL validate release status transitions: rollout only for production, halt only for inProgress releases
4. THE gpd SHALL support `gpd publish rollout` command with --percentage flag (0.01-100.00 granularity) setting status to inProgress
5. THE gpd SHALL support `gpd publish promote` command copying release notes and artifacts but requiring new rollout percentage
6. THE gpd SHALL support `gpd publish halt` command to halt production rollouts (sets status to halted)
7. THE gpd SHALL support `gpd publish rollback` command with explicit --version-code from same track history
8. THE gpd SHALL require --package flag for all publishing operations
9. WHEN rollback target is ambiguous, THE gpd SHALL require explicit --version-code specification

### Requirement 9: Store Listing and Asset Management

**User Story:** As a developer, I want to manage my app's store listing metadata and assets, so that I can update descriptions, screenshots, and localization without using the web console.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd publish listing` command to update title, short description, and full description
2. THE gpd SHALL support `gpd publish assets` command with directory conventions: assets/{locale}/{category}/
3. THE gpd SHALL support `gpd publish assets spec` command to output machine-readable asset validation matrix
4. THE gpd SHALL perform best-effort local validation of asset dimensions per documented matrix
5. THE gpd SHALL support partial screenshot replacement while maintaining ordering within categories
6. THE gpd SHALL use standard locale codes (en-US format) with normalization from en_US format
7. THE gpd SHALL support default locale fallback for missing localizations
8. THE gpd SHALL support per-locale assets where supported by the Android Publisher API

### Requirement 10: Track and Testing Management

**User Story:** As a developer, I want to manage testing tracks and testers, so that I can control who has access to pre-release versions.

#### Acceptance Criteria

1. THE gpd SHALL validate track names against: internal, alpha, beta, production only
2. THE gpd SHALL NOT support custom tracks or closed testing tracks initially
3. THE gpd SHALL support `gpd publish testers` command with --dry-run showing add/remove diff counts
4. THE gpd SHALL support CSV upload for bulk internal tester management with email validation and deduplication
5. THE gpd SHALL enforce maximum 200 internal testers per Google Play limits
6. THE gpd SHALL NOT support Google Groups integration or organization-managed testers initially

### Requirement 11: In-App Product Catalog Management

**User Story:** As a developer, I want to manage in-app product catalogs, so that I can configure monetization without manual console access.

#### Acceptance Criteria

1. THE gpd SHALL support monetization products command for basic in-app products (managed/consumable only)
2. THE gpd SHALL support basic subscription products with read-only access to existing subscriptions via Android Publisher API v3
3. THE gpd SHALL NOT support creating or modifying subscription products initially
4. THE gpd SHALL NOT support modern base plans/offers, regional pricing, or introductory pricing initially
5. THE gpd SHALL validate product configurations against Google Play requirements before submission
6. THE gpd SHALL separate catalog management from purchase verification operations

### Requirement 12: Purchase State Verification

**User Story:** As a developer, I want to verify purchase tokens and subscription states, so that I can validate user entitlements programmatically.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd purchases verify` command with --environment (sandbox/production/auto) and --package flags
2. THE gpd SHALL use Purchases.products:get for one-time product verification
3. THE gpd SHALL use Purchases.subscriptionsv2:get for subscription state verification (v2 API)
4. THE gpd SHALL treat Purchases.subscriptions:get as legacy/deprecated
5. THE gpd SHALL implement retry policy with maximum 3 retries and exponential backoff for transient errors
6. THE gpd SHALL return purchase state, validity, consumption status, and acknowledgement state in structured format
7. THE gpd SHALL require different service account permissions for purchase verification vs catalog management

### Requirement 13: Review Management and Response

**User Story:** As a developer, I want to read user reviews and respond to them safely, so that I can engage with my users while avoiding bulk response issues.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd reviews list` command with filtering by rating, date range (ISO 8601), and language
2. THE gpd SHALL support `gpd reviews reply` command with per-process rate limiting (default: 1 per 5 seconds, applies to reply only)
3. THE gpd SHALL implement --max-actions flag with default limit of 10 responses per execution
4. THE gpd SHALL implement configurable --rate-limit flag for review operations
5. THE gpd SHALL support --template-file with variables {{appName}}, {{rating}}, {{locale}} and deterministic failure for missing variables
6. THE gpd SHALL provide preview plan output in --dry-run mode showing intended responses
7. THE gpd SHALL distinguish between creating new responses and updating existing ones

### Requirement 14: Analytics and Reporting

**User Story:** As a developer, I want to access app performance data and crash reports, so that I can monitor my app's health and user engagement.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd analytics` and `gpd vitals` commands using Play Developer Reporting API
2. THE gpd SHALL separate console analytics (installs, ratings) from Android vitals (crashes, ANRs)
3. THE gpd SHALL use ISO 8601 format for date ranges with timezone conversion to UTC
4. THE gpd SHALL support JSON and CSV export formats with consistent field naming
5. THE gpd SHALL document data freshness as best-effort with typical 24-48 hour delays
6. THE gpd SHALL provide `gpd analytics capabilities` and `gpd vitals capabilities` commands with supported metrics, dimensions, granularities, and maximum lookback

### Requirement 15: Performance and Resource Management

**User Story:** As a developer, I want fast command execution with minimal resource usage, so that the CLI integrates efficiently into my development workflow.

#### Acceptance Criteria

1. THE gpd SHALL achieve cold start under 200ms plus network time for simple commands (excluding first install and first token mint)
2. THE gpd SHALL add less than 100ms command overhead excluding network operations and token minting (measured per OS)
3. THE gpd SHALL use resumable uploads with retry and exponential backoff for large files
4. THE gpd SHALL display upload progress to stderr or disable progress in JSON mode
5. THE gpd SHALL implement token caching to avoid repeated authentication overhead
6. THE gpd SHALL respect API rate limits with appropriate concurrency controls
7. THE gpd SHALL consume minimal memory during normal operations (under 50MB)

### Requirement 16: Configuration and Resource Addressing

**User Story:** As a developer, I want consistent configuration and resource addressing, so that I can use the CLI predictably across different projects.

#### Acceptance Criteria

1. THE gpd SHALL support configuration files in JSON or YAML format in standard locations
2. THE gpd SHALL follow precedence order: command flags > environment variables > configuration file > defaults
3. THE gpd SHALL require --package flag or global package configuration for all app-specific operations
4. THE gpd SHALL validate track names against allowed values and provide clear error messages
5. THE gpd SHALL normalize locale codes from en_US to en-US format automatically
6. THE gpd SHALL use ISO 8601 format for all date range specifications
7. THE gpd SHALL never store service account keys in configuration files by default

### Requirement 17: Security and Data Protection

**User Story:** As a developer, I want secure handling of sensitive data, so that my credentials and user data remain protected.

#### Acceptance Criteria

1. THE gpd SHALL redact PII (usernames, orderIds, purchaseTokens) at structured field level in logs only
2. THE gpd SHALL exclude review text from stdout data by default for reviews list command
3. THE gpd SHALL support --include-review-text flag for explicit review text inclusion in stdout data
4. THE gpd SHALL redact access tokens and service account keys from all log output
5. THE gpd SHALL use HTTPS for all API communications with certificate validation
6. THE gpd SHALL validate service account permissions and provide actionable error messages
7. THE gpd SHALL support credential rotation without data loss or service interruption
8. THE gpd SHALL suppress sensitive information in --verbose mode according to defined redaction rules

### Requirement 18: Platform Compatibility and Distribution

**User Story:** As a developer, I want the CLI to work consistently across different platforms, so that I can use it in various development environments.

#### Acceptance Criteria

1. THE gpd SHALL support macOS, Linux, and Windows operating systems
2. THE gpd SHALL require no external dependencies (JDK, Gradle, or other runtime tools)
3. THE gpd SHALL support minimum Android Publisher API v3 and Play Developer Reporting API versions
4. THE gpd SHALL provide single-binary distribution for each supported platform
5. THE gpd SHALL publish checksums and signatures for releases with `gpd --version --json` including build hash
6. THE gpd SHALL support manual replacement or third-party distribution channels for updates
7. THE gpd SHALL provide unambiguous `gpd --version` output with project name and build identifier
8. THE gpd SHALL detect multiple gpd binaries in PATH via `gpd config doctor` and report their locations

### Requirement 20: AI Agent Ergonomics and Quickstart

**User Story:** As an AI agent, I want optimized defaults and clear guidance for automated workflows, so that I can reliably interact with the CLI without human intervention.

#### Acceptance Criteria

1. THE gpd SHALL output minified JSON by default (single-line) for predictable machine parsing
2. THE gpd SHALL support --pretty flag to enable human-readable JSON formatting
3. THE gpd SHALL support `gpd help agent` command with quickstart documentation for AI agents
4. THE gpd SHALL support --paginate/--all flag to automatically drain all pages for agent workflows
5. THE gpd SHALL make "no next page" state unambiguous by clearing meta.nextPageToken when pagination is complete
6. THE gpd SHALL support --next flag as friendly alias for --page-token for manual pagination
7. THE gpd SHALL support --sort flag with -field syntax for descending order and per-command validation

### Requirement 21: Network and Timeout Configuration

**User Story:** As a developer, I want configurable network timeouts and download controls, so that I can tune the CLI for different environments and use cases.

#### Acceptance Criteria

1. THE gpd SHALL support GPD_TIMEOUT environment variable and --timeout flag for network operations
2. THE gpd SHALL provide output path control for file downloads
3. THE gpd SHALL handle long-running report/list commands with appropriate timeout defaults
4. THE gpd SHALL support connection retry with exponential backoff for transient network failures

### Requirement 22: Authentication User Experience

**User Story:** As a developer, I want simple authentication status management, so that I can validate and manage credentials without performing real operations.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd auth status` command to check current authentication state
2. THE gpd SHALL support `gpd auth logout` command to clear stored credentials
3. THE gpd SHALL support credential fallback hierarchy: secure storage > local config > environment variables
4. THE gpd SHALL support GPD_SERVICE_ACCOUNT_KEY environment variable for service account JSON
5. THE gpd SHALL support GPD_PACKAGE environment variable for default package identifier
6. THE gpd SHALL validate authentication state without performing API operations when possible

### Requirement 23: Safety Controls for Destructive Operations

**User Story:** As a developer, I want explicit confirmation for destructive operations, so that I can prevent accidental changes while maintaining non-interactive operation.

#### Acceptance Criteria

1. THE gpd SHALL require --confirm flag for destructive operations: rollback, halt, delete operations
2. THE gpd SHALL support --yes flag as alias for --confirm for scripting convenience
3. THE gpd SHALL remain non-interactive and never prompt for user input
4. THE gpd SHALL provide clear error messages when confirmation flags are missing
5. THE gpd SHALL document which operations require confirmation flags in help output

### Requirement 24: Command Interface Standards

**User Story:** As a developer, I want consistent and descriptive command interfaces, so that I can use the CLI predictably across different operations.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd version` command alongside --version flag for scripting convenience
2. THE gpd SHALL prefer long, descriptive flags over terse abbreviations
3. THE gpd SHALL provide examples in help output for common operations
4. THE gpd SHALL use consistent flag naming patterns across all commands
5. THE gpd SHALL validate sort keys per command and emit clear errors for unsupported keys
6. THE gpd SHALL support --output flag with values: json, table, markdown (csv only for analytics/vitals exports)

### Requirement 19: Capabilities Discovery and System Health

**User Story:** As a developer, I want to discover CLI capabilities and validate my setup, so that I can troubleshoot issues and understand what operations are available.

#### Acceptance Criteria

1. THE gpd SHALL support `gpd auth check` command to validate service account permissions
2. THE gpd SHALL support `gpd config doctor` command to diagnose configuration and credential issues
3. THE gpd SHALL support `gpd publish capabilities` command to list available publishing operations
4. THE gpd SHALL support `gpd analytics capabilities` and `gpd vitals capabilities` commands for reporting features
5. THE gpd SHALL provide structured JSON output for all capability discovery commands
6. THE gpd SHALL NOT collect or transmit telemetry data

### Requirement 25: Edit Transaction and Command Integration

**User Story:** As a developer, I want clear rules about which commands require explicit edits, so that I can plan my publishing workflows appropriately.

#### Acceptance Criteria

1. THE gpd SHALL run `gpd publish upload` in auto-edit mode by default (uploads artifacts only)
2. THE gpd SHALL run `gpd publish release` in auto-edit mode by default (creates release from uploaded artifacts)
3. THE gpd SHALL support explicit --edit-id for `gpd publish upload` and `gpd publish release` commands
4. THE gpd SHALL support --release-notes-file flag with JSON format: {"locale": "text"} for localized release notes
5. THE gpd SHALL support --version-code flag as repeatable for multi-APK releases
6. THE gpd SHALL copy release notes and artifacts during promote but require explicit --percentage for rollout

### Requirement 26: Configuration and Concurrency Management

**User Story:** As a developer, I want predictable configuration locations and concurrency behavior, so that the CLI works consistently across environments.

#### Acceptance Criteria

1. THE gpd SHALL use OS-appropriate directories: {configDir} = ~/.config/gpd (Linux), ~/Library/Application Support/gpd (macOS), %APPDATA%/gpd (Windows)
2. THE gpd SHALL use OS-appropriate cache directories: {cacheDir} = ~/.cache/gpd (Linux), ~/Library/Caches/gpd (macOS), %LOCALAPPDATA%/gpd (Windows)
3. THE gpd SHALL treat ~/.gpd as compatibility alias for config directory
4. THE gpd SHALL default to maximum 3 parallel API calls for list operations
5. THE gpd SHALL use single-threaded uploads for reliability
6. THE gpd SHALL use file locking with 30-second timeout for edit mappings, exiting with code 8 on contention

### Requirement 27: Asset Management and Tester Constraints

**User Story:** As a developer, I want clear asset organization and tester management rules, so that I can manage app content predictably.

#### Acceptance Criteria

1. THE gpd SHALL determine asset ordering by lexicographic filename sorting within categories
2. THE gpd SHALL support --replace flag with category values (phone, tablet, tv, wear) for partial replacement
3. THE gpd SHALL handle per-locale assets by directory structure when API supports locale-specific assets
4. WHEN internal tester limit (200) is exceeded, THE gpd SHALL return hard error with current count
5. THE gpd SHALL deduplicate email addresses before counting against limits
6. THE gpd SHALL define "auto" environment detection as: call API and interpret response (no token prefix rules)

### Requirement 28: Review Template and Concurrency Safety

**User Story:** As a developer, I want safe review response templating, so that I can respond to users without exposing sensitive information.

#### Acceptance Criteria

1. THE gpd SHALL support template variables: {{appName}}, {{rating}}, {{locale}} only (no review text access)
2. THE gpd SHALL use proper escaping for template variables to prevent injection
3. THE gpd SHALL return meta.action field with values: created, updated, skipped for each review response
4. THE gpd SHALL apply rate limiting per-process (not per-package) for review operations
5. THE gpd SHALL apply rate limiting to both list and reply endpoints

### Requirement 29: Global Options and Configuration Management

**User Story:** As a developer, I want consistent global options and configuration management, so that I can use the CLI predictably across all commands.

#### Acceptance Criteria

1. THE gpd SHALL support global flags: --package, --output, --pretty, --timeout, --store-tokens, --fields, --quiet, --verbose
2. THE gpd SHALL support shared list flags: --all, --paginate, --page-size, --page-token, --next, --sort, --limit
3. THE gpd SHALL support `gpd init` command to scaffold config, sample release-notes.json, assets/ layout, and .gitignore
4. THE gpd SHALL support `gpd config path` command to show configuration file locations
5. THE gpd SHALL support `gpd config get/set` commands for configuration management
6. THE gpd SHALL support `gpd config print --resolved` command showing precedence resolution
7. THE gpd SHALL support shell completion generation for bash, zsh, and fish
8. THE gpd SHALL define JSON envelope schemas: collections use {data: [], error, meta}, singletons use {data: {}, error, meta}

### Requirement 30: Command Reordering and Cleanup

**User Story:** As a developer, I want logically ordered requirements, so that the specification is easy to follow and implement.

#### Acceptance Criteria

1. THE gpd SHALL implement all requirements in monotonic order (1-30)
2. THE gpd SHALL remove duplicate content and inconsistent numbering
3. THE gpd SHALL maintain requirement traceability through implementation
4. THE gpd SHALL defer download-style commands and --decompress flag until specific commands are defined
5. THE gpd SHALL focus initial scope on core publishing, reviews, analytics, and monetization operations

The following features are explicitly excluded from the initial scope:
The following features are explicitly excluded from the initial scope:

- **Play Console UI Parity**: Not all Play Console features will be implemented
- **Device Automation**: No device testing orchestration or device management
- **Custom Testing Tracks**: Only standard tracks (internal, alpha, beta, production) supported
- **Legacy API Support**: Focus on current API versions only
- **Proprietary Endpoints**: No undocumented or unofficial API usage
- **Advanced Monetization**: No modern base plans/offers, regional pricing, legacy subscriptions, or complex pricing models initially (requires different API integration)
- **Google Groups Integration**: Tester management limited to direct email lists
- **Real-time Analytics**: Analytics data subject to Google's standard reporting delays
- **Telemetry**: No data collection or transmission of usage metrics
## Non-Goals

The following features are explicitly excluded from the initial scope:

- **Play Console UI Parity**: Not all Play Console features will be implemented
- **Device Automation**: No device testing orchestration or device management
- **Interactive Flows**: No interactive prompts, TUI interfaces, or browser-based OAuth
- **Custom Testing Tracks**: Only standard tracks (internal, alpha, beta, production) supported
- **Legacy API Support**: Focus on current API versions only
- **Proprietary Endpoints**: No undocumented or unofficial API usage
- **Advanced Monetization**: No modern base plans/offers, regional pricing, legacy subscriptions, or complex pricing models initially (requires different API integration)
- **Google Groups Integration**: Tester management limited to direct email lists
- **Real-time Analytics**: Analytics data subject to Google's standard reporting delays
- **Telemetry**: No data collection or transmission of usage metrics
- **Download Commands**: File download operations deferred until specific use cases are defined