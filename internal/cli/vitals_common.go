package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

const (
	timePeriodLast7Days  = "last7Days"
	timePeriodLast30Days = "last30Days"
	timePeriodLast90Days = "last90Days"
)

func parseYear(date string) int64 {
	parts := strings.Split(date, "-")
	if len(parts) >= 1 {
		y, _ := strconv.ParseInt(parts[0], 10, 64)
		return y
	}
	return 0
}

func parseMonth(date string) int64 {
	parts := strings.Split(date, "-")
	if len(parts) >= 2 {
		m, _ := strconv.ParseInt(parts[1], 10, 64)
		return m
	}
	return 0
}

func parseDay(date string) int64 {
	parts := strings.Split(date, "-")
	if len(parts) >= 3 {
		d, _ := strconv.ParseInt(parts[2], 10, 64)
		return d
	}
	return 0
}

func formatReportText(report string) string {
	lines := strings.Split(report, "\n")
	var formatted []string
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		formatted = append(formatted, fmt.Sprintf("%d: %s", i+1, trimmed))
	}
	if len(formatted) == 0 {
		return report
	}
	return "  " + strings.Join(formatted, "\n  ")
}

func (c *CLI) vitalsQuery(ctx context.Context, startDate, endDate string, metrics, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	var allResults []map[string]interface{}

	for _, metric := range metrics {
		switch metric {
		case "crashRate":
			err := c.vitalsCrashes(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil
		case "anrRate":
			err := c.vitalsANRs(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil
		case "excessiveWakeups":
			err := c.vitalsExcessiveWakeups(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil
		case "lmkRate":
			err := c.vitalsLmkRate(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil
		case "slowRendering":
			err := c.vitalsSlowRendering(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil
		case "slowStart":
			err := c.vitalsSlowStart(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil
		case "stuckWakelocks":
			err := c.vitalsStuckWakelocks(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil
		default:
			allResults = append(allResults, map[string]interface{}{
				"metric": metric,
				"status": "unsupported",
				"hint":   "Use 'gpd vitals capabilities' to see supported metrics",
			})
		}
	}

	result := output.NewResult(map[string]interface{}{
		"metrics":    metrics,
		"startDate":  startDate,
		"endDate":    endDate,
		"dimensions": dimensions,
		"package":    c.packageName,
		"results":    allResults,
		"dataFreshness": map[string]interface{}{
			"note": "Vitals data may be delayed by 24-48 hours",
		},
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsCapabilities(_ context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"metrics": []map[string]interface{}{
			{"name": "crashRate", "description": "Crash rate per 1000 sessions", "unit": "percentage"},
			{"name": "anrRate", "description": "ANR rate per 1000 sessions", "unit": "percentage"},
			{"name": "userPerceivedCrashRate", "description": "User-perceived crash rate", "unit": "percentage"},
			{"name": "userPerceivedAnrRate", "description": "User-perceived ANR rate", "unit": "percentage"},
			{"name": "excessiveWakeups", "description": "Excessive wakeups", "unit": "count"},
			{"name": "stuckWakeLocks", "description": "Stuck wake locks", "unit": "count"},
			{"name": "lmkRate", "description": "Low memory kill rate", "unit": "percentage"},
			{"name": "slowRendering", "description": "Slow rendering rate", "unit": "percentage"},
			{"name": "slowStart", "description": "Slow start rate", "unit": "percentage"},
		},
		"dimensions": []map[string]interface{}{
			{"name": "country", "description": "Country code"},
			{"name": "device", "description": "Device model"},
			{"name": "androidVersion", "description": "Android OS version"},
			{"name": "appVersion", "description": "App version code"},
		},
		"granularities":   []string{"daily"},
		"maxLookbackDays": 28,
		"dataFreshness": map[string]interface{}{
			"typical": "24-48 hours",
			"note":    "Vitals data freshness depends on user opt-in and reporting",
		},
		"thresholds": map[string]interface{}{
			"crashRateBad":       1.09,
			"crashRateExcessive": 8.0,
			"anrRateBad":         0.47,
			"anrRateExcessive":   4.0,
		},
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}
