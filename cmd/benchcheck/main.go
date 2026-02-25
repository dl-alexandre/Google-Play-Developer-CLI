// benchcheck is a CLI tool for comparing Go benchmark results and detecting regressions.
//
// Usage:
//
//	benchcheck --baseline baseline.txt --current current.txt [--threshold 1.20]
//	benchcheck --compare-with-main  # Compare current branch against main
//
// Exit codes:
//
//	0 - No regressions or only notices
//	1 - Critical regressions detected
//	2 - Usage error
//	3 - File read error
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/dl-alexandre/gpd/internal/apitest/benchcheck"
)

func main() {
	var (
		baselinePath = flag.String("baseline", "", "Path to baseline benchmark output file")
		currentPath  = flag.String("current", "", "Path to current benchmark output file")
		threshold    = flag.Float64("threshold", 1.20, "Regression threshold (e.g., 1.20 = 20%)")
		outputFormat = flag.String("format", "text", "Output format: text, json, github")
		failOnWarn   = flag.Bool("fail-on-warning", false, "Exit with error on warnings (not just critical)")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Compare Go benchmark results and detect performance regressions.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Compare two benchmark files\n")
		fmt.Fprintf(os.Stderr, "  %s --baseline main.bench --current pr.bench\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # Compare with stricter threshold\n")
		fmt.Fprintf(os.Stderr, "  %s --baseline main.bench --current pr.bench --threshold 1.10\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  # GitHub Actions output format\n")
		fmt.Fprintf(os.Stderr, "  %s --baseline main.bench --current pr.bench --format github\n", os.Args[0])
	}

	flag.Parse()

	if *baselinePath == "" || *currentPath == "" {
		flag.Usage()
		os.Exit(2)
	}

	// Read and parse baseline
	baselineFile, err := os.Open(*baselinePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open baseline file: %v\n", err)
		os.Exit(3)
	}

	parser := benchcheck.NewParser()
	baselineResults, err := parser.Parse(baselineFile)
	if err != nil {
		_ = baselineFile.Close() //nolint:errcheck
		fmt.Fprintf(os.Stderr, "Error: Failed to parse baseline: %v\n", err)
		os.Exit(3)
	}
	_ = baselineFile.Close() //nolint:errcheck

	// Read and parse current
	currentFile, err := os.Open(*currentPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Failed to open current file: %v\n", err)
		os.Exit(3)
	}

	currentResults, err := parser.Parse(currentFile)
	if err != nil {
		_ = currentFile.Close() //nolint:errcheck
		fmt.Fprintf(os.Stderr, "Error: Failed to parse current results: %v\n", err)
		os.Exit(3)
	}
	_ = currentFile.Close() //nolint:errcheck

	// Configure comparator
	config := &benchcheck.Config{
		NsPerOpThreshold:     *threshold,
		BytesPerOpThreshold:  *threshold,
		AllocsPerOpThreshold: *threshold,
		MinIterations:        10,
	}

	comparator := benchcheck.NewComparator(config)
	regressions := comparator.Compare(baselineResults, currentResults)

	// Output results
	switch *outputFormat {
	case "github":
		printGitHubOutput(regressions)
	case "json":
		printJSONOutput(regressions)
	default:
		printTextOutput(regressions, baselineResults, currentResults)
	}

	// Determine exit code
	if benchcheck.HasCriticalRegressions(regressions) {
		os.Exit(1)
	}

	if *failOnWarn {
		for _, r := range regressions {
			if r.Severity == benchcheck.SeverityWarning {
				os.Exit(1)
			}
		}
	}

	os.Exit(0)
}

func printTextOutput(regressions []benchcheck.Regression, baseline, current map[string]benchcheck.BenchmarkResult) {
	reporter := benchcheck.NewReporter(os.Stdout)

	fmt.Println("=== Benchmark Regression Report ===")

	fmt.Printf("Baseline: %d benchmarks\n", len(baseline))
	fmt.Printf("Current:  %d benchmarks\n", len(current))
	fmt.Println()

	reporter.PrintRegressions(regressions)

	// Summary
	fmt.Printf("Summary: %d total regressions (", len(regressions))

	critical, warnings, notices := 0, 0, 0
	for _, r := range regressions {
		switch r.Severity {
		case benchcheck.SeverityCritical:
			critical++
		case benchcheck.SeverityWarning:
			warnings++
		case benchcheck.SeverityNotice:
			notices++
		}
	}

	parts := []string{}
	if critical > 0 {
		parts = append(parts, fmt.Sprintf("%d critical", critical))
	}
	if warnings > 0 {
		parts = append(parts, fmt.Sprintf("%d warnings", warnings))
	}
	if notices > 0 {
		parts = append(parts, fmt.Sprintf("%d notices", notices))
	}
	if len(parts) == 0 {
		parts = append(parts, "0 issues")
	}

	fmt.Printf("%s)\n", strings.Join(parts, ", "))
}

func printGitHubOutput(regressions []benchcheck.Regression) {
	// GitHub Actions workflow commands for annotations
	// https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions

	for _, r := range regressions {
		var level string
		switch r.Severity {
		case benchcheck.SeverityCritical:
			level = "error"
		case benchcheck.SeverityWarning:
			level = "warning"
		default:
			level = "notice"
		}

		if r.Metric == "removed" {
			fmt.Printf("::%s::Benchmark %s was removed\n", level, r.Benchmark)
		} else {
			msg := fmt.Sprintf("Benchmark %s: %s increased by %.1f%% (%.2f → %.2f)",
				r.Benchmark, r.Metric, r.ChangePct, r.OldValue, r.NewValue)
			fmt.Printf("::%s::%s\n", level, msg)
		}
	}

	// Summary as a notice
	if len(regressions) == 0 {
		fmt.Println("::notice::✅ No performance regressions detected")
	} else {
		fmt.Printf("::notice::Found %d performance regressions\n", len(regressions))
	}
}

func printJSONOutput(regressions []benchcheck.Regression) {
	// Simple JSON output
	fmt.Println("{")
	fmt.Printf(`  "regressions": [`)

	for i, r := range regressions {
		if i > 0 {
			fmt.Println(",")
		}
		fmt.Printf(`    {"benchmark": %q, "metric": %q, "severity": %q, "change_pct": %.2f}`,
			r.Benchmark, r.Metric, r.Severity.String(), r.ChangePct)
	}

	fmt.Println("\n  ]")
	fmt.Println("}")
}
