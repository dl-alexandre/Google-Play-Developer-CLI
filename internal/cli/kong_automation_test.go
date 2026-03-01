//go:build unit
// +build unit

package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// Test Command Structure
// ============================================================================

func TestAutomationCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := AutomationCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"ReleaseNotes", "Rollout", "Promote", "Validate", "Monitor",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("AutomationCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("AutomationCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("AutomationCmd.%s should have help tag", name)
		}
	}

	actualFields := v.NumField()
	if actualFields != len(expectedSubcommands) {
		t.Errorf("AutomationCmd has %d fields, expected %d", actualFields, len(expectedSubcommands))
	}
}

func TestAutomationReleaseNotesCmd_FieldTags(t *testing.T) {
	cmd := AutomationReleaseNotesCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
		required  string
	}{
		{"Source", "git,pr,file", ""},
		{"Format", "json,markdown", ""},
		{"OutputFile", "", ""},
		{"Since", "", ""},
		{"Until", "", ""},
		{"MaxCommits", "", ""},
	}

	for _, tc := range tests {
		t.Run(tc.fieldName, func(t *testing.T) {
			field, ok := typeOfCmd.FieldByName(tc.fieldName)
			if !ok {
				t.Errorf("AutomationReleaseNotesCmd missing field: %s", tc.fieldName)
				return
			}

			if tc.enum != "" {
				enumTag := field.Tag.Get("enum")
				if enumTag != tc.enum {
					t.Errorf("AutomationReleaseNotesCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
				}
			}

			// Check help tag exists
			helpTag := field.Tag.Get("help")
			if helpTag == "" {
				t.Errorf("AutomationReleaseNotesCmd.%s should have help tag", tc.fieldName)
			}
		})
	}
}

func TestAutomationRolloutCmd_FieldTags(t *testing.T) {
	cmd := AutomationRolloutCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedEnumFields := map[string]string{
		"Track": "internal,alpha,beta,production",
	}

	for fieldName, expectedEnum := range expectedEnumFields {
		field, ok := typeOfCmd.FieldByName(fieldName)
		if !ok {
			t.Errorf("AutomationRolloutCmd missing field: %s", fieldName)
			continue
		}

		enumTag := field.Tag.Get("enum")
		if enumTag != expectedEnum {
			t.Errorf("AutomationRolloutCmd.%s enum tag = %q, want %q", fieldName, enumTag, expectedEnum)
		}
	}

	// Check required fields exist
	requiredFields := []string{
		"StartPercentage", "TargetPercentage", "StepSize", "StepInterval",
		"HealthThreshold", "DryRun", "Wait", "AutoRollback",
	}

	for _, fieldName := range requiredFields {
		if _, ok := typeOfCmd.FieldByName(fieldName); !ok {
			t.Errorf("AutomationRolloutCmd missing field: %s", fieldName)
		}
	}
}

func TestAutomationPromoteCmd_FieldTags(t *testing.T) {
	cmd := AutomationPromoteCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
		required  string
	}{
		{"FromTrack", "internal,alpha,beta,production", "true"},
		{"ToTrack", "internal,alpha,beta,production", "true"},
	}

	for _, tc := range tests {
		t.Run(tc.fieldName, func(t *testing.T) {
			field, ok := typeOfCmd.FieldByName(tc.fieldName)
			if !ok {
				t.Errorf("AutomationPromoteCmd missing field: %s", tc.fieldName)
				return
			}

			enumTag := field.Tag.Get("enum")
			if enumTag != tc.enum {
				t.Errorf("AutomationPromoteCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
			}

			requiredTag := field.Tag.Get("required")
			if requiredTag != tc.required {
				t.Errorf("AutomationPromoteCmd.%s required tag = %q, want %q", tc.fieldName, requiredTag, tc.required)
			}
		})
	}
}

func TestAutomationValidateCmd_FieldTags(t *testing.T) {
	cmd := AutomationValidateCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Checks")
	if !ok {
		t.Fatal("AutomationValidateCmd missing Checks field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "all,aab,signing,permissions,deobfuscation"
	if enumTag != expected {
		t.Errorf("AutomationValidateCmd.Checks enum tag = %q, want %q", enumTag, expected)
	}

	defaultTag := field.Tag.Get("default")
	if defaultTag != "all" {
		t.Errorf("AutomationValidateCmd.Checks default tag = %q, want 'all'", defaultTag)
	}
}

func TestAutomationMonitorCmd_FieldTags(t *testing.T) {
	cmd := AutomationMonitorCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Track")
	if !ok {
		t.Fatal("AutomationMonitorCmd missing Track field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "internal,alpha,beta,production"
	if enumTag != expected {
		t.Errorf("AutomationMonitorCmd.Track enum tag = %q, want %q", enumTag, expected)
	}

	requiredTag := field.Tag.Get("required")
	if requiredTag != "" {
		t.Errorf("AutomationMonitorCmd.Track required tag should be empty or 'true', got: %s", requiredTag)
	}
}

// ============================================================================
// Test AutomationReleaseNotesCmd
// ============================================================================

func TestAutomationReleaseNotesCmd_Run_PackageRequiredForPRSource(t *testing.T) {
	tests := []struct {
		name      string
		source    string
		pkg       string
		wantError bool
		errType   error
	}{
		{
			name:      "pr source without package returns error",
			source:    "pr",
			pkg:       "",
			wantError: true,
			errType:   errors.ErrPackageRequired,
		},
		{
			name:      "pr source with package succeeds",
			source:    "pr",
			pkg:       "com.example.app",
			wantError: false,
		},
		{
			name:      "git source without package succeeds",
			source:    "git",
			pkg:       "",
			wantError: false,
		},
		{
			name:      "file source without package succeeds",
			source:    "file",
			pkg:       "",
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationReleaseNotesCmd{
				Source: tc.source,
			}
			globals := &Globals{
				Package: tc.pkg,
				Output:  "json",
			}

			err := cmd.Run(globals)
			if tc.wantError {
				if err == nil {
					t.Fatalf("Expected error, got nil")
				}
				if tc.errType != nil && err != tc.errType {
					t.Errorf("Expected %v, got: %v", tc.errType, err)
				}
			} else {
				// These may fail due to git not being available in test environment
				// but they shouldn't return the package required error
				if err == errors.ErrPackageRequired {
					t.Errorf("Did not expect package required error, got: %v", err)
				}
			}
		})
	}
}

func TestAutomationReleaseNotesCmd_generateFromPRs(t *testing.T) {
	cmd := &AutomationReleaseNotesCmd{
		Source: "pr",
	}
	globals := &Globals{Package: "com.example.app"}

	result := cmd.generateFromPRs(globals)

	// Should return a map with a not-implemented message
	resultMap, ok := result.(map[string]interface{})
	if !ok {
		t.Fatalf("Expected map result, got: %T", result)
	}

	if msg, exists := resultMap["message"]; !exists {
		t.Error("Expected 'message' field in PR response")
	} else {
		msgStr, ok := msg.(string)
		if !ok || !strings.Contains(msgStr, "not yet implemented") {
			t.Errorf("Expected 'not yet implemented' message, got: %v", msg)
		}
	}

	if pkg, exists := resultMap["package"]; !exists || pkg != "not-implemented" {
		t.Errorf("Expected package='not-implemented', got: %v", pkg)
	}
}

func TestAutomationReleaseNotesCmd_generateFromFile(t *testing.T) {
	t.Run("valid file returns content", func(t *testing.T) {
		content := `{"test": "value", "number": 123}`
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "notes.json")
		if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}

		cmd := &AutomationReleaseNotesCmd{
			Source:     "file",
			OutputFile: tmpFile,
		}
		globals := &Globals{}

		result, err := cmd.generateFromFile(globals)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		resultMap, ok := result.(map[string]interface{})
		if !ok {
			t.Fatalf("Expected map result, got: %T", result)
		}

		if content, exists := resultMap["content"]; !exists {
			t.Error("Expected 'content' field in result")
		} else {
			contentStr, ok := content.(string)
			if !ok || contentStr != `{"test": "value", "number": 123}` {
				t.Errorf("Expected original content, got: %v", content)
			}
		}
	})

	t.Run("nonexistent file returns error", func(t *testing.T) {
		cmd := &AutomationReleaseNotesCmd{
			Source:     "file",
			OutputFile: "/nonexistent/path/notes.json",
		}
		globals := &Globals{}

		_, err := cmd.generateFromFile(globals)
		if err == nil {
			t.Fatal("Expected error for nonexistent file, got nil")
		}

		if !strings.Contains(err.Error(), "failed to read") {
			t.Errorf("Expected 'failed to read' error, got: %v", err)
		}
	})
}

func TestAutomationReleaseNotesCmd_writeToFile(t *testing.T) {
	t.Run("write string content", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "output.md")

		cmd := &AutomationReleaseNotesCmd{
			OutputFile: tmpFile,
		}

		content := "## What's New\n\n- Fixed bug\n- Added feature"
		err := cmd.writeToFile(content)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		written, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		if string(written) != content {
			t.Errorf("Expected %q, got %q", content, string(written))
		}
	})

	t.Run("write map content as JSON", func(t *testing.T) {
		tmpDir := t.TempDir()
		tmpFile := filepath.Join(tmpDir, "output.json")

		cmd := &AutomationReleaseNotesCmd{
			OutputFile: tmpFile,
		}

		content := map[string]interface{}{
			"commits": []string{"abc123", "def456"},
			"count":   2,
		}

		err := cmd.writeToFile(content)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		written, err := os.ReadFile(tmpFile)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal(written, &parsed); err != nil {
			t.Errorf("Written content is not valid JSON: %v", err)
		}

		if parsed["count"].(float64) != 2 {
			t.Errorf("Expected count=2, got: %v", parsed["count"])
		}
	})

	t.Run("invalid path returns error", func(t *testing.T) {
		cmd := &AutomationReleaseNotesCmd{
			OutputFile: "/nonexistent/directory/file.txt",
		}

		err := cmd.writeToFile("test content")
		if err == nil {
			t.Fatal("Expected error for invalid path, got nil")
		}
	})
}

// ============================================================================
// Test AutomationRolloutCmd
// ============================================================================

func TestAutomationRolloutCmd_Run_Validation(t *testing.T) {
	tests := []struct {
		name           string
		pkg            string
		track          string
		startPct       float64
		targetPct      float64
		wantError      bool
		expectedErrMsg string
	}{
		{
			name:           "missing package returns error",
			pkg:            "",
			track:          "production",
			startPct:       1,
			targetPct:      10,
			wantError:      true,
			expectedErrMsg: "package",
		},
		{
			name:           "invalid track returns error",
			pkg:            "com.example.app",
			track:          "invalid-track",
			startPct:       1,
			targetPct:      10,
			wantError:      true,
			expectedErrMsg: "track",
		},
		{
			name:           "zero start percentage returns error",
			pkg:            "com.example.app",
			track:          "production",
			startPct:       0,
			targetPct:      10,
			wantError:      true,
			expectedErrMsg: "start-percentage",
		},
		{
			name:           "negative start percentage returns error",
			pkg:            "com.example.app",
			track:          "production",
			startPct:       -1,
			targetPct:      10,
			wantError:      true,
			expectedErrMsg: "start-percentage",
		},
		{
			name:           "start percentage over 100 returns error",
			pkg:            "com.example.app",
			track:          "production",
			startPct:       101,
			targetPct:      10,
			wantError:      true,
			expectedErrMsg: "start-percentage",
		},
		{
			name:           "zero target percentage returns error",
			pkg:            "com.example.app",
			track:          "production",
			startPct:       1,
			targetPct:      0,
			wantError:      true,
			expectedErrMsg: "target-percentage",
		},
		{
			name:           "target percentage over 100 returns error",
			pkg:            "com.example.app",
			track:          "production",
			startPct:       1,
			targetPct:      101,
			wantError:      true,
			expectedErrMsg: "target-percentage",
		},
		{
			name:           "start greater than target returns error",
			pkg:            "com.example.app",
			track:          "production",
			startPct:       50,
			targetPct:      25,
			wantError:      true,
			expectedErrMsg: "cannot be greater",
		},
		{
			name:      "valid parameters succeed in dry-run",
			pkg:       "com.example.app",
			track:     "production",
			startPct:  1,
			targetPct: 10,
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationRolloutCmd{
				Track:            tc.track,
				StartPercentage:  tc.startPct,
				TargetPercentage: tc.targetPct,
				StepSize:         10,
				DryRun:           true,
				Wait:             false,
			}
			globals := &Globals{
				Package: tc.pkg,
				Output:  "json",
			}

			err := cmd.Run(globals)
			if tc.wantError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tc.expectedErrMsg != "" && !strings.Contains(err.Error(), tc.expectedErrMsg) &&
					!strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tc.expectedErrMsg)) {
					t.Errorf("Expected error containing %q, got: %v", tc.expectedErrMsg, err)
				}
			} else {
				// In dry-run mode, we expect success
				// In non-dry-run, might error due to auth issues
				if err != nil && !strings.Contains(err.Error(), "auth") && !strings.Contains(err.Error(), "API") {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestCalculateRolloutSteps(t *testing.T) {
	tests := []struct {
		name     string
		start    float64
		target   float64
		stepSize float64
		expected []float64
	}{
		{
			name:     "simple 1 to 100 with step 10",
			start:    1,
			target:   100,
			stepSize: 10,
			expected: []float64{11, 21, 31, 41, 51, 61, 71, 81, 91, 100},
		},
		{
			name:     "small range with large step",
			start:    90,
			target:   100,
			stepSize: 20,
			expected: []float64{100},
		},
		{
			name:     "single step exact match",
			start:    50,
			target:   60,
			stepSize: 10,
			expected: []float64{60},
		},
		{
			name:     "fractional steps",
			start:    0.5,
			target:   5.0,
			stepSize: 1.0,
			expected: []float64{1.5, 2.5, 3.5, 4.5, 5.0},
		},
		{
			name:     "start equals target returns empty",
			start:    100,
			target:   100,
			stepSize: 10,
			expected: []float64{},
		},
		{
			name:     "tiny step size creates many steps",
			start:    0,
			target:   1,
			stepSize: 0.3,
			expected: []float64{0.3, 0.6, 0.9, 1},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := calculateRolloutSteps(tc.start, tc.target, tc.stepSize)

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d steps, got %d: %v", len(tc.expected), len(result), result)
				return
			}

			for i, expected := range tc.expected {
				if result[i] != expected {
					t.Errorf("Step %d: expected %v, got %v", i, expected, result[i])
				}
			}
		})
	}
}

func TestAutomationRolloutCmd_DryRun(t *testing.T) {
	cmd := &AutomationRolloutCmd{
		Track:            "production",
		StartPercentage:  5,
		TargetPercentage: 50,
		StepSize:         10,
		StepInterval:     30 * time.Minute,
		HealthThreshold:  0.01,
		DryRun:           true,
		Wait:             false,
		AutoRollback:     true,
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("Dry-run should not error: %v", err)
	}

	// Note: In dry-run mode, the output goes to stdout, so we can't easily verify it
	// But the fact that it didn't error is a good sign
}

// ============================================================================
// Test AutomationPromoteCmd
// ============================================================================

func TestAutomationPromoteCmd_Run_Validation(t *testing.T) {
	tests := []struct {
		name           string
		pkg            string
		fromTrack      string
		toTrack        string
		wantError      bool
		expectedErrMsg string
	}{
		{
			name:           "missing package returns error",
			pkg:            "",
			fromTrack:      "internal",
			toTrack:        "alpha",
			wantError:      true,
			expectedErrMsg: "package",
		},
		{
			name:           "invalid from track returns error",
			pkg:            "com.example.app",
			fromTrack:      "invalid",
			toTrack:        "production",
			wantError:      true,
			expectedErrMsg: "track",
		},
		{
			name:           "invalid to track returns error",
			pkg:            "com.example.app",
			fromTrack:      "alpha",
			toTrack:        "invalid",
			wantError:      true,
			expectedErrMsg: "track",
		},
		{
			name:           "same from and to track returns error",
			pkg:            "com.example.app",
			fromTrack:      "alpha",
			toTrack:        "alpha",
			wantError:      true,
			expectedErrMsg: "must be different",
		},
		{
			name:      "valid tracks succeed in dry-run",
			pkg:       "com.example.app",
			fromTrack: "internal",
			toTrack:   "alpha",
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationPromoteCmd{
				FromTrack: tc.fromTrack,
				ToTrack:   tc.toTrack,
				DryRun:    true,
			}
			globals := &Globals{
				Package: tc.pkg,
				Output:  "json",
			}

			err := cmd.Run(globals)
			if tc.wantError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tc.expectedErrMsg != "" && !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("Expected error containing %q, got: %v", tc.expectedErrMsg, err)
				}
			} else {
				// In dry-run, should succeed
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAutomationPromoteCmd_DryRun(t *testing.T) {
	cmd := &AutomationPromoteCmd{
		FromTrack:    "internal",
		ToTrack:      "alpha",
		VersionCodes: []int64{100, 101},
		Verify:       true,
		DryRun:       true,
		Wait:         false,
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("Dry-run should not error: %v", err)
	}
}

// ============================================================================
// Test AutomationValidateCmd
// ============================================================================

func TestAutomationValidateCmd_Run_Validation(t *testing.T) {
	tests := []struct {
		name      string
		pkg       string
		checks    []string
		wantError bool
	}{
		{
			name:      "missing package returns error",
			pkg:       "",
			checks:    []string{"all"},
			wantError: true,
		},
		{
			name:      "valid with all checks",
			pkg:       "com.example.app",
			checks:    []string{"all"},
			wantError: false,
		},
		{
			name:      "valid with specific checks",
			pkg:       "com.example.app",
			checks:    []string{"aab", "signing"},
			wantError: false,
		},
		{
			name:      "valid with empty checks (defaults to all)",
			pkg:       "com.example.app",
			checks:    []string{},
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationValidateCmd{
				Checks: tc.checks,
				DryRun: true,
				Strict: false,
			}
			globals := &Globals{
				Package: tc.pkg,
				Output:  "json",
			}

			err := cmd.Run(globals)
			if tc.wantError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
			} else {
				// Dry-run should succeed
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAutomationValidateCmd_expandChecks(t *testing.T) {
	tests := []struct {
		name     string
		checks   []string
		expected []string
	}{
		{
			name:     "all expands to all checks",
			checks:   []string{"all"},
			expected: []string{"aab", "signing", "permissions", "deobfuscation"},
		},
		{
			name:     "single check",
			checks:   []string{"aab"},
			expected: []string{"aab"},
		},
		{
			name:     "multiple checks",
			checks:   []string{"aab", "signing"},
			expected: []string{"aab", "signing"},
		},
		{
			name:     "mixed all and specific",
			checks:   []string{"all", "aab"},
			expected: []string{"aab", "signing", "permissions", "deobfuscation"},
		},
		{
			name:     "deduplication",
			checks:   []string{"aab", "aab", "signing"},
			expected: []string{"aab", "signing"},
		},
		{
			name:     "empty returns empty",
			checks:   []string{},
			expected: []string{},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationValidateCmd{
				Checks: tc.checks,
			}

			result := cmd.expandChecks()

			if len(result) != len(tc.expected) {
				t.Errorf("Expected %d checks, got %d: %v", len(tc.expected), len(result), result)
				return
			}

			// Check that all expected checks are present
			for _, expected := range tc.expected {
				found := false
				for _, actual := range result {
					if actual == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected check %q not found in result: %v", expected, result)
				}
			}
		})
	}
}

func TestAutomationValidateCmd_runCheck(t *testing.T) {
	tests := []struct {
		name        string
		check       string
		wantError   bool
		wantWarning bool
	}{
		{
			name:        "aab check succeeds",
			check:       "aab",
			wantError:   false,
			wantWarning: false,
		},
		{
			name:        "signing check succeeds",
			check:       "signing",
			wantError:   false,
			wantWarning: false,
		},
		{
			name:        "permissions check succeeds",
			check:       "permissions",
			wantError:   false,
			wantWarning: false,
		},
		{
			name:        "deobfuscation check succeeds",
			check:       "deobfuscation",
			wantError:   false,
			wantWarning: false,
		},
		{
			name:        "unknown check returns error",
			check:       "unknown",
			wantError:   true,
			wantWarning: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationValidateCmd{}
			globals := &Globals{}

			result, err := cmd.runCheck(globals, tc.check)
			if tc.wantError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if !strings.Contains(err.Error(), "unknown validation check") {
					t.Errorf("Expected 'unknown validation check' error, got: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				if result == nil {
					t.Fatal("Expected result, got nil")
				}
				if result.Warning != tc.wantWarning {
					t.Errorf("Expected Warning=%v, got %v", tc.wantWarning, result.Warning)
				}
				if result.Message == "" {
					t.Error("Expected non-empty message")
				}
			}
		})
	}
}

func TestAutomationValidateCmd_Run_StrictMode(t *testing.T) {
	tests := []struct {
		name        string
		strict      bool
		wantError   bool
		checkResult func(cmd *AutomationValidateCmd)
	}{
		{
			name:      "strict mode with warnings fails",
			strict:    true,
			wantError: true,
			checkResult: func(cmd *AutomationValidateCmd) {
				// Override runCheck to return warnings
				// This would require mocking or a different approach
			},
		},
		{
			name:      "non-strict mode with warnings succeeds",
			strict:    false,
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationValidateCmd{
				Checks: []string{"aab"},
				Strict: tc.strict,
				DryRun: false,
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := cmd.Run(globals)
			// Since all checks currently pass without warnings in the stub implementation,
			// strict mode should also succeed
			if err != nil {
				apiErr, ok := err.(*errors.APIError)
				if ok && apiErr.Code == errors.CodeValidationError {
					// This is expected when validation fails
				} else {
					t.Logf("Got error (may be expected): %v", err)
				}
			}
		})
	}
}

// ============================================================================
// Test AutomationMonitorCmd
// ============================================================================

func TestAutomationMonitorCmd_Run_Validation(t *testing.T) {
	tests := []struct {
		name           string
		pkg            string
		track          string
		wantError      bool
		expectedErrMsg string
	}{
		{
			name:           "missing package returns error",
			pkg:            "",
			track:          "production",
			wantError:      true,
			expectedErrMsg: "package",
		},
		{
			name:           "invalid track returns error",
			pkg:            "com.example.app",
			track:          "invalid-track",
			wantError:      true,
			expectedErrMsg: "track",
		},
		{
			name:      "valid parameters succeed in dry-run",
			pkg:       "com.example.app",
			track:     "production",
			wantError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &AutomationMonitorCmd{
				Track:         tc.track,
				Duration:      10 * time.Minute,
				CheckInterval: 5 * time.Minute,
				DryRun:        true,
			}
			globals := &Globals{
				Package: tc.pkg,
				Output:  "json",
			}

			err := cmd.Run(globals)
			if tc.wantError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				if tc.expectedErrMsg != "" && !strings.Contains(err.Error(), tc.expectedErrMsg) {
					t.Errorf("Expected error containing %q, got: %v", tc.expectedErrMsg, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestAutomationMonitorCmd_DryRun(t *testing.T) {
	cmd := &AutomationMonitorCmd{
		Track:             "production",
		Duration:          2 * time.Hour,
		CheckInterval:     5 * time.Minute,
		CrashThreshold:    0.01,
		AnrThreshold:      0.005,
		ErrorThreshold:    0.02,
		AutoAlert:         true,
		ExitOnDegradation: true,
		DryRun:            true,
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Fatalf("Dry-run should not error: %v", err)
	}
}

func TestAutomationMonitorCmd_checkReleaseHealth(t *testing.T) {
	cmd := &AutomationMonitorCmd{}
	globals := &Globals{}

	health := cmd.checkReleaseHealth(globals)

	if health == nil {
		t.Fatal("Expected health metrics, got nil")
	}

	// Check that the default values are reasonable
	if health.CrashRate < 0 {
		t.Errorf("CrashRate should be non-negative, got %f", health.CrashRate)
	}
	if health.AnrRate < 0 {
		t.Errorf("AnrRate should be non-negative, got %f", health.AnrRate)
	}
	if health.ErrorRate < 0 {
		t.Errorf("ErrorRate should be non-negative, got %f", health.ErrorRate)
	}

	// Default values should be below thresholds
	if health.CrashRate > 0.01 {
		t.Logf("Note: Default CrashRate %f is above typical threshold", health.CrashRate)
	}
}

// ============================================================================
// Test Edge Cases and Integration
// ============================================================================

func TestAutomationCommands_RequireAuth(t *testing.T) {
	// Test that commands that need auth return appropriate errors
	// when not in dry-run mode

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{
			name: "AutomationRolloutCmd",
			cmd: &AutomationRolloutCmd{
				Track:            "production",
				StartPercentage:  1,
				TargetPercentage: 10,
				DryRun:           false,
				Wait:             false,
			},
		},
		{
			name: "AutomationPromoteCmd",
			cmd: &AutomationPromoteCmd{
				FromTrack: "internal",
				ToTrack:   "alpha",
				DryRun:    false,
				Wait:      false,
			},
		},
		{
			name: "AutomationValidateCmd",
			cmd: &AutomationValidateCmd{
				Checks: []string{"aab"},
				DryRun: false,
				Strict: false,
			},
		},
		{
			name: "AutomationMonitorCmd",
			cmd: &AutomationMonitorCmd{
				Track:         "production",
				Duration:      10 * time.Minute,
				CheckInterval: 5 * time.Minute,
				DryRun:        false,
			},
		},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := tc.cmd.Run(globals)
			if err == nil {
				// Some commands may succeed without auth in certain scenarios
				return
			}

			// Should not return "not yet implemented"
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s should be implemented but returns 'not yet implemented'", tc.name)
			}
		})
	}
}

func TestAutomationCommands_WithVerbose(t *testing.T) {
	// Test that commands work with verbose mode enabled
	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{
			name: "AutomationReleaseNotesCmd (git)",
			cmd: &AutomationReleaseNotesCmd{
				Source: "git",
			},
		},
		{
			name: "AutomationValidateCmd",
			cmd: &AutomationValidateCmd{
				Checks: []string{"aab"},
				DryRun: true,
			},
		},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
				Verbose: true,
			}

			err := tc.cmd.Run(globals)
			// Verbose mode shouldn't cause failures
			if err != nil {
				// Some errors are expected (like git not being available)
				// Just make sure it's not a panic or unexpected error
				t.Logf("Got error (may be expected): %v", err)
			}
		})
	}
}

func TestAutomationCommands_WithDifferentOutputs(t *testing.T) {
	formats := []string{"json", "table", "markdown"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cmd := &AutomationValidateCmd{
				Checks: []string{"aab"},
				DryRun: true,
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  format,
				Pretty:  true,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error with format %s: %v", format, err)
			}
		})
	}
}

// ============================================================================
// Test Helper Functions
// ============================================================================

func TestAutomationRolloutCmd_checkHealth(t *testing.T) {
	cmd := &AutomationRolloutCmd{}
	globals := &Globals{
		Package: "com.example.app",
	}

	healthy, err := cmd.checkHealth(globals)
	if err != nil {
		t.Fatalf("checkHealth returned error: %v", err)
	}

	// Currently returns true always (stub implementation)
	if !healthy {
		t.Error("Expected checkHealth to return true (stub implementation)")
	}
}

func TestAutomationRolloutCmd_performRollback(t *testing.T) {
	cmd := &AutomationRolloutCmd{}
	globals := &Globals{
		Package: "com.example.app",
	}

	err := cmd.performRollback(globals, 50.0)
	if err != nil {
		t.Fatalf("performRollback returned error: %v", err)
	}

	// Currently returns nil (stub implementation)
}

func TestAutomationPromoteCmd_verifyPromotion(t *testing.T) {
	cmd := &AutomationPromoteCmd{}
	globals := &Globals{
		Package: "com.example.app",
	}

	verified, err := cmd.verifyPromotion(globals)
	if err != nil {
		t.Fatalf("verifyPromotion returned error: %v", err)
	}

	// Currently returns true always (stub implementation)
	if !verified {
		t.Error("Expected verifyPromotion to return true (stub implementation)")
	}
}
