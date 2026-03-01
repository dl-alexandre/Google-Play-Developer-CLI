# Analytics, Apps & Games Implementation Plan

**Status**: Not started
**Priority**: Low — specialized use cases
**Date**: 2026-03-01

## Background

Analytics query, app discovery, and games management commands are stubs. Analytics and apps are useful but lower priority since vitals commands already cover the most common metrics. Games commands serve a niche audience.

All stubs in `internal/cli/kong_analytics_apps_games.go`.

## Analytics Commands (2)

| Command | Line | API Method |
|---------|------|------------|
| `analytics query` | 31 | Play Developer Reporting API — custom metrics query |
| `analytics capabilities` | 39 | List available metrics and dimensions |

### Details

- **`analytics query`**: Run custom analytics queries against the Play Developer Reporting API
  - `--metrics` — comma-separated metric names
  - `--dimensions` — comma-separated dimension names
  - `--start-date`, `--end-date` — date range
  - `--filters` — dimension filters (JSON)
  - Uses the same reporting API as vitals but with user-specified metrics
  - Return tabular data in JSON envelope

- **`analytics capabilities`**: List available metrics, dimensions, and filter options
  - Informational, helps users build queries
  - Could be generated from API discovery or hardcoded from docs

### Notes

- The Play Developer Reporting API is already initialized (`PlayReporting()` service)
- Vitals commands already use this API for crashes/ANRs/errors — analytics query generalizes it
- Consider supporting `--format csv` properly here (currently broken in vitals too)

## Apps Commands (2)

| Command | Line | API Method |
|---------|------|------------|
| `apps list` | 61 | List apps accessible to the authenticated account |
| `apps get` | 71 | Get details for a specific app |

### Details

- **`apps list`**: List all applications the service account/user has access to
  - No direct "list apps" endpoint in Android Publisher API
  - Options: use `Reviews.List()` across known packages, or require developer ID and use the developer API
  - May need to use a different approach — check API capabilities
  - Useful for discovery: "what packages can I manage?"

- **`apps get`**: Get app details (title, description, icon, etc.)
  - Maps to `Edits.Details.Get()` for the specified `--package`
  - Could also aggregate info from multiple endpoints (details + tracks + latest release)

### Notes

- `apps list` is tricky — the Android Publisher API doesn't have a straightforward "list all apps" endpoint
- May need to document limitations or use alternative approaches

## Games Commands (6)

| Command | Line | API Method |
|---------|------|------------|
| `games achievements reset` | 101 | `GamesManagement.Achievements.Reset(achievementId)` |
| `games scores reset` | 118 | `GamesManagement.Scores.Reset(leaderboardId)` |
| `games events reset` | 135 | `GamesManagement.Events.Reset(eventId)` |
| `games players hide` | 152 | `GamesManagement.Players.Hide(applicationId, playerId)` |
| `games players unhide` | 163 | `GamesManagement.Players.Unhide(applicationId, playerId)` |
| `games capabilities` | 171 | Informational |

### Details

- Reset commands clear test data during development (achievements, leaderboards, events)
- `--achievement-id`, `--leaderboard-id`, `--event-id` for respective commands
- `--for-all` flag to reset all (e.g., `Achievements.ResetAll()`)
- Player hide/unhide manages visibility in leaderboards
- These use the Games Management API (separate from Android Publisher)

### Notes

- Games Management API is being deprecated in favor of Play Games Services v2
- Consider checking API availability before implementing
- Low priority unless there's specific demand

## Implementation Order

1. `apps get` — useful, straightforward
2. `analytics capabilities` — informational, helps users
3. `analytics query` — custom metrics access
4. `apps list` — investigate API limitations first
5. Games commands — only if demand exists

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_analytics_apps_games.go` | Implement `Run()` for all 10 commands |
| `internal/api/client.go` | Verify GamesManagement service methods accessible |

## Testing

- Mock PlayReporting query/capabilities responses
- Mock GamesManagement reset/hide/unhide operations
- Test analytics query with various metric/dimension combinations
- Test games reset with --for-all flag
