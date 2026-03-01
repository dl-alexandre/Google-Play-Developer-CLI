# Multi-Account / Profile Support Implementation Plan

**Status**: Blocked — waiting on asccli multi-account bugfixes to land first
**Priority**: High
**Date**: 2026-03-01

## Background

The backend infrastructure for multi-account support already exists but the CLI surface is incomplete. Users currently can't switch between accounts without manually swapping `--key` flags or environment variables.

### What Already Works

| Component | Location | Status |
|-----------|----------|--------|
| Profile-keyed token storage | `internal/auth/token_storage.go` | Working |
| `tokenStorageKey("{profile}--{hash}")` | `token_storage.go:112-116` | Working |
| `Manager.ListProfiles()` | `token_storage.go:69` | Working |
| `SetActiveProfile()` / `GetActiveProfile()` | `internal/auth/auth.go` | Working |
| `--profile` global CLI flag | `internal/cli/kong_cli.go:27` | Exists (not wired) |
| `GPD_AUTH_PROFILE` env var | `internal/config/` | Defined (not read) |
| `activeProfile` in config file | `internal/config/config.go` | Exists |
| Token metadata files (`.meta.json`) | `internal/auth/token_storage.go` | Working |
| Platform-specific secure storage | `internal/storage/` | Working |

### What's Missing

1. **CLI commands**: `auth switch`, `auth list`, `auth init` — documented in parity guides but not implemented
2. **`--profile` flag propagation** — only `kong_reviews.go` calls `SetActiveProfile(globals.Profile)`; other commands ignore it
3. **`GPD_AUTH_PROFILE` env var** — `GetEnvAuthProfile()` exists but is never called during CLI init
4. **Profile deletion/cleanup** — no mechanism to remove a stored profile
5. **Active profile persistence** — no way to remember which profile was last used across invocations

## Implementation Steps

### Phase 1: Wire Up Existing Infrastructure

**Goal**: Make `--profile` and `GPD_AUTH_PROFILE` actually work everywhere.

1. **Centralize profile resolution in CLI initialization**
   - In the Kong CLI setup (after parsing globals), resolve the active profile from:
     1. `--profile` flag (highest priority)
     2. `GPD_AUTH_PROFILE` env var
     3. `config.ActiveProfile` from config file
     4. `"default"` fallback
   - Call `authMgr.SetActiveProfile()` once, centrally — remove the one-off call in `kong_reviews.go`

2. **Persist active profile in config**
   - When a profile is explicitly set via `auth switch`, write it to `config.ActiveProfile`
   - On next invocation, load it as the default (overridable by flag/env)

### Phase 2: Implement Auth Profile Commands

3. **`gpd auth list`**
   - Call `Manager.ListProfiles()` to enumerate stored profiles
   - Show: profile name, email (from token metadata), origin (service-account vs oauth), last used timestamp
   - Mark the active profile with an indicator
   - Output as standard JSON envelope

4. **`gpd auth init <profile>`**
   - Create a named profile and run the auth flow for it
   - Accept `--key` for service account profiles
   - Store credentials under the profile name in token storage
   - Set as active profile if `--set-active` flag is passed (or if it's the first profile)

5. **`gpd auth switch <profile>`**
   - Validate profile exists (via `ListProfiles()`)
   - Update `config.ActiveProfile`
   - Verify the stored token is still valid (refresh if needed)
   - Output confirmation with profile details

6. **`gpd auth delete <profile>`**
   - Remove token + metadata files for the profile
   - Remove from secure storage (keychain)
   - Refuse to delete the currently active profile unless `--force`
   - If deleting the active profile with `--force`, switch to "default"

### Phase 3: Polish

7. **`gpd auth status` enhancements**
   - Show current profile name in status output
   - Show total number of stored profiles
   - Add `--all` flag to show status for every profile

8. **`gpd auth logout` profile awareness**
   - Logout from the current profile only (not all)
   - Add `--all` flag to clear all profiles
   - Add `--profile <name>` to target a specific profile

9. **Documentation**
   - Update CLI help text for all auth commands
   - Add examples to `docs/examples/`
   - Update `docs/auth-parity-guide.md` to mark features as implemented

## Dependencies

- **asccli**: Multi-account bugfixes need to land first so we can align behavior and avoid repeating the same issues
- **No new Go dependencies needed** — all infrastructure already exists

## Testing

- Unit tests for profile resolution priority chain (flag > env > config > default)
- Unit tests for each new auth subcommand
- Integration test: create profile A, create profile B, switch between them, verify correct credentials used
- Edge cases: delete active profile, switch to nonexistent profile, concurrent profile access

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_auth.go` | Add `auth list`, `auth init`, `auth switch`, `auth delete` commands |
| `internal/cli/kong_cli.go` | Centralize profile resolution in CLI init |
| `internal/cli/kong_reviews.go` | Remove one-off `SetActiveProfile` call |
| `internal/auth/auth.go` | Add `DeleteProfile()` method, enhance `Authenticate()` to use active profile |
| `internal/auth/token_storage.go` | Add `DeleteProfileTokens()`, enhance `ListProfiles()` with metadata |
| `internal/config/config.go` | Wire `GPD_AUTH_PROFILE` reading, add `SetActiveProfile()` persistence |
| `internal/storage/` | Add `DeleteByPrefix()` for profile cleanup |
