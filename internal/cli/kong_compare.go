// Package cli provides app comparison commands for analyzing multiple apps.
package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

const (
	statusInProgress = "inProgress"
	statusHalted     = "halted"
)

// CompareCmd contains app comparison commands.
type CompareCmd struct {
	Vitals        CompareVitalsCmd        `cmd:"" help:"Compare vitals metrics across multiple apps"`
	Reviews       CompareReviewsCmd       `cmd:"" help:"Compare review metrics across apps"`
	Releases      CompareReleasesCmd      `cmd:"" help:"Compare release history across apps"`
	Subscriptions CompareSubscriptionsCmd `cmd:"" help:"Compare subscription metrics"`
}

// CompareVitalsCmd compares vitals metrics across multiple apps.
type CompareVitalsCmd struct {
	Packages  []string `help:"Package names to compare (repeatable)" required:""`
	Metric    string   `help:"Metric to compare" default:"all" enum:"crash-rate,anr-rate,error-rate,all"`
	StartDate string   `help:"Start date (ISO 8601)"`
	EndDate   string   `help:"End date (ISO 8601)"`
	Format    string   `help:"Output format" default:"table" enum:"json,table,csv"`
}

// compareVitalsResult represents vitals comparison result.
type compareVitalsResult struct {
	Metric       string                 `json:"metric"`
	Period       string                 `json:"period"`
	Apps         []compareVitalsAppData `json:"apps"`
	BestApp      string                 `json:"bestApp,omitempty"`
	WorstApp     string                 `json:"worstApp,omitempty"`
	ComparisonAt time.Time              `json:"comparisonAt"`
}

// compareVitalsAppData represents data for a single app.
type compareVitalsAppData struct {
	Package    string  `json:"package"`
	CrashRate  float64 `json:"crashRate,omitempty"`
	AnrRate    float64 `json:"anrRate,omitempty"`
	ErrorCount int64   `json:"errorCount,omitempty"`
	Score      float64 `json:"score"`
	Rank       int     `json:"rank"`
}

// Run executes the compare vitals command.
func (cmd *CompareVitalsCmd) Run(globals *Globals) error {
	if len(cmd.Packages) < 2 {
		return errors.NewAPIError(errors.CodeValidationError, "at least 2 packages are required for comparison").
			WithHint("Provide multiple package names with --packages flag")
	}

	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Comparing vitals for %d apps\n", len(cmd.Packages))
	}

	// Create authenticated API client
	ctx := context.Background()
	authMgr := newAuthManager()
	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
	if err != nil {
		return err
	}

	svc, err := client.PlayReporting()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get reporting service: %v", err))
	}

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	startTime := time.Now()

	// Build period string for result
	period := "last 30 days"
	if cmd.StartDate != "" || cmd.EndDate != "" {
		startStr := cmd.StartDate
		if startStr == "" {
			startStr = time.Now().UTC().AddDate(0, 0, -30).Format("2006-01-02")
		}
		endStr := cmd.EndDate
		if endStr == "" {
			endStr = time.Now().UTC().Format("2006-01-02")
		}
		period = startStr + " to " + endStr
	}

	// Query vitals for each package
	result := &compareVitalsResult{
		Metric:       cmd.Metric,
		Period:       period,
		Apps:         make([]compareVitalsAppData, 0, len(cmd.Packages)),
		ComparisonAt: time.Now(),
	}

	for _, pkg := range cmd.Packages {
		appData := compareVitalsAppData{
			Package: pkg,
		}

		if err := client.Acquire(ctx); err != nil {
			return err
		}

		// Query crash rate if requested
		if cmd.Metric == "crash-rate" || cmd.Metric == checkAll {
			crashRate, qerr := cmd.queryCrashRate(ctx, client, svc, pkg, timelineSpec)
			if qerr != nil {
				client.Release()
				return qerr
			}
			appData.CrashRate = crashRate
		}

		// Query ANR rate if requested
		if cmd.Metric == "anr-rate" || cmd.Metric == checkAll {
			anrRate, qerr := cmd.queryAnrRate(ctx, client, svc, pkg, timelineSpec)
			if qerr != nil {
				client.Release()
				return qerr
			}
			appData.AnrRate = anrRate
		}

		client.Release()

		// Calculate composite score (lower is better for crash/anr rates)
		// Score = 100 - (crashRate * 50000 + anrRate * 50000), clamped to [0, 100]
		score := 100.0 - (appData.CrashRate*50000 + appData.AnrRate*50000)
		if score < 0 {
			score = 0
		}
		if score > 100 {
			score = 100
		}
		appData.Score = score

		result.Apps = append(result.Apps, appData)
	}

	// Rank apps by score (higher score = better = lower rank number)
	sort.Slice(result.Apps, func(i, j int) bool {
		return result.Apps[i].Score > result.Apps[j].Score
	})
	for i := range result.Apps {
		result.Apps[i].Rank = i + 1
	}

	if len(result.Apps) > 0 {
		result.BestApp = result.Apps[0].Package
		result.WorstApp = result.Apps[len(result.Apps)-1].Package
	}

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting"))
}

func (cmd *CompareVitalsCmd) queryCrashRate(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) (float64, error) {
	crashName := fmt.Sprintf("apps/%s/crashRateMetricSet", pkg)
	crashReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{metricCrashRate, "crashRate7dUserWeighted"},
		PageSize:     1,
	}

	var crashRate float64
	err := client.DoWithRetry(ctx, func() error {
		resp, qerr := svc.Vitals.Crashrate.Query(crashName, crashReq).Context(ctx).Do()
		if qerr != nil {
			return qerr
		}
		if len(resp.Rows) > 0 {
			for _, metric := range resp.Rows[0].Metrics {
				if metric.Metric == metricCrashRate && metric.DecimalValue != nil {
					if val, perr := strconv.ParseFloat(metric.DecimalValue.Value, 64); perr == nil {
						crashRate = val
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to query crash rate for %s: %v", pkg, err))
	}
	return crashRate, nil
}

func (cmd *CompareVitalsCmd) queryAnrRate(ctx context.Context, client *api.Client, svc *playdeveloperreporting.Service, pkg string, timelineSpec *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec) (float64, error) {
	anrName := fmt.Sprintf("apps/%s/anrRateMetricSet", pkg)
	anrReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{metricAnrRate, "anrRate7dUserWeighted"},
		PageSize:     1,
	}

	var anrRate float64
	err := client.DoWithRetry(ctx, func() error {
		resp, qerr := svc.Vitals.Anrrate.Query(anrName, anrReq).Context(ctx).Do()
		if qerr != nil {
			return qerr
		}
		if len(resp.Rows) > 0 {
			for _, metric := range resp.Rows[0].Metrics {
				if metric.Metric == metricAnrRate && metric.DecimalValue != nil {
					if val, perr := strconv.ParseFloat(metric.DecimalValue.Value, 64); perr == nil {
						anrRate = val
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return 0, errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to query ANR rate for %s: %v", pkg, err))
	}
	return anrRate, nil
}

// CompareReviewsCmd compares review metrics across apps.
type CompareReviewsCmd struct {
	Packages         []string `help:"Package names to compare (repeatable)" required:""`
	StartDate        string   `help:"Start date (ISO 8601)"`
	EndDate          string   `help:"End date (ISO 8601)"`
	IncludeSentiment bool     `help:"Include sentiment analysis"`
	Format           string   `help:"Output format" default:"table" enum:"json,table,csv"`
}

// compareReviewsResult represents reviews comparison result.
type compareReviewsResult struct {
	Period       string                  `json:"period"`
	Apps         []compareReviewsAppData `json:"apps"`
	ComparisonAt time.Time               `json:"comparisonAt"`
}

// compareReviewsAppData represents review data for a single app.
type compareReviewsAppData struct {
	Package        string        `json:"package"`
	AverageRating  float64       `json:"averageRating"`
	TotalReviews   int64         `json:"totalReviews"`
	RatingsDist    map[int]int64 `json:"ratingsDistribution"`
	SentimentScore float64       `json:"sentimentScore,omitempty"`
}

// Run executes the compare reviews command.
func (cmd *CompareReviewsCmd) Run(globals *Globals) error {
	if len(cmd.Packages) < 2 {
		return errors.NewAPIError(errors.CodeValidationError, "at least 2 packages are required for comparison")
	}

	// Create authenticated API client
	ctx := context.Background()
	authMgr := newAuthManager()
	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get publisher service: %v", err))
	}

	startTime := time.Now()

	result := &compareReviewsResult{
		Apps:         make([]compareReviewsAppData, 0, len(cmd.Packages)),
		ComparisonAt: time.Now(),
	}

	for _, pkg := range cmd.Packages {
		appData := compareReviewsAppData{
			Package:     pkg,
			RatingsDist: make(map[int]int64),
		}

		if err := client.Acquire(ctx); err != nil {
			return err
		}

		var allReviews []*androidpublisher.Review

		err = client.DoWithRetry(ctx, func() error {
			resp, qerr := svc.Reviews.List(pkg).Context(ctx).Do()
			if qerr != nil {
				return qerr
			}
			allReviews = append(allReviews, resp.Reviews...)

			// Fetch additional pages for comprehensive data
			pageToken := resp.TokenPagination
			for pageToken != nil && pageToken.NextPageToken != "" {
				nextResp, nerr := svc.Reviews.List(pkg).Token(pageToken.NextPageToken).Context(ctx).Do()
				if nerr != nil {
					return nerr
				}
				allReviews = append(allReviews, nextResp.Reviews...)
				pageToken = nextResp.TokenPagination
			}
			return nil
		})

		client.Release()

		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to list reviews for %s: %v", pkg, err))
		}

		// Calculate average rating and distribution from returned reviews
		var totalRating int64
		var reviewCount int64
		var sentimentSum float64

		for _, review := range allReviews {
			if review.Comments == nil {
				continue
			}
			for _, comment := range review.Comments {
				if comment.UserComment == nil {
					continue
				}
				rating := int(comment.UserComment.StarRating)
				appData.RatingsDist[rating]++
				totalRating += comment.UserComment.StarRating
				reviewCount++

				// Use star rating as a proxy for sentiment if requested
				if cmd.IncludeSentiment {
					// Normalize star rating to [-1, 1] range
					sentimentSum += (float64(comment.UserComment.StarRating) - 3.0) / 2.0
				}
			}
		}

		appData.TotalReviews = reviewCount
		if reviewCount > 0 {
			appData.AverageRating = float64(totalRating) / float64(reviewCount)
			if cmd.IncludeSentiment {
				appData.SentimentScore = sentimentSum / float64(reviewCount)
			}
		}

		result.Apps = append(result.Apps, appData)
	}

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher"))
}

// CompareReleasesCmd compares release history across apps.
type CompareReleasesCmd struct {
	Packages []string `help:"Package names to compare (repeatable)" required:""`
	Track    string   `help:"Track to compare" default:"production" enum:"internal,alpha,beta,production"`
	Since    string   `help:"Compare releases since this date"`
	Limit    int      `help:"Maximum releases per app" default:"10"`
}

// compareReleasesResult represents releases comparison result.
type compareReleasesResult struct {
	Track        string                   `json:"track"`
	Apps         []compareReleasesAppData `json:"apps"`
	Timeline     []compareReleaseEvent    `json:"timeline"`
	ComparisonAt time.Time                `json:"comparisonAt"`
}

// compareReleasesAppData represents release data for a single app.
type compareReleasesAppData struct {
	Package       string               `json:"package"`
	ReleaseCount  int                  `json:"releaseCount"`
	LatestVersion string               `json:"latestVersion,omitempty"`
	LatestDate    string               `json:"latestDate,omitempty"`
	Releases      []compareReleaseInfo `json:"releases"`
}

// compareReleaseInfo represents a single release.
type compareReleaseInfo struct {
	VersionCodes []string `json:"versionCodes"`
	Status       string   `json:"status"`
	Date         string   `json:"date,omitempty"`
	Name         string   `json:"name,omitempty"`
}

// compareReleaseEvent represents an event on the timeline.
type compareReleaseEvent struct {
	Date    string `json:"date"`
	Package string `json:"package"`
	Release string `json:"release"`
	Type    string `json:"type"`
}

// Run executes the compare releases command.
func (cmd *CompareReleasesCmd) Run(globals *Globals) error {
	if len(cmd.Packages) < 2 {
		return errors.NewAPIError(errors.CodeValidationError, "at least 2 packages are required for comparison")
	}

	// Create authenticated API client
	ctx := context.Background()
	authMgr := newAuthManager()
	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get publisher service: %v", err))
	}

	startTime := time.Now()

	result := &compareReleasesResult{
		Track:        cmd.Track,
		Apps:         make([]compareReleasesAppData, 0, len(cmd.Packages)),
		Timeline:     make([]compareReleaseEvent, 0),
		ComparisonAt: time.Now(),
	}

	for _, pkg := range cmd.Packages {
		appData := compareReleasesAppData{
			Package:  pkg,
			Releases: make([]compareReleaseInfo, 0),
		}

		if err := client.Acquire(ctx); err != nil {
			return err
		}

		// Create temporary edit
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		if err != nil {
			client.Release()
			return errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to create edit for %s: %v", pkg, err))
		}

		editID := edit.Id

		// Get track info
		var track *androidpublisher.Track
		err = client.DoWithRetry(ctx, func() error {
			track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
			return err
		})

		// Clean up the temporary edit
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})

		client.Release()

		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to get track for %s: %v", pkg, err))
		}

		// Process releases from the track
		if track != nil && track.Releases != nil {
			count := 0
			for _, release := range track.Releases {
				if cmd.Limit > 0 && count >= cmd.Limit {
					break
				}

				versionCodes := make([]string, 0, len(release.VersionCodes))
				for _, vc := range release.VersionCodes {
					versionCodes = append(versionCodes, fmt.Sprintf("%d", vc))
				}

				releaseInfo := compareReleaseInfo{
					VersionCodes: versionCodes,
					Status:       release.Status,
					Name:         release.Name,
				}

				appData.Releases = append(appData.Releases, releaseInfo)
				count++

				// Build timeline events
				eventType := "release"
				switch release.Status {
				case statusInProgress:
					eventType = "rollout"
				case statusHalted:
					eventType = statusHalted
				}

				releaseName := release.Name
				if releaseName == "" && len(versionCodes) > 0 {
					releaseName = "v" + versionCodes[0]
				}

				result.Timeline = append(result.Timeline, compareReleaseEvent{
					Date:    time.Now().Format("2006-01-02"),
					Package: pkg,
					Release: releaseName,
					Type:    eventType,
				})
			}
			appData.ReleaseCount = count

			// Set latest version from the first release (most recent)
			if len(track.Releases) > 0 {
				latest := track.Releases[0]
				if len(latest.VersionCodes) > 0 {
					appData.LatestVersion = fmt.Sprintf("%d", latest.VersionCodes[0])
				}
				if latest.Name != "" {
					appData.LatestVersion = latest.Name
				}
			}
		}

		result.Apps = append(result.Apps, appData)
	}

	// Sort timeline by date
	sort.Slice(result.Timeline, func(i, j int) bool {
		return result.Timeline[i].Date > result.Timeline[j].Date
	})

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher"))
}

// CompareSubscriptionsCmd compares subscription metrics across apps.
type CompareSubscriptionsCmd struct {
	Packages      []string `help:"Package names to compare (repeatable)" required:""`
	Subscriptions []string `help:"Specific subscription IDs to compare"`
	Period        string   `help:"Comparison period" default:"30d" enum:"7d,30d,90d"`
}

// compareSubscriptionsResult represents subscriptions comparison result.
type compareSubscriptionsResult struct {
	Period       string                        `json:"period"`
	Apps         []compareSubscriptionsAppData `json:"apps"`
	ComparisonAt time.Time                     `json:"comparisonAt"`
}

// compareSubscriptionsAppData represents subscription data for a single app.
type compareSubscriptionsAppData struct {
	Package    string  `json:"package"`
	TotalSubs  int64   `json:"totalSubscriptions"`
	ActiveSubs int64   `json:"activeSubscriptions"`
	ChurnRate  float64 `json:"churnRate"`
	Mrr        float64 `json:"mrr,omitempty"`  // Monthly Recurring Revenue
	ARPU       float64 `json:"arpu,omitempty"` // Average Revenue Per User
}

// Run executes the compare subscriptions command.
func (cmd *CompareSubscriptionsCmd) Run(globals *Globals) error {
	if len(cmd.Packages) < 2 {
		return errors.NewAPIError(errors.CodeValidationError, "at least 2 packages are required for comparison")
	}

	// Create authenticated API client
	ctx := context.Background()
	authMgr := newAuthManager()
	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get publisher service: %v", err))
	}

	startTime := time.Now()

	result := &compareSubscriptionsResult{
		Period:       cmd.Period,
		Apps:         make([]compareSubscriptionsAppData, 0, len(cmd.Packages)),
		ComparisonAt: time.Now(),
	}

	var warnings []string

	for _, pkg := range cmd.Packages {
		appData := compareSubscriptionsAppData{
			Package: pkg,
		}

		if err := client.Acquire(ctx); err != nil {
			return err
		}

		// List subscriptions using Monetization API
		var subscriptionCount int64
		err = client.DoWithRetry(ctx, func() error {
			// No direct filter on the API; we filter after retrieval if specific subscriptions are requested
			call := svc.Monetization.Subscriptions.List(pkg)
			resp, lerr := call.Context(ctx).Do()
			if lerr != nil {
				return lerr
			}
			if resp != nil && resp.Subscriptions != nil {
				subscriptionCount = int64(len(resp.Subscriptions))

				// Filter by specific subscription IDs if provided
				if len(cmd.Subscriptions) > 0 {
					filteredCount := int64(0)
					subSet := make(map[string]bool)
					for _, subID := range cmd.Subscriptions {
						subSet[subID] = true
					}
					for _, sub := range resp.Subscriptions {
						if subSet[sub.ProductId] {
							filteredCount++
						}
					}
					subscriptionCount = filteredCount
				}
			}
			return nil
		})

		client.Release()

		if err != nil {
			// Monetization API may not be available for all accounts
			warnings = append(warnings, fmt.Sprintf("could not list subscriptions for %s: %v", pkg, err))
			appData.TotalSubs = 0
		} else {
			appData.TotalSubs = subscriptionCount
			appData.ActiveSubs = subscriptionCount // Active count approximation from list
		}

		// MRR/ARPU/churn data is not available via the Monetization API
		// These would require financial reports or Play Console export data
		appData.ChurnRate = 0
		appData.Mrr = 0
		appData.ARPU = 0

		result.Apps = append(result.Apps, appData)
	}

	r := output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher")

	if len(warnings) > 0 {
		r = r.WithWarnings(warnings...)
	}
	if result.Apps[0].Mrr == 0 {
		r = r.WithWarnings("MRR, ARPU, and churn rate data are not available via the Google Play API; these fields are set to 0")
	}

	return writeOutput(globals, r)
}
