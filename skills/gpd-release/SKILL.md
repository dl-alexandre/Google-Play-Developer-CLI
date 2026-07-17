---
name: gpd-release
description: >-
  Upload, validate, publish, promote, and stage rollouts for Google Play apps
  with gpd (Google Play Developer CLI). Use for release pipelines: validate
  readiness, publish play one-shot jobs, track status, staged rollouts, promote
  between tracks, halt/rollback. Prefer --dry-run before mutating production.
---

# gpd-release

End-to-end **Play release** workflows with **gpd**: validate → upload/assign track → status → rollout/promote/halt.

## When to use

- Shipping an APK/AAB to a track (`internal`, `alpha`, `beta`, `production`, or custom)
- Pre-publish readiness checks (`gpd validate`)
- One-shot “upload → track → status” jobs (`gpd publish play`)
- Staged rollouts, promotions, halt, or rollback
- CI/CD release steps that must stay non-interactive and JSON-parseable

For credentials first, use the **gpd-auth** skill.

## Safe defaults

| Practice | Guidance |
| --- | --- |
| Package | Always pass `-p` / `--package <applicationId>`. |
| Output | Prefer `--output json` (default in pipes/CI). Add `--pretty` only for humans. |
| Dry-run | Run mutating commands with `--dry-run` first (especially production). |
| Track | Start on `internal` / closed testing; promote later. |
| Production | Prefer staged `--percentage` before full release; use `--confirm` on halt/rollback. |
| Auth | Service account with Play access; see **gpd-auth**. |
| Source of truth | Prefer `gpd <cmd> --help` over memorized flags. Do not invent flags. |

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
    --key-path=STRING
    --profile=STRING
    --cache-dir=STRING       # $GPD_CACHE_DIR
```

## 1. Validate (pre-publish)

```bash
# Local / plan-only readiness (dry-run defaults to true)
gpd validate --package com.example.app --track internal --output json

# Include optional local artifact checks
gpd validate --package com.example.app --file ./app-release.aab --track internal --output json

# Treat warnings as failures
gpd validate --package com.example.app --track production --strict --output json

# Opt-in network package-access probe (needs credentials + --package)
gpd validate --package com.example.app --network --output json

# Explicit dry-run (default true) — plan without network side effects
gpd validate --package com.example.app --dry-run --output json
```

| Flag | Notes |
| --- | --- |
| `--track` | Default `internal` |
| `--file` | Optional APK/AAB path for local checks |
| `--strict` | Warnings → failures |
| `--dry-run` | Plan checks without network side effects (**default true**) |
| `--network` | Opt-in package access probe (requires `--package` + credentials) |

## 2. One-shot publish job (`publish play`)

High-level **upload → track → status** (ASC-style publish analogue).

```bash
# Plan only
gpd publish play ./app-release.aab \
  --package com.example.app \
  --track internal \
  --dry-run \
  --output json

# Upload to internal, completed assignment
gpd publish play ./app-release.aab \
  --package com.example.app \
  --track internal \
  --status completed \
  --output json

# Staged production rollout at 10%
gpd publish play ./app-release.aab \
  --package com.example.app \
  --track production \
  --percentage 10 \
  --output json
```

| Flag | Notes |
| --- | --- |
| `--track` | Default `internal` |
| `--percentage` | `0–100`. When `>0`, release is `inProgress` with userFraction; `0` uses `--status` |
| `--status` | Default `completed` (used when `--percentage` is `0`) |
| `--dry-run` | Plan without network side effects |

## 3. Granular publish steps

### Upload

```bash
gpd publish upload ./app-release.aab \
  --package com.example.app \
  --track internal \
  --dry-run \
  --output json

gpd publish upload ./app-release.aab \
  --package com.example.app \
  --track internal \
  --output json
```

Notable flags: `--track` (default `internal`), `--edit-id`, `--obb-main`, `--obb-patch`, `--obb-main-ref-version`, `--obb-patch-ref-version`, `--no-auto-commit`, `--in-progress-review-behaviour`, `--dry-run`.

### Create / update release

```bash
gpd publish release \
  --package com.example.app \
  --track internal \
  --status draft \
  --version-codes 123 \
  --dry-run \
  --output json

gpd publish release \
  --package com.example.app \
  --track production \
  --status inProgress \
  --version-codes 123 \
  --name "1.2.3" \
  --release-notes-file ./release-notes.json \
  --output json
```

Notable flags: `--track` (default `internal`), `--name`, `--status` (default `draft`), `--version-codes` (repeatable), `--retain-version-codes`, `--in-app-update-priority`, `--release-notes-file`, `--edit-id`, `--no-auto-commit`, `--in-progress-review-behaviour`, `--dry-run`, `--wait`, `--wait-timeout` (default `30m`).

> Use **`--version-codes`** (plural) on `publish release`. Do not invent `--version-code` for this command.

### Status & tracks

```bash
gpd publish tracks --package com.example.app --output json
gpd publish status --package com.example.app --output json
gpd publish status --package com.example.app --track production --output json
```

### Staged rollout

```bash
gpd publish rollout \
  --package com.example.app \
  --track production \
  --percentage 10 \
  --dry-run \
  --output json

gpd publish rollout \
  --package com.example.app \
  --track production \
  --percentage 50 \
  --output json
```

Notable flags: `--track` (default `production`), `--percentage` (`0.01–100.00`), `--edit-id`, `--no-auto-commit`, `--dry-run`.

### Promote between tracks

```bash
gpd publish promote \
  --package com.example.app \
  --from-track internal \
  --to-track production \
  --percentage 10 \
  --dry-run \
  --output json

gpd publish promote \
  --package com.example.app \
  --from-track beta \
  --to-track production \
  --output json
```

Notable flags: `--from-track`, `--to-track`, `--percentage` (default `0`), `--edit-id`, `--no-auto-commit`, `--dry-run`.

### Halt / rollback (destructive)

```bash
# Halt in-progress production rollout
gpd publish halt \
  --package com.example.app \
  --track production \
  --dry-run \
  --output json

gpd publish halt \
  --package com.example.app \
  --track production \
  --confirm \
  --output json

# Roll back to a prior version code
gpd publish rollback \
  --package com.example.app \
  --track production \
  --version-code 122 \
  --dry-run \
  --output json

gpd publish rollback \
  --package com.example.app \
  --track production \
  --version-code 122 \
  --confirm \
  --output json
```

Halt/rollback require `--confirm` for the real mutating path; always dry-run first.

## Recommended agent workflow

1. **Auth** — `gpd auth status` / `gpd auth check --package …` (see **gpd-auth**).
2. **Validate** — `gpd validate --package … --file ./app.aab --output json` (add `--network` when probing access).
3. **Dry-run publish** — `gpd publish play ./app.aab --package … --track internal --dry-run --output json`.
4. **Execute** — drop `--dry-run` for the intended track.
5. **Verify** — `gpd publish status --package … --track … --output json`.
6. **Widen** — `gpd publish rollout --percentage N` or `gpd publish promote …`.
7. **Incident** — halt with `--confirm`, or rollback with `--version-code` + `--confirm`.

## Exit codes (shared)

`0` success · `1` API · `2` Auth · `3` Permission · `4` Validation · `5` Rate limit · `6` Network · `7` Not found · `8` Conflict

## Related skills

- **gpd-auth** — service accounts, profiles, doctor/check
- **gpd-reviews-vitals** — post-release review replies and crash/ANR vitals

## Notes

- Prefer live `gpd publish --help` / `gpd validate --help` when flags change.
- JSON envelope shape: `{ "data": …, "error": …, "meta": … }`.
