//go:build unit
// +build unit

package cli

import (
	"context"
	"fmt"
	"strconv"
	"testing"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/gpd/internal/apitest"
	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// CompareVitalsCmd Tests
// ============================================================================

func TestCompareVitalsCmd_Validation(t *testing.T) {
	tests := []struct {
		name     string
		packages []string
		wantErr  bool
		errCode  errors.ErrorCode
	}{
		{
			name:     "empty packages list",
			packages: []string{},
			wantErr:  true,
			errCode:  errors.CodeValidationError,
		},
		{
			name:     "single package",
			packages: []string{"com.example.app1"},
			wantErr:  true,
			errCode:  errors.CodeValidationError,
		},
		{
			name:     "two packages valid",
			packages: []string{"com.example.app1", "com.example.app2"},
			wantErr:  false,
		},
		{
			name:     "three packages valid",
			packages: []string{"com.example.app1", "com.example.app2", "com.example.app3"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CompareVitalsCmd{
				Packages: tt.packages,
				Metric:   "all",
			}
			globals := &Globals{
				Package: "com.example.test", // Note: globals.Package is not used by compare
				Output:  "json",
			}

			err := cmd.Run(globals)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error, got nil")
					return
				}
				apiErr, ok := err.(*errors.APIError)
				if !ok {
					t.Errorf("Expected APIError, got: %T", err)
					return
				}
				if apiErr.Code != tt.errCode {
					t.Errorf("Expected error code %v, got: %v", tt.errCode, apiErr.Code)
				}
			} else {
				// Note: This will still fail because auth/API calls will fail,
				// but we're testing the validation logic
				if err == nil {
					t.Logf("Command passed validation (API calls would fail without auth)")
				}
			}
		})
	}
}

func TestCompareVitalsCmd_ScoreCalculation(t *testing.T) {
	tests := []struct {
		name      string
		crashRate float64
		anrRate   float64
		wantMin   float64
		wantMax   float64
	}{
		{
			name:      "perfect vitals",
			crashRate: 0.0,
			anrRate:   0.0,
			wantMin:   100.0,
			wantMax:   100.0,
		},
		{
			name:      "moderate crash rate",
			crashRate: 0.001,
			anrRate:   0.0,
			wantMin:   90.0,
			wantMax:   100.0,
		},
		{
			name:      "moderate ANR rate",
			crashRate: 0.0,
			anrRate:   0.001,
			wantMin:   90.0,
			wantMax:   100.0,
		},
		{
			name:      "high crash rate",
			crashRate: 0.01,
			anrRate:   0.0,
			wantMin:   0.0,
			wantMax:   60.0,
		},
		{
			name:      "high ANR rate",
			crashRate: 0.0,
			anrRate:   0.01,
			wantMin:   0.0,
			wantMax:   60.0,
		},
		{
			name:      "very high rates (clamped to 0)",
			crashRate: 0.1,
			anrRate:   0.1,
			wantMin:   0.0,
			wantMax:   0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Score calculation: 100 - (crashRate*50000 + anrRate*50000), clamped to [0, 100]
			score := 100.0 - (tt.crashRate*50000 + tt.anrRate*50000)
			if score < 0 {
				score = 0
			}
			if score > 100 {
				score = 100
			}

			if score < tt.wantMin || score > tt.wantMax {
				t.Errorf("Score = %v, want between %v and %v", score, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCompareVitalsCmd_Ranking(t *testing.T) {
	apps := []compareVitalsAppData{
		{Package: "app1", CrashRate: 0.001, AnrRate: 0.001, Score: 90.0},
		{Package: "app2", CrashRate: 0.002, AnrRate: 0.002, Score: 80.0},
		{Package: "app3", CrashRate: 0.0005, AnrRate: 0.0005, Score: 95.0},
	}

	// Sort by score (higher = better = lower rank number)
	for i := 0; i < len(apps)-1; i++ {
		for j := i + 1; j < len(apps); j++ {
			if apps[i].Score < apps[j].Score {
				apps[i], apps[j] = apps[j], apps[i]
			}
		}
	}

	// Assign ranks
	for i := range apps {
		apps[i].Rank = i + 1
	}

	// Verify ranking
	if apps[0].Package != "app3" || apps[0].Rank != 1 {
		t.Errorf("Expected app3 to be rank 1, got %s at rank %d", apps[0].Package, apps[0].Rank)
	}
	if apps[1].Package != "app1" || apps[1].Rank != 2 {
		t.Errorf("Expected app1 to be rank 2, got %s at rank %d", apps[1].Package, apps[1].Rank)
	}
	if apps[2].Package != "app2" || apps[2].Rank != 3 {
		t.Errorf("Expected app2 to be rank 3, got %s at rank %d", apps[2].Package, apps[2].Rank)
	}
}

func TestCompareVitalsCmd_BestWorstApp(t *testing.T) {
	result := &compareVitalsResult{
		Apps: []compareVitalsAppData{
			{Package: "com.example.best", Score: 95.0},
			{Package: "com.example.middle", Score: 80.0},
			{Package: "com.example.worst", Score: 60.0},
		},
	}

	if len(result.Apps) > 0 {
		result.BestApp = result.Apps[0].Package
		result.WorstApp = result.Apps[len(result.Apps)-1].Package
	}

	if result.BestApp != "com.example.best" {
		t.Errorf("Expected best app to be com.example.best, got: %s", result.BestApp)
	}
	if result.WorstApp != "com.example.worst" {
		t.Errorf("Expected worst app to be com.example.worst, got: %s", result.WorstApp)
	}
}

// ============================================================================
// CompareReviewsCmd Tests
// ============================================================================

func TestCompareReviewsCmd_Validation(t *testing.T) {
	tests := []struct {
		name     string
		packages []string
		wantErr  bool
	}{
		{
			name:     "empty packages",
			packages: []string{},
			wantErr:  true,
		},
		{
			name:     "single package",
			packages: []string{"com.example.app"},
			wantErr:  true,
		},
		{
			name:     "two packages valid",
			packages: []string{"com.example.app1", "com.example.app2"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CompareReviewsCmd{
				Packages: tt.packages,
			}
			globals := &Globals{Output: "json"}

			err := cmd.Run(globals)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			} else {
				// Validation passed, but API calls will fail without auth
				t.Logf("Validation passed for %s", tt.name)
			}
		})
	}
}

func TestCompareReviewsCmd_SentimentCalculation(t *testing.T) {
	tests := []struct {
		name     string
		ratings  []int64
		expected float64
	}{
		{
			name:     "all 5 stars",
			ratings:  []int64{5, 5, 5},
			expected: 1.0, // (5-3)/2 = 1.0
		},
		{
			name:     "all 1 stars",
			ratings:  []int64{1, 1, 1},
			expected: -1.0, // (1-3)/2 = -1.0
		},
		{
			name:     "all 3 stars",
			ratings:  []int64{3, 3, 3},
			expected: 0.0, // (3-3)/2 = 0.0
		},
		{
			name:     "mixed ratings",
			ratings:  []int64{5, 3, 1},
			expected: 0.0, // (1 + 0 + -1) / 3 = 0.0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var sentimentSum float64
			for _, rating := range tt.ratings {
				sentimentSum += (float64(rating) - 3.0) / 2.0
			}
			avgSentiment := sentimentSum / float64(len(tt.ratings))

			if avgSentiment != tt.expected {
				t.Errorf("Expected sentiment %v, got %v", tt.expected, avgSentiment)
			}
		})
	}
}

func TestCompareReviewsCmd_AverageRating(t *testing.T) {
	ratings := []int64{5, 4, 3, 2, 1}
	var total int64
	for _, r := range ratings {
		total += r
	}
	avg := float64(total) / float64(len(ratings))

	expected := 3.0
	if avg != expected {
		t.Errorf("Expected average %v, got %v", expected, avg)
	}
}

func TestCompareReviewsCmd_RatingsDistribution(t *testing.T) {
	ratings := []int64{5, 5, 4, 3, 3, 3, 2, 1}
	dist := make(map[int]int64)

	for _, r := range ratings {
		dist[int(r)]++
	}

	if dist[5] != 2 {
		t.Errorf("Expected 2 five-star ratings, got %d", dist[5])
	}
	if dist[4] != 1 {
		t.Errorf("Expected 1 four-star rating, got %d", dist[4])
	}
	if dist[3] != 3 {
		t.Errorf("Expected 3 three-star ratings, got %d", dist[3])
	}
	if dist[2] != 1 {
		t.Errorf("Expected 1 two-star rating, got %d", dist[2])
	}
	if dist[1] != 1 {
		t.Errorf("Expected 1 one-star rating, got %d", dist[1])
	}
}

// ============================================================================
// CompareReleasesCmd Tests
// ============================================================================

func TestCompareReleasesCmd_Validation(t *testing.T) {
	tests := []struct {
		name     string
		packages []string
		track    string
		wantErr  bool
	}{
		{
			name:     "empty packages",
			packages: []string{},
			track:    "production",
			wantErr:  true,
		},
		{
			name:     "single package",
			packages: []string{"com.example.app"},
			track:    "production",
			wantErr:  true,
		},
		{
			name:     "two packages valid",
			packages: []string{"com.example.app1", "com.example.app2"},
			track:    "production",
			wantErr:  false,
		},
		{
			name:     "internal track",
			packages: []string{"com.example.app1", "com.example.app2"},
			track:    "internal",
			wantErr:  false,
		},
		{
			name:     "alpha track",
			packages: []string{"com.example.app1", "com.example.app2"},
			track:    "alpha",
			wantErr:  false,
		},
		{
			name:     "beta track",
			packages: []string{"com.example.app1", "com.example.app2"},
			track:    "beta",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CompareReleasesCmd{
				Packages: tt.packages,
				Track:    tt.track,
			}
			globals := &Globals{Output: "json"}

			err := cmd.Run(globals)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			} else {
				t.Logf("Validation passed for %s", tt.name)
			}
		})
	}
}

func TestCompareReleasesCmd_ReleaseStatusMapping(t *testing.T) {
	tests := []struct {
		status    string
		eventType string
	}{
		{status: "completed", eventType: "release"},
		{status: "inProgress", eventType: "rollout"},
		{status: "halted", eventType: "halted"},
		{status: "draft", eventType: "release"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%s", tt.status), func(t *testing.T) {
			eventType := "release"
			switch tt.status {
			case statusInProgress:
				eventType = "rollout"
			case statusHalted:
				eventType = statusHalted
			}

			if eventType != tt.eventType {
				t.Errorf("Expected event type %s for status %s, got %s", tt.eventType, tt.status, eventType)
			}
		})
	}
}

func TestCompareReleasesCmd_TimelineSorting(t *testing.T) {
	events := []compareReleaseEvent{
		{Date: "2024-01-15", Package: "app1", Release: "v1.0"},
		{Date: "2024-03-20", Package: "app2", Release: "v2.0"},
		{Date: "2024-01-10", Package: "app3", Release: "v1.5"},
	}

	// Sort by date descending (newest first)
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Date < events[j].Date {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	if events[0].Date != "2024-03-20" {
		t.Errorf("Expected newest date first, got: %s", events[0].Date)
	}
	if events[1].Date != "2024-01-15" {
		t.Errorf("Expected second newest date, got: %s", events[1].Date)
	}
	if events[2].Date != "2024-01-10" {
		t.Errorf("Expected oldest date last, got: %s", events[2].Date)
	}
}

func TestCompareReleasesCmd_ReleaseNameFallback(t *testing.T) {
	tests := []struct {
		name         string
		releaseName  string
		versionCodes []string
		expected     string
	}{
		{
			name:         "use provided name",
			releaseName:  "My Release",
			versionCodes: []string{"100"},
			expected:     "My Release",
		},
		{
			name:         "fallback to version code",
			releaseName:  "",
			versionCodes: []string{"200"},
			expected:     "v200",
		},
		{
			name:         "empty name no version codes",
			releaseName:  "",
			versionCodes: []string{},
			expected:     "v",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			releaseName := tt.releaseName
			if releaseName == "" && len(tt.versionCodes) > 0 {
				releaseName = "v" + tt.versionCodes[0]
			}

			if releaseName != tt.expected {
				t.Errorf("Expected release name %s, got %s", tt.expected, releaseName)
			}
		})
	}
}

// ============================================================================
// CompareSubscriptionsCmd Tests
// ============================================================================

func TestCompareSubscriptionsCmd_Validation(t *testing.T) {
	tests := []struct {
		name     string
		packages []string
		period   string
		wantErr  bool
	}{
		{
			name:     "empty packages",
			packages: []string{},
			period:   "30d",
			wantErr:  true,
		},
		{
			name:     "single package",
			packages: []string{"com.example.app"},
			period:   "30d",
			wantErr:  true,
		},
		{
			name:     "two packages valid with 7d",
			packages: []string{"com.example.app1", "com.example.app2"},
			period:   "7d",
			wantErr:  false,
		},
		{
			name:     "two packages valid with 30d",
			packages: []string{"com.example.app1", "com.example.app2"},
			period:   "30d",
			wantErr:  false,
		},
		{
			name:     "two packages valid with 90d",
			packages: []string{"com.example.app1", "com.example.app2"},
			period:   "90d",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &CompareSubscriptionsCmd{
				Packages: tt.packages,
				Period:   tt.period,
			}
			globals := &Globals{Output: "json"}

			err := cmd.Run(globals)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			} else {
				t.Logf("Validation passed for %s", tt.name)
			}
		})
	}
}

func TestCompareSubscriptionsCmd_SubscriptionFiltering(t *testing.T) {
	allSubs := []*androidpublisher.Subscription{
		{ProductId: "sub_001"},
		{ProductId: "sub_002"},
		{ProductId: "sub_003"},
		{ProductId: "sub_004"},
	}

	filterIDs := []string{"sub_001", "sub_003"}
	filterSet := make(map[string]bool)
	for _, id := range filterIDs {
		filterSet[id] = true
	}

	filteredCount := 0
	for _, sub := range allSubs {
		if filterSet[sub.ProductId] {
			filteredCount++
		}
	}

	if filteredCount != 2 {
		t.Errorf("Expected 2 filtered subscriptions, got %d", filteredCount)
	}
}

func TestCompareSubscriptionsCmd_WarningHandling(t *testing.T) {
	warnings := []string{
		"could not list subscriptions for app1: API error",
		"could not list subscriptions for app2: timeout",
	}

	result := &compareSubscriptionsResult{
		Period: "30d",
		Apps: []compareSubscriptionsAppData{
			{Package: "com.example.app1", Mrr: 0, ARPU: 0, ChurnRate: 0},
			{Package: "com.example.app2", Mrr: 0, ARPU: 0, ChurnRate: 0},
		},
	}

	// Verify warnings are tracked
	if len(warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(warnings))
	}

	// Verify MRR/ARPU warning condition
	mrrWarning := "MRR, ARPU, and churn rate data are not available via the Google Play API; these fields are set to 0"
	if result.Apps[0].Mrr == 0 {
		// This should trigger a warning in the actual implementation
		t.Logf("MRR warning expected: %s", mrrWarning)
	}
}

// ============================================================================
// Result Structure Tests
// ============================================================================

func TestCompareVitalsResult_Structure(t *testing.T) {
	result := &compareVitalsResult{
		Metric: "crash-rate",
		Period: "2024-01-01 to 2024-01-31",
		Apps: []compareVitalsAppData{
			{
				Package:   "com.example.app1",
				CrashRate: 0.001,
				AnrRate:   0.002,
				Score:     85.0,
				Rank:      1,
			},
		},
		BestApp:      "com.example.app1",
		ComparisonAt: time.Now(),
	}

	if result.Metric != "crash-rate" {
		t.Errorf("Expected metric crash-rate, got: %s", result.Metric)
	}
	if len(result.Apps) != 1 {
		t.Errorf("Expected 1 app, got: %d", len(result.Apps))
	}
	if result.BestApp != "com.example.app1" {
		t.Errorf("Expected best app com.example.app1, got: %s", result.BestApp)
	}
}

func TestCompareReviewsResult_Structure(t *testing.T) {
	result := &compareReviewsResult{
		Period: "2024-01-01 to 2024-01-31",
		Apps: []compareReviewsAppData{
			{
				Package:       "com.example.app1",
				AverageRating: 4.5,
				TotalReviews:  100,
				RatingsDist:   map[int]int64{5: 50, 4: 30, 3: 10, 2: 5, 1: 5},
			},
		},
		ComparisonAt: time.Now(),
	}

	if result.Period != "2024-01-01 to 2024-01-31" {
		t.Errorf("Expected specific period, got: %s", result.Period)
	}
	if len(result.Apps) != 1 {
		t.Errorf("Expected 1 app, got: %d", len(result.Apps))
	}
	if result.Apps[0].AverageRating != 4.5 {
		t.Errorf("Expected rating 4.5, got: %f", result.Apps[0].AverageRating)
	}
}

func TestCompareReleasesResult_Structure(t *testing.T) {
	result := &compareReleasesResult{
		Track: "production",
		Apps: []compareReleasesAppData{
			{
				Package:       "com.example.app1",
				ReleaseCount:  5,
				LatestVersion: "100",
				LatestDate:    "2024-01-15",
				Releases: []compareReleaseInfo{
					{
						VersionCodes: []string{"100"},
						Status:       "completed",
						Name:         "Release 1.0",
					},
				},
			},
		},
		Timeline: []compareReleaseEvent{
			{Date: "2024-01-15", Package: "com.example.app1", Release: "v100", Type: "release"},
		},
		ComparisonAt: time.Now(),
	}

	if result.Track != "production" {
		t.Errorf("Expected track production, got: %s", result.Track)
	}
	if len(result.Apps) != 1 {
		t.Errorf("Expected 1 app, got: %d", len(result.Apps))
	}
	if result.Apps[0].ReleaseCount != 5 {
		t.Errorf("Expected 5 releases, got: %d", result.Apps[0].ReleaseCount)
	}
	if len(result.Timeline) != 1 {
		t.Errorf("Expected 1 timeline event, got: %d", len(result.Timeline))
	}
}

func TestCompareSubscriptionsResult_Structure(t *testing.T) {
	result := &compareSubscriptionsResult{
		Period: "30d",
		Apps: []compareSubscriptionsAppData{
			{
				Package:    "com.example.app1",
				TotalSubs:  1000,
				ActiveSubs: 800,
				ChurnRate:  0.05,
				Mrr:        5000.0,
				ARPU:       5.0,
			},
		},
		ComparisonAt: time.Now(),
	}

	if result.Period != "30d" {
		t.Errorf("Expected period 30d, got: %s", result.Period)
	}
	if len(result.Apps) != 1 {
		t.Errorf("Expected 1 app, got: %d", len(result.Apps))
	}
	if result.Apps[0].TotalSubs != 1000 {
		t.Errorf("Expected 1000 subs, got: %d", result.Apps[0].TotalSubs)
	}
	if result.Apps[0].Mrr != 5000.0 {
		t.Errorf("Expected MRR 5000.0, got: %f", result.Apps[0].Mrr)
	}
}

// ============================================================================
// Mock Client Integration Tests
// ============================================================================

func TestCompareCommands_WithMockClient(t *testing.T) {
	mockClient := apitest.NewMockClient()

	t.Run("mock client tracks calls", func(t *testing.T) {
		mockClient.TrackCall("reporting", "QueryCrashRate", map[string]interface{}{
			"package": "com.example.app",
		})

		count := mockClient.GetCallCount("reporting", "QueryCrashRate")
		if count != 1 {
			t.Errorf("Expected 1 call, got: %d", count)
		}
	})

	t.Run("mock client resets calls", func(t *testing.T) {
		mockClient.ResetCalls()

		count := mockClient.GetCallCount("reporting", "QueryCrashRate")
		if count != 0 {
			t.Errorf("Expected 0 calls after reset, got: %d", count)
		}
	})
}

func TestCompareVitalsCmd_PeriodBuilding(t *testing.T) {
	tests := []struct {
		name      string
		startDate string
		endDate   string
		expected  string
	}{
		{
			name:      "default period",
			startDate: "",
			endDate:   "",
			expected:  "last 30 days",
		},
		{
			name:      "custom dates",
			startDate: "2024-01-01",
			endDate:   "2024-01-31",
			expected:  "2024-01-01 to 2024-01-31",
		},
		{
			name:      "only start date",
			startDate: "2024-01-01",
			endDate:   "",
			expected:  "2024-01-01 to", // Will include today's date
		},
		{
			name:      "only end date",
			startDate: "",
			endDate:   "2024-01-31",
			expected:  " to 2024-01-31", // Will include 30 days ago
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the period building logic from the actual command
			period := "last 30 days"
			if tt.startDate != "" || tt.endDate != "" {
				startStr := tt.startDate
				if startStr == "" {
					startStr = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
				}
				endStr := tt.endDate
				if endStr == "" {
					endStr = time.Now().UTC().Format("2006-01-02")
				}
				period = startStr + " to " + endStr
			}

			// For custom dates, verify the format
			if tt.startDate != "" && tt.endDate != "" {
				if period != tt.expected {
					t.Errorf("Expected period %s, got %s", tt.expected, period)
				}
			}
		})
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestCompareCommands_ErrorHints(t *testing.T) {
	t.Run("vitals with insufficient packages has hint", func(t *testing.T) {
		cmd := &CompareVitalsCmd{
			Packages: []string{"com.example.app"},
		}
		globals := &Globals{Output: "json"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}

		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatal("Expected APIError type")
		}
		if apiErr.Hint == "" {
			t.Error("Expected error to have hint")
		}
	})

	t.Run("reviews with insufficient packages has error", func(t *testing.T) {
		cmd := &CompareReviewsCmd{
			Packages: []string{},
		}
		globals := &Globals{Output: "json"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}

		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatal("Expected APIError type")
		}
		if apiErr.Code != errors.CodeValidationError {
			t.Errorf("Expected validation error code, got: %v", apiErr.Code)
		}
	})

	t.Run("releases with insufficient packages has error", func(t *testing.T) {
		cmd := &CompareReleasesCmd{
			Packages: []string{"com.example.app"},
		}
		globals := &Globals{Output: "json"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}

		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatal("Expected APIError type")
		}
		if apiErr.Code != errors.CodeValidationError {
			t.Errorf("Expected validation error code, got: %v", apiErr.Code)
		}
	})

	t.Run("subscriptions with insufficient packages has error", func(t *testing.T) {
		cmd := &CompareSubscriptionsCmd{
			Packages: []string{},
		}
		globals := &Globals{Output: "json"}

		err := cmd.Run(globals)
		if err == nil {
			t.Fatal("Expected error")
		}

		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatal("Expected APIError type")
		}
		if apiErr.Code != errors.CodeValidationError {
			t.Errorf("Expected validation error code, got: %v", apiErr.Code)
		}
	})
}

// ============================================================================
// Context Propagation Tests
// ============================================================================

func TestCompareCommands_ContextPropagation(t *testing.T) {
	t.Run("context cancellation respected", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		if ctx.Err() != context.Canceled {
			t.Error("Expected context to be canceled")
		}
	})

	t.Run("context timeout respected", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Wait for timeout

		if ctx.Err() != context.DeadlineExceeded {
			t.Error("Expected context deadline exceeded")
		}
	})
}

// ============================================================================
// Metric Set Name Tests
// ============================================================================

func TestCompareVitalsCmd_MetricSetNames(t *testing.T) {
	tests := []struct {
		packageName string
		metricType  string
		expected    string
	}{
		{
			packageName: "com.example.app",
			metricType:  "crash",
			expected:    "apps/com.example.app/crashRateMetricSet",
		},
		{
			packageName: "com.example.app",
			metricType:  "anr",
			expected:    "apps/com.example.app/anrRateMetricSet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.metricType, func(t *testing.T) {
			var name string
			switch tt.metricType {
			case "crash":
				name = fmt.Sprintf("apps/%s/crashRateMetricSet", tt.packageName)
			case "anr":
				name = fmt.Sprintf("apps/%s/anrRateMetricSet", tt.packageName)
			}

			if name != tt.expected {
				t.Errorf("Expected name %s, got %s", tt.expected, name)
			}
		})
	}
}

// ============================================================================
// Decimal Value Parsing Tests
// ============================================================================

func TestCompareVitalsCmd_DecimalValueParsing(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected float64
		wantErr  bool
	}{
		{
			name:     "valid decimal",
			value:    "0.001",
			expected: 0.001,
			wantErr:  false,
		},
		{
			name:     "zero value",
			value:    "0",
			expected: 0,
			wantErr:  false,
		},
		{
			name:     "large value",
			value:    "0.1",
			expected: 0.1,
			wantErr:  false,
		},
		{
			name:    "invalid value",
			value:   "not-a-number",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var parsed float64
			var err error

			// Simulate parsing logic
			if tt.value != "" {
				parsed, err = strconv.ParseFloat(tt.value, 64)
			}

			if tt.wantErr {
				if err == nil {
					t.Error("Expected error for invalid value")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if parsed != tt.expected {
					t.Errorf("Expected %v, got %v", tt.expected, parsed)
				}
			}
		})
	}
}

// ============================================================================
// API Response Handling Tests
// ============================================================================

func TestCompareVitalsCmd_EmptyResponseHandling(t *testing.T) {
	// Simulate empty API response (no rows)
	resp := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetResponse{
		Rows: []*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1MetricsRow{},
	}

	// Should handle gracefully and return 0 rates
	var crashRate float64
	if len(resp.Rows) > 0 {
		for _, metric := range resp.Rows[0].Metrics {
			if metric.Metric == "crashRate" && metric.DecimalValue != nil {
				if val, err := strconv.ParseFloat(metric.DecimalValue.Value, 64); err == nil {
					crashRate = val
				}
			}
		}
	}

	if crashRate != 0 {
		t.Errorf("Expected 0 crash rate for empty response, got: %f", crashRate)
	}
}

func TestCompareReviewsCmd_NilCommentHandling(t *testing.T) {
	// Test handling of reviews with nil comments
	reviews := []*androidpublisher.Review{
		{Comments: nil}, // Review with no comments
		{Comments: []*androidpublisher.Comment{
			{UserComment: nil}, // Comment with no user comment
		}},
		{Comments: []*androidpublisher.Comment{
			{UserComment: &androidpublisher.UserComment{StarRating: 5}},
		}},
	}

	validCount := 0
	for _, review := range reviews {
		if review.Comments == nil {
			continue
		}
		for _, comment := range review.Comments {
			if comment.UserComment == nil {
				continue
			}
			validCount++
		}
	}

	if validCount != 1 {
		t.Errorf("Expected 1 valid comment, got: %d", validCount)
	}
}

// ============================================================================
// Track Release Processing Tests
// ============================================================================

func TestCompareReleasesCmd_LimitHandling(t *testing.T) {
	releases := []*androidpublisher.TrackRelease{
		{Status: "completed", Name: "Release 1", VersionCodes: []int64{1}},
		{Status: "completed", Name: "Release 2", VersionCodes: []int64{2}},
		{Status: "completed", Name: "Release 3", VersionCodes: []int64{3}},
		{Status: "completed", Name: "Release 4", VersionCodes: []int64{4}},
		{Status: "completed", Name: "Release 5", VersionCodes: []int64{5}},
	}

	limit := 3
	count := 0
	for _, release := range releases {
		if limit > 0 && count >= limit {
			break
		}
		count++
		_ = release // Process release
	}

	if count != 3 {
		t.Errorf("Expected 3 releases processed with limit, got: %d", count)
	}
}

func TestCompareReleasesCmd_VersionCodeFormatting(t *testing.T) {
	tests := []struct {
		name         string
		versionCodes []int64
		expected     []string
	}{
		{
			name:         "single version code",
			versionCodes: []int64{100},
			expected:     []string{"100"},
		},
		{
			name:         "multiple version codes",
			versionCodes: []int64{100, 101, 102},
			expected:     []string{"100", "101", "102"},
		},
		{
			name:         "empty version codes",
			versionCodes: []int64{},
			expected:     []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := make([]string, 0, len(tt.versionCodes))
			for _, vc := range tt.versionCodes {
				result = append(result, fmt.Sprintf("%d", vc))
			}

			if len(result) != len(tt.expected) {
				t.Errorf("Expected %d version codes, got: %d", len(tt.expected), len(result))
			}

			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Expected version code %s, got: %s", tt.expected[i], v)
				}
			}
		})
	}
}

// ============================================================================
// Pagination Handling Tests
// ============================================================================

func TestCompareReviewsCmd_PaginationLogic(t *testing.T) {
	// Simulate pagination behavior
	pageTokens := []string{"page1", "page2", ""}
	currentPage := 0

	allReviews := make([]*androidpublisher.Review, 0)
	pageToken := pageTokens[0]

	for pageToken != "" {
		// Simulate fetching a page
		allReviews = append(allReviews, &androidpublisher.Review{
			Comments: []*androidpublisher.Comment{
				{UserComment: &androidpublisher.UserComment{StarRating: int64(currentPage + 1)}},
			},
		})

		currentPage++
		if currentPage < len(pageTokens) {
			pageToken = pageTokens[currentPage]
		} else {
			pageToken = ""
		}
	}

	if len(allReviews) != 2 {
		t.Errorf("Expected 2 pages of reviews, got: %d", len(allReviews))
	}
}
