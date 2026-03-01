# Publish Release Workflow Implementation Plan

**Status**: Not started
**Priority**: High — unlocks core release lifecycle beyond upload
**Date**: 2026-03-01

## Background

Users can upload APKs/AABs and create releases, but cannot manage the release lifecycle after that. The 5 missing commands (`rollout`, `promote`, `halt`, `rollback`, `status`) are the most impactful gap in the CLI.

All commands live in `internal/cli/kong_publish.go` as stubs.

## Commands to Implement

### 1. `gpd publish status`
**Line**: `kong_publish.go:727`
**Purpose**: Get current track status (release info, version codes, rollout fraction)

- Accepts `--track` (production, beta, alpha, internal) or `--all`
- Uses `Edits.Tracks.Get()` or `Edits.Tracks.List()`
- Shows: track name, release status, version codes, rollout percentage, in-app update priority
- No edit commit needed (read-only)

### 2. `gpd publish rollout`
**Line**: `kong_publish.go:673`
**Purpose**: Update rollout percentage for a staged release

- Accepts `--track` and `--fraction` (0.0–1.0)
- Uses `Edits.Tracks.Update()` to modify `userFraction` on the `inProgress` release
- Validate: release must be in `inProgress` status
- Requires edit commit

### 3. `gpd publish promote`
**Line**: `kong_publish.go:688`
**Purpose**: Promote a release from one track to another

- Accepts `--from-track` and `--to-track`
- Optionally `--fraction` for staged rollout on the target track
- Get release from source track, create/update release on target track
- Requires edit commit

### 4. `gpd publish halt`
**Line**: `kong_publish.go:702`
**Purpose**: Halt a staged rollout (pause at current percentage)

- Accepts `--track`
- Uses `Edits.Tracks.Update()` to set release status to `halted`
- Validate: release must be in `inProgress` status
- Requires edit commit

### 5. `gpd publish rollback`
**Line**: `kong_publish.go:717`
**Purpose**: Resume the previous release (effectively rolling back)

- Accepts `--track`
- Google Play doesn't have a true "rollback" — this halts the current release, which resumes serving the previous version to users not in the rollout
- Should document this behavior clearly in help text
- Requires edit commit

### 6. `gpd publish capabilities`
**Line**: `kong_publish.go:837`
**Purpose**: List available publish capabilities for the authenticated account

- Read-only, no edit needed
- Show available tracks, supported artifact types, listing locales

## Shared Concerns

- All mutating commands need edit transaction handling (create edit, operate, commit)
- Respect `--edit-id` and `--no-auto-commit` flags already defined on the parent
- Use `edits.Manager` for lock/idempotency
- Return structured errors for: no active rollout, wrong release status, track not found

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_publish.go` | Implement `Run()` for all 6 commands |
| `internal/api/client.go` | May need helper methods for track operations |

## Testing

- Mock `Edits.Tracks.Get/Update/List` responses
- Test state transitions: inProgress → halted, halted → inProgress
- Test validation: rollout on non-staged release, promote to same track
- Test edit transaction lifecycle for each command
