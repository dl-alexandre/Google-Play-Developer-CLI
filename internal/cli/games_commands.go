package cli

import (
	"context"

	"github.com/spf13/cobra"
	gamesmanagement "google.golang.org/api/gamesmanagement/v1management"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addGamesCommands() {
	gamesCmd := &cobra.Command{
		Use:   "games",
		Short: "Play Games Services management commands",
		Long:  "Manage Play Games Services including achievements, scores, events, and players.",
	}

	c.addAchievementsCommands(gamesCmd)
	c.addScoresCommands(gamesCmd)
	c.addEventsCommands(gamesCmd)
	c.addPlayersCommands(gamesCmd)
	c.addApplicationsCommands(gamesCmd)
	c.addGamesCapabilitiesCommand(gamesCmd)

	c.rootCmd.AddCommand(gamesCmd)
}

func (c *CLI) addAchievementsCommands(parent *cobra.Command) {
	achievementsCmd := &cobra.Command{
		Use:   "achievements",
		Short: "Manage game achievements",
		Long:  "Reset achievements for testing purposes. Requires Games scope.",
	}

	var (
		allPlayers     bool
		achievementIDs []string
	)

	resetCmd := &cobra.Command{
		Use:   "reset [achievementId]",
		Short: "Reset an achievement",
		Long:  "Reset achievement progress for the currently authenticated player or all players.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !allPlayers && len(achievementIDs) == 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"achievement ID required or use --all-players with no args to reset all"))
			}
			if len(achievementIDs) > 0 && allPlayers {
				return c.gamesAchievementsResetMultipleForAllPlayers(cmd.Context(), achievementIDs)
			}
			if len(args) > 0 && allPlayers {
				return c.gamesAchievementsResetForAllPlayers(cmd.Context(), args[0])
			}
			if len(args) > 0 {
				return c.gamesAchievementsReset(cmd.Context(), args[0])
			}
			return c.gamesAchievementsResetAll(cmd.Context(), allPlayers)
		},
	}
	resetCmd.Flags().BoolVar(&allPlayers, "all-players", false, "Reset for all players (requires admin)")
	resetCmd.Flags().StringSliceVar(&achievementIDs, "ids", nil, "Multiple achievement IDs to reset (comma-separated)")

	achievementsCmd.AddCommand(resetCmd)
	parent.AddCommand(achievementsCmd)
}

func (c *CLI) addScoresCommands(parent *cobra.Command) {
	scoresCmd := &cobra.Command{
		Use:   "scores",
		Short: "Manage leaderboard scores",
		Long:  "Reset leaderboard scores for testing purposes. Requires Games scope.",
	}

	var (
		allPlayers     bool
		leaderboardIDs []string
	)

	resetCmd := &cobra.Command{
		Use:   "reset [leaderboardId]",
		Short: "Reset scores on a leaderboard",
		Long:  "Reset scores for the currently authenticated player or all players.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !allPlayers && len(leaderboardIDs) == 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"leaderboard ID required or use --all-players with no args to reset all"))
			}
			if len(leaderboardIDs) > 0 && allPlayers {
				return c.gamesScoresResetMultipleForAllPlayers(cmd.Context(), leaderboardIDs)
			}
			if len(args) > 0 && allPlayers {
				return c.gamesScoresResetForAllPlayers(cmd.Context(), args[0])
			}
			if len(args) > 0 {
				return c.gamesScoresReset(cmd.Context(), args[0])
			}
			return c.gamesScoresResetAll(cmd.Context(), allPlayers)
		},
	}
	resetCmd.Flags().BoolVar(&allPlayers, "all-players", false, "Reset for all players (requires admin)")
	resetCmd.Flags().StringSliceVar(&leaderboardIDs, "ids", nil, "Multiple leaderboard IDs to reset (comma-separated)")

	scoresCmd.AddCommand(resetCmd)
	parent.AddCommand(scoresCmd)
}

func (c *CLI) addEventsCommands(parent *cobra.Command) {
	eventsCmd := &cobra.Command{
		Use:   "events",
		Short: "Manage game events",
		Long:  "Reset game events for testing purposes. Requires Games scope.",
	}

	var (
		allPlayers bool
		eventIDs   []string
	)

	resetCmd := &cobra.Command{
		Use:   "reset [eventId]",
		Short: "Reset an event",
		Long:  "Reset event progress for the currently authenticated player or all players.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && !allPlayers && len(eventIDs) == 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"event ID required or use --all-players with no args to reset all"))
			}
			if len(eventIDs) > 0 && allPlayers {
				return c.gamesEventsResetMultipleForAllPlayers(cmd.Context(), eventIDs)
			}
			if len(args) > 0 && allPlayers {
				return c.gamesEventsResetForAllPlayers(cmd.Context(), args[0])
			}
			if len(args) > 0 {
				return c.gamesEventsReset(cmd.Context(), args[0])
			}
			return c.gamesEventsResetAll(cmd.Context(), allPlayers)
		},
	}
	resetCmd.Flags().BoolVar(&allPlayers, "all-players", false, "Reset for all players (requires admin)")
	resetCmd.Flags().StringSliceVar(&eventIDs, "ids", nil, "Multiple event IDs to reset (comma-separated)")

	eventsCmd.AddCommand(resetCmd)
	parent.AddCommand(eventsCmd)
}

func (c *CLI) addPlayersCommands(parent *cobra.Command) {
	playersCmd := &cobra.Command{
		Use:   "players",
		Short: "Manage player visibility",
		Long:  "Hide or unhide players from leaderboards and social features.",
	}

	var applicationID string

	hideCmd := &cobra.Command{
		Use:   "hide [playerId]",
		Short: "Hide a player",
		Long:  "Hide the given player's leaderboard scores from other players.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if applicationID == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--application-id is required"))
			}
			return c.gamesPlayersHide(cmd.Context(), applicationID, args[0])
		},
	}
	hideCmd.Flags().StringVar(&applicationID, "application-id", "", "Game application ID (required)")
	_ = hideCmd.MarkFlagRequired("application-id")

	unhideCmd := &cobra.Command{
		Use:   "unhide [playerId]",
		Short: "Unhide a player",
		Long:  "Unhide the given player's leaderboard scores from other players.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if applicationID == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--application-id is required"))
			}
			return c.gamesPlayersUnhide(cmd.Context(), applicationID, args[0])
		},
	}
	unhideCmd.Flags().StringVar(&applicationID, "application-id", "", "Game application ID (required)")
	_ = unhideCmd.MarkFlagRequired("application-id")

	playersCmd.AddCommand(hideCmd, unhideCmd)
	parent.AddCommand(playersCmd)
}

func (c *CLI) addApplicationsCommands(parent *cobra.Command) {
	applicationsCmd := &cobra.Command{
		Use:   "applications",
		Short: "Manage game applications",
		Long:  "List hidden players for a game application.",
	}

	var (
		pageSize  int64
		pageToken string
		all       bool
	)

	listHiddenCmd := &cobra.Command{
		Use:   "list-hidden [applicationId]",
		Short: "List hidden players",
		Long:  "Get the list of players hidden from the given game application.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.gamesApplicationsListHidden(cmd.Context(), args[0], pageSize, pageToken, all)
		},
	}
	listHiddenCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	listHiddenCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(listHiddenCmd, &all)

	applicationsCmd.AddCommand(listHiddenCmd)
	parent.AddCommand(applicationsCmd)
}

func (c *CLI) addGamesCapabilitiesCommand(parent *cobra.Command) {
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List Games management capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.gamesCapabilities(cmd.Context())
		},
	}
	parent.AddCommand(capabilitiesCmd)
}

func (c *CLI) getGamesManagementService(ctx context.Context) (*gamesmanagement.Service, error) {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.GamesManagement()
}

func (c *CLI) gamesAchievementsReset(ctx context.Context, achievementID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	resp, err := svc.Achievements.Reset(achievementID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"achievementId":  achievementID,
		"kind":           resp.Kind,
		"currentState":   resp.CurrentState,
		"definitionId":   resp.DefinitionId,
		"updateOccurred": resp.UpdateOccurred,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesAchievementsResetAll(ctx context.Context, allPlayers bool) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if allPlayers {
		err := svc.Achievements.ResetAllForAllPlayers().Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		result := output.NewResult(map[string]interface{}{
			"success":    true,
			"operation":  "resetAllForAllPlayers",
			"allPlayers": true,
		})
		return c.Output(result.WithServices("gamesmanagement"))
	}

	resp, err := svc.Achievements.ResetAll().Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var achievements []map[string]interface{}
	for _, a := range resp.Results {
		achievements = append(achievements, map[string]interface{}{
			"kind":           a.Kind,
			"currentState":   a.CurrentState,
			"definitionId":   a.DefinitionId,
			"updateOccurred": a.UpdateOccurred,
		})
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"kind":         resp.Kind,
		"achievements": achievements,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesAchievementsResetForAllPlayers(ctx context.Context, achievementID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	err = svc.Achievements.ResetForAllPlayers(achievementID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":       true,
		"achievementId": achievementID,
		"allPlayers":    true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesAchievementsResetMultipleForAllPlayers(ctx context.Context, achievementIDs []string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	req := &gamesmanagement.AchievementResetMultipleForAllRequest{
		AchievementIds: achievementIDs,
	}

	err = svc.Achievements.ResetMultipleForAllPlayers(req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"achievementIds": achievementIDs,
		"allPlayers":     true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesScoresReset(ctx context.Context, leaderboardID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	resp, err := svc.Scores.Reset(leaderboardID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var resetScoreTimeSpans []map[string]interface{}
	for _, ts := range resp.ResetScoreTimeSpans {
		resetScoreTimeSpans = append(resetScoreTimeSpans, map[string]interface{}{
			"timeSpan": ts,
		})
	}

	result := output.NewResult(map[string]interface{}{
		"success":             true,
		"leaderboardId":       leaderboardID,
		"kind":                resp.Kind,
		"definitionId":        resp.DefinitionId,
		"resetScoreTimeSpans": resetScoreTimeSpans,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesScoresResetAll(ctx context.Context, allPlayers bool) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if allPlayers {
		err := svc.Scores.ResetAllForAllPlayers().Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		result := output.NewResult(map[string]interface{}{
			"success":    true,
			"operation":  "resetAllForAllPlayers",
			"allPlayers": true,
		})
		return c.Output(result.WithServices("gamesmanagement"))
	}

	resp, err := svc.Scores.ResetAll().Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var leaderboards []map[string]interface{}
	for _, l := range resp.Results {
		leaderboards = append(leaderboards, map[string]interface{}{
			"kind":                l.Kind,
			"definitionId":        l.DefinitionId,
			"resetScoreTimeSpans": l.ResetScoreTimeSpans,
		})
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"kind":         resp.Kind,
		"leaderboards": leaderboards,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesScoresResetForAllPlayers(ctx context.Context, leaderboardID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	err = svc.Scores.ResetForAllPlayers(leaderboardID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":       true,
		"leaderboardId": leaderboardID,
		"allPlayers":    true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesScoresResetMultipleForAllPlayers(ctx context.Context, leaderboardIDs []string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	req := &gamesmanagement.ScoresResetMultipleForAllRequest{
		LeaderboardIds: leaderboardIDs,
	}

	err = svc.Scores.ResetMultipleForAllPlayers(req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"leaderboardIds": leaderboardIDs,
		"allPlayers":     true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesEventsReset(ctx context.Context, eventID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	err = svc.Events.Reset(eventID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"eventId": eventID,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesEventsResetAll(ctx context.Context, allPlayers bool) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if allPlayers {
		err := svc.Events.ResetAllForAllPlayers().Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		result := output.NewResult(map[string]interface{}{
			"success":    true,
			"operation":  "resetAllForAllPlayers",
			"allPlayers": true,
		})
		return c.Output(result.WithServices("gamesmanagement"))
	}

	err = svc.Events.ResetAll().Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesEventsResetForAllPlayers(ctx context.Context, eventID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	err = svc.Events.ResetForAllPlayers(eventID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"eventId":    eventID,
		"allPlayers": true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesEventsResetMultipleForAllPlayers(ctx context.Context, eventIDs []string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	req := &gamesmanagement.EventsResetMultipleForAllRequest{
		EventIds: eventIDs,
	}

	err = svc.Events.ResetMultipleForAllPlayers(req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"eventIds":   eventIDs,
		"allPlayers": true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesPlayersHide(ctx context.Context, applicationID, playerID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	err = svc.Players.Hide(applicationID, playerID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"playerId":      playerID,
		"hidden":        true,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesPlayersUnhide(ctx context.Context, applicationID, playerID string) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	err = svc.Players.Unhide(applicationID, playerID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":       true,
		"applicationId": applicationID,
		"playerId":      playerID,
		"hidden":        false,
	})
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesApplicationsListHidden(ctx context.Context, applicationID string, pageSize int64, pageToken string, all bool) error {
	svc, err := c.getGamesManagementService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	req := svc.Applications.ListHidden(applicationID)
	if pageSize > 0 {
		req = req.MaxResults(pageSize)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	startToken := pageToken
	nextToken := ""
	var allPlayers []interface{}
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}

		for _, item := range resp.Items {
			player := map[string]interface{}{
				"kind":             item.Kind,
				"hiddenTimeMillis": item.HiddenTimeMillis,
			}
			if item.Player != nil {
				player["playerId"] = item.Player.PlayerId
				player["displayName"] = item.Player.DisplayName
				player["avatarImageUrl"] = item.Player.AvatarImageUrl
				player["bannerUrlPortrait"] = item.Player.BannerUrlPortrait
				player["bannerUrlLandscape"] = item.Player.BannerUrlLandscape
			}
			allPlayers = append(allPlayers, player)
		}

		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		req = req.PageToken(nextToken)
	}

	result := output.NewResult(allPlayers)
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("gamesmanagement"))
}

func (c *CLI) gamesCapabilities(_ context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"achievements": map[string]interface{}{
			"operations": []string{
				"reset",
				"resetAll",
				"resetForAllPlayers",
				"resetAllForAllPlayers",
				"resetMultipleForAllPlayers",
			},
			"flags": []string{
				"--all-players: Reset for all players (admin only)",
				"--ids: Multiple achievement IDs (comma-separated)",
			},
		},
		"scores": map[string]interface{}{
			"operations": []string{
				"reset",
				"resetAll",
				"resetForAllPlayers",
				"resetAllForAllPlayers",
				"resetMultipleForAllPlayers",
			},
			"flags": []string{
				"--all-players: Reset for all players (admin only)",
				"--ids: Multiple leaderboard IDs (comma-separated)",
			},
		},
		"events": map[string]interface{}{
			"operations": []string{
				"reset",
				"resetAll",
				"resetForAllPlayers",
				"resetAllForAllPlayers",
				"resetMultipleForAllPlayers",
			},
			"flags": []string{
				"--all-players: Reset for all players (admin only)",
				"--ids: Multiple event IDs (comma-separated)",
			},
		},
		"players": map[string]interface{}{
			"operations": []string{"hide", "unhide"},
			"flags": []string{
				"--application-id: Game application ID (required)",
			},
		},
		"applications": map[string]interface{}{
			"operations": []string{"list-hidden"},
			"flags": []string{
				"--page-size: Results per page",
				"--page-token: Pagination token",
				"--all: Fetch all pages",
			},
		},
		"scope": "https://www.googleapis.com/auth/games",
		"notes": []string{
			"Games Management API is for testing purposes",
			"Most operations require the Games OAuth scope",
			"--all-players operations require admin privileges",
			"Reset operations clear progress for testing",
		},
	})
	return c.Output(result.WithServices("gamesmanagement"))
}
