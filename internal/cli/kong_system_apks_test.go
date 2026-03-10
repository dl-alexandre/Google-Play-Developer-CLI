//go:build unit
// +build unit

package cli

import (
	"testing"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// ============================================================================
// SystemApksVariantsListCmd Tests
// ============================================================================

func TestSystemApksVariantsListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &SystemApksVariantsListCmd{
		VersionCode: 123,
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestSystemApksVariantsListCmd_InvalidVersionCode(t *testing.T) {
	cmd := &SystemApksVariantsListCmd{
		VersionCode: -1,
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	// Should fail on API call (invalid version code validation happens at API level)
	// or auth failure in test environment
	if err == nil {
		t.Fatal("Expected error for invalid version code (API/auth failure expected)")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	// In test without auth, expect auth failure
	// With auth but invalid version code, could be validation or not found
	if apiErr.Code != errors.CodeAuthFailure && apiErr.Code != errors.CodeValidationError && apiErr.Code != errors.CodeNotFound {
		t.Errorf("Expected auth, validation, or not found error, got: %v", apiErr.Code)
	}
}

func TestSystemApksVariantsListCmd_Validation(t *testing.T) {
	tests := []struct {
		name        string
		versionCode int64
		wantErr     bool
	}{
		{
			name:        "zero version code",
			versionCode: 0,
			wantErr:     true,
		},
		{
			name:        "valid version code",
			versionCode: 1,
			wantErr:     false, // Will fail on API auth, not validation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &SystemApksVariantsListCmd{
				VersionCode: tt.versionCode,
			}
			globals := &Globals{Package: "com.example.app"}

			err := cmd.Run(globals)
			if tt.wantErr {
				// Should fail on API or auth
				if err == nil {
					t.Fatal("Expected error")
				}
			} else {
				// Should fail on API auth, not validation
				if err == nil {
					t.Fatal("Expected API/auth error")
				}
				apiErr, ok := err.(*errors.APIError)
				if !ok {
					t.Fatalf("Expected APIError, got: %T", err)
				}
				// Should not be validation error for valid version code
				if apiErr.Code == errors.CodeValidationError {
					t.Errorf("Did not expect validation error for valid version code %d", tt.versionCode)
				}
			}
		})
	}
}

// ============================================================================
// SystemApksVariantsGetCmd Tests
// ============================================================================

func TestSystemApksVariantsGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &SystemApksVariantsGetCmd{
		VersionCode: 123,
		VariantId:   456,
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestSystemApksVariantsGetCmd_InvalidVariantID(t *testing.T) {
	cmd := &SystemApksVariantsGetCmd{
		VersionCode: 123,
		VariantId:   -1,
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	// Should fail on API call or auth in test environment
	if err == nil {
		t.Fatal("Expected error for invalid variant ID (API/auth failure expected)")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	// In test without auth, expect auth failure
	// With auth but invalid variant ID, could be validation or not found
	if apiErr.Code != errors.CodeAuthFailure && apiErr.Code != errors.CodeValidationError && apiErr.Code != errors.CodeNotFound {
		t.Errorf("Expected auth, validation, or not found error, got: %v", apiErr.Code)
	}
}

// ============================================================================
// SystemApksVariantsCreateCmd Tests
// ============================================================================

func TestSystemApksVariantsCreateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &SystemApksVariantsCreateCmd{
		VersionCode: 123,
		DeviceSpec:  "density=480,abis=arm64-v8a,locales=en-US",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestSystemApksVariantsCreateCmd_MissingDeviceSpec(t *testing.T) {
	cmd := &SystemApksVariantsCreateCmd{
		VersionCode: 123,
		// No DeviceSpec or File provided
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing device specification")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
}

func TestSystemApksVariantsCreateCmd_InvalidDeviceSpec(t *testing.T) {
	tests := []struct {
		name       string
		deviceSpec string
		wantErr    bool
	}{
		{
			name:       "empty spec",
			deviceSpec: "",
			wantErr:    true,
		},
		{
			name:       "invalid format - no equals",
			deviceSpec: "invalidformat",
			wantErr:    true,
		},
		{
			name:       "invalid density value",
			deviceSpec: "density=invalid",
			wantErr:    true,
		},
		{
			name:       "unknown key",
			deviceSpec: "unknown=value",
			wantErr:    true,
		},
		{
			name:       "valid spec with density",
			deviceSpec: "density=480",
			wantErr:    false,
		},
		{
			name:       "valid spec with abis",
			deviceSpec: "abis=arm64-v8a|armeabi-v7a",
			wantErr:    false,
		},
		{
			name:       "valid spec with locales",
			deviceSpec: "locales=en-US|es-ES",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &SystemApksVariantsCreateCmd{
				VersionCode: 123,
				DeviceSpec:  tt.deviceSpec,
			}
			globals := &Globals{Package: "com.example.app"}

			err := cmd.Run(globals)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Expected error for invalid device spec")
				}
				// For empty spec, we get a specific validation error from command validation
				// For invalid formats, we get it from parseDeviceSpec
				apiErr, ok := err.(*errors.APIError)
				if !ok {
					// Empty spec case returns error before parseDeviceSpec is called
					if tt.deviceSpec == "" {
						t.Fatalf("Expected APIError, got: %T", err)
					}
					// parseDeviceSpec returns errors.NewAPIError which should be *APIError
					t.Fatalf("Expected APIError from parseDeviceSpec, got: %T", err)
				}
				if apiErr.Code != errors.CodeValidationError {
					t.Errorf("Expected validation error, got: %v", apiErr.Code)
				}
			} else {
				// Valid spec should fail on API auth, not validation
				if err == nil {
					t.Fatal("Expected API/auth error for valid spec")
				}
				apiErr, ok := err.(*errors.APIError)
				if !ok {
					t.Fatalf("Expected APIError, got: %T", err)
				}
				// Should not be validation error
				if apiErr.Code == errors.CodeValidationError {
					t.Errorf("Did not expect validation error for valid device spec: %s", tt.deviceSpec)
				}
			}
		})
	}
}

// ============================================================================
// SystemApksVariantsDownloadCmd Tests
// ============================================================================

func TestSystemApksVariantsDownloadCmd_Run_PackageRequired(t *testing.T) {
	cmd := &SystemApksVariantsDownloadCmd{
		VersionCode: 123,
		VariantID:   456,
		OutputFile:  "test.apk",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestSystemApksVariantsDownloadCmd_InvalidVariantID(t *testing.T) {
	cmd := &SystemApksVariantsDownloadCmd{
		VersionCode: 123,
		VariantID:   -1,
		OutputFile:  "test.apk",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	// Should fail on API call or auth in test environment
	if err == nil {
		t.Fatal("Expected error for invalid variant ID (API/auth failure expected)")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	// In test without auth, expect auth failure
	// With auth but invalid variant ID, could be validation or not found
	if apiErr.Code != errors.CodeAuthFailure && apiErr.Code != errors.CodeValidationError && apiErr.Code != errors.CodeNotFound {
		t.Errorf("Expected auth, validation, or not found error, got: %v", apiErr.Code)
	}
}
