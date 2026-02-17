# Outstanding Improvement Plan for Google-Play-Developer-CLI (gpd)

This plan consolidates open work from:
- `docs/cli-assessment.md`
- `docs/asc-parity.md`
- `docs/api-coverage-matrix.md`

Completed items already logged in `cli-assessment.md` are intentionally excluded.

## Priority 1: API Coverage Gap

1. Add LMK rate support in vitals commands when the Play Developer Reporting API exposes it.
- Gap: `gpd vitals lmk-rate` is currently marked unsupported.
- Source: `docs/api-coverage-matrix.md` (Play Developer Reporting API v1beta1 / LMK Rate).
- Exit criteria:
  - Implement command wiring and API integration when endpoint is available.
  - Add tests and docs updates.
  - Remove unsupported marker from the coverage matrix.

## Priority 2: ASC Parity Gaps (High-Value, Feasible)

1. Strengthen beta testing workflows beyond track-scoped tester operations.
- Gap: No ASC-style beta group lifecycle parity (group CRUD, richer assignment flows).
- Current: `gpd publish testers list/get/add/remove` only.
- Source: `docs/asc-parity.md` (Beta Groups, Beta Testers: Partial).
- Exit criteria:
  - Define Play-equivalent abstractions and UX for group-like management where possible.
  - Document supported and unsupported semantics explicitly.

2. Expand review response parity handling.
- Gap: response delete path is documented as unsupported.
- Current: list/get/reply/response get/for-review/delete command set exists, but parity notes still indicate a gap.
- Source: `docs/asc-parity.md` (App Store reviews: Partial).
- Exit criteria:
  - Verify actual API capability and command behavior.
  - Either implement missing behavior or document hard platform limitation with exact guidance.

3. Improve submission/release workflow parity documentation and ergonomics.
- Gap: release/submission mapping is partial due workflow differences.
- Current: `publish release/rollout/promote/halt/rollback/status`.
- Source: `docs/asc-parity.md` (Versions, Submit: Partial).
- Exit criteria:
  - Publish a clear workflow mapping doc for common ASC submit/release journeys.
  - Add examples for staged rollout, halt, rollback, and promotion decision paths.

## Priority 3: Parity Gaps (Platform-Limit or UX Clarification)

1. Clarify authentication parity boundaries.
- Gap: ASC browser login flow parity is partial.
- Current: device-code OAuth + service account model.
- Source: `docs/asc-parity.md` (Authentication: Partial).
- Exit criteria:
  - Document explicit auth decision tree and migration guidance for ASC users.

2. Clarify app/build model differences.
- Gap: no global build registry and no build-level beta group assignment equivalent.
- Source: `docs/asc-parity.md` (Apps & Builds: Partial).
- Exit criteria:
  - Document limitations and recommended Play-native alternatives.

3. Clarify analytics/reporting scope differences.
- Gap: ASC analytics/sales scope is broader than current Play Reporting coverage.
- Source: `docs/asc-parity.md` (Analytics & Sales: Partial).
- Exit criteria:
  - Document supported datasets and unsupported ASC analogs in one place.

4. Clarify metadata/localization workflow differences.
- Gap: app setup, app info, localizations remain partial due model differences.
- Source: `docs/asc-parity.md` (App Setup, App Info, Localizations: Partial).
- Exit criteria:
  - Add side-by-side task mapping examples (ASC task -> gpd command sequence).

## Backlog Hygiene

1. Keep `docs/api-coverage-matrix.md` and `docs/asc-parity.md` synchronized after every feature change.
2. For each open parity item, mark one of:
- Implementable now
- Blocked by Google Play API capability
- Intentional non-goal

3. Add explicit target status and owner fields when converting this plan into execution tickets.
