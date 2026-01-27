package cli

import (
	"context"
	"encoding/json"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

var validDeveloperPermissions = []string{
	"CAN_VIEW_FINANCIAL_DATA_GLOBAL",
	"CAN_MANAGE_PERMISSIONS_GLOBAL",
	"CAN_EDIT_GAMES_GLOBAL",
	"CAN_PUBLISH_GAMES_GLOBAL",
	"CAN_REPLY_TO_REVIEWS_GLOBAL",
	"CAN_MANAGE_PUBLIC_APKS_GLOBAL",
	"CAN_MANAGE_TRACK_APKS_GLOBAL",
	"CAN_MANAGE_TRACK_USERS_GLOBAL",
	"CAN_MANAGE_PUBLIC_LISTING_GLOBAL",
	"CAN_MANAGE_DRAFT_APPS_GLOBAL",
	"CAN_CREATE_MANAGED_PLAY_APPS_GLOBAL",
	"CAN_CHANGE_MANAGED_PLAY_SETTING_GLOBAL",
	"CAN_MANAGE_ORDERS_GLOBAL",
	"CAN_MANAGE_APP_CONTENT_GLOBAL",
	"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
	"CAN_VIEW_APP_QUALITY_GLOBAL",
	"CAN_MANAGE_DEEPLINKS_GLOBAL",
}

func (c *CLI) addPermissionsCommands() {
	permissionsCmd := &cobra.Command{
		Use:   "permissions",
		Short: "Permissions management commands",
		Long:  "Manage users and grants for developer accounts and apps.",
	}

	c.addUsersCommands(permissionsCmd)
	c.addGrantsCommands(permissionsCmd)
	c.addPermissionsCapabilitiesCommand(permissionsCmd)

	c.rootCmd.AddCommand(permissionsCmd)
}

func (c *CLI) addUsersCommands(parent *cobra.Command) {
	usersCmd := &cobra.Command{
		Use:   "users",
		Short: "Manage developer account users",
		Long:  "List, create, update, and delete users with access to the developer account.",
	}

	var (
		developerID          string
		email                string
		developerPermissions []string
		expirationTime       string
		pageSize             int64
		pageToken            string
		all                  bool
		userFile             string
	)

	usersCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a user in the developer account",
		Long:  "Grants access for a user to the developer account by email.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if developerID == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--developer-id is required"))
			}
			if email == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--email is required"))
			}
			return c.permissionsUsersCreate(cmd.Context(), developerID, email, developerPermissions, expirationTime)
		},
	}
	usersCreateCmd.Flags().StringVar(&developerID, "developer-id", "", "Developer account ID (required)")
	usersCreateCmd.Flags().StringVar(&email, "email", "", "User email address (required)")
	usersCreateCmd.Flags().StringSliceVar(&developerPermissions, "developer-permissions", nil, "Developer-level permissions (comma-separated)")
	usersCreateCmd.Flags().StringVar(&expirationTime, "expiration-time", "", "Access expiration time (RFC3339 format)")
	_ = usersCreateCmd.MarkFlagRequired("developer-id")
	_ = usersCreateCmd.MarkFlagRequired("email")

	usersListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all users in the developer account",
		Long:  "Lists all users with access to the developer account with pagination support.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if developerID == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--developer-id is required"))
			}
			return c.permissionsUsersList(cmd.Context(), developerID, pageSize, pageToken, all)
		},
	}
	usersListCmd.Flags().StringVar(&developerID, "developer-id", "", "Developer account ID (required)")
	usersListCmd.Flags().Int64Var(&pageSize, "page-size", 100, "Results per page")
	usersListCmd.Flags().StringVar(&pageToken, "page-token", "", "Pagination token")
	addPaginationFlags(usersListCmd, &all)
	_ = usersListCmd.MarkFlagRequired("developer-id")

	usersDeleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Remove a user from the developer account",
		Long:  "Removes all access for the user to the developer account. Name format: developers/{developer}/users/{email}",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.permissionsUsersDelete(cmd.Context(), args[0])
		},
	}

	usersPatchCmd := &cobra.Command{
		Use:   "patch [name]",
		Short: "Update user permissions",
		Long:  "Updates access for the user to the developer account. Name format: developers/{developer}/users/{email}",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if userFile == "" && len(developerPermissions) == 0 && expirationTime == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"at least one of --developer-permissions, --expiration-time, or --file is required"))
			}
			return c.permissionsUsersPatch(cmd.Context(), args[0], developerPermissions, expirationTime, userFile)
		},
	}
	usersPatchCmd.Flags().StringSliceVar(&developerPermissions, "developer-permissions", nil, "Developer-level permissions (comma-separated)")
	usersPatchCmd.Flags().StringVar(&expirationTime, "expiration-time", "", "Access expiration time (RFC3339 format)")
	usersPatchCmd.Flags().StringVar(&userFile, "file", "", "User JSON file for patch data")

	usersGetCmd := &cobra.Command{
		Use:   "get [name]",
		Short: "Get a user's details",
		Long:  "Gets details for a specific user. Name format: developers/{developer}/users/{email}",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.permissionsUsersGet(cmd.Context(), args[0])
		},
	}

	usersCmd.AddCommand(usersCreateCmd, usersListCmd, usersDeleteCmd, usersPatchCmd, usersGetCmd)
	parent.AddCommand(usersCmd)
}

func (c *CLI) addGrantsCommands(parent *cobra.Command) {
	grantsCmd := &cobra.Command{
		Use:   "grants",
		Short: "Manage app-level permission grants",
		Long:  "Create, update, and delete app-level permission grants for users.",
	}

	var (
		email           string
		appPermissions  []string
		grantFile       string
		listPermissions bool
	)

	grantsCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Grant user access to an app",
		Long:  "Creates an app-level grant for a user. Requires --package flag.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if listPermissions {
				return c.permissionsListAvailable(cmd.Context())
			}
			if err := c.requirePackage(); err != nil {
				return c.OutputError(err.(*errors.APIError))
			}
			if email == "" {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--email is required"))
			}
			return c.permissionsGrantsCreate(cmd.Context(), email, appPermissions)
		},
	}
	grantsCreateCmd.Flags().StringVar(&email, "email", "", "User email address (required)")
	grantsCreateCmd.Flags().StringSliceVar(&appPermissions, "app-permissions", nil, "App-level permissions (comma-separated)")
	grantsCreateCmd.Flags().BoolVar(&listPermissions, "list-permissions", false, "List available app-level permission names")
	_ = grantsCreateCmd.MarkFlagRequired("email")

	grantsDeleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Revoke a user's grant",
		Long:  "Deletes an app-level grant. Name format: developers/{developer}/users/{email}/grants/{package_name}",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.permissionsGrantsDelete(cmd.Context(), args[0])
		},
	}

	grantsPatchCmd := &cobra.Command{
		Use:   "patch [name]",
		Short: "Update a user's grant",
		Long:  "Updates an app-level grant. Name format: developers/{developer}/users/{email}/grants/{package_name}",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if listPermissions {
				return c.permissionsListAvailable(cmd.Context())
			}
			if len(args) == 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"grant name argument is required"))
			}
			if grantFile == "" && len(appPermissions) == 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"at least one of --app-permissions or --file is required"))
			}
			return c.permissionsGrantsPatch(cmd.Context(), args[0], appPermissions, grantFile)
		},
	}
	grantsPatchCmd.Flags().StringSliceVar(&appPermissions, "app-permissions", nil, "App-level permissions (comma-separated)")
	grantsPatchCmd.Flags().StringVar(&grantFile, "file", "", "Grant JSON file for patch data")
	grantsPatchCmd.Flags().BoolVar(&listPermissions, "list-permissions", false, "List available app-level permission names")

	grantsCmd.AddCommand(grantsCreateCmd, grantsDeleteCmd, grantsPatchCmd)
	parent.AddCommand(grantsCmd)
}

func (c *CLI) addPermissionsCapabilitiesCommand(parent *cobra.Command) {
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List permissions management capabilities",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.permissionsCapabilities(cmd.Context())
		},
	}
	parent.AddCommand(capabilitiesCmd)
}

func (c *CLI) permissionsUsersCreate(ctx context.Context, developerID, email string, permissions []string, expirationTime string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	user := &androidpublisher.User{
		Email: email,
	}

	if len(permissions) > 0 {
		for _, perm := range permissions {
			if !isValidDeveloperPermission(perm) {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"invalid developer permission: "+perm).
					WithHint("Valid permissions: " + strings.Join(validDeveloperPermissions, ", ")))
			}
		}
		user.DeveloperAccountPermissions = permissions
	}

	if expirationTime != "" {
		user.ExpirationTime = expirationTime
	}

	parent := "developers/" + developerID
	created, err := publisher.Users.Create(parent, user).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":                     true,
		"name":                        created.Name,
		"email":                       created.Email,
		"accessState":                 created.AccessState,
		"developerAccountPermissions": created.DeveloperAccountPermissions,
		"expirationTime":              created.ExpirationTime,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) permissionsUsersList(ctx context.Context, developerID string, pageSize int64, pageToken string, all bool) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	parent := "developers/" + developerID
	req := publisher.Users.List(parent)
	if pageSize > 0 {
		req = req.PageSize(pageSize)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	startToken := pageToken
	nextToken := ""
	var allUsers []interface{}
	for {
		resp, err := req.Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}

		for _, user := range resp.Users {
			allUsers = append(allUsers, map[string]interface{}{
				"name":                        user.Name,
				"email":                       user.Email,
				"accessState":                 user.AccessState,
				"developerAccountPermissions": user.DeveloperAccountPermissions,
				"expirationTime":              user.ExpirationTime,
				"partial":                     user.Partial,
			})
		}

		nextToken = resp.NextPageToken
		if nextToken == "" || !all {
			break
		}
		req = req.PageToken(nextToken)
	}

	result := output.NewResult(allUsers)
	result.WithPagination(startToken, nextToken)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) permissionsUsersDelete(ctx context.Context, name string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	err = publisher.Users.Delete(name).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"name":    name,
		"deleted": true,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) permissionsUsersPatch(ctx context.Context, name string, permissions []string, expirationTime, userFile string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var user androidpublisher.User
	var updateMask []string

	if userFile != "" {
		data, err := os.ReadFile(userFile)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				"failed to read file: "+userFile))
		}
		if err := json.Unmarshal(data, &user); err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid JSON file"))
		}
	}

	if len(permissions) > 0 {
		for _, perm := range permissions {
			if !isValidDeveloperPermission(perm) {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"invalid developer permission: "+perm).
					WithHint("Valid permissions: " + strings.Join(validDeveloperPermissions, ", ")))
			}
		}
		user.DeveloperAccountPermissions = permissions
		updateMask = append(updateMask, "developerAccountPermissions")
	}

	if expirationTime != "" {
		user.ExpirationTime = expirationTime
		updateMask = append(updateMask, "expirationTime")
	}

	call := publisher.Users.Patch(name, &user)
	if len(updateMask) > 0 {
		call = call.UpdateMask(strings.Join(updateMask, ","))
	}

	updated, err := call.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":                     true,
		"name":                        updated.Name,
		"email":                       updated.Email,
		"accessState":                 updated.AccessState,
		"developerAccountPermissions": updated.DeveloperAccountPermissions,
		"expirationTime":              updated.ExpirationTime,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) permissionsUsersGet(ctx context.Context, name string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	parts := strings.Split(name, "/")
	if len(parts) < 4 || parts[0] != "developers" || parts[2] != "users" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"invalid user name format").
			WithHint("Expected format: developers/{developer}/users/{email}"))
	}

	developerID := parts[1]
	parent := "developers/" + developerID

	resp, err := publisher.Users.List(parent).PageSize(100).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	for _, user := range resp.Users {
		if user.Name == name {
			result := output.NewResult(map[string]interface{}{
				"name":                        user.Name,
				"email":                       user.Email,
				"accessState":                 user.AccessState,
				"developerAccountPermissions": user.DeveloperAccountPermissions,
				"expirationTime":              user.ExpirationTime,
				"partial":                     user.Partial,
				"grants":                      user.Grants,
			})
			return c.Output(result.WithServices("androidpublisher"))
		}
	}

	return c.OutputError(errors.NewAPIError(errors.CodeNotFound, "user not found: "+name))
}

func (c *CLI) permissionsGrantsCreate(ctx context.Context, email string, permissions []string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	grant := &androidpublisher.Grant{
		PackageName: c.packageName,
	}

	if len(permissions) > 0 {
		grant.AppLevelPermissions = permissions
	}

	parent := "developers/-/users/" + email
	created, err := publisher.Grants.Create(parent, grant).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":             true,
		"name":                created.Name,
		"packageName":         created.PackageName,
		"appLevelPermissions": created.AppLevelPermissions,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) permissionsGrantsDelete(ctx context.Context, name string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	err = publisher.Grants.Delete(name).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"name":    name,
		"deleted": true,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) permissionsGrantsPatch(ctx context.Context, name string, permissions []string, grantFile string) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var grant androidpublisher.Grant
	var updateMask []string

	if grantFile != "" {
		data, err := os.ReadFile(grantFile)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				"failed to read file: "+grantFile))
		}
		if err := json.Unmarshal(data, &grant); err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid JSON file"))
		}
	}

	if len(permissions) > 0 {
		grant.AppLevelPermissions = permissions
		updateMask = append(updateMask, "appLevelPermissions")
	}

	call := publisher.Grants.Patch(name, &grant)
	if len(updateMask) > 0 {
		call = call.UpdateMask(strings.Join(updateMask, ","))
	}

	updated, err := call.Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"success":             true,
		"name":                updated.Name,
		"packageName":         updated.PackageName,
		"appLevelPermissions": updated.AppLevelPermissions,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

var validAppLevelPermissions = []string{
	"CAN_VIEW_FINANCIAL_DATA",
	"CAN_MANAGE_PERMISSIONS",
	"CAN_REPLY_TO_REVIEWS",
	"CAN_MANAGE_PUBLIC_APKS",
	"CAN_MANAGE_TRACK_APKS",
	"CAN_MANAGE_TRACK_USERS",
	"CAN_MANAGE_PUBLIC_LISTING",
	"CAN_MANAGE_DRAFT_APPS",
	"CAN_MANAGE_ORDERS",
	"CAN_MANAGE_APP_CONTENT",
	"CAN_VIEW_NON_FINANCIAL_DATA",
	"CAN_VIEW_APP_QUALITY",
	"CAN_MANAGE_DEEPLINKS",
}

func (c *CLI) permissionsListAvailable(_ context.Context) error {
	appPermDescriptions := map[string]string{
		"CAN_VIEW_FINANCIAL_DATA":     "View financial data and reports",
		"CAN_MANAGE_PERMISSIONS":      "Admin - manage all permissions",
		"CAN_REPLY_TO_REVIEWS":        "Reply to user reviews",
		"CAN_MANAGE_PUBLIC_APKS":      "Release to production, manage app signing",
		"CAN_MANAGE_TRACK_APKS":       "Release to testing tracks",
		"CAN_MANAGE_TRACK_USERS":      "Manage testing tracks and tester lists",
		"CAN_MANAGE_PUBLIC_LISTING":   "Manage store listing and presence",
		"CAN_MANAGE_DRAFT_APPS":       "Edit and delete draft apps",
		"CAN_MANAGE_ORDERS":           "Manage orders and subscriptions",
		"CAN_MANAGE_APP_CONTENT":      "Manage policy-related pages",
		"CAN_VIEW_NON_FINANCIAL_DATA": "View app information (read-only)",
		"CAN_VIEW_APP_QUALITY":        "View app quality data (Vitals, Crashes)",
		"CAN_MANAGE_DEEPLINKS":        "Manage deep links setup",
	}

	devPermDescriptions := map[string]string{
		"CAN_VIEW_FINANCIAL_DATA_GLOBAL":         "View financial data globally",
		"CAN_MANAGE_PERMISSIONS_GLOBAL":          "Admin - manage all permissions globally",
		"CAN_EDIT_GAMES_GLOBAL":                  "Edit games globally",
		"CAN_PUBLISH_GAMES_GLOBAL":               "Publish games globally",
		"CAN_REPLY_TO_REVIEWS_GLOBAL":            "Reply to reviews globally",
		"CAN_MANAGE_PUBLIC_APKS_GLOBAL":          "Manage production releases globally",
		"CAN_MANAGE_TRACK_APKS_GLOBAL":           "Manage testing releases globally",
		"CAN_MANAGE_TRACK_USERS_GLOBAL":          "Manage track users globally",
		"CAN_MANAGE_PUBLIC_LISTING_GLOBAL":       "Manage store listings globally",
		"CAN_MANAGE_DRAFT_APPS_GLOBAL":           "Manage draft apps globally",
		"CAN_CREATE_MANAGED_PLAY_APPS_GLOBAL":    "Create managed Play apps globally",
		"CAN_CHANGE_MANAGED_PLAY_SETTING_GLOBAL": "Change managed Play settings globally",
		"CAN_MANAGE_ORDERS_GLOBAL":               "Manage orders globally",
		"CAN_MANAGE_APP_CONTENT_GLOBAL":          "Manage app content globally",
		"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL":     "View non-financial data globally",
		"CAN_VIEW_APP_QUALITY_GLOBAL":            "View app quality data globally",
		"CAN_MANAGE_DEEPLINKS_GLOBAL":            "Manage deep links globally",
	}

	appPerms := make([]map[string]interface{}, 0, len(validAppLevelPermissions))
	for _, perm := range validAppLevelPermissions {
		appPerms = append(appPerms, map[string]interface{}{
			"name":        perm,
			"description": appPermDescriptions[perm],
		})
	}

	devPerms := make([]map[string]interface{}, 0, len(validDeveloperPermissions))
	for _, perm := range validDeveloperPermissions {
		devPerms = append(devPerms, map[string]interface{}{
			"name":        perm,
			"description": devPermDescriptions[perm],
		})
	}

	result := output.NewResult(map[string]interface{}{
		"appLevelPermissions":  appPerms,
		"developerPermissions": devPerms,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) permissionsCapabilities(_ context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"users": map[string]interface{}{
			"operations":           []string{"create", "list", "get", "patch", "delete"},
			"developerPermissions": validDeveloperPermissions,
			"accessStates": []string{
				"ACCESS_STATE_UNSPECIFIED",
				"INVITED",
				"INVITATION_EXPIRED",
				"ACCESS_GRANTED",
				"ACCESS_EXPIRED",
			},
		},
		"grants": map[string]interface{}{
			"operations":          []string{"create", "patch", "delete"},
			"appLevelPermissions": validAppLevelPermissions,
		},
		"notes": []string{
			"User name format: developers/{developer}/users/{email}",
			"Grant name format: developers/{developer}/users/{email}/grants/{package_name}",
			"Developer ID can be found in Play Console URL",
			"Use --list-permissions flag to see detailed permission descriptions",
		},
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func isValidDeveloperPermission(perm string) bool {
	for _, valid := range validDeveloperPermissions {
		if valid == perm {
			return true
		}
	}
	return false
}
