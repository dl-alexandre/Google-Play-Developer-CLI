// Package cli provides monetization commands for gpd.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	_ = productsCreateCmd.MarkFlagRequired("product-id")

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
		Short: "Manage subscriptions",
		Long:  "List, create, update, and archive subscription products.",
	}
	var (
		subscriptionID   string
		subscriptionFile string
		updateMask       string
		allowMissing     bool
		ids              []string
		batchFile        string
		confirm          bool
		showArchived     bool
	)

	// subscriptions list
	subscriptionsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List subscription products",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsList(cmd.Context(), pageSize, pageToken, all, showArchived)
		},
	}
	subscriptionsListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	subscriptionsListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	subscriptionsListCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")
	subscriptionsListCmd.Flags().BoolVar(&showArchived, "show-archived", false, "Include archived subscriptions")

	// subscriptions get
	subscriptionsGetCmd := &cobra.Command{
		Use:   "get [subscription-id]",
		Short: "Get a subscription product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsGet(cmd.Context(), args[0])
		},
	}

	subscriptionsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a subscription",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsCreate(cmd.Context(), subscriptionID, subscriptionFile)
		},
	}
	subscriptionsCreateCmd.Flags().StringVar(&subscriptionID, "product-id", "", "Subscription product ID")
	subscriptionsCreateCmd.Flags().StringVar(&subscriptionFile, "file", "", "Subscription JSON file")
	_ = subscriptionsCreateCmd.MarkFlagRequired("product-id")
	_ = subscriptionsCreateCmd.MarkFlagRequired("file")

	subscriptionsUpdateCmd := &cobra.Command{
		Use:   "update [subscription-id]",
		Short: "Update a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsUpdate(cmd.Context(), args[0], subscriptionFile)
		},
	}
	subscriptionsUpdateCmd.Flags().StringVar(&subscriptionFile, "file", "", "Subscription JSON file")
	_ = subscriptionsUpdateCmd.MarkFlagRequired("file")

	subscriptionsPatchCmd := &cobra.Command{
		Use:   "patch [subscription-id]",
		Short: "Patch a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsPatch(cmd.Context(), args[0], subscriptionFile, updateMask, allowMissing)
		},
	}
	subscriptionsPatchCmd.Flags().StringVar(&subscriptionFile, "file", "", "Subscription JSON file")
	subscriptionsPatchCmd.Flags().StringVar(&updateMask, "update-mask", "", "Fields to update (comma-separated)")
	subscriptionsPatchCmd.Flags().BoolVar(&allowMissing, "allow-missing", false, "Create if missing")
	_ = subscriptionsPatchCmd.MarkFlagRequired("file")

	subscriptionsDeleteCmd := &cobra.Command{
		Use:   "delete [subscription-id]",
		Short: "Delete a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations"))
			}
			return c.monetizationSubscriptionsDelete(cmd.Context(), args[0])
		},
	}
	subscriptionsDeleteCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")

	subscriptionsArchiveCmd := &cobra.Command{
		Use:   "archive [subscription-id]",
		Short: "Archive a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsArchive(cmd.Context(), args[0])
		},
	}

	subscriptionsBatchGetCmd := &cobra.Command{
		Use:   "batchGet",
		Short: "Batch get subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsBatchGet(cmd.Context(), ids)
		},
	}
	subscriptionsBatchGetCmd.Flags().StringSliceVar(&ids, "ids", nil, "Subscription IDs")
	_ = subscriptionsBatchGetCmd.MarkFlagRequired("ids")

	subscriptionsBatchUpdateCmd := &cobra.Command{
		Use:   "batchUpdate",
		Short: "Batch update subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsBatchUpdate(cmd.Context(), batchFile)
		},
	}
	subscriptionsBatchUpdateCmd.Flags().StringVar(&batchFile, "file", "", "Batch update JSON file")
	_ = subscriptionsBatchUpdateCmd.MarkFlagRequired("file")

	subscriptionsCmd.AddCommand(subscriptionsListCmd, subscriptionsGetCmd, subscriptionsCreateCmd, subscriptionsUpdateCmd,
		subscriptionsPatchCmd, subscriptionsDeleteCmd, subscriptionsArchiveCmd, subscriptionsBatchGetCmd, subscriptionsBatchUpdateCmd)

	// monetization base plans
	basePlansCmd := &cobra.Command{
		Use:   "baseplans",
		Short: "Manage subscription base plans",
	}
	var basePlanFile string
	var basePlanRegion string
	var basePlanPriceMicros int64

	basePlansActivateCmd := &cobra.Command{
		Use:   "activate [subscription-id] [plan-id]",
		Short: "Activate a base plan",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationBasePlansActivate(cmd.Context(), args[0], args[1])
		},
	}
	basePlansDeactivateCmd := &cobra.Command{
		Use:   "deactivate [subscription-id] [plan-id]",
		Short: "Deactivate a base plan",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationBasePlansDeactivate(cmd.Context(), args[0], args[1])
		},
	}
	basePlansDeleteCmd := &cobra.Command{
		Use:   "delete [subscription-id] [plan-id]",
		Short: "Delete a base plan",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations"))
			}
			return c.monetizationBasePlansDelete(cmd.Context(), args[0], args[1])
		},
	}
	basePlansDeleteCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")

	basePlansMigratePricesCmd := &cobra.Command{
		Use:   "migrate-prices [subscription-id] [plan-id]",
		Short: "Migrate base plan prices",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationBasePlansMigratePrices(cmd.Context(), args[0], args[1], basePlanRegion, basePlanPriceMicros)
		},
	}
	basePlansMigratePricesCmd.Flags().StringVar(&basePlanRegion, "region-code", "", "Region code")
	basePlansMigratePricesCmd.Flags().Int64Var(&basePlanPriceMicros, "price-micros", 0, "Price in micros")
	_ = basePlansMigratePricesCmd.MarkFlagRequired("region-code")
	_ = basePlansMigratePricesCmd.MarkFlagRequired("price-micros")

	basePlansBatchMigrateCmd := &cobra.Command{
		Use:   "batch-migrate-prices [subscription-id]",
		Short: "Batch migrate base plan prices",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationBasePlansBatchMigratePrices(cmd.Context(), args[0], basePlanFile)
		},
	}
	basePlansBatchMigrateCmd.Flags().StringVar(&basePlanFile, "file", "", "Batch migrate JSON file")
	_ = basePlansBatchMigrateCmd.MarkFlagRequired("file")

	basePlansBatchUpdateStatesCmd := &cobra.Command{
		Use:   "batch-update-states [subscription-id]",
		Short: "Batch update base plan states",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationBasePlansBatchUpdateStates(cmd.Context(), args[0], basePlanFile)
		},
	}
	basePlansBatchUpdateStatesCmd.Flags().StringVar(&basePlanFile, "file", "", "Batch update JSON file")
	_ = basePlansBatchUpdateStatesCmd.MarkFlagRequired("file")

	basePlansCmd.AddCommand(basePlansActivateCmd, basePlansDeactivateCmd, basePlansDeleteCmd, basePlansMigratePricesCmd,
		basePlansBatchMigrateCmd, basePlansBatchUpdateStatesCmd)

	// monetization offers
	offersCmd := &cobra.Command{
		Use:   "offers",
		Short: "Manage subscription offers",
	}
	var offerID string
	var offerFile string
	var offerIDs []string

	offersCreateCmd := &cobra.Command{
		Use:   "create [subscription-id] [plan-id]",
		Short: "Create an offer",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersCreate(cmd.Context(), args[0], args[1], offerID, offerFile)
		},
	}
	offersCreateCmd.Flags().StringVar(&offerID, "offer-id", "", "Offer ID")
	offersCreateCmd.Flags().StringVar(&offerFile, "file", "", "Offer JSON file")
	_ = offersCreateCmd.MarkFlagRequired("offer-id")
	_ = offersCreateCmd.MarkFlagRequired("file")

	offersGetCmd := &cobra.Command{
		Use:   "get [subscription-id] [plan-id] [offer-id]",
		Short: "Get an offer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersGet(cmd.Context(), args[0], args[1], args[2])
		},
	}

	offersListCmd := &cobra.Command{
		Use:   "list [subscription-id] [plan-id]",
		Short: "List offers",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersList(cmd.Context(), args[0], args[1], pageSize, pageToken, all)
		},
	}
	offersListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	offersListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	offersListCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	offersDeleteCmd := &cobra.Command{
		Use:   "delete [subscription-id] [plan-id] [offer-id]",
		Short: "Delete an offer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations"))
			}
			return c.monetizationOffersDelete(cmd.Context(), args[0], args[1], args[2])
		},
	}
	offersDeleteCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")

	offersActivateCmd := &cobra.Command{
		Use:   "activate [subscription-id] [plan-id] [offer-id]",
		Short: "Activate an offer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersActivate(cmd.Context(), args[0], args[1], args[2])
		},
	}

	offersDeactivateCmd := &cobra.Command{
		Use:   "deactivate [subscription-id] [plan-id] [offer-id]",
		Short: "Deactivate an offer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersDeactivate(cmd.Context(), args[0], args[1], args[2])
		},
	}

	offersBatchGetCmd := &cobra.Command{
		Use:   "batchGet [subscription-id] [plan-id]",
		Short: "Batch get offers",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersBatchGet(cmd.Context(), args[0], args[1], offerIDs)
		},
	}
	offersBatchGetCmd.Flags().StringSliceVar(&offerIDs, "offer-ids", nil, "Offer IDs")
	_ = offersBatchGetCmd.MarkFlagRequired("offer-ids")

	offersBatchUpdateCmd := &cobra.Command{
		Use:   "batchUpdate [subscription-id] [plan-id]",
		Short: "Batch update offers",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersBatchUpdate(cmd.Context(), args[0], args[1], offerFile)
		},
	}
	offersBatchUpdateCmd.Flags().StringVar(&offerFile, "file", "", "Batch update JSON file")
	_ = offersBatchUpdateCmd.MarkFlagRequired("file")

	offersBatchUpdateStatesCmd := &cobra.Command{
		Use:   "batchUpdateStates [subscription-id] [plan-id]",
		Short: "Batch update offer states",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationOffersBatchUpdateStates(cmd.Context(), args[0], args[1], offerFile)
		},
	}
	offersBatchUpdateStatesCmd.Flags().StringVar(&offerFile, "file", "", "Batch update states JSON file")
	_ = offersBatchUpdateStatesCmd.MarkFlagRequired("file")

	offersCmd.AddCommand(offersCreateCmd, offersGetCmd, offersListCmd, offersDeleteCmd, offersActivateCmd, offersDeactivateCmd,
		offersBatchGetCmd, offersBatchUpdateCmd, offersBatchUpdateStatesCmd)

	// monetization onetime products (alias to legacy products)
	onetimeProductsCmd := &cobra.Command{
		Use:   "onetimeproducts",
		Short: "Manage one-time products",
		Long:  "Alias of legacy in-app products for managed/consumable items.",
	}
	onetimeProductsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List one-time products",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsList(cmd.Context(), pageSize, pageToken, all)
		},
	}
	onetimeProductsListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	onetimeProductsListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	onetimeProductsListCmd.Flags().BoolVar(&all, "all", false, "Fetch all pages")

	onetimeProductsGetCmd := &cobra.Command{
		Use:   "get [product-id]",
		Short: "Get a one-time product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsGet(cmd.Context(), args[0])
		},
	}

	onetimeProductsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a one-time product",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsCreate(cmd.Context(), productID, productType, defaultPrice, status)
		},
	}
	onetimeProductsCreateCmd.Flags().StringVar(&productID, "product-id", "", "Product SKU")
	onetimeProductsCreateCmd.Flags().StringVar(&productType, "type", "managed", "Product type: managed, consumable")
	onetimeProductsCreateCmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default price in micros (e.g., 990000 for $0.99)")
	onetimeProductsCreateCmd.Flags().StringVar(&status, "status", "active", "Product status: active, inactive")
	_ = onetimeProductsCreateCmd.MarkFlagRequired("product-id")

	onetimeProductsUpdateCmd := &cobra.Command{
		Use:   "update [product-id]",
		Short: "Update a one-time product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsUpdate(cmd.Context(), args[0], defaultPrice, status)
		},
	}
	onetimeProductsUpdateCmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default price in micros")
	onetimeProductsUpdateCmd.Flags().StringVar(&status, "status", "", "Product status: active, inactive")

	onetimeProductsDeleteCmd := &cobra.Command{
		Use:   "delete [product-id]",
		Short: "Delete a one-time product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsDelete(cmd.Context(), args[0])
		},
	}

	onetimeProductsCmd.AddCommand(onetimeProductsListCmd, onetimeProductsGetCmd, onetimeProductsCreateCmd, onetimeProductsUpdateCmd, onetimeProductsDeleteCmd)

	// monetization convert-region-prices
	var priceMicros int64
	var currencyCode string
	var regionFilter []string
	convertCmd := &cobra.Command{
		Use:   "convert-region-prices",
		Short: "Convert region prices",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationConvertRegionPrices(cmd.Context(), priceMicros, currencyCode, regionFilter)
		},
	}
	convertCmd.Flags().Int64Var(&priceMicros, "price-micros", 0, "Price in micros")
	convertCmd.Flags().StringVar(&currencyCode, "currency", "USD", "Currency code")
	convertCmd.Flags().StringSliceVar(&regionFilter, "to-regions", nil, "Region codes to include")
	_ = convertCmd.MarkFlagRequired("price-micros")

	// monetization capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List monetization capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationCapabilities(cmd.Context())
		},
	}

	monetizationCmd.AddCommand(productsCmd, subscriptionsCmd, basePlansCmd, offersCmd, onetimeProductsCmd, convertCmd, capabilitiesCmd)
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

func (c *CLI) monetizationSubscriptionsList(ctx context.Context, pageSize int64, pageToken string, all bool, showArchived bool) error {
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
	req = req.ShowArchived(showArchived)
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

func (c *CLI) monetizationSubscriptionsCreate(ctx context.Context, subscriptionID, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	var sub androidpublisher.Subscription
	if err := loadJSONFile(filePath, &sub); err != nil {
		return c.OutputError(err)
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	created, err := publisher.Monetization.Subscriptions.Create(c.packageName, &sub).
		ProductId(subscriptionID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(created)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsUpdate(ctx context.Context, subscriptionID, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	var sub androidpublisher.Subscription
	if err := loadJSONFile(filePath, &sub); err != nil {
		return c.OutputError(err)
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	updated, err := publisher.Monetization.Subscriptions.Patch(c.packageName, subscriptionID, &sub).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(updated)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsPatch(ctx context.Context, subscriptionID, filePath, updateMask string, allowMissing bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	var sub androidpublisher.Subscription
	if err := loadJSONFile(filePath, &sub); err != nil {
		return c.OutputError(err)
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	call := publisher.Monetization.Subscriptions.Patch(c.packageName, subscriptionID, &sub).
		AllowMissing(allowMissing)
	if updateMask != "" {
		call = call.UpdateMask(updateMask)
	}

	updated, err := call.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(updated)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsDelete(ctx context.Context, subscriptionID string) error {
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

	if err := publisher.Monetization.Subscriptions.Delete(c.packageName, subscriptionID).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"productId":    subscriptionID,
		"package":      c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsArchive(ctx context.Context, subscriptionID string) error {
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

	resp, err := publisher.Monetization.Subscriptions.Archive(c.packageName, subscriptionID, &androidpublisher.ArchiveSubscriptionRequest{}).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(resp)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsBatchGet(ctx context.Context, ids []string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if len(ids) == 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "ids are required"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	resp, err := publisher.Monetization.Subscriptions.BatchGet(c.packageName).ProductIds(ids...).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(resp)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationSubscriptionsBatchUpdate(ctx context.Context, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	var req androidpublisher.BatchUpdateSubscriptionsRequest
	if err := loadJSONFile(filePath, &req); err != nil {
		return c.OutputError(err)
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	resp, err := publisher.Monetization.Subscriptions.BatchUpdate(c.packageName, &req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(resp)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansActivate(ctx context.Context, subscriptionID, basePlanID string) error {
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
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Activate(c.packageName, subscriptionID, basePlanID, &androidpublisher.ActivateBasePlanRequest{}).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansDeactivate(ctx context.Context, subscriptionID, basePlanID string) error {
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
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Deactivate(c.packageName, subscriptionID, basePlanID, &androidpublisher.DeactivateBasePlanRequest{}).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansDelete(ctx context.Context, subscriptionID, basePlanID string) error {
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
	if err := publisher.Monetization.Subscriptions.BasePlans.Delete(c.packageName, subscriptionID, basePlanID).Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"basePlanId": basePlanID,
		"productId":  subscriptionID,
		"package":    c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansMigratePrices(ctx context.Context, subscriptionID, basePlanID, regionCode string, priceMicros int64) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if priceMicros <= 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "price-micros must be greater than zero"))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	// Note: The migration API migrates existing prices, not sets new ones directly.
	// This requires OldestAllowedPriceVersionTime to specify which price cohorts to migrate.
	// Using current time as a reasonable default - all existing price cohorts will be migrated.
	req := &androidpublisher.MigrateBasePlanPricesRequest{
		RegionalPriceMigrations: []*androidpublisher.RegionalPriceMigrationConfig{
			{
				RegionCode:                regionCode,
				OldestAllowedPriceVersionTime: "1970-01-01T00:00:00Z", // Migrate all existing price cohorts
			},
		},
		RegionsVersion: &androidpublisher.RegionsVersion{
			Version: "2022/02", // Latest regions version
		},
	}
	resp, err := publisher.Monetization.Subscriptions.BasePlans.MigratePrices(c.packageName, subscriptionID, basePlanID, req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansBatchMigratePrices(ctx context.Context, subscriptionID, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	var req androidpublisher.BatchMigrateBasePlanPricesRequest
	if err := loadJSONFile(filePath, &req); err != nil {
		return c.OutputError(err)
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	resp, err := publisher.Monetization.Subscriptions.BasePlans.BatchMigratePrices(c.packageName, subscriptionID, &req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansBatchUpdateStates(ctx context.Context, subscriptionID, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	var req androidpublisher.BatchUpdateBasePlanStatesRequest
	if err := loadJSONFile(filePath, &req); err != nil {
		return c.OutputError(err)
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	resp, err := publisher.Monetization.Subscriptions.BasePlans.BatchUpdateStates(c.packageName, subscriptionID, &req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersCreate(ctx context.Context, subscriptionID, basePlanID, offerID, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	var offer androidpublisher.SubscriptionOffer
	if err := loadJSONFile(filePath, &offer); err != nil {
		return c.OutputError(err)
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Offers.Create(c.packageName, subscriptionID, basePlanID, &offer).
		OfferId(offerID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersGet(ctx context.Context, subscriptionID, basePlanID, offerID string) error {
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
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Offers.Get(c.packageName, subscriptionID, basePlanID, offerID).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersList(ctx context.Context, subscriptionID, basePlanID string, pageSize int64, pageToken string, all bool) error {
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
	req := publisher.Monetization.Subscriptions.BasePlans.Offers.List(c.packageName, subscriptionID, basePlanID)
	if pageSize > 0 {
		req = req.PageSize(pageSize)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	var offers []interface{}
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		for _, offer := range resp.SubscriptionOffers {
			offers = append(offers, offer)
		}
		if resp.NextPageToken == "" || !all {
			pageToken = resp.NextPageToken
			break
		}
		req = req.PageToken(resp.NextPageToken)
	}
	result := output.NewResult(offers)
	if pageToken != "" {
		result.WithPagination("", pageToken)
	}
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersDelete(ctx context.Context, subscriptionID, basePlanID, offerID string) error {
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
	if err := publisher.Monetization.Subscriptions.BasePlans.Offers.Delete(c.packageName, subscriptionID, basePlanID, offerID).
		Context(ctx).Do(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success": true,
		"offerId": offerID,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersActivate(ctx context.Context, subscriptionID, basePlanID, offerID string) error {
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
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Offers.Activate(c.packageName, subscriptionID, basePlanID, offerID, &androidpublisher.ActivateSubscriptionOfferRequest{}).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersDeactivate(ctx context.Context, subscriptionID, basePlanID, offerID string) error {
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
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Offers.Deactivate(c.packageName, subscriptionID, basePlanID, offerID, &androidpublisher.DeactivateSubscriptionOfferRequest{}).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersBatchGet(ctx context.Context, subscriptionID, basePlanID string, offerIDs []string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if len(offerIDs) == 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "offer-ids are required"))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	requests := make([]*androidpublisher.GetSubscriptionOfferRequest, 0, len(offerIDs))
	for _, offerID := range offerIDs {
		requests = append(requests, &androidpublisher.GetSubscriptionOfferRequest{
			PackageName: c.packageName,
			ProductId:  subscriptionID,
			BasePlanId:  basePlanID,
			OfferId:     offerID,
		})
	}
	req := &androidpublisher.BatchGetSubscriptionOffersRequest{
		Requests: requests,
	}
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Offers.BatchGet(c.packageName, subscriptionID, basePlanID, req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersBatchUpdate(ctx context.Context, subscriptionID, basePlanID, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	var req androidpublisher.BatchUpdateSubscriptionOffersRequest
	if err := loadJSONFile(filePath, &req); err != nil {
		return c.OutputError(err)
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Offers.BatchUpdate(c.packageName, subscriptionID, basePlanID, &req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationOffersBatchUpdateStates(ctx context.Context, subscriptionID, basePlanID, filePath string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	var req androidpublisher.BatchUpdateSubscriptionOfferStatesRequest
	if err := loadJSONFile(filePath, &req); err != nil {
		return c.OutputError(err)
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	resp, err := publisher.Monetization.Subscriptions.BasePlans.Offers.BatchUpdateStates(c.packageName, subscriptionID, basePlanID, &req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationConvertRegionPrices(ctx context.Context, priceMicros int64, currencyCode string, regionFilter []string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if priceMicros <= 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "price-micros must be greater than zero"))
	}
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	units := priceMicros / 1_000_000
	nanos := (priceMicros % 1_000_000) * 1000
	req := &androidpublisher.ConvertRegionPricesRequest{
		Price: &androidpublisher.Money{
			CurrencyCode: strings.ToUpper(currencyCode),
			Units:        units,
			Nanos:        nanos,
		},
	}
	resp, err := publisher.Monetization.ConvertRegionPrices(c.packageName, req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if len(regionFilter) > 0 && resp.ConvertedRegionPrices != nil {
		filtered := make(map[string]androidpublisher.ConvertedRegionPrice)
		for _, code := range regionFilter {
			if val, ok := resp.ConvertedRegionPrices[strings.ToUpper(code)]; ok {
				filtered[strings.ToUpper(code)] = val
			}
		}
		resp.ConvertedRegionPrices = filtered
	}

	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
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
			"operations": []string{"list", "get", "create", "update", "patch", "delete", "archive", "batchGet", "batchUpdate"},
		},
		"basePlans": map[string]interface{}{
			"operations": []string{"activate", "deactivate", "delete", "migrate-prices", "batch-migrate-prices", "batch-update-states"},
		},
		"offers": map[string]interface{}{
			"operations": []string{"create", "get", "list", "delete", "activate", "deactivate", "batchGet", "batchUpdate", "batchUpdateStates"},
		},
		"regionalPricing": map[string]interface{}{
			"operations": []string{"convert-region-prices"},
		},
		"apiLimitations": []string{
			"Offer updates use batch update endpoint",
		},
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func loadJSONFile(path string, out interface{}) *errors.APIError {
	data, err := os.ReadFile(path)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read file: %s", path))
	}
	if err := json.Unmarshal(data, out); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, "invalid JSON file")
	}
	return nil
}
