---
name: AuthRefreshMitigations
overview: Add centralized auth error classification, remediation output, an auth diagnose command, and stabilize token storage by profile+client hash in both gdrv and gpd, plus doc updates for testing-mode expiry and refresh-token caps.
todos:
  - id: classify-auth
    content: Add auth error classifiers and remediation output in gdrv+gpd
    status: pending
  - id: diagnose-cmd
    content: Add auth diagnose commands with refresh-check/json
    status: pending
  - id: stable-storage
    content: Make token storage profile+client-stable with metadata
    status: pending
  - id: docs-update
    content: Update docs for testing-mode expiry and token cap
    status: pending
isProject: false
---

# Auth Refresh Mitigations Plan

## Scope

Implement the requested mitigations in both repos:

- Google Drive CLI (`gdrv`): [internal/auth/manager.go](/Users/developer/Documents/GitHub/Google-Drive-CLI/internal/auth/manager.go), [internal/cli/auth.go](/Users/developer/Documents/GitHub/Google-Drive-CLI/internal/cli/auth.go), auth storage in [internal/auth/storage.go](/Users/developer/Documents/GitHub/Google-Drive-CLI/internal/auth/storage.go), and error classification in [internal/errors/google_api.go](/Users/developer/Documents/GitHub/Google-Drive-CLI/internal/errors/google_api.go).
- Google Play Developer CLI (`gpd`): auth flow in [internal/auth/auth.go](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/internal/auth/auth.go), auth commands in [internal/cli/auth_commands.go](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/internal/cli/auth_commands.go), error codes in [internal/errors/codes.go](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/internal/errors/codes.go), and config doctor reference in [internal/cli/config_commands.go](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/internal/cli/config_commands.go).

## Plan

- Add a shared auth-error classifier per repo that normalizes:
- OAuth refresh failures (e.g. `invalid_grant`, `invalid_client`, `unauthorized_client`).
- API failures (401/403) into consistent remediation output.
- Optional clock-skew detection using response `Date` headers when present.
- Wire the classifier into the refresh path so all commands surface the same focused remediation text.
- Implement `auth diagnose` in both CLIs that prints:
- Active profile, token location, client-id hash fingerprint, scopes, token timestamps, and refresh-token presence.
- Optional `--refresh-check` to attempt a refresh and reuse the classifier output on failure.
- Optional `--json` output for automation.
- Stabilize token storage paths to include both profile and client-id hash, and add a small metadata file alongside stored tokens to detect client mismatches.
- Update docs in both repos to note testing-mode refresh token expiry (7 days) and the documented refresh-token issuance cap (100), replacing older 50 references if present.

## Notes on Implementation€

- ™In `gdrv`, use the existing auth manager refresh flow in [internal/auth/manager.go](/Users/developer/Documents/GitHub/Google-Drive-CLI/internal/auth/manager.go) and CLI auth commands in [internal/cli/auth.go](/Users/developer/Documents/GitHub/Google-Drive-CLI/internal/cli/auth.go) as the integration points.€
- In `gpd`, integrate classification at the token refresh boundary in [internal/auth/auth.go](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/internal/auth/auth.go) and add the command under [internal/cli/auth_commands.go](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/internal/cli/auth_commands.go).

## Validation

- Add/adjust unit tests around new classifier behavior and diagnose output formatting.
- Manual: run `gdrv auth diagnose --refresh-check` and `gpd auth diagnose --refresh-check` with valid and invalid creds to confirm remediation output.