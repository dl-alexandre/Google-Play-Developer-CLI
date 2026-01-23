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
	queryCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	queryCmd.MarkFlagRequired("start-date")
	queryCmd.MarkFlagRequired("end-date")

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

	// Note: This is a simplified implementation
	// Actual implementation would use proper Play Developer Reporting API calls
	_ = reporting

	result := output.NewResult(map[string]interface{}{
		"startDate":  startDate,
		"endDate":    endDate,
		"metrics":    metrics,
		"dimensions": dimensions,
		"package":    c.packageName,
		"rows":       []interface{}{},
		"dataFreshness": map[string]interface{}{
			"note": "Data may be delayed by 24-48 hours",
		},
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) analyticsCapabilities(ctx context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"metrics": []map[string]interface{}{
			{"name": "installs", "description": "Total installs"},
			{"name": "uninstalls", "description": "Total uninstalls"},
			{"name": "updates", "description": "App updates"},
			{"name": "activeDevices", "description": "Active devices"},
			{"name": "crashes", "description": "Crash count"},
			{"name": "anrs", "description": "ANR count"},
		},
		"dimensions": []map[string]interface{}{
			{"name": "country", "description": "Country code"},
			{"name": "device", "description": "Device model"},
			{"name": "androidVersion", "description": "Android OS version"},
			{"name": "appVersion", "description": "App version code"},
		},
		"granularities":   []string{"daily", "weekly", "monthly"},
		"maxLookbackDays": 365,
		"dataFreshness": map[string]interface{}{
			"typical": "24-48 hours",
			"note":    "Data freshness varies by metric type",
		},
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}
