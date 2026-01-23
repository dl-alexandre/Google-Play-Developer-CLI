// Package main provides the entry point for the gpd CLI.
package main

import (
	"os"

	"github.com/dl-alexandre/gpd/internal/cli"
)

func main() {
	app := cli.New()
	os.Exit(app.Execute())
}
