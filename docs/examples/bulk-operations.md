# Bulk Operations Guide

This guide demonstrates how to perform batch operations with the Google Play Developer CLI, allowing you to efficiently manage multiple resources in a single command.

## Table of Contents

1. [Overview](#overview)
2. [Bulk Upload Operations](#bulk-upload-operations)
3. [Bulk Listing Updates](#bulk-listing-updates)
4. [Bulk Image Uploads](#bulk-image-uploads)
5. [Bulk Track Updates](#bulk-track-updates)
6. [Best Practices](#best-practices)

---

## Overview

Bulk operations allow you to perform multiple actions efficiently by processing them in parallel or as atomic transactions. This is especially useful when:

- Uploading multiple AAB/APK variants (e.g., different architectures)
- Localizing your app for multiple markets
- Managing store assets across different locales
- Updating multiple release tracks simultaneously

### Key Benefits

- **Efficiency**: Process multiple operations in parallel (up to 3 concurrent by default)
- **Atomicity**: All operations succeed or fail together within an edit transaction
- **Consistency**: Maintain consistency across related resources
- **Time Savings**: Significantly reduce total operation time

---

## Bulk Upload Operations

Upload multiple APK or AAB files at once, perfect for multi-architecture builds or variant releases.

### Basic Multi-File Upload

Upload multiple AAB files to the internal testing track:

```bash
gpd bulk upload \
  app-arm64-v8a.aab \
  app-armeabi-v7a.aab \
  app-x86_64.aab \
  --package com.example.app \
  --track internal
```

**Output:**
```json
{
  "data": {
    "successCount": 3,
    "failureCount": 0,
    "skippedCount": 0,
    "uploads": [
      {
        "file": "app-arm64-v8a.aab",
        "versionCode": 1048576,
        "status": "success",
        "sha1": "a1b2c3d4..."
      },
      {
        "file": "app-armeabi-v7a.aab",
        "versionCode": 1048577,
        "status": "success",
        "sha1": "e5f6g7h8..."
      },
      {
        "file": "app-x86_64.aab",
        "versionCode": 1048578,
        "status": "success",
        "sha1": "i9j0k1l2..."
      }
    ],
    "editId": "12345678901234567890",
    "committed": true,
    "processingTime": "12.34s"
  }
}
```

### Upload with Custom Edit ID

Use a specific edit ID for tracking:

```bash
gpd bulk upload \
  build/*.aab \
  --package com.example.app \
  --edit-id "multi-arch-release-v2.0" \
  --track beta \
  --no-auto-commit
```

### Parallel Upload Control

Adjust parallel upload concurrency:

```bash
gpd bulk upload \
  builds/*.aab \
  --package com.example.app \
  --max-parallel 5 \
  --track internal
```

### Dry Run Mode

Preview what would be uploaded without making changes:

```bash
gpd bulk upload \
  app.aab app2.aab \
  --package com.example.app \
  --track internal \
  --dry-run
```

**Output:**
```json
{
  "data": {
    "files": ["app.aab", "app2.aab"],
    "track": "internal",
    "dryRun": true,
    "wouldUpload": 2
  },
  "meta": {
    "noop": "dry run - no files uploaded"
  }
}
```

### CI/CD Integration Example

```bash
#!/bin/bash
# multi-arch-upload.sh

PACKAGE="com.example.app"
VERSION=$(grep versionCode app/build.gradle | grep -o '[0-9]\+')

# Build multiple architectures
./gradlew bundleRelease

# Upload all variants at once
gpd bulk upload \
  app/build/outputs/bundle/*/*.aab \
  --package "$PACKAGE" \
  --edit-id "release-v$VERSION" \
  --track internal \
  --max-parallel 4

echo "All variants uploaded successfully"
```

---

## Bulk Listing Updates

Update store listings across multiple locales from a single JSON file.

### Preparing the Listings Data File

Create a JSON file with locale mappings:

**`listings-update.json`:**
```json
{
  "en-US": {
    "title": "My Awesome App",
    "shortDescription": "The best productivity app for professionals",
    "fullDescription": "My Awesome App helps you stay organized and productive...",
    "video": "https://www.youtube.com/watch?v=example"
  },
  "de-DE": {
    "title": "Meine Tolle App",
    "shortDescription": "Die beste Produktivitäts-App für Profis",
    "fullDescription": "Meine Tolle App hilft Ihnen, organisiert und produktiv zu bleiben..."
  },
  "fr-FR": {
    "title": "Mon Application Géniale",
    "shortDescription": "La meilleure application de productivité",
    "fullDescription": "Mon Application Géniale vous aide à rester organisé et productif..."
  },
  "ja-JP": {
    "title": "素晴らしいアプリ",
    "shortDescription": "プロフェッショナル向けの最高の生産性アプリ",
    "fullDescription": "素晴らしいアプリは、整理されて生産的な状態を保つのに役立ちます..."
  },
  "es-ES": {
    "title": "Mi App Increíble",
    "shortDescription": "La mejor app de productividad para profesionales",
    "fullDescription": "Mi App Increíble te ayuda a mantenerte organizado y productivo..."
  }
}
```

### Apply Bulk Listing Update

```bash
gpd bulk listings \
  --package com.example.app \
  --data-file listings-update.json
```

**Output:**
```json
{
  "data": {
    "successCount": 5,
    "failureCount": 0,
    "locales": [
      {
        "locale": "en-US",
        "status": "success"
      },
      {
        "locale": "de-DE",
        "status": "success"
      },
      {
        "locale": "fr-FR",
        "status": "success"
      },
      {
        "locale": "ja-JP",
        "status": "success"
      },
      {
        "locale": "es-ES",
        "status": "success"
      }
    ],
    "editId": "12345678901234567890"
  }
}
```

### Using with Edit ID

```bash
# Create an edit first
EDIT_ID=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.editId')

# Update listings
gpd bulk listings \
  --package com.example.app \
  --data-file listings-update.json \
  --edit-id "$EDIT_ID"

# Later, commit with other changes
gpd publish edit commit "$EDIT_ID" --package com.example.app
```

### Dry Run Validation

Validate your JSON structure before applying:

```bash
gpd bulk listings \
  --package com.example.app \
  --data-file listings-update.json \
  --dry-run
```

### Localization Workflow Script

```bash
#!/bin/bash
# localize-release.sh

PACKAGE="com.example.app"

# Validate localization file
if ! jq empty listings-update.json 2>/dev/null; then
  echo "Error: Invalid JSON in listings-update.json"
  exit 1
fi

# Get locale count
LOCALE_COUNT=$(jq 'length' listings-update.json)
echo "Updating listings for $LOCALE_COUNT locales..."

# Create edit for atomic update
EDIT_ID=$(gpd publish edit create --package "$PACKAGE" --output json | jq -r '.data.editId')
echo "Created edit: $EDIT_ID"

# Apply listing updates
gpd bulk listings \
  --package "$PACKAGE" \
  --data-file listings-update.json \
  --edit-id "$EDIT_ID"

# Validate before committing
gpd publish edit validate "$EDIT_ID" --package "$PACKAGE"

# Commit all changes
gpd publish edit commit "$EDIT_ID" --package "$PACKAGE"

echo "Listings updated successfully!"
```

---

## Bulk Image Uploads

Upload multiple store listing images organized by type and locale.

### Directory Structure

Organize your images following this structure:

```
store-images/
├── featureGraphic/
│   ├── en-US.png
│   ├── de-DE.png
│   └── ja-JP.png
├── icon/
│   └── universal.png
├── phoneScreenshots/
│   ├── en-US/
│   │   ├── 1_login.png
│   │   ├── 2_dashboard.png
│   │   └── 3_settings.png
│   ├── de-DE/
│   │   ├── 1_login.png
│   │   ├── 2_dashboard.png
│   │   └── 3_settings.png
│   └── ja-JP/
│       ├── 1_login.png
│       ├── 2_dashboard.png
│       └── 3_settings.png
├── tabletScreenshots/
│   └── en-US/
│       ├── 1_tablet_home.png
│       └── 2_tablet_detail.png
└── tvBanner/
    └── en-US.png
```

### Bulk Image Upload

```bash
gpd bulk images \
  --package com.example.app \
  --image-dir ./store-images \
  --max-parallel 3
```

**Output:**
```json
{
  "data": {
    "successCount": 15,
    "failureCount": 0,
    "images": [
      {
        "type": "featureGraphic",
        "locale": "en-US",
        "filename": "store-images/featureGraphic/en-US.png",
        "status": "success"
      },
      {
        "type": "phoneScreenshots",
        "locale": "en-US",
        "filename": "store-images/phoneScreenshots/en-US/1_login.png",
        "status": "success"
      },
      {
        "type": "phoneScreenshots",
        "locale": "de-DE",
        "filename": "store-images/phoneScreenshots/de-DE/1_login.png",
        "status": "success"
      }
    ],
    "editId": "12345678901234567890"
  }
}
```

### Override Default Locale

If your directory structure doesn't include locale subdirectories, specify a default:

```bash
gpd bulk images \
  --package com.example.app \
  --image-dir ./universal-images \
  --locale en-US
```

### Preview with Dry Run

```bash
gpd bulk images \
  --package com.example.app \
  --image-dir ./store-images \
  --dry-run
```

**Output:**
```json
{
  "data": {
    "images": [
      {
        "type": "featureGraphic",
        "locale": "en-US",
        "filename": "store-images/featureGraphic/en-US.png",
        "status": "pending"
      },
      {
        "type": "phoneScreenshots",
        "locale": "en-US",
        "filename": "store-images/phoneScreenshots/en-US/1_login.png",
        "status": "pending"
      }
    ],
    "count": 15,
    "dryRun": true,
    "wouldUpload": 15
  },
  "meta": {
    "noop": "dry run - no images uploaded"
  }
}
```

### Asset Management Script

```bash
#!/bin/bash
# upload-store-assets.sh

PACKAGE="com.example.app"
IMAGE_DIR="./store-assets"

# Verify directory exists
if [ ! -d "$IMAGE_DIR" ]; then
  echo "Error: Image directory $IMAGE_DIR not found"
  exit 1
fi

# Count images
IMAGE_COUNT=$(find "$IMAGE_DIR" -name "*.png" -o -name "*.jpg" | wc -l)
echo "Found $IMAGE_COUNT images to upload"

# Create edit for atomic update
EDIT_ID=$(gpd publish edit create --package "$PACKAGE" --output json | jq -r '.data.editId')

# Upload all images
gpd bulk images \
  --package "$PACKAGE" \
  --image-dir "$IMAGE_DIR" \
  --edit-id "$EDIT_ID" \
  --max-parallel 4

# Validate and commit
gpd publish edit validate "$EDIT_ID" --package "$PACKAGE"
gpd publish edit commit "$EDIT_ID" --package "$PACKAGE"

echo "Store assets uploaded successfully!"
```

---

## Bulk Track Updates

Update multiple release tracks simultaneously with the same version codes.

### Update Multiple Tracks

Deploy to internal and beta tracks at once:

```bash
gpd bulk tracks \
  --package com.example.app \
  --tracks internal beta \
  --version-codes 1048576 1048577 \
  --status completed
```

**Output:**
```json
{
  "data": {
    "successCount": 2,
    "failureCount": 0,
    "tracks": [
      {
        "track": "internal",
        "status": "success",
        "versionCodes": ["1048576", "1048577"]
      },
      {
        "track": "beta",
        "status": "success",
        "versionCodes": ["1048576", "1048577"]
      }
    ],
    "editId": "12345678901234567890",
    "committed": true
  }
}
```

### Multi-Track Release with Custom Name

```bash
gpd bulk tracks \
  --package com.example.app \
  --tracks internal alpha beta \
  --version-codes 1048576 \
  --status draft \
  --name "Version 2.0 Release Candidate"
```

### Staged Multi-Track Update

Create releases without committing for staged deployment:

```bash
gpd bulk tracks \
  --package com.example.app \
  --tracks internal beta \
  --version-codes 1048576 \
  --status draft \
  --no-auto-commit
```

Then commit separately:

```bash
gpd publish edit commit <edit-id> --package com.example.app
```

### Dry Run Multi-Track Planning

```bash
gpd bulk tracks \
  --package com.example.app \
  --tracks internal alpha beta production \
  --version-codes 1048576 1048577 \
  --status inProgress \
  --dry-run
```

**Output:**
```json
{
  "data": {
    "tracks": ["internal", "alpha", "beta", "production"],
    "versionCodes": ["1048576", "1048577"],
    "status": "inProgress",
    "dryRun": true,
    "wouldUpdate": 4
  },
  "meta": {
    "noop": "dry run - no tracks updated"
  }
}
```

### Complete Multi-Track Workflow

```bash
#!/bin/bash
# multi-track-release.sh

PACKAGE="com.example.app"
VERSION_CODE=$(grep versionCode app/build.gradle | grep -o '[0-9]\+')

# 1. Upload AAB
gpd publish upload app/build/outputs/bundle/release/app-release.aab \
  --package "$PACKAGE" \
  --edit-id "multi-track-$VERSION_CODE" \
  --no-auto-commit

# 2. Create releases on internal and beta tracks
gpd bulk tracks \
  --package "$PACKAGE" \
  --tracks internal beta \
  --version-codes "$VERSION_CODE" \
  --status completed \
  --edit-id "multi-track-$VERSION_CODE" \
  --no-auto-commit

# 3. Validate
gpd publish edit validate "multi-track-$VERSION_CODE" --package "$PACKAGE"

# 4. Commit all changes
gpd publish edit commit "multi-track-$VERSION_CODE" --package "$PACKAGE"

echo "Released version $VERSION_CODE to internal and beta tracks"
```

---

## Best Practices

### 1. Use Edit Transactions for Atomicity

Always use `--edit-id` for multi-step bulk operations to ensure atomic commits:

```bash
EDIT_ID=$(gpd publish edit create --package com.example.app --output json | jq -r '.data.editId')

# All bulk operations use the same edit ID
gpd bulk upload ... --edit-id "$EDIT_ID" --no-auto-commit
gpd bulk listings ... --edit-id "$EDIT_ID" --no-auto-commit
gpd bulk images ... --edit-id "$EDIT_ID" --no-auto-commit

# Commit everything at once
gpd publish edit commit "$EDIT_ID" --package com.example.app
```

### 2. Validate Before Committing

Always validate bulk operations before committing:

```bash
gpd bulk upload ... --dry-run
gpd bulk listings ... --dry-run
gpd bulk images ... --dry-run

# Then commit if everything looks correct
gpd bulk upload ... --edit-id "$EDIT_ID"
```

### 3. Monitor Parallel Uploads

Adjust `--max-parallel` based on your file sizes and network conditions:

```bash
# For small files (APKs < 10MB)
gpd bulk upload *.apk --max-parallel 5

# For large files (AABs > 50MB)
gpd bulk upload *.aab --max-parallel 2
```

### 4. Error Handling in Scripts

```bash
#!/bin/bash
set -e

PACKAGE="com.example.app"
EDIT_ID=$(gpd publish edit create --package "$PACKAGE" --output json | jq -r '.data.editId')

cleanup() {
  if [ $? -ne 0 ]; then
    echo "Error occurred, cleaning up edit $EDIT_ID"
    gpd publish edit delete "$EDIT_ID" --package "$PACKAGE" || true
  fi
}
trap cleanup EXIT

gpd bulk upload builds/*.aab \
  --package "$PACKAGE" \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

gpd bulk listings \
  --package "$PACKAGE" \
  --data-file listings.json \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

gpd publish edit validate "$EDIT_ID" --package "$PACKAGE"
gpd publish edit commit "$EDIT_ID" --package "$PACKAGE"
```

### 5. Progress Tracking

For large bulk operations, monitor progress with verbose output:

```bash
gpd bulk upload builds/*.aab \
  --package com.example.app \
  --verbose \
  2>&1 | tee bulk-upload.log
```

### 6. Batch Size Management

When dealing with hundreds of images, consider splitting into batches:

```bash
# Process images in batches of 50
for batch in store-images/batch-{1..10}; do
  if [ -d "$batch" ]; then
    gpd bulk images \
      --package com.example.app \
      --image-dir "$batch" \
      --max-parallel 3
  fi
done
```

---

## Summary

| Operation | Command | Example Count |
|-----------|---------|---------------|
| Upload multiple AABs | `gpd bulk upload` | 6 examples |
| Update listings | `gpd bulk listings` | 5 examples |
| Upload images | `gpd bulk images` | 5 examples |
| Update tracks | `gpd bulk tracks` | 5 examples |
| **Total Examples** | | **21 examples** |

---

## Related Commands

- [`gpd publish edit create`](./edit-workflow.md) - Create edit transactions
- [`gpd publish upload`](./edit-workflow.md) - Single file upload
- [`gpd publish release`](./release-workflow.md) - Create releases
- [`gpd config doctor`](./error-debugging.md) - Diagnose configuration issues
