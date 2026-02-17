# Error Debugging with Google Play Developer CLI

This guide demonstrates how to use `gpd` for comprehensive error monitoring, debugging, and app quality analysis using the Android Vitals API.

## Overview

The Google Play Developer CLI provides powerful tools for monitoring app quality through error reporting and vitals metrics. These features help you:

- **Identify issues quickly**: Search and filter error reports to find problematic patterns
- **Monitor app health**: Track crash rates, ANR rates, and performance metrics over time
- **Detect anomalies**: Automatically identify unusual patterns in your app's behavior
- **Debug effectively**: Access detailed error reports with stack traces and device information

### Prerequisites

Before using vitals commands, ensure you have:

1. **Authentication configured**: Set up service account credentials
   ```bash
   gpd auth status
   ```

2. **Package name set**: Either via `--package` flag or configuration
   ```bash
   gpd config set defaultPackage com.example.app
   ```

3. **Required permissions**: Your service account needs access to the Play Developer Reporting API

## Error Search Commands

### Finding Error Issues

The `gpd vitals errors issues search` command helps you discover and categorize error issues in your app.

#### Basic Usage

```bash
# Search for all error issues in the last 30 days
gpd vitals errors issues search --package com.example.app --interval last30Days

# Search for specific error types
gpd vitals errors issues search --package com.example.app \
  --query "NullPointerException" \
  --interval last7Days

# Get more results per page
gpd vitals errors issues search --package com.example.app \
  --interval last30Days \
  --page-size 100
```

#### Command Parameters

| Parameter | Description | Default | Options |
|-----------|-------------|---------|---------|
| `--query` | Search query filter | - | Any text query |
| `--interval` | Time interval | `last30Days` | `last7Days`, `last30Days`, `last90Days` |
| `--page-size` | Results per page | 50 | 1-1000 |
| `--page-token` | Pagination token | - | From previous response |

#### Example Output

```json
{
  "data": {
    "query": "NullPointerException",
    "interval": "last30Days",
    "package": "com.example.app",
    "issues": [
      {
        "name": "apps/com.example.app/errorIssues/123456789",
        "errorIssue": {
          "issueId": "123456789",
          "type": "CRASH",
          "severity": "HIGH",
          "firstErrorReportTime": "2024-01-15T10:30:00Z",
          "lastErrorReportTime": "2024-01-20T14:22:00Z",
          "errorCount": 1250,
          "affectedUsers": 450
        }
      }
    ],
    "rowCount": 1,
    "nextPageToken": "CAESFwo..."
  }
}
```

#### Common Query Patterns

```bash
# Find crashes only
gpd vitals errors issues search --package com.example.app \
  --query "type:CRASH" \
  --interval last30Days

# Find high-severity issues
gpd vitals errors issues search --package com.example.app \
  --query "severity:HIGH" \
  --interval last30Days

# Find issues affecting many users
gpd vitals errors issues search --package com.example.app \
  --query "affectedUsers>100" \
  --interval last30Days
```

### Getting Detailed Error Reports

The `gpd vitals errors reports search` command provides detailed error reports with stack traces and device information.

#### Basic Usage

```bash
# Get all error reports from the last 7 days
gpd vitals errors reports search --package com.example.app \
  --interval last7Days

# Search for specific error patterns
gpd vitals errors reports search --package com.example.app \
  --query "OutOfMemoryError" \
  --interval last30Days

# Format reports for readability (deobfuscate)
gpd vitals errors reports search --package com.example.app \
  --query "crash" \
  --interval last7Days \
  --deobfuscate
```

#### Command Parameters

| Parameter | Description | Default | Options |
|-----------|-------------|---------|---------|
| `--query` | Search query filter | - | Any text query |
| `--interval` | Time interval | `last30Days` | `last7Days`, `last30Days`, `last90Days` |
| `--page-size` | Results per page | 50 | 1-1000 |
| `--page-token` | Pagination token | - | From previous response |
| `--deobfuscate` | Format report text for readability | false | true/false |

#### Example Output

```json
{
  "data": {
    "query": "crash",
    "interval": "last7Days",
    "package": "com.example.app",
    "reports": [
      {
        "name": "apps/com.example.app/errorReports/abc123",
        "errorReport": {
          "reportTime": "2024-01-20T14:22:00Z",
          "issueId": "123456789",
          "deviceModel": "Pixel 7",
          "osVersion": "Android 14",
          "appVersion": "1.2.3",
          "reportText": "java.lang.NullPointerException\n  at com.example.app.MainActivity.onCreate(MainActivity.java:42)\n  ..."
        }
      }
    ],
    "rowCount": 1,
    "nextPageToken": "CAESFwo..."
  }
}
```

### Error Statistics

#### Get Error Counts Summary

The `gpd vitals errors counts get` command provides a summary of error counts.

```bash
gpd vitals errors counts get --package com.example.app
```

**Example Output:**

```json
{
  "data": {
    "package": "com.example.app",
    "counts": {
      "distinctUsers": 1250,
      "errorCount": 5432,
      "distinctIssues": 15
    }
  }
}
```

#### Query Error Counts Over Time

The `gpd vitals errors counts query` command allows you to track error counts over time with optional dimension grouping.

```bash
# Query error counts for a date range
gpd vitals errors counts query --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31

# Group by app version
gpd vitals errors counts query --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --dimensions appVersion

# Group by multiple dimensions
gpd vitals errors counts query --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --dimensions appVersion,device,country
```

#### Command Parameters

| Parameter | Description | Required | Options |
|-----------|-------------|----------|---------|
| `--start-date` | Start date (ISO 8601: YYYY-MM-DD) | Yes | Any valid date |
| `--end-date` | End date (ISO 8601: YYYY-MM-DD) | Yes | Any valid date |
| `--dimensions` | Grouping dimensions | No | `appVersion`, `device`, `country`, `androidVersion` |
| `--page-size` | Results per page | No | Default: 100 |
| `--page-token` | Pagination token | No | From previous response |

## Vitals Metrics

Beyond error reporting, `gpd` provides access to comprehensive Android Vitals metrics that help monitor overall app health.

### Crash Rate Analysis

Monitor crash rates to identify stability issues:

```bash
# Get crash rate for a date range
gpd vitals crashes --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31

# Group by app version to compare releases
gpd vitals crashes --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --dimensions appVersion

# Group by device to identify problematic devices
gpd vitals crashes --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --dimensions device

# Fetch all pages automatically
gpd vitals crashes --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --all
```

**Understanding Crash Rate Thresholds:**

- **Good**: < 0.47% crash rate
- **Bad**: 0.47% - 1.09% crash rate
- **Excessive**: > 1.09% crash rate

### ANR Rate Monitoring

Application Not Responding (ANR) rates indicate when your app becomes unresponsive:

```bash
# Get ANR rate data
gpd vitals anrs --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31

# Compare ANR rates across Android versions
gpd vitals anrs --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --dimensions androidVersion
```

**Understanding ANR Rate Thresholds:**

- **Good**: < 0.47% ANR rate
- **Bad**: 0.47% - 1.09% ANR rate
- **Excessive**: > 1.09% ANR rate

### Excessive Wakeups

Monitor apps that wake up the device too frequently, which drains battery:

```bash
gpd vitals excessive-wakeups --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31
```

### Slow Rendering

Track slow rendering rates to identify UI performance issues:

```bash
gpd vitals slow-rendering --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --dimensions device
```

### Slow Start Times

Monitor app startup performance:

```bash
gpd vitals slow-start --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31
```

### Stuck Wakelocks

Identify wakelock issues that prevent devices from sleeping:

```bash
gpd vitals stuck-wakelocks --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31
```

### Generic Query Command

Use `gpd vitals query` to query multiple metrics at once:

```bash
# Query crash rate (default)
gpd vitals query --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --metrics crashRate

# Query multiple metrics
gpd vitals query --package com.example.app \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --metrics crashRate,anrRate,slowRendering
```

### Available Dimensions

All vitals metrics support grouping by these dimensions:

- `appVersion`: App version code
- `device`: Device model
- `country`: Country code (ISO 3166-1 alpha-2)
- `androidVersion`: Android OS version

### Viewing Capabilities

To see all available metrics and dimensions:

```bash
gpd vitals capabilities --package com.example.app
```

## Anomalies Detection

The Play Developer Reporting API can automatically detect anomalies in your vitals metrics, helping you identify unusual patterns without manual analysis.

### Listing Anomalies

```bash
# List all anomalies from the last 30 days
gpd vitals anomalies list --package com.example.app \
  --time-period last30Days

# Filter by specific metric
gpd vitals anomalies list --package com.example.app \
  --metric crashRate \
  --time-period last30Days

# Get anomalies from the last 7 days
gpd vitals anomalies list --package com.example.app \
  --time-period last7Days

# Get anomalies from the last 90 days
gpd vitals anomalies list --package com.example.app \
  --time-period last90Days
```

#### Command Parameters

| Parameter | Description | Default | Options |
|-----------|-------------|---------|---------|
| `--metric` | Filter by metric name | - | `crashRate`, `anrRate`, etc. |
| `--time-period` | Time period | `last30Days` | `last7Days`, `last30Days`, `last90Days` |
| `--page-size` | Results per page | 20 | 1-100 |
| `--page-token` | Pagination token | - | From previous response |

### Understanding Anomaly Data

Anomaly objects contain:

- **Metric information**: Which metric triggered the anomaly
- **Timeline**: When the anomaly was detected
- **Severity**: How significant the anomaly is
- **Dimension values**: Which app version, device, country, etc. is affected

**Example Output:**

```json
{
  "data": {
    "anomalies": [
      {
        "name": "apps/com.example.app/anomalies/abc123",
        "metric": {
          "metric": "crashRate",
          "metricSet": "apps/com.example.app/crashRateMetricSet"
        },
        "dimensions": {
          "appVersion": "123",
          "device": "Pixel 7"
        },
        "startTime": "2024-01-15T00:00:00Z",
        "endTime": "2024-01-20T23:59:59Z",
        "severity": "HIGH"
      }
    ],
    "nextPageToken": "CAESFwo..."
  }
}
```

## Common Workflows

### Daily Crash Monitoring Script

Create a script to monitor crashes daily and alert on spikes:

```bash
#!/bin/bash
# daily-crash-monitor.sh

PACKAGE="com.example.app"
YESTERDAY=$(date -v-1d +%Y-%m-%d 2>/dev/null || date -d "yesterday" +%Y-%m-%d)
TODAY=$(date +%Y-%m-%d)

# Get crash rate for yesterday
CRASH_DATA=$(gpd vitals crashes --package "$PACKAGE" \
  --start-date "$YESTERDAY" \
  --end-date "$TODAY" \
  --output json)

# Extract crash rate (adjust based on your JSON parsing tool)
CRASH_RATE=$(echo "$CRASH_DATA" | jq -r '.data.rows[0].crashRate // 0')

# Threshold: 1.09% (excessive)
THRESHOLD=1.09

if (( $(echo "$CRASH_RATE > $THRESHOLD" | bc -l) )); then
  echo "ALERT: Crash rate is $CRASH_RATE%, exceeding threshold of $THRESHOLD%"
  # Send notification (Slack, email, etc.)
  exit 1
else
  echo "OK: Crash rate is $CRASH_RATE%"
  exit 0
fi
```

### Investigating a Spike in Crashes

When you notice a spike in crashes, follow this workflow:

```bash
#!/bin/bash
# investigate-crash-spike.sh

PACKAGE="com.example.app"
START_DATE="2024-01-15"
END_DATE="2024-01-20"

echo "1. Getting crash rate over time..."
gpd vitals crashes --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --output json > crash-timeline.json

echo "2. Identifying top error issues..."
gpd vitals errors issues search --package "$PACKAGE" \
  --interval last7Days \
  --page-size 10 \
  --output json > top-issues.json

echo "3. Getting detailed reports for top issue..."
TOP_ISSUE_ID=$(jq -r '.data.issues[0].errorIssue.issueId' top-issues.json)
gpd vitals errors reports search --package "$PACKAGE" \
  --query "issueId:$TOP_ISSUE_ID" \
  --interval last7Days \
  --deobfuscate \
  --output json > detailed-reports.json

echo "4. Checking if issue is version-specific..."
gpd vitals crashes --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --dimensions appVersion \
  --output json > crashes-by-version.json

echo "Investigation complete. Check the generated JSON files."
```

### Comparing Metrics Across Versions

Compare app performance between versions:

```bash
#!/bin/bash
# compare-versions.sh

PACKAGE="com.example.app"
START_DATE="2024-01-01"
END_DATE="2024-01-31"

echo "Comparing versions for date range: $START_DATE to $END_DATE"

echo "Crash rates by version:"
gpd vitals crashes --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --dimensions appVersion \
  --output table

echo -e "\nANR rates by version:"
gpd vitals anrs --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --dimensions appVersion \
  --output table

echo -e "\nSlow rendering by version:"
gpd vitals slow-rendering --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --dimensions appVersion \
  --output table
```

### Weekly Health Report

Generate a comprehensive weekly health report:

```bash
#!/bin/bash
# weekly-health-report.sh

PACKAGE="com.example.app"
LAST_WEEK=$(date -v-7d +%Y-%m-%d 2>/dev/null || date -d "7 days ago" +%Y-%m-%d)
TODAY=$(date +%Y-%m-%d)

echo "=== Weekly Health Report for $PACKAGE ==="
echo "Period: $LAST_WEEK to $TODAY"
echo ""

echo "--- Error Summary ---"
gpd vitals errors counts get --package "$PACKAGE" --output table

echo -e "\n--- Crash Rate ---"
gpd vitals crashes --package "$PACKAGE" \
  --start-date "$LAST_WEEK" \
  --end-date "$TODAY" \
  --output table

echo -e "\n--- ANR Rate ---"
gpd vitals anrs --package "$PACKAGE" \
  --start-date "$LAST_WEEK" \
  --end-date "$TODAY" \
  --output table

echo -e "\n--- Top 5 Error Issues ---"
gpd vitals errors issues search --package "$PACKAGE" \
  --interval last7Days \
  --page-size 5 \
  --output table

echo -e "\n--- Detected Anomalies ---"
gpd vitals anomalies list --package "$PACKAGE" \
  --time-period last7Days \
  --output table
```

## Integration Ideas

### Slack Notifications

Send crash alerts to Slack:

```bash
#!/bin/bash
# slack-crash-alert.sh

PACKAGE="com.example.app"
SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Get crash rate
CRASH_DATA=$(gpd vitals crashes --package "$PACKAGE" \
  --start-date "$(date -v-1d +%Y-%m-%d)" \
  --end-date "$(date +%Y-%m-%d)" \
  --output json)

CRASH_RATE=$(echo "$CRASH_DATA" | jq -r '.data.rows[0].crashRate // 0')

if (( $(echo "$CRASH_RATE > 1.09" | bc -l) )); then
  MESSAGE="🚨 *High Crash Rate Alert*\nPackage: $PACKAGE\nCrash Rate: ${CRASH_RATE}%\nThreshold: 1.09%"
  
  curl -X POST -H 'Content-type: application/json' \
    --data "{\"text\":\"$MESSAGE\"}" \
    "$SLACK_WEBHOOK_URL"
fi
```

### Discord Webhook Integration

Send metrics to Discord:

```bash
#!/bin/bash
# discord-metrics.sh

PACKAGE="com.example.app"
DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/YOUR/WEBHOOK"

# Get error summary
ERROR_SUMMARY=$(gpd vitals errors counts get --package "$PACKAGE" --output json)

ERROR_COUNT=$(echo "$ERROR_SUMMARY" | jq -r '.data.counts.errorCount')
DISTINCT_USERS=$(echo "$ERROR_SUMMARY" | jq -r '.data.counts.distinctUsers')

MESSAGE="📊 *Daily Error Summary*\n**Package:** $PACKAGE\n**Error Count:** $ERROR_COUNT\n**Affected Users:** $DISTINCT_USERS"

curl -X POST -H 'Content-type: application/json' \
  --data "{\"content\":\"$MESSAGE\"}" \
  "$DISCORD_WEBHOOK_URL"
```

### Dashboard Integration

Export data for dashboard visualization:

```bash
#!/bin/bash
# export-dashboard-data.sh

PACKAGE="com.example.app"
OUTPUT_DIR="./dashboard-data"
START_DATE="2024-01-01"
END_DATE="2024-01-31"

mkdir -p "$OUTPUT_DIR"

# Export crash data
gpd vitals crashes --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --dimensions appVersion \
  --output json > "$OUTPUT_DIR/crashes.json"

# Export ANR data
gpd vitals anrs --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --dimensions appVersion \
  --output json > "$OUTPUT_DIR/anrs.json"

# Export error counts
gpd vitals errors counts query --package "$PACKAGE" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --dimensions appVersion \
  --output json > "$OUTPUT_DIR/error-counts.json"

echo "Data exported to $OUTPUT_DIR/"
```

### Automated Alerting with Cron

Set up automated monitoring with cron:

```bash
# Add to crontab (crontab -e)
# Run daily crash check at 9 AM
0 9 * * * /path/to/daily-crash-monitor.sh >> /var/log/gpd-monitor.log 2>&1

# Run weekly health report every Monday at 8 AM
0 8 * * 1 /path/to/weekly-health-report.sh | mail -s "Weekly Health Report" team@example.com
```

### CI/CD Integration

Add vitals checks to your CI/CD pipeline:

```yaml
# .github/workflows/vitals-check.yml
name: Vitals Check

on:
  schedule:
    - cron: '0 9 * * *'  # Daily at 9 AM
  workflow_dispatch:

jobs:
  check-vitals:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Check crash rate
        run: |
          gpd vitals crashes \
            --package ${{ secrets.PACKAGE_NAME }} \
            --start-date $(date -d "yesterday" +%Y-%m-%d) \
            --end-date $(date +%Y-%m-%d) \
            --output json > crash-data.json
          
          CRASH_RATE=$(jq -r '.data.rows[0].crashRate // 0' crash-data.json)
          
          if (( $(echo "$CRASH_RATE > 1.09" | bc -l) )); then
            echo "::error::Crash rate ($CRASH_RATE%) exceeds threshold (1.09%)"
            exit 1
          fi
        env:
          GPD_SERVICE_ACCOUNT_KEY: ${{ secrets.GPD_SERVICE_ACCOUNT_KEY }}
```

## Tips and Best Practices

1. **Data Freshness**: Vitals data may be delayed by 24-48 hours. Keep this in mind when monitoring recent releases.

2. **Pagination**: Use `--all` flag or handle `nextPageToken` when fetching large datasets.

3. **Dimension Grouping**: Use dimensions strategically to identify patterns:
   - `appVersion`: Compare releases
   - `device`: Identify device-specific issues
   - `country`: Regional performance differences
   - `androidVersion`: OS version compatibility

4. **Query Optimization**: Use specific date ranges and queries to reduce API calls and response times.

5. **Error Filtering**: Combine error search with vitals metrics to get a complete picture of app health.

6. **Automation**: Set up automated monitoring for critical metrics to catch issues early.

7. **Historical Comparison**: Always compare current metrics against historical baselines to identify trends.

## Additional Resources

- [Google Play Developer Reporting API Documentation](https://developers.google.com/play/developer/reporting)
- [Android Vitals Best Practices](https://developer.android.com/topic/performance/vitals)
- [gpd CLI Documentation](../README.md)

## Troubleshooting

### No Data Available

If you're not seeing data:

1. **Check data freshness**: Vitals data is delayed by 24-48 hours
2. **Verify date range**: Ensure your date range includes days with sufficient data
3. **Check app version**: Ensure the app version has been released and has users
4. **Verify permissions**: Ensure your service account has access to the Play Developer Reporting API

### Authentication Issues

```bash
# Check authentication status
gpd auth status

# Verify package access
gpd auth check --package com.example.app
```

### Rate Limiting

If you encounter rate limits:

1. Reduce the frequency of API calls
2. Use pagination tokens instead of fetching all pages at once
3. Cache results when possible
4. Use `--page-size` to optimize requests
