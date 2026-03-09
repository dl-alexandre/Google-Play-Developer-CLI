# API Drift Detection & Orchestration

This document describes the comprehensive API drift detection and orchestration system integrated into the Google Play Developer CLI (gpd).

## Overview

The API drift detection system automatically monitors the Google Play Developer API (and other Google APIs) for changes between the published discovery document and the Go client library implementation. When drift is detected, the system can:

1. Alert via GitHub issues
2. Generate detailed reports
3. Automatically create PRs with dependency updates
4. Track drift trends over time

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                      API Orchestration                         │
│  (Workflow: .github/workflows/api-orchestration.yml)          │
├─────────────────────────────────────────────────────────────────┤
│  Phase 1: Health Check                                          │
│  Phase 2: Dependency Analysis                                  │
│  Phase 3: Drift Detection (Multi-API)                         │
│  Phase 4: Report Generation                                    │
│  Phase 5: Issue/PR Creation (if drift detected)               │
│  Phase 6: Auto-Update (optional)                              │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Drift Detection Engine                        │
│         (internal/apidrift/detector.go)                        │
│  • Fetches discovery documents                                  │
│  • Parses API schemas and endpoints                            │
│  • Compares with Go client library                             │
│  • Generates drift reports                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     CLI Integration                              │
│           (cmd/gpd maintenance drift)                           │
│  • Single API drift detection                                  │
│  • Multi-API monitoring                                          │
│  • Health checks                                                  │
│  • Various output formats (text, JSON, markdown)              │
└─────────────────────────────────────────────────────────────────┘
```

## Components

### 1. Drift Detection Engine

**Location:** `internal/apidrift/detector.go`

The core drift detection engine:
- Fetches Google API Discovery documents
- Parses REST API schemas and endpoints
- Compares discovery API with Go client library
- Calculates drift scores

**Key Types:**

```go
// DiscoveryDocument represents the full discovery API response
type DiscoveryDocument struct {
    Revision    string              `json:"revision"`
    Resources   map[string]Resource `json:"resources"`
    Schemas     map[string]Schema   `json:"schemas"`
    // ...
}

// DriftReport contains detection results
type DriftReport struct {
    Timestamp         time.Time
    DiscoveryRevision string
    GoModVersion      string
    DriftScore        int
    DriftDetected     bool
    Endpoints         EndpointAnalysis
    Schemas           SchemaAnalysis
}
```

### 2. CLI Commands

**Location:** `internal/cli/kong_maintenance.go`

#### Single API Drift Detection

```bash
# Basic drift check
gpd maintenance drift

# JSON output
gpd maintenance drift --format json

# Markdown report to file
gpd maintenance drift --format markdown --output drift-report.md

# Fail if drift score exceeds threshold (for CI/CD)
gpd maintenance drift --threshold 10

# Check specific API
gpd maintenance drift --discovery-url https://www.googleapis.com/discovery/v1/apis/drive/v3/rest
```

#### Multi-API Monitoring

```bash
# Check multiple APIs (table output)
gpd maintenance multi-drift --apis androidpublisher,drive,gmail

# All supported APIs (JSON)
gpd maintenance multi-drift --apis androidpublisher,drive,gmail,calendar,sheets,docs,slides,people,tasks,youtube,analytics,bigquery,storage,compute --format json

# Save reports to directory
gpd maintenance multi-drift --output-dir ./drift-reports
```

#### System Health Check

```bash
gpd maintenance health
gpd maintenance health --check-api --check-auth --check-config
```

### 3. Standalone Tool

**Location:** `cmd/apidrift/main.go`

A standalone binary for CI/CD pipelines:

```bash
# Build
go build -o bin/apidrift ./cmd/apidrift

# Run drift detection
./bin/apidrift

# Options
./bin/apidrift -help
./bin/apidrift -verbose
./bin/apidrift -format json -output report.json
./bin/apidrift -threshold 5  # Fail if drift > 5
```

### 4. GitHub Actions Workflows

#### API Drift Workflow

**Location:** `.github/workflows/api-drift.yml`

Runs on:
- Schedule: Weekly (Mondays 9 AM UTC)
- Manual trigger via `workflow_dispatch`

Features:
- Monitors 14 Google APIs
- Creates GitHub issues when drift detected
- Generates artifact reports
- Configurable threshold for CI failure

#### API Orchestration Workflow

**Location:** `.github/workflows/api-orchestration.yml`

A comprehensive 6-phase orchestration:

1. **Health Check** - System health validation
2. **Dependency Analysis** - Check for available updates
3. **Drift Detection** - Multi-API drift analysis
4. **Report Generation** - Comprehensive report
5. **Ticket Creation** - Auto-create GitHub issues
6. **Auto-Update** - Optional automatic PR creation

**Modes:**
- `full` - Run all phases
- `drift-only` - Just drift detection
- `health-only` - Just health check
- `deps-only` - Just dependency analysis

## Supported APIs

The system can monitor these Google APIs:

| API | Priority | Discovery URL |
|-----|----------|---------------|
| androidpublisher | Critical | `androidpublisher/v3` |
| drive | High | `drive/v3` |
| storage | Medium | `storage/v1` |
| gmail | Low | `gmail/v1` |
| calendar | Low | `calendar/v3` |
| sheets | Low | `sheets/v4` |
| docs | Low | `docs/v1` |
| slides | Low | `slides/v1` |
| people | Low | `people/v1` |
| tasks | Low | `tasks/v1` |
| youtube | Low | `youtube/v3` |
| analytics | Low | `analytics/v3` |
| bigquery | Low | `bigquery/v2` |
| compute | Low | `compute/v1` |

## Drift Score Calculation

The drift score is calculated as:

```
Drift Score = (Missing Endpoints) + (Deprecated Endpoints)
```

**Thresholds:**
- 0: No drift ✅
- 1-5: Minor drift (informational) ℹ️
- 6-20: Moderate drift (should update) ⚠️
- 21+: Significant drift (critical) 🚨

## CI/CD Integration

### GitHub Actions

```yaml
- name: Check API Drift
  run: |
    go build -o gpd ./cmd/gpd
    ./gpd maintenance drift --threshold 10
```

### Makefile Targets

```bash
# Quick drift check
make apidrift

# Check with threshold
make apidrift-check

# Multi-API monitoring
make apidrift-multi

# All formats
make apidrift-json
make apidrift-markdown
```

## Report Formats

### Text Format (Default)

```
=============================================================
Google Play Developer API Drift Detection Report
=============================================================
Timestamp: 2026-03-09T23:06:59Z

Discovery URL:      https://www.googleapis.com/discovery/v1/apis/androidpublisher/v3/rest
Discovery Revision: 20260309
Go Module Version:  0.270.0

Endpoints:
  Discovery Total:   136
  Implemented Total: 66
  Missing in Client: 76
  Deprecated:        6

Drift Score: 82
Status: ⚠️ DRIFT DETECTED
```

### JSON Format

```json
{
  "timestamp": "2026-03-09T23:06:59Z",
  "discovery_url": "https://www.googleapis.com/discovery/v1/apis/androidpublisher/v3/rest",
  "discovery_revision": "20260309",
  "go_mod_version": "0.270.0",
  "drift_detected": true,
  "drift_score": 82,
  "endpoints": {
    "discovery_total": 136,
    "implemented_total": 66,
    "missing_in_client": ["edits.bundles.upload", ...],
    "deprecated": ["old.endpoint", ...]
  }
}
```

### Markdown Format

```markdown
# Google Play Developer API Drift Report

**Generated:** 2026-03-09 23:06:59 UTC

## Summary

- **Discovery Revision:** 20260309
- **Go Module Version:** 0.270.0
- **Drift Score:** 82
- **Status:** ⚠️ DRIFT DETECTED

## Endpoint Analysis

| Metric | Count |
|--------|-------|
| Discovery Total | 136 |
| Implemented | 66 |
| Missing | 76 |
| Deprecated | 6 |
```

## Automation Features

### Auto-Issue Creation

When drift is detected, the workflow automatically creates a GitHub issue with:
- Drift score and revision
- Missing endpoints list
- Action items checklist
- Link to workflow run

### Auto-PR Creation (Optional)

With `auto_update: true`, the workflow will:
1. Update `google.golang.org/api` in `go.mod`
2. Run `go mod tidy`
3. Execute tests
4. Create a PR with detailed description

## Exit Codes

The drift detection tool uses these exit codes:

- `0` - Success (no drift or within threshold)
- `1` - Drift detected
- `2` - Drift score exceeded threshold

## Best Practices

1. **Run Weekly** - Schedule drift detection to run weekly
2. **Review Reports** - Check generated reports for trends
3. **Update Promptly** - Address drift within 2 weeks of detection
4. **Monitor Trends** - Track drift score over time
5. **CI Integration** - Add to CI pipeline for PR checks

## Troubleshooting

### Common Issues

**High Drift Score**
- This is expected when Google releases new API features
- Update the Go client library: `go get -u google.golang.org/api`
- Re-run drift detection to verify

**Discovery API Unavailable**
- Check network connectivity
- Verify discovery URL is correct
- The discovery API is rarely down (99.9% uptime)

**Client Source Not Found**
- The detector attempts to find the Go client library in:
  1. Specified path (`--client-dir`)
  2. GOPATH/pkg/mod
  3. Local `internal/api` directory

## Related Documentation

- [CONTRIBUTING.md](./CONTRIBUTING.md) - Development setup
- [docs/api-coverage-matrix.md](./docs/api-coverage-matrix.md) - API coverage
- [Workflows](./.github/workflows/) - All CI/CD workflows

## Future Enhancements

- [ ] gRPC API drift detection
- [ ] Automatic endpoint stub generation
- [ ] Drift trend visualization
- [ ] Slack/Discord notifications
- [ ] Prometheus metrics export
