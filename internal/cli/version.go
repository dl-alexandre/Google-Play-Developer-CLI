package cli

import "github.com/dl-alexandre/cli-tools/version"

// Build-time variables (re-exported from cli-tools/version for compatibility).
var (
	Version    = version.Version
	BinaryName = version.BinaryName
	GitHubRepo = "dl-alexandre/Google-Play-Developer-CLI"
	GitCommit  = version.GitCommit
	BuildTime  = version.BuildTime
)

func init() {
	version.BinaryName = "gpd"
	BinaryName = version.BinaryName
}
