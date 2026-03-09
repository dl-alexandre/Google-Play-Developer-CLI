// apidrift is a CLI tool for detecting API drift between the Google Play Developer
// API discovery document and the Go client library.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/apidrift"
)

const (
	defaultDiscoveryURL = "https://www.googleapis.com/discovery/v1/apis/androidpublisher/v3/rest"
	defaultGoModPath    = "go.mod"
)

func main() {
	var (
		discoveryURL    = flag.String("discovery-url", defaultDiscoveryURL, "URL to the API discovery document")
		goModPath       = flag.String("go-mod", defaultGoModPath, "Path to go.mod file")
		clientSourceDir = flag.String("client-dir", "", "Path to Go client library source (if local)")
		output          = flag.String("output", "", "Output file path (JSON format)")
		format          = flag.String("format", "text", "Output format: text, json, markdown")
		threshold       = flag.Int("threshold", 0, "Fail if drift score exceeds this threshold")
		verbose         = flag.Bool("verbose", false, "Enable verbose output")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Detect API drift between Google Play Developer API discovery document\n")
		fmt.Fprintf(os.Stderr, "and the Go client library.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Basic drift detection\n")
		fmt.Fprintf(os.Stderr, "  %s\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Output JSON report\n")
		fmt.Fprintf(os.Stderr, "  %s -output report.json -format json\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Fail if more than 5 endpoints drift\n")
		fmt.Fprintf(os.Stderr, "  %s -threshold 5\n", os.Args[0])
	}

	flag.Parse()

	// Determine client source directory
	clientDir := *clientSourceDir
	if clientDir == "" {
		// Try to find it in GOPATH/pkg/mod
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			gopath = filepath.Join(os.Getenv("HOME"), "go")
		}

		// Extract version from go.mod to find the right directory
		goModContent, err := os.ReadFile(*goModPath)
		if err == nil {
			version := extractVersion(string(goModContent))
			if version != "" {
				clientDir = filepath.Join(gopath, "pkg", "mod", "google.golang.org", "api@"+version, "androidpublisher", "v3")
			}
		}

		// Fallback: use local internal/api as approximation
		if clientDir == "" || !dirExists(clientDir) {
			clientDir = "internal/api"
		}
	}

	if *verbose {
		fmt.Fprintf(os.Stderr, "Discovery URL: %s\n", *discoveryURL)
		fmt.Fprintf(os.Stderr, "Go mod path: %s\n", *goModPath)
		fmt.Fprintf(os.Stderr, "Client source: %s\n", clientDir)
		fmt.Fprintln(os.Stderr)
	}

	// Create detector
	detector := apidrift.NewDetector(*discoveryURL, *goModPath, clientDir)

	// Run detection
	report, err := detector.Detect()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Output report
	switch *format {
	case "json":
		if err := printJSONReport(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "markdown":
		printMarkdownReport(report)
	default:
		report.PrintReport()
	}

	// Save to file if requested
	if *output != "" {
		if err := report.SaveReport(*output); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving report: %v\n", err)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Report saved to: %s\n", *output)
	}

	// Check threshold
	if *threshold > 0 && report.DriftScore > *threshold {
		fmt.Fprintf(os.Stderr, "\n❌ Drift score %d exceeds threshold %d\n", report.DriftScore, *threshold)
		os.Exit(2)
	}

	// Exit with appropriate code
	if report.DriftDetected {
		os.Exit(1)
	}
	os.Exit(0)
}

func extractVersion(goModContent string) string {
	// Look for google.golang.org/api vX.Y.Z
	lines := strings.Split(goModContent, "\n")
	for _, line := range lines {
		if strings.Contains(line, "google.golang.org/api") && strings.Contains(line, "v0.") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "v0.") || strings.HasPrefix(part, "v1.") {
					return part
				}
			}
		}
	}
	return ""
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func printJSONReport(report *apidrift.DriftReport) error {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func printMarkdownReport(report *apidrift.DriftReport) {
	fmt.Println("# Google Play Developer API Drift Report")
	fmt.Printf("\n**Generated:** %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05 UTC"))

	fmt.Println("## Summary")
	fmt.Printf("\n- **Discovery Revision:** %s\n", report.DiscoveryRevision)
	fmt.Printf("- **Go Module Version:** %s\n", report.GoModVersion)
	fmt.Printf("- **Drift Score:** %d\n", report.DriftScore)
	if report.DriftDetected {
		fmt.Printf("- **Status:** ⚠️ DRIFT DETECTED\n")
	} else {
		fmt.Printf("- **Status:** ✅ No drift\n")
	}

	fmt.Println("\n## Endpoint Analysis")
	fmt.Printf("\n| Metric | Count |\n")
	fmt.Printf("|--------|-------|\n")
	fmt.Printf("| Discovery Total | %d |\n", report.Endpoints.DiscoveryTotal)
	fmt.Printf("| Implemented | %d |\n", report.Endpoints.ImplementedTotal)
	fmt.Printf("| Missing | %d |\n", len(report.Endpoints.MissingInClient))
	fmt.Printf("| Deprecated | %d |\n", len(report.Endpoints.Deprecated))

	if len(report.Endpoints.MissingInClient) > 0 {
		fmt.Println("\n### Missing Endpoints")
		for _, ep := range report.Endpoints.MissingInClient {
			fmt.Printf("- `%s`\n", ep)
		}
	}

	if len(report.Endpoints.Deprecated) > 0 {
		fmt.Println("\n### Deprecated Endpoints")
		for _, ep := range report.Endpoints.Deprecated {
			fmt.Printf("- `%s`\n", ep)
		}
	}

	fmt.Println("\n## Recommendations")
	if report.DriftDetected {
		fmt.Println("\n1. Update the Go client library to the latest version")
		fmt.Println("2. Review missing endpoints for new features to implement")
		fmt.Println("3. Remove or deprecate endpoints no longer in the API")
	} else {
		fmt.Println("\n✅ API is up to date. No action required.")
	}
}
