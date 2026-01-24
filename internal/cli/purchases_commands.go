// Package cli provides purchases commands for gpd.
package cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addPurchasesCommands() {
	purchasesCmd := &cobra.Command{
		Use:   "purchases",
		Short: "Purchase verification commands",
		Long:  "Verify purchase tokens and subscription states.",
	}

	var (
		productID   string
		token       string
		environment string
		productType string
		developerPayload string
		subscriptionID string
		startTime string
		endTime string
		kind string
		maxResults int64
		pageToken string
		expectedExpiry string
		desiredExpiry string
		revokeType string
	)

	// purchases verify
	verifyCmd := &cobra.Command{
		Use:   "verify",
		Short: "Verify a purchase token",
		Long:  "Verify a purchase or subscription token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesVerify(cmd.Context(), productID, token, environment, productType)
		},
	}
	verifyCmd.Flags().StringVar(&productID, "product-id", "", "Product ID")
	verifyCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	verifyCmd.Flags().StringVar(&environment, "environment", "auto", "Environment: sandbox, production, auto")
	verifyCmd.Flags().StringVar(&productType, "type", "auto", "Product type: product, subscription, auto")
	_ = verifyCmd.MarkFlagRequired("token")

	// purchases capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List purchase verification capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesCapabilities(cmd.Context())
		},
	}

	voidedCmd := &cobra.Command{
		Use:   "voided",
		Short: "Voided purchases",
	}

	voidedListCmd := &cobra.Command{
		Use:   "list",
		Short: "List voided purchases",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesVoidedList(cmd.Context(), startTime, endTime, kind, maxResults, pageToken)
		},
	}
	voidedListCmd.Flags().StringVar(&startTime, "start-time", "", "Start time (RFC3339 or millis)")
	voidedListCmd.Flags().StringVar(&endTime, "end-time", "", "End time (RFC3339 or millis)")
	voidedListCmd.Flags().StringVar(&kind, "type", "", "Type: product or subscription")
	voidedListCmd.Flags().Int64Var(&maxResults, "max-results", 0, "Max results per page")
	voidedListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	voidedCmd.AddCommand(voidedListCmd)

	productsCmd := &cobra.Command{
		Use:   "products",
		Short: "Product purchase actions",
	}

	productsAcknowledgeCmd := &cobra.Command{
		Use:   "acknowledge",
		Short: "Acknowledge a product purchase",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesProductsAcknowledge(cmd.Context(), productID, token, developerPayload)
		},
	}
	productsAcknowledgeCmd.Flags().StringVar(&productID, "product-id", "", "Product ID")
	productsAcknowledgeCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	productsAcknowledgeCmd.Flags().StringVar(&developerPayload, "developer-payload", "", "Developer payload")
	_ = productsAcknowledgeCmd.MarkFlagRequired("product-id")
	_ = productsAcknowledgeCmd.MarkFlagRequired("token")

	productsConsumeCmd := &cobra.Command{
		Use:   "consume",
		Short: "Consume a product purchase",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesProductsConsume(cmd.Context(), productID, token)
		},
	}
	productsConsumeCmd.Flags().StringVar(&productID, "product-id", "", "Product ID")
	productsConsumeCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	_ = productsConsumeCmd.MarkFlagRequired("product-id")
	_ = productsConsumeCmd.MarkFlagRequired("token")

	productsCmd.AddCommand(productsAcknowledgeCmd, productsConsumeCmd)

	subscriptionsCmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "Subscription purchase actions",
	}

	subscriptionsAcknowledgeCmd := &cobra.Command{
		Use:   "acknowledge",
		Short: "Acknowledge a subscription purchase",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesSubscriptionsAcknowledge(cmd.Context(), subscriptionID, token, developerPayload)
		},
	}
	subscriptionsAcknowledgeCmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Subscription ID")
	subscriptionsAcknowledgeCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	subscriptionsAcknowledgeCmd.Flags().StringVar(&developerPayload, "developer-payload", "", "Developer payload")
	_ = subscriptionsAcknowledgeCmd.MarkFlagRequired("subscription-id")
	_ = subscriptionsAcknowledgeCmd.MarkFlagRequired("token")

	subscriptionsCancelCmd := &cobra.Command{
		Use:   "cancel",
		Short: "Cancel a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesSubscriptionsCancel(cmd.Context(), subscriptionID, token)
		},
	}
	subscriptionsCancelCmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Subscription ID")
	subscriptionsCancelCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	_ = subscriptionsCancelCmd.MarkFlagRequired("subscription-id")
	_ = subscriptionsCancelCmd.MarkFlagRequired("token")

	subscriptionsDeferCmd := &cobra.Command{
		Use:   "defer",
		Short: "Defer a subscription renewal",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesSubscriptionsDefer(cmd.Context(), subscriptionID, token, expectedExpiry, desiredExpiry)
		},
	}
	subscriptionsDeferCmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Subscription ID")
	subscriptionsDeferCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	subscriptionsDeferCmd.Flags().StringVar(&expectedExpiry, "expected-expiry-time", "", "Expected expiry time (RFC3339 or millis)")
	subscriptionsDeferCmd.Flags().StringVar(&desiredExpiry, "desired-expiry-time", "", "Desired expiry time (RFC3339 or millis)")
	_ = subscriptionsDeferCmd.MarkFlagRequired("subscription-id")
	_ = subscriptionsDeferCmd.MarkFlagRequired("token")
	_ = subscriptionsDeferCmd.MarkFlagRequired("expected-expiry-time")
	_ = subscriptionsDeferCmd.MarkFlagRequired("desired-expiry-time")

	subscriptionsRefundCmd := &cobra.Command{
		Use:   "refund",
		Short: "Refund a subscription (v1)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesSubscriptionsRefund(cmd.Context(), subscriptionID, token)
		},
	}
	subscriptionsRefundCmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Subscription ID")
	subscriptionsRefundCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	_ = subscriptionsRefundCmd.MarkFlagRequired("subscription-id")
	_ = subscriptionsRefundCmd.MarkFlagRequired("token")

	subscriptionsRevokeCmd := &cobra.Command{
		Use:   "revoke",
		Short: "Revoke a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesSubscriptionsRevoke(cmd.Context(), subscriptionID, token, revokeType)
		},
	}
	subscriptionsRevokeCmd.Flags().StringVar(&subscriptionID, "subscription-id", "", "Subscription ID (v1)")
	subscriptionsRevokeCmd.Flags().StringVar(&token, "token", "", "Purchase token")
	subscriptionsRevokeCmd.Flags().StringVar(&revokeType, "revoke-type", "", "Revoke type for v2: fullRefund, partialRefund, itemBasedRefund")
	_ = subscriptionsRevokeCmd.MarkFlagRequired("token")

	subscriptionsCmd.AddCommand(subscriptionsAcknowledgeCmd, subscriptionsCancelCmd, subscriptionsDeferCmd, subscriptionsRefundCmd, subscriptionsRevokeCmd)

	purchasesCmd.AddCommand(verifyCmd, voidedCmd, productsCmd, subscriptionsCmd, capabilitiesCmd)
	c.rootCmd.AddCommand(purchasesCmd)
}

func (c *CLI) purchasesVerify(ctx context.Context, productID, token, environment, productType string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if token == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "purchase token is required"))
	}

	// Get API client
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var purchaseData interface{}
	var purchaseType string

	// Determine product type
	if productType == "auto" || productType == "product" {
		// Try product purchase first
		if productID != "" {
			productPurchase, err := publisher.Purchases.Products.Get(c.packageName, productID, token).Context(ctx).Do()
			if err == nil {
				purchaseData = map[string]interface{}{
					"kind":                        productPurchase.Kind,
					"purchaseTimeMillis":          productPurchase.PurchaseTimeMillis,
					"purchaseState":               productPurchase.PurchaseState,
					"consumptionState":            productPurchase.ConsumptionState,
					"developerPayload":            productPurchase.DeveloperPayload,
					"orderId":                     productPurchase.OrderId,
					"purchaseType":                productPurchase.PurchaseType,
					"acknowledgementState":        productPurchase.AcknowledgementState,
					"quantity":                    productPurchase.Quantity,
					"obfuscatedExternalAccountId": productPurchase.ObfuscatedExternalAccountId,
					"obfuscatedExternalProfileId": productPurchase.ObfuscatedExternalProfileId,
				}
				purchaseType = "product"
			}
		}
	}

	// Try subscription if product failed or type is subscription
	if purchaseData == nil && (productType == "auto" || productType == "subscription") {
		if productID != "" {
			// Use subscriptions v2 API
			subPurchase, err := publisher.Purchases.Subscriptionsv2.Get(c.packageName, token).Context(ctx).Do()
			if err == nil {
				purchaseData = map[string]interface{}{
					"kind":                       subPurchase.Kind,
					"subscriptionState":          subPurchase.SubscriptionState,
					"acknowledgementState":       subPurchase.AcknowledgementState,
					"externalAccountIdentifiers": subPurchase.ExternalAccountIdentifiers,
					"linkedPurchaseToken":        subPurchase.LinkedPurchaseToken,
					"latestOrderId":              subPurchase.LatestOrderId,
				}
				purchaseType = "subscription"
			} else if productType == "subscription" {
				return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
					fmt.Sprintf("subscription not found: %v", err)))
			}
		}
	}

	if purchaseData == nil {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			"purchase not found").WithHint("Check that the token and product ID are correct"))
	}

	result := output.NewResult(map[string]interface{}{
		"valid":       true,
		"type":        purchaseType,
		"environment": environment,
		"productId":   productID,
		"purchase":    purchaseData,
		"package":     c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesCapabilities(ctx context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"supportedProductTypes": []string{"product", "subscription"},
		"supportedEnvironments": []string{"sandbox", "production", "auto"},
		"apis": map[string]interface{}{
			"products": map[string]interface{}{
				"endpoints": []string{"purchases.products.get", "purchases.products.acknowledge", "purchases.products.consume"},
				"purpose":  "One-time product verification",
			},
			"subscriptionsV2": map[string]interface{}{
				"endpoints":   []string{"purchases.subscriptionsv2.get", "purchases.subscriptionsv2.cancel", "purchases.subscriptionsv2.revoke"},
				"purpose":    "Subscription state verification (v2 API)",
				"deprecated": false,
			},
			"subscriptions": map[string]interface{}{
				"endpoints":   []string{"purchases.subscriptions.acknowledge", "purchases.subscriptions.cancel", "purchases.subscriptions.defer", "purchases.subscriptions.refund", "purchases.subscriptions.revoke"},
				"purpose":    "Legacy subscription actions",
				"deprecated": true,
			},
			"voidedPurchases": map[string]interface{}{
				"endpoints": []string{"purchases.voidedpurchases.list"},
				"purpose":   "Voided purchases list",
			},
		},
		"retryPolicy": map[string]interface{}{
			"maxRetries":     3,
			"backoffType":    "exponential",
			"initialDelayMs": 1000,
			"maxDelayMs":     30000,
		},
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesVoidedList(ctx context.Context, startTime, endTime, kind string, maxResults int64, pageToken string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	req := publisher.Purchases.Voidedpurchases.List(c.packageName)
	if startTime != "" {
		ms, err := parseTimeMillis(startTime)
		if err != nil {
			return c.OutputError(err)
		}
		req = req.StartTime(ms)
	}
	if endTime != "" {
		ms, err := parseTimeMillis(endTime)
		if err != nil {
			return c.OutputError(err)
		}
		req = req.EndTime(ms)
	}
	if kind != "" {
		switch kind {
		case "product":
			req = req.Type(0)
		case "subscription":
			req = req.Type(1)
		default:
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "type must be product or subscription"))
		}
	}
	if maxResults > 0 {
		req = req.MaxResults(maxResults)
	}
	if pageToken != "" {
		req = req.Token(pageToken)
	}
	resp, err := req.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	nextToken := ""
	if resp.TokenPagination != nil {
		nextToken = resp.TokenPagination.NextPageToken
	}
	result := output.NewResult(map[string]interface{}{
		"voidedPurchases": resp.VoidedPurchases,
		"nextPageToken":   nextToken,
		"package":         c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesProductsAcknowledge(ctx context.Context, productID, token, developerPayload string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	req := &androidpublisher.ProductPurchasesAcknowledgeRequest{
		DeveloperPayload: developerPayload,
	}
	if err := publisher.Purchases.Products.Acknowledge(c.packageName, productID, token, req).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"productId": productID,
		"token":     token,
		"package":   c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesProductsConsume(ctx context.Context, productID, token string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if err := publisher.Purchases.Products.Consume(c.packageName, productID, token).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"productId": productID,
		"token":     token,
		"package":   c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesSubscriptionsAcknowledge(ctx context.Context, subscriptionID, token, developerPayload string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	req := &androidpublisher.SubscriptionPurchasesAcknowledgeRequest{
		DeveloperPayload: developerPayload,
	}
	if err := publisher.Purchases.Subscriptions.Acknowledge(c.packageName, subscriptionID, token, req).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"subscriptionId": subscriptionID,
		"token":          token,
		"package":        c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesSubscriptionsCancel(ctx context.Context, subscriptionID, token string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if err := publisher.Purchases.Subscriptions.Cancel(c.packageName, subscriptionID, token).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"subscriptionId": subscriptionID,
		"token":          token,
		"package":        c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesSubscriptionsDefer(ctx context.Context, subscriptionID, token, expected, desired string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	expectedMillis, apiErr := parseTimeMillis(expected)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}
	desiredMillis, apiErr := parseTimeMillis(desired)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	req := &androidpublisher.SubscriptionPurchasesDeferRequest{
		DeferralInfo: &androidpublisher.SubscriptionDeferralInfo{
			ExpectedExpiryTimeMillis: expectedMillis,
			DesiredExpiryTimeMillis:  desiredMillis,
		},
	}
	resp, err := publisher.Purchases.Subscriptions.Defer(c.packageName, subscriptionID, token, req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) purchasesSubscriptionsRefund(ctx context.Context, subscriptionID, token string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if err := publisher.Purchases.Subscriptions.Refund(c.packageName, subscriptionID, token).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"subscriptionId": subscriptionID,
		"token":          token,
		"package":        c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) purchasesSubscriptionsRevoke(ctx context.Context, subscriptionID, token, revokeType string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if revokeType != "" {
		var revocationContext *androidpublisher.RevocationContext
		switch revokeType {
		case "fullRefund":
			revocationContext = &androidpublisher.RevocationContext{
				FullRefund: &androidpublisher.RevocationContextFullRefund{},
			}
		case "partialRefund", "proratedRefund":
			revocationContext = &androidpublisher.RevocationContext{
				ProratedRefund: &androidpublisher.RevocationContextProratedRefund{},
			}
		default:
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid revoke-type: must be fullRefund or partialRefund"))
		}
		req := &androidpublisher.RevokeSubscriptionPurchaseRequest{
			RevocationContext: revocationContext,
		}
		if _, err := publisher.Purchases.Subscriptionsv2.Revoke(c.packageName, token, req).Context(ctx).Do(); err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		result := output.NewResult(map[string]interface{}{
			"success":   true,
			"token":     token,
			"package":   c.packageName,
			"api":       "subscriptionsv2",
			"revokeType": revokeType,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}
	if subscriptionID == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "subscription-id is required for v1 revoke"))
	}
	if err := publisher.Purchases.Subscriptions.Revoke(c.packageName, subscriptionID, token).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"subscriptionId": subscriptionID,
		"token":          token,
		"package":        c.packageName,
		"api":            "subscriptions",
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func parseTimeMillis(value string) (int64, *errors.APIError) {
	if value == "" {
		return 0, errors.NewAPIError(errors.CodeValidationError, "time value is required")
	}
	if millis, err := strconv.ParseInt(value, 10, 64); err == nil {
		return millis, nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return 0, errors.NewAPIError(errors.CodeValidationError, "time must be RFC3339 or milliseconds")
	}
	return parsed.UnixMilli(), nil
}
