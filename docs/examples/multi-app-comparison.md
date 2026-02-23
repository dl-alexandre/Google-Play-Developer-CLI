# Multi-App Comparison Guide

This guide demonstrates how to compare metrics across multiple Google Play apps, enabling competitive analysis, portfolio management, and cross-app health assessment.

## Table of Contents

1. [Overview](#overview)
2. [Vitals Comparison](#vitals-comparison)
3. [Reviews Comparison](#reviews-comparison)
4. [Release Timeline Comparison](#release-timeline-comparison)
5. [Subscription Metrics Comparison](#subscription-metrics-comparison)
6. [Best Practices](#best-practices)

---

## Overview

The `gpd compare` commands allow you to analyze and compare metrics across multiple apps in your portfolio. This is valuable for:

- **Portfolio Management**: Compare health across your app ecosystem
- **Competitive Analysis**: Benchmark against competitor apps
- **Release Planning**: Compare release cadence and timing
- **User Sentiment**: Analyze review sentiment across apps
- **Revenue Tracking**: Compare subscription and monetization metrics

### Comparison Command Structure

```
gpd compare
├── vitals         # Compare vitals metrics across apps
├── reviews        # Compare review metrics across apps
├── releases       # Compare release history across apps
└── subscriptions  # Compare subscription metrics
```

---

## Vitals Comparison

Compare crash rates, ANR rates, and error metrics across multiple apps.

### Basic Vitals Comparison

Compare vitals for 2-5 apps:

```bash
gpd compare vitals \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --metric all \
  --format table
```

**Output (table format):**
```
| Package           | Crash Rate | ANR Rate | Error Count | Score | Rank |
|-------------------|------------|----------|-------------|-------|------|
| com.company.app2  | 0.001      | 0.0005   | 125         | 98.5  | 1    |
| com.company.app1  | 0.003      | 0.001    | 340         | 92.0  | 2    |
| com.company.app3  | 0.008      | 0.003    | 890         | 78.5  | 3    |
```

**Output (JSON format):**
```json
{
  "data": {
    "metric": "all",
    "period": "last-30-days",
    "apps": [
      {
        "package": "com.company.app2",
        "crashRate": 0.001,
        "anrRate": 0.0005,
        "errorCount": 125,
        "score": 98.5,
        "rank": 1
      },
      {
        "package": "com.company.app1",
        "crashRate": 0.003,
        "anrRate": 0.001,
        "errorCount": 340,
        "score": 92.0,
        "rank": 2
      },
      {
        "package": "com.company.app3",
        "crashRate": 0.008,
        "anrRate": 0.003,
        "errorCount": 890,
        "score": 78.5,
        "rank": 3
      }
    ],
    "bestApp": "com.company.app2",
    "worstApp": "com.company.app3",
    "comparisonAt": "2024-01-20T10:30:00Z"
  }
}
```

### Compare Specific Metrics

Focus on crash rates only:

```bash
gpd compare vitals \
  --packages com.company.app1 com.company.app2 \
  --metric crash-rate \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --format csv > crash-comparison.csv
```

### Date Range Comparison

Compare vitals over a specific period:

```bash
gpd compare vitals \
  --packages com.company.app1 com.company.app2 com.company.app3 com.company.app4 \
  --metric all \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --format json
```

### CSV Export for Analysis

Export to CSV for spreadsheet analysis:

```bash
gpd compare vitals \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --metric all \
  --format csv > vitals-comparison-$(date +%Y%m%d).csv
```

**CSV Output:**
```csv
package,crash_rate,anr_rate,error_count,score,rank
com.company.app2,0.001,0.0005,125,98.5,1
com.company.app1,0.003,0.001,340,92.0,2
com.company.app3,0.008,0.003,890,78.5,3
```

### Weekly Health Comparison Script

```bash
#!/bin/bash
# weekly-vitals-comparison.sh

PACKAGES=(
  "com.company.app1"
  "com.company.app2"
  "com.company.app3"
)

START_DATE=$(date -v-7d +%Y-%m-%d 2>/dev/null || date -d "7 days ago" +%Y-%m-%d)
END_DATE=$(date +%Y-%m-%d)

echo "=== Weekly Vitals Comparison ==="
echo "Period: $START_DATE to $END_DATE"
echo ""

gpd compare vitals \
  --packages "${PACKAGES[@]}" \
  --metric all \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --format table

echo ""
echo "=== Ranking ==="
gpd compare vitals \
  --packages "${PACKAGES[@]}" \
  --metric all \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --format json | jq -r '.data.apps[] | "\(.rank). \(.package) (Score: \(.score))"'
```

### Portfolio Health Dashboard

```bash
#!/bin/bash
# portfolio-dashboard.sh

PORTFOLIO_APPS=(
  "com.company.game1"
  "com.company.game2"
  "com.company.productivity1"
  "com.company.productivity2"
  "com.company.social1"
)

echo "Generating Portfolio Health Report..."

# Generate JSON for dashboard
gpd compare vitals \
  --packages "${PORTFOLIO_APPS[@]}" \
  --metric all \
  --format json > portfolio-health.json

# Generate CSV for executives
gpd compare vitals \
  --packages "${PORTFOLIO_APPS[@]}" \
  --metric all \
  --format csv > portfolio-health.csv

# Display summary
echo ""
echo "Portfolio Summary:"
echo "-----------------"
jq -r '.data.apps[] | "\(.package): Score \(.score)/100, Rank #\(.rank)"' portfolio-health.json

echo ""
echo "Best Performing: $(jq -r '.data.bestApp' portfolio-health.json)"
echo "Needs Attention: $(jq -r '.data.worstApp' portfolio-health.json)"
```

---

## Reviews Comparison

Compare review metrics, ratings, and sentiment across multiple apps.

### Basic Reviews Comparison

Compare review metrics for multiple apps:

```bash
gpd compare reviews \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --format table
```

**Output:**
```
| Package           | Avg Rating | Total Reviews | 5★   | 4★   | 3★   | 2★   | 1★   |
|-------------------|------------|---------------|------|------|------|------|------|
| com.company.app1  | 4.5        | 12,450        | 8,712| 2,490| 623  | 249  | 376  |
| com.company.app2  | 4.2        | 8,320         | 4,992| 2,080| 832  | 416  | 0    |
| com.company.app3  | 3.8        | 5,670         | 1,701| 1,418| 1,418| 567  | 567  |
```

**JSON Output:**
```json
{
  "data": {
    "period": "all-time",
    "apps": [
      {
        "package": "com.company.app1",
        "averageRating": 4.5,
        "totalReviews": 12450,
        "ratingsDistribution": {
          "1": 376,
          "2": 249,
          "3": 623,
          "4": 2490,
          "5": 8712
        }
      },
      {
        "package": "com.company.app2",
        "averageRating": 4.2,
        "totalReviews": 8320,
        "ratingsDistribution": {
          "1": 0,
          "2": 0,
          "3": 832,
          "4": 2080,
          "5": 4992
        }
      },
      {
        "package": "com.company.app3",
        "averageRating": 3.8,
        "totalReviews": 5670,
        "ratingsDistribution": {
          "1": 567,
          "2": 567,
          "3": 1418,
          "4": 1418,
          "5": 1701
        }
      }
    ],
    "comparisonAt": "2024-01-20T10:30:00Z"
  }
}
```

### Compare with Sentiment Analysis

Include sentiment scores in the comparison:

```bash
gpd compare reviews \
  --packages com.company.app1 com.company.app2 \
  --include-sentiment \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --format json
```

**Output:**
```json
{
  "data": {
    "period": "2024-01-01 to 2024-01-31",
    "apps": [
      {
        "package": "com.company.app1",
        "averageRating": 4.5,
        "totalReviews": 1250,
        "ratingsDistribution": {
          "1": 38,
          "2": 25,
          "3": 63,
          "4": 250,
          "5": 875
        },
        "sentimentScore": 0.87
      },
      {
        "package": "com.company.app2",
        "averageRating": 4.2,
        "totalReviews": 832,
        "ratingsDistribution": {
          "1": 0,
          "2": 0,
          "3": 83,
          "4": 208,
          "5": 541
        },
        "sentimentScore": 0.82
      }
    ]
  }
}
```

### Time-Range Comparison

Compare reviews from a specific period:

```bash
gpd compare reviews \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --start-date 2024-01-01 \
  --end-date 2024-01-31 \
  --format table
```

### CSV Export for Analysis

```bash
gpd compare reviews \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --include-sentiment \
  --format csv > reviews-comparison.csv
```

### Monthly Reviews Report

```bash
#!/bin/bash
# monthly-reviews-report.sh

PACKAGES=(
  "com.company.app1"
  "com.company.app2"
  "com.company.app3"
)

START_DATE=$(date -v-30d +%Y-%m-%d 2>/dev/null || date -d "30 days ago" +%Y-%m-%d)
END_DATE=$(date +%Y-%m-%d)

echo "=== Monthly Reviews Comparison ==="
echo "Period: $START_DATE to $END_DATE"
echo ""

gpd compare reviews \
  --packages "${PACKAGES[@]}" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --include-sentiment \
  --format table

echo ""
echo "Detailed JSON data:"
gpd compare reviews \
  --packages "${PACKAGES[@]}" \
  --start-date "$START_DATE" \
  --end-date "$END_DATE" \
  --include-sentiment \
  --format json > "reviews-comparison-$(date +%Y%m).json"
```

---

## Release Timeline Comparison

Compare release cadence, timing, and version history across apps.

### Compare Release History

Compare release history for multiple apps:

```bash
gpd compare releases \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --track production \
  --limit 5 \
  --format table
```

**Output:**
```
| Package           | Release Count | Latest Version | Latest Date |
|-------------------|---------------|----------------|-------------|
| com.company.app1  | 12            | 1.5.3          | 2024-01-15  |
| com.company.app2  | 8             | 2.1.0          | 2024-01-18  |
| com.company.app3  | 15            | 3.2.1          | 2024-01-20  |
```

**JSON Output:**
```json
{
  "data": {
    "track": "production",
    "apps": [
      {
        "package": "com.company.app1",
        "releaseCount": 12,
        "latestVersion": "1.5.3",
        "latestDate": "2024-01-15",
        "releases": [
          {
            "versionCodes": ["1048576"],
            "status": "completed",
            "date": "2024-01-15",
            "name": "Version 1.5.3"
          }
        ]
      },
      {
        "package": "com.company.app2",
        "releaseCount": 8,
        "latestVersion": "2.1.0",
        "latestDate": "2024-01-18",
        "releases": []
      },
      {
        "package": "com.company.app3",
        "releaseCount": 15,
        "latestVersion": "3.2.1",
        "latestDate": "2024-01-20",
        "releases": []
      }
    ],
    "timeline": [
      {
        "date": "2024-01-20",
        "package": "com.company.app3",
        "release": "3.2.1",
        "type": "production"
      },
      {
        "date": "2024-01-18",
        "package": "com.company.app2",
        "release": "2.1.0",
        "type": "production"
      }
    ],
    "comparisonAt": "2024-01-20T10:30:00Z"
  }
}
```

### Compare Beta Track Releases

```bash
gpd compare releases \
  --packages com.company.app1 com.company.app2 \
  --track beta \
  --since 2024-01-01 \
  --limit 10 \
  --format json
```

### Release Timeline Visualization

Generate data for timeline charts:

```bash
gpd compare releases \
  --packages com.company.app1 com.company.app2 com.company.app3 com.company.app4 \
  --track production \
  --since 2024-01-01 \
  --format json > release-timeline-data.json

# View timeline
echo "Release Timeline (2024):"
jq -r '.data.timeline[] | "\(.date): \(.package) v\(.release)"' release-timeline-data.json
```

### Release Cadence Analysis

```bash
#!/bin/bash
# release-cadence-analysis.sh

PACKAGES=(
  "com.company.app1"
  "com.company.app2"
  "com.company.app3"
)

START_DATE="2024-01-01"

echo "=== Release Cadence Analysis ==="
echo "Since: $START_DATE"
echo ""

# Get comparison data
DATA=$(gpd compare releases \
  --packages "${PACKAGES[@]}" \
  --track production \
  --since "$START_DATE" \
  --format json)

# Calculate release frequency
DAYS_SINCE=$(echo "$(date +%s) - $(date -d "$START_DATE" +%s)" | bc)
DAYS_SINCE=$((DAYS_SINCE / 86400))

echo "Release Frequency (production, since $START_DATE):"
echo "------------------------------------------------"

for PACKAGE in "${PACKAGES[@]}"; do
  COUNT=$(echo "$DATA" | jq -r ".data.apps[] | select(.package==\"$PACKAGE\") | .releaseCount")
  if [ "$COUNT" -gt 0 ] && [ "$DAYS_SINCE" -gt 0 ]; then
    FREQUENCY=$(echo "scale=2; $DAYS_SINCE / $COUNT" | bc)
    echo "$PACKAGE: $COUNT releases (every ${FREQUENCY} days)"
  else
    echo "$PACKAGE: $COUNT releases"
  fi
done
```

---

## Subscription Metrics Comparison

Compare subscription metrics across apps with monetization features.

### Basic Subscription Comparison

```bash
gpd compare subscriptions \
  --packages com.company.app1 com.company.app2 \
  --period 30d \
  --format table
```

**Output:**
```
| Package           | Total Subs | Active Subs | Churn Rate | MRR    | ARPU  |
|-------------------|------------|-------------|------------|--------|-------|
| com.company.app1  | 5,230      | 4,450       | 4.8%       | $8,900 | $2.00 |
| com.company.app2  | 3,120      | 2,650       | 6.2%       | $5,300 | $2.00 |
```

**JSON Output:**
```json
{
  "data": {
    "period": "30d",
    "apps": [
      {
        "package": "com.company.app1",
        "totalSubscriptions": 5230,
        "activeSubscriptions": 4450,
        "churnRate": 0.048,
        "mrr": 8900.0,
        "arpu": 2.0
      },
      {
        "package": "com.company.app2",
        "totalSubscriptions": 3120,
        "activeSubscriptions": 2650,
        "churnRate": 0.062,
        "mrr": 5300.0,
        "arpu": 2.0
      }
    ],
    "comparisonAt": "2024-01-20T10:30:00Z"
  }
}
```

### Compare Specific Subscriptions

Focus on specific subscription products:

```bash
gpd compare subscriptions \
  --packages com.company.app1 com.company.app2 \
  --subscriptions premium_subscription basic_subscription \
  --period 30d \
  --format json
```

### Time Period Comparison

Compare over different time periods:

```bash
# 7-day comparison
gpd compare subscriptions \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --period 7d \
  --format table

# 90-day comparison
gpd compare subscriptions \
  --packages com.company.app1 com.company.app2 com.company.app3 \
  --period 90d \
  --format json > quarterly-subscription-data.json
```

### CSV Export for Financial Analysis

```bash
gpd compare subscriptions \
  --packages com.company.app1 com.company.app2 \
  --period 30d \
  --format csv > subscription-metrics.csv
```

### Revenue Dashboard Script

```bash
#!/bin/bash
# revenue-dashboard.sh

PACKAGES=(
  "com.company.app1"
  "com.company.app2"
  "com.company.app3"
)

echo "=== Subscription Revenue Dashboard ==="
echo "Generated: $(date)"
echo ""

gpd compare subscriptions \
  --packages "${PACKAGES[@]}" \
  --period 30d \
  --format table

echo ""
echo "Total Portfolio Metrics:"
gpd compare subscriptions \
  --packages "${PACKAGES[@]}" \
  --period 30d \
  --format json | jq -r '
    .data.apps | 
    {
      totalSubs: map(.totalSubscriptions) | add,
      totalActive: map(.activeSubscriptions) | add,
      totalMrr: map(.mrr) | add,
      avgChurn: (map(.churnRate) | add / length)
    } |
    "Total Subscriptions: \(.totalSubs)" +
    "\nTotal Active: \(.totalActive)" +
    "\nTotal MRR: $\(.totalMrr | round)" +
    "\nAverage Churn Rate: \(.avgChurn * 100 | round)%"
  '
```

---

## Best Practices

### 1. Regular Portfolio Reviews

```bash
#!/bin/bash
# weekly-portfolio-review.sh

PORTFOLIO=(
  "com.company.app1"
  "com.company.app2"
  "com.company.app3"
  "com.company.app4"
  "com.company.app5"
)

REPORT_FILE="/tmp/portfolio-review-$(date +%Y%m%d).md"

echo "# Weekly Portfolio Review" > "$REPORT_FILE"
echo "Date: $(date)" >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "## Vitals Comparison" >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"
gpd compare vitals \
  --packages "${PORTFOLIO[@]}" \
  --metric all \
  --start-date "$(date -v-7d +%Y-%m-%d)" \
  --end-date "$(date +%Y-%m-%d)" \
  --format table >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

echo "## Reviews Comparison" >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"
gpd compare reviews \
  --packages "${PORTFOLIO[@]}" \
  --start-date "$(date -v-7d +%Y-%m-%d)" \
  --end-date "$(date +%Y-%m-%d)" \
  --format table >> "$REPORT_FILE"
echo '```' >> "$REPORT_FILE"

cat "$REPORT_FILE"
```

### 2. Benchmarking Against Competitors

```bash
#!/bin/bash
# competitive-analysis.sh

# Your apps
YOUR_APPS=("com.yourcompany.app1" "com.yourcompany.app2")

# Competitor apps (public data only)
COMPETITOR_APPS=("com.competitor.app1" "com.competitor.app2")

ALL_APPS=("${YOUR_APPS[@]}" "${COMPETITOR_APPS[@]}")

echo "=== Competitive Vitals Analysis ==="
gpd compare vitals \
  --packages "${ALL_APPS[@]}" \
  --metric all \
  --format table

echo ""
echo "=== Competitive Reviews Analysis ==="
gpd compare reviews \
  --packages "${ALL_APPS[@]}" \
  --include-sentiment \
  --format table
```

### 3. Multi-Format Reporting

```bash
#!/bin/bash
# multi-format-comparison.sh

PACKAGES=("com.company.app1" "com.company.app2" "com.company.app3")
DATE=$(date +%Y%m%d)
OUTPUT_DIR="./comparison-reports"
mkdir -p "$OUTPUT_DIR"

# Generate all formats
for FORMAT in json table csv; do
  gpd compare vitals \
    --packages "${PACKAGES[@]}" \
    --metric all \
    --format "$FORMAT" > "$OUTPUT_DIR/vitals-$FORMAT-$DATE.$FORMAT"
done

# Create markdown report
gpd compare vitals \
  --packages "${PACKAGES[@]}" \
  --metric all \
  --format table > "$OUTPUT_DIR/vitals-comparison-$DATE.md"

echo "Reports generated in $OUTPUT_DIR/"
ls -la "$OUTPUT_DIR/"
```

### 4. Trend Analysis Over Time

```bash
#!/bin/bash
# trend-analysis.sh

PACKAGES=("com.company.app1" "com.company.app2")
OUTPUT_DIR="./trend-data"
mkdir -p "$OUTPUT_DIR"

# Collect weekly comparisons for 4 weeks
for WEEKS_AGO in 0 1 2 3; do
  END_DATE=$(date -v-${WEEKS_AGO}w +%Y-%m-%d 2>/dev/null || date -d "${WEEKS_AGO} weeks ago" +%Y-%m-%d)
  START_DATE=$(date -v-$((WEEKS_AGO+1))w +%Y-%m-%d 2>/dev/null || date -d "$((WEEKS_AGO+1)) weeks ago" +%Y-%m-%d)
  
  gpd compare vitals \
    --packages "${PACKAGES[@]}" \
    --metric all \
    --start-date "$START_DATE" \
    --end-date "$END_DATE" \
    --format json > "$OUTPUT_DIR/vitals-week-$((WEEKS_AGO+1)).json"
done

echo "Trend data collected. Generate charts from JSON files in $OUTPUT_DIR/"
```

### 5. Automated Comparison Alerts

```bash
#!/bin/bash
# comparison-alerts.sh

PACKAGES=("com.company.app1" "com.company.app2" "com.company.app3")
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/WEBHOOK"

# Get comparison data
RESULT=$(gpd compare vitals \
  --packages "${PACKAGES[@]}" \
  --metric all \
  --format json)

# Check for underperforming apps
WORST_APP=$(echo "$RESULT" | jq -r '.data.worstApp')
WORST_SCORE=$(echo "$RESULT" | jq -r '.data.apps[] | select(.package=="'"$WORST_APP"'") | .score')

if (( $(echo "$WORST_SCORE < 70" | bc -l) )); then
  MESSAGE="⚠️ App performance alert: $WORST_APP has low health score ($WORST_SCORE/100)"
  
  curl -X POST -H 'Content-type: application/json' \
    --data "{\"text\":\"$MESSAGE\"}" \
    "$SLACK_WEBHOOK"
fi
```

---

## Summary

| Comparison Feature | Command | Example Count |
|-------------------|---------|---------------|
| Vitals Comparison | `gpd compare vitals` | 7 examples |
| Reviews Comparison | `gpd compare reviews` | 6 examples |
| Releases Comparison | `gpd compare releases` | 5 examples |
| Subscriptions Comparison | `gpd compare subscriptions` | 6 examples |
| **Total Examples** | | **24 examples** |

---

## Related Commands

- [`gpd monitor dashboard`](./monitoring-setup.md) - Generate monitoring dashboards
- [`gpd monitor report`](./monitoring-setup.md) - Scheduled health reports
- [`gpd vitals crashes`](./error-debugging.md) - Detailed crash analysis
- [`gpd reviews list`](./error-debugging.md) - Review management
- [`gpd monetization subscriptions`](./subscription-management.md) - Subscription details
