package cli

import (
	"github.com/spf13/cobra"
)

func (c *CLI) addVitalsCommands() {
	vitalsCmd := &cobra.Command{
		Use:   "vitals",
		Short: "Android vitals commands",
		Long:  "Access crash rates, ANR rates, and performance metrics.",
	}

	c.addVitalsMetricsCommands(vitalsCmd)
	c.addVitalsErrorsCommands(vitalsCmd)
	c.addVitalsAnomaliesCommands(vitalsCmd)
	c.addVitalsCapabilitiesCommand(vitalsCmd)

	c.rootCmd.AddCommand(vitalsCmd)
}

func (c *CLI) addVitalsMetricsCommands(vitalsCmd *cobra.Command) {
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
	addPaginationFlags(crashesCmd, &all)
	_ = crashesCmd.MarkFlagRequired("start-date")
	_ = crashesCmd.MarkFlagRequired("end-date")

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
	addPaginationFlags(anrsCmd, &all)
	_ = anrsCmd.MarkFlagRequired("start-date")
	_ = anrsCmd.MarkFlagRequired("end-date")

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
	addPaginationFlags(queryCmd, &all)
	_ = queryCmd.MarkFlagRequired("start-date")
	_ = queryCmd.MarkFlagRequired("end-date")

	vitalsCmd.AddCommand(crashesCmd, anrsCmd, queryCmd)
	c.addVitalsPerformanceCommands(vitalsCmd)
}

func (c *CLI) addVitalsPerformanceCommands(vitalsCmd *cobra.Command) {
	var (
		startDate  string
		endDate    string
		dimensions []string
		outputFmt  string
		pageSize   int64
		pageToken  string
		all        bool
	)

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
	addPaginationFlags(excessiveWakeupsCmd, &all)
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
	addPaginationFlags(lmkRateCmd, &all)
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
	addPaginationFlags(slowRenderingCmd, &all)
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
	addPaginationFlags(slowStartCmd, &all)
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
	addPaginationFlags(stuckWakelocksCmd, &all)
	_ = stuckWakelocksCmd.MarkFlagRequired("start-date")
	_ = stuckWakelocksCmd.MarkFlagRequired("end-date")

	vitalsCmd.AddCommand(excessiveWakeupsCmd, lmkRateCmd, slowRenderingCmd, slowStartCmd, stuckWakelocksCmd)
}

func (c *CLI) addVitalsErrorsCommands(vitalsCmd *cobra.Command) {
	errorsCmd := &cobra.Command{
		Use:   "errors",
		Short: "Search and report errors",
		Long:  "Search error issues, reports, and query error counts.",
	}

	var (
		errorQuery     string
		errorInterval  string
		errorPageSize  int64
		errorPageToken string
		errorAll       bool
		deobfuscate    bool
	)

	errorsIssuesSearchCmd := &cobra.Command{
		Use:   "issues search",
		Short: "Search error issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsErrorsIssuesSearch(cmd.Context(), errorQuery, errorInterval, errorPageSize, errorPageToken, errorAll)
		},
	}
	errorsIssuesSearchCmd.Flags().StringVar(&errorQuery, "query", "", "Search query")
	errorsIssuesSearchCmd.Flags().StringVar(&errorInterval, "interval", "last30Days", "Time interval")
	errorsIssuesSearchCmd.Flags().Int64Var(&errorPageSize, "page-size", 50, "Results per page")
	errorsIssuesSearchCmd.Flags().StringVar(&errorPageToken, "page-token", "", "Pagination token")
	addPaginationFlags(errorsIssuesSearchCmd, &errorAll)

	errorsReportsSearchCmd := &cobra.Command{
		Use:   "reports search",
		Short: "Search error reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsErrorsReportsSearch(cmd.Context(), errorQuery, errorInterval, errorPageSize, errorPageToken, errorAll, deobfuscate)
		},
	}
	errorsReportsSearchCmd.Flags().StringVar(&errorQuery, "query", "", "Search query")
	errorsReportsSearchCmd.Flags().StringVar(&errorInterval, "interval", "last30Days", "Time interval")
	errorsReportsSearchCmd.Flags().Int64Var(&errorPageSize, "page-size", 50, "Results per page")
	errorsReportsSearchCmd.Flags().StringVar(&errorPageToken, "page-token", "", "Pagination token")
	errorsReportsSearchCmd.Flags().BoolVar(&deobfuscate, "deobfuscate", false, "Format report text for readability")
	addPaginationFlags(errorsReportsSearchCmd, &errorAll)

	errorsCmd.AddCommand(errorsIssuesSearchCmd, errorsReportsSearchCmd)
	c.addVitalsErrorsCountsCommands(errorsCmd)
	vitalsCmd.AddCommand(errorsCmd)
}

func (c *CLI) addVitalsErrorsCountsCommands(errorsCmd *cobra.Command) {
	var (
		countsStartDate string
		countsEndDate   string
		countsDims      []string
		errorPageSize   int64
		errorPageToken  string
		errorAll        bool
	)

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
			return c.vitalsErrorsCountsQuery(cmd.Context(), countsStartDate, countsEndDate, countsDims, errorPageSize, errorPageToken, errorAll)
		},
	}
	errorsCountsQueryCmd.Flags().StringVar(&countsStartDate, "start-date", "", "Start date (ISO 8601)")
	errorsCountsQueryCmd.Flags().StringVar(&countsEndDate, "end-date", "", "End date (ISO 8601)")
	errorsCountsQueryCmd.Flags().StringSliceVar(&countsDims, "dimensions", nil, "Dimensions for grouping")
	errorsCountsQueryCmd.Flags().Int64Var(&errorPageSize, "page-size", 100, "Results per page")
	errorsCountsQueryCmd.Flags().StringVar(&errorPageToken, "page-token", "", "Pagination token")
	addPaginationFlags(errorsCountsQueryCmd, &errorAll)
	_ = errorsCountsQueryCmd.MarkFlagRequired("start-date")
	_ = errorsCountsQueryCmd.MarkFlagRequired("end-date")

	errorsCmd.AddCommand(errorsCountsGetCmd, errorsCountsQueryCmd)
}

func (c *CLI) addVitalsAnomaliesCommands(vitalsCmd *cobra.Command) {
	anomaliesCmd := &cobra.Command{
		Use:   "anomalies",
		Short: "Anomalies in vitals metrics",
	}

	var (
		anomalyMetric      string
		anomalyTimePeriod  string
		anomalyMinSeverity string
		anomalyPageSize    int64
		anomalyPageToken   string
		anomalyAll         bool
	)

	anomaliesListCmd := &cobra.Command{
		Use:   "list",
		Short: "List anomalies",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsAnomaliesList(cmd.Context(), anomalyMetric, anomalyTimePeriod, anomalyMinSeverity, anomalyPageSize, anomalyPageToken, anomalyAll)
		},
	}
	anomaliesListCmd.Flags().StringVar(&anomalyMetric, "metric", "", "Metric name filter")
	anomaliesListCmd.Flags().StringVar(&anomalyTimePeriod, "time-period", "last30Days", "Time period: last7Days, last30Days, last90Days")
	anomaliesListCmd.Flags().StringVar(&anomalyMinSeverity, "min-severity", "", "Minimum severity")
	anomaliesListCmd.Flags().Int64Var(&anomalyPageSize, "page-size", 20, "Results per page")
	anomaliesListCmd.Flags().StringVar(&anomalyPageToken, "page-token", "", "Pagination token")
	addPaginationFlags(anomaliesListCmd, &anomalyAll)

	anomaliesCmd.AddCommand(anomaliesListCmd)
	vitalsCmd.AddCommand(anomaliesCmd)
}

func (c *CLI) addVitalsCapabilitiesCommand(vitalsCmd *cobra.Command) {
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List vitals capabilities",
		Long:  "List available vitals metrics and dimensions.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.vitalsCapabilities(cmd.Context())
		},
	}
	vitalsCmd.AddCommand(capabilitiesCmd)
}
