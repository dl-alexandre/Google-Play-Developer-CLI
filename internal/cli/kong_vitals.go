package cli

import (
	"github.com/dl-alexandre/gpd/internal/errors"
)

// VitalsCrashesCmd queries crash rate data.
type VitalsCrashesCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the crashes command.
func (cmd *VitalsCrashesCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsAnrsCmd queries ANR rate data.
type VitalsAnrsCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the anrs command.
func (cmd *VitalsAnrsCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsErrorsCmd contains error commands.
type VitalsErrorsCmd struct {
	Issues  VitalsErrorsIssuesCmd  `cmd:"" help:"Search error issues"`
	Reports VitalsErrorsReportsCmd `cmd:"" help:"Search error reports"`
	Counts  VitalsErrorsCountsCmd  `cmd:"" help:"Error count metrics"`
}

// VitalsErrorsIssuesCmd searches error issues.
type VitalsErrorsIssuesCmd struct {
	Query     string `help:"Search query"`
	Interval  string `help:"Time interval: last7Days, last30Days, last90Days" default:"last30Days"`
	PageSize  int64  `help:"Results per page" default:"50"`
	PageToken string `help:"Pagination token"`
	All       bool   `help:"Fetch all pages"`
}

// Run executes the errors issues search command.
func (cmd *VitalsErrorsIssuesCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsErrorsReportsCmd searches error reports.
type VitalsErrorsReportsCmd struct {
	Query       string `help:"Search query"`
	Interval    string `help:"Time interval: last7Days, last30Days, last90Days" default:"last30Days"`
	PageSize    int64  `help:"Results per page" default:"50"`
	PageToken   string `help:"Pagination token"`
	All         bool   `help:"Fetch all pages"`
	Deobfuscate bool   `help:"Format report text for readability"`
}

// Run executes the errors reports search command.
func (cmd *VitalsErrorsReportsCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsErrorsCountsCmd contains error counts commands.
type VitalsErrorsCountsCmd struct {
	Get   VitalsErrorsCountsGetCmd   `cmd:"" help:"Get error count metrics"`
	Query VitalsErrorsCountsQueryCmd `cmd:"" help:"Query error counts over time"`
}

// VitalsErrorsCountsGetCmd gets error count metrics.
type VitalsErrorsCountsGetCmd struct{}

// Run executes the errors counts get command.
func (cmd *VitalsErrorsCountsGetCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsErrorsCountsQueryCmd queries error counts over time.
type VitalsErrorsCountsQueryCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the errors counts query command.
func (cmd *VitalsErrorsCountsQueryCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsMetricsCmd contains performance metrics commands.
type VitalsMetricsCmd struct {
	ExcessiveWakeups VitalsMetricsExcessiveWakeupsCmd `cmd:"" help:"Query excessive wakeups data"`
	LmkRate          VitalsMetricsLmkRateCmd          `cmd:"" help:"Query LMK rate data"`
	SlowRendering    VitalsMetricsSlowRenderingCmd    `cmd:"" help:"Query slow rendering data"`
	SlowStart        VitalsMetricsSlowStartCmd        `cmd:"" help:"Query slow start data"`
	StuckWakelocks   VitalsMetricsStuckWakelocksCmd   `cmd:"" help:"Query stuck wakelocks data"`
}

// VitalsMetricsExcessiveWakeupsCmd queries excessive wakeups data.
type VitalsMetricsExcessiveWakeupsCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the excessive-wakeups command.
func (cmd *VitalsMetricsExcessiveWakeupsCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsMetricsLmkRateCmd queries LMK rate data.
type VitalsMetricsLmkRateCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the lmk-rate command.
func (cmd *VitalsMetricsLmkRateCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsMetricsSlowRenderingCmd queries slow rendering data.
type VitalsMetricsSlowRenderingCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the slow-rendering command.
func (cmd *VitalsMetricsSlowRenderingCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	return errors.NewAPIError(errors.CodeGeneralError, "not yet implemented")
}

// VitalsMetricsSlowStartCmd queries slow start data.
type VitalsMetricsSlowStartCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the slow-start command.
func (cmd *VitalsMetricsSlowStartCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals slow-start not yet implemented")
}

// VitalsMetricsStuckWakelocksCmd queries stuck wakelocks data.
type VitalsMetricsStuckWakelocksCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the stuck-wakelocks command.
func (cmd *VitalsMetricsStuckWakelocksCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals stuck-wakelocks not yet implemented")
}

// VitalsAnomaliesCmd contains anomaly commands.
type VitalsAnomaliesCmd struct {
	List VitalsAnomaliesListCmd `cmd:"" help:"List anomalies"`
}

// VitalsAnomaliesListCmd lists anomalies.
type VitalsAnomaliesListCmd struct {
	Metric      string `help:"Metric name filter"`
	TimePeriod  string `help:"Time period: last7Days, last30Days, last90Days" default:"last30Days"`
	MinSeverity string `help:"Minimum severity filter"`
	PageSize    int64  `help:"Results per page" default:"20"`
	PageToken   string `help:"Pagination token"`
	All         bool   `help:"Fetch all pages"`
}

// Run executes the anomalies list command.
func (cmd *VitalsAnomaliesListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals anomalies list not yet implemented")
}

// VitalsQueryCmd queries vitals metrics.
type VitalsQueryCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Metrics    []string `help:"Metrics to retrieve" default:"crashRate"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the query command.
func (cmd *VitalsQueryCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals query not yet implemented")
}

// VitalsCapabilitiesCmd lists available vitals metrics.
type VitalsCapabilitiesCmd struct{}

// Run executes the capabilities command.
func (cmd *VitalsCapabilitiesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals capabilities not yet implemented")
}

// Ensure the KongCLI Vitals field is properly typed
type VitalsCmd struct {
	Crashes      VitalsCrashesCmd      `cmd:"" help:"Query crash rate data"`
	Anrs         VitalsAnrsCmd         `cmd:"" help:"Query ANR rate data"`
	Errors       VitalsErrorsCmd       `cmd:"" help:"Search and report errors"`
	Metrics      VitalsMetricsCmd      `cmd:"" help:"Query performance metrics"`
	Anomalies    VitalsAnomaliesCmd    `cmd:"" help:"Anomalies in vitals metrics"`
	Query        VitalsQueryCmd        `cmd:"" help:"Query vitals metrics"`
	Capabilities VitalsCapabilitiesCmd `cmd:"" help:"List available vitals metrics"`
}
