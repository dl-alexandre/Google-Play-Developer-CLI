// Package cli provides monetization commands for gpd.
package cli

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addMonetizationCommands() {
	monetizationCmd := &cobra.Command{
		Use:   "monetization",
		Short: "Monetization commands",
		Long:  "Manage in-app products and subscriptions.",
	}

	// monetization products
	productsCmd := &cobra.Command{
		Use:   "products",
		Short: "Manage in-app products",
		Long:  "List, create, and update in-app products.",
	}

	var (
		productID    string
		productType  string
		defaultPrice string
		status       string
		pageSize     int64
		pageToken    string
		all          bool
	)

	// products list
	productsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List in-app products",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsList(cmd.Context(), pageSize, pageToken, all)
		},
	}
	productsListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	productsListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	productsListCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	// products get
	productsGetCmd := &cobra.Command{
		Use:   "get [product-id]",
		Short: "Get an in-app product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsGet(cmd.Context(), args[0])
		},
	}

	// products create
	productsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an in-app product",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsCreate(cmd.Context(), productID, productType, defaultPrice, status)
		},
	}
	productsCreateCmd.Flags().StringVar(&productID, "product-id", "", "Product SKU")
	productsCreateCmd.Flags().StringVar(&productType, "type", "managed", "Product type: managed, consumable")
	productsCreateCmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default price in micros (e.g., 990000 for $0.99)")
	productsCreateCmd.Flags().StringVar(&status, "status", "active", "Product status: active, inactive")
	productsCreateCmd.MarkFlagRequired("product-id")

	// products update
	productsUpdateCmd := &cobra.Command{
		Use:   "update [product-id]",
		Short: "Update an in-app product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsUpdate(cmd.Context(), args[0], defaultPrice, status)
		},
	}
	productsUpdateCmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default price in micros")
	productsUpdateCmd.Flags().StringVar(&status, "status", "", "Product status: active, inactive")

	// products delete
	productsDeleteCmd := &cobra.Command{
		Use:   "delete [product-id]",
		Short: "Delete an in-app product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsDelete(cmd.Context(), args[0])
		},
	}

	productsCmd.AddCommand(productsListCmd, productsGetCmd, productsCreateCmd, productsUpdateCmd, productsDeleteCmd)

	// monetization subscriptions (read-only)
	subscriptionsCmd := &cobra.Command{
		Use:   "subscriptions",
		Short: "View subscriptions (read-only)",
		Long:  "List and view subscription products. Creating/modifying subscriptions is not supported.",
	}

	// subscriptions list
	subscriptionsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List subscription products",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsList(cmd.Context(), pageSize, pageToken, all)
		},
	}
	subscriptionsListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	subscriptionsListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	subscriptionsListCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	// subscriptions get
	subscriptionsGetCmd := &cobra.Command{
		Use:   "get [subscription-id]",
		Short: "Get a subscription product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsGet(cmd.Context(), args[0])
		},
	}

	subscriptionsCmd.AddCommand(subscriptionsListCmd, subscriptionsGetCmd)

	// monetization capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List monetization capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationCapabilities(cmd.Context())
		},
	}

	monetizationCmd.AddCommand(productsCmd, subscriptionsCmd, capabilitiesCmd)
	c.rootCmd.AddCommand(monetizationCmd)
}

func (c *CLI) monetizationProductsList(ctx context.Context, pageSize int64, pageToken string, all bool) error {
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

	req := publisher.Inappproducts.List(c.packageName)
	if pageToken != "" {
		req = req.Token(pageToken)
	}

	var allProducts []interface{}
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}

		for _, product := range resp.Inappproduct {
			allProducts = append(allProducts, map[string]interface{}{
				"sku":             product.Sku,
				"status":          product.Status,
				"purchaseType":    product.PurchaseType,
				"defaultPrice":    product.DefaultPrice,
				"defaultLanguage": product.DefaultLanguage,
			})
		}

		if resp.TokenPagination == nil || resp.TokenPagination.NextPageToken == "" || !all {
			if resp.TokenPagination != nil {
				pageToken = resp.TokenPagination.NextPageToken
			}
			break
		}
		req = req.Token(resp.TokenPagination.NextPageToken)
	}

	result := output.NewResult(allProducts)
	if pageToken != "" {
		result.WithPagination("", pageToken)
	}
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationProductsGet(ctx context.Context, productID string) error {
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

	product, err := publisher.Inappproducts.Get(c.packageName, productID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"sku":             product.Sku,
		"status":          product.Status,
		"purchaseType":    product.PurchaseType,
		"defaultPrice":    product.DefaultPrice,
		"defaultLanguage": product.DefaultLanguage,
		"listings":        product.Listings,
		"prices":          product.Prices,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationProductsCreate(ctx context.Context, productID, productType, defaultPrice, status string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if productID == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"product ID is required"))
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

	// Build product
	product := &androidpublisher.InAppProduct{
		PackageName:     c.packageName,
		Sku:             productID,
		Status:          status,
		DefaultLanguage: "en-US",
	}

	// Set purchase type (managed or consumable -> managedUser or subscription)
	if productType == "consumable" {
		product.PurchaseType = "managedUser" // Consumable in Play Store API
	} else {
		product.PurchaseType = "managedUser" // Managed non-consumable
	}

	// Parse and set default price if provided
	if defaultPrice != "" {
		priceMicros, err := strconv.ParseInt(defaultPrice, 10, 64)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				"invalid price format - use micros (e.g., 990000 for $0.99)"))
		}
		product.DefaultPrice = &androidpublisher.Price{
			Currency:    "USD",
			PriceMicros: strconv.FormatInt(priceMicros, 10),
		}
	}

	// Create product
	created, err := publisher.Inappproducts.Insert(c.packageName, product).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"productId":    created.Sku,
		"status":       created.Status,
		"purchaseType": created.PurchaseType,
		"defaultPrice": created.DefaultPrice,
		"package":      c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationProductsUpdate(ctx context.Context, productID, defaultPrice, status string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
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

	// Get existing product
	existing, err := publisher.Inappproducts.Get(c.packageName, productID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			"product not found: "+productID))
	}

	// Update fields if provided
	if status != "" {
		existing.Status = status
	}
	if defaultPrice != "" {
		priceMicros, err := strconv.ParseInt(defaultPrice, 10, 64)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				"invalid price format - use micros (e.g., 990000 for $0.99)"))
		}
		existing.DefaultPrice = &androidpublisher.Price{
			Currency:    "USD",
			PriceMicros: strconv.FormatInt(priceMicros, 10),
		}
	}

	// Update product
	updated, err := publisher.Inappproducts.Update(c.packageName, productID, existing).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"productId":    updated.Sku,
		"status":       updated.Status,
		"defaultPrice": updated.DefaultPrice,
		"package":      c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationProductsDelete(ctx context.Context, productID string) error {
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

	err = publisher.Inappproducts.Delete(c.packageName, productID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"productId": productID,
		"deleted":   true,
		"package":   c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsList(ctx context.Context, pageSize int64, pageToken string, all bool) error {
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

	req := publisher.Monetization.Subscriptions.List(c.packageName)
	if pageSize > 0 {
		req = req.PageSize(int64(pageSize))
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	var allSubscriptions []interface{}
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}

		for _, sub := range resp.Subscriptions {
			allSubscriptions = append(allSubscriptions, map[string]interface{}{
				"productId":   sub.ProductId,
				"packageName": sub.PackageName,
				"archived":    sub.Archived,
			})
		}

		if resp.NextPageToken == "" || !all {
			pageToken = resp.NextPageToken
			break
		}
		req = req.PageToken(resp.NextPageToken)
	}

	result := output.NewResult(allSubscriptions)
	if pageToken != "" {
		result.WithPagination("", pageToken)
	}
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsGet(ctx context.Context, subscriptionID string) error {
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

	sub, err := publisher.Monetization.Subscriptions.Get(c.packageName, subscriptionID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"productId":   sub.ProductId,
		"packageName": sub.PackageName,
		"archived":    sub.Archived,
		"basePlans":   sub.BasePlans,
		"listings":    sub.Listings,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationCapabilities(ctx context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"products": map[string]interface{}{
			"supportedTypes": []string{"managed", "consumable"},
			"operations":     []string{"list", "get", "create", "update", "delete"},
		},
		"subscriptions": map[string]interface{}{
			"note":       "Subscription management is read-only in this version",
			"operations": []string{"list", "get"},
			"notSupported": []string{
				"create",
				"update",
				"delete",
				"base plans management",
				"offers management",
				"regional pricing",
				"introductory pricing",
			},
		},
		"apiLimitations": []string{
			"Subscription creation/modification requires Play Console UI",
			"Modern base plans and offers require separate API integration",
			"Regional pricing requires locale-specific configuration",
		},
	})
	return c.Output(result.WithServices("androidpublisher"))
}
