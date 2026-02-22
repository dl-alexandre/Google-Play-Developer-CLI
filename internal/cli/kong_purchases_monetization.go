package cli

import (
	"github.com/dl-alexandre/gpd/internal/errors"
)

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
	return errors.NewAPIError(errors.CodeGeneralError, "purchases products acknowledge not yet implemented")
}

// PurchasesProductsConsumeCmd consumes a product purchase.
type PurchasesProductsConsumeCmd struct {
	ProductID string `help:"Product ID" required:""`
	Token     string `help:"Purchase token" required:""`
}

// Run executes the consume command.
func (cmd *PurchasesProductsConsumeCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "purchases products consume not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "purchases subscriptions acknowledge not yet implemented")
}

// PurchasesSubscriptionsCancelCmd cancels a subscription.
type PurchasesSubscriptionsCancelCmd struct {
	SubscriptionID string `help:"Subscription ID" required:""`
	Token          string `help:"Purchase token" required:""`
}

// Run executes the cancel subscription command.
func (cmd *PurchasesSubscriptionsCancelCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "purchases subscriptions cancel not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "purchases subscriptions defer not yet implemented")
}

// PurchasesSubscriptionsRefundCmd refunds a subscription.
type PurchasesSubscriptionsRefundCmd struct {
	SubscriptionID string `help:"Subscription ID" required:""`
	Token          string `help:"Purchase token" required:""`
}

// Run executes the refund subscription command.
func (cmd *PurchasesSubscriptionsRefundCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "purchases subscriptions refund not yet implemented")
}

// PurchasesSubscriptionsRevokeCmd revokes a subscription.
type PurchasesSubscriptionsRevokeCmd struct {
	SubscriptionID string `help:"Subscription ID (v1 API)"`
	Token          string `help:"Purchase token" required:""`
	RevokeType     string `help:"Revoke type for v2: fullRefund, partialRefund, itemBasedRefund"`
}

// Run executes the revoke subscription command.
func (cmd *PurchasesSubscriptionsRevokeCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "purchases subscriptions revoke not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "purchases verify not yet implemented")
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

// Run executes the voided list command.
func (cmd *PurchasesVoidedListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "purchases voided list not yet implemented")
}

// PurchasesCapabilitiesCmd lists purchase verification capabilities.
type PurchasesCapabilitiesCmd struct{}

// Run executes the capabilities command.
func (cmd *PurchasesCapabilitiesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "purchases capabilities not yet implemented")
}

// ============================================================================
// Monetization Commands
// ============================================================================

// MonetizationCmd contains monetization commands.
type MonetizationCmd struct {
	Products      MonetizationProductsCmd      `cmd:"" help:"Manage products"`
	Subscriptions MonetizationSubscriptionsCmd `cmd:"" help:"Manage subscriptions"`
	BasePlans     MonetizationBasePlansCmd     `cmd:"" help:"Manage base plans"`
	Offers        MonetizationOffersCmd        `cmd:"" help:"Manage offers"`
	Capabilities  MonetizationCapabilitiesCmd  `cmd:"" help:"List monetization capabilities"`
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

// Run executes the list products command.
func (cmd *MonetizationProductsListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization products list not yet implemented")
}

// MonetizationProductsGetCmd gets an in-app product.
type MonetizationProductsGetCmd struct {
	ProductID string `arg:"" help:"Product ID (SKU)" required:""`
}

// Run executes the get product command.
func (cmd *MonetizationProductsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization products get not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "monetization products create not yet implemented")
}

// MonetizationProductsUpdateCmd updates an in-app product.
type MonetizationProductsUpdateCmd struct {
	ProductID    string `arg:"" help:"Product ID (SKU)" required:""`
	DefaultPrice string `help:"Default price in micros"`
	Status       string `help:"Product status: active, inactive"`
}

// Run executes the update product command.
func (cmd *MonetizationProductsUpdateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization products update not yet implemented")
}

// MonetizationProductsDeleteCmd deletes an in-app product.
type MonetizationProductsDeleteCmd struct {
	ProductID string `arg:"" help:"Product ID (SKU)" required:""`
}

// Run executes the delete product command.
func (cmd *MonetizationProductsDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization products delete not yet implemented")
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

// Run executes the list subscriptions command.
func (cmd *MonetizationSubscriptionsListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions list not yet implemented")
}

// MonetizationSubscriptionsGetCmd gets a subscription product.
type MonetizationSubscriptionsGetCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
}

// Run executes the get subscription command.
func (cmd *MonetizationSubscriptionsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions get not yet implemented")
}

// MonetizationSubscriptionsCreateCmd creates a subscription.
type MonetizationSubscriptionsCreateCmd struct {
	SubscriptionID string `help:"Subscription product ID" required:""`
	File           string `help:"Subscription JSON file" required:"" type:"existingfile"`
}

// Run executes the create subscription command.
func (cmd *MonetizationSubscriptionsCreateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions create not yet implemented")
}

// MonetizationSubscriptionsUpdateCmd updates a subscription.
type MonetizationSubscriptionsUpdateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	File           string `help:"Subscription JSON file" required:"" type:"existingfile"`
}

// Run executes the update subscription command.
func (cmd *MonetizationSubscriptionsUpdateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions update not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions patch not yet implemented")
}

// MonetizationSubscriptionsDeleteCmd deletes a subscription.
type MonetizationSubscriptionsDeleteCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	Confirm        bool   `help:"Confirm destructive operation" required:""`
}

// Run executes the delete subscription command.
func (cmd *MonetizationSubscriptionsDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions delete not yet implemented")
}

// MonetizationSubscriptionsArchiveCmd archives a subscription.
type MonetizationSubscriptionsArchiveCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
}

// Run executes the archive subscription command.
func (cmd *MonetizationSubscriptionsArchiveCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions archive not yet implemented")
}

// MonetizationSubscriptionsBatchGetCmd batch gets subscriptions.
type MonetizationSubscriptionsBatchGetCmd struct {
	IDs []string `help:"Subscription IDs" required:""`
}

// Run executes the batch get subscriptions command.
func (cmd *MonetizationSubscriptionsBatchGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions batch get not yet implemented")
}

// MonetizationSubscriptionsBatchUpdateCmd batch updates subscriptions.
type MonetizationSubscriptionsBatchUpdateCmd struct {
	File string `help:"Batch update JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update subscriptions command.
func (cmd *MonetizationSubscriptionsBatchUpdateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization subscriptions batch update not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "monetization base-plans activate not yet implemented")
}

// MonetizationBasePlansDeactivateCmd deactivates a base plan.
type MonetizationBasePlansDeactivateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
}

// Run executes the deactivate base plan command.
func (cmd *MonetizationBasePlansDeactivateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization base-plans deactivate not yet implemented")
}

// MonetizationBasePlansDeleteCmd deletes a base plan.
type MonetizationBasePlansDeleteCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	Confirm        bool   `help:"Confirm destructive operation" required:""`
}

// Run executes the delete base plan command.
func (cmd *MonetizationBasePlansDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization base-plans delete not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "monetization base-plans migrate-prices not yet implemented")
}

// MonetizationBasePlansBatchMigrateCmd batch migrates base plan prices.
type MonetizationBasePlansBatchMigrateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	File           string `help:"Batch migrate JSON file" required:"" type:"existingfile"`
}

// Run executes the batch migrate prices command.
func (cmd *MonetizationBasePlansBatchMigrateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization base-plans batch-migrate not yet implemented")
}

// MonetizationBasePlansBatchUpdateStatesCmd batch updates base plan states.
type MonetizationBasePlansBatchUpdateStatesCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	File           string `help:"Batch update JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update states command.
func (cmd *MonetizationBasePlansBatchUpdateStatesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization base-plans batch-update-states not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers create not yet implemented")
}

// MonetizationOffersGetCmd gets an offer.
type MonetizationOffersGetCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `arg:"" help:"Offer ID" required:""`
}

// Run executes the get offer command.
func (cmd *MonetizationOffersGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers get not yet implemented")
}

// MonetizationOffersListCmd lists offers.
type MonetizationOffersListCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	PageSize       int64  `help:"Results per page" default:"100"`
	PageToken      string `help:"Pagination token"`
	All            bool   `help:"Fetch all pages"`
}

// Run executes the list offers command.
func (cmd *MonetizationOffersListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers list not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers delete not yet implemented")
}

// MonetizationOffersActivateCmd activates an offer.
type MonetizationOffersActivateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `arg:"" help:"Offer ID" required:""`
}

// Run executes the activate offer command.
func (cmd *MonetizationOffersActivateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers activate not yet implemented")
}

// MonetizationOffersDeactivateCmd deactivates an offer.
type MonetizationOffersDeactivateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	OfferID        string `arg:"" help:"Offer ID" required:""`
}

// Run executes the deactivate offer command.
func (cmd *MonetizationOffersDeactivateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers deactivate not yet implemented")
}

// MonetizationOffersBatchGetCmd batch gets offers.
type MonetizationOffersBatchGetCmd struct {
	SubscriptionID string   `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string   `arg:"" help:"Base plan ID" required:""`
	OfferIDs       []string `help:"Offer IDs" required:""`
}

// Run executes the batch get offers command.
func (cmd *MonetizationOffersBatchGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers batch get not yet implemented")
}

// MonetizationOffersBatchUpdateCmd batch updates offers.
type MonetizationOffersBatchUpdateCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	File           string `help:"Batch update JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update offers command.
func (cmd *MonetizationOffersBatchUpdateCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers batch update not yet implemented")
}

// MonetizationOffersBatchUpdateStatesCmd batch updates offer states.
type MonetizationOffersBatchUpdateStatesCmd struct {
	SubscriptionID string `arg:"" help:"Subscription ID" required:""`
	BasePlanID     string `arg:"" help:"Base plan ID" required:""`
	File           string `help:"Batch update states JSON file" required:"" type:"existingfile"`
}

// Run executes the batch update states command.
func (cmd *MonetizationOffersBatchUpdateStatesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization offers batch update states not yet implemented")
}

// MonetizationCapabilitiesCmd lists monetization capabilities.
type MonetizationCapabilitiesCmd struct{}

// Run executes the capabilities command.
func (cmd *MonetizationCapabilitiesCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "monetization capabilities not yet implemented")
}
