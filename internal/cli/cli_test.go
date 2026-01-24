package cli

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

	requiredCommands := []string{"permissions", "recovery", "games"}
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
