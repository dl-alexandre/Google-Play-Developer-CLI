// Package version provides build version information for gpd.
package version

import (
	"fmt"
	"runtime"
)

// Build information set by ldflags during compilation.
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// Info contains version information.
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
	GoVersion string `json:"goVersion"`
	Platform  string `json:"platform"`
}

// Get returns the current version information.
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a human-readable version string.
func (i Info) String() string {
	return fmt.Sprintf("gpd %s (%s) built %s", i.Version, i.GitCommit, i.BuildTime)
}

// Short returns a short version string.
func (i Info) Short() string {
	return i.Version
}
