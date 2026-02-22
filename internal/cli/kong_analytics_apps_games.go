package cli

import (
	"github.com/dl-alexandre/gpd/internal/errors"
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
	return errors.NewAPIError(errors.CodeGeneralError, "analytics query not yet implemented")
}

// AnalyticsCapabilitiesCmd lists analytics capabilities.
type AnalyticsCapabilitiesCmd struct{}

// Run executes the analytics capabilities command.
func (cmd *AnalyticsCapabilitiesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "analytics capabilities not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "apps list not yet implemented")
}

// AppsGetCmd gets app details.
type AppsGetCmd struct {
	Package string `arg:"" help:"App package name (uses --package when omitted)"`
}

// Run executes the apps get command.
func (cmd *AppsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "apps get not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "games achievements reset not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "games scores reset not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "games events reset not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "games players hide not yet implemented")
}

// GamesPlayersUnhideCmd unhides a player.
type GamesPlayersUnhideCmd struct {
	PlayerID      string `arg:"" help:"Player ID to unhide"`
	ApplicationID string `help:"Game application ID (required)" required:""`
}

// Run executes the players unhide command.
func (cmd *GamesPlayersUnhideCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games players unhide not yet implemented")
}

// GamesCapabilitiesCmd lists Games management capabilities.
type GamesCapabilitiesCmd struct{}

// Run executes the games capabilities command.
func (cmd *GamesCapabilitiesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games capabilities not yet implemented")
}
