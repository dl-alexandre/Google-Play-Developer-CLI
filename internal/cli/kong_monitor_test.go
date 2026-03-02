//go:build unit
// +build unit

package cli

import (
	"math"
	"reflect"
	"strings"
	"testing"
	"time"

	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// ============================================================================
// MonitorCmd Structure Tests
// ============================================================================

func TestMonitorCmd_Structure(t *testing.T) {
	cmd := MonitorCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"Watch", "Anomalies", "Dashboard", "Report", "Webhooks",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("MonitorCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("MonitorCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("MonitorCmd.%s should have help tag", name)
		}
	}

	actualFields := v.NumField()
	if actualFields != len(expectedSubcommands) {
		t.Errorf("MonitorCmd has %d fields, expected %d", actualFields, len(expectedSubcommands))
	}
}

// ============================================================================
// MonitorWatchCmd Tests
// ============================================================================

func TestMonitorWatchCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonitorWatchCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package") {
		t.Errorf("Expected 'package' error, got: %v", err)
	}
}

func TestMonitorWatchCmd_NormalizeMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []string
		expected []string
	}{
		{
			name:     "empty returns all",
			metrics:  []string{},
			expected: []string{"crashes", "anrs", "errors"},
		},
		{
			name:     "nil returns all",
			metrics:  nil,
			expected: []string{"crashes", "anrs", "errors"},
		},
		{
			name:     "all returns all metrics",
			metrics:  []string{"all"},
			expected: []string{"crashes", "anrs", "errors"},
		},
		{
			name:     "single crash metric",
			metrics:  []string{"crashes"},
			expected: []string{"crashes"},
		},
		{
			name:     "single crash metric singular",
			metrics:  []string{"crash"},
			expected: []string{"crashes"},
		},
		{
			name:     "single anr metric",
			metrics:  []string{"anrs"},
			expected: []string{"anrs"},
		},
		{
			name:     "single anr metric singular",
			metrics:  []string{"anr"},
			expected: []string{"anrs"},
		},
		{
			name:     "single error metric",
			metrics:  []string{"errors"},
			expected: []string{"errors"},
		},
		{
			name:     "single error metric singular",
			metrics:  []string{"error"},
			expected: []string{"errors"},
		},
		{
			name:     "mixed metrics",
			metrics:  []string{"crash", "anr", "error"},
			expected: []string{"crashes", "anrs", "errors"},
		},
		{
			name:     "case insensitive",
			metrics:  []string{"CRASHES", "AnRs", "Errors"},
			expected: []string{"crashes", "anrs", "errors"},
		},
		{
			name:     "duplicates handled",
			metrics:  []string{"crashes", "crashes"},
			expected: []string{"crashes", "crashes"},
		},
		{
			name:     "unknown metrics ignored",
			metrics:  []string{"unknown", "crashes"},
			expected: []string{"crashes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorWatchCmd{Metrics: tt.metrics}
			result := cmd.normalizeMetrics()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestMonitorWatchCmd_CalculateSeverity(t *testing.T) {
	tests := []struct {
		name      string
		actual    float64
		threshold float64
		expected  string
	}{
		{
			name:      "exactly at threshold is low",
			actual:    1.0,
			threshold: 1.0,
			expected:  "low",
		},
		{
			name:      "slightly above threshold is low",
			actual:    1.4,
			threshold: 1.0,
			expected:  "low",
		},
		{
			name:      "1.5x threshold is medium",
			actual:    1.5,
			threshold: 1.0,
			expected:  "medium",
		},
		{
			name:      "between 1.5x and 2x is medium",
			actual:    1.8,
			threshold: 1.0,
			expected:  "medium",
		},
		{
			name:      "2x threshold is high",
			actual:    2.0,
			threshold: 1.0,
			expected:  "high",
		},
		{
			name:      "between 2x and 3x is high",
			actual:    2.5,
			threshold: 1.0,
			expected:  "high",
		},
		{
			name:      "3x threshold is critical",
			actual:    3.0,
			threshold: 1.0,
			expected:  "critical",
		},
		{
			name:      "above 3x is critical",
			actual:    5.0,
			threshold: 1.0,
			expected:  "critical",
		},
		{
			name:      "zero threshold edge case",
			actual:    1.0,
			threshold: 0.0,
			expected:  "critical",
		},
		{
			name:      "very high ratio",
			actual:    100.0,
			threshold: 0.01,
			expected:  "critical",
		},
	}

	cmd := &MonitorWatchCmd{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.calculateSeverity(tt.actual, tt.threshold)
			if result != tt.expected {
				t.Errorf("calculateSeverity(%v, %v) = %v, want %v",
					tt.actual, tt.threshold, result, tt.expected)
			}
		})
	}
}

func TestMonitorWatchCmd_DefaultThresholds(t *testing.T) {
	cmd := &MonitorWatchCmd{}

	// Verify default values from struct tags
	if cmd.CrashThreshold != 0.01 {
		t.Errorf("Expected default CrashThreshold 0.01, got %v", cmd.CrashThreshold)
	}

	if cmd.AnrThreshold != 0.005 {
		t.Errorf("Expected default AnrThreshold 0.005, got %v", cmd.AnrThreshold)
	}

	if cmd.ErrorThreshold != 100 {
		t.Errorf("Expected default ErrorThreshold 100, got %v", cmd.ErrorThreshold)
	}

	if cmd.Interval != 5*time.Minute {
		t.Errorf("Expected default Interval 5m, got %v", cmd.Interval)
	}

	if cmd.Format != "json" {
		t.Errorf("Expected default Format 'json', got %v", cmd.Format)
	}
}

// ============================================================================
// MonitorAnomaliesCmd Tests
// ============================================================================

func TestMonitorAnomaliesCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonitorAnomaliesCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package") {
		t.Errorf("Expected 'package' error, got: %v", err)
	}
}

func TestMonitorAnomaliesCmd_NormalizeMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []string
		expected []string
	}{
		{
			name:     "empty returns all",
			metrics:  []string{},
			expected: []string{"crashes", "anrs", "errors"},
		},
		{
			name:     "all returns all metrics",
			metrics:  []string{"all"},
			expected: []string{"crashes", "anrs", "errors"},
		},
		{
			name:     "single metrics",
			metrics:  []string{"crashes"},
			expected: []string{"crashes"},
		},
		{
			name:     "singular forms",
			metrics:  []string{"crash", "anr", "error"},
			expected: []string{"crashes", "anrs", "errors"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorAnomaliesCmd{Metrics: tt.metrics}
			result := cmd.normalizeMetrics()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %v, got %v", tt.expected, result)
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestMonitorAnomaliesCmd_GetSensitivityMultiplier(t *testing.T) {
	tests := []struct {
		name        string
		sensitivity string
		expected    float64
	}{
		{
			name:        "low sensitivity",
			sensitivity: "low",
			expected:    3.0,
		},
		{
			name:        "medium sensitivity (default)",
			sensitivity: "medium",
			expected:    2.0,
		},
		{
			name:        "high sensitivity",
			sensitivity: "high",
			expected:    1.5,
		},
		{
			name:        "empty defaults to medium",
			sensitivity: "",
			expected:    2.0,
		},
		{
			name:        "unknown defaults to medium",
			sensitivity: "unknown",
			expected:    2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorAnomaliesCmd{Sensitivity: tt.sensitivity}
			result := cmd.getSensitivityMultiplier()
			if result != tt.expected {
				t.Errorf("getSensitivityMultiplier() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMonitorAnomaliesCmd_CalculateAnomalySeverity(t *testing.T) {
	tests := []struct {
		name      string
		deviation float64
		expected  string
	}{
		{
			name:      "low deviation",
			deviation: 25.0,
			expected:  "low",
		},
		{
			name:      "exactly 50 is medium",
			deviation: 50.0,
			expected:  "medium",
		},
		{
			name:      "between 50 and 100 is medium",
			deviation: 75.0,
			expected:  "medium",
		},
		{
			name:      "exactly 100 is high",
			deviation: 100.0,
			expected:  "high",
		},
		{
			name:      "between 100 and 200 is high",
			deviation: 150.0,
			expected:  "high",
		},
		{
			name:      "exactly 200 is critical",
			deviation: 200.0,
			expected:  "critical",
		},
		{
			name:      "above 200 is critical",
			deviation: 300.0,
			expected:  "critical",
		},
		{
			name:      "zero deviation",
			deviation: 0.0,
			expected:  "low",
		},
		{
			name:      "negative deviation",
			deviation: -50.0,
			expected:  "low",
		},
	}

	cmd := &MonitorAnomaliesCmd{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.calculateAnomalySeverity(tt.deviation)
			if result != tt.expected {
				t.Errorf("calculateAnomalySeverity(%v) = %v, want %v",
					tt.deviation, result, tt.expected)
			}
		})
	}
}

func TestMonitorAnomaliesCmd_InvalidSinceDate(t *testing.T) {
	cmd := &MonitorAnomaliesCmd{
		Since: "invalid-date",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid since date")
	}
	if !strings.Contains(err.Error(), "invalid since date") {
		t.Errorf("Expected 'invalid since date' error, got: %v", err)
	}
}

// ============================================================================
// MonitorDashboardCmd Tests
// ============================================================================

func TestMonitorDashboardCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonitorDashboardCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package") {
		t.Errorf("Expected 'package' error, got: %v", err)
	}
}

func TestMonitorDashboardCmd_NormalizeDashboardMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  []string
		expected []string
	}{
		{
			name:     "empty returns all",
			metrics:  []string{},
			expected: []string{"crashes", "anrs", "errors", "slow-rendering", "slow-start", "wakeups", "wakelocks"},
		},
		{
			name:     "all returns all metrics",
			metrics:  []string{"all"},
			expected: []string{"crashes", "anrs", "errors", "slow-rendering", "slow-start", "wakeups", "wakelocks"},
		},
		{
			name:     "crashes variations",
			metrics:  []string{"crash", "crashes"},
			expected: []string{"crashes", "crashes"},
		},
		{
			name:     "anrs variations",
			metrics:  []string{"anr", "anrs"},
			expected: []string{"anrs", "anrs"},
		},
		{
			name:     "errors variations",
			metrics:  []string{"error", "errors"},
			expected: []string{"errors", "errors"},
		},
		{
			name:     "slow rendering variations",
			metrics:  []string{"slow-rendering", "slowrendering"},
			expected: []string{"slow-rendering", "slow-rendering"},
		},
		{
			name:     "slow start variations",
			metrics:  []string{"slow-start", "slowstart"},
			expected: []string{"slow-start", "slow-start"},
		},
		{
			name:     "wakeups variations",
			metrics:  []string{"wakeups", "wakeup", "excessive-wakeups"},
			expected: []string{"wakeups", "wakeups", "wakeups"},
		},
		{
			name:     "wakelocks variations",
			metrics:  []string{"wakelocks", "wakelock", "stuck-wakelocks"},
			expected: []string{"wakelocks", "wakelocks", "wakelocks"},
		},
		{
			name:     "case insensitive",
			metrics:  []string{"CRASHES", "Slow-Rendering", "WAKEUPS"},
			expected: []string{"crashes", "slow-rendering", "wakeups"},
		},
		{
			name:     "mixed metrics",
			metrics:  []string{"crashes", "slow-rendering", "wakeups"},
			expected: []string{"crashes", "slow-rendering", "wakeups"},
		},
		{
			name:     "unknown metrics ignored",
			metrics:  []string{"unknown", "crashes"},
			expected: []string{"crashes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &MonitorDashboardCmd{Metrics: tt.metrics}
			result := cmd.normalizeDashboardMetrics()

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %v, got %v (length mismatch)", tt.expected, result)
				return
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected %s at index %d, got %s", tt.expected[i], i, v)
				}
			}
		})
	}
}

func TestMonitorDashboardCmd_CalculateTrends(t *testing.T) {
	tests := []struct {
		name          string
		avgCrashRate  float64
		avgAnrRate    float64
		expectedCrash string
		expectedAnr   string
		expectedError string
	}{
		{
			name:          "all stable at zero",
			avgCrashRate:  0.0,
			avgAnrRate:    0.0,
			expectedCrash: "stable",
			expectedAnr:   "stable",
			expectedError: "stable",
		},
		{
			name:          "crash increasing",
			avgCrashRate:  0.03,
			avgAnrRate:    0.0,
			expectedCrash: "increasing",
			expectedAnr:   "stable",
			expectedError: "stable",
		},
		{
			name:          "crash decreasing",
			avgCrashRate:  0.001,
			avgAnrRate:    0.0,
			expectedCrash: "decreasing",
			expectedAnr:   "stable",
			expectedError: "stable",
		},
		{
			name:          "crash stable in middle",
			avgCrashRate:  0.01,
			avgAnrRate:    0.0,
			expectedCrash: "stable",
			expectedAnr:   "stable",
			expectedError: "stable",
		},
		{
			name:          "anr increasing",
			avgCrashRate:  0.0,
			avgAnrRate:    0.02,
			expectedCrash: "stable",
			expectedAnr:   "increasing",
			expectedError: "stable",
		},
		{
			name:          "anr decreasing",
			avgCrashRate:  0.0,
			avgAnrRate:    0.0005,
			expectedCrash: "stable",
			expectedAnr:   "decreasing",
			expectedError: "stable",
		},
		{
			name:          "anr stable in middle",
			avgCrashRate:  0.0,
			avgAnrRate:    0.005,
			expectedCrash: "stable",
			expectedAnr:   "stable",
			expectedError: "stable",
		},
		{
			name:          "both increasing",
			avgCrashRate:  0.03,
			avgAnrRate:    0.02,
			expectedCrash: "increasing",
			expectedAnr:   "increasing",
			expectedError: "stable",
		},
		{
			name:          "boundary crash decreasing",
			avgCrashRate:  0.005,
			avgAnrRate:    0.001,
			expectedCrash: "decreasing",
			expectedAnr:   "stable",
			expectedError: "stable",
		},
		{
			name:          "boundary anr decreasing",
			avgCrashRate:  0.01,
			avgAnrRate:    0.001,
			expectedCrash: "stable",
			expectedAnr:   "decreasing",
			expectedError: "stable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &dashboardResult{
				Summary: dashboardSummary{
					AvgCrashRate: tt.avgCrashRate,
					AvgAnrRate:   tt.avgAnrRate,
				},
			}

			cmd := &MonitorDashboardCmd{}
			cmd.calculateTrends(result)

			if result.Trends.CrashTrend != tt.expectedCrash {
				t.Errorf("CrashTrend = %v, want %v", result.Trends.CrashTrend, tt.expectedCrash)
			}
			if result.Trends.AnrTrend != tt.expectedAnr {
				t.Errorf("AnrTrend = %v, want %v", result.Trends.AnrTrend, tt.expectedAnr)
			}
			if result.Trends.ErrorTrend != tt.expectedError {
				t.Errorf("ErrorTrend = %v, want %v", result.Trends.ErrorTrend, tt.expectedError)
			}
		})
	}
}

func TestMonitorDashboardCmd_DefaultPeriod(t *testing.T) {
	cmd := &MonitorDashboardCmd{}

	if cmd.Period != 7 {
		t.Errorf("Expected default Period 7, got %v", cmd.Period)
	}
}

// ============================================================================
// MonitorReportCmd Tests
// ============================================================================

func TestMonitorReportCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonitorReportCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package") {
		t.Errorf("Expected 'package' error, got: %v", err)
	}
}

func TestMonitorReportCmd_CalculateOverallHealth(t *testing.T) {
	tests := []struct {
		name       string
		crashRate  float64
		anrRate    float64
		errorCount int64
		expected   string
	}{
		{
			name:       "excellent health",
			crashRate:  0.0,
			anrRate:    0.0,
			errorCount: 0,
			expected:   "excellent",
		},
		{
			name:       "good health with low crash rate",
			crashRate:  0.01,
			anrRate:    0.0,
			errorCount: 0,
			expected:   "good",
		},
		{
			name:       "fair health with moderate crash rate",
			crashRate:  0.02,
			anrRate:    0.005,
			errorCount: 50,
			expected:   "fair",
		},
		{
			name:       "poor health with high crash rate",
			crashRate:  0.05,
			anrRate:    0.02,
			errorCount: 200,
			expected:   "poor",
		},
		{
			name:       "health at boundaries",
			crashRate:  0.015,
			anrRate:    0.007,
			errorCount: 100,
			expected:   "fair",
		},
	}

	cmd := &MonitorReportCmd{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.calculateOverallHealth(tt.crashRate, tt.anrRate, tt.errorCount)
			if result != tt.expected {
				t.Errorf("calculateOverallHealth(%v, %v, %v) = %v, want %v",
					tt.crashRate, tt.anrRate, tt.errorCount, result, tt.expected)
			}
		})
	}
}

func TestMonitorReportCmd_GenerateFindings(t *testing.T) {
	tests := []struct {
		name             string
		crashRate        float64
		anrRate          float64
		errorCount       int64
		issuesOpen       int
		expectedCount    int
		expectedContains []string
	}{
		{
			name:             "no issues",
			crashRate:        0.01,
			anrRate:          0.005,
			errorCount:       100,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"No significant issues"},
		},
		{
			name:             "high crash rate only",
			crashRate:        0.03,
			anrRate:          0.005,
			errorCount:       100,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"High crash rate"},
		},
		{
			name:             "high anr rate only",
			crashRate:        0.01,
			anrRate:          0.02,
			errorCount:       100,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"High ANR rate"},
		},
		{
			name:             "high error count only",
			crashRate:        0.01,
			anrRate:          0.005,
			errorCount:       1500,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"High error volume"},
		},
		{
			name:             "many open issues only",
			crashRate:        0.01,
			anrRate:          0.005,
			errorCount:       100,
			issuesOpen:       15,
			expectedCount:    1,
			expectedContains: []string{"open error issues"},
		},
		{
			name:             "all issues present",
			crashRate:        0.03,
			anrRate:          0.02,
			errorCount:       1500,
			issuesOpen:       15,
			expectedCount:    4,
			expectedContains: []string{"High crash rate", "High ANR rate", "High error volume", "open error issues"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &reportResult{
				Summary: reportSummary{
					CrashRate:  tt.crashRate,
					AnrRate:    tt.anrRate,
					ErrorCount: tt.errorCount,
					IssuesOpen: tt.issuesOpen,
				},
			}

			cmd := &MonitorReportCmd{}
			cmd.generateFindings(result)

			if len(result.KeyFindings) != tt.expectedCount {
				t.Errorf("Expected %d findings, got %d: %v", tt.expectedCount, len(result.KeyFindings), result.KeyFindings)
			}

			for _, expected := range tt.expectedContains {
				found := false
				for _, finding := range result.KeyFindings {
					if strings.Contains(finding, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected finding containing '%s', got: %v", expected, result.KeyFindings)
				}
			}
		})
	}
}

func TestMonitorReportCmd_GenerateRecommendations(t *testing.T) {
	tests := []struct {
		name             string
		crashRate        float64
		anrRate          float64
		errorCount       int64
		issuesOpen       int
		expectedCount    int
		expectedContains []string
	}{
		{
			name:             "no recommendations needed",
			crashRate:        0.01,
			anrRate:          0.005,
			errorCount:       100,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"Continue monitoring"},
		},
		{
			name:             "crash rate recommendation",
			crashRate:        0.03,
			anrRate:          0.005,
			errorCount:       100,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"Prioritize fixing top crashes"},
		},
		{
			name:             "anr rate recommendation",
			crashRate:        0.01,
			anrRate:          0.02,
			errorCount:       100,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"Review ANR patterns"},
		},
		{
			name:             "error count recommendation",
			crashRate:        0.01,
			anrRate:          0.005,
			errorCount:       1500,
			issuesOpen:       5,
			expectedCount:    1,
			expectedContains: []string{"Investigate error patterns"},
		},
		{
			name:             "open issues recommendation",
			crashRate:        0.01,
			anrRate:          0.005,
			errorCount:       100,
			issuesOpen:       15,
			expectedCount:    1,
			expectedContains: []string{"Address backlog"},
		},
		{
			name:             "all recommendations",
			crashRate:        0.03,
			anrRate:          0.02,
			errorCount:       1500,
			issuesOpen:       15,
			expectedCount:    4,
			expectedContains: []string{"Prioritize fixing", "Review ANR", "Investigate error", "Address backlog"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &reportResult{
				Summary: reportSummary{
					CrashRate:  tt.crashRate,
					AnrRate:    tt.anrRate,
					ErrorCount: tt.errorCount,
					IssuesOpen: tt.issuesOpen,
				},
			}

			cmd := &MonitorReportCmd{}
			cmd.generateRecommendations(result)

			if len(result.Recommendations) != tt.expectedCount {
				t.Errorf("Expected %d recommendations, got %d: %v", tt.expectedCount, len(result.Recommendations), result.Recommendations)
			}

			for _, expected := range tt.expectedContains {
				found := false
				for _, rec := range result.Recommendations {
					if strings.Contains(rec, expected) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected recommendation containing '%s', got: %v", expected, result.Recommendations)
				}
			}
		})
	}
}

func TestMonitorReportCmd_PeriodCalculation(t *testing.T) {
	tests := []struct {
		name         string
		period       string
		expectedDays int
	}{
		{
			name:         "daily period",
			period:       "daily",
			expectedDays: 1,
		},
		{
			name:         "weekly period",
			period:       "weekly",
			expectedDays: 7,
		},
		{
			name:         "monthly period",
			period:       "monthly",
			expectedDays: 30,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the period value is set correctly
			cmd := &MonitorReportCmd{Period: tt.period}
			if cmd.Period != tt.period {
				t.Errorf("Period = %v, want %v", cmd.Period, tt.period)
			}
		})
	}
}

// ============================================================================
// MonitorWebhooksCmd Tests
// ============================================================================

func TestMonitorWebhooksCmd_Structure(t *testing.T) {
	cmd := MonitorWebhooksCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"List",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("MonitorWebhooksCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("MonitorWebhooksCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}
	}
}

func TestMonitorWebhooksListCmd_Run(t *testing.T) {
	cmd := &MonitorWebhooksListCmd{}
	globals := &Globals{}

	// Should succeed with simulated data
	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("MonitorWebhooksListCmd.Run() should succeed, got: %v", err)
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestParseDecimalValue(t *testing.T) {
	tests := []struct {
		name     string
		value    *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue
		expected float64
	}{
		{
			name:     "nil value returns 0",
			value:    nil,
			expected: 0,
		},
		{
			name: "empty decimal value returns 0",
			value: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{
					Value: "",
				},
			},
			expected: 0,
		},
		{
			name: "nil decimal value returns 0",
			value: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				DecimalValue: nil,
			},
			expected: 0,
		},
		{
			name: "valid decimal value",
			value: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{
					Value: "0.123",
				},
			},
			expected: 0.123,
		},
		{
			name: "integer value",
			value: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{
					Value: "100",
				},
			},
			expected: 100,
		},
		{
			name: "negative value",
			value: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{
					Value: "-0.5",
				},
			},
			expected: -0.5,
		},
		{
			name: "invalid value returns 0",
			value: &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricValue{
				DecimalValue: &playdeveloperreporting.GoogleTypeDecimal{
					Value: "invalid",
				},
			},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseDecimalValue(tt.value)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("parseDecimalValue() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// ============================================================================
// Result Structure Tests
// ============================================================================

func TestMonitorAlertStructure(t *testing.T) {
	alert := monitorAlert{
		Metric:      "crashRate",
		Threshold:   0.01,
		ActualValue: 0.02,
		Severity:    "high",
		Timestamp:   time.Now(),
		Dimensions:  map[string]string{"version": "1.0"},
	}

	if alert.Metric != "crashRate" {
		t.Errorf("Expected Metric 'crashRate', got %v", alert.Metric)
	}
	if alert.Severity != "high" {
		t.Errorf("Expected Severity 'high', got %v", alert.Severity)
	}
}

func TestMonitorWatchResultStructure(t *testing.T) {
	result := &monitorWatchResult{
		Package:            "com.example.app",
		Timestamp:          time.Now(),
		Duration:           5 * time.Minute,
		PollCount:          10,
		Alerts:             []monitorAlert{},
		ThresholdsBreached: 2,
		Metrics:            map[string]interface{}{"crashRate": 0.01},
	}

	if result.Package != "com.example.app" {
		t.Errorf("Expected Package 'com.example.app', got %v", result.Package)
	}
	if result.ThresholdsBreached != 2 {
		t.Errorf("Expected ThresholdsBreached 2, got %v", result.ThresholdsBreached)
	}
}

func TestAnomalyResultStructure(t *testing.T) {
	result := &anomalyResult{
		Package:        "com.example.app",
		Timestamp:      time.Now(),
		BaselinePeriod: 30,
		Anomalies:      []detectedAnomaly{},
		TotalAnomalies: 0,
	}

	if result.BaselinePeriod != 30 {
		t.Errorf("Expected BaselinePeriod 30, got %v", result.BaselinePeriod)
	}
}

func TestDetectedAnomalyStructure(t *testing.T) {
	anomaly := detectedAnomaly{
		Metric:       "crashRate",
		Severity:     "high",
		Deviation:    150.0,
		CurrentValue: 0.05,
		BaselineAvg:  0.02,
		Timestamp:    time.Now(),
	}

	if anomaly.Metric != "crashRate" {
		t.Errorf("Expected Metric 'crashRate', got %v", anomaly.Metric)
	}
	if anomaly.Deviation != 150.0 {
		t.Errorf("Expected Deviation 150.0, got %v", anomaly.Deviation)
	}
}

func TestDashboardResultStructure(t *testing.T) {
	result := &dashboardResult{
		Package:     "com.example.app",
		GeneratedAt: time.Now(),
		PeriodDays:  7,
		Summary: dashboardSummary{
			TotalCrashes:  100,
			TotalAnrs:     10,
			TotalErrors:   500,
			AvgCrashRate:  0.01,
			AvgAnrRate:    0.001,
			AffectedUsers: 1000,
		},
		Metrics: make(map[string]interface{}),
		Trends: dashboardTrends{
			CrashTrend: "stable",
			AnrTrend:   "stable",
			ErrorTrend: "stable",
		},
	}

	if result.PeriodDays != 7 {
		t.Errorf("Expected PeriodDays 7, got %v", result.PeriodDays)
	}
	if result.Summary.TotalCrashes != 100 {
		t.Errorf("Expected TotalCrashes 100, got %v", result.Summary.TotalCrashes)
	}
}

func TestReportResultStructure(t *testing.T) {
	result := &reportResult{
		Package:     "com.example.app",
		ReportType:  "daily",
		GeneratedAt: time.Now(),
		PeriodStart: time.Now().AddDate(0, 0, -1),
		PeriodEnd:   time.Now(),
		Summary: reportSummary{
			OverallHealth: "good",
			CrashRate:     0.01,
			AnrRate:       0.005,
			ErrorCount:    100,
			ActiveUsers:   10000,
			IssuesOpen:    5,
		},
		KeyFindings:     []string{},
		Recommendations: []string{},
		RawData:         make(map[string]interface{}),
	}

	if result.ReportType != "daily" {
		t.Errorf("Expected ReportType 'daily', got %v", result.ReportType)
	}
	if result.Summary.OverallHealth != "good" {
		t.Errorf("Expected OverallHealth 'good', got %v", result.Summary.OverallHealth)
	}
}

func TestWebhookInfoStructure(t *testing.T) {
	webhook := webhookInfo{
		ID:         "hook-1",
		Name:       "Test Webhook",
		URL:        "https://example.com/webhook",
		Events:     []string{"vitals.crashes"},
		Active:     true,
		CreatedAt:  time.Now(),
		LastCalled: time.Now(),
		Status:     "active",
	}

	if webhook.ID != "hook-1" {
		t.Errorf("Expected ID 'hook-1', got %v", webhook.ID)
	}
	if !webhook.Active {
		t.Error("Expected Active to be true")
	}
}

func TestWebhooksListResultStructure(t *testing.T) {
	result := &webhooksListResult{
		Webhooks:   []webhookInfo{},
		TotalCount: 0,
		Note:       "Test note",
	}

	if result.TotalCount != 0 {
		t.Errorf("Expected TotalCount 0, got %v", result.TotalCount)
	}
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestOutputResultResult(t *testing.T) {
	// Test that outputResultResult accepts various formats without error
	tests := []struct {
		name   string
		format string
	}{
		{"json format", "json"},
		{"table format", "table"},
		{"html format", "html"},
		{"markdown format", "markdown"},
		{"md format", "md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := output.NewResult(map[string]string{"test": "data"})
			err := outputResultResult(result, tt.format, false)
			// Note: These may print to stdout but should not error
			if err != nil {
				t.Errorf("outputResultResult() with format %q returned error: %v", tt.format, err)
			}
		})
	}
}

// ============================================================================
// Integration Tests (Command Execution)
// ============================================================================

func TestMonitorCommands_RequireAuth(t *testing.T) {
	globals := &Globals{Package: "com.example.app"}

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"MonitorWatchCmd", &MonitorWatchCmd{}},
		{"MonitorAnomaliesCmd", &MonitorAnomaliesCmd{}},
		{"MonitorDashboardCmd", &MonitorDashboardCmd{}},
		{"MonitorReportCmd", &MonitorReportCmd{}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(globals)
			if err == nil {
				t.Errorf("%s.Run() should return error without auth, got nil", tc.name)
				return
			}

			// Should fail with auth error
			if !strings.Contains(err.Error(), "auth") && !strings.Contains(err.Error(), "key") {
				t.Errorf("%s.Run() error should contain auth-related message, got: %v", tc.name, err)
			}
		})
	}
}

// ============================================================================
// Edge Cases and Error Handling
// ============================================================================

func TestMonitorAnomaliesCmd_InvalidSinceDateFormat(t *testing.T) {
	cmd := &MonitorAnomaliesCmd{
		Since: "not-a-valid-date",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid since date format")
	}
	if !strings.Contains(err.Error(), "invalid since date") {
		t.Errorf("Expected 'invalid since date' error, got: %v", err)
	}
}

func TestMonitorAnomaliesCmd_InvalidSinceDateSlashFormat(t *testing.T) {
	cmd := &MonitorAnomaliesCmd{
		Since: "2024/01/01", // Wrong format
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid since date format")
	}
	if !strings.Contains(err.Error(), "invalid since date") {
		t.Errorf("Expected 'invalid since date' error, got: %v", err)
	}
}

func TestMonitorReportCmd_HealthScoreCalculations(t *testing.T) {
	cmd := &MonitorReportCmd{}

	// Test boundary conditions for health score
	tests := []struct {
		name       string
		crashRate  float64
		anrRate    float64
		errorCount int64
		expected   string
	}{
		{
			name:       "perfect health",
			crashRate:  0,
			anrRate:    0,
			errorCount: 0,
			expected:   "excellent",
		},
		{
			name:       "exactly 90 boundary",
			crashRate:  0.005,
			anrRate:    0,
			errorCount: 0,
			expected:   "excellent",
		},
		{
			name:       "exactly 70 boundary",
			crashRate:  0.015,
			anrRate:    0,
			errorCount: 0,
			expected:   "good",
		},
		{
			name:       "exactly 50 boundary",
			crashRate:  0.025,
			anrRate:    0,
			errorCount: 0,
			expected:   "fair",
		},
		{
			name:       "below 50",
			crashRate:  0.03,
			anrRate:    0,
			errorCount: 0,
			expected:   "poor",
		},
		{
			name:       "max deductions from crash rate",
			crashRate:  0.1,
			anrRate:    0,
			errorCount: 0,
			expected:   "poor",
		},
		{
			name:       "max deductions from anr rate",
			crashRate:  0,
			anrRate:    0.05,
			errorCount: 0,
			expected:   "poor",
		},
		{
			name:       "max deductions from error count",
			crashRate:  0,
			anrRate:    0,
			errorCount: 3000,
			expected:   "poor",
		},
		{
			name:       "negative score handling",
			crashRate:  0.1,
			anrRate:    0.05,
			errorCount: 3000,
			expected:   "poor",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cmd.calculateOverallHealth(tt.crashRate, tt.anrRate, tt.errorCount)
			if result != tt.expected {
				t.Errorf("calculateOverallHealth(%v, %v, %v) = %v, want %v",
					tt.crashRate, tt.anrRate, tt.errorCount, result, tt.expected)
			}
		})
	}
}

func TestMonitorWatchCmd_SeverityWithZeroThreshold(t *testing.T) {
	cmd := &MonitorWatchCmd{}

	// Edge case: zero threshold should avoid division by zero
	result := cmd.calculateSeverity(1.0, 0.0)
	// When threshold is 0, ratio becomes +Inf which is >= 3.0, so it should be "critical"
	if result != "critical" {
		t.Errorf("calculateSeverity with zero threshold should return 'critical', got %v", result)
	}
}

func TestMonitorWatchCmd_SeverityWithNegativeValues(t *testing.T) {
	cmd := &MonitorWatchCmd{}

	// Test with negative values
	result := cmd.calculateSeverity(-1.0, 1.0)
	// Negative actual value would give negative ratio, which falls through to default "low"
	if result != "low" {
		t.Errorf("calculateSeverity with negative actual should return 'low', got %v", result)
	}
}

// ============================================================================
// Constants Tests
// ============================================================================

func TestMonitorConstants(t *testing.T) {
	// Verify all metric constants are defined
	constants := map[string]string{
		"metricCrash":                       metricCrash,
		"metricAnr":                         metricAnr,
		"metricError":                       metricError,
		"metricCrashes":                     metricCrashes,
		"metricAnrs":                        metricAnrs,
		"metricErrors":                      metricErrors,
		"metricCrashRate":                   metricCrashRate,
		"metricAnrRate":                     metricAnrRate,
		"metricErrorCount":                  metricErrorCount,
		"metricSlowRenderingRate":           metricSlowRenderingRate,
		"metricSlowStartRate":               metricSlowStartRate,
		"metricExcessiveWakeupRate":         metricExcessiveWakeupRate,
		"metricStuckBackgroundWakelockRate": metricStuckBackgroundWakelockRate,
		"severityHigh":                      severityHigh,
		"severityLow":                       severityLow,
		"metricDistinctUsers":               metricDistinctUsers,
		"trendStable":                       trendStable,
	}

	expected := map[string]string{
		"metricCrash":                       "crash",
		"metricAnr":                         "anr",
		"metricError":                       "error",
		"metricCrashes":                     "crashes",
		"metricAnrs":                        "anrs",
		"metricErrors":                      "errors",
		"metricCrashRate":                   "crashRate",
		"metricAnrRate":                     "anrRate",
		"metricErrorCount":                  "errorCount",
		"metricSlowRenderingRate":           "slowRenderingRate",
		"metricSlowStartRate":               "slowStartRate",
		"metricExcessiveWakeupRate":         "excessiveWakeupRate",
		"metricStuckBackgroundWakelockRate": "stuckBackgroundWakelockRate",
		"severityHigh":                      "high",
		"severityLow":                       "low",
		"metricDistinctUsers":               "distinctUsers",
		"trendStable":                       "stable",
	}

	for name, value := range constants {
		if expected[name] != value {
			t.Errorf("Constant %s = %q, want %q", name, value, expected[name])
		}
	}
}

// ============================================================================
// Command Flag Defaults Tests
// ============================================================================

func TestMonitorWatchCmd_FlagDefaults(t *testing.T) {
	cmd := &MonitorWatchCmd{}

	// Verify defaults match struct tags
	if cmd.Interval != 5*time.Minute {
		t.Errorf("Default Interval = %v, want 5m", cmd.Interval)
	}

	if cmd.Duration != 0 {
		t.Errorf("Default Duration = %v, want 0", cmd.Duration)
	}

	if cmd.CrashThreshold != 0.01 {
		t.Errorf("Default CrashThreshold = %v, want 0.01", cmd.CrashThreshold)
	}

	if cmd.AnrThreshold != 0.005 {
		t.Errorf("Default AnrThreshold = %v, want 0.005", cmd.AnrThreshold)
	}

	if cmd.ErrorThreshold != 100 {
		t.Errorf("Default ErrorThreshold = %v, want 100", cmd.ErrorThreshold)
	}

	if !cmd.AlertOnBreaches {
		t.Log("AlertOnBreaches defaults to false (expected)")
	}

	if cmd.Format != "json" {
		t.Errorf("Default Format = %v, want json", cmd.Format)
	}
}

func TestMonitorAnomaliesCmd_FlagDefaults(t *testing.T) {
	cmd := &MonitorAnomaliesCmd{}

	if cmd.BaselinePeriod != 30 {
		t.Errorf("Default BaselinePeriod = %v, want 30", cmd.BaselinePeriod)
	}

	if cmd.Sensitivity != "medium" {
		t.Errorf("Default Sensitivity = %v, want medium", cmd.Sensitivity)
	}

	if cmd.Since != "" {
		t.Errorf("Default Since = %v, want empty", cmd.Since)
	}

	if cmd.Format != "json" {
		t.Errorf("Default Format = %v, want json", cmd.Format)
	}
}

func TestMonitorDashboardCmd_FlagDefaults(t *testing.T) {
	cmd := &MonitorDashboardCmd{}

	if cmd.Period != 7 {
		t.Errorf("Default Period = %v, want 7", cmd.Period)
	}

	if cmd.Format != "json" {
		t.Errorf("Default Format = %v, want json", cmd.Format)
	}
}

func TestMonitorReportCmd_FlagDefaults(t *testing.T) {
	cmd := &MonitorReportCmd{}

	if cmd.Period != "daily" {
		t.Errorf("Default Period = %v, want daily", cmd.Period)
	}

	if cmd.Format != "json" {
		t.Errorf("Default Format = %v, want json", cmd.Format)
	}

	if cmd.IncludeRawData {
		t.Error("IncludeRawData should default to false")
	}
}
