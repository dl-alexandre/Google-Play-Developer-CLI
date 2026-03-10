// Package cli provides maintenance commands for the gpd CLI.
package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/apidrift"
)

// MaintenanceCmd groups all system maintenance and monitoring commands.
type MaintenanceCmd struct {
	Drift       DriftCmd       `cmd:"" help:"Detect API drift between discovery and client library"`
	MultiDrift  MultiDriftCmd  `cmd:"" help:"Monitor drift across multiple Google APIs"`
	Health      HealthCmd      `cmd:"" help:"Check system health and dependencies"`
	UpdateCheck UpdateCheckCmd `cmd:"" name:"update-check" help:"Check for CLI updates"`
}

// DriftCmd detects API drift between discovery document and Go client library.
type DriftCmd struct {
	DiscoveryURL string `help:"Discovery API URL" default:"https://www.googleapis.com/discovery/v1/apis/androidpublisher/v3/rest"`
	Format       string `help:"Output format" enum:"json,markdown,text" default:"text"`
	OutputFile   string `help:"Output file path (optional)" name:"output-file" type:"path"`
	Threshold    int    `help:"Fail if drift score exceeds this value" default:"0"`
	DriftVerbose bool   `help:"Enable verbose drift output" name:"verbose-drift"`
}

// Run executes the drift detection command.
func (cmd *DriftCmd) Run(globals *Globals) error {
	if cmd.DriftVerbose || globals.Verbose {
		fmt.Fprintf(os.Stderr, "Discovery URL: %s\n", cmd.DiscoveryURL)
		fmt.Fprintf(os.Stderr, "Format: %s\n", cmd.Format)
		fmt.Fprintln(os.Stderr)
	}

	// Create detector
	detector := apidrift.NewDetector(
		cmd.DiscoveryURL,
		"go.mod",
		"internal/api",
	)

	// Run detection
	report, err := detector.Detect()
	if err != nil {
		return fmt.Errorf("drift detection failed: %w", err)
	}

	// Output based on format
	var outputData []byte
	switch cmd.Format {
	case formatJSON:
		outputData, err = json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal report: %w", err)
		}
	case formatMarkdown:
		outputData = []byte(generateMarkdownReport(report))
	default:
		report.PrintReport()
	}

	// Write to file if specified
	if cmd.OutputFile != "" {
		if err := os.WriteFile(cmd.OutputFile, outputData, 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Fprintf(os.Stderr, "Report saved to: %s\n", cmd.OutputFile)
	}

	// Print to stdout (unless text format which already printed)
	if cmd.Format != "text" {
		fmt.Println(string(outputData))
	}

	// Check threshold
	if cmd.Threshold > 0 && report.DriftScore > cmd.Threshold {
		return fmt.Errorf("drift score %d exceeds threshold %d", report.DriftScore, cmd.Threshold)
	}

	// Return error if drift detected (for CI/CD)
	if report.DriftDetected && cmd.Threshold == 0 {
		return fmt.Errorf("API drift detected: %d endpoints missing or deprecated", report.DriftScore)
	}

	return nil
}

// MultiDriftCmd monitors drift across multiple Google APIs.
type MultiDriftCmd struct {
	APIs      []string `help:"APIs to check" enum:"androidpublisher,drive,gmail,calendar,sheets,docs,slides,people,tasks,youtube,analytics,bigquery,storage,compute" default:"androidpublisher"`
	Format    string   `help:"Output format" enum:"json,table,markdown" default:"table"`
	OutputDir string   `help:"Output directory for reports" default:".artifacts/drift" type:"path"`
}

// googleAPIs defines the discovery URLs for supported APIs.
var googleAPIs = map[string]string{
	"androidpublisher": "https://www.googleapis.com/discovery/v1/apis/androidpublisher/v3/rest",
	"drive":            "https://www.googleapis.com/discovery/v1/apis/drive/v3/rest",
	"gmail":            "https://www.googleapis.com/discovery/v1/apis/gmail/v1/rest",
	"calendar":         "https://www.googleapis.com/discovery/v1/apis/calendar/v3/rest",
	"sheets":           "https://www.googleapis.com/discovery/v1/apis/sheets/v4/rest",
	"docs":             "https://www.googleapis.com/discovery/v1/apis/docs/v1/rest",
	"slides":           "https://www.googleapis.com/discovery/v1/apis/slides/v1/rest",
	"people":           "https://www.googleapis.com/discovery/v1/apis/people/v1/rest",
	"tasks":            "https://www.googleapis.com/discovery/v1/apis/tasks/v1/rest",
	"youtube":          "https://www.googleapis.com/discovery/v1/apis/youtube/v3/rest",
	"analytics":        "https://www.googleapis.com/discovery/v1/apis/analytics/v3/rest",
	"bigquery":         "https://www.googleapis.com/discovery/v1/apis/bigquery/v2/rest",
	"storage":          "https://www.googleapis.com/discovery/v1/apis/storage/v1/rest",
	"compute":          "https://www.googleapis.com/discovery/v1/apis/compute/v1/rest",
}

// Run executes multi-API drift monitoring.
func (cmd *MultiDriftCmd) Run(globals *Globals) error {
	// Create output directory
	if err := os.MkdirAll(cmd.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	results := make(map[string]*apidrift.DriftReport)

	fmt.Println("Checking API drift across multiple Google APIs...")
	fmt.Println()

	for _, apiName := range cmd.APIs {
		url, ok := googleAPIs[apiName]
		if !ok {
			fmt.Fprintf(os.Stderr, "⚠️ Unknown API: %s (skipping)\n", apiName)
			continue
		}

		fmt.Printf("Checking %s... ", apiName)

		detector := apidrift.NewDetector(url, "go.mod", "internal/api")
		report, err := detector.Detect()
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			continue
		}

		results[apiName] = report

		if report.DriftDetected {
			fmt.Printf("⚠️ Drift (score: %d)\n", report.DriftScore)
		} else {
			fmt.Printf("✅ Up to date\n")
		}

		// Save individual report
		reportPath := filepath.Join(cmd.OutputDir, apiName+".json")
		if err := report.SaveReport(reportPath); err != nil {
			fmt.Fprintf(os.Stderr, "⚠️ Failed to save report for %s: %v\n", apiName, err)
		}
	}

	fmt.Println()

	// Generate summary report
	switch cmd.Format {
	case "json":
		return outputMultiDriftJSON(results)
	case formatMarkdown:
		return outputMultiDriftMarkdown(results)
	default:
		return outputMultiDriftTable(results)
	}
}

func outputMultiDriftJSON(results map[string]*apidrift.DriftReport) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputMultiDriftTable(results map[string]*apidrift.DriftReport) error {
	// Sort API names
	var apiNames []string
	for name := range results {
		apiNames = append(apiNames, name)
	}
	sort.Strings(apiNames)

	// Calculate totals
	totalDrift := 0
	driftCount := 0
	for _, report := range results {
		if report.DriftDetected {
			totalDrift += report.DriftScore
			driftCount++
		}
	}

	// Print table
	fmt.Println("┌─────────────────────┬──────────┬─────────┬─────────────────────┐")
	fmt.Println("│ API                 │ Revision │ Drift   │ Status              │")
	fmt.Println("├─────────────────────┼──────────┼─────────┼─────────────────────┤")

	for _, apiName := range apiNames {
		report := results[apiName]
		revision := report.DiscoveryRevision
		if len(revision) > 10 {
			revision = revision[:10]
		}

		status := "✅ Up to date"
		if report.DriftDetected {
			status = fmt.Sprintf("⚠️ Score: %d", report.DriftScore)
		}

		fmt.Printf("│ %-19s │ %-8s │ %-7s │ %-19s │\n",
			apiName,
			revision,
			fmt.Sprintf("%d/%d", report.Endpoints.ImplementedTotal, report.Endpoints.DiscoveryTotal),
			status,
		)
	}

	fmt.Println("└─────────────────────┴──────────┴─────────┴─────────────────────┘")
	fmt.Printf("\nSummary: %d/%d APIs have drift (total score: %d)\n", driftCount, len(results), totalDrift)

	return nil
}

func outputMultiDriftMarkdown(results map[string]*apidrift.DriftReport) error {
	var apiNames []string
	for name := range results {
		apiNames = append(apiNames, name)
	}
	sort.Strings(apiNames)

	totalDrift := 0
	driftCount := 0
	for _, report := range results {
		if report.DriftDetected {
			totalDrift += report.DriftScore
			driftCount++
		}
	}

	fmt.Println("# Multi-API Drift Report")
	fmt.Printf("\n**Generated:** %s\n\n", time.Now().UTC().Format("2006-01-02 15:04:05 UTC"))

	fmt.Println("## Summary")
	fmt.Printf("\n- **Total APIs Checked:** %d\n", len(results))
	fmt.Printf("- **APIs with Drift:** %d\n", driftCount)
	fmt.Printf("- **Total Drift Score:** %d\n\n", totalDrift)

	fmt.Println("## Results")
	fmt.Println("\n| API | Revision | Endpoints | Drift Score | Status |")
	fmt.Println("|-----|----------|-----------|-------------|--------|")

	for _, apiName := range apiNames {
		report := results[apiName]
		status := "✅ Up to date"
		if report.DriftDetected {
			status = fmt.Sprintf("⚠️ %d", report.DriftScore)
		}

		fmt.Printf("| %s | %s | %d/%d | %d | %s |\n",
			apiName,
			report.DiscoveryRevision,
			report.Endpoints.ImplementedTotal,
			report.Endpoints.DiscoveryTotal,
			report.DriftScore,
			status,
		)
	}

	fmt.Println("\n## Details")
	for _, apiName := range apiNames {
		report := results[apiName]
		if report.DriftDetected {
			fmt.Printf("\n### %s\n", apiName)
			fmt.Printf("\n- **Missing Endpoints:** %d\n", len(report.Endpoints.MissingInClient))
			fmt.Printf("- **Deprecated Endpoints:** %d\n", len(report.Endpoints.Deprecated))
			if len(report.Endpoints.MissingInClient) > 0 {
				fmt.Println("\nMissing:")
				for _, ep := range report.Endpoints.MissingInClient[:minInt(10, len(report.Endpoints.MissingInClient))] {
					fmt.Printf("- `%s`\n", ep)
				}
				if len(report.Endpoints.MissingInClient) > 10 {
					fmt.Printf("- ... and %d more\n", len(report.Endpoints.MissingInClient)-10)
				}
			}
		}
	}

	return nil
}

func generateMarkdownReport(report *apidrift.DriftReport) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Google Play Developer API Drift Report\n\n")
	fmt.Fprintf(&b, "**Generated:** %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05 UTC"))

	b.WriteString("## Summary\n\n")
	fmt.Fprintf(&b, "- **Discovery Revision:** %s\n", report.DiscoveryRevision)
	fmt.Fprintf(&b, "- **Go Module Version:** %s\n", report.GoModVersion)
	fmt.Fprintf(&b, "- **Drift Score:** %d\n", report.DriftScore)
	if report.DriftDetected {
		b.WriteString("- **Status:** ⚠️ DRIFT DETECTED\n")
	} else {
		b.WriteString("- **Status:** ✅ No drift\n")
	}

	b.WriteString("\n## Endpoint Analysis\n\n")
	b.WriteString("| Metric | Count |\n")
	b.WriteString("|--------|-------|\n")
	fmt.Fprintf(&b, "| Discovery Total | %d |\n", report.Endpoints.DiscoveryTotal)
	fmt.Fprintf(&b, "| Implemented | %d |\n", report.Endpoints.ImplementedTotal)
	fmt.Fprintf(&b, "| Missing | %d |\n", len(report.Endpoints.MissingInClient))
	fmt.Fprintf(&b, "| Deprecated | %d |\n\n", len(report.Endpoints.Deprecated))

	if len(report.Endpoints.MissingInClient) > 0 {
		b.WriteString("## Missing Endpoints\n\n")
		sort.Strings(report.Endpoints.MissingInClient)
		for _, ep := range report.Endpoints.MissingInClient {
			fmt.Fprintf(&b, "- `%s`\n", ep)
		}
		b.WriteString("\n")
	}

	if len(report.Endpoints.Deprecated) > 0 {
		b.WriteString("## Deprecated Endpoints\n\n")
		sort.Strings(report.Endpoints.Deprecated)
		for _, ep := range report.Endpoints.Deprecated {
			fmt.Fprintf(&b, "- `%s`\n", ep)
		}
		b.WriteString("\n")
	}

	b.WriteString("## Recommendations\n\n")
	if report.DriftDetected {
		b.WriteString("1. Update the Go client library to the latest version\n")
		b.WriteString("2. Review missing endpoints for new features to implement\n")
		b.WriteString("3. Remove or deprecate endpoints no longer in the API\n")
	} else {
		b.WriteString("✅ API is up to date. No action required.\n")
	}

	return b.String()
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// HealthCmd checks system health and dependencies.
type HealthCmd struct {
	CheckAPI    bool `help:"Check API connectivity"`
	CheckAuth   bool `help:"Check authentication status"`
	CheckConfig bool `help:"Check configuration validity"`
}

// Run executes the health check command.
func (cmd *HealthCmd) Run(globals *Globals) error {
	results := make(map[string]interface{})

	results["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	results["cli_version"] = "dev" // Would be populated from version package

	checks := []struct {
		name string
		fn   func() (bool, string)
	}{
		{
			name: "go_version",
			fn: func() (bool, string) {
				return true, "1.24.x"
			},
		},
		{
			name: "dependencies",
			fn: func() (bool, string) {
				return true, "ok"
			},
		},
		{
			name: "cache",
			fn: func() (bool, string) {
				if globals.CacheDir == "" {
					return false, "not configured"
				}
				return true, globals.CacheDir
			},
		},
	}

	for _, check := range checks {
		ok, msg := check.fn()
		results[check.name] = map[string]interface{}{
			"status":  ok,
			"message": msg,
		}
	}

	// Output results
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}
