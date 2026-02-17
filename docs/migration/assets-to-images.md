# Migration Guide: Assets to Images API

## Overview

The Google Play Developer CLI has migrated from the legacy `assets` command to a new `images` command that provides more granular control over store images. The new `images` API offers:

- **Better control**: Upload, list, and delete individual images
- **Type safety**: Explicit image type specification
- **Improved validation**: Built-in dimension and format validation
- **Edit workflow integration**: Full support for edit transactions
- **Locale support**: Per-locale image management

This guide will help you migrate your existing workflows from `gpd publish assets` to `gpd publish images`.

## What Changed

### Old Approach: `gpd publish assets upload`

The legacy `assets` command used a directory-based convention where images were organized in a specific folder structure:

```
assets/
  {locale}/
    {category}/
      image1.png
      image2.png
```

This approach:
- Required a specific directory structure
- Uploaded all images in a category at once
- Had limited control over individual images
- Used category-based organization (phone, tablet, tv, wear)

### New Approach: `gpd publish images upload`

The new `images` command uses explicit image types and file paths:

```bash
gpd publish images upload <type> <file> --locale <locale>
```

This approach:
- Allows uploading individual images
- Uses explicit image type specification
- Provides better error messages and validation
- Supports listing and deleting specific images
- Integrates with the edit workflow

## Command Mapping

| Old Command | New Command |
|-------------|-------------|
| `gpd publish assets upload [directory]` | `gpd publish images upload <type> <file> --locale <locale>` |
| `gpd publish assets spec` | See [Image Types Supported](#image-types-supported) section below |
| N/A (not available) | `gpd publish images list <type> --locale <locale>` |
| N/A (not available) | `gpd publish images delete <type> <id> --locale <locale>` |
| N/A (not available) | `gpd publish images deleteall <type> --locale <locale>` |

### Key Differences

1. **Explicit type required**: You must specify the image type (e.g., `phoneScreenshots`, `icon`) as a positional argument
2. **Single file upload**: Upload one image at a time instead of batch directory uploads
3. **Locale flag**: Use `--locale` flag instead of directory structure
4. **Edit workflow**: Full support for `--edit-id` and `--no-auto-commit` flags

## Image Types Supported

The following image types are supported in the new `images` command:

| Image Type | Description | Use Case |
|------------|-------------|----------|
| `phoneScreenshots` | Phone screenshots | Screenshots for phones (320-3840px) |
| `tabletScreenshots` | Tablet screenshots | Screenshots for tablets |
| `sevenInchScreenshots` | 7-inch tablet screenshots | Screenshots for 7-inch tablets |
| `tenInchScreenshots` | 10-inch tablet screenshots | Screenshots for 10-inch tablets |
| `tvScreenshots` | TV screenshots | Screenshots for Android TV |
| `wearScreenshots` | Wear OS screenshots | Screenshots for Wear OS devices |
| `icon` | App icon | 512x512 PNG icon |
| `featureGraphic` | Feature graphic | 1024x500 promotional graphic |
| `tvBanner` | TV banner | 1280x720 banner for Android TV |

**Note**: The legacy `assets` command used categories like `phone`, `tablet`, `tv`, `wear`. The new API uses more specific types like `phoneScreenshots`, `sevenInchScreenshots`, `tenInchScreenshots`, etc.

## Validation Changes

The new `images` command includes enhanced validation that happens before upload:

### Dimension Validation

Each image type has specific dimension requirements:

| Image Type | Min Width | Max Width | Min Height | Max Height |
|------------|-----------|-----------|------------|------------|
| `icon` | 512 | 512 | 512 | 512 |
| `featureGraphic` | 1024 | 1024 | 500 | 500 |
| `tvBanner` | 1280 | 1280 | 720 | 720 |
| `phoneScreenshots` | 320 | 3840 | 320 | 3840 |
| `tabletScreenshots` | 320 | 3840 | 320 | 3840 |
| `sevenInchScreenshots` | 320 | 3840 | 320 | 3840 |
| `tenInchScreenshots` | 320 | 3840 | 320 | 3840 |
| `tvScreenshots` | 320 | 3840 | 320 | 3840 |
| `wearScreenshots` | 320 | 3840 | 320 | 3840 |

### File Size Limits

| Image Type | Max Size |
|------------|----------|
| `icon` | 1 MB |
| `featureGraphic` | 15 MB |
| `tvBanner` | 15 MB |
| All screenshot types | 8 MB |

### Format Requirements

- **PNG**: Supported for all image types
- **JPEG**: Supported for all image types except `icon` (PNG only)

The CLI will validate these requirements before uploading and provide clear error messages if validation fails.

## Migration Steps

### Step 1: Identify Your Current Asset Structure

First, identify how your assets are currently organized:

```bash
# Example old structure
assets/
  en-US/
    phone/
      screenshot1.png
      screenshot2.png
    tablet/
      screenshot1.png
    icon.png
    featureGraphic.png
```

### Step 2: Map Categories to Image Types

Map your old category structure to the new image types:

| Old Category | New Image Type |
|--------------|----------------|
| `phone/` | `phoneScreenshots` |
| `tablet/` | `tabletScreenshots`, `sevenInchScreenshots`, or `tenInchScreenshots` |
| `tv/` | `tvScreenshots` |
| `wear/` | `wearScreenshots` |
| `icon.png` | `icon` |
| `featureGraphic.png` | `featureGraphic` |

### Step 3: Update Your Scripts

#### Before (Old Assets Command)

```bash
# Upload all assets from directory
gpd publish assets upload ./assets --package com.example.app

# Or with specific category
gpd publish assets upload ./assets --package com.example.app --category phone
```

#### After (New Images Command)

```bash
# Upload individual images
gpd publish images upload phoneScreenshots ./screenshots/phone1.png \
  --package com.example.app --locale en-US

gpd publish images upload phoneScreenshots ./screenshots/phone2.png \
  --package com.example.app --locale en-US

gpd publish images upload icon ./icon.png \
  --package com.example.app --locale en-US

gpd publish images upload featureGraphic ./feature-graphic.png \
  --package com.example.app --locale en-US
```

### Step 4: Batch Upload Script Example

Since the new API requires individual file uploads, you'll need to create a script for batch operations:

```bash
#!/bin/bash
# migrate-assets.sh

PACKAGE="com.example.app"
LOCALE="en-US"
ASSETS_DIR="./assets"

# Upload phone screenshots
for file in "$ASSETS_DIR/$LOCALE/phone"/*.png; do
  if [ -f "$file" ]; then
    echo "Uploading $file as phoneScreenshots..."
    gpd publish images upload phoneScreenshots "$file" \
      --package "$PACKAGE" --locale "$LOCALE"
  fi
done

# Upload tablet screenshots (7-inch)
for file in "$ASSETS_DIR/$LOCALE/tablet"/*.png; do
  if [ -f "$file" ]; then
    echo "Uploading $file as sevenInchScreenshots..."
    gpd publish images upload sevenInchScreenshots "$file" \
      --package "$PACKAGE" --locale "$LOCALE"
  fi
done

# Upload icon
if [ -f "$ASSETS_DIR/$LOCALE/icon.png" ]; then
  echo "Uploading icon..."
  gpd publish images upload icon "$ASSETS_DIR/$LOCALE/icon.png" \
    --package "$PACKAGE" --locale "$LOCALE"
fi

# Upload feature graphic
if [ -f "$ASSETS_DIR/$LOCALE/featureGraphic.png" ]; then
  echo "Uploading feature graphic..."
  gpd publish images upload featureGraphic "$ASSETS_DIR/$LOCALE/featureGraphic.png" \
    --package "$PACKAGE" --locale "$LOCALE"
fi
```

### Step 5: Using Edit Workflow (Optional)

The new `images` command fully supports the edit workflow for batching multiple changes:

```bash
# Create an edit
EDIT_ID=$(gpd publish edit create --package com.example.app | jq -r '.data.id')

# Upload multiple images in the same edit
gpd publish images upload phoneScreenshots screenshot1.png \
  --package com.example.app --locale en-US --edit-id "$EDIT_ID" --no-auto-commit

gpd publish images upload phoneScreenshots screenshot2.png \
  --package com.example.app --locale en-US --edit-id "$EDIT_ID" --no-auto-commit

gpd publish images upload icon icon.png \
  --package com.example.app --locale en-US --edit-id "$EDIT_ID" --no-auto-commit

# Commit all changes at once
gpd publish edit commit "$EDIT_ID" --package com.example.app
```

### Step 6: List and Manage Images

The new API provides commands to list and manage existing images:

```bash
# List all phone screenshots
gpd publish images list phoneScreenshots \
  --package com.example.app --locale en-US

# Delete a specific image
gpd publish images delete phoneScreenshots <image-id> \
  --package com.example.app --locale en-US

# Delete all images of a type
gpd publish images deleteall phoneScreenshots \
  --package com.example.app --locale en-US
```

## Example: Complete Migration

### Before

```bash
# Old directory structure
assets/
  en-US/
    phone/
      screenshot1.png
      screenshot2.png
      screenshot3.png
    icon.png
    featureGraphic.png

# Old upload command
gpd publish assets upload ./assets --package com.example.app
```

### After

```bash
# New individual uploads
gpd publish images upload phoneScreenshots assets/en-US/phone/screenshot1.png \
  --package com.example.app --locale en-US

gpd publish images upload phoneScreenshots assets/en-US/phone/screenshot2.png \
  --package com.example.app --locale en-US

gpd publish images upload phoneScreenshots assets/en-US/phone/screenshot3.png \
  --package com.example.app --locale en-US

gpd publish images upload icon assets/en-US/icon.png \
  --package com.example.app --locale en-US

gpd publish images upload featureGraphic assets/en-US/featureGraphic.png \
  --package com.example.app --locale en-US
```

## Deprecation Timeline

### Current Status

- **Legacy `assets` command**: Still available but deprecated
- **New `images` command**: Fully supported and recommended for all new workflows
- **Migration period**: Both commands are available during the transition period

### When Old Commands Will Be Removed

The `gpd publish assets` commands will be removed in a future major version release. We recommend migrating to the new `images` API as soon as possible.

**Timeline**:
- **v1.x**: Both `assets` and `images` commands available
- **v2.0**: `assets` commands will be removed (target: Q2 2026)

### Migration Checklist

- [ ] Identify all scripts using `gpd publish assets`
- [ ] Map old category structure to new image types
- [ ] Update upload scripts to use `gpd publish images upload`
- [ ] Test image uploads with the new command
- [ ] Verify image dimensions meet new validation requirements
- [ ] Update CI/CD pipelines if applicable
- [ ] Remove old `assets` directory structure (optional)

## Troubleshooting

### Common Issues

1. **"Invalid image type"**
   - Ensure you're using the correct image type name (e.g., `phoneScreenshots`, not `phone`)
   - Check the [Image Types Supported](#image-types-supported) section for valid types

2. **"Image width/height too small/large"**
   - Verify your image dimensions match the requirements in the [Validation Changes](#validation-changes) section
   - Use an image editor to resize if needed

3. **"Image exceeds size limit"**
   - Compress your images to meet the size requirements
   - Consider using PNG optimization tools or JPEG compression

4. **"Invalid image format"**
   - Ensure images are PNG or JPEG
   - For icons, only PNG is supported

5. **"Locale not found"**
   - Use the `--locale` flag with a valid locale code (e.g., `en-US`, `fr-FR`)
   - Default locale is `en-US` if not specified

## Additional Resources

- [Google Play Console Help: Store Listing Images](https://support.google.com/googleplay/android-developer/answer/9866151)
- [Google Play Developer API: Images](https://developers.google.com/android-publisher/api-ref/rest/v3/edits.images)
- CLI Help: `gpd publish images --help`

## Need Help?

If you encounter issues during migration:

1. Check the validation errors - they provide specific guidance
2. Use `--dry-run` flag to test commands without making changes
3. Review the [Troubleshooting](#troubleshooting) section
4. Open an issue on the GitHub repository
