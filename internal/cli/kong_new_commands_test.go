package cli

import (
	"reflect"
	"strings"
	"testing"
)

// ============================================================================
// Bulk Commands Tests
// ============================================================================

func TestBulkCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := BulkCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Upload", "Listings", "Images", "Tracks"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("BulkCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("BulkCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("BulkCmd.%s should have help tag", name)
		}
	}
}

func TestBulkUploadCmd_FieldTags(t *testing.T) {
	cmd := BulkUploadCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
		default_  string
	}{
		{"Track", "internal,alpha,beta,production", "internal"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("BulkUploadCmd missing field: %s", tc.fieldName)
			continue
		}

		enumTag := field.Tag.Get("enum")
		if enumTag != tc.enum {
			t.Errorf("BulkUploadCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
		}

		if tc.default_ != "" {
			defaultTag := field.Tag.Get("default")
			if defaultTag != tc.default_ {
				t.Errorf("BulkUploadCmd.%s default tag = %q, want %q", tc.fieldName, defaultTag, tc.default_)
			}
		}
	}
}

func TestBulkTracksCmd_FieldTags(t *testing.T) {
	cmd := BulkTracksCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Status")
	if !ok {
		t.Fatal("BulkTracksCmd missing Status field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "draft,completed,halted,inProgress"
	if enumTag != expected {
		t.Errorf("BulkTracksCmd.Status enum tag = %q, want %q", enumTag, expected)
	}
}

// ============================================================================
// Compare Commands Tests
// ============================================================================

func TestCompareCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := CompareCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Vitals", "Reviews", "Releases", "Subscriptions"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("CompareCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("CompareCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("CompareCmd.%s should have help tag", name)
		}
	}
}

func TestCompareVitalsCmd_FieldTags(t *testing.T) {
	cmd := CompareVitalsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Metric")
	if !ok {
		t.Fatal("CompareVitalsCmd missing Metric field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "crash-rate,anr-rate,error-rate,all"
	if enumTag != expected {
		t.Errorf("CompareVitalsCmd.Metric enum tag = %q, want %q", enumTag, expected)
	}

	defaultTag := field.Tag.Get("default")
	if defaultTag != "all" {
		t.Errorf("CompareVitalsCmd.Metric default tag = %q, want \"all\"", defaultTag)
	}
}

// ============================================================================
// Automation Commands Tests
// ============================================================================

func TestAutomationCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := AutomationCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"ReleaseNotes", "Rollout", "Promote", "Validate", "Monitor"}

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
}

func TestAutomationReleaseNotesCmd_FieldTags(t *testing.T) {
	cmd := AutomationReleaseNotesCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		enum      string
	}{
		{"Source", "git,pr,file"},
		{"Format", "json,markdown"},
	}

	for _, tc := range tests {
		field, ok := typeOfCmd.FieldByName(tc.fieldName)
		if !ok {
			t.Errorf("AutomationReleaseNotesCmd missing field: %s", tc.fieldName)
			continue
		}

		enumTag := field.Tag.Get("enum")
		if enumTag != tc.enum {
			t.Errorf("AutomationReleaseNotesCmd.%s enum tag = %q, want %q", tc.fieldName, enumTag, tc.enum)
		}
	}
}

// ============================================================================
// Monitor Commands Tests
// ============================================================================

func TestMonitorCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := MonitorCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Watch", "Anomalies", "Dashboard", "Report", "Webhooks"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("MonitorCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("MonitorCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("MonitorCmd.%s should have help tag", name)
		}
	}
}

func TestMonitorWatchCmd_FieldTags(t *testing.T) {
	cmd := MonitorWatchCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Format")
	if !ok {
		t.Fatal("MonitorWatchCmd missing Format field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "json,table,html"
	if enumTag != expected {
		t.Errorf("MonitorWatchCmd.Format enum tag = %q, want %q", enumTag, expected)
	}
}

// ============================================================================
// Testing Commands Tests
// ============================================================================

func TestTestingCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := TestingCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Prelaunch", "DeviceLab", "Screenshots", "Validate", "Compatibility"}

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
}

func TestTestingPrelaunchCmd_FieldTags(t *testing.T) {
	cmd := TestingPrelaunchCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Action")
	if !ok {
		t.Fatal("TestingPrelaunchCmd missing Action field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "trigger,check,wait"
	if enumTag != expected {
		t.Errorf("TestingPrelaunchCmd.Action enum tag = %q, want %q", enumTag, expected)
	}

	defaultTag := field.Tag.Get("default")
	if defaultTag != "check" {
		t.Errorf("TestingPrelaunchCmd.Action default tag = %q, want \"check\"", defaultTag)
	}
}

// ============================================================================
// Release Management Commands Tests
// ============================================================================

func TestReleaseMgmtCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := ReleaseMgmtCmd{}
	v := reflect.ValueOf(cmd)
	typeOfCmd := v.Type()

	expectedSubcommands := []string{"Calendar", "Conflicts", "Strategy", "History", "Notes"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("ReleaseMgmtCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("ReleaseMgmtCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("ReleaseMgmtCmd.%s should have help tag", name)
		}
	}
}

func TestReleaseCalendarCmd_FieldTags(t *testing.T) {
	cmd := ReleaseCalendarCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Track")
	if !ok {
		t.Fatal("ReleaseCalendarCmd missing Track field")
	}

	enumTag := field.Tag.Get("enum")
	expected := "internal,alpha,beta,production,all"
	if enumTag != expected {
		t.Errorf("ReleaseCalendarCmd.Track enum tag = %q, want %q", enumTag, expected)
	}
}

func TestReleaseNotesCmd_FieldTags(t *testing.T) {
	cmd := ReleaseNotesCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Action")
	if !ok {
		t.Fatal("ReleaseNotesCmd missing Action field")
	}

	// Check that the required tag exists (it will have empty value "")
	if !strings.Contains(string(field.Tag), "required") {
		t.Error("ReleaseNotesCmd.Action should have required tag")
	}

	enumTag := field.Tag.Get("enum")
	expected := "get,set,copy,list"
	if enumTag != expected {
		t.Errorf("ReleaseNotesCmd.Action enum tag = %q, want %q", enumTag, expected)
	}
}
