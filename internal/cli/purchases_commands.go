// Package cli provides purchases commands for gpd.
package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/google-play-cli/gpd/internal/errors"
	"github.com/google-play-cli/gpd/internal/output"
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
	verifyCmd.MarkFlagRequired("token")

	// purchases capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List purchase verification capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.purchasesCapabilities(cmd.Context())
		},
	}

	purchasesCmd.AddCommand(verifyCmd, capabilitiesCmd)
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
				"endpoint": "purchases.products.get",
				"purpose":  "One-time product verification",
			},
			"subscriptionsV2": map[string]interface{}{
				"endpoint":   "purchases.subscriptionsv2.get",
				"purpose":    "Subscription state verification (v2 API)",
				"deprecated": false,
			},
			"subscriptions": map[string]interface{}{
				"endpoint":   "purchases.subscriptions.get",
				"purpose":    "Legacy subscription verification",
				"deprecated": true,
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
