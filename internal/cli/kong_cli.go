// Package cli provides the Kong-based CLI framework for gpd.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/alecthomas/kong"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/cache"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/extensions"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/logging"
)

// Globals contains all global flags shared across all commands.
type Globals struct {
	Package     string        `help:"App package name" short:"p"`
	Output      string        `help:"Output format: json, table, markdown, csv, excel" default:"json" enum:"json,table,markdown,csv,excel"`
	Pretty      bool          `help:"Pretty print JSON output"`
	Timeout     time.Duration `help:"Network timeout" default:"30s"`
	StoreTokens string        `help:"Token storage: auto, never, secure" default:"auto" enum:"auto,never,secure"`
	Fields      string        `help:"JSON field projection (comma-separated paths)"`
	Quiet       bool          `help:"Suppress non-error output"`
	Verbose     bool          `help:"Enable verbose logging" short:"v"`
	KeyPath     string        `help:"Path to service account key file"`
	Profile     string        `help:"Configuration profile to use"`
	CacheDir    string        `help:"Cache directory for temporary data" env:"GPD_CACHE_DIR"`

	// Context is set by RunKongCLI and propagated to commands
	Context context.Context `kong:"-"`

	// Cache is initialized by RunKongCLI
	Cache *cache.Cache `kong:"-"`
}

// KongCLI represents the complete Kong CLI structure.
type KongCLI struct {
	Globals

	// Top-level commands
	Auth          AuthCmd          `cmd:"" help:"Authentication commands"`
	Config        ConfigCmd        `cmd:"" help:"Configuration commands"`
	Publish       PublishCmd       `cmd:"" help:"Publishing commands"`
	Reviews       ReviewsCmd       `cmd:"" help:"Review management commands"`
	Vitals        VitalsCmd        `cmd:"" help:"Android vitals commands"`
	Monitor       MonitorCmd       `cmd:"" help:"Monitoring and alerting commands"`
	Analytics     AnalyticsCmd     `cmd:"" help:"Analytics commands"`
	Purchases     PurchasesCmd     `cmd:"" help:"Purchase verification commands"`
	Monetization  MonetizationCmd  `cmd:"" help:"Monetization commands"`
	Permissions   PermissionsCmd   `cmd:"" help:"Permissions management"`
	Recovery      RecoveryCmd      `cmd:"" help:"App recovery commands"`
	Apps          AppsCmd          `cmd:"" help:"App discovery commands"`
	Games         GamesCmd         `cmd:"" help:"Google Play Games services"`
	Integrity     IntegrityCmd     `cmd:"" help:"Play Integrity API commands"`
	Migrate       MigrateCmd       `cmd:"" help:"Migration commands"`
	CustomApp     CustomAppCmd     `cmd:"" help:"Custom app publishing" aliases:"customapp"`
	GeneratedApks GeneratedApksCmd `cmd:"" help:"Generated APKs management"`
	SystemApks    SystemApksCmd    `cmd:"" help:"System APKs management"`
	Grouping      GroupingCmd      `cmd:"" help:"App access grouping"`
	Version       VersionCmd       `cmd:"" help:"Show version information"`
	CheckUpdate   UpdateCheckCmd   `cmd:"" name:"check-update" help:"Check for available updates"`
	Completion    CompletionCmd    `cmd:"" help:"Generate shell completion scripts"`
	Maintenance   MaintenanceCmd   `cmd:"" help:"System maintenance and monitoring commands"`

	// New advanced commands
	Bulk        BulkCmd        `cmd:"" help:"Batch operations for uploads and updates"`
	Compare     CompareCmd     `cmd:"" help:"Compare metrics across multiple apps"`
	ReleaseMgmt ReleaseMgmtCmd `cmd:"" name:"release-mgmt" help:"Advanced release management"`
	Testing     TestingCmd     `cmd:"" help:"Testing and QA tools"`
	Automation  AutomationCmd  `cmd:"" help:"CI/CD release automation"`
	Workflow    WorkflowCmd    `cmd:"" help:"Declarative workflow execution"`

	// Extension commands
	Extension ExtensionCmd `cmd:"" help:"Manage CLI extensions"`

	// Help is automatically provided by Kong
}

// Run executes the Kong CLI and returns the exit code.
func RunKongCLI() int {
	// Check if first argument is an extension to run
	// This must happen before Kong parsing
	if tryRunExtension(os.Args[1:]) {
		return 0 // Extension handled execution
	}

	var cli KongCLI

	// Set default cache directory
	if cli.CacheDir == "" {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			cli.CacheDir = filepath.Join(homeDir, ".gpd", "cache")
		}
	}

	// Initialize cache
	if cli.CacheDir != "" {
		cli.Cache = cache.New(cli.CacheDir, 24*time.Hour)
	}

	// Perform automatic update check (non-blocking)
	AutoUpdateCheck(cli.CacheDir)

	parser, err := kong.New(&cli,
		kong.Name("gpd"),
		kong.Description("Google Play Developer CLI - A fast, lightweight command-line interface for the Google Play Developer Console."),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Compact: true,
			Summary: true,
		}),
		kong.Help(buildHelpWithExtensions),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating parser: %v\n", err)
		return errors.ExitGeneralError
	}

	kongCtx, err := parser.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return errors.ExitValidationError
	}

	// Set up verbose logging if requested
	if cli.Verbose {
		logger := logging.NewLogger(os.Stderr, true)
		logger.SetLevel(logging.LevelDebug)
		logging.SetDefault(logger)
		logging.Debug("Verbose logging enabled")
	}

	// Create context with timeout from globals
	ctx, cancel := context.WithTimeout(context.Background(), cli.Timeout)
	defer cancel()
	cli.Context = ctx

	logging.Debug("Command execution started",
		logging.String("timeout", cli.Timeout.String()),
	)

	// Execute the selected command
	err = kongCtx.Run(&cli.Globals)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			return apiErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return errors.ExitGeneralError
	}

	return errors.ExitSuccess
}

// buildHelpWithExtensions creates a custom help function that includes installed extensions.
func buildHelpWithExtensions(options kong.HelpOptions, ctx *kong.Context) error {
	// First, print the standard help
	if err := kong.DefaultHelpPrinter(options, ctx); err != nil {
		return err
	}

	// Query installed extensions
	extList, err := extensions.List()
	if err != nil {
		// Silently skip if we can't list extensions
		return nil
	}

	// If no extensions installed, just return
	if len(extList) == 0 {
		return nil
	}

	// Build and print extension commands section
	fmt.Fprintln(ctx.Stdout)
	fmt.Fprintln(ctx.Stdout, "Extension Commands:")

	w := tabwriter.NewWriter(ctx.Stdout, 0, 0, 2, ' ', 0)
	for _, ext := range extList {
		desc := ext.Description
		if len(desc) > 50 {
			desc = desc[:47] + "..."
		}
		if desc == "" {
			desc = "Extension command"
		}
		fmt.Fprintf(w, "  gpd %s\t%s\n", ext.Name, desc)
	}
	w.Flush()

	return nil
}
