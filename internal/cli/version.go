package cli

import (
	"github.com/dl-alexandre/Google-Play-Developer-CLI/pkg/version"
)

// Build-time variables (set by GoReleaser or build flags)
// These are initialized from the pkg/version package
var (
	// Version is the current version of the CLI
	Version = version.Version

	// BinaryName is the name of the binary
	BinaryName = "gpd"

	// GitHubRepo is the GitHub repository name
	GitHubRepo = "Google-Play-Developer-CLI"

	// GitCommit is the git commit hash
	GitCommit = version.GitCommit

	// BuildTime is the build timestamp
	BuildTime = version.BuildTime
)
