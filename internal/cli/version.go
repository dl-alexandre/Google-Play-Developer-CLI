package cli

import (
	"github.com/dl-alexandre/cli-tools/version"
)

var (
	// GitHubRepo is the GitHub repository name
	GitHubRepo = "Google-Play-Developer-CLI"
)

// VersionInfo returns the current version from cli-tools
func Version() string {
	return version.Version
}

// GitCommit returns the git commit hash from cli-tools
func GitCommit() string {
	return version.GitCommit
}

// BuildTime returns the build timestamp from cli-tools
func BuildTime() string {
	return version.BuildTime
}

// BinaryName returns the binary name
func BinaryName() string {
	return version.BinaryName
}

func init() {
	// Set CLI-specific metadata
	version.BinaryName = "gpd"
}
