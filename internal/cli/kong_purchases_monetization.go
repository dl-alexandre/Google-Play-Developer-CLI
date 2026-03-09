package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

const purchaseTypeSubscription = "subscription"

// ============================================================================
// Purchases Commands
// ============================================================================

// PurchasesCmd contains purchase verification commands.
type PurchasesCmd struct {
	Products      PurchasesProductsCmd      `cmd:"" help:"Product purchases"`
	Subscriptions PurchasesSubscriptionsCmd `cmd:"" help:"Subscription purchases"`
	Verify        PurchasesVerifyCmd        `cmd:"" help:"Verify purchase"`
	Voided        PurchasesVoidedCmd        `cmd:"" help:"Voided purchases"`
	Capabilities  PurchasesCapabilitiesCmd  `cmd:"" help:"List purchase verification capabilities"`
}

// ============================================================================
// Purchases Products Commands
// ============================================================================

// PurchasesProductsCmd contains product purchase actions.
type PurchasesProductsCmd struct {
	Acknowledge PurchasesProductsAcknowledgeCmd `cmd:"" help:"Acknowledge a product purchase"`
	Consume     PurchasesProductsConsumeCmd     `cmd:"" help:"Consume a product purchase"`
}

// PurchasesProductsAcknowledgeCmd acknowledges a product purchase.
type PurchasesProductsAcknowledgeCmd struct {
	ProductID        string `help:"Product ID" required:""`
	Token            string `help:"Purchase token" required:""`
	DeveloperPayload string `help:"Developer payload"`
}

// Run executes the acknowledge command.
func (cmd *PurchasesProductsAcknowledgeCmd) Run(globals *Globals) error {
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

	req := &androidpublisher.ProductPurchasesAcknowledgeRequest{
		DeveloperPayload: cmd.DeveloperPayload,
	}

	err = client.DoWithRetry(ctx, func() error {
		return svc.Purchases.Products.Acknowledge(globals.Package, cmd.ProductID, cmd.Token, req).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to acknowledge product purchase: %v", err))
	}

	data := map[string]interface{}{
		"productId":    cmd.ProductID,
		"acknowledged": true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesProductsConsumeCmd consumes a product purchase.
type PurchasesProductsConsumeCmd struct {
	ProductID string `help:"Product ID" required:""`
	Token     string `help:"Purchase token" required:""`
}

// Run executes the consume command.
func (cmd *PurchasesProductsConsumeCmd) Run(globals *Globals) error {
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Purchases.Products.Consume(globals.Package, cmd.ProductID, cmd.Token).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to consume product purchase: %v", err))
	}

	data := map[string]interface{}{
		"productId": cmd.ProductID,
		"consumed":  true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Purchases Subscriptions Commands
// ============================================================================

// PurchasesSubscriptionsCmd contains subscription purchase actions.
type PurchasesSubscriptionsCmd struct {
	Acknowledge PurchasesSubscriptionsAcknowledgeCmd `cmd:"" help:"Acknowledge a subscription purchase"`
	Cancel      PurchasesSubscriptionsCancelCmd      `cmd:"" help:"Cancel a subscription"`
	Defer       PurchasesSubscriptionsDeferCmd       `cmd:"" help:"Defer a subscription renewal"`
	Refund      PurchasesSubscriptionsRefundCmd      `cmd:"" help:"Refund a subscription"`
	Revoke      PurchasesSubscriptionsRevokeCmd      `cmd:"" help:"Revoke a subscription"`
}

// PurchasesSubscriptionsAcknowledgeCmd acknowledges a subscription purchase.
type PurchasesSubscriptionsAcknowledgeCmd struct {
	SubscriptionID   string `help:"Subscription ID" required:""`
	Token            string `help:"Purchase token" required:""`
	DeveloperPayload string `help:"Developer payload"`
}

// Run executes the acknowledge subscription command.
func (cmd *PurchasesSubscriptionsAcknowledgeCmd) Run(globals *Globals) error {
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

	req := &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{
		DeveloperPayload: cmd.DeveloperPayload,
	}

	err = client.DoWithRetry(ctx, func() error {
		return svc.Purchases.Subscriptions.Acknowledge(globals.Package, cmd.SubscriptionID, cmd.Token, req).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to acknowledge subscription purchase: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId": cmd.SubscriptionID,
		"acknowledged":   true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesSubscriptionsCancelCmd cancels a subscription.
type PurchasesSubscriptionsCancelCmd struct {
	SubscriptionID string `help:"Subscription ID" required:""`
	Token          string `help:"Purchase token" required:""`
}

// Run executes the cancel subscription command.
func (cmd *PurchasesSubscriptionsCancelCmd) Run(globals *Globals) error {
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Purchases.Subscriptions.Cancel(globals.Package, cmd.SubscriptionID, cmd.Token).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to cancel subscription: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId": cmd.SubscriptionID,
		"cancelled":      true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesSubscriptionsDeferCmd defers a subscription renewal.
type PurchasesSubscriptionsDeferCmd struct {
	SubscriptionID string `help:"Subscription ID" required:""`
	Token          string `help:"Purchase token" required:""`
	ExpectedExpiry string `help:"Expected expiry time (RFC3339 or millis)" required:""`
	DesiredExpiry  string `help:"Desired expiry time (RFC3339 or millis)" required:""`
}

// Run executes the defer subscription command.
func (cmd *PurchasesSubscriptionsDeferCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	expectedMillis, err := parseTimeToMillis(cmd.ExpectedExpiry)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid expected expiry time: %v", err)).
			WithHint("Provide time as RFC3339 (e.g., 2024-01-01T00:00:00Z) or milliseconds since epoch")
	}

	desiredMillis, err := parseTimeToMillis(cmd.DesiredExpiry)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid desired expiry time: %v", err)).
			WithHint("Provide time as RFC3339 (e.g., 2024-01-01T00:00:00Z) or milliseconds since epoch")
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

	req := &androidpublisher.SubscriptionPurchasesDeferRequest{
		DeferralInfo: &androidpublisher.SubscriptionDeferralInfo{
			ExpectedExpiryTimeMillis: expectedMillis,
			DesiredExpiryTimeMillis:  desiredMillis,
		},
	}

	var resp *androidpublisher.SubscriptionPurchasesDeferResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Purchases.Subscriptions.Defer(globals.Package, cmd.SubscriptionID, cmd.Token, req).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to defer subscription: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId":      cmd.SubscriptionID,
		"deferred":            true,
		"newExpiryTimeMillis": resp.NewExpiryTimeMillis,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesSubscriptionsRefundCmd refunds a subscription.
type PurchasesSubscriptionsRefundCmd struct {
	SubscriptionID string `help:"Subscription ID" required:""`
	Token          string `help:"Purchase token" required:""`
}

// Run executes the refund subscription command.
func (cmd *PurchasesSubscriptionsRefundCmd) Run(globals *Globals) error {
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Purchases.Subscriptions.Refund(globals.Package, cmd.SubscriptionID, cmd.Token).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to refund subscription: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId": cmd.SubscriptionID,
		"refunded":       true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// PurchasesSubscriptionsRevokeCmd revokes a subscription.
type PurchasesSubscriptionsRevokeCmd struct {
	SubscriptionID string `help:"Subscription ID (v1 API)"`
	Token          string `help:"Purchase token" required:""`
	RevokeType     string `help:"Revoke type for v2: fullRefund, partialRefund, itemBasedRefund"`
}

// Run executes the revoke subscription command.
func (cmd *PurchasesSubscriptionsRevokeCmd) Run(globals *Globals) error {
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

	// Use v2 API if RevokeType is specified, otherwise fall back to v1
	if cmd.RevokeType != "" {
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
		default:
			return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid revoke type: %s", cmd.RevokeType)).
				WithHint("Valid types: fullRefund, partialRefund")
		}

		var resp *androidpublisher.RevokeSubscriptionPurchaseResponse
		err = client.DoWithRetry(ctx, func() error {
			var callErr error
			resp, callErr = svc.Purchases.Subscriptionsv2.Revoke(globals.Package, cmd.Token, revokeReq).Context(ctx).Do()
			return callErr
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to revoke subscription (v2): %v", err))
		}

		data := map[string]interface{}{
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

	// v1 API requires SubscriptionID
	if cmd.SubscriptionID == "" {
		return errors.NewAPIError(errors.CodeValidationError, "subscription ID is required for v1 revoke").
			WithHint("Provide --subscription-id or use --revoke-type for v2 API")
	}

	err = client.DoWithRetry(ctx, func() error {
		return svc.Purchases.Subscriptions.Revoke(globals.Package, cmd.SubscriptionID, cmd.Token).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to revoke subscription: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId": cmd.SubscriptionID,
		"revoked":        true,
		"apiVersion":     "v1",
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Purchases Verify Command
// ============================================================================

// PurchasesVerifyCmd verifies a purchase token.
type PurchasesVerifyCmd struct {
	ProductID   string `help:"Product ID"`
	Token       string `help:"Purchase token" required:""`
	Environment string `help:"Environment: sandbox, production, auto" default:"auto"`
	Type        string `help:"Product type: product, subscription, auto" default:"auto"`
}

// Run executes the verify command.
func (cmd *PurchasesVerifyCmd) Run(globals *Globals) error {
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

	switch cmd.Type {
	case "product":
		if cmd.ProductID == "" {
			return errors.NewAPIError(errors.CodeValidationError, "product ID is required when type is 'product'").
				WithHint("Provide --product-id flag")
		}
		return cmd.verifyProduct(ctx, client, svc, globals, start)

	case purchaseTypeSubscription:
		return cmd.verifySubscription(ctx, client, svc, globals, start)

	case "auto", "":
		// Try product first if ProductID is provided, then fall back to subscription
		if cmd.ProductID != "" {
			result := cmd.verifyProduct(ctx, client, svc, globals, start)
			if result == nil {
				return nil
			}
		}
		// Try subscription v2
		return cmd.verifySubscription(ctx, client, svc, globals, start)

	default:
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid type: %s", cmd.Type)).
			WithHint("Valid types: product, subscription, auto")
	}
}

func (cmd *PurchasesVerifyCmd) verifyProduct(ctx context.Context, client interface {
	DoWithRetry(context.Context, func() error) error
}, svc *androidpublisher.Service, globals *Globals, start time.Time) error {
	var purchase *androidpublisher.ProductPurchase
	err := client.DoWithRetry(ctx, func() error {
		var callErr error
		purchase, callErr = svc.Purchases.Products.Get(globals.Package, cmd.ProductID, cmd.Token).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to verify product purchase: %v", err))
	}

	data := map[string]interface{}{
		"type":                 "product",
		"productId":            cmd.ProductID,
		"purchaseState":        purchase.PurchaseState,
		"consumptionState":     purchase.ConsumptionState,
		"acknowledgementState": purchase.AcknowledgementState,
		"orderId":              purchase.OrderId,
		"purchaseTimeMillis":   purchase.PurchaseTimeMillis,
		"kind":                 purchase.Kind,
		"developerPayload":     purchase.DeveloperPayload,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

func (cmd *PurchasesVerifyCmd) verifySubscription(ctx context.Context, client interface {
	DoWithRetry(context.Context, func() error) error
}, svc *androidpublisher.Service, globals *Globals, start time.Time) error {
	var purchase *androidpublisher.SubscriptionPurchaseV2
	err := client.DoWithRetry(ctx, func() error {
		var callErr error
		purchase, callErr = svc.Purchases.Subscriptionsv2.Get(globals.Package, cmd.Token).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to verify subscription purchase: %v", err))
	}

	data := map[string]interface{}{
		"type":                 purchaseTypeSubscription,
		"acknowledgementState": purchase.AcknowledgementState,
		"subscriptionState":    purchase.SubscriptionState,
		"latestOrderId":        purchase.LatestOrderId,
		"linkedPurchaseToken":  purchase.LinkedPurchaseToken,
		"kind":                 purchase.Kind,
	}
	if purchase.StartTime != "" {
		data["startTime"] = purchase.StartTime
	}
	if purchase.RegionCode != "" {
		data["regionCode"] = purchase.RegionCode
	}
	if len(purchase.LineItems) > 0 {
		data["lineItems"] = purchase.LineItems
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Purchases Voided Commands
// ============================================================================

// PurchasesVoidedCmd contains voided purchase commands.
type PurchasesVoidedCmd struct {
	List PurchasesVoidedListCmd `cmd:"" help:"List voided purchases"`
}

// PurchasesVoidedListCmd lists voided purchases.
type PurchasesVoidedListCmd struct {
	StartTime  string `help:"Start time (RFC3339 or millis)"`
	EndTime    string `help:"End time (RFC3339 or millis)"`
	Type       string `help:"Type: product or subscription"`
	MaxResults int64  `help:"Max results per page"`
	PageToken  string `help:"Pagination token"`
	All        bool   `help:"Fetch all pages"`
}

// voidedPurchasesPageResponse wraps the voided purchases list response for pagination.
type voidedPurchasesPageResponse struct {
	resp *androidpublisher.VoidedPurchasesListResponse
}

func (r voidedPurchasesPageResponse) GetNextPageToken() string {
	if r.resp.TokenPagination != nil {
		return r.resp.TokenPagination.NextPageToken
	}
	return ""
}

func (r voidedPurchasesPageResponse) GetItems() []*androidpublisher.VoidedPurchase {
	return r.resp.VoidedPurchases
}

// Run executes the voided list command.
func (cmd *PurchasesVoidedListCmd) Run(globals *Globals) error {
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

	var allVoided []*androidpublisher.VoidedPurchase
	var nextPageToken string

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Purchases.Voidedpurchases.List(globals.Package).Context(ctx)

		if cmd.StartTime != "" {
			millis, parseErr := parseTimeToMillis(cmd.StartTime)
			if parseErr != nil {
				return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid start time: %v", parseErr))
			}
			call = call.StartTime(millis)
		}
		if cmd.EndTime != "" {
			millis, parseErr := parseTimeToMillis(cmd.EndTime)
			if parseErr != nil {
				return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid end time: %v", parseErr))
			}
			call = call.EndTime(millis)
		}
		if cmd.MaxResults > 0 {
			call = call.MaxResults(cmd.MaxResults)
		}
		if cmd.PageToken != "" {
			call = call.Token(cmd.PageToken)
		}
		if cmd.Type != "" {
			var typeVal int64
			switch cmd.Type {
			case "product":
				typeVal = 0
			case purchaseTypeSubscription:
				typeVal = 1
			default:
				return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid type: %s", cmd.Type)).
					WithHint("Valid types: product, subscription")
			}
			call = call.Type(typeVal)
		}

		resp, callErr := call.Do()
		if callErr != nil {
			return callErr
		}

		allVoided = append(allVoided, resp.VoidedPurchases...)
		if resp.TokenPagination != nil {
			nextPageToken = resp.TokenPagination.NextPageToken
		}

		if cmd.All && nextPageToken != "" {
			query := func(pageToken string) (voidedPurchasesPageResponse, error) {
				pageCall := svc.Purchases.Voidedpurchases.List(globals.Package).
					Token(pageToken).Context(ctx)
				if cmd.StartTime != "" {
					millis, _ := parseTimeToMillis(cmd.StartTime)
					pageCall = pageCall.StartTime(millis)
				}
				if cmd.EndTime != "" {
					millis, _ := parseTimeToMillis(cmd.EndTime)
					pageCall = pageCall.EndTime(millis)
				}
				if cmd.MaxResults > 0 {
					pageCall = pageCall.MaxResults(cmd.MaxResults)
				}
				pageResp, pageErr := pageCall.Do()
				return voidedPurchasesPageResponse{resp: pageResp}, pageErr
			}

			additionalItems, remainingToken, fetchErr := fetchAllPages(ctx, query, nextPageToken, 0)
			if fetchErr != nil {
				return fetchErr
			}
			allVoided = append(allVoided, additionalItems...)
			nextPageToken = remainingToken
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list voided purchases: %v", err))
	}

	data := map[string]interface{}{
		"voidedPurchases": allVoided,
		"totalCount":      len(allVoided),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	if nextPageToken != "" {
		result = result.WithPagination(cmd.PageToken, nextPageToken)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// PurchasesCapabilitiesCmd lists purchase verification capabilities.
type PurchasesCapabilitiesCmd struct{}

// Run executes the capabilities command.
func (cmd *PurchasesCapabilitiesCmd) Run(globals *Globals) error {
	start := time.Now()

	data := map[string]interface{}{
		"capabilities": []map[string]interface{}{
			{
				"name":        "products.acknowledge",
				"description": "Acknowledge a product purchase",
				"apiVersion":  "v3",
			},
			{
				"name":        "products.consume",
				"description": "Consume a product purchase",
				"apiVersion":  "v3",
			},
			{
				"name":        "products.get",
				"description": "Get product purchase details (verify)",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptions.acknowledge",
				"description": "Acknowledge a subscription purchase",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptions.cancel",
				"description": "Cancel a subscription",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptions.defer",
				"description": "Defer a subscription renewal",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptions.get",
				"description": "Get subscription purchase details (v1)",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptions.refund",
				"description": "Refund a subscription",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptions.revoke",
				"description": "Revoke a subscription (v1)",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptionsv2.get",
				"description": "Get subscription purchase details (v2)",
				"apiVersion":  "v3",
			},
			{
				"name":        "subscriptionsv2.revoke",
				"description": "Revoke a subscription with refund context (v2)",
				"apiVersion":  "v3",
			},
			{
				"name":        "voidedpurchases.list",
				"description": "List voided purchases",
				"apiVersion":  "v3",
			},
		},
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Monetization Commands
// ============================================================================

// MonetizationCmd contains monetization commands.
type MonetizationCmd struct {
	Products        MonetizationProductsCmd        `cmd:"" help:"Manage products"`
	Subscriptions   MonetizationSubscriptionsCmd   `cmd:"" help:"Manage subscriptions"`
	OneTimeProducts MonetizationOneTimeProductsCmd `cmd:"" help:"Manage one-time products"`
	BasePlans       MonetizationBasePlansCmd       `cmd:"" help:"Manage base plans"`
	Offers          MonetizationOffersCmd          `cmd:"" help:"Manage offers"`
	Capabilities    MonetizationCapabilitiesCmd    `cmd:"" help:"List monetization capabilities"`
}

// ============================================================================
// Monetization Products Commands
// ============================================================================

// MonetizationProductsCmd contains product management commands.
type MonetizationProductsCmd struct {
	List   MonetizationProductsListCmd   `cmd:"" help:"List in-app products"`
	Get    MonetizationProductsGetCmd    `cmd:"" help:"Get an in-app product"`
	Create MonetizationProductsCreateCmd `cmd:"" help:"Create an in-app product"`
	Update MonetizationProductsUpdateCmd `cmd:"" help:"Update an in-app product"`
	Delete MonetizationProductsDeleteCmd `cmd:"" help:"Delete an in-app product"`
}

// MonetizationProductsListCmd lists in-app products.
type MonetizationProductsListCmd struct {
	PageSize  int64  `help:"Results per page" default:"100"`
	PageToken string `help:"Pagination token"`
	All       bool   `help:"Fetch all pages"`
}

// inappProductsPageResponse wraps the in-app products list response for pagination.
type inappProductsPageResponse struct {
	resp *androidpublisher.InappproductsListResponse
}

func (r inappProductsPageResponse) GetNextPageToken() string {
	if r.resp.TokenPagination != nil {
		return r.resp.TokenPagination.NextPageToken
	}
	return ""
}

func (r inappProductsPageResponse) GetItems() []*androidpublisher.InAppProduct {
	return r.resp.Inappproduct
}

// Run executes the list products command.
func (cmd *MonetizationProductsListCmd) Run(globals *Globals) error {
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

	var allProducts []*androidpublisher.InAppProduct
	var nextPageToken string

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Inappproducts.List(globals.Package).Context(ctx)
		if cmd.PageSize > 0 {
			call = call.MaxResults(cmd.PageSize)
		}
		if cmd.PageToken != "" {
			call = call.Token(cmd.PageToken)
		}

		resp, callErr := call.Do()
		if callErr != nil {
			return callErr
		}

		allProducts = append(allProducts, resp.Inappproduct...)
		if resp.TokenPagination != nil {
			nextPageToken = resp.TokenPagination.NextPageToken
		}

		if cmd.All && nextPageToken != "" {
			query := func(pageToken string) (inappProductsPageResponse, error) {
				pageCall := svc.Inappproducts.List(globals.Package).
					Token(pageToken).Context(ctx)
				if cmd.PageSize > 0 {
					pageCall = pageCall.MaxResults(cmd.PageSize)
				}
				pageResp, pageErr := pageCall.Do()
				return inappProductsPageResponse{resp: pageResp}, pageErr
			}

			additionalItems, remainingToken, fetchErr := fetchAllPages(ctx, query, nextPageToken, 0)
			if fetchErr != nil {
				return fetchErr
			}
			allProducts = append(allProducts, additionalItems...)
			nextPageToken = remainingToken
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list in-app products: %v", err))
	}

	data := map[string]interface{}{
		"products":   allProducts,
		"totalCount": len(allProducts),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	if nextPageToken != "" {
		result = result.WithPagination(cmd.PageToken, nextPageToken)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// MonetizationProductsGetCmd gets an in-app product.
type MonetizationProductsGetCmd struct {
	ProductID string `arg:"" help:"Product ID (SKU)" required:""`
}

// Run executes the get product command.
func (cmd *MonetizationProductsGetCmd) Run(globals *Globals) error {
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

	var product *androidpublisher.InAppProduct
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		product, callErr = svc.Inappproducts.Get(globals.Package, cmd.ProductID).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get in-app product: %v", err))
	}

	return outputResult(
		output.NewResult(product).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationProductsCreateCmd creates an in-app product.
type MonetizationProductsCreateCmd struct {
	ProductID    string `help:"Product SKU" required:""`
	Type         string `help:"Product type: managed, consumable" default:"managed"`
	DefaultPrice string `help:"Default price in micros (e.g., 990000 for $0.99)"`
	Status       string `help:"Product status: active, inactive" default:"active"`
}

// Run executes the create product command.
func (cmd *MonetizationProductsCreateCmd) Run(globals *Globals) error {
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

	product := &androidpublisher.InAppProduct{
		PackageName: globals.Package,
		Sku:         cmd.ProductID,
		Status:      cmd.Status,
	}

	// Map user-friendly type to API type
	switch cmd.Type {
	case "managed", "consumable":
		product.PurchaseType = "managedUser"
	case purchaseTypeSubscription:
		product.PurchaseType = purchaseTypeSubscription
	default:
		product.PurchaseType = "managedUser"
	}

	if cmd.DefaultPrice != "" {
		product.DefaultPrice = &androidpublisher.Price{
			PriceMicros: cmd.DefaultPrice,
			Currency:    "USD",
		}
	}

	var created *androidpublisher.InAppProduct
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		created, callErr = svc.Inappproducts.Insert(globals.Package, product).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create in-app product: %v", err))
	}

	return outputResult(
		output.NewResult(created).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationProductsUpdateCmd updates an in-app product.
type MonetizationProductsUpdateCmd struct {
	ProductID    string `arg:"" help:"Product ID (SKU)" required:""`
	DefaultPrice string `help:"Default price in micros"`
	Status       string `help:"Product status: active, inactive"`
}

// Run executes the update product command.
func (cmd *MonetizationProductsUpdateCmd) Run(globals *Globals) error {
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

	// Get existing product first
	var existing *androidpublisher.InAppProduct
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		existing, callErr = svc.Inappproducts.Get(globals.Package, cmd.ProductID).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get existing product: %v", err))
	}

	// Apply updates
	if cmd.DefaultPrice != "" {
		existing.DefaultPrice = &androidpublisher.Price{
			PriceMicros: cmd.DefaultPrice,
			Currency:    existing.DefaultPrice.Currency,
		}
	}
	if cmd.Status != "" {
		existing.Status = cmd.Status
	}

	var updated *androidpublisher.InAppProduct
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		updated, callErr = svc.Inappproducts.Update(globals.Package, cmd.ProductID, existing).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update in-app product: %v", err))
	}

	return outputResult(
		output.NewResult(updated).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationProductsDeleteCmd deletes an in-app product.
type MonetizationProductsDeleteCmd struct {
	ProductID string `arg:"" help:"Product ID (SKU)" required:""`
}

// Run executes the delete product command.
func (cmd *MonetizationProductsDeleteCmd) Run(globals *Globals) error {
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Inappproducts.Delete(globals.Package, cmd.ProductID).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete in-app product: %v", err))
	}

	data := map[string]interface{}{
		"productId": cmd.ProductID,
		"deleted":   true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Monetization Subscriptions Commands
// ============================================================================

// MonetizationSubscriptionsCmd contains subscription management commands.
type MonetizationSubscriptionsCmd struct {
	List        MonetizationSubscriptionsListCmd        `cmd:"" help:"List subscription products"`
	Get         MonetizationSubscriptionsGetCmd         `cmd:"" help:"Get a subscription product"`
	Create      MonetizationSubscriptionsCreateCmd      `cmd:"" help:"Create a subscription"`
	Update      MonetizationSubscriptionsUpdateCmd      `cmd:"" help:"Update a subscription"`
	Patch       MonetizationSubscriptionsPatchCmd       `cmd:"" help:"Patch a subscription"`
	Delete      MonetizationSubscriptionsDeleteCmd      `cmd:"" help:"Delete a subscription"`
	Archive     MonetizationSubscriptionsArchiveCmd     `cmd:"" help:"Archive a subscription"`
	BatchGet    MonetizationSubscriptionsBatchGetCmd    `cmd:"" help:"Batch get subscriptions"`
	BatchUpdate MonetizationSubscriptionsBatchUpdateCmd `cmd:"" help:"Batch update subscriptions"`
}

// MonetizationSubscriptionsListCmd lists subscription products.
type MonetizationSubscriptionsListCmd struct {
	PageSize     int64  `help:"Results per page" default:"100"`
	PageToken    string `help:"Pagination token"`
	All          bool   `help:"Fetch all pages"`
	ShowArchived bool   `help:"Include archived subscriptions"`
}

// subscriptionsPageResponse wraps the subscriptions list response for pagination.
type subscriptionsPageResponse struct {
	resp *androidpublisher.ListSubscriptionsResponse
}

func (r subscriptionsPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r subscriptionsPageResponse) GetItems() []*androidpublisher.Subscription {
	return r.resp.Subscriptions
}

// Run executes the list subscriptions command.
func (cmd *MonetizationSubscriptionsListCmd) Run(globals *Globals) error {
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

	var allSubscriptions []*androidpublisher.Subscription
	var nextPageToken string

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Monetization.Subscriptions.List(globals.Package).Context(ctx)
		if cmd.PageSize > 0 {
			call = call.PageSize(cmd.PageSize)
		}
		if cmd.PageToken != "" {
			call = call.PageToken(cmd.PageToken)
		}
		if cmd.ShowArchived {
			call = call.ShowArchived(true)
		}

		resp, callErr := call.Do()
		if callErr != nil {
			return callErr
		}

		allSubscriptions = append(allSubscriptions, resp.Subscriptions...)
		nextPageToken = resp.NextPageToken

		if cmd.All && nextPageToken != "" {
			query := func(pageToken string) (subscriptionsPageResponse, error) {
				pageCall := svc.Monetization.Subscriptions.List(globals.Package).
					PageToken(pageToken).Context(ctx)
				if cmd.PageSize > 0 {
					pageCall = pageCall.PageSize(cmd.PageSize)
				}
				if cmd.ShowArchived {
					pageCall = pageCall.ShowArchived(true)
				}
				pageResp, pageErr := pageCall.Do()
				return subscriptionsPageResponse{resp: pageResp}, pageErr
			}

			additionalItems, remainingToken, fetchErr := fetchAllPages(ctx, query, nextPageToken, 0)
			if fetchErr != nil {
				return fetchErr
			}
			allSubscriptions = append(allSubscriptions, additionalItems...)
			nextPageToken = remainingToken
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list subscriptions: %v", err))
	}

	data := map[string]interface{}{
		"subscriptions": allSubscriptions,
		"totalCount":    len(allSubscriptions),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	if nextPageToken != "" {
		result = result.WithPagination(cmd.PageToken, nextPageToken)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// MonetizationSubscriptionsGetCmd gets a subscription product.
type MonetizationSubscriptionsGetCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
}

// Run executes the get subscription command.
func (cmd *MonetizationSubscriptionsGetCmd) Run(globals *Globals) error {
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

	var subscription *androidpublisher.Subscription
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		subscription, callErr = svc.Monetization.Subscriptions.Get(globals.Package, cmd.SubscriptionID).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get subscription: %v", err))
	}

	return outputResult(
		output.NewResult(subscription).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationSubscriptionsCreateCmd creates a subscription.
type MonetizationSubscriptionsCreateCmd struct {
	SubscriptionID string `help:"Subscription product ID" required:""`
	File           string `help:"Subscription JSON file" required:"" type:"existingfile"`
}

// Run executes the create subscription command.
func (cmd *MonetizationSubscriptionsCreateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var subscription androidpublisher.Subscription
	if err := json.Unmarshal(fileData, &subscription); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse subscription JSON: %v", err)).
			WithHint("Ensure the file contains valid Subscription JSON")
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

	var created *androidpublisher.Subscription
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		created, callErr = svc.Monetization.Subscriptions.Create(globals.Package, &subscription).
			ProductId(cmd.SubscriptionID).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create subscription: %v", err))
	}

	return outputResult(
		output.NewResult(created).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationSubscriptionsUpdateCmd updates a subscription.
type MonetizationSubscriptionsUpdateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	File           string `help:"Subscription JSON file" required:"" type:"existingfile"`
}

// Run executes the update subscription command.
func (cmd *MonetizationSubscriptionsUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var subscription androidpublisher.Subscription
	if err := json.Unmarshal(fileData, &subscription); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse subscription JSON: %v", err)).
			WithHint("Ensure the file contains valid Subscription JSON")
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

	var updated *androidpublisher.Subscription
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		updated, callErr = svc.Monetization.Subscriptions.Patch(globals.Package, cmd.SubscriptionID, &subscription).
			Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update subscription: %v", err))
	}

	return outputResult(
		output.NewResult(updated).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationSubscriptionsPatchCmd patches a subscription.
type MonetizationSubscriptionsPatchCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	File           string `help:"Subscription JSON file" required:"" type:"existingfile"`
	UpdateMask     string `help:"Fields to update (comma-separated)"`
	AllowMissing   bool   `help:"Create if missing"`
}

// Run executes the patch subscription command.
func (cmd *MonetizationSubscriptionsPatchCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var subscription androidpublisher.Subscription
	if err := json.Unmarshal(fileData, &subscription); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse subscription JSON: %v", err)).
			WithHint("Ensure the file contains valid Subscription JSON")
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

	var patched *androidpublisher.Subscription
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		call := svc.Monetization.Subscriptions.Patch(globals.Package, cmd.SubscriptionID, &subscription).
			Context(ctx)
		if cmd.UpdateMask != "" {
			call = call.UpdateMask(cmd.UpdateMask)
		}
		if cmd.AllowMissing {
			call = call.AllowMissing(true)
		}
		patched, callErr = call.Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to patch subscription: %v", err))
	}

	return outputResult(
		output.NewResult(patched).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationSubscriptionsDeleteCmd deletes a subscription.
type MonetizationSubscriptionsDeleteCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	Confirm        bool   `help:"Confirm destructive operation" required:""`
}

// Run executes the delete subscription command.
func (cmd *MonetizationSubscriptionsDeleteCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "deletion requires confirmation").
			WithHint("Pass --confirm to confirm the destructive operation")
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Monetization.Subscriptions.Delete(globals.Package, cmd.SubscriptionID).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete subscription: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId": cmd.SubscriptionID,
		"deleted":        true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationSubscriptionsArchiveCmd archives a subscription.
type MonetizationSubscriptionsArchiveCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
}

// Run executes the archive subscription command.
func (cmd *MonetizationSubscriptionsArchiveCmd) Run(globals *Globals) error {
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

	var archived *androidpublisher.Subscription
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		archived, callErr = svc.Monetization.Subscriptions.Archive(globals.Package, cmd.SubscriptionID,
			&androidpublisher.ArchiveSubscriptionRequest{}).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to archive subscription: %v", err))
	}

	return outputResult(
		output.NewResult(archived).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationSubscriptionsBatchGetCmd batch gets subscriptions.
type MonetizationSubscriptionsBatchGetCmd struct {
	IDs []string `help:"Subscription IDs" required:""`
}

// Run executes the batch get subscriptions command.
func (cmd *MonetizationSubscriptionsBatchGetCmd) Run(globals *Globals) error {
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

	var resp *androidpublisher.BatchGetSubscriptionsResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BatchGet(globals.Package).
			ProductIds(cmd.IDs...).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch get subscriptions: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationSubscriptionsBatchUpdateCmd batch updates subscriptions.
type MonetizationSubscriptionsBatchUpdateCmd struct {
	File string `help:"Batch update JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update subscriptions command.
func (cmd *MonetizationSubscriptionsBatchUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var batchReq androidpublisher.BatchUpdateSubscriptionsRequest
	if err := json.Unmarshal(fileData, &batchReq); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse batch update JSON: %v", err)).
			WithHint("Ensure the file contains valid BatchUpdateSubscriptionsRequest JSON")
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

	var resp *androidpublisher.BatchUpdateSubscriptionsResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BatchUpdate(globals.Package, &batchReq).
			Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch update subscriptions: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Monetization OneTimeProducts Commands
// ============================================================================

// MonetizationOneTimeProductsCmd contains one-time product management commands.
type MonetizationOneTimeProductsCmd struct {
	List        MonetizationOneTimeProductsListCmd        `cmd:"" help:"List one-time products"`
	Get         MonetizationOneTimeProductsGetCmd         `cmd:"" help:"Get a one-time product"`
	Create      MonetizationOneTimeProductsCreateCmd      `cmd:"" help:"Create a one-time product"`
	Update      MonetizationOneTimeProductsUpdateCmd      `cmd:"" help:"Update a one-time product"`
	Delete      MonetizationOneTimeProductsDeleteCmd      `cmd:"" help:"Delete a one-time product"`
	BatchGet    MonetizationOneTimeProductsBatchGetCmd    `cmd:"" help:"Batch get one-time products"`
	BatchUpdate MonetizationOneTimeProductsBatchUpdateCmd `cmd:"" help:"Batch update one-time products"`
}

// MonetizationOneTimeProductsListCmd lists one-time products.
type MonetizationOneTimeProductsListCmd struct {
	PageSize  int64  `help:"Results per page" default:"100"`
	PageToken string `help:"Pagination token"`
	All       bool   `help:"Fetch all pages"`
}

// oneTimeProductsPageResponse wraps the one-time products list response for pagination.
type oneTimeProductsPageResponse struct {
	resp *androidpublisher.ListOneTimeProductsResponse
}

func (r oneTimeProductsPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r oneTimeProductsPageResponse) GetItems() []*androidpublisher.OneTimeProduct {
	return r.resp.OneTimeProducts
}

// Run executes the list one-time products command.
func (cmd *MonetizationOneTimeProductsListCmd) Run(globals *Globals) error {
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

	var allProducts []*androidpublisher.OneTimeProduct
	var nextPageToken string

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Monetization.Onetimeproducts.List(globals.Package).Context(ctx)
		if cmd.PageSize > 0 {
			call = call.PageSize(cmd.PageSize)
		}
		if cmd.PageToken != "" {
			call = call.PageToken(cmd.PageToken)
		}

		resp, callErr := call.Do()
		if callErr != nil {
			return callErr
		}

		allProducts = append(allProducts, resp.OneTimeProducts...)
		nextPageToken = resp.NextPageToken

		if cmd.All && nextPageToken != "" {
			query := func(pageToken string) (oneTimeProductsPageResponse, error) {
				pageCall := svc.Monetization.Onetimeproducts.List(globals.Package).
					PageToken(pageToken).Context(ctx)
				if cmd.PageSize > 0 {
					pageCall = pageCall.PageSize(cmd.PageSize)
				}
				pageResp, pageErr := pageCall.Do()
				return oneTimeProductsPageResponse{resp: pageResp}, pageErr
			}

			additionalItems, remainingToken, fetchErr := fetchAllPages(ctx, query, nextPageToken, 0)
			if fetchErr != nil {
				return fetchErr
			}
			allProducts = append(allProducts, additionalItems...)
			nextPageToken = remainingToken
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list one-time products: %v", err))
	}

	data := map[string]interface{}{
		"oneTimeProducts": allProducts,
		"totalCount":      len(allProducts),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	if nextPageToken != "" {
		result = result.WithPagination(cmd.PageToken, nextPageToken)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// MonetizationOneTimeProductsGetCmd gets a one-time product.
type MonetizationOneTimeProductsGetCmd struct {
	ProductID string `arg:"" help:"Product ID" required:""`
}

// Run executes the get one-time product command.
func (cmd *MonetizationOneTimeProductsGetCmd) Run(globals *Globals) error {
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

	var product *androidpublisher.OneTimeProduct
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		product, callErr = svc.Monetization.Onetimeproducts.Get(globals.Package, cmd.ProductID).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get one-time product: %v", err))
	}

	return outputResult(
		output.NewResult(product).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOneTimeProductsCreateCmd creates a one-time product.
type MonetizationOneTimeProductsCreateCmd struct {
	ProductID string `help:"Product ID" required:""`
	File      string `help:"One-time product JSON file" required:"" type:"existingfile"`
}

// Run executes the create one-time product command.
func (cmd *MonetizationOneTimeProductsCreateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var product androidpublisher.OneTimeProduct
	if err := json.Unmarshal(fileData, &product); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse one-time product JSON: %v", err)).
			WithHint("Ensure the file contains valid OneTimeProduct JSON")
	}

	product.PackageName = globals.Package
	product.ProductId = cmd.ProductID

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	var created *androidpublisher.OneTimeProduct
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		created, callErr = svc.Monetization.Onetimeproducts.Patch(globals.Package, cmd.ProductID, &product).
			Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create one-time product: %v", err))
	}

	return outputResult(
		output.NewResult(created).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOneTimeProductsUpdateCmd updates a one-time product.
type MonetizationOneTimeProductsUpdateCmd struct {
	ProductID string `arg:"" help:"Product ID" required:""`
	File      string `help:"One-time product JSON file" required:"" type:"existingfile"`
}

// Run executes the update one-time product command.
func (cmd *MonetizationOneTimeProductsUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var product androidpublisher.OneTimeProduct
	if err := json.Unmarshal(fileData, &product); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse one-time product JSON: %v", err)).
			WithHint("Ensure the file contains valid OneTimeProduct JSON")
	}

	product.PackageName = globals.Package
	product.ProductId = cmd.ProductID

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	var updated *androidpublisher.OneTimeProduct
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		updated, callErr = svc.Monetization.Onetimeproducts.Patch(globals.Package, cmd.ProductID, &product).
			Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update one-time product: %v", err))
	}

	return outputResult(
		output.NewResult(updated).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOneTimeProductsDeleteCmd deletes a one-time product.
type MonetizationOneTimeProductsDeleteCmd struct {
	ProductID string `arg:"" help:"Product ID" required:""`
}

// Run executes the delete one-time product command.
func (cmd *MonetizationOneTimeProductsDeleteCmd) Run(globals *Globals) error {
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Monetization.Onetimeproducts.Delete(globals.Package, cmd.ProductID).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete one-time product: %v", err))
	}

	data := map[string]interface{}{
		"productId": cmd.ProductID,
		"deleted":   true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOneTimeProductsBatchGetCmd batch gets one-time products.
type MonetizationOneTimeProductsBatchGetCmd struct {
	IDs []string `help:"Product IDs" required:""`
}

// Run executes the batch get one-time products command.
func (cmd *MonetizationOneTimeProductsBatchGetCmd) Run(globals *Globals) error {
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

	var resp *androidpublisher.BatchGetOneTimeProductsResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Onetimeproducts.BatchGet(globals.Package).
			ProductIds(cmd.IDs...).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch get one-time products: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOneTimeProductsBatchUpdateCmd batch updates one-time products.
type MonetizationOneTimeProductsBatchUpdateCmd struct {
	File string `help:"Batch update JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update one-time products command.
func (cmd *MonetizationOneTimeProductsBatchUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var batchReq androidpublisher.BatchUpdateOneTimeProductsRequest
	if err := json.Unmarshal(fileData, &batchReq); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse batch update JSON: %v", err)).
			WithHint("Ensure the file contains valid BatchUpdateOneTimeProductsRequest JSON")
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

	var resp *androidpublisher.BatchUpdateOneTimeProductsResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Onetimeproducts.BatchUpdate(globals.Package, &batchReq).
			Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch update one-time products: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Monetization Base Plans Commands
// ============================================================================

// MonetizationBasePlansCmd contains base plan management commands.
type MonetizationBasePlansCmd struct {
	Activate          MonetizationBasePlansActivateCmd          `cmd:"" help:"Activate a base plan"`
	Deactivate        MonetizationBasePlansDeactivateCmd        `cmd:"" help:"Deactivate a base plan"`
	Delete            MonetizationBasePlansDeleteCmd            `cmd:"" help:"Delete a base plan"`
	MigratePrices     MonetizationBasePlansMigratePricesCmd     `cmd:"" help:"Migrate base plan prices"`
	BatchMigrate      MonetizationBasePlansBatchMigrateCmd      `cmd:"" help:"Batch migrate base plan prices"`
	BatchUpdateStates MonetizationBasePlansBatchUpdateStatesCmd `cmd:"" help:"Batch update base plan states"`
}

// MonetizationBasePlansActivateCmd activates a base plan.
type MonetizationBasePlansActivateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
}

// Run executes the activate base plan command.
func (cmd *MonetizationBasePlansActivateCmd) Run(globals *Globals) error {
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

	var subscription *androidpublisher.Subscription
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		subscription, callErr = svc.Monetization.Subscriptions.BasePlans.Activate(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID,
			&androidpublisher.ActivateBasePlanRequest{},
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to activate base plan: %v", err))
	}

	return outputResult(
		output.NewResult(subscription).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationBasePlansDeactivateCmd deactivates a base plan.
type MonetizationBasePlansDeactivateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
}

// Run executes the deactivate base plan command.
func (cmd *MonetizationBasePlansDeactivateCmd) Run(globals *Globals) error {
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

	var subscription *androidpublisher.Subscription
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		subscription, callErr = svc.Monetization.Subscriptions.BasePlans.Deactivate(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID,
			&androidpublisher.DeactivateBasePlanRequest{},
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to deactivate base plan: %v", err))
	}

	return outputResult(
		output.NewResult(subscription).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationBasePlansDeleteCmd deletes a base plan.
type MonetizationBasePlansDeleteCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	Confirm        bool   `help:"Confirm destructive operation" required:""`
}

// Run executes the delete base plan command.
func (cmd *MonetizationBasePlansDeleteCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "deletion requires confirmation").
			WithHint("Pass --confirm to confirm the destructive operation")
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Monetization.Subscriptions.BasePlans.Delete(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID,
		).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete base plan: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId": cmd.SubscriptionID,
		"basePlanId":     cmd.BasePlanID,
		"deleted":        true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationBasePlansMigratePricesCmd migrates base plan prices.
type MonetizationBasePlansMigratePricesCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	RegionCode     string `help:"Region code" required:""`
	PriceMicros    int64  `help:"Price in micros" required:""`
}

// Run executes the migrate prices command.
func (cmd *MonetizationBasePlansMigratePricesCmd) Run(globals *Globals) error {
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

	req := &androidpublisher.MigrateBasePlanPricesRequest{
		RegionalPriceMigrations: []*androidpublisher.RegionalPriceMigrationConfig{
			{
				RegionCode:                    cmd.RegionCode,
				OldestAllowedPriceVersionTime: time.Now().Format(time.RFC3339),
			},
		},
	}

	var resp *androidpublisher.MigrateBasePlanPricesResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BasePlans.MigratePrices(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, req,
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to migrate base plan prices: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationBasePlansBatchMigrateCmd batch migrates base plan prices.
type MonetizationBasePlansBatchMigrateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	File           string `help:"Batch migrate JSON file" required:"" type:"existingfile"`
}

// Run executes the batch migrate prices command.
func (cmd *MonetizationBasePlansBatchMigrateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var batchReq androidpublisher.BatchMigrateBasePlanPricesRequest
	if err := json.Unmarshal(fileData, &batchReq); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse batch migrate JSON: %v", err)).
			WithHint("Ensure the file contains valid BatchMigrateBasePlanPricesRequest JSON")
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

	var resp *androidpublisher.BatchMigrateBasePlanPricesResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BasePlans.BatchMigratePrices(
			globals.Package, cmd.SubscriptionID, &batchReq,
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch migrate base plan prices: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationBasePlansBatchUpdateStatesCmd batch updates base plan states.
type MonetizationBasePlansBatchUpdateStatesCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	File           string `help:"Batch update JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update states command.
func (cmd *MonetizationBasePlansBatchUpdateStatesCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var batchReq androidpublisher.BatchUpdateBasePlanStatesRequest
	if err := json.Unmarshal(fileData, &batchReq); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse batch update JSON: %v", err)).
			WithHint("Ensure the file contains valid BatchUpdateBasePlanStatesRequest JSON")
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

	var resp *androidpublisher.BatchUpdateBasePlanStatesResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BasePlans.BatchUpdateStates(
			globals.Package, cmd.SubscriptionID, &batchReq,
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch update base plan states: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// ============================================================================
// Monetization Offers Commands
// ============================================================================

// MonetizationOffersCmd contains offer management commands.
type MonetizationOffersCmd struct {
	Create            MonetizationOffersCreateCmd            `cmd:"" help:"Create an offer"`
	Get               MonetizationOffersGetCmd               `cmd:"" help:"Get an offer"`
	List              MonetizationOffersListCmd              `cmd:"" help:"List offers"`
	Delete            MonetizationOffersDeleteCmd            `cmd:"" help:"Delete an offer"`
	Activate          MonetizationOffersActivateCmd          `cmd:"" help:"Activate an offer"`
	Deactivate        MonetizationOffersDeactivateCmd        `cmd:"" help:"Deactivate an offer"`
	BatchGet          MonetizationOffersBatchGetCmd          `cmd:"" help:"Batch get offers"`
	BatchUpdate       MonetizationOffersBatchUpdateCmd       `cmd:"" help:"Batch update offers"`
	BatchUpdateStates MonetizationOffersBatchUpdateStatesCmd `cmd:"" help:"Batch update offer states"`
}

// MonetizationOffersCreateCmd creates an offer.
type MonetizationOffersCreateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `help:"Offer ID" required:""`
	File           string `help:"Offer JSON file" required:"" type:"existingfile"`
}

// Run executes the create offer command.
func (cmd *MonetizationOffersCreateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var offer androidpublisher.SubscriptionOffer
	if err := json.Unmarshal(fileData, &offer); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse offer JSON: %v", err)).
			WithHint("Ensure the file contains valid SubscriptionOffer JSON")
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

	var created *androidpublisher.SubscriptionOffer
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		created, callErr = svc.Monetization.Subscriptions.BasePlans.Offers.Create(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, &offer,
		).OfferId(cmd.OfferID).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create offer: %v", err))
	}

	return outputResult(
		output.NewResult(created).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOffersGetCmd gets an offer.
type MonetizationOffersGetCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `arg:"" help:"Offer ID" required:""`
}

// Run executes the get offer command.
func (cmd *MonetizationOffersGetCmd) Run(globals *Globals) error {
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

	var offer *androidpublisher.SubscriptionOffer
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		offer, callErr = svc.Monetization.Subscriptions.BasePlans.Offers.Get(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, cmd.OfferID,
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get offer: %v", err))
	}

	return outputResult(
		output.NewResult(offer).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOffersListCmd lists offers.
type MonetizationOffersListCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	PageSize       int64  `help:"Results per page" default:"100"`
	PageToken      string `help:"Pagination token"`
	All            bool   `help:"Fetch all pages"`
}

// offersPageResponse wraps the offers list response for pagination.
type offersPageResponse struct {
	resp *androidpublisher.ListSubscriptionOffersResponse
}

func (r offersPageResponse) GetNextPageToken() string {
	return r.resp.NextPageToken
}

func (r offersPageResponse) GetItems() []*androidpublisher.SubscriptionOffer {
	return r.resp.SubscriptionOffers
}

// Run executes the list offers command.
func (cmd *MonetizationOffersListCmd) Run(globals *Globals) error {
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

	var allOffers []*androidpublisher.SubscriptionOffer
	var nextPageToken string

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Monetization.Subscriptions.BasePlans.Offers.List(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID,
		).Context(ctx)
		if cmd.PageSize > 0 {
			call = call.PageSize(cmd.PageSize)
		}
		if cmd.PageToken != "" {
			call = call.PageToken(cmd.PageToken)
		}

		resp, callErr := call.Do()
		if callErr != nil {
			return callErr
		}

		allOffers = append(allOffers, resp.SubscriptionOffers...)
		nextPageToken = resp.NextPageToken

		if cmd.All && nextPageToken != "" {
			query := func(pageToken string) (offersPageResponse, error) {
				pageCall := svc.Monetization.Subscriptions.BasePlans.Offers.List(
					globals.Package, cmd.SubscriptionID, cmd.BasePlanID,
				).PageToken(pageToken).Context(ctx)
				if cmd.PageSize > 0 {
					pageCall = pageCall.PageSize(cmd.PageSize)
				}
				pageResp, pageErr := pageCall.Do()
				return offersPageResponse{resp: pageResp}, pageErr
			}

			additionalItems, remainingToken, fetchErr := fetchAllPages(ctx, query, nextPageToken, 0)
			if fetchErr != nil {
				return fetchErr
			}
			allOffers = append(allOffers, additionalItems...)
			nextPageToken = remainingToken
		}

		return nil
	})

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list offers: %v", err))
	}

	data := map[string]interface{}{
		"offers":     allOffers,
		"totalCount": len(allOffers),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	if nextPageToken != "" {
		result = result.WithPagination(cmd.PageToken, nextPageToken)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// MonetizationOffersDeleteCmd deletes an offer.
type MonetizationOffersDeleteCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `arg:"" help:"Offer ID" required:""`
	Confirm        bool   `help:"Confirm destructive operation" required:""`
}

// Run executes the delete offer command.
func (cmd *MonetizationOffersDeleteCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "deletion requires confirmation").
			WithHint("Pass --confirm to confirm the destructive operation")
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

	err = client.DoWithRetry(ctx, func() error {
		return svc.Monetization.Subscriptions.BasePlans.Offers.Delete(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, cmd.OfferID,
		).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete offer: %v", err))
	}

	data := map[string]interface{}{
		"subscriptionId": cmd.SubscriptionID,
		"basePlanId":     cmd.BasePlanID,
		"offerId":        cmd.OfferID,
		"deleted":        true,
	}

	return outputResult(
		output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOffersActivateCmd activates an offer.
type MonetizationOffersActivateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `arg:"" help:"Offer ID" required:""`
}

// Run executes the activate offer command.
func (cmd *MonetizationOffersActivateCmd) Run(globals *Globals) error {
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

	var offer *androidpublisher.SubscriptionOffer
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		offer, callErr = svc.Monetization.Subscriptions.BasePlans.Offers.Activate(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, cmd.OfferID,
			&androidpublisher.ActivateSubscriptionOfferRequest{},
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to activate offer: %v", err))
	}

	return outputResult(
		output.NewResult(offer).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOffersDeactivateCmd deactivates an offer.
type MonetizationOffersDeactivateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `arg:"" help:"Offer ID" required:""`
}

// Run executes the deactivate offer command.
func (cmd *MonetizationOffersDeactivateCmd) Run(globals *Globals) error {
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

	var offer *androidpublisher.SubscriptionOffer
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		offer, callErr = svc.Monetization.Subscriptions.BasePlans.Offers.Deactivate(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, cmd.OfferID,
			&androidpublisher.DeactivateSubscriptionOfferRequest{},
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to deactivate offer: %v", err))
	}

	return outputResult(
		output.NewResult(offer).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOffersBatchGetCmd batch gets offers.
type MonetizationOffersBatchGetCmd struct {
	SubscriptionID string   `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string   `arg:"" help:"Base plan ID" required:""`
	OfferIDs       []string `help:"Offer IDs" required:""`
}

// Run executes the batch get offers command.
func (cmd *MonetizationOffersBatchGetCmd) Run(globals *Globals) error {
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

	// Build batch get request with individual offer requests
	requests := make([]*androidpublisher.GetSubscriptionOfferRequest, 0, len(cmd.OfferIDs))
	for _, offerID := range cmd.OfferIDs {
		requests = append(requests, &androidpublisher.GetSubscriptionOfferRequest{
			PackageName: globals.Package,
			ProductId:   cmd.SubscriptionID,
			BasePlanId:  cmd.BasePlanID,
			OfferId:     offerID,
		})
	}

	batchReq := &androidpublisher.BatchGetSubscriptionOffersRequest{
		Requests: requests,
	}

	var resp *androidpublisher.BatchGetSubscriptionOffersResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BasePlans.Offers.BatchGet(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, batchReq,
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch get offers: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOffersBatchUpdateCmd batch updates offers.
type MonetizationOffersBatchUpdateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	File           string `help:"Batch update JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update offers command.
func (cmd *MonetizationOffersBatchUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var batchReq androidpublisher.BatchUpdateSubscriptionOffersRequest
	if err := json.Unmarshal(fileData, &batchReq); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse batch update JSON: %v", err)).
			WithHint("Ensure the file contains valid BatchUpdateSubscriptionOffersRequest JSON")
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

	var resp *androidpublisher.BatchUpdateSubscriptionOffersResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, &batchReq,
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch update offers: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationOffersBatchUpdateStatesCmd batch updates offer states.
type MonetizationOffersBatchUpdateStatesCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	File           string `help:"Batch update states JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update states command.
func (cmd *MonetizationOffersBatchUpdateStatesCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()

	fileData, err := os.ReadFile(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %v", err))
	}

	var batchReq androidpublisher.BatchUpdateSubscriptionOfferStatesRequest
	if err := json.Unmarshal(fileData, &batchReq); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse batch update states JSON: %v", err)).
			WithHint("Ensure the file contains valid BatchUpdateSubscriptionOfferStatesRequest JSON")
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

	var resp *androidpublisher.BatchUpdateSubscriptionOfferStatesResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Monetization.Subscriptions.BasePlans.Offers.BatchUpdateStates(
			globals.Package, cmd.SubscriptionID, cmd.BasePlanID, &batchReq,
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to batch update offer states: %v", err))
	}

	return outputResult(
		output.NewResult(resp).WithDuration(time.Since(start)).WithServices("androidpublisher"),
		globals.Output, globals.Pretty,
	)
}

// MonetizationCapabilitiesCmd lists monetization capabilities.
type MonetizationCapabilitiesCmd struct{}

// getProductCapabilities returns product management capabilities.
func getProductCapabilities() []map[string]interface{} {
	return []map[string]interface{}{
		{"category": "products", "name": "inappproducts.list", "description": "List in-app products"},
		{"category": "products", "name": "inappproducts.get", "description": "Get an in-app product"},
		{"category": "products", "name": "inappproducts.insert", "description": "Create an in-app product"},
		{"category": "products", "name": "inappproducts.update", "description": "Update an in-app product"},
		{"category": "products", "name": "inappproducts.delete", "description": "Delete an in-app product"},
	}
}

// getSubscriptionCapabilities returns subscription management capabilities.
func getSubscriptionCapabilities() []map[string]interface{} {
	return []map[string]interface{}{
		{"category": "subscriptions", "name": "monetization.subscriptions.list", "description": "List subscription products"},
		{"category": "subscriptions", "name": "monetization.subscriptions.get", "description": "Get a subscription product"},
		{"category": "subscriptions", "name": "monetization.subscriptions.create", "description": "Create a subscription product"},
		{"category": "subscriptions", "name": "monetization.subscriptions.patch", "description": "Update a subscription product"},
		{"category": "subscriptions", "name": "monetization.subscriptions.delete", "description": "Delete a subscription product"},
		{"category": "subscriptions", "name": "monetization.subscriptions.archive", "description": "Archive a subscription product"},
		{"category": "subscriptions", "name": "monetization.subscriptions.batchGet", "description": "Batch get subscription products"},
		{"category": "subscriptions", "name": "monetization.subscriptions.batchUpdate", "description": "Batch update subscription products"},
	}
}

// getBasePlanCapabilities returns base plan management capabilities.
func getBasePlanCapabilities() []map[string]interface{} {
	return []map[string]interface{}{
		{"category": "basePlans", "name": "monetization.subscriptions.basePlans.activate", "description": "Activate a base plan"},
		{"category": "basePlans", "name": "monetization.subscriptions.basePlans.deactivate", "description": "Deactivate a base plan"},
		{"category": "basePlans", "name": "monetization.subscriptions.basePlans.delete", "description": "Delete a base plan"},
		{"category": "basePlans", "name": "monetization.subscriptions.basePlans.migratePrices", "description": "Migrate base plan prices"},
		{"category": "basePlans", "name": "monetization.subscriptions.basePlans.batchMigratePrices", "description": "Batch migrate base plan prices"},
		{"category": "basePlans", "name": "monetization.subscriptions.basePlans.batchUpdateStates", "description": "Batch update base plan states"},
	}
}

// getOfferCapabilities returns subscription offer management capabilities.
func getOfferCapabilities() []map[string]interface{} {
	return []map[string]interface{}{
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.create", "description": "Create a subscription offer"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.get", "description": "Get a subscription offer"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.list", "description": "List subscription offers"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.delete", "description": "Delete a subscription offer"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.activate", "description": "Activate a subscription offer"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.deactivate", "description": "Deactivate a subscription offer"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.batchGet", "description": "Batch get subscription offers"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.batchUpdate", "description": "Batch update subscription offers"},
		{"category": "offers", "name": "monetization.subscriptions.basePlans.offers.batchUpdateStates", "description": "Batch update subscription offer states"},
	}
}

// getOneTimeProductCapabilities returns one-time product management capabilities.
func getOneTimeProductCapabilities() []map[string]interface{} {
	return []map[string]interface{}{
		{"category": "oneTimeProducts", "name": "monetization.onetimeproducts.list", "description": "List one-time products"},
		{"category": "oneTimeProducts", "name": "monetization.onetimeproducts.get", "description": "Get a one-time product"},
		{"category": "oneTimeProducts", "name": "monetization.onetimeproducts.patch", "description": "Create or update a one-time product"},
		{"category": "oneTimeProducts", "name": "monetization.onetimeproducts.delete", "description": "Delete a one-time product"},
		{"category": "oneTimeProducts", "name": "monetization.onetimeproducts.batchGet", "description": "Batch get one-time products"},
		{"category": "oneTimeProducts", "name": "monetization.onetimeproducts.batchUpdate", "description": "Batch update one-time products"},
	}
}

// getMonetizationCapabilities returns the list of monetization API capabilities.
func getMonetizationCapabilities() []map[string]interface{} {
	var capabilities []map[string]interface{}
	capabilities = append(capabilities, getProductCapabilities()...)
	capabilities = append(capabilities, getSubscriptionCapabilities()...)
	capabilities = append(capabilities, getBasePlanCapabilities()...)
	capabilities = append(capabilities, getOfferCapabilities()...)
	capabilities = append(capabilities, getOneTimeProductCapabilities()...)
	return capabilities
}

// Run executes the capabilities command.
func (cmd *MonetizationCapabilitiesCmd) Run(globals *Globals) error {
	start := time.Now()
	data := map[string]interface{}{
		"capabilities": getMonetizationCapabilities(),
	}
	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// ============================================================================
// Helper Functions
// ============================================================================

// parseTimeToMillis parses a time string as either RFC3339 or milliseconds since epoch.
func parseTimeToMillis(s string) (int64, error) {
	// Try parsing as milliseconds first
	if millis, err := strconv.ParseInt(s, 10, 64); err == nil {
		return millis, nil
	}

	// Try parsing as RFC3339
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return 0, fmt.Errorf("cannot parse %q as RFC3339 or milliseconds", s)
	}
	return t.UnixMilli(), nil
}
