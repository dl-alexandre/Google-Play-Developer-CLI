# Publish Metadata (Listings, Details, Images, Assets) Implementation Plan

**Status**: Not started
**Priority**: High — core store presence management
**Date**: 2026-03-01

## Background

Store listing and app detail management commands are all stubs. Users can't manage screenshots, descriptions, or app metadata from the CLI. These map directly to the Android Publisher API `Edits.Listings`, `Edits.Details`, and `Edits.Images` endpoints.

All stubs in `internal/cli/kong_publish.go`.

## Commands to Implement

### Listings (3 commands)

| Command | Line | API Method |
|---------|------|------------|
| `publish listing get` | 870 | `Edits.Listings.Get(language)` |
| `publish listing update` | 860 | `Edits.Listings.Update(language, listing)` |
| `publish listing delete` | 885 | `Edits.Listings.Delete(language)` |

- `--language` flag (BCP-47 locale code, e.g., `en-US`)
- `listing get` with no language → `Edits.Listings.List()` (all locales)
- `listing update` accepts `--title`, `--short-description`, `--full-description`, `--video-url`
- All require edit transaction

### Details (3 commands)

| Command | Line | API Method |
|---------|------|------------|
| `publish details get` | 900 | `Edits.Details.Get()` |
| `publish details update` | 916 | `Edits.Details.Update(details)` |
| `publish details patch` | 933 | `Edits.Details.Patch(details)` |

- Fields: `contactEmail`, `contactPhone`, `contactWebsite`, `defaultLanguage`
- `update` replaces all fields, `patch` merges partial updates
- All require edit transaction

### Images (4 commands)

| Command | Line | API Method |
|---------|------|------------|
| `publish images list` | 969 | `Edits.Images.List(language, imageType)` |
| `publish images upload` | 957 | `Edits.Images.Upload(language, imageType, file)` |
| `publish images delete` | 984 | `Edits.Images.Delete(language, imageType, imageId)` |
| `publish images deleteall` | 998 | `Edits.Images.Deleteall(language, imageType)` |

- `--image-type`: `featureGraphic`, `icon`, `phoneScreenshots`, `sevenInchScreenshots`, `tenInchScreenshots`, `tvBanner`, `tvScreenshots`, `wearScreenshots`
- `--language` flag for locale
- `upload` reads file from `--file` path, validates format (PNG/JPEG) and dimensions
- All require edit transaction

### Assets (2 commands)

| Command | Line | Notes |
|---------|------|-------|
| `publish assets upload` | 1019 | Expansion files (OBB) upload |
| `publish assets spec` | 1027 | Show asset requirements/specs |

- `assets upload` handles `Edits.Expansionfiles.Upload()`
- `assets spec` is informational (no API call, just outputs requirements)

### Deobfuscation (1 command)

| Command | Line | API Method |
|---------|------|------------|
| `publish deobfuscation upload` | 1048 | `Edits.Deobfuscationfiles.Upload(versionCode, file)` |

- Upload ProGuard/R8 mapping files for crash symbolication
- `--version-code` and `--file` flags
- Requires edit transaction

### Internal Sharing (1 command)

| Command | Line | API Method |
|---------|------|------------|
| `publish internal-share upload` | 1275 | `Internalappsharingartifacts.Uploadapk/Uploadbundle` |

- Upload APK/AAB for internal sharing (bypasses tracks)
- Returns download URL
- Does NOT require edit transaction (separate API)

## Implementation Order

1. `listing get` + `details get` (read-only, simplest)
2. `listing update` + `details update/patch` (write operations)
3. `images list` + `images upload` (file handling)
4. `images delete` + `images deleteall`
5. `listing delete`
6. `deobfuscation upload`
7. `assets upload/spec`
8. `internal-share upload`

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_publish.go` | Implement `Run()` for all 14 commands |
| `internal/api/client.go` | Possible helpers for image upload with content type detection |

## Testing

- Mock all Edits.Listings/Details/Images API responses
- Test image type validation and file format checks
- Test listing CRUD cycle: create → get → update → delete
- Test internal sharing (separate from edit flow)
- Test deobfuscation upload with version code validation
