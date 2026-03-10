package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

const (
	formatCSV = "csv"

	metricSetCrashRate               = "crashRateMetricSet"
	metricSetAnrRate                 = "anrRateMetricSet"
	metricSetSlowRendering           = "slowRenderingRateMetricSet"
	metricSetSlowStart               = "slowStartRateMetricSet"
	metricSetStuckBackgroundWakelock = "stuckBackgroundWakelockRateMetricSet"
	metricSetExcessiveWakeup         = "excessiveWakeupRateMetricSet"
	metricSetErrorCount              = "errorCountMetricSet"
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

	name := fmt.Sprintf("apps/%s/crashRateMetricSet", globals.Package)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics: []string{
			"crashRate",
			"crashRate7dUserWeighted",
			"crashRate28dUserWeighted",
			"userPerceivedCrashRate",
			"userPerceivedCrashRate7dUserWeighted",
			"userPerceivedCrashRate28dUserWeighted",
			"distinctUsers",
		},
		PageSize:  cmd.PageSize,
		PageToken: cmd.PageToken,
	}

	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Crashrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}

		allRows = append(allRows, resp.Rows...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (crashRatePageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, err := svc.Vitals.Crashrate.Query(name, pageReq).Context(ctx).Do()
				return crashRatePageResponse{resp: pageResp}, err
			}
			additionalRows, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allRows = append(allRows, additionalRows...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query crash rate: %v", err))
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// crashRatePageResponse wraps the crash rate query response to implement PageResponse.
type crashRatePageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetResponse
}

func (r crashRatePageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r crashRatePageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow {
	return r.resp.Rows
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

	name := fmt.Sprintf("apps/%s/anrRateMetricSet", globals.Package)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics: []string{
			"anrRate",
			"anrRate7dUserWeighted",
			"anrRate28dUserWeighted",
			"userPerceivedAnrRate",
			"userPerceivedAnrRate7dUserWeighted",
			"userPerceivedAnrRate28dUserWeighted",
			"distinctUsers",
		},
		PageSize:  cmd.PageSize,
		PageToken: cmd.PageToken,
	}

	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Anrrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}

		allRows = append(allRows, resp.Rows...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (anrRatePageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, err := svc.Vitals.Anrrate.Query(name, pageReq).Context(ctx).Do()
				return anrRatePageResponse{resp: pageResp}, err
			}
			additionalRows, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allRows = append(allRows, additionalRows...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query ANR rate: %v", err))
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// anrRatePageResponse wraps the ANR rate query response to implement PageResponse.
type anrRatePageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetResponse
}

func (r anrRatePageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r anrRatePageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow {
	return r.resp.Rows
}

// VitalsErrorsCmd contains error commands.
type VitalsErrorsCmd struct {
	Issues  VitalsErrorsIssuesCmd  `cmd:"" help:"Search error issues"`
	Reports VitalsErrorsReportsCmd `cmd:"" help:"Search error reports"`
	Counts  VitalsErrorsCountsCmd  `cmd:"" help:"Error count metrics"`
}

// VitalsErrorsIssuesCmd searches error issues.
type VitalsErrorsIssuesCmd struct {
	Query                  string `help:"Search query"`
	Interval               string `help:"Time interval: last7Days, last30Days, last90Days" default:"last30Days"`
	PageSize               int64  `help:"Results per page" default:"50"`
	PageToken              string `help:"Pagination token"`
	All                    bool   `help:"Fetch all pages"`
	SampleErrorReportLimit int64  `help:"Limit for sample error reports per issue"`
}

// Run executes the errors issues search command.
func (cmd *VitalsErrorsIssuesCmd) Run(globals *Globals) error {
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

	parent := fmt.Sprintf("apps/%s/errorIssues", globals.Package)

	filter := cmd.buildFilter()
	startTime := time.Now()

	var allIssues []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Vitals.Errors.Issues.Search(parent).Context(ctx).
			Filter(filter).
			PageSize(cmd.PageSize).
			PageToken(cmd.PageToken)

		if cmd.SampleErrorReportLimit > 0 {
			call = call.SampleErrorReportLimit(cmd.SampleErrorReportLimit)
		}

		resp, err := call.Do()
		if err != nil {
			return err
		}

		allIssues = append(allIssues, resp.ErrorIssues...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (errorIssuesPageResponse, error) {
				pageCall := svc.Vitals.Errors.Issues.Search(parent).Context(ctx).
					Filter(filter).
					PageSize(cmd.PageSize).
					PageToken(pageToken)
				pageResp, err := pageCall.Do()
				return errorIssuesPageResponse{resp: pageResp}, err
			}
			additionalIssues, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allIssues = append(allIssues, additionalIssues...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to search error issues: %v", err))
	}

	data := formatErrorIssues(allIssues)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResult(result, globals.Output, globals.Pretty)
}

func (cmd *VitalsErrorsIssuesCmd) buildFilter() string {
	var filters []string
	if cmd.Query != "" {
		filters = append(filters, fmt.Sprintf("(cause =~ %q OR location =~ %q)", cmd.Query, cmd.Query))
	}
	if cmd.Interval != "" {
		filters = append(filters, fmt.Sprintf("activeBetween(%s)", cmd.intervalToDateRange()))
	}
	return strings.Join(filters, " AND ")
}

func (cmd *VitalsErrorsIssuesCmd) intervalToDateRange() string {
	endDate := time.Now().UTC().Format("2006-01-02")
	var startDate string

	switch cmd.Interval {
	case "last7Days":
		startDate = time.Now().UTC().AddDate(0, 0, -7).Format("2006-01-02")
	case "last30Days":
		startDate = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	case "last90Days":
		startDate = time.Now().UTC().AddDate(0, 0, -90).Format("2006-01-02")
	default:
		startDate = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
	}

	return fmt.Sprintf("\"%sT00:00:00Z\",\"%sT00:00:00Z\"", startDate, endDate)
}

// errorIssuesPageResponse wraps the error issues search response to implement PageResponse.
type errorIssuesPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SearchErrorIssuesResponse
}

func (r errorIssuesPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r errorIssuesPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue {
	return r.resp.ErrorIssues
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

	parent := fmt.Sprintf("apps/%s", globals.Package)

	startTime := time.Now()

	var allReports []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Vitals.Errors.Reports.Search(parent).Context(ctx).
			PageSize(cmd.PageSize).
			PageToken(cmd.PageToken)

		// Set interval parameters
		startDate, endDate := cmd.getIntervalDates()
		call = cmd.setIntervalParams(call, startDate, endDate)

		if cmd.Query != "" {
			call = call.Filter(cmd.Query)
		}

		resp, err := call.Do()
		if err != nil {
			return err
		}

		allReports = append(allReports, resp.ErrorReports...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (errorReportsPageResponse, error) {
				pageCall := svc.Vitals.Errors.Reports.Search(parent).Context(ctx).
					PageSize(cmd.PageSize).
					PageToken(pageToken)
				pageCall = cmd.setIntervalParams(pageCall, startDate, endDate)
				pageResp, err := pageCall.Do()
				return errorReportsPageResponse{resp: pageResp}, err
			}
			additionalReports, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allReports = append(allReports, additionalReports...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to search error reports: %v", err))
	}

	data := formatErrorReports(allReports)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResult(result, globals.Output, globals.Pretty)
}

func (cmd *VitalsErrorsReportsCmd) getIntervalDates() (startDate, endDate time.Time) {
	endDate = time.Now().UTC()

	switch cmd.Interval {
	case "last7Days":
		startDate = endDate.AddDate(0, 0, -7)
	case "last90Days":
		startDate = endDate.AddDate(0, 0, -90)
	default:
		startDate = endDate.AddDate(0, 0, -30)
	}

	return startDate, endDate
}

func (cmd *VitalsErrorsReportsCmd) setIntervalParams(call *playdeveloperreporting.VitalsErrorsReportsSearchCall, startDate, endDate time.Time) *playdeveloperreporting.VitalsErrorsReportsSearchCall {
	return call.
		IntervalStartTimeYear(int64(startDate.Year())).
		IntervalStartTimeMonth(int64(startDate.Month())).
		IntervalStartTimeDay(int64(startDate.Day())).
		IntervalStartTimeTimeZoneId("UTC").
		IntervalEndTimeYear(int64(endDate.Year())).
		IntervalEndTimeMonth(int64(endDate.Month())).
		IntervalEndTimeDay(int64(endDate.Day())).
		IntervalEndTimeTimeZoneId("UTC")
}

// errorReportsPageResponse wraps the error reports search response to implement PageResponse.
type errorReportsPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SearchErrorReportsResponse
}

func (r errorReportsPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r errorReportsPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport {
	return r.resp.ErrorReports
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

	name := fmt.Sprintf("apps/%s/errorCountMetricSet", globals.Package)
	startTime := time.Now()

	var metricSet *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorCountMetricSet
	err = client.DoWithRetry(ctx, func() error {
		var err error
		metricSet, err = svc.Vitals.Errors.Counts.Get(name).Context(ctx).Do()
		return err
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get error count metrics: %v", err))
	}

	data := map[string]interface{}{
		"name":          metricSet.Name,
		"freshnessInfo": metricSet.FreshnessInfo,
	}

	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResult(result, globals.Output, globals.Pretty)
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

	name := fmt.Sprintf("apps/%s/errorCountMetricSet", globals.Package)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics: []string{
			"errorCount",
			"distinctUsers",
			"errorReportCount",
		},
		PageSize:  cmd.PageSize,
		PageToken: cmd.PageToken,
	}

	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Errors.Counts.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}

		allRows = append(allRows, resp.Rows...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (errorCountPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, err := svc.Vitals.Errors.Counts.Query(name, pageReq).Context(ctx).Do()
				return errorCountPageResponse{resp: pageResp}, err
			}
			additionalRows, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allRows = append(allRows, additionalRows...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query error counts: %v", err))
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResult(result, globals.Output, globals.Pretty)
}

// errorCountPageResponse wraps the error count query response to implement PageResponse.
type errorCountPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetResponse
}

func (r errorCountPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r errorCountPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow {
	return r.resp.Rows
}

// VitalsMetricsCmd contains performance metrics commands.
type VitalsMetricsCmd struct {
	ExcessiveWakeups VitalsMetricsExcessiveWakeupsCmd `cmd:"" help:"Query excessive wakeups data"`
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

	name := fmt.Sprintf("apps/%s/excessiveWakeupRateMetricSet", globals.Package)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics: []string{
			"excessiveWakeupRate",
			"excessiveWakeupRate7dUserWeighted",
			"excessiveWakeupRate28dUserWeighted",
			"distinctUsers",
		},
		PageSize:  cmd.PageSize,
		PageToken: cmd.PageToken,
	}

	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Excessivewakeuprate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}

		allRows = append(allRows, resp.Rows...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (excessiveWakeupPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, err := svc.Vitals.Excessivewakeuprate.Query(name, pageReq).Context(ctx).Do()
				return excessiveWakeupPageResponse{resp: pageResp}, err
			}
			additionalRows, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allRows = append(allRows, additionalRows...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query excessive wakeup rate: %v", err))
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// excessiveWakeupPageResponse wraps the excessive wakeup query response to implement PageResponse.
type excessiveWakeupPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetResponse
}

func (r excessiveWakeupPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r excessiveWakeupPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow {
	return r.resp.Rows
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

	name := fmt.Sprintf("apps/%s/slowRenderingRateMetricSet", globals.Package)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics: []string{
			"slowRenderingRate",
			"slowRenderingRate7dUserWeighted",
			"slowRenderingRate28dUserWeighted",
			"distinctUsers",
		},
		PageSize:  cmd.PageSize,
		PageToken: cmd.PageToken,
	}

	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Slowrenderingrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}

		allRows = append(allRows, resp.Rows...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (slowRenderingPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, err := svc.Vitals.Slowrenderingrate.Query(name, pageReq).Context(ctx).Do()
				return slowRenderingPageResponse{resp: pageResp}, err
			}
			additionalRows, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allRows = append(allRows, additionalRows...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query slow rendering rate: %v", err))
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// slowRenderingPageResponse wraps the slow rendering query response to implement PageResponse.
type slowRenderingPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetResponse
}

func (r slowRenderingPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r slowRenderingPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow {
	return r.resp.Rows
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

	name := fmt.Sprintf("apps/%s/slowStartRateMetricSet", globals.Package)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics: []string{
			"slowStartRate",
			"slowStartRate7dUserWeighted",
			"slowStartRate28dUserWeighted",
			"distinctUsers",
		},
		PageSize:  cmd.PageSize,
		PageToken: cmd.PageToken,
	}

	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Slowstartrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}

		allRows = append(allRows, resp.Rows...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (slowStartPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, err := svc.Vitals.Slowstartrate.Query(name, pageReq).Context(ctx).Do()
				return slowStartPageResponse{resp: pageResp}, err
			}
			additionalRows, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allRows = append(allRows, additionalRows...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query slow start rate: %v", err))
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// slowStartPageResponse wraps the slow start query response to implement PageResponse.
type slowStartPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetResponse
}

func (r slowStartPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r slowStartPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow {
	return r.resp.Rows
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

	name := fmt.Sprintf("apps/%s/stuckBackgroundWakelockRateMetricSet", globals.Package)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics: []string{
			"stuckBackgroundWakelockRate",
			"stuckBackgroundWakelockRate7dUserWeighted",
			"stuckBackgroundWakelockRate28dUserWeighted",
			"distinctUsers",
		},
		PageSize:  cmd.PageSize,
		PageToken: cmd.PageToken,
	}

	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		resp, err := svc.Vitals.Stuckbackgroundwakelockrate.Query(name, req).Context(ctx).Do()
		if err != nil {
			return err
		}

		allRows = append(allRows, resp.Rows...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (stuckWakelockPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, err := svc.Vitals.Stuckbackgroundwakelockrate.Query(name, pageReq).Context(ctx).Do()
				return stuckWakelockPageResponse{resp: pageResp}, err
			}
			additionalRows, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allRows = append(allRows, additionalRows...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query stuck wakelock rate: %v", err))
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// stuckWakelockPageResponse wraps the stuck wakelock query response to implement PageResponse.
type stuckWakelockPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetResponse
}

func (r stuckWakelockPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r stuckWakelockPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow {
	return r.resp.Rows
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

	parent := fmt.Sprintf("apps/%s", globals.Package)
	startTime := time.Now()

	var allAnomalies []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Anomalies.List(parent).Context(ctx).
			PageSize(cmd.PageSize).
			PageToken(cmd.PageToken)

		if cmd.Metric != "" {
			call = call.Filter(fmt.Sprintf("metric = %q", cmd.Metric))
		}

		resp, err := call.Do()
		if err != nil {
			return err
		}

		allAnomalies = append(allAnomalies, resp.Anomalies...)

		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (anomaliesPageResponse, error) {
				pageCall := svc.Anomalies.List(parent).Context(ctx).
					PageSize(cmd.PageSize).
					PageToken(pageToken)
				pageResp, err := pageCall.Do()
				return anomaliesPageResponse{resp: pageResp}, err
			}
			additionalAnomalies, _, err := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if err != nil {
				return err
			}
			allAnomalies = append(allAnomalies, additionalAnomalies...)
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list anomalies: %v", err))
	}

	data := formatAnomalies(allAnomalies)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResult(result, globals.Output, globals.Pretty)
}

// anomaliesPageResponse wraps the anomalies list response to implement PageResponse.
type anomaliesPageResponse struct {
	resp *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ListAnomaliesResponse
}

func (r anomaliesPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r anomaliesPageResponse) GetItems() []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly {
	return r.resp.Anomalies
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

	// Determine metric set from the first metric requested
	metricSetName := metricSetCrashRate
	if len(cmd.Metrics) > 0 {
		switch cmd.Metrics[0] {
		case metricCrashRate, "userPerceivedCrashRate":
			metricSetName = metricSetCrashRate
		case metricAnrRate, "userPerceivedAnrRate":
			metricSetName = metricSetAnrRate
		case "slowRenderingRate":
			metricSetName = metricSetSlowRendering
		case "slowStartRate":
			metricSetName = metricSetSlowStart
		case "stuckBackgroundWakelockRate":
			metricSetName = metricSetStuckBackgroundWakelock
		case "excessiveWakeupRate":
			metricSetName = metricSetExcessiveWakeup
		case "errorCount", "errorReportCount":
			metricSetName = metricSetErrorCount
		}
	}

	name := fmt.Sprintf("apps/%s/%s", globals.Package, metricSetName)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	startTime := time.Now()

	allRows, err := cmd.queryMetricSet(ctx, client, svc, name, metricSetName, timelineSpec)
	if err != nil {
		return err
	}

	data := formatMetricsRows(allRows)
	result := output.NewResult(data).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// queryMetricSet dispatches the vitals query to the appropriate metric set API endpoint.
func (cmd *VitalsQueryCmd) queryMetricSet(
	ctx context.Context,
	client *api.Client,
	svc *playdeveloperreporting.Service,
	name, metricSetName string,
	timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec,
) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	var err error

	switch metricSetName {
	case metricSetCrashRate:
		allRows, err = cmd.queryCrashRateMetrics(ctx, client, svc, name, timelineSpec)
	case metricSetAnrRate:
		allRows, err = cmd.queryAnrRateMetrics(ctx, client, svc, name, timelineSpec)
	case metricSetSlowRendering:
		allRows, err = cmd.querySlowRenderingMetrics(ctx, client, svc, name, timelineSpec)
	case metricSetSlowStart:
		allRows, err = cmd.querySlowStartMetrics(ctx, client, svc, name, timelineSpec)
	case metricSetStuckBackgroundWakelock:
		allRows, err = cmd.queryStuckWakelockMetrics(ctx, client, svc, name, timelineSpec)
	case metricSetExcessiveWakeup:
		allRows, err = cmd.queryExcessiveWakeupMetrics(ctx, client, svc, name, timelineSpec)
	case metricSetErrorCount:
		allRows, err = cmd.queryErrorCountMetrics(ctx, client, svc, name, timelineSpec)
	default:
		return nil, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("unsupported metric set: %s", metricSetName)).
			WithHint("Supported metric sets: crashRate, anrRate, slowRenderingRate, slowStartRate, stuckBackgroundWakelockRate, excessiveWakeupRate, errorCount")
	}

	if err != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query vitals: %v", err))
	}
	return allRows, nil
}

func (cmd *VitalsQueryCmd) queryCrashRateMetrics(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, name string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics:      cmd.Metrics,
		PageSize:     cmd.PageSize,
		PageToken:    cmd.PageToken,
	}
	err := client.DoWithRetry(ctx, func() error {
		resp, qErr := svc.Vitals.Crashrate.Query(name, req).Context(ctx).Do()
		if qErr != nil {
			return qErr
		}
		allRows = append(allRows, resp.Rows...)
		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (crashRatePageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, pErr := svc.Vitals.Crashrate.Query(name, pageReq).Context(ctx).Do()
				return crashRatePageResponse{resp: pageResp}, pErr
			}
			additionalRows, _, fErr := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if fErr != nil {
				return fErr
			}
			allRows = append(allRows, additionalRows...)
		}
		return nil
	})
	return allRows, err
}

func (cmd *VitalsQueryCmd) queryAnrRateMetrics(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, name string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics:      cmd.Metrics,
		PageSize:     cmd.PageSize,
		PageToken:    cmd.PageToken,
	}
	err := client.DoWithRetry(ctx, func() error {
		resp, qErr := svc.Vitals.Anrrate.Query(name, req).Context(ctx).Do()
		if qErr != nil {
			return qErr
		}
		allRows = append(allRows, resp.Rows...)
		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (anrRatePageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, pErr := svc.Vitals.Anrrate.Query(name, pageReq).Context(ctx).Do()
				return anrRatePageResponse{resp: pageResp}, pErr
			}
			additionalRows, _, fErr := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if fErr != nil {
				return fErr
			}
			allRows = append(allRows, additionalRows...)
		}
		return nil
	})
	return allRows, err
}

func (cmd *VitalsQueryCmd) querySlowRenderingMetrics(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, name string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics:      cmd.Metrics,
		PageSize:     cmd.PageSize,
		PageToken:    cmd.PageToken,
	}
	err := client.DoWithRetry(ctx, func() error {
		resp, qErr := svc.Vitals.Slowrenderingrate.Query(name, req).Context(ctx).Do()
		if qErr != nil {
			return qErr
		}
		allRows = append(allRows, resp.Rows...)
		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (slowRenderingPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, pErr := svc.Vitals.Slowrenderingrate.Query(name, pageReq).Context(ctx).Do()
				return slowRenderingPageResponse{resp: pageResp}, pErr
			}
			additionalRows, _, fErr := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if fErr != nil {
				return fErr
			}
			allRows = append(allRows, additionalRows...)
		}
		return nil
	})
	return allRows, err
}

func (cmd *VitalsQueryCmd) querySlowStartMetrics(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, name string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics:      cmd.Metrics,
		PageSize:     cmd.PageSize,
		PageToken:    cmd.PageToken,
	}
	err := client.DoWithRetry(ctx, func() error {
		resp, qErr := svc.Vitals.Slowstartrate.Query(name, req).Context(ctx).Do()
		if qErr != nil {
			return qErr
		}
		allRows = append(allRows, resp.Rows...)
		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (slowStartPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, pErr := svc.Vitals.Slowstartrate.Query(name, pageReq).Context(ctx).Do()
				return slowStartPageResponse{resp: pageResp}, pErr
			}
			additionalRows, _, fErr := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if fErr != nil {
				return fErr
			}
			allRows = append(allRows, additionalRows...)
		}
		return nil
	})
	return allRows, err
}

func (cmd *VitalsQueryCmd) queryStuckWakelockMetrics(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, name string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics:      cmd.Metrics,
		PageSize:     cmd.PageSize,
		PageToken:    cmd.PageToken,
	}
	err := client.DoWithRetry(ctx, func() error {
		resp, qErr := svc.Vitals.Stuckbackgroundwakelockrate.Query(name, req).Context(ctx).Do()
		if qErr != nil {
			return qErr
		}
		allRows = append(allRows, resp.Rows...)
		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (stuckWakelockPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, pErr := svc.Vitals.Stuckbackgroundwakelockrate.Query(name, pageReq).Context(ctx).Do()
				return stuckWakelockPageResponse{resp: pageResp}, pErr
			}
			additionalRows, _, fErr := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if fErr != nil {
				return fErr
			}
			allRows = append(allRows, additionalRows...)
		}
		return nil
	})
	return allRows, err
}

func (cmd *VitalsQueryCmd) queryExcessiveWakeupMetrics(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, name string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics:      cmd.Metrics,
		PageSize:     cmd.PageSize,
		PageToken:    cmd.PageToken,
	}
	err := client.DoWithRetry(ctx, func() error {
		resp, qErr := svc.Vitals.Excessivewakeuprate.Query(name, req).Context(ctx).Do()
		if qErr != nil {
			return qErr
		}
		allRows = append(allRows, resp.Rows...)
		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (excessiveWakeupPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, pErr := svc.Vitals.Excessivewakeuprate.Query(name, pageReq).Context(ctx).Do()
				return excessiveWakeupPageResponse{resp: pageResp}, pErr
			}
			additionalRows, _, fErr := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if fErr != nil {
				return fErr
			}
			allRows = append(allRows, additionalRows...)
		}
		return nil
	})
	return allRows, err
}

func (cmd *VitalsQueryCmd) queryErrorCountMetrics(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, name string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) ([]*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow, error) {
	var allRows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
	req := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
		TimelineSpec: timelineSpec,
		Dimensions:   cmd.Dimensions,
		Metrics:      cmd.Metrics,
		PageSize:     cmd.PageSize,
		PageToken:    cmd.PageToken,
	}
	err := client.DoWithRetry(ctx, func() error {
		resp, qErr := svc.Vitals.Errors.Counts.Query(name, req).Context(ctx).Do()
		if qErr != nil {
			return qErr
		}
		allRows = append(allRows, resp.Rows...)
		if cmd.All && resp.NextPageToken != "" {
			query := func(pageToken string) (errorCountPageResponse, error) {
				pageReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetRequest{
					TimelineSpec: req.TimelineSpec,
					Dimensions:   req.Dimensions,
					Metrics:      req.Metrics,
					PageSize:     req.PageSize,
					PageToken:    pageToken,
				}
				pageResp, pErr := svc.Vitals.Errors.Counts.Query(name, pageReq).Context(ctx).Do()
				return errorCountPageResponse{resp: pageResp}, pErr
			}
			additionalRows, _, fErr := fetchAllPages(ctx, query, resp.NextPageToken, 0)
			if fErr != nil {
				return fErr
			}
			allRows = append(allRows, additionalRows...)
		}
		return nil
	})
	return allRows, err
}

// VitalsCapabilitiesCmd lists available vitals metrics.
type VitalsCapabilitiesCmd struct{}

// Run executes the capabilities command.
func (cmd *VitalsCapabilitiesCmd) Run(globals *Globals) error {
	startTime := time.Now()

	capabilities := map[string]interface{}{
		"metricSets": []map[string]interface{}{
			{
				"name":        metricSetCrashRate,
				"description": "Crash rate metrics including user-perceived crash rates",
				"command":     "gpd vitals crashes",
				"freshness":   "Data available within 24 hours",
				"metrics": []string{
					"crashRate", "crashRate7dUserWeighted", "crashRate28dUserWeighted",
					"userPerceivedCrashRate", "userPerceivedCrashRate7dUserWeighted",
					"userPerceivedCrashRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        metricSetAnrRate,
				"description": "ANR (Application Not Responding) rate metrics",
				"command":     "gpd vitals anrs",
				"freshness":   "Data available within 24 hours",
				"metrics": []string{
					"anrRate", "anrRate7dUserWeighted", "anrRate28dUserWeighted",
					"userPerceivedAnrRate", "userPerceivedAnrRate7dUserWeighted",
					"userPerceivedAnrRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        metricSetSlowRendering,
				"description": "Slow rendering (UI jank) rate metrics",
				"command":     "gpd vitals metrics slow-rendering",
				"freshness":   "Data available within 24 hours",
				"metrics": []string{
					"slowRenderingRate", "slowRenderingRate7dUserWeighted",
					"slowRenderingRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        metricSetSlowStart,
				"description": "Slow app startup rate metrics",
				"command":     "gpd vitals metrics slow-start",
				"freshness":   "Data available within 24 hours",
				"metrics": []string{
					"slowStartRate", "slowStartRate7dUserWeighted",
					"slowStartRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        metricSetStuckBackgroundWakelock,
				"description": "Stuck background wakelock rate metrics",
				"command":     "gpd vitals metrics stuck-wakelocks",
				"freshness":   "Data available within 24 hours",
				"metrics": []string{
					"stuckBackgroundWakelockRate", "stuckBackgroundWakelockRate7dUserWeighted",
					"stuckBackgroundWakelockRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        metricSetExcessiveWakeup,
				"description": "Excessive wakeup rate metrics",
				"command":     "gpd vitals metrics excessive-wakeups",
				"freshness":   "Data available within 24 hours",
				"metrics": []string{
					"excessiveWakeupRate", "excessiveWakeupRate7dUserWeighted",
					"excessiveWakeupRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        metricSetErrorCount,
				"description": "Error count metrics for crashes and ANRs",
				"command":     "gpd vitals errors counts",
				"freshness":   "Data available within 24 hours",
				"metrics": []string{
					"errorReportCount", "distinctUsers",
				},
			},
		},
		"dimensions": []string{
			"apiLevel", "versionCode", "deviceModel", "deviceBrand",
			"deviceType", "countryCode", "deviceRamBucket",
			"deviceSocMake", "deviceSocModel", "deviceCpuMake",
			"deviceCpuModel", "deviceGpuMake", "deviceGpuModel",
			"deviceGpuVersion", "deviceVulkanVersion", "deviceGlEsVersion",
			"deviceScreenSize", "deviceScreenDpi",
		},
		"aggregationPeriods": []string{
			"DAILY", "HOURLY",
		},
		"additionalFeatures": []map[string]interface{}{
			{
				"name":        "anomalies",
				"description": "Detect anomalies in vitals metrics",
				"command":     "gpd vitals anomalies list",
			},
			{
				"name":        "errorIssues",
				"description": "Search and list error issues",
				"command":     "gpd vitals errors issues",
			},
			{
				"name":        "errorReports",
				"description": "View individual error reports",
				"command":     "gpd vitals errors reports",
			},
		},
		"notes": []string{
			"Use 'gpd vitals query --metrics <metric>' for generic vitals queries",
			"Dedicated subcommands (crashes, anrs, etc.) provide pre-configured metric lists",
			"All metric sets support dimension-based grouping via --dimensions",
			"Pagination is supported via --page-size, --page-token, and --all flags",
			"Data is typically available within 24 hours of collection",
		},
	}

	result := output.NewResult(capabilities).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResult(result, globals.Output, globals.Pretty)
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

// Helper functions

func buildTimelineSpec(startDate, endDate string) (*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec, error) {
	var start, end time.Time
	var err error

	if startDate == "" {
		start = time.Now().UTC().AddDate(0, 0, -30)
	} else {
		start, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid start date: %v", err))
		}
	}

	if endDate == "" {
		end = time.Now().UTC()
	} else {
		end, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid end date: %v", err))
		}
	}

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
	}, nil
}

func formatMetricsRows(rows []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		data := map[string]interface{}{
			"aggregationPeriod": row.AggregationPeriod,
			"startTime":         row.StartTime,
		}

		dimensions := make(map[string]interface{})
		for _, dim := range row.Dimensions {
			if dim.StringValue != "" {
				dimensions[dim.Dimension] = dim.StringValue
			} else {
				dimensions[dim.Dimension] = dim.Int64Value
			}
		}
		data["dimensions"] = dimensions

		metrics := make(map[string]interface{})
		for _, metric := range row.Metrics {
			if metric.DecimalValue != nil {
				metrics[metric.Metric] = metric.DecimalValue.Value
			}
		}
		data["metrics"] = metrics

		result = append(result, data)
	}
	return result
}

func formatErrorIssues(issues []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(issues))
	for _, issue := range issues {
		data := map[string]interface{}{
			"name":             issue.Name,
			"cause":            issue.Cause,
			"type":             issue.Type,
			"location":         issue.Location,
			"distinctUsers":    issue.DistinctUsers,
			"errorReportCount": issue.ErrorReportCount,
			"issueUri":         issue.IssueUri,
		}
		result = append(result, data)
	}
	return result
}

func formatErrorReports(reports []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(reports))
	for _, report := range reports {
		data := map[string]interface{}{
			"name":       report.Name,
			"issue":      report.Issue,
			"eventTime":  report.EventTime,
			"appVersion": report.AppVersion,
			"osVersion":  report.OsVersion,
			"deviceModel": func() string {
				if report.DeviceModel != nil && report.DeviceModel.DeviceId != nil {
					return fmt.Sprintf("%s/%s", report.DeviceModel.DeviceId.BuildBrand, report.DeviceModel.DeviceId.BuildDevice)
				}
				return ""
			}(),
			"type":       report.Type,
			"reportText": report.ReportText,
		}
		result = append(result, data)
	}
	return result
}

func formatAnomalies(anomalies []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(anomalies))
	for _, anomaly := range anomalies {
		data := map[string]interface{}{
			"name":         anomaly.Name,
			"metric":       anomaly.Metric,
			"metricSet":    anomaly.MetricSet,
			"dimensions":   anomaly.Dimensions,
			"timelineSpec": anomaly.TimelineSpec,
		}
		result = append(result, data)
	}
	return result
}
