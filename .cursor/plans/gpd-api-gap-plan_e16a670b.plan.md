---
name: gpd-api-gap-plan
overview: Comprehensive implementation plan for Google Play CLI roadmap with detailed milestones, acceptance criteria, and implementation specs across all phases.
todos:
  - id: phase1-foundation
    content: Implement Phase 1 - Edit lifecycle, error search, deobfuscation
    status: completed
  - id: phase2-monetization
    content: Implement Phase 2 - Modern subscriptions, base plans, offers
    status: completed
  - id: phase3-purchases
    content: Implement Phase 3 - Voided purchases and purchase actions
    status: completed
  - id: phase4-vitals
    content: Implement Phase 4 - Additional vitals metrics and anomalies
    status: completed
  - id: phase5-publishing
    content: Implement Phase 5 - Images API, internal sharing, app details
    status: completed
  - id: phase6-access
    content: Implement Phase 6 - Grants and users management
    status: completed
  - id: phase7-optional
    content: Implement Phase 7 - Games, recovery, generated APKs (optional)
    status: completed
isProject: true
---

# Google Play CLI - Comprehensive Implementation Roadmap

**Status:** Completed
**Start Date:** 2026-01-24
**Completion Date:** 2026-01-24
**Target APIs:** Android Publisher v3, Play Developer Reporting v1beta1, Play Games Management v1

---

## Implementation Principles

1. **Backward Compatibility:** All new features must not break existing commands
2. **Consistent UX:** Follow existing CLI patterns (flags, output, error handling)
3. **Test-Driven:** Write tests before implementation
4. **Documentation-First:** Update docs as features are added
5. **Atomic Releases:** Each phase can be released independently

---

## Phase 1: Foundation & Workflows (P0 - Critical)

**Estimated Effort:** 4-6 weeks
**Priority:** Critical - Blocks core workflows
**Dependencies:** None

### 1.1 Edit Lifecycle Exposure

**Goal:** Enable atomic multi-step releases and explicit edit management

#### Implementation Tasks:

**A. Edit Manager Refactoring** (`internal/edits/manager.go`)
- [ ] Add `GetOrCreateEdit()` method with caching
- [ ] Add `ListEdits()` to show active edits
- [ ] Add `ValidateEdit()` to pre-validate before commit
- [ ] Add `DeleteEdit()` for cleanup
- [ ] Add edit state tracking (draft, validating, committed)
- [ ] Implement edit TTL tracking (edits expire after inactivity)

**B. New Command File** (`internal/cli/edit_commands.go`)
- [ ] Create `edit_commands.go` with base command structure
- [ ] Implement `gpd publish edit create --package <pkg>`
  - Returns: `{"editId": "...", "expiryTime": "...", "id": "..."}`
- [ ] Implement `gpd publish edit list --package <pkg>`
  - Shows all active edits for package
- [ ] Implement `gpd publish edit get <edit-id> --package <pkg>`
  - Shows edit details and attached changes
- [ ] Implement `gpd publish edit commit <edit-id> --package <pkg>`
  - Commits edit and returns AppEdit response
- [ ] Implement `gpd publish edit validate <edit-id> --package <pkg>`
  - Pre-validates without committing
- [ ] Implement `gpd publish edit delete <edit-id> --package <pkg>`
  - Aborts and deletes edit

**C. Existing Command Integration**
- [ ] Add `--edit-id` flag to all `gpd publish` commands
- [ ] Add `--no-auto-commit` flag for manual commit workflow
- [ ] Update `upload`, `release`, `listing`, etc. to support explicit edit IDs

**D. Edit State Management & Error Handling**
- [ ] Define edit TTL: expire after 7 days or 1 hour since last activity
- [ ] Implement state transitions: draft → validating → committed | draft → aborted
- [ ] Handle edit conflicts with clear error messaging
- [ ] Add retry logic with exponential backoff for transient API errors
- [ ] Cache edit metadata locally to reduce API calls

**Example Workflow:**
```bash
# Manual edit workflow
EDIT=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.id')
gpd publish upload app.aab --package com.example.app --edit-id $EDIT --no-auto-commit
gpd publish listing update --title "New Title" --package com.example.app --edit-id $EDIT --no-auto-commit
gpd publish edit validate $EDIT --package com.example.app
gpd publish edit commit $EDIT --package com.example.app
```

**Acceptance Criteria:**
- [ ] Can create edit and get unique ID
- [ ] Can list all active edits for a package
- [ ] Can attach multiple operations to single edit
- [ ] Can validate without committing
- [ ] Can commit manually
- [ ] Can abort/delete edit
- [ ] Existing auto-commit behavior unchanged by default
- [ ] Unit tests with mocked API
- [ ] Integration tests with real API

**Files to Modify:**
- `internal/edits/manager.go` - Core logic
- `internal/cli/edit_commands.go` - New file
- `internal/cli/publish_commands.go` - Add `--edit-id` flag support
- `docs/api-coverage-matrix.md` - Update status for edits.*

---

### 1.2 Error Search & Reporting (Vitals)

**Goal:** Enable detailed crash analysis and stack trace retrieval

#### Implementation Tasks:

**A. API Client Enhancement** (`internal/api/client.go`)
- [ ] Add error search query methods to PlayReporting service
- [ ] Implement pagination for error results
- [ ] Add error report filtering (OS version, device, date range)

**B. Command Implementation** (`internal/cli/vitals_commands.go`)
- [ ] Implement `gpd vitals errors issues search --package <pkg>`
  - Flags: `--query`, `--interval`, `--page-size`, `--page-token`
  - Returns: Grouped error issues with signatures
- [ ] Implement `gpd vitals errors reports search --package <pkg>`
  - Flags: `--query`, `--interval`, `--page-size`
  - Returns: Individual error reports with stack traces
- [ ] Implement `gpd vitals errors counts get --package <pkg>`
  - Returns: Error count metrics
- [ ] Implement `gpd vitals errors counts query --package <pkg>`
  - Flags: `--start-date`, `--end-date`, `--dimensions`
  - Returns: Time-series error counts

**C. Output Formatting**
- [ ] Add table format for error issues (issue ID, type, count, devices)
- [ ] Add detailed view for stack traces (pretty-printed)
- [ ] Add CSV export for error counts

**D. Query Syntax & Formatting**
- [ ] Document query syntax: `issueId:<id>`, `type:CRASH|ANR`, `osVersion:>=14`, `deviceModel:<model>`
- [ ] Implement error grouping by stack trace signature (first 5 frames)
- [ ] Format stack traces with 2-space indentation, line numbers, and deobfuscation status
- [ ] Add `--deobfuscate` flag to attempt symbol resolution when mappings available

**Example Commands:**
```bash
# Search for crash issues
gpd vitals errors issues search --package com.example.app \
  --interval last30Days --page-size 50

# Get detailed reports for an issue
gpd vitals errors reports search --package com.example.app \
  --query "issueId:123456"

# Query error counts over time
gpd vitals errors counts query --package com.example.app \
  --start-date 2026-01-01 --end-date 2026-01-31 \
  --dimensions apiLevel,deviceModel
```

**Acceptance Criteria:**
- [ ] Can search and list error issues
- [ ] Can retrieve individual error reports
- [ ] Stack traces properly formatted
- [ ] Can filter by date range, OS version, device
- [ ] Pagination works correctly
- [ ] Can export to CSV
- [ ] Error grouping matches Play Console
- [ ] Unit tests for query building
- [ ] Integration tests with sample errors

**Files to Modify:**
- `internal/api/client.go` - Add error search methods
- `internal/cli/vitals_commands.go` - Add error commands
- `internal/output/formatter.go` - Add stack trace formatting
- `docs/api-coverage-matrix.md` - Update vitals.errors.* status

---

### 1.3 Deobfuscation File Upload

**Goal:** Enable automated ProGuard/R8 mapping and native symbol upload

#### Implementation Tasks:

**A. Command Implementation** (`internal/cli/publish_commands.go`)
- [ ] Add `deobfuscationCmd` subcommand to publish
- [ ] Implement `gpd publish deobfuscation upload <file> --package <pkg>`
  - Flags: `--version-code`, `--type` (proguard|nativeCode), `--edit-id`
  - Validates file exists and has correct format
  - Uploads via `edits.deobfuscationfiles.upload`

**B. File Validation & Upload Resilience**
- [ ] ProGuard mapping: max 50MB, validate format (class mappings present)
- [ ] Native symbols: max 100MB, validate .so.sym or ZIP containing .so.sym files
- [ ] Add `--chunk-size` flag for large uploads (default 10MB)
- [ ] Implement resume support for interrupted uploads
- [ ] Validate version code exists before upload

**C. CI/CD Integration**
- [ ] Add example GitHub Actions workflow
- [ ] Add example GitLab CI config
- [ ] Document integration with Gradle builds

**Example Commands:**
```bash
# Upload ProGuard mapping after build
gpd publish deobfuscation upload mapping.txt \
  --package com.example.app \
  --version-code 123 \
  --type proguard

# Upload native symbols
gpd publish deobfuscation upload symbols.zip \
  --package com.example.app \
  --version-code 123 \
  --type nativeCode

# With explicit edit
gpd publish deobfuscation upload mapping.txt \
  --package com.example.app \
  --version-code 123 \
  --type proguard \
  --edit-id abc123
```

**Acceptance Criteria:**
- [ ] Can upload ProGuard mapping files
- [ ] Can upload native code symbols
- [ ] File format validation works
- [ ] Works with explicit edit ID
- [ ] Auto-creates edit if needed
- [ ] Proper error messages for invalid files
- [ ] Example CI/CD configs documented
- [ ] Unit tests for file validation
- [ ] Integration tests with sample mappings

**Files to Modify:**
- `internal/cli/publish_commands.go` - Add deobfuscation commands
- `.github/workflows/ci.yml` - Add example usage
- `docs/examples/ci-cd-integration.md` - New file
- `docs/api-coverage-matrix.md` - Update deobfuscationfiles status

---

### Phase 1 Deliverables

- [ ] Edit lifecycle fully exposed and tested
- [ ] Error search with stack traces working
- [ ] Deobfuscation upload in CI/CD pipelines
- [ ] All commands documented with examples
- [ ] Updated coverage matrix showing Phase 1 complete
- [ ] Release notes prepared
- [ ] Migration guide for edit workflow (optional feature)

---

## Phase 2: Monetization Modernization (P0/P1 - High Priority)

**Estimated Effort:** 6-8 weeks
**Priority:** High - Required for modern subscription management
**Dependencies:** None (can run in parallel with Phase 1)

### 2.1 Modern Subscriptions API

**Goal:** Full CRUD for subscriptions with modern base plan/offer model

#### Implementation Tasks:

**A. API Client Enhancement** (`internal/api/client.go`)
- [ ] Add `MonetizationSubscriptions()` service accessor
- [ ] Add `MonetizationBasePlans()` service accessor
- [ ] Add `MonetizationOffers()` service accessor
- [ ] Implement batch operation helpers

**B. Core Subscription Commands** (`internal/cli/monetization_commands.go`)
- [ ] Refactor existing subscriptions from read-only
- [ ] Implement `gpd monetization subscriptions create --package <pkg>`
  - Flags: `--product-id`, `--package-name`
  - Returns: Created subscription object
- [ ] Implement `gpd monetization subscriptions update <sub-id> --package <pkg>`
  - Flags: `--archived`, `--listings-file` (JSON)
  - Supports partial updates
- [ ] Implement `gpd monetization subscriptions patch <sub-id> --package <pkg>`
  - Similar to update but uses PATCH semantics
- [ ] Implement `gpd monetization subscriptions delete <sub-id> --package <pkg>`
  - Requires confirmation flag
- [ ] Implement `gpd monetization subscriptions archive <sub-id> --package <pkg>`
  - Soft delete / archive
- [ ] Implement `gpd monetization subscriptions batchGet --ids <id1,id2> --package <pkg>`
  - Bulk retrieval
- [ ] Implement `gpd monetization subscriptions batchUpdate --file <json> --package <pkg>`
  - Bulk updates from JSON file

**C. Validation & Batch Limits**
- [ ] Product ID validation: `^[a-zA-Z0-9_]+$`, 1-100 characters, unique per package
- [ ] Conflict resolution: return 409 with current state on concurrent updates
- [ ] Batch operations: max 50 subscriptions per batch, max 10 concurrent requests
- [ ] Add `--dry-run` flag to validate without applying changes

**Example Commands:**
```bash
# Create new subscription
gpd monetization subscriptions create \
  --product-id premium_monthly \
  --package com.example.app

# Update subscription listings
gpd monetization subscriptions update premium_monthly \
  --package com.example.app \
  --listings-file listings.json

# Archive subscription
gpd monetization subscriptions archive premium_monthly \
  --package com.example.app

# Batch operations
gpd monetization subscriptions batchGet \
  --ids premium_monthly,premium_yearly \
  --package com.example.app
```

**Acceptance Criteria:**
- [ ] Can create subscriptions via CLI
- [ ] Can update subscription metadata
- [ ] Can archive/delete subscriptions
- [ ] Batch operations work correctly
- [ ] Proper validation for product IDs
- [ ] Error handling for conflicts
- [ ] Unit tests for all CRUD operations
- [ ] Integration tests with test products

**Files to Modify:**
- `internal/cli/monetization_commands.go` - Expand subscriptions
- `internal/api/client.go` - Add monetization services
- `docs/api-coverage-matrix.md` - Update monetization.subscriptions.*

---

### 2.2 Base Plans Management

**Goal:** Manage subscription base plans (pricing tiers)

#### Implementation Tasks:

**A. Base Plan Commands** (`internal/cli/monetization_commands.go`)
- [ ] Implement `gpd monetization baseplans activate <sub-id> <plan-id> --package <pkg>`
  - Activates base plan for sale
- [ ] Implement `gpd monetization baseplans deactivate <sub-id> <plan-id> --package <pkg>`
  - Deactivates base plan
- [ ] Implement `gpd monetization baseplans delete <sub-id> <plan-id> --package <pkg>`
  - Deletes base plan (with confirmation)
- [ ] Implement `gpd monetization baseplans migrate-prices <sub-id> <plan-id> --package <pkg>`
  - Flags: `--region-code`, `--price-micros`
  - Migrates pricing for a region
- [ ] Implement `gpd monetization baseplans batch-migrate-prices --file <json> --package <pkg>`
  - Bulk price migration
- [ ] Implement `gpd monetization baseplans batch-update-states --file <json> --package <pkg>`
  - Bulk state changes (activate/deactivate)

**B. Price Migration & State Rules**
- [ ] Warn if price change exceeds 50%; require `--force` to proceed
- [ ] Prevent deactivation when active subscriptions exist
- [ ] Add `--grace-period` flag to schedule deactivation
- [ ] Validate regional pricing completeness before activation

**Example Commands:**
```bash
# Activate base plan
gpd monetization baseplans activate premium_monthly monthly_plan \
  --package com.example.app

# Migrate pricing for a region
gpd monetization baseplans migrate-prices premium_monthly monthly_plan \
  --package com.example.app \
  --region-code US \
  --price-micros 4990000

# Batch update states
gpd monetization baseplans batch-update-states \
  --file baseplan-states.json \
  --package com.example.app
```

**Acceptance Criteria:**
- [ ] Can activate/deactivate base plans
- [ ] Can migrate regional pricing
- [ ] Batch operations work
- [ ] Price validation (micros format)
- [ ] Proper error handling
- [ ] Unit tests for state changes
- [ ] Integration tests with pricing

**Files to Modify:**
- `internal/cli/monetization_commands.go` - Add baseplans commands
- `docs/api-coverage-matrix.md` - Update basePlans.* status

---

### 2.3 Offers Management

**Goal:** Create and manage subscription offers (trials, discounts)

#### Implementation Tasks:

**A. Offers Commands** (`internal/cli/monetization_commands.go`)
- [ ] Implement `gpd monetization offers create <sub-id> <plan-id> --package <pkg>`
  - Flags: `--offer-id`, `--phases-file` (JSON with offer phases)
  - Creates new offer
- [ ] Implement `gpd monetization offers get <sub-id> <plan-id> <offer-id> --package <pkg>`
  - Retrieves offer details
- [ ] Implement `gpd monetization offers list <sub-id> <plan-id> --package <pkg>`
  - Lists all offers for a base plan
- [ ] Implement `gpd monetization offers update <sub-id> <plan-id> <offer-id> --package <pkg>`
  - Updates offer configuration
- [ ] Implement `gpd monetization offers delete <sub-id> <plan-id> <offer-id> --package <pkg>`
  - Deletes offer
- [ ] Implement `gpd monetization offers activate <sub-id> <plan-id> <offer-id> --package <pkg>`
  - Makes offer available
- [ ] Implement `gpd monetization offers deactivate <sub-id> <plan-id> <offer-id> --package <pkg>`
  - Disables offer
- [ ] Implement `gpd monetization offers batchGet --package <pkg>`
  - Bulk retrieval across subscriptions
- [ ] Implement `gpd monetization offers batchUpdate --file <json> --package <pkg>`
  - Bulk updates
- [ ] Implement `gpd monetization offers batchUpdateStates --file <json> --package <pkg>`
  - Bulk state changes

**B. Offer Phase Schema & Validation**
- [ ] Document offer phase JSON schema in examples
- [ ] Validate: trial phases first; discount phases require end date; recurring phases require duration
- [ ] Add `--validate-phases` flag to verify offer configuration

**Example Offer Phase JSON:**
```json
{
  "phases": [
    {
      "duration": "P7D",
      "price": {"currencyCode": "USD", "priceMicros": "0"},
      "billingPeriod": "P1M"
    },
    {
      "duration": "P1M",
      "price": {"currencyCode": "USD", "priceMicros": "4990000"},
      "billingPeriod": "P1M"
    }
  ]
}
```

**Example Commands:**
```bash
# Create trial offer
gpd monetization offers create premium_monthly monthly_plan \
  --offer-id trial_7day \
  --package com.example.app \
  --phases-file trial-offer.json

# Activate offer
gpd monetization offers activate premium_monthly monthly_plan trial_7day \
  --package com.example.app

# List all offers
gpd monetization offers list premium_monthly monthly_plan \
  --package com.example.app
```

**Acceptance Criteria:**
- [ ] Can create offers with phases
- [ ] Can activate/deactivate offers
- [ ] Can list and retrieve offers
- [ ] Batch operations work
- [ ] Offer phase validation
- [ ] Proper error messages
- [ ] Unit tests for offer lifecycle
- [ ] Integration tests with trial offers

**Files to Modify:**
- `internal/cli/monetization_commands.go` - Add offers commands
- `docs/examples/subscription-offers.md` - New file with examples
- `docs/api-coverage-matrix.md` - Update offers.* status

---

### 2.4 Regional Pricing & One-Time Products

**Goal:** Price conversion and modern one-time product management

#### Implementation Tasks:

**A. Regional Pricing** (`internal/cli/monetization_commands.go`)
- [ ] Implement `gpd monetization convert-region-prices --package <pkg>`
  - Flags: `--price-micros`, `--from-region`, `--to-regions`
  - Returns: Converted prices for target regions

**B. Regional Pricing Details**
- [ ] Supported regions: US, GB, DE, FR, JP, CA, AU, BR, IN, MX, KR, and others in Play Console
- [ ] Use Google Play daily exchange rates for conversions
- [ ] Add `--list-regions` flag to show supported region codes
- [ ] Validate region codes before conversion

**C. One-Time Products** (`internal/cli/monetization_commands.go`)
- [ ] Implement `gpd monetization onetimeproducts create --package <pkg>`
- [ ] Implement `gpd monetization onetimeproducts get <id> --package <pkg>`
- [ ] Implement `gpd monetization onetimeproducts list --package <pkg>`
- [ ] Implement `gpd monetization onetimeproducts update <id> --package <pkg>`
- [ ] Implement `gpd monetization onetimeproducts delete <id> --package <pkg>`
- [ ] Implement batch operations

**D. One-Time Product Types**
- [ ] Types: `managed` (non-consumable), `consumable` (repeatable)
- [ ] Add `--purchase-type` flag with validation against Play Console config

**Example Commands:**
```bash
# Convert pricing across regions
gpd monetization convert-region-prices \
  --package com.example.app \
  --price-micros 4990000 \
  --from-region US \
  --to-regions GB,DE,FR,JP

# Create one-time product
gpd monetization onetimeproducts create \
  --package com.example.app \
  --product-id coins_100 \
  --type managed
```

**Acceptance Criteria:**
- [ ] Price conversion works across regions
- [ ] One-time products CRUD complete
- [ ] Batch operations functional
- [ ] Currency handling correct
- [ ] Unit tests for conversions
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/monetization_commands.go` - Add pricing & one-time products
- `docs/api-coverage-matrix.md` - Update status

---

### Phase 2 Deliverables

- [ ] Modern subscription management fully functional
- [ ] Base plans and offers can be managed via CLI
- [ ] Regional pricing tools available
- [ ] One-time products migrated to modern API
- [ ] Comprehensive examples and guides
- [ ] Updated coverage matrix
- [ ] Release notes with migration guide from old API

---

## Phase 3: Purchase Management (P1 - High Priority)

**Estimated Effort:** 3-4 weeks
**Priority:** High - Complete purchase lifecycle
**Dependencies:** None

### 3.1 Voided Purchases

**Goal:** Track refunded/cancelled purchases for entitlement revocation

#### Implementation Tasks:

**A. Command Implementation** (`internal/cli/purchases_commands.go`)
- [ ] Implement `gpd purchases voided list --package <pkg>`
  - Flags: `--start-time`, `--end-time`, `--type` (product|subscription), `--max-results`, `--page-token`
  - Returns: List of voided purchase tokens with void times

**B. Integration Helpers**
- [ ] Add CSV export for voided purchases
- [ ] Add filtering by product ID
- [ ] Add date range validation

**Example Commands:**
```bash
# List voided purchases in date range
gpd purchases voided list \
  --package com.example.app \
  --start-time 2026-01-01T00:00:00Z \
  --end-time 2026-01-31T23:59:59Z

# Export to CSV for processing
gpd purchases voided list \
  --package com.example.app \
  --start-time 2026-01-01T00:00:00Z \
  --output csv > voided.csv
```

**Acceptance Criteria:**
- [ ] Can list voided purchases in time range
- [ ] Pagination works correctly
- [ ] Can filter by product type
- [ ] CSV export functional
- [ ] Timestamps correctly formatted
- [ ] Unit tests for time parsing
- [ ] Integration tests with voided purchases

**Files to Modify:**
- `internal/cli/purchases_commands.go` - Add voided list
- `docs/api-coverage-matrix.md` - Update voidedpurchases status

---

### 3.2 Product Purchase Actions

**Goal:** Complete product purchase lifecycle

#### Implementation Tasks:

**A. Acknowledge Command** (`internal/cli/purchases_commands.go`)
- [ ] Implement `gpd purchases products acknowledge --token <token> --product-id <id> --package <pkg>`
  - Acknowledges purchase to prevent refund window
  - Flags: `--developer-payload`

**B. Consume Command** (`internal/cli/purchases_commands.go`)
- [ ] Implement `gpd purchases products consume --token <token> --product-id <id> --package <pkg>`
  - Marks consumable purchase as consumed

**Example Commands:**
```bash
# Acknowledge purchase
gpd purchases products acknowledge \
  --token abc123token \
  --product-id premium_upgrade \
  --package com.example.app

# Consume consumable
gpd purchases products consume \
  --token def456token \
  --product-id coins_100 \
  --package com.example.app
```

**Acceptance Criteria:**
- [ ] Can acknowledge product purchases
- [ ] Can consume consumables
- [ ] Proper error handling for invalid tokens
- [ ] Works with both verify and standalone
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/purchases_commands.go` - Add acknowledge/consume
- `docs/api-coverage-matrix.md` - Update products.* status

---

### 3.3 Subscription Purchase Actions

**Goal:** Complete subscription management actions

#### Implementation Tasks:

**A. Subscription Commands** (`internal/cli/purchases_commands.go`)
- [ ] Implement `gpd purchases subscriptions acknowledge --token <token> --package <pkg>`
- [ ] Implement `gpd purchases subscriptions cancel --token <token> --package <pkg>`
  - Cancels subscription at end of period
- [ ] Implement `gpd purchases subscriptions defer --token <token> --package <pkg>`
  - Flags: `--expected-expiry-time`, `--desired-expiry-time`
  - Defers subscription renewal
- [ ] Implement `gpd purchases subscriptions refund --token <token> --package <pkg>`
  - Issues refund for subscription
- [ ] Implement `gpd purchases subscriptions revoke --token <token> --package <pkg>`
  - Immediately revokes subscription access

**B. v2 API Support**
- [ ] Implement v2 equivalents for cancel and revoke
- [ ] Add auto-detection of API version to use

**Example Commands:**
```bash
# Cancel subscription (end of period)
gpd purchases subscriptions cancel \
  --token sub_token_123 \
  --package com.example.app

# Defer subscription
gpd purchases subscriptions defer \
  --token sub_token_123 \
  --package com.example.app \
  --expected-expiry-time 2026-02-01T00:00:00Z \
  --desired-expiry-time 2026-03-01T00:00:00Z

# Refund subscription
gpd purchases subscriptions refund \
  --token sub_token_123 \
  --package com.example.app

# Revoke immediately
gpd purchases subscriptions revoke \
  --token sub_token_123 \
  --package com.example.app
```

**Acceptance Criteria:**
- [ ] All subscription actions work
- [ ] Both v1 and v2 APIs supported
- [ ] Proper validation for expiry times
- [ ] Confirmation required for destructive ops
- [ ] Error handling for invalid states
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/purchases_commands.go` - Add subscription actions
- `docs/api-coverage-matrix.md` - Update subscriptions.* and subscriptionsv2.* status

---

### Phase 3 Deliverables

- [ ] Voided purchases tracking implemented
- [ ] Product purchase actions complete
- [ ] Subscription actions complete
- [ ] Both v1 and v2 subscription APIs supported
- [ ] Documentation with examples
- [ ] Updated coverage matrix
- [ ] Release notes

---

## Phase 4: Vitals Expansion (P2 - Medium Priority)

**Estimated Effort:** 2-3 weeks
**Priority:** Medium - Enhanced vitals monitoring
**Dependencies:** Phase 1 (for error search pattern)

### 4.1 Additional Metric Sets

**Goal:** Complete vitals coverage beyond crashes/ANRs

#### Implementation Tasks:

**A. New Metric Commands** (`internal/cli/vitals_commands.go`)
- [ ] Implement `gpd vitals excessive-wakeups --start-date <date> --end-date <date> --package <pkg>`
- [ ] Implement `gpd vitals lmk-rate --start-date <date> --end-date <date> --package <pkg>`
  - Low Memory Kill rate metrics
- [ ] Implement `gpd vitals slow-rendering --start-date <date> --end-date <date> --package <pkg>`
- [ ] Implement `gpd vitals slow-start --start-date <date> --end-date <date> --package <pkg>`
- [ ] Implement `gpd vitals stuck-wakelocks --start-date <date> --end-date <date> --package <pkg>`

**B. Unified Query Interface**
- [ ] Enhance `gpd vitals query` to support all metric types
- [ ] Add metric auto-completion
- [ ] Standardize output format across metrics

**C. Metric Definitions & Dimensions**
- [ ] Excessive wakeups: >10 wakeups per hour per device
- [ ] LMK rate: low memory kills per 1000 sessions
- [ ] Slow rendering: frames >16ms
- [ ] Slow start: app start >3 seconds
- [ ] Stuck wakelocks: held >1 hour
- [ ] Supported dimensions: `apiLevel`, `deviceModel`, `countryCode`, `appVersion`, `osVersion`
- [ ] Add `--list-dimensions` flag to show dimensions per metric

**Example Commands:**
```bash
# Query excessive wakeups
gpd vitals excessive-wakeups \
  --package com.example.app \
  --start-date 2026-01-01 \
  --end-date 2026-01-31 \
  --dimensions apiLevel,deviceModel

# Slow start metrics
gpd vitals slow-start \
  --package com.example.app \
  --start-date 2026-01-01 \
  --end-date 2026-01-31
```

**Acceptance Criteria:**
- [ ] All 5 additional metrics implemented
- [ ] Consistent query interface
- [ ] Proper data visualization
- [ ] CSV export works
- [ ] Dimensions filtering works
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/vitals_commands.go` - Add metric commands
- `docs/api-coverage-matrix.md` - Update vitals.* status

---

### 4.2 Anomalies Detection

**Goal:** Surface detected anomalies in app metrics

#### Implementation Tasks:

**A. Anomalies Command** (`internal/cli/vitals_commands.go`)
- [ ] Implement `gpd vitals anomalies list --package <pkg>`
  - Lists detected anomalies with severity
  - Flags: `--metric`, `--time-period`

**B. Anomaly Severity & Time Periods**
- [ ] Severity: HIGH (>50% change), MEDIUM (20-50%), LOW (10-20%)
- [ ] Time periods: `last7Days`, `last30Days`, `last90Days`, `custom`
- [ ] Add `--min-severity` flag to filter results

**Example Commands:**
```bash
# List all anomalies
gpd vitals anomalies list --package com.example.app

# Filter by metric
gpd vitals anomalies list \
  --package com.example.app \
  --metric crashRate
```

**Acceptance Criteria:**
- [ ] Can list anomalies
- [ ] Filtering works
- [ ] Severity levels shown
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/vitals_commands.go` - Add anomalies
- `docs/api-coverage-matrix.md` - Update anomalies status

---

### Phase 4 Deliverables

- [ ] All vitals metric sets covered
- [ ] Anomalies detection available
- [ ] Comprehensive vitals guide
- [ ] Updated coverage matrix
- [ ] Release notes

---

## Phase 5: Publishing Enhancements (P1/P2 - Medium Priority)

**Estimated Effort:** 4-5 weeks
**Priority:** Medium - Quality of life improvements
**Dependencies:** Phase 1 (edit lifecycle)

### 5.1 Images API (Proper Implementation)

**Goal:** Replace simplified assets with proper edits.images API

#### Implementation Tasks:

**A. Images Commands** (`internal/cli/publish_commands.go`)
- [ ] Implement `gpd publish images upload <type> <file> --package <pkg> --locale <locale>`
  - Types: icon, featureGraphic, promoGraphic, tvBanner, phoneScreenshots, etc.
  - Returns: Image ID and URL
- [ ] Implement `gpd publish images list <type> --package <pkg> --locale <locale>`
  - Lists all images of type for locale
- [ ] Implement `gpd publish images delete <type> <id> --package <pkg> --locale <locale>`
  - Deletes specific image
- [ ] Implement `gpd publish images deleteall <type> --package <pkg> --locale <locale>`
  - Deletes all images of type

**B. Migration from Old Assets**
- [ ] Add deprecation warning to old `assets` commands
- [ ] Provide migration script/guide
- [ ] Support both APIs during transition

**C. Validation**
- [ ] Implement image dimension validation
- [ ] Implement file size validation
- [ ] Add format validation (PNG, JPEG)

**D. Image Requirements**
- [ ] icon: 512x512px, max 1MB
- [ ] featureGraphic: 1024x500px, max 15MB
- [ ] phoneScreenshots: 320-3840px width, max 8MB each
- [ ] tabletScreenshots: 600-3840px width, max 8MB each
- [ ] tvBanner: 1280x720px, max 15MB
- [ ] Validate format before upload
- [ ] Add `--validate-only` flag to check without uploading

**Example Commands:**
```bash
# Upload screenshot
gpd publish images upload phoneScreenshots screenshot1.png \
  --package com.example.app \
  --locale en-US

# List all phone screenshots
gpd publish images list phoneScreenshots \
  --package com.example.app \
  --locale en-US

# Delete specific image
gpd publish images delete phoneScreenshots img_abc123 \
  --package com.example.app \
  --locale en-US
```

**Acceptance Criteria:**
- [ ] All image types supported
- [ ] Per-locale management works
- [ ] Validation catches invalid images
- [ ] Delete operations work
- [ ] Migration guide complete
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/publish_commands.go` - Add images commands
- `docs/migration/assets-to-images.md` - New file
- `docs/api-coverage-matrix.md` - Update images.* status

---

### 5.2 Internal App Sharing

**Goal:** Quick internal testing workflow

#### Implementation Tasks:

**A. Internal Share Command** (`internal/cli/publish_commands.go`)
- [ ] Implement `gpd publish internal-share upload <file> --package <pkg>`
  - Supports APK and AAB
  - Returns: Download URL for internal sharing
  - No edit needed, immediate availability

**Example Commands:**
```bash
# Share APK internally
gpd publish internal-share upload app-debug.apk \
  --package com.example.app

# Returns download URL
{
  "downloadUrl": "https://play.google.com/apps/internaltest/...",
  "expiresAt": "2026-02-24T00:00:00Z"
}
```

**Acceptance Criteria:**
- [ ] Can upload APK/AAB
- [ ] Returns shareable URL
- [ ] Fast upload (no edit overhead)
- [ ] Expiry time shown
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/publish_commands.go` - Add internal-share
- `docs/api-coverage-matrix.md` - Update internalappsharingartifacts status

---

### 5.3 App Details Management

**Goal:** Manage app contact info and settings

#### Implementation Tasks:

**A. Details Commands** (`internal/cli/publish_commands.go`)
- [ ] Implement `gpd publish details get --package <pkg>`
  - Shows contact email, phone, website, default language
- [ ] Implement `gpd publish details update --package <pkg>`
  - Flags: `--contact-email`, `--contact-phone`, `--contact-website`, `--default-language`
- [ ] Implement `gpd publish details patch --package <pkg>`
  - Partial updates

**Example Commands:**
```bash
# Get app details
gpd publish details get --package com.example.app

# Update contact info
gpd publish details update \
  --package com.example.app \
  --contact-email support@example.com \
  --contact-website https://example.com
```

**Acceptance Criteria:**
- [ ] Can get details
- [ ] Can update contact info
- [ ] Validation for email/URL formats
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/publish_commands.go` - Add details commands
- `docs/api-coverage-matrix.md` - Update details.* status

---

### Phase 5 Deliverables

- [ ] Images API properly implemented
- [ ] Internal sharing workflow available
- [ ] App details management complete
- [ ] Migration guide from old assets
- [ ] Updated coverage matrix
- [ ] Release notes

---

## Phase 6: Access Control (P2 - Medium Priority)

**Estimated Effort:** 2-3 weeks
**Priority:** Medium - Team management automation
**Dependencies:** None

### 6.1 Grants Management

**Goal:** Automate app-level permission grants

#### Implementation Tasks:

**A. New Command File** (`internal/cli/permissions_commands.go`)
- [ ] Create permissions command group
- [ ] Implement `gpd permissions grants create --package <pkg>`
  - Flags: `--email`, `--app-permissions` (JSON or comma-separated)
  - Grants user access to app
- [ ] Implement `gpd permissions grants delete <name> --package <pkg>`
  - Revokes grant
- [ ] Implement `gpd permissions grants patch <name> --package <pkg>`
  - Updates grant permissions

**B. Permission Scopes**
- [ ] Valid permissions: `VIEW_APP_INFORMATION`, `MANAGE_PRODUCTION_RELEASES`, `MANAGE_TEST_RELEASES`, `VIEW_FINANCIAL_DATA`, `MANAGE_ORDERS`, `REPLY_TO_REVIEWS`
- [ ] Add `--list-permissions` flag to show available permissions
- [ ] Validate permission names (case-sensitive) before grants

**Example Commands:**
```bash
# Grant user access to app
gpd permissions grants create \
  --package com.example.app \
  --email user@example.com \
  --app-permissions "VIEW_APP_INFORMATION,MANAGE_PRODUCTION_RELEASES"

# Revoke grant
gpd permissions grants delete grants/abc123 \
  --package com.example.app
```

**Acceptance Criteria:**
- [ ] Can create grants
- [ ] Can revoke grants
- [ ] Can update permissions
- [ ] Permission validation
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/permissions_commands.go` - New file
- `internal/api/client.go` - Add grants methods if needed
- `docs/api-coverage-matrix.md` - Update grants.* status

---

### 6.2 Users Management

**Goal:** Manage developer account users

#### Implementation Tasks:

**A. Users Commands** (`internal/cli/permissions_commands.go`)
- [ ] Implement `gpd permissions users create --email <email>`
  - Creates user in developer account
- [ ] Implement `gpd permissions users list`
  - Lists all users
- [ ] Implement `gpd permissions users delete <name>`
  - Removes user
- [ ] Implement `gpd permissions users patch <name>`
  - Updates user permissions

**Example Commands:**
```bash
# Add user to account
gpd permissions users create --email newuser@example.com

# List all users
gpd permissions users list

# Remove user
gpd permissions users delete users/abc123
```

**Acceptance Criteria:**
- [ ] User CRUD operations work
- [ ] List shows all users
- [ ] Proper validation
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/permissions_commands.go` - Add users commands
- `docs/api-coverage-matrix.md` - Update users.* status

---

### Phase 6 Deliverables

- [ ] Grants management implemented
- [ ] Users management implemented
- [ ] Automation examples (onboarding/offboarding)
- [ ] Updated coverage matrix
- [ ] Release notes

---

## Phase 7: Optional Features (P3/Optional - Low Priority)

**Estimated Effort:** 6-8 weeks (if implemented)
**Priority:** Optional - Game-specific and edge cases
**Dependencies:** None

### 7.1 Play Games Services (Optional)

**Goal:** Support game developers with achievements, leaderboards

#### Implementation Tasks:

**A. API Client Enhancement** (`internal/api/client.go`)
- [ ] Add `GamesManagement()` service accessor
- [ ] Add auth scope: `https://www.googleapis.com/auth/games`

**B. New Command File** (`internal/cli/games_commands.go`)
- [ ] Create games command group
- [ ] Implement achievements reset commands
- [ ] Implement scores reset commands
- [ ] Implement events reset commands
- [ ] Implement players hide/unhide
- [ ] Implement rooms reset

**Example Commands:**
```bash
# Reset achievement for tester
gpd games achievements reset achievement_123 \
  --package com.example.game

# Hide player scores
gpd games players hide player_456 \
  --package com.example.game
```

**Acceptance Criteria:**
- [ ] All game management operations work
- [ ] Only for games (not general apps)
- [ ] Proper scoping
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/games_commands.go` - New file
- `internal/api/client.go` - Add games service
- `internal/auth/manager.go` - Add games scope
- `docs/api-coverage-matrix.md` - Update games.* status

---

### 7.2 App Recovery (Optional)

**Goal:** Emergency rollback capabilities

#### Implementation Tasks:

**A. New Command File** (`internal/cli/recovery_commands.go`)
- [ ] Implement `gpd recovery create --package <pkg> --version-code <code>`
- [ ] Implement `gpd recovery list --package <pkg>`
- [ ] Implement `gpd recovery deploy <id> --package <pkg>`
- [ ] Implement `gpd recovery cancel <id> --package <pkg>`
- [ ] Implement `gpd recovery add-targeting <id> --package <pkg>`

**Example Commands:**
```bash
# Create recovery action
gpd recovery create \
  --package com.example.app \
  --version-code 122

# Deploy recovery
gpd recovery deploy recovery_abc123 \
  --package com.example.app
```

**Acceptance Criteria:**
- [ ] Recovery CRUD works
- [ ] Targeting works
- [ ] Confirmation for deploy
- [ ] Unit tests
- [ ] Integration tests

**Files to Modify:**
- `internal/cli/recovery_commands.go` - New file
- `docs/api-coverage-matrix.md` - Update apprecovery.* status

---

### 7.3 Other Optional Features

**A. Generated APKs**
- [ ] Implement download of split APKs from bundles
- [ ] List generated variants

**B. Expansion Files**
- [ ] Legacy OBB file management (low priority)

**C. Country Availability**
- [ ] Query country availability

---

### Phase 7 Deliverables

- [ ] Games management (if applicable)
- [ ] App recovery (if needed)
- [ ] Other features as requested
- [ ] Updated coverage matrix
- [ ] Release notes

---

## Cross-Cutting Concerns

### Testing Strategy

**Unit Tests:**
- [ ] Every new command has unit tests with mocked API
- [ ] Test flag parsing and validation
- [ ] Test error handling paths
- [ ] Minimum 80% code coverage for new code

**Integration Tests:**
- [ ] Critical flows tested against real API
- [ ] Use test packages/products
- [ ] Automated in CI/CD
- [ ] Run on every PR

**Test Data Management:**
- [ ] Use dedicated test package (configurable via env var)
- [ ] Create test products/subscriptions with `_test` suffix
- [ ] Add `gpd test setup` to initialize test environment
- [ ] Add `gpd test cleanup` to remove test data

**CI/CD Test Execution:**
- [ ] Unit tests on every commit
- [ ] Integration tests on PRs when credentials exist
- [ ] E2E tests nightly or on release branches
- [ ] Skip integration/E2E if `GPD_API_CREDENTIALS` missing

**E2E Tests:**
- [ ] Full workflows tested (upload → release → verify)
- [ ] Run nightly or on release branches

### Documentation Requirements

**Per Phase:**
- [ ] Update `docs/api-coverage-matrix.md` with new statuses
- [ ] Update `docs/api-gap-analysis.md` with implementation notes
- [ ] Update `docs/gap-analysis-summary.md` with progress
- [ ] Add command examples to README or separate guides

**New Docs to Create:**
- [ ] `docs/examples/edit-workflow.md` - Edit lifecycle examples
- [ ] `docs/examples/subscription-management.md` - Modern monetization
- [ ] `docs/examples/error-debugging.md` - Error search guide
- [ ] `docs/examples/ci-cd-integration.md` - CI/CD patterns
- [ ] `docs/migration/assets-to-images.md` - Migration guide

### Release Strategy

**Versioning:**
- Use semantic versioning
- Each phase = minor version bump
- Breaking changes = major version bump

**Release Checklist per Phase:**
- [ ] All tests passing
- [ ] Documentation updated
- [ ] Release notes prepared
- [ ] Migration guide (if breaking changes)
- [ ] Changelog updated
- [ ] GitHub release created
- [ ] Homebrew formula updated

### Backward Compatibility

**Rules:**
- Existing commands must not break
- New flags added as optional
- Deprecation warnings before removal
- Migration guides for breaking changes

**Deprecation Process:**
1. Add deprecation warning (1 release)
2. Mark as deprecated in docs
3. Remove in next major version

---

## Success Metrics

### Coverage Goals

**Phase 1:** 30% → 35% coverage
**Phase 2:** 35% → 50% coverage
**Phase 3:** 50% → 55% coverage
**Phase 4:** 55% → 60% coverage
**Phase 5:** 60% → 70% coverage
**Phase 6:** 70% → 75% coverage
**Phase 7:** 75% → 85% coverage (if implemented)

### Quality Metrics

- 80%+ code coverage
- <5% error rate in integration tests
- <100ms added latency per command
- Zero security vulnerabilities
- All critical paths documented

---

## Dependencies & Prerequisites

### External Dependencies

- Go 1.21+ for implementation
- Google Play API access
- Test packages for integration testing
- CI/CD infrastructure for automated testing

### Internal Dependencies

- Current codebase stable and tested
- Auth flow working correctly
- Output formatting consistent

---

## Risk Management

### Technical Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| API changes/deprecations | High | Monitor API changelog, add version detection |
| Breaking changes | High | Strict backward compatibility, deprecation process |
| Performance degradation | Medium | Benchmark tests, optimize before release |
| Auth scope issues | Medium | Clear scope documentation, validation |
| Error format drift | Low | Centralize error templates, add tests |

### Schedule Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| Phase overruns | Medium | Buffer time in estimates, prioritize ruthlessly |
| Resource constraints | High | Clear priorities, phase independence |
| API rate limits | Low | Implement backoff, caching, explicit limits |

### API Rate Limits

- Android Publisher API: 1000 requests per 100 seconds per project
- Play Reporting API: 100 requests per 100 seconds per project
- Implement token bucket rate limiting with exponential backoff
- Add `--rate-limit-info` flag to show current status

### Required Auth Scopes by Phase

- Phase 1: `https://www.googleapis.com/auth/androidpublisher`
- Phase 2: `https://www.googleapis.com/auth/androidpublisher`
- Phase 3: `https://www.googleapis.com/auth/androidpublisher`
- Phase 4: `https://www.googleapis.com/auth/playdeveloperreporting`
- Phase 5: `https://www.googleapis.com/auth/androidpublisher`
- Phase 6: `https://www.googleapis.com/auth/androidpublisher`
- Phase 7: `https://www.googleapis.com/auth/games`

---

## Sign-off Criteria

### Per Phase

- [ ] All acceptance criteria met
- [ ] Tests passing (unit + integration)
- [ ] Documentation complete
- [ ] Code reviewed
- [ ] Coverage matrix updated
- [ ] Release notes drafted

### Overall Project

- [ ] All P0-P2 phases complete
- [ ] 70%+ API coverage achieved
- [ ] Zero critical bugs
- [ ] Documentation comprehensive
- [ ] Community feedback addressed

---

## Timeline Summary

| Phase | Duration | Dependencies | Priority |
|-------|----------|--------------|----------|
| Phase 1 | 4-6 weeks | None | P0 |
| Phase 2 | 6-8 weeks | None | P0/P1 |
| Phase 3 | 3-4 weeks | None | P1 |
| Phase 4 | 2-3 weeks | Phase 1 | P2 |
| Phase 5 | 4-5 weeks | Phase 1 | P1/P2 |
| Phase 6 | 2-3 weeks | None | P2 |
| Phase 7 | 6-8 weeks | None | P3/Optional |

**Total Estimated Duration:** 21-31 weeks (5-8 months)
**With parallelization:** 12-18 weeks (3-4.5 months)

---

## Next Steps

1. **Review & Approve Plan:** Stakeholder sign-off
2. **Set Up Project Board:** GitHub Projects with phases as milestones
3. **Create Initial Issues:** Break down Phase 1 into implementable tasks
4. **Establish Baseline:** Ensure current tests pass, coverage measured
5. **Start Phase 1:** Begin implementation

**Issue Structure:**
- [ ] Include phase number and priority
- [ ] Add acceptance criteria checklist
- [ ] List API endpoints to implement
- [ ] Estimate effort and dependencies

**Baseline Metrics:**
- [ ] Run `gpd capabilities` to record current API coverage
- [ ] Run `go test -cover ./...` to capture baseline coverage
- [ ] Document current command count and coverage percentage

---

## Appendix: Implementation Standards

### Error Message Format

- Format: `[Category] Description: Details (Error Code: XXX)`
- Example: `[Validation] Invalid product ID: Must match pattern ^[a-zA-Z0-9_]+$ (Error Code: VAL001)`

### Output Format

- Default: table format for human-readable output
- JSON: `--output json`
- CSV: `--output csv` for list commands
- Consistent field names across commands

### Logging Levels

- `--verbose`: show API requests/responses
- `--debug`: show internal state and caching details
- Default: errors and warnings only

---

**Plan Status:** ✅ Ready for Execution
**Last Updated:** 2026-01-24
**Next Review:** Start of each phase
