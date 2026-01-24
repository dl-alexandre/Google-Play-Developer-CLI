// Package cli provides vitals commands for gpd.
package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

// parseYear extracts year from ISO 8601 date (YYYY-MM-DD)
func parseYear(date string) int64 {
	parts := strings.Split(date, "-")
	if len(parts) >= 1 {
		y, _ := strconv.ParseInt(parts[0], 10, 64)
		return y
	}
	return 0
}

// parseMonth extracts month from ISO 8601 date (YYYY-MM-DD)
func parseMonth(date string) int64 {
	parts := strings.Split(date, "-")
	if len(parts) >= 2 {
		m, _ := strconv.ParseInt(parts[1], 10, 64)
		return m
	}
	return 0
}

// parseDay extracts day from ISO 8601 date (YYYY-MM-DD)
func parseDay(date string) int64 {
	parts := strings.Split(date, "-")
	if len(parts) >= 3 {
		d, _ := strconv.ParseInt(parts[2], 10, 64)
		return d
	}
	return 0
}

func (c *CLI) addVitalsCommands() {
	vitalsCmd := &cobra.Command{
		Use:   "vitals",
		Short: "Android vitals commands",
		Long:  "Access crash rates, ANR rates, and performance metrics.",
	}

	var (
		startDate  string
		endDate    string
		metrics    []string
		dimensions []string
		outputFmt  string
		pageSize   int64
		pageToken  string
		all        bool
	)
	var (
		errorQuery      string
		errorInterval   string
		errorPageSize   int64
		errorPageToken  string
		deobfuscate     bool
		countsStartDate string
		countsEndDate   string
		countsDims      []string
	)
	var (
		anomalyMetric     string
		anomalyTimePeriod string
		anomalyMinSeverity string
		anomalyPageSize   int64
		anomalyPageToken  string
	)

	// vitals crashes
	crashesCmd := &cobra.Command{
		Use:   "crashes",
		Short: "Query crash rate data",
		Long:  "Query crash rate metrics for a date range.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsCrashes(cmd.Context(), startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	crashesCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	crashesCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	crashesCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	crashesCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	crashesCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	crashesCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	crashesCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = crashesCmd.MarkFlagRequired("start-date")
	_ = crashesCmd.MarkFlagRequired("end-date")

	// vitals anrs
	anrsCmd := &cobra.Command{
		Use:   "anrs",
		Short: "Query ANR rate data",
		Long:  "Query Application Not Responding (ANR) metrics.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsANRs(cmd.Context(), startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	anrsCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	anrsCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	anrsCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	anrsCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	anrsCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	anrsCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	anrsCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = anrsCmd.MarkFlagRequired("start-date")
	_ = anrsCmd.MarkFlagRequired("end-date")

	// vitals query (generic)
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Query vitals metrics",
		Long:  "Query vitals metrics for a date range.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsQuery(cmd.Context(), startDate, endDate, metrics, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	queryCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	queryCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	queryCmd.Flags().StringSliceVar(&metrics, "metrics", []string{"crashRate"}, "Metrics to retrieve")
	queryCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	queryCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	queryCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	queryCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	queryCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = queryCmd.MarkFlagRequired("start-date")
	_ = queryCmd.MarkFlagRequired("end-date")

	excessiveWakeupsCmd := &cobra.Command{
		Use:   "excessive-wakeups",
		Short: "Query excessive wakeups data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsExcessiveWakeups(cmd.Context(), startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	excessiveWakeupsCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	excessiveWakeupsCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	excessiveWakeupsCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	excessiveWakeupsCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	excessiveWakeupsCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	excessiveWakeupsCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	excessiveWakeupsCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = excessiveWakeupsCmd.MarkFlagRequired("start-date")
	_ = excessiveWakeupsCmd.MarkFlagRequired("end-date")

	lmkRateCmd := &cobra.Command{
		Use:   "lmk-rate",
		Short: "Query LMK rate data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsLmkRate(cmd.Context(), startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	lmkRateCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	lmkRateCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	lmkRateCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	lmkRateCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	lmkRateCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	lmkRateCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	lmkRateCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = lmkRateCmd.MarkFlagRequired("start-date")
	_ = lmkRateCmd.MarkFlagRequired("end-date")

	slowRenderingCmd := &cobra.Command{
		Use:   "slow-rendering",
		Short: "Query slow rendering data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsSlowRendering(cmd.Context(), startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	slowRenderingCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	slowRenderingCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	slowRenderingCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	slowRenderingCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	slowRenderingCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	slowRenderingCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	slowRenderingCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = slowRenderingCmd.MarkFlagRequired("start-date")
	_ = slowRenderingCmd.MarkFlagRequired("end-date")

	slowStartCmd := &cobra.Command{
		Use:   "slow-start",
		Short: "Query slow start data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsSlowStart(cmd.Context(), startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	slowStartCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	slowStartCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	slowStartCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	slowStartCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	slowStartCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	slowStartCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	slowStartCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = slowStartCmd.MarkFlagRequired("start-date")
	_ = slowStartCmd.MarkFlagRequired("end-date")

	stuckWakelocksCmd := &cobra.Command{
		Use:   "stuck-wakelocks",
		Short: "Query stuck wakelocks data",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsStuckWakelocks(cmd.Context(), startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	stuckWakelocksCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	stuckWakelocksCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	stuckWakelocksCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	stuckWakelocksCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	stuckWakelocksCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	stuckWakelocksCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	stuckWakelocksCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	_ = stuckWakelocksCmd.MarkFlagRequired("start-date")
	_ = stuckWakelocksCmd.MarkFlagRequired("end-date")

	// vitals capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List vitals capabilities",
		Long:  "List available vitals metrics and dimensions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsCapabilities(cmd.Context())
		},
	}

	errorsCmd := &cobra.Command{
		Use:   "errors",
		Short: "Search and report errors",
		Long:  "Search error issues, reports, and query error counts.",
	}

	errorsIssuesSearchCmd := &cobra.Command{
		Use:   "issues search",
		Short: "Search error issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsErrorsIssuesSearch(cmd.Context(), errorQuery, errorInterval, errorPageSize, errorPageToken)
		},
	}
	errorsIssuesSearchCmd.Flags().StringVar(&errorQuery, "query", "", "Search query")
	errorsIssuesSearchCmd.Flags().StringVar(&errorInterval, "interval", "last30Days", "Time interval")
	errorsIssuesSearchCmd.Flags().Int64Var(&errorPageSize, "page-size", 50, "Results per page")
	errorsIssuesSearchCmd.Flags().StringVar(&errorPageToken, "page-token", "", "Pagination token")

	errorsReportsSearchCmd := &cobra.Command{
		Use:   "reports search",
		Short: "Search error reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsErrorsReportsSearch(cmd.Context(), errorQuery, errorInterval, errorPageSize, errorPageToken, deobfuscate)
		},
	}
	errorsReportsSearchCmd.Flags().StringVar(&errorQuery, "query", "", "Search query")
	errorsReportsSearchCmd.Flags().StringVar(&errorInterval, "interval", "last30Days", "Time interval")
	errorsReportsSearchCmd.Flags().Int64Var(&errorPageSize, "page-size", 50, "Results per page")
	errorsReportsSearchCmd.Flags().StringVar(&errorPageToken, "page-token", "", "Pagination token")
	errorsReportsSearchCmd.Flags().BoolVar(&deobfuscate, "deobfuscate", false, "Format report text for readability")

	errorsCountsGetCmd := &cobra.Command{
		Use:   "counts get",
		Short: "Get error count metrics",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsErrorsCountsGet(cmd.Context())
		},
	}

	errorsCountsQueryCmd := &cobra.Command{
		Use:   "counts query",
		Short: "Query error counts over time",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsErrorsCountsQuery(cmd.Context(), countsStartDate, countsEndDate, countsDims, errorPageSize, errorPageToken)
		},
	}
	errorsCountsQueryCmd.Flags().StringVar(&countsStartDate, "start-date", "", "Start date (ISO 8601)")
	errorsCountsQueryCmd.Flags().StringVar(&countsEndDate, "end-date", "", "End date (ISO 8601)")
	errorsCountsQueryCmd.Flags().StringSliceVar(&countsDims, "dimensions", nil, "Dimensions for grouping")
	errorsCountsQueryCmd.Flags().Int64Var(&errorPageSize, "page-size", 100, "Results per page")
	errorsCountsQueryCmd.Flags().StringVar(&errorPageToken, "page-token", "", "Pagination token")
	_ = errorsCountsQueryCmd.MarkFlagRequired("start-date")
	_ = errorsCountsQueryCmd.MarkFlagRequired("end-date")

	errorsCmd.AddCommand(errorsIssuesSearchCmd, errorsReportsSearchCmd, errorsCountsGetCmd, errorsCountsQueryCmd)

	anomaliesCmd := &cobra.Command{
		Use:   "anomalies",
		Short: "Anomalies in vitals metrics",
	}
	anomaliesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List anomalies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsAnomaliesList(cmd.Context(), anomalyMetric, anomalyTimePeriod, anomalyMinSeverity, anomalyPageSize, anomalyPageToken)
		},
	}
	anomaliesListCmd.Flags().StringVar(&anomalyMetric, "metric", "", "Metric name filter")
	anomaliesListCmd.Flags().StringVar(&anomalyTimePeriod, "time-period", "last30Days", "Time period: last7Days, last30Days, last90Days")
	anomaliesListCmd.Flags().StringVar(&anomalyMinSeverity, "min-severity", "", "Minimum severity")
	anomaliesListCmd.Flags().Int64Var(&anomalyPageSize, "page-size", 20, "Results per page")
	anomaliesListCmd.Flags().StringVar(&anomalyPageToken, "page-token", "", "Pagination token")
	anomaliesCmd.AddCommand(anomaliesListCmd)

	vitalsCmd.AddCommand(crashesCmd, anrsCmd, excessiveWakeupsCmd, lmkRateCmd, slowRenderingCmd, slowStartCmd, stuckWakelocksCmd, queryCmd, errorsCmd, anomaliesCmd, capabilitiesCmd)
	c.rootCmd.AddCommand(vitalsCmd)
}

func (c *CLI) vitalsCrashes(ctx context.Context, startDate, endDate string, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Get API client
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	// Build the app name for the API
	appName := fmt.Sprintf("apps/%s", c.packageName)

	// Build timeline spec
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(startDate),
			Month: parseMonth(startDate),
			Day:   parseDay(startDate),
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(endDate),
			Month: parseMonth(endDate),
			Day:   parseDay(endDate),
		},
	}

	// Build query request
	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}

	// Add dimensions if specified
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}

	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	// Query crash rate metric set
	var allRows []map[string]interface{}
	for {
		resp, err := reporting.Vitals.Crashrate.Query(appName, queryReq).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to query crash rate: %v", err)))
		}

		for _, row := range resp.Rows {
			rowData := map[string]interface{}{
				"startTime": row.StartTime,
			}
			// Extract metrics by name
			for _, m := range row.Metrics {
				if m.DecimalValue != nil {
					rowData[m.Metric] = m.DecimalValue.Value
				}
			}
			// Extract dimensions by name
			for _, d := range row.Dimensions {
				rowData[d.Dimension] = d.StringValue
			}
			allRows = append(allRows, rowData)
		}

		if !all || resp.NextPageToken == "" {
			pageToken = resp.NextPageToken
			break
		}
		queryReq.PageToken = resp.NextPageToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "crashRate",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": pageToken,
		"dataFreshness": map[string]interface{}{
			"note": "Vitals data may be delayed by 24-48 hours",
		},
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsANRs(ctx context.Context, startDate, endDate string, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Get API client
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	// Build the app name for the API
	appName := fmt.Sprintf("apps/%s", c.packageName)

	// Build timeline spec
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(startDate),
			Month: parseMonth(startDate),
			Day:   parseDay(startDate),
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(endDate),
			Month: parseMonth(endDate),
			Day:   parseDay(endDate),
		},
	}

	// Build query request
	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}

	// Add dimensions if specified
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}

	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	// Query ANR rate metric set
	var allRows []map[string]interface{}
	for {
		resp, err := reporting.Vitals.Anrrate.Query(appName, queryReq).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to query ANR rate: %v", err)))
		}

		for _, row := range resp.Rows {
			rowData := map[string]interface{}{
				"startTime": row.StartTime,
			}
			// Extract metrics by name
			for _, m := range row.Metrics {
				if m.DecimalValue != nil {
					rowData[m.Metric] = m.DecimalValue.Value
				}
			}
			// Extract dimensions by name
			for _, d := range row.Dimensions {
				rowData[d.Dimension] = d.StringValue
			}
			allRows = append(allRows, rowData)
		}

		if !all || resp.NextPageToken == "" {
			pageToken = resp.NextPageToken
			break
		}
		queryReq.PageToken = resp.NextPageToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "anrRate",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": pageToken,
		"dataFreshness": map[string]interface{}{
			"note": "Vitals data may be delayed by 24-48 hours",
		},
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsQuery(ctx context.Context, startDate, endDate string, metrics, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// The generic query routes to specific metric implementations
	// based on the requested metrics
	var allResults []map[string]interface{}

	for _, metric := range metrics {
		switch metric {
		case "crashRate":
			// Route to crash rate query
			err := c.vitalsCrashes(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil // Already output result
		case "anrRate":
			// Route to ANR rate query
			err := c.vitalsANRs(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
			if err != nil {
				return err
			}
			return nil // Already output result
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

func (c *CLI) vitalsCapabilities(ctx context.Context) error {
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

func (c *CLI) vitalsExcessiveWakeups(ctx context.Context, startDate, endDate string, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(startDate),
			Month: parseMonth(startDate),
			Day:   parseDay(startDate),
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(endDate),
			Month: parseMonth(endDate),
			Day:   parseDay(endDate),
		},
	}

	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}
	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	var allRows []map[string]interface{}
	for {
		resp, err := reporting.Vitals.Excessivewakeuprate.Query(appName, queryReq).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to query excessive wakeups: %v", err)))
		}

		for _, row := range resp.Rows {
			rowData := map[string]interface{}{
				"startTime": row.StartTime,
			}
			for _, m := range row.Metrics {
				if m.DecimalValue != nil {
					rowData[m.Metric] = m.DecimalValue.Value
				}
			}
			for _, d := range row.Dimensions {
				rowData[d.Dimension] = d.StringValue
			}
			allRows = append(allRows, rowData)
		}

		if !all || resp.NextPageToken == "" {
			pageToken = resp.NextPageToken
			break
		}
		queryReq.PageToken = resp.NextPageToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "excessiveWakeups",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": pageToken,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsLmkRate(ctx context.Context, startDate, endDate string, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {
	// LMK rate metric is not available in the Play Developer Reporting API v1beta1
	// The API does not provide a dedicated LMK rate metric set endpoint
	return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
		"LMK rate metric is not available in the Play Developer Reporting API. "+
			"Please use other available metrics such as crashRate, anrRate, excessiveWakeups, etc."))
}

func (c *CLI) vitalsSlowRendering(ctx context.Context, startDate, endDate string, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(startDate),
			Month: parseMonth(startDate),
			Day:   parseDay(startDate),
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(endDate),
			Month: parseMonth(endDate),
			Day:   parseDay(endDate),
		},
	}

	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}
	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	var allRows []map[string]interface{}
	for {
		resp, err := reporting.Vitals.Slowrenderingrate.Query(appName, queryReq).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to query slow rendering: %v", err)))
		}

		for _, row := range resp.Rows {
			rowData := map[string]interface{}{
				"startTime": row.StartTime,
			}
			for _, m := range row.Metrics {
				if m.DecimalValue != nil {
					rowData[m.Metric] = m.DecimalValue.Value
				}
			}
			for _, d := range row.Dimensions {
				rowData[d.Dimension] = d.StringValue
			}
			allRows = append(allRows, rowData)
		}

		if !all || resp.NextPageToken == "" {
			pageToken = resp.NextPageToken
			break
		}
		queryReq.PageToken = resp.NextPageToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "slowRendering",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": pageToken,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsSlowStart(ctx context.Context, startDate, endDate string, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(startDate),
			Month: parseMonth(startDate),
			Day:   parseDay(startDate),
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(endDate),
			Month: parseMonth(endDate),
			Day:   parseDay(endDate),
		},
	}

	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}
	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	var allRows []map[string]interface{}
	for {
		resp, err := reporting.Vitals.Slowstartrate.Query(appName, queryReq).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to query slow start: %v", err)))
		}

		for _, row := range resp.Rows {
			rowData := map[string]interface{}{
				"startTime": row.StartTime,
			}
			for _, m := range row.Metrics {
				if m.DecimalValue != nil {
					rowData[m.Metric] = m.DecimalValue.Value
				}
			}
			for _, d := range row.Dimensions {
				rowData[d.Dimension] = d.StringValue
			}
			allRows = append(allRows, rowData)
		}

		if !all || resp.NextPageToken == "" {
			pageToken = resp.NextPageToken
			break
		}
		queryReq.PageToken = resp.NextPageToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "slowStart",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": pageToken,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsStuckWakelocks(ctx context.Context, startDate, endDate string, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {

	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(startDate),
			Month: parseMonth(startDate),
			Day:   parseDay(startDate),
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(endDate),
			Month: parseMonth(endDate),
			Day:   parseDay(endDate),
		},
	}

	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}
	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	var allRows []map[string]interface{}
	for {
		resp, err := reporting.Vitals.Stuckbackgroundwakelockrate.Query(appName, queryReq).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to query stuck wakelocks: %v", err)))
		}

		for _, row := range resp.Rows {
			rowData := map[string]interface{}{
				"startTime": row.StartTime,
			}
			for _, m := range row.Metrics {
				if m.DecimalValue != nil {
					rowData[m.Metric] = m.DecimalValue.Value
				}
			}
			for _, d := range row.Dimensions {
				rowData[d.Dimension] = d.StringValue
			}
			allRows = append(allRows, rowData)
		}

		if !all || resp.NextPageToken == "" {
			pageToken = resp.NextPageToken
			break
		}
		queryReq.PageToken = resp.NextPageToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "stuckWakelocks",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": pageToken,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsAnomaliesList(ctx context.Context, metric, timePeriod, minSeverity string, pageSize int64, pageToken string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	appName := fmt.Sprintf("apps/%s", c.packageName)
	req := reporting.Anomalies.List(appName)
	filter := buildAnomalyFilter(timePeriod)
	if filter != "" {
		req = req.Filter(filter)
	}
	if pageSize > 0 {
		req = req.PageSize(pageSize)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, err := req.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	anomalies := resp.Anomalies
	if metric != "" {
		filtered := make([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly, 0, len(anomalies))
		for _, anomaly := range anomalies {
			if anomaly == nil {
				continue
			}
			if anomaly.Metric != nil && strings.EqualFold(anomaly.Metric.Metric, metric) {
				filtered = append(filtered, anomaly)
				continue
			}
			if strings.Contains(strings.ToLower(anomaly.MetricSet), strings.ToLower(metric)) {
				filtered = append(filtered, anomaly)
			}
		}
		anomalies = filtered
	}
	result := output.NewResult(map[string]interface{}{
		"anomalies":     anomalies,
		"metric":        metric,
		"timePeriod":    timePeriod,
		"nextPageToken": resp.NextPageToken,
		"package":       c.packageName,
	})
	if minSeverity != "" {
		result.WithWarnings("min-severity filtering is not supported by the API")
	}
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func buildAnomalyFilter(timePeriod string) string {
	now := time.Now().UTC()
	switch timePeriod {
	case "last7Days":
		return fmt.Sprintf("activeBetween(\"%s\", \"%s\")", now.AddDate(0, 0, -7).Format(time.RFC3339), now.Format(time.RFC3339))
	case "last30Days":
		return fmt.Sprintf("activeBetween(\"%s\", \"%s\")", now.AddDate(0, 0, -30).Format(time.RFC3339), now.Format(time.RFC3339))
	case "last90Days":
		return fmt.Sprintf("activeBetween(\"%s\", \"%s\")", now.AddDate(0, 0, -90).Format(time.RFC3339), now.Format(time.RFC3339))
	case "", "all":
		return ""
	default:
		return ""
	}
}

func (c *CLI) vitalsErrorsIssuesSearch(ctx context.Context, query, interval string, pageSize int64, pageToken string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	searchCall := reporting.Vitals.Errors.Issues.Search(appName)

	// Set filter if query is provided
	if query != "" {
		searchCall = searchCall.Filter(query)
	}

	// Set interval if provided (convert interval string to start/end times)
	if interval != "" {
		now := time.Now().UTC()
		var startTime time.Time
		switch interval {
		case "last7Days":
			startTime = now.AddDate(0, 0, -7)
		case "last30Days":
			startTime = now.AddDate(0, 0, -30)
		case "last90Days":
			startTime = now.AddDate(0, 0, -90)
		default:
			// If interval is not recognized, try to parse as date or use default
			startTime = now.AddDate(0, 0, -30)
		}
		searchCall = searchCall.IntervalStartTimeYear(int64(startTime.Year())).
			IntervalStartTimeMonth(int64(startTime.Month())).
			IntervalStartTimeDay(int64(startTime.Day())).
			IntervalEndTimeYear(int64(now.Year())).
			IntervalEndTimeMonth(int64(now.Month())).
			IntervalEndTimeDay(int64(now.Day()))
	}

	// Set pagination parameters
	if pageSize > 0 {
		searchCall = searchCall.PageSize(pageSize)
	}
	if pageToken != "" {
		searchCall = searchCall.PageToken(pageToken)
	}

	resp, err := searchCall.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to search error issues: %v", err)))
	}

	result := output.NewResult(map[string]interface{}{
		"query":        query,
		"interval":     interval,
		"package":      c.packageName,
		"issues":       resp.ErrorIssues,
		"rowCount":     len(resp.ErrorIssues),
		"nextPageToken": resp.NextPageToken,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsErrorsReportsSearch(ctx context.Context, query, interval string, pageSize int64, pageToken string, formatReport bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	searchCall := reporting.Vitals.Errors.Reports.Search(appName)

	// Set filter if query is provided
	if query != "" {
		searchCall = searchCall.Filter(query)
	}

	// Set interval if provided (convert interval string to start/end times)
	if interval != "" {
		now := time.Now().UTC()
		var startTime time.Time
		switch interval {
		case "last7Days":
			startTime = now.AddDate(0, 0, -7)
		case "last30Days":
			startTime = now.AddDate(0, 0, -30)
		case "last90Days":
			startTime = now.AddDate(0, 0, -90)
		default:
			// If interval is not recognized, try to parse as date or use default
			startTime = now.AddDate(0, 0, -30)
		}
		searchCall = searchCall.IntervalStartTimeYear(int64(startTime.Year())).
			IntervalStartTimeMonth(int64(startTime.Month())).
			IntervalStartTimeDay(int64(startTime.Day())).
			IntervalEndTimeYear(int64(now.Year())).
			IntervalEndTimeMonth(int64(now.Month())).
			IntervalEndTimeDay(int64(now.Day()))
	}

	// Set pagination parameters
	if pageSize > 0 {
		searchCall = searchCall.PageSize(pageSize)
	}
	if pageToken != "" {
		searchCall = searchCall.PageToken(pageToken)
	}

	resp, err := searchCall.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to search error reports: %v", err)))
	}

	if formatReport {
		for _, report := range resp.ErrorReports {
			if report != nil && report.ReportText != "" {
				report.ReportText = formatReportText(report.ReportText)
			}
		}
	}

	result := output.NewResult(map[string]interface{}{
		"query":        query,
		"interval":     interval,
		"package":      c.packageName,
		"reports":      resp.ErrorReports,
		"rowCount":     len(resp.ErrorReports),
		"nextPageToken": resp.NextPageToken,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsErrorsCountsGet(ctx context.Context) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	resp, err := reporting.Vitals.Errors.Counts.Get(appName).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to get error counts: %v", err)))
	}

	result := output.NewResult(map[string]interface{}{
		"package": c.packageName,
		"counts":  resp,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsErrorsCountsQuery(ctx context.Context, startDate, endDate string, dimensions []string, pageSize int64, pageToken string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appName := fmt.Sprintf("apps/%s", c.packageName)
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(startDate),
			Month: parseMonth(startDate),
			Day:   parseDay(startDate),
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  parseYear(endDate),
			Month: parseMonth(endDate),
			Day:   parseDay(endDate),
		},
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}
	if len(dimensions) > 0 {
		req.Dimensions = dimensions
	}
	if pageToken != "" {
		req.PageToken = pageToken
	}

	resp, err := reporting.Vitals.Errors.Counts.Query(appName, req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to query error counts: %v", err)))
	}

	var rows []map[string]interface{}
	for _, row := range resp.Rows {
		rowData := map[string]interface{}{
			"startTime": row.StartTime,
		}
		for _, m := range row.Metrics {
			if m.DecimalValue != nil {
				rowData[m.Metric] = m.DecimalValue.Value
			}
		}
		for _, d := range row.Dimensions {
			rowData[d.Dimension] = d.StringValue
		}
		rows = append(rows, rowData)
	}

	result := output.NewResult(map[string]interface{}{
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          rows,
		"rowCount":      len(rows),
		"nextPageToken": resp.NextPageToken,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
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
