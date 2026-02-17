package cli

import (
	"context"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) monetizationBasePlansActivate(ctx context.Context, subscriptionID, basePlanID string) error {
	publisher, errResult := c.requirePublisherService(ctx)
	if errResult != nil {
		return c.Output(errResult)
	}

	resp, err := publisher.Monetization.Subscriptions.BasePlans.Activate(c.packageName, subscriptionID, basePlanID, &androidpublisher.ActivateBasePlanRequest{}).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansDeactivate(ctx context.Context, subscriptionID, basePlanID string) error {
	publisher, errResult := c.requirePublisherService(ctx)
	if errResult != nil {
		return c.Output(errResult)
	}

	resp, err := publisher.Monetization.Subscriptions.BasePlans.Deactivate(c.packageName, subscriptionID, basePlanID, &androidpublisher.DeactivateBasePlanRequest{}).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansDelete(ctx context.Context, subscriptionID, basePlanID string) error {
	publisher, errResult := c.requirePublisherService(ctx)
	if errResult != nil {
		return c.Output(errResult)
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
	if priceMicros <= 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "price-micros must be greater than zero"))
	}

	publisher, errResult := c.requirePublisherService(ctx)
	if errResult != nil {
		return c.Output(errResult)
	}

	req := &androidpublisher.MigrateBasePlanPricesRequest{
		RegionalPriceMigrations: []*androidpublisher.RegionalPriceMigrationConfig{
			{
				RegionCode:                    regionCode,
				OldestAllowedPriceVersionTime: "1970-01-01T00:00:00Z",
			},
		},
		RegionsVersion: &androidpublisher.RegionsVersion{
			Version: "2022/02",
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
	var req androidpublisher.BatchMigrateBasePlanPricesRequest
	if err := loadJSONFile(filePath, &req); err != nil {
		return c.OutputError(err)
	}

	publisher, errResult := c.requirePublisherService(ctx)
	if errResult != nil {
		return c.Output(errResult)
	}

	resp, err := publisher.Monetization.Subscriptions.BasePlans.BatchMigratePrices(c.packageName, subscriptionID, &req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}

func (c *CLI) monetizationBasePlansBatchUpdateStates(ctx context.Context, subscriptionID, filePath string) error {
	var req androidpublisher.BatchUpdateBasePlanStatesRequest
	if err := loadJSONFile(filePath, &req); err != nil {
		return c.OutputError(err)
	}

	publisher, errResult := c.requirePublisherService(ctx)
	if errResult != nil {
		return c.Output(errResult)
	}

	resp, err := publisher.Monetization.Subscriptions.BasePlans.BatchUpdateStates(c.packageName, subscriptionID, &req).
		Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(resp).WithServices("androidpublisher"))
}
