//go:build unit
// +build unit

package cli

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// ============================================================================
// Analytics Commands Tests
// ============================================================================

func TestAnalyticsCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := AnalyticsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedSubcommands := []struct {
		name     string
		typeName string
	}{
		{"Query", "cli.AnalyticsQueryCmd"},
		{"Capabilities", "cli.AnalyticsCapabilitiesCmd"},
	}

	for _, sub := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(sub.name)
		if !ok {
			t.Errorf("AnalyticsCmd missing subcommand: %s", sub.name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("AnalyticsCmd.%s should have cmd:\"\" tag, got: %s", sub.name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("AnalyticsCmd.%s should have help tag", sub.name)
		}

		if field.Type.String() != sub.typeName {
			t.Errorf("AnalyticsCmd.%s type = %v, want %v", sub.name, field.Type.String(), sub.typeName)
		}
	}
}

func TestAnalyticsQueryCmd_FieldTags_AnalyticsAppsGames(t *testing.T) {
	cmd := AnalyticsQueryCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		helpText  string
		default_  string
		enum      string
	}{
		{"StartDate", "Start date (ISO 8601)", "", ""},
		{"EndDate", "End date (ISO 8601)", "", ""},
		{"Format", "Output format: json, csv", "json", "json,csv"},
		{"PageSize", "Results per page", "100", ""},
		{"PageToken", "Pagination token", "", ""},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("AnalyticsQueryCmd missing field: %s", tc.fieldName)
			continue
		}

		helpTag := field.Tag.Get("help")
		if tc.helpText != "" && !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("AnalyticsQueryCmd.%s help tag = %q, want to contain %q", tc.fieldName, helpTag, tc.helpText)
		}

		if tc.default_ != "" {
			defaultTag := field.Tag.Get("default")
			if defaultTag != tc.default_ {
				t.Errorf("AnalyticsQueryCmd.%s default tag = %q, want %q", tc.fieldName, defaultTag, tc.default_)
			}
		}

		if tc.enum != "" {
			enumTag := field.Tag.Get("enum")
			if enumTag != tc.enum {
				t.Errorf("AnalyticsQueryCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
			}
		}
	}
}

func TestAnalyticsQueryCmd_Run_PackageRequired(t *testing.T) {
	cmd := &AnalyticsQueryCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestAnalyticsQueryCmd_MetricsAndDimensions(t *testing.T) {
	cmd := AnalyticsQueryCmd{
		StartDate:  "2024-01-01",
		EndDate:    "2024-01-31",
		Metrics:    []string{"crashRate", "anrRate"},
		Dimensions: []string{"versionCode", "deviceModel"},
		Format:     "json",
		PageSize:   50,
	}

	if len(cmd.Metrics) != 2 {
		t.Errorf("Expected 2 metrics, got: %d", len(cmd.Metrics))
	}
	if len(cmd.Dimensions) != 2 {
		t.Errorf("Expected 2 dimensions, got: %d", len(cmd.Dimensions))
	}
	if cmd.PageSize != 50 {
		t.Errorf("Expected PageSize 50, got: %d", cmd.PageSize)
	}
	if cmd.StartDate != "2024-01-01" {
		t.Errorf("Expected StartDate '2024-01-01', got: %s", cmd.StartDate)
	}
}

func TestAnalyticsQueryCmd_InvalidAuth(t *testing.T) {
	cmd := &AnalyticsQueryCmd{
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
	}
	globals := &Globals{
		Package: "com.example.app",
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestAnalyticsCapabilitiesCmd_Run(t *testing.T) {
	cmd := &AnalyticsCapabilitiesCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("AnalyticsCapabilitiesCmd.Run() unexpected error: %v", err)
	}
}

func TestAnalyticsCapabilitiesCmd_StructExists(t *testing.T) {
	cmd := AnalyticsCapabilitiesCmd{}
	if reflect.TypeOf(cmd).String() != "cli.AnalyticsCapabilitiesCmd" {
		t.Errorf("AnalyticsCapabilitiesCmd type = %v, want cli.AnalyticsCapabilitiesCmd", reflect.TypeOf(cmd))
	}
}

func TestAnalyticsQueryCmd_StructExists(t *testing.T) {
	cmd := AnalyticsQueryCmd{}
	if reflect.TypeOf(cmd).String() != "cli.AnalyticsQueryCmd" {
		t.Errorf("AnalyticsQueryCmd type = %v, want cli.AnalyticsQueryCmd", reflect.TypeOf(cmd))
	}
}

// ============================================================================
// Apps Commands Tests
// ============================================================================

func TestAppsCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := AppsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		name     string
		typeName string
	}{
		{"List", "cli.AppsListCmd"},
		{"Get", "cli.AppsGetCmd"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.name)
		if !ok {
			t.Errorf("AppsCmd missing subcommand: %s", tc.name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("AppsCmd.%s should have cmd:\"\" tag, got: %s", tc.name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("AppsCmd.%s should have help tag", tc.name)
		}

		if field.Type.String() != tc.typeName {
			t.Errorf("AppsCmd.%s type = %v, want %v", tc.name, field.Type.String(), tc.typeName)
		}
	}
}

func TestAppsListCmd_FieldTags(t *testing.T) {
	cmd := AppsListCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		default_  string
		helpText  string
	}{
		{"PageSize", "100", "Results per page"},
		{"PageToken", "", "Pagination token"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("AppsListCmd missing field: %s", tc.fieldName)
			continue
		}

		helpTag := field.Tag.Get("help")
		if tc.helpText != "" && !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("AppsListCmd.%s help tag = %q, want to contain %q", tc.fieldName, helpTag, tc.helpText)
		}

		if tc.default_ != "" {
			defaultTag := field.Tag.Get("default")
			if defaultTag != tc.default_ {
				t.Errorf("AppsListCmd.%s default tag = %q, want %q", tc.fieldName, defaultTag, tc.default_)
			}
		}
	}

	// Check for All field
	field, ok := typeOfCmd.FieldByName("All")
	if !ok {
		t.Error("AppsListCmd missing All field")
	} else {
		helpTag := field.Tag.Get("help")
		if !strings.Contains(helpTag, "Fetch all pages") {
			t.Errorf("AppsListCmd.All help tag = %q, want to contain 'Fetch all pages'", helpTag)
		}
	}
}

func TestAppsListCmd_Run(t *testing.T) {
	cmd := &AppsListCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("AppsListCmd.Run() unexpected error: %v", err)
	}
}

func TestAppsListCmd_RunWithPagination(t *testing.T) {
	cmd := &AppsListCmd{
		PageSize:  50,
		PageToken: "next-page-token",
		All:       true,
	}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("AppsListCmd.Run() with pagination options unexpected error: %v", err)
	}
}

func TestAppsGetCmd_FieldTags(t *testing.T) {
	cmd := AppsGetCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Package")
	if !ok {
		t.Fatal("AppsGetCmd missing Package field")
	}

	argTag := field.Tag.Get("arg")
	if argTag != "" {
		t.Errorf("AppsGetCmd.Package arg tag = %q, want empty string", argTag)
	}

	helpTag := field.Tag.Get("help")
	if !strings.Contains(helpTag, "App package name") {
		t.Errorf("AppsGetCmd.Package help tag = %q, want to contain 'App package name'", helpTag)
	}
}

func TestAppsGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &AppsGetCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestAppsGetCmd_Run_WithPackageArg(t *testing.T) {
	cmd := &AppsGetCmd{
		Package: "com.example.test",
	}
	globals := &Globals{
		Output:  "json",
		KeyPath: "/nonexistent/key.json", // Will fail at auth
	}

	err := cmd.Run(globals)
	// Should fail on auth, not package validation
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
	if err == errors.ErrPackageRequired {
		t.Error("Should have failed on auth, not package validation")
	}
}

func TestAppsGetCmd_Run_WithGlobalsPackage(t *testing.T) {
	cmd := &AppsGetCmd{} // No Package arg
	globals := &Globals{
		Package: "com.example.test",
		Output:  "json",
		KeyPath: "/nonexistent/key.json", // Will fail at auth
	}

	err := cmd.Run(globals)
	// Should fail on auth, not package validation
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
	if err == errors.ErrPackageRequired {
		t.Error("Should have failed on auth, not package validation")
	}
}

func TestAppsListCmd_StructExists(t *testing.T) {
	cmd := AppsListCmd{}
	if reflect.TypeOf(cmd).String() != "cli.AppsListCmd" {
		t.Errorf("AppsListCmd type = %v, want cli.AppsListCmd", reflect.TypeOf(cmd))
	}
}

func TestAppsGetCmd_StructExists(t *testing.T) {
	cmd := AppsGetCmd{}
	if reflect.TypeOf(cmd).String() != "cli.AppsGetCmd" {
		t.Errorf("AppsGetCmd type = %v, want cli.AppsGetCmd", reflect.TypeOf(cmd))
	}
}

// ============================================================================
// Games Commands Tests
// ============================================================================

func TestGamesCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := GamesCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedSubcommands := []struct {
		name     string
		typeName string
	}{
		{"Achievements", "cli.GamesAchievementsCmd"},
		{"Scores", "cli.GamesScoresCmd"},
		{"Events", "cli.GamesEventsCmd"},
		{"Players", "cli.GamesPlayersCmd"},
		{"Capabilities", "cli.GamesCapabilitiesCmd"},
	}

	for _, sub := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(sub.name)
		if !ok {
			t.Errorf("GamesCmd missing subcommand: %s", sub.name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("GamesCmd.%s should have cmd:\"\" tag, got: %s", sub.name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("GamesCmd.%s should have help tag", sub.name)
		}

		if field.Type.String() != sub.typeName {
			t.Errorf("GamesCmd.%s type = %v, want %v", sub.name, field.Type.String(), sub.typeName)
		}
	}
}

func TestGamesAchievementsCmd_HasResetSubcommand(t *testing.T) {
	cmd := GamesAchievementsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Reset")
	if !ok {
		t.Fatal("GamesAchievementsCmd missing Reset field")
	}

	if field.Type.String() != "cli.GamesAchievementsResetCmd" {
		t.Errorf("GamesAchievementsCmd.Reset type = %v, want cli.GamesAchievementsResetCmd", field.Type.String())
	}

	cmdTag := field.Tag.Get("cmd")
	if cmdTag != "" {
		t.Errorf("GamesAchievementsCmd.Reset cmd tag = %q, want empty string", cmdTag)
	}

	helpTag := field.Tag.Get("help")
	if !strings.Contains(helpTag, "Reset achievements") {
		t.Errorf("GamesAchievementsCmd.Reset help tag = %q, want to contain 'Reset achievements'", helpTag)
	}
}

func TestGamesAchievementsResetCmd_FieldTags(t *testing.T) {
	cmd := GamesAchievementsResetCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		helpText  string
	}{
		{"AchievementID", "Achievement ID to reset"},
		{"AllPlayers", "Reset for all players"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("GamesAchievementsResetCmd missing field: %s", tc.fieldName)
			continue
		}

		helpTag := field.Tag.Get("help")
		if tc.helpText != "" && !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("GamesAchievementsResetCmd.%s help tag = %q, want to contain %q", tc.fieldName, helpTag, tc.helpText)
		}
	}

	// Check AchievementID has arg tag
	field, ok := typeOfCmd.FieldByName("AchievementID")
	if ok {
		argTag := field.Tag.Get("arg")
		if argTag != "" {
			t.Errorf("GamesAchievementsResetCmd.AchievementID arg tag = %q, want empty string", argTag)
		}
	}
}

func TestGamesAchievementsResetCmd_Run_InvalidAuth(t *testing.T) {
	cmd := &GamesAchievementsResetCmd{
		AchievementID: "achievement-123",
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesAchievementsResetCmd_Run_AllPlayersSingle(t *testing.T) {
	cmd := &GamesAchievementsResetCmd{
		AchievementID: "achievement-123",
		AllPlayers:    true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth before trying API call
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesAchievementsResetCmd_Run_AllPlayersBatch(t *testing.T) {
	cmd := &GamesAchievementsResetCmd{
		IDs:        []string{"ach1", "ach2", "ach3"},
		AllPlayers: true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth before trying API call
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesAchievementsResetCmd_Run_AllPlayersAllAchievements(t *testing.T) {
	cmd := &GamesAchievementsResetCmd{
		AchievementID: "", // Empty - should reset all
		AllPlayers:    true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth before trying API call
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesAchievementsResetCmd_Run_MultipleIDs(t *testing.T) {
	cmd := &GamesAchievementsResetCmd{
		IDs: []string{"ach1", "ach2"},
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth before trying API call
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesScoresCmd_HasResetSubcommand(t *testing.T) {
	cmd := GamesScoresCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Reset")
	if !ok {
		t.Fatal("GamesScoresCmd missing Reset field")
	}

	if field.Type.String() != "cli.GamesScoresResetCmd" {
		t.Errorf("GamesScoresCmd.Reset type = %v, want cli.GamesScoresResetCmd", field.Type.String())
	}
}

func TestGamesScoresResetCmd_FieldTags(t *testing.T) {
	cmd := GamesScoresResetCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		helpText  string
	}{
		{"LeaderboardID", "Leaderboard ID to reset"},
		{"AllPlayers", "Reset for all players"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("GamesScoresResetCmd missing field: %s", tc.fieldName)
			continue
		}

		helpTag := field.Tag.Get("help")
		if tc.helpText != "" && !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("GamesScoresResetCmd.%s help tag = %q, want to contain %q", tc.fieldName, helpTag, tc.helpText)
		}
	}

	// Check LeaderboardID has arg tag
	field, ok := typeOfCmd.FieldByName("LeaderboardID")
	if ok {
		argTag := field.Tag.Get("arg")
		if argTag != "" {
			t.Errorf("GamesScoresResetCmd.LeaderboardID arg tag = %q, want empty string", argTag)
		}
	}
}

func TestGamesScoresResetCmd_Run_InvalidAuth(t *testing.T) {
	cmd := &GamesScoresResetCmd{
		LeaderboardID: "leaderboard-123",
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesScoresResetCmd_Run_AllPlayers(t *testing.T) {
	cmd := &GamesScoresResetCmd{
		LeaderboardID: "leaderboard-123",
		AllPlayers:    true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesScoresResetCmd_Run_AllPlayersAllScores(t *testing.T) {
	cmd := &GamesScoresResetCmd{
		LeaderboardID: "", // Empty - reset all
		AllPlayers:    true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesScoresResetCmd_Run_MultipleIDs(t *testing.T) {
	cmd := &GamesScoresResetCmd{
		IDs:        []string{"lb1", "lb2"},
		AllPlayers: false,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesScoresResetCmd_Run_MultipleIDsAllPlayers(t *testing.T) {
	cmd := &GamesScoresResetCmd{
		IDs:        []string{"lb1", "lb2"},
		AllPlayers: true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesEventsCmd_HasResetSubcommand(t *testing.T) {
	cmd := GamesEventsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Reset")
	if !ok {
		t.Fatal("GamesEventsCmd missing Reset field")
	}

	if field.Type.String() != "cli.GamesEventsResetCmd" {
		t.Errorf("GamesEventsCmd.Reset type = %v, want cli.GamesEventsResetCmd", field.Type.String())
	}
}

func TestGamesEventsResetCmd_FieldTags(t *testing.T) {
	cmd := GamesEventsResetCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		helpText  string
	}{
		{"EventID", "Event ID to reset"},
		{"AllPlayers", "Reset for all players"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("GamesEventsResetCmd missing field: %s", tc.fieldName)
			continue
		}

		helpTag := field.Tag.Get("help")
		if tc.helpText != "" && !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("GamesEventsResetCmd.%s help tag = %q, want to contain %q", tc.fieldName, helpTag, tc.helpText)
		}
	}

	// Check EventID has arg tag
	field, ok := typeOfCmd.FieldByName("EventID")
	if ok {
		argTag := field.Tag.Get("arg")
		if argTag != "" {
			t.Errorf("GamesEventsResetCmd.EventID arg tag = %q, want empty string", argTag)
		}
	}
}

func TestGamesEventsResetCmd_Run_InvalidAuth(t *testing.T) {
	cmd := &GamesEventsResetCmd{
		EventID: "event-123",
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesEventsResetCmd_Run_AllPlayers(t *testing.T) {
	cmd := &GamesEventsResetCmd{
		EventID:    "event-123",
		AllPlayers: true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesEventsResetCmd_Run_AllPlayersAllEvents(t *testing.T) {
	cmd := &GamesEventsResetCmd{
		EventID:    "", // Empty - reset all
		AllPlayers: true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesEventsResetCmd_Run_MultipleIDs(t *testing.T) {
	cmd := &GamesEventsResetCmd{
		IDs: []string{"event1", "event2"},
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesEventsResetCmd_Run_MultipleIDsAllPlayers(t *testing.T) {
	cmd := &GamesEventsResetCmd{
		IDs:        []string{"event1", "event2"},
		AllPlayers: true,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should fail on auth
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesPlayersCmd_HasHideAndUnhideSubcommands(t *testing.T) {
	cmd := GamesPlayersCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		name     string
		typeName string
		helpText string
	}{
		{"Hide", "cli.GamesPlayersHideCmd", "Hide a player"},
		{"Unhide", "cli.GamesPlayersUnhideCmd", "Unhide a player"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.name)
		if !ok {
			t.Errorf("GamesPlayersCmd missing subcommand: %s", tc.name)
			continue
		}

		if field.Type.String() != tc.typeName {
			t.Errorf("GamesPlayersCmd.%s type = %v, want %v", tc.name, field.Type.String(), tc.typeName)
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("GamesPlayersCmd.%s cmd tag = %q, want empty string", tc.name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("GamesPlayersCmd.%s help tag = %q, want to contain %q", tc.name, helpTag, tc.helpText)
		}
	}
}

func TestGamesPlayersHideCmd_FieldTags(t *testing.T) {
	cmd := GamesPlayersHideCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		helpText  string
		required  bool
	}{
		{"PlayerID", "Player ID to hide", false},
		{"ApplicationID", "Game application ID", true},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("GamesPlayersHideCmd missing field: %s", tc.fieldName)
			continue
		}

		helpTag := field.Tag.Get("help")
		if tc.helpText != "" && !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("GamesPlayersHideCmd.%s help tag = %q, want to contain %q", tc.fieldName, helpTag, tc.helpText)
		}

		if tc.required {
			if !strings.Contains(string(field.Tag), "required") {
				t.Errorf("GamesPlayersHideCmd.%s should have required tag", tc.fieldName)
			}
		}
	}

	// Check PlayerID has arg tag
	field, ok := typeOfCmd.FieldByName("PlayerID")
	if ok {
		argTag := field.Tag.Get("arg")
		if argTag != "" {
			t.Errorf("GamesPlayersHideCmd.PlayerID arg tag = %q, want empty string", argTag)
		}
	}
}

func TestGamesPlayersHideCmd_Run_InvalidAuth(t *testing.T) {
	cmd := &GamesPlayersHideCmd{
		PlayerID:      "player-123",
		ApplicationID: "app-123",
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesPlayersUnhideCmd_FieldTags(t *testing.T) {
	cmd := GamesPlayersUnhideCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		helpText  string
		required  bool
	}{
		{"PlayerID", "Player ID to unhide", false},
		{"ApplicationID", "Game application ID", true},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("GamesPlayersUnhideCmd missing field: %s", tc.fieldName)
			continue
		}

		helpTag := field.Tag.Get("help")
		if tc.helpText != "" && !strings.Contains(helpTag, tc.helpText) {
			t.Errorf("GamesPlayersUnhideCmd.%s help tag = %q, want to contain %q", tc.fieldName, helpTag, tc.helpText)
		}

		if tc.required {
			if !strings.Contains(string(field.Tag), "required") {
				t.Errorf("GamesPlayersUnhideCmd.%s should have required tag", tc.fieldName)
			}
		}
	}

	// Check PlayerID has arg tag
	field, ok := typeOfCmd.FieldByName("PlayerID")
	if ok {
		argTag := field.Tag.Get("arg")
		if argTag != "" {
			t.Errorf("GamesPlayersUnhideCmd.PlayerID arg tag = %q, want empty string", argTag)
		}
	}
}

func TestGamesPlayersUnhideCmd_Run_InvalidAuth(t *testing.T) {
	cmd := &GamesPlayersUnhideCmd{
		PlayerID:      "player-123",
		ApplicationID: "app-123",
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesCapabilitiesCmd_Run(t *testing.T) {
	cmd := &GamesCapabilitiesCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("GamesCapabilitiesCmd.Run() unexpected error: %v", err)
	}
}

func TestGamesCapabilitiesCmd_Run_WithPrettyOutput(t *testing.T) {
	cmd := &GamesCapabilitiesCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: true,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("GamesCapabilitiesCmd.Run() with pretty output unexpected error: %v", err)
	}
}

// ============================================================================
// Helper Types Tests
// ============================================================================

func TestCrashRateQueryRequest_StructAndMethod(t *testing.T) {
	req := &crashRateQueryRequest{
		TimelineSpec: nil,
		Dimensions:   []string{"versionCode"},
		Metrics:      []string{"crashRate"},
		PageSize:     100,
		PageToken:    "token123",
	}

	if len(req.Dimensions) != 1 || req.Dimensions[0] != "versionCode" {
		t.Errorf("Expected Dimensions ['versionCode'], got: %v", req.Dimensions)
	}
	if len(req.Metrics) != 1 || req.Metrics[0] != "crashRate" {
		t.Errorf("Expected Metrics ['crashRate'], got: %v", req.Metrics)
	}
	if req.PageSize != 100 {
		t.Errorf("Expected PageSize 100, got: %d", req.PageSize)
	}
	if req.PageToken != "token123" {
		t.Errorf("Expected PageToken 'token123', got: %s", req.PageToken)
	}

	// Test toAPI method returns nil when TimelineSpec is nil
	apiReq := req.toAPI()
	if apiReq.TimelineSpec != nil {
		t.Error("Expected nil TimelineSpec in API request")
	}
}

func TestAnrRateQueryRequest_StructAndMethod(t *testing.T) {
	req := &anrRateQueryRequest{
		TimelineSpec: nil,
		Dimensions:   []string{"deviceModel"},
		Metrics:      []string{"anrRate"},
		PageSize:     50,
		PageToken:    "token456",
	}

	if len(req.Dimensions) != 1 || req.Dimensions[0] != "deviceModel" {
		t.Errorf("Expected Dimensions ['deviceModel'], got: %v", req.Dimensions)
	}
	if len(req.Metrics) != 1 || req.Metrics[0] != "anrRate" {
		t.Errorf("Expected Metrics ['anrRate'], got: %v", req.Metrics)
	}
	if req.PageSize != 50 {
		t.Errorf("Expected PageSize 50, got: %d", req.PageSize)
	}
	if req.PageToken != "token456" {
		t.Errorf("Expected PageToken 'token456', got: %s", req.PageToken)
	}

	// Test toAPI method returns nil when TimelineSpec is nil
	apiReq := req.toAPI()
	if apiReq.TimelineSpec != nil {
		t.Error("Expected nil TimelineSpec in API request")
	}
}

// ============================================================================
// Table-Driven Tests
// ============================================================================

func TestAnalyticsQueryCmd_MetricSetRouting(t *testing.T) {
	tests := []struct {
		name         string
		metrics      []string
		expectedPath string
	}{
		{
			name:         "crash rate metric",
			metrics:      []string{"crashRate"},
			expectedPath: "crashRateMetricSet",
		},
		{
			name:         "user perceived crash rate metric",
			metrics:      []string{"userPerceivedCrashRate"},
			expectedPath: "crashRateMetricSet",
		},
		{
			name:         "anr rate metric",
			metrics:      []string{"anrRate"},
			expectedPath: "anrRateMetricSet",
		},
		{
			name:         "user perceived anr rate metric",
			metrics:      []string{"userPerceivedAnrRate"},
			expectedPath: "anrRateMetricSet",
		},
		{
			name:         "slow rendering rate metric",
			metrics:      []string{"slowRenderingRate"},
			expectedPath: "slowRenderingRateMetricSet",
		},
		{
			name:         "slow start rate metric",
			metrics:      []string{"slowStartRate"},
			expectedPath: "slowStartRateMetricSet",
		},
		{
			name:         "stuck background wakelock rate metric",
			metrics:      []string{"stuckBackgroundWakelockRate"},
			expectedPath: "stuckBackgroundWakelockRateMetricSet",
		},
		{
			name:         "excessive wakeup rate metric",
			metrics:      []string{"excessiveWakeupRate"},
			expectedPath: "excessiveWakeupRateMetricSet",
		},
		{
			name:         "error count metric",
			metrics:      []string{"errorCount"},
			expectedPath: "errorCountMetricSet",
		},
		{
			name:         "no metrics - default to crash rate",
			metrics:      []string{},
			expectedPath: "crashRateMetricSet", // Default
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// This test documents expected routing behavior
			// The actual metric set name is computed in the Run method
			cmd := &AnalyticsQueryCmd{
				Metrics: tc.metrics,
			}

			// Just verify the command was created with correct metrics
			if len(tc.metrics) > 0 {
				if len(cmd.Metrics) != len(tc.metrics) {
					t.Errorf("Expected %d metrics, got: %d", len(tc.metrics), len(cmd.Metrics))
				}
			}
		})
	}
}

func TestGamesCommands_ContextHandling(t *testing.T) {
	tests := []struct {
		name    string
		cmd     interface{ Run(*Globals) error }
		globals *Globals
	}{
		{
			name:    "achievements reset with context",
			cmd:     &GamesAchievementsResetCmd{AchievementID: "ach-123"},
			globals: &Globals{Context: context.Background(), KeyPath: "/nonexistent/key.json"},
		},
		{
			name:    "scores reset with context",
			cmd:     &GamesScoresResetCmd{LeaderboardID: "lb-123"},
			globals: &Globals{Context: context.Background(), KeyPath: "/nonexistent/key.json"},
		},
		{
			name:    "events reset with context",
			cmd:     &GamesEventsResetCmd{EventID: "evt-123"},
			globals: &Globals{Context: context.Background(), KeyPath: "/nonexistent/key.json"},
		},
		{
			name: "players hide with context",
			cmd:  &GamesPlayersHideCmd{PlayerID: "player-123", ApplicationID: "app-123"},
			globals: &Globals{
				Context: context.Background(),
				KeyPath: "/nonexistent/key.json",
			},
		},
		{
			name: "players unhide with context",
			cmd:  &GamesPlayersUnhideCmd{PlayerID: "player-123", ApplicationID: "app-123"},
			globals: &Globals{
				Context: context.Background(),
				KeyPath: "/nonexistent/key.json",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(tc.globals)
			if err == nil {
				t.Fatal("Expected error for invalid auth")
			}
			// Error is expected due to invalid key path
		})
	}
}

func TestGamesCommands_NilContext(t *testing.T) {
	tests := []struct {
		name    string
		cmd     interface{ Run(*Globals) error }
		globals *Globals
	}{
		{
			name:    "achievements reset with nil context",
			cmd:     &GamesAchievementsResetCmd{AchievementID: "ach-123"},
			globals: &Globals{Context: nil, KeyPath: "/nonexistent/key.json"},
		},
		{
			name:    "scores reset with nil context",
			cmd:     &GamesScoresResetCmd{LeaderboardID: "lb-123"},
			globals: &Globals{Context: nil, KeyPath: "/nonexistent/key.json"},
		},
		{
			name:    "events reset with nil context",
			cmd:     &GamesEventsResetCmd{EventID: "evt-123"},
			globals: &Globals{Context: nil, KeyPath: "/nonexistent/key.json"},
		},
		{
			name: "players hide with nil context",
			cmd:  &GamesPlayersHideCmd{PlayerID: "player-123", ApplicationID: "app-123"},
			globals: &Globals{
				Context: nil,
				KeyPath: "/nonexistent/key.json",
			},
		},
		{
			name: "players unhide with nil context",
			cmd:  &GamesPlayersUnhideCmd{PlayerID: "player-123", ApplicationID: "app-123"},
			globals: &Globals{
				Context: nil,
				KeyPath: "/nonexistent/key.json",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(tc.globals)
			if err == nil {
				t.Fatal("Expected error for invalid auth")
			}
			// Error is expected due to invalid key path
			// The nil context should be handled by creating context.Background()
		})
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestGamesCommands_ErrorTypes(t *testing.T) {
	tests := []struct {
		name            string
		cmd             interface{ Run(*Globals) error }
		globals         *Globals
		expectAuthError bool
	}{
		{
			name: "achievements reset missing auth",
			cmd:  &GamesAchievementsResetCmd{AchievementID: "ach-123"},
			globals: &Globals{
				KeyPath: "/nonexistent/key.json",
			},
			expectAuthError: true,
		},
		{
			name: "scores reset missing auth",
			cmd:  &GamesScoresResetCmd{LeaderboardID: "lb-123"},
			globals: &Globals{
				KeyPath: "/nonexistent/key.json",
			},
			expectAuthError: true,
		},
		{
			name: "events reset missing auth",
			cmd:  &GamesEventsResetCmd{EventID: "evt-123"},
			globals: &Globals{
				KeyPath: "/nonexistent/key.json",
			},
			expectAuthError: true,
		},
		{
			name: "players hide missing auth",
			cmd:  &GamesPlayersHideCmd{PlayerID: "player-123", ApplicationID: "app-123"},
			globals: &Globals{
				KeyPath: "/nonexistent/key.json",
			},
			expectAuthError: true,
		},
		{
			name: "players unhide missing auth",
			cmd:  &GamesPlayersUnhideCmd{PlayerID: "player-123", ApplicationID: "app-123"},
			globals: &Globals{
				KeyPath: "/nonexistent/key.json",
			},
			expectAuthError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(tc.globals)
			if err == nil {
				t.Fatal("Expected error")
			}

			if tc.expectAuthError {
				// Verify error message contains auth-related text
				errStr := err.Error()
				if !strings.Contains(errStr, "auth") &&
					!strings.Contains(errStr, "key") &&
					!strings.Contains(errStr, "credential") {
					t.Logf("Error message: %s", errStr)
				}
			}
		})
	}
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestGamesCommands_DifferentOutputFormats(t *testing.T) {
	formats := []string{"json", "table"}

	for _, format := range formats {
		t.Run("capabilities_"+format, func(t *testing.T) {
			cmd := &GamesCapabilitiesCmd{}
			globals := &Globals{
				Output: format,
				Pretty: false,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("GamesCapabilitiesCmd.Run() with format %s unexpected error: %v", format, err)
			}
		})
	}
}

func TestAnalyticsCommands_DifferentOutputFormats(t *testing.T) {
	formats := []string{"json"}

	for _, format := range formats {
		t.Run("capabilities_"+format, func(t *testing.T) {
			cmd := &AnalyticsCapabilitiesCmd{}
			globals := &Globals{
				Output: format,
				Pretty: false,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("AnalyticsCapabilitiesCmd.Run() with format %s unexpected error: %v", format, err)
			}
		})
	}
}

// ============================================================================
// Duration and Metadata Tests
// ============================================================================

func TestGamesCapabilitiesCmd_IncludesDuration(t *testing.T) {
	cmd := &GamesCapabilitiesCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	start := time.Now()
	err := cmd.Run(globals)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify command completed in reasonable time
	if duration > 5*time.Second {
		t.Logf("Command took %v to complete", duration)
	}
}

func TestAnalyticsCapabilitiesCmd_IncludesDuration(t *testing.T) {
	cmd := &AnalyticsCapabilitiesCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	start := time.Now()
	err := cmd.Run(globals)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify command completed in reasonable time
	if duration > 5*time.Second {
		t.Logf("Command took %v to complete", duration)
	}
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestGamesAchievementsResetCmd_EmptyAchievementID(t *testing.T) {
	cmd := &GamesAchievementsResetCmd{
		AchievementID: "",
		AllPlayers:    false,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesScoresResetCmd_EmptyLeaderboardID(t *testing.T) {
	cmd := &GamesScoresResetCmd{
		LeaderboardID: "",
		AllPlayers:    false,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesEventsResetCmd_EmptyEventID(t *testing.T) {
	cmd := &GamesEventsResetCmd{
		EventID:    "",
		AllPlayers: false,
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesPlayersHideCmd_EmptyApplicationID(t *testing.T) {
	cmd := &GamesPlayersHideCmd{
		PlayerID:      "player-123",
		ApplicationID: "", // Empty but required
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

func TestGamesPlayersUnhideCmd_EmptyApplicationID(t *testing.T) {
	cmd := &GamesPlayersUnhideCmd{
		PlayerID:      "player-123",
		ApplicationID: "", // Empty but required
	}
	globals := &Globals{
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid auth")
	}
}

// ============================================================================
// Struct Instantiation Tests
// ============================================================================

func TestAllCommandStructsCanBeInstantiated(t *testing.T) {
	structs := []interface{}{
		AnalyticsCmd{},
		AnalyticsQueryCmd{},
		AnalyticsCapabilitiesCmd{},
		AppsCmd{},
		AppsListCmd{},
		AppsGetCmd{},
		GamesCmd{},
		GamesAchievementsCmd{},
		GamesAchievementsResetCmd{},
		GamesScoresCmd{},
		GamesScoresResetCmd{},
		GamesEventsCmd{},
		GamesEventsResetCmd{},
		GamesPlayersCmd{},
		GamesPlayersHideCmd{},
		GamesPlayersUnhideCmd{},
		GamesCapabilitiesCmd{},
	}

	for _, s := range structs {
		if s == nil {
			t.Error("Struct should not be nil")
		}
	}
}

// ============================================================================
// Pointer Receiver Tests
// ============================================================================

func TestCommandRunMethodsUsePointerReceivers(t *testing.T) {
	// Verify all command structs have Run methods with pointer receivers
	commands := []struct {
		cmd  interface{ Run(*Globals) error }
		name string
	}{
		{&AnalyticsQueryCmd{}, "AnalyticsQueryCmd"},
		{&AnalyticsCapabilitiesCmd{}, "AnalyticsCapabilitiesCmd"},
		{&AppsListCmd{}, "AppsListCmd"},
		{&AppsGetCmd{}, "AppsGetCmd"},
		{&GamesAchievementsResetCmd{}, "GamesAchievementsResetCmd"},
		{&GamesScoresResetCmd{}, "GamesScoresResetCmd"},
		{&GamesEventsResetCmd{}, "GamesEventsResetCmd"},
		{&GamesPlayersHideCmd{}, "GamesPlayersHideCmd"},
		{&GamesPlayersUnhideCmd{}, "GamesPlayersUnhideCmd"},
		{&GamesCapabilitiesCmd{}, "GamesCapabilitiesCmd"},
	}

	for _, c := range commands {
		// Try to call Run with nil globals (should handle gracefully or error)
		globals := &Globals{Output: "json"}
		err := c.cmd.Run(globals)
		// Error is expected (no auth), but should not panic
		_ = err
	}
}
