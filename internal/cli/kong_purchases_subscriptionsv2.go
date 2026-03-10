package cli

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// ============================================================================
// Purchases SubscriptionsV2 Commands
// ============================================================================

// PurchasesSubscriptionsV2Cmd contains subscriptionsv2 purchase actions.
type PurchasesSubscriptionsV2Cmd struct {
	Get    PurchasesSubscriptionsV2GetCmd    `cmd:"" help:"Get a subscription purchase (v2 API)"`
	Cancel PurchasesSubscriptionsV2CancelCmd `cmd:"" help:"Cancel a subscription (v2 API)"`
	Defer  PurchasesSubscriptionsV2DeferCmd  `cmd:"" help:"Defer a subscription renewal (v2 API)"`
	Revoke PurchasesSubscriptionsV2RevokeCmd `cmd:"" help:"Revoke a subscription with refund (v2 API)"`
}

// PurchasesSubscriptionsV2GetCmd gets a subscription purchase using the v2 API.
type PurchasesSubscriptionsV2GetCmd struct {
	Token string `help:"Purchase token" required:""`
}

// Run executes the get subscriptionsv2 command.
func (cmd *PurchasesSubscriptionsV2GetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	var purchase *androidpublisher.SubscriptionPurchaseV2
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		purchase, callErr = svc.Purchases.Subscriptionsv2.Get(globals.Package, cmd.Token).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get subscription purchase: %v", err))
	}

	return outputResult(
		output.NewResult(purchase).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesSubscriptionsV2CancelCmd cancels a subscription using the v2 API.
type PurchasesSubscriptionsV2CancelCmd struct {
	Token            string `help:"Purchase token" required:""`
	CancellationType string `help:"Cancellation type: userRequestedStopRenewals, developerRequestedStopPayments" required:""`
}

// Run executes the cancel subscriptionsv2 command.
func (cmd *PurchasesSubscriptionsV2CancelCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	// Validate cancellation type
	var cancellationType string
	switch cmd.CancellationType {
	case "userRequestedStopRenewals", "user":
		cancellationType = "USER_REQUESTED_STOP_RENEWALS"
	case "developerRequestedStopPayments", "developer":
		cancellationType = "DEVELOPER_REQUESTED_STOP_PAYMENTS"
	default:
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid cancellation type: %s", cmd.CancellationType)).
			WithHint("Valid types: userRequestedStopRenewals (or user), developerRequestedStopPayments (or developer)")
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	req := &androidpublisher.CancelSubscriptionPurchaseRequest{
		CancellationContext: &androidpublisher.CancellationContext{
			CancellationType: cancellationType,
		},
	}

	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		_, callErr = svc.Purchases.Subscriptionsv2.Cancel(globals.Package, cmd.Token, req).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to cancel subscription: %v", err))
	}

	data := map[string]interface{}{
		"token":            cmd.Token,
		"cancelled":        true,
		"cancellationType": cancellationType,
		"apiVersion":       "v2",
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesSubscriptionsV2DeferCmd defers a subscription renewal using the v2 API.
type PurchasesSubscriptionsV2DeferCmd struct {
	Token         string `help:"Purchase token" required:""`
	DeferDuration string `help:"Deferral duration (e.g., '7d', '1w', '30d')" required:""`
	Etag          string `help:"ETag from subscriptionsv2.get (required for consistency)" required:""`
	ValidateOnly  bool   `help:"Dry run - validate only without applying"`
}

// Run executes the defer subscriptionsv2 command.
func (cmd *PurchasesSubscriptionsV2DeferCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	req := &androidpublisher.DeferSubscriptionPurchaseRequest{
		DeferralContext: &androidpublisher.DeferralContext{
			DeferDuration: cmd.DeferDuration,
			Etag:          cmd.Etag,
			ValidateOnly:  cmd.ValidateOnly,
		},
	}

	var resp *androidpublisher.DeferSubscriptionPurchaseResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Purchases.Subscriptionsv2.Defer(globals.Package, cmd.Token, req).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to defer subscription: %v", err))
	}

	data := map[string]interface{}{
		"token":                 cmd.Token,
		"deferred":              true,
		"itemExpiryTimeDetails": resp.ItemExpiryTimeDetails,
		"apiVersion":            "v2",
	}
	if cmd.ValidateOnly {
		data["validatedOnly"] = true
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesSubscriptionsV2RevokeCmd revokes a subscription using the v2 API.
type PurchasesSubscriptionsV2RevokeCmd struct {
	Token      string `help:"Purchase token" required:""`
	RevokeType string `help:"Revoke type: fullRefund, partialRefund (prorated), itemBasedRefund" required:""`
}

// Run executes the revoke subscriptionsv2 command.
func (cmd *PurchasesSubscriptionsV2RevokeCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	revokeReq := &androidpublisher.RevokeSubscriptionPurchaseRequest{}
	switch cmd.RevokeType {
	case "fullRefund":
		revokeReq.RevocationContext = &androidpublisher.RevocationContext{
			FullRefund: &androidpublisher.RevocationContextFullRefund{},
		}
	case "partialRefund", "proratedRefund":
		revokeReq.RevocationContext = &androidpublisher.RevocationContext{
			ProratedRefund: &androidpublisher.RevocationContextProratedRefund{},
		}
	case "itemBasedRefund":
		revokeReq.RevocationContext = &androidpublisher.RevocationContext{
			ItemBasedRefund: &androidpublisher.RevocationContextItemBasedRefund{},
		}
	default:
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid revoke type: %s", cmd.RevokeType)).
			WithHint("Valid types: fullRefund, partialRefund (or proratedRefund), itemBasedRefund")
	}

	var resp *androidpublisher.RevokeSubscriptionPurchaseResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Purchases.Subscriptionsv2.Revoke(globals.Package, cmd.Token, revokeReq).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to revoke subscription: %v", err))
	}

	data := map[string]interface{}{
		"token":      cmd.Token,
		"revoked":    true,
		"revokeType": cmd.RevokeType,
		"apiVersion": "v2",
	}
	_ = resp // response is empty on success

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}
