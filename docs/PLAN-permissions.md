# Permissions Implementation Plan

**Status**: Not started
**Priority**: High — team management workflows
**Date**: 2026-03-01

## Background

All permission/grant management commands are stubs. These enable managing developer account users and their access levels programmatically — critical for organizations managing Play Console access via automation.

All stubs in `internal/cli/kong_permissions_recovery_integrity.go`.

## Commands to Implement

### Users (3 commands)

| Command | Line | API Method |
|---------|------|------------|
| `permissions users list` | 51 | `Users.List(parent)` |
| `permissions users add` | 33 | `Users.Create(parent, user)` |
| `permissions users remove` | 43 | `Users.Delete(name)` |

- `parent` = `developers/{developerId}`
- `list` returns all users with their emails and access levels
- `add` accepts `--email` and `--role` (admin, developer, etc.)
- `remove` accepts `--email` or user resource name

### Grants (3 commands)

| Command | Line | API Method |
|---------|------|------------|
| `permissions grants list` | 91 | `Grants.List(parent)` |
| `permissions grants add` | 70 | `Grants.Create(parent, grant)` |
| `permissions grants remove` | 81 | `Grants.Delete(name)` |

- Grants are per-app permission assignments
- `parent` = `developers/{developerId}/users/{userId}`
- `add` accepts `--app` (package name) and `--permissions` (list of permission strings)
- Google Play permissions: `ACCESS_MANAGED_RESOURCES`, `MANAGE_STORE_LISTING`, `MANAGE_PRODUCTION_RELEASES`, etc.
- `remove` accepts grant resource name or `--app` + `--user` to resolve it

### General (1 command)

| Command | Line | Notes |
|---------|------|-------|
| `permissions list` | 99 | List all available permission types |

- Informational — outputs the set of grantable permissions with descriptions
- No API call needed, can be hardcoded from API docs

## Notes

- Uses the Google Play Developer API v3 `Users` and `Grants` resources (not Edits)
- No edit transaction needed — these are direct API calls
- Developer ID is required — may need `--developer-id` flag or resolve from auth context
- Consider `--dry-run` for add/remove operations

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_permissions_recovery_integrity.go` | Implement `Run()` for all 7 commands |
| `internal/api/client.go` | Ensure Users/Grants service methods are accessible |

## Testing

- Mock Users.List/Create/Delete
- Mock Grants.List/Create/Delete
- Test user add with various roles
- Test grant add with multiple permissions
- Test remove with user not found
