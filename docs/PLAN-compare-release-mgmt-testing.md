# Compare, Release Management & Testing Implementation Plan

**Status**: Not started
**Priority**: Low — advanced/composite features, depends on underlying commands
**Date**: 2026-03-01

## Background

These are higher-level composite commands that aggregate data from multiple API calls. They depend on the underlying basic commands being implemented first. All return NoOp stubs today.

## Compare Commands (4)

File: `internal/cli/kong_compare.go`

| Command | Line | Dependencies |
|---------|------|-------------|
| `compare vitals` | 102 | Vitals commands (implemented) |
| `compare reviews` | 179 | Reviews commands (implemented) |
| `compare releases` | 268 | Publish status (NOT implemented) |
| `compare subscriptions` | 337 | Monetization commands (NOT implemented) |

### Details

- Compare metrics **across multiple packages** side by side
- `--packages` flag (comma-separated list of package names)
- `--period` for time range comparison
- Each command queries the respective API for each package and assembles a comparison table

### Blockers

- `compare vitals` and `compare reviews` could be implemented now (underlying APIs work)
- `compare releases` needs `publish status` first
- `compare subscriptions` needs monetization commands first

## Release Management Commands (5)

File: `internal/cli/kong_release_mgmt.go`

| Command | Line | Dependencies |
|---------|------|-------------|
| `release-mgmt calendar` | 95 | Publish status, track listing |
| `release-mgmt conflicts` | 166 | Publish status (version codes across tracks) |
| `release-mgmt strategy` | 246 | Vitals + publish status |
| `release-mgmt history` | 340 | Track history + optional vitals |
| `release-mgmt notes` | 416 | Edits.Listings (release notes are per-track per-locale) |

### Details

- **`calendar`**: Timeline view of upcoming/past releases across tracks
- **`conflicts`**: Detect version code collisions between tracks
- **`strategy`**: AI-assisted rollout recommendations based on crash rates/ANRs (suggest rollback vs continue)
- **`history`**: Detailed release history with optional vitals overlay
- **`notes`**: Manage release notes — get/set/copy across locales for a track

### Blockers

- All depend on `publish status` and track listing being implemented
- `strategy` is the most complex — needs vitals correlation with release timeline
- `notes` depends on listing/metadata commands

## Testing Commands (4)

File: `internal/cli/kong_testing.go`

| Command | Line | Notes |
|---------|------|-------|
| `testing prelaunch` | 91 | API-limited — Play Console UI only |
| `testing device-lab` | 164 | Needs Firebase Test Lab integration |
| `testing screenshots` | 221 | Needs Firebase Test Lab integration |
| `testing compatibility` | 395 | Needs device catalog API |

### Details

- **`testing prelaunch`**: Pre-launch reports from Google Play automated testing
  - API access is limited; most data only available in Play Console UI
  - Could potentially surface basic pass/fail status

- **`testing device-lab`**: Run tests on Firebase Test Lab
  - Requires separate Firebase authentication
  - `gcloud firebase test` already exists — may not be worth duplicating
  - Consider wrapping `gcloud` instead of reimplementing

- **`testing screenshots`**: Capture screenshots across devices/locales
  - Depends on device-lab integration
  - Alternative: integrate with Fastlane screengrab

- **`testing compatibility`**: Check device compatibility
  - Uses device catalog to validate APK/AAB compatibility
  - `Edits.Devicetierconfigs` API

### Recommendation

- `testing prelaunch`: Implement basic status check, document limitations
- `testing device-lab` and `testing screenshots`: Consider deferring — `gcloud` covers this better
- `testing compatibility`: Implement if device catalog API is accessible

## Implementation Order

1. `compare vitals` + `compare reviews` — no blockers, underlying APIs work
2. `release-mgmt notes` — after listing commands land
3. `release-mgmt calendar` + `history` — after publish status lands
4. `release-mgmt conflicts` — after publish status lands
5. `compare releases` — after publish status lands
6. `testing prelaunch` (basic) + `testing compatibility`
7. `release-mgmt strategy` — most complex, last
8. `compare subscriptions` — after monetization lands
9. `testing device-lab` + `testing screenshots` — evaluate if worth doing vs deferring to gcloud

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_compare.go` | Implement `Run()` for 4 commands |
| `internal/cli/kong_release_mgmt.go` | Implement `Run()` for 5 commands |
| `internal/cli/kong_testing.go` | Implement `Run()` for 2-4 commands |

## Testing

- Mock multi-package API calls for compare commands
- Test comparison output formatting
- Test release calendar date range handling
- Test conflict detection logic with overlapping version codes
- Test release notes locale management
