# Miscellaneous Fixes & Polish Plan

**Status**: Not started
**Priority**: Low — quality-of-life improvements
**Date**: 2026-03-01

## Issues to Address

### 1. CSV Output Not Implemented

**Location**: Vitals commands accept `--format csv` but don't convert output.

- `kong_vitals.go` — all vitals commands
- Need a CSV formatter in `internal/output/` or within vitals command handlers
- Should also work for `analytics query` when implemented

### 2. `automation release-notes --from-prs` Stub

**Location**: `kong_automation.go:156`

- `generateFromPRs()` returns "PR-based release notes generation not yet implemented"
- Needs Git integration to read PR titles/descriptions since last release tag
- Lower priority — `--from-commits` and `--from-changelog` may cover most cases

### 3. `bulk upload` `uploadFile()` Stub

**Location**: `kong_bulk.go:~197`

- Returns placeholder data instead of actually uploading
- Wire to `Edits.Apks.Upload()` or `Edits.Bundles.Upload()` based on file extension
- Wire `--in-progress-review-behaviour` flag (currently unused)
- Covered in PLAN-bulk-operations.md but calling out here as a standalone quick fix

### 4. Documentation Accuracy

**Location**: `docs/api-coverage-matrix.md`

- Some stubs are marked ✅ (fully implemented) when they're not
- Needs an accuracy pass to match actual implementation status
- Also update `docs/asc-parity.md` and `docs/cli-assessment.md`

### 5. `--strict` Flag Unused in Testing Validate

**Location**: `kong_testing.go:227`

- `TestingValidateCmd` defines `--strict` but doesn't check it in `Run()`
- Either implement strict validation mode or remove the flag

### 6. Empty Placeholder Structs

**Location**: `kong_commands.go:6-29`

- `MigrateCmd`, `CustomAppCmd`, `GroupingCmd` — empty structs
- Either implement or remove to avoid dead code
- `CustomAppCmd` maps to Play Custom App Publishing API (niche)
- `MigrateCmd` — check if migration commands in `internal/migrate/` are wired up
- `GroupingCmd` — unclear purpose, investigate or remove

### 7. Consistent `--profile` Flag Wiring

Covered in PLAN-multi-account.md but noting here: only `kong_reviews.go` calls `SetActiveProfile(globals.Profile)`. This should be centralized.

## Implementation Order

These are independent and can be tackled in any order:

1. CSV output formatter (impacts multiple commands)
2. Documentation accuracy pass (no code changes)
3. `bulk upload` `uploadFile()` fix
4. Remove/implement dead placeholder structs
5. `--strict` flag fix
6. `--from-prs` automation
