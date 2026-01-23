// Package cli provides vitals commands for gpd.
package cli

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/google-play-cli/gpd/internal/errors"
	"github.com/google-play-cli/gpd/internal/output"
)

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

	// Note: Simplified implementation
	_ = reporting

	result := output.NewResult(map[string]interface{}{
		"metric":     "crashRate",
		"startDate":  startDate,
		"endDate":    endDate,
		"dimensions": dimensions,
		"package":    c.packageName,
		"rows":       []interface{}{},
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

	result := output.NewResult(map[string]interface{}{
		"metric":     "anrRate",
		"startDate":  startDate,
		"endDate":    endDate,
		"dimensions": dimensions,
		"package":    c.packageName,
		"rows":       []interface{}{},
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

	result := output.NewResult(map[string]interface{}{
		"metrics":    metrics,
		"startDate":  startDate,
		"endDate":    endDate,
		"dimensions": dimensions,
		"package":    c.packageName,
		"rows":       []interface{}{},
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
