# Automation Workflows Guide

This guide demonstrates how to automate CI/CD pipelines, staged rollouts, release notes generation, and post-release monitoring using the Google Play Developer CLI.

## Table of Contents

1. [Overview](#overview)
2. [CI/CD Pipeline Integration](#cicd-pipeline-integration)
3. [Automated Staged Rollout](#automated-staged-rollout)
4. [Release Notes Generation](#release-notes-generation)
5. [Post-Release Monitoring](#post-release-monitoring)
6. [Best Practices](#best-practices)

---

## Overview

The `gpd` CLI provides powerful automation commands that integrate seamlessly with CI/CD pipelines. These commands enable:

- **Automated validation** before releases
- **Smart promotion** between tracks with verification
- **Staged rollouts** with health checks
- **Release notes generation** from git history
- **Post-release monitoring** with alerting

### Automation Command Structure

```
gpd automation
├── release-notes    # Generate release notes from git/PRs
├── rollout          # Automated staged rollout with health checks
├── promote          # Smart promote with verification
├── validate         # Pre-release validation
└── monitor          # Post-release health monitoring
```

---

## CI/CD Pipeline Integration

### Complete CI/CD Pipeline Example

A complete workflow from build to production:

```bash
#!/bin/bash
# complete-cicd-pipeline.sh

set -e

PACKAGE="com.example.app"
VERSION=$(grep versionCode app/build.gradle | grep -o '[0-9]\+')
EDIT_ID="release-v${VERSION}"

# 1. Pre-deployment validation
echo "=== Step 1: Pre-deployment Validation ==="
gpd automation validate \
  --package "$PACKAGE" \
  --checks all \
  --strict

# 2. Create edit transaction
echo "=== Step 2: Creating Edit Transaction ==="
gpd publish edit create \
  --package "$PACKAGE" \
  --edit-id "$EDIT_ID"

# 3. Upload AAB
echo "=== Step 3: Uploading AAB ==="
gpd publish upload \
  app/build/outputs/bundle/release/app-release.aab \
  --package "$PACKAGE" \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# 4. Generate and apply release notes
echo "=== Step 4: Generating Release Notes ==="
gpd automation release-notes \
  --source git \
  --since "v$(($VERSION-1))" \
  --format markdown \
  --output-file /tmp/release-notes.md

# 5. Create internal release
echo "=== Step 5: Creating Internal Release ==="
gpd publish release \
  --package "$PACKAGE" \
  --track internal \
  --status completed \
  --version-code "$VERSION" \
  --release-notes-file /tmp/release-notes.md \
  --edit-id "$EDIT_ID" \
  --no-auto-commit

# 6. Validate and commit
echo "=== Step 6: Validating and Committing ==="
gpd publish edit validate "$EDIT_ID" --package "$PACKAGE"
gpd publish edit commit "$EDIT_ID" --package "$PACKAGE"

# 7. Wait for internal testing
echo "=== Step 7: Waiting for Internal Testing (30 minutes) ==="
sleep 1800

# 8. Promote to beta
echo "=== Step 8: Promoting to Beta ==="
gpd automation promote \
  --package "$PACKAGE" \
  --from-track internal \
  --to-track beta \
  --verify \
  --verify-timeout 15m

# 9. Start staged rollout to production
echo "=== Step 9: Starting Production Rollout ==="
gpd automation rollout \
  --package "$PACKAGE" \
  --track production \
  --start-percentage 1 \
  --target-percentage 100 \
  --step-size 10 \
  --step-interval 2h \
  --health-threshold 0.01 \
  --auto-rollback

echo "=== Deployment Pipeline Complete ==="
```

### GitHub Actions Workflow

```yaml
name: Automated Release Pipeline

on:
  push:
    branches:
      - main
    tags:
      - 'v*'

env:
  PACKAGE: com.example.app
  GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}

jobs:
  validate:
    name: Pre-Release Validation
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Play-Developer-CLI/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH
      
      - name: Validate Release
        run: |
          gpd automation validate \
            --package "$PACKAGE" \
            --checks all \
            --strict

  deploy-internal:
    name: Deploy to Internal Testing
    needs: validate
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Java
        uses: actions/setup-java@v4
        with:
          distribution: 'temurin'
          java-version: '17'
      
      - name: Build AAB
        run: ./gradlew bundleRelease
      
      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Play-Developer-CLI/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH
      
      - name: Deploy to Internal
        run: |
          EDIT_ID="internal-$(date +%s)"
          
          gpd publish upload \
            app/build/outputs/bundle/release/app-release.aab \
            --package "$PACKAGE" \
            --edit-id "$EDIT_ID" \
            --no-auto-commit
          
          gpd automation release-notes \
            --source git \
            --since "$(git describe --tags --abbrev=0)" \
            --format markdown \
            --output-file release-notes.md
          
          gpd publish release \
            --package "$PACKAGE" \
            --track internal \
            --status completed \
            --release-notes-file release-notes.md \
            --edit-id "$EDIT_ID"

  promote-beta:
    name: Promote to Beta
    needs: deploy-internal
    runs-on: ubuntu-latest
    environment: beta
    steps:
      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Play-Developer-CLI/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH
      
      - name: Promote with Verification
        run: |
          gpd automation promote \
            --package "$PACKAGE" \
            --from-track internal \
            --to-track beta \
            --verify \
            --verify-timeout 30m

  rollout-production:
    name: Production Rollout
    needs: promote-beta
    runs-on: ubuntu-latest
    environment: production
    steps:
      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Play-Developer-CLI/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH
      
      - name: Start Staged Rollout
        run: |
          gpd automation rollout \
            --package "$PACKAGE" \
            --track production \
            --start-percentage 1 \
            --target-percentage 100 \
            --step-size 5 \
            --step-interval 4h \
            --health-threshold 0.01 \
            --auto-rollback \
            --wait

  monitor:
    name: Post-Release Monitoring
    needs: rollout-production
    runs-on: ubuntu-latest
    steps:
      - name: Install gpd
        run: |
          curl -fsSL https://raw.githubusercontent.com/dl-alexandre/Google-Play-Developer-CLI/main/install.sh | bash
          echo "$HOME/.local/bin" >> $GITHUB_PATH
      
      - name: Monitor Release Health
        run: |
          gpd automation monitor \
            --package "$PACKAGE" \
            --track production \
            --duration 2h \
            --check-interval 10m \
            --crash-threshold 0.01 \
            --exit-on-degradation
```

---

## Automated Staged Rollout

The `gpd automation rollout` command performs automated staged rollouts with health checks and optional auto-rollback.

### Basic Staged Rollout

Start a rollout from 1% to 100% in 10% increments:

```bash
gpd automation rollout \
  --package com.example.app \
  --track production \
  --start-percentage 1 \
  --target-percentage 100 \
  --step-size 10 \
  --step-interval 2h \
  --wait
```

**Output:**
```json
{
  "data": {
    "track": "production",
    "finalPercentage": 100,
    "stepsCompleted": 10,
    "steps": [1, 10, 20, 30, 40, 50, 60, 70, 80, 90, 100],
    "healthThreshold": 0.01
  }
}
```

### Rollout with Health Checks

Monitor crash rates during rollout:

```bash
gpd automation rollout \
  --package com.example.app \
  --track production \
  --start-percentage 5 \
  --target-percentage 100 \
  --step-size 5 \
  --step-interval 1h \
  --health-threshold 0.01 \
  --wait
```

### Auto-Rollback on Health Failure

Automatically rollback if crash rates exceed threshold:

```bash
gpd automation rollout \
  --package com.example.app \
  --track production \
  --start-percentage 10 \
  --target-percentage 100 \
  --step-size 10 \
  --step-interval 30m \
  --health-threshold 0.005 \
  --auto-rollback \
  --wait
```

**On Health Check Failure:**
```json
{
  "data": {
    "track": "production",
    "finalPercentage": 30,
    "stepsCompleted": 3,
    "healthCheckFailed": true,
    "rollbackInitiated": true,
    "message": "Health check failed at 30% rollout - automatically rolling back"
  },
  "error": {
    "code": "GENERAL_ERROR",
    "message": "Health check failed at rollout percentage",
    "details": {
      "currentPercentage": 30,
      "crashRate": 0.015
    }
  }
}
```

### Dry Run Rollout Planning

Preview rollout steps without executing:

```bash
gpd automation rollout \
  --package com.example.app \
  --track production \
  --start-percentage 1 \
  --target-percentage 100 \
  --step-size 25 \
  --step-interval 1h \
  --dry-run
```

**Output:**
```json
{
  "data": {
    "plan": {
      "track": "production",
      "startPercentage": 1,
      "targetPercentage": 100,
      "steps": [1, 25, 50, 75, 100],
      "stepInterval": "1h0m0s",
      "healthThreshold": 0.01,
      "autoRollback": false
    }
  },
  "meta": {
    "noop": "dry-run mode"
  }
}
```

### Conservative Rollout Strategy

For high-risk releases, use smaller steps with longer intervals:

```bash
gpd automation rollout \
  --package com.example.app \
  --track production \
  --start-percentage 1 \
  --target-percentage 100 \
  --step-size 5 \
  --step-interval 4h \
  --health-threshold 0.005 \
  --auto-rollback \
  --wait
```

### Rollout Monitoring Script

```bash
#!/bin/bash
# rollout-monitor.sh

PACKAGE="com.example.app"
LOG_FILE="/var/log/gpd-rollout-$(date +%Y%m%d-%H%M%S).log"

# Function to send notification
notify() {
  local message="$1"
  echo "$(date): $message" | tee -a "$LOG_FILE"
  # Add Slack/Discord webhook here
  # curl -X POST -H 'Content-type: application/json' \
  #   --data "{\"text\":\"$message\"}" \
  #   "$SLACK_WEBHOOK_URL"
}

# Start rollout
notify "Starting staged rollout for $PACKAGE"

if gpd automation rollout \
  --package "$PACKAGE" \
  --track production \
  --start-percentage 1 \
  --target-percentage 100 \
  --step-size 10 \
  --step-interval 2h \
  --health-threshold 0.01 \
  --auto-rollback \
  --wait \
  --output json > /tmp/rollout-result.json 2>&1; then
  
  FINAL_PCT=$(jq -r '.data.finalPercentage' /tmp/rollout-result.json)
  notify "✅ Rollout completed successfully at ${FINAL_PCT}%"
else
  notify "❌ Rollout failed - check logs at $LOG_FILE"
  exit 1
fi
```

---

## Release Notes Generation

Automatically generate release notes from git commits or pull requests.

### Generate from Git History

Create release notes from commits since the last tag:

```bash
gpd automation release-notes \
  --source git \
  --format markdown \
  --output-file release-notes.md
```

**Output (Markdown):**
```markdown
## What's New

- feat: Add dark mode support
- fix: Resolve memory leak in image loading
- perf: Optimize database queries
- docs: Update API documentation
- feat: Implement push notifications
```

**Output (JSON):**
```json
{
  "data": {
    "commits": [
      {
        "hash": "a1b2c3d4",
        "message": "feat: Add dark mode support",
        "author": "Jane Smith",
        "email": "jane@example.com",
        "date": "2024-01-15"
      },
      {
        "hash": "e5f6g7h8",
        "message": "fix: Resolve memory leak in image loading",
        "author": "John Doe",
        "email": "john@example.com",
        "date": "2024-01-14"
      }
    ],
    "count": 5,
    "since": "v1.2.0",
    "until": "HEAD"
  }
}
```

### Custom Git Range

Generate notes from a specific commit range:

```bash
gpd automation release-notes \
  --source git \
  --since "v1.2.0" \
  --until "v1.3.0-rc1" \
  --format markdown \
  --output-file release-notes-v1.3.0.md
```

### Limit Commit Count

For large releases, limit the number of commits:

```bash
gpd automation release-notes \
  --source git \
  --since "v1.0.0" \
  --max-commits 20 \
  --format markdown
```

### Generate from Pull Requests

Create release notes from merged PRs (requires package specification):

```bash
gpd automation release-notes \
  --source pr \
  --package com.example.app \
  --format markdown \
  --output-file release-notes.md
```

**Note:** PR-based generation requires additional setup for GitHub API access.

### Use with Release Command

Combine with release creation:

```bash
# Generate release notes
gpd automation release-notes \
  --source git \
  --since "$(git describe --tags --abbrev=0)" \
  --format markdown \
  --output-file /tmp/notes.md

# Create release with notes
gpd publish release \
  --package com.example.app \
  --track production \
  --status draft \
  --release-notes-file /tmp/notes.md
```

### CI/CD Integration

```yaml
- name: Generate Release Notes
  id: release_notes
  run: |
    gpd automation release-notes \
      --source git \
      --since "${{ github.event.before }}" \
      --format markdown \
      --output-file release-notes.md
    
    # Preview in workflow logs
    cat release-notes.md

- name: Create Release with Notes
  run: |
    gpd publish release \
      --package ${{ env.PACKAGE }} \
      --track internal \
      --status completed \
      --release-notes-file release-notes.md
```

---

## Post-Release Monitoring

Monitor release health after deployment with automated health checks.

### Basic Post-Release Monitoring

Monitor a production release for 2 hours:

```bash
gpd automation monitor \
  --package com.example.app \
  --track production \
  --duration 2h \
  --check-interval 5m
```

**Output:**
```json
{
  "data": {
    "monitoring": {
      "track": "production",
      "duration": "2h0m0s",
      "checksPerformed": 24,
      "status": "healthy",
      "degradations": 0,
      "thresholds": {
        "crash": 0.01,
        "anr": 0.005,
        "error": 0.02
      },
      "checks": [
        {
          "timestamp": "2024-01-20T10:00:00Z",
          "crashRate": 0.002,
          "anrRate": 0.001,
          "errorRate": 0.005,
          "status": "healthy"
        }
      ]
    }
  }
}
```

### Monitor with Custom Thresholds

Set custom health thresholds:

```bash
gpd automation monitor \
  --package com.example.app \
  --track production \
  --duration 4h \
  --check-interval 10m \
  --crash-threshold 0.005 \
  --anr-threshold 0.003 \
  --error-threshold 0.01
```

### Exit on Degradation

Fail CI/CD pipeline if health degrades:

```bash
gpd automation monitor \
  --package com.example.app \
  --track production \
  --duration 2h \
  --crash-threshold 0.01 \
  --exit-on-degradation
```

**On Degradation:**
```json
{
  "data": {
    "monitoring": {
      "track": "production",
      "duration": "45m0s",
      "checksPerformed": 9,
      "status": "degraded",
      "degradations": 1,
      "thresholds": {
        "crash": 0.01,
        "anr": 0.005,
        "error": 0.02
      }
    }
  },
  "error": {
    "code": "GENERAL_ERROR",
    "message": "release health degraded during monitoring",
    "details": {
      "degradations": 1,
      "thresholds": {
        "crash": 0.01,
        "anr": 0.005,
        "error": 0.02
      }
    }
  }
}
```

### Enable Auto-Alerts

Send notifications when thresholds are breached:

```bash
gpd automation monitor \
  --package com.example.app \
  --track beta \
  --duration 1h \
  --check-interval 5m \
  --crash-threshold 0.01 \
  --auto-alert \
  --exit-on-degradation
```

### Dry Run Monitoring Plan

Preview monitoring plan:

```bash
gpd automation monitor \
  --package com.example.app \
  --track production \
  --duration 2h \
  --check-interval 5m \
  --dry-run
```

**Output:**
```json
{
  "data": {
    "monitoringPlan": {
      "track": "production",
      "duration": "2h0m0s",
      "checkInterval": "5m0s",
      "crashThreshold": 0.01,
      "anrThreshold": 0.005,
      "errorThreshold": 0.02,
      "autoAlert": false,
      "exitOnDegradation": false
    }
  },
  "meta": {
    "noop": "dry-run mode"
  }
}
```

### Continuous Monitoring Script

```bash
#!/bin/bash
# continuous-monitor.sh

PACKAGE="com.example.app"
TRACK="production"

# Monitor continuously with alerts
while true; do
  echo "Starting monitoring cycle at $(date)"
  
  if gpd automation monitor \
    --package "$PACKAGE" \
    --track "$TRACK" \
    --duration 1h \
    --check-interval 5m \
    --crash-threshold 0.01 \
    --auto-alert \
    --output json > /tmp/monitor-result.json; then
    
    echo "Monitoring cycle completed successfully"
  else
    echo "ALERT: Health degradation detected at $(date)"
    # Send alert to Slack/PagerDuty
    # curl -X POST ...
  fi
  
  # Wait before next cycle
  sleep 300
done
```

---

## Best Practices

### 1. Always Validate Before Automation

```bash
# Validate before starting automation
gpd automation validate --package com.example.app --checks all --strict

# Use dry-run to preview changes
gpd automation rollout ... --dry-run
gpd automation release-notes ... --dry-run
```

### 2. Use Appropriate Health Thresholds

```bash
# Conservative thresholds for production
--crash-threshold 0.005 \
--anr-threshold 0.003 \
--error-threshold 0.01

# Relaxed thresholds for internal testing
--crash-threshold 0.02 \
--anr-threshold 0.01 \
--error-threshold 0.05
```

### 3. Implement Gradual Rollouts

```bash
# Week 1: Small percentage, longer intervals
gpd automation rollout \
  --start-percentage 1 \
  --step-size 5 \
  --step-interval 4h

# Week 2: If healthy, increase pace
--start-percentage 25 \
--step-size 25 \
--step-interval 2h
```

### 4. Combine Commands for Complete Workflows

```bash
#!/bin/bash
set -e

PACKAGE="com.example.app"
VERSION=$(grep versionCode app/build.gradle | grep -o '[0-9]\+')

# Validate
gpd automation validate --package "$PACKAGE" --checks all

# Generate notes
gpd automation release-notes \
  --source git \
  --output-file notes.md \
  --since "v$(($VERSION-1))"

# Deploy
gpd publish upload app.aab --package "$PACKAGE"

# Monitor
gpd automation monitor \
  --package "$PACKAGE" \
  --track production \
  --duration 2h \
  --exit-on-degradation
```

### 5. Error Handling and Recovery

```bash
#!/bin/bash
set -e

PACKAGE="com.example.app"
ROLLBACK_VERSION=""

# Capture current production version for potential rollback
ROLLBACK_VERSION=$(gpd publish status \
  --package "$PACKAGE" \
  --track production \
  --output json | jq -r '.data.releases[0].versionCodes[0]')

echo "Current production version: $ROLLBACK_VERSION"

# Attempt rollout
if ! gpd automation rollout \
  --package "$PACKAGE" \
  --track production \
  --start-percentage 10 \
  --target-percentage 100 \
  --step-size 10 \
  --health-threshold 0.01 \
  --auto-rollback \
  --wait; then
  
  echo "Rollout failed - initiating manual rollback"
  gpd publish rollback \
    --package "$PACKAGE" \
    --track production \
    --version-code "$ROLLBACK_VERSION" \
    --confirm
fi
```

### 6. Document Your Automation

```bash
# Add comments to complex automation scripts
#!/bin/bash
# Production Release Automation
# 
# Prerequisites:
#   - gpd CLI installed and authenticated
#   - Service account with Release Manager permissions
#   - Slack webhook configured for alerts
#
# Usage: ./release-production.sh <version-code>

set -e

# ... script continues
```

---

## Summary

| Workflow | Command | Example Count |
|----------|---------|---------------|
| CI/CD Pipeline | Complete workflow | 2 examples |
| Staged Rollout | `gpd automation rollout` | 7 examples |
| Release Notes | `gpd automation release-notes` | 6 examples |
| Post-Release Monitor | `gpd automation monitor` | 7 examples |
| **Total Examples** | | **22 examples** |

---

## Related Commands

- [`gpd publish promote`](./release-workflow.md) - Manual track promotion
- [`gpd publish rollout`](./release-workflow.md) - Manual rollout control
- [`gpd vitals anomalies`](./monitoring-setup.md) - Anomaly detection
- [`gpd config doctor`](./error-debugging.md) - Configuration validation
