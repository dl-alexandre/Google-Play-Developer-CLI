package cli

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func TestCommandRegistration(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	if rootCmd == nil {
		t.Fatal("rootCmd is nil")
	}

	commands := make(map[string]bool)
	for _, cmd := range rootCmd.Commands() {
		commands[cmd.Use] = true
	}

	requiredCommands := []string{"permissions", "recovery", "games", "customapp", "integrity", "grouping", "apps"}
	for _, cmdName := range requiredCommands {
		if !commands[cmdName] {
			t.Errorf("Command %q not found in root commands", cmdName)
		}
	}
}

func TestPermissionsCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	permissionsCmd := findCommand(rootCmd, "permissions")
	if permissionsCmd == nil {
		t.Fatal("permissions command not found")
	}

	subcommands := make(map[string]bool)
	for _, cmd := range permissionsCmd.Commands() {
		subcommands[cmd.Use] = true
	}

	requiredSubcommands := []string{"users", "grants", "capabilities"}
	for _, subcmdName := range requiredSubcommands {
		if !subcommands[subcmdName] {
			t.Errorf("Permissions subcommand %q not found", subcmdName)
		}
	}

	usersCmd := findCommand(permissionsCmd, "users")
	if usersCmd == nil {
		t.Fatal("permissions users command not found")
	}

	usersSubcommands := make(map[string]bool)
	for _, cmd := range usersCmd.Commands() {
		cmdName := cmd.Name()
		usersSubcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			usersSubcommands[parts[0]] = true
		}
	}

	requiredUsersSubcommands := []string{"create", "list", "get", "patch", "delete"}
	for _, subcmdName := range requiredUsersSubcommands {
		if !usersSubcommands[subcmdName] {
			t.Errorf("Users subcommand %q not found", subcmdName)
		}
	}

	grantsCmd := findCommand(permissionsCmd, "grants")
	if grantsCmd == nil {
		t.Fatal("permissions grants command not found")
	}

	grantsSubcommands := make(map[string]bool)
	for _, cmd := range grantsCmd.Commands() {
		cmdName := cmd.Name()
		grantsSubcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			grantsSubcommands[parts[0]] = true
		}
	}

	requiredGrantsSubcommands := []string{"create", "patch", "delete"}
	for _, subcmdName := range requiredGrantsSubcommands {
		if !grantsSubcommands[subcmdName] {
			t.Errorf("Grants subcommand %q not found", subcmdName)
		}
	}
}

func TestAuthCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	authCmd := requireCommand(t, rootCmd, "auth")
	checkRequiredSubcommands(t, getSubcommandNames(authCmd),
		[]string{"status", "check", "logout", "diagnose", "doctor", "login", "init", "switch", "list"},
		"Auth")
}

func TestReviewsCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	reviewsCmd := requireCommand(t, rootCmd, "reviews")
	checkRequiredSubcommands(t, getSubcommandNames(reviewsCmd), []string{"list", "reply", "get", "response", "capabilities"}, "Reviews")

	responseCmd := requireCommand(t, reviewsCmd, "response")
	checkRequiredSubcommands(t, getSubcommandNames(responseCmd), []string{"get", "delete"}, "Reviews response")
}

func TestRecoveryCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	recoveryCmd := findCommand(rootCmd, "recovery")
	if recoveryCmd == nil {
		t.Fatal("recovery command not found")
	}

	subcommands := make(map[string]bool)
	for _, cmd := range recoveryCmd.Commands() {
		cmdName := cmd.Name()
		subcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			subcommands[parts[0]] = true
		}
	}

	requiredSubcommands := []string{"create", "list", "deploy", "cancel", "add-targeting", "capabilities"}
	for _, subcmdName := range requiredSubcommands {
		if !subcommands[subcmdName] {
			t.Errorf("Recovery subcommand %q not found", subcmdName)
		}
	}
}

func TestGamesCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	gamesCmd := requireCommand(t, rootCmd, "games")

	subcommands := getSubcommandNames(gamesCmd)
	requiredSubcommands := []string{"achievements", "scores", "events", "players", "applications", "capabilities"}
	checkRequiredSubcommands(t, subcommands, requiredSubcommands, "Games")

	achievementsCmd := requireCommand(t, gamesCmd, "achievements")
	checkSubcommandExists(t, getSubcommandNames(achievementsCmd), "reset", "Achievements")

	scoresCmd := requireCommand(t, gamesCmd, "scores")
	checkSubcommandExists(t, getSubcommandNames(scoresCmd), "reset", "Scores")

	eventsCmd := requireCommand(t, gamesCmd, "events")
	checkSubcommandExists(t, getSubcommandNames(eventsCmd), "reset", "Events")

	playersCmd := requireCommand(t, gamesCmd, "players")
	checkRequiredSubcommands(t, getSubcommandNames(playersCmd), []string{"hide", "unhide"}, "Players")

	applicationsCmd := requireCommand(t, gamesCmd, "applications")
	checkSubcommandExists(t, getSubcommandNames(applicationsCmd), "list-hidden", "Applications")
}

func TestCustomAppCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	customAppCmd := requireCommand(t, rootCmd, "customapp")
	checkSubcommandExists(t, getSubcommandNames(customAppCmd), "create", "CustomApp")
}

func TestAppsCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	appsCmd := requireCommand(t, rootCmd, "apps")
	checkRequiredSubcommands(t, getSubcommandNames(appsCmd), []string{"list", "get"}, "Apps")
}

func TestIntegrityCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	integrityCmd := requireCommand(t, rootCmd, "integrity")
	checkRequiredSubcommands(t, getSubcommandNames(integrityCmd), []string{"decode"}, "Integrity")
}

func TestGroupingCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	groupingCmd := requireCommand(t, rootCmd, "grouping")
	checkRequiredSubcommands(t, getSubcommandNames(groupingCmd), []string{"token", "token-recall"}, "Grouping")
}

func TestPermissionsFlags(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	permissionsCmd := findCommand(rootCmd, "permissions")
	if permissionsCmd == nil {
		t.Fatal("permissions command not found")
	}

	usersCmd := findCommand(permissionsCmd, "users")
	if usersCmd == nil {
		t.Fatal("permissions users command not found")
	}

	usersCreateCmd := findCommand(usersCmd, "create")
	if usersCreateCmd == nil {
		t.Fatal("permissions users create command not found")
	}

	developerIDFlag := usersCreateCmd.Flag("developer-id")
	if developerIDFlag == nil {
		t.Error("--developer-id flag not found on permissions users create")
	}

	emailFlag := usersCreateCmd.Flag("email")
	if emailFlag == nil {
		t.Error("--email flag not found on permissions users create")
	}

	usersListCmd := findCommand(usersCmd, "list")
	if usersListCmd == nil {
		t.Fatal("permissions users list command not found")
	}

	developerIDFlagList := usersListCmd.Flag("developer-id")
	if developerIDFlagList == nil {
		t.Error("--developer-id flag not found on permissions users list")
	}

	grantsCmd := findCommand(permissionsCmd, "grants")
	if grantsCmd == nil {
		t.Fatal("permissions grants command not found")
	}

	grantsCreateCmd := findCommand(grantsCmd, "create")
	if grantsCreateCmd == nil {
		t.Fatal("permissions grants create command not found")
	}

	emailFlagGrants := grantsCreateCmd.Flag("email")
	if emailFlagGrants == nil {
		t.Error("--email flag not found on permissions grants create")
	}
}

func TestRecoveryFlags(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	recoveryCmd := findCommand(rootCmd, "recovery")
	if recoveryCmd == nil {
		t.Fatal("recovery command not found")
	}

	createCmd := findCommand(recoveryCmd, "create")
	if createCmd == nil {
		t.Fatal("recovery create command not found")
	}

	versionCodeFlag := createCmd.Flag("version-code")
	if versionCodeFlag == nil {
		t.Error("--version-code flag not found on recovery create")
	}

	fileFlag := createCmd.Flag("file")
	if fileFlag == nil {
		t.Error("--file flag not found on recovery create")
	}

	allUsersFlag := createCmd.Flag("all-users")
	if allUsersFlag == nil {
		t.Error("--all-users flag not found on recovery create")
	}

	listCmd := findCommand(recoveryCmd, "list")
	if listCmd == nil {
		t.Fatal("recovery list command not found")
	}

	versionCodeFlagList := listCmd.Flag("version-code")
	if versionCodeFlagList == nil {
		t.Error("--version-code flag not found on recovery list")
	}
}

func TestGamesFlags(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	gamesCmd := findCommand(rootCmd, "games")
	if gamesCmd == nil {
		t.Fatal("games command not found")
	}

	playersCmd := findCommand(gamesCmd, "players")
	if playersCmd == nil {
		t.Fatal("games players command not found")
	}

	hideCmd := findCommand(playersCmd, "hide")
	if hideCmd == nil {
		t.Fatal("games players hide command not found")
	}

	applicationIDFlag := hideCmd.Flag("application-id")
	if applicationIDFlag == nil {
		t.Error("--application-id flag not found on games players hide")
	}

	unhideCmd := findCommand(playersCmd, "unhide")
	if unhideCmd == nil {
		t.Fatal("games players unhide command not found")
	}

	applicationIDFlagUnhide := unhideCmd.Flag("application-id")
	if applicationIDFlagUnhide == nil {
		t.Error("--application-id flag not found on games players unhide")
	}

	achievementsCmd := findCommand(gamesCmd, "achievements")
	if achievementsCmd == nil {
		t.Fatal("games achievements command not found")
	}

	resetCmd := findCommand(achievementsCmd, "reset")
	if resetCmd == nil {
		t.Fatal("games achievements reset command not found")
	}

	allPlayersFlag := resetCmd.Flag("all-players")
	if allPlayersFlag == nil {
		t.Error("--all-players flag not found on games achievements reset")
	}

	idsFlag := resetCmd.Flag("ids")
	if idsFlag == nil {
		t.Error("--ids flag not found on games achievements reset")
	}
}

func TestCapabilitiesCommandsExist(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	permissionsCmd := findCommand(rootCmd, "permissions")
	if permissionsCmd == nil {
		t.Fatal("permissions command not found")
	}

	permissionsCapabilitiesCmd := findCommand(permissionsCmd, "capabilities")
	if permissionsCapabilitiesCmd == nil {
		t.Error("permissions capabilities command not found")
	}

	recoveryCmd := findCommand(rootCmd, "recovery")
	if recoveryCmd == nil {
		t.Fatal("recovery command not found")
	}

	recoveryCapabilitiesCmd := findCommand(recoveryCmd, "capabilities")
	if recoveryCapabilitiesCmd == nil {
		t.Error("recovery capabilities command not found")
	}

	gamesCmd := findCommand(rootCmd, "games")
	if gamesCmd == nil {
		t.Fatal("games command not found")
	}

	gamesCapabilitiesCmd := findCommand(gamesCmd, "capabilities")
	if gamesCapabilitiesCmd == nil {
		t.Error("games capabilities command not found")
	}
}

func TestSetupAppliesEnvAndOutput(t *testing.T) {
	t.Setenv("GPD_PACKAGE", "com.example.env")
	t.Setenv("GPD_STORE_TOKENS", "never")

	buf := &bytes.Buffer{}
	cli := New()
	cli.stdout = buf
	cli.outputMgr = output.NewManager(buf)
	cli.startTime = time.Now()

	cli.outputFormat = "markdown"
	cli.pretty = true
	cli.fields = "hello,world"

	if err := cli.setup(nil); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	if cli.packageName != "com.example.env" {
		t.Fatalf("packageName = %q, want %q", cli.packageName, "com.example.env")
	}

	result := output.NewResult(map[string]interface{}{"hello": "world"})
	if err := cli.Output(result); err != nil {
		t.Fatalf("Output error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "- **hello:** world") {
		t.Fatalf("expected markdown output, got %q", out)
	}
}

func TestOutputErrorWritesMarkdown(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := New()
	cli.stdout = buf
	cli.outputMgr = output.NewManager(buf)
	cli.outputFormat = "markdown"

	if err := cli.setup(nil); err != nil {
		t.Fatalf("setup error: %v", err)
	}

	apiErr := errors.NewAPIError(errors.CodeValidationError, "bad input")
	if err := cli.OutputError(apiErr); err != nil {
		t.Fatalf("OutputError error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "## Error") || !strings.Contains(out, "bad input") {
		t.Fatalf("expected markdown error output, got %q", out)
	}
}

func TestRequirePackage(t *testing.T) {
	cli := New()
	cli.packageName = ""
	if err := cli.requirePackage(); err == nil {
		t.Fatal("expected error when package is missing")
	}
	cli.packageName = "com.example.app"
	if err := cli.requirePackage(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHelpAgentCommand(t *testing.T) {
	buf := &bytes.Buffer{}
	cli := New()
	cli.stdout = buf
	cli.outputMgr = output.NewManager(buf)

	cli.rootCmd.SetArgs([]string{"help", "agent"})
	if err := cli.rootCmd.Execute(); err != nil {
		t.Fatalf("execute help agent error: %v", err)
	}

	if !strings.Contains(buf.String(), "AI Agent Quickstart Guide") {
		t.Fatalf("expected agent help output, got %q", buf.String())
	}
}

// getSubcommandNames extracts all subcommand names from a command.
// It includes both the command name and the first word of the Use field.
func getSubcommandNames(cmd *cobra.Command) map[string]bool {
	subcommands := make(map[string]bool)
	for _, c := range cmd.Commands() {
		subcommands[c.Name()] = true
		subcommands[c.Use] = true
		parts := strings.Fields(c.Use)
		if len(parts) > 0 {
			subcommands[parts[0]] = true
		}
	}
	return subcommands
}

// requireCommand finds a subcommand by name or fails the test.
func requireCommand(t *testing.T, parent *cobra.Command, name string) *cobra.Command {
	t.Helper()
	cmd := findCommand(parent, name)
	if cmd == nil {
		t.Fatalf("%s %s command not found", parent.Name(), name)
	}
	return cmd
}

// checkRequiredSubcommands verifies that all required subcommands exist.
func checkRequiredSubcommands(t *testing.T, subcommands map[string]bool, required []string, parentName string) {
	t.Helper()
	for _, name := range required {
		if !subcommands[name] {
			t.Errorf("%s subcommand %q not found", parentName, name)
		}
	}
}

// checkSubcommandExists verifies that a single subcommand exists.
func checkSubcommandExists(t *testing.T, subcommands map[string]bool, name, parentName string) {
	t.Helper()
	if !subcommands[name] {
		t.Errorf("%s %s subcommand not found", parentName, name)
	}
}

func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, cmd := range parent.Commands() {
		cmdName := cmd.Name()
		if cmdName == name {
			return cmd
		}
		if cmd.Use == name {
			return cmd
		}
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 && parts[0] == name {
			return cmd
		}
	}
	return nil
}

func TestAuthDiagnoseCommandFlags(t *testing.T) {
	cli := New()
	rootCmd := cli.rootCmd

	authCmd := requireCommand(t, rootCmd, "auth")
	diagnoseCmd := requireCommand(t, authCmd, "diagnose")

	refreshCheckFlag := diagnoseCmd.Flag("refresh-check")
	if refreshCheckFlag == nil {
		t.Fatal("--refresh-check flag not found on auth diagnose")
	}
	if refreshCheckFlag.Shorthand != "" {
		t.Error("--refresh-check should not have a shorthand")
	}
	if refreshCheckFlag.DefValue != "false" {
		t.Errorf("expected default value 'false', got %q", refreshCheckFlag.DefValue)
	}

	doctorCmd := requireCommand(t, authCmd, "doctor")
	doctorRefreshCheckFlag := doctorCmd.Flag("refresh-check")
	if doctorRefreshCheckFlag == nil {
		t.Error("--refresh-check flag not found on auth doctor")
	}
}

func TestAuthDiagnoseOutputFields(t *testing.T) {
	cli := New()

	buf := &bytes.Buffer{}
	cli.stdout = buf
	cli.outputMgr = output.NewManager(buf)

	t.Setenv("GPD_SERVICE_ACCOUNT_KEY", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
	t.Setenv("HOME", t.TempDir())

	cli.rootCmd.SetArgs([]string{"auth", "diagnose"})
	err := cli.rootCmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := buf.String()
	expectedFields := []string{
		"authenticated",
		"error",
	}

	for _, field := range expectedFields {
		if !strings.Contains(outputStr, field) {
			t.Errorf("expected output to contain field %q, got: %s", field, outputStr)
		}
	}

	if !strings.Contains(outputStr, `"authenticated":false`) {
		t.Errorf("expected authenticated:false in output, got: %s", outputStr)
	}
}

func TestAuthDiagnoseRefreshCheckFlag(t *testing.T) {
	cli := New()

	buf := &bytes.Buffer{}
	cli.stdout = buf
	cli.outputMgr = output.NewManager(buf)

	t.Setenv("GPD_SERVICE_ACCOUNT_KEY", "")
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
	t.Setenv("HOME", t.TempDir())

	cli.rootCmd.SetArgs([]string{"auth", "diagnose", "--refresh-check"})
	err := cli.rootCmd.Execute()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	outputStr := buf.String()
	if !strings.Contains(outputStr, "error") {
		t.Errorf("expected error field in output with no credentials, got: %s", outputStr)
	}
}
