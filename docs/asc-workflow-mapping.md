# ASC Workflow Mapping (Submit, Builds, Analytics, Localization)

**Last verified:** 2026-07-17 against live `gpd --help`, `gpd validate`, and `gpd publish play`.

Maps common App Store Connect (`asc`) operational workflows to `gpd` sequences, including Google Play model differences.

See also: [ASC Parity Matrix](asc-parity.md), [Auth Parity Guide](auth-parity-guide.md).

---

## Mental model

| Concept | ASC | Google Play / gpd |
| --- | --- | --- |
| App identity | Numeric App Store Connect app ID (`--app`) | Package name (`--package` / `-p`) |
| Binary | Build → version → review | AAB/APK upload inside an **edit**, assigned to a **track** |
| Beta | TestFlight groups | Tracks: `internal` → `alpha` → `beta` → `production` |
| Staged release | Phased release on a version | Rollout **user fraction** on a track |
| Submit for review | `submit` / `publish appstore --submit` | Promoting/completing a track release (no Apple-style review queue) |
| High-level ship | `asc publish`, `asc release stage`, `asc status --watch` | Compose `publish *` or use `automation` / `workflow` |

---

## 1) Submit / release workflow

### ASC-style goal: upload, attach, submit, watch

ASC (illustrative):

```bash
asc publish appstore --app "123" --ipa app.ipa --version "1.2.3" --submit --confirm
asc status --app "123" --watch
asc validate --app "123" --version "1.2.3"
```

### gpd Play-native sequence

```bash
# Upload to internal track (creates/commits edit by default)
gpd publish upload ./app.aab --package com.example.app --track internal

# Create/update a release on a track
gpd publish release --package com.example.app --track internal --status draft

# Staged production rollout
gpd publish rollout --package com.example.app --track production --percentage 10

# Status
gpd publish status --package com.example.app --track production
gpd publish tracks --package com.example.app

# Halt / rollback / promote
gpd publish halt --package com.example.app --track production --confirm
gpd publish rollback --package com.example.app --track production --confirm
gpd publish promote --package com.example.app --from-track internal --to-track production
```

### Higher-level gpd helpers (ASC “job command” spirit)

```bash
# Readiness report (ASC validate analogue; dry-run by default)
gpd validate --package com.example.app --track internal --dry-run
gpd validate --package com.example.app --track production --file ./app.aab --dry-run

# Composed ship job (ASC publish analogue)
gpd publish play ./app.aab --package com.example.app --track internal --dry-run
gpd publish play ./app.aab --package com.example.app --track production --percentage 10

gpd automation validate --package com.example.app
gpd automation rollout --package com.example.app --track production
gpd automation promote --package com.example.app
gpd automation monitor --package com.example.app
gpd release-mgmt history --package com.example.app
gpd release-mgmt strategy --package com.example.app
```

### Declarative workflow (both CLIs)

```bash
gpd workflow validate --file ./workflows/production-release.json
gpd workflow run --file ./workflows/production-release.json
gpd workflow status
```

Example workflows live under `docs/examples/workflows/`.

### Decision paths (Play)

- Increase staged rollout when vitals/reviews look healthy (`publish rollout`, `automation rollout`).  
- Halt when regressions appear (`publish halt`).  
- Roll back if impact is severe (`publish rollback`).  
- Promote when lower tracks validate (`publish promote` / `automation promote`).

---

## 2) Builds & TestFlight → tracks & testers

### ASC

```bash
asc builds upload --app "123" --ipa app.ipa
asc builds list --app "123"
asc testflight groups list --app "123"
asc builds add-groups --app "123" --group "Internal Testers" ...
```

### gpd

```bash
gpd publish upload ./app.aab --package com.example.app --track internal
gpd publish builds list --package com.example.app
gpd publish builds get --package com.example.app
gpd publish testers list --package com.example.app
gpd publish testers add --package com.example.app --track internal --group "testers@example.com"
gpd publish beta-groups list --package com.example.app   # ASC-compat naming over tracks
```

**Model difference:** ASC attaches builds to TestFlight groups. Play grants testers access via **track** membership (and Google Groups). `beta-groups` commands are a compatibility veneer, not a second object model.

Internal sharing (no ASC twin):

```bash
gpd publish internal-share upload ./app.aab --package com.example.app
```

---

## 3) Analytics & crash feedback

### ASC

```bash
asc analytics sales ...
asc testflight feedback list --app "123"
asc testflight crashes list --app "123"
```

### gpd

```bash
gpd analytics query --package com.example.app ...
gpd analytics capabilities
gpd vitals crashes --package com.example.app
gpd vitals anrs --package com.example.app
gpd vitals errors issues --package com.example.app
gpd vitals anomalies list --package com.example.app
gpd monitor watch --package com.example.app
```

**Scope difference:** ASC sales/finance reports and TestFlight feedback are not the same APIs as Play Reporting + Android vitals. For metrics outside API coverage, use Play Console exports / BI.

---

## 4) Metadata & localization

### ASC

```bash
asc localizations list --app "123" --type app-info
asc metadata apply --app "123" --version "1.2.3" --dir ./metadata --dry-run
asc screenshots apply --app "123" --version "1.2.3" ...
```

### gpd

```bash
# Listing copy
gpd publish listing get --package com.example.app --language en-US
gpd publish listing update --package com.example.app --language en-US \
  --title "..." --short-description "..."

# Contact / app details
gpd publish details get --package com.example.app
gpd publish details update --package com.example.app --contact-email ops@example.com

# Images / assets
gpd publish images upload --package com.example.app --language en-US \
  --image-type phone-screenshots --file ./shot1.png
gpd publish images list --package com.example.app --language en-US
gpd publish assets upload --package com.example.app --dir ./store-assets
gpd publish assets spec

# Bulk locale updates
gpd bulk listings --package com.example.app --data-file ./listings.json
```

**Boundaries:** No ASC metadata directory sync or keyword audit. Pricing/availability and categories do not map 1:1.

---

## 5) Reviews (customer)

### ASC

```bash
asc reviews list --app "123"
asc reviews respond ...
```

### gpd

```bash
gpd reviews list --package com.example.app --min-rating 1
gpd reviews get --package com.example.app ...
gpd reviews reply --package com.example.app ...
gpd reviews response-get --package com.example.app ...
# response-delete exists but Play API may not support deletion
gpd reviews response-delete --package com.example.app ...
```

---

## 6) Auth for workflows (CI)

### ASC

```bash
asc auth login --bypass-keychain --name ci --key-id ... --issuer-id ... --private-key ./AuthKey.p8
asc auth status --validate
```

### gpd

```bash
gpd auth login ci --key-path ./sa.json
# or: export GOOGLE_APPLICATION_CREDENTIALS=./sa.json
gpd auth status
gpd auth check --package com.example.app
gpd auth doctor --refresh-check
```

In CI, prefer:

```bash
gpd --key-path "$SA_JSON" --store-tokens never publish upload ...
```

---

## 7) What not to force-map

These ASC workflows should stay Apple-side:

- Signing / devices / profiles / certificates  
- Xcode archive/export / Xcode Cloud  
- Sandbox testers, App Clips, App Events, alternative distribution  
- Apple Ads campaigns, ASC finance downloads  
- StoreKit retention messaging  

Play-native replacements live under vitals, monetization, integrity, recovery, and permissions — see the [parity matrix](asc-parity.md).

---

## Quick cheat sheet

| Job | Start here |
| --- | --- |
| Readiness | `gpd validate --package … --dry-run` |
| Ship a build | `gpd publish play app.aab --package … --track …` (or `publish upload`) |
| Staged production | `gpd publish play … --percentage N` / `publish rollout` |
| Stop a bad rollout | `gpd publish halt --confirm` |
| Beta testers | `gpd publish testers` / `beta-groups` |
| Crashes | `gpd vitals crashes` / `errors` |
| Store copy | `gpd publish listing` / `images` / `assets` |
| Reviews | `gpd reviews list` / `reply` |
| Multi-step CI | `gpd workflow run` or `gpd automation *` |
| Auth health | `gpd auth doctor --refresh-check` |
