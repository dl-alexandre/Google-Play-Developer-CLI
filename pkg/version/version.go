// Package version provides build version information for gpd.
package version

import (
	"fmt"
	"runtime"

	"github.com/dl-alexandre/cli-tools/version"
)

var (
	Version   = version.Version
	GitCommit = version.GitCommit
	BuildTime = version.BuildTime
)

type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

func Get() *Info {
	return &Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

func (i *Info) String() string {
	return fmt.Sprintf("gpd %s (%s) built %s", i.Version, i.GitCommit, i.BuildTime)
}

func (i *Info) Short() string {
	return i.Version
}

func init() {
	version.BinaryName = "gpd"
}
