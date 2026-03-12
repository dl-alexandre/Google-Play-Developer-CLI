# Workflow Documentation

The `gpd workflow` command provides declarative workflow execution for Google Play Developer Console operations. Define multi-step workflows in JSON files, capture step outputs, reference them in subsequent steps, and resume failed runs.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Workflow Definition](#workflow-definition)
- [Variable Interpolation](#variable-interpolation)
- [CLI Commands](#cli-commands)
- [Examples](#examples)
- [Resume and Recovery](#resume-and-recovery)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)

## Overview

Workflows enable you to:

- Define multi-step release pipelines in JSON
- Capture outputs from gpd CLI commands
- Reference captured outputs in subsequent steps via `${steps.<name>.<field>}`
- Resume failed runs without re-executing completed steps
- Persist workflow state locally for recovery

### Command Structure

```
gpd workflow
├── run       # Execute a workflow from JSON file
├── list      # List available workflows and runs
├── show      # Display workflow definition
├── status    # Show detailed run status
├── validate  # Validate workflow file for errors
└── logs      # Show logs from a workflow run
```

## Quick Start

### 1. Create a Workflow

Create `release.json`:

```json
{
  "name": "production-release",
  "description": "Release app to production with validation",
  "env": {
    "PACKAGE": "com.example.app"
  },
  "steps": [
    {
      "name": "upload",
      "command": "gpd publish upload app.aab --package ${env.PACKAGE} --output json",
      "captureOutputs": ["versionCode", "editId"]
    },
    {
      "name": "validate",
      "command": "gpd automation validate --package ${env.PACKAGE} --checks all --strict",
      "dependsOn": ["upload"]
    },
    {
      "name": "release",
      "command": "gpd publish release --package ${env.PACKAGE} --track internal --version-code ${steps.upload.versionCode}",
      "dependsOn": ["validate"]
    }
  ]
}
```

### 2. Run the Workflow

```bash
gpd workflow run --file release.json
```

### 3. Resume if Needed

If a step fails, get the run ID from the output and resume:

```bash
gpd workflow run --file release.json --resume <run-id>
```

## Workflow Definition

### Schema

```json
{
  "name": "workflow-name",
  "description": "Optional description",
  "env": {
    "KEY": "value"
  },
  "steps": [
    {
      "name": "step-name",
      "type": "gpd|shell",
      "command": "command to execute",
      "workingDir": "optional/working/directory",
      "env": {
        "STEP_VAR": "value"
      },
      "dependsOn": ["previous-step"],
      "captureOutputs": ["field1", "field2"],
      "condition": "${env.SHOULD_RUN}",
      "continueOnError": false,
      "timeout": "5m"
    }
  ]
}
```

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | **Required.** Workflow identifier |
| `description` | string | Human-readable description |
| `env` | object | Environment variables available to all steps |
| `steps` | array | **Required.** List of workflow steps |

### Step Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | **Required.** Unique step identifier |
| `type` | string | Command type: `gpd` (default) or `shell` |
| `command` | string | **Required.** Command to execute |
| `workingDir` | string | Working directory for execution |
| `env` | object | Step-specific environment variables |
| `dependsOn` | array | Names of steps that must complete first |
| `captureOutputs` | array | JSON fields to capture from stdout |
| `condition` | string | Condition to check before running |
| `continueOnError` | boolean | Continue workflow even if step fails |
| `timeout` | string | Maximum execution time (e.g., "5m", "30s") |
| `retryCount` | int | Number of retry attempts on failure |
| `retryDelay` | string | Delay between retry attempts (e.g., "5s", "1m") |
| `retryBackoff` | string | Backoff strategy: `linear` (default) or `exponential` |
| `parallel` | boolean | Run step concurrently with other parallel steps at the same level |

### Parallel Execution

Steps can be executed in parallel when they:
1. Have no dependencies on each other
2. Are marked with `"parallel": true`

**Workflow-level settings:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `maxParallel` | int | 4 | Maximum number of parallel steps to execute simultaneously |

**Example - Parallel uploads for multiple apps:**

```json
{
  "name": "multi-package-parallel",
  "description": "Upload multiple apps in parallel",
  "maxParallel": 3,
  "steps": [
    {
      "name": "upload_app1",
      "command": "gpd publish upload app1.aab --package com.example.app1 --output json",
      "parallel": true,
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "upload_app2",
      "command": "gpd publish upload app2.aab --package com.example.app2 --output json",
      "parallel": true,
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "upload_app3",
      "command": "gpd publish upload app3.aab --package com.example.app3 --output json",
      "parallel": true,
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "notify",
      "command": "gpd automation notify --message \"All apps uploaded\"",
      "dependsOn": ["upload_app1", "upload_app2", "upload_app3"]
    }
  ]
}
```

**How parallel execution works:**

1. Steps are grouped into "levels" based on their dependencies
2. Within each level, steps marked as `parallel: true` execute concurrently (up to `maxParallel` at a time)
3. Steps without `parallel: true` execute sequentially after parallel steps complete
4. All steps in a level must complete before any dependent steps in the next level start

**Example - Mixed parallel and sequential:**

```json
{
  "name": "build-and-deploy",
  "maxParallel": 4,
  "steps": [
    {
      "name": "lint",
      "command": "./gradlew lint",
      "parallel": true
    },
    {
      "name": "unit_tests",
      "command": "./gradlew test",
      "parallel": true
    },
    {
      "name": "build",
      "command": "./gradlew bundleRelease",
      "dependsOn": ["lint", "unit_tests"],
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "upload_prod",
      "command": "gpd publish upload app.aab --package com.example.app --track production",
      "dependsOn": ["build"],
      "parallel": true
    },
    {
      "name": "upload_beta",
      "command": "gpd publish upload app.aab --package com.example.app --track beta",
      "dependsOn": ["build"],
      "parallel": true
    }
  ]
}
```

In this example:
- `lint` and `unit_tests` run in parallel (level 0)
- `build` runs after both complete (level 1)
- `upload_prod` and `upload_beta` run in parallel after build (level 2)

### Step Types

**GPD commands** (`type: "gpd"`):
```json
{
  "name": "upload",
  "type": "gpd",
  "command": "gpd publish upload app.aab --package com.example.app"
}
```

**Shell commands** (`type: "shell"` or inferred):
```json
{
  "name": "build",
  "type": "shell",
  "command": "./gradlew assembleRelease"
}
```

If `type` is omitted, it defaults to `gpd` if the command starts with "gpd ", otherwise `shell`.

## Variable Interpolation

Workflows support variable interpolation using `${...}` syntax.

### Step Outputs

Reference outputs from previous steps:

```json
{
  "command": "gpd publish release --version-code ${steps.build.versionCode}"
}
```

Supported fields:
- `${steps.<name>.<field>}` - Access captured JSON output fields
- `${steps.<name>.exitCode}` - Access exit code
- `${steps.<name>.stdout}` - Access full stdout
- `${steps.<name>.stderr}` - Access full stderr

### Environment Variables

```json
{
  "command": "gpd publish upload --package ${env.PACKAGE}",
  "condition": "${env.SHOULD_RELEASE}"
}
```

Environment variable resolution:
1. Step's `env` field
2. Workflow's `env` field
3. System environment variables

### Nested Fields

Access nested JSON output:

```json
{
  "captureOutputs": ["data.versionCode"],
  "command": "echo ${steps.upload.data.versionCode}"
}
```

## Retry Logic

Steps can be configured to automatically retry on failure. This is useful for handling transient failures like network issues or rate limiting.

### Basic Retry Configuration

```json
{
  "name": "upload-with-retry",
  "command": "gpd publish upload app.aab --package com.example.app",
  "retryCount": 3,
  "retryDelay": "5s"
}
```

### Retry Fields

| Field | Type | Description |
|-------|------|-------------|
| `retryCount` | integer | Number of retry attempts (default: 0) |
| `retryDelay` | string | Base delay between attempts (e.g., "5s", "1m") |
| `retryBackoff` | string | Backoff strategy: `linear` or `exponential` |

### Backoff Strategies

**Linear** (default): Each retry waits `retryDelay * attempt_number`
```json
{
  "retryCount": 3,
  "retryDelay": "5s",
  "retryBackoff": "linear"
}
```
Retries after: 5s, 10s, 15s

**Exponential**: Each retry waits `retryDelay * 2^(attempt_number-1)`
```json
{
  "retryCount": 3,
  "retryDelay": "5s",
  "retryBackoff": "exponential"
}
```
Retries after: 5s, 10s, 20s

### Retry with API Calls

```json
{
  "name": "api-call-with-retry",
  "command": "gpd automation validate --package ${env.PACKAGE}",
  "retryCount": 5,
  "retryDelay": "10s",
  "retryBackoff": "exponential",
  "timeout": "5m"
}
```

### Monitoring Retries

When a step retries, the runner logs each attempt:
```
[workflow] Executing step: api-call
[workflow] ERROR: Step api-call failed (attempt 1/4): connection timeout
[workflow] Retrying step api-call (attempt 1/3) after 10s delay
[workflow] Executing step: api-call
[workflow] Step api-call completed successfully after 2 retries
```

The final step output includes retry information:
```json
{
  "stepName": "api-call",
  "exitCode": 0,
  "retryCount": 3,
  "retries": 2
}
```

## CLI Commands

### `gpd workflow run`

Execute a workflow from a JSON file.

```bash
gpd workflow run --file workflow.json [flags]
```

**Flags:**
- `--file, -f` - Path to workflow JSON file (required)
- `--resume` - Resume a previous run by ID
- `--dry-run` - Validate workflow without executing
- `--force` - Re-run completed steps even in resume mode
- `--watch` - Watch workflow execution with real-time progress updates
- `--watch-format` - Watch mode output format: `text` (default), `json`, or `tui`
- `--env KEY=VALUE` - Set environment variables

**Examples:**

```bash
# Run a workflow
gpd workflow run --file release.json

# Validate only
gpd workflow run --file release.json --dry-run

# Resume from failure
gpd workflow run --file release.json --resume 1710201600-abc123

# Force re-run all steps
gpd workflow run --file release.json --resume 1710201600-abc123 --force

# Set environment variables
gpd workflow run --file release.json --env PACKAGE=com.example.app --env VERSION=1.0.0

# Watch with simple text progress
gpd workflow run --file release.json --watch

# Watch with JSON output for automation
gpd workflow run --file release.json --watch --watch-format json

# Watch with TUI-style progress bar
gpd workflow run --file release.json --watch --watch-format tui
```

### `gpd workflow list`

List available workflows and run history.

```bash
gpd workflow list [flags]
```

**Flags:**
- `--all, -a` - Include run history

**Examples:**

```bash
# List workflow definitions
gpd workflow list

# Include run history
gpd workflow list --all
```

### `gpd workflow show`

Display workflow definition and details.

```bash
gpd workflow show <workflow-name-or-path>
```

**Examples:**

```bash
# Show workflow by name
gpd workflow show release

# Show workflow by path
gpd workflow show ./workflows/production.json
```

### `gpd workflow status`

Show detailed status of a workflow run.

```bash
gpd workflow status <run-id>
```

**Examples:**

```bash
# Check run status
gpd workflow status 1710201600-abc123
```

### `gpd workflow validate`

Validate a workflow file for errors without executing it. Performs comprehensive validation including:

- **JSON Schema Validation** - Ensures the file is valid JSON with required fields
- **Circular Dependency Detection** - Checks for cycles in step dependencies
- **Step Name Uniqueness** - Verifies all step names are unique
- **Dependency Resolution** - Confirms all dependencies reference existing steps
- **Command Syntax Validation** - Checks for unbalanced braces and empty commands
- **Variable Interpolation Validation** - Verifies step references and warns about unset environment variables

```bash
gpd workflow validate <workflow-file-path>
```

**Examples:**

```bash
# Validate a workflow file
gpd workflow validate ./release.json

# Validate with JSON output for programmatic use
gpd workflow validate ./release.json --output json

# Validate with table output for readability
gpd workflow validate ./release.json --output table
```

**Validation Output:**

The command returns a detailed validation report:

```json
{
  "data": {
    "file": "./release.json",
    "valid": true,
    "workflow": "production-release",
    "steps": 5,
    "issues": [],
    "summary": {
      "total": 0,
      "errors": 0,
      "warnings": 0
    }
  },
  "meta": {
    "services": ["workflow"]
  }
}
```

**Validation Issues:**

When issues are found, they are categorized as:

- **Errors** - Critical issues that would prevent the workflow from running (duplicate step names, unknown dependencies, circular dependencies, malformed commands)
- **Warnings** - Non-critical issues that might cause problems at runtime (references to environment variables not defined in the workflow)

Example error output:

```json
{
  "data": {
    "file": "./broken.json",
    "valid": false,
    "issues": [
      {
        "type": "error",
        "field": "dependsOn",
        "step": "deploy",
        "message": "step 'deploy' has unknown dependency: 'nonexistent'"
      }
    ],
    "summary": {
      "total": 1,
      "errors": 1,
      "warnings": 0
    }
  },
  "meta": {
    "services": ["workflow"]
  }
}
```

## Watch Mode

The `--watch` flag provides real-time monitoring of workflow execution with progress updates streamed to stdout.

### Watch Formats

**Text Format** (default): Simple text-based progress with visual indicators
```bash
gpd workflow run --file release.json --watch
```

Output:
```
▶ Starting workflow: production-release (3 steps)
→ [1/3] Executing step: validate
✓ [1/3] Step completed: validate (duration: 2.345s)
→ [2/3] Executing step: upload
✓ [2/3] Step completed: upload (duration: 15.234s)
→ [3/3] Executing step: release
✓ [3/3] Step completed: release (duration: 1.456s)
✓ Workflow completed successfully: production-release (duration: 19.035s)
```

**JSON Format**: Structured JSON events for automation and tooling
```bash
gpd workflow run --file release.json --watch --watch-format json
```

Output:
```json
{"type":"workflow_started","timestamp":"2024-03-12T10:30:00Z","workflow":"production-release","runId":"1710238200-abc123","totalSteps":3}
{"type":"step_started","timestamp":"2024-03-12T10:30:00Z","workflow":"production-release","runId":"1710238200-abc123","stepName":"validate","stepNum":1,"totalSteps":3}
{"type":"step_completed","timestamp":"2024-03-12T10:30:02Z","workflow":"production-release","runId":"1710238200-abc123","stepName":"validate","stepNum":1,"totalSteps":3,"duration":2345000000}
```

**TUI Format**: Interactive progress bar display
```bash
gpd workflow run --file release.json --watch --watch-format tui
```

Output:
```
[████████████████████████████░░░░░░░░░░░░░░░░░░░░]  60% | 3/5 | upload | 23s
```

### Event Types

Watch mode emits the following event types:

| Event | Description |
|-------|-------------|
| `workflow_started` | Workflow execution has begun |
| `workflow_completed` | Workflow finished successfully |
| `workflow_failed` | Workflow failed with an error |
| `step_started` | A step has started execution |
| `step_completed` | A step completed successfully |
| `step_failed` | A step failed with an error |
| `step_skipped` | A step was skipped (condition not met or already completed) |

### Using Watch Mode with CI/CD

For CI/CD pipelines, JSON format is recommended:

```bash
# In your CI pipeline
gpd workflow run --file release.json --watch --watch-format json | tee workflow.log

# Parse the JSON to extract metrics or send notifications
gpd workflow run --file release.json --watch --watch-format json | \
  jq -r 'select(.type == "step_failed") | "Step failed: \(.stepName) - \(.error)"' | \
  notify-error.sh
```

## Examples

### Complete CI/CD Pipeline

```json
{
  "name": "cicd-pipeline",
  "description": "Complete CI/CD pipeline from build to production",
  "env": {
    "PACKAGE": "com.example.app"
  },
  "steps": [
    {
      "name": "validate",
      "command": "gpd automation validate --package ${env.PACKAGE} --checks all --strict"
    },
    {
      "name": "build",
      "type": "shell",
      "command": "./gradlew bundleRelease",
      "captureOutputs": ["versionCode", "versionName"]
    },
    {
      "name": "upload",
      "command": "gpd publish upload app/build/outputs/bundle/release/app-release.aab --package ${env.PACKAGE} --output json",
      "dependsOn": ["validate", "build"],
      "captureOutputs": ["versionCode", "editId"]
    },
    {
      "name": "release_internal",
      "command": "gpd publish release --package ${env.PACKAGE} --track internal --version-code ${steps.upload.versionCode}",
      "dependsOn": ["upload"]
    },
    {
      "name": "promote_beta",
      "command": "gpd automation promote --package ${env.PACKAGE} --from-track internal --to-track beta --verify",
      "dependsOn": ["release_internal"],
      "condition": "${env.PROMOTE_TO_BETA}"
    }
  ]
}
```

### Conditional Steps

```json
{
  "name": "conditional-release",
  "steps": [
    {
      "name": "check_prerelease",
      "type": "shell",
      "command": "echo ${env.IS_PRERELEASE:-false}",
      "captureOutputs": []
    },
    {
      "name": "release_production",
      "command": "gpd publish release --package ${env.PACKAGE} --track production",
      "condition": "${env.IS_PRERELEASE} != true",
      "dependsOn": ["check_prerelease"]
    },
    {
      "name": "release_beta",
      "command": "gpd publish release --package ${env.PACKAGE} --track beta",
      "condition": "${env.IS_PRERELEASE} == true",
      "dependsOn": ["check_prerelease"]
    }
  ]
}
```

### Multi-Package Workflow

See `docs/examples/workflows/multi-app-release.json` for a complete multi-package workflow example.

```json
{
  "name": "multi-package-release",
  "steps": [
    {
      "name": "upload_app1",
      "command": "gpd publish upload app1.aab --package com.example.app1 --output json",
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "upload_app2",
      "command": "gpd publish upload app2.aab --package com.example.app2 --output json",
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "release_app1",
      "command": "gpd publish release --package com.example.app1 --track internal --version-code ${steps.upload_app1.versionCode}",
      "dependsOn": ["upload_app1"]
    },
    {
      "name": "release_app2",
      "command": "gpd publish release --package com.example.app2 --track internal --version-code ${steps.upload_app2.versionCode}",
      "dependsOn": ["upload_app2"]
    }
  ]
}
```

### Staged Rollout with Monitoring

See `docs/examples/workflows/staged-rollout-monitoring.json` for a comprehensive staged rollout with full health monitoring at each stage.

```json
{
  "name": "staged-rollout",
  "description": "Start staged rollout with health monitoring",
  "env": {
    "PACKAGE": "com.example.app",
    "START_PERCENTAGE": "1",
    "TARGET_PERCENTAGE": "100"
  },
  "steps": [
    {
      "name": "upload",
      "command": "gpd publish upload app.aab --package ${env.PACKAGE} --output json",
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "create_release",
      "command": "gpd publish release --package ${env.PACKAGE} --track production --status inProgress --version-code ${steps.upload.versionCode}",
      "dependsOn": ["upload"]
    },
    {
      "name": "rollout",
      "command": "gpd automation rollout --package ${env.PACKAGE} --track production --start-percentage ${env.START_PERCENTAGE} --target-percentage ${env.TARGET_PERCENTAGE} --step-size 10 --step-interval 2h --auto-rollback",
      "dependsOn": ["create_release"],
      "timeout": "24h"
    }
  ]
}
```

## Workflow Example Files

The following production-ready workflow examples are available in `docs/examples/workflows/`:

### Basic Workflows

- **`simple-release.json`** - Simple upload and release to internal track
- **`production-release.json`** - Complete production release with validation and staged rollout
- **`cicd-pipeline.json`** - Full CI/CD pipeline from build to production

### Advanced Workflows

- **`staged-rollout-monitoring.json`** - Production rollout with automated monitoring and rollback. Implements 5% → 25% → 50% → 100% staged rollout with health monitoring at each stage and auto-rollback on issues.
  
- **`multi-track-release.json`** - Release to multiple tracks simultaneously (internal, alpha, beta) with different release notes per track. Uploads once and releases in parallel.

- **`rollback-workflow.json`** - Emergency rollback procedure. Halts current rollout, identifies previous stable version, performs rollback, and verifies success.

- **`release-with-testing.json`** - Full release with automated testing. Runs Firebase Test Lab tests and pre-launch reports, only proceeds to release if tests pass.

- **`version-bump-release.json`** - CI/CD style with version management. Reads version from Gradle, validates it's new, builds, uploads, creates GitHub release, and updates release notes from git commits.

- **`conditional-beta-release.json`** - Release to beta only if IS_BETA environment variable is set, with conditional production release.

- **`multi-app-release.json`** - Release multiple apps in parallel with cross-dependencies.

### Using Example Workflows

```bash
# Run an example workflow
gpd workflow run --file docs/examples/workflows/simple-release.json

# Run with environment variables
gpd workflow run --file docs/examples/workflows/staged-rollout-monitoring.json \
  --env PACKAGE=com.example.app \
  --env AAB_PATH=./app-release.aab \
  --env AUTO_ROLLBACK=true

# Validate an example workflow first
gpd workflow validate docs/examples/workflows/release-with-testing.json

# Dry run to preview actions
gpd workflow run --file docs/examples/workflows/rollback-workflow.json --dry-run
```

## Resume and Recovery

Workflow state is automatically persisted to `~/.gpd/workflows/runs/<run-id>.json`.

### When Runs Can Be Resumed

A run can be resumed when:
- Status is `running` or `failed`
- Not all steps completed successfully
- Not manually cancelled

### Resuming a Failed Run

```bash
# Check status
gpd workflow status <run-id>

# Resume from failed step
gpd workflow run --file workflow.json --resume <run-id>

# Force re-run all steps
gpd workflow run --file workflow.json --resume <run-id> --force
```

### Run State Storage

State is stored in:
- `~/.gpd/workflows/definitions/` - Workflow JSON files
- `~/.gpd/workflows/runs/` - Persisted run states

You can change the base directory with `--cache-dir`:

```bash
gpd workflow run --file workflow.json --cache-dir /custom/path
```

### Cleaning Up Old Runs

Run states are preserved indefinitely. To clean up:

```bash
# List all runs
gpd workflow list --all

# Remove specific run (manual deletion)
rm ~/.gpd/workflows/runs/<run-id>.json
```

## Best Practices

1. **Use descriptive step names** - Makes debugging and status checking easier
2. **Capture only necessary outputs** - Reduces state size and improves clarity
3. **Set appropriate timeouts** - Prevents stuck workflows
4. **Use conditions sparingly** - Too many conditions make workflows hard to reason about
5. **Validate before running** - Use `gpd workflow validate` to check for errors before executing (provides more detailed feedback than `--dry-run`)
6. **Use environment variables** - Keep secrets and environment-specific values out of workflow files
7. **Commit workflow files** - Store workflow definitions in version control
8. **Monitor long-running workflows** - Use `gpd workflow status` to check progress

## Troubleshooting

### "workflow file not found"
- Verify the path to the JSON file is correct
- Use absolute paths or paths relative to current directory

### "step X has unknown dependency"
- Check that all dependencies reference existing step names
- Ensure dependent steps are defined before or at the same level

### "failed to resolve ${steps.X.Y}"
- Verify the referenced step ran successfully
- Check the captured output field names match the JSON output

### "circular dependency detected"
- Review step dependencies to ensure no cycles exist
- Break cycles by removing unnecessary dependencies

### Run stuck or hanging
- Check if step has appropriate timeout set
- Use `gpd workflow status` to see which step is running
- Cancel with Ctrl+C and resume with the run ID
