# Authentication Parity Guide (ASC → gpd)

**Last verified:** 2026-07-17 against live `gpd auth --help` and generated [COMMANDS.md](COMMANDS.md).

This guide maps App Store Connect CLI auth workflows to `gpd`, documents real commands (not aspirational ones), and explains migration choices.

## Auth model differences

| | ASC | gpd |
| --- | --- | --- |
| Primary credential | App Store Connect API key (`.p8` + key ID + issuer ID) | Google **service account** JSON (recommended for CI) |
| Interactive human auth | Key registration; optional web session flows elsewhere | OAuth **device flow** when client ID is configured |
| Storage | System keychain with config fallback (`--bypass-keychain`) | Platform secure storage (`--store-tokens auto\|never\|secure`) |
| Profiles | Named keys (`--name`, `auth switch`) | Named profiles (`auth login <profile>`, `auth switch`, `--profile`, `GPD_AUTH_PROFILE`) |
| Diagnostics | `asc auth status --validate`, `asc auth doctor` / `asc doctor` | `gpd auth status`, `gpd auth doctor` / `diagnose`, `gpd config doctor` |
| Package access check | Network validate on login (`--network`) | `gpd auth check --package <pkg>` (edits insert/delete probe) |

`gpd` does **not** use ASC-style browser redirect login for Play automation. Prefer service accounts in CI.

## Implemented commands (source of truth)

```bash
gpd auth status                 # Auth status (profile, origin, token validity)
gpd auth login [profile]        # Authenticate (service account --key / env / ADC)
gpd auth init [profile]         # Alias of login (ASC naming parity)
gpd auth logout [--name name | --all]     # Clear stored credentials for active/named/all profiles
gpd auth delete <profile> [--force]       # Remove a profile (refuse active unless --force)
gpd auth list                   # List stored profiles + active marker
gpd auth switch <profile>       # Persist active profile to config
gpd auth check --package ...    # Validate Play API access for a package
gpd auth doctor                 # Sectioned diagnostics report
gpd auth diagnose               # Alias of doctor
```

Global flags that affect auth:

| Flag / env | Role |
| --- | --- |
| `--key-path` / service account path on login | Explicit key file |
| `--profile` | Override active profile for this invocation |
| `GPD_AUTH_PROFILE` | Env override for profile |
| `GPD_SERVICE_ACCOUNT_KEY` | Inline JSON key |
| `GOOGLE_APPLICATION_CREDENTIALS` | ADC key path |
| `GPD_CLIENT_ID` / `GPD_CLIENT_SECRET` | Device-flow OAuth (when used) |
| `--store-tokens` | `auto` (default), `never`, `secure` |

### Profile resolution order

1. `--profile` flag  
2. `GPD_AUTH_PROFILE`  
3. `activeProfile` in config file  
4. `default`

`gpd auth switch <profile>` and successful `auth login` / `auth init` persist `activeProfile` via `config.SetActiveProfile`.

## Command mapping (ASC → gpd)

| ASC | gpd | Notes |
| --- | --- | --- |
| `asc auth login --name ... --key-id ... --issuer-id ... --private-key ...` | `gpd auth login [profile] --key-path key.json` | Different credential material |
| `asc auth init` | `gpd auth init [profile]` | gpd init authenticates; ASC init scaffolds config template |
| `asc auth switch` | `gpd auth switch <profile>` | Implemented |
| `asc auth status` / `--validate` | `gpd auth status` | Use `auth check` for package validation |
| `asc auth doctor` / `asc doctor` | `gpd auth doctor` / `gpd auth diagnose` | Structured sections + summary; `--refresh-check`, `--network` |
| `asc auth logout` | `gpd auth logout [--name name \| --all]` | Clears stored tokens for the active profile by default |
| (profile delete) | `gpd auth delete <profile> [--force]` | Refuses the active profile unless `--force` (then switches to `default`) |
| (login `--network`) | `gpd auth check --package ...` | Explicit package permission probe |
| `asc auth login --bypass-keychain` | `--store-tokens never` or CI env without secure storage | Closest operational equivalent |

## Decision tree

1. **Unattended CI/CD?**  
   Use a service account with Android Publisher access.  
   `gpd --key-path /secrets/sa.json ...` or `GOOGLE_APPLICATION_CREDENTIALS`.  
   Prefer `--store-tokens never` in CI.

2. **Human ops / support on a laptop?**  
   `gpd auth login team-ops --key-path ./sa.json`  
   or device flow when OAuth client is configured.

3. **Multiple apps/teams?**  
   ```bash
   gpd auth login team-a --key-path ./team-a.json
   gpd auth login team-b --key-path ./team-b.json
   gpd auth switch team-a
   gpd --profile team-b publish status --package com.example.b
   ```

4. **Auth failures / invalid_grant (OAuth)?**  
   Re-login, revoke stale tokens, move OAuth consent to production if stuck in testing mode (7-day refresh tokens, 100-token cap).

## Doctor report shape

`gpd auth doctor` prints a JSON envelope (`data`) with:

- `sections[]` — Environment, Storage, Profiles, Credentials, optional Network, Runtime  
- `summary` — counts of ok / info / warn / fail  
- `recommendations[]`  
- `activeProfile`

Useful flags:

```bash
gpd auth doctor --refresh-check
gpd auth doctor --refresh-check --network --package com.example.app
gpd auth diagnose --refresh-check   # alias
```

`gpd config doctor` remains available for config-path and store-tokens issues; prefer **auth doctor** for credential health.

## Migration guidance for ASC users

### Service-account first (recommended)

1. Create a Google Cloud service account.  
2. Enable Android Publisher API (and other APIs you need).  
3. Invite the service account in Play Console with the right app permissions.  
4. ```bash
   gpd auth login ci --key-path ./sa.json
   gpd auth status
   gpd auth check --package com.example.app
   ```

### Profile hygiene

```bash
gpd auth list
gpd auth switch ci
gpd auth status --pretty
gpd auth logout --name staging      # clear one profile's stored tokens
gpd auth logout --all               # clear every profile
gpd auth delete staging             # remove a non-active profile
gpd auth delete ci --force          # delete active profile; switches to default
```

### Diagnostics before filing bugs

```bash
gpd auth doctor --refresh-check --output json --pretty
gpd config doctor --pretty
```

## Known boundaries

- No ASC-identical `.p8` / issuer workflow — different platforms.  
- No browser redirect login for Play automation.  
- `auth logout` clears stored tokens/metadata for the active profile (or `--name` / `--all`); `auth delete` removes a named profile and refuses the active profile unless `--force`.  
- OAuth device flow depends on `GPD_CLIENT_ID` (and optional secret) being configured.  
- Testing-mode OAuth apps: refresh tokens can expire after 7 days.

## Related

- [ASC Parity Matrix](asc-parity.md)  
- [ASC Workflow Mapping](asc-workflow-mapping.md)  
- Multi-account implementation notes: [PLAN-multi-account.md](PLAN-multi-account.md)
