// Package benchcheck provides tools for comparing Go benchmark results
// and detecting performance regressions.
package benchcheck

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
)

// BenchmarkResult represents a single benchmark run
type BenchmarkResult struct {
	Name        string
	N           int     // Number of iterations
	NsPerOp     float64 // Nanoseconds per operation
	BytesPerOp  int64   // Bytes allocated per operation
	AllocsPerOp int64   // Allocations per operation
}

// Regression represents a detected performance regression
type Regression struct {
	Benchmark string
	Metric    string
	OldValue  float64
	NewValue  float64
	ChangePct float64
	Severity  Severity
}

// Severity levels for regressions
type Severity int

const (
	SeverityNotice   Severity = iota // < 10% change
	SeverityWarning                  // 10-20% change
	SeverityCritical                 // > 20% change
)

func (s Severity) String() string {
	switch s {
	case SeverityNotice:
		return "NOTICE"
	case SeverityWarning:
		return "WARNING"
	case SeverityCritical:
		return "CRITICAL"
	default:
		return "UNKNOWN"
	}
}

// Config for regression detection
type Config struct {
	NsPerOpThreshold     float64 // Threshold for ns/op regressions (default 1.20 = 20%)
	BytesPerOpThreshold  float64 // Threshold for B/op regressions
	AllocsPerOpThreshold float64 // Threshold for allocs/op regressions
	MinIterations        int     // Minimum iterations for valid comparison
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		NsPerOpThreshold:     1.20,
		BytesPerOpThreshold:  1.20,
		AllocsPerOpThreshold: 1.20,
		MinIterations:        10,
	}
}

// Parser handles parsing benchmark output
type Parser struct {
	lineRegex *regexp.Regexp
}

// NewParser creates a new benchmark parser
func NewParser() *Parser {
	// Matches: BenchmarkName-CPU  1000000  1234 ns/op  456 B/op  7 allocs/op
	// Or: BenchmarkName  1000000  1234 ns/op
	lineRegex := regexp.MustCompile(`^(Benchmark\S+)\s+(\d+)\s+([\d.]+)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`)

	return &Parser{lineRegex: lineRegex}
}

// Parse reads benchmark output and extracts results
func (p *Parser) Parse(r io.Reader) (map[string]BenchmarkResult, error) {
	results := make(map[string]BenchmarkResult)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()
		matches := p.lineRegex.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		n, _ := strconv.Atoi(matches[2])
		nsPerOp, _ := strconv.ParseFloat(matches[3], 64)

		result := BenchmarkResult{
			Name:    matches[1],
			N:       n,
			NsPerOp: nsPerOp,
		}

		if matches[4] != "" {
			result.BytesPerOp, _ = strconv.ParseInt(matches[4], 10, 64)
		}

		if matches[5] != "" {
			result.AllocsPerOp, _ = strconv.ParseInt(matches[5], 10, 64)
		}

		results[result.Name] = result
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning benchmark output: %w", err)
	}

	return results, nil
}

// Comparator compares benchmark results between two runs
type Comparator struct {
	config *Config
}

// NewComparator creates a new benchmark comparator
func NewComparator(config *Config) *Comparator {
	if config == nil {
		config = DefaultConfig()
	}
	return &Comparator{config: config}
}

// Compare checks for regressions between baseline and current results
func (c *Comparator) Compare(baseline, current map[string]BenchmarkResult) []Regression {
	var regressions []Regression

	for name, currentResult := range current {
		baselineResult, ok := baseline[name]
		if !ok {
			// New benchmark, not a regression
			continue
		}

		// Skip if not enough iterations
		if baselineResult.N < c.config.MinIterations || currentResult.N < c.config.MinIterations {
			continue
		}

		// Check ns/op regression (higher is worse)
		// Detect any change > 10%, then categorize severity
		if baselineResult.NsPerOp > 0 && currentResult.NsPerOp > 0 {
			ratio := currentResult.NsPerOp / baselineResult.NsPerOp
			if ratio > 1.10 { // 10% change detection threshold
				severity := c.severityFromRatio(ratio, c.config.NsPerOpThreshold)
				regressions = append(regressions, Regression{
					Benchmark: name,
					Metric:    "ns/op",
					OldValue:  baselineResult.NsPerOp,
					NewValue:  currentResult.NsPerOp,
					ChangePct: (ratio - 1.0) * 100,
					Severity:  severity,
				})
			}
		}

		// Check B/op regression
		if baselineResult.BytesPerOp > 0 && currentResult.BytesPerOp > 0 {
			ratio := float64(currentResult.BytesPerOp) / float64(baselineResult.BytesPerOp)
			if ratio > 1.10 {
				severity := c.severityFromRatio(ratio, c.config.BytesPerOpThreshold)
				regressions = append(regressions, Regression{
					Benchmark: name,
					Metric:    "B/op",
					OldValue:  float64(baselineResult.BytesPerOp),
					NewValue:  float64(currentResult.BytesPerOp),
					ChangePct: (ratio - 1.0) * 100,
					Severity:  severity,
				})
			}
		}

		// Check allocs/op regression
		if baselineResult.AllocsPerOp > 0 && currentResult.AllocsPerOp > 0 {
			ratio := float64(currentResult.AllocsPerOp) / float64(baselineResult.AllocsPerOp)
			if ratio > 1.10 {
				severity := c.severityFromRatio(ratio, c.config.AllocsPerOpThreshold)
				regressions = append(regressions, Regression{
					Benchmark: name,
					Metric:    "allocs/op",
					OldValue:  float64(baselineResult.AllocsPerOp),
					NewValue:  float64(currentResult.AllocsPerOp),
					ChangePct: (ratio - 1.0) * 100,
					Severity:  severity,
				})
			}
		}
	}

	// Check for removed benchmarks
	for name := range baseline {
		if _, ok := current[name]; !ok {
			regressions = append(regressions, Regression{
				Benchmark: name,
				Metric:    "removed",
				Severity:  SeverityWarning,
			})
		}
	}

	return regressions
}

func (c *Comparator) severityFromRatio(ratio, threshold float64) Severity {
	// Critical if exceeds the configured threshold (default 1.20 = 20%)
	if ratio > threshold {
		return SeverityCritical
	} else if ratio > 1.10 { // Warning at 10%
		return SeverityWarning
	}
	// Notice for 10% or less (but we shouldn't reach here as 10% is our detection floor)
	return SeverityNotice
}

// Reporter generates human-readable reports
type Reporter struct {
	w io.Writer
}

// NewReporter creates a new report generator
func NewReporter(w io.Writer) *Reporter {
	return &Reporter{w: w}
}

// PrintResults outputs benchmark results in a formatted table
func (r *Reporter) PrintResults(results map[string]BenchmarkResult) {
	_, _ = fmt.Fprintln(r.w, "Benchmark Results:")
	_, _ = fmt.Fprintln(r.w, strings.Repeat("-", 80))
	_, _ = fmt.Fprintf(r.w, "%-50s %12s %12s %12s\n", "Benchmark", "N", "ns/op", "B/op")
	_, _ = fmt.Fprintln(r.w, strings.Repeat("-", 80))

	for name, result := range results {
		_, _ = fmt.Fprintf(r.w, "%-50s %12d %12.2f", name, result.N, result.NsPerOp)
		if result.BytesPerOp > 0 {
			_, _ = fmt.Fprintf(r.w, " %12d", result.BytesPerOp)
		}
		_, _ = fmt.Fprintln(r.w)
	}
	_, _ = fmt.Fprintln(r.w)
}

// PrintRegressions outputs regression report
func (r *Reporter) PrintRegressions(regressions []Regression) {
	if len(regressions) == 0 {
		_, _ = fmt.Fprintln(r.w, "✓ No performance regressions detected")
		return
	}

	_, _ = fmt.Fprintln(r.w, "\n⚠ Performance Regressions Detected:")
	_, _ = fmt.Fprintln(r.w, strings.Repeat("=", 100))

	// Group by severity
	critical := filterBySeverity(regressions, SeverityCritical)
	warnings := filterBySeverity(regressions, SeverityWarning)
	notices := filterBySeverity(regressions, SeverityNotice)

	if len(critical) > 0 {
		_, _ = fmt.Fprintf(r.w, "\n🔴 CRITICAL (%d):\n", len(critical))
		for _, reg := range critical {
			r.printRegression(reg)
		}
	}

	if len(warnings) > 0 {
		_, _ = fmt.Fprintf(r.w, "\n🟡 WARNING (%d):\n", len(warnings))
		for _, reg := range warnings {
			r.printRegression(reg)
		}
	}

	if len(notices) > 0 {
		_, _ = fmt.Fprintf(r.w, "\n📝 NOTICE (%d):\n", len(notices))
		for _, reg := range notices {
			r.printRegression(reg)
		}
	}

	_, _ = fmt.Fprintln(r.w)
}

func (r *Reporter) printRegression(reg Regression) {
	if reg.Metric == "removed" {
		_, _ = fmt.Fprintf(r.w, "  - %s: benchmark removed\n", reg.Benchmark)
		return
	}

	_, _ = fmt.Fprintf(r.w, "  - %s\n", reg.Benchmark)
	_, _ = fmt.Fprintf(r.w, "    Metric: %s | Change: +%.1f%% | Old: %.2f → New: %.2f\n",
		reg.Metric, reg.ChangePct, reg.OldValue, reg.NewValue)
}

// HasCriticalRegressions returns true if any critical regressions exist
func HasCriticalRegressions(regressions []Regression) bool {
	for _, r := range regressions {
		if r.Severity == SeverityCritical {
			return true
		}
	}
	return false
}

func filterBySeverity(regressions []Regression, severity Severity) []Regression {
	var filtered []Regression
	for _, r := range regressions {
		if r.Severity == severity {
			filtered = append(filtered, r)
		}
	}
	return filtered
}

// StatisticalAnalysis provides Welch's t-test for benchmark significance
type StatisticalAnalysis struct {
	Benchmark    string
	Metric       string
	BaselineMean float64
	BaselineStd  float64
	CurrentMean  float64
	CurrentStd   float64
	TStat        float64
	PValue       float64
	Significant  bool // true if p < 0.05
}

// WelchTTest performs Welch's t-test for two samples with potentially different variances
// Returns t-statistic and degrees of freedom
func WelchTTest(baseline, current []float64) (tStat float64, df float64) {
	if len(baseline) < 2 || len(current) < 2 {
		return 0, 0
	}

	// Calculate means
	baselineMean := mean(baseline)
	currentMean := mean(current)

	// Calculate variances
	baselineVar := variance(baseline, baselineMean)
	currentVar := variance(current, currentMean)

	// Welch's t-test formula
	diff := baselineMean - currentMean
	se := baselineVar/float64(len(baseline)) + currentVar/float64(len(current))

	if se == 0 {
		return 0, 0
	}

	tStat = diff / sqrt(se)

	// Welch-Satterthwaite degrees of freedom
	n1, n2 := float64(len(baseline)), float64(len(current))
	numerator := (baselineVar/n1 + currentVar/n2) * (baselineVar/n1 + currentVar/n2)
	denominator := (baselineVar/n1)*(baselineVar/n1)/(n1-1) + (currentVar/n2)*(currentVar/n2)/(n2-1)

	if denominator == 0 {
		return tStat, 0
	}

	df = numerator / denominator

	return tStat, df
}

// TwoTailedPValue approximates the two-tailed p-value from t-statistic and df
// Using a simple approximation (not exact, but good enough for benchmark comparison)
func TwoTailedPValue(tStat, df float64) float64 {
	if df <= 0 {
		return 1.0
	}

	// For large df, approximate with normal distribution
	if df > 30 {
		// Two-tailed test: P(|Z| > |t|) = 2 * (1 - Φ(|t|))
		// Using approximation: 1 - Φ(z) ≈ 0.5 * exp(-0.717*z - 0.416*z*z)
		z := abs(tStat)
		if z > 3.5 {
			return 0.0005 // Very significant
		}
		p := 0.5 * exp(-0.717*z-0.416*z*z)
		return 2 * p
	}

	// For small df, use rough t-distribution approximation
	// This is a simplified version - for production, use a proper implementation
	criticalValues := map[float64]map[float64]float64{
		5:  {1.0: 0.37, 2.0: 0.10, 3.0: 0.03},
		10: {1.0: 0.34, 2.0: 0.07, 3.0: 0.01},
		20: {1.0: 0.33, 2.0: 0.06, 3.0: 0.005},
	}

	// Find closest df
	closestDF := 5.0
	if df > 15 {
		closestDF = 20
	} else if df > 7 {
		closestDF = 10
	}

	absT := abs(tStat)
	for t, p := range criticalValues[closestDF] {
		if absT <= t {
			return p
		}
	}

	return 0.001 // Very significant
}

// IsSignificant returns true if p < alpha (typically 0.05)
func IsSignificant(pValue, alpha float64) bool {
	return pValue < alpha
}

// Helper functions
func mean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func variance(data []float64, mean float64) float64 {
	if len(data) < 2 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		d := v - mean
		sum += d * d
	}
	return sum / float64(len(data)-1) // Sample variance
}

func sqrt(x float64) float64 {
	if x <= 0 {
		return 0
	}
	// Simple Newton-Raphson square root
	z := x
	for i := 0; i < 10; i++ {
		z = (z + x/z) / 2
	}
	return z
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func exp(x float64) float64 {
	// Simple Taylor series approximation for e^x
	if x == 0 {
		return 1.0
	}
	if x < -10 {
		return 0
	}
	result := 1.0
	term := 1.0
	for i := 1; i < 20; i++ {
		term *= x / float64(i)
		result += term
		if term < 1e-15 {
			break
		}
	}
	return result
}
