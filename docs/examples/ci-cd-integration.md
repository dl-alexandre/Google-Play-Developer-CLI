# CI/CD Integration Guide

This guide demonstrates how to integrate the Google Play Developer CLI (`gpd`) into your CI/CD pipelines for automated Android app deployments.

## Table of Contents

1. [Overview](#overview)
2. [Authentication Setup](#authentication-setup)
3. [GitHub Actions Examples](#github-actions-examples)
4. [GitLab CI Examples](#gitlab-ci-examples)
5. [Common Patterns](#common-patterns)
6. [Best Practices](#best-practices)
7. [Troubleshooting](#troubleshooting)

## Overview

Automating Play Store deployments with `gpd` provides several benefits:

- **Consistency**: Eliminates manual errors and ensures reproducible releases
- **Speed**: Reduces deployment time from minutes to seconds
- **Traceability**: Every release is tied to a git commit and CI/CD run
- **Safety**: Enables staged rollouts, automated testing, and rollback capabilities
- **Compliance**: Maintains audit trails for regulatory requirements

The `gpd` CLI is designed for CI/CD environments with:
- JSON-first output for easy parsing
- Predictable exit codes for error handling
- No interactive prompts (all flags are explicit)
- Support for service account authentication
- Edit transaction management for atomic releases

## Authentication Setup

### Service Account Creation

1. **Create a Google Cloud Project** (if you don't have one):
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one

2. **Enable the Google Play Android Publisher API**:
   - Navigate to "APIs & Services" > "Library"
   - Search for "Google Play Android Publisher API"
   - Click "Enable"

3. **Create a Service Account**:
   - Go to "APIs & Services" > "Credentials"
   - Click "Create Credentials" > "Service Account"
   - Provide a name (e.g., "play-store-deployer")
   - Click "Create and Continue"
   - Skip role assignment (not needed for Play Store API)
   - Click "Done"

4. **Create and Download a Key**:
   - Click on the created service account
   - Go to the "Keys" tab
   - Click "Add Key" > "Create new key"
   - Select "JSON" format
   - Download the JSON key file

5. **Grant Play Console Access**:
   - Go to [Google Play Console](https://play.google.com/console/)
   - Navigate to "Users and permissions"
   - Click "Invite new users"
   - Enter the service account email (from the JSON key file)
   - Grant appropriate permissions:
     - **Release apps**: For uploading and releasing
     - **View app information**: For reading app details
     - **Manage production releases**: For production deployments
     - **Manage testing track releases**: For internal/alpha/beta tracks
   - Save the invitation

### Storing Credentials Securely

#### GitHub Actions

Use GitHub Secrets to store your service account key:

1. **Base64 Encode the JSON Key** (optional, for multiline storage):
   ```bash
   base64 -i service-account.json
   ```

2. **Add as Repository Secret**:
   - Go to your repository Settings > Secrets and variables > Actions
   - Click "New repository secret"
   - Name: `GPD_SERVICE_ACCOUNT_KEY`
   - Value: Paste the entire JSON content (or base64-encoded value)
   - Click "Add secret"

#### GitLab CI

Use GitLab CI/CD Variables:

1. **Add CI/CD Variable**:
   - Go to your project Settings > CI/CD > Variables
   - Click "Add variable"
   - Key: `GPD_SERVICE_ACCOUNT_KEY`
   - Value: Paste the entire JSON content
   - Type: Variable
   - Environment scope: All (or specific environments)
   - Protect variable: ✅ (recommended)
   - Mask variable: ❌ (JSON is too large to mask)
   - Click "Add variable"

#### Alternative: Key File Storage

For environments where you can securely store files:

1. **Store in CI/CD Secure Storage**:
   - Upload the JSON key to your CI/CD platform's secure file storage
   - Reference it in your pipeline configuration

2. **Use Environment Variable**:
   ```yaml
   env:
     GOOGLE_APPLICATION_CREDENTIALS: /path/to/service-account.json
   ```

### Environment Variables

`gpd` supports multiple authentication methods (in priority order):

1. **`--key` flag**: Explicit key file path
   ```bash
   gpd --key /path/to/service-account.json auth status
   ```

2. **`GPD_SERVICE_ACCOUNT_KEY`**: JSON key content as environment variable
   ```bash
   export GPD_SERVICE_ACCOUNT_KEY='{"type": "service_account", ...}'
   gpd auth status
   ```

3. **`GOOGLE_APPLICATION_CREDENTIALS`**: Path to key file
   ```bash
   export GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json
   gpd auth status
   ```

4. **Application Default Credentials**: For GCP environments

## GitHub Actions Examples

### Basic Upload Workflow

This workflow uploads an AAB file to the internal testing track:

```yaml
name: Deploy to Play Store

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      track:
        description: 'Release track'
        required: true
        default: 'internal'
        type: choice
        options:
          - internal
          - alpha
          - beta
          - production

jobs:
  deploy:
    name: Deploy to Play Store
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22.x'

      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Verify authentication
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
        run: |
          gpd auth status
          gpd auth check --package ${{ secrets.APP_PACKAGE_NAME }}

      - name: Build AAB
        run: |
          # Your build command here
          ./gradlew bundleRelease

      - name: Upload to Play Store
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
        run: |
          gpd publish upload \
            app/build/outputs/bundle/release/app-release.aab \
            --package ${{ secrets.APP_PACKAGE_NAME }}

      - name: Create Release
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
        run: |
          TRACK="${{ github.event.inputs.track || 'internal' }}"
          gpd publish release \
            --package ${{ secrets.APP_PACKAGE_NAME }} \
            --track "$TRACK" \
            --status draft
```

### Release to Internal Testing

Automated release to internal testing on every push to main:

```yaml
name: Internal Testing Release

on:
  push:
    branches: [main]
    paths:
      - 'app/**'

jobs:
  deploy-internal:
    name: Deploy to Internal Testing
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Java
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'

      - name: Cache Gradle dependencies
        uses: actions/cache@v4
        with:
          path: |
            ~/.gradle/caches
            ~/.gradle/wrapper
          key: ${{ runner.os }}-gradle-${{ hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') }}
          restore-keys: |
            ${{ runner.os }}-gradle-

      - name: Build AAB
        run: ./gradlew bundleRelease

      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Upload and Release
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
          APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
        run: |
          # Upload AAB
          gpd publish upload \
            app/build/outputs/bundle/release/app-release.aab \
            --package "$APP_PACKAGE"

          # Get version code from AAB (using aapt2 or parsing build.gradle)
          VERSION_CODE=$(grep -oP 'versionCode\s+\K\d+' app/build.gradle || echo "auto")
          
          # Create release
          gpd publish release \
            --package "$APP_PACKAGE" \
            --track internal \
            --status completed \
            --version-code "$VERSION_CODE"

      - name: Notify on Success
        if: success()
        uses: 8398a7/action-slack@v3
        with:
          status: custom
          custom_payload: |
            {
              text: "✅ Successfully deployed to Internal Testing",
              attachments: [{
                color: 'good',
                text: `Version: ${process.env.VERSION_CODE}\nCommit: ${process.env.GITHUB_SHA}`
              }]
            }
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### Staged Rollout Workflow

Gradual rollout to production with automatic promotion:

```yaml
name: Staged Production Rollout

on:
  workflow_dispatch:
    inputs:
      initial_percentage:
        description: 'Initial rollout percentage'
        required: true
        default: '10'
        type: choice
        options:
          - '10'
          - '20'
          - '50'

jobs:
  staged-rollout:
    name: Staged Production Rollout
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Java
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'

      - name: Build AAB
        run: ./gradlew bundleRelease

      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Promote from Beta to Production
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
          APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
        run: |
          # Promote latest beta release to production
          gpd publish promote \
            --package "$APP_PACKAGE" \
            --from-track beta \
            --to-track production

      - name: Start Staged Rollout
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
          APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
          INITIAL_PERCENTAGE: ${{ github.event.inputs.initial_percentage }}
        run: |
          gpd publish rollout \
            --package "$APP_PACKAGE" \
            --track production \
            --percentage "$INITIAL_PERCENTAGE"

      - name: Wait for Rollout Period
        run: |
          echo "Waiting 24 hours before next stage..."
          # In a real scenario, you might use a scheduled workflow
          # or manual approval for the next stage

      - name: Increase Rollout (Manual Trigger)
        if: false  # Set to true when ready to increase
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
          APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
        run: |
          gpd publish rollout \
            --package "$APP_PACKAGE" \
            --track production \
            --percentage 50
```

### Multi-Environment Deployment

Deploy to different tracks based on git branch:

```yaml
name: Multi-Environment Deployment

on:
  push:
    branches:
      - develop
      - staging
      - main

jobs:
  determine-environment:
    name: Determine Environment
    runs-on: ubuntu-latest
    outputs:
      track: ${{ steps.env.outputs.track }}
      status: ${{ steps.env.outputs.status }}
    steps:
      - name: Set environment
        id: env
        run: |
          if [[ "${{ github.ref }}" == "refs/heads/develop" ]]; then
            echo "track=internal" >> $GITHUB_OUTPUT
            echo "status=draft" >> $GITHUB_OUTPUT
          elif [[ "${{ github.ref }}" == "refs/heads/staging" ]]; then
            echo "track=beta" >> $GITHUB_OUTPUT
            echo "status=completed" >> $GITHUB_OUTPUT
          elif [[ "${{ github.ref }}" == "refs/heads/main" ]]; then
            echo "track=production" >> $GITHUB_OUTPUT
            echo "status=inProgress" >> $GITHUB_OUTPUT
          fi

  deploy:
    name: Deploy to ${{ needs.determine-environment.outputs.track }}
    needs: determine-environment
    runs-on: ubuntu-latest
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Java
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'

      - name: Build AAB
        run: ./gradlew bundleRelease

      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH

      - name: Deploy
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
          APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
          TRACK: ${{ needs.determine-environment.outputs.track }}
          STATUS: ${{ needs.determine-environment.outputs.status }}
        run: |
          # Upload AAB
          gpd publish upload \
            app/build/outputs/bundle/release/app-release.aab \
            --package "$APP_PACKAGE"

          # Create release
          gpd publish release \
            --package "$APP_PACKAGE" \
            --track "$TRACK" \
            --status "$STATUS"

          # For production, start with 10% rollout
          if [[ "$TRACK" == "production" ]]; then
            gpd publish rollout \
              --package "$APP_PACKAGE" \
              --track production \
              --percentage 10
          fi
```

## GitLab CI Examples

### Basic Pipeline Configuration

```yaml
stages:
  - build
  - deploy

variables:
  APP_PACKAGE: "com.example.app"
  GPD_VERSION: "latest"

build:
  stage: build
  image: gradle:8-jdk17
  script:
    - ./gradlew bundleRelease
  artifacts:
    paths:
      - app/build/outputs/bundle/release/app-release.aab
    expire_in: 1 hour

deploy:
  stage: deploy
  image: golang:1.22
  before_script:
    - curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
  script:
    - |
      # Verify authentication
      gpd auth status
      gpd auth check --package "$APP_PACKAGE"
      
      # Upload AAB
      gpd publish upload \
        app/build/outputs/bundle/release/app-release.aab \
        --package "$APP_PACKAGE"
      
      # Create release
      gpd publish release \
        --package "$APP_PACKAGE" \
        --track internal \
        --status completed
  only:
    - tags
  environment:
    name: play-store/internal
```

### Release Workflow with Manual Approval

```yaml
stages:
  - build
  - test
  - deploy-internal
  - deploy-production

variables:
  APP_PACKAGE: "com.example.app"

build:
  stage: build
  image: gradle:8-jdk17
  script:
    - ./gradlew bundleRelease
  artifacts:
    paths:
      - app/build/outputs/bundle/release/app-release.aab

test:
  stage: test
  image: gradle:8-jdk17
  script:
    - ./gradlew test
    - ./gradlew lint

deploy-internal:
  stage: deploy-internal
  image: golang:1.22
  before_script:
    - curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
  script:
    - |
      gpd publish upload \
        app/build/outputs/bundle/release/app-release.aab \
        --package "$APP_PACKAGE"
      
      gpd publish release \
        --package "$APP_PACKAGE" \
        --track internal \
        --status completed
  only:
    - develop
  environment:
    name: play-store/internal
    url: https://play.google.com/console

deploy-production:
  stage: deploy-production
  image: golang:1.22
  before_script:
    - curl -fsSL https://raw.githubusercontent.com/dl-alexandre/gpd/main/install.sh | bash
    - export PATH="$HOME/.local/bin:$PATH"
  script:
    - |
      # Promote from beta to production
      gpd publish promote \
        --package "$APP_PACKAGE" \
        --from-track beta \
        --to-track production
      
      # Start staged rollout
      gpd publish rollout \
        --package "$APP_PACKAGE" \
        --track production \
        --percentage 10
  when: manual
  only:
    - main
  environment:
    name: play-store/production
    url: https://play.google.com/console
```

## Common Patterns

### Using Edit Workflow for Atomic Releases

The edit workflow allows you to make multiple changes atomically before committing:

```yaml
# Example: Upload AAB and update release notes in a single transaction
- name: Atomic Release
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    # Create edit
    EDIT_ID=$(gpd publish edit create --package "$APP_PACKAGE" --output json | jq -r '.data.id')
    
    # Upload AAB within edit
    gpd publish upload \
      app/build/outputs/bundle/release/app-release.aab \
      --package "$APP_PACKAGE" \
      --edit-id "$EDIT_ID" \
      --no-auto-commit
    
    # Update release notes within same edit
    gpd publish release \
      --package "$APP_PACKAGE" \
      --track production \
      --status draft \
      --release-notes "Release notes here" \
      --edit-id "$EDIT_ID" \
      --no-auto-commit
    
    # Validate edit before committing
    gpd publish edit validate "$EDIT_ID" --package "$APP_PACKAGE"
    
    # Commit edit (all changes go live together)
    gpd publish edit commit "$EDIT_ID" --package "$APP_PACKAGE"
```

### Automated Version Code Management

Extract and use version codes from your build:

```yaml
- name: Get Version Information
  id: version
  run: |
    # Extract version code from build.gradle or build output
    VERSION_CODE=$(grep -oP 'versionCode\s+\K\d+' app/build.gradle)
    VERSION_NAME=$(grep -oP 'versionName\s+"\K[^"]+' app/build.gradle)
    
    echo "version_code=$VERSION_CODE" >> $GITHUB_OUTPUT
    echo "version_name=$VERSION_NAME" >> $GITHUB_OUTPUT

- name: Deploy with Version
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
    VERSION_CODE: ${{ steps.version.outputs.version_code }}
  run: |
    gpd publish upload \
      app/build/outputs/bundle/release/app-release.aab \
      --package "$APP_PACKAGE"
    
    gpd publish release \
      --package "$APP_PACKAGE" \
      --track internal \
      --status completed \
      --version-code "$VERSION_CODE"
```

### Release Notes from Git Commits

Generate release notes from git commit messages:

```yaml
- name: Generate Release Notes
  id: release_notes
  run: |
    # Get commits since last tag
    LAST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    if [ -z "$LAST_TAG" ]; then
      COMMITS=$(git log --pretty=format:"- %s" HEAD~10..HEAD)
    else
      COMMITS=$(git log --pretty=format:"- %s" "$LAST_TAG"..HEAD)
    fi
    
    # Format release notes
    RELEASE_NOTES=$(cat <<EOF
    ## What's New
    
    $COMMITS
    
    ## Build Information
    - Commit: ${{ github.sha }}
    - Build: ${{ github.run_number }}
    EOF
    )
    
    echo "notes<<EOF" >> $GITHUB_OUTPUT
    echo "$RELEASE_NOTES" >> $GITHUB_OUTPUT
    echo "EOF" >> $GITHUB_OUTPUT

- name: Deploy with Release Notes
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
    RELEASE_NOTES: ${{ steps.release_notes.outputs.notes }}
  run: |
    gpd publish upload \
      app/build/outputs/bundle/release/app-release.aab \
      --package "$APP_PACKAGE"
    
    # Save release notes to file
    echo "$RELEASE_NOTES" > /tmp/release_notes.txt
    
    gpd publish release \
      --package "$APP_PACKAGE" \
      --track internal \
      --status completed \
      --release-notes-file /tmp/release_notes.txt
```

### Conditional Deployment Based on Changes

Only deploy when relevant files change:

```yaml
name: Smart Deploy

on:
  push:
    branches: [main]
    paths:
      - 'app/**'
      - 'build.gradle'
      - 'gradle.properties'

jobs:
  check-changes:
    runs-on: ubuntu-latest
    outputs:
      should-deploy: ${{ steps.filter.outputs.should-deploy }}
    steps:
      - uses: actions/checkout@v4
      - uses: dorny/paths-filter@v2
        id: filter
        with:
          filters: |
            deploy:
              - 'app/**'
              - 'build.gradle'

  deploy:
    needs: check-changes
    if: needs.check-changes.outputs.should-deploy == 'true'
    runs-on: ubuntu-latest
    steps:
      # ... deployment steps
```

## Best Practices

### Security Considerations

1. **Never Commit Credentials**:
   - Always use CI/CD secrets/variables
   - Never hardcode service account keys
   - Rotate keys periodically

2. **Least Privilege Principle**:
   - Grant only necessary permissions to service account
   - Use separate service accounts for different environments
   - Review permissions regularly

3. **Secure Secret Storage**:
   - Use platform-native secret management (GitHub Secrets, GitLab Variables)
   - Enable secret scanning in repositories
   - Use encrypted variables when possible

4. **Audit Logging**:
   ```yaml
   - name: Log Deployment
     run: |
       echo "Deployed by: ${{ github.actor }}"
       echo "Commit: ${{ github.sha }}"
       echo "Track: production"
       # Send to your logging system
   ```

### Error Handling in CI

Handle errors gracefully and provide actionable feedback:

```yaml
- name: Deploy with Error Handling
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    set -e  # Exit on error
    
    # Check authentication first
    if ! gpd auth status --output json | jq -e '.data.authenticated == true' > /dev/null; then
      echo "❌ Authentication failed"
      exit 1
    fi
    
    # Upload with retry logic
    MAX_RETRIES=3
    RETRY_COUNT=0
    
    while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
      if gpd publish upload \
        app/build/outputs/bundle/release/app-release.aab \
        --package "$APP_PACKAGE"; then
        echo "✅ Upload successful"
        break
      else
        RETRY_COUNT=$((RETRY_COUNT + 1))
        if [ $RETRY_COUNT -lt $MAX_RETRIES ]; then
          echo "⏳ Retrying upload (attempt $RETRY_COUNT/$MAX_RETRIES)..."
          sleep 5
        else
          echo "❌ Upload failed after $MAX_RETRIES attempts"
          exit 1
        fi
      fi
    done
    
    # Create release
    if ! gpd publish release \
      --package "$APP_PACKAGE" \
      --track internal \
      --status completed; then
      echo "❌ Release creation failed"
      exit 1
    fi
```

### Rollback Strategies

Implement automated rollback capabilities:

```yaml
- name: Rollback on Failure
  if: failure()
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    # Get current production version
    CURRENT_VERSION=$(gpd publish status \
      --package "$APP_PACKAGE" \
      --track production \
      --output json | jq -r '.data.releases[0].versionCodes[0]')
    
    # Rollback to previous version
    gpd publish rollback \
      --package "$APP_PACKAGE" \
      --track production \
      --version-code "$PREVIOUS_VERSION" \
      --confirm
    
    # Notify team
    echo "🚨 Rollback executed: $CURRENT_VERSION -> $PREVIOUS_VERSION"
```

### Pre-deployment Validation

Validate before deploying:

```yaml
- name: Pre-deployment Checks
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    # Check app capabilities
    CAPABILITIES=$(gpd publish capabilities \
      --package "$APP_PACKAGE" \
      --output json)
    
    # Verify we can create releases
    if ! echo "$CAPABILITIES" | jq -e '.data.canCreateReleases == true' > /dev/null; then
      echo "❌ Cannot create releases"
      exit 1
    fi
    
    # Check current track status
    STATUS=$(gpd publish status \
      --package "$APP_PACKAGE" \
      --track production \
      --output json)
    
    # Warn if production has issues
    if echo "$STATUS" | jq -e '.data.releases[0].status == "halted"' > /dev/null; then
      echo "⚠️  Warning: Production rollout is halted"
    fi
```

### Staged Rollout Automation

Automate gradual rollouts:

```yaml
- name: Staged Rollout
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    # Start at 10%
    gpd publish rollout \
      --package "$APP_PACKAGE" \
      --track production \
      --percentage 10
    
    # Wait and monitor (use scheduled workflows for automation)
    echo "Deployed to 10% of users"
    echo "Monitor metrics before increasing rollout"
    
    # Subsequent stages can be triggered manually or via scheduled workflows
    # Stage 2: 50%
    # Stage 3: 100%
```

## Troubleshooting

### Common CI/CD Issues and Solutions

#### Authentication Failures

**Problem**: `gpd auth status` returns `authenticated: false`

**Solutions**:
1. Verify the service account key is correctly set in secrets
2. Check that the JSON key is valid (no extra whitespace, proper escaping)
3. Ensure the service account has been granted access in Play Console
4. Verify the service account email matches the one in Play Console

```yaml
- name: Debug Authentication
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
  run: |
    # Check if key is set
    if [ -z "$GPD_SERVICE_ACCOUNT_KEY" ]; then
      echo "❌ GPD_SERVICE_ACCOUNT_KEY is not set"
      exit 1
    fi
    
    # Validate JSON structure
    echo "$GPD_SERVICE_ACCOUNT_KEY" | jq -e '.type == "service_account"' || {
      echo "❌ Invalid service account key format"
      exit 1
    }
    
    # Check authentication
    gpd auth status --output json | jq '.'
```

#### Permission Denied Errors

**Problem**: `permission denied` or `403 Forbidden` errors

**Solutions**:
1. Verify service account permissions in Play Console
2. Check that the app package name matches
3. Ensure the service account has access to the specific track

```yaml
- name: Verify Permissions
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    # Check permissions
    gpd auth check --package "$APP_PACKAGE" --output json | jq '.'
```

#### Upload Failures

**Problem**: AAB upload fails with validation errors

**Solutions**:
1. Verify AAB file exists and is valid
2. Check version code is higher than existing releases
3. Ensure AAB is signed correctly
4. Validate AAB format with `bundletool`

```yaml
- name: Validate AAB Before Upload
  run: |
    # Check file exists
    if [ ! -f "app/build/outputs/bundle/release/app-release.aab" ]; then
      echo "❌ AAB file not found"
      exit 1
    fi
    
    # Check file size (should be > 0)
    SIZE=$(stat -f%z app/build/outputs/bundle/release/app-release.aab 2>/dev/null || stat -c%s app/build/outputs/bundle/release/app-release.aab)
    if [ "$SIZE" -eq 0 ]; then
      echo "❌ AAB file is empty"
      exit 1
    fi
    
    echo "✅ AAB file validated: ${SIZE} bytes"
```

#### Edit Transaction Issues

**Problem**: Edit transactions expire or conflict

**Solutions**:
1. Use `--edit-id` to reuse existing edits
2. Commit edits promptly (they expire after 7 days)
3. Use `--no-auto-commit` only when necessary
4. Check for stale edits and clean them up

```yaml
- name: Clean Up Stale Edits
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    # List all edits
    EDITS=$(gpd publish edit list --package "$APP_PACKAGE" --output json)
    
    # Delete expired edits
    echo "$EDITS" | jq -r '.data.edits[] | select(.expiryTime < (now | strftime("%Y-%m-%dT%H:%M:%SZ"))) | .id' | \
      while read edit_id; do
        echo "Deleting expired edit: $edit_id"
        gpd publish edit delete "$edit_id" --package "$APP_PACKAGE" || true
      done
```

#### Network Timeouts

**Problem**: Requests timeout in CI environment

**Solutions**:
1. Increase timeout with `--timeout` flag
2. Add retry logic
3. Check CI runner network connectivity

```yaml
- name: Deploy with Extended Timeout
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    gpd publish upload \
      app/build/outputs/bundle/release/app-release.aab \
      --package "$APP_PACKAGE" \
      --timeout 5m
```

#### Version Code Conflicts

**Problem**: Version code already exists error

**Solutions**:
1. Extract version code from AAB or build configuration
2. Verify it's higher than existing releases
3. Use `gpd publish status` to check current version codes

```yaml
- name: Check Version Code
  env:
    GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
    APP_PACKAGE: ${{ secrets.APP_PACKAGE_NAME }}
  run: |
    # Get current version codes
    CURRENT_VERSIONS=$(gpd publish status \
      --package "$APP_PACKAGE" \
      --track production \
      --output json | jq -r '.data.releases[].versionCodes[]')
    
    # Get new version code
    NEW_VERSION=$(grep -oP 'versionCode\s+\K\d+' app/build.gradle)
    
    # Verify it's higher
    MAX_VERSION=$(echo "$CURRENT_VERSIONS" | sort -n | tail -1)
    if [ "$NEW_VERSION" -le "$MAX_VERSION" ]; then
      echo "❌ Version code $NEW_VERSION must be > $MAX_VERSION"
      exit 1
    fi
    
    echo "✅ Version code $NEW_VERSION is valid"
```

### Getting Help

If you encounter issues not covered here:

1. **Check CLI Help**:
   ```bash
   gpd --help
   gpd publish --help
   ```

2. **Enable Verbose Output**:
   ```yaml
   - name: Debug with Verbose Output
     run: |
       gpd --verbose publish upload app.aab --package com.example.app
   ```

3. **Check Exit Codes**:
   - `0`: Success
   - `1`: General API error
   - `2`: Authentication failure
   - `3`: Permission denied
   - `4`: Validation error
   - `5`: Rate limited
   - `6`: Network error
   - `7`: Not found
   - `8`: Conflict

4. **Review Logs**: Check CI/CD logs for detailed error messages

5. **Validate Configuration**: Use `gpd config doctor` to diagnose issues

---

For more information, see the [main README](../README.md) and [command reference](../README.md#command-reference).
