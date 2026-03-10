package cli

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

const (
	metricCrash                       = "crash"
	metricAnr                         = "anr"
	metricError                       = "error"
	metricCrashes                     = "crashes"
	metricAnrs                        = "anrs"
	metricErrors                      = "errors"
	metricCrashRate                   = "crashRate"
	metricAnrRate                     = "anrRate"
	metricErrorCount                  = "errorCount"
	metricSlowRenderingRate           = "slowRenderingRate"
	metricSlowStartRate               = "slowStartRate"
	metricExcessiveWakeupRate         = "excessiveWakeupRate"
	metricStuckBackgroundWakelockRate = "stuckBackgroundWakelockRate"
	severityHigh                      = "high"
	severityLow                       = "low"
	metricDistinctUsers               = "distinctUsers"
	trendStable                       = "stable"
)

// MonitorCmd contains monitoring and alerting commands.
type MonitorCmd struct {
	Watch     MonitorWatchCmd     `cmd:"" help:"Continuous vitals monitoring with threshold alerts"`
	Anomalies MonitorAnomaliesCmd `cmd:"" help:"Detect statistical anomalies in vitals metrics"`
	Dashboard MonitorDashboardCmd `cmd:"" help:"Generate monitoring dashboard data"`
	Report    MonitorReportCmd    `cmd:"" help:"Generate scheduled monitoring reports"`
	Webhooks  MonitorWebhooksCmd  `cmd:"" help:"Manage monitoring webhooks (simulated - no Play API)"`
}

// MonitorWatchCmd continuously monitors vitals and alerts on threshold breaches.
type MonitorWatchCmd struct {
	Metrics         []string      `help:"Metrics to monitor: crashes, anrs, errors, all" default:"all"`
	Interval        time.Duration `help:"Poll interval for continuous monitoring" default:"5m"`
	Duration        time.Duration `help:"Total monitoring duration (0 = one-shot)" default:"0"`
	CrashThreshold  float64       `help:"Crash rate threshold for alerting (0-1)" default:"0.01"`
	AnrThreshold    float64       `help:"ANR rate threshold for alerting (0-1)" default:"0.005"`
	ErrorThreshold  float64       `help:"Error count threshold for alerting" default:"100"`
	AlertOnBreaches bool          `help:"Exit with error code when thresholds breached"`
	Format          string        `help:"Output format: json, table, html" default:"json" enum:"json,table,html"`
}

// monitorAlert represents a threshold breach alert.
type monitorAlert struct {
	Metric      string            `json:"metric"`
	Threshold   float64           `json:"threshold"`
	ActualValue float64           `json:"actualValue"`
	Severity    string            `json:"severity"`
	Timestamp   time.Time         `json:"timestamp"`
	Dimensions  map[string]string `json:"dimensions,omitempty"`
}

// monitorWatchResult represents the monitoring result.
type monitorWatchResult struct {
	Package            string                 `json:"package"`
	Timestamp          time.Time              `json:"timestamp"`
	Duration           time.Duration          `json:"duration"`
	PollCount          int                    `json:"pollCount"`
	Alerts             []monitorAlert         `json:"alerts"`
	ThresholdsBreached int                    `json:"thresholdsBreached"`
	Metrics            map[string]interface{} `json:"metrics"`
}

// Run executes the watch command.
func (cmd *MonitorWatchCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	authMgr := newAuthManager()

	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource,
		api.WithTimeout(globals.Timeout),
		api.WithVerboseLogging(globals.Verbose))
	if err != nil {
		return errors.NewAPIError(errors.CodeAuthFailure, fmt.Sprintf("failed to create API client: %v", err))
	}

	svc, err := client.PlayReporting()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get reporting service: %v", err))
	}

	// Normalize metrics
	metrics := cmd.normalizeMetrics()

	startTime := time.Now()
	result := &monitorWatchResult{
		Package:   globals.Package,
		Timestamp: startTime,
		Metrics:   make(map[string]interface{}),
		Alerts:    []monitorAlert{},
	}

	// Determine if continuous or one-shot
	isContinuous := cmd.Duration > 0
	ticker := time.NewTicker(cmd.Interval)
	defer ticker.Stop()

	done := make(chan bool)
	if isContinuous {
		go func() {
			time.Sleep(cmd.Duration)
			done <- true
		}()
	}

	pollCount := 0
	for {
		pollCount++

		// Poll all requested metrics
		for _, metric := range metrics {
			if err := cmd.pollMetric(ctx, client, svc, globals.Package, metric, result); err != nil {
				return err
			}
		}

		result.PollCount = pollCount
		result.Duration = time.Since(startTime)

		// If one-shot, break after first poll
		if !isContinuous {
			break
		}

		// Wait for next poll or timeout
		select {
		case <-ticker.C:
			continue
		case <-done:
			goto finish
		case <-ctx.Done():
			goto finish
		}
	}

finish:
	// Generate output
	outputResult := output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if result.ThresholdsBreached > 0 && cmd.AlertOnBreaches {
		outputResult = outputResult.WithWarnings(
			fmt.Sprintf("%d threshold breaches detected", result.ThresholdsBreached),
		)
	}

	return outputResultResult(outputResult, cmd.Format, globals.Pretty)
}

func (cmd *MonitorWatchCmd) normalizeMetrics() []string {
	if len(cmd.Metrics) == 0 {
		return []string{metricCrashes, metricAnrs, metricErrors}
	}

	var result []string
	for _, m := range cmd.Metrics {
		switch strings.ToLower(m) {
		case checkAll:
			return []string{metricCrashes, metricAnrs, metricErrors}
		case metricCrashes, metricCrash:
			result = append(result, metricCrashes)
		case metricAnrs, metricAnr:
			result = append(result, metricAnrs)
		case metricErrors, metricError:
			result = append(result, metricErrors)
		}
	}
	return result
}

func (cmd *MonitorWatchCmd) pollMetric(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg, metric string, result *monitorWatchResult) error {
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -1)

	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(startDate.Year()),
			Month: int64(startDate.Month()),
			Day:   int64(startDate.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(endDate.Year()),
			Month: int64(endDate.Month()),
			Day:   int64(endDate.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
	}

	switch metric {
	case metricCrashes:
		return cmd.checkCrashThreshold(ctx, client, svc, pkg, timelineSpec, result)
	case metricAnrs:
		return cmd.checkAnrThreshold(ctx, client, svc, pkg, timelineSpec, result)
	case metricErrors:
		return cmd.checkErrorThreshold(ctx, client, svc, pkg, timelineSpec, result)
	}
	return nil
}

func (cmd *MonitorWatchCmd) checkCrashThreshold(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec, result *monitorWatchResult) error {
	name := fmt.Sprintf("apps/%s/crashRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{metricCrashRate, "userPerceivedCrashRate", metricDistinctUsers},
		PageSize:     1,
	}

	var crashRate float64
	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Crashrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		if len(resp.Rows) > 0 && len(resp.Rows[0].Metrics) > 0 {
			for _, m := range resp.Rows[0].Metrics {
				if m.Metric == metricCrashRate {
					crashRate = parseDecimalValue(m)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	result.Metrics[metricCrashRate] = crashRate

	if crashRate > cmd.CrashThreshold {
		alert := monitorAlert{
			Metric:      metricCrashRate,
			Threshold:   cmd.CrashThreshold,
			ActualValue: crashRate,
			Severity:    cmd.calculateSeverity(crashRate, cmd.CrashThreshold),
			Timestamp:   time.Now(),
		}
		result.Alerts = append(result.Alerts, alert)
		result.ThresholdsBreached++
	}

	return nil
}

func (cmd *MonitorWatchCmd) checkAnrThreshold(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec, result *monitorWatchResult) error {
	name := fmt.Sprintf("apps/%s/anrRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{metricAnrRate, "userPerceivedAnrRate", metricDistinctUsers},
		PageSize:     1,
	}

	var anrRate float64
	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Anrrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		if len(resp.Rows) > 0 && len(resp.Rows[0].Metrics) > 0 {
			for _, m := range resp.Rows[0].Metrics {
				if m.Metric == metricAnrRate {
					anrRate = parseDecimalValue(m)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	result.Metrics[metricAnrRate] = anrRate

	if anrRate > cmd.AnrThreshold {
		alert := monitorAlert{
			Metric:      metricAnrRate,
			Threshold:   cmd.AnrThreshold,
			ActualValue: anrRate,
			Severity:    cmd.calculateSeverity(anrRate, cmd.AnrThreshold),
			Timestamp:   time.Now(),
		}
		result.Alerts = append(result.Alerts, alert)
		result.ThresholdsBreached++
	}

	return nil
}

func (cmd *MonitorWatchCmd) checkErrorThreshold(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec, result *monitorWatchResult) error {
	name := fmt.Sprintf("apps/%s/errorCountMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{metricErrorCount, metricDistinctUsers},
		PageSize:     1,
	}

	var errorCount float64
	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Errors.Counts.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		if len(resp.Rows) > 0 && len(resp.Rows[0].Metrics) > 0 {
			for _, m := range resp.Rows[0].Metrics {
				if m.Metric == metricErrorCount {
					errorCount = parseDecimalValue(m)
					break
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	result.Metrics[metricErrorCount] = errorCount

	if float64(errorCount) > cmd.ErrorThreshold {
		alert := monitorAlert{
			Metric:      metricErrorCount,
			Threshold:   cmd.ErrorThreshold,
			ActualValue: float64(errorCount),
			Severity:    cmd.calculateSeverity(float64(errorCount), cmd.ErrorThreshold),
			Timestamp:   time.Now(),
		}
		result.Alerts = append(result.Alerts, alert)
		result.ThresholdsBreached++
	}

	return nil
}

func (cmd *MonitorWatchCmd) calculateSeverity(actual, threshold float64) string {
	ratio := actual / threshold
	switch {
	case ratio >= 3.0:
		return "critical"
	case ratio >= 2.0:
		return severityHigh
	case ratio >= 1.5:
		return "medium"
	default:
		return severityLow
	}
}

// MonitorAnomaliesCmd detects statistical anomalies in vitals metrics.
type MonitorAnomaliesCmd struct {
	Metrics        []string `help:"Metrics to analyze: crashes, anrs, errors, all" default:"all"`
	BaselinePeriod int      `help:"Days of baseline data for comparison" default:"30"`
	Sensitivity    string   `help:"Anomaly sensitivity: low, medium, high" default:"medium" enum:"low,medium,high"`
	Since          string   `help:"Start date for anomaly detection (ISO 8601)"`
	Format         string   `help:"Output format: json, table, html" default:"json" enum:"json,table,html"`
}

// anomalyResult represents detected anomalies.
type anomalyResult struct {
	Package        string            `json:"package"`
	Timestamp      time.Time         `json:"timestamp"`
	BaselinePeriod int               `json:"baselinePeriodDays"`
	Anomalies      []detectedAnomaly `json:"anomalies"`
	TotalAnomalies int               `json:"totalAnomalies"`
}

// detectedAnomaly represents a single detected anomaly.
type detectedAnomaly struct {
	Metric       string    `json:"metric"`
	Severity     string    `json:"severity"`
	Deviation    float64   `json:"deviationPercent"`
	CurrentValue float64   `json:"currentValue"`
	BaselineAvg  float64   `json:"baselineAverage"`
	Timestamp    time.Time `json:"timestamp"`
}

// Run executes the anomalies detection command.
func (cmd *MonitorAnomaliesCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	// Validate date format before authentication
	if cmd.Since != "" {
		_, err := time.Parse("2006-01-02", cmd.Since)
		if err != nil {
			return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid since date: %v", err))
		}
	}

	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	authMgr := newAuthManager()

	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource,
		api.WithTimeout(globals.Timeout),
		api.WithVerboseLogging(globals.Verbose))
	if err != nil {
		return errors.NewAPIError(errors.CodeAuthFailure, fmt.Sprintf("failed to create API client: %v", err))
	}

	svc, err := client.PlayReporting()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get reporting service: %v", err))
	}

	metrics := cmd.normalizeMetrics()
	startTime := time.Now()

	result := &anomalyResult{
		Package:        globals.Package,
		Timestamp:      startTime,
		BaselinePeriod: cmd.BaselinePeriod,
		Anomalies:      []detectedAnomaly{},
	}

	// Calculate date ranges
	var currentStart, currentEnd time.Time
	if cmd.Since != "" {
		currentStart, _ = time.Parse("2006-01-02", cmd.Since)
	} else {
		currentStart = time.Now().UTC().AddDate(0, 0, -7)
	}
	currentEnd = time.Now().UTC()

	baselineEnd := currentStart.AddDate(0, 0, -1)
	baselineStart := baselineEnd.AddDate(0, 0, -cmd.BaselinePeriod)

	// Detect anomalies for each metric
	for _, metric := range metrics {
		if err := cmd.detectAnomalies(ctx, client, svc, globals.Package, metric, baselineStart, baselineEnd, currentStart, currentEnd, result); err != nil {
			return err
		}
	}

	result.TotalAnomalies = len(result.Anomalies)

	outputResult := output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if result.TotalAnomalies > 0 {
		outputResult = outputResult.WithWarnings(
			fmt.Sprintf("%d anomalies detected in vitals metrics", result.TotalAnomalies),
		)
	}

	return outputResultResult(outputResult, cmd.Format, globals.Pretty)
}

func (cmd *MonitorAnomaliesCmd) normalizeMetrics() []string {
	if len(cmd.Metrics) == 0 {
		return []string{metricCrashes, metricAnrs, metricErrors}
	}

	var result []string
	for _, m := range cmd.Metrics {
		switch strings.ToLower(m) {
		case checkAll:
			return []string{metricCrashes, metricAnrs, metricErrors}
		case metricCrashes, metricCrash:
			result = append(result, metricCrashes)
		case metricAnrs, metricAnr:
			result = append(result, metricAnrs)
		case metricErrors, metricError:
			result = append(result, metricErrors)
		}
	}
	return result
}

func (cmd *MonitorAnomaliesCmd) detectAnomalies(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg, metric string, baselineStart, baselineEnd, currentStart, currentEnd time.Time, result *anomalyResult) error {
	multiplier := cmd.getSensitivityMultiplier()

	switch metric {
	case metricCrashes:
		return cmd.detectCrashAnomalies(ctx, client, svc, pkg, baselineStart, baselineEnd, currentStart, currentEnd, multiplier, result)
	case metricAnrs:
		return cmd.detectAnrAnomalies(ctx, client, svc, pkg, baselineStart, baselineEnd, currentStart, currentEnd, multiplier, result)
	case metricErrors:
		return cmd.detectErrorAnomalies(ctx, client, svc, pkg, baselineStart, baselineEnd, currentStart, currentEnd, multiplier, result)
	}
	return nil
}

func (cmd *MonitorAnomaliesCmd) getSensitivityMultiplier() float64 {
	switch cmd.Sensitivity {
	case "low":
		return 3.0
	case "high":
		return 1.5
	default:
		return 2.0
	}
}

func (cmd *MonitorAnomaliesCmd) detectCrashAnomalies(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, baselineStart, baselineEnd, currentStart, currentEnd time.Time, multiplier float64, result *anomalyResult) error {
	// Get baseline data
	baselineAvg, err := cmd.getCrashRate(ctx, client, svc, pkg, baselineStart, baselineEnd)
	if err != nil {
		return err
	}

	// Get current data
	currentRate, err := cmd.getCrashRate(ctx, client, svc, pkg, currentStart, currentEnd)
	if err != nil {
		return err
	}

	// Check for anomaly
	if baselineAvg > 0 && currentRate > baselineAvg*multiplier {
		deviation := ((currentRate - baselineAvg) / baselineAvg) * 100
		severity := cmd.calculateAnomalySeverity(deviation)

		anomaly := detectedAnomaly{
			Metric:       "crashRate",
			Severity:     severity,
			Deviation:    deviation,
			CurrentValue: currentRate,
			BaselineAvg:  baselineAvg,
			Timestamp:    time.Now(),
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}

	return nil
}

func (cmd *MonitorAnomaliesCmd) detectAnrAnomalies(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, baselineStart, baselineEnd, currentStart, currentEnd time.Time, multiplier float64, result *anomalyResult) error {
	baselineAvg, err := cmd.getAnrRate(ctx, client, svc, pkg, baselineStart, baselineEnd)
	if err != nil {
		return err
	}

	currentRate, err := cmd.getAnrRate(ctx, client, svc, pkg, currentStart, currentEnd)
	if err != nil {
		return err
	}

	if baselineAvg > 0 && currentRate > baselineAvg*multiplier {
		deviation := ((currentRate - baselineAvg) / baselineAvg) * 100
		severity := cmd.calculateAnomalySeverity(deviation)

		anomaly := detectedAnomaly{
			Metric:       "anrRate",
			Severity:     severity,
			Deviation:    deviation,
			CurrentValue: currentRate,
			BaselineAvg:  baselineAvg,
			Timestamp:    time.Now(),
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}

	return nil
}

func (cmd *MonitorAnomaliesCmd) detectErrorAnomalies(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, baselineStart, baselineEnd, currentStart, currentEnd time.Time, multiplier float64, result *anomalyResult) error {
	baselineAvg, err := cmd.getErrorCount(ctx, client, svc, pkg, baselineStart, baselineEnd)
	if err != nil {
		return err
	}

	currentCount, err := cmd.getErrorCount(ctx, client, svc, pkg, currentStart, currentEnd)
	if err != nil {
		return err
	}

	if baselineAvg > 0 && currentCount > baselineAvg*multiplier {
		deviation := ((float64(currentCount) - baselineAvg) / baselineAvg) * 100
		severity := cmd.calculateAnomalySeverity(deviation)

		anomaly := detectedAnomaly{
			Metric:       "errorCount",
			Severity:     severity,
			Deviation:    deviation,
			CurrentValue: float64(currentCount),
			BaselineAvg:  baselineAvg,
			Timestamp:    time.Now(),
		}
		result.Anomalies = append(result.Anomalies, anomaly)
	}

	return nil
}

func (cmd *MonitorAnomaliesCmd) getCrashRate(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time) (float64, error) {
	name := fmt.Sprintf("apps/%s/crashRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{"crashRate"},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Crashrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == "crashRate" {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}
	return totalRate / float64(count), nil
}

func (cmd *MonitorAnomaliesCmd) getAnrRate(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time) (float64, error) {
	name := fmt.Sprintf("apps/%s/anrRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{"anrRate"},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Anrrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == "anrRate" {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}
	return totalRate / float64(count), nil
}

func (cmd *MonitorAnomaliesCmd) getErrorCount(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time) (float64, error) {
	name := fmt.Sprintf("apps/%s/errorCountMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricErrorCount},
		PageSize:     100,
	}

	var totalCount float64

	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Errors.Counts.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricErrorCount {
					totalCount += parseDecimalValue(m)
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	// Return average per day
	days := int(end.Sub(start).Hours() / 24)
	if days == 0 {
		days = 1
	}
	return float64(totalCount) / float64(days), nil
}

func (cmd *MonitorAnomaliesCmd) buildTimelineSpec(start, end time.Time) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(start.Year()),
			Month: int64(start.Month()),
			Day:   int64(start.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(end.Year()),
			Month: int64(end.Month()),
			Day:   int64(end.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
	}
}

func (cmd *MonitorAnomaliesCmd) calculateAnomalySeverity(deviation float64) string {
	switch {
	case deviation >= 200:
		return "critical"
	case deviation >= 100:
		return "high"
	case deviation >= 50:
		return "medium"
	default:
		return "low"
	}
}

// MonitorDashboardCmd generates monitoring dashboard data.
type MonitorDashboardCmd struct {
	Metrics []string `help:"Metrics to include: crashes, anrs, errors, slow-rendering, slow-start, wakeups, wakelocks, all" default:"all"`
	Period  int      `help:"Days of data to include" default:"7"`
	Format  string   `help:"Output format: json, html, markdown" default:"json" enum:"json,html,markdown"`
}

// dashboardResult represents dashboard data.
type dashboardResult struct {
	Package     string                 `json:"package"`
	GeneratedAt time.Time              `json:"generatedAt"`
	PeriodDays  int                    `json:"periodDays"`
	Summary     dashboardSummary       `json:"summary"`
	Metrics     map[string]interface{} `json:"metrics"`
	Trends      dashboardTrends        `json:"trends"`
}

// dashboardSummary represents summary statistics.
type dashboardSummary struct {
	TotalCrashes  int64   `json:"totalCrashes"`
	TotalAnrs     int64   `json:"totalAnrs"`
	TotalErrors   int64   `json:"totalErrors"`
	AvgCrashRate  float64 `json:"averageCrashRate"`
	AvgAnrRate    float64 `json:"averageAnrRate"`
	AffectedUsers int64   `json:"affectedUsers"`
}

// dashboardTrends represents metric trends.
type dashboardTrends struct {
	CrashTrend string `json:"crashTrend"`
	AnrTrend   string `json:"anrTrend"`
	ErrorTrend string `json:"errorTrend"`
}

// Run executes the dashboard command.
func (cmd *MonitorDashboardCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	authMgr := newAuthManager()

	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource,
		api.WithTimeout(globals.Timeout),
		api.WithVerboseLogging(globals.Verbose))
	if err != nil {
		return errors.NewAPIError(errors.CodeAuthFailure, fmt.Sprintf("failed to create API client: %v", err))
	}

	svc, err := client.PlayReporting()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get reporting service: %v", err))
	}

	metrics := cmd.normalizeDashboardMetrics()
	startTime := time.Now()
	endDate := time.Now().UTC()
	startDate := endDate.AddDate(0, 0, -cmd.Period)

	result := &dashboardResult{
		Package:     globals.Package,
		GeneratedAt: startTime,
		PeriodDays:  cmd.Period,
		Metrics:     make(map[string]interface{}),
	}

	// Aggregate all requested metrics
	for _, metric := range metrics {
		switch metric {
		case metricCrashes:
			cmd.aggregateCrashes(ctx, client, svc, globals.Package, startDate, endDate, result)
		case metricAnrs:
			cmd.aggregateAnrs(ctx, client, svc, globals.Package, startDate, endDate, result)
		case metricErrors:
			cmd.aggregateErrors(ctx, client, svc, globals.Package, startDate, endDate, result)
		case "slow-rendering":
			cmd.aggregateSlowRendering(ctx, client, svc, globals.Package, startDate, endDate, result)
		case "slow-start":
			cmd.aggregateSlowStart(ctx, client, svc, globals.Package, startDate, endDate, result)
		case "wakeups":
			cmd.aggregateWakeups(ctx, client, svc, globals.Package, startDate, endDate, result)
		case "wakelocks":
			cmd.aggregateWakelocks(ctx, client, svc, globals.Package, startDate, endDate, result)
		}
	}

	// Calculate trends
	cmd.calculateTrends(result)

	outputResult := output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResultResult(outputResult, cmd.Format, globals.Pretty)
}

func (cmd *MonitorDashboardCmd) normalizeDashboardMetrics() []string {
	if len(cmd.Metrics) == 0 {
		return []string{metricCrashes, metricAnrs, metricErrors, "slow-rendering", "slow-start", "wakeups", "wakelocks"}
	}

	var result []string
	for _, m := range cmd.Metrics {
		switch strings.ToLower(m) {
		case checkAll:
			return []string{metricCrashes, metricAnrs, metricErrors, "slow-rendering", "slow-start", "wakeups", "wakelocks"}
		case metricCrashes, metricCrash:
			result = append(result, metricCrashes)
		case metricAnrs, metricAnr:
			result = append(result, metricAnrs)
		case metricErrors, metricError:
			result = append(result, metricErrors)
		case "slow-rendering", "slowrendering":
			result = append(result, "slow-rendering")
		case "slow-start", "slowstart":
			result = append(result, "slow-start")
		case "wakeups", "wakeup", "excessive-wakeups":
			result = append(result, "wakeups")
		case "wakelocks", "wakelock", "stuck-wakelocks":
			result = append(result, "wakelocks")
		}
	}
	return result
}

func (cmd *MonitorDashboardCmd) aggregateCrashes(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *dashboardResult) {
	name := fmt.Sprintf("apps/%s/crashRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricCrashRate, metricDistinctUsers},
		PageSize:     100,
	}

	var totalRate float64
	var count int
	var affectedUsers float64

	_ = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Crashrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricCrashRate {
					totalRate += parseDecimalValue(m)
					count++
				}
				if m.Metric == metricDistinctUsers {
					affectedUsers += parseDecimalValue(m)
				}
			}
		}
		return nil
	})

	if count > 0 {
		result.Summary.AvgCrashRate = totalRate / float64(count)
		result.Summary.AffectedUsers += int64(affectedUsers)
	}

	result.Metrics["crashes"] = map[string]interface{}{
		"averageCrashRate": result.Summary.AvgCrashRate,
		"affectedUsers":    int64(affectedUsers),
	}
}

func (cmd *MonitorDashboardCmd) aggregateAnrs(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *dashboardResult) {
	name := fmt.Sprintf("apps/%s/anrRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricAnrRate, metricDistinctUsers},
		PageSize:     100,
	}

	var totalRate float64
	var count int
	var affectedUsers float64

	_ = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Anrrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricAnrRate {
					totalRate += parseDecimalValue(m)
					count++
				}
				if m.Metric == metricDistinctUsers {
					affectedUsers += parseDecimalValue(m)
				}
			}
		}
		return nil
	})

	if count > 0 {
		result.Summary.AvgAnrRate = totalRate / float64(count)
		result.Summary.AffectedUsers += int64(affectedUsers)
	}

	result.Metrics["anrs"] = map[string]interface{}{
		"averageAnrRate": result.Summary.AvgAnrRate,
		"affectedUsers":  int64(affectedUsers),
	}
}

func (cmd *MonitorDashboardCmd) aggregateErrors(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *dashboardResult) {
	name := fmt.Sprintf("apps/%s/errorCountMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricErrorCount, metricDistinctUsers},
		PageSize:     100,
	}

	var totalErrors float64
	var affectedUsers float64

	_ = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Errors.Counts.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricErrorCount {
					totalErrors += parseDecimalValue(m)
				}
				if m.Metric == metricDistinctUsers {
					affectedUsers += parseDecimalValue(m)
				}
			}
		}
		return nil
	})

	result.Summary.TotalErrors = int64(totalErrors)
	result.Summary.AffectedUsers += int64(affectedUsers)

	result.Metrics["errors"] = map[string]interface{}{
		"totalErrors":   int64(totalErrors),
		"affectedUsers": int64(affectedUsers),
	}
}

func (cmd *MonitorDashboardCmd) aggregateSlowRendering(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *dashboardResult) {
	name := fmt.Sprintf("apps/%s/slowRenderingRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricSlowRenderingRate},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	_ = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Slowrenderingrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricSlowRenderingRate {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	avgRate := 0.0
	if count > 0 {
		avgRate = totalRate / float64(count)
	}

	result.Metrics["slowRendering"] = map[string]interface{}{
		"averageSlowRenderingRate": avgRate,
	}
}

func (cmd *MonitorDashboardCmd) aggregateSlowStart(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *dashboardResult) {
	name := fmt.Sprintf("apps/%s/slowStartRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricSlowStartRate},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	_ = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Slowstartrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricSlowStartRate {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	avgRate := 0.0
	if count > 0 {
		avgRate = totalRate / float64(count)
	}

	result.Metrics["slowStart"] = map[string]interface{}{
		"averageSlowStartRate": avgRate,
	}
}

func (cmd *MonitorDashboardCmd) aggregateWakeups(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *dashboardResult) {
	name := fmt.Sprintf("apps/%s/excessiveWakeupRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricExcessiveWakeupRate},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	_ = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Excessivewakeuprate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricExcessiveWakeupRate {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	avgRate := 0.0
	if count > 0 {
		avgRate = totalRate / float64(count)
	}

	result.Metrics["excessiveWakeups"] = map[string]interface{}{
		"averageExcessiveWakeupRate": avgRate,
	}
}

func (cmd *MonitorDashboardCmd) aggregateWakelocks(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *dashboardResult) {
	name := fmt.Sprintf("apps/%s/stuckBackgroundWakelockRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricStuckBackgroundWakelockRate},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	_ = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Stuckbackgroundwakelockrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricStuckBackgroundWakelockRate {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	avgRate := 0.0
	if count > 0 {
		avgRate = totalRate / float64(count)
	}

	result.Metrics["stuckWakelocks"] = map[string]interface{}{
		"averageStuckWakelockRate": avgRate,
	}
}

func (cmd *MonitorDashboardCmd) calculateTrends(result *dashboardResult) {
	// Simplified trend calculation based on metric values
	// In a real implementation, you'd compare current period vs previous period
	result.Trends.CrashTrend = trendStable
	result.Trends.AnrTrend = trendStable
	result.Trends.ErrorTrend = trendStable

	// Crash trend: stable at 0 or between 0.005 and 0.02, increasing above 0.02, decreasing between 0 and 0.005
	if result.Summary.AvgCrashRate > 0.02 {
		result.Trends.CrashTrend = "increasing"
	} else if result.Summary.AvgCrashRate > 0 && result.Summary.AvgCrashRate <= 0.005 {
		result.Trends.CrashTrend = "decreasing"
	}

	// ANR trend: stable at 0 or between 0.001 and 0.01, increasing above 0.01, decreasing between 0 and 0.001 (inclusive)
	if result.Summary.AvgAnrRate > 0.01 {
		result.Trends.AnrTrend = "increasing"
	} else if result.Summary.AvgAnrRate > 0 && result.Summary.AvgAnrRate <= 0.001 {
		result.Trends.AnrTrend = "decreasing"
	}
}

func (cmd *MonitorDashboardCmd) buildTimelineSpec(start, end time.Time) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(start.Year()),
			Month: int64(start.Month()),
			Day:   int64(start.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(end.Year()),
			Month: int64(end.Month()),
			Day:   int64(end.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
	}
}

// MonitorReportCmd generates scheduled monitoring reports.
type MonitorReportCmd struct {
	Period         string `help:"Report period: daily, weekly, monthly" default:"daily" enum:"daily,weekly,monthly"`
	Format         string `help:"Output format: json, html, markdown" default:"json" enum:"json,html,markdown"`
	IncludeRawData bool   `help:"Include raw metric data in report"`
}

// reportResult represents a monitoring report.
type reportResult struct {
	Package         string                 `json:"package"`
	ReportType      string                 `json:"reportType"`
	GeneratedAt     time.Time              `json:"generatedAt"`
	PeriodStart     time.Time              `json:"periodStart"`
	PeriodEnd       time.Time              `json:"periodEnd"`
	Summary         reportSummary          `json:"summary"`
	KeyFindings     []string               `json:"keyFindings"`
	Recommendations []string               `json:"recommendations"`
	RawData         map[string]interface{} `json:"rawData,omitempty"`
}

// reportSummary represents report summary statistics.
type reportSummary struct {
	OverallHealth  string  `json:"overallHealth"`
	CrashRate      float64 `json:"crashRate"`
	AnrRate        float64 `json:"anrRate"`
	ErrorCount     int64   `json:"errorCount"`
	ActiveUsers    int64   `json:"activeUsers"`
	IssuesResolved int     `json:"issuesResolved"`
	IssuesOpen     int     `json:"issuesOpen"`
}

// Run executes the report command.
func (cmd *MonitorReportCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	ctx := context.Background()
	authMgr := newAuthManager()

	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
	if err != nil {
		return errors.NewAPIError(errors.CodeAuthFailure, fmt.Sprintf("failed to create API client: %v", err))
	}

	svc, err := client.PlayReporting()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get reporting service: %v", err))
	}

	startTime := time.Now()

	// Calculate period
	periodEnd := time.Now().UTC()
	var periodStart time.Time
	switch cmd.Period {
	case "daily":
		periodStart = periodEnd.AddDate(0, 0, -1)
	case "weekly":
		periodStart = periodEnd.AddDate(0, 0, -7)
	case "monthly":
		periodStart = periodEnd.AddDate(0, -1, 0)
	}

	result := &reportResult{
		Package:     globals.Package,
		ReportType:  cmd.Period,
		GeneratedAt: startTime,
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		RawData:     make(map[string]interface{}),
	}

	// Gather metrics
	if err := cmd.gatherReportData(ctx, client, svc, globals.Package, periodStart, periodEnd, result); err != nil {
		return err
	}

	// Generate findings and recommendations
	cmd.generateFindings(result)
	cmd.generateRecommendations(result)

	if !cmd.IncludeRawData {
		result.RawData = nil
	}

	outputResult := output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResultResult(outputResult, cmd.Format, globals.Pretty)
}

func (cmd *MonitorReportCmd) gatherReportData(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time, result *reportResult) error {
	// Get crash data
	crashRate, err := cmd.getReportCrashRate(ctx, client, svc, pkg, start, end)
	if err != nil {
		return err
	}
	result.Summary.CrashRate = crashRate

	// Get ANR data
	anrRate, err := cmd.getReportAnrRate(ctx, client, svc, pkg, start, end)
	if err != nil {
		return err
	}
	result.Summary.AnrRate = anrRate

	// Get error data
	errorCount, err := cmd.getReportErrorCount(ctx, client, svc, pkg, start, end)
	if err != nil {
		return err
	}
	result.Summary.ErrorCount = errorCount

	// Get active users (estimate from crash metric users)
	activeUsers, err := cmd.getReportActiveUsers(ctx, client, svc, pkg, start, end)
	if err != nil {
		return err
	}
	result.Summary.ActiveUsers = activeUsers

	// Calculate overall health
	result.Summary.OverallHealth = cmd.calculateOverallHealth(crashRate, anrRate, errorCount)

	// Get error issues count
	issuesOpen, err := cmd.getErrorIssuesStatus(ctx, client, svc, pkg)
	if err == nil {
		result.Summary.IssuesOpen = issuesOpen
	}

	if cmd.IncludeRawData {
		result.RawData["crashRate"] = crashRate
		result.RawData["anrRate"] = anrRate
		result.RawData["errorCount"] = errorCount
		result.RawData["activeUsers"] = activeUsers
	}

	return nil
}

func (cmd *MonitorReportCmd) getReportCrashRate(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time) (float64, error) {
	name := fmt.Sprintf("apps/%s/crashRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{"crashRate"},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Crashrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == "crashRate" {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}
	return totalRate / float64(count), nil
}

func (cmd *MonitorReportCmd) getReportAnrRate(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time) (float64, error) {
	name := fmt.Sprintf("apps/%s/anrRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{"anrRate"},
		PageSize:     100,
	}

	var totalRate float64
	var count int

	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Anrrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == "anrRate" {
					totalRate += parseDecimalValue(m)
					count++
				}
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	if count == 0 {
		return 0, nil
	}
	return totalRate / float64(count), nil
}

func (cmd *MonitorReportCmd) getReportErrorCount(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time) (int64, error) {
	name := fmt.Sprintf("apps/%s/errorCountMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{"errorCount"},
		PageSize:     100,
	}

	var totalCount float64

	err := client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Errors.Counts.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == "errorCount" {
					totalCount += parseDecimalValue(m)
				}
			}
		}
		return nil
	})

	return int64(totalCount), err
}

func (cmd *MonitorReportCmd) getReportActiveUsers(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, start, end time.Time) (totalUsers int64, err error) {
	name := fmt.Sprintf("apps/%s/crashRateMetricSet", pkg)
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: cmd.buildTimelineSpec(start, end),
		Metrics:      []string{metricDistinctUsers},
		PageSize:     100,
	}

	var users float64

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Crashrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}
		for _, row := range resp.Rows {
			for _, m := range row.Metrics {
				if m.Metric == metricDistinctUsers {
					users += parseDecimalValue(m)
				}
			}
		}
		return nil
	})

	return int64(users), err
}

func (cmd *MonitorReportCmd) getErrorIssuesStatus(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string) (openCount int, err error) {
	parent := fmt.Sprintf("apps/%s/errorIssues", pkg)

	var allIssues []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Errors.Issues.Search(parent).Context(ctx).
			Filter(fmt.Sprintf("activeBetween(%q, %q)",
				time.Now().AddDate(0, 0, -30).Format("2006-01-02")+"T00:00:00Z",
				time.Now().Format("2006-01-02")+"T00:00:00Z")).
			PageSize(100).
			Do()
		if err != nil {
			return err
		}
		allIssues = append(allIssues, resp.ErrorIssues...)
		return nil
	})

	if err != nil {
		return 0, err
	}

	openCount = len(allIssues)

	return openCount, nil
}

func (cmd *MonitorReportCmd) calculateOverallHealth(crashRate, anrRate float64, errorCount int64) string {
	score := 100.0

	// Deduct for crash rate (max 40 points)
	score -= math.Min(crashRate*2000, 40)

	// Deduct for ANR rate (max 30 points)
	score -= math.Min(anrRate*3000, 30)

	// Deduct for error count (max 30 points)
	score -= math.Min(float64(errorCount)/100, 30)

	switch {
	case score >= 90:
		return "excellent"
	case score >= 70:
		return "good"
	case score >= 50:
		return "fair"
	default:
		return "poor"
	}
}

func (cmd *MonitorReportCmd) generateFindings(result *reportResult) {
	findings := []string{}

	if result.Summary.CrashRate > 0.02 {
		findings = append(findings, fmt.Sprintf("High crash rate detected: %.2f%% (threshold: 2%%)", result.Summary.CrashRate*100))
	}

	if result.Summary.AnrRate > 0.01 {
		findings = append(findings, fmt.Sprintf("High ANR rate detected: %.2f%% (threshold: 1%%)", result.Summary.AnrRate*100))
	}

	if result.Summary.ErrorCount > 1000 {
		findings = append(findings, fmt.Sprintf("High error volume: %d errors reported", result.Summary.ErrorCount))
	}

	if result.Summary.IssuesOpen > 10 {
		findings = append(findings, fmt.Sprintf("%d open error issues require attention", result.Summary.IssuesOpen))
	}

	if len(findings) == 0 {
		findings = append(findings, "No significant issues detected during this period")
	}

	result.KeyFindings = findings
}

func (cmd *MonitorReportCmd) generateRecommendations(result *reportResult) {
	recommendations := []string{}

	if result.Summary.CrashRate > 0.02 {
		recommendations = append(recommendations, "Prioritize fixing top crashes - consider using gpd vitals crashes to identify patterns")
	}

	if result.Summary.AnrRate > 0.01 {
		recommendations = append(recommendations, "Review ANR patterns - check main thread blocking operations and I/O on UI thread")
	}

	if result.Summary.ErrorCount > 1000 {
		recommendations = append(recommendations, "Investigate error patterns - use gpd vitals errors to analyze error clusters")
	}

	if result.Summary.IssuesOpen > 10 {
		recommendations = append(recommendations, "Address backlog of open issues - consider team sprint prioritization")
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Continue monitoring - current metrics are within healthy ranges")
	}

	result.Recommendations = recommendations
}

func (cmd *MonitorReportCmd) buildTimelineSpec(start, end time.Time) *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec {
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(start.Year()),
			Month: int64(start.Month()),
			Day:   int64(start.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  int64(end.Year()),
			Month: int64(end.Month()),
			Day:   int64(end.Day()),
			TimeZone: &playdeveloperreporting.GoogleTypeTimeZone{
				Id: "America/Los_Angeles",
			},
		},
	}
}

// MonitorWebhooksCmd contains webhook management commands.
type MonitorWebhooksCmd struct {
	List MonitorWebhooksListCmd `cmd:"" help:"List configured webhooks (simulated)"`
}

// MonitorWebhooksListCmd lists configured webhooks.
type MonitorWebhooksListCmd struct{}

// webhookInfo represents webhook configuration.
type webhookInfo struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	URL        string    `json:"url"`
	Events     []string  `json:"events"`
	Active     bool      `json:"active"`
	CreatedAt  time.Time `json:"createdAt"`
	LastCalled time.Time `json:"lastCalled,omitempty"`
	Status     string    `json:"status"`
}

// webhooksListResult represents the webhook list result.
type webhooksListResult struct {
	Webhooks   []webhookInfo `json:"webhooks"`
	TotalCount int           `json:"totalCount"`
	Note       string        `json:"note"`
}

// Run executes the webhooks list command.
func (cmd *MonitorWebhooksListCmd) Run(globals *Globals) error {
	// Google Play Console doesn't provide a public API for webhook management
	// This is a simulated command that shows what webhooks would look like

	startTime := time.Now()

	result := &webhooksListResult{
		Webhooks:   []webhookInfo{},
		TotalCount: 0,
		Note: "Google Play Console webhooks are configured through the Play Console UI, not via API. " +
			"Use the Play Console > Setup > API access > Webhooks section to configure.",
	}

	// Example webhook entries showing what the format would be
	exampleWebhooks := []webhookInfo{
		{
			ID:        "example-1",
			Name:      "Crash Alerts",
			URL:       "https://hooks.example.com/play/crashes",
			Events:    []string{"vitals.crashes", "vitals.anrs"},
			Active:    true,
			CreatedAt: time.Now().AddDate(0, -3, 0),
			Status:    "example-only",
		},
		{
			ID:         "example-2",
			Name:       "Release Notifications",
			URL:        "https://hooks.example.com/play/releases",
			Events:     []string{"publishing.releases"},
			Active:     true,
			CreatedAt:  time.Now().AddDate(0, -1, 0),
			LastCalled: time.Now().AddDate(0, 0, -2),
			Status:     "example-only",
		},
	}

	result.Webhooks = exampleWebhooks
	result.TotalCount = len(exampleWebhooks)

	outputResult := output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithWarnings("Webhooks API not available - showing example format only").
		WithServices("playdeveloperreporting")

	return outputResultResult(outputResult, "json", globals.Pretty)
}

// parseDecimalValue parses a decimal value from the API response.
func parseDecimalValue(m *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue) float64 {
	if m == nil || m.DecimalValue == nil || m.DecimalValue.Value == "" {
		return 0
	}
	val, _ := strconv.ParseFloat(m.DecimalValue.Value, 64)
	return val
}

// outputResultResult is a helper to output results with format support.
func outputResultResult(result *output.Result, format string, pretty bool) error {
	switch format {
	case "html":
		return outputHTML(result)
	case "markdown", "md":
		return outputMarkdown(result)
	default:
		return outputResult(result, format, pretty)
	}
}

// outputHTML outputs result as HTML.
func outputHTML(result *output.Result) error {
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Google Play Monitoring Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; padding: 30px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        h1 { color: #333; border-bottom: 2px solid #4CAF50; padding-bottom: 10px; }
        .metric { background: #f9f9f9; padding: 15px; margin: 10px 0; border-left: 4px solid #4CAF50; }
        .alert { background: #fff3cd; border-left-color: #ffc107; }
        .error { background: #f8d7da; border-left-color: #dc3545; }
        pre { background: #f4f4f4; padding: 15px; overflow-x: auto; }
        .timestamp { color: #666; font-size: 0.9em; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Google Play Monitoring Report</h1>
        <p class="timestamp">Generated: %s</p>
        <pre>%s</pre>
    </div>
</body>
</html>`, time.Now().Format(time.RFC3339), result.Data)

	fmt.Println(html)
	return nil
}

// outputMarkdown outputs result as Markdown.
func outputMarkdown(result *output.Result) error {
	fmt.Println("# Google Play Monitoring Report")
	fmt.Println()
	fmt.Printf("*Generated: %s*\n\n", time.Now().Format(time.RFC3339))

	switch data := result.Data.(type) {
	case map[string]interface{}:
		for k, v := range data {
			fmt.Printf("## %s\n\n", cases.Title(language.English).String(k))
			fmt.Printf("```\n%v\n```\n\n", v)
		}
	default:
		fmt.Printf("```json\n%+v\n```\n", data)
	}

	return nil
}
