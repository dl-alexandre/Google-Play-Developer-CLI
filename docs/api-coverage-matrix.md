# API Coverage Matrix

This document provides a comprehensive overview of Google Play API coverage in the Google Play Developer CLI (gpd).

## Android Publisher API v3

### Edits

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.insert` | `gpd publish edit create` | ✅ | Create a new edit transaction |
| `edits.list` | `gpd publish edit list` | ✅ | List cached edits (local cache) |
| `edits.get` | `gpd publish edit get <edit-id>` | ✅ | Get edit details |
| `edits.commit` | `gpd publish edit commit <edit-id>` | ✅ | Commit an edit transaction |
| `edits.validate` | `gpd publish edit validate <edit-id>` | ✅ | Validate an edit before committing |
| `edits.delete` | `gpd publish edit delete <edit-id>` | ✅ | Delete an edit transaction |

### Bundles/APKs

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.bundles.upload` | `gpd publish upload <file.aab>` | ✅ | Upload Android App Bundle (AAB) |
| `edits.bundles.list` | `gpd publish builds list --type bundle` | ✅ | List bundles in an edit |
| `edits.bundles.get` | `gpd publish builds get <version-code> --type bundle` | ✅ | Get bundle details |
| `edits.apks.upload` | `gpd publish upload <file.apk>` | ✅ | Upload APK file |
| `edits.apks.list` | `gpd publish builds list --type apk` | ✅ | List APKs in an edit |
| `edits.apks.get` | `gpd publish builds get <version-code> --type apk` | ✅ | Get APK details |

### Expansion Files (OBB)

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.expansionfiles.upload` | `gpd publish upload --obb-main/--obb-patch <file.obb>` | ✅ | Upload expansion files (OBB) |
| `edits.expansionfiles.update` | `gpd publish upload --obb-main-references-version/--obb-patch-references-version <version>` | ✅ | Update expansion file references |

### Tracks

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.tracks.list` | `gpd publish tracks` | ✅ | List all release tracks |
| `edits.tracks.get` | `gpd publish status --track <track>` | ✅ | Get specific track status |
| `edits.tracks.update` | `gpd publish release`, `gpd publish rollout`, `gpd publish promote`, `gpd publish halt`, `gpd publish rollback` | ✅ | Update track releases (via release commands). See examples/release-workflow.md for ASC mapping |

### Testers (Track-based)

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.testers.get` | `gpd publish testers get --track <track>`, `gpd publish testers list [--track <track>]` | ✅ | Retrieve tester groups for internal/alpha/beta tracks |
| `edits.testers.update` | `gpd publish testers add`, `gpd publish testers remove` | ✅ | Update tester Google Groups for track |
| N/A (ASC compatibility layer) | `gpd publish beta-groups list/get/create/update/delete/add-testers/remove-testers` | ✅ | ASC-style compatibility commands mapped to track testers; no standalone beta-group object in Play API |

### Listings

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.listings.get` | `gpd publish listing get [--locale <locale>]` | ✅ | Get store listing for locale(s) |
| `edits.listings.update` | `gpd publish listing update` | ✅ | Update store listing (title, descriptions) |
| `edits.listings.patch` | `gpd publish listing update` | ✅ | Uses update endpoint with partial fields |

### Images

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.images.upload` | `gpd publish images upload <type> <file>` | ✅ | Upload store images (screenshots, icons, etc.) |
| `edits.images.list` | `gpd publish images list <type>` | ✅ | List images for a type and locale |
| `edits.images.delete` | `gpd publish images delete <type> <id>` | ✅ | Delete a specific image |
| `edits.images.deleteall` | `gpd publish images deleteall <type>` | ✅ | Delete all images for a type |

### Deobfuscation Files

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.deobfuscationfiles.upload` | `gpd publish deobfuscation upload <file>` | ✅ | Upload ProGuard/R8 mappings or native debug symbols |

### Internal App Sharing

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `internalappsharingartifacts.uploadapk` | `gpd publish internal-share upload <file.apk>` | ✅ | Upload APK for internal testing |
| `internalappsharingartifacts.uploadbundle` | `gpd publish internal-share upload <file.aab>` | ✅ | Upload AAB for internal testing |

### App Details

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.details.get` | `gpd publish details get` | ✅ | Get app contact information |
| `edits.details.update` | `gpd publish details update` | ✅ | Update app details |
| `edits.details.patch` | `gpd publish details patch` | ✅ | Patch app details with update mask |

### Monetization - Subscriptions

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `monetization.subscriptions.list` | `gpd monetization subscriptions list` | ✅ | List subscription products |
| `monetization.subscriptions.get` | `gpd monetization subscriptions get <id>` | ✅ | Get subscription details |
| `monetization.subscriptions.create` | `gpd monetization subscriptions create` | ✅ | Create a subscription product |
| `monetization.subscriptions.patch` | `gpd monetization subscriptions patch <id>` | ✅ | Patch subscription with update mask |
| `monetization.subscriptions.update` | `gpd monetization subscriptions update <id>` | ✅ | Update subscription (uses patch internally) |
| `monetization.subscriptions.delete` | `gpd monetization subscriptions delete <id>` | ✅ | Delete a subscription |
| `monetization.subscriptions.archive` | `gpd monetization subscriptions archive <id>` | ✅ | Archive a subscription |
| `monetization.subscriptions.batchGet` | `gpd monetization subscriptions batchGet` | ✅ | Batch get multiple subscriptions |
| `monetization.subscriptions.batchUpdate` | `gpd monetization subscriptions batchUpdate` | ✅ | Batch update subscriptions |
| `monetization.convertRegionPrices` | `gpd monetization subscriptions convert-prices` | ✅ | Convert subscription prices across regions |

### Monetization - Base Plans

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `monetization.subscriptions.basePlans.activate` | `gpd monetization baseplans activate <sub-id> <plan-id>` | ✅ | Activate a base plan |
| `monetization.subscriptions.basePlans.deactivate` | `gpd monetization baseplans deactivate <sub-id> <plan-id>` | ✅ | Deactivate a base plan |
| `monetization.subscriptions.basePlans.delete` | `gpd monetization baseplans delete <sub-id> <plan-id>` | ✅ | Delete a base plan |
| `monetization.subscriptions.basePlans.migratePrices` | `gpd monetization baseplans migrate-prices <sub-id> <plan-id>` | ✅ | Migrate base plan prices |
| `monetization.subscriptions.basePlans.batchMigratePrices` | `gpd monetization baseplans batch-migrate-prices <sub-id>` | ✅ | Batch migrate prices for multiple plans |
| `monetization.subscriptions.basePlans.batchUpdateStates` | `gpd monetization baseplans batch-update-states <sub-id>` | ✅ | Batch update base plan states |

### Monetization - Offers

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `monetization.subscriptions.basePlans.offers.create` | `gpd monetization offers create <sub-id> <plan-id>` | ✅ | Create a subscription offer |
| `monetization.subscriptions.basePlans.offers.get` | `gpd monetization offers get <sub-id> <plan-id> <offer-id>` | ✅ | Get offer details |
| `monetization.subscriptions.basePlans.offers.list` | `gpd monetization offers list <sub-id> <plan-id>` | ✅ | List offers for a base plan |
| `monetization.subscriptions.basePlans.offers.delete` | `gpd monetization offers delete <sub-id> <plan-id> <offer-id>` | ✅ | Delete an offer |
| `monetization.subscriptions.basePlans.offers.activate` | `gpd monetization offers activate <sub-id> <plan-id> <offer-id>` | ✅ | Activate an offer |
| `monetization.subscriptions.basePlans.offers.deactivate` | `gpd monetization offers deactivate <sub-id> <plan-id> <offer-id>` | ✅ | Deactivate an offer |
| `monetization.subscriptions.basePlans.offers.batchGet` | `gpd monetization offers batchGet <sub-id> <plan-id>` | ✅ | Batch get multiple offers |
| `monetization.subscriptions.basePlans.offers.batchUpdate` | `gpd monetization offers batchUpdate <sub-id> <plan-id>` | ✅ | Batch update offers |
| `monetization.subscriptions.basePlans.offers.batchUpdateStates` | `gpd monetization offers batchUpdateStates <sub-id> <plan-id>` | ✅ | Batch update offer states |

### Monetization - In-app Products

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `inappproducts.list` | `gpd monetization products list` | ✅ | List in-app products |
| `inappproducts.get` | `gpd monetization products get <id>` | ✅ | Get product details |
| `inappproducts.insert` | `gpd monetization products create` | ✅ | Create an in-app product |
| `inappproducts.update` | `gpd monetization products update <id>` | ✅ | Update an in-app product |
| `inappproducts.delete` | `gpd monetization products delete <id>` | ✅ | Delete an in-app product |

### Purchases - Products

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `purchases.products.get` | `gpd purchases verify --type product` | ✅ | Get product purchase details |
| `purchases.products.acknowledge` | `gpd purchases products acknowledge` | ✅ | Acknowledge a product purchase |
| `purchases.products.consume` | `gpd purchases products consume` | ✅ | Consume a consumable product |

### Purchases - Subscriptions

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `purchases.subscriptionsv2.get` | `gpd purchases verify --type subscription` | ✅ | Get subscription purchase (v2 API) |
| `purchases.subscriptions.acknowledge` | `gpd purchases subscriptions acknowledge` | ✅ | Acknowledge a subscription purchase |
| `purchases.subscriptions.cancel` | `gpd purchases subscriptions cancel` | ✅ | Cancel a subscription |
| `purchases.subscriptions.defer` | `gpd purchases subscriptions defer` | ✅ | Defer subscription renewal |
| `purchases.subscriptions.refund` | `gpd purchases subscriptions refund` | ✅ | Refund a subscription (v1) |
| `purchases.subscriptions.revoke` | `gpd purchases subscriptions revoke` | ✅ | Revoke subscription (v1) |
| `purchases.subscriptionsv2.revoke` | `gpd purchases subscriptions revoke` | ✅ | Revoke subscription (v2 API) |

### Purchases - Voided

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `purchases.voidedpurchases.list` | `gpd purchases voided list` | ✅ | List voided purchases |

### Reviews

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `reviews.list` | `gpd reviews list` | ✅ | List user reviews with filtering |
| `reviews.get` | `gpd reviews get` | ✅ | Get a specific review by ID |
| `reviews.reply` | `gpd reviews reply` | ✅ | Reply to a user review |
| N/A (delete operation not exposed by Play API) | `gpd reviews response delete` | ⚠️ | Command exists for parity but returns explicit platform limitation guidance |

### Users

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `users.create` | `gpd permissions users create` | ✅ | Create a developer account user |
| `users.list` | `gpd permissions users list` | ✅ | List users in developer account |
| `users.get` | `gpd permissions users get <name>` | ✅ | Get user details |
| `users.delete` | `gpd permissions users delete <name>` | ✅ | Delete a user |
| `users.patch` | `gpd permissions users patch <name>` | ✅ | Update user permissions |

### Grants

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `grants.create` | `gpd permissions grants create` | ✅ | Create app-level permission grant |
| `grants.delete` | `gpd permissions grants delete <name>` | ✅ | Delete a grant |
| `grants.patch` | `gpd permissions grants patch <name>` | ✅ | Update grant permissions |

### App Recovery

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `apprecovery.create` | `gpd recovery create` | ✅ | Create a draft recovery action |
| `apprecovery.list` | `gpd recovery list` | ✅ | List recovery actions |
| `apprecovery.deploy` | `gpd recovery deploy <recovery-id>` | ✅ | Deploy a recovery action |
| `apprecovery.cancel` | `gpd recovery cancel <recovery-id>` | ✅ | Cancel a recovery action |
| `apprecovery.addTargeting` | `gpd recovery add-targeting <recovery-id>` | ✅ | Add targeting to recovery action |

## Play Developer Reporting API v1beta1

### Apps

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `apps.search` | `gpd apps list` | ✅ | List apps accessible to the account |

### Crash Rate

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.crashrate.query` | `gpd vitals crashes` | ✅ | Query crash rate metrics |

### ANR Rate

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.anrrate.query` | `gpd vitals anrs` | ✅ | Query ANR rate metrics |

### Excessive Wakeups

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.excessivewakeuprate.query` | `gpd vitals excessive-wakeups` | ✅ | Query excessive wakeup rate |

### Slow Rendering

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.slowrenderingrate.query` | `gpd vitals slow-rendering` | ✅ | Query slow rendering rate |

### Slow Start

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.slowstartrate.query` | `gpd vitals slow-start` | ✅ | Query slow start rate |

### Stuck Wakelocks

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.stuckbackgroundwakelockrate.query` | `gpd vitals stuck-wakelocks` | ✅ | Query stuck wakelock rate |

### LMK Rate

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| N/A | `gpd vitals lmk-rate` | ❌ | Not available in Play Developer Reporting API v1beta1 |

### Error Issues Search

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.errors.issues.search` | `gpd vitals errors issues search` | ✅ | Search error issues |

### Error Reports Search

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.errors.reports.search` | `gpd vitals errors reports search` | ✅ | Search error reports with deobfuscation support |

### Error Counts

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.errors.counts.get` | `gpd vitals errors counts get` | ✅ | Get error count summary |
| `vitals.errors.counts.query` | `gpd vitals errors counts query` | ✅ | Query error counts over time |

### Anomalies

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `anomalies.list` | `gpd vitals anomalies list` | ✅ | List anomalies in vitals metrics |

## Play Integrity API v1

### Integrity Token Decoding

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `v1.decodeIntegrityToken` | `gpd integrity decode` | ✅ | Decode Play Integrity tokens |

## Play Games Services API v1

### Access Tokens (Play Grouping)

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `accesstokens.generatePlayGroupingApiToken` | `gpd grouping token` | ✅ | Generate Play Grouping API token |
| `accesstokens.generateRecallPlayGroupingApiToken` | `gpd grouping token-recall` | ✅ | Generate Play Grouping API token via Recall |

## Play Custom App Publishing API v1

### Custom Apps

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `accounts.customApps.create` | `gpd customapp create` | ✅ | Create and upload custom app |

## Games Management API v1

### Achievements Reset

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `achievements.reset` | `gpd games achievements reset <id>` | ✅ | Reset achievement for current player |
| `achievements.resetAll` | `gpd games achievements reset` | ✅ | Reset all achievements for current player |
| `achievements.resetForAllPlayers` | `gpd games achievements reset <id> --all-players` | ✅ | Reset achievement for all players |
| `achievements.resetAllForAllPlayers` | `gpd games achievements reset --all-players` | ✅ | Reset all achievements for all players |
| `achievements.resetMultipleForAllPlayers` | `gpd games achievements reset --ids <ids> --all-players` | ✅ | Reset multiple achievements for all players |

### Scores Reset

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `scores.reset` | `gpd games scores reset <leaderboard-id>` | ✅ | Reset scores for current player |
| `scores.resetAll` | `gpd games scores reset` | ✅ | Reset all scores for current player |
| `scores.resetForAllPlayers` | `gpd games scores reset <leaderboard-id> --all-players` | ✅ | Reset scores for all players |
| `scores.resetAllForAllPlayers` | `gpd games scores reset --all-players` | ✅ | Reset all scores for all players |
| `scores.resetMultipleForAllPlayers` | `gpd games scores reset --ids <ids> --all-players` | ✅ | Reset multiple leaderboards for all players |

### Events Reset

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `events.reset` | `gpd games events reset <event-id>` | ✅ | Reset event for current player |
| `events.resetAll` | `gpd games events reset` | ✅ | Reset all events for current player |
| `events.resetForAllPlayers` | `gpd games events reset <event-id> --all-players` | ✅ | Reset event for all players |
| `events.resetAllForAllPlayers` | `gpd games events reset --all-players` | ✅ | Reset all events for all players |
| `events.resetMultipleForAllPlayers` | `gpd games events reset --ids <ids> --all-players` | ✅ | Reset multiple events for all players |

### Players Hide/Unhide

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `players.hide` | `gpd games players hide <player-id>` | ✅ | Hide player from leaderboards |
| `players.unhide` | `gpd games players unhide <player-id>` | ✅ | Unhide player from leaderboards |
| `applications.listHidden` | `gpd games applications list-hidden <app-id>` | ✅ | List hidden players for an application |

## Bulk Operations

### Bulk Upload

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.bundles.upload` (batch) | `gpd bulk upload <files...>` | ⚠️ | Upload multiple APKs/AABs in parallel with controlled concurrency |

### Bulk Listings

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.listings.update` (batch) | `gpd bulk listings --data-file <file>` | ⚠️ | Update store listings across multiple locales from JSON file |

### Bulk Images

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.images.upload` (batch) | `gpd bulk images --image-dir <dir>` | ⚠️ | Batch upload images organized by type/locale directory structure |

### Bulk Tracks

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.tracks.update` (batch) | `gpd bulk tracks --tracks <tracks...>` | ⚠️ | Update multiple tracks with same release simultaneously |

## App Comparison

### Compare Vitals

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.crashrate.query` + `vitals.anrrate.query` | `gpd compare vitals --packages <pkgs...>` | ⚠️ | Compare crash/ANR metrics across multiple apps |

### Compare Reviews

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `reviews.list` (aggregated) | `gpd compare reviews --packages <pkgs...>` | ⚠️ | Compare review metrics and ratings across apps |

### Compare Releases

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.tracks.get` (multiple) | `gpd compare releases --packages <pkgs...>` | ⚠️ | Compare release history and timelines across apps |

### Compare Subscriptions

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `monetization.subscriptions.list` (aggregated) | `gpd compare subscriptions --packages <pkgs...>` | ⚠️ | Compare subscription metrics (MRR, churn, ARPU) across apps |

## Release Management

### Release Calendar

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.tracks.get` (timeline view) | `gpd release-mgmt calendar` | ⚠️ | Show upcoming and past releases in calendar format |

### Release Conflicts

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.bundles.list` + `edits.tracks.get` | `gpd release-mgmt conflicts --version-codes <codes...>` | ⚠️ | Detect version code conflicts across tracks |

### Release Strategy

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.*.query` + `edits.tracks.get` | `gpd release-mgmt strategy --track <track>` | ⚠️ | Get rollback/roll-forward recommendations based on health metrics |

### Release History

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.tracks.get` (historical) | `gpd release-mgmt history --track <track>` | ⚠️ | Show detailed release history with optional vitals |

### Release Notes

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| N/A (Play Console feature) | `gpd release-mgmt notes <action>` | ⚠️ | Manage release notes across locales (get, set, copy, list) |

## Testing Tools

### Pre-launch Report

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| Limited API access | `gpd testing prelaunch <action>` | ⚠️ | Trigger or check pre-launch report. Limited API access, primarily Play Console UI |

### Device Lab

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| Firebase Test Lab API | `gpd testing device-lab --app-file <file>` | ⚠️ | Run tests on Firebase Test Lab. Requires Firebase project setup |

### Screenshots

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| Firebase Test Lab / Play Console | `gpd testing screenshots --app-file <file>` | ⚠️ | Capture screenshots across devices. Requires Firebase Test Lab integration |

### Validate

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| Local validation | `gpd testing validate --app-file <file>` | ✅ | Comprehensive app validation (AAB format, signing, permissions, size, API level) |

### Compatibility

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| Device catalog API | `gpd testing compatibility --app-file <file>` | ⚠️ | Check device compatibility against Play Store device catalog |

## Release Automation

### Automated Release Notes

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| Git / GitHub API | `gpd automation release-notes --source <git|pr|file>` | ✅ | Generate release notes from git history, PRs, or file |

### Automated Rollout

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.tracks.update` (staged) | `gpd automation rollout --track <track>` | ✅ | Automated staged rollout with health checks and auto-rollback |

### Smart Promote

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `edits.tracks.update` (promotion) | `gpd automation promote --from-track <track> --to-track <track>` | ✅ | Smart promote with optional verification and health checks |

### Pre-release Validation

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| Multiple validation checks | `gpd automation validate --checks <checks...>` | ✅ | Comprehensive pre-release validation (AAB, signing, permissions, deobfuscation) |

### Post-release Monitoring

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.*.query` | `gpd automation monitor --track <track>` | ✅ | Monitor release health after rollout with threshold-based alerting |

## Monitoring

### Continuous Watch

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.crashrate.query` + `vitals.anrrate.query` + `vitals.errors.counts.query` | `gpd monitor watch` | ✅ | Continuous vitals monitoring with configurable thresholds and alerting |

### Anomaly Detection

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| `vitals.*.query` (statistical analysis) | `gpd monitor anomalies` | ✅ | Detect statistical anomalies in vitals metrics with baseline comparison |

### Dashboard

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| All vitals metrics APIs | `gpd monitor dashboard` | ✅ | Generate comprehensive monitoring dashboard data with trends |

### Scheduled Reports

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| All vitals metrics APIs | `gpd monitor report --period <period>` | ✅ | Generate scheduled monitoring reports (daily, weekly, monthly) with findings and recommendations |

### Webhooks

| API Endpoint | CLI Command | Status | Notes |
|--------------|-------------|--------|-------|
| N/A (Play Console UI only) | `gpd monitor webhooks list` | ⚠️ | Google Play webhooks are configured via Play Console UI, not API. Command shows example format only |

## Summary

- **Android Publisher API v3**: Comprehensive coverage of publishing, monetization, purchases, reviews, permissions, and recovery operations
- **Play Developer Reporting API v1beta1**: Full coverage of vitals metrics, error reporting, and anomalies (except LMK Rate which is not available in the API)
- **Games Management API v1**: Complete coverage of achievement, score, event resets, and player visibility management
- **Bulk Operations**: Batch processing commands for uploads, listings, images, and tracks (stubs ready for implementation)
- **App Comparison**: Cross-app analysis for vitals, reviews, releases, and subscriptions (stubs ready for implementation)
- **Release Management**: Enhanced release lifecycle tools including calendar, conflict detection, strategy recommendations, history, and notes management (stubs ready for implementation)
- **Testing Tools**: Pre-launch reports, Firebase Test Lab integration, screenshot capture, validation, and compatibility checking (partial implementation)
- **Release Automation**: CI/CD-ready commands for automated release notes generation, staged rollouts with health checks, smart promotion, validation, and monitoring (fully implemented)
- **Monitoring**: Real-time vitals monitoring with threshold alerting, anomaly detection, dashboard generation, and scheduled reporting (fully implemented)

## Notes

- All commands require authentication via `gpd auth login` or service account key
- Most publishing operations use edit transactions which must be committed
- Vitals data may be delayed by 24-48 hours
- Games Management API requires Games OAuth scope
- Some operations (like `--all-players` in games commands) require admin privileges
