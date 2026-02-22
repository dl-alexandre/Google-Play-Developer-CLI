package cli

// Placeholder command structs for Kong CLI migration
// These need to be implemented with proper Run() methods

// ConfigCmd contains configuration commands.
// Subcommands are defined in kong_config.go.
type ConfigCmd struct {
	Init       ConfigInitCmd       `cmd:"" help:"Initialize project configuration"`
	Doctor     ConfigDoctorCmd     `cmd:"" help:"Diagnose configuration and credential issues"`
	Path       ConfigPathCmd       `cmd:"" help:"Show configuration paths"`
	Get        ConfigGetCmd        `cmd:"" help:"Get configuration value"`
	Set        ConfigSetCmd        `cmd:"" help:"Set configuration value"`
	Print      ConfigPrintCmd      `cmd:"" help:"Print current configuration"`
	Export     ConfigExportCmd     `cmd:"" help:"Export configuration to file"`
	Import     ConfigImportCmd     `cmd:"" help:"Import configuration from file"`
	Completion ConfigCompletionCmd `cmd:"" help:"Generate shell completion script"`
}

// Note: PurchasesCmd and MonetizationCmd are defined in kong_purchases_monetization.go

// MigrateCmd contains migration commands.
type MigrateCmd struct{}

// CustomAppCmd contains custom app publishing commands.
type CustomAppCmd struct{}

// GroupingCmd contains app access grouping commands.
type GroupingCmd struct{}
