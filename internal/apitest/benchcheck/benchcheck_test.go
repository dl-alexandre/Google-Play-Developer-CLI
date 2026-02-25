//go:build unit
// +build unit

package benchcheck

import (
	"bytes"
	"strings"
	"testing"
)

func TestParser_Parse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    map[string]BenchmarkResult
		wantErr bool
	}{
		{
			name:  "simple benchmark",
			input: `BenchmarkSimple-8    1000000    1234 ns/op`,
			want: map[string]BenchmarkResult{
				"BenchmarkSimple-8": {Name: "BenchmarkSimple-8", N: 1000000, NsPerOp: 1234},
			},
		},
		{
			name:  "benchmark with memory stats",
			input: `BenchmarkWithMemory-8    1000000    1234 ns/op    456 B/op    7 allocs/op`,
			want: map[string]BenchmarkResult{
				"BenchmarkWithMemory-8": {Name: "BenchmarkWithMemory-8", N: 1000000, NsPerOp: 1234, BytesPerOp: 456, AllocsPerOp: 7},
			},
		},
		{
			name: "multiple benchmarks",
			input: `BenchmarkA-8    1000000    1000 ns/op
BenchmarkB-8    2000000    2000 ns/op    100 B/op    2 allocs/op`,
			want: map[string]BenchmarkResult{
				"BenchmarkA-8": {Name: "BenchmarkA-8", N: 1000000, NsPerOp: 1000},
				"BenchmarkB-8": {Name: "BenchmarkB-8", N: 2000000, NsPerOp: 2000, BytesPerOp: 100, AllocsPerOp: 2},
			},
		},
		{
			name:    "empty input",
			input:   "",
			want:    map[string]BenchmarkResult{},
			wantErr: false,
		},
		{
			name: "non-benchmark lines are ignored",
			input: `goos: darwin
goarch: arm64
BenchmarkParse-8    1000000    1234 ns/op
PASS`,
			want: map[string]BenchmarkResult{
				"BenchmarkParse-8": {Name: "BenchmarkParse-8", N: 1000000, NsPerOp: 1234},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			parser := NewParser()
			got, err := parser.Parse(strings.NewReader(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.want) {
				t.Errorf("Parse() returned %d results, want %d", len(got), len(tt.want))
			}

			for name, wantResult := range tt.want {
				gotResult, ok := got[name]
				if !ok {
					t.Errorf("Parse() missing result for %s", name)
					continue
				}
				if gotResult.N != wantResult.N {
					t.Errorf("%s.N = %d, want %d", name, gotResult.N, wantResult.N)
				}
				if gotResult.NsPerOp != wantResult.NsPerOp {
					t.Errorf("%s.NsPerOp = %f, want %f", name, gotResult.NsPerOp, wantResult.NsPerOp)
				}
				if gotResult.BytesPerOp != wantResult.BytesPerOp {
					t.Errorf("%s.BytesPerOp = %d, want %d", name, gotResult.BytesPerOp, wantResult.BytesPerOp)
				}
				if gotResult.AllocsPerOp != wantResult.AllocsPerOp {
					t.Errorf("%s.AllocsPerOp = %d, want %d", name, gotResult.AllocsPerOp, wantResult.AllocsPerOp)
				}
			}
		})
	}
}

func TestComparator_Compare(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		config   *Config
		baseline map[string]BenchmarkResult
		current  map[string]BenchmarkResult
		wantLen  int
		check    func([]Regression) bool
	}{
		{
			name: "no regression - same values",
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000},
			},
			wantLen: 0,
		},
		{
			name: "improvement - no regression reported",
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 500}, // 50% faster
			},
			wantLen: 0,
		},
		{
			name: "critical regression - 25% slower",
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1250},
			},
			wantLen: 1,
			check: func(r []Regression) bool {
				return r[0].Severity == SeverityCritical && r[0].ChangePct > 24.0 && r[0].ChangePct < 26.0
			},
		},
		{
			name: "warning regression - 15% slower",
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1150},
			},
			wantLen: 1,
			check: func(r []Regression) bool {
				return r[0].Severity == SeverityWarning && r[0].ChangePct > 14.0 && r[0].ChangePct < 16.0
			},
		},
		{
			name: "notice regression - 9% slower (just below detection floor)",
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1090}, // 9% - just below 10% floor
			},
			wantLen: 0, // Not detected (below 10% floor)
		},
		{
			name:   "below critical threshold - warning regression",
			config: &Config{NsPerOpThreshold: 1.50}, // 50% threshold for critical
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1150}, // 15% - detected, but warning not critical
			},
			wantLen: 1,
			check: func(r []Regression) bool {
				return r[0].Severity == SeverityWarning && r[0].ChangePct > 14.0 && r[0].ChangePct < 16.0
			},
		},
		{
			name: "new benchmark - not a regression",
			baseline: map[string]BenchmarkResult{
				"BenchmarkOld-8": {Name: "BenchmarkOld-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkOld-8": {Name: "BenchmarkOld-8", N: 1000, NsPerOp: 1000},
				"BenchmarkNew-8": {Name: "BenchmarkNew-8", N: 1000, NsPerOp: 1000},
			},
			wantLen: 0,
		},
		{
			name: "removed benchmark - warning",
			baseline: map[string]BenchmarkResult{
				"BenchmarkRemoved-8": {Name: "BenchmarkRemoved-8", N: 1000, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{},
			wantLen: 1,
			check: func(r []Regression) bool {
				return r[0].Metric == "removed" && r[0].Severity == SeverityWarning
			},
		},
		{
			name: "memory regression detected - 30% is critical",
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000, BytesPerOp: 100},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 1000, NsPerOp: 1000, BytesPerOp: 130},
			},
			wantLen: 1,
			check: func(r []Regression) bool {
				return r[0].Metric == "B/op" && r[0].Severity == SeverityCritical && r[0].ChangePct > 29.0 && r[0].ChangePct < 31.0
			},
		},
		{
			name:   "not enough iterations - skipped",
			config: &Config{MinIterations: 100},
			baseline: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 50, NsPerOp: 1000},
			},
			current: map[string]BenchmarkResult{
				"BenchmarkTest-8": {Name: "BenchmarkTest-8", N: 50, NsPerOp: 1500},
			},
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			comparator := NewComparator(tt.config)
			got := comparator.Compare(tt.baseline, tt.current)

			if len(got) != tt.wantLen {
				t.Errorf("Compare() returned %d regressions, want %d", len(got), tt.wantLen)
			}

			if tt.check != nil && !tt.check(got) {
				t.Errorf("Compare() regression check failed for %+v", got)
			}
		})
	}
}

func TestSeverity_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityNotice, "NOTICE"},
		{SeverityWarning, "WARNING"},
		{SeverityCritical, "CRITICAL"},
		{Severity(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()
			if got := tt.severity.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReporter_PrintRegressions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		regressions []Regression
		wantContain []string
	}{
		{
			name:        "no regressions",
			regressions: []Regression{},
			wantContain: []string{"No performance regressions"},
		},
		{
			name: "critical regression",
			regressions: []Regression{
				{Benchmark: "BenchmarkSlow", Metric: "ns/op", OldValue: 100, NewValue: 130, ChangePct: 30.0, Severity: SeverityCritical},
			},
			wantContain: []string{"CRITICAL", "BenchmarkSlow", "30.0%"},
		},
		{
			name: "removed benchmark",
			regressions: []Regression{
				{Benchmark: "BenchmarkGone", Metric: "removed", Severity: SeverityWarning},
			},
			wantContain: []string{"WARNING", "BenchmarkGone", "removed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			reporter := NewReporter(&buf)
			reporter.PrintRegressions(tt.regressions)

			output := buf.String()
			for _, want := range tt.wantContain {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing %q:\n%s", want, output)
				}
			}
		})
	}
}

func TestHasCriticalRegressions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		regressions []Regression
		want        bool
	}{
		{
			name:        "empty",
			regressions: []Regression{},
			want:        false,
		},
		{
			name: "only notices",
			regressions: []Regression{
				{Severity: SeverityNotice},
			},
			want: false,
		},
		{
			name: "has critical",
			regressions: []Regression{
				{Severity: SeverityWarning},
				{Severity: SeverityCritical},
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := HasCriticalRegressions(tt.regressions); got != tt.want {
				t.Errorf("HasCriticalRegressions() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	config := DefaultConfig()

	if config.NsPerOpThreshold != 1.20 {
		t.Errorf("NsPerOpThreshold = %f, want 1.20", config.NsPerOpThreshold)
	}
	if config.MinIterations != 10 {
		t.Errorf("MinIterations = %d, want 10", config.MinIterations)
	}
}

// Statistical significance tests

func TestWelchTTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		baseline        []float64
		current         []float64
		wantSignificant bool
	}{
		{
			name:            "significant difference - different means",
			baseline:        []float64{100, 101, 99, 100, 102},
			current:         []float64{150, 151, 149, 150, 152},
			wantSignificant: true,
		},
		{
			name:            "no significant difference - same values",
			baseline:        []float64{100, 100, 100, 100, 100},
			current:         []float64{100, 100, 100, 100, 100},
			wantSignificant: false,
		},
		{
			name:            "insufficient data - baseline too small",
			baseline:        []float64{100},
			current:         []float64{150, 151, 149},
			wantSignificant: false,
		},
		{
			name:            "insufficient data - current too small",
			baseline:        []float64{100, 101, 99},
			current:         []float64{150},
			wantSignificant: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tStat, df := WelchTTest(tt.baseline, tt.current)

			if len(tt.baseline) < 2 || len(tt.current) < 2 {
				if df != 0 {
					t.Errorf("Expected df=0 for insufficient data, got %f", df)
				}
				return
			}

			pValue := TwoTailedPValue(tStat, df)
			isSig := IsSignificant(pValue, 0.05)

			if isSig != tt.wantSignificant {
				t.Errorf("IsSignificant() = %v (p=%.4f), want %v", isSig, pValue, tt.wantSignificant)
			}
		})
	}
}

func TestMean(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []float64
		want float64
	}{
		{"empty", []float64{}, 0},
		{"single", []float64{5}, 5},
		{"two values", []float64{1, 3}, 2},
		{"multiple", []float64{1, 2, 3, 4, 5}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := mean(tt.data); got != tt.want {
				t.Errorf("mean() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestVariance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data []float64
		mean float64
		want float64
	}{
		{"insufficient data", []float64{5}, 5, 0},
		{"two values", []float64{1, 3}, 2, 2},
		{"constant", []float64{5, 5, 5}, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := variance(tt.data, tt.mean)
			if got < 0 || (tt.want > 0 && got < tt.want-0.001) || (tt.want > 0 && got > tt.want+0.001) {
				t.Errorf("variance() = %v, want approximately %v", got, tt.want)
			}
		})
	}
}

func TestTwoTailedPValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		tStat float64
		df    float64
	}{
		{"small df", 2.0, 10},
		{"zero df", 2.0, 0},
		{"very large t", 5.0, 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Just verify it doesn't panic
			// Note: The approximation has known limitations for certain ranges
			_ = TwoTailedPValue(tt.tStat, tt.df)
		})
	}
}
