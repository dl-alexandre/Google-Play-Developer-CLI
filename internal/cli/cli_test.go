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

	gamesCmd := findCommand(rootCmd, "games")
	if gamesCmd == nil {
		t.Fatal("games command not found")
	}

	subcommands := make(map[string]bool)
	for _, cmd := range gamesCmd.Commands() {
		subcommands[cmd.Use] = true
	}

	requiredSubcommands := []string{"achievements", "scores", "events", "players", "applications", "capabilities"}
	for _, subcmdName := range requiredSubcommands {
		if !subcommands[subcmdName] {
			t.Errorf("Games subcommand %q not found", subcmdName)
		}
	}

	achievementsCmd := findCommand(gamesCmd, "achievements")
	if achievementsCmd == nil {
		t.Fatal("games achievements command not found")
	}

	achievementsSubcommands := make(map[string]bool)
	for _, cmd := range achievementsCmd.Commands() {
		cmdName := cmd.Name()
		achievementsSubcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			achievementsSubcommands[parts[0]] = true
		}
	}

	if !achievementsSubcommands["reset"] {
		t.Error("Achievements reset subcommand not found")
	}

	scoresCmd := findCommand(gamesCmd, "scores")
	if scoresCmd == nil {
		t.Fatal("games scores command not found")
	}

	scoresSubcommands := make(map[string]bool)
	for _, cmd := range scoresCmd.Commands() {
		cmdName := cmd.Name()
		scoresSubcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			scoresSubcommands[parts[0]] = true
		}
	}

	if !scoresSubcommands["reset"] {
		t.Error("Scores reset subcommand not found")
	}

	eventsCmd := findCommand(gamesCmd, "events")
	if eventsCmd == nil {
		t.Fatal("games events command not found")
	}

	eventsSubcommands := make(map[string]bool)
	for _, cmd := range eventsCmd.Commands() {
		cmdName := cmd.Name()
		eventsSubcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			eventsSubcommands[parts[0]] = true
		}
	}

	if !eventsSubcommands["reset"] {
		t.Error("Events reset subcommand not found")
	}

	playersCmd := findCommand(gamesCmd, "players")
	if playersCmd == nil {
		t.Fatal("games players command not found")
	}

	playersSubcommands := make(map[string]bool)
	for _, cmd := range playersCmd.Commands() {
		cmdName := cmd.Name()
		playersSubcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			playersSubcommands[parts[0]] = true
		}
	}

	requiredPlayersSubcommands := []string{"hide", "unhide"}
	for _, subcmdName := range requiredPlayersSubcommands {
		if !playersSubcommands[subcmdName] {
			t.Errorf("Players subcommand %q not found", subcmdName)
		}
	}

	applicationsCmd := findCommand(gamesCmd, "applications")
	if applicationsCmd == nil {
		t.Fatal("games applications command not found")
	}

	applicationsSubcommands := make(map[string]bool)
	for _, cmd := range applicationsCmd.Commands() {
		cmdName := cmd.Name()
		applicationsSubcommands[cmdName] = true
		parts := strings.Fields(cmd.Use)
		if len(parts) > 0 {
			applicationsSubcommands[parts[0]] = true
		}
	}

	if !applicationsSubcommands["list-hidden"] {
		t.Error("Applications list-hidden subcommand not found")
	}
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
