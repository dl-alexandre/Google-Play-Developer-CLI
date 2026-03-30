package cli

import (
	"github.com/dl-alexandre/cli-tools/update"
	"github.com/dl-alexandre/cli-tools/version"
)

// UpdateCheckCmd wraps cli-tools update functionality
type UpdateCheckCmd struct {
	Force  bool   `help:"Force check, bypassing cache" flag:"force"`
	Format string `help:"Output format" enum:"json,table" default:"table"`
}

// Run executes the update check
func (c *UpdateCheckCmd) Run(globals *Globals) error {
	checker := update.New(update.Config{
		CurrentVersion: version.Version,
		BinaryName:     version.BinaryName,
		GitHubRepo:     "dl-alexandre/Google-Play-Developer-CLI",
		InstallCommand: "brew upgrade gpd",
	})

	info, err := checker.Check(c.Force)
	if err != nil {
		return err
	}

	return update.DisplayUpdate(info, version.BinaryName, c.Format)
}

// AutoUpdateCheck performs a background update check (for use at startup)
// It returns immediately and doesn't block
func AutoUpdateCheck(cacheDir string) {
	checker := update.New(update.Config{
		CurrentVersion: version.Version,
		BinaryName:     version.BinaryName,
		GitHubRepo:     "dl-alexandre/Google-Play-Developer-CLI",
		InstallCommand: "brew upgrade gpd",
	})
	checker.AutoCheck()
}

// UpdateInfo is re-exported from cli-tools for backward compatibility
type UpdateInfo = update.Info
