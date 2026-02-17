# Authentication Parity Guide (ASC -> gpd)

This guide explains how App Store Connect (ASC) auth workflows map to `gpd`, where behavior differs, and what to do in migration scenarios.

## Auth Model Differences

- ASC commonly uses browser-based login sessions and key-based automation.
- `gpd` supports:
  - Service account credentials (recommended for CI/automation).
  - OAuth device flow (`gpd auth login`) for user-based interactive auth.
- `gpd` does not use browser redirect login in the same way as ASC CLI flows.

## Decision Tree

Use this to pick the right auth path.

1. Need unattended CI/CD automation?
   - Use service account auth (`--key` or `GOOGLE_APPLICATION_CREDENTIALS`).
2. Need human interactive access for support/ops workflows?
   - Use `gpd auth login` (device flow).
3. Need to manage multiple personas or environments?
   - Use profile-based auth (`gpd auth login <profile>`, `gpd auth switch <profile>`, `gpd auth list`).
4. Seeing refresh failures (`invalid_grant`) during OAuth usage?
   - Re-authenticate, revoke stale tokens, and verify OAuth consent screen mode.

## Command Mapping (ASC -> gpd)

- `asc auth login` -> `gpd auth login`
- `asc auth init` -> `gpd auth init`
- `asc auth switch` -> `gpd auth switch <profile>`
- `asc auth status` -> `gpd auth status`
- `asc auth doctor` -> `gpd auth diagnose` (or `gpd auth doctor`)
- `asc auth logout` -> `gpd auth logout`

## Migration Guidance for ASC Users

1. Service-account-first migration (recommended)
   - Provision a Google service account with Android Publisher permissions.
   - Prefer `gpd --key /path/key.json ...` in CI jobs.
   - Validate with `gpd auth status` and `gpd auth check --package <pkg>`.

2. User flow migration
   - Use `gpd auth login` to authorize via device flow.
   - Use profile names to separate teams/apps: `gpd auth login team-a`, `gpd auth switch team-a`.

3. Operational diagnostics
   - Use `gpd auth diagnose --refresh-check` to inspect token source, scopes, storage, and refresh outcomes.

## OAuth Testing-Mode Constraints

If your OAuth consent screen is in testing mode:

- Refresh tokens can expire after 7 days.
- Google enforces a 100 refresh-token issuance cap per OAuth client.

If you hit recurring `invalid_grant` errors:

1. Re-authenticate with `gpd auth login`.
2. Revoke stale tokens in Google Cloud Console.
3. Move OAuth consent screen to production for stable long-lived behavior.

## Known Boundaries

- No ASC-identical browser login UX; device flow and service accounts are the supported models.
- `gpd` auth behavior is optimized for explicit CLI/automation usage with JSON output and diagnosable failure modes.
