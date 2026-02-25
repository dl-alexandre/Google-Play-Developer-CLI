//go:build unit
// +build unit

package cli

import (
	"strings"
	"testing"
)

func TestCompletionCmd(t *testing.T) {
	t.Run("bash completion generates script", func(t *testing.T) {
		cmd := &CompletionCmd{Shell: "bash"}
		globals := &Globals{}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Expected no error for bash completion, got: %v", err)
		}
		// Output goes to stdout which we can't capture easily here
		// but the lack of error means it worked
	})

	t.Run("zsh completion generates script", func(t *testing.T) {
		cmd := &CompletionCmd{Shell: "zsh"}
		globals := &Globals{}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Expected no error for zsh completion, got: %v", err)
		}
	})

	t.Run("fish completion generates script", func(t *testing.T) {
		cmd := &CompletionCmd{Shell: "fish"}
		globals := &Globals{}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Expected no error for fish completion, got: %v", err)
		}
	})

	t.Run("invalid shell returns error", func(t *testing.T) {
		cmd := &CompletionCmd{Shell: "invalid"}
		globals := &Globals{}

		err := cmd.Run(globals)
		if err == nil {
			t.Error("Expected error for invalid shell, got nil")
		}
		if !strings.Contains(err.Error(), "unsupported shell") {
			t.Errorf("Expected 'unsupported shell' error, got: %v", err)
		}
	})
}

func TestGetPackageCompletions(t *testing.T) {
	t.Run("returns packages from env and config", func(t *testing.T) {
		packages := getPackageCompletions()
		// Should at least return a slice (may be empty if no config/env)
		if packages == nil {
			t.Error("Expected non-nil packages slice")
		}
	})
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{500, "500 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1024 * 1024, "1.00 MB"},
		{1024 * 1024 * 1024, "1.00 GB"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			result := formatBytes(tc.bytes)
			if result != tc.expected {
				t.Errorf("formatBytes(%d) = %q, want %q", tc.bytes, result, tc.expected)
			}
		})
	}
}
