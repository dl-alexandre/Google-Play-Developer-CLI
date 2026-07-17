# CLI Structure (Kong + domain packages)

This document describes how the gpd CLI is organized today and the incremental plan for moving toward domain packages under `internal/cli/<domain>/`.

## Current Kong layout

Entry point: `cmd/gpd` → `cli.RunKongCLI()` in [`internal/cli/kong_cli.go`](../../internal/cli/kong_cli.go).

| Layer | Location | Role |
|-------|----------|------|
| Root + globals | `kong_cli.go` | `KongCLI`, `Globals`, parser setup, profile/output resolution, command execution |
| Command groups | `kong_*.go` | Nested Kong command structs (`AuthCmd`, `PublishCmd`, …) and `Run` methods |
| Shared helpers | `kong_helpers.go`, `pagination.go`, `progress.go`, … | Cross-command utilities still in package `cli` |
| Domain packages | `internal/cli/<domain>/` | Extracted pure helpers / focused subsystems (growing) |

### Registration model

- Command **types and handlers** live in `kong_<group>.go` files in package `cli`.
- Command **registration** is declarative: fields on `KongCLI` in `kong_cli.go` (Kong struct tags).
- Global flags (`--package`, `--output`, `--profile`, …) live on `Globals` and are injected into every `Run(globals *Globals)`.

Example shape:

```go
// kong_cli.go
type KongCLI struct {
    Globals
    Auth    AuthCmd    `cmd:"" help:"Authentication commands"`
    Publish PublishCmd `cmd:"" help:"Publishing commands"`
    // ...
}

// kong_auth.go (same package cli)
type AuthCmd struct {
    Status AuthStatusCmd `cmd:"" help:"Check authentication status"`
}
```

There is no separate “register commands” function; adding a top-level command means adding a field on `KongCLI` and implementing the nested types’ `Run` methods.

## Rule: new command families use domain packages

**Prefer this for new work:**

1. Put pure logic, plans, and non-Kong helpers in `internal/cli/<domain>/` (own package name, e.g. `outfmt`, `playship`).
2. Keep thin Kong adapter types in `internal/cli/kong_<domain>.go` (or extend an existing `kong_*.go`) that call into the domain package.
3. Register the command group on `KongCLI` in `kong_cli.go` only when introducing a new top-level command.

### Reference extraction: `outfmt`

TTY-aware default output resolution was moved out of package `cli` as a low-churn example:

- Package: [`internal/cli/outfmt`](../../internal/cli/outfmt)
- API: `ResolveOutputFromEnvAndTTY`, `ResolveDefaultOutput`, `OutputFlagSet`, `EnvDefaultOutput`
- Caller: `RunKongCLI` in `kong_cli.go` imports `outfmt` and sets `cli.Output`

Behavior is unchanged: explicit `--output` > `GPD_DEFAULT_OUTPUT` > `table` on TTY / `json` in pipes/CI.

### What stays in package `cli` (for now)

- Large existing groups (`kong_publish.go`, `kong_auth.go`, `kong_validate.go`, …)
- Types that embed Kong tags and need access to unexported package helpers
- Cross-cutting wiring (auth globals, extension help, update checks)

Do **not** rewrite multi-thousand-line files in one pass. Extract pure helpers first; move command structs later if needed.

## Migration plan (auth / publish over time)

Full split is **incremental**. Suggested order:

### Phase 0 — Pattern established (this doc + `outfmt`)

- Domain package layout documented.
- One real extraction (`outfmt`) proves import/test patterns.

### Phase 1 — Pure helpers from publish / play (done for playship)

Package [`internal/cli/playship`](../../internal/cli/playship):

- `ResolveReleaseParams`, `BuildTrackRelease`, `BuildPlan`
- Kong adapter: `PublishPlayCmd` in `kong_publish.go` calls `playship.*`

Further candidates when touch risk is low:
- `buildPlayTrackRelease`
- `buildPublishPlayPlan`

Target package name (suggested): `internal/cli/playship/` (or `play`).

Constraints:

- Keep unit tests green (`go test -tags=unit ./internal/cli/...`).
- Prefer moving pure functions first; leave `Run` methods in `kong_publish.go`.
- Avoid simultaneous large edits to `kong_auth.go` / `kong_validate.go` when those files are under concurrent work.

### Phase 2 — Auth domain package

- Extract non-UI auth helpers (profile resolution wiring already leans on `internal/config` and `internal/auth`).
- Leave Kong command structs in `kong_auth.go` until helpers are stable.
- `internal/auth` remains the credential/token subsystem; CLI domain packages only adapt CLI concerns.

### Phase 3 — Optional command-struct relocation

Only if package `cli` remains hard to navigate:

- Move selected `*Cmd` types into domain packages **if** Kong + globals injection stays clean.
- Re-export or thin-wrap in `kong_cli.go` / `kong_*.go` as needed.
- No user-facing flag or subcommand path changes.

## Testing expectations

```bash
# Package-scoped unit tests (preferred for domain packages)
go test -tags=unit ./internal/cli/outfmt/...

# Full CLI unit suite after extractions
go test -tags=unit ./internal/cli/...
```

Domain package tests should use `//go:build unit` consistently with existing CLI tests.

## Conflict-minimization guidelines

| Do | Don’t |
|----|--------|
| Extract small, pure helpers into new packages | Rewrite all of `kong_publish.go` in one PR |
| Update only direct callers of moved symbols | Drive-by renames across auth/validate |
| Add docs for the pattern when establishing it | Duplicate env/format constants in many places without reason |
| Keep user-facing CLI behavior identical | Change flag defaults or help text as a side effect of moves |

## Related docs

- [KONG_MIGRATION.md](../KONG_MIGRATION.md) — original Cobra → Kong migration notes
- [AGENTS.md](./AGENTS.md) — agent-oriented project guide
- [COMMANDS.md](../COMMANDS.md) — generated/command reference
