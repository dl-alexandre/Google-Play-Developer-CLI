package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/dl-alexandre/gpd/internal/auth"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
)

// newAuthManager creates a new auth manager instance.
func newAuthManager() *auth.Manager {
	secureStorage := storage.New()
	return auth.NewManager(secureStorage)
}

// outputResult formats and outputs a result based on the format.
func outputResult(result *output.Result, format string, pretty bool) error {
	switch format {
	case "json":
		return outputJSON(result, pretty)
	case "table":
		return outputTable(result)
	default:
		return outputJSON(result, pretty)
	}
}

// outputJSON outputs result as JSON.
func outputJSON(result *output.Result, pretty bool) error {
	var data []byte
	var err error

	if pretty {
		data, err = json.MarshalIndent(result, "", "  ")
	} else {
		data, err = json.Marshal(result)
	}

	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

// outputTable outputs result as a table.
func outputTable(result *output.Result) error {
	// Extract data from result
	data, ok := result.Data.(map[string]interface{})
	if !ok {
		// Fall back to JSON if we can't format as table
		return outputJSON(result, false)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"Key", "Value"})

	for key, value := range data {
		_ = table.Append([]string{key, fmt.Sprintf("%v", value)})
	}

	_ = table.Render()
	return nil
}

// requirePackage validates that a package name is provided.
func requirePackage(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name is required")
	}
	return nil
}
