package cli

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/playcustomapp/v1"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addCustomAppCommands() {
	customAppCmd := &cobra.Command{
		Use:   "customapp",
		Short: "Custom app publishing commands",
		Long:  "Create and publish custom apps via the Play Custom App Publishing API.",
	}

	var (
		accountID int64
		title     string
		language  string
		apkPath   string
		orgIDs    []string
	)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom app",
		Long:  "Create a custom app and upload an APK to publish it.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if accountID == 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "--account is required"))
			}
			if strings.TrimSpace(title) == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "--title is required"))
			}
			if strings.TrimSpace(language) == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "--language is required"))
			}
			if strings.TrimSpace(apkPath) == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "--apk is required"))
			}
			if err := validateCustomAppAPK(apkPath); err != nil {
				return c.OutputError(err)
			}
			return c.customAppCreate(cmd.Context(), accountID, title, language, apkPath, orgIDs)
		},
	}

	createCmd.Flags().Int64Var(&accountID, "account", 0, "Developer account ID (required)")
	createCmd.Flags().StringVar(&title, "title", "", "App title (required)")
	createCmd.Flags().StringVar(&language, "language", "", "Default listing language (BCP 47, required)")
	createCmd.Flags().StringVar(&apkPath, "apk", "", "Path to APK to upload (required)")
	createCmd.Flags().StringSliceVar(&orgIDs, "org-id", nil, "Organization IDs to grant access")

	_ = createCmd.MarkFlagRequired("account")
	_ = createCmd.MarkFlagRequired("title")
	_ = createCmd.MarkFlagRequired("language")
	_ = createCmd.MarkFlagRequired("apk")

	customAppCmd.AddCommand(createCmd)
	c.rootCmd.AddCommand(customAppCmd)
}

func validateCustomAppAPK(filePath string) *errors.APIError {
	info, err := os.Stat(filePath)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, "file not found").
			WithDetails(map[string]interface{}{"path": filePath})
	}
	if info.IsDir() {
		return errors.NewAPIError(errors.CodeValidationError, "APK path is a directory").
			WithDetails(map[string]interface{}{"path": filePath})
	}
	if strings.ToLower(filepath.Ext(filePath)) != ".apk" {
		return errors.NewAPIError(errors.CodeValidationError, "custom app upload must be an APK").
			WithHint("Provide a .apk file for --apk")
	}
	return nil
}

func (c *CLI) getPlayCustomAppService(ctx context.Context) (*playcustomapp.Service, error) {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return nil, err
	}
	svc, svcErr := client.PlayCustomApp()
	if svcErr != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, svcErr.Error())
	}
	return svc, nil
}

func (c *CLI) customAppCreate(ctx context.Context, accountID int64, title, language, apkPath string, orgIDs []string) error {
	svc, err := c.getPlayCustomAppService(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	f, err := os.Open(apkPath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	orgs := make([]*playcustomapp.Organization, 0, len(orgIDs))
	for _, orgID := range orgIDs {
		if strings.TrimSpace(orgID) == "" {
			continue
		}
		orgs = append(orgs, &playcustomapp.Organization{OrganizationId: orgID})
	}

	customApp := &playcustomapp.CustomApp{
		Title:        title,
		LanguageCode: language,
	}
	if len(orgs) > 0 {
		customApp.Organizations = orgs
	}

	resp, uploadErr := svc.Accounts.CustomApps.Create(accountID, customApp).Media(f).Context(ctx).Do()
	closeErr := f.Close()
	if uploadErr != nil {
		if closeErr != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, uploadErr.Error()).
				WithDetails(map[string]interface{}{"closeError": closeErr.Error()}))
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, uploadErr.Error()))
	}
	if closeErr != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, closeErr.Error()))
	}

	return c.Output(output.NewResult(resp).WithServices("playcustomapp"))
}
