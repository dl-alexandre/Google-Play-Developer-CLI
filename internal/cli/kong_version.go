package cli

import (
	"fmt"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/pkg/version"
)

// VersionCmd shows version information.
type VersionCmd struct{}

// Run executes the version command.
func (cmd *VersionCmd) Run(globals *Globals) error {
	info := version.Get()
	fmt.Println(info.String())
	return nil
}
