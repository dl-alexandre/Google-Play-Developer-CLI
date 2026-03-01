// Package cli provides release management commands for release lifecycle.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

// actionGet is the constant for the "get" release notes action.
const actionGet = "get"

// Constants for release management operations.
const (
	trackAll          = "all"
	eventTypeRelease  = "release"
	eventTypeRollout  = "rollout"
	recommendContinue = "continue"
	actionCopy        = "copy"
)

// ReleaseMgmtCmd contains release management commands.
type ReleaseMgmtCmd struct {
	Calendar  ReleaseCalendarCmd  `cmd:"" help:"Show upcoming and past releases"`
	Conflicts ReleaseConflictsCmd `cmd:"" help:"Detect version code conflicts"`
	Strategy  ReleaseStrategyCmd  `cmd:"" help:"Get rollback/roll-forward recommendations"`
	History   ReleaseHistoryCmd   `cmd:"" help:"Show detailed release history"`
	Notes     ReleaseNotesCmd     `cmd:"" help:"Manage release notes across locales"`
}

// ReleaseCalendarCmd shows upcoming and past releases.
type ReleaseCalendarCmd struct {
	Track      string `help:"Track to show calendar for" default:"all" enum:"internal,alpha,beta,production,all"`
	DaysAhead  int    `help:"Days to look ahead" default:"30"`
	DaysBehind int    `help:"Days to look back" default:"30"`
	Format     string `help:"Output format" default:"table" enum:"json,table,markdown"`
}

// releaseCalendarResult represents the calendar result.
type releaseCalendarResult struct {
	Track       string                 `json:"track,omitempty"`
	PeriodStart string                 `json:"periodStart"`
	PeriodEnd   string                 `json:"periodEnd"`
	Events      []releaseCalendarEvent `json:"events"`
	GeneratedAt time.Time              `json:"generatedAt"`
}

// releaseCalendarEvent represents a calendar event.
type releaseCalendarEvent struct {
	Date        string `json:"date"`
	Type        string `json:"type"`
	Track       string `json:"track"`
	VersionCode string `json:"versionCode,omitempty"`
	Description string `json:"description"`
}

// Run executes the release calendar command.
func (cmd *ReleaseCalendarCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
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
	now := time.Now()
	startDate := now.AddDate(0, 0, -cmd.DaysBehind)
	endDate := now.AddDate(0, 0, cmd.DaysAhead)

	result := &releaseCalendarResult{
		Track:       cmd.Track,
		PeriodStart: startDate.Format("2006-01-02"),
		PeriodEnd:   endDate.Format("2006-01-02"),
		Events:      make([]releaseCalendarEvent, 0),
		GeneratedAt: now,
	}

	// Create temporary edit to read track data
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	// List all tracks
	var tracksList *androidpublisher.TracksListResponse
	err = client.DoWithRetry(ctx, func() error {
		tracksList, err = svc.Edits.Tracks.List(globals.Package, editID).Context(ctx).Do()
		return err
	})

	// Clean up the temporary edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(globals.Package, editID).Context(ctx).Do()
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
	}

	// Build calendar events from track/release data
	for _, track := range tracksList.Tracks {
		// Filter by track if specified
		if cmd.Track != trackAll && track.Track != cmd.Track {
			continue
		}

		for _, release := range track.Releases {
			versionCode := ""
			if len(release.VersionCodes) > 0 {
				versionCode = fmt.Sprintf("%d", release.VersionCodes[0])
			}

			eventType := eventTypeRelease
			description := ""

			switch release.Status {
			case releaseCompleted:
				eventType = releaseCompleted
				description = fmt.Sprintf("Completed release %s on %s", release.Name, track.Track)
			case statusInProgress:
				eventType = eventTypeRollout
				rolloutPct := release.UserFraction * 100
				description = fmt.Sprintf("Rolling out %s on %s (%.1f%%)", release.Name, track.Track, rolloutPct)
			case statusHalted:
				eventType = statusHalted
				description = fmt.Sprintf("Halted release %s on %s", release.Name, track.Track)
			case releaseStatusDraft:
				eventType = releaseStatusDraft
				description = fmt.Sprintf("Draft release %s on %s", release.Name, track.Track)
			default:
				description = fmt.Sprintf("Release %s on %s (status: %s)", release.Name, track.Track, release.Status)
			}

			result.Events = append(result.Events, releaseCalendarEvent{
				Date:        now.Format("2006-01-02"),
				Type:        eventType,
				Track:       track.Track,
				VersionCode: versionCode,
				Description: description,
			})
		}
	}

	// Sort events by date
	sort.Slice(result.Events, func(i, j int) bool {
		return result.Events[i].Date < result.Events[j].Date
	})

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher"))
}

// ReleaseConflictsCmd detects version code conflicts.
type ReleaseConflictsCmd struct {
	VersionCodes []string `help:"Version codes to check (repeatable)"`
	CheckTrack   string   `help:"Specific track to check" default:"all" enum:"internal,alpha,beta,production,all"`
	SuggestFix   bool     `help:"Suggest fixes for conflicts"`
}

// releaseConflictsResult represents the conflicts check result.
type releaseConflictsResult struct {
	HasConflicts bool              `json:"hasConflicts"`
	Conflicts    []releaseConflict `json:"conflicts"`
	Suggestions  []string          `json:"suggestions,omitempty"`
	CheckedAt    time.Time         `json:"checkedAt"`
}

// releaseConflict represents a single conflict.
type releaseConflict struct {
	VersionCode     string `json:"versionCode"`
	Track           string `json:"track"`
	Status          string `json:"status"`
	ExistingVersion string `json:"existingVersion,omitempty"`
}

// Run executes the release conflicts command.
func (cmd *ReleaseConflictsCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if len(cmd.VersionCodes) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "at least one version code is required").
			WithHint("Provide version codes with --version-codes flag")
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

	result := &releaseConflictsResult{
		Conflicts:   make([]releaseConflict, 0),
		Suggestions: make([]string, 0),
		CheckedAt:   time.Now(),
	}

	// Build a set of requested version codes
	requestedVCs := make(map[string]bool)
	for _, vc := range cmd.VersionCodes {
		requestedVCs[vc] = true
	}

	// Create temporary edit
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	// List all tracks and collect version codes
	var tracksList *androidpublisher.TracksListResponse
	err = client.DoWithRetry(ctx, func() error {
		tracksList, err = svc.Edits.Tracks.List(globals.Package, editID).Context(ctx).Do()
		return err
	})

	// Clean up the temporary edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(globals.Package, editID).Context(ctx).Do()
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
	}

	// Check for version code conflicts across tracks
	// Track which version codes appear on which tracks
	var maxExistingVC int64
	for _, track := range tracksList.Tracks {
		// Filter by track if specified
		if cmd.CheckTrack != trackAll && track.Track != cmd.CheckTrack {
			continue
		}

		for _, release := range track.Releases {
			for _, vc := range release.VersionCodes {
				vcStr := fmt.Sprintf("%d", vc)
				if vc > maxExistingVC {
					maxExistingVC = vc
				}

				if requestedVCs[vcStr] {
					result.Conflicts = append(result.Conflicts, releaseConflict{
						VersionCode:     vcStr,
						Track:           track.Track,
						Status:          release.Status,
						ExistingVersion: release.Name,
					})
				}
			}
		}
	}

	result.HasConflicts = len(result.Conflicts) > 0

	if cmd.SuggestFix && result.HasConflicts {
		// Find the highest conflicting version code
		suggestedVC := maxExistingVC + 1
		result.Suggestions = append(result.Suggestions,
			fmt.Sprintf("Use version code %d or higher to avoid conflicts", suggestedVC))

		// Check for multi-track conflicts
		trackMap := make(map[string][]string) // vc -> list of tracks
		for _, conflict := range result.Conflicts {
			trackMap[conflict.VersionCode] = append(trackMap[conflict.VersionCode], conflict.Track)
		}
		for vc, tracks := range trackMap {
			if len(tracks) > 1 {
				result.Suggestions = append(result.Suggestions,
					fmt.Sprintf("Version code %s exists on multiple tracks (%s); promote or remove from older tracks", vc, strings.Join(tracks, ", ")))
			}
		}
	}

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher"))
}

// ReleaseStrategyCmd provides rollback/roll-forward recommendations.
type ReleaseStrategyCmd struct {
	Track           string  `help:"Track to analyze" default:"production" enum:"internal,alpha,beta,production"`
	CurrentVersion  string  `help:"Current release version code"`
	HealthThreshold float64 `help:"Health score threshold (0-1)" default:"0.95"`
	DryRun          bool    `help:"Show strategy without executing"`
}

// releaseStrategyResult represents the strategy recommendation.
type releaseStrategyResult struct {
	Track          string                 `json:"track"`
	CurrentVersion string                 `json:"currentVersion"`
	HealthScore    float64                `json:"healthScore"`
	Recommendation string                 `json:"recommendation" enum:"continue,rollback,monitor,investigate"`
	Reasoning      string                 `json:"reasoning"`
	Actions        []string               `json:"actions"`
	Risks          []string               `json:"risks,omitempty"`
	Metrics        releaseStrategyMetrics `json:"metrics"`
	AnalyzedAt     time.Time              `json:"analyzedAt"`
}

// releaseStrategyMetrics contains health metrics.
type releaseStrategyMetrics struct {
	CrashRate    float64 `json:"crashRate"`
	AnrRate      float64 `json:"anrRate"`
	ErrorRate    float64 `json:"errorRate"`
	UserFeedback float64 `json:"userFeedback"`
}

// queryStrategyVitals queries crash and ANR rates for strategy analysis.
func (cmd *ReleaseStrategyCmd) queryStrategyVitals(ctx context.Context, client *api.Client, globals *Globals) (crashRate, anrRate float64, err error) {
	reportingSvc, err := client.PlayReporting()
	if err != nil {
		return 0, 0, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get reporting service: %v", err))
	}

	timelineSpec, err := buildTimelineSpec("", "")
	if err != nil {
		return 0, 0, err
	}

	if err := client.Acquire(ctx); err != nil {
		return 0, 0, err
	}

	crashName := fmt.Sprintf("apps/%s/crashRateMetricSet", globals.Package)
	crashReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{"crashRate", "crashRate7dUserWeighted"},
		PageSize:     1,
	}

	qerr := client.DoWithRetry(ctx, func() error {
		resp, rerr := reportingSvc.Vitals.Crashrate.Query(crashName, crashReq).Context(ctx).Do()
		if rerr != nil {
			return rerr
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
	if qerr != nil {
		crashRate = 0
	}

	anrName := fmt.Sprintf("apps/%s/anrRateMetricSet", globals.Package)
	anrReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{"anrRate", "anrRate7dUserWeighted"},
		PageSize:     1,
	}

	qerr = client.DoWithRetry(ctx, func() error {
		resp, rerr := reportingSvc.Vitals.Anrrate.Query(anrName, anrReq).Context(ctx).Do()
		if rerr != nil {
			return rerr
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
	if qerr != nil {
		anrRate = 0
	}

	client.Release()
	return crashRate, anrRate, nil
}

// queryStrategyTrack queries the track for current version and rollout info.
func (cmd *ReleaseStrategyCmd) queryStrategyTrack(ctx context.Context, client *api.Client, globals *Globals) (currentVersion string, userFraction float64, err error) {
	pubSvc, err := client.AndroidPublisher()
	if err != nil {
		return "", 0, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get publisher service: %v", err))
	}

	if err := client.Acquire(ctx); err != nil {
		return "", 0, err
	}

	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		edit, err = pubSvc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return "", 0, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = pubSvc.Edits.Tracks.Get(globals.Package, editID, cmd.Track).Context(ctx).Do()
		return err
	})

	_ = client.DoWithRetry(ctx, func() error {
		return pubSvc.Edits.Delete(globals.Package, editID).Context(ctx).Do()
	})

	client.Release()

	if err != nil {
		return "", 0, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track: %v", err))
	}

	if track != nil && len(track.Releases) > 0 {
		latest := track.Releases[0]
		if len(latest.VersionCodes) > 0 {
			currentVersion = fmt.Sprintf("%d", latest.VersionCodes[0])
		}
		if latest.Name != "" {
			currentVersion = latest.Name
		}
		userFraction = latest.UserFraction
	}

	if cmd.CurrentVersion != "" {
		currentVersion = cmd.CurrentVersion
	}

	return currentVersion, userFraction, nil
}

// buildStrategyRecommendation generates the recommendation based on health score.
func (cmd *ReleaseStrategyCmd) buildStrategyRecommendation(healthScore, userFraction float64) (recommendation, reasoning string, actions, risks []string) {
	actions = make([]string, 0)
	risks = make([]string, 0)

	switch {
	case healthScore >= cmd.HealthThreshold:
		recommendation = recommendContinue
		reasoning = fmt.Sprintf("Health score %.2f is above threshold %.2f. Metrics are within acceptable thresholds.", healthScore, cmd.HealthThreshold)
		actions = append(actions, "Continue monitoring crash rate and ANR rate")
		if userFraction > 0 && userFraction < 1.0 {
			actions = append(actions, fmt.Sprintf("Consider increasing rollout from %.0f%% to next stage", userFraction*100))
		}
	case healthScore >= cmd.HealthThreshold*0.85:
		recommendation = "monitor"
		reasoning = fmt.Sprintf("Health score %.2f is slightly below threshold %.2f. Close monitoring recommended.", healthScore, cmd.HealthThreshold)
		actions = append(actions,
			"Monitor crash rate and ANR rate closely for the next 24 hours",
			"Do not increase rollout percentage")
		risks = append(risks, "Metrics may continue to degrade if underlying issues are not addressed")
	case healthScore >= cmd.HealthThreshold*0.7:
		recommendation = "investigate"
		reasoning = fmt.Sprintf("Health score %.2f is significantly below threshold %.2f. Investigation needed.", healthScore, cmd.HealthThreshold)
		actions = append(actions,
			"Investigate crash and ANR reports for the current release",
			"Consider halting rollout while investigating")
		if userFraction > 0 && userFraction < 1.0 {
			actions = append(actions, fmt.Sprintf("Halt rollout at current %.0f%%", userFraction*100))
		}
		risks = append(risks,
			"User experience is degraded for affected users",
			"Bad reviews may increase if rollout continues")
	default:
		recommendation = "rollback"
		reasoning = fmt.Sprintf("Health score %.2f is critically below threshold %.2f. Rollback recommended.", healthScore, cmd.HealthThreshold)
		actions = append(actions,
			"Immediately halt the current rollout",
			"Prepare a hotfix or rollback to the previous stable version",
			"Investigate root cause of crash/ANR spike")
		risks = append(risks,
			"Continued rollout will impact a growing number of users",
			"App rating may be significantly impacted")
	}

	return recommendation, reasoning, actions, risks
}

// Run executes the release strategy command.
func (cmd *ReleaseStrategyCmd) Run(globals *Globals) error {
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
		return err
	}

	startTime := time.Now()

	crashRate, anrRate, err := cmd.queryStrategyVitals(ctx, client, globals)
	if err != nil {
		return err
	}

	currentVersion, userFraction, err := cmd.queryStrategyTrack(ctx, client, globals)
	if err != nil {
		return err
	}

	// Calculate health score based on metrics vs thresholds
	// Health score: 1.0 = perfect, 0.0 = critical
	// Crash rate bad threshold: 0.01 (1%), ANR rate bad threshold: 0.005 (0.5%)
	crashPenalty := crashRate / 0.01
	if crashPenalty > 1 {
		crashPenalty = 1
	}
	anrPenalty := anrRate / 0.005
	if anrPenalty > 1 {
		anrPenalty = 1
	}
	healthScore := 1.0 - (crashPenalty*0.5 + anrPenalty*0.5)
	if healthScore < 0 {
		healthScore = 0
	}

	recommendation, reasoning, actions, risks := cmd.buildStrategyRecommendation(healthScore, userFraction)

	result := &releaseStrategyResult{
		Track:          cmd.Track,
		CurrentVersion: currentVersion,
		HealthScore:    healthScore,
		Recommendation: recommendation,
		Reasoning:      reasoning,
		Actions:        actions,
		Risks:          risks,
		Metrics: releaseStrategyMetrics{
			CrashRate:    crashRate,
			AnrRate:      anrRate,
			ErrorRate:    0, // Error rate requires separate query
			UserFeedback: 0, // Would need reviews API data
		},
		AnalyzedAt: time.Now(),
	}

	r := output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher", "playdeveloperreporting")

	if cmd.DryRun {
		r = r.WithWarnings("dry run - strategy analysis preview, no actions taken")
	}

	return writeOutput(globals, r)
}

// ReleaseHistoryCmd shows detailed release history.
type ReleaseHistoryCmd struct {
	Track         string `help:"Track to show history for" default:"production" enum:"internal,alpha,beta,production"`
	Limit         int    `help:"Maximum releases to show" default:"20"`
	IncludeVitals bool   `help:"Include health metrics for each release"`
	Format        string `help:"Output format" default:"table" enum:"json,table,csv"`
}

// releaseHistoryResult represents the release history.
type releaseHistoryResult struct {
	Track       string               `json:"track"`
	Count       int                  `json:"count"`
	Releases    []releaseHistoryItem `json:"releases"`
	GeneratedAt time.Time            `json:"generatedAt"`
}

// releaseHistoryItem represents a single release in history.
type releaseHistoryItem struct {
	VersionCodes      []string              `json:"versionCodes"`
	Name              string                `json:"name,omitempty"`
	Status            string                `json:"status"`
	ReleaseDate       string                `json:"releaseDate"`
	RolloutPercentage float64               `json:"rolloutPercentage,omitempty"`
	Vitals            *releaseHistoryVitals `json:"vitals,omitempty"`
}

// releaseHistoryVitals contains health metrics for a release.
type releaseHistoryVitals struct {
	CrashRate float64 `json:"crashRate"`
	AnrRate   float64 `json:"anrRate"`
	Stability float64 `json:"stability"`
}

// attachHistoryVitals queries vitals and attaches them to the most recent release.
func (cmd *ReleaseHistoryCmd) attachHistoryVitals(ctx context.Context, client *api.Client, globals *Globals, result *releaseHistoryResult) {
	reportingSvc, verr := client.PlayReporting()
	if verr != nil {
		return
	}

	timelineSpec, terr := buildTimelineSpec("", "")
	if terr != nil {
		return
	}

	if err := client.Acquire(ctx); err != nil {
		return
	}

	var crashRate, anrRate float64

	crashName := fmt.Sprintf("apps/%s/crashRateMetricSet", globals.Package)
	crashReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{"crashRate"},
		PageSize:     1,
	}
	_ = client.DoWithRetry(ctx, func() error {
		resp, qerr := reportingSvc.Vitals.Crashrate.Query(crashName, crashReq).Context(ctx).Do()
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

	anrName := fmt.Sprintf("apps/%s/anrRateMetricSet", globals.Package)
	anrReq := &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: timelineSpec,
		Metrics:      []string{"anrRate"},
		PageSize:     1,
	}
	_ = client.DoWithRetry(ctx, func() error {
		resp, qerr := reportingSvc.Vitals.Anrrate.Query(anrName, anrReq).Context(ctx).Do()
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

	client.Release()

	stability := 1.0 - crashRate - anrRate
	if stability < 0 {
		stability = 0
	}

	if len(result.Releases) > 0 {
		result.Releases[0].Vitals = &releaseHistoryVitals{
			CrashRate: crashRate,
			AnrRate:   anrRate,
			Stability: stability,
		}
	}
}

// Run executes the release history command.
func (cmd *ReleaseHistoryCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
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

	result := &releaseHistoryResult{
		Track:       cmd.Track,
		Count:       0,
		Releases:    make([]releaseHistoryItem, 0, cmd.Limit),
		GeneratedAt: time.Now(),
	}

	// Create temporary edit
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	// Get track releases
	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(globals.Package, editID, cmd.Track).Context(ctx).Do()
		return err
	})

	// Clean up the temporary edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(globals.Package, editID).Context(ctx).Do()
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track: %v", err))
	}

	// Build history from real release data
	if track != nil && track.Releases != nil {
		for i, release := range track.Releases {
			if cmd.Limit > 0 && i >= cmd.Limit {
				break
			}

			versionCodes := make([]string, 0, len(release.VersionCodes))
			for _, vc := range release.VersionCodes {
				versionCodes = append(versionCodes, fmt.Sprintf("%d", vc))
			}

			var rolloutPct float64
			switch release.Status {
			case statusInProgress:
				rolloutPct = release.UserFraction * 100
			case releaseCompleted:
				rolloutPct = 100.0
			}

			item := releaseHistoryItem{
				VersionCodes:      versionCodes,
				Name:              release.Name,
				Status:            release.Status,
				ReleaseDate:       time.Now().Format("2006-01-02"),
				RolloutPercentage: rolloutPct,
			}

			result.Releases = append(result.Releases, item)
			result.Count++
		}
	}

	// If IncludeVitals, query vitals for the package
	if cmd.IncludeVitals && result.Count > 0 {
		cmd.attachHistoryVitals(ctx, client, globals, result)
	}

	// Sort by date descending (most recent first)
	sort.Slice(result.Releases, func(i, j int) bool {
		return result.Releases[i].ReleaseDate > result.Releases[j].ReleaseDate
	})

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher"))
}

// ReleaseNotesCmd manages release notes across locales.
type ReleaseNotesCmd struct {
	Action        string   `help:"Action to perform" enum:"get,set,copy,list" required:""`
	Track         string   `help:"Track for the release" default:"production" enum:"internal,alpha,beta,production"`
	VersionCode   string   `help:"Version code for the release"`
	SourceLocale  string   `help:"Source locale (for copy action)" default:"en-US"`
	TargetLocales []string `help:"Target locales (for copy action, repeatable)"`
	File          string   `help:"JSON file with release notes (for set action)" type:"existingfile"`
	Format        string   `help:"Output format" default:"json" enum:"json,table"`
}

// releaseNotesResult represents the release notes operation result.
type releaseNotesResult struct {
	Action      string                      `json:"action"`
	Track       string                      `json:"track"`
	VersionCode string                      `json:"versionCode,omitempty"`
	Locales     map[string]releaseNotesData `json:"locales,omitempty"`
	Source      string                      `json:"source,omitempty"`
	Targets     []string                    `json:"targets,omitempty"`
	ModifiedAt  time.Time                   `json:"modifiedAt,omitempty"`
}

// releaseNotesData represents release notes for a locale.
type releaseNotesData struct {
	Text string `json:"text"`
}

// releaseNotesActionGet handles the "get" and "list" actions for release notes.
func (cmd *ReleaseNotesCmd) releaseNotesActionGet(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg string, result *releaseNotesResult) error {
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	var err error
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})

	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track: %v", err))
	}

	result.Locales = make(map[string]releaseNotesData)

	if track != nil && track.Releases != nil {
		for _, release := range track.Releases {
			if cmd.Action == actionGet && cmd.VersionCode != "" {
				if !releaseMatchesVersionCode(release, cmd.VersionCode) {
					continue
				}
			}

			if release.ReleaseNotes != nil {
				for _, note := range release.ReleaseNotes {
					result.Locales[note.Language] = releaseNotesData{Text: note.Text}
				}
			}

			if cmd.Action == actionGet {
				break
			}
		}
	}

	return nil
}

// releaseNotesActionSet handles the "set" action for release notes.
func (cmd *ReleaseNotesCmd) releaseNotesActionSet(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg string, result *releaseNotesResult) error {
	if cmd.File == "" {
		return errors.NewAPIError(errors.CodeValidationError, "--file is required for set action")
	}

	fileData, ferr := os.ReadFile(cmd.File)
	if ferr != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", ferr))
	}

	var notesMap map[string]string
	if jerr := json.Unmarshal(fileData, &notesMap); jerr != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse release notes JSON: %v", jerr))
	}

	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	var err error
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})
	if err != nil {
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track: %v", err))
	}

	if track != nil && track.Releases != nil {
		for i, release := range track.Releases {
			if cmd.VersionCode != "" && !releaseMatchesVersionCode(release, cmd.VersionCode) {
				continue
			}

			localizedNotes := make([]*androidpublisher.LocalizedText, 0, len(notesMap))
			for locale, text := range notesMap {
				localizedNotes = append(localizedNotes, &androidpublisher.LocalizedText{
					Language: locale,
					Text:     text,
				})
			}

			track.Releases[i].ReleaseNotes = localizedNotes
			break
		}
	}

	err = client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Tracks.Update(pkg, editID, cmd.Track, track).Context(ctx).Do()
		return uerr
	})
	if err != nil {
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update track: %v", err))
	}

	err = client.DoWithRetry(ctx, func() error {
		_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
		return cerr
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err))
	}

	result.Locales = make(map[string]releaseNotesData)
	for locale, text := range notesMap {
		result.Locales[locale] = releaseNotesData{Text: text}
	}

	return nil
}

// releaseNotesActionCopy handles the "copy" action for release notes.
func (cmd *ReleaseNotesCmd) releaseNotesActionCopy(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg string, result *releaseNotesResult) error {
	if len(cmd.TargetLocales) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "--target-locales is required for copy action")
	}

	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	var err error
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})
	if err != nil {
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track: %v", err))
	}

	result.Source = cmd.SourceLocale
	result.Targets = cmd.TargetLocales
	result.Locales = make(map[string]releaseNotesData)

	if track != nil && len(track.Releases) > 0 {
		release := track.Releases[0]

		var sourceText string
		for _, note := range release.ReleaseNotes {
			if note.Language == cmd.SourceLocale {
				sourceText = note.Text
				result.Locales[cmd.SourceLocale] = releaseNotesData{Text: sourceText}
				break
			}
		}

		if sourceText == "" {
			_ = client.DoWithRetry(ctx, func() error {
				return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
			})
			client.Release()
			return errors.NewAPIError(errors.CodeNotFound,
				fmt.Sprintf("no release notes found for source locale %s", cmd.SourceLocale))
		}

		existingNotes := make(map[string]*androidpublisher.LocalizedText)
		for _, note := range release.ReleaseNotes {
			existingNotes[note.Language] = note
		}

		for _, targetLocale := range cmd.TargetLocales {
			if existing, ok := existingNotes[targetLocale]; ok {
				existing.Text = sourceText
			} else {
				release.ReleaseNotes = append(release.ReleaseNotes, &androidpublisher.LocalizedText{
					Language: targetLocale,
					Text:     sourceText,
				})
			}
			result.Locales[targetLocale] = releaseNotesData{Text: sourceText}
		}

		track.Releases[0] = release
	}

	err = client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Tracks.Update(pkg, editID, cmd.Track, track).Context(ctx).Do()
		return uerr
	})
	if err != nil {
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update track: %v", err))
	}

	err = client.DoWithRetry(ctx, func() error {
		_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
		return cerr
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err))
	}

	return nil
}

// releaseMatchesVersionCode checks if a release contains the given version code.
func releaseMatchesVersionCode(release *androidpublisher.TrackRelease, versionCode string) bool {
	for _, vc := range release.VersionCodes {
		if fmt.Sprintf("%d", vc) == versionCode {
			return true
		}
	}
	return false
}

// Run executes the release notes command.
func (cmd *ReleaseNotesCmd) Run(globals *Globals) error {
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
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get publisher service: %v", err))
	}

	startTime := time.Now()
	pkg := globals.Package

	result := &releaseNotesResult{
		Action:      cmd.Action,
		Track:       cmd.Track,
		VersionCode: cmd.VersionCode,
		ModifiedAt:  time.Now(),
	}

	switch cmd.Action {
	case actionGet, "list":
		if err := cmd.releaseNotesActionGet(ctx, client, svc, pkg, result); err != nil {
			return err
		}
	case "set":
		if err := cmd.releaseNotesActionSet(ctx, client, svc, pkg, result); err != nil {
			return err
		}
	case actionCopy:
		if err := cmd.releaseNotesActionCopy(ctx, client, svc, pkg, result); err != nil {
			return err
		}
	}

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher"))
}
