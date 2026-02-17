// Package cli provides analytics commands for gpd.
package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addAnalyticsCommands() {
	analyticsCmd := &cobra.Command{
		Use:   "analytics",
		Short: "Analytics commands",
		Long:  "Access app analytics and install statistics.",
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

	// analytics query
	queryCmd := &cobra.Command{
		Use:   "query",
		Short: "Query analytics data",
		Long:  "Query app analytics data for a date range.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.analyticsQuery(cmd.Context(), startDate, endDate, metrics, dimensions, outputFmt, pageSize, pageToken, all)
		},
	}
	queryCmd.Flags().StringVar(&startDate, "start-date", "", "Start date (ISO 8601)")
	queryCmd.Flags().StringVar(&endDate, "end-date", "", "End date (ISO 8601)")
	queryCmd.Flags().StringSliceVar(&metrics, "metrics", []string{"installs"}, "Metrics to retrieve")
	queryCmd.Flags().StringSliceVar(&dimensions, "dimensions", nil, "Dimensions for grouping")
	queryCmd.Flags().StringVar(&outputFmt, "format", "json", "Output format: json, csv")
	queryCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	queryCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(queryCmd, &all)

	// analytics capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List analytics capabilities",
		Long:  "List available metrics, dimensions, and granularities.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.analyticsCapabilities(cmd.Context())
		},
	}

	analyticsCmd.AddCommand(queryCmd, capabilitiesCmd)
	c.rootCmd.AddCommand(analyticsCmd)
}

func (c *CLI) analyticsQuery(ctx context.Context, startDate, endDate string, metrics, dimensions []string,
	outputFmt string, pageSize int64, pageToken string, all bool) error {
	if apiErr := validateReportingDates(startDate, endDate); apiErr != nil {
		result := output.NewErrorResult(apiErr).WithServices("playdeveloperreporting")
		return c.Output(result)
	}
	if len(metrics) != 1 {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "exactly one metric is required").
			WithHint("Use --metrics with a single value. For multiple metrics, run separate queries.")).
			WithServices("playdeveloperreporting")
		return c.Output(result)
	}

	metric := metrics[0]
	switch metric {
	case "crashRate":
		return c.vitalsCrashes(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
	case "anrRate":
		return c.vitalsANRs(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
	case "excessiveWakeups":
		return c.vitalsExcessiveWakeups(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
	case "slowRendering":
		return c.vitalsSlowRendering(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
	case "slowStart":
		return c.vitalsSlowStart(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
	case "stuckWakelocks":
		return c.vitalsStuckWakelocks(ctx, startDate, endDate, dimensions, outputFmt, pageSize, pageToken, all)
	default:
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
			"requested metric is not available in public Play Reporting APIs").
			WithHint("Supported metrics: crashRate, anrRate, excessiveWakeups, slowRendering, slowStart, stuckWakelocks")).
			WithServices("playdeveloperreporting")
		return c.Output(result)
	}
}

func (c *CLI) analyticsCapabilities(_ context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"requiredScopes": []string{"https://www.googleapis.com/auth/playdeveloperreporting"},
		"metrics": []map[string]interface{}{
			{"name": "crashRate", "description": "Crash rate per 1000 sessions"},
			{"name": "anrRate", "description": "ANR rate per 1000 sessions"},
			{"name": "excessiveWakeups", "description": "Excessive wakeups"},
			{"name": "slowRendering", "description": "Slow rendering rate"},
			{"name": "slowStart", "description": "Slow start rate"},
			{"name": "stuckWakelocks", "description": "Stuck wakelocks"},
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
			"note":    "Analytics queries proxy to Play Reporting vitals metrics",
		},
		"notes": []string{
			"Install, revenue, and discovery analytics are not available via public APIs.",
			"Use gpd vitals for full app quality metrics and error reports.",
		},
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}
