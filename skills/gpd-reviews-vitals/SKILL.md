---
name: gpd-reviews-vitals
description: >-
  List/reply to Google Play user reviews and query Android Vitals (crashes, ANRs,
  error issues/reports, metrics) via gpd. Use for support triage, review response
  automation, and post-release stability monitoring. Always pass --package and
  prefer JSON output; dry-run replies before sending.
---

# gpd-reviews-vitals

**Reviews** and **Android Vitals** operations with **gpd**.

## When to use

- Listing or filtering user reviews (rating, language, date)
- Drafting / sending developer replies (single or batched)
- Checking crash rate, ANR rate, or other vitals after a release
- Searching error issues/reports or listing anomalies
- Support / quality triage agents that need structured JSON

Auth first with **gpd-auth**. For shipping builds use **gpd-release**.

## Safe defaults

| Practice | Guidance |
| --- | --- |
| Package | Always pass `-p` / `--package <applicationId>`. |
| Output | Prefer `--output json` (default in pipes/CI). Use `--pretty` for humans. |
| Replies | Always `--dry-run` before live `reviews reply`; respect `--max-actions` / `--rate-limit`. |
| Scope | Start with small `--page-size` / date windows; use `--all` only when needed. |
| PII | Review text may contain personal data — avoid logging full payloads in shared traces. |
| Source of truth | Prefer `gpd reviews --help` / `gpd vitals --help`. Do not invent flags. |

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

---

## Reviews

### List

```bash
# Recent / filtered reviews as JSON
gpd reviews list \
  --package com.example.app \
  --min-rating 1 \
  --max-rating 2 \
  --page-size 50 \
  --output json

# Include body text and optional translation
gpd reviews list \
  --package com.example.app \
  --include-review-text \
  --translation-language en \
  --language en \
  --start-date 2026-01-01 \
  --end-date 2026-01-31 \
  --scan-limit 100 \
  --output json --pretty

# Paginate or fetch all pages
gpd reviews list --package com.example.app --page-token TOKEN --output json
gpd reviews list --package com.example.app --all --output json
```

| Flag | Notes |
| --- | --- |
| `--min-rating` / `--max-rating` | 1–5 |
| `--language` | Filter by review language |
| `--start-date` / `--end-date` | ISO 8601 |
| `--scan-limit` | Default `100` |
| `--include-review-text` | Include review body |
| `--translation-language` | Translate reviews |
| `--page-size` | Default `50` |
| `--page-token` | Pagination |
| `--all` | Fetch all pages |

### Get one review

```bash
gpd reviews get REVIEW_ID \
  --package com.example.app \
  --include-review-text \
  --output json
```

Optional: `--translation-language`.

### Reply (mutating)

```bash
# Plan only
gpd reviews reply REVIEW_ID \
  --package com.example.app \
  --text "Thanks for the feedback — we're looking into this." \
  --dry-run \
  --output json

# Send reply
gpd reviews reply REVIEW_ID \
  --package com.example.app \
  --text "Thanks for the feedback — we're looking into this." \
  --output json

# Template-driven / batch-friendly
gpd reviews reply \
  --package com.example.app \
  --template-file ./reply.txt \
  --max-actions 10 \
  --rate-limit 5s \
  --dry-run \
  --output json
```

| Flag | Notes |
| --- | --- |
| `--text` | Reply body |
| `--template-file` | Template file for reply |
| `--max-actions` | Default `10` — max replies per execution |
| `--rate-limit` | Default `5s` between replies |
| `--dry-run` | Show intended actions without executing |

### Response get / delete

```bash
gpd reviews response-get REVIEW_ID --package com.example.app --output json
gpd reviews response-delete REVIEW_ID --package com.example.app --output json
```

`response-delete` is destructive; confirm intent with the operator before running.

---

## Vitals

### Crashes & ANRs

```bash
gpd vitals crashes \
  --package com.example.app \
  --start-date 2026-01-01 \
  --end-date 2026-01-31 \
  --output json

gpd vitals anrs \
  --package com.example.app \
  --start-date 2026-01-01 \
  --end-date 2026-01-31 \
  --output json

# Optional grouping / pagination
gpd vitals crashes \
  --package com.example.app \
  --dimensions apiLevel,versionCode \
  --page-size 100 \
  --output json
```

Shared flags for `crashes` / `anrs`:

| Flag | Notes |
| --- | --- |
| `--start-date` / `--end-date` | ISO 8601 |
| `--dimensions` | Repeatable grouping dimensions |
| `--format` | `json` (default) or `csv` — command-local format |
| `--page-size` | Default `100` |
| `--page-token` | Pagination |
| `--all` | Fetch all pages |

> Prefer global `--output json` for the standard result envelope. Command `--format` is for vitals export shape (`json`/`csv`).

### Generic vitals query

```bash
gpd vitals query \
  --package com.example.app \
  --metrics crashRate,anrRate \
  --start-date 2026-01-01 \
  --end-date 2026-01-31 \
  --output json

gpd vitals capabilities --output json
```

`vitals query` flags: `--start-date`, `--end-date`, `--metrics` (default includes `crashRate`), `--dimensions`, `--format`, `--page-size`, `--page-token`, `--all`.

### Errors & other metrics (overview)

```bash
gpd vitals errors issues --package com.example.app --output json
gpd vitals errors reports --package com.example.app --output json
gpd vitals errors counts get --package com.example.app --output json
gpd vitals errors counts query --package com.example.app --output json

gpd vitals metrics excessive-wakeups --package com.example.app --output json
gpd vitals metrics slow-rendering --package com.example.app --output json
gpd vitals metrics slow-start --package com.example.app --output json
gpd vitals metrics stuck-wakelocks --package com.example.app --output json

gpd vitals anomalies list --package com.example.app --output json
```

Run `gpd vitals <cmd> --help` for command-specific filters before automating.

---

## Recommended agent workflows

### Support triage

1. `gpd reviews list --package … --min-rating 1 --max-rating 2 --include-review-text --output json`
2. `gpd reviews get <id> --package … --include-review-text --output json`
3. Draft reply → `gpd reviews reply <id> --package … --text "…" --dry-run --output json`
4. On approval → same command without `--dry-run`

### Post-release stability

1. Confirm package auth (`gpd auth check --package …`).
2. `gpd vitals crashes` / `gpd vitals anrs` for the rollout window.
3. Optionally `gpd vitals query --metrics crashRate,anrRate …`.
4. If rates spike, use **gpd-release** to halt/rollback.

## Exit codes (shared)

`0` success · `1` API · `2` Auth · `3` Permission · `4` Validation · `5` Rate limit · `6` Network · `7` Not found · `8` Conflict

## Related skills

- **gpd-auth** — credentials, profiles, doctor/check
- **gpd-release** — validate / publish / rollout / halt

## Notes

- Prefer live `gpd reviews --help` and `gpd vitals --help`.
- JSON envelope: `{ "data": …, "error": …, "meta": … }`.
