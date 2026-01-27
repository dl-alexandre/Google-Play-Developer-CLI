package cli

import (
	"context"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

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
	startToken := pageToken
	nextToken := ""
	var offers []interface{}
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		for _, offer := range resp.SubscriptionOffers {
			offers = append(offers, offer)
		}
		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		req = req.PageToken(nextToken)
	}
	result := output.NewResult(offers)
	result.WithPagination(startToken, nextToken)
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
			ProductId:   subscriptionID,
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
