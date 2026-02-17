package cli

import (
	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addMonetizationCommands() {
	monetizationCmd := &cobra.Command{
		Use:   "monetization",
		Short: "Monetization commands",
		Long:  "Manage in-app products and subscriptions.",
	}

	c.addMonetizationProductsCommands(monetizationCmd)
	c.addMonetizationSubscriptionsCommands(monetizationCmd)
	c.addMonetizationBasePlansCommands(monetizationCmd)
	c.addMonetizationOffersCommands(monetizationCmd)
	c.addMonetizationOnetimeProductsCommands(monetizationCmd)
	c.addMonetizationUtilityCommands(monetizationCmd)

	c.rootCmd.AddCommand(monetizationCmd)
}

func (c *CLI) addMonetizationProductsCommands(monetizationCmd *cobra.Command) {
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

	productsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List in-app products",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsList(cmd.Context(), pageSize, pageToken, all)
		},
	}
	productsListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	productsListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(productsListCmd, &all)

	productsGetCmd := &cobra.Command{
		Use:   "get [product-id]",
		Short: "Get an in-app product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsGet(cmd.Context(), args[0])
		},
	}

	productsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an in-app product",
		RunE: func(cmd *cobra.Command, args []string) error {
			if productID == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--product-id is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationProductsCreate(cmd.Context(), productID, productType, defaultPrice, status)
		},
	}
	productsCreateCmd.Flags().StringVar(&productID, "product-id", "", "Product SKU")
	productsCreateCmd.Flags().StringVar(&productType, "type", "managed", "Product type: managed, consumable")
	productsCreateCmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default price in micros (e.g., 990000 for $0.99)")
	productsCreateCmd.Flags().StringVar(&status, "status", "active", "Product status: active, inactive")

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

	productsDeleteCmd := &cobra.Command{
		Use:   "delete [product-id]",
		Short: "Delete an in-app product",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsDelete(cmd.Context(), args[0])
		},
	}

	productsCmd.AddCommand(productsListCmd, productsGetCmd, productsCreateCmd, productsUpdateCmd, productsDeleteCmd)
	monetizationCmd.AddCommand(productsCmd)
}

func (c *CLI) addMonetizationSubscriptionsCommands(monetizationCmd *cobra.Command) {
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
		pageSize         int64
		pageToken        string
		all              bool
	)

	subscriptionsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List subscription products",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationSubscriptionsList(cmd.Context(), pageSize, pageToken, all, showArchived)
		},
	}
	subscriptionsListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	subscriptionsListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(subscriptionsListCmd, &all)
	subscriptionsListCmd.Flags().BoolVar(&showArchived, "show-archived", false, "Include archived subscriptions")

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
			if subscriptionID == "" || subscriptionFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--product-id and --file are required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationSubscriptionsCreate(cmd.Context(), subscriptionID, subscriptionFile)
		},
	}
	subscriptionsCreateCmd.Flags().StringVar(&subscriptionID, "product-id", "", "Subscription product ID")
	subscriptionsCreateCmd.Flags().StringVar(&subscriptionFile, "file", "", "Subscription JSON file")

	subscriptionsUpdateCmd := &cobra.Command{
		Use:   "update [subscription-id]",
		Short: "Update a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if subscriptionFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--file is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationSubscriptionsUpdate(cmd.Context(), args[0], subscriptionFile)
		},
	}
	subscriptionsUpdateCmd.Flags().StringVar(&subscriptionFile, "file", "", "Subscription JSON file")

	subscriptionsPatchCmd := &cobra.Command{
		Use:   "patch [subscription-id]",
		Short: "Patch a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if subscriptionFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--file is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationSubscriptionsPatch(cmd.Context(), args[0], subscriptionFile, updateMask, allowMissing)
		},
	}
	subscriptionsPatchCmd.Flags().StringVar(&subscriptionFile, "file", "", "Subscription JSON file")
	subscriptionsPatchCmd.Flags().StringVar(&updateMask, "update-mask", "", "Fields to update (comma-separated)")
	subscriptionsPatchCmd.Flags().BoolVar(&allowMissing, "allow-missing", false, "Create if missing")

	subscriptionsDeleteCmd := &cobra.Command{
		Use:   "delete [subscription-id]",
		Short: "Delete a subscription",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations")).WithServices("androidpublisher")
				return c.Output(result)
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
			if len(ids) == 0 {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--ids is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationSubscriptionsBatchGet(cmd.Context(), ids)
		},
	}
	subscriptionsBatchGetCmd.Flags().StringSliceVar(&ids, "ids", nil, "Subscription IDs")

	subscriptionsBatchUpdateCmd := &cobra.Command{
		Use:   "batchUpdate",
		Short: "Batch update subscriptions",
		RunE: func(cmd *cobra.Command, args []string) error {
			if batchFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--file is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationSubscriptionsBatchUpdate(cmd.Context(), batchFile)
		},
	}
	subscriptionsBatchUpdateCmd.Flags().StringVar(&batchFile, "file", "", "Batch update JSON file")

	subscriptionsCmd.AddCommand(subscriptionsListCmd, subscriptionsGetCmd, subscriptionsCreateCmd, subscriptionsUpdateCmd,
		subscriptionsPatchCmd, subscriptionsDeleteCmd, subscriptionsArchiveCmd, subscriptionsBatchGetCmd, subscriptionsBatchUpdateCmd)
	monetizationCmd.AddCommand(subscriptionsCmd)
}

func (c *CLI) addMonetizationBasePlansCommands(monetizationCmd *cobra.Command) {
	basePlansCmd := &cobra.Command{
		Use:   "baseplans",
		Short: "Manage subscription base plans",
	}
	var (
		basePlanFile        string
		basePlanRegion      string
		basePlanPriceMicros int64
		confirm             bool
	)

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
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations")).WithServices("androidpublisher")
				return c.Output(result)
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
			if basePlanRegion == "" || basePlanPriceMicros <= 0 {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--region-code and --price-micros are required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationBasePlansMigratePrices(cmd.Context(), args[0], args[1], basePlanRegion, basePlanPriceMicros)
		},
	}
	basePlansMigratePricesCmd.Flags().StringVar(&basePlanRegion, "region-code", "", "Region code")
	basePlansMigratePricesCmd.Flags().Int64Var(&basePlanPriceMicros, "price-micros", 0, "Price in micros")

	basePlansBatchMigrateCmd := &cobra.Command{
		Use:   "batch-migrate-prices [subscription-id]",
		Short: "Batch migrate base plan prices",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if basePlanFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--file is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationBasePlansBatchMigratePrices(cmd.Context(), args[0], basePlanFile)
		},
	}
	basePlansBatchMigrateCmd.Flags().StringVar(&basePlanFile, "file", "", "Batch migrate JSON file")

	basePlansBatchUpdateStatesCmd := &cobra.Command{
		Use:   "batch-update-states [subscription-id]",
		Short: "Batch update base plan states",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if basePlanFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--file is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationBasePlansBatchUpdateStates(cmd.Context(), args[0], basePlanFile)
		},
	}
	basePlansBatchUpdateStatesCmd.Flags().StringVar(&basePlanFile, "file", "", "Batch update JSON file")

	basePlansCmd.AddCommand(basePlansActivateCmd, basePlansDeactivateCmd, basePlansDeleteCmd, basePlansMigratePricesCmd,
		basePlansBatchMigrateCmd, basePlansBatchUpdateStatesCmd)
	monetizationCmd.AddCommand(basePlansCmd)
}

func (c *CLI) addMonetizationOffersCommands(monetizationCmd *cobra.Command) {
	offersCmd := &cobra.Command{
		Use:   "offers",
		Short: "Manage subscription offers",
	}
	var (
		offerID   string
		offerFile string
		offerIDs  []string
		confirm   bool
		pageSize  int64
		pageToken string
		all       bool
	)

	offersCreateCmd := &cobra.Command{
		Use:   "create [subscription-id] [plan-id]",
		Short: "Create an offer",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if offerID == "" || offerFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--offer-id and --file are required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationOffersCreate(cmd.Context(), args[0], args[1], offerID, offerFile)
		},
	}
	offersCreateCmd.Flags().StringVar(&offerID, "offer-id", "", "Offer ID")
	offersCreateCmd.Flags().StringVar(&offerFile, "file", "", "Offer JSON file")

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
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"subscription-id and plan-id are required").
					WithHint("Usage: gpd monetization offers list <subscription-id> <plan-id>"))
				return c.Output(result.WithServices("androidpublisher"))
			}
			return c.monetizationOffersList(cmd.Context(), args[0], args[1], pageSize, pageToken, all)
		},
	}
	offersListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	offersListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(offersListCmd, &all)

	offersDeleteCmd := &cobra.Command{
		Use:   "delete [subscription-id] [plan-id] [offer-id]",
		Short: "Delete an offer",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations")).WithServices("androidpublisher")
				return c.Output(result)
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
			if len(offerIDs) == 0 {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--offer-ids is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationOffersBatchGet(cmd.Context(), args[0], args[1], offerIDs)
		},
	}
	offersBatchGetCmd.Flags().StringSliceVar(&offerIDs, "offer-ids", nil, "Offer IDs")

	offersBatchUpdateCmd := &cobra.Command{
		Use:   "batchUpdate [subscription-id] [plan-id]",
		Short: "Batch update offers",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if offerFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--file is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationOffersBatchUpdate(cmd.Context(), args[0], args[1], offerFile)
		},
	}
	offersBatchUpdateCmd.Flags().StringVar(&offerFile, "file", "", "Batch update JSON file")

	offersBatchUpdateStatesCmd := &cobra.Command{
		Use:   "batchUpdateStates [subscription-id] [plan-id]",
		Short: "Batch update offer states",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if offerFile == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--file is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationOffersBatchUpdateStates(cmd.Context(), args[0], args[1], offerFile)
		},
	}
	offersBatchUpdateStatesCmd.Flags().StringVar(&offerFile, "file", "", "Batch update states JSON file")

	offersCmd.AddCommand(offersCreateCmd, offersGetCmd, offersListCmd, offersDeleteCmd, offersActivateCmd, offersDeactivateCmd,
		offersBatchGetCmd, offersBatchUpdateCmd, offersBatchUpdateStatesCmd)
	monetizationCmd.AddCommand(offersCmd)
}

func (c *CLI) addMonetizationOnetimeProductsCommands(monetizationCmd *cobra.Command) {
	onetimeProductsCmd := &cobra.Command{
		Use:   "onetimeproducts",
		Short: "Manage one-time products",
		Long:  "Alias of legacy in-app products for managed/consumable items.",
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

	onetimeProductsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List one-time products",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationProductsList(cmd.Context(), pageSize, pageToken, all)
		},
	}
	onetimeProductsListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	onetimeProductsListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(onetimeProductsListCmd, &all)

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
			if productID == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--product-id is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationProductsCreate(cmd.Context(), productID, productType, defaultPrice, status)
		},
	}
	onetimeProductsCreateCmd.Flags().StringVar(&productID, "product-id", "", "Product SKU")
	onetimeProductsCreateCmd.Flags().StringVar(&productType, "type", "managed", "Product type: managed, consumable")
	onetimeProductsCreateCmd.Flags().StringVar(&defaultPrice, "default-price", "", "Default price in micros (e.g., 990000 for $0.99)")
	onetimeProductsCreateCmd.Flags().StringVar(&status, "status", "active", "Product status: active, inactive")

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
	monetizationCmd.AddCommand(onetimeProductsCmd)
}

func (c *CLI) addMonetizationUtilityCommands(monetizationCmd *cobra.Command) {
	var (
		priceMicros  int64
		currencyCode string
		regionFilter []string
	)

	convertCmd := &cobra.Command{
		Use:   "convert-region-prices",
		Short: "Convert region prices",
		RunE: func(cmd *cobra.Command, args []string) error {
			if priceMicros == 0 {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"--price-micros is required")).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.monetizationConvertRegionPrices(cmd.Context(), priceMicros, currencyCode, regionFilter)
		},
	}
	convertCmd.Flags().Int64Var(&priceMicros, "price-micros", 0, "Price in micros")
	convertCmd.Flags().StringVar(&currencyCode, "currency", "USD", "Currency code")
	convertCmd.Flags().StringSliceVar(&regionFilter, "to-regions", nil, "Region codes to include")

	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List monetization capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.monetizationCapabilities(cmd.Context())
		},
	}

	monetizationCmd.AddCommand(convertCmd, capabilitiesCmd)
}
