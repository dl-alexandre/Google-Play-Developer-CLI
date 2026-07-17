# Multi-Account / Profile Support Implementation Plan

**Status**: Phase 1–2 largely complete (2026-07-17) — CLI surface + profile resolution shipped  
**Priority**: Medium (remaining polish)  
**Date**: 2026-03-01 (updated 2026-07-17)

## Background

The backend infrastructure for multi-account support already existed; the CLI surface was incomplete and parity docs overstated capabilities. As of 2026-07-17 the core profile commands and global resolution are implemented.

### What Already Works

| Component | Location | Status |
|-----------|----------|--------|
| Profile-keyed token storage | `internal/auth/token_storage.go` | Working |
| `tokenStorageKey("{profile}--{hash}")` | `token_storage.go` | Working |
| `Manager.ListProfiles()` | `token_storage.go` | Working |
| `SetActiveProfile()` / `GetActiveProfile()` | `internal/auth/auth.go` | Working |
| `--profile` global CLI flag | `internal/cli/kong_cli.go` | Wired via `ResolveAuthProfile` + `applyAuthGlobals` |
| `GPD_AUTH_PROFILE` env var | `internal/config` | Read in `ResolveAuthProfile` |
| `activeProfile` in config file | `internal/config/config.go` | Read + written by `SetActiveProfile` |
| Token metadata files (`.meta.json`) | `internal/auth/token_storage.go` | Working |
| Platform-specific secure storage | `internal/storage/` | Working |
| `auth list/switch/init/login/check/doctor/diagnose` | `internal/cli/kong_auth.go` | Implemented |

### Remaining gaps

1. **Profile deletion/cleanup** — no `auth delete` yet  
2. **Logout profile targeting** — logout is not yet fully profile-file-aware (`--all` / secure-storage wipe)  
3. **Per-command explicit profile on every API path** — globals are resolved centrally; some older helpers still construct managers independently (should keep using `newAuthManager()`)

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
