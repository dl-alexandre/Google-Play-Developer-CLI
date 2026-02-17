package version

import "testing"

func TestGetAndString(t *testing.T) {
	origVersion := Version
	origCommit := GitCommit
	origBuild := BuildTime
	defer func() {
		Version = origVersion
		GitCommit = origCommit
		BuildTime = origBuild
	}()

	Version = "1.2.3"
	GitCommit = "abc123"
	BuildTime = "now"

	info := Get()
	if info.Version != "1.2.3" || info.GitCommit != "abc123" || info.BuildTime != "now" {
		t.Fatalf("unexpected info: %+v", info)
	}
	if info.Short() != "1.2.3" {
		t.Fatalf("unexpected short: %s", info.Short())
	}
	if info.String() == "" {
		t.Fatalf("expected non-empty string")
	}
}
