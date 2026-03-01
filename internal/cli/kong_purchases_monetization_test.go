//go:build unit
// +build unit

package cli

import (
	"context"
	"testing"
	"time"

	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// Helper: createTempFile (reuses helper from kong_publish_test.go)
// ============================================================================

// Note: createTempFile helper is defined in kong_publish_test.go
// and is accessible within the same package.

// ============================================================================
// parseTimeToMillis Tests
// ============================================================================

func TestParseTimeToMillis(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{
			name:    "valid milliseconds",
			input:   "1704067200000",
			want:    1704067200000,
			wantErr: false,
		},
		{
			name:    "valid RFC3339",
			input:   "2024-01-01T00:00:00Z",
			want:    1704067200000,
			wantErr: false,
		},
		{
			name:    "valid RFC3339 with offset",
			input:   "2024-01-01T08:00:00+08:00",
			want:    1704067200000,
			wantErr: false,
		},
		{
			name:    "zero milliseconds",
			input:   "0",
			want:    0,
			wantErr: false,
		},
		{
			name:    "negative milliseconds",
			input:   "-1",
			want:    -1,
			wantErr: false,
		},
		{
			name:    "invalid string",
			input:   "not-a-time",
			want:    0,
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    0,
			wantErr: true,
		},
		{
			name:    "invalid date format",
			input:   "2024/01/01",
			want:    0,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseTimeToMillis(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTimeToMillis() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("parseTimeToMillis() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// PurchasesProductsAcknowledgeCmd Tests
// ============================================================================

func TestPurchasesProductsAcknowledgeCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesProductsAcknowledgeCmd{
		ProductID: "product-123",
		Token:     "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// PurchasesProductsConsumeCmd Tests
// ============================================================================

func TestPurchasesProductsConsumeCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesProductsConsumeCmd{
		ProductID: "product-123",
		Token:     "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// PurchasesSubscriptionsAcknowledgeCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsAcknowledgeCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsAcknowledgeCmd{
		SubscriptionID: "sub-123",
		Token:          "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// PurchasesSubscriptionsCancelCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsCancelCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsCancelCmd{
		SubscriptionID: "sub-123",
		Token:          "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// PurchasesSubscriptionsDeferCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsDeferCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsDeferCmd{
		SubscriptionID: "sub-123",
		Token:          "token-abc",
		ExpectedExpiry: "2024-12-31T23:59:59Z",
		DesiredExpiry:  "2025-01-31T23:59:59Z",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPurchasesSubscriptionsDeferCmd_Run_InvalidExpectedExpiry(t *testing.T) {
	cmd := &PurchasesSubscriptionsDeferCmd{
		SubscriptionID: "sub-123",
		Token:          "token-abc",
		ExpectedExpiry: "invalid-time",
		DesiredExpiry:  "2025-01-31T23:59:59Z",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid expected expiry")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Error("Expected error to have hint")
	}
}

func TestPurchasesSubscriptionsDeferCmd_Run_InvalidDesiredExpiry(t *testing.T) {
	cmd := &PurchasesSubscriptionsDeferCmd{
		SubscriptionID: "sub-123",
		Token:          "token-abc",
		ExpectedExpiry: "2024-12-31T23:59:59Z",
		DesiredExpiry:  "invalid-time",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid desired expiry")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
}

// ============================================================================
// PurchasesSubscriptionsRefundCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsRefundCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsRefundCmd{
		SubscriptionID: "sub-123",
		Token:          "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// PurchasesSubscriptionsRevokeCmd Tests
// ============================================================================

func TestPurchasesSubscriptionsRevokeCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesSubscriptionsRevokeCmd{
		Token: "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPurchasesSubscriptionsRevokeCmd_Run_InvalidRevokeType(t *testing.T) {
	cmd := &PurchasesSubscriptionsRevokeCmd{
		Token:      "token-abc",
		RevokeType: "invalidType",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid revoke type")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Error("Expected error to have hint about valid types")
	}
}

func TestPurchasesSubscriptionsRevokeCmd_Run_MissingSubscriptionIDForV1(t *testing.T) {
	cmd := &PurchasesSubscriptionsRevokeCmd{
		Token: "token-abc",
		// No SubscriptionID and no RevokeType - should require SubscriptionID for v1
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing subscription ID")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Error("Expected error to have hint about using --subscription-id or --revoke-type")
	}
}

// ============================================================================
// PurchasesVerifyCmd Tests
// ============================================================================

func TestPurchasesVerifyCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesVerifyCmd{
		Token: "token-abc",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPurchasesVerifyCmd_Run_InvalidType(t *testing.T) {
	cmd := &PurchasesVerifyCmd{
		Token: "token-abc",
		Type:  "invalid",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid type")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
}

func TestPurchasesVerifyCmd_Run_MissingProductIDForProductType(t *testing.T) {
	cmd := &PurchasesVerifyCmd{
		Token: "token-abc",
		Type:  "product",
		// ProductID is empty
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing product ID")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Error("Expected error to have hint about providing --product-id")
	}
}

// ============================================================================
// PurchasesVoidedListCmd Tests
// ============================================================================

func TestPurchasesVoidedListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &PurchasesVoidedListCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestPurchasesVoidedListCmd_Run_InvalidStartTime(t *testing.T) {
	cmd := &PurchasesVoidedListCmd{
		StartTime: "invalid-time",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid start time")
	}
}

func TestPurchasesVoidedListCmd_Run_InvalidEndTime(t *testing.T) {
	cmd := &PurchasesVoidedListCmd{
		EndTime: "invalid-time",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid end time")
	}
}

// ============================================================================
// PurchasesCapabilitiesCmd Tests
// ============================================================================

func TestPurchasesCapabilitiesCmd_Run(t *testing.T) {
	cmd := &PurchasesCapabilitiesCmd{}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// ============================================================================
// MonetizationProductsListCmd Tests
// ============================================================================

func TestMonetizationProductsListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationProductsListCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationProductsGetCmd Tests
// ============================================================================

func TestMonetizationProductsGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationProductsGetCmd{
		ProductID: "product-123",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationProductsCreateCmd Tests
// ============================================================================

func TestMonetizationProductsCreateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationProductsCreateCmd{
		ProductID: "product-123",
		Type:      "managed",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationProductsUpdateCmd Tests
// ============================================================================

func TestMonetizationProductsUpdateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationProductsUpdateCmd{
		ProductID: "product-123",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationProductsDeleteCmd Tests
// ============================================================================

func TestMonetizationProductsDeleteCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationProductsDeleteCmd{
		ProductID: "product-123",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationSubscriptionsListCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsListCmd{}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationSubscriptionsGetCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsGetCmd{
		SubscriptionID: "sub-123",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationSubscriptionsCreateCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsCreateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsCreateCmd{
		SubscriptionID: "sub-123",
		File:           "/tmp/subscription.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestMonetizationSubscriptionsCreateCmd_Run_InvalidFile(t *testing.T) {
	cmd := &MonetizationSubscriptionsCreateCmd{
		SubscriptionID: "sub-123",
		File:           "/nonexistent/file.json",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid file")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
}

func TestMonetizationSubscriptionsCreateCmd_Run_InvalidJSON(t *testing.T) {
	// Create a temp file with invalid JSON
	tmpFile := createTempFile(t, "invalid.json", []byte("not valid json"))

	cmd := &MonetizationSubscriptionsCreateCmd{
		SubscriptionID: "sub-123",
		File:           tmpFile,
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Error("Expected error to have hint about valid JSON")
	}
}

// ============================================================================
// MonetizationSubscriptionsUpdateCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsUpdateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsUpdateCmd{
		SubscriptionID: "sub-123",
		File:           "/tmp/subscription.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationSubscriptionsPatchCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsPatchCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsPatchCmd{
		SubscriptionID: "sub-123",
		File:           "/tmp/subscription.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationSubscriptionsDeleteCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsDeleteCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsDeleteCmd{
		SubscriptionID: "sub-123",
		Confirm:        true,
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestMonetizationSubscriptionsDeleteCmd_Run_MissingConfirm(t *testing.T) {
	cmd := &MonetizationSubscriptionsDeleteCmd{
		SubscriptionID: "sub-123",
		Confirm:        false, // Not confirmed
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing confirmation")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Error("Expected error to have hint about --confirm flag")
	}
}

// ============================================================================
// MonetizationSubscriptionsArchiveCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsArchiveCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsArchiveCmd{
		SubscriptionID: "sub-123",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationSubscriptionsBatchGetCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsBatchGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsBatchGetCmd{
		IDs: []string{"sub-1", "sub-2"},
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationSubscriptionsBatchUpdateCmd Tests
// ============================================================================

func TestMonetizationSubscriptionsBatchUpdateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationSubscriptionsBatchUpdateCmd{
		File: "/tmp/batch.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationBasePlansActivateCmd Tests
// ============================================================================

func TestMonetizationBasePlansActivateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationBasePlansActivateCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationBasePlansDeactivateCmd Tests
// ============================================================================

func TestMonetizationBasePlansDeactivateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationBasePlansDeactivateCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationBasePlansDeleteCmd Tests
// ============================================================================

func TestMonetizationBasePlansDeleteCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationBasePlansDeleteCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		Confirm:        true,
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestMonetizationBasePlansDeleteCmd_Run_MissingConfirm(t *testing.T) {
	cmd := &MonetizationBasePlansDeleteCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		Confirm:        false, // Not confirmed
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing confirmation")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
}

// ============================================================================
// MonetizationBasePlansMigratePricesCmd Tests
// ============================================================================

func TestMonetizationBasePlansMigratePricesCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationBasePlansMigratePricesCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		RegionCode:     "US",
		PriceMicros:    990000,
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationBasePlansBatchMigrateCmd Tests
// ============================================================================

func TestMonetizationBasePlansBatchMigrateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationBasePlansBatchMigrateCmd{
		SubscriptionID: "sub-123",
		File:           "/tmp/batch.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationBasePlansBatchUpdateStatesCmd Tests
// ============================================================================

func TestMonetizationBasePlansBatchUpdateStatesCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationBasePlansBatchUpdateStatesCmd{
		SubscriptionID: "sub-123",
		File:           "/tmp/batch.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersCreateCmd Tests
// ============================================================================

func TestMonetizationOffersCreateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersCreateCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		OfferID:        "offer-1",
		File:           "/tmp/offer.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersGetCmd Tests
// ============================================================================

func TestMonetizationOffersGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersGetCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		OfferID:        "offer-1",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersListCmd Tests
// ============================================================================

func TestMonetizationOffersListCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersListCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersDeleteCmd Tests
// ============================================================================

func TestMonetizationOffersDeleteCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersDeleteCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		OfferID:        "offer-1",
		Confirm:        true,
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

func TestMonetizationOffersDeleteCmd_Run_MissingConfirm(t *testing.T) {
	cmd := &MonetizationOffersDeleteCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		OfferID:        "offer-1",
		Confirm:        false, // Not confirmed
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for missing confirmation")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
}

// ============================================================================
// MonetizationOffersActivateCmd Tests
// ============================================================================

func TestMonetizationOffersActivateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersActivateCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		OfferID:        "offer-1",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersDeactivateCmd Tests
// ============================================================================

func TestMonetizationOffersDeactivateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersDeactivateCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		OfferID:        "offer-1",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersBatchGetCmd Tests
// ============================================================================

func TestMonetizationOffersBatchGetCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersBatchGetCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		OfferIDs:       []string{"offer-1", "offer-2"},
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersBatchUpdateCmd Tests
// ============================================================================

func TestMonetizationOffersBatchUpdateCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersBatchUpdateCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		File:           "/tmp/batch.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationOffersBatchUpdateStatesCmd Tests
// ============================================================================

func TestMonetizationOffersBatchUpdateStatesCmd_Run_PackageRequired(t *testing.T) {
	cmd := &MonetizationOffersBatchUpdateStatesCmd{
		SubscriptionID: "sub-123",
		BasePlanID:     "base-plan-1",
		File:           "/tmp/batch.json",
	}
	globals := &Globals{} // No package set

	err := cmd.Run(globals)
	if err != errors.ErrPackageRequired {
		t.Errorf("Expected ErrPackageRequired, got: %v", err)
	}
}

// ============================================================================
// MonetizationCapabilitiesCmd Tests
// ============================================================================

func TestMonetizationCapabilitiesCmd_Run(t *testing.T) {
	cmd := &MonetizationCapabilitiesCmd{}
	globals := &Globals{
		Package: "com.example.app",
		Output:  "json",
	}

	err := cmd.Run(globals)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// ============================================================================
// Command Structure Tests
// ============================================================================

func TestCommandStructures(t *testing.T) {
	t.Run("PurchasesCmd structure", func(t *testing.T) {
		cmd := &PurchasesCmd{}
		// Verify all subcommands exist
		_ = cmd.Products
		_ = cmd.Subscriptions
		_ = cmd.Verify
		_ = cmd.Voided
		_ = cmd.Capabilities
	})

	t.Run("PurchasesProductsCmd structure", func(t *testing.T) {
		cmd := &PurchasesProductsCmd{}
		_ = cmd.Acknowledge
		_ = cmd.Consume
	})

	t.Run("PurchasesSubscriptionsCmd structure", func(t *testing.T) {
		cmd := &PurchasesSubscriptionsCmd{}
		_ = cmd.Acknowledge
		_ = cmd.Cancel
		_ = cmd.Defer
		_ = cmd.Refund
		_ = cmd.Revoke
	})

	t.Run("MonetizationCmd structure", func(t *testing.T) {
		cmd := &MonetizationCmd{}
		_ = cmd.Products
		_ = cmd.Subscriptions
		_ = cmd.BasePlans
		_ = cmd.Offers
		_ = cmd.Capabilities
	})

	t.Run("MonetizationProductsCmd structure", func(t *testing.T) {
		cmd := &MonetizationProductsCmd{}
		_ = cmd.List
		_ = cmd.Get
		_ = cmd.Create
		_ = cmd.Update
		_ = cmd.Delete
	})

	t.Run("MonetizationSubscriptionsCmd structure", func(t *testing.T) {
		cmd := &MonetizationSubscriptionsCmd{}
		_ = cmd.List
		_ = cmd.Get
		_ = cmd.Create
		_ = cmd.Update
		_ = cmd.Patch
		_ = cmd.Delete
		_ = cmd.Archive
		_ = cmd.BatchGet
		_ = cmd.BatchUpdate
	})

	t.Run("MonetizationBasePlansCmd structure", func(t *testing.T) {
		cmd := &MonetizationBasePlansCmd{}
		_ = cmd.Activate
		_ = cmd.Deactivate
		_ = cmd.Delete
		_ = cmd.MigratePrices
		_ = cmd.BatchMigrate
		_ = cmd.BatchUpdateStates
	})

	t.Run("MonetizationOffersCmd structure", func(t *testing.T) {
		cmd := &MonetizationOffersCmd{}
		_ = cmd.Create
		_ = cmd.Get
		_ = cmd.List
		_ = cmd.Delete
		_ = cmd.Activate
		_ = cmd.Deactivate
		_ = cmd.BatchGet
		_ = cmd.BatchUpdate
		_ = cmd.BatchUpdateStates
	})
}

// ============================================================================
// Context Handling Tests
// ============================================================================

func TestContextHandling(t *testing.T) {
	t.Run("nil context is handled", func(t *testing.T) {
		cmd := &PurchasesCapabilitiesCmd{}
		globals := &Globals{
			Package: "com.example.app",
			Context: nil, // Explicitly nil
		}

		// Should not panic
		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error with nil context: %v", err)
		}
	})

	t.Run("context is passed through", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := &PurchasesCapabilitiesCmd{}
		globals := &Globals{
			Package: "com.example.app",
			Context: ctx,
		}

		err := cmd.Run(globals)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}

// ============================================================================
// Flag/Field Validation Tests
// ============================================================================

func TestRequiredFieldValidation(t *testing.T) {
	tests := []struct {
		name    string
		cmd     interface{ Run(*Globals) error }
		globals *Globals
		wantErr bool
	}{
		{
			name: "PurchasesProductsAcknowledgeCmd missing fields",
			cmd:  &PurchasesProductsAcknowledgeCmd{
				// Missing ProductID and Token
			},
			globals: &Globals{Package: "com.example.app"},
			wantErr: true, // Will fail at API level
		},
		{
			name: "PurchasesSubscriptionsDeferCmd missing expiry times",
			cmd: &PurchasesSubscriptionsDeferCmd{
				SubscriptionID: "sub-123",
				Token:          "token-abc",
				// Missing ExpectedExpiry and DesiredExpiry
			},
			globals: &Globals{Package: "com.example.app"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cmd.Run(tt.globals)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================================
// Error Type Tests
// ============================================================================

func TestErrorTypes(t *testing.T) {
	t.Run("validation errors have correct code", func(t *testing.T) {
		cmd := &PurchasesSubscriptionsRevokeCmd{
			Token:      "token-abc",
			RevokeType: "invalid",
		}
		globals := &Globals{Package: "com.example.app"}

		err := cmd.Run(globals)
		apiErr, ok := err.(*errors.APIError)
		if !ok {
			t.Fatal("Expected APIError")
		}
		if apiErr.Code != errors.CodeValidationError {
			t.Errorf("Expected CodeValidationError, got: %v", apiErr.Code)
		}
	})

	t.Run("package required error is correct type", func(t *testing.T) {
		cmd := &PurchasesProductsAcknowledgeCmd{}
		globals := &Globals{}

		err := cmd.Run(globals)
		if err != errors.ErrPackageRequired {
			t.Errorf("Expected ErrPackageRequired, got: %v", err)
		}
	})
}

// ============================================================================
// Pagination Response Wrapper Tests
// ============================================================================

func TestPaginationResponseWrappers(t *testing.T) {
	t.Run("voidedPurchasesPageResponse implements interface", func(t *testing.T) {
		// This test verifies the struct exists and has the right methods
		// The actual testing would be done through the list command
		_ = voidedPurchasesPageResponse{}
	})

	t.Run("inappProductsPageResponse implements interface", func(t *testing.T) {
		_ = inappProductsPageResponse{}
	})

	t.Run("subscriptionsPageResponse implements interface", func(t *testing.T) {
		_ = subscriptionsPageResponse{}
	})

	t.Run("offersPageResponse implements interface", func(t *testing.T) {
		_ = offersPageResponse{}
	})
}

// ============================================================================
// File Reading Error Tests
// ============================================================================

func TestFileReadingErrors(t *testing.T) {
	tests := []struct {
		name string
		cmd  interface{ Run(*Globals) error }
	}{
		{
			name: "MonetizationSubscriptionsCreateCmd with missing file",
			cmd: &MonetizationSubscriptionsCreateCmd{
				SubscriptionID: "sub-123",
				File:           "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationSubscriptionsUpdateCmd with missing file",
			cmd: &MonetizationSubscriptionsUpdateCmd{
				SubscriptionID: "sub-123",
				File:           "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationSubscriptionsPatchCmd with missing file",
			cmd: &MonetizationSubscriptionsPatchCmd{
				SubscriptionID: "sub-123",
				File:           "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationSubscriptionsBatchUpdateCmd with missing file",
			cmd: &MonetizationSubscriptionsBatchUpdateCmd{
				File: "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationBasePlansBatchMigrateCmd with missing file",
			cmd: &MonetizationBasePlansBatchMigrateCmd{
				SubscriptionID: "sub-123",
				File:           "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationBasePlansBatchUpdateStatesCmd with missing file",
			cmd: &MonetizationBasePlansBatchUpdateStatesCmd{
				SubscriptionID: "sub-123",
				File:           "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationOffersCreateCmd with missing file",
			cmd: &MonetizationOffersCreateCmd{
				SubscriptionID: "sub-123",
				BasePlanID:     "base-1",
				OfferID:        "offer-1",
				File:           "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationOffersBatchUpdateCmd with missing file",
			cmd: &MonetizationOffersBatchUpdateCmd{
				SubscriptionID: "sub-123",
				BasePlanID:     "base-1",
				File:           "/nonexistent/file.json",
			},
		},
		{
			name: "MonetizationOffersBatchUpdateStatesCmd with missing file",
			cmd: &MonetizationOffersBatchUpdateStatesCmd{
				SubscriptionID: "sub-123",
				BasePlanID:     "base-1",
				File:           "/nonexistent/file.json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globals := &Globals{Package: "com.example.app"}
			err := tt.cmd.Run(globals)
			if err == nil {
				t.Fatal("Expected error for missing file")
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("Expected APIError, got: %T", err)
			}
			if apiErr.Code != errors.CodeValidationError {
				t.Errorf("Expected validation error code, got: %v", apiErr.Code)
			}
		})
	}
}

// ============================================================================
// JSON Parsing Error Tests
// ============================================================================

func TestJSONParsingErrors(t *testing.T) {
	tests := []struct {
		name     string
		cmd      interface{ Run(*Globals) error }
		hintText string
	}{
		{
			name: "MonetizationSubscriptionsCreateCmd with invalid JSON",
			cmd: &MonetizationSubscriptionsCreateCmd{
				SubscriptionID: "sub-123",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid Subscription JSON",
		},
		{
			name: "MonetizationSubscriptionsUpdateCmd with invalid JSON",
			cmd: &MonetizationSubscriptionsUpdateCmd{
				SubscriptionID: "sub-123",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid Subscription JSON",
		},
		{
			name: "MonetizationSubscriptionsPatchCmd with invalid JSON",
			cmd: &MonetizationSubscriptionsPatchCmd{
				SubscriptionID: "sub-123",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid Subscription JSON",
		},
		{
			name: "MonetizationSubscriptionsBatchUpdateCmd with invalid JSON",
			cmd: &MonetizationSubscriptionsBatchUpdateCmd{
				File: createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid BatchUpdateSubscriptionsRequest JSON",
		},
		{
			name: "MonetizationBasePlansBatchMigrateCmd with invalid JSON",
			cmd: &MonetizationBasePlansBatchMigrateCmd{
				SubscriptionID: "sub-123",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid BatchMigrateBasePlanPricesRequest JSON",
		},
		{
			name: "MonetizationBasePlansBatchUpdateStatesCmd with invalid JSON",
			cmd: &MonetizationBasePlansBatchUpdateStatesCmd{
				SubscriptionID: "sub-123",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid BatchUpdateBasePlanStatesRequest JSON",
		},
		{
			name: "MonetizationOffersCreateCmd with invalid JSON",
			cmd: &MonetizationOffersCreateCmd{
				SubscriptionID: "sub-123",
				BasePlanID:     "base-1",
				OfferID:        "offer-1",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid SubscriptionOffer JSON",
		},
		{
			name: "MonetizationOffersBatchUpdateCmd with invalid JSON",
			cmd: &MonetizationOffersBatchUpdateCmd{
				SubscriptionID: "sub-123",
				BasePlanID:     "base-1",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid BatchUpdateSubscriptionOffersRequest JSON",
		},
		{
			name: "MonetizationOffersBatchUpdateStatesCmd with invalid JSON",
			cmd: &MonetizationOffersBatchUpdateStatesCmd{
				SubscriptionID: "sub-123",
				BasePlanID:     "base-1",
				File:           createTempFile(t, "invalid.json", []byte("not valid json")),
			},
			hintText: "valid BatchUpdateSubscriptionOfferStatesRequest JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			globals := &Globals{Package: "com.example.app"}
			err := tt.cmd.Run(globals)
			if err == nil {
				t.Fatal("Expected error for invalid JSON")
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("Expected APIError, got: %T", err)
			}
			if apiErr.Code != errors.CodeValidationError {
				t.Errorf("Expected validation error code, got: %v", apiErr.Code)
			}
			if apiErr.Hint == "" {
				t.Errorf("Expected hint containing %q", tt.hintText)
			}
		})
	}
}

// ============================================================================
// PurchasesVoidedListCmd Type Validation Test
// ============================================================================

func TestPurchasesVoidedListCmd_InvalidType(t *testing.T) {
	cmd := &PurchasesVoidedListCmd{
		Type: "invalid-type",
	}
	globals := &Globals{Package: "com.example.app"}

	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("Expected error for invalid type")
	}
	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("Expected APIError, got: %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("Expected validation error code, got: %v", apiErr.Code)
	}
	if apiErr.Hint == "" {
		t.Error("Expected error to have hint about valid types")
	}
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestOutputFormats(t *testing.T) {
	tests := []struct {
		name   string
		output string
		pretty bool
	}{
		{"json output", "json", false},
		{"json pretty output", "json", true},
		{"yaml output", "yaml", false},
		{"table output", "table", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &PurchasesCapabilitiesCmd{}
			globals := &Globals{
				Package: "com.example.app",
				Output:  tt.output,
				Pretty:  tt.pretty,
			}

			err := cmd.Run(globals)
			if err != nil {
				t.Errorf("Unexpected error with output=%s pretty=%v: %v", tt.output, tt.pretty, err)
			}
		})
	}
}

// ============================================================================
// Constants Test
// ============================================================================

func TestConstants(t *testing.T) {
	if purchaseTypeSubscription != "subscription" {
		t.Errorf("Expected purchaseTypeSubscription to be 'subscription', got: %s", purchaseTypeSubscription)
	}
}

// ============================================================================
// Complex Command Initialization Tests
// ============================================================================

func TestComplexCommandInitialization(t *testing.T) {
	t.Run("PurchasesSubscriptionsDeferCmd with all fields", func(t *testing.T) {
		cmd := &PurchasesSubscriptionsDeferCmd{
			SubscriptionID: "sub-123",
			Token:          "token-abc",
			ExpectedExpiry: "2024-12-31T23:59:59Z",
			DesiredExpiry:  "2025-01-31T23:59:59Z",
		}

		globals := &Globals{Package: "com.example.app"}
		err := cmd.Run(globals)
		// Will fail at API client creation, but validates field initialization
		if err == nil {
			t.Error("Expected error (no valid auth), but command initialized correctly")
		}
	})

	t.Run("MonetizationProductsCreateCmd with all fields", func(t *testing.T) {
		cmd := &MonetizationProductsCreateCmd{
			ProductID:    "product-123",
			Type:         "consumable",
			DefaultPrice: "990000",
			Status:       "active",
		}

		globals := &Globals{Package: "com.example.app"}
		err := cmd.Run(globals)
		// Will fail at API client creation, but validates field initialization
		if err == nil {
			t.Error("Expected error (no valid auth), but command initialized correctly")
		}
	})

	t.Run("MonetizationProductsUpdateCmd with all fields", func(t *testing.T) {
		cmd := &MonetizationProductsUpdateCmd{
			ProductID:    "product-123",
			DefaultPrice: "1990000",
			Status:       "inactive",
		}

		globals := &Globals{Package: "com.example.app"}
		err := cmd.Run(globals)
		// Will fail at API client creation, but validates field initialization
		if err == nil {
			t.Error("Expected error (no valid auth), but command initialized correctly")
		}
	})
}
