package cli

import (
	"fmt"

	"github.com/dl-alexandre/cli-tools/version"
)

// VersionCmd shows version information.
type VersionCmd struct{}

// Run executes the version command.
func (cmd *VersionCmd) Run(globals *Globals) error {
	info := version.GetInfo()
	fmt.Printf("gpd %s (%s) built %s\n", info.Version, info.GitCommit, info.BuildTime)
	return nil
}
