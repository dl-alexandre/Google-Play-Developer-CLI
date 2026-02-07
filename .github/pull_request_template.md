## Summary

<!-- Describe what changed and why. -->

## Risk and Scope

- [ ] No breaking CLI command/flag changes, or clearly documented if intentional.
- [ ] Auth and permissions behavior reviewed for regressions.
- [ ] Error output/JSON contract remains backward compatible.

## Required Quality Gates

- [ ] `make test`
- [ ] `make build-all`
- [ ] `make lint`
- [ ] `make security`

## Validation

- [ ] Manual smoke check completed (example: `./bin/gpd version`).
- [ ] New/updated tests included for changed behavior.
- [ ] Docs updated (README/CHANGELOG/help text) when user-visible behavior changed.

## Release Readiness

- [ ] Version metadata injection remains valid (`Version`, `GitCommit`, `BuildTime`).
- [ ] Checksum behavior unchanged or intentionally updated.
- [ ] CI `test/build/lint/security` expected to pass.
