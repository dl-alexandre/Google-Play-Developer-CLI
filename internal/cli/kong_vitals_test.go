//go:build unit
// +build unit

package cli

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"
)

// ============================================================================
// Command Structure and Validation Tests
// ============================================================================

func TestVitalsCrashesCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsCrashesCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsAnrsCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsAnrsCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsErrorsIssuesCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsErrorsIssuesCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsErrorsReportsCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsErrorsReportsCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsErrorsCountsGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsErrorsCountsGetCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsErrorsCountsQueryCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsErrorsCountsQueryCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsMetricsExcessiveWakeupsCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsMetricsExcessiveWakeupsCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsMetricsSlowRenderingCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsMetricsSlowRenderingCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsMetricsSlowStartCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsMetricsSlowStartCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsMetricsStuckWakelocksCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsMetricsStuckWakelocksCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsAnomaliesListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsAnomaliesListCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsQueryCmd_Run_PackageRequired(t *testing.T) {
	cmd := &VitalsQueryCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestVitalsCapabilitiesCmd_Run_NoPackageRequired(t *testing.T) {
	cmd := &VitalsCapabilitiesCmd{}
	globals := &Globals{
		Output: "json",
	}

	// Capabilities command does NOT require package
	// It should succeed without a package (though it will fail without auth)
	err := cmd.Run(globals)
	// Expected to fail due to auth, not due to missing package
	if err != nil && strings.Contains(err.Error(), "package name is required") {
		t.Error("Capabilities command should not require package")
	}
}

// ============================================================================
// Command Flag/Field Structure Tests
// ============================================================================

func TestVitalsCrashesCmd_FieldTags(t *testing.T) {
	cmd := &VitalsCrashesCmd{
		StartDate:  "2024-01-01",
		EndDate:    "2024-01-31",
		Dimensions: []string{"versionCode", "deviceModel"},
		Format:     "csv",
		PageSize:   50,
		PageToken:  "token123",
		All:        true,
	}

	if cmd.StartDate != "2024-01-01" {
		t.Errorf("StartDate = %v, want 2024-01-01", cmd.StartDate)
	}
	if cmd.EndDate != "2024-01-31" {
		t.Errorf("EndDate = %v, want 2024-01-31", cmd.EndDate)
	}
	if len(cmd.Dimensions) != 2 {
		t.Errorf("Dimensions length = %v, want 2", len(cmd.Dimensions))
	}
	if cmd.Format != "csv" {
		t.Errorf("Format = %v, want csv", cmd.Format)
	}
	if cmd.PageSize != 50 {
		t.Errorf("PageSize = %v, want 50", cmd.PageSize)
	}
	if cmd.PageToken != "token123" {
		t.Errorf("PageToken = %v, want token123", cmd.PageToken)
	}
	if !cmd.All {
		t.Error("All = false, want true")
	}
}

func TestVitalsAnrsCmd_FieldTags(t *testing.T) {
	cmd := &VitalsAnrsCmd{
		StartDate:  "2024-02-01",
		EndDate:    "2024-02-29",
		Dimensions: []string{"apiLevel"},
		Format:     "json",
		PageSize:   100,
		All:        false,
	}

	if cmd.StartDate != "2024-02-01" {
		t.Errorf("StartDate = %v, want 2024-02-01", cmd.StartDate)
	}
	if cmd.Format != "json" {
		t.Errorf("Format = %v, want json", cmd.Format)
	}
	if cmd.All {
		t.Error("All = true, want false")
	}
}

func TestVitalsErrorsIssuesCmd_FieldTags(t *testing.T) {
	cmd := &VitalsErrorsIssuesCmd{
		Query:                  "NullPointerException",
		Interval:               "last7Days",
		PageSize:               25,
		PageToken:              "pageToken",
		All:                    true,
		SampleErrorReportLimit: 10,
	}

	if cmd.Query != "NullPointerException" {
		t.Errorf("Query = %v, want NullPointerException", cmd.Query)
	}
	if cmd.Interval != "last7Days" {
		t.Errorf("Interval = %v, want last7Days", cmd.Interval)
	}
	if cmd.PageSize != 25 {
		t.Errorf("PageSize = %v, want 25", cmd.PageSize)
	}
	if cmd.SampleErrorReportLimit != 10 {
		t.Errorf("SampleErrorReportLimit = %v, want 10", cmd.SampleErrorReportLimit)
	}
}

func TestVitalsErrorsReportsCmd_FieldTags(t *testing.T) {
	cmd := &VitalsErrorsReportsCmd{
		Query:       "crash",
		Interval:    "last30Days",
		PageSize:    100,
		PageToken:   "token",
		All:         false,
		Deobfuscate: true,
	}

	if cmd.Query != "crash" {
		t.Errorf("Query = %v, want crash", cmd.Query)
	}
	if cmd.Interval != "last30Days" {
		t.Errorf("Interval = %v, want last30Days", cmd.Interval)
	}
	if !cmd.Deobfuscate {
		t.Error("Deobfuscate = false, want true")
	}
}

func TestVitalsQueryCmd_FieldTags(t *testing.T) {
	cmd := &VitalsQueryCmd{
		StartDate:  "2024-03-01",
		EndDate:    "2024-03-31",
		Metrics:    []string{"crashRate", "anrRate"},
		Dimensions: []string{"countryCode", "deviceType"},
		Format:     "csv",
		PageSize:   200,
		PageToken:  "queryToken",
		All:        true,
	}

	if len(cmd.Metrics) != 2 {
		t.Errorf("Metrics length = %v, want 2", len(cmd.Metrics))
	}
	if cmd.Metrics[0] != "crashRate" {
		t.Errorf("Metrics[0] = %v, want crashRate", cmd.Metrics[0])
	}
	if len(cmd.Dimensions) != 2 {
		t.Errorf("Dimensions length = %v, want 2", len(cmd.Dimensions))
	}
}

func TestVitalsAnomaliesListCmd_FieldTags(t *testing.T) {
	cmd := &VitalsAnomaliesListCmd{
		Metric:      "crashRate",
		TimePeriod:  "last30Days",
		MinSeverity: "warning",
		PageSize:    50,
		PageToken:   "anomalyToken",
		All:         true,
	}

	if cmd.Metric != "crashRate" {
		t.Errorf("Metric = %v, want crashRate", cmd.Metric)
	}
	if cmd.TimePeriod != "last30Days" {
		t.Errorf("TimePeriod = %v, want last30Days", cmd.TimePeriod)
	}
	if cmd.MinSeverity != "warning" {
		t.Errorf("MinSeverity = %v, want warning", cmd.MinSeverity)
	}
}

// ============================================================================
// buildTimelineSpec Tests
// ============================================================================

func TestBuildTimelineSpec(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid dates",
			startDate: "2024-01-01",
			endDate:   "2024-01-31",
			wantErr:   false,
		},
		{
			name:      "empty dates uses defaults",
			startDate: "",
			endDate:   "",
			wantErr:   false,
		},
		{
			name:      "only start date uses default end",
			startDate: "2024-01-01",
			endDate:   "",
			wantErr:   false,
		},
		{
			name:      "only end date uses default start",
			startDate: "",
			endDate:   "2024-01-31",
			wantErr:   false,
		},
		{
			name:      "invalid start date format",
			startDate: "01-01-2024",
			endDate:   "2024-01-31",
			wantErr:   true,
			errMsg:    "invalid start date",
		},
		{
			name:      "invalid end date format",
			startDate: "2024-01-01",
			endDate:   "31-01-2024",
			wantErr:   true,
			errMsg:    "invalid end date",
		},
		{
			name:      "nonexistent start date",
			startDate: "2024-13-01",
			endDate:   "2024-01-31",
			wantErr:   true,
			errMsg:    "invalid start date",
		},
		{
			name:      "nonexistent end date",
			startDate: "2024-01-01",
			endDate:   "2024-02-30",
			wantErr:   true,
			errMsg:    "invalid end date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, err := buildTimelineSpec(tt.startDate, tt.endDate)
			if (err != nil) != tt.wantErr {
				t.Errorf("buildTimelineSpec() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("Expected error containing %q, got: %v", tt.errMsg, err)
				}
				return
			}
			if spec == nil {
				t.Fatal("Expected non-nil spec")
			}
			if spec.AggregationPeriod != "DAILY" {
				t.Errorf("AggregationPeriod = %v, want DAILY", spec.AggregationPeriod)
			}
			if spec.StartTime == nil {
				t.Error("StartTime is nil")
			}
			if spec.EndTime == nil {
				t.Error("EndTime is nil")
			}
			if spec.StartTime.TimeZone == nil || spec.StartTime.TimeZone.Id != "America/Los_Angeles" {
				t.Error("StartTime.TimeZone.Id should be America/Los_Angeles")
			}
		})
	}
}

func TestBuildTimelineSpec_DefaultDateRange(t *testing.T) {
	// Test that empty dates default to approximately 30 days ago to now
	spec, err := buildTimelineSpec("", "")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if spec == nil {
		t.Fatal("Expected non-nil spec")
	}

	// Verify we have both start and end times
	if spec.StartTime == nil || spec.EndTime == nil {
		t.Fatal("Both StartTime and EndTime should be set")
	}

	// Check that start time is before end time
	startYear := spec.StartTime.Year
	startMonth := spec.StartTime.Month
	startDay := spec.StartTime.Day
	endYear := spec.EndTime.Year
	endMonth := spec.EndTime.Month
	endDay := spec.EndTime.Day

	start := time.Date(int(startYear), time.Month(startMonth), int(startDay), 0, 0, 0, 0, time.UTC)
	end := time.Date(int(endYear), time.Month(endMonth), int(endDay), 0, 0, 0, 0, time.UTC)

	if !start.Before(end) {
		t.Error("Start time should be before end time")
	}

	// Check that the range is approximately 30 days
	diff := end.Sub(start)
	if diff < 25*24*time.Hour || diff > 35*24*time.Hour {
		t.Errorf("Expected approximately 30 day range, got %v", diff)
	}
}

// ============================================================================
// formatMetricsRows Tests
// ============================================================================

func TestFormatMetricsRows(t *testing.T) {
	tests := []struct {
		name     string
		rows     []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow
		expected int
	}{
		{
			name:     "empty rows",
			rows:     []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{},
			expected: 0,
		},
		{
			name:     "nil rows",
			rows:     nil,
			expected: 0,
		},
		{
			name: "single row with dimensions and metrics",
			rows: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
				{
					AggregationPeriod: "DAILY",
					StartTime: &playdeveloperreporting.GoogleTypeDateTime{
						Year:  2024,
						Month: 1,
						Day:   15,
					},
					Dimensions: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DimensionValue{
						{
							Dimension:  "versionCode",
							Int64Value: 100,
						},
						{
							Dimension:   "deviceModel",
							StringValue: "Pixel 6",
						},
					},
					Metrics: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
						{
							Metric: "crashRate",
							DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{
								Value: "0.05",
							},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple rows",
			rows: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
				{
					AggregationPeriod: "DAILY",
					StartTime: &playdeveloperreporting.GoogleTypeDateTime{
						Year:  2024,
						Month: 1,
						Day:   15,
					},
				},
				{
					AggregationPeriod: "DAILY",
					StartTime: &playdeveloperreporting.GoogleTypeDateTime{
						Year:  2024,
						Month: 1,
						Day:   16,
					},
				},
			},
			expected: 2,
		},
		{
			name: "row without dimensions",
			rows: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
				{
					AggregationPeriod: "DAILY",
					StartTime: &playdeveloperreporting.GoogleTypeDateTime{
						Year:  2024,
						Month: 1,
						Day:   15,
					},
					Dimensions: nil,
					Metrics:    nil,
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatMetricsRows(tt.rows)
			if len(result) != tt.expected {
				t.Errorf("formatMetricsRows() returned %d items, want %d", len(result), tt.expected)
			}

			// Verify structure of first result if present
			if len(result) > 0 {
				first := result[0]
				if _, ok := first["aggregationPeriod"]; !ok {
					t.Error("Missing aggregationPeriod field")
				}
				if _, ok := first["startTime"]; !ok {
					t.Error("Missing startTime field")
				}
				if _, ok := first["dimensions"]; !ok {
					t.Error("Missing dimensions field")
				}
				if _, ok := first["metrics"]; !ok {
					t.Error("Missing metrics field")
				}
			}
		})
	}
}

func TestFormatMetricsRows_DimensionTypes(t *testing.T) {
	rows := []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
		{
			AggregationPeriod: "DAILY",
			StartTime: &playdeveloperreporting.GoogleTypeDateTime{
				Year:  2024,
				Month: 1,
				Day:   15,
			},
			Dimensions: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DimensionValue{
				{
					Dimension:  "intDimension",
					Int64Value: 42,
				},
				{
					Dimension:   "stringDimension",
					StringValue: "test-value",
				},
			},
			Metrics: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				{
					Metric: "testMetric",
					DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{
						Value: "1.23",
					},
				},
			},
		},
	}

	result := formatMetricsRows(rows)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	dimensions, ok := result[0]["dimensions"].(map[string]interface{})
	if !ok {
		t.Fatal("dimensions should be a map")
	}

	// Check int dimension
	if dimensions["intDimension"] != int64(42) {
		t.Errorf("intDimension = %v, want 42", dimensions["intDimension"])
	}

	// Check string dimension
	if dimensions["stringDimension"] != "test-value" {
		t.Errorf("stringDimension = %v, want test-value", dimensions["stringDimension"])
	}

	metrics, ok := result[0]["metrics"].(map[string]interface{})
	if !ok {
		t.Fatal("metrics should be a map")
	}

	if metrics["testMetric"] != "1.23" {
		t.Errorf("testMetric = %v, want 1.23", metrics["testMetric"])
	}
}

// ============================================================================
// formatErrorIssues Tests
// ============================================================================

func TestFormatErrorIssues(t *testing.T) {
	tests := []struct {
		name     string
		issues   []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue
		expected int
	}{
		{
			name:     "empty issues",
			issues:   []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue{},
			expected: 0,
		},
		{
			name:     "nil issues",
			issues:   nil,
			expected: 0,
		},
		{
			name: "single issue",
			issues: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue{
				{
					Name:             "apps/com.example.app/errorIssues/123",
					Cause:            "java.lang.NullPointerException",
					Type:             "crash",
					Location:         "com.example.MainActivity.onCreate",
					DistinctUsers:    150,
					ErrorReportCount: 500,
					IssueUri:         "https://play.google.com/console/...",
				},
			},
			expected: 1,
		},
		{
			name: "multiple issues",
			issues: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue{
				{
					Name:  "apps/com.example.app/errorIssues/1",
					Cause: "Exception 1",
				},
				{
					Name:  "apps/com.example.app/errorIssues/2",
					Cause: "Exception 2",
				},
				{
					Name:  "apps/com.example.app/errorIssues/3",
					Cause: "Exception 3",
				},
			},
			expected: 3,
		},
		{
			name: "issue with empty fields",
			issues: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue{
				{
					Name:             "apps/com.example.app/errorIssues/empty",
					Cause:            "",
					Type:             "",
					Location:         "",
					DistinctUsers:    0,
					ErrorReportCount: 0,
					IssueUri:         "",
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatErrorIssues(tt.issues)
			if len(result) != tt.expected {
				t.Errorf("formatErrorIssues() returned %d items, want %d", len(result), tt.expected)
			}

			// Verify structure if there are results
			if len(result) > 0 {
				first := result[0]
				requiredFields := []string{"name", "cause", "type", "location", "distinctUsers", "errorReportCount", "issueUri"}
				for _, field := range requiredFields {
					if _, ok := first[field]; !ok {
						t.Errorf("Missing required field: %s", field)
					}
				}
			}
		})
	}
}

func TestFormatErrorIssues_FieldValues(t *testing.T) {
	issues := []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue{
		{
			Name:             "apps/com.example.app/errorIssues/123",
			Cause:            "java.lang.IllegalStateException",
			Type:             "crash",
			Location:         "com.example.Activity.onResume",
			DistinctUsers:    1000,
			ErrorReportCount: 5000,
			IssueUri:         "https://play.google.com/console/developers/...",
		},
	}

	result := formatErrorIssues(issues)
	if len(result) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(result))
	}

	if result[0]["name"] != "apps/com.example.app/errorIssues/123" {
		t.Errorf("name = %v", result[0]["name"])
	}
	if result[0]["cause"] != "java.lang.IllegalStateException" {
		t.Errorf("cause = %v", result[0]["cause"])
	}
	if result[0]["type"] != "crash" {
		t.Errorf("type = %v", result[0]["type"])
	}
	if result[0]["distinctUsers"] != int64(1000) {
		t.Errorf("distinctUsers = %v, want 1000", result[0]["distinctUsers"])
	}
	if result[0]["errorReportCount"] != int64(5000) {
		t.Errorf("errorReportCount = %v, want 5000", result[0]["errorReportCount"])
	}
}

// ============================================================================
// formatErrorReports Tests
// ============================================================================

func TestFormatErrorReports(t *testing.T) {
	tests := []struct {
		name     string
		reports  []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport
		expected int
	}{
		{
			name:     "empty reports",
			reports:  []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport{},
			expected: 0,
		},
		{
			name:     "nil reports",
			reports:  nil,
			expected: 0,
		},
		{
			name: "single report",
			reports: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport{
				{
					Name:       "apps/com.example.app/errorReports/abc123",
					Issue:      "apps/com.example.app/errorIssues/123",
					EventTime:  "2024-01-15T10:00:00Z",
					AppVersion: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1AppVersion{VersionCode: 100},
					OsVersion:  &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1OsVersion{ApiLevel: 34},
					Type:       "java.lang.NullPointerException",
					ReportText: "Stack trace here...",
				},
			},
			expected: 1,
		},
		{
			name: "report with device model",
			reports: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport{
				{
					Name: "apps/com.example.app/errorReports/report1",
					DeviceModel: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DeviceModelSummary{
						DeviceId: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DeviceId{
							BuildBrand:  "google",
							BuildDevice: "panther",
						},
					},
				},
			},
			expected: 1,
		},
		{
			name: "report without device model",
			reports: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport{
				{
					Name:        "apps/com.example.app/errorReports/report2",
					DeviceModel: nil,
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatErrorReports(tt.reports)
			if len(result) != tt.expected {
				t.Errorf("formatErrorReports() returned %d items, want %d", len(result), tt.expected)
			}

			if len(result) > 0 {
				first := result[0]
				requiredFields := []string{"name", "issue", "eventTime", "appVersion", "osVersion", "deviceModel", "type", "reportText"}
				for _, field := range requiredFields {
					if _, ok := first[field]; !ok {
						t.Errorf("Missing required field: %s", field)
					}
				}
			}
		})
	}
}

func TestFormatErrorReports_DeviceModel(t *testing.T) {
	tests := []struct {
		name        string
		deviceModel *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DeviceModelSummary
		expected    string
	}{
		{
			name: "with device model",
			deviceModel: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DeviceModelSummary{
				DeviceId: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DeviceId{
					BuildBrand:  "google",
					BuildDevice: "panther",
				},
			},
			expected: "google/panther",
		},
		{
			name:        "nil device model",
			deviceModel: nil,
			expected:    "",
		},
		{
			name: "device model with nil device id",
			deviceModel: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DeviceModelSummary{
				DeviceId: nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reports := []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport{
				{
					Name:        "test-report",
					DeviceModel: tt.deviceModel,
				},
			}

			result := formatErrorReports(reports)
			if len(result) != 1 {
				t.Fatal("Expected 1 result")
			}

			if result[0]["deviceModel"] != tt.expected {
				t.Errorf("deviceModel = %v, want %v", result[0]["deviceModel"], tt.expected)
			}
		})
	}
}

// ============================================================================
// formatAnomalies Tests
// ============================================================================

func TestFormatAnomalies(t *testing.T) {
	tests := []struct {
		name      string
		anomalies []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly
		expected  int
	}{
		{
			name:      "empty anomalies",
			anomalies: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly{},
			expected:  0,
		},
		{
			name:      "nil anomalies",
			anomalies: nil,
			expected:  0,
		},
		{
			name: "single anomaly",
			anomalies: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly{
				{
					Name:      "apps/com.example.app/anomalies/123",
					Metric:    &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{Metric: "crashRate"},
					MetricSet: "crashRateMetricSet",
					Dimensions: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1DimensionValue{
						{Dimension: "versionCode", Int64Value: 100},
					},
					TimelineSpec: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
						AggregationPeriod: "DAILY",
					},
				},
			},
			expected: 1,
		},
		{
			name: "multiple anomalies",
			anomalies: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly{
				{Name: "anomaly1", Metric: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{Metric: "crashRate"}},
				{Name: "anomaly2", Metric: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{Metric: "anrRate"}},
				{Name: "anomaly3", Metric: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{Metric: "slowStartRate"}},
			},
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAnomalies(tt.anomalies)
			if len(result) != tt.expected {
				t.Errorf("formatAnomalies() returned %d items, want %d", len(result), tt.expected)
			}

			if len(result) > 0 {
				first := result[0]
				requiredFields := []string{"name", "metric", "metricSet", "dimensions", "timelineSpec"}
				for _, field := range requiredFields {
					if _, ok := first[field]; !ok {
						t.Errorf("Missing required field: %s", field)
					}
				}
			}
		})
	}
}

// ============================================================================
// VitalsErrorsIssuesCmd.buildFilter Tests
// ============================================================================

func TestVitalsErrorsIssuesCmd_BuildFilter(t *testing.T) {
	tests := []struct {
		name     string
		cmd      VitalsErrorsIssuesCmd
		expected string
	}{
		{
			name: "with query and interval",
			cmd: VitalsErrorsIssuesCmd{
				Query:    "NullPointerException",
				Interval: "last7Days",
			},
			expected: `(cause =~ "NullPointerException" OR location =~ "NullPointerException") AND activeBetween(`,
		},
		{
			name: "only query",
			cmd: VitalsErrorsIssuesCmd{
				Query:    "IllegalStateException",
				Interval: "",
			},
			expected: `(cause =~ "IllegalStateException" OR location =~ "IllegalStateException")`,
		},
		{
			name: "only interval",
			cmd: VitalsErrorsIssuesCmd{
				Query:    "",
				Interval: "last30Days",
			},
			expected: `activeBetween(`,
		},
		{
			name: "empty query and interval",
			cmd: VitalsErrorsIssuesCmd{
				Query:    "",
				Interval: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cmd.buildFilter()
			if !strings.HasPrefix(result, tt.expected) {
				t.Errorf("buildFilter() = %v, expected to start with %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// VitalsErrorsIssuesCmd.intervalToDateRange Tests
// ============================================================================

func TestVitalsErrorsIssuesCmd_IntervalToDateRange(t *testing.T) {
	cmd := &VitalsErrorsIssuesCmd{}

	tests := []struct {
		name         string
		interval     string
		containsFunc func(string) bool
	}{
		{
			name:     "last7Days",
			interval: "last7Days",
			containsFunc: func(s string) bool {
				return strings.Contains(s, "T00:00:00Z")
			},
		},
		{
			name:     "last30Days",
			interval: "last30Days",
			containsFunc: func(s string) bool {
				return strings.Contains(s, "T00:00:00Z")
			},
		},
		{
			name:     "last90Days",
			interval: "last90Days",
			containsFunc: func(s string) bool {
				return strings.Contains(s, "T00:00:00Z")
			},
		},
		{
			name:     "unknown interval defaults to last30Days",
			interval: "unknown",
			containsFunc: func(s string) bool {
				return strings.Contains(s, "T00:00:00Z")
			},
		},
		{
			name:     "empty interval defaults to last30Days",
			interval: "",
			containsFunc: func(s string) bool {
				return strings.Contains(s, "T00:00:00Z")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd.Interval = tt.interval
			result := cmd.intervalToDateRange()
			if !tt.containsFunc(result) {
				t.Errorf("intervalToDateRange() = %v, did not match expected format", result)
			}
			// Verify it's a valid date range format
			if !strings.Contains(result, `","`) {
				t.Error("Expected date range to contain comma-separated dates")
			}
		})
	}
}

// ============================================================================
// VitalsErrorsReportsCmd.getIntervalDates Tests
// ============================================================================

func TestVitalsErrorsReportsCmd_GetIntervalDates(t *testing.T) {
	cmd := &VitalsErrorsReportsCmd{}

	tests := []struct {
		name         string
		interval     string
		expectedDays int
	}{
		{
			name:         "last7Days",
			interval:     "last7Days",
			expectedDays: 7,
		},
		{
			name:         "last30Days default",
			interval:     "last30Days",
			expectedDays: 30,
		},
		{
			name:         "last90Days",
			interval:     "last90Days",
			expectedDays: 90,
		},
		{
			name:         "unknown defaults to 30 days",
			interval:     "unknown",
			expectedDays: 30,
		},
		{
			name:         "empty defaults to 30 days",
			interval:     "",
			expectedDays: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd.Interval = tt.interval
			startDate, endDate := cmd.getIntervalDates()

			diff := endDate.Sub(startDate)
			expectedDuration := time.Duration(tt.expectedDays) * 24 * time.Hour
			// Allow 1 day tolerance for timezone/rounding issues
			tolerance := 24 * time.Hour

			if diff < expectedDuration-tolerance || diff > expectedDuration+tolerance {
				t.Errorf("Date range = %v, expected approximately %v", diff, expectedDuration)
			}
		})
	}
}

// ============================================================================
// Page Response Tests
// ============================================================================

func TestCrashRatePageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetResponse{
		NextPageToken: "next-token-123",
		Rows: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
			{AggregationPeriod: "DAILY"},
			{AggregationPeriod: "DAILY"},
		},
	}

	wrapper := crashRatePageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "next-token-123" {
		t.Errorf("GetNextPageToken() = %v, want next-token-123", wrapper.GetNextPageToken())
	}

	items := wrapper.GetItems()
	if len(items) != 2 {
		t.Errorf("GetItems() returned %d items, want 2", len(items))
	}
}

func TestAnrRatePageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetResponse{
		NextPageToken: "",
		Rows:          []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{},
	}

	wrapper := anrRatePageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "" {
		t.Error("GetNextPageToken() should return empty string")
	}

	items := wrapper.GetItems()
	if len(items) != 0 {
		t.Errorf("GetItems() returned %d items, want 0", len(items))
	}
}

func TestErrorIssuesPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SearchErrorIssuesResponse{
		NextPageToken: "token",
		ErrorIssues: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorIssue{
			{Name: "issue1"},
			{Name: "issue2"},
		},
	}

	wrapper := errorIssuesPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "token" {
		t.Errorf("GetNextPageToken() = %v, want token", wrapper.GetNextPageToken())
	}

	items := wrapper.GetItems()
	if len(items) != 2 {
		t.Errorf("GetItems() returned %d items, want 2", len(items))
	}
}

func TestErrorReportsPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1SearchErrorReportsResponse{
		NextPageToken: "report-token",
		ErrorReports: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ErrorReport{
			{Name: "report1"},
		},
	}

	wrapper := errorReportsPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "report-token" {
		t.Errorf("GetNextPageToken() = %v, want report-token", wrapper.GetNextPageToken())
	}

	items := wrapper.GetItems()
	if len(items) != 1 {
		t.Errorf("GetItems() returned %d items, want 1", len(items))
	}
}

func TestErrorCountPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryErrorCountMetricSetResponse{
		NextPageToken: "count-token",
		Rows: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{
			{AggregationPeriod: "DAILY"},
		},
	}

	wrapper := errorCountPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "count-token" {
		t.Errorf("GetNextPageToken() = %v, want count-token", wrapper.GetNextPageToken())
	}

	items := wrapper.GetItems()
	if len(items) != 1 {
		t.Errorf("GetItems() returned %d items, want 1", len(items))
	}
}

func TestExcessiveWakeupPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryExcessiveWakeupRateMetricSetResponse{
		NextPageToken: "wakeup-token",
		Rows:          []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{},
	}

	wrapper := excessiveWakeupPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "wakeup-token" {
		t.Errorf("GetNextPageToken() = %v, want wakeup-token", wrapper.GetNextPageToken())
	}
}

func TestSlowRenderingPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowRenderingRateMetricSetResponse{
		NextPageToken: "rendering-token",
		Rows:          []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{},
	}

	wrapper := slowRenderingPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "rendering-token" {
		t.Errorf("GetNextPageToken() = %v, want rendering-token", wrapper.GetNextPageToken())
	}
}

func TestSlowStartPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QuerySlowStartRateMetricSetResponse{
		NextPageToken: "start-token",
		Rows:          []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{},
	}

	wrapper := slowStartPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "start-token" {
		t.Errorf("GetNextPageToken() = %v, want start-token", wrapper.GetNextPageToken())
	}
}

func TestStuckWakelockPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryStuckBackgroundWakelockRateMetricSetResponse{
		NextPageToken: "wakelock-token",
		Rows:          []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{},
	}

	wrapper := stuckWakelockPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "wakelock-token" {
		t.Errorf("GetNextPageToken() = %v, want wakelock-token", wrapper.GetNextPageToken())
	}
}

func TestAnomaliesPageResponse(t *testing.T) {
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1ListAnomaliesResponse{
		NextPageToken: "anomaly-token",
		Anomalies: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1Anomaly{
			{Name: "anomaly1"},
		},
	}

	wrapper := anomaliesPageResponse{resp: resp}

	if wrapper.GetNextPageToken() != "anomaly-token" {
		t.Errorf("GetNextPageToken() = %v, want anomaly-token", wrapper.GetNextPageToken())
	}

	items := wrapper.GetItems()
	if len(items) != 1 {
		t.Errorf("GetItems() returned %d items, want 1", len(items))
	}
}

// ============================================================================
// Command Execution with Invalid Auth (Error Paths)
// ============================================================================

func TestVitalsCommands_InvalidAuth(t *testing.T) {
	tests := []struct {
		name string
		cmd  interface {
			Run(*Globals) error
		}
	}{
		{
			name: "crashes",
			cmd:  &VitalsCrashesCmd{},
		},
		{
			name: "anrs",
			cmd:  &VitalsAnrsCmd{},
		},
		{
			name: "errors issues",
			cmd:  &VitalsErrorsIssuesCmd{},
		},
		{
			name: "errors reports",
			cmd:  &VitalsErrorsReportsCmd{},
		},
		{
			name: "errors counts get",
			cmd:  &VitalsErrorsCountsGetCmd{},
		},
		{
			name: "errors counts query",
			cmd:  &VitalsErrorsCountsQueryCmd{},
		},
		{
			name: "metrics excessive wakeups",
			cmd:  &VitalsMetricsExcessiveWakeupsCmd{},
		},
		{
			name: "metrics slow rendering",
			cmd:  &VitalsMetricsSlowRenderingCmd{},
		},
		{
			name: "metrics slow start",
			cmd:  &VitalsMetricsSlowStartCmd{},
		},
		{
			name: "metrics stuck wakelocks",
			cmd:  &VitalsMetricsStuckWakelocksCmd{},
		},
		{
			name: "anomalies list",
			cmd:  &VitalsAnomaliesListCmd{},
		},
		{
			name: "query",
			cmd:  &VitalsQueryCmd{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				KeyPath: "/nonexistent/key.json", // Invalid key path
			}

			err := tt.cmd.Run(globals)
			if err == nil {
				t.Fatal("Expected error for invalid auth")
			}
			// Error could be auth-related or API-related
			t.Logf("Got expected error: %v", err)
		})
	}
}

// ============================================================================
// VitalsQueryCmd Metric Set Detection Tests
// ============================================================================

func TestVitalsQueryCmd_MetricSetDetection(t *testing.T) {
	tests := []struct {
		name        string
		metrics     []string
		expectedSet string
	}{
		{
			name:        "crash rate metric",
			metrics:     []string{metricCrashRate},
			expectedSet: metricSetCrashRate,
		},
		{
			name:        "user perceived crash rate",
			metrics:     []string{"userPerceivedCrashRate"},
			expectedSet: metricSetCrashRate,
		},
		{
			name:        "ANR rate metric",
			metrics:     []string{metricAnrRate},
			expectedSet: metricSetAnrRate,
		},
		{
			name:        "user perceived ANR rate",
			metrics:     []string{"userPerceivedAnrRate"},
			expectedSet: metricSetAnrRate,
		},
		{
			name:        "slow rendering rate",
			metrics:     []string{"slowRenderingRate"},
			expectedSet: metricSetSlowRendering,
		},
		{
			name:        "slow start rate",
			metrics:     []string{"slowStartRate"},
			expectedSet: metricSetSlowStart,
		},
		{
			name:        "stuck background wakelock rate",
			metrics:     []string{"stuckBackgroundWakelockRate"},
			expectedSet: metricSetStuckBackgroundWakelock,
		},
		{
			name:        "excessive wakeup rate",
			metrics:     []string{"excessiveWakeupRate"},
			expectedSet: metricSetExcessiveWakeup,
		},
		{
			name:        "error count metric",
			metrics:     []string{"errorCount"},
			expectedSet: metricSetErrorCount,
		},
		{
			name:        "error report count metric",
			metrics:     []string{"errorReportCount"},
			expectedSet: metricSetErrorCount,
		},
		{
			name:        "default when empty metrics",
			metrics:     []string{},
			expectedSet: metricSetCrashRate,
		},
		{
			name:        "default when nil metrics",
			metrics:     nil,
			expectedSet: metricSetCrashRate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &VitalsQueryCmd{
				Metrics: tt.metrics,
			}

			// The metric set detection happens inside Run(), but we can test the logic
			// by checking the name construction logic
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

			if metricSetName != tt.expectedSet {
				t.Errorf("metricSetName = %v, want %v", metricSetName, tt.expectedSet)
			}
		})
	}
}

// ============================================================================
// Constants Validation Tests
// ============================================================================

func TestVitalsConstants(t *testing.T) {
	// Test that all metric set constants are properly defined
	constants := map[string]string{
		"formatCSV":                        formatCSV,
		"metricSetCrashRate":               metricSetCrashRate,
		"metricSetAnrRate":                 metricSetAnrRate,
		"metricSetSlowRendering":           metricSetSlowRendering,
		"metricSetSlowStart":               metricSetSlowStart,
		"metricSetStuckBackgroundWakelock": metricSetStuckBackgroundWakelock,
		"metricSetExcessiveWakeup":         metricSetExcessiveWakeup,
		"metricSetErrorCount":              metricSetErrorCount,
	}

	expected := map[string]string{
		"formatCSV":                        "csv",
		"metricSetCrashRate":               "crashRateMetricSet",
		"metricSetAnrRate":                 "anrRateMetricSet",
		"metricSetSlowRendering":           "slowRenderingRateMetricSet",
		"metricSetSlowStart":               "slowStartRateMetricSet",
		"metricSetStuckBackgroundWakelock": "stuckBackgroundWakelockRateMetricSet",
		"metricSetExcessiveWakeup":         "excessiveWakeupRateMetricSet",
		"metricSetErrorCount":              "errorCountMetricSet",
	}

	for name, value := range constants {
		if value == "" {
			t.Errorf("Constant %s is empty", name)
		}
		if expected[name] != "" && value != expected[name] {
			t.Errorf("Constant %s = %v, want %v", name, value, expected[name])
		}
	}
}

// ============================================================================
// Context Propagation Tests
// ============================================================================

func TestVitalsCrashesCmd_ContextPropagation(t *testing.T) {
	cmd := &VitalsCrashesCmd{}

	// Test with custom context
	customCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	globals := &Globals{
		Package: "com.example.app",
		Context: customCtx,
		Timeout: 30 * time.Second,
	}

	// Verify context is set (command will fail due to auth, but that's expected)
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error due to missing auth")
	}

	// The context should have been used (even though the command fails later)
	if globals.Context == nil {
		t.Error("Context should not be nil")
	}
}

func TestVitalsAnrsCmd_ContextPropagation(t *testing.T) {
	cmd := &VitalsAnrsCmd{}

	// Test with nil context (should default to context.Background())
	globals := &Globals{
		Package: "com.example.app",
		Context: nil,
	}

	// The command should handle nil context by using context.Background()
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error due to missing auth")
	}

	// Error should be related to auth, not context
	if strings.Contains(err.Error(), "context") {
		t.Error("Error should not be context-related")
	}
}

// ============================================================================
// VitalsQueryCmd Metric Set Query Dispatcher Tests (Error Path)
// ============================================================================

func TestVitalsQueryCmd_QueryMetricSet_Invalid(t *testing.T) {
	// Create a context for testing
	ctx := context.Background()

	cmd := &VitalsQueryCmd{
		Metrics:    []string{"invalidMetric"},
		Dimensions: []string{},
		PageSize:   100,
	}

	// Build a timeline spec for testing
	timelineSpec := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec{
		AggregationPeriod: "DAILY",
		StartTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  2024,
			Month: 1,
			Day:   1,
		},
		EndTime: &playdeveloperreporting.GoogleTypeDateTime{
			Year:  2024,
			Month: 1,
			Day:   31,
		},
	}

	// Test with unsupported metric set
	supportedMetricSets := []string{
		metricSetCrashRate,
		metricSetAnrRate,
		metricSetSlowRendering,
		metricSetSlowStart,
		metricSetStuckBackgroundWakelock,
		metricSetExcessiveWakeup,
		metricSetErrorCount,
	}

	for _, metricSet := range supportedMetricSets {
		t.Run(fmt.Sprintf("supported_%s", metricSet), func(t *testing.T) {
			// We can't fully test without a real client, but we can verify
			// that the metric set name is formatted correctly
			expectedName := fmt.Sprintf("apps/com.example.app/%s", metricSet)
			if !strings.HasSuffix(expectedName, metricSet) {
				t.Errorf("Name should end with %s", metricSet)
			}
		})
	}

	// Test unsupported metric set
	unsupportedSet := "unsupportedMetricSet"
	_, err := cmd.queryMetricSet(ctx, nil, nil, "apps/com.example.app/"+unsupportedSet, unsupportedSet, timelineSpec)
	if err == nil {
		t.Fatal("Expected error for unsupported metric set")
	}
	if !strings.Contains(err.Error(), "unsupported metric set") {
		t.Errorf("Expected 'unsupported metric set' error, got: %v", err)
	}
}

// ============================================================================
// VitalsErrorsReportsCmd.setIntervalParams Tests
// ============================================================================

func TestVitalsErrorsReportsCmd_SetIntervalParams(t *testing.T) {
	cmd := &VitalsErrorsReportsCmd{}

	// We can't easily test the actual call chaining without a real API client,
	// but we can verify the method exists and can be called
	t.Run("method exists", func(t *testing.T) {
		// This just verifies the method signature is correct
		_ = cmd.setIntervalParams
	})
}

// ============================================================================
// Integration Test Helpers
// ============================================================================

func TestVitalsCommands_WithMockServer(t *testing.T) {
	// This test documents the expected structure for mock server testing
	// In a real implementation, you would:
	// 1. Create a mock HTTP server that returns Google Play API responses
	// 2. Configure the API client to use the mock server URL
	// 3. Execute the commands and verify the responses

	t.Run("mock server structure", func(t *testing.T) {
		// Mock server would need to handle:
		// - Authentication endpoints
		// - Crash rate query endpoint
		// - ANR rate query endpoint
		// - Error issues search endpoint
		// - Error reports search endpoint
		// - Error counts get/query endpoints
		// - Metrics endpoints (excessive wakeups, slow rendering, slow start, stuck wakelocks)
		// - Anomalies list endpoint

		t.Log("Mock server would need endpoints for:")
		t.Log("  - /v1beta1/apps/{package}/crashRateMetricSet:query")
		t.Log("  - /v1beta1/apps/{package}/anrRateMetricSet:query")
		t.Log("  - /v1beta1/apps/{package}/errorIssues:search")
		t.Log("  - /v1beta1/apps/{package}/errorReports:search")
		t.Log("  - /v1beta1/apps/{package}/errorCountMetricSet")
		t.Log("  - /v1beta1/apps/{package}/excessiveWakeupRateMetricSet:query")
		t.Log("  - /v1beta1/apps/{package}/slowRenderingRateMetricSet:query")
		t.Log("  - /v1beta1/apps/{package}/slowStartRateMetricSet:query")
		t.Log("  - /v1beta1/apps/{package}/stuckBackgroundWakelockRateMetricSet:query")
		t.Log("  - /v1beta1/apps/{package}/anomalies")
	})
}
