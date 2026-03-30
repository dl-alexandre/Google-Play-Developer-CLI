package cli

import (
	"github.com/dl-alexandre/cli-tools/version"
)

// Build-time variables (re-exported from cli-tools/version for backward compatibility)
var (
	// Version is the current version of the CLI
	Version = version.Version

	// BinaryName is the name of the binary
	BinaryName = version.BinaryName

	// GitHubRepo is the GitHub repository name
	GitHubRepo = "Google-Play-Developer-CLI"

	// GitCommit is the git commit hash
	GitCommit = version.GitCommit

	// BuildTime is the build timestamp
	BuildTime = version.BuildTime
)

func init() {
	// Set CLI-specific metadata
	version.BinaryName = "gpd"
}
