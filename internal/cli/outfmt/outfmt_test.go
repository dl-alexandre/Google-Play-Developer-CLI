//go:build unit
// +build unit

package outfmt

import (
	"testing"
)

func TestResolveDefaultOutput_PriorityChain(t *testing.T) {
	tests := []struct {
		name     string
		explicit string
		userSet  bool
		isTTY    bool
		env      string
		want     string
	}{
		{
			name:     "explicit json wins over tty table",
			explicit: "json",
			userSet:  true,
			isTTY:    true,
			env:      "table",
			want:     "json",
		},
		{
			name:     "explicit table wins over non-tty json",
			explicit: "table",
			userSet:  true,
			isTTY:    false,
			env:      "json",
			want:     "table",
		},
		{
			name:     "explicit markdown wins over env",
			explicit: "markdown",
			userSet:  true,
			isTTY:    false,
			env:      "csv",
			want:     "markdown",
		},
		{
			name:     "env wins when flag not set",
			explicit: "json", // kong structural default, not user-set
			userSet:  false,
			isTTY:    true,
			env:      "markdown",
			want:     "markdown",
		},
		{
			name:     "env csv when non-tty",
			explicit: "",
			userSet:  false,
			isTTY:    false,
			env:      "csv",
			want:     "csv",
		},
		{
			name:     "tty defaults to table",
			explicit: "json",
			userSet:  false,
			isTTY:    true,
			env:      "",
			want:     "table",
		},
		{
			name:     "non-tty defaults to json",
			explicit: "json",
			userSet:  false,
			isTTY:    false,
			env:      "",
			want:     "json",
		},
		{
			name:     "invalid env falls through to tty table",
			explicit: "",
			userSet:  false,
			isTTY:    true,
			env:      "yaml",
			want:     "table",
		},
		{
			name:     "invalid env falls through to non-tty json",
			explicit: "",
			userSet:  false,
			isTTY:    false,
			env:      "not-a-format",
			want:     "json",
		},
		{
			name:     "explicit empty with userSet falls through to env",
			explicit: "",
			userSet:  true,
			isTTY:    false,
			env:      "excel",
			want:     "excel",
		},
		{
			name:     "case insensitive explicit",
			explicit: "TABLE",
			userSet:  true,
			isTTY:    false,
			env:      "",
			want:     "table",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveDefaultOutput(tt.explicit, tt.userSet, tt.isTTY, tt.env)
			if got != tt.want {
				t.Fatalf("ResolveDefaultOutput(%q, userSet=%v, tty=%v, env=%q) = %q, want %q",
					tt.explicit, tt.userSet, tt.isTTY, tt.env, got, tt.want)
			}
		})
	}
}

func TestOutputFlagSet(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want bool
	}{
		{"empty", nil, false},
		{"no output", []string{"auth", "status"}, false},
		{"long flag", []string{"auth", "status", "--output", "json"}, true},
		{"equals form", []string{"--output=table", "auth"}, true},
		{"similar but not", []string{"--output-dir", "x"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OutputFlagSet(tt.args); got != tt.want {
				t.Fatalf("OutputFlagSet(%v) = %v, want %v", tt.args, got, tt.want)
			}
		})
	}
}

func TestResolveOutputFromEnvAndTTY_UsesEnv(t *testing.T) {
	// Force non-TTY path for determinism, then set env.
	origTTY := stdoutIsTerminal
	stdoutIsTerminal = func() bool { return false }
	t.Cleanup(func() { stdoutIsTerminal = origTTY })

	t.Setenv(EnvDefaultOutput, "markdown")
	got := ResolveOutputFromEnvAndTTY("json", []string{"version"})
	if got != "markdown" {
		t.Fatalf("got %q, want markdown from env", got)
	}
}

func TestResolveOutputFromEnvAndTTY_ExplicitWins(t *testing.T) {
	origTTY := stdoutIsTerminal
	stdoutIsTerminal = func() bool { return true }
	t.Cleanup(func() { stdoutIsTerminal = origTTY })

	t.Setenv(EnvDefaultOutput, "table")
	got := ResolveOutputFromEnvAndTTY("json", []string{"--output", "json", "version"})
	if got != "json" {
		t.Fatalf("got %q, want json from explicit flag", got)
	}
}

func TestResolveOutputFromEnvAndTTY_TTYDefault(t *testing.T) {
	origTTY := stdoutIsTerminal
	stdoutIsTerminal = func() bool { return true }
	t.Cleanup(func() { stdoutIsTerminal = origTTY })

	t.Setenv(EnvDefaultOutput, "")
	got := ResolveOutputFromEnvAndTTY("json", []string{"version"})
	if got != "table" {
		t.Fatalf("got %q, want table for TTY", got)
	}
}

func TestResolveOutputFromEnvAndTTY_NonTTYDefault(t *testing.T) {
	origTTY := stdoutIsTerminal
	stdoutIsTerminal = func() bool { return false }
	t.Cleanup(func() { stdoutIsTerminal = origTTY })

	t.Setenv(EnvDefaultOutput, "")
	got := ResolveOutputFromEnvAndTTY("json", []string{"version"})
	if got != "json" {
		t.Fatalf("got %q, want json for non-TTY", got)
	}
}
