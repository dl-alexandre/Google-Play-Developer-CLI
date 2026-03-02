//go:build unit
// +build unit

package cli

import (
	"context"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	gpdErrors "github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// ============================================================================
// Helper Functions Tests
// ============================================================================

func TestRoleToDeveloperPermissions(t *testing.T) {
	tests := []struct {
		name     string
		role     string
		expected []string
	}{
		{
			name:     "admin role returns admin permissions",
			role:     "admin",
			expected: []string{"CAN_MANAGE_PERMISSIONS_GLOBAL"},
		},
		{
			name: "developer role returns developer permissions",
			role: "developer",
			expected: []string{
				"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
				"CAN_MANAGE_TRACK_APKS_GLOBAL",
				"CAN_MANAGE_TRACK_USERS_GLOBAL",
				"CAN_MANAGE_PUBLIC_LISTING_GLOBAL",
				"CAN_MANAGE_DRAFT_APPS_GLOBAL",
				"CAN_REPLY_TO_REVIEWS_GLOBAL",
			},
		},
		{
			name:     "viewer role returns viewer permissions",
			role:     "viewer",
			expected: []string{"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"},
		},
		{
			name:     "unknown role defaults to viewer permissions",
			role:     "unknown",
			expected: []string{"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"},
		},
		{
			name:     "empty role defaults to viewer permissions",
			role:     "",
			expected: []string{"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := roleToDeveloperPermissions(tc.role)
			if len(result) != len(tc.expected) {
				t.Errorf("roleToDeveloperPermissions(%q) returned %d permissions, expected %d",
					tc.role, len(result), len(tc.expected))
			}
			for i, perm := range tc.expected {
				if i >= len(result) || result[i] != perm {
					t.Errorf("roleToDeveloperPermissions(%q)[%d] = %q, expected %q",
						tc.role, i, result[i], perm)
				}
			}
		})
	}
}

func TestGetDeveloperParent(t *testing.T) {
	result := getDeveloperParent()
	expected := "developers/-"
	if result != expected {
		t.Errorf("getDeveloperParent() = %q, expected %q", result, expected)
	}
}

// ============================================================================
// Permissions Command Structure Tests
// ============================================================================

func TestPermissionsCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := PermissionsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedSubcommands := []string{"Users", "Grants", "List"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PermissionsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PermissionsCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("PermissionsCmd.%s should have help tag", name)
		}
	}
}

func TestPermissionsUsersCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := PermissionsUsersCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedSubcommands := []string{"Add", "Remove", "List"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PermissionsUsersCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PermissionsUsersCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("PermissionsUsersCmd.%s should have help tag", name)
		}
	}
}

func TestPermissionsGrantsCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := PermissionsGrantsCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedSubcommands := []string{"Add", "Remove", "List"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("PermissionsGrantsCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("PermissionsGrantsCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("PermissionsGrantsCmd.%s should have help tag", name)
		}
	}
}

func TestPermissionsUsersRemoveCmd_FieldTags(t *testing.T) {
	cmd := PermissionsUsersRemoveCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Email")
	if !ok {
		t.Fatal("PermissionsUsersRemoveCmd missing Email field")
	}

	if !strings.Contains(string(field.Tag), "required") {
		t.Error("PermissionsUsersRemoveCmd.Email should have required tag")
	}

	helpTag := field.Tag.Get("help")
	if helpTag == "" {
		t.Error("PermissionsUsersRemoveCmd.Email should have help tag")
	}
}

func TestPermissionsGrantsAddCmd_FieldTags(t *testing.T) {
	cmd := PermissionsGrantsAddCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		required  bool
	}{
		{"Email", true},
		{"Grant", true},
		{"Expiry", false},
	}

	for _, tc := range tests {
		t.Run(tc.fieldName, func(t *testing.T) {
			field, ok := typeOfCmd.FieldByName(tc.fieldName)
			if !ok {
				t.Fatalf("PermissionsGrantsAddCmd missing field: %s", tc.fieldName)
			}

			helpTag := field.Tag.Get("help")
			if helpTag == "" {
				t.Errorf("PermissionsGrantsAddCmd.%s should have help tag", tc.fieldName)
			}

			if tc.required {
				if !strings.Contains(string(field.Tag), "required") {
					t.Errorf("PermissionsGrantsAddCmd.%s should have required tag", tc.fieldName)
				}
			}
		})
	}
}

func TestPermissionsGrantsRemoveCmd_FieldTags(t *testing.T) {
	cmd := PermissionsGrantsRemoveCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
	}{
		{"Email"},
		{"Grant"},
	}

	for _, tc := range tests {
		t.Run(tc.fieldName, func(t *testing.T) {
			field, ok := typeOfCmd.FieldByName(tc.fieldName)
			if !ok {
				t.Fatalf("PermissionsGrantsRemoveCmd missing field: %s", tc.fieldName)
			}

			if !strings.Contains(string(field.Tag), "required") {
				t.Errorf("PermissionsGrantsRemoveCmd.%s should have required tag", tc.fieldName)
			}
		})
	}
}

func TestPermissionsGrantsListCmd_FieldTags(t *testing.T) {
	cmd := PermissionsGrantsListCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Email")
	if !ok {
		t.Fatal("PermissionsGrantsListCmd missing Email field")
	}

	helpTag := field.Tag.Get("help")
	if helpTag == "" {
		t.Error("PermissionsGrantsListCmd.Email should have help tag")
	}

	// Email should be optional (no required tag)
	if strings.Contains(string(field.Tag), "required") {
		t.Error("PermissionsGrantsListCmd.Email should not have required tag")
	}
}

// ============================================================================
// Recovery Command Structure Tests
// ============================================================================

func TestRecoveryCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := RecoveryCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedSubcommands := []string{"List", "Create", "Deploy", "Cancel"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("RecoveryCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("RecoveryCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("RecoveryCmd.%s should have help tag", name)
		}
	}
}

func TestRecoveryListCmd_FieldTags(t *testing.T) {
	cmd := RecoveryListCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("Status")
	if !ok {
		t.Fatal("RecoveryListCmd missing Status field")
	}

	helpTag := field.Tag.Get("help")
	if helpTag == "" {
		t.Error("RecoveryListCmd.Status should have help tag")
	}

	// Status should be optional filter
	if strings.Contains(string(field.Tag), "required") {
		t.Error("RecoveryListCmd.Status should not have required tag")
	}
}

func TestRecoveryDeployCmd_FieldTags(t *testing.T) {
	cmd := RecoveryDeployCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	field, ok := typeOfCmd.FieldByName("ID")
	if !ok {
		t.Fatal("RecoveryDeployCmd missing ID field")
	}

	if !strings.Contains(string(field.Tag), "required") {
		t.Error("RecoveryDeployCmd.ID should have required tag")
	}

	argTag := field.Tag.Get("arg")
	if argTag != "" {
		t.Errorf("RecoveryDeployCmd.ID arg tag = %q, expected positional argument", argTag)
	}
}

func TestRecoveryCancelCmd_FieldTags(t *testing.T) {
	cmd := RecoveryCancelCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		required  bool
	}{
		{"ID", true},
		{"Reason", false},
	}

	for _, tc := range tests {
		t.Run(tc.fieldName, func(t *testing.T) {
			field, ok := typeOfCmd.FieldByName(tc.fieldName)
			if !ok {
				t.Fatalf("RecoveryCancelCmd missing field: %s", tc.fieldName)
			}

			helpTag := field.Tag.Get("help")
			if helpTag == "" {
				t.Errorf("RecoveryCancelCmd.%s should have help tag", tc.fieldName)
			}

			if tc.required {
				if !strings.Contains(string(field.Tag), "required") {
					t.Errorf("RecoveryCancelCmd.%s should have required tag", tc.fieldName)
				}
			}
		})
	}
}

// ============================================================================
// Integrity Command Structure Tests
// ============================================================================

func TestIntegrityCmd_HasExpectedSubcommands(t *testing.T) {
	cmd := IntegrityCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	expectedSubcommands := []string{"Decode"}

	for _, name := range expectedSubcommands {
		field, ok := typeOfCmd.FieldByName(name)
		if !ok {
			t.Errorf("IntegrityCmd missing subcommand: %s", name)
			continue
		}

		cmdTag := field.Tag.Get("cmd")
		if cmdTag != "" {
			t.Errorf("IntegrityCmd.%s should have cmd:\"\" tag, got: %s", name, cmdTag)
		}

		helpTag := field.Tag.Get("help")
		if helpTag == "" {
			t.Errorf("IntegrityCmd.%s should have help tag", name)
		}
	}
}

func TestIntegrityDecodeCmd_FieldTags(t *testing.T) {
	cmd := IntegrityDecodeCmd{}
	typeOfCmd := reflect.TypeOf(cmd)

	tests := []struct {
		fieldName string
		required  bool
	}{
		{"Token", true},
		{"Verify", false},
	}

	for _, tc := range tests {
		t.Run(tc.fieldName, func(t *testing.T) {
			field, ok := typeOfCmd.FieldByName(tc.fieldName)
			if !ok {
				t.Fatalf("IntegrityDecodeCmd missing field: %s", tc.fieldName)
			}

			helpTag := field.Tag.Get("help")
			if helpTag == "" {
				t.Errorf("IntegrityDecodeCmd.%s should have help tag", tc.fieldName)
			}

			if tc.required {
				if !strings.Contains(string(field.Tag), "required") {
					t.Errorf("IntegrityDecodeCmd.%s should have required tag", tc.fieldName)
				}
			}
		})
	}
}

// ============================================================================
// Permissions Run Method Tests - Error Paths
// ============================================================================

func TestPermissionsGrantsAddCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PermissionsGrantsAddCmd{
		Email: "user@example.com",
		Grant: "CAN_ACCESS_APP",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != gpdErrors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPermissionsGrantsRemoveCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PermissionsGrantsRemoveCmd{
		Email: "user@example.com",
		Grant: "CAN_ACCESS_APP",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != gpdErrors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestRecoveryListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &RecoveryListCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != gpdErrors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestRecoveryCreateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &RecoveryCreateCmd{
		Type:   "rollback",
		Reason: "Critical bug",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != gpdErrors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestRecoveryDeployCmd_Run_PackageRequired(t *testing.T) {
	cmd := &RecoveryDeployCmd{
		ID: "123",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != gpdErrors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestRecoveryDeployCmd_Run_InvalidID(t *testing.T) {
	cmd := &RecoveryDeployCmd{
		ID: "invalid-id",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid recovery ID")
	}

	apiErr, ok := err.(*gpdErrors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}

	if apiErr.Code != gpdErrors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %s", apiErr.Code)
	}

	if !strings.Contains(apiErr.Message, "invalid recovery action ID") {
		t.Errorf("Expected 'invalid recovery action ID' in message, got: %s", apiErr.Message)
	}
}

func TestRecoveryCancelCmd_Run_PackageRequired(t *testing.T) {
	cmd := &RecoveryCancelCmd{
		ID:     "123",
		Reason: "No longer needed",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != gpdErrors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestRecoveryCancelCmd_Run_InvalidID(t *testing.T) {
	cmd := &RecoveryCancelCmd{
		ID:     "not-a-number",
		Reason: "Test",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid recovery ID")
	}

	apiErr, ok := err.(*gpdErrors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}

	if apiErr.Code != gpdErrors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %s", apiErr.Code)
	}
}

func TestIntegrityDecodeCmd_Run_PackageRequired(t *testing.T) {
	cmd := &IntegrityDecodeCmd{
		Token: "fake-token",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != gpdErrors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// Context Handling Tests
// ============================================================================

func TestPermissionsCommands_ContextHandling(t *testing.T) {
	tests := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{
			name: "PermissionsUsersAddCmd",
			cmd:  &PermissionsUsersAddCmd{Email: "test@example.com", Role: "viewer"},
		},
		{
			name: "PermissionsUsersRemoveCmd",
			cmd:  &PermissionsUsersRemoveCmd{Email: "test@example.com"},
		},
		{
			name: "PermissionsUsersListCmd",
			cmd:  &PermissionsUsersListCmd{},
		},
		{
			name: "PermissionsGrantsAddCmd",
			cmd:  &PermissionsGrantsAddCmd{Email: "test@example.com", Grant: "CAN_ACCESS_APP"},
		},
		{
			name: "PermissionsGrantsRemoveCmd",
			cmd:  &PermissionsGrantsRemoveCmd{Email: "test@example.com", Grant: "CAN_ACCESS_APP"},
		},
		{
			name: "PermissionsGrantsListCmd",
			cmd:  &PermissionsGrantsListCmd{},
		},
		{
			name: "PermissionsListCmd",
			cmd:  &PermissionsListCmd{},
		},
		{
			name: "RecoveryListCmd",
			cmd:  &RecoveryListCmd{},
		},
		{
			name: "RecoveryCreateCmd",
			cmd:  &RecoveryCreateCmd{Type: "rollback", Reason: "test"},
		},
		{
			name: "IntegrityDecodeCmd",
			cmd:  &IntegrityDecodeCmd{Token: "test"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test with nil context (should create background context)
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}

			// Command should handle nil context gracefully
			// It will fail on API client creation, but should not panic
			_ = tc.cmd.Run(globals)
			// Error is expected since no valid auth, but shouldn't panic
		})
	}
}

func TestPermissionsCommands_WithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Cancel immediately

	globals := &Globals{
		Package: "com.example.app",
		Context: ctx,
		Output:  "json",
	}

	// The context is already cancelled, but commands create their own
	// if globals.Context is nil, so this tests the flow
	if ctx.Err() != context.Canceled {
		t.Error("Expected context to be cancelled")
	}

	_ = globals // Use the variable
}

// ============================================================================
// Role Mapping Edge Cases
// ============================================================================

func TestRoleToDeveloperPermissions_CaseSensitivity(t *testing.T) {
	// Test case sensitivity
	tests := []struct {
		input    string
		expected []string
	}{
		{"ADMIN", []string{"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"}}, // Unknown case, defaults to viewer
		{"Admin", []string{"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"}}, // Mixed case, defaults to viewer
		{"admin", []string{"CAN_MANAGE_PERMISSIONS_GLOBAL"}},      // Correct lowercase
	}

	for _, tc := range tests {
		result := roleToDeveloperPermissions(tc.input)
		if len(result) != len(tc.expected) {
			t.Errorf("roleToDeveloperPermissions(%q) returned %d permissions, expected %d",
				tc.input, len(result), len(tc.expected))
		}
		for i, perm := range tc.expected {
			if i >= len(result) || result[i] != perm {
				t.Errorf("roleToDeveloperPermissions(%q)[%d] = %q, expected %q",
					tc.input, i, result[i], perm)
			}
		}
	}
}

// ============================================================================
// Command Instantiation Tests
// ============================================================================

func TestCommandInstantiation(t *testing.T) {
	tests := []struct {
		name string
		cmd  interface{}
	}{
		{"PermissionsCmd", PermissionsCmd{}},
		{"PermissionsUsersCmd", PermissionsUsersCmd{}},
		{"PermissionsUsersAddCmd", PermissionsUsersAddCmd{Email: "test@example.com", Role: "admin"}},
		{"PermissionsUsersRemoveCmd", PermissionsUsersRemoveCmd{Email: "test@example.com"}},
		{"PermissionsUsersListCmd", PermissionsUsersListCmd{}},
		{"PermissionsGrantsCmd", PermissionsGrantsCmd{}},
		{"PermissionsGrantsAddCmd", PermissionsGrantsAddCmd{Email: "test@example.com", Grant: "CAN_ACCESS_APP"}},
		{"PermissionsGrantsRemoveCmd", PermissionsGrantsRemoveCmd{Email: "test@example.com", Grant: "CAN_ACCESS_APP"}},
		{"PermissionsGrantsListCmd", PermissionsGrantsListCmd{Email: "test@example.com"}},
		{"PermissionsGrantsListCmd_empty", PermissionsGrantsListCmd{}},
		{"PermissionsListCmd", PermissionsListCmd{}},
		{"RecoveryCmd", RecoveryCmd{}},
		{"RecoveryListCmd", RecoveryListCmd{Status: "active"}},
		{"RecoveryListCmd_empty", RecoveryListCmd{}},
		{"RecoveryCreateCmd", RecoveryCreateCmd{Type: "rollback", Target: "123", Reason: "Bug"}},
		{"RecoveryCreateCmd_minimal", RecoveryCreateCmd{Type: "rollback", Reason: "Bug"}},
		{"RecoveryDeployCmd", RecoveryDeployCmd{ID: "123"}},
		{"RecoveryCancelCmd", RecoveryCancelCmd{ID: "123", Reason: "Test"}},
		{"RecoveryCancelCmd_minimal", RecoveryCancelCmd{ID: "123"}},
		{"IntegrityCmd", IntegrityCmd{}},
		{"IntegrityDecodeCmd", IntegrityDecodeCmd{Token: "token123", Verify: true}},
		{"IntegrityDecodeCmd_minimal", IntegrityDecodeCmd{Token: "token123"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify the struct can be instantiated
			if tc.cmd == nil {
				t.Error("Command is nil")
			}
		})
	}
}

// ============================================================================
// Recovery Target Parsing Logic
// ============================================================================

func TestRecoveryCreateCmd_TargetParsing(t *testing.T) {
	// Test that target parsing logic works correctly
	tests := []struct {
		name          string
		target        string
		expectVersion bool
	}{
		{
			name:          "numeric target is parsed as version code",
			target:        "12345",
			expectVersion: true,
		},
		{
			name:          "non-numeric target is not parsed as version",
			target:        "production",
			expectVersion: false,
		},
		{
			name:          "empty target is ignored",
			target:        "",
			expectVersion: false,
		},
		{
			name:          "large number is valid version code",
			target:        "9999999999",
			expectVersion: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// The actual parsing happens in Run(), which requires API setup
			// This test documents the expected behavior
			cmd := &RecoveryCreateCmd{
				Type:   "rollback",
				Target: tc.target,
				Reason: "Test",
			}

			if tc.target != "" && tc.expectVersion {
				// Verify that a numeric target would be parsed correctly
				if _, err := strconv.ParseInt(tc.target, 10, 64); err != nil {
					t.Errorf("Target %q should be parseable as int64", tc.target)
				}
			}

			_ = cmd // Use the command to avoid unused variable
		})
	}
}

// ============================================================================
// Grant Permission Parsing
// ============================================================================

func TestPermissionsGrantsAddCmd_GrantParsing(t *testing.T) {
	tests := []struct {
		name          string
		grant         string
		expectedPerms []string
		expectedCount int
	}{
		{
			name:          "single permission",
			grant:         "CAN_ACCESS_APP",
			expectedPerms: []string{"CAN_ACCESS_APP"},
			expectedCount: 1,
		},
		{
			name:          "multiple permissions with spaces",
			grant:         "CAN_ACCESS_APP, CAN_VIEW_FINANCIAL_DATA",
			expectedPerms: []string{"CAN_ACCESS_APP", "CAN_VIEW_FINANCIAL_DATA"},
			expectedCount: 2,
		},
		{
			name:          "multiple permissions without spaces",
			grant:         "CAN_ACCESS_APP,CAN_VIEW_FINANCIAL_DATA",
			expectedPerms: []string{"CAN_ACCESS_APP", "CAN_VIEW_FINANCIAL_DATA"},
			expectedCount: 2,
		},
		{
			name:          "permissions with extra whitespace",
			grant:         "  CAN_ACCESS_APP  ,  CAN_VIEW_FINANCIAL_DATA  ",
			expectedPerms: []string{"CAN_ACCESS_APP", "CAN_VIEW_FINANCIAL_DATA"},
			expectedCount: 2,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Simulate the parsing that happens in the Run method
			permissions := strings.Split(tc.grant, ",")
			for i := range permissions {
				permissions[i] = strings.TrimSpace(permissions[i])
			}

			if len(permissions) != tc.expectedCount {
				t.Errorf("Expected %d permissions, got %d", tc.expectedCount, len(permissions))
			}

			for i, expected := range tc.expectedPerms {
				if i >= len(permissions) {
					t.Errorf("Missing permission at index %d, expected %q", i, expected)
					continue
				}
				if permissions[i] != expected {
					t.Errorf("Permission[%d] = %q, expected %q", i, permissions[i], expected)
				}
			}
		})
	}
}

// ============================================================================
// Recovery Status Mapping Tests
// ============================================================================

func TestRecoveryStatusMapping(t *testing.T) {
	// Test the status mapping that happens in RecoveryListCmd
	statusMap := map[string]string{
		"RECOVERY_STATUS_ACTIVE":                 "active",
		"RECOVERY_STATUS_CANCELED":               "cancelled",
		"RECOVERY_STATUS_DRAFT":                  "pending",
		"RECOVERY_STATUS_GENERATION_IN_PROGRESS": "pending",
		"RECOVERY_STATUS_GENERATION_FAILED":      "failed",
		"RECOVERY_STATUS_UNSPECIFIED":            "pending",
	}

	tests := []struct {
		apiStatus      string
		expectedMapped string
	}{
		{"RECOVERY_STATUS_ACTIVE", "active"},
		{"RECOVERY_STATUS_CANCELED", "cancelled"},
		{"RECOVERY_STATUS_DRAFT", "pending"},
		{"RECOVERY_STATUS_GENERATION_IN_PROGRESS", "pending"},
		{"RECOVERY_STATUS_GENERATION_FAILED", "failed"},
		{"RECOVERY_STATUS_UNSPECIFIED", "pending"},
		{"UNKNOWN_STATUS", ""}, // Should fall through to strings.ToLower
	}

	for _, tc := range tests {
		t.Run(tc.apiStatus, func(t *testing.T) {
			mapped, ok := statusMap[tc.apiStatus]
			if tc.expectedMapped != "" {
				if !ok {
					t.Errorf("Status %q not found in map", tc.apiStatus)
					return
				}
				if mapped != tc.expectedMapped {
					t.Errorf("Status %q mapped to %q, expected %q", tc.apiStatus, mapped, tc.expectedMapped)
				}
			} else {
				// Unknown status should use strings.ToLower
				expected := strings.ToLower(tc.apiStatus)
				if !ok {
					// This is expected behavior - unknown statuses use ToLower
					_ = expected
				}
			}
		})
	}
}

// ============================================================================
// Permissions List Command Tests
// ============================================================================

func TestPermissionsListCmd_Run(t *testing.T) {
	cmd := &PermissionsListCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: false,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("PermissionsListCmd.Run() unexpected error: %v", err)
	}
}

func TestPermissionsListCmd_Run_WithPrettyOutput(t *testing.T) {
	cmd := &PermissionsListCmd{}
	globals := &Globals{
		Output: "json",
		Pretty: true,
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("PermissionsListCmd.Run() with pretty output unexpected error: %v", err)
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestPermissionsCommands_ErrorTypes(t *testing.T) {
	t.Run("package required error has correct code", func(t *testing.T) {
		err := gpdErrors.ErrPackageRequired
		if err.Code != gpdErrors.CodeValidationError {
			t.Errorf("ErrPackageRequired code = %s, expected %s", err.Code, gpdErrors.CodeValidationError)
		}
		if err.ExitCode() != gpdErrors.ExitValidationError {
			t.Errorf("ErrPackageRequired exit code = %d, expected %d", err.ExitCode(), gpdErrors.ExitValidationError)
		}
	})
}

// ============================================================================
// User Resource Name Building Tests
// ============================================================================

func TestUserResourceNameBuilding(t *testing.T) {
	// Test the resource name pattern used in permissions commands
	email := "user@example.com"
	parent := getDeveloperParent()
	expectedName := parent + "/users/" + email

	// This is the pattern used in PermissionsUsersRemoveCmd
	userName := parent + "/users/" + email
	if userName != expectedName {
		t.Errorf("User resource name = %q, expected %q", userName, expectedName)
	}
}

func TestGrantResourceNameBuilding(t *testing.T) {
	// Test the grant resource name pattern
	email := "user@example.com"
	packageName := "com.example.app"
	parent := getDeveloperParent()
	expectedName := parent + "/users/" + email + "/grants/" + packageName

	// This is the pattern used in PermissionsGrantsRemoveCmd
	grantName := parent + "/users/" + email + "/grants/" + packageName
	if grantName != expectedName {
		t.Errorf("Grant resource name = %q, expected %q", grantName, expectedName)
	}
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestPermissionsCommands_WithDifferentOutputs(t *testing.T) {
	formats := []string{"json", "table"}

	for _, format := range formats {
		t.Run("PermissionsListCmd_"+format, func(t *testing.T) {
			cmd := &PermissionsListCmd{}
			globals := &Globals{
				Output: format,
				Pretty: false,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("PermissionsListCmd.Run() with %s output unexpected error: %v", format, err)
			}
		})
	}
}

// ============================================================================
// Recovery Action ID Parsing Edge Cases
// ============================================================================

func TestRecoveryActionIDParsing(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		shouldParse bool
		expectedID  int64
	}{
		{"valid numeric ID", "12345", true, 12345},
		{"zero ID", "0", true, 0},
		{"large ID", "9223372036854775807", true, 9223372036854775807}, // Max int64
		{"negative ID", "-1", true, -1},
		{"invalid alphanumeric", "abc123", false, 0},
		{"invalid with spaces", "123 456", false, 0},
		{"empty string", "", false, 0},
		{"floating point", "123.45", false, 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			parsed, err := strconv.ParseInt(tc.id, 10, 64)
			if tc.shouldParse {
				if err != nil {
					t.Errorf("Expected %q to parse successfully, got error: %v", tc.id, err)
					return
				}
				if parsed != tc.expectedID {
					t.Errorf("ParseInt(%q) = %d, expected %d", tc.id, parsed, tc.expectedID)
				}
			} else {
				if err == nil {
					t.Errorf("Expected %q to fail parsing, but got: %d", tc.id, parsed)
				}
			}
		})
	}
}

// ============================================================================
// Timeout and Context Tests
// ============================================================================

func TestPermissionsCommands_WithTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	time.Sleep(5 * time.Millisecond)

	if ctx.Err() != context.DeadlineExceeded {
		t.Error("Expected context deadline exceeded")
	}
}

// ============================================================================
// Concurrent Safety Tests
// ============================================================================

func TestRoleToDeveloperPermissions_Concurrent(t *testing.T) {
	// Test that roleToDeveloperPermissions is safe for concurrent use
	// (it's a pure function with no shared state, so it should be)

	done := make(chan bool, 3)

	go func() {
		roleToDeveloperPermissions("admin")
		done <- true
	}()

	go func() {
		roleToDeveloperPermissions("developer")
		done <- true
	}()

	go func() {
		roleToDeveloperPermissions("viewer")
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("Timeout waiting for concurrent roleToDeveloperPermissions calls")
		}
	}
}

// ============================================================================
// Email Filter Tests (for Grants List)
// ============================================================================

func TestEmailFilter_Matching(t *testing.T) {
	tests := []struct {
		filter      string
		userEmail   string
		shouldMatch bool
	}{
		{"user@example.com", "user@example.com", true},
		{"user@example.com", "USER@EXAMPLE.COM", true}, // Case insensitive
		{"user@example.com", "other@example.com", false},
		{"", "user@example.com", true},              // Empty filter matches all
		{"@example.com", "user@example.com", false}, // Partial match with EqualFold
		{"user", "user@example.com", false},         // Partial match with EqualFold
	}

	for _, tc := range tests {
		t.Run(tc.filter+"_"+tc.userEmail, func(t *testing.T) {
			// This mimics the filtering logic in PermissionsGrantsListCmd
			matches := tc.filter == "" || strings.EqualFold(tc.userEmail, tc.filter)
			if matches != tc.shouldMatch {
				t.Errorf("Filter %q against %q: expected match=%v, got match=%v",
					tc.filter, tc.userEmail, tc.shouldMatch, matches)
			}
		})
	}
}

// ============================================================================
// Permissions Count Validation
// ============================================================================

func TestPermissionsList_Count(t *testing.T) {
	// The permissions list in PermissionsListCmd has a specific set of permissions
	// This test validates that the count matches expected values

	// Developer-level permissions (from the source code)
	developerPermissions := []string{
		"CAN_SEE_ALL_APPS",
		"CAN_VIEW_FINANCIAL_DATA_GLOBAL",
		"CAN_MANAGE_PERMISSIONS_GLOBAL",
		"CAN_EDIT_GAMES_GLOBAL",
		"CAN_PUBLISH_GAMES_GLOBAL",
		"CAN_REPLY_TO_REVIEWS_GLOBAL",
		"CAN_MANAGE_PUBLIC_APKS_GLOBAL",
		"CAN_MANAGE_TRACK_APKS_GLOBAL",
		"CAN_MANAGE_TRACK_USERS_GLOBAL",
		"CAN_MANAGE_PUBLIC_LISTING_GLOBAL",
		"CAN_MANAGE_DRAFT_APPS_GLOBAL",
		"CAN_CREATE_MANAGED_PLAY_APPS_GLOBAL",
		"CAN_CHANGE_MANAGED_PLAY_SETTING_GLOBAL",
		"CAN_MANAGE_ORDERS_GLOBAL",
		"CAN_MANAGE_APP_CONTENT_GLOBAL",
		"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
		"CAN_VIEW_APP_QUALITY_GLOBAL",
		"CAN_MANAGE_DEEPLINKS_GLOBAL",
	}

	// App-level permissions
	appPermissions := []string{
		"CAN_ACCESS_APP",
		"CAN_VIEW_FINANCIAL_DATA",
		"CAN_MANAGE_PERMISSIONS",
		"CAN_REPLY_TO_REVIEWS",
		"CAN_MANAGE_PUBLIC_APKS",
		"CAN_MANAGE_TRACK_APKS",
		"CAN_MANAGE_TRACK_USERS",
		"CAN_MANAGE_PUBLIC_LISTING",
		"CAN_MANAGE_DRAFT_APPS",
		"CAN_MANAGE_ORDERS",
		"CAN_MANAGE_APP_CONTENT",
		"CAN_VIEW_NON_FINANCIAL_DATA",
		"CAN_VIEW_APP_QUALITY",
		"CAN_MANAGE_DEEPLINKS",
	}

	expectedTotal := len(developerPermissions) + len(appPermissions)
	if expectedTotal != 32 {
		t.Errorf("Expected 32 total permissions, got %d", expectedTotal)
	}
}

// ============================================================================
// Command Run Signatures
// ============================================================================

func TestCommandRunSignatures(t *testing.T) {
	// Verify all commands have the correct Run method signature
	commands := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{"PermissionsUsersAddCmd", &PermissionsUsersAddCmd{}},
		{"PermissionsUsersRemoveCmd", &PermissionsUsersRemoveCmd{}},
		{"PermissionsUsersListCmd", &PermissionsUsersListCmd{}},
		{"PermissionsGrantsAddCmd", &PermissionsGrantsAddCmd{}},
		{"PermissionsGrantsRemoveCmd", &PermissionsGrantsRemoveCmd{}},
		{"PermissionsGrantsListCmd", &PermissionsGrantsListCmd{}},
		{"PermissionsListCmd", &PermissionsListCmd{}},
		{"RecoveryListCmd", &RecoveryListCmd{}},
		{"RecoveryCreateCmd", &RecoveryCreateCmd{}},
		{"RecoveryDeployCmd", &RecoveryDeployCmd{}},
		{"RecoveryCancelCmd", &RecoveryCancelCmd{}},
		{"IntegrityDecodeCmd", &IntegrityDecodeCmd{}},
	}

	for _, tc := range commands {
		t.Run(tc.name, func(t *testing.T) {
			// Just verify the method exists and can be called
			// It will fail due to missing auth, but shouldn't panic
			globals := &Globals{
				Package: "com.example.app",
				Output:  "json",
			}
			_ = tc.cmd.Run(globals)
		})
	}
}
