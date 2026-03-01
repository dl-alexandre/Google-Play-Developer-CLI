# Bulk Operations Implementation Plan

**Status**: Not started
**Priority**: Medium — CI/CD and batch workflow enablement
**Date**: 2026-03-01

## Background

Three bulk operation commands exist as stubs returning NoOp. Additionally, `bulk upload`'s `uploadFile()` helper is a stub. These commands enable batch management of store listings, images, and track configurations across multiple locales/apps.

Stubs in `internal/cli/kong_bulk.go`.

## Commands to Implement

### 1. `bulk listings` (line 319)

**Purpose**: Update store listings for multiple locales in one operation.

- `--from-dir` — directory with `{locale}/listing.json` files
- `--from-csv` — CSV with columns: locale, title, shortDescription, fullDescription, videoUrl
- `--locales` — filter to specific locales (comma-separated)
- Each locale file/row maps to `Edits.Listings.Update(locale, listing)`
- Requires single edit transaction wrapping all updates
- Report per-locale success/failure in output

### 2. `bulk images` (line 427)

**Purpose**: Upload images for multiple locales and image types in one operation.

- `--from-dir` — directory structure: `{locale}/{imageType}/{filename.png}`
- `--replace` — delete existing images before uploading (default: append)
- `--locales` and `--image-types` filters
- Each file maps to `Edits.Images.Upload(locale, imageType, file)`
- Validate file formats (PNG/JPEG) and dimensions before uploading
- Requires single edit transaction
- Use semaphore for concurrent uploads (respect `--max-parallel` from bulk cmd)
- Report per-file success/failure

### 3. `bulk tracks` (line 501)

**Purpose**: Update releases across multiple tracks in one operation.

- `--from-json` — JSON file defining track configurations
- `--tracks` — filter to specific tracks
- Each track entry maps to `Edits.Tracks.Update(track, trackConfig)`
- Requires single edit transaction
- Useful for promoting across multiple tracks simultaneously

### 4. Fix `bulk upload` `uploadFile()` (line ~197)

**Purpose**: The `bulk upload` command structure exists but `uploadFile()` returns placeholder data.

- Wire `uploadFile()` to actual `Edits.Apks.Upload()` or `Edits.Bundles.Upload()` based on file extension
- Respect `--in-progress-review-behaviour` flag (currently unused)
- Handle file type detection (`.apk` vs `.aab`)

## Shared Concerns

- All bulk commands should wrap operations in a single edit transaction
- Use `--dry-run` to preview what would happen without committing
- `--max-parallel` controls concurrency (already defined on parent)
- Output should include per-item results array showing success/failure/skipped
- Failed items should not abort the entire batch (continue on error by default, `--fail-fast` to stop)

## Implementation Order

1. Fix `bulk upload` `uploadFile()` — smallest change, biggest impact
2. `bulk listings` — straightforward CRUD batch
3. `bulk images` — file handling + concurrent uploads
4. `bulk tracks` — track management batch

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_bulk.go` | Implement `Run()` for 3 commands, fix `uploadFile()` |

## Testing

- Test directory structure parsing for listings and images
- Test CSV parsing for listings
- Test concurrent upload with semaphore limits
- Test partial failure handling (some locales succeed, others fail)
- Test dry-run mode
- Test edit transaction wrapping (single commit for all operations)
