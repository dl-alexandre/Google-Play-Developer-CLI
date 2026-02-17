# Edit Workflow Documentation

## Overview

Edits are **transactional units** in the Google Play Developer API that allow you to make multiple changes to your app before committing them. Think of an edit as a "staging area" where you can:

- Upload multiple artifacts (AABs/APKs)
- Update store listings
- Modify release tracks
- Change app details
- Upload images and assets

All changes are grouped together and only applied when you **commit** the edit. This provides several benefits:

1. **Atomicity**: All changes succeed or fail together
2. **Validation**: Test your changes before committing
3. **Rollback**: Delete an edit to discard all changes
4. **Multi-step workflows**: Build complex releases across multiple commands

The Google Play Developer CLI (`gpd`) provides comprehensive edit lifecycle management with automatic edit creation, local caching, and expiration handling.

## Basic Usage

### Create an Edit

Create a new edit transaction:

```bash
gpd publish edit create --package com.example.app
```

**Output:**
```json
{
  "data": {
    "editId": "12345678901234567890",
    "package": "com.example.app",
    "createdAt": "2024-01-15T10:30:00Z",
    "lastUsedAt": "2024-01-15T10:30:00Z",
    "state": "draft"
  }
}
```

The `editId` is automatically saved locally and can be reused with the `--edit-id` flag.

### List Edits

View all cached edits for your package:

```bash
gpd publish edit list --package com.example.app
```

**Output:**
```json
{
  "data": {
    "edits": [
      {
        "editId": "12345678901234567890",
        "handle": "12345678901234567890",
        "package": "com.example.app",
        "createdAt": "2024-01-15T10:30:00Z",
        "lastUsedAt": "2024-01-15T10:30:00Z",
        "state": "draft",
        "expired": false
      }
    ],
    "count": 1,
    "package": "com.example.app"
  }
}
```

### Get Edit Details

Inspect a specific edit's local and remote state:

```bash
gpd publish edit get 12345678901234567890 --package com.example.app
```

**Output:**
```json
{
  "data": {
    "editId": "12345678901234567890",
    "local": {
      "handle": "12345678901234567890",
      "serverId": "12345678901234567890",
      "packageName": "com.example.app",
      "createdAt": "2024-01-15T10:30:00Z",
      "lastUsedAt": "2024-01-15T10:30:00Z",
      "state": "draft"
    },
    "remote": {
      "id": "12345678901234567890",
      "expiryTimeSeconds": "1736966400"
    },
    "package": "com.example.app"
  }
}
```

### Validate an Edit

Check if an edit is valid before committing:

```bash
gpd publish edit validate 12345678901234567890 --package com.example.app
```

**Output:**
```json
{
  "data": {
    "success": true,
    "editId": "12345678901234567890",
    "package": "com.example.app"
  }
}
```

Validation checks for:
- Required fields are present
- Version codes are valid
- Track configurations are correct
- No conflicting changes

### Commit an Edit

Apply all changes in an edit to your app:

```bash
gpd publish edit commit 12345678901234567890 --package com.example.app
```

**Output:**
```json
{
  "data": {
    "success": true,
    "editId": "12345678901234567890",
    "package": "com.example.app"
  }
}
```

⚠️ **Warning**: Committing an edit is **irreversible**. Once committed, changes are live in the Play Console.

### Delete an Edit

Discard an edit and all its changes:

```bash
gpd publish edit delete 12345678901234567890 --package com.example.app
```

**Output:**
```json
{
  "data": {
    "success": true,
    "editId": "12345678901234567890",
    "package": "com.example.app"
  }
}
```

This deletes the edit both locally and remotely, discarding all uncommitted changes.

## Advanced Workflows

### Multi-Step Release Workflow

Build a complete release with multiple operations in a single edit:

```bash
# 1. Create an edit (or use existing)
EDIT_ID=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.editId')

# 2. Upload AAB
gpd publish upload app.aab \
  --package com.example.app \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# 3. Update store listing
gpd publish listing update \
  --package com.example.app \
  --locale en-US \
  --title "My Awesome App v2.0" \
  --short-description "New features and improvements" \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# 4. Create release on internal track
gpd publish release \
  --package com.example.app \
  --track internal \
  --status draft \
  --version-code 42 \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# 5. Validate before committing
gpd publish edit validate "$EDIT_ID" --package com.example.app

# 6. Commit all changes atomically
gpd publish edit commit "$EDIT_ID" --package com.example.app
```

This workflow ensures all changes are applied together. If any step fails, you can delete the edit and start over.

### Using `--edit-id` Flag

Reuse an existing edit across multiple commands:

```bash
# Create edit once
EDIT_ID="my-release-v2"

# Use the same edit for multiple operations
gpd publish upload app.aab \
  --package com.example.app \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

gpd publish release \
  --package com.example.app \
  --track beta \
  --status completed \
  --version-code 42 \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# Commit when ready
gpd publish edit commit "$EDIT_ID" --package com.example.app
```

The `--edit-id` flag allows you to:
- Use a custom handle/name for your edit
- Share edits across scripts or CI/CD pipelines
- Resume work on an existing edit

### Using `--no-auto-commit` Flag

By default, most publish commands automatically commit edits after completion. Use `--no-auto-commit` to keep edits open:

```bash
# Upload without auto-commit
gpd publish upload app.aab \
  --package com.example.app \
  --no-auto-commit

# The edit remains open for additional changes
# You can add more operations or commit manually later
```

**When to use `--no-auto-commit`:**
- Building multi-step releases
- Testing changes before committing
- Coordinating changes across team members
- CI/CD pipelines that need validation steps

### Atomic Release with Rollback on Failure

Create a robust release script with error handling:

```bash
#!/bin/bash
set -e

PACKAGE="com.example.app"
TRACK="production"

# Create edit
EDIT_ID=$(gpd publish edit create --package "$PACKAGE" --output json | jq -r '.data.editId')
echo "Created edit: $EDIT_ID"

# Function to cleanup on error
cleanup() {
  echo "Error occurred, cleaning up edit $EDIT_ID"
  gpd publish edit delete "$EDIT_ID" --package "$PACKAGE" || true
  exit 1
}

trap cleanup ERR

# Upload AAB
echo "Uploading AAB..."
gpd publish upload app.aab \
  --package "$PACKAGE" \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# Create release
echo "Creating release..."
gpd publish release \
  --package "$PACKAGE" \
  --track "$TRACK" \
  --status inProgress \
  --version-code 42 \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# Validate
echo "Validating edit..."
if ! gpd publish edit validate "$EDIT_ID" --package "$PACKAGE" --output json | jq -e '.data.success' > /dev/null; then
  echo "Validation failed!"
  cleanup
fi

# Commit
echo "Committing edit..."
gpd publish edit commit "$EDIT_ID" --package "$PACKAGE"

echo "Release successful!"
```

This script ensures that if any step fails, the edit is automatically deleted, preventing partial releases.

### Staged Rollout Workflow

Gradually roll out a release using edits:

```bash
# Create edit for staged rollout
EDIT_ID=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.editId')

# Upload and create release with 5% rollout
gpd publish upload app.aab \
  --package com.example.app \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

gpd publish release \
  --package com.example.app \
  --track production \
  --status inProgress \
  --version-code 42 \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# Start with 5% rollout
gpd publish rollout \
  --package com.example.app \
  --track production \
  --percentage 5 \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# Commit
gpd publish edit commit "$EDIT_ID" --package com.example.app

# Later, increase to 50% (creates new edit automatically)
gpd publish rollout \
  --package com.example.app \
  --track production \
  --percentage 50

# Finally, complete rollout
gpd publish rollout \
  --package com.example.app \
  --track production \
  --percentage 100
```

## Best Practices

### 1. Always Validate Before Committing

```bash
# Validate first
gpd publish edit validate "$EDIT_ID" --package com.example.app

# Only commit if validation succeeds
if [ $? -eq 0 ]; then
  gpd publish edit commit "$EDIT_ID" --package com.example.app
fi
```

### 2. Use Descriptive Edit IDs

Use meaningful names for your edits:

```bash
# Good: Descriptive name
gpd publish upload app.aab \
  --package com.example.app \
  --edit-id "release-v2.0.0-beta" \
  --no-auto-commit

# Bad: Random or unclear names
gpd publish upload app.aab \
  --package com.example.app \
  --edit-id "temp123" \
  --no-auto-commit
```

### 3. Clean Up Unused Edits

Edits expire after 7 days of creation or 1 hour of inactivity, but manually delete unused edits:

```bash
# List all edits
gpd publish edit list --package com.example.app

# Delete expired or unused edits
gpd publish edit delete "$EDIT_ID" --package com.example.app
```

### 4. Handle Edit Expiration

Edits expire after:
- **7 days** from creation
- **1 hour** of inactivity

Check expiration status:

```bash
gpd publish edit list --package com.example.app | jq '.data.edits[] | select(.expired == true)'
```

If an edit expires, create a new one and redo your changes.

### 5. Use Dry-Run for Testing

Test your workflow without making changes:

```bash
gpd publish upload app.aab \
  --package com.example.app \
  --edit-id "$EDIT_ID" \
  --dry-run

gpd publish release \
  --package com.example.app \
  --track internal \
  --status draft \
  --version-code 42 \
  --edit-id "$EDIT_ID" \
  --dry-run
```

### 6. Coordinate Team Edits

When multiple team members work on the same app:

1. Use descriptive `--edit-id` values with your name/feature
2. Check for existing edits before creating new ones
3. Delete your edits when done
4. Use locks (automatic) to prevent conflicts

```bash
# Check existing edits
gpd publish edit list --package com.example.app

# Use your own edit ID
gpd publish upload app.aab \
  --package com.example.app \
  --edit-id "alice-feature-x" \
  --no-auto-commit
```

### 7. CI/CD Integration

In CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Create Edit
  id: edit
  run: |
    EDIT_ID=$(gpd publish edit create --package ${{ env.PACKAGE }} --output json | jq -r '.data.editId')
    echo "edit_id=$EDIT_ID" >> $GITHUB_OUTPUT

- name: Upload AAB
  run: |
    gpd publish upload app/build/outputs/bundle/release/app.aab \
      --package ${{ env.PACKAGE }} \
      --edit-id ${{ steps.edit.outputs.edit_id }} \
      --no-auto-commit

- name: Create Release
  run: |
    gpd publish release \
      --package ${{ env.PACKAGE }} \
      --track internal \
      --status draft \
      --version-code ${{ github.run_number }} \
      --edit-id ${{ steps.edit.outputs.edit_id }} \
      --no-auto-commit

- name: Validate and Commit
  run: |
    gpd publish edit validate ${{ steps.edit.outputs.edit_id }} --package ${{ env.PACKAGE }}
    gpd publish edit commit ${{ steps.edit.outputs.edit_id }} --package ${{ env.PACKAGE }}
```

## Error Handling

### Common Errors

#### Edit Not Found

**Error:**
```
failed to get edit: edit not found
```

**Solution:**
- Check if the edit ID is correct
- Verify the edit hasn't expired
- List all edits: `gpd publish edit list --package com.example.app`

#### Edit Expired

**Error:**
```
edit has expired
```

**Solution:**
- Edits expire after 7 days or 1 hour of inactivity
- Create a new edit and redo your changes
- Check expiration: `gpd publish edit list --package com.example.app | jq '.data.edits[] | select(.expired == true)'`

#### Validation Failed

**Error:**
```
failed to validate edit: validation error
```

**Solution:**
- Check required fields are present
- Verify version codes are valid
- Ensure track configurations are correct
- Review the edit contents: `gpd publish edit get "$EDIT_ID" --package com.example.app`

#### Lock Timeout

**Error:**
```
file lock timeout
```

**Solution:**
- Another process is using an edit for this package
- Wait for the other process to finish
- Check for stale locks (they auto-expire after 4 hours)
- Kill the blocking process if necessary

#### Commit Failed

**Error:**
```
failed to commit edit: conflict
```

**Solution:**
- Another edit may have been committed
- Refresh your edit state: `gpd publish edit get "$EDIT_ID" --package com.example.app`
- Validate the edit: `gpd publish edit validate "$EDIT_ID" --package com.example.app`
- Create a new edit if conflicts persist

### Error Recovery Patterns

#### Pattern 1: Retry with New Edit

```bash
# Attempt operation
if ! gpd publish upload app.aab --package com.example.app --edit-id "$EDIT_ID" --no-auto-commit; then
  # Delete failed edit
  gpd publish edit delete "$EDIT_ID" --package com.example.app
  
  # Create new edit and retry
  EDIT_ID=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.editId')
  gpd publish upload app.aab --package com.example.app --edit-id "$EDIT_ID" --no-auto-commit
fi
```

#### Pattern 2: Check Before Operations

```bash
# Verify edit exists and is valid
EDIT_INFO=$(gpd publish edit get "$EDIT_ID" --package com.example.app --output json)
if [ $? -ne 0 ] || [ "$(echo "$EDIT_INFO" | jq -r '.data.local.expired')" = "true" ]; then
  echo "Edit expired or not found, creating new one"
  EDIT_ID=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.editId')
fi
```

#### Pattern 3: Cleanup on Script Exit

```bash
# Trap errors and cleanup
cleanup() {
  if [ -n "$EDIT_ID" ]; then
    echo "Cleaning up edit $EDIT_ID"
    gpd publish edit delete "$EDIT_ID" --package com.example.app || true
  fi
}

trap cleanup EXIT ERR

# Your workflow here
EDIT_ID=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.editId')
# ... operations ...
gpd publish edit commit "$EDIT_ID" --package com.example.app

# Clear trap on success
trap - EXIT ERR
```

## Edit Lifecycle States

Edits progress through these states:

1. **draft** - Initial state, changes can be made
2. **validating** - Edit is being validated (temporary state)
3. **committed** - Edit has been committed (final state)
4. **aborted** - Edit was deleted (final state)

View state with:
```bash
gpd publish edit get "$EDIT_ID" --package com.example.app | jq '.data.local.state'
```

## Summary

The edit workflow provides powerful transactional capabilities for managing Google Play releases:

- ✅ **Atomic operations**: All changes commit together
- ✅ **Validation**: Test before committing
- ✅ **Multi-step workflows**: Build complex releases
- ✅ **Error recovery**: Delete and retry on failure
- ✅ **Team coordination**: Use descriptive edit IDs
- ✅ **CI/CD friendly**: Perfect for automation

Remember:
- Always validate before committing
- Use `--no-auto-commit` for multi-step workflows
- Clean up unused edits
- Handle expiration gracefully
- Use descriptive edit IDs for team coordination
