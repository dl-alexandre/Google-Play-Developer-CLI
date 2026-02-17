package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) vitalsAnomaliesList(ctx context.Context, metric, timePeriod, minSeverity string, pageSize int64, pageToken string, all bool) error {
	if err := c.requirePackage(); err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("playdeveloperreporting")
		return c.Output(result)
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

	startToken := pageToken
	nextToken := ""
	var allAnomalies []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly
	for {
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
		allAnomalies = append(allAnomalies, anomalies...)
		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		req = req.PageToken(nextToken)
	}
	result := output.NewResult(map[string]interface{}{
		"anomalies":     allAnomalies,
		"metric":        metric,
		"timePeriod":    timePeriod,
		"nextPageToken": nextToken,
		"package":       c.packageName,
	})
	if minSeverity != "" {
		result.WithWarnings("min-severity filtering is not supported by the API")
	}
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func buildAnomalyFilter(timePeriod string) string {
	now := time.Now().UTC()
	switch timePeriod {
	case timePeriodLast7Days:
		return fmt.Sprintf("activeBetween(%q, %q)", now.AddDate(0, 0, -7).Format(time.RFC3339), now.Format(time.RFC3339))
	case timePeriodLast30Days:
		return fmt.Sprintf("activeBetween(%q, %q)", now.AddDate(0, 0, -30).Format(time.RFC3339), now.Format(time.RFC3339))
	case timePeriodLast90Days:
		return fmt.Sprintf("activeBetween(%q, %q)", now.AddDate(0, 0, -90).Format(time.RFC3339), now.Format(time.RFC3339))
	case "", "all":
		return ""
	default:
		return ""
	}
}

func (c *CLI) vitalsErrorsIssuesSearch(ctx context.Context, query, interval string, pageSize int64, pageToken string, all bool) error {
	if err := c.requirePackage(); err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("playdeveloperreporting")
		return c.Output(result)
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

	if query != "" {
		searchCall = searchCall.Filter(query)
	}

	if interval != "" {
		now := time.Now().UTC()
		var startTime time.Time
		switch interval {
		case timePeriodLast7Days:
			startTime = now.AddDate(0, 0, -7)
		case timePeriodLast30Days:
			startTime = now.AddDate(0, 0, -30)
		case timePeriodLast90Days:
			startTime = now.AddDate(0, 0, -90)
		default:
			startTime = now.AddDate(0, 0, -30)
		}
		searchCall = searchCall.IntervalStartTimeYear(int64(startTime.Year())).
			IntervalStartTimeMonth(int64(startTime.Month())).
			IntervalStartTimeDay(int64(startTime.Day())).
			IntervalEndTimeYear(int64(now.Year())).
			IntervalEndTimeMonth(int64(now.Month())).
			IntervalEndTimeDay(int64(now.Day()))
	}

	if pageSize > 0 {
		searchCall = searchCall.PageSize(pageSize)
	}
	if pageToken != "" {
		searchCall = searchCall.PageToken(pageToken)
	}

	startToken := pageToken
	nextToken := ""
	var allIssues []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue
	for {
		resp, err := searchCall.Context(ctx).Do()
		if err != nil {
			return c.outputReportingQueryError(err, "failed to search error issues")
		}
		allIssues = append(allIssues, resp.ErrorIssues...)
		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		searchCall = searchCall.PageToken(nextToken)
	}

	result := output.NewResult(map[string]interface{}{
		"query":         query,
		"interval":      interval,
		"package":       c.packageName,
		"issues":        allIssues,
		"rowCount":      len(allIssues),
		"nextPageToken": nextToken,
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsErrorsReportsSearch(ctx context.Context, query, interval string, pageSize int64, pageToken string, all, formatReport bool) error {
	if err := c.requirePackage(); err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("playdeveloperreporting")
		return c.Output(result)
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

	if query != "" {
		searchCall = searchCall.Filter(query)
	}

	if interval != "" {
		now := time.Now().UTC()
		var startTime time.Time
		switch interval {
		case timePeriodLast7Days:
			startTime = now.AddDate(0, 0, -7)
		case timePeriodLast30Days:
			startTime = now.AddDate(0, 0, -30)
		case timePeriodLast90Days:
			startTime = now.AddDate(0, 0, -90)
		default:
			startTime = now.AddDate(0, 0, -30)
		}
		searchCall = searchCall.IntervalStartTimeYear(int64(startTime.Year())).
			IntervalStartTimeMonth(int64(startTime.Month())).
			IntervalStartTimeDay(int64(startTime.Day())).
			IntervalEndTimeYear(int64(now.Year())).
			IntervalEndTimeMonth(int64(now.Month())).
			IntervalEndTimeDay(int64(now.Day()))
	}

	if pageSize > 0 {
		searchCall = searchCall.PageSize(pageSize)
	}
	if pageToken != "" {
		searchCall = searchCall.PageToken(pageToken)
	}

	startToken := pageToken
	nextToken := ""
	var allReports []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport
	for {
		resp, err := searchCall.Context(ctx).Do()
		if err != nil {
			return c.outputReportingQueryError(err, "failed to search error reports")
		}
		if formatReport {
			for _, report := range resp.ErrorReports {
				if report != nil && report.ReportText != "" {
					report.ReportText = formatReportText(report.ReportText)
				}
			}
		}
		allReports = append(allReports, resp.ErrorReports...)
		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		searchCall = searchCall.PageToken(nextToken)
	}

	result := output.NewResult(map[string]interface{}{
		"query":         query,
		"interval":      interval,
		"package":       c.packageName,
		"reports":       allReports,
		"rowCount":      len(allReports),
		"nextPageToken": nextToken,
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsErrorsCountsGet(ctx context.Context) error {
	if err := c.requirePackage(); err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("playdeveloperreporting")
		return c.Output(result)
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
		return c.outputReportingQueryError(err, "failed to get error counts")
	}

	result := output.NewResult(map[string]interface{}{
		"package": c.packageName,
		"counts":  resp,
	})
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsErrorsCountsQuery(ctx context.Context, startDate, endDate string, dimensions []string, pageSize int64, pageToken string, all bool) error {
	if apiErr := validateReportingDates(startDate, endDate); apiErr != nil {
		result := output.NewErrorResult(apiErr).WithServices("playdeveloperreporting")
		return c.Output(result)
	}
	if err := c.requirePackage(); err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("playdeveloperreporting")
		return c.Output(result)
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

	startToken := pageToken
	nextToken := ""
	var rows []map[string]interface{}
	for {
		resp, err := reporting.Vitals.Errors.Counts.Query(appName, req).Context(ctx).Do()
		if err != nil {
			return c.outputReportingQueryError(err, "failed to query error counts")
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
			rows = append(rows, rowData)
		}
		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		req.PageToken = nextToken
	}

	result := output.NewResult(map[string]interface{}{
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          rows,
		"rowCount":      len(rows),
		"nextPageToken": nextToken,
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}
