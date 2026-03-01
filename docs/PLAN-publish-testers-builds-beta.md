# Publish Testers, Builds & Beta Groups Implementation Plan

**Status**: Not started
**Priority**: High — enables testing workflows
**Date**: 2026-03-01

## Background

Tester management, build lifecycle, and beta group commands are all stubs. These are essential for teams managing testing workflows from CI/CD.

All stubs in `internal/cli/kong_publish.go`.

## Commands to Implement

### Testers (4 commands)

| Command | Line | API Method |
|---------|------|------------|
| `publish testers list` | 1094 | `Edits.Testers.Get(track)` |
| `publish testers get` | 1104 | `Edits.Testers.Get(track)` |
| `publish testers add` | 1070 | `Edits.Testers.Update(track, testers)` |
| `publish testers remove` | 1084 | `Edits.Testers.Update(track, testers)` |

- `--track` flag (which test track to manage testers for)
- `--email` flag (single or comma-separated list)
- `add`/`remove` fetch current list, modify, then update (no atomic add/remove in API)
- Google Groups supported via `googleGroups` field
- All require edit transaction

### Builds (4 commands)

| Command | Line | Notes |
|---------|------|-------|
| `publish builds list` | 1123 | List uploaded APKs/bundles with version codes |
| `publish builds get` | 1135 | Get details for a specific version code |
| `publish builds expire` | 1149 | Expire a specific build |
| `publish builds expire-all` | 1162 | Expire all builds on a track |

- `list` uses `Edits.Apks.List()` and `Edits.Bundles.List()`
- `get` fetches details for `--version-code`
- `expire` marks builds as not servable
- Read operations don't need commit; expire operations do

### Beta Groups (7 commands)

| Command | Line | Notes |
|---------|------|-------|
| `publish beta-groups list` | 1183 | List all test tracks/groups |
| `publish beta-groups get` | 1193 | Get details for a track |
| `publish beta-groups create` | 1207 | Create closed test track |
| `publish beta-groups update` | 1221 | Update track settings |
| `publish beta-groups delete` | 1231 | Remove test track |
| `publish beta-groups add-testers` | 1245 | Add testers to group |
| `publish beta-groups remove-testers` | 1259 | Remove testers from group |

- Maps to closed testing tracks in Google Play
- `create` uses `Edits.Tracks.Update()` with a new track config
- Tester operations overlap with `publish testers` — beta-groups scopes it to a named group
- All mutating ops require edit transaction

## Notes

- Google Play's tester model differs from App Store Connect — testers are managed per-track, not as standalone groups
- Beta groups map to closed testing tracks; open/internal testing are separate
- Consider whether `beta-groups` commands should be aliases for track-scoped tester operations

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_publish.go` | Implement `Run()` for all 15 commands |

## Testing

- Mock Edits.Testers.Get/Update and Edits.Apks.List/Bundles.List
- Test add/remove tester flow (fetch → modify → update)
- Test build listing with mixed APK/AAB results
- Test beta group lifecycle: create → add testers → remove testers → delete
