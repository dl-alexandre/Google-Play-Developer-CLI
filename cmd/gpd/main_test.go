package main

import (
	"os"
	"testing"
)

func TestRunVersion(t *testing.T) {
	origArgs := os.Args
	t.Cleanup(func() { os.Args = origArgs })

	os.Args = []string{"gpd", "version"}
	if code := run(); code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
}
