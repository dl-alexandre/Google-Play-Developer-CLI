# Consolidated Master Plan - Google Play Developer CLI (gpd)

**Last Updated:** 2026-02-10
**Source Plans:**
- `docs/PLAN.md` - Outstanding Improvement Plan
- `.github/plans/authrefreshmitigations_cbc5d107.plan.md` - Auth Refresh Mitigations

---

## 📋 Executive Summary

This document consolidates all active plans for the Google Play Developer CLI project. Currently there are **2 active plans** with work spanning API coverage gaps, App Store Connect parity improvements, and authentication system enhancements.

**Quick Stats:**
- 🔴 Priority 1 Items: 1
- 🟡 Priority 2 Items: 3
- 🟢 Priority 3 Items: 4
- 🔧 Active Todos: 4 (implementation complete, validation pending)
- ⚠️ Empty Plans: 0

---

## 🔴 PRIORITY 1: API Coverage Gap

### 1. LMK Rate Support in Vitals Commands
**Status:** `BLOCKED` (waiting on Google Play Developer Reporting API)
**Gap:** `gpd vitals lmk-rate` currently marked unsupported
**Source:** `docs/api-coverage-matrix.md` (Play Developer Reporting API v1beta1 / LMK Rate)

**Exit Criteria:**
- [ ] Implement command wiring when endpoint becomes available
- [ ] Add comprehensive tests
- [ ] Update documentation
- [ ] Remove unsupported marker from coverage matrix

**Notes:** This is blocked by Google Play API capability. Monitor API release notes.

---

## 🟡 PRIORITY 2: ASC Parity Gaps (High-Value, Feasible)

### 2. Beta Testing Workflows Enhancement
**Status:** `IN PROGRESS` / `PARTIAL` (compatibility layer added)
**Gap:** No ASC-style beta group lifecycle parity (group CRUD, richer assignment flows)
**Current:** `gpd publish testers list/get/add/remove` only
**Source:** `docs/asc-parity.md` (Beta Groups, Beta Testers: Partial)

**Exit Criteria:**
- [x] Define Play-equivalent abstractions for group-like management
- [x] Design UX for group management where Play supports it
- [x] Document supported vs unsupported semantics explicitly
- [x] Add examples for common beta testing workflows

**Platform Note:** Google Play's track-based model differs from ASC's group model. Need clear mapping.

---

### 3. Review Response Parity Handling
**Status:** `DOCUMENTED LIMITATION (API DOES NOT SUPPORT DELETE)`
**Gap:** Response delete path documented as unsupported
**Current:** Commands exist: list/get/reply/response get/for-review/delete
**Source:** `docs/asc-parity.md` (App Store reviews: Partial)

**Exit Criteria:**
- [x] Verify actual API capability vs documented behavior (code path returns explicit unsupported API limitation)
- [x] Test delete response functionality path at CLI level (returns validation error with workaround)
- [x] Implement missing behavior OR document hard platform limitation
- [x] Provide exact guidance for workaround if blocked

**Action Required:** Optional live API reconfirmation during next release validation cycle.

---

### 4. Submission/Release Workflow Parity
**Status:** `DOCS IMPLEMENTED`
**Gap:** Release/submission mapping incomplete due to workflow differences
**Current:** `publish release/rollout/promote/halt/rollback/status`
**Source:** `docs/asc-parity.md` (Versions, Submit: Partial)

**Exit Criteria:**
- [x] Publish clear workflow mapping doc for common ASC journeys
- [ ] Add examples for:
  - [x] Staged rollout decision paths
  - [x] Halt scenarios and recovery
  - [x] Rollback procedures
  - [x] Promotion workflows (alpha → beta → production)
- [x] Create side-by-side comparison: ASC command → gpd equivalent

---

## 🟢 PRIORITY 3: Parity Gaps (Platform-Limit or UX Clarification)

### 5. Authentication Parity Boundaries
**Status:** `COMPLETE`
**Gap:** ASC browser login flow parity is partial
**Current:** Device-code OAuth + service account model
**Source:** `docs/asc-parity.md` (Authentication: Partial)

**Exit Criteria:**
- [x] Document explicit auth decision tree
- [x] Create migration guidance for ASC users
- [x] Explain why browser flow differs (security model)
- [x] Add troubleshooting guide for auth issues

---

### 6. App/Build Model Differences
**Status:** `DOCS IMPLEMENTED`
**Gap:** No global build registry, no build-level beta group assignment equivalent
**Source:** `docs/asc-parity.md` (Apps & Builds: Partial)

**Exit Criteria:**
- [x] Document Google Play build model limitations
- [x] Document recommended Play-native alternatives
- [x] Explain track-based vs artifact-based distribution
- [x] Create migration guide for ASC build workflows

---

### 7. Analytics/Reporting Scope Differences
**Status:** `DOCS IMPLEMENTED`
**Gap:** ASC analytics/sales scope broader than current Play Reporting coverage
**Source:** `docs/asc-parity.md` (Analytics & Sales: Partial)

**Exit Criteria:**
- [x] Document supported Play Reporting datasets
- [x] Create unsupported ASC analogs reference
- [x] Map ASC metrics → available Play metrics
- [x] Document gaps that require Console access

---

### 8. Metadata/Localization Workflow Differences
**Status:** `DOCS IMPLEMENTED`
**Gap:** App setup, app info, localizations remain partial due to model differences
**Source:** `docs/asc-parity.md` (App Setup, App Info, Localizations: Partial)

**Exit Criteria:**
- [x] Add side-by-side task mapping: ASC task → gpd command sequence
- [x] Document Google Play listing model (track-based localizations)
- [x] Create examples for common metadata operations
- [x] Add troubleshooting for localization edge cases

---

## 🔐 AUTH REFRESH MITIGATIONS PLAN

**Plan ID:** `authrefreshmitigations_cbc5d107`
**Scope:** Cross-repo implementation (gdrv + gpd)
**Status:** `IN PROGRESS`

This plan addresses authentication error handling, diagnostics, and token storage stability.

### Todos

#### TODO-1: Auth Error Classifiers and Remediation
**Status:** ✅ `IMPLEMENTED (CODE)`, tests/validation pending
**Scope:** Both gdrv and gpd

**Implementation:**
- [x] Add shared auth-error classifier per repo that normalizes:
  - [x] OAuth refresh failures: `invalid_grant`, `invalid_client`, `unauthorized_client`
  - [x] API failures (401/403) into consistent remediation output
  - [x] Optional clock-skew detection using response `Date` headers
- [x] Wire classifier into refresh path so all commands surface consistent remediation text

**Test Requirements:**
- [ ] Unit tests for classifier behavior
- [ ] Test with simulated 401/403 responses
- [ ] Test clock-skew detection scenarios

---

#### TODO-2: Auth Diagnose Commands
**Status:** ✅ `IMPLEMENTED (CODE)`, manual validation pending
**Scope:** Both gdrv and gpd

**Command:** `auth diagnose`

**Requirements:**
- [x] Print active profile
- [x] Print token storage location
- [x] Print client-id hash fingerprint
- [x] Print authorized scopes
- [x] Print token timestamps
- [x] Print refresh-token presence

**Flags:**
- [x] `--refresh-check` - Attempt a refresh and show classifier output on failure
- [ ] `--json` - Output for automation/CI integration

**Integration Points:**
- `gdrv`: `internal/auth/manager.go`, `internal/cli/auth.go`
- `gpd`: `internal/auth/auth.go`, `internal/cli/auth_commands.go`

**Validation:**
- [ ] Manual test: `gdrv auth diagnose --refresh-check` with valid creds
- [ ] Manual test: `gdrv auth diagnose --refresh-check` with invalid creds
- [ ] Manual test: `gpd auth diagnose --refresh-check` with valid creds
- [ ] Manual test: `gpd auth diagnose --refresh-check` with invalid creds

---

#### TODO-3: Stable Token Storage
**Status:** ✅ `IMPLEMENTED (CODE)`, migration validation pending
**Scope:** Both gdrv and gpd

**Implementation:**
- [x] Change token storage paths to include both profile AND client-id hash
- [x] Add metadata file alongside stored tokens to detect client mismatches
- [x] Migration: detect old storage format and migrate to new format
- [x] Handle edge case: user switches client IDs

**Files to Modify:**
- `gdrv`: `internal/auth/storage.go`
- `gpd`: Token storage implementation (verify location)

---

#### TODO-4: Documentation Updates
**Status:** 🟡 `PARTIAL` (gdrv done, gpd updated)
**Scope:** Both gdrv and gpd

**Updates Required:**
- [x] Document testing-mode refresh token expiry (7 days)
- [x] Document refresh-token issuance cap (100 tokens)
- [x] Replace older "50 tokens" references if present
- [x] Add troubleshooting section for "invalid_grant" errors
- [x] Add guide on when to revoke and re-authenticate

---

## ⚠️ EMPTY PLANS

No empty plan files currently remain in `.github/plans`.

---

## 📊 Backlog Hygiene Requirements

### Ongoing Maintenance Tasks

1. **Synchronization Rule**
   - [ ] Keep `docs/api-coverage-matrix.md` synchronized after every feature change
   - [ ] Keep `docs/asc-parity.md` synchronized after every feature change

2. **Status Tracking Rule**
   For each open parity item, mark ONE of:
   - ✅ `Implementable now` - Ready for development
   - ⏸️ `Blocked by Google Play API capability` - Waiting on Google
   - 🚫 `Intentional non-goal` - Explicitly not implementing

3. **Ticket Conversion Rule**
   - [ ] Add explicit target status field when converting this plan into execution tickets
   - [ ] Add owner field when converting into execution tickets
   - [ ] Link tickets back to this master plan

---

## 🎯 Quick Reference: Priority Matrix

| Priority | Item | Status | Blocker | Owner |
|----------|------|--------|---------|-------|
| P1 | LMK Rate Support | 🔴 Blocked | Google API | TBD |
| P2 | Beta Testing Workflows | 🟡 Partial (compatibility commands added) | Play API model limits | TBD |
| P2 | Review Response Parity | ✅ Documented Limitation | Google Play API | TBD |
| P2 | Submission/Release Workflow | ✅ Docs Implemented | None | TBD |
| P3 | Auth Parity Documentation | ✅ Complete | None | TBD |
| P3 | App/Build Model Docs | ✅ Docs Implemented | None | TBD |
| P3 | Analytics Scope Docs | ✅ Docs Implemented | None | TBD |
| P3 | Metadata/Localization Docs | ✅ Docs Implemented | None | TBD |
| Auth | Error Classifiers | ✅ Implemented (Code) | Validation | TBD |
| Auth | Diagnose Commands | ✅ Implemented (Code) | Manual verification | TBD |
| Auth | Stable Token Storage | ✅ Implemented (Code) | Migration verification | TBD |
| Auth | Documentation Updates | 🟡 Partial Complete | Remaining docs sweep | TBD |

---

## 📝 Notes

### Implementation Guidelines
- Follow existing codebase patterns (see AGENTS.md in each repo)
- All CLIs use Cobra framework
- Error handling via `internal/errors` package
- Structured JSON output for AI-agent compatibility
- Pass `context.Context` as first parameter to all I/O operations
- Use `sync.Once` for lazy initialization
- Test files co-located: `package.go` → `package_test.go`

### Cross-Repo Considerations
- Auth changes must be implemented in BOTH gdrv and gpd
- Maintain consistency in command naming and flags
- Shared patterns should be documented in workspace AGENTS.md

### Documentation References
- API Coverage Matrix: `docs/api-coverage-matrix.md`
- ASC Parity Document: `docs/asc-parity.md`
- CLI Assessment: `docs/cli-assessment.md`

---

**Next Review Date:** TBD
**Plan Owner:** TBD
