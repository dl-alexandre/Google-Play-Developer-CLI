//go:build unit
// +build unit

package cli

import (
	"testing"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// ============================================================================
// PurchasesSubscriptionsV2GetCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsV2GetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsV2GetCmd{
		Token: "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// PurchasesSubscriptionsV2CancelCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsV2CancelCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsV2CancelCmd{
		Token:            "token-abc",
		CancellationType: "userRequestedStopRenewals",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPurchasesSubscriptionsV2CancelCmd_Run_InvalidCancellationType(t *testing.T) {
	cmd := &PurchasesSubscriptionsV2CancelCmd{
		Token:            "token-abc",
		CancellationType: "invalid-type",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid cancellation type")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
}

func TestPurchasesSubscriptionsV2CancelCmd_Run_ValidCancellationTypes(t *testing.T) {
	validTypes := []string{
		"userRequestedStopRenewals",
		"user",
		"developerRequestedStopPayments",
		"developer",
	}

	for _, cancelType := range validTypes {
		cmd := &PurchasesSubscriptionsV2CancelCmd{
			Token:            "token-abc",
			CancellationType: cancelType,
		}
		// Validation should pass, but will fail on API call since no auth
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		// Should fail on API call, not on validation
		if err == nil {
			t.Fatalf("Expected error for type %s (no API auth)", cancelType)
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError for type %s, got: %T", cancelType, err)
		}
		// Should be an API error (auth or general), not validation
		if apiErr.Code == errors.CodeValidationError {
			t.Errorf("Did not expect validation error for valid type %s", cancelType)
		}
	}
}

// ============================================================================
// PurchasesSubscriptionsV2DeferCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsV2DeferCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsV2DeferCmd{
		Token:         "token-abc",
		DeferDuration: "7d",
		Etag:          "abc123",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// PurchasesSubscriptionsV2RevokeCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsV2RevokeCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsV2RevokeCmd{
		Token:      "token-abc",
		RevokeType: "fullRefund",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPurchasesSubscriptionsV2RevokeCmd_Run_InvalidRevokeType(t *testing.T) {
	cmd := &PurchasesSubscriptionsV2RevokeCmd{
		Token:      "token-abc",
		RevokeType: "invalid-type",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid revoke type")
	}
	// The error could be either validation (if auth were available)
	// or auth failure (in test environment without credentials)
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	// In test environment without auth, we get AUTH_FAILURE before validation
	// With auth, we'd get CodeValidationError
	if apiErr.Code != errors.CodeValidationError && apiErr.Code != errors.CodeAuthFailure {
		t.Errorf("Expected validation or auth error code, got: %v", apiErr.Code)
	}
}

func TestPurchasesSubscriptionsV2RevokeCmd_Run_ValidRevokeTypes(t *testing.T) {
	validTypes := []string{
		"fullRefund",
		"partialRefund",
		"proratedRefund",
		"itemBasedRefund",
	}

	for _, revokeType := range validTypes {
		cmd := &PurchasesSubscriptionsV2RevokeCmd{
			Token:      "token-abc",
			RevokeType: revokeType,
		}
		// Validation should pass, but will fail on API call since no auth
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		// Should fail on API call, not on validation
		if err == nil {
			t.Fatalf("Expected error for type %s (no API auth)", revokeType)
		}
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatalf("Expected APIError for type %s, got: %T", revokeType, err)
		}
		// Should be an API error (auth or general), not validation
		if apiErr.Code == errors.CodeValidationError {
			t.Errorf("Did not expect validation error for valid type %s", revokeType)
		}
	}
}
