package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/auth"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/storage"
)

const (
	formatJSON  = "json"
	formatTable = "table"
	formatExcel = "excel"
)

// newAuthManager creates a new auth manager instance.
func newAuthManager() *auth.Manager {
	secureStorage := storage.New()
	return auth.NewManager(secureStorage)
}

// outputResult formats and outputs a result based on the format.
func outputResult(result *output.Result, format string, pretty bool) error {
	manager := output.NewManager(os.Stdout)
	manager.SetPretty(pretty)

	// Convert string format to output.Format type
	switch strings.ToLower(format) {
	case formatTable:
		manager.SetFormat(output.FormatTable)
	case formatExcel:
		manager.SetFormat(output.FormatExcel)
	default:
		manager.SetFormat(output.FormatJSON)
	}

	return manager.Write(result)
}

// requirePackage validates that a package name is provided.
func requirePackage(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name is required")
	}
	return nil
}
