## CLI Assessment Log

Purpose: Track CLI runs and improvement opportunities discovered during testing.

### 2026-02-02

#### Run: Auth + Apps + Capabilities baseline
- Commands:
  - `gpd auth status`
  - `gpd apps list`
  - `gpd analytics capabilities`
  - `gpd vitals capabilities`
  - `gpd reviews list --page-size 1`
  - `gpd publish tracks`
  - `gpd permissions capabilities`
  - `gpd monetization capabilities`
  - `gpd purchases capabilities`
  - `gpd games capabilities`
- Observations:
  - `reviews list` returned `data: null` with `partial: true` and no warning.
- Improvements applied:
  - Added review list warnings when no data returned or filtered.
  - Added required scopes to analytics/vitals/reviews capabilities.
  - Removed duplicate `help` entry in root help output.
  - Link: TODO (add commit/PR)

#### Run: Read-only sweep with app data
- Commands:
  - `gpd apps list`
  - `gpd apps get com.milcgroup.animal`
  - `gpd publish tracks --package com.milcgroup.animal`
  - `gpd publish status --package com.milcgroup.animal --track production`
  - `gpd reviews list --package com.milcgroup.animal --page-size 1`
  - `gpd permissions users list --developer-id 123`
  - `gpd monetization products list --package com.milcgroup.animal --page-size 1`
- Observations:
  - `permissions users list` returned 400 with little guidance.
  - `monetization products list` returned 403 "Please migrate to the new publishing API."
- Improvements applied:
  - Added targeted hint for invalid developer ID (Play Console URL).
  - Added targeted hint for legacy monetization endpoint.
  - Ensured error responses include `services` metadata.
  - Link: TODO (add commit/PR)

#### Run: Output format and pagination ergonomics
- Commands:
  - `gpd --help`
  - `gpd reviews list --package com.milcgroup.animal --page-size 1`
  - `gpd analytics capabilities`
  - `gpd vitals capabilities`
  - `gpd reviews capabilities`
- Observations:
  - `--fields` was a no-op, `--output` help omitted `csv`, warnings were hidden in non-JSON.
  - `--all` and `--paginate` were ambiguous duplicates.
- Improvements applied:
  - Implemented `--fields` projection support (dotted paths + array indexes).
  - Updated `--output` help to include `csv` note.
  - Display warnings in table/markdown, emit to stderr for csv.
  - Deprecated `--paginate` in favor of `--all`.
  - Link: TODO (add commit/PR)

#### Run: Config + listing + voided purchases
- Commands:
  - `gpd config show`
  - `gpd publish listing get --package com.milcgroup.one --language en-US`
  - `gpd purchases voided --package com.milcgroup.one --page-size 1`
- Observations:
  - `config show` did not exist.
  - `listing get` had no `--language` alias.
  - `purchases voided` required `list` subcommand and returned `null` when empty.
- Improvements applied:
  - Added `config show` alias for `config print`.
  - Added deprecated `--language` alias for listing get.
  - Added direct `purchases voided` execution and `--page-size` alias.
  - Ensure `voidedPurchases` returns `[]` when empty.
  - Link: TODO (add commit/PR)

#### Run: Config doctor + details + assets + reporting queries
- Commands:
  - `gpd config doctor`
  - `gpd publish details get --package com.milcgroup.one`
  - `gpd publish assets spec`
  - `gpd analytics query --package com.milcgroup.one --start-date 2026-01-26 --end-date 2026-02-01 --metrics crashRate --dimensions country --page-size 1`
  - `gpd vitals query --package com.milcgroup.one --start-date 2026-01-26 --end-date 2026-02-01 --metrics crashRate --dimensions country --page-size 1`
- Observations:
  - Analytics/vitals queries returned 404 with HTML body and no hint.
- Improvements applied:
  - Added Play Reporting query error handling to include hint when 404 occurs.
  - Ensured errors include `services: ["playdeveloperreporting"]`.
  - Link: TODO (add commit/PR)

#### Run: Images, reviews, purchases, monetization, reporting (broad sweep)
- Commands:
  - `gpd publish images list phone --package com.milcgroup.one --locale en-US`
  - `gpd publish images list phoneScreenshots --package com.milcgroup.one --locale en-US`
  - `gpd reviews list --package com.milcgroup.one --page-size 1 --output table`
  - `gpd reviews list --package com.milcgroup.one --page-size 1 --output markdown`
  - `gpd reviews get --package com.milcgroup.one --review-id invalid`
  - `gpd purchases capabilities --output table`
  - `gpd purchases voided --package com.milcgroup.one --max-results 1 --output table`
  - `gpd monetization subscriptions list --package com.milcgroup.one --page-size 1 --output table`
  - `gpd monetization offers list --package com.milcgroup.one --page-size 1`
  - `gpd analytics query --package com.milcgroup.one --start-date 2026-01-26 --end-date 2026-02-01 --metrics crashRate --dimensions country --page-size 1 --output table`
  - `gpd vitals query --package com.milcgroup.one --start-date 2026-01-26 --end-date 2026-02-01 --metrics crashRate --dimensions country --page-size 1 --output table`
- Observations:
  - `publish images list phone` failed with a 400 due to invalid image type.
  - `reviews get` with invalid ID returned 400 with no hint.
  - `monetization offers list` required positional args and returned a generic error with no services.
  - Table output for `reviews list` showed JSON because slices were not rendered as tables.
- Improvements applied:
  - Added image type validation hint for image list errors.
  - Added hint for invalid review IDs.
  - Added explicit validation error (with services) for `offers list` missing args.
  - Table/markdown/csv now render `[]map[string]interface{}` properly.
  - Link: TODO (add commit/PR)

#### Run: Publish testers + purchases verify
- Commands:
  - `gpd publish testers list --package com.milcgroup.one`
  - `gpd purchases verify --package com.milcgroup.one --product-id fake_product --token fake_token`
- Observations:
  - `testers list` hit a 400 when attempting production track (non-testing).
  - `purchases verify` errors lacked `services` metadata.
- Improvements applied:
  - Skip production in testers list and add warning.
  - Ensure purchases verify error includes `services: ["androidpublisher"]`.
  - Link: TODO (add commit/PR)

#### Run: Assets category, recovery list, fastlane validate
- Commands:
  - `gpd publish assets upload --package com.milcgroup.one --dry-run --category phone`
  - `gpd recovery list --package com.milcgroup.one`
  - `gpd migrate fastlane validate --dir fastlane/metadata/android`
- Observations:
  - `publish assets upload` used a confusing `--replace` flag for category.
  - `recovery list` returned 400 about version code when no filter provided.
  - `migrate fastlane validate` failed with missing directory and no guidance.
- Improvements applied:
  - Added `--category` flag and deprecated `--replace`.
  - Added recovery list hint for version code errors and ensured services metadata.
  - Added migrate fastlane validate hints and `services: ["migrate"]`.
  - Link: TODO (add commit/PR)

#### Run: Integrity, grouping, games management, internal-share, deobfuscation
- Commands:
  - `gpd integrity decode --package com.milcgroup.one --token fake`
  - `gpd grouping token --package com.milcgroup.one --persona test`
  - `gpd games applications list-hidden dummy --page-size 1`
  - `gpd publish internal-share upload missing.aab --package com.milcgroup.one --dry-run`
  - `gpd publish deobfuscation upload missing.txt --package com.milcgroup.one --type proguard --version-code 1 --dry-run`
- Observations:
  - Integrity/Games APIs return 403 when disabled but lacked actionable hints.
  - Games management list-hidden returned a raw 403 without services metadata.
  - Internal-share/deobfuscation validation errors lacked services metadata.
- Improvements applied:
  - Added API-enable hints for integrity and games grouping tokens.
  - Added games management list-hidden hint + services metadata.
  - Added services metadata for internal-share/deobfuscation validation errors.
  - Link: TODO (add commit/PR)

#### Run: Custom app + recovery validation checks
- Commands:
  - `gpd customapp create`
  - `gpd recovery create`
  - `gpd purchases verify --package com.milcgroup.one --product-id dummy --token fake`
  - `gpd analytics capabilities`
  - `gpd vitals capabilities`
- Observations:
  - `customapp create` previously exited with no output because Cobra flag requirements short-circuited.
  - `recovery create` showed package-required error without services metadata.
  - `purchases verify` returned structured NOT_FOUND as expected.
  - Analytics/Vitals capabilities outputs look complete with scopes.
- Improvements applied:
  - Removed Cobra required-flag enforcement for customapp create and returned structured validation errors with services metadata.
  - Added services metadata for recovery create validation errors.
  - Link: TODO (add commit/PR)

#### Run: Games management resets + players validation
- Commands:
  - `gpd customapp create --account 1 --title "Test" --language en-US --apk ./not-an-apk.txt`
  - `gpd recovery create --package com.milcgroup.one`
  - `gpd games achievements reset "dummy-achievement-id"`
  - `gpd games scores reset "dummy-leaderboard-id"`
  - `gpd games events reset "dummy-event-id"`
  - `gpd games players hide`
  - `gpd games players unhide`
- Observations:
  - Games management reset errors lacked services metadata and actionable hints.
  - Players hide/unhide failed via Cobra arg/flag enforcement with no structured output.
- Improvements applied:
  - Added services metadata and API-enable hints for games management reset errors.
  - Replaced Cobra arg/flag enforcement with structured validation errors for players hide/unhide.
  - Link: TODO (add commit/PR)

#### Run: Games reset-all validation paths
- Commands:
  - `gpd games achievements reset`
  - `gpd games scores reset`
  - `gpd games events reset`
- Observations:
  - Validation errors returned without services metadata.
- Improvements applied:
  - Added `services: ["gamesmanagement"]` to reset-all validation errors.
  - Link: TODO (add commit/PR)

#### Run: Custom app directory + recovery file validation
- Commands:
  - `gpd customapp create --account 1 --title "Test" --language en-US --apk ./creds`
  - `gpd recovery create --package com.milcgroup.one --file ./missing.json`
- Observations:
  - Recovery create file validation lacked services metadata.
- Improvements applied:
  - Added `services: ["androidpublisher"]` to recovery create file/JSON validation errors.
  - Link: TODO (add commit/PR)

#### Run: Custom app non-APK + recovery invalid JSON
- Commands:
  - `gpd customapp create --account 1 --title "Test" --language en-US --apk /tmp/gpd-invalid.txt`
  - `gpd recovery create --package com.milcgroup.one --file /tmp/gpd-invalid.json`
- Observations:
  - Custom app and recovery validation errors already include services metadata and hints as expected.

#### Run: Integrity empty token file + grouping recall validation
- Commands:
  - `gpd integrity decode --package com.milcgroup.one --token-file /tmp/gpd-empty.txt`
  - `gpd grouping token-recall --package com.milcgroup.one --persona test`
- Observations:
  - Grouping token-recall previously exited without structured error due to required flag enforcement.
- Improvements applied:
  - Removed required-flag enforcement and returned structured validation errors with `services: ["games"]` for grouping token/recall.
  - Link: TODO (add commit/PR)

#### Run: Integrity token input conflicts + grouping token validation
- Commands:
  - `gpd integrity decode --package com.milcgroup.one --token fake --token-file /tmp/gpd-empty.txt`
  - `gpd grouping token --package com.milcgroup.one`
- Observations:
  - Validation errors returned with services metadata as expected.

#### Run: Analytics/Vitals date validation
- Commands:
  - `gpd analytics query --metrics crashRate`
  - `gpd vitals crashes`
- Observations:
  - Previously, missing start/end dates were enforced by Cobra and could exit silently.
- Improvements applied:
  - Added structured validation errors for missing dates with `services: ["playdeveloperreporting"]`.
  - Removed Cobra required-flag enforcement for analytics/vitals date flags.
  - Link: TODO (add commit/PR)

#### Run: Permissions required flag validation
- Commands:
  - `gpd permissions users list`
  - `gpd permissions grants create`
  - `gpd permissions grants create --package com.milcgroup.one`
- Observations:
  - Cobra required-flag enforcement could exit without structured output.
- Improvements applied:
  - Removed required flag enforcement for permissions users/grants creation and returned structured validation errors with `services: ["androidpublisher"]`.
  - Link: TODO (add commit/PR)

#### Run: Publish required flag validation
- Commands:
  - `gpd publish release`
  - `gpd publish rollout`
  - `gpd publish promote`
  - `gpd publish rollback`
  - `gpd publish deobfuscation upload dummy --package com.milcgroup.one --dry-run`
  - `gpd publish testers get`
- Observations:
  - Cobra required-flag enforcement could exit without structured output for publish commands.
- Improvements applied:
  - Added structured validation errors with `services: ["androidpublisher"]` for publish release/rollout/promote/rollback, deobfuscation upload, and testers get.
  - Link: TODO (add commit/PR)

#### Run: Purchases required flag validation
- Commands:
  - `gpd purchases verify`
  - `gpd purchases products acknowledge`
  - `gpd purchases subscriptions defer`
- Observations:
  - Cobra required-flag enforcement could exit without structured output for purchases commands.
- Improvements applied:
  - Added structured validation errors with `services: ["androidpublisher"]` for purchases verify, products acknowledge/consume, and subscriptions acknowledge/cancel/defer/refund/revoke.
  - Link: TODO (add commit/PR)

#### Run: Monetization required flag validation
- Commands:
  - `gpd monetization products create`
  - `gpd monetization subscriptions create`
  - `gpd monetization baseplans migrate-prices dummy dummy`
  - `gpd monetization offers create dummy dummy`
  - `gpd monetization offers batchGet dummy dummy`
  - `gpd monetization convert-region-prices`
- Observations:
  - Cobra required-flag enforcement could exit without structured output for monetization commands.
- Improvements applied:
  - Added structured validation errors with `services: ["androidpublisher"]` for monetization create/batch/convert operations and removed required-flag enforcement.
  - Link: TODO (add commit/PR)

#### Run: Vitals errors counts + LMK validation
- Commands:
  - `gpd vitals lmk-rate`
  - `gpd vitals errors counts get`
  - `gpd vitals errors counts query --package com.milcgroup.one --start-date 2024-01-01 --end-date 2024-01-02`
- Observations:
  - Remaining vitals date validation relied on Cobra required flags.
  - Errors counts API responses could return HTML 404s without standardized hints.
- Improvements applied:
  - Removed remaining required-flag enforcement for vitals LMK/errors counts.
  - Added structured validation errors with `services: ["playdeveloperreporting"]` for errors counts get.
  - Routed errors counts get/query failures through reporting error helper for consistent hints.
  - Replaced `counts get/query` command strings with a real `counts` parent command for reliable parsing.
  - Link: TODO (add commit/PR)

#### Run: Remaining validation metadata sweep
- Commands:
  - `gpd analytics query --start-date 2024-01-01 --end-date 2024-01-02 --metrics crashRate,anrRate`
  - `gpd reviews get`
  - `gpd publish testers add --track internal`
  - `gpd monetization subscriptions delete dummy`
  - `gpd purchases voided list --type badtype --package com.milcgroup.one`
  - `gpd purchases subscriptions revoke --revoke-type badtype --token t --package com.milcgroup.one`
- Observations:
  - Validation errors now consistently include `services` metadata across analytics/reviews/publish/monetization/purchases.
- Improvements applied:
  - Added `services` metadata to remaining validation paths in analytics, reviews, purchases, monetization, recovery add-targeting, permissions patch, publish testers, and vitals errors.
  - Link: TODO (add commit/PR)

#### Run: Validation spot checks
- Commands:
  - `gpd permissions users patch developers/123/users/test@example.com`
  - `gpd recovery add-targeting 123 --package com.milcgroup.one`
  - `gpd publish testers add --track internal --package com.milcgroup.one`
  - `gpd reviews reply --review-id dummy --rate-limit notaduration --text "x" --package com.milcgroup.one`
- Observations:
  - Validation errors now return `services` metadata for permissions, recovery, publish testers, and reviews reply.

#### Run: Vitals error search + testers dry-run
- Commands:
  - `gpd vitals errors issues search --package com.milcgroup.one --query "crash" --interval last30Days --page-size 10`
  - `gpd vitals errors reports search --package com.milcgroup.one --query "ANR" --interval last30Days --page-size 10 --deobfuscate`
  - `gpd publish testers remove --package com.milcgroup.one --track internal --group "testers@example.com" --dry-run`
- Observations:
  - Vitals error searches return structured validation errors with services metadata for invalid query restrictions.
  - Testers dry-run outputs structured result with services metadata.

#### Run: Validation spot checks (permissions/reviews/monetization/purchases)
- Commands:
  - `gpd permissions grants patch`
  - `gpd reviews response delete`
  - `gpd monetization offers batchUpdate dummy dummy`
  - `gpd purchases subscriptions revoke --token t --package com.milcgroup.one`
- Observations:
  - Validation errors return `services` metadata across these commands.

#### Run: Auth/config/version checks
- Commands:
  - `gpd auth status`
  - `gpd config print`
  - `gpd version`
- Observations:
  - Version output now includes `services: ["version"]`.

#### Run: Apps list + publish status/tracks
- Commands:
  - `gpd apps list --output json`
  - `gpd publish status --package com.example.fake --output json`
  - `gpd publish tracks list --package com.example.fake --output json`
- Observations:
  - Apps list returns app inventory with `services: ["playdeveloperreporting"]`.
  - Publish status/tracks now include `services: ["androidpublisher"]` on 404 errors.

### 2026-02-03

#### Run: Publish happy path (upload -> release -> rollout -> status)
- Commands:
  - `gpd publish upload ./artifacts/app.aab --package com.milcgroup.one`
  - `gpd publish release --package com.milcgroup.one --track production --status inProgress --version-code <version-code>`
  - `gpd publish rollout --package com.milcgroup.one --track production --percentage 10`
  - `gpd publish status --package com.milcgroup.one --track production`
- Observations:
  - Pending execution (requires a real AAB, version code, and production track).
- Improvements applied:
  - None (happy-path validation only).
  - Link: TODO (add commit/PR)

#### Run: Monetization happy path (create/update + list/get)
- Commands:
  - `gpd monetization products create --package com.milcgroup.one --product-id gpd.test.product --type managed --default-price 990000 --status active`
  - `gpd monetization products update gpd.test.product --package com.milcgroup.one --default-price 1990000 --status active`
  - `gpd monetization products list --package com.milcgroup.one --page-size 1`
  - `gpd monetization products get gpd.test.product --package com.milcgroup.one`
- Observations:
  - Pending execution (requires a real package with monetization access).
- Improvements applied:
  - None (happy-path validation only).
  - Link: TODO (add commit/PR)

#### Run: Reviews lifecycle (list -> get -> reply -> response get/delete)
- Commands:
  - `gpd reviews list --package com.milcgroup.one --page-size 1 --include-review-text`
  - `gpd reviews get --package com.milcgroup.one --review-id <review-id>`
  - `gpd reviews reply --package com.milcgroup.one --review-id <review-id> --text "Thanks for the feedback!"`
  - `gpd reviews response get --package com.milcgroup.one --review-id <review-id>`
  - `gpd reviews response delete --package com.milcgroup.one --review-id <review-id>`
- Observations:
  - Pending execution (requires a valid review ID and reply permissions).
- Improvements applied:
  - None (happy-path validation only).
  - Link: TODO (add commit/PR)

#### Assessment Summary
- Strength: Wide surface-area coverage with consistent error-shaping and hints.
- Gap: Success-path workflows are underrepresented (publish + monetization).
- Risk: "Improvements applied" entries are not linked to commits/PRs, making regressions hard to detect.

#### Follow-ups (Verification)
- [ ] Link each "Improvements applied" entry to a commit or PR.
- [ ] Execute the 2026-02-03 publish happy-path run and replace placeholders with real IDs.
- [ ] Execute the 2026-02-03 monetization happy-path run and replace placeholders with real IDs.
- [ ] Execute the 2026-02-03 reviews lifecycle run and replace placeholders with real IDs.
