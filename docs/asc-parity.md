# App Store Connect CLI Parity Matrix

**Last verified:** 2026-07-17 against:

- live `gpd --help` / `gpd auth --help` / `gpd publish --help` / `gpd validate --help`
- generated [COMMANDS.md](COMMANDS.md) (`make generate-command-docs`)
- [rorkai/App-Store-Connect-CLI](https://github.com/rorkai/App-Store-Connect-CLI) `docs/COMMANDS.md` (reference)

This matrix maps App Store Connect CLI (`asc`) feature groups to `gpd` equivalents. It is intentionally **honest about model differences** and about what is implemented versus platform-impossible.

Related docs:

- [Auth Parity Guide](auth-parity-guide.md)
- [ASC Workflow Mapping](asc-workflow-mapping.md)
- [Command taxonomy (generated)](COMMANDS.md)
- [API Coverage Matrix](api-coverage-matrix.md)

## Status meanings

| Status | Meaning |
| --- | --- |
| **Full** | Comparable capability for day-to-day automation |
| **Partial** | Similar goals with platform or product gaps |
| **Not applicable** | No Google Play equivalent (Apple-only) |
| **gpd-only** | Play capability with no ASC equivalent |

## CLI contract (aligned where it matters)

| Topic | ASC | gpd |
| --- | --- | --- |
| Framework | `ffcli`, domain packages | Kong (`internal/cli/kong_*.go`) |
| Default `--output` | TTY-aware: `table` / pipe `json` | **TTY-aware:** `table` on TTY, `json` in pipes/CI; `GPD_DEFAULT_OUTPUT` + explicit `--output` win |
| Destructive ops | `--confirm` | `--confirm` / `--dry-run` on mutating commands |
| Auth | `.p8` API key + profiles + doctor | Service account / ADC / device flow + profiles + doctor |
| High-level ship | `asc publish`, `asc validate`, `asc status` | `gpd publish play`, `gpd validate`, `gpd publish status` |
| Command docs | generated `docs/COMMANDS.md` | generated `docs/COMMANDS.md` (`make generate-command-docs` / `make check-docs`) |
| Install trust | checksum-verified install script | checksum-verified `install.sh` (`GPD_INSTALL_INSECURE=1` opt-out) |

---

## Parity matrix

### Getting started & auth

| ASC Feature Group | ASC Commands (examples) | gpd Equivalent | Status |
| --- | --- | --- | --- |
| Authentication | `asc auth login/init/switch/status/doctor/logout` | `gpd auth login/init/switch/list/status/check/doctor/diagnose/logout` ([Auth Parity Guide](auth-parity-guide.md)) | **Partial** — multi-profile CLI implemented; credentials are service-account/ADC/device-flow, not ASC `.p8` |
| Config / doctor | `asc doctor`, `asc init`, `asc docs` | `gpd config *`, `gpd auth doctor` | **Partial** — no embedded docs browser / install-skills |
| Output formats | `--output table\|json\|markdown`, TTY defaults | `--output json\|table\|markdown\|csv\|excel`, TTY defaults, `GPD_DEFAULT_OUTPUT` | **Full** for day-to-day use |
| Version / completion | `asc version`, `asc completion` | `gpd version`, `gpd completion`, `gpd check-update` | **Full** |

### Apps, builds, distribution

| ASC Feature Group | ASC Commands (examples) | gpd Equivalent | Status |
| --- | --- | --- | --- |
| Apps | `asc apps list/get` | `gpd apps list`, `gpd apps get` | **Partial** |
| Builds | `asc builds upload/list/...` | `gpd publish upload`, `gpd publish builds *`, `gpd publish play` | **Partial** — track/edit model, not global build registry |
| TestFlight | `asc testflight`, feedback, crashes | `gpd vitals *`, `gpd publish testers`, `gpd publish beta-groups` | **Partial** |
| Beta groups / testers | `asc beta-groups`, `asc beta-testers` | `gpd publish beta-groups *`, `gpd publish testers *` | **Partial** |
| Internal sharing | N/A | `gpd publish internal-share upload` | **gpd-only** |
| Deobfuscation | N/A | `gpd publish deobfuscation upload` | **gpd-only** |
| Devices / signing / Xcode / sandbox | `asc devices`, certificates, xcode-cloud, sandbox | N/A | **Not applicable** |

### Metadata & media

| ASC Feature Group | ASC Commands (examples) | gpd Equivalent | Status |
| --- | --- | --- | --- |
| Listing / localizations | `asc localizations`, `asc metadata` | `gpd publish listing *`, `gpd publish details *` | **Partial** |
| Screenshots / assets | `asc screenshots` | `gpd publish images *`, `gpd publish assets *` | **Partial** |
| App tags / events / clips / categories / pre-orders | ASC-only | N/A | **Not applicable** |

### Review & release

| ASC Feature Group | ASC Commands (examples) | gpd Equivalent | Status |
| --- | --- | --- | --- |
| Customer reviews | `asc reviews` | `gpd reviews list/get/reply/response-*` | **Partial** — response delete limited by Play API |
| Validate / readiness | `asc validate` | `gpd validate` (dry-run readiness report) | **Partial** — local + plan; live network probes via `auth check` / publish status |
| High-level publish | `asc publish appstore\|testflight` | `gpd publish play` (upload→track→status), plus primitives | **Partial** — Play track model |
| Status | `asc status --watch` | `gpd publish status`, `gpd monitor *`, `gpd automation monitor` | **Partial** |
| Versions / phased release | `asc versions`, phased-release | `gpd publish rollout`, `gpd release-mgmt *` | **Partial** |

### Analytics, finance, ads

| ASC Feature Group | gpd Equivalent | Status |
| --- | --- | --- |
| Analytics & sales | `gpd analytics *`, `gpd vitals *` | **Partial** |
| Finance / Apple Ads | N/A | **Not applicable** |

### Monetization & games

| ASC Feature Group | gpd Equivalent | Status |
| --- | --- | --- |
| IAP & subscriptions | `gpd monetization *` | **Partial** |
| Purchase verification | `gpd purchases *` | **gpd-only** (stronger on Play) |
| StoreKit retention | N/A | **Not applicable** |
| Game Center | `gpd games *`, `gpd grouping` | **Partial** |

### Automation & tooling

| ASC Feature Group | gpd Equivalent | Status |
| --- | --- | --- |
| Workflows | `gpd workflow *` | **Full** for declarative multi-step runs (schema differs) |
| Migrate (Fastlane) | `gpd migrate` | **Partial** |
| Agent skills | [`skills/`](../skills/README.md) (`gpd-auth`, `gpd-release`, `gpd-reviews-vitals`) | **Partial** (packaged skills; not a runtime installer) |
| Extensions | `gpd extension *` | **gpd-only** |
| Telemetry / snitch / schema search | N/A or not implemented | **Not applicable** |

---

## gpd features without ASC equivalent

Permissions, edit lifecycle depth, internal sharing, deobfuscation, purchases, integrity, recovery, vitals depth, custom apps, generated/system APKs, bulk, compare, release-mgmt, testing, automation, monitor, maintenance/API drift, extensions.

---

## Intentionally not mirrored (Apple-only)

Devices, certificates, profiles, bundle IDs, notarization, Xcode / Xcode Cloud, sandbox testers, App Tags/Events/Clips, alternative distribution, Apple Ads, ASC finance, StoreKit retention, web-session scraping.

---

## Remaining product gaps

Optional / later:

1. Broader domain package migration for remaining large `kong_*.go` families (beyond `outfmt` + `playship`)  
2. Even deeper `validate` (media matrix, content rating, all listing locales)  
3. ~~Publish production tag `v0.6.5+`~~ — **v0.6.5** shipped with new archive + checksum names  

**Delivered recently:** TTY-aware output; checksum install + GoReleaser SHA-256 (snapshot verified); generated command docs; `validate` (package/track/listing network probes) + `publish play`; multi-profile auth including **delete/logout**; `setup-gpd` action; agent **skills/** pack; `outfmt` + `playship` extractions.

---

## Maintenance rule

When adding or removing a user-facing command:

1. `make generate-command-docs` and commit `docs/COMMANDS.md`  
2. `make check-docs` must pass  
3. Update this matrix if ASC parity status changes  
4. Prefer live `gpd <cmd> --help` over inventing flags in docs  

For Play API endpoint mapping (not ASC), see [API Coverage Matrix](api-coverage-matrix.md).
