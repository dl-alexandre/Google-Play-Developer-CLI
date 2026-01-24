package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addRecoveryCommands() {
	recoveryCmd := &cobra.Command{
		Use:   "recovery",
		Short: "App recovery commands",
		Long:  "Create and manage app recovery actions for remote in-app updates.",
	}

	var (
		versionCode     int64
		targetingFile   string
		allUsers        bool
		regions         []string
		androidSdkLevels []int64
		versionCodes    []int64
		versionRangeMin int64
		versionRangeMax int64
	)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a draft recovery action",
		Long:  "Create an app recovery action with DRAFT status. Use deploy to activate.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			if versionCode <= 0 && targetingFile == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--version-code or --file is required"))
			}
			return c.recoveryCreate(cmd.Context(), versionCode, targetingFile, allUsers, regions, androidSdkLevels, versionCodes, versionRangeMin, versionRangeMax)
		},
	}
	createCmd.Flags().Int64Var(&versionCode, "version-code", 0, "Target app version code")
	createCmd.Flags().StringVar(&targetingFile, "file", "", "JSON file with targeting configuration")
	createCmd.Flags().BoolVar(&allUsers, "all-users", true, "Target all users")
	createCmd.Flags().StringSliceVar(&regions, "regions", nil, "Target regions (comma-separated ISO codes)")
	createCmd.Flags().Int64SliceVar(&androidSdkLevels, "android-sdk-levels", nil, "Target Android SDK levels (comma-separated)")
	createCmd.Flags().Int64SliceVar(&versionCodes, "version-codes", nil, "Target specific version codes")
	createCmd.Flags().Int64Var(&versionRangeMin, "version-range-min", 0, "Minimum version code in range")
	createCmd.Flags().Int64Var(&versionRangeMax, "version-range-max", 0, "Maximum version code in range")

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List recovery actions",
		Long:  "List all app recovery actions for a package.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			return c.recoveryList(cmd.Context(), versionCode)
		},
	}
	listCmd.Flags().Int64Var(&versionCode, "version-code", 0, "Filter by version code")

	deployCmd := &cobra.Command{
		Use:   "deploy [recovery-id]",
		Short: "Deploy a recovery action",
		Long:  "Deploy a draft recovery action to activate it for targeted users.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			return c.recoveryDeploy(cmd.Context(), args[0])
		},
	}

	cancelCmd := &cobra.Command{
		Use:   "cancel [recovery-id]",
		Short: "Cancel a recovery action",
		Long:  "Cancel an active or draft recovery action.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			return c.recoveryCancel(cmd.Context(), args[0])
		},
	}

	addTargetingCmd := &cobra.Command{
		Use:   "add-targeting [recovery-id]",
		Short: "Add targeting to a recovery action",
		Long:  "Incrementally update targeting for an existing recovery action.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			if targetingFile == "" && !allUsers && len(regions) == 0 && len(androidSdkLevels) == 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"at least one targeting option is required: --file, --all-users, --regions, or --android-sdk-levels"))
			}
			return c.recoveryAddTargeting(cmd.Context(), args[0], targetingFile, allUsers, regions, androidSdkLevels)
		},
	}
	addTargetingCmd.Flags().StringVar(&targetingFile, "file", "", "JSON file with targeting update configuration")
	addTargetingCmd.Flags().BoolVar(&allUsers, "all-users", false, "Target all users")
	addTargetingCmd.Flags().StringSliceVar(&regions, "regions", nil, "Additional regions to target (comma-separated ISO codes)")
	addTargetingCmd.Flags().Int64SliceVar(&androidSdkLevels, "android-sdk-levels", nil, "Target Android SDK levels (comma-separated)")

	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List recovery capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.recoveryCapabilities(cmd.Context())
		},
	}

	recoveryCmd.AddCommand(createCmd, listCmd, deployCmd, cancelCmd, addTargetingCmd, capabilitiesCmd)
	c.rootCmd.AddCommand(recoveryCmd)
}

func (c *CLI) recoveryCreate(ctx context.Context, versionCode int64, targetingFile string, allUsers bool, regions []string, androidSdkLevels []int64, versionCodes []int64, versionRangeMin, versionRangeMax int64) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var req androidpublisher.CreateDraftAppRecoveryRequest

	if targetingFile != "" {
		data, err := os.ReadFile(targetingFile)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				"failed to read file: "+targetingFile))
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid JSON file"))
		}
	} else {
		req.RemoteInAppUpdate = &androidpublisher.RemoteInAppUpdate{
			IsRemoteInAppUpdateRequested: true,
		}

		targeting := &androidpublisher.Targeting{}

		if allUsers {
			targeting.AllUsers = &androidpublisher.AllUsers{
				IsAllUsersRequested: true,
			}
		}

		if len(regions) > 0 {
			targeting.Regions = &androidpublisher.Regions{
				RegionCode: regions,
			}
		}

		if len(androidSdkLevels) > 0 {
			targeting.AndroidSdks = &androidpublisher.AndroidSdks{
				SdkLevels: androidSdkLevels,
			}
		}

		if len(versionCodes) > 0 {
			targeting.VersionList = &androidpublisher.AppVersionList{
				VersionCodes: versionCodes,
			}
		} else if versionCode > 0 {
			targeting.VersionList = &androidpublisher.AppVersionList{
				VersionCodes: []int64{versionCode},
			}
		}

		if versionRangeMin > 0 || versionRangeMax > 0 {
			targeting.VersionRange = &androidpublisher.AppVersionRange{}
			if versionRangeMin > 0 {
				targeting.VersionRange.VersionCodeStart = versionRangeMin
			}
			if versionRangeMax > 0 {
				targeting.VersionRange.VersionCodeEnd = versionRangeMax
			}
		}

		req.Targeting = targeting
	}

	action, err := publisher.Apprecovery.Create(c.packageName, &req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":        true,
		"appRecoveryId":  action.AppRecoveryId,
		"status":         action.Status,
		"createTime":     action.CreateTime,
		"lastUpdateTime": action.LastUpdateTime,
		"targeting":      action.Targeting,
		"package":        c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) recoveryList(ctx context.Context, versionCode int64) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	call := publisher.Apprecovery.List(c.packageName)
	if versionCode > 0 {
		call = call.VersionCode(versionCode)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var actions []interface{}
	for _, action := range resp.RecoveryActions {
		actions = append(actions, map[string]interface{}{
			"appRecoveryId":         action.AppRecoveryId,
			"status":                action.Status,
			"createTime":            action.CreateTime,
			"deployTime":            action.DeployTime,
			"cancelTime":            action.CancelTime,
			"lastUpdateTime":        action.LastUpdateTime,
			"targeting":             action.Targeting,
			"remoteInAppUpdateData": action.RemoteInAppUpdateData,
		})
	}

	result := output.NewResult(actions)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) recoveryDeploy(ctx context.Context, recoveryID string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appRecoveryID, err := parseRecoveryID(recoveryID)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, err.Error()))
	}

	_, err = publisher.Apprecovery.Deploy(c.packageName, appRecoveryID, &androidpublisher.DeployAppRecoveryRequest{}).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":       true,
		"appRecoveryId": appRecoveryID,
		"deployed":      true,
		"package":       c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) recoveryCancel(ctx context.Context, recoveryID string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appRecoveryID, err := parseRecoveryID(recoveryID)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, err.Error()))
	}

	_, err = publisher.Apprecovery.Cancel(c.packageName, appRecoveryID, &androidpublisher.CancelAppRecoveryRequest{}).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":       true,
		"appRecoveryId": appRecoveryID,
		"cancelled":     true,
		"package":       c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) recoveryAddTargeting(ctx context.Context, recoveryID, targetingFile string, allUsers bool, regions []string, androidSdkLevels []int64) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	appRecoveryID, err := parseRecoveryID(recoveryID)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, err.Error()))
	}

	var req androidpublisher.AddTargetingRequest

	if targetingFile != "" {
		data, err := os.ReadFile(targetingFile)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				"failed to read file: "+targetingFile))
		}
		if err := json.Unmarshal(data, &req); err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid JSON file"))
		}
	} else {
		targetingUpdate := &androidpublisher.TargetingUpdate{}

		if allUsers {
			targetingUpdate.AllUsers = &androidpublisher.AllUsers{
				IsAllUsersRequested: true,
			}
		}

		if len(regions) > 0 {
			targetingUpdate.Regions = &androidpublisher.Regions{
				RegionCode: regions,
			}
		}

		if len(androidSdkLevels) > 0 {
			targetingUpdate.AndroidSdks = &androidpublisher.AndroidSdks{
				SdkLevels: androidSdkLevels,
			}
		}

		req.TargetingUpdate = targetingUpdate
	}

	_, err = publisher.Apprecovery.AddTargeting(c.packageName, appRecoveryID, &req).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":         true,
		"appRecoveryId":   appRecoveryID,
		"targetingAdded":  true,
		"package":         c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) recoveryCapabilities(ctx context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"operations": []string{"create", "list", "deploy", "cancel", "add-targeting"},
		"recoveryStatuses": []string{
			"RECOVERY_STATUS_UNSPECIFIED",
			"RECOVERY_STATUS_ACTIVE",
			"RECOVERY_STATUS_CANCELED",
			"RECOVERY_STATUS_DRAFT",
			"RECOVERY_STATUS_GENERATION_IN_PROGRESS",
		},
		"targetingOptions": map[string]interface{}{
			"allUsers":    "Target all users",
			"regions":     "Target specific regions by ISO country codes",
			"sdkLevels":   "Target specific Android SDK levels",
			"versionList": "Target specific app version codes",
			"versionRange": "Target app version code ranges",
		},
		"notes": []string{
			"Create returns a draft recovery action",
			"Deploy activates the recovery for targeted users",
			"Only criteria selected during creation can be expanded via add-targeting",
			"Cancelled actions cannot be resumed",
		},
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func parseRecoveryID(recoveryID string) (int64, error) {
	id, err := strconv.ParseInt(recoveryID, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid recovery ID: %s", recoveryID)
	}
	return id, nil
}
