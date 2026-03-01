# Recovery & Integrity Implementation Plan

**Status**: Not started
**Priority**: Medium — incident response tooling
**Date**: 2026-03-01

## Background

App recovery and Play Integrity commands are stubs. Recovery commands enable responding to critical issues (force-update users to a patched version). Integrity decoding validates device/app attestation tokens.

All stubs in `internal/cli/kong_permissions_recovery_integrity.go`.

## Recovery Commands (4)

| Command | Line | API Method |
|---------|------|------------|
| `recovery list` | 121 | `Apprecovery.List(packageName)` |
| `recovery create` | 133 | `Apprecovery.Create(packageName, recoveryAction)` |
| `recovery deploy` | 143 | `Apprecovery.Deploy(packageName, recoveryId)` |
| `recovery cancel` | 154 | `Apprecovery.Cancel(packageName, recoveryId)` |

### Details

- **`recovery list`**: List active/past recovery actions for the app
  - Show: recovery ID, status, target version codes, affected users estimate
  - Support `--status` filter (active, deployed, cancelled)

- **`recovery create`**: Create a new recovery action
  - `--target-version-codes` (which versions to target)
  - `--recovery-type` (e.g., `FORCE_UPDATE`)
  - `--targeting` (device/OS filters)
  - Returns recovery ID for subsequent deploy/cancel

- **`recovery deploy`**: Deploy a created recovery action to affected users
  - `--recovery-id` required
  - Confirm action (prompts by default, `--yes` to skip)
  - High-impact operation — should warn user

- **`recovery cancel`**: Cancel an active recovery action
  - `--recovery-id` required

### Notes

- Recovery API is relatively new — check API availability
- These are direct API calls, no edit transaction
- Deploy is a high-impact operation — always confirm unless `--yes`

## Integrity Command (1)

| Command | Line | Notes |
|---------|------|-------|
| `integrity decode` | 174 | `PlayIntegrity.DecodeIntegrityToken(packageName, token)` |

### Details

- Decodes a Play Integrity attestation token
- `--token` flag (the integrity token from client)
- Returns: device integrity verdict, app integrity verdict, account details, request details
- Useful for server-side validation of client attestations

### Notes

- Requires Play Integrity API to be enabled in Google Cloud Console
- Token is a JWS — API decrypts and verifies server-side
- Consider `--raw` flag to output full decoded payload

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_permissions_recovery_integrity.go` | Implement `Run()` for all 5 commands |
| `internal/api/client.go` | Ensure AppRecovery and PlayIntegrity services are accessible |

## Testing

- Mock Apprecovery.List/Create/Deploy/Cancel
- Mock PlayIntegrity.DecodeIntegrityToken
- Test recovery lifecycle: create → deploy → cancel
- Test integrity decode with valid/invalid tokens
- Test deploy confirmation prompt behavior
