// Package cli provides vitals commands for gpd.
package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

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
	crashesCmd.MarkFlagRequired("start-date")
	crashesCmd.MarkFlagRequired("end-date")

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
	anrsCmd.MarkFlagRequired("start-date")
	anrsCmd.MarkFlagRequired("end-date")

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
	queryCmd.MarkFlagRequired("start-date")
	queryCmd.MarkFlagRequired("end-date")

	// vitals capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List vitals capabilities",
		Long:  "List available vitals metrics and dimensions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsCapabilities(cmd.Context())
		},
	}

	vitalsCmd.AddCommand(crashesCmd, anrsCmd, queryCmd, capabilitiesCmd)
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
