# App Store Connect CLI Parity Matrix

This matrix maps App Store Connect CLI feature groups to `gpd` equivalents. Where Google Play has no direct analogue, the status is marked as not applicable. Links in the gpd column point to the most relevant reference or example documentation in this repo.

Status meanings:
- Full: Comparable capability and scope
- Partial: Similar capability with notable gaps or model differences
- Not applicable: No Google Play equivalent
- gpd-only: Google Play capability with no ASC equivalent

## Parity Matrix

| ASC Feature Group | ASC Commands (examples) | gpd Equivalent (docs) | Status |
| --- | --- | --- | --- |
| Authentication | `asc auth login`, `asc auth switch`, `asc auth init`, `asc auth status`, `asc auth doctor`, `asc auth logout` | `gpd auth login`, `gpd auth init`, `gpd auth switch`, `gpd auth list`, `gpd auth status`, `gpd auth check`, `gpd auth diagnose`, `gpd auth doctor`, `gpd auth logout` ([Command Reference](../README.md#command-reference), [Auth Parity Guide](auth-parity-guide.md)) | Partial (device code OAuth + service accounts; no browser auth) |
| Apps & Builds | `asc apps`, `asc builds list`, `asc builds info`, `asc builds expire`, `asc builds expire-all`, `asc builds upload`, `asc builds add-groups`, `asc builds remove-groups` | `gpd apps list`, `gpd apps get`, `gpd publish builds list`, `gpd publish builds get`, `gpd publish builds expire`, `gpd publish builds expire-all`, `gpd publish upload`, `gpd publish status`, `gpd publish tracks`, `gpd publish capabilities` ([API: Apps](api-coverage-matrix.md#apps), [API: Bundles/APKs](api-coverage-matrix.md#bundlesapks), [API: Tracks](api-coverage-matrix.md#tracks)) | Partial (no global build registry; no build-level beta group assignment) |
| TestFlight | `asc feedback`, `asc crashes`, `asc testflight apps list`, `asc testflight apps get`, `asc testflight sync pull` | `gpd vitals crashes` ([Command Reference](../README.md#command-reference), [Error Debugging](examples/error-debugging.md)) | Partial (no feedback/testflight sync) |
| Beta Groups | `asc beta-groups list`, `asc beta-groups create`, `asc beta-groups get`, `asc beta-groups update`, `asc beta-groups delete`, `asc beta-groups add-testers`, `asc beta-groups remove-testers` | `gpd publish beta-groups list/get/create/update/delete/add-testers/remove-testers`, `gpd publish testers list/get/add/remove` ([Command Reference](../README.md#command-reference)) | Partial (compatibility commands map groups to internal/alpha/beta tracks; no standalone group object in Play API) |
| Beta Testers | `asc beta-testers list`, `asc beta-testers get`, `asc beta-testers add`, `asc beta-testers remove`, `asc beta-testers invite`, `asc beta-testers add-groups`, `asc beta-testers remove-groups` | `gpd publish testers list`, `gpd publish testers get`, `gpd publish testers add`, `gpd publish testers remove` ([Command Reference](../README.md#command-reference)) | Partial (track-based, no invite lifecycle) |
| Devices | `asc devices list`, `asc devices get`, `asc devices register`, `asc devices update` | N/A | Not applicable |
| App Store | `asc reviews`, `asc reviews respond`, `asc reviews response get`, `asc reviews response for-review`, `asc reviews response delete` | `gpd reviews list`, `gpd reviews get`, `gpd reviews reply`, `gpd reviews response get`, `gpd reviews response for-review`, `gpd reviews response delete`, `gpd reviews capabilities` ([API: Reviews](api-coverage-matrix.md#reviews)) | Partial (`gpd reviews response delete` exists but returns platform limitation: deletion unsupported by Google Play API) |
| App Tags | `asc app-tags list`, `asc app-tags get`, `asc app-tags update`, `asc app-tags territories`, `asc app-tags territories-relationships`, `asc app-tags relationships` | N/A | Not applicable |
| App Events | `asc app-events list`, `asc app-events localizations list`, `asc app-events localizations screenshots list`, `asc app-events localizations video-clips list`, `asc app-events relationships`, `asc app-events localizations screenshots-relationships`, `asc app-events localizations video-clips-relationships` | N/A | Not applicable |
| Alternative Distribution | `asc alternative-distribution domains list`, `asc alternative-distribution domains create`, `asc alternative-distribution domains delete`, `asc alternative-distribution keys list`, `asc alternative-distribution keys create`, `asc alternative-distribution keys app`, `asc alternative-distribution packages create`, `asc alternative-distribution packages get`, `asc alternative-distribution packages versions list`, `asc alternative-distribution packages versions get`, `asc alternative-distribution packages versions deltas`, `asc alternative-distribution packages versions variants` | N/A | Not applicable |
| Analytics & Sales | `asc analytics sales`, `asc analytics request`, `asc analytics requests`, `asc analytics get`, `asc analytics download` | `gpd analytics query`, `gpd analytics capabilities`, `gpd vitals crashes`, `gpd vitals anrs` ([Command Reference](../README.md#command-reference), [Error Debugging](examples/error-debugging.md)) | Partial (Play Reporting focus) |
| Finance Reports | `asc finance reports`, `asc finance regions` | N/A | Not applicable |
| Sandbox Testers | `asc sandbox list`, `asc sandbox get`, `asc sandbox update`, `asc sandbox clear-history` | N/A | Not applicable |
| Xcode Cloud | `asc xcode-cloud workflows`, `asc xcode-cloud build-runs`, `asc xcode-cloud run`, `asc xcode-cloud status` | N/A | Not applicable |
| Game Center | `asc game-center achievements`, `asc game-center leaderboards`, `asc game-center leaderboard-sets` | `gpd games`, `gpd grouping` ([API: Play Games Services](api-coverage-matrix.md#play-games-services-api-v1), [API: Games Management](api-coverage-matrix.md#games-management-api-v1)) | Partial (different feature set) |
| App Setup | `asc app-setup info set`, `asc app-setup categories set`, `asc app-setup availability set`, `asc app-setup pricing set`, `asc app-setup localizations upload` | `gpd publish listing update`, `gpd publish details update`, `gpd publish images upload`, `gpd publish assets upload`, `gpd publish assets spec` ([API: Listings](api-coverage-matrix.md#listings), [API: Images](api-coverage-matrix.md#images), [API: App Details](api-coverage-matrix.md#app-details)) | Partial (pricing/availability differ) |
| Categories | `asc categories list`, `asc categories set` | N/A | Not applicable |
| Versions | `asc versions list`, `asc versions get`, `asc versions attach-build`, `asc versions release`, `asc versions phased-release get`, `asc versions phased-release create`, `asc versions phased-release update`, `asc versions phased-release delete`, `asc versions promotions create` | `gpd publish release`, `gpd publish rollout`, `gpd publish promote`, `gpd publish halt`, `gpd publish rollback`, `gpd publish status` ([API: Tracks](api-coverage-matrix.md#tracks), [Edit Workflow](examples/edit-workflow.md)) | Partial (workflow and concepts differ) |
| App Info | `asc app-info get`, `asc app-info set` | `gpd publish listing get`, `gpd publish listing update`, `gpd publish details get`, `gpd publish details update` ([API: Listings](api-coverage-matrix.md#listings), [API: App Details](api-coverage-matrix.md#app-details)) | Partial (workflow and scope differ) |
| Pre-Release Versions | `asc pre-release-versions list`, `asc pre-release-versions get` | N/A | Not applicable |
| Localizations | `asc localizations list`, `asc localizations download`, `asc localizations upload` | `gpd publish listing get`, `gpd publish listing update`, `gpd publish images upload` ([API: Listings](api-coverage-matrix.md#listings), [API: Images](api-coverage-matrix.md#images)) | Partial (scope and workflow differ) |
| Build Localizations | `asc build-localizations list`, `asc build-localizations create`, `asc build-localizations update`, `asc build-localizations delete`, `asc build-localizations get` | N/A | Not applicable |
| Offer Codes (Subscriptions) | `asc offer-codes list`, `asc offer-codes generate`, `asc offer-codes values` | N/A | Not applicable |
| In-App Purchases & Subscriptions | `asc in-app-purchases`, `asc subscriptions`, `asc subscription-groups` | `gpd monetization products`, `gpd monetization subscriptions`, `gpd monetization baseplans`, `gpd monetization offers` ([API: Monetization - Subscriptions](api-coverage-matrix.md#monetization---subscriptions), [API: Base Plans](api-coverage-matrix.md#monetization---base-plans), [API: Offers](api-coverage-matrix.md#monetization---offers), [API: In-app Products](api-coverage-matrix.md#monetization---in-app-products), [Subscription Management](examples/subscription-management.md)) | Partial (model differs) |
| Migrate (Fastlane Compatibility) | `asc migrate validate`, `asc migrate import`, `asc migrate export` | `gpd migrate fastlane` ([Command Reference](../README.md#command-reference)) | Full |
| Submit | `asc submit create`, `asc submit status`, `asc submit cancel` | `gpd publish release`, `gpd publish rollout`, `gpd publish promote`, `gpd publish halt`, `gpd publish rollback` ([API: Tracks](api-coverage-matrix.md#tracks), [Edit Workflow](examples/edit-workflow.md)) | Partial (workflow and concepts differ) |
| Utilities | `asc version` | `gpd version`, `gpd config init`, `gpd config doctor`, `gpd config path`, `gpd config get`, `gpd config set`, `gpd config completion` ([Command Reference](../README.md#command-reference)) | Partial (additional utilities available) |
| Output Formats | `asc --output table`, `asc --output markdown` | `gpd --output json`, `gpd --output table`, `gpd --output markdown` ([Command Reference](../README.md#command-reference)) | Full |

## gpd Features Without ASC Equivalent

| gpd Feature Group | gpd Commands (examples) | Docs | Status |
| --- | --- | --- | --- |
| Permissions & Access | `gpd permissions users`, `gpd permissions grants` | [API: Users](api-coverage-matrix.md#users), [API: Grants](api-coverage-matrix.md#grants) | gpd-only |
| Edit Transactions | `gpd publish edit create`, `gpd publish edit validate`, `gpd publish edit commit` | [API: Edits](api-coverage-matrix.md#edits), [Edit Workflow](examples/edit-workflow.md) | gpd-only |
| Internal App Sharing | `gpd publish internal-share upload` | [API: Internal App Sharing](api-coverage-matrix.md#internal-app-sharing) | gpd-only |
| Deobfuscation Uploads | `gpd publish deobfuscation upload` | [API: Deobfuscation Files](api-coverage-matrix.md#deobfuscation-files) | gpd-only |
| Purchases Verification | `gpd purchases verify`, `gpd purchases products acknowledge`, `gpd purchases subscriptions revoke`, `gpd purchases voided list` | [API: Purchases - Products](api-coverage-matrix.md#purchases---products), [API: Purchases - Subscriptions](api-coverage-matrix.md#purchases---subscriptions), [API: Purchases - Voided](api-coverage-matrix.md#purchases---voided) | gpd-only |
| Play Integrity | `gpd integrity decode` | [API: Play Integrity](api-coverage-matrix.md#play-integrity-api-v1) | gpd-only |
| App Recovery | `gpd recovery create`, `gpd recovery deploy`, `gpd recovery cancel` | [API: App Recovery](api-coverage-matrix.md#app-recovery) | gpd-only |
| Android Vitals Error Reporting | `gpd vitals errors issues search`, `gpd vitals errors reports search`, `gpd vitals anomalies list` | [API: Error Issues Search](api-coverage-matrix.md#error-issues-search), [API: Anomalies](api-coverage-matrix.md#anomalies), [Error Debugging](examples/error-debugging.md) | gpd-only |
| Custom App Publishing | `gpd customapp create` | [API: Custom App Publishing](api-coverage-matrix.md#play-custom-app-publishing-api-v1) | gpd-only |

## Notes

- Some ASC features (Devices, Xcode Cloud, Sandbox Testers, App Tags, App Events, Finance Reports, Alternative Distribution) have no Google Play equivalents.
- Some gpd features (Play Integrity, App Recovery, Deobfuscation, Purchases, Edit Transactions, Permissions & Access) have no ASC equivalents.
- For the full Google Play API-to-command mapping, see [API Coverage Matrix](api-coverage-matrix.md).
