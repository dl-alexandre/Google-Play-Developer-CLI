// Package main provides the entry point for the gpd CLI.
package main

import (
	"os"

	"github.com/dl-alexandre/gpd/internal/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	app := cli.New()
	return app.Execute()
}
