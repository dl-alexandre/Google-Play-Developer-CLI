package cli

import (
	"context"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/playintegrity/v1"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addIntegrityCommands() {
	integrityCmd := &cobra.Command{
		Use:   "integrity",
		Short: "Play Integrity API commands",
		Long:  "Decode Play Integrity tokens and inspect integrity payloads.",
	}

	var token string
	var tokenFile string

	decodeCmd := &cobra.Command{
		Use:   "decode",
		Short: "Decode a Play Integrity token",
		Long:  "Decode a standard Play Integrity token for the configured package.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			tokenValue, err := resolveTokenInput(token, tokenFile)
			if err != nil {
				return c.OutputError(err)
			}
			return c.integrityDecode(cmd.Context(), c.packageName, tokenValue)
		},
	}
	decodeCmd.Flags().StringVar(&token, "token", "", "Integrity token value")
	decodeCmd.Flags().StringVar(&tokenFile, "token-file", "", "File containing the integrity token")

	integrityCmd.AddCommand(decodeCmd)
	c.rootCmd.AddCommand(integrityCmd)
}

func resolveTokenInput(token, tokenFile string) (string, *errors.APIError) {
	if token != "" && tokenFile != "" {
		return "", errors.NewAPIError(errors.CodeValidationError, "provide --token or --token-file, not both")
	}
	if token != "" {
		value := strings.TrimSpace(token)
		if value == "" {
			return "", errors.NewAPIError(errors.CodeValidationError, "integrity token is required").
				WithHint("Provide a non-empty value for --token")
		}
		return value, nil
	}
	if tokenFile == "" {
		return "", errors.NewAPIError(errors.CodeValidationError, "integrity token is required").
			WithHint("Provide --token or --token-file")
	}
	data, err := os.ReadFile(tokenFile)
	if err != nil {
		return "", errors.NewAPIError(errors.CodeValidationError, "failed to read token file").
			WithDetails(map[string]interface{}{"path": tokenFile, "error": err.Error()})
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", errors.NewAPIError(errors.CodeValidationError, "token file is empty").
			WithDetails(map[string]interface{}{"path": tokenFile})
	}
	return value, nil
}

func (c *CLI) getPlayIntegrityService(ctx context.Context) (*playintegrity.Service, error) {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return nil, err
	}
	svc, svcErr := client.PlayIntegrity()
	if svcErr != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, svcErr.Error())
	}
	return svc, nil
}

func (c *CLI) integrityDecode(ctx context.Context, packageName, token string) error {
	svc, err := c.getPlayIntegrityService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	resp, err := svc.V1.DecodeIntegrityToken(packageName, &playintegrity.DecodeIntegrityTokenRequest{
		IntegrityToken: token,
	}).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	return c.Output(output.NewResult(resp).WithServices("playintegrity"))
}
