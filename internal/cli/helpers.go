package cli

import (
	"context"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) getPublisherService(ctx context.Context) (*androidpublisher.Service, error) {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return nil, err
	}
	return client.AndroidPublisher()
}

func (c *CLI) requirePublisherService(ctx context.Context) (*androidpublisher.Service, *output.Result) {
	if err := c.requirePackage(); err != nil {
		return nil, output.NewErrorResult(err.(*errors.APIError))
	}

	publisher, err := c.getPublisherService(ctx)
	if err != nil {
		return nil, output.NewErrorResult(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	return publisher, nil
}
