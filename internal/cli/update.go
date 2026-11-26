package cli

import (
	"github.com/dl-alexandre/cli-tools/update"
	"github.com/dl-alexandre/cli-tools/version"
)

// UpdateCheckCmd wraps the shared cli-tools update checker.
type UpdateCheckCmd struct {
	Force  bool   `help:"Force check, bypassing cache" flag:"force"`
	Format string `help:"Output format" enum:"json,table" default:"table"`
}

// UpdateInfo is re-exported for compatibility with callers of the CLI package.
type UpdateInfo = update.Info

func updateChecker(cacheDir string) *update.Checker {
	return update.New(update.Config{
		CurrentVersion: version.Version,
		BinaryName:     version.BinaryName,
		GitHubRepo:     GitHubRepo,
		CacheDir:       cacheDir,
		InstallCommand: "brew upgrade gpd",
	})
}

// Run executes the update check.
func (c *UpdateCheckCmd) Run(globals *Globals) error {
	info, err := updateChecker(globals.CacheDir).Check(c.Force)
	if err != nil {
		return err
	}
	return update.DisplayUpdate(info, version.BinaryName, c.Format)
}

// AutoUpdateCheck starts the shared non-blocking update check.
func AutoUpdateCheck(cacheDir string) {
	updateChecker(cacheDir).AutoCheck()
}
