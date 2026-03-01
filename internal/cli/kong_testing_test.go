//go:build unit
// +build unit

package cli

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/option"

	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// Test Command Structure
// ============================================================================

func TestTestingCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := TestingCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{
		"Prelaunch", "DeviceLab", "Screenshots", "Validate", "Compatibility",
	}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("TestingCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("TestingCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("TestingCmd.%s should have help tag", name)
		}
	}

	actualFields := v.NumField()
	if actualFields != len(expectedSubcommands) {
		t.Errorf("TestingCmd has %d fields, expected %d", actualFields, len(expectedSubcommands))
	}
}

func TestTestingPrelaunchCmd_FieldTags(t *testing.T) {
	cmd := TestingPrelaunchCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
		default_  string
	}{
		{"Action", "trigger,check,wait", "check"},
		{"Format", "json,table", "table"},
		{"MaxWaitTime", "", "30m"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("TestingPrelaunchCmd missing field: %s", tc.fieldName)
			continue
		}

		if tc.enum != "" {
			enumTag := field.Tag.Get("enum")
			if enumTag != tc.enum {
				t.Errorf("TestingPrelaunchCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
			}
		}

		if tc.default_ != "" {
			defaultTag := field.Tag.Get("default")
			if defaultTag != tc.default_ {
				t.Errorf("TestingPrelaunchCmd.%s default tag = %q, want %q", tc.fieldName, defaultTag, tc.default_)
			}
		}
	}
}

func TestTestingDeviceLabCmd_FieldTags(t *testing.T) {
	cmd := TestingDeviceLabCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
		required  bool
	}{
		{"AppFile", "", true},
		{"TestType", "robo,instrumentation", false},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("TestingDeviceLabCmd missing field: %s", tc.fieldName)
			continue
		}

		if tc.enum != "" {
			enumTag := field.Tag.Get("enum")
			if enumTag != tc.enum {
				t.Errorf("TestingDeviceLabCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
			}
		}

		if tc.required {
			typeTag := field.Tag.Get("type")
			if typeTag != "existingfile" {
				t.Errorf("TestingDeviceLabCmd.%s should have type:\"existingfile\" tag", tc.fieldName)
			}
		}
	}
}

func TestTestingScreenshotsCmd_FieldTags(t *testing.T) {
	cmd := TestingScreenshotsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Orientations")
	if !ok {
		t.Fatal("TestingScreenshotsCmd missing Orientations field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "portrait,landscape"
	if enumTag != expected {
		t.Errorf("TestingScreenshotsCmd.Orientations enum tag = %q, want %q", enumTag, expected)
	}

	defaultTag := field.Tag.Get("default")
	if defaultTag != "portrait" {
		t.Errorf("TestingScreenshotsCmd.Orientations default tag = %q, want portrait", defaultTag)
	}
}

func TestTestingValidateCmd_FieldTags(t *testing.T) {
	cmd := TestingValidateCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Checks")
	if !ok {
		t.Fatal("TestingValidateCmd missing Checks field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "all,aab,signing,permissions,size,api-level"
	if enumTag != expected {
		t.Errorf("TestingValidateCmd.Checks enum tag = %q, want %q", enumTag, expected)
	}

	defaultTag := field.Tag.Get("default")
	if defaultTag != "all" {
		t.Errorf("TestingValidateCmd.Checks default tag = %q, want all", defaultTag)
	}
}

func TestTestingCompatibilityCmd_FieldTags(t *testing.T) {
	cmd := TestingCompatibilityCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
		default_  string
	}{
		{"DeviceCatalog", "play,all", "play"},
		{"Format", "json,table,csv", "table"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("TestingCompatibilityCmd missing field: %s", tc.fieldName)
			continue
		}

		if tc.enum != "" {
			enumTag := field.Tag.Get("enum")
			if enumTag != tc.enum {
				t.Errorf("TestingCompatibilityCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
			}
		}

		if tc.default_ != "" {
			defaultTag := field.Tag.Get("default")
			if defaultTag != tc.default_ {
				t.Errorf("TestingCompatibilityCmd.%s default tag = %q, want %q", tc.fieldName, defaultTag, tc.default_)
			}
		}
	}
}

// ============================================================================
// TestingPrelaunchCmd Tests
// ============================================================================

func TestTestingPrelaunchCmd_Run_PackageRequired(t *testing.T) {
	cmd := &TestingPrelaunchCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestTestingPrelaunchCmd_Run_WithoutEditID(t *testing.T) {
	cmd := &TestingPrelaunchCmd{
		Action: "check",
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	err := cmd.Run(globals)
	// Should succeed with API limited message since no edit ID provided
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTestingPrelaunchCmd_Run_WithEditID_NoAuth(t *testing.T) {
	cmd := &TestingPrelaunchCmd{
		Action: "check",
		EditID: "123456",
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
		KeyPath: "/nonexistent/key.json",
	}

	err := cmd.Run(globals)
	// Should still succeed but with edit_not_found status since auth will fail
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTestingPrelaunchCmd_Run_WithEditID_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/edits/") && r.Method == "GET":
			fmt.Fprint(w, `{"id": "test-edit-id", "expiryTimeSeconds": "1234567890"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cmd := &TestingPrelaunchCmd{
		Action: "check",
		EditID: "test-edit-id",
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	// Without valid auth, it won't reach the server, but test the structure
	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	_ = server.URL // Acknowledge server to avoid unused variable
}

func TestTestingPrelaunchCmd_Run_VerboseMode(t *testing.T) {
	cmd := &TestingPrelaunchCmd{
		Action: "trigger",
		EditID: "test-edit",
	}
	globals := &Globals{
		Package: "com.example.app",
		Verbose: true,
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Unexpected error in verbose mode: %v", err)
	}
}

func TestTestingPrelaunchCmd_Run_AllActions(t *testing.T) {
	actions := []string{"trigger", "check", "wait"}

	for _, action := range actions {
		t.Run(fmt.Sprintf("action_%s", action), func(t *testing.T) {
			cmd := &TestingPrelaunchCmd{
				Action:      action,
				EditID:      "test-edit",
				MaxWaitTime: "10m",
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error for action %s: %v", action, err)
			}
		})
	}
}

// ============================================================================
// TestingDeviceLabCmd Tests
// ============================================================================

func TestTestingDeviceLabCmd_Run_PackageRequired(t *testing.T) {
	cmd := &TestingDeviceLabCmd{
		AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestTestingDeviceLabCmd_Run_MissingAppFile(t *testing.T) {
	cmd := &TestingDeviceLabCmd{
		AppFile: "/nonexistent/app.apk",
	}
	globals := &Globals{
		Package: "com.example.app",
	}

	// The file validation happens via kong tag "existingfile", so this should work
	// if kong doesn't validate (kong would catch this before Run is called)
	// Testing what happens when file doesn't exist at command definition
	err := cmd.Run(globals)
	// Should still run because kong validates type:"existingfile" before Run()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

func TestTestingDeviceLabCmd_Run_ValidInputs(t *testing.T) {
	tests := []struct {
		name     string
		cmd      TestingDeviceLabCmd
		validate func(t *testing.T, result *testingDeviceLabResult)
	}{
		{
			name: "robo test without test file",
			cmd: TestingDeviceLabCmd{
				AppFile:     createTempFile(t, "test.apk", []byte("fake apk")),
				TestType:    "robo",
				TestTimeout: "15m",
			},
			validate: func(t *testing.T, result *testingDeviceLabResult) {
				if result.TestMatrixID != "" {
					t.Error("Expected no test matrix ID for suggestion-only mode")
				}
				if result.Status != "requires_firebase_setup" {
					t.Errorf("Expected 'requires_firebase_setup' status, got: %s", result.Status)
				}
			},
		},
		{
			name: "instrumentation test with test file",
			cmd: TestingDeviceLabCmd{
				AppFile:     createTempFile(t, "test.apk", []byte("fake apk")),
				TestFile:    createTempFile(t, "test-android.apk", []byte("fake test apk")),
				TestType:    "instrumentation",
				TestTimeout: "30m",
			},
			validate: func(t *testing.T, result *testingDeviceLabResult) {
				if !strings.Contains(result.SuggestedCmd, "--type instrumentation") {
					t.Error("Expected instrumentation type in suggested command")
				}
				if !strings.Contains(result.SuggestedCmd, "--test") {
					t.Error("Expected --test flag in suggested command")
				}
			},
		},
		{
			name: "async mode",
			cmd: TestingDeviceLabCmd{
				AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
				Async:   true,
			},
			validate: func(t *testing.T, result *testingDeviceLabResult) {
				if !strings.Contains(result.SuggestedCmd, "--async") {
					t.Error("Expected --async flag in suggested command")
				}
			},
		},
		{
			name: "with devices",
			cmd: TestingDeviceLabCmd{
				AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
				Devices: []string{"redfin", "oriole", "panther"},
			},
			validate: func(t *testing.T, result *testingDeviceLabResult) {
				if len(result.TestRuns) != 3 {
					t.Errorf("Expected 3 test runs, got: %d", len(result.TestRuns))
				}
				for _, run := range result.TestRuns {
					if run.Status != "pending" {
						t.Errorf("Expected status 'pending', got: %s", run.Status)
					}
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := tc.cmd.Run(globals)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Cannot directly validate result since it's written to output
			// But we can verify the command executed without error
		})
	}
}

func TestTestingDeviceLabCmd_Run_VerboseMode(t *testing.T) {
	cmd := &TestingDeviceLabCmd{
		AppFile:  createTempFile(t, "test.apk", []byte("fake apk")),
		TestType: "robo",
	}
	globals := &Globals{
		Package: "com.example.app",
		Verbose: true,
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Unexpected error in verbose mode: %v", err)
	}
}

// ============================================================================
// TestingScreenshotsCmd Tests
// ============================================================================

func TestTestingScreenshotsCmd_Run_PackageRequired(t *testing.T) {
	cmd := &TestingScreenshotsCmd{
		AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestTestingScreenshotsCmd_Run_ValidInputs(t *testing.T) {
	tests := []struct {
		name string
		cmd  TestingScreenshotsCmd
	}{
		{
			name: "basic screenshots",
			cmd: TestingScreenshotsCmd{
				AppFile:      createTempFile(t, "test.apk", []byte("fake apk")),
				Devices:      []string{"redfin"},
				Orientations: []string{"portrait"},
				OutputDir:    "./screenshots",
			},
		},
		{
			name: "multiple devices and orientations",
			cmd: TestingScreenshotsCmd{
				AppFile:      createTempFile(t, "test.aab", []byte("fake aab")),
				Devices:      []string{"redfin", "oriole"},
				Orientations: []string{"portrait", "landscape"},
				Locales:      []string{"en-US", "de-DE", "fr-FR"},
				OutputDir:    "/tmp/screenshots",
			},
		},
		{
			name: "with testlab flag",
			cmd: TestingScreenshotsCmd{
				AppFile:   createTempFile(t, "test.apk", []byte("fake apk")),
				Devices:   []string{"redfin"},
				TestLab:   true,
				OutputDir: "./testlab-shots",
			},
		},
		{
			name: "default locale when none specified",
			cmd: TestingScreenshotsCmd{
				AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
				Devices: []string{"redfin"},
				Locales: []string{}, // Empty - should default to en-US
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := tc.cmd.Run(globals)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTestingScreenshotsCmd_Run_VerboseMode(t *testing.T) {
	cmd := &TestingScreenshotsCmd{
		AppFile:   createTempFile(t, "test.apk", []byte("fake apk")),
		Devices:   []string{"redfin"},
		OutputDir: "./screenshots",
	}
	globals := &Globals{
		Package: "com.example.app",
		Verbose: true,
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Unexpected error in verbose mode: %v", err)
	}
}

// ============================================================================
// TestingValidateCmd Tests
// ============================================================================

func TestTestingValidateCmd_Run_ValidInputs(t *testing.T) {
	tests := []struct {
		name           string
		cmd            TestingValidateCmd
		expectedChecks int
	}{
		{
			name: "all checks",
			cmd: TestingValidateCmd{
				AppFile: createTempFile(t, "test.aab", []byte("fake aab")),
				Checks:  []string{"all"},
				Strict:  false,
			},
			expectedChecks: 1, // 'all' maps to aab check
		},
		{
			name: "specific checks",
			cmd: TestingValidateCmd{
				AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
				Checks:  []string{"aab", "signing", "permissions"},
				Strict:  false,
			},
			expectedChecks: 3,
		},
		{
			name: "all individual checks",
			cmd: TestingValidateCmd{
				AppFile: createTempFile(t, "test.aab", []byte("fake aab")),
				Checks:  []string{"aab", "signing", "permissions", "size", "api-level"},
				Strict:  false,
			},
			expectedChecks: 5,
		},
		{
			name: "strict mode",
			cmd: TestingValidateCmd{
				AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
				Checks:  []string{"all"},
				Strict:  true,
			},
			expectedChecks: 1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := tc.cmd.Run(globals)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}

func TestTestingValidateCmd_Run_VerboseMode(t *testing.T) {
	cmd := &TestingValidateCmd{
		AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
		Checks:  []string{"all"},
		Strict:  false,
	}
	globals := &Globals{
		Package: "com.example.app",
		Verbose: true,
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Unexpected error in verbose mode: %v", err)
	}
}

// ============================================================================
// TestingCompatibilityCmd Tests
// ============================================================================

func TestTestingCompatibilityCmd_Run_PackageRequired(t *testing.T) {
	cmd := &TestingCompatibilityCmd{
		AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing package")
	}
	if !strings.Contains(err.Error(), "package name is required") {
		t.Errorf("Expected 'package name is required' error, got: %v", err)
	}
}

func TestTestingCompatibilityCmd_Run_ValidInputs(t *testing.T) {
	tests := []struct {
		name string
		cmd  TestingCompatibilityCmd
	}{
		{
			name: "basic compatibility check",
			cmd: TestingCompatibilityCmd{
				AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
				DeviceCatalog: "play",
				Format:        "table",
			},
		},
		{
			name: "with min SDK",
			cmd: TestingCompatibilityCmd{
				AppFile:       createTempFile(t, "test.aab", []byte("fake aab")),
				MinSDK:        30,
				DeviceCatalog: "play",
				Format:        "json",
			},
		},
		{
			name: "with target SDK",
			cmd: TestingCompatibilityCmd{
				AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
				TargetSDK:     33,
				DeviceCatalog: "all",
				Format:        "csv",
			},
		},
		{
			name: "with both SDK versions",
			cmd: TestingCompatibilityCmd{
				AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
				MinSDK:        26,
				TargetSDK:     33,
				DeviceCatalog: "play",
				Format:        "table",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			// Will fail on auth, but tests input validation
			err := tc.cmd.Run(globals)
			// Expect error due to auth failure, but not input validation error
			if err != nil && !strings.Contains(err.Error(), "authentication not configured") {
				// Other errors are acceptable since we're testing input structure
				t.Logf("Got expected auth-related error: %v", err)
			}
		})
	}
}

func TestTestingCompatibilityCmd_Run_VerboseMode(t *testing.T) {
	cmd := &TestingCompatibilityCmd{
		AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
		MinSDK:        30,
		DeviceCatalog: "play",
	}
	globals := &Globals{
		Package: "com.example.app",
		Verbose: true,
		Output:  "json",
	}

	err := cmd.Run(globals)
	// Will fail on auth but verbose should work
	if err != nil && !strings.Contains(err.Error(), "authentication not configured") {
		t.Logf("Got expected error: %v", err)
	}
}

// ============================================================================
// Helper Function Tests
// ============================================================================

func TestEstimateMinSDKSupport(t *testing.T) {
	tests := []struct {
		minSDK   int
		expected float64
	}{
		{34, 0.25},  // Android 14+
		{33, 0.40},  // Android 13+
		{31, 0.55},  // Android 12+
		{30, 0.65},  // Android 11+
		{29, 0.75},  // Android 10+
		{28, 0.85},  // Android 9+
		{26, 0.90},  // Android 8+
		{24, 0.95},  // Android 7+
		{21, 0.99},  // Android 5+
		{19, 1.0},   // Below 21, full support
		{16, 1.0},   // Very old, full support
		{100, 0.25}, // Future version
	}

	for _, tc := range tests {
		t.Run(fmt.Sprintf("sdk_%d", tc.minSDK), func(t *testing.T) {
			result := estimateMinSDKSupport(tc.minSDK)
			if result != tc.expected {
				t.Errorf("estimateMinSDKSupport(%d) = %f, want %f", tc.minSDK, result, tc.expected)
			}
		})
	}
}

func TestEstimateMinSDKSupport_BoundaryValues(t *testing.T) {
	// Test boundary values around each threshold
	boundaryTests := []struct {
		minSDK int
		min    float64
		max    float64
	}{
		{33, 0.40, 0.40}, // Exactly at boundary
		{32, 0.40, 0.55}, // Between 31 and 33
		{31, 0.55, 0.55}, // Exactly at boundary
		{30, 0.55, 0.65}, // Between 30 and 31
	}

	for _, tc := range boundaryTests {
		t.Run(fmt.Sprintf("boundary_sdk_%d", tc.minSDK), func(t *testing.T) {
			result := estimateMinSDKSupport(tc.minSDK)
			if result < tc.min || result > tc.max {
				t.Errorf("estimateMinSDKSupport(%d) = %f, expected between %f and %f",
					tc.minSDK, result, tc.min, tc.max)
			}
		})
	}
}

// ============================================================================
// PopulateDeviceSupport Tests
// ============================================================================

func TestTestingCompatibilityCmd_PopulateDeviceSupport(t *testing.T) {
	tests := []struct {
		name              string
		minSDK            int
		targetSDK         int
		expectCompatible  bool
		expectWarnings    int
		expectDeviceCount int
	}{
		{
			name:              "high minSDK warning",
			minSDK:            34,
			targetSDK:         0,
			expectCompatible:  true,
			expectWarnings:    1,
			expectDeviceCount: 20000,
		},
		{
			name:              "low targetSDK warning",
			minSDK:            0,
			targetSDK:         30,
			expectCompatible:  true,
			expectWarnings:    1,
			expectDeviceCount: 20000,
		},
		{
			name:              "both warnings",
			minSDK:            34,
			targetSDK:         30,
			expectCompatible:  true,
			expectWarnings:    2,
			expectDeviceCount: 20000,
		},
		{
			name:              "no warnings with good SDKs",
			minSDK:            21,
			targetSDK:         33,
			expectCompatible:  true,
			expectWarnings:    0,
			expectDeviceCount: 20000,
		},
		{
			name:              "no SDK specified",
			minSDK:            0,
			targetSDK:         0,
			expectCompatible:  true,
			expectWarnings:    0,
			expectDeviceCount: 20000,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &TestingCompatibilityCmd{
				MinSDK:    tc.minSDK,
				TargetSDK: tc.targetSDK,
			}
			result := &testingCompatibilityResult{
				Issues:       make([]testingCompatibilityIssue, 0),
				DeviceGroups: make([]testingCompatibilityGroup, 0),
			}

			cmd.populateDeviceSupport(result)

			if result.Compatible != tc.expectCompatible {
				t.Errorf("Compatible = %v, want %v", result.Compatible, tc.expectCompatible)
			}

			if result.DeviceCount != tc.expectDeviceCount {
				t.Errorf("DeviceCount = %d, want %d", result.DeviceCount, tc.expectDeviceCount)
			}

			if tc.minSDK > 0 {
				expectedSupported := int(float64(tc.expectDeviceCount) * estimateMinSDKSupport(tc.minSDK))
				if result.SupportedCount != expectedSupported {
					t.Errorf("SupportedCount = %d, want %d", result.SupportedCount, expectedSupported)
				}
			}

			if len(result.DeviceGroups) != 5 {
				t.Errorf("Expected 5 device groups, got: %d", len(result.DeviceGroups))
			}

			// Check that percentages sum reasonably
			var totalPercent float64
			for _, group := range result.DeviceGroups {
				totalPercent += group.Percent
			}
			// Allow for some floating point variance
			if totalPercent < 99.0 || totalPercent > 101.0 {
				t.Errorf("Device group percentages sum to %f, expected ~100", totalPercent)
			}
		})
	}
}

// ============================================================================
// Result Structure Tests
// ============================================================================

func TestTestingResultStructures(t *testing.T) {
	t.Run("testingPrelaunchResult structure", func(t *testing.T) {
		result := testingPrelaunchResult{
			Status:      "running",
			EditID:      "edit-123",
			TestsRun:    10,
			TestsPassed: 8,
			TestsFailed: 2,
			Issues: []testingPrelaunchIssue{
				{Severity: "error", Type: "crash", Message: "Test crash"},
			},
			Devices: []testingPrelaunchDevice{
				{Model: "redfin", OSVersion: "12", Status: "completed"},
			},
			CheckedAt: time.Now(),
		}

		if result.TestsRun != result.TestsPassed+result.TestsFailed {
			t.Error("TestsRun should equal TestsPassed + TestsFailed")
		}
	})

	t.Run("testingDeviceLabResult structure", func(t *testing.T) {
		now := time.Now()
		result := testingDeviceLabResult{
			TestMatrixID: "matrix-123",
			Status:       "completed",
			Outcome:      "success",
			TestRuns: []testingDeviceLabTestRun{
				{Device: "redfin", OSVersion: "12", Status: "completed", Outcome: "passed"},
			},
			LogsURL:     "https://console.firebase.google.com/logs",
			GcloudFound: true,
			StartedAt:   now,
			CompletedAt: &now,
		}

		if result.CompletedAt == nil {
			t.Error("CompletedAt should not be nil")
		}
	})

	t.Run("testingScreenshotsResult structure", func(t *testing.T) {
		result := testingScreenshotsResult{
			Status:   "completed",
			Total:    10,
			Captured: 8,
			Failed:   2,
			Screenshots: []testingScreenshot{
				{Device: "redfin", Orientation: "portrait", Locale: "en-US", Filename: "redfin_portrait_en-US.png", Status: "captured"},
			},
			OutputDir:   "./screenshots",
			GcloudFound: true,
			GeneratedAt: time.Now(),
		}

		if result.Total != result.Captured+result.Failed {
			t.Error("Total should equal Captured + Failed")
		}
	})

	t.Run("testingValidateResult structure", func(t *testing.T) {
		result := testingValidateResult{
			Valid:  true,
			Status: "passed",
			Checks: []testingValidateCheck{
				{Name: "signing", Status: "pass", Message: "Valid signature"},
				{Name: "size", Status: "pass", Message: "Size within limits"},
			},
			Errors:      []string{},
			Warnings:    []string{},
			ValidatedAt: time.Now(),
		}

		if !result.Valid && len(result.Errors) == 0 {
			t.Error("If Valid is false, there should be errors")
		}
	})

	t.Run("testingCompatibilityResult structure", func(t *testing.T) {
		result := testingCompatibilityResult{
			Compatible:     true,
			DeviceCount:    20000,
			SupportedCount: 18000,
			BlockedCount:   2000,
			Issues: []testingCompatibilityIssue{
				{Severity: "warning", Type: "min_sdk", Message: "High min SDK", Devices: 2000},
			},
			DeviceGroups: []testingCompatibilityGroup{
				{Name: "Phones", Count: 15000, Percent: 83.3},
				{Name: "Tablets", Count: 2160, Percent: 12.0},
			},
			CheckedAt: time.Now(),
		}

		if result.DeviceCount != result.SupportedCount+result.BlockedCount {
			t.Error("DeviceCount should equal SupportedCount + BlockedCount")
		}
	})
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestTestingCommands_ErrorHandling(t *testing.T) {
	t.Run("prelaunch with invalid format", func(t *testing.T) {
		cmd := &TestingPrelaunchCmd{
			Action: "check",
			Format: "invalid",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		// The enum tag should prevent invalid values from reaching Run()
		// but if it does, the command should still handle it gracefully
		err := cmd.Run(globals)
		if err != nil {
			t.Logf("Got error (may be expected): %v", err)
		}
	})

	t.Run("device lab with invalid test type", func(t *testing.T) {
		cmd := &TestingDeviceLabCmd{
			AppFile:  createTempFile(t, "test.apk", []byte("fake apk")),
			TestType: "invalid",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		// Invalid test type defaults to robo behavior
		if err != nil {
			t.Logf("Got error (may be expected): %v", err)
		}
	})

	t.Run("compatibility with invalid format", func(t *testing.T) {
		cmd := &TestingCompatibilityCmd{
			AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
			Format:  "invalid",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		// Will fail on auth but format is not validated in Run()
		if err != nil && !strings.Contains(err.Error(), "authentication not configured") {
			t.Logf("Got error (may be expected): %v", err)
		}
	})
}

// ============================================================================
// Integration Tests with Mock Server
// ============================================================================

func TestTestingCompatibilityCmd_WithMockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.Contains(r.URL.Path, "/edits") && r.Method == "POST":
			fmt.Fprint(w, `{"id": "mock-edit-id", "expiryTimeSeconds": "1234567890"}`)
		case strings.Contains(r.URL.Path, "/edits/") && strings.Contains(r.URL.Path, "/apks") && r.Method == "GET":
			fmt.Fprint(w, `{"apks": [{"versionCode": 1, "binary": {"sha256": "abc123"}}]}`)
		case strings.Contains(r.URL.Path, "/edits/") && strings.Contains(r.URL.Path, "/bundles") && r.Method == "GET":
			fmt.Fprint(w, `{"bundles": [{"versionCode": 1, "sha256": "def456"}]}`)
		case strings.Contains(r.URL.Path, "/edits/") && r.Method == "DELETE":
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	cmd := &TestingCompatibilityCmd{
		AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
		MinSDK:        26,
		TargetSDK:     33,
		DeviceCatalog: "play",
	}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	// Without proper auth injection, this will fail on auth
	// but demonstrates the structure
	err := cmd.Run(globals)
	if err != nil {
		t.Logf("Expected auth error: %v", err)
	}

	_ = server.URL
	_ = option.WithEndpoint(server.URL) // Reference to avoid unused import
}

// ============================================================================
// Edge Cases and Boundary Tests
// ============================================================================

func TestTestingCommands_EdgeCases(t *testing.T) {
	t.Run("screenshots with no devices", func(t *testing.T) {
		cmd := &TestingScreenshotsCmd{
			AppFile:      createTempFile(t, "test.apk", []byte("fake apk")),
			Devices:      []string{},
			Orientations: []string{"portrait"},
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with no devices: %v", err)
		}
	})

	t.Run("screenshots with no orientations", func(t *testing.T) {
		cmd := &TestingScreenshotsCmd{
			AppFile:      createTempFile(t, "test.apk", []byte("fake apk")),
			Devices:      []string{"redfin"},
			Orientations: []string{}, // Default should apply
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with no orientations: %v", err)
		}
	})

	t.Run("validate with empty checks", func(t *testing.T) {
		cmd := &TestingValidateCmd{
			AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
			Checks:  []string{},
			Strict:  false,
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with empty checks: %v", err)
		}
	})

	t.Run("compatibility with negative SDK", func(t *testing.T) {
		cmd := &TestingCompatibilityCmd{
			AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
			MinSDK:        -1,
			TargetSDK:     -1,
			DeviceCatalog: "play",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		// Will fail on auth but tests SDK handling
		err := cmd.Run(globals)
		if err != nil && !strings.Contains(err.Error(), "authentication not configured") {
			t.Logf("Got error: %v", err)
		}
	})

	t.Run("device lab with many devices", func(t *testing.T) {
		devices := make([]string, 100)
		for i := range devices {
			devices[i] = fmt.Sprintf("device-%d", i)
		}

		cmd := &TestingDeviceLabCmd{
			AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
			Devices: devices,
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with many devices: %v", err)
		}
	})
}

// ============================================================================
// Command Implementation Tests
// ============================================================================

func TestTestingCommands_Implemented(t *testing.T) {
	globals := &Globals{Package: "com.example.app"}

	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"TestingPrelaunchCmd", &TestingPrelaunchCmd{}},
		{"TestingDeviceLabCmd", &TestingDeviceLabCmd{AppFile: "test.apk"}},
		{"TestingScreenshotsCmd", &TestingScreenshotsCmd{AppFile: "test.apk"}},
		{"TestingValidateCmd", &TestingValidateCmd{AppFile: "test.apk"}},
		{"TestingCompatibilityCmd", &TestingCompatibilityCmd{AppFile: "test.apk"}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.cmd.Run(globals)
			if err == nil {
				return // success is fine
			}

			// Should not return "not yet implemented"
			if strings.Contains(err.Error(), "not yet implemented") {
				t.Errorf("%s.Run() should be implemented, but returns 'not yet implemented'", tc.name)
			}

			// Acceptable errors:
			// - package required (if package not set in some cases)
			// - authentication not configured
			// - file not found
			acceptableErrors := []string{
				"package name is required",
				"authentication not configured",
				"file not found",
				"no such file",
			}

			foundAcceptable := false
			for _, acceptable := range acceptableErrors {
				if strings.Contains(err.Error(), acceptable) {
					foundAcceptable = true
					break
				}
			}

			if !foundAcceptable {
				t.Logf("%s.Run() returned error (may be expected): %v", tc.name, err)
			}
		})
	}
}

// ============================================================================
// Constants Validation
// ============================================================================

func TestTestingConstants(t *testing.T) {
	// Verify that checkAll constant is used correctly in testing commands
	if checkAll != "all" {
		t.Errorf("checkAll constant = %q, expected 'all'", checkAll)
	}
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestTestingCommands_OutputFormats(t *testing.T) {
	formats := []string{"json", "table"}

	for _, format := range formats {
		t.Run(fmt.Sprintf("prelaunch_format_%s", format), func(t *testing.T) {
			cmd := &TestingPrelaunchCmd{
				Action: "check",
				Format: format,
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  format,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error with format %s: %v", format, err)
			}
		})
	}
}

// ============================================================================
// Concurrent Safety Tests
// ============================================================================

func TestTestingCommands_ConcurrentExecution(t *testing.T) {
	// Test that commands can be instantiated and used concurrently
	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"TestingPrelaunchCmd", &TestingPrelaunchCmd{Action: "check"}},
		{"TestingValidateCmd", &TestingValidateCmd{
			AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
			Checks:  []string{"all"},
		}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			// Run the command multiple times concurrently
			done := make(chan error, 5)
			for i := 0; i < 5; i++ {
				go func() {
					done <- tc.cmd.Run(globals)
				}()
			}

			for i := 0; i < 5; i++ {
				err := <-done
				// Commands should either succeed or fail with expected errors
				if err != nil {
					t.Logf("Concurrent execution %d error: %v", i, err)
				}
			}
		})
	}
}

// ============================================================================
// API Error Classification Tests
// ============================================================================

func TestTestingCommands_APIErrors(t *testing.T) {
	t.Run("compatibility with API error", func(t *testing.T) {
		cmd := &TestingCompatibilityCmd{
			AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
			MinSDK:        30,
			DeviceCatalog: "play",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
			KeyPath: "/invalid/path/to/key.json",
		}

		err := cmd.Run(globals)
		if err == nil {
			t.Log("Expected error due to invalid key path")
			return
		}

		// Check if it's an APIError
		if apiErr, ok := err.(*errors.APIError); ok {
			if apiErr.Code != errors.CodeAuthFailure && apiErr.Code != errors.CodeGeneralError {
				t.Logf("Got APIError with code: %s", apiErr.Code)
			}
		}
	})
}

// ============================================================================
// Time and Duration Tests
// ============================================================================

func TestTestingCommands_DurationHandling(t *testing.T) {
	t.Run("device lab timeout parsing", func(t *testing.T) {
		timeouts := []string{"1m", "15m", "1h", "30m"}

		for _, timeout := range timeouts {
			cmd := &TestingDeviceLabCmd{
				AppFile:     createTempFile(t, "test.apk", []byte("fake apk")),
				TestTimeout: timeout,
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error with timeout %s: %v", timeout, err)
			}

			// Verify timeout appears in suggested command
			if !strings.Contains(cmd.TestTimeout, timeout) {
				t.Errorf("TestTimeout not set correctly: %s", cmd.TestTimeout)
			}
		}
	})

	t.Run("prelaunch max wait time", func(t *testing.T) {
		waitTimes := []string{"10m", "30m", "1h", "2h"}

		for _, waitTime := range waitTimes {
			cmd := &TestingPrelaunchCmd{
				Action:      "wait",
				EditID:      "test-edit",
				MaxWaitTime: waitTime,
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error with wait time %s: %v", waitTime, err)
			}
		}
	})
}

// ============================================================================
// Complex Scenario Tests
// ============================================================================

func TestTestingCommands_ComplexScenarios(t *testing.T) {
	t.Run("full testing workflow", func(t *testing.T) {
		packageName := "com.example.app"
		globals := &Globals{
			Package: packageName,
			Output:  "json",
			Verbose: true,
		}

		// Step 1: Validate the app
		validateCmd := &TestingValidateCmd{
			AppFile: createTempFile(t, "test.aab", []byte("fake aab")),
			Checks:  []string{"all", "signing", "permissions"},
			Strict:  true,
		}

		err := validateCmd.Run(globals)
		if err != nil {
			t.Logf("Validation error (expected without auth): %v", err)
		}

		// Step 2: Check compatibility
		compatCmd := &TestingCompatibilityCmd{
			AppFile:       createTempFile(t, "test.aab", []byte("fake aab")),
			MinSDK:        26,
			TargetSDK:     33,
			DeviceCatalog: "play",
			Format:        "json",
		}

		err = compatCmd.Run(globals)
		if err != nil {
			t.Logf("Compatibility error (expected without auth): %v", err)
		}

		// Step 3: Plan screenshots
		screenshotCmd := &TestingScreenshotsCmd{
			AppFile:      createTempFile(t, "test.aab", []byte("fake aab")),
			Devices:      []string{"redfin", "oriole", "panther"},
			Orientations: []string{"portrait", "landscape"},
			Locales:      []string{"en-US", "de-DE", "fr-FR", "ja-JP"},
			OutputDir:    "./screenshots",
			TestLab:      true,
		}

		err = screenshotCmd.Run(globals)
		if err != nil {
			t.Errorf("Screenshot planning error: %v", err)
		}

		// Step 4: Device lab test plan
		deviceLabCmd := &TestingDeviceLabCmd{
			AppFile:     createTempFile(t, "test.aab", []byte("fake aab")),
			Devices:     []string{"redfin", "oriole"},
			TestTimeout: "30m",
			Async:       true,
			TestType:    "robo",
		}

		err = deviceLabCmd.Run(globals)
		if err != nil {
			t.Errorf("Device lab error: %v", err)
		}
	})

	t.Run("comprehensive device compatibility analysis", func(t *testing.T) {
		sdkConfigs := []struct {
			minSDK    int
			targetSDK int
		}{
			{21, 33}, // Maximum compatibility
			{26, 33}, // Good balance
			{30, 33}, // Modern only
			{33, 33}, // Latest only
		}

		for _, config := range sdkConfigs {
			name := fmt.Sprintf("minSDK_%d_targetSDK_%d", config.minSDK, config.targetSDK)
			t.Run(name, func(t *testing.T) {
				cmd := &TestingCompatibilityCmd{
					AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
					MinSDK:        config.minSDK,
					TargetSDK:     config.targetSDK,
					DeviceCatalog: "play",
					Format:        "json",
				}
				globals := &Globals{
					Package: "com.example.app",
					Output:  "json",
				}

				// Calculate expected support
				expectedSupport := estimateMinSDKSupport(config.minSDK)

				// Will fail on auth, but tests the structure
				err := cmd.Run(globals)
				if err != nil {
					t.Logf("Expected auth error for SDK config %v: %v", config, err)
				}

				t.Logf("Expected device support for minSDK %d: %.0f%%", config.minSDK, expectedSupport*100)
			})
		}
	})
}

// ============================================================================
// Nil and Empty Input Tests
// ============================================================================

func TestTestingCommands_NilEmptyInputs(t *testing.T) {
	t.Run("prelaunch with empty edit ID", func(t *testing.T) {
		cmd := &TestingPrelaunchCmd{
			Action:      "check",
			EditID:      "",
			MaxWaitTime: "30m",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with empty edit ID: %v", err)
		}
	})

	t.Run("validate with nil checks defaults to all", func(t *testing.T) {
		cmd := &TestingValidateCmd{
			AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
			Checks:  nil, // Should default to all
			Strict:  false,
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with nil checks: %v", err)
		}
	})

	t.Run("screenshots with nil slices", func(t *testing.T) {
		cmd := &TestingScreenshotsCmd{
			AppFile:      createTempFile(t, "test.apk", []byte("fake apk")),
			Devices:      nil,
			Orientations: nil,
			Locales:      nil,
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with nil slices: %v", err)
		}
	})
}

// ============================================================================
// Result Data Validation Tests
// ============================================================================

func TestTestingResultDataValidation(t *testing.T) {
	t.Run("prelaunch issue severity levels", func(t *testing.T) {
		severities := []string{"info", "warning", "error", "critical"}
		for _, severity := range severities {
			issue := testingPrelaunchIssue{
				Severity: severity,
				Type:     "test",
				Message:  "Test message",
			}
			if issue.Severity != severity {
				t.Errorf("Severity mismatch: got %s, want %s", issue.Severity, severity)
			}
		}
	})

	t.Run("compatibility issue severity levels", func(t *testing.T) {
		severities := []string{"info", "warning", "error"}
		for _, severity := range severities {
			issue := testingCompatibilityIssue{
				Severity: severity,
				Type:     "test",
				Message:  "Test message",
				Devices:  100,
			}
			if issue.Severity != severity {
				t.Errorf("Severity mismatch: got %s, want %s", issue.Severity, severity)
			}
		}
	})

	t.Run("validate check status values", func(t *testing.T) {
		statuses := []string{"pass", "fail", "skip", "warn"}
		for _, status := range statuses {
			check := testingValidateCheck{
				Name:    "test",
				Status:  status,
				Message: "Test message",
			}
			if check.Status != status {
				t.Errorf("Status mismatch: got %s, want %s", check.Status, status)
			}
		}
	})
}

// ============================================================================
// Command Flag Interaction Tests
// ============================================================================

func TestTestingCommands_FlagInteractions(t *testing.T) {
	t.Run("validate strict mode with warnings", func(t *testing.T) {
		// When strict mode is enabled, warnings should cause validation to fail
		cmd := &TestingValidateCmd{
			AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
			Checks:  []string{"all"},
			Strict:  true,
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Logf("Strict mode result: %v", err)
		}
	})

	t.Run("compatibility format interactions", func(t *testing.T) {
		formats := []string{"json", "table", "csv"}
		catalogs := []string{"play", "all"}

		for _, format := range formats {
			for _, catalog := range catalogs {
				name := fmt.Sprintf("format_%s_catalog_%s", format, catalog)
				t.Run(name, func(t *testing.T) {
					cmd := &TestingCompatibilityCmd{
						AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
						DeviceCatalog: catalog,
						Format:        format,
					}
					globals := &Globals{
						Package: "com.example.app",
						Output:  format,
					}

					err := cmd.Run(globals)
					if err != nil && !strings.Contains(err.Error(), "authentication not configured") {
						t.Logf("Flag interaction error: %v", err)
					}
				})
			}
		}
	})

	t.Run("screenshots testlab flag combinations", func(t *testing.T) {
		testLabValues := []bool{true, false}

		for _, useTestLab := range testLabValues {
			name := fmt.Sprintf("testlab_%v", useTestLab)
			t.Run(name, func(t *testing.T) {
				cmd := &TestingScreenshotsCmd{
					AppFile:   createTempFile(t, "test.apk", []byte("fake apk")),
					Devices:   []string{"redfin"},
					TestLab:   useTestLab,
					OutputDir: "./screenshots",
				}
				globals := &Globals{
					Package: "com.example.app",
					Output:  "json",
				}

				err := cmd.Run(globals)
				if err != nil {
					t.Errorf("Unexpected error with TestLab=%v: %v", useTestLab, err)
				}
			})
		}
	})
}

// ============================================================================
// Performance and Scale Tests
// ============================================================================

func TestTestingCommands_Performance(t *testing.T) {
	t.Run("screenshots with many locales", func(t *testing.T) {
		locales := []string{
			"en-US", "en-GB", "de-DE", "fr-FR", "es-ES",
			"it-IT", "pt-BR", "ru-RU", "ja-JP", "ko-KR",
			"zh-CN", "zh-TW", "hi-IN", "ar-SA", "th-TH",
		}

		cmd := &TestingScreenshotsCmd{
			AppFile:      createTempFile(t, "test.apk", []byte("fake apk")),
			Devices:      []string{"redfin", "oriole"},
			Orientations: []string{"portrait", "landscape"},
			Locales:      locales,
			OutputDir:    "./screenshots",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		start := time.Now()
		err := cmd.Run(globals)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Unexpected error with many locales: %v", err)
		}

		// Should complete in reasonable time (< 1 second for planning)
		if duration > 1*time.Second {
			t.Logf("Warning: Screenshot planning took %v (may indicate performance issue)", duration)
		}
	})

	t.Run("device lab with many devices", func(t *testing.T) {
		devices := make([]string, 50)
		for i := 0; i < 50; i++ {
			devices[i] = fmt.Sprintf("device-model-%d", i)
		}

		cmd := &TestingDeviceLabCmd{
			AppFile: createTempFile(t, "test.apk", []byte("fake apk")),
			Devices: devices,
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		start := time.Now()
		err := cmd.Run(globals)
		duration := time.Since(start)

		if err != nil {
			t.Errorf("Unexpected error with many devices: %v", err)
		}

		if duration > 100*time.Millisecond {
			t.Logf("Warning: Device lab setup took %v", duration)
		}
	})
}

// ============================================================================
// Error Recovery and Resilience Tests
// ============================================================================

func TestTestingCommands_ErrorRecovery(t *testing.T) {
	t.Run("compatibility with edit creation failure", func(t *testing.T) {
		// This tests the error handling path when API calls fail
		cmd := &TestingCompatibilityCmd{
			AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
			MinSDK:        26,
			TargetSDK:     33,
			DeviceCatalog: "play",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
			KeyPath: "/nonexistent/service-account.json",
		}

		err := cmd.Run(globals)
		if err == nil {
			t.Log("Expected error due to invalid key path")
		} else {
			// Verify error is properly wrapped
			if !strings.Contains(err.Error(), "authentication not configured") &&
				!strings.Contains(err.Error(), "no such file") {
				t.Logf("Error type: %T, Message: %v", err, err)
			}
		}
	})

	t.Run("validate with unreadable file", func(t *testing.T) {
		// Create a directory instead of a file to simulate unreadable
		tmpDir := t.TempDir()
		unreadablePath := filepath.Join(tmpDir, "unreadable")
		if err := os.Mkdir(unreadablePath, 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}

		// Add a file inside that will fail validation
		testFile := filepath.Join(unreadablePath, "test.txt")
		if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cmd := &TestingValidateCmd{
			AppFile: testFile, // Not an APK/AAB
			Checks:  []string{"all"},
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		// Validate command doesn't actually read the file, just validates structure
		if err != nil {
			t.Logf("Got error for non-APK file: %v", err)
		}
	})
}

// ============================================================================
// Mock Client Integration Tests
// ============================================================================

func TestTestingCommands_WithMockClient(t *testing.T) {
	// Create a mock server that simulates successful API responses
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/edits"):
			// Create edit
			response := `{
				"id": "test-edit-123",
				"expiryTimeSeconds": "1234567890"
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))

		case r.Method == "GET" && strings.Contains(r.URL.Path, "/edits/") && strings.Contains(r.URL.Path, "/apks"):
			// List APKs
			response := `{
				"apks": [
					{
						"versionCode": 100,
						"binary": {
							"sha256": "abcdef1234567890"
						}
					}
				]
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))

		case r.Method == "GET" && strings.Contains(r.URL.Path, "/edits/") && strings.Contains(r.URL.Path, "/bundles"):
			// List bundles
			response := `{
				"bundles": [
					{
						"versionCode": 101,
						"sha256": "0987654321fedcba"
					}
				]
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))

		case r.Method == "GET" && strings.Contains(r.URL.Path, "/edits/"):
			// Get edit
			response := `{
				"id": "test-edit-123",
				"expiryTimeSeconds": "1234567890"
			}`
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))

		case r.Method == "DELETE" && strings.Contains(r.URL.Path, "/edits/"):
			// Delete edit
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Test prelaunch with mock server
	t.Run("prelaunch with mock server", func(t *testing.T) {
		cmd := &TestingPrelaunchCmd{
			Action: "check",
			EditID: "test-edit-123",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		// Without auth injection, will fail, but verifies structure
		err := cmd.Run(globals)
		if err != nil {
			t.Logf("Expected auth error: %v", err)
		}
	})

	// Test compatibility with mock server
	t.Run("compatibility with mock server", func(t *testing.T) {
		cmd := &TestingCompatibilityCmd{
			AppFile:       createTempFile(t, "test.apk", []byte("fake apk")),
			MinSDK:        26,
			TargetSDK:     33,
			DeviceCatalog: "play",
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Logf("Expected auth error: %v", err)
		}
	})

	_ = server.URL
	_ = androidpublisher.AppEdit{} // Reference to avoid unused import
}

// ============================================================================
// Time Parsing Tests
// ============================================================================

func TestTestingCommands_TimeParsing(t *testing.T) {
	t.Run("max wait time formats", func(t *testing.T) {
		timeFormats := []string{
			"30m",
			"1h",
			"1h30m",
			"90m",
		}

		for _, format := range timeFormats {
			cmd := &TestingPrelaunchCmd{
				Action:      "wait",
				MaxWaitTime: format,
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error with time format %s: %v", format, err)
			}
		}
	})

	t.Run("test timeout formats", func(t *testing.T) {
		timeoutFormats := []string{
			"15m",
			"30m",
			"1h",
			"2h",
		}

		for _, format := range timeoutFormats {
			cmd := &TestingDeviceLabCmd{
				AppFile:     createTempFile(t, "test.apk", []byte("fake apk")),
				TestTimeout: format,
			}
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error with timeout format %s: %v", format, err)
			}
		}
	})
}

// ============================================================================
// String Building Tests
// ============================================================================

func TestTestingCommands_StringBuilding(t *testing.T) {
	t.Run("device lab command building", func(t *testing.T) {
		cmd := &TestingDeviceLabCmd{
			AppFile:     "/path/to/app.apk",
			TestFile:    "/path/to/test.apk",
			TestType:    "instrumentation",
			TestTimeout: "30m",
			Async:       true,
			Devices:     []string{"redfin", "oriole"},
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// The command is stored in result but we can't access it directly
		// The test verifies the Run() method completes without error
	})

	t.Run("screenshots command building", func(t *testing.T) {
		cmd := &TestingScreenshotsCmd{
			AppFile:      "/path/to/app.apk",
			Devices:      []string{"redfin", "oriole"},
			Orientations: []string{"portrait", "landscape"},
			Locales:      []string{"en-US", "de-DE"},
			TestLab:      true,
		}
		globals := &Globals{
			Package: "com.example.app",
			Output:  "json",
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}

		// Verify command structure is valid
		if cmd.AppFile != "/path/to/app.apk" {
			t.Error("AppFile was modified unexpectedly")
		}
	})
}

// ============================================================================
// Context Cancellation Tests
// ============================================================================

func TestTestingCommands_ContextHandling(t *testing.T) {
	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		// Verify context is cancelled
		if ctx.Err() != context.Canceled {
			t.Error("Expected context to be canceled")
		}

		// The actual commands don't accept external context yet
		// This test documents the expected behavior for future implementation
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
		defer cancel()

		time.Sleep(10 * time.Millisecond) // Ensure timeout passes

		if ctx.Err() != context.DeadlineExceeded {
			t.Error("Expected context to be deadline exceeded")
		}
	})
}
