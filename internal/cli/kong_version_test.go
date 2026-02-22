package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gpd/pkg/version"
)

func TestVersionCmd_Run_ReturnsNil(t *testing.T) {
	cmd := &VersionCmd{}
	globals := &Globals{}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Run() error = %v, want nil", err)
	}
}

func TestVersionCmd_Run_OutputFormat(t *testing.T) {
	origVersion := version.Version
	origCommit := version.GitCommit
	origBuild := version.BuildTime
	defer func() {
		version.Version = origVersion
		version.GitCommit = origCommit
		version.BuildTime = origBuild
	}()

	version.Version = "1.0.0"
	version.GitCommit = "abc1234"
	version.BuildTime = "2024-01-15T10:30:00Z"

	cmd := &VersionCmd{}
	globals := &Globals{}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(globals)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	expected := fmt.Sprintf("gpd %s (%s) built %s", version.Version, version.GitCommit, version.BuildTime)
	if output != expected {
		t.Errorf("output = %q, want %q", output, expected)
	}
}

func TestVersionCmd_Run_IgnoresGlobals(t *testing.T) {
	testCases := []struct {
		name    string
		globals *Globals
	}{
		{
			name:    "default globals",
			globals: &Globals{},
		},
		{
			name: "verbose enabled",
			globals: &Globals{
				Verbose: true,
			},
		},
		{
			name: "quiet enabled",
			globals: &Globals{
				Quiet: true,
			},
		},
		{
			name: "with package",
			globals: &Globals{
				Package: "com.example.app",
			},
		},
		{
			name: "with output format",
			globals: &Globals{
				Output: "table",
			},
		},
		{
			name: "with timeout",
			globals: &Globals{
				Timeout: 60 * time.Second,
			},
		},
		{
			name: "all globals set",
			globals: &Globals{
				Package:     "com.test.app",
				Output:      "json",
				Pretty:      true,
				Timeout:     120 * time.Second,
				StoreTokens: "secure",
				Fields:      "version,commit",
				Quiet:       true,
				Verbose:     true,
				KeyPath:     "/path/to/key.json",
				Profile:     "production",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &VersionCmd{}

			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := cmd.Run(tc.globals)

			_ = w.Close()
			os.Stdout = oldStdout

			if err != nil {
				t.Errorf("Run() error = %v, want nil", err)
			}

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			if !strings.Contains(output, "gpd") {
				t.Errorf("output %q does not contain 'gpd'", output)
			}
		})
	}
}

func TestKongCLI_VersionCmdAccessible(t *testing.T) {
	cli := KongCLI{}

	// Verify Version field is of correct type via compile-time type assertion
	_ = cli.Version

	var versionField interface{} = cli.Version
	if _, ok := versionField.(VersionCmd); !ok {
		t.Error("KongCLI.Version is not of type VersionCmd")
	}
}

func TestVersionCmd_Run_OutputContainsExpectedFields(t *testing.T) {
	cmd := &VersionCmd{}
	globals := &Globals{}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(globals)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "gpd") {
		t.Errorf("output %q does not contain 'gpd'", output)
	}

	if !strings.Contains(output, "built") {
		t.Errorf("output %q does not contain 'built'", output)
	}

	if !strings.Contains(output, "(") || !strings.Contains(output, ")") {
		t.Errorf("output %q does not contain parentheses for commit hash", output)
	}
}

func TestVersionCmd_Run_ProducesNonEmptyOutput(t *testing.T) {
	cmd := &VersionCmd{}
	globals := &Globals{}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(globals)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	if output == "" {
		t.Error("output is empty, expected non-empty version string")
	}
}

func TestVersionCmd_StructType(t *testing.T) {
	var cmd VersionCmd

	cmdType := fmt.Sprintf("%T", cmd)
	if cmdType != "cli.VersionCmd" {
		t.Errorf("VersionCmd type = %s, want cli.VersionCmd", cmdType)
	}
}

func TestVersionCmd_Run_MultipleCalls(t *testing.T) {
	cmd := &VersionCmd{}
	globals := &Globals{}

	for i := 0; i < 3; i++ {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := cmd.Run(globals)

		_ = w.Close()
		os.Stdout = oldStdout

		if err != nil {
			t.Errorf("Run() call %d error = %v, want nil", i+1, err)
			continue
		}

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := strings.TrimSpace(buf.String())

		if !strings.HasPrefix(output, "gpd ") {
			t.Errorf("call %d: output %q does not start with 'gpd '", i+1, output)
		}
	}
}

func TestVersionCmd_Run_WithNilGlobals(t *testing.T) {
	cmd := &VersionCmd{}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := cmd.Run(nil)

	_ = w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("Run(nil) error = %v, want nil", err)
	}

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := strings.TrimSpace(buf.String())

	if !strings.HasPrefix(output, "gpd ") {
		t.Errorf("output %q does not start with 'gpd '", output)
	}
}
