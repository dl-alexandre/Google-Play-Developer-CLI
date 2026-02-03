package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addAppsCommands() {
	appsCmd := &cobra.Command{
		Use:   "apps",
		Short: "App discovery commands",
		Long:  "List apps accessible in the Google Play developer account.",
	}

	var (
		pageSize  int64
		pageToken string
		all       bool
	)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List apps in the developer account",
		Long:  "List apps accessible to the authenticated account with pagination support.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.appsList(cmd.Context(), pageSize, pageToken, all)
		},
	}
	listCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	listCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(listCmd, &all)

	getCmd := &cobra.Command{
		Use:   "get [package]",
		Short: "Get app details",
		Long:  "Get details for a specific app package. Uses --package when omitted.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			packageName := c.packageName
			if len(args) > 0 {
				packageName = args[0]
			}
			if packageName == "" {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "package name is required").
					WithHint("Provide a package as an argument or set --package")).WithServices("playdeveloperreporting")
				return c.Output(result)
			}
			return c.appsGet(cmd.Context(), packageName)
		},
	}

	appsCmd.AddCommand(listCmd, getCmd)
	c.rootCmd.AddCommand(appsCmd)
}

func (c *CLI) appsList(ctx context.Context, pageSize int64, pageToken string, all bool) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("playdeveloperreporting")
		return c.Output(result)
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeGeneralError, err.Error())).
			WithServices("playdeveloperreporting")
		return c.Output(result)
	}

	req := reporting.Apps.Search()
	if pageSize > 0 {
		req = req.PageSize(pageSize)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	startToken := pageToken
	nextToken := ""
	apps := make([]interface{}, 0)
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			result := output.NewErrorResult(errors.NewAPIError(errors.CodeGeneralError, err.Error())).
				WithServices("playdeveloperreporting")
			return c.Output(result)
		}

		for _, app := range resp.Apps {
			apps = append(apps, app)
		}

		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		req = req.PageToken(nextToken)
	}

	result := output.NewResult(apps)
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("playdeveloperreporting"))
}

func (c *CLI) appsGet(ctx context.Context, packageName string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("playdeveloperreporting")
		return c.Output(result)
	}

	reporting, err := client.PlayReporting()
	if err != nil {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeGeneralError, err.Error())).
			WithServices("playdeveloperreporting")
		return c.Output(result)
	}

	req := reporting.Apps.Search().PageSize(100)
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			result := output.NewErrorResult(errors.NewAPIError(errors.CodeGeneralError, err.Error())).
				WithServices("playdeveloperreporting")
			return c.Output(result)
		}

		for _, app := range resp.Apps {
			if app.PackageName == packageName {
				return c.Output(output.NewResult(app).WithServices("playdeveloperreporting"))
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		req = req.PageToken(resp.NextPageToken)
	}

	result := output.NewErrorResult(errors.NewAPIError(errors.CodeNotFound,
		fmt.Sprintf("app not found for package %s", packageName))).WithServices("playdeveloperreporting")
	return c.Output(result)
}
