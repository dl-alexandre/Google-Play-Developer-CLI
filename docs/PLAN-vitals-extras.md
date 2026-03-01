# Vitals Extras Implementation Plan

**Status**: Not started
**Priority**: Low
**Date**: 2026-03-01

## Background

Two generic vitals commands are stubs: `vitals query` and `vitals capabilities`. The specific vitals commands (crashes, ANRs, errors, anomalies) are already implemented — these two provide generic/discovery access.

Stubs in `internal/cli/kong_vitals.go`.

## Commands to Implement

### 1. `vitals query` (line 1294)

**Purpose**: Run arbitrary vitals queries beyond the predefined crash/ANR/error commands.

- `--metric-set` — which metric set to query (e.g., `crashRateMetricSet`, `anrRateMetricSet`, `stuckBackgroundWakelockRateMetricSet`, `slowRenderingMetricSet`, etc.)
- `--dimensions`, `--metrics`, `--filters` — query parameters
- `--start-date`, `--end-date` — date range
- `--timeline-spec` — aggregation period
- Uses `PlayReporting.Vitals.{MetricSet}.Query()`
- Generalizes what the specific vitals commands do

### 2. `vitals capabilities` (line 1302)

**Purpose**: List available vitals metric sets, dimensions, and freshness info.

- `--metric-set` — optional, show capabilities for a specific set
- Without flag, list all available metric sets with their dimensions and metrics
- Show data freshness information per metric set
- Informational, no heavy API call

## Notes

- These are low priority because the specific commands (`vitals crashes`, `vitals anr-rate`, etc.) cover the most common use cases
- `vitals query` is useful for metrics not yet wrapped in dedicated commands (e.g., slow rendering, wakelock)
- Also fix `--format csv` output for all vitals commands while here (currently accepted but not implemented)

## Files to Modify

| File | Changes |
|------|---------|
| `internal/cli/kong_vitals.go` | Implement `Run()` for 2 commands |

## Testing

- Mock PlayReporting vitals query responses
- Test with various metric sets
- Test capabilities output format
