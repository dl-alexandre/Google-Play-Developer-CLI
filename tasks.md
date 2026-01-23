# Implementation Plan: Google Play Developer CLI

## Overview

This implementation plan follows a vertical slice approach, building a complete but minimal foundation first, then expanding surface area. The plan prioritizes contract stability and early validation over feature breadth, ensuring the framework is proven before implementing many commands.

## Phase 0: Specification Lock (Mandatory)

- [x] 0.1 Finalize specification inconsistencies ✅
  - [x] Tester limit default standardized to 200 (internal/config/config.go:52)
  - [x] Meta schema fields locked down (internal/output/result.go:34-40):
    - partial, scannedCount, filteredCount, totalAvailable, retries, dataFreshnessUtc, noOpReason
  - [x] CI detection with specific environment variables (internal/config/config.go:230-240):
    - CI, GITHUB_ACTIONS, JENKINS_URL, BUILDKITE, CIRCLECI, TRAVIS, GITLAB_CI, GPD_CI
  - [x] OAuth scope mapping complete (internal/auth/auth.go:17-27):
    - ScopeAndroidPublisher: publish, reviews, monetization, purchases
    - ScopePlayReporting: analytics, vitals
  - _Requirements: All specification consistency_

## Phase 1: Foundation and Contracts

- [x] 1. Project Setup and Core Infrastructure ✅
  - Initialize Go module with proper structure (cmd/, internal/, pkg/)
  - Set up build system with cross-platform binary generation
  - Configure CI/CD pipeline with automated testing
  - _Requirements: 18.1, 18.4_

- [x] 2. Core CLI Framework with Command Registry ✅ (implementation only, tests pending)
  - [x] 2.1 Implement command framework with namespace organization
    - Command registry with namespace validation (publish/*, reviews/*, etc.)
    - Command interface with validation returning *Result
    - JSON envelope structure with comprehensive metadata schema
    - Exit code taxonomy and error mapping
    - _Requirements: 1.5, 5.1-5.10, 6.1-6.4_

  - [x] 2.2 Implement output formatting system
    - JSON (minified by default), table, markdown formats
    - Field projection with --fields flag
    - Pretty printing with --pretty flag
    - _Requirements: 6.1, 6.2, 6.6_

  - [x] 2.3 Write core property tests (non-optional) ✅
    - **Property 1: Command Namespace Organization**
    - **Property 13: JSON Output Format Consistency**
    - **Property 15: Error Schema Guarantee**
    - **Property 16: Meta Schema Guarantee**
    - **Validates: Requirements 1.5, 6.1-6.4**

- [x] 3. Mock Transport Harness and Test Infrastructure ✅
  - [x] 3.1 Implement test harness with mocked HTTP transport
    - Golden request/response fixtures loader
    - Request recorder for contract validation
    - Deterministic RNG for property tests
    - _Requirements: Testing infrastructure_

  - [x] 3.2 Create initial golden fixtures
    - Auth check API responses
    - Basic edits API responses
    - Error response examples
    - _Requirements: Testing infrastructure_

- [x] 4. Credentials Subsystem (Unified) ✅ (implementation only, tests pending)
  - [x] 4.1 Implement complete authentication system
    - Configuration system with OS-appropriate directories and precedence rules
    - Credential origin selection (ADC, keyfile, environment)
    - OAuth 2.0 TokenSource creation with early refresh wrapper (300s leeway)
    - Platform-specific secure storage integration
    - CI environment detection for --store-tokens=auto
    - _Requirements: 3.1-3.3, 4.1-4.4, 16.1-16.2_

  - [x] 4.2 Write authentication property tests (non-optional) ✅
    - **Property 9: Service Account Authentication**
    - **Property 10: Credential Discovery Support**
    - **Property 11: Token Storage Flag Validation**
    - **Validates: Requirements 3.1-3.3**

- [x] 5. First Vertical Slice - Diagnostic Commands ✅ (implementation only, tests pending)
  - [x] 5.1 Implement auth check command
    - Per-surface permission validation with minimal API calls
    - Structured capability reporting with actionable hints
    - OAuth scope validation per command namespace
    - _Requirements: 1.6, 19.1_

  - [x] 5.2 Implement config doctor command
    - Configuration validation and diagnosis
    - Credential health checking
    - Basic system health validation
    - _Requirements: 19.2_

  - [x] 5.3 Write diagnostic command property tests (non-optional) ✅
    - **Property 2: Service Account Permission Validation**
    - **Property 28: Auth Check Command Functionality**
    - **Property 29: Config Doctor Command Functionality**
    - **Validates: Requirements 1.6, 19.1, 19.2**

- [x] 6. Checkpoint - Validate Core Framework ✅
  - Ensure all core property tests pass
  - Validate auth check and config doctor work end-to-end
  - Confirm output envelope and error handling contracts

## Phase 2: Stateful Operations Foundation

- [x] 7. Edit Transaction Management ✅ (implementation only, tests pending)
  - [x] 7.1 Implement edit manager with file locking
    - Auto-edit and explicit edit modes
    - Cross-platform file locking with conservative stale lock policy (PID/hostname checks)
    - SHA256-based artifact caching with TTL
    - _Requirements: 2.1-2.6_

  - [x] 7.2 Write edit transaction property tests (non-optional) ✅
    - **Property 3: Auto-Edit Mode Consistency**
    - **Property 5: Edit Mapping Persistence**
    - **Property 6: Concurrent Edit Protection**
    - **Property 8: Idempotent Operation Detection**
    - **Validates: Requirements 2.1-2.6**

- [x] 8. API Client Infrastructure ✅
  - [x] 8.1 Implement unified API client with retry logic
    - Android Publisher API v3 client
    - Play Developer Reporting API client
    - Exponential backoff retry with metadata tracking
    - _Requirements: 1.1, 1.2_

## Phase 3: First Real Play Operation

- [x] 9. Basic Publishing Commands ✅ (implementation only, tests pending)
  - [x] 9.1 Implement publish upload command (basic)
    - AAB/APK file upload without resumable uploads initially
    - Artifact validation and SHA256 caching
    - _Requirements: 8.1_

  - [x] 9.2 Implement publish release command (minimal)
    - Release creation with track, status, version code validation
    - Basic release notes support
    - _Requirements: 8.2, 8.3_

  - [x] 9.3 Write publishing property tests (non-optional) ✅
    - **Property 17: Artifact Upload Functionality**
    - **Property 18: Release Command Flag Validation**
    - **Property 19: Release Status Transition Validation**
    - **Validates: Requirements 8.1-8.3**

- [x] 10. Checkpoint - Validate Publishing Foundation ✅
  - Ensure publishing commands work with mocked APIs
  - Validate edit transaction behavior
  - Confirm idempotency and caching work correctly

## Phase 4: Expand Surface Area

- [x] 11. Review Management System ✅
  - [x] 11.1 Implement reviews list command
    - Client-side filtering with scan limits and partial results metadata
    - Translation language support (server-side)
    - Review text inclusion controls with PII protection
    - _Requirements: 13.1, 17.2, 17.3_

  - [x] 11.2 Implement reviews reply command
    - Rate limiting with configurable delays
    - Template variable substitution with safety checks
    - Idempotency based on reviewId + text hash
    - _Requirements: 13.2-13.7_

  - [x] 11.3 Write review operation property tests ✅
    - **Property 22: Review List Command Client-Side Filtering**
    - **Property 23: Review Reply Rate Limiting**
    - **Property 24: Template Variable Processing**
    - **Validates: Requirements 13.1, 13.2, 13.5**

- [x] 12. Purchase Verification ✅
  - [x] 12.1 Implement purchase verification commands
    - Products API integration for one-time purchases
    - Subscriptions v2 API for subscription verification (explicit routing)
    - Environment detection with API-based auto detection
    - _Requirements: 12.1-12.7_

  - [x] 12.2 Write purchase verification property tests ✅
    - **Property 20: Purchase Verification Command Support**
    - **Property 21: Purchase API Integration**
    - **Validates: Requirements 12.1-12.3**

- [x] 13. Analytics and Reporting ✅
  - [x] 13.1 Define capabilities discovery output format
    - Data freshness indicators
    - Supported metrics and dimensions
    - Availability gap reporting
    - _Requirements: 14.6_

  - [x] 13.2 Implement analytics and vitals commands
    - Play Developer Reporting API integration
    - Date range handling with timezone conversion
    - CSV export support for analytics data
    - _Requirements: 14.1-14.6_

- [x] 14. Store Listing and Asset Management ✅
  - [x] 14.1 Define asset validation constraints table
    - Authoritative dimension requirements per asset type
    - Reference source for validation rules
    - _Requirements: 9.4_

  - [x] 14.2 Implement store listing commands
    - Title, description, and localization management
    - Locale normalization (en_US to en-US)
    - _Requirements: 9.1, 9.6, 9.7_

  - [x] 14.3 Implement asset management commands
    - Screenshot, feature graphic, and icon upload
    - Asset validation with dimension checking
    - Directory convention support
    - _Requirements: 9.2-9.5, 9.8_

- [x] 15. Monetization and Tester Management ✅
  - [x] 15.1 Implement monetization commands
    - In-app product management (managed/consumable)
    - Read-only subscription access
    - Product validation against Play requirements
    - _Requirements: 11.1-11.6_

  - [x] 15.2 Implement tester management commands
    - Google Groups tester management (API supports groups, not individual emails)
    - Add/remove groups, replace all groups
    - Note: Individual email testers managed via Play Console UI
    - _Requirements: 10.1-10.6_

- [x] 16. Advanced Publishing Features ✅
  - [x] 16.1 Implement rollout management commands
    - Implemented promote, complete, halt, rollout, tracks, status commands
    - Rollout percentage configuration with validation (0.01-100%)
    - Promote between tracks with artifact/release notes copying
    - Halt and rollback via halt + promote workflow
    - Production operations require --confirm flag
    - _Requirements: 8.4-8.8_

  - [x] 16.2 Confirm resumable upload support for artifact types
    - Google API client library supports resumable uploads by default for large files
    - Added UploadOptions with ChunkSize (default 8MB) and ProgressFunc callback
    - Added UploadBundleWithOptions, UploadAPKWithOptions, UploadArtifactWithOptions
    - Progress indicators supported via googleapi.ProgressUpdater
    - _Requirements: 15.3_

## Phase 5: Polish and Optimization

- [x] 17. Security and Privacy Implementation ✅
  - [x] 17.1 Implement structured logging with PII redaction
    - Created internal/logging package with Logger and PIIRedactor
    - Allowlisted fields: packageName, versionCode, track, status, etc.
    - Sensitive fields: email, userName, reviewText, token, etc.
    - Pattern-based redaction for emails, IPs, phone numbers, JWT tokens
    - Verbose mode outputs JSON, normal mode outputs simple format
    - _Requirements: 17.1, 17.4-17.8_

  - [x]* 17.2 Write security property tests
    - TestPIIRedactorAllowlistedFields - verifies safe fields not redacted
    - TestPIIRedactorSensitiveFields - verifies PII fields redacted
    - TestReviewTextDefaultExclusion - verifies review text redacted by default
    - TestGoogleGroupsRedaction - verifies email lists redacted
    - 19 total tests covering all redaction scenarios
    - **Validates: Requirements 17.1-17.3**

- [x] 18. Performance Optimization ✅
  - [x] 18.1 Implement lazy initialization patterns
    - API services (Publisher, Reporting) created on first access via sync.Once
    - Thread-safe lazy initialization prevents race conditions
    - HTTP client created once, services instantiated only when needed
    - _Requirements: 15.1, 15.7_

  - [x] 18.2 Implement concurrent API call limits
    - DefaultConcurrentCalls = 3 for metadata operations
    - Semaphore-based rate limiting via Acquire(ctx) method
    - AcquireForUpload(ctx) provides exclusive access (single-threaded uploads)
    - WithConcurrentCalls(n) option to customize limit
    - _Requirements: 15.6_

- [x] 19. Documentation and User Experience ✅
  - [x] 19.1 Implement help system and shell completion
    - Created gpd config completion command
    - Bash completion with _gpd_completions function
    - Zsh completion with _gpd function and _arguments
    - Fish completion with complete commands
    - All namespaces and commands have completions
    - _Requirements: 20.3, 29.6_

  - [x] 19.2 Implement init command for project scaffolding
    - Already implemented in config.go
    - Creates config directory and cache directory
    - Creates sample config.json, release-notes.json
    - Creates assets/ directory structure (locale/category)
    - Creates .gitignore for sensitive files
    - _Requirements: 29.3_

- [x] 20. Extended Property Test Suite (Optional) ✅
  - [x]* 20.1 Implement remaining property tests by subsystem
    - Output and exit code properties (internal/errors/codes_test.go - 17 tests)
    - Config and flags properties (internal/config/config_test.go, flags_test.go - 40 tests)
    - API package properties (internal/api/api_test.go - 13 tests)
    - CLI package properties (internal/cli/cli_test.go - 17 tests)
    - All remaining correctness properties covered

  - [x]* 20.2 Implement comprehensive unit test suite
    - Error condition handling and validation (17 tests)
    - File system operations and config management (40 tests)
    - API track validation and release config (13 tests)
    - CLI integration and flag parsing (17 tests)
    - Total: 307 tests across all packages
  
  - [ ]* 20.3 Optional live API integration tests
    - Sandbox environment testing (manual/nightly)
    - Basic connectivity and authentication validation

- [x] 21. Final Integration and Release ✅
  - [x] 21.1 Implement version and build information
    - Created Makefile with VERSION, GIT_COMMIT, BUILD_TIME variables
    - Version injected via ldflags at build time
    - Supports cross-platform builds (linux/darwin/windows, amd64/arm64)
    - SHA256 checksums generation via `make checksums`
    - _Requirements: 18.5, 18.7_

  - [x] 21.2 Final end-to-end testing and validation
    - All tests pass (make test)
    - Binary builds successfully with version info
    - JSON output correctly shows version, commit, build time
    - _Requirements: 15.1, 18.1_

- [x] 22. Final checkpoint - Ensure all tests pass ✅
  - All tests pass successfully
  - Build completes without errors
  - Version information correctly injected

## Notes

- Phase 0 is mandatory and must be completed before implementation begins
- Tasks marked with `*` are optional and can be skipped for faster MVP
- Non-optional property tests are integrated into the critical path to prevent contract drift
- Each task references specific requirements for traceability
- The plan follows vertical slice approach: complete thin functionality before expanding breadth
- Live API tests are optional and should be run manually or in nightly builds
- Resumable uploads and advanced features are deferred until core functionality is proven