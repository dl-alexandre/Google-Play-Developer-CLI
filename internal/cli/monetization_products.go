package cli

import (
	"context"
	stdErrors "errors"
	"net/http"
	"strconv"
	"strings"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) monetizationProductsList(ctx context.Context, _ int64, pageToken string, all bool) error {
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

	startToken := pageToken
	nextToken := ""
	var allProducts []interface{}
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			apiErr := errors.ClassifyAuthError(err)
			if apiErr == nil {
				apiErr = errors.NewAPIError(errors.CodeGeneralError, err.Error())
			}
			var gapiErr *googleapi.Error
			if stdErrors.As(err, &gapiErr) && gapiErr.Code == http.StatusForbidden &&
				strings.Contains(gapiErr.Message, "Please migrate to the new publishing API") {
				apiErr = apiErr.WithHint("This endpoint is legacy. Migrate to the new Play Publishing APIs or use monetization subscriptions/baseplans if applicable.")
			}
			result := output.NewErrorResult(apiErr).WithServices("androidpublisher")
			return c.Output(result)
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

		nextToken = ""
		if resp.TokenPagination != nil {
			nextToken = resp.TokenPagination.NextPageToken
		}
		if nextToken == "" || !all {
			break
		}
		req = req.Token(nextToken)
	}

	result := output.NewResult(allProducts)
	result.WithPagination(startToken, nextToken)
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

func (c *CLI) monetizationProductsCreate(ctx context.Context, productID, _, defaultPrice, status string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if productID == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"product ID is required"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	product := &androidpublisher.InAppProduct{
		PackageName:     c.packageName,
		Sku:             productID,
		Status:          status,
		DefaultLanguage: "en-US",
	}

	product.PurchaseType = "managedUser"

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

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	existing, err := publisher.Inappproducts.Get(c.packageName, productID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			"product not found: "+productID))
	}

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
