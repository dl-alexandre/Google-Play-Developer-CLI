// Package cli provides the main CLI framework for gpd.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/auth"
	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
	"github.com/dl-alexandre/gpd/pkg/version"
)

// CLI represents the main CLI application.
type CLI struct {
	rootCmd   *cobra.Command
	config    *config.Config
	authMgr   *auth.Manager
	apiClient *api.Client
	outputMgr *output.Manager
	stdout    io.Writer
	stderr    io.Writer
	startTime time.Time

	// Global flags
	packageName  string
	outputFormat string
	pretty       bool
	timeout      time.Duration
	storeTokens  string
	fields       string
	quiet        bool
	verbose      bool
	keyPath      string
	profile      string
}

// New creates a new CLI instance.
func New() *CLI {
	cli := &CLI{
		stdout:    os.Stdout,
		stderr:    os.Stderr,
		startTime: time.Now(),
	}

	cli.outputMgr = output.NewManager(cli.stdout)

	// Initialize authentication manager with secure storage
	secureStorage := storage.New()
	cli.authMgr = auth.NewManager(secureStorage)

	cli.buildCommands()
	return cli
}

// Execute runs the CLI.
func (c *CLI) Execute() int {
	if err := c.rootCmd.Execute(); err != nil {
		return errors.ExitGeneralError
	}
	return errors.ExitSuccess
}

func (c *CLI) buildCommands() {
	c.rootCmd = &cobra.Command{
		Use:   "gpd",
		Short: "Google Play Developer CLI",
		Long: `gpd is a fast, lightweight command-line interface for the Google Play Developer Console.

It provides programmatic access to Google Play Developer Console functionality
for automating Android app publishing and management tasks.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return c.setup(cmd)
		},
	}

	// Global flags
	pf := c.rootCmd.PersistentFlags()
	pf.StringVar(&c.packageName, "package", "", "App package name")
	pf.StringVar(&c.outputFormat, "output", "json", "Output format: json, table, markdown, csv (analytics/vitals only)")
	pf.BoolVar(&c.pretty, "pretty", false, "Pretty print JSON output")
	pf.DurationVar(&c.timeout, "timeout", 30*time.Second, "Network timeout")
	pf.StringVar(&c.storeTokens, "store-tokens", "auto", "Token storage: auto, never, secure")
	pf.StringVar(&c.fields, "fields", "", "JSON field projection (comma-separated paths)")
	pf.BoolVar(&c.quiet, "quiet", false, "Suppress stderr except errors")
	pf.BoolVar(&c.verbose, "verbose", false, "Verbose output")
	pf.StringVar(&c.keyPath, "key", "", "Service account key file path")
	pf.StringVar(&c.profile, "profile", "", "Authentication profile name")

	// Add command groups
	c.addVersionCommand()
	c.addAuthCommands()
	c.addConfigCommands()
	c.addAppsCommands()
	c.addPublishCommands()
	c.addCustomAppCommands()
	c.addMigrateCommands()
	c.addReviewsCommands()
	c.addPurchasesCommands()
	c.addAnalyticsCommands()
	c.addVitalsCommands()
	c.addMonetizationCommands()
	c.addPermissionsCommands()
	c.addGamesCommands()
	c.addGroupingCommands()
	c.addIntegrityCommands()
	c.addRecoveryCommands()
	c.addHelpCommands()
}

func (c *CLI) setup(_ *cobra.Command) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		cfg = config.DefaultConfig()
	}
	c.config = cfg

	// Apply environment variable overrides
	if envPkg := config.GetEnvPackage(); envPkg != "" && c.packageName == "" {
		c.packageName = envPkg
	}
	if c.packageName == "" && c.config.DefaultPackage != "" {
		c.packageName = c.config.DefaultPackage
	}
	if envStore := config.GetEnvStoreTokens(); envStore != "" && c.storeTokens == "auto" {
		c.storeTokens = envStore
	}
	if c.storeTokens == "auto" && c.config.StoreTokens != "" {
		c.storeTokens = c.config.StoreTokens
	}
	if envProfile := config.GetEnvAuthProfile(); envProfile != "" && c.profile == "" {
		c.profile = envProfile
	}
	if c.profile == "" && c.config.ActiveProfile != "" {
		c.profile = c.config.ActiveProfile
	}
	c.authMgr.SetStoreTokens(c.storeTokens)
	c.authMgr.SetActiveProfile(c.profile)

	// Configure output manager
	c.outputMgr.SetFormat(output.ParseFormat(c.outputFormat))
	c.outputMgr.SetPretty(c.pretty)
	if c.fields != "" {
		c.outputMgr.SetFields(strings.Split(c.fields, ","))
	}

	return nil
}

// Output writes a result to stdout.
func (c *CLI) Output(r *output.Result) error {
	r.WithDuration(time.Since(c.startTime))
	return c.outputMgr.Write(r)
}

// OutputError writes an error result.
func (c *CLI) OutputError(err *errors.APIError) error {
	r := output.NewErrorResult(err)
	r.WithDuration(time.Since(c.startTime))
	return c.outputMgr.Write(r)
}

// getAPIClient lazily initializes and returns the API client.
func (c *CLI) getAPIClient(ctx context.Context) (*api.Client, error) {
	if c.apiClient != nil {
		return c.apiClient, nil
	}

	// Authenticate first
	creds, err := c.authMgr.Authenticate(ctx, c.keyPath)
	if err != nil {
		return nil, err
	}

	// Create API client
	client, err := api.NewClient(ctx, creds.TokenSource,
		api.WithTimeout(c.timeout),
	)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create API client: %v", err))
	}

	c.apiClient = client
	return client, nil
}

// requirePackage ensures a package name is provided.
func (c *CLI) requirePackage() error {
	if c.packageName == "" {
		return errors.ErrPackageRequired
	}
	return nil
}

// addVersionCommand adds the version command.
func (c *CLI) addVersionCommand() {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			info := version.Get()
			result := output.NewResult(map[string]interface{}{
				"version":   info.Version,
				"gitCommit": info.GitCommit,
				"buildTime": info.BuildTime,
				"goVersion": info.GoVersion,
				"platform":  info.Platform,
			})
			return c.Output(result.WithServices("version"))
		},
	}
	c.rootCmd.AddCommand(versionCmd)

	// Also add --version flag
	c.rootCmd.Version = version.Get().Short()
	c.rootCmd.SetVersionTemplate(`{{.Version}}
`)
}

// addHelpCommands adds help-related commands.
func (c *CLI) addHelpCommands() {
	helpCmd := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			if len(args) == 0 {
				return root.Help()
			}
			target, _, err := root.Find(args)
			if err != nil {
				return err
			}
			if target == nil {
				return root.Help()
			}
			target.InitDefaultHelpFlag()
			return target.Help()
		},
	}

	agentCmd := &cobra.Command{
		Use:   "agent",
		Short: "AI agent quickstart guide",
		Long: `AI Agent Quickstart Guide for gpd

gpd is designed for programmatic access by AI agents and automation systems.

Key Features for AI Agents:
1. Minified JSON output by default (single-line)
2. Predictable exit codes for error handling
3. Explicit flags over interactive prompts
4. No browser-based authentication

Example Workflow:
  # Check authentication
  gpd auth status

  # Upload an artifact
  gpd publish upload app.aab --package com.example.app

  # Create a release
  gpd publish release --package com.example.app --track internal --status draft

  # Check release status
  gpd publish status --package com.example.app --track internal

Exit Codes:
  0 - Success
  1 - General API error
  2 - Authentication failure
  3 - Permission denied
  4 - Validation error
  5 - Rate limited
  6 - Network error
  7 - Not found
  8 - Conflict

Output Format:
  All responses follow the envelope structure: {data, error, meta}
  Use --pretty for human-readable JSON
  Use --output table for tabular output
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(c.stdout, cmd.Long)
			return err
		},
	}

	helpCmd.AddCommand(agentCmd)
	c.rootCmd.SetHelpCommand(helpCmd)
}
