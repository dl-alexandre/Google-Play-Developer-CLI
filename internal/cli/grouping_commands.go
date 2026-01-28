package cli

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/games/v1"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addGroupingCommands() {
	groupingCmd := &cobra.Command{
		Use:   "grouping",
		Short: "Play Grouping API commands",
		Long:  "Generate Play Grouping API tokens via Play Games Services.",
	}

	var persona string
	var recallSessionID string

	tokenCmd := &cobra.Command{
		Use:   "token",
		Short: "Generate a Play Grouping API token",
		Long:  "Generate a Play Grouping API token for the configured package.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			if strings.TrimSpace(persona) == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "--persona is required"))
			}
			return c.groupingToken(cmd.Context(), c.packageName, persona)
		},
	}
	tokenCmd.Flags().StringVar(&persona, "persona", "", "Persona identifier for the user (required)")
	_ = tokenCmd.MarkFlagRequired("persona")

	tokenRecallCmd := &cobra.Command{
		Use:   "token-recall",
		Short: "Generate a Play Grouping API token using Recall",
		Long:  "Generate a Play Grouping API token for a Recall session.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			if strings.TrimSpace(persona) == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "--persona is required"))
			}
			if strings.TrimSpace(recallSessionID) == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "--recall-session-id is required"))
			}
			return c.groupingTokenRecall(cmd.Context(), c.packageName, persona, recallSessionID)
		},
	}
	tokenRecallCmd.Flags().StringVar(&persona, "persona", "", "Persona identifier for the user (required)")
	tokenRecallCmd.Flags().StringVar(&recallSessionID, "recall-session-id", "", "Recall session ID (required)")
	_ = tokenRecallCmd.MarkFlagRequired("persona")
	_ = tokenRecallCmd.MarkFlagRequired("recall-session-id")

	groupingCmd.AddCommand(tokenCmd, tokenRecallCmd)
	c.rootCmd.AddCommand(groupingCmd)
}

func (c *CLI) getGamesService(ctx context.Context) (*games.Service, error) {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return nil, err
	}
	svc, svcErr := client.Games()
	if svcErr != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, svcErr.Error())
	}
	return svc, nil
}

func (c *CLI) groupingToken(ctx context.Context, packageName, persona string) error {
	svc, err := c.getGamesService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	resp, err := svc.Accesstokens.GeneratePlayGroupingApiToken().
		PackageName(packageName).
		Persona(persona).
		Context(ctx).
		Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	return c.Output(output.NewResult(resp).WithServices("games"))
}

func (c *CLI) groupingTokenRecall(ctx context.Context, packageName, persona, recallSessionID string) error {
	svc, err := c.getGamesService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	resp, err := svc.Accesstokens.GenerateRecallPlayGroupingApiToken().
		PackageName(packageName).
		Persona(persona).
		RecallSessionId(recallSessionID).
		Context(ctx).
		Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	return c.Output(output.NewResult(resp).WithServices("games"))
}
