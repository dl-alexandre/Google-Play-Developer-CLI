---
name: gpd-auth
description: >-
  Authenticate and manage gpd (Google Play Developer CLI) credentials —
  service accounts, named profiles, auth doctor/check/status, list/switch/delete/logout.
  Use when setting up CI keys, diagnosing invalid credentials, switching teams/apps,
  or verifying package access before publish/reviews/vitals calls.
---

# gpd-auth

Service-account and multi-profile authentication for **gpd**.

## When to use

- First-time setup or CI credential wiring
- “Permission denied” / auth failures before API calls
- Multiple service accounts (teams, apps, environments)
- Verifying the active profile can reach a package
- Clearing or deleting stored profiles

Do **not** invent OAuth browser flows for automation. Prefer service accounts.

## Safe defaults

| Practice | Guidance |
| --- | --- |
| Output | Prefer `--output json` (default in pipes/CI). Use `--pretty` only for humans. |
| Package | Always pass `--package <appId>` for `auth check` and for doctor network probes. |
| CI storage | Use `--store-tokens never` (or env without secure storage) so keys are not written to keychain. |
| Key material | Prefer `--key-path` / `GOOGLE_APPLICATION_CREDENTIALS` / `GPD_SERVICE_ACCOUNT_KEY` over interactive login. |
| Destructive | Prefer `auth logout --name …` before `auth delete`; use `--force` only when deleting the active profile. |
| Source of truth | Run `gpd auth <cmd> --help` if unsure; do not invent flags. |

### Global flags (all commands)

```
-p, --package=STRING
    --output="json"          # json|table|markdown|csv|excel
    --pretty
    --timeout=30s
    --store-tokens="auto"    # auto|never|secure
    --fields=STRING
    --quiet
-v, --verbose
    --key-path=STRING        # service account key file
    --profile=STRING         # profile for this invocation
    --cache-dir=STRING       # $GPD_CACHE_DIR
```

### Profile resolution order

1. `--profile`
2. `GPD_AUTH_PROFILE`
3. config `activeProfile`
4. `default`

### Useful env vars

| Env | Role |
| --- | --- |
| `GPD_AUTH_PROFILE` | Profile override |
| `GPD_SERVICE_ACCOUNT_KEY` | Inline JSON key |
| `GOOGLE_APPLICATION_CREDENTIALS` | ADC key path |
| `GPD_CLIENT_ID` / `GPD_CLIENT_SECRET` | Device-flow OAuth (optional) |

## Commands (exact help surface)

### Status / list / switch

```bash
# Check whether credentials load and look valid
gpd auth status --output json

# List stored profiles (active marker in output)
gpd auth list --output json

# Make a profile active for subsequent commands
gpd auth switch <profile> --output json

# One-off profile without switching active
gpd --profile team-b auth status --output json
```

### Login / init (service account)

`auth init` is an alias of `auth login`.

```bash
# Store credentials under a named profile
gpd auth login ci --key ./sa.json --output json
# equivalent:
gpd auth login ci --key-path ./sa.json --output json

gpd auth init ci --key ./sa.json --output json

# CI: authenticate without persisting tokens to secure storage
gpd auth login ci --key ./sa.json --store-tokens never --output json

# Device-flow OAuth only when client ID/secret are configured (not preferred for CI)
gpd auth login ops --output json
```

Command-specific flags on login/init:

| Flag | Meaning |
| --- | --- |
| `--key` | Path to service account key file (command-local; `--key-path` also works globally) |

### Package permission check

Requires global `--package`.

```bash
gpd auth check --package com.example.app --output json
gpd --profile ci auth check --package com.example.app --output json
```

### Doctor / diagnose

`auth diagnose` is an alias of `auth doctor`.

```bash
# Local diagnostics (no network probe)
gpd auth doctor --output json

# Attempt credential/token load refresh path
gpd auth doctor --refresh-check --output json

# Network package probe (requires --package + credentials)
gpd auth doctor --refresh-check --network --package com.example.app --output json

gpd auth diagnose --refresh-check --output json
```

Command-specific flags:

| Flag | Meaning |
| --- | --- |
| `--refresh-check` | Attempt token refresh / credential load |
| `--network` | Lightweight network permission probe (requires `--package`) |

Also available (config-level): `gpd config doctor` — prefer **auth doctor** for credential health.

### Logout / delete

```bash
# Sign out active profile
gpd auth logout --output json

# Sign out a named profile
gpd auth logout --name ci --output json

# Sign out all profiles
gpd auth logout --all --output json

# Delete a stored profile (refuses active unless --force)
gpd auth delete staging --output json
gpd auth delete staging --force --output json
```

| Command | Notable flags / args |
| --- | --- |
| `auth logout` | `--name=STRING` (default: active), `--all` |
| `auth delete <profile>` | `--force` (allow deleting active; switches to `default`) |

## Recommended agent workflow

1. `gpd auth status --output json` — is anything loaded?
2. If empty / broken: `gpd auth login <profile> --key ./sa.json --output json`
3. Multi-account: `gpd auth list` → `gpd auth switch <profile>` (or `--profile` per call)
4. Before write ops: `gpd auth check --package com.example.app --output json`
5. On mystery failures: `gpd auth doctor --refresh-check --network --package com.example.app --output json`
6. Cleanup: `gpd auth logout --name <profile>` or `gpd auth delete <profile>`

## Exit codes (shared)

`0` success · `1` API · `2` Auth · `3` Permission · `4` Validation · `5` Rate limit · `6` Network · `7` Not found · `8` Conflict

## Related skills

- **gpd-release** — validate / upload / publish play / rollout after auth works
- **gpd-reviews-vitals** — reviews and Android vitals queries

## Notes

- Prefer live `gpd auth --help` / `gpd auth <cmd> --help` over memorized flags.
- Play Console must grant the service account access to the app; API enablement alone is not enough.
- For broader operator docs see `docs/auth-parity-guide.md` in the gpd repo.
