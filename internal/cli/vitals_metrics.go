package cli

import (
	"context"
	"fmt"

	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) vitalsCrashes(ctx context.Context, startDate, endDate string, dimensions []string,
	_ string, pageSize int64, pageToken string, all bool) error {
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

	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}
	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	startToken := pageToken
	nextToken := ""
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

		nextToken = resp.NextPageToken
		if !all || nextToken == "" {
			break
		}
		queryReq.PageToken = nextToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "crashRate",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": nextToken,
		"dataFreshness": map[string]interface{}{
			"note": "Vitals data may be delayed by 24-48 hours",
		},
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsANRs(ctx context.Context, startDate, endDate string, dimensions []string,
	_ string, pageSize int64, pageToken string, all bool) error {
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

	queryReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		PageSize:     pageSize,
	}
	if len(dimensions) > 0 {
		queryReq.Dimensions = dimensions
	}
	if pageToken != "" {
		queryReq.PageToken = pageToken
	}

	startToken := pageToken
	nextToken := ""
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

		nextToken = resp.NextPageToken
		if !all || nextToken == "" {
			break
		}
		queryReq.PageToken = nextToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "anrRate",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": nextToken,
		"dataFreshness": map[string]interface{}{
			"note": "Vitals data may be delayed by 24-48 hours",
		},
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsExcessiveWakeups(ctx context.Context, startDate, endDate string, dimensions []string,
	_ string, pageSize int64, pageToken string, all bool) error {
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

	startToken := pageToken
	nextToken := ""
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

		nextToken = resp.NextPageToken
		if !all || nextToken == "" {
			break
		}
		queryReq.PageToken = nextToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "excessiveWakeups",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": nextToken,
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsLmkRate(_ context.Context, _, _ string, _ []string,
	_ string, _ int64, _ string, _ bool) error {
	return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
		"LMK rate metric is not available in the Play Developer Reporting API. "+
			"Please use other available metrics such as crashRate, anrRate, excessiveWakeups, etc."))
}

func (c *CLI) vitalsSlowRendering(ctx context.Context, startDate, endDate string, dimensions []string,
	_ string, pageSize int64, pageToken string, all bool) error {
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

	startToken := pageToken
	nextToken := ""
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

		nextToken = resp.NextPageToken
		if !all || nextToken == "" {
			break
		}
		queryReq.PageToken = nextToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "slowRendering",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": nextToken,
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsSlowStart(ctx context.Context, startDate, endDate string, dimensions []string,
	_ string, pageSize int64, pageToken string, all bool) error {
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

	startToken := pageToken
	nextToken := ""
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

		nextToken = resp.NextPageToken
		if !all || nextToken == "" {
			break
		}
		queryReq.PageToken = nextToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "slowStart",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": nextToken,
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) vitalsStuckWakelocks(ctx context.Context, startDate, endDate string, dimensions []string,
	_ string, pageSize int64, pageToken string, all bool) error {
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

	startToken := pageToken
	nextToken := ""
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

		nextToken = resp.NextPageToken
		if !all || nextToken == "" {
			break
		}
		queryReq.PageToken = nextToken
	}

	result := output.NewResult(map[string]interface{}{
		"metric":        "stuckWakelocks",
		"startDate":     startDate,
		"endDate":       endDate,
		"dimensions":    dimensions,
		"package":       c.packageName,
		"rows":          allRows,
		"rowCount":      len(allRows),
		"nextPageToken": nextToken,
	})
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}
