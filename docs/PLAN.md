# Outstanding Improvement Plan for Google-Play-Developer-CLI (gpd)

Last reviewed: 2026-02-19

This plan consolidates open work from:
- `docs/cli-assessment.md`
- `docs/asc-parity.md`
- `docs/api-coverage-matrix.md`

Completed items already logged in `cli-assessment.md` are intentionally excluded.

## Evidence refresh (2026-02-19)

- ASC (remote): GitHub commits page could not be opened (cache miss), so ASC remote verification is still pending for this run.
- GPD: new mock API testing infrastructure and broader unit test coverage landed (`internal/apitest/mock_client.go`, `internal/api/client_test.go`, `internal/output/result_test.go`, `internal/cli/publish_upload_test.go`).
- GPD: auth/config normalization and timeout lifecycle handling were improved (`internal/config/config.go`, `internal/api/client.go`, `internal/cli/cli.go`).
- Delta: GPD moved from "no testing infrastructure" to "foundation in place", but golden output snapshots for regression detection are still missing.

## Matrix updates

| Parity target | ASC evidence | GPD evidence | Delta | Plan impact |
|---|---|---|---|---|
| Testing: CLI golden tests + fixtures | No new ASC testing-specific evidence found (remote commit details not accessible). | New mock API infra and expanded tests in `internal/apitest/mock_client.go` and `internal/api/client_test.go`. | Moved from "no infra" to "foundation in place"; golden snapshots still absent. | Elevate golden/fixture harness execution using the new mock infrastructure. |
| Auth/config precedence + strict conflict detection | No new ASC auth/config precedence evidence found (remote commit details not accessible). | Validation and normalization added in `internal/config/config.go`. | Partial improvement; strict conflict detection and precedence docs still missing. | Keep this as a gap and refine acceptance criteria to include explicit conflict cases. |

## Key gaps / highest ROI (current)

1. Golden/fixture regression harness for CLI outputs is still missing; extend mock API infra into snapshot tests.
2. Explicit bulk tester lifecycle parity (CSV import/export), or a clear non-goal declaration, remains undefined.
3. Play OpenAPI/spec snapshot policy is still missing (implement or document as a non-goal).
4. Auth/config precedence improved with validation, but strict conflict detection and explicit precedence docs are still missing.
5. Safety flags for destructive/irreversible operations remain uneven across commands.
6. JSON envelope output contract exists, but list/table/markdown golden coverage is still too narrow.
7. Docs cookbook workflows and troubleshooting templates remain partial.

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
- Gap: explicit bulk tester lifecycle parity is undefined (CSV import/export and invite-oriented flows), and unsupported semantics are not consistently marked as non-goals.
- Current: `gpd publish beta-groups list/get/create/update/delete/add-testers/remove-testers` compatibility commands + `gpd publish testers list/get/add/remove`.
- Source: `docs/asc-parity.md` (Beta Groups, Beta Testers: Partial).
- Exit criteria:
  - Add bulk tester lifecycle commands/docs equivalent to ASC CSV workflows where Play API supports it.
  - If unsupported by Play API, explicitly declare CSV parity as an intentional non-goal with rationale.
  - Keep track-mapping semantics explicit for unsupported ASC concepts.
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
  - Add strict conflict detection for overlapping auth/config sources and document precedence in one canonical location.

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

4. Add a CLI output regression harness (golden/fixture snapshots) for high-churn list/table/markdown commands.

5. Decide whether to add and maintain a Play API/OpenAPI snapshot process (or explicitly mark as intentional non-goal).

6. Normalize destructive-operation safety flags (`--confirm` / `--dry-run`) across commands that mutate or irreversibly delete resources.

7. Expand cookbook workflows and troubleshooting templates for parity-critical flows (auth precedence, tester lifecycle, output format validation).
