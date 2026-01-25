package cli

import "github.com/spf13/cobra"

func (c *CLI) addMigrateCommands() {
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate metadata between formats",
		Long:  "Migrate metadata between external CI/CD formats and Google Play.",
	}

	c.addMigrateFastlaneCommands(migrateCmd)
	c.rootCmd.AddCommand(migrateCmd)
}
