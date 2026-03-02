// Package main provides the entry point for the gpd CLI.
package main

import (
	"os"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	// Use Kong-based CLI
	return cli.RunKongCLI()
}
