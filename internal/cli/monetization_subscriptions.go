package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) monetizationSubscriptionsList(ctx context.Context, pageSize int64, pageToken string, all, showArchived bool) error {
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
		req = req.PageSize(pageSize)
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
		"success":   true,
		"productId": subscriptionID,
		"package":   c.packageName,
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

func (c *CLI) monetizationCapabilities(_ context.Context) error {
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
