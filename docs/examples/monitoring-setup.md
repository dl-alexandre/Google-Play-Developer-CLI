# Monitoring Setup Guide

This guide demonstrates how to set up continuous monitoring, anomaly detection, and automated health reporting for your Google Play apps using the `gpd` CLI.

## Table of Contents

1. [Overview](#overview)
2. [Continuous Vitals Monitoring](#continuous-vitals-monitoring)
3. [Anomaly Detection](#anomaly-detection)
4. [Monitoring Dashboards](#monitoring-dashboards)
5. [Scheduled Health Reports](#scheduled-health-reports)
6. [Alerting and Notifications](#alerting-and-notifications)
7. [Best Practices](#best-practices)

---

## Overview

The `gpd monitor` commands provide comprehensive monitoring capabilities for Android apps, including:

- **Real-time vitals monitoring**: Track crashes, ANRs, and errors continuously
- **Anomaly detection**: Automatically identify unusual patterns in metrics
- **Dashboard generation**: Export data for visualization dashboards
- **Scheduled reports**: Automated daily, weekly, or monthly health reports
- **Webhook integration**: Set up alerts and notifications

### Monitoring Command Structure

```
gpd monitor
├── watch      # Continuous vitals monitoring with threshold alerts
├── anomalies  # Detect statistical anomalies in vitals metrics
├── dashboard  # Generate monitoring dashboard data
├── report     # Generate scheduled monitoring reports
└── webhooks   # Manage monitoring webhooks
```

---

## Continuous Vitals Monitoring

The `gpd monitor watch` command provides continuous monitoring of your app's vitals with configurable thresholds and alerting.

### Basic One-Shot Monitoring

Check current vitals status:

```bash
gpd monitor watch \
  --package com.example.app \
  --metrics all
```

**Output:**
```json
{
  "data": {
    "package": "com.example.app",
    "timestamp": "2024-01-20T10:30:00Z",
    "duration": "0s",
    "pollCount": 1,
    "alerts": [],
    "thresholdsBreached": 0,
    "metrics": {
      "crashRate": 0.002,
      "anrRate": 0.001,
      "errorCount": 45
    }
  }
}
```

### Continuous Monitoring with Alerts

Monitor for 1 hour with 5-minute intervals:

```bash
gpd monitor watch \
  --package com.example.app \
  --metrics crashes anrs errors \
  --interval 5m \
  --duration 1h \
  --crash-threshold 0.01 \
  --anr-threshold 0.005 \
  --error-threshold 100 \
  --alert-on-breaches
```

**Output (with alert):**
```json
{
  "data": {
    "package": "com.example.app",
    "timestamp": "2024-01-20T10:30:00Z",
    "duration": "30m0s",
    "pollCount": 6,
    "alerts": [
      {
        "metric": "crashRate",
        "threshold": 0.01,
        "actualValue": 0.015,
        "severity": "high",
        "timestamp": "2024-01-20T10:45:00Z",
        "dimensions": {}
      }
    ],
    "thresholdsBreached": 1,
    "metrics": {
      "crashRate": 0.015,
      "anrRate": 0.002,
      "errorCount": 78
    }
  },
  "meta": {
    "warnings": ["1 threshold breaches detected"]
  }
}
```

### Monitor Specific Metrics

Focus on crash monitoring only:

```bash
gpd monitor watch \
  --package com.example.app \
  --metrics crashes \
  --interval 10m \
  --duration 2h \
  --crash-threshold 0.005 \
  --format table
```

**Output (table format):**
```
| Timestamp           | Crash Rate | Status   |
|---------------------|------------|----------|
| 2024-01-20 10:00:00 | 0.002      | Healthy  |
| 2024-01-20 10:10:00 | 0.003      | Healthy  |
| 2024-01-20 10:20:00 | 0.008      | Warning  |
```

### 24-Hour Monitoring Session

Set up extended monitoring for critical releases:

```bash
gpd monitor watch \
  --package com.example.app \
  --metrics all \
  --interval 15m \
  --duration 24h \
  --crash-threshold 0.01 \
  --anr-threshold 0.005 \
  --error-threshold 200 \
  --alert-on-breaches \
  --format json > /var/log/gpd-monitor-$(date +%Y%m%d).json
```

### Monitoring with Multiple Output Formats

```bash
# JSON for programmatic processing
gpd monitor watch \
  --package com.example.app \
  --duration 1h \
  --format json > vitals-data.json

# Table for human reading
gpd monitor watch \
  --package com.example.app \
  --duration 1h \
  --format table

# HTML for dashboards
gpd monitor watch \
  --package com.example.app \
  --duration 1h \
  --format html > vitals-dashboard.html
```

### Real-Time Monitoring Script

```bash
#!/bin/bash
# realtime-monitor.sh

PACKAGE="com.example.app"
LOG_DIR="/var/log/gpd-monitor"
mkdir -p "$LOG_DIR"

LOG_FILE="$LOG_DIR/monitor-$(date +%Y%m%d-%H%M%S).log"

# Start continuous monitoring
gpd monitor watch \
  --package "$PACKAGE" \
  --metrics all \
  --interval 5m \
  --crash-threshold 0.01 \
  --anr-threshold 0.005 \
  --alert-on-breaches \
  --format json 2>&1 | tee "$LOG_FILE" | while read -r line; do
    
    # Check for alerts in real-time
    if echo "$line" | jq -e '.data.alerts | length > 0' > /dev/null 2>&1; then
      ALERT_COUNT=$(echo "$line" | jq '.data.alerts | length')
      echo "$(date): ALERT! $ALERT_COUNT threshold(s) breached"
      
      # Send notification
      # curl -X POST ...
    fi
  done
```

---

## Anomaly Detection

The `gpd monitor anomalies` command automatically detects statistical anomalies in your app's vitals metrics.

### Basic Anomaly Detection

Detect anomalies in the last 30 days:

```bash
gpd monitor anomalies \
  --package com.example.app \
  --metrics all
```

**Output (no anomalies):**
```json
{
  "data": {
    "package": "com.example.app",
    "timestamp": "2024-01-20T10:30:00Z",
    "baselinePeriodDays": 30,
    "anomalies": [],
    "totalAnomalies": 0
  }
}
```

**Output (with anomalies):**
```json
{
  "data": {
    "package": "com.example.app",
    "timestamp": "2024-01-20T10:30:00Z",
    "baselinePeriodDays": 30,
    "anomalies": [
      {
        "metric": "crashRate",
        "severity": "high",
        "deviationPercent": 150.5,
        "currentValue": 0.015,
        "baselineAverage": 0.006,
        "timestamp": "2024-01-19T00:00:00Z"
      },
      {
        "metric": "anrRate",
        "severity": "medium",
        "deviationPercent": 85.2,
        "currentValue": 0.009,
        "baselineAverage": 0.0049,
        "timestamp": "2024-01-18T00:00:00Z"
      }
    ],
    "totalAnomalies": 2
  },
  "meta": {
    "warnings": ["2 anomalies detected in vitals metrics"]
  }
}
```

### Detect Specific Metric Anomalies

Focus on crash anomalies only:

```bash
gpd monitor anomalies \
  --package com.example.app \
  --metrics crashes \
  --baseline-period 30
```

### Adjust Detection Sensitivity

Control anomaly detection sensitivity:

```bash
# High sensitivity - detects smaller deviations
gpd monitor anomalies \
  --package com.example.app \
  --sensitivity high \
  --baseline-period 14

# Low sensitivity - only major deviations
gpd monitor anomalies \
  --package com.example.app \
  --sensitivity low \
  --baseline-period 90
```

### Custom Date Range

Detect anomalies since a specific date:

```bash
gpd monitor anomalies \
  --package com.example.app \
  --metrics all \
  --since 2024-01-01 \
  --baseline-period 30 \
  --format table
```

### Daily Anomaly Check Script

```bash
#!/bin/bash
# daily-anomaly-check.sh

PACKAGE="com.example.app"
REPORT_FILE="/tmp/anomaly-report-$(date +%Y%m%d).json"

# Run anomaly detection
gpd monitor anomalies \
  --package "$PACKAGE" \
  --metrics all \
  --baseline-period 30 \
  --format json > "$REPORT_FILE"

ANOMALY_COUNT=$(jq '.data.totalAnomalies' "$REPORT_FILE")

if [ "$ANOMALY_COUNT" -gt 0 ]; then
  echo "⚠️  $ANOMALY_COUNT anomaly(s) detected for $PACKAGE"
  
  # Extract details
  jq -r '.data.anomalies[] | "\(.metric): \(.severity) (\(.deviationPercent)% deviation)"' "$REPORT_FILE"
  
  # Send alert
  # curl -X POST -H 'Content-type: application/json' \
  #   --data "{\"text\":\"Anomaly detected in $PACKAGE: $ANOMALY_COUNT issues\"}" \
  #   "$SLACK_WEBHOOK"
else
  echo "✅ No anomalies detected for $PACKAGE"
fi
```

---

## Monitoring Dashboards

The `gpd monitor dashboard` command generates comprehensive dashboard data for visualization tools.

### Generate Basic Dashboard Data

```bash
gpd monitor dashboard \
  --package com.example.app \
  --period 7 \
  --format json
```

**Output:**
```json
{
  "data": {
    "package": "com.example.app",
    "generatedAt": "2024-01-20T10:30:00Z",
    "periodDays": 7,
    "summary": {
      "totalCrashes": 1250,
      "totalAnrs": 420,
      "totalErrors": 5432,
      "averageCrashRate": 0.008,
      "averageAnrRate": 0.003,
      "affectedUsers": 8750
    },
    "metrics": {
      "crashes": {
        "averageCrashRate": 0.008,
        "affectedUsers": 8750
      },
      "anrs": {
        "averageAnrRate": 0.003,
        "affectedUsers": 4200
      },
      "errors": {
        "totalErrors": 5432,
        "affectedUsers": 12500
      },
      "slowRendering": {
        "averageSlowRenderingRate": 0.12
      },
      "slowStart": {
        "averageSlowStartRate": 0.05
      },
      "excessiveWakeups": {
        "averageExcessiveWakeupRate": 0.02
      },
      "stuckWakelocks": {
        "averageStuckWakelockRate": 0.01
      }
    },
    "trends": {
      "crashTrend": "stable",
      "anrTrend": "stable",
      "errorTrend": "stable"
    }
  }
}
```

### Custom Metrics Selection

Include only specific metrics:

```bash
gpd monitor dashboard \
  --package com.example.app \
  --metrics crashes anrs errors \
  --period 14 \
  --format json > dashboard-core.json
```

### HTML Dashboard Output

Generate an HTML dashboard for viewing:

```bash
gpd monitor dashboard \
  --package com.example.app \
  --period 30 \
  --format html > dashboard.html

# Open in browser
open dashboard.html
```

### Markdown Report

Generate a markdown report for documentation:

```bash
gpd monitor dashboard \
  --package com.example.app \
  --period 7 \
  --format markdown > weekly-report.md
```

**Output:**
```markdown
# App Health Dashboard

**Package:** com.example.app  
**Period:** Last 7 days  
**Generated:** 2024-01-20 10:30:00

## Summary

| Metric | Value | Status |
|--------|-------|--------|
| Total Crashes | 1,250 | ⚠️ |
| Total ANRs | 420 | ✅ |
| Total Errors | 5,432 | ⚠️ |
| Avg Crash Rate | 0.8% | ✅ |
| Avg ANR Rate | 0.3% | ✅ |
| Affected Users | 8,750 | - |

## Trends

- **Crash Trend:** Stable
- **ANR Trend:** Stable
- **Error Trend:** Stable
```

### Export for External Dashboards

Export data for Grafana, Datadog, or other tools:

```bash
#!/bin/bash
# export-dashboard-data.sh

PACKAGE="com.example.app"
OUTPUT_DIR="./dashboard-exports"
mkdir -p "$OUTPUT_DIR"

# Daily export for real-time dashboards
gpd monitor dashboard \
  --package "$PACKAGE" \
  --period 1 \
  --format json > "$OUTPUT_DIR/daily-$(date +%Y%m%d).json"

# Weekly export for trend analysis
gpd monitor dashboard \
  --package "$PACKAGE" \
  --period 7 \
  --format json > "$OUTPUT_DIR/weekly-$(date +%Y%m%d).json"

# Monthly export for reporting
gpd monitor dashboard \
  --package "$PACKAGE" \
  --period 30 \
  --format json > "$OUTPUT_DIR/monthly-$(date +%Y%m%d).json"

echo "Dashboard data exported to $OUTPUT_DIR/"
```

---

## Scheduled Health Reports

The `gpd monitor report` command generates scheduled health reports (daily, weekly, monthly).

### Daily Health Report

```bash
gpd monitor report \
  --package com.example.app \
  --period daily \
  --format json
```

**Output:**
```json
{
  "data": {
    "package": "com.example.app",
    "reportType": "daily",
    "generatedAt": "2024-01-20T10:30:00Z",
    "periodStart": "2024-01-19T00:00:00Z",
    "periodEnd": "2024-01-20T00:00:00Z",
    "summary": {
      "overallHealth": "good",
      "crashRate": 0.008,
      "anrRate": 0.003,
      "errorCount": 125,
      "activeUsers": 25000,
      "issuesResolved": 3,
      "issuesOpen": 12
    },
    "keyFindings": [
      "No significant issues detected during this period"
    ],
    "recommendations": [
      "Continue monitoring - current metrics are within healthy ranges"
    ]
  }
}
```

### Weekly Report with Raw Data

```bash
gpd monitor report \
  --package com.example.app \
  --period weekly \
  --format html \
  --include-raw-data > weekly-report.html
```

### Monthly Executive Report

Generate a markdown report for stakeholders:

```bash
gpd monitor report \
  --package com.example.app \
  --period monthly \
  --format markdown > monthly-executive-report.md
```

**Sample Output:**
```markdown
# Monthly Health Report: com.example.app

**Report Period:** 2024-01-01 to 2024-01-31  
**Generated:** January 31, 2024

## Executive Summary

**Overall Health:** Good ✅

| Metric | Value | Threshold | Status |
|--------|-------|-----------|--------|
| Crash Rate | 0.8% | < 0.47% | ⚠️ |
| ANR Rate | 0.3% | < 0.47% | ✅ |
| Error Count | 3,456 | < 1,000 | ⚠️ |
| Active Users | 1.2M | - | - |

## Key Findings

1. Crash rate is above optimal threshold but within acceptable range
2. Error volume is elevated - recommend prioritizing top issues
3. User growth is strong with 1.2M active users

## Recommendations

1. Prioritize fixing top crashes - use `gpd vitals crashes` to identify patterns
2. Address error backlog of 12 open issues
3. Continue current monitoring cadence
```

### Automated Report Scheduling

Set up cron jobs for automated reports:

```bash
# Add to crontab (crontab -e)

# Daily report at 9 AM
0 9 * * * /usr/local/bin/gpd monitor report \
  --package com.example.app \
  --period daily \
  --format json > /var/reports/daily-$(date +\%Y\%m\%d).json

# Weekly report every Monday at 8 AM
0 8 * * 1 /usr/local/bin/gpd monitor report \
  --package com.example.app \
  --period weekly \
  --format html > /var/reports/weekly-$(date +\%Y\%m\%d).html

# Monthly report on 1st at 7 AM
0 7 1 * * /usr/local/bin/gpd monitor report \
  --package com.example.app \
  --period monthly \
  --format markdown | mail -s "Monthly App Health Report" team@example.com
```

### Report Distribution Script

```bash
#!/bin/bash
# generate-and-distribute-report.sh

PACKAGE="com.example.app"
REPORT_DIR="/var/reports"
DATE=$(date +%Y%m%d)

mkdir -p "$REPORT_DIR"

# Generate reports in multiple formats
gpd monitor report \
  --package "$PACKAGE" \
  --period weekly \
  --format json > "$REPORT_DIR/weekly-$DATE.json"

gpd monitor report \
  --package "$PACKAGE" \
  --period weekly \
  --format html > "$REPORT_DIR/weekly-$DATE.html"

gpd monitor report \
  --package "$PACKAGE" \
  --period weekly \
  --format markdown > "$REPORT_DIR/weekly-$DATE.md"

# Send email with HTML report
if [ -f "$REPORT_DIR/weekly-$DATE.html" ]; then
  (echo "Weekly app health report for $PACKAGE"; \
   echo; \
   cat "$REPORT_DIR/weekly-$DATE.html") | \
  mail -s "Weekly Report: $PACKAGE" \
    -a "Content-Type: text/html" \
    team@example.com
fi

echo "Reports generated and distributed"
```

---

## Alerting and Notifications

### Slack Integration

Send alerts to Slack when thresholds are breached:

```bash
#!/bin/bash
# slack-alert.sh

PACKAGE="com.example.app"
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK/TOKEN"

# Monitor and alert
gpd monitor watch \
  --package "$PACKAGE" \
  --metrics all \
  --interval 10m \
  --duration 2h \
  --crash-threshold 0.01 \
  --alert-on-breaches \
  --format json 2>&1 | while read -r line; do
    
    # Check for alerts
    if echo "$line" | jq -e '.data.alerts | length > 0' > /dev/null 2>&1; then
      ALERTS=$(echo "$line" | jq -r '.data.alerts[] | "• \(.metric): \(.actualValue) (threshold: \(.threshold))"' | tr '\n' '\\n')
      
      curl -X POST -H 'Content-type: application/json' \
        --data "{\"text\":\"🚨 *App Health Alert*\\nPackage: $PACKAGE\\n\\n$ALERTS\"}" \
        "$SLACK_WEBHOOK"
    fi
  done
```

### Discord Webhook

Send notifications to Discord:

```bash
#!/bin/bash
# discord-notification.sh

PACKAGE="com.example.app"
DISCORD_WEBHOOK="https://discord.com/api/webhooks/YOUR/WEBHOOK/TOKEN"

# Get anomaly count
ANOMALIES=$(gpd monitor anomalies \
  --package "$PACKAGE" \
  --metrics all \
  --format json | jq '.data.totalAnomalies')

if [ "$ANOMALIES" -gt 0 ]; then
  curl -X POST -H 'Content-type: application/json' \
    --data "{\"content\":\"📊 Anomaly Alert\\n**Package:** $PACKAGE\\n**Anomalies:** $ANOMALIES detected\"}" \
    "$DISCORD_WEBHOOK"
fi
```

### PagerDuty Integration

Trigger incidents for critical issues:

```bash
#!/bin/bash
# pagerduty-trigger.sh

PACKAGE="com.example.app"
PAGERDUTY_KEY="your-integration-key"

# Check crash rate
CRASH_DATA=$(gpd vitals crashes \
  --package "$PACKAGE" \
  --start-date $(date -v-1d +%Y-%m-%d) \
  --end-date $(date +%Y-%m-%d) \
  --format json)

CRASH_RATE=$(echo "$CRASH_DATA" | jq -r '.data.rows[0].crashRate // 0')

if (( $(echo "$CRASH_RATE > 0.02" | bc -l) )); then
  curl -X POST \
    -H "Content-Type: application/json" \
    -H "Authorization: Token token=$PAGERDUTY_KEY" \
    --data '{
      "incident": {
        "type": "incident",
        "title": "High Crash Rate: '"$PACKAGE"'",
        "service": {"id": "your-service-id", "type": "service_reference"},
        "urgency": "high",
        "body": {
          "type": "incident_body",
          "details": "Crash rate is '"$CRASH_RATE"' (threshold: 0.02)"
        }
      }
    }' \
    https://api.pagerduty.com/incidents
fi
```

---

## Best Practices

### 1. Set Realistic Thresholds

```bash
# Conservative production thresholds
--crash-threshold 0.005 \
--anr-threshold 0.003 \
--error-threshold 50

# More relaxed thresholds for development
--crash-threshold 0.02 \
--anr-threshold 0.01 \
--error-threshold 200
```

### 2. Use Appropriate Baseline Periods

```bash
# Short period for recent releases
--baseline-period 7

# Standard period for normal operations
--baseline-period 30

# Long period for seasonal analysis
--baseline-period 90
```

### 3. Combine Monitoring Approaches

```bash
#!/bin/bash
# comprehensive-monitoring.sh

PACKAGE="com.example.app"

# 1. Real-time monitoring
gpd monitor watch \
  --package "$PACKAGE" \
  --duration 2h \
  --interval 5m \
  --alert-on-breaches

# 2. Anomaly detection
gpd monitor anomalies \
  --package "$PACKAGE" \
  --baseline-period 30 \
  --sensitivity medium

# 3. Generate dashboard data
gpd monitor dashboard \
  --package "$PACKAGE" \
  --period 7 \
  --format json > dashboard-data.json

# 4. Create weekly report
gpd monitor report \
  --package "$PACKAGE" \
  --period weekly \
  --format html > weekly-report.html
```

### 4. Automate Report Distribution

```bash
#!/bin/bash
# automated-reporting.sh

PACKAGES=("com.example.app1" "com.example.app2" "com.example.app3")

for PACKAGE in "${PACKAGES[@]}"; do
  echo "Generating reports for $PACKAGE..."
  
  # Daily dashboard
  gpd monitor dashboard \
    --package "$PACKAGE" \
    --period 1 \
    --format json > "/var/dashboards/$PACKAGE-$(date +%Y%m%d).json"
  
  # Weekly report
  if [ "$(date +%u)" = "1" ]; then
    gpd monitor report \
      --package "$PACKAGE" \
      --period weekly \
      --format html > "/var/reports/$PACKAGE-weekly-$(date +%Y%m%d).html"
  fi
done
```

### 5. Version Control Your Monitoring Config

```yaml
# monitoring-config.yaml
monitoring:
  package: com.example.app
  thresholds:
    crash_rate: 0.01
    anr_rate: 0.005
    error_count: 100
  schedules:
    watch:
      interval: 5m
      duration: 2h
    anomaly_check:
      baseline_period: 30
      sensitivity: medium
    reports:
      daily: "0 9 * * *"
      weekly: "0 8 * * 1"
```

### 6. Monitor Multiple Apps

```bash
#!/bin/bash
# multi-app-monitoring.sh

PACKAGES=(
  "com.company.app1"
  "com.company.app2"
  "com.company.app3"
)

for PACKAGE in "${PACKAGES[@]}"; do
  echo "Monitoring $PACKAGE..."
  
  # Background monitoring for each app
  gpd monitor watch \
    --package "$PACKAGE" \
    --duration 1h \
    --interval 10m \
    --crash-threshold 0.01 \
    --alert-on-breaches \
    --format json > "/var/log/$PACKAGE-$(date +%Y%m%d).json" &
done

# Wait for all background jobs
wait
echo "Multi-app monitoring complete"
```

---

## Summary

| Monitoring Feature | Command | Example Count |
|-------------------|---------|---------------|
| Continuous Monitoring | `gpd monitor watch` | 7 examples |
| Anomaly Detection | `gpd monitor anomalies` | 6 examples |
| Dashboard Generation | `gpd monitor dashboard` | 7 examples |
| Scheduled Reports | `gpd monitor report` | 7 examples |
| Alerting Setup | Webhook integration | 4 examples |
| **Total Examples** | | **31 examples** |

---

## Related Commands

- [`gpd vitals crashes`](./error-debugging.md) - Detailed crash analysis
- [`gpd vitals anrs`](./error-debugging.md) - ANR rate monitoring
- [`gpd vitals errors`](./error-debugging.md) - Error report search
- [`gpd automation monitor`](./automation-workflows.md) - Post-release monitoring
- [`gpd compare vitals`](./multi-app-comparison.md) - Compare vitals across apps
