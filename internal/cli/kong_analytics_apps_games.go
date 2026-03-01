package cli

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	gamesmanagement "google.golang.org/api/gamesmanagement/v1management"
	playdeveloperreporting "google.golang.org/api/playdeveloperreporting/v1beta1"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

// ============================================================================
// Analytics Commands
// ============================================================================

// AnalyticsCmd contains analytics commands.
type AnalyticsCmd struct {
	Query        AnalyticsQueryCmd        `cmd:"" help:"Query analytics data"`
	Capabilities AnalyticsCapabilitiesCmd `cmd:"" help:"List analytics capabilities"`
}

// AnalyticsQueryCmd runs an analytics query.
type AnalyticsQueryCmd struct {
	StartDate  string   `help:"Start date (ISO 8601)"`
	EndDate    string   `help:"End date (ISO 8601)"`
	Metrics    []string `help:"Metrics to retrieve"`
	Dimensions []string `help:"Dimensions for grouping"`
	Format     string   `help:"Output format: json, csv" default:"json" enum:"json,csv"`
	PageSize   int64    `help:"Results per page" default:"100"`
	PageToken  string   `help:"Pagination token"`
	All        bool     `help:"Fetch all pages"`
}

// Run executes the analytics query command.
func (cmd *AnalyticsQueryCmd) Run(globals *Globals) error {
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

	// Determine which metric set to query based on requested metrics
	// The Play Developer Reporting API uses specific metric set endpoints.
	// We route to the crash rate metric set as a default/general analytics entry point.
	metricSetName := metricSetCrashRate
	if len(cmd.Metrics) > 0 {
		// Map common metric names to metric sets
		switch cmd.Metrics[0] {
		case metricCrashRate, "userPerceivedCrashRate":
			metricSetName = metricSetCrashRate
		case metricAnrRate, "userPerceivedAnrRate":
			metricSetName = metricSetAnrRate
		case metricSlowRenderingRate:
			metricSetName = metricSetSlowRendering
		case metricSlowStartRate:
			metricSetName = metricSetSlowStart
		case metricStuckBackgroundWakelockRate:
			metricSetName = metricSetStuckBackgroundWakelock
		case metricExcessiveWakeupRate:
			metricSetName = metricSetExcessiveWakeup
		case metricErrorCount:
			metricSetName = metricSetErrorCount
		}
	}

	name := fmt.Sprintf("apps/%s/%s", globals.Package, metricSetName)

	timelineSpec, err := buildTimelineSpec(cmd.StartDate, cmd.EndDate)
	if err != nil {
		return err
	}

	startTime := time.Now()

	// Use the crash rate query as the generic query path (all metric sets return MetricsRows)
	// The Play Reporting API requires specific query endpoints per metric set.
	var allRows []map[string]interface{}

	switch metricSetName {
	case "crashRateMetricSet":
		req := &crashRateQueryRequest{
			TimelineSpec: timelineSpec,
			Dimensions:   cmd.Dimensions,
			Metrics:      cmd.Metrics,
			PageSize:     cmd.PageSize,
			PageToken:    cmd.PageToken,
		}
		err = client.DoWithRetry(ctx, func() error {
			resp, qErr := svc.Vitals.Crashrate.Query(name, req.toAPI()).Context(ctx).Do()
			if qErr != nil {
				return qErr
			}
			allRows = append(allRows, formatMetricsRows(resp.Rows)...)
			return nil
		})
	case "anrRateMetricSet":
		req := &anrRateQueryRequest{
			TimelineSpec: timelineSpec,
			Dimensions:   cmd.Dimensions,
			Metrics:      cmd.Metrics,
			PageSize:     cmd.PageSize,
			PageToken:    cmd.PageToken,
		}
		err = client.DoWithRetry(ctx, func() error {
			resp, qErr := svc.Vitals.Anrrate.Query(name, req.toAPI()).Context(ctx).Do()
			if qErr != nil {
				return qErr
			}
			allRows = append(allRows, formatMetricsRows(resp.Rows)...)
			return nil
		})
	default:
		// For other metric sets, return informational guidance
		result := output.NewResult(map[string]interface{}{
			"message":   fmt.Sprintf("Use 'gpd vitals' subcommands for specific metric sets. Metric set %q requires a dedicated query endpoint.", metricSetName),
			"metricSet": metricSetName,
			"hint":      "Try: gpd vitals crashes, gpd vitals anrs, gpd vitals metrics slow-rendering, etc.",
		}).WithDuration(time.Since(startTime)).
			WithServices("playdeveloperreporting")
		return outputResult(result, globals.Output, globals.Pretty)
	}

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to query analytics: %v", err))
	}

	result := output.NewResult(allRows).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	if cmd.Format == formatCSV {
		return outputResult(result, formatCSV, globals.Pretty)
	}
	return outputResult(result, globals.Output, globals.Pretty)
}

// AnalyticsCapabilitiesCmd lists analytics capabilities.
type AnalyticsCapabilitiesCmd struct{}

// Run executes the analytics capabilities command.
func (cmd *AnalyticsCapabilitiesCmd) Run(globals *Globals) error {
	startTime := time.Now()

	capabilities := map[string]interface{}{
		"metricSets": []map[string]interface{}{
			{
				"name":        "crashRateMetricSet",
				"description": "Crash rate metrics including user-perceived crash rates",
				"metrics": []string{
					"crashRate", "crashRate7dUserWeighted", "crashRate28dUserWeighted",
					"userPerceivedCrashRate", "userPerceivedCrashRate7dUserWeighted",
					"userPerceivedCrashRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        "anrRateMetricSet",
				"description": "ANR (Application Not Responding) rate metrics",
				"metrics": []string{
					"anrRate", "anrRate7dUserWeighted", "anrRate28dUserWeighted",
					"userPerceivedAnrRate", "userPerceivedAnrRate7dUserWeighted",
					"userPerceivedAnrRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        "slowRenderingRateMetricSet",
				"description": "Slow rendering rate metrics (UI jank)",
				"metrics": []string{
					"slowRenderingRate", "slowRenderingRate7dUserWeighted",
					"slowRenderingRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        "slowStartRateMetricSet",
				"description": "Slow app start rate metrics",
				"metrics": []string{
					"slowStartRate", "slowStartRate7dUserWeighted",
					"slowStartRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        "stuckBackgroundWakelockRateMetricSet",
				"description": "Stuck background wakelock rate metrics",
				"metrics": []string{
					"stuckBackgroundWakelockRate", "stuckBackgroundWakelockRate7dUserWeighted",
					"stuckBackgroundWakelockRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        "excessiveWakeupRateMetricSet",
				"description": "Excessive wakeup rate metrics",
				"metrics": []string{
					"excessiveWakeupRate", "excessiveWakeupRate7dUserWeighted",
					"excessiveWakeupRate28dUserWeighted", "distinctUsers",
				},
			},
			{
				"name":        "errorCountMetricSet",
				"description": "Error count metrics for crashes and ANRs",
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
		"filters": []string{
			"apiLevel", "versionCode", "deviceModel", "countryCode",
		},
		"notes": "Use 'gpd analytics query --metrics <metric>' to query specific metrics. " +
			"Metrics are grouped by metric sets - each query targets one metric set. " +
			"Use 'gpd vitals' subcommands for dedicated metric set queries.",
	}

	result := output.NewResult(capabilities).
		WithDuration(time.Since(startTime)).
		WithServices("playdeveloperreporting")

	return outputResult(result, globals.Output, globals.Pretty)
}

// ============================================================================
// Apps Commands
// ============================================================================

// AppsCmd contains app discovery commands.
type AppsCmd struct {
	List AppsListCmd `cmd:"" help:"List apps in the developer account"`
	Get  AppsGetCmd  `cmd:"" help:"Get app details"`
}

// AppsListCmd lists apps in the developer account.
type AppsListCmd struct {
	PageSize  int64  `help:"Results per page" default:"100"`
	PageToken string `help:"Pagination token"`
	All       bool   `help:"Fetch all pages"`
}

// Run executes the apps list command.
func (cmd *AppsListCmd) Run(globals *Globals) error {
	startTime := time.Now()

	// The Android Publisher API does not provide a "list apps" endpoint.
	// Service accounts are granted access to specific apps via the Google Play Console.
	result := output.NewResult(map[string]interface{}{
		"message": "The Google Play Developer API does not provide a 'list apps' endpoint. " +
			"Access is granted per-app via the Google Play Console.",
		"hints": []string{
			"Use '--package <package.name>' to query a specific app",
			"Set GPD_PACKAGE environment variable for a default package",
			"Use 'gpd config set package <package.name>' to set a default",
			"Check Google Play Console > Settings > API access for configured apps",
		},
		"workaround": "To verify access to a specific app, use: gpd apps get <package.name>",
	}).WithDuration(time.Since(startTime)).
		WithServices("androidpublisher").
		WithWarnings("The Android Publisher API does not support listing all apps for an account")

	return outputResult(result, globals.Output, globals.Pretty)
}

// AppsGetCmd gets app details.
type AppsGetCmd struct {
	Package string `arg:"" help:"App package name (uses --package when omitted)"`
}

// Run executes the apps get command.
func (cmd *AppsGetCmd) Run(globals *Globals) error {
	pkg := cmd.Package
	if pkg == "" {
		pkg = globals.Package
	}
	if err := requirePackage(pkg); err != nil {
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

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get publisher service: %v", err))
	}

	startTime := time.Now()

	// Create a temporary edit to query app details
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		var e error
		edit, e = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return e
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit (verify package access): %v", err)).
			WithHint("Ensure the service account has access to this app in Google Play Console")
	}

	editID := edit.Id

	// Get app details
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var details *androidpublisher.AppDetails
	err = client.DoWithRetry(ctx, func() error {
		var e error
		details, e = svc.Edits.Details.Get(pkg, editID).Context(ctx).Do()
		return e
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get app details: %v", err))
	}

	// Get track info for latest releases
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var tracksResp *androidpublisher.TracksListResponse
	err = client.DoWithRetry(ctx, func() error {
		var e error
		tracksResp, e = svc.Edits.Tracks.List(pkg, editID).Context(ctx).Do()
		return e
	})

	client.Release()

	trackInfo := make([]map[string]interface{}, 0)
	if err == nil && tracksResp != nil {
		for _, track := range tracksResp.Tracks {
			ti := map[string]interface{}{
				"track": track.Track,
			}
			if len(track.Releases) > 0 {
				latest := track.Releases[0]
				ti["latestRelease"] = map[string]interface{}{
					"name":         latest.Name,
					"status":       latest.Status,
					"versionCodes": latest.VersionCodes,
				}
			}
			trackInfo = append(trackInfo, ti)
		}
	}

	// Delete the temporary edit (best effort)
	if acquireErr := client.Acquire(ctx); acquireErr == nil {
		_ = client.DoWithRetry(ctx, func() error {
			e := svc.Edits.Delete(pkg, editID).Context(ctx).Do()
			return e
		})
		client.Release()
	}

	appData := map[string]interface{}{
		"packageName":    pkg,
		"defaultLocale":  details.DefaultLanguage,
		"contactEmail":   details.ContactEmail,
		"contactPhone":   details.ContactPhone,
		"contactWebsite": details.ContactWebsite,
		"tracks":         trackInfo,
	}

	result := output.NewResult(appData).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// ============================================================================
// Games Commands
// ============================================================================

// GamesCmd contains Google Play Games services commands.
type GamesCmd struct {
	Achievements GamesAchievementsCmd `cmd:"" help:"Manage game achievements"`
	Scores       GamesScoresCmd       `cmd:"" help:"Manage leaderboard scores"`
	Events       GamesEventsCmd       `cmd:"" help:"Manage game events"`
	Players      GamesPlayersCmd      `cmd:"" help:"Manage player visibility"`
	Capabilities GamesCapabilitiesCmd `cmd:"" help:"List Games management capabilities"`
}

// GamesAchievementsCmd manages game achievements.
type GamesAchievementsCmd struct {
	Reset GamesAchievementsResetCmd `cmd:"" help:"Reset achievements"`
}

// GamesAchievementsResetCmd resets achievements.
type GamesAchievementsResetCmd struct {
	AchievementID string   `arg:"" help:"Achievement ID to reset (uses --all-players to reset all)"`
	AllPlayers    bool     `help:"Reset for all players (requires admin)"`
	IDs           []string `help:"Multiple achievement IDs to reset (comma-separated)"`
}

// Run executes the achievements reset command.
func (cmd *GamesAchievementsResetCmd) Run(globals *Globals) error {
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

	gmSvc, err := client.GamesManagement()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get games management service: %v", err))
	}

	startTime := time.Now()

	// If multiple IDs provided and AllPlayers, use batch reset
	if len(cmd.IDs) > 0 && cmd.AllPlayers {
		err = client.DoWithRetry(ctx, func() error {
			return gmSvc.Achievements.ResetMultipleForAllPlayers(
				&gamesmanagement.AchievementResetMultipleForAllRequest{
					AchievementIds: cmd.IDs,
				}).Context(ctx).Do()
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset achievements for all players: %v", err))
		}

		result := output.NewResult(map[string]interface{}{
			"action":         "resetMultipleForAllPlayers",
			"achievementIds": cmd.IDs,
			"count":          len(cmd.IDs),
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset for all players (single achievement or all)
	if cmd.AllPlayers {
		if cmd.AchievementID == "" {
			// Reset all achievements for all players
			err = client.DoWithRetry(ctx, func() error {
				return gmSvc.Achievements.ResetAllForAllPlayers().Context(ctx).Do()
			})
			if err != nil {
				return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset all achievements for all players: %v", err))
			}

			result := output.NewResult(map[string]interface{}{
				"action": "resetAllForAllPlayers",
			}).WithDuration(time.Since(startTime)).
				WithServices("gamesmanagement")

			return outputResult(result, globals.Output, globals.Pretty)
		}

		// Reset single achievement for all players
		err = client.DoWithRetry(ctx, func() error {
			return gmSvc.Achievements.ResetForAllPlayers(cmd.AchievementID).Context(ctx).Do()
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset achievement for all players: %v", err))
		}

		result := output.NewResult(map[string]interface{}{
			"action":        "resetForAllPlayers",
			"achievementId": cmd.AchievementID,
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset for current player (single or multiple IDs)
	if len(cmd.IDs) > 0 {
		results := make([]map[string]interface{}, 0, len(cmd.IDs))
		for _, id := range cmd.IDs {
			var resp *gamesmanagement.AchievementResetResponse
			resetErr := client.DoWithRetry(ctx, func() error {
				var e error
				resp, e = gmSvc.Achievements.Reset(id).Context(ctx).Do()
				return e
			})
			if resetErr != nil {
				results = append(results, map[string]interface{}{
					"achievementId": id,
					"status":        "failed",
					"error":         resetErr.Error(),
				})
			} else {
				results = append(results, map[string]interface{}{
					"achievementId":  resp.DefinitionId,
					"currentState":   resp.CurrentState,
					"updateOccurred": resp.UpdateOccurred,
					"status":         "success",
				})
			}
		}

		result := output.NewResult(map[string]interface{}{
			"action":  "resetMultiple",
			"results": results,
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset single achievement for current player
	var resp *gamesmanagement.AchievementResetResponse
	err = client.DoWithRetry(ctx, func() error {
		var e error
		resp, e = gmSvc.Achievements.Reset(cmd.AchievementID).Context(ctx).Do()
		return e
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset achievement: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"action":         "reset",
		"achievementId":  resp.DefinitionId,
		"currentState":   resp.CurrentState,
		"updateOccurred": resp.UpdateOccurred,
	}).WithDuration(time.Since(startTime)).
		WithServices("gamesmanagement")

	return outputResult(result, globals.Output, globals.Pretty)
}

// GamesScoresCmd manages leaderboard scores.
type GamesScoresCmd struct {
	Reset GamesScoresResetCmd `cmd:"" help:"Reset scores on a leaderboard"`
}

// GamesScoresResetCmd resets leaderboard scores.
type GamesScoresResetCmd struct {
	LeaderboardID string   `arg:"" help:"Leaderboard ID to reset (uses --all-players to reset all)"`
	AllPlayers    bool     `help:"Reset for all players (requires admin)"`
	IDs           []string `help:"Multiple leaderboard IDs to reset (comma-separated)"`
}

// Run executes the scores reset command.
func (cmd *GamesScoresResetCmd) Run(globals *Globals) error {
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

	gmSvc, err := client.GamesManagement()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get games management service: %v", err))
	}

	startTime := time.Now()

	// If multiple IDs provided and AllPlayers, use batch reset
	if len(cmd.IDs) > 0 && cmd.AllPlayers {
		err = client.DoWithRetry(ctx, func() error {
			return gmSvc.Scores.ResetMultipleForAllPlayers(
				&gamesmanagement.ScoresResetMultipleForAllRequest{
					LeaderboardIds: cmd.IDs,
				}).Context(ctx).Do()
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset scores for all players: %v", err))
		}

		result := output.NewResult(map[string]interface{}{
			"action":         "resetMultipleForAllPlayers",
			"leaderboardIds": cmd.IDs,
			"count":          len(cmd.IDs),
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset for all players
	if cmd.AllPlayers {
		if cmd.LeaderboardID == "" {
			// Reset all scores for all players
			err = client.DoWithRetry(ctx, func() error {
				return gmSvc.Scores.ResetAllForAllPlayers().Context(ctx).Do()
			})
			if err != nil {
				return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset all scores for all players: %v", err))
			}

			result := output.NewResult(map[string]interface{}{
				"action": "resetAllForAllPlayers",
			}).WithDuration(time.Since(startTime)).
				WithServices("gamesmanagement")

			return outputResult(result, globals.Output, globals.Pretty)
		}

		// Reset single leaderboard for all players
		err = client.DoWithRetry(ctx, func() error {
			return gmSvc.Scores.ResetForAllPlayers(cmd.LeaderboardID).Context(ctx).Do()
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset scores for all players: %v", err))
		}

		result := output.NewResult(map[string]interface{}{
			"action":        "resetForAllPlayers",
			"leaderboardId": cmd.LeaderboardID,
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset for current player (single or multiple)
	if len(cmd.IDs) > 0 {
		results := make([]map[string]interface{}, 0, len(cmd.IDs))
		for _, id := range cmd.IDs {
			var resp *gamesmanagement.PlayerScoreResetResponse
			resetErr := client.DoWithRetry(ctx, func() error {
				var e error
				resp, e = gmSvc.Scores.Reset(id).Context(ctx).Do()
				return e
			})
			if resetErr != nil {
				results = append(results, map[string]interface{}{
					"leaderboardId": id,
					"status":        "failed",
					"error":         resetErr.Error(),
				})
			} else {
				results = append(results, map[string]interface{}{
					"leaderboardId":       resp.DefinitionId,
					"resetScoreTimeSpans": resp.ResetScoreTimeSpans,
					"status":              "success",
				})
			}
		}

		result := output.NewResult(map[string]interface{}{
			"action":  "resetMultiple",
			"results": results,
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset single leaderboard for current player
	var resp *gamesmanagement.PlayerScoreResetResponse
	err = client.DoWithRetry(ctx, func() error {
		var e error
		resp, e = gmSvc.Scores.Reset(cmd.LeaderboardID).Context(ctx).Do()
		return e
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset scores: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"action":              "reset",
		"leaderboardId":       resp.DefinitionId,
		"resetScoreTimeSpans": resp.ResetScoreTimeSpans,
	}).WithDuration(time.Since(startTime)).
		WithServices("gamesmanagement")

	return outputResult(result, globals.Output, globals.Pretty)
}

// GamesEventsCmd manages game events.
type GamesEventsCmd struct {
	Reset GamesEventsResetCmd `cmd:"" help:"Reset game events"`
}

// GamesEventsResetCmd resets game events.
type GamesEventsResetCmd struct {
	EventID    string   `arg:"" help:"Event ID to reset (uses --all-players to reset all)"`
	AllPlayers bool     `help:"Reset for all players (requires admin)"`
	IDs        []string `help:"Multiple event IDs to reset (comma-separated)"`
}

// Run executes the events reset command.
func (cmd *GamesEventsResetCmd) Run(globals *Globals) error {
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

	gmSvc, err := client.GamesManagement()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get games management service: %v", err))
	}

	startTime := time.Now()

	// If multiple IDs provided and AllPlayers, use batch reset
	if len(cmd.IDs) > 0 && cmd.AllPlayers {
		err = client.DoWithRetry(ctx, func() error {
			return gmSvc.Events.ResetMultipleForAllPlayers(
				&gamesmanagement.EventsResetMultipleForAllRequest{
					EventIds: cmd.IDs,
				}).Context(ctx).Do()
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset events for all players: %v", err))
		}

		result := output.NewResult(map[string]interface{}{
			"action":   "resetMultipleForAllPlayers",
			"eventIds": cmd.IDs,
			"count":    len(cmd.IDs),
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset for all players
	if cmd.AllPlayers {
		if cmd.EventID == "" {
			// Reset all events for all players
			err = client.DoWithRetry(ctx, func() error {
				return gmSvc.Events.ResetAllForAllPlayers().Context(ctx).Do()
			})
			if err != nil {
				return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset all events for all players: %v", err))
			}

			result := output.NewResult(map[string]interface{}{
				"action": "resetAllForAllPlayers",
			}).WithDuration(time.Since(startTime)).
				WithServices("gamesmanagement")

			return outputResult(result, globals.Output, globals.Pretty)
		}

		// Reset single event for all players
		err = client.DoWithRetry(ctx, func() error {
			return gmSvc.Events.ResetForAllPlayers(cmd.EventID).Context(ctx).Do()
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset event for all players: %v", err))
		}

		result := output.NewResult(map[string]interface{}{
			"action":  "resetForAllPlayers",
			"eventId": cmd.EventID,
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset for current player (single or multiple)
	if len(cmd.IDs) > 0 {
		results := make([]map[string]interface{}, 0, len(cmd.IDs))
		for _, id := range cmd.IDs {
			resetErr := client.DoWithRetry(ctx, func() error {
				return gmSvc.Events.Reset(id).Context(ctx).Do()
			})
			if resetErr != nil {
				results = append(results, map[string]interface{}{
					"eventId": id,
					"status":  "failed",
					"error":   resetErr.Error(),
				})
			} else {
				results = append(results, map[string]interface{}{
					"eventId": id,
					"status":  "success",
				})
			}
		}

		result := output.NewResult(map[string]interface{}{
			"action":  "resetMultiple",
			"results": results,
		}).WithDuration(time.Since(startTime)).
			WithServices("gamesmanagement")

		return outputResult(result, globals.Output, globals.Pretty)
	}

	// Reset single event for current player
	err = client.DoWithRetry(ctx, func() error {
		return gmSvc.Events.Reset(cmd.EventID).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to reset event: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"action":  "reset",
		"eventId": cmd.EventID,
	}).WithDuration(time.Since(startTime)).
		WithServices("gamesmanagement")

	return outputResult(result, globals.Output, globals.Pretty)
}

// GamesPlayersCmd manages player visibility.
type GamesPlayersCmd struct {
	Hide   GamesPlayersHideCmd   `cmd:"" help:"Hide a player"`
	Unhide GamesPlayersUnhideCmd `cmd:"" help:"Unhide a player"`
}

// GamesPlayersHideCmd hides a player.
type GamesPlayersHideCmd struct {
	PlayerID      string `arg:"" help:"Player ID to hide"`
	ApplicationID string `help:"Game application ID (required)" required:""`
}

// Run executes the players hide command.
func (cmd *GamesPlayersHideCmd) Run(globals *Globals) error {
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

	gmSvc, err := client.GamesManagement()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get games management service: %v", err))
	}

	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		return gmSvc.Players.Hide(cmd.ApplicationID, cmd.PlayerID).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to hide player: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"action":        "hide",
		"playerId":      cmd.PlayerID,
		"applicationId": cmd.ApplicationID,
	}).WithDuration(time.Since(startTime)).
		WithServices("gamesmanagement")

	return outputResult(result, globals.Output, globals.Pretty)
}

// GamesPlayersUnhideCmd unhides a player.
type GamesPlayersUnhideCmd struct {
	PlayerID      string `arg:"" help:"Player ID to unhide"`
	ApplicationID string `help:"Game application ID (required)" required:""`
}

// Run executes the players unhide command.
func (cmd *GamesPlayersUnhideCmd) Run(globals *Globals) error {
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

	gmSvc, err := client.GamesManagement()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get games management service: %v", err))
	}

	startTime := time.Now()

	err = client.DoWithRetry(ctx, func() error {
		return gmSvc.Players.Unhide(cmd.ApplicationID, cmd.PlayerID).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to unhide player: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"action":        "unhide",
		"playerId":      cmd.PlayerID,
		"applicationId": cmd.ApplicationID,
	}).WithDuration(time.Since(startTime)).
		WithServices("gamesmanagement")

	return outputResult(result, globals.Output, globals.Pretty)
}

// GamesCapabilitiesCmd lists Games management capabilities.
type GamesCapabilitiesCmd struct{}

// Run executes the games capabilities command.
func (cmd *GamesCapabilitiesCmd) Run(globals *Globals) error {
	startTime := time.Now()

	capabilities := map[string]interface{}{
		"service": "Google Play Games Management API",
		"version": "v1management",
		"resources": map[string]interface{}{
			"achievements": map[string]interface{}{
				"description": "Reset achievement progress for testing",
				"operations": []map[string]interface{}{
					{"method": "reset", "description": "Reset a single achievement for the current player"},
					{"method": "resetAll", "description": "Reset all achievements for the current player"},
					{"method": "resetForAllPlayers", "description": "Reset a single achievement for all players (admin)"},
					{"method": "resetAllForAllPlayers", "description": "Reset all achievements for all players (admin)"},
					{"method": "resetMultipleForAllPlayers", "description": "Reset multiple achievements for all players (admin)"},
				},
			},
			"scores": map[string]interface{}{
				"description": "Reset leaderboard scores for testing",
				"operations": []map[string]interface{}{
					{"method": "reset", "description": "Reset scores for a single leaderboard for the current player"},
					{"method": "resetAll", "description": "Reset all scores for the current player"},
					{"method": "resetForAllPlayers", "description": "Reset scores for a single leaderboard for all players (admin)"},
					{"method": "resetAllForAllPlayers", "description": "Reset all scores for all players (admin)"},
					{"method": "resetMultipleForAllPlayers", "description": "Reset multiple leaderboard scores for all players (admin)"},
				},
			},
			"events": map[string]interface{}{
				"description": "Reset game events for testing",
				"operations": []map[string]interface{}{
					{"method": "reset", "description": "Reset a single event for the current player"},
					{"method": "resetAll", "description": "Reset all events for the current player"},
					{"method": "resetForAllPlayers", "description": "Reset a single event for all players (admin)"},
					{"method": "resetAllForAllPlayers", "description": "Reset all events for all players (admin)"},
					{"method": "resetMultipleForAllPlayers", "description": "Reset multiple events for all players (admin)"},
				},
			},
			"players": map[string]interface{}{
				"description": "Manage player visibility in the game",
				"operations": []map[string]interface{}{
					{"method": "hide", "description": "Hide a player from the game (ban)"},
					{"method": "unhide", "description": "Unhide a player (unban)"},
				},
			},
		},
		"notes": []string{
			"The Games Management API is primarily used for testing and development",
			"Reset operations for 'all players' require admin/publisher access",
			"Player hide/unhide operations require the game's application ID",
			"This API is separate from the Play Games Services API",
		},
	}

	result := output.NewResult(capabilities).
		WithDuration(time.Since(startTime)).
		WithServices("gamesmanagement")

	return outputResult(result, globals.Output, globals.Pretty)
}

// analyticsQueryHelper types used by AnalyticsQueryCmd
type crashRateQueryRequest struct {
	TimelineSpec interface{}
	Dimensions   []string
	Metrics      []string
	PageSize     int64
	PageToken    string
}

func (r *crashRateQueryRequest) toAPI() *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest {
	ts, _ := r.TimelineSpec.(*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec)
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryCrashRateMetricSetRequest{
		TimelineSpec: ts,
		Dimensions:   r.Dimensions,
		Metrics:      r.Metrics,
		PageSize:     r.PageSize,
		PageToken:    r.PageToken,
	}
}

type anrRateQueryRequest struct {
	TimelineSpec interface{}
	Dimensions   []string
	Metrics      []string
	PageSize     int64
	PageToken    string
}

func (r *anrRateQueryRequest) toAPI() *playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest {
	ts, _ := r.TimelineSpec.(*playdeveloperreporting.GooglePlayDeveloperReportingV1beta1TimelineSpec)
	return &playdeveloperreporting.GooglePlayDeveloperReportingV1beta1QueryAnrRateMetricSetRequest{
		TimelineSpec: ts,
		Dimensions:   r.Dimensions,
		Metrics:      r.Metrics,
		PageSize:     r.PageSize,
		PageToken:    r.PageToken,
	}
}
