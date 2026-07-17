// Package outfmt resolves the CLI's default --output format from flags, env, and TTY.
//
// This is a small domain package under internal/cli/ illustrating the preferred
// layout for reusable CLI helpers. New command families should live in
// internal/cli/<domain>/ with registration remaining in kong_cli.go.
package outfmt

import (
	"os"
	"strings"

	"golang.org/x/term"
)

// EnvDefaultOutput is the environment variable that overrides auto output format.
const EnvDefaultOutput = "GPD_DEFAULT_OUTPUT"

// validOutputFormats are accepted --output / GPD_DEFAULT_OUTPUT values.
var validOutputFormats = map[string]bool{
	"json":     true,
	"table":    true,
	"markdown": true,
	"csv":      true,
	"excel":    true,
}

// stdoutIsTerminal reports whether fd 1 is a terminal. Overridable in tests.
var stdoutIsTerminal = func() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// ResolveDefaultOutput chooses the effective output format.
//
// Priority:
//  1. explicitFlag — when the user passed --output on the command line
//  2. envValue (GPD_DEFAULT_OUTPUT) when set to a valid format
//  3. "table" when stdout is a TTY
//  4. "json" otherwise (pipes, files, CI)
//
// explicitFlag is only honored when userSet is true so Kong's structural
// default does not defeat TTY/env resolution.
func ResolveDefaultOutput(explicitFlag string, userSet bool, isTTY bool, envValue string) string {
	if userSet {
		normalized := normalizeOutputFormat(explicitFlag)
		if normalized != "" {
			return normalized
		}
	}

	if env := normalizeOutputFormat(envValue); env != "" {
		return env
	}

	if isTTY {
		return "table"
	}
	return "json"
}

// OutputFlagSet reports whether args include an explicit --output flag.
func OutputFlagSet(args []string) bool {
	for _, a := range args {
		if a == "--output" {
			return true
		}
		if strings.HasPrefix(a, "--output=") {
			return true
		}
	}
	return false
}

// ResolveOutputFromEnvAndTTY is the production helper used by RunKongCLI.
func ResolveOutputFromEnvAndTTY(current string, args []string) string {
	return ResolveDefaultOutput(
		current,
		OutputFlagSet(args),
		stdoutIsTerminal(),
		os.Getenv(EnvDefaultOutput),
	)
}

func normalizeOutputFormat(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	if validOutputFormats[v] {
		return v
	}
	return ""
}
