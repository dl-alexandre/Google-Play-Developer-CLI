---
name: SharedOAuthClientPlan
overview: Finalize shared OAuth client onboarding with PKCE/public-client framing, explicit precedence and fallback rules, clear token storage defaults, and conditional gpd docs guidance.
todos:
  - id: gdrv-docs
    content: Update gdrv README with PKCE/public-client guidance and overrides
    status: pending
  - id: gpd-docs
    content: "Update gpd README: OAuth note or full OAuth docs (conditional)"
    status: pending
  - id: cli-messaging
    content: Unify OAuth credential messaging and add dev/CI enforcement
    status: pending
isProject: false
---

# Shared OAuth Client Plan

## Decisions (explicit)

- **Auth flow**: installed/desktop OAuth client with Authorization Code + PKCE (public client). Any bundled "secret" is not relied on for security and may be omitted depending on implementation constraints.
- **Credential precedence**: env vars → config file → bundled default client. **No partial overrides**: if any OAuth client env var is set, all required OAuth client vars must be set.
- **Config file paths** (defaults):
- Linux: `~/.config/gdrv/config.toml`
- macOS: `~/Library/Application Support/gdrv/config.toml`
- Windows: `%APPDATA%\gdrv\config.json`
- Permissions: `0600` (owner read/write only)
- **Token storage** (defaults):
- Default to file-based storage (later: OS credential store with file fallback)
- Linux: `~/.config/gdrv/tokens.json`
- macOS: `~/Library/Application Support/gdrv/tokens.json`
- Windows: `%APPDATA%\gdrv\tokens.json`
- Permissions: `0600` (owner read/write only, critical for token files)
- **Redirect behavior**: primary loopback (`127.0.0.1` + ephemeral port) with **manual copy/paste fallback**. Fallback triggers: cannot bind port, browser open fails, `--headless` / `--no-browser` or no GUI detected.
- **Operational posture**: host the shared client in a dedicated Google Cloud project with tight quota monitoring/alerts and a rotation plan.
- **Contributor/CI policy**: contributors and CI must use their own OAuth client to avoid accidental dependence on the shared one; enforce via a dev/CI flag.

## Scope and goals

- Make public-user onboarding in both CLIs explicit about the bundled OAuth client and how to override it.
- Keep contributor guidance separate: contributors/CI must use their own OAuth client credentials.
- Ensure CLI errors and help text clearly describe the decision tree.

## Proposed changes

### gdrv (Google-Drive-CLI)

- Documentation updates:
- [README.md](/Users/developer/Documents/GitHub/Google-Drive-CLI/README.md)
- Quick Start: "By default, `gdrv` uses the bundled OAuth client for personal use; to use your own client, set `GDRV_OAUTH_CLIENT_ID` (and `GDRV_OAUTH_CLIENT_SECRET` if required by the current implementation—treat it as public in distributed binaries)."
- Authentication: explicit note that for installed apps using PKCE, some Google Cloud consoles still ask for a 'secret,' but it is not a security boundary for distributed binaries.
- Add a **Contributors** note: must use personal OAuth clients and should not depend on the bundled credentials (enforce in CI).
- Add operational details: redirect behavior (primary + fallback), scopes requested by presets (least privilege + rationale), token storage mechanism/paths/permissions, multi-profile behavior, logout/revoke behavior.
- **Scopes source of truth**: Docs list scopes by referencing named presets defined in code; the CLI can print active scopes via `gdrv auth scopes`. (Even if command doesn't exist yet, establish that scopes must come from one enumerated place and docs mirror that list.)
- CLI messaging updates:
- Ensure a single error/hint is emitted when client credentials are missing or invalid.
- Include "bundled default" vs "override with env/config" in the hint.
- Likely touchpoints: [internal/cli/auth.go](/Users/developer/Documents/GitHub/Google-Drive-CLI/internal/cli/auth.go) and auth validation in `internal/auth`.

### gpd (Google-Play-Developer-CLI)

- **Gating step**: Before editing `gpd` docs, confirm an OAuth codepath is user-reachable (command or config flag). If not, add only the 'OAuth not used' clarification.
- Documentation updates (conditional based on actual OAuth support):
- [README.md](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/README.md)
- **Option A** (if OAuth is supported and user-reachable): add a concise OAuth section (when used, redirect behavior, scope expectations, and differences vs service-account auth), plus the same public-client notes and precedence rules.
- **Option B** (if OAuth is not supported): add a brief "OAuth not used; service accounts only" clarification to prevent confusion.
- Default to **Option B** unless proven otherwise.
- CLI messaging updates (only where OAuth is actually triggered):
- Surface the same decision tree and bundled-client guidance in auth errors/help.
- Likely touchpoint: [internal/cli/auth_commands.go](/Users/developer/Documents/GitHub/Google-Play-Developer-CLI/internal/cli/auth_commands.go)

## Security & quota notes (docs)

1. **Client secret is public**: Document that shipping a client secret means it is public and can be extracted; this is expected for PKCE/public clients.
2. **Least-privilege scopes**: Clarify scope presets and rationale (read-only vs read-write, minimal required permissions).
3. **Quota monitoring**: Add shared-client project notes with quota monitoring/alerts configuration.
4. **Rotation playbook**: Document rotation triggers (quota anomalies/abuse detected), how users are informed of rotation, how old versions fail with actionable guidance (error message includes upgrade instructions), and rotation process (rotate client, adjust scopes if needed, consider verification status).

## Implementation detail to enforce consistency

- Introduce/centralize a single "missing OAuth client credentials" error/hint used by all auth entrypoints (login/status/doctor).
- Add dev/CI enforcement mechanism: `GDRV_REQUIRE_CUSTOM_OAUTH=1` (explicit intent; good for CI)
- **Behavior**: When set, `auth login` refuses bundled credentials and errors with: "Custom OAuth client required; set env/config with GDRV_OAUTH_CLIENT_ID and GDRV_OAUTH_CLIENT_SECRET."
- This prevents accidental dependence on the shared client in contributor/CI flows.

## Acceptance Criteria

This plan is complete when all of the following are true:

1. **Docs clarity**: Both CLIs document PKCE/public-client model, bundled vs custom credentials, and exact config file paths with permissions.
2. **Centralized errors**: Single error message/hint for missing/invalid OAuth credentials used consistently across all auth entrypoints.
3. **Fallback works headless**: Manual copy/paste fallback triggers correctly when browser launch fails or `--headless` / `--no-browser` is used.
4. **Override precedence enforced**: Partial overrides (only ID or only secret) are rejected with clear error; full env override takes precedence over config/bundled.
5. **CI flag enforced**: `GDRV_REQUIRE_CUSTOM_OAUTH=1` refuses bundled credentials and errors with actionable message.

## Validation

- Fresh machine (no token cache): `auth login` uses bundled client.
- Override path: set `GDRV_OAUTH_CLIENT_ID/GDRV_OAUTH_CLIENT_SECRET` and confirm they take precedence.
- **Partial override rejection** (explicit negative tests):
- Set only `GDRV_OAUTH_CLIENT_ID` (no secret/config) → must fail with centralized message
- Set only `GDRV_OAUTH_CLIENT_SECRET` (no ID/config) → must fail with centralized message
- Expected error: "No partial overrides allowed; set all required OAuth client vars."
- Token refresh test (expired access token).
- Revoked consent behavior (re-auth guidance shown).
- Headless/device flow behavior when browser launch fails.
- Error copy: verify exact wording for missing credentials and invalid client.
- Dev/CI enforcement: set `GDRV_REQUIRE_CUSTOM_OAUTH=1` and confirm bundled credentials are refused.