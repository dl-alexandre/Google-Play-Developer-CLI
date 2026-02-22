// Package cli provides the Kong-based CLI framework for gpd.
package cli

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/dl-alexandre/gpd/internal/errors"
)

// Globals contains all global flags shared across all commands.
type Globals struct {
	Package     string        `help:"App package name" short:"p"`
	Output      string        `help:"Output format: json, table, markdown, csv" default:"json" enum:"json,table,markdown,csv"`
	Pretty      bool          `help:"Pretty print JSON output"`
	Timeout     time.Duration `help:"Network timeout" default:"30s"`
	StoreTokens string        `help:"Token storage: auto, never, secure" default:"auto" enum:"auto,never,secure"`
	Fields      string        `help:"JSON field projection (comma-separated paths)"`
	Quiet       bool          `help:"Suppress non-error output"`
	Verbose     bool          `help:"Enable verbose logging" short:"v"`
	KeyPath     string        `help:"Path to service account key file"`
	Profile     string        `help:"Configuration profile to use"`
}

// KongCLI represents the complete Kong CLI structure.
type KongCLI struct {
	Globals

	// Top-level commands
	Auth         AuthCmd         `cmd:"" help:"Authentication commands"`
	Config       ConfigCmd       `cmd:"" help:"Configuration commands"`
	Publish      PublishCmd      `cmd:"" help:"Publishing commands"`
	Reviews      ReviewsCmd      `cmd:"" help:"Review management commands"`
	Vitals       VitalsCmd       `cmd:"" help:"Android vitals commands"`
	Analytics    AnalyticsCmd    `cmd:"" help:"Analytics commands"`
	Purchases    PurchasesCmd    `cmd:"" help:"Purchase verification commands"`
	Monetization MonetizationCmd `cmd:"" help:"Monetization commands"`
	Permissions  PermissionsCmd  `cmd:"" help:"Permissions management"`
	Recovery     RecoveryCmd     `cmd:"" help:"App recovery commands"`
	Apps         AppsCmd         `cmd:"" help:"App discovery commands"`
	Games        GamesCmd        `cmd:"" help:"Google Play Games services"`
	Integrity    IntegrityCmd    `cmd:"" help:"Play Integrity API commands"`
	Migrate      MigrateCmd      `cmd:"" help:"Migration commands"`
	CustomApp    CustomAppCmd    `cmd:"" help:"Custom app publishing" aliases:"customapp"`
	Grouping     GroupingCmd     `cmd:"" help:"App access grouping"`
	Version      VersionCmd      `cmd:"" help:"Show version information"`

	// Help is automatically provided by Kong
}

// Run executes the Kong CLI and returns the exit code.
func RunKongCLI() int {
	var cli KongCLI

	parser, err := kong.New(&cli,
		kong.Name("gpd"),
		kong.Description("Google Play Developer CLI - A fast, lightweight command-line interface for the Google Play Developer Console."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating parser: %v\n", err)
		return errors.ExitGeneralError
	}

	ctx, err := parser.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return errors.ExitValidationError
	}

	// Note: Verbose logging setup would go here
	// Currently using default logging behavior

	// Execute the selected command
	err = ctx.Run(&cli.Globals)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			return apiErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return errors.ExitGeneralError
	}

	return errors.ExitSuccess
}
