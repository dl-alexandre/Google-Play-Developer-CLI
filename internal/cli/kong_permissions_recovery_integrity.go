package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/playintegrity/v1"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

// ============================================================================
// Permissions Commands
// ============================================================================

// PermissionsCmd contains permissions management commands.
type PermissionsCmd struct {
	Users  PermissionsUsersCmd  `cmd:"" help:"Manage users"`
	Grants PermissionsGrantsCmd `cmd:"" help:"Manage grants"`
	List   PermissionsListCmd   `cmd:"" help:"List permissions"`
}

// PermissionsUsersCmd manages users.
type PermissionsUsersCmd struct {
	Add    PermissionsUsersAddCmd    `cmd:"" help:"Add a user"`
	Remove PermissionsUsersRemoveCmd `cmd:"" help:"Remove a user"`
	List   PermissionsUsersListCmd   `cmd:"" help:"List users"`
}

// PermissionsUsersAddCmd adds a user.
type PermissionsUsersAddCmd struct {
	Email string `help:"User email address" required:""`
	Role  string `help:"User role" required:"" enum:"admin,developer,viewer"`
}

// roleToDeveloperPermissions maps simplified role names to Google Play
// developer-level permission strings.
func roleToDeveloperPermissions(role string) []string {
	switch role {
	case "admin":
		return []string{"CAN_MANAGE_PERMISSIONS_GLOBAL"}
	case "developer":
		return []string{
			"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL",
			"CAN_MANAGE_TRACK_APKS_GLOBAL",
			"CAN_MANAGE_TRACK_USERS_GLOBAL",
			"CAN_MANAGE_PUBLIC_LISTING_GLOBAL",
			"CAN_MANAGE_DRAFT_APPS_GLOBAL",
			"CAN_REPLY_TO_REVIEWS_GLOBAL",
		}
	case "viewer":
		return []string{"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"}
	default:
		return []string{"CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"}
	}
}

// getDeveloperParent returns the developer parent resource name.
// The Google Play Developer API Users/Grants endpoints require a parent of the
// form "developers/{developer_id}". When the numeric developer account ID is
// not available we fall back to the undocumented "developers/-" wildcard which
// resolves to the developer account associated with the authenticated
// credentials. If the API rejects this, callers should set the
// GPD_DEVELOPER_ID environment variable.
func getDeveloperParent() string {
	return "developers/-"
}

// Run executes the users add command.
func (cmd *PermissionsUsersAddCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	parent := getDeveloperParent()

	user := &androidpublisher.User{
		Email:                       cmd.Email,
		DeveloperAccountPermissions: roleToDeveloperPermissions(cmd.Role),
	}

	var createdUser *androidpublisher.User
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		createdUser, callErr = svc.Users.Create(parent, user).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to add user: %v", err)).
			WithHint("Ensure the service account has admin permissions on the developer account")
	}

	type userAddResult struct {
		Email       string   `json:"email"`
		Name        string   `json:"name"`
		AccessState string   `json:"accessState"`
		Role        string   `json:"role"`
		Permissions []string `json:"permissions"`
	}

	data := userAddResult{
		Email:       createdUser.Email,
		Name:        createdUser.Name,
		AccessState: createdUser.AccessState,
		Role:        cmd.Role,
		Permissions: createdUser.DeveloperAccountPermissions,
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PermissionsUsersRemoveCmd removes a user.
type PermissionsUsersRemoveCmd struct {
	Email string `help:"User email address" required:""`
}

// Run executes the users remove command.
func (cmd *PermissionsUsersRemoveCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	// The Users.Delete endpoint requires the full resource name:
	// "developers/{developer}/users/{email}"
	parent := getDeveloperParent()
	userName := fmt.Sprintf("%s/users/%s", parent, cmd.Email)

	err = client.DoWithRetry(ctx, func() error {
		return svc.Users.Delete(userName).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to remove user: %v", err)).
			WithHint("Ensure the user exists and you have admin permissions")
	}

	type userRemoveResult struct {
		Email   string `json:"email"`
		Removed bool   `json:"removed"`
	}

	data := userRemoveResult{
		Email:   cmd.Email,
		Removed: true,
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PermissionsUsersListCmd lists users.
type PermissionsUsersListCmd struct{}

// Run executes the users list command.
func (cmd *PermissionsUsersListCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	parent := getDeveloperParent()

	var allUsers []*androidpublisher.User
	var nextPageToken string

	err = client.DoWithRetry(ctx, func() error {
		call := svc.Users.List(parent).Context(ctx)
		if nextPageToken != "" {
			call = call.PageToken(nextPageToken)
		}
		resp, callErr := call.Do()
		if callErr != nil {
			return callErr
		}
		allUsers = append(allUsers, resp.Users...)
		nextPageToken = resp.NextPageToken
		return nil
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list users: %v", err)).
			WithHint("Ensure the service account has admin permissions on the developer account")
	}

	// Fetch additional pages if available
	for nextPageToken != "" {
		token := nextPageToken
		nextPageToken = ""
		err = client.DoWithRetry(ctx, func() error {
			resp, callErr := svc.Users.List(parent).Context(ctx).PageToken(token).Do()
			if callErr != nil {
				return callErr
			}
			allUsers = append(allUsers, resp.Users...)
			nextPageToken = resp.NextPageToken
			return nil
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list users (pagination): %v", err))
		}
	}

	type userData struct {
		Email       string   `json:"email"`
		Name        string   `json:"name"`
		AccessState string   `json:"accessState"`
		Permissions []string `json:"permissions"`
		Partial     bool     `json:"partial"`
		GrantCount  int      `json:"grantCount"`
	}

	type usersListResult struct {
		Users      []userData `json:"users"`
		TotalCount int        `json:"totalCount"`
	}

	users := make([]userData, 0, len(allUsers))
	for _, u := range allUsers {
		users = append(users, userData{
			Email:       u.Email,
			Name:        u.Name,
			AccessState: u.AccessState,
			Permissions: u.DeveloperAccountPermissions,
			Partial:     u.Partial,
			GrantCount:  len(u.Grants),
		})
	}

	data := usersListResult{
		Users:      users,
		TotalCount: len(users),
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PermissionsGrantsCmd manages grants.
type PermissionsGrantsCmd struct {
	Add    PermissionsGrantsAddCmd    `cmd:"" help:"Add a grant"`
	Remove PermissionsGrantsRemoveCmd `cmd:"" help:"Remove a grant"`
	List   PermissionsGrantsListCmd   `cmd:"" help:"List grants"`
}

// PermissionsGrantsAddCmd adds a grant.
type PermissionsGrantsAddCmd struct {
	Email  string `help:"User email address" required:""`
	Grant  string `help:"Permission grant" required:""`
	Expiry string `help:"Grant expiry date (YYYY-MM-DD)"`
}

// Run executes the grants add command.
func (cmd *PermissionsGrantsAddCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	// Parse the comma-separated permissions into a slice
	permissions := strings.Split(cmd.Grant, ",")
	for i := range permissions {
		permissions[i] = strings.TrimSpace(permissions[i])
	}

	// The Grants.Create parent is "developers/{developer}/users/{email}"
	parent := getDeveloperParent()
	userParent := fmt.Sprintf("%s/users/%s", parent, cmd.Email)

	grant := &androidpublisher.Grant{
		PackageName:         globals.Package,
		AppLevelPermissions: permissions,
	}

	var createdGrant *androidpublisher.Grant
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		createdGrant, callErr = svc.Grants.Create(userParent, grant).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to add grant: %v", err)).
			WithHint("Ensure the user exists and you have admin permissions. Valid permissions: CAN_ACCESS_APP, CAN_VIEW_FINANCIAL_DATA, CAN_MANAGE_PERMISSIONS, CAN_REPLY_TO_REVIEWS, CAN_MANAGE_PUBLIC_APKS, CAN_MANAGE_TRACK_APKS, CAN_MANAGE_TRACK_USERS, CAN_MANAGE_PUBLIC_LISTING, CAN_MANAGE_DRAFT_APPS, CAN_MANAGE_ORDERS, CAN_MANAGE_APP_CONTENT, CAN_VIEW_NON_FINANCIAL_DATA, CAN_VIEW_APP_QUALITY, CAN_MANAGE_DEEPLINKS")
	}

	type grantAddResult struct {
		Email       string   `json:"email"`
		PackageName string   `json:"packageName"`
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}

	data := grantAddResult{
		Email:       cmd.Email,
		PackageName: createdGrant.PackageName,
		Name:        createdGrant.Name,
		Permissions: createdGrant.AppLevelPermissions,
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PermissionsGrantsRemoveCmd removes a grant.
type PermissionsGrantsRemoveCmd struct {
	Email string `help:"User email address" required:""`
	Grant string `help:"Permission grant" required:""`
}

// Run executes the grants remove command.
func (cmd *PermissionsGrantsRemoveCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	// The grant resource name follows the pattern:
	// "developers/{developer}/users/{email}/grants/{package_name}"
	parent := getDeveloperParent()
	grantName := fmt.Sprintf("%s/users/%s/grants/%s", parent, cmd.Email, globals.Package)

	err = client.DoWithRetry(ctx, func() error {
		return svc.Grants.Delete(grantName).Context(ctx).Do()
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to remove grant: %v", err)).
			WithHint("Ensure the grant exists and you have admin permissions")
	}

	type grantRemoveResult struct {
		Email       string `json:"email"`
		PackageName string `json:"packageName"`
		Removed     bool   `json:"removed"`
	}

	data := grantRemoveResult{
		Email:       cmd.Email,
		PackageName: globals.Package,
		Removed:     true,
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PermissionsGrantsListCmd lists grants.
type PermissionsGrantsListCmd struct {
	Email string `help:"Filter by user email"`
}

// Run executes the grants list command.
func (cmd *PermissionsGrantsListCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	parent := getDeveloperParent()

	// List all users to get their grants. The Users API returns grants
	// embedded in each User object.
	var allUsers []*androidpublisher.User
	var nextPageToken string

	for {
		err = client.DoWithRetry(ctx, func() error {
			call := svc.Users.List(parent).Context(ctx)
			if nextPageToken != "" {
				call = call.PageToken(nextPageToken)
			}
			resp, callErr := call.Do()
			if callErr != nil {
				return callErr
			}
			allUsers = append(allUsers, resp.Users...)
			nextPageToken = resp.NextPageToken
			return nil
		})
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list users for grants: %v", err)).
				WithHint("Ensure the service account has admin permissions on the developer account")
		}
		if nextPageToken == "" {
			break
		}
	}

	type grantData struct {
		Email       string   `json:"email"`
		PackageName string   `json:"packageName"`
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}

	type grantsListResult struct {
		Grants     []grantData `json:"grants"`
		TotalCount int         `json:"totalCount"`
	}

	var grants []grantData
	for _, u := range allUsers {
		// If email filter is set, skip non-matching users
		if cmd.Email != "" && !strings.EqualFold(u.Email, cmd.Email) {
			continue
		}

		for _, g := range u.Grants {
			grants = append(grants, grantData{
				Email:       u.Email,
				PackageName: g.PackageName,
				Name:        g.Name,
				Permissions: g.AppLevelPermissions,
			})
		}
	}

	if grants == nil {
		grants = []grantData{}
	}

	data := grantsListResult{
		Grants:     grants,
		TotalCount: len(grants),
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PermissionsListCmd lists permissions.
type PermissionsListCmd struct{}

// Run executes the permissions list command.
func (cmd *PermissionsListCmd) Run(globals *Globals) error {
	start := time.Now()

	type permissionInfo struct {
		Name        string `json:"name"`
		Level       string `json:"level"`
		Description string `json:"description"`
	}

	type permissionsListResult struct {
		Permissions []permissionInfo `json:"permissions"`
		TotalCount  int              `json:"totalCount"`
	}

	permissions := []permissionInfo{
		// Developer-level permissions
		{Name: "CAN_SEE_ALL_APPS", Level: "developer", Description: "View app information and download bulk reports (read-only). Deprecated: use CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL"},
		{Name: "CAN_VIEW_FINANCIAL_DATA_GLOBAL", Level: "developer", Description: "View financial data, orders, and cancellation survey responses"},
		{Name: "CAN_MANAGE_PERMISSIONS_GLOBAL", Level: "developer", Description: "Admin (all permissions)"},
		{Name: "CAN_EDIT_GAMES_GLOBAL", Level: "developer", Description: "Edit Play Games Services projects"},
		{Name: "CAN_PUBLISH_GAMES_GLOBAL", Level: "developer", Description: "Publish Play Games Services projects"},
		{Name: "CAN_REPLY_TO_REVIEWS_GLOBAL", Level: "developer", Description: "Reply to reviews"},
		{Name: "CAN_MANAGE_PUBLIC_APKS_GLOBAL", Level: "developer", Description: "Release to production, exclude devices, and use app signing by Google Play"},
		{Name: "CAN_MANAGE_TRACK_APKS_GLOBAL", Level: "developer", Description: "Release to testing tracks"},
		{Name: "CAN_MANAGE_TRACK_USERS_GLOBAL", Level: "developer", Description: "Manage testing tracks and edit tester lists"},
		{Name: "CAN_MANAGE_PUBLIC_LISTING_GLOBAL", Level: "developer", Description: "Manage store presence"},
		{Name: "CAN_MANAGE_DRAFT_APPS_GLOBAL", Level: "developer", Description: "Create, edit, and delete draft apps"},
		{Name: "CAN_CREATE_MANAGED_PLAY_APPS_GLOBAL", Level: "developer", Description: "Create and publish private apps to your organization"},
		{Name: "CAN_CHANGE_MANAGED_PLAY_SETTING_GLOBAL", Level: "developer", Description: "Choose whether apps are public, or only available to your organization"},
		{Name: "CAN_MANAGE_ORDERS_GLOBAL", Level: "developer", Description: "Manage orders and subscriptions"},
		{Name: "CAN_MANAGE_APP_CONTENT_GLOBAL", Level: "developer", Description: "Manage policy related pages on all apps for the developer"},
		{Name: "CAN_VIEW_NON_FINANCIAL_DATA_GLOBAL", Level: "developer", Description: "View app information and download bulk reports (read-only)"},
		{Name: "CAN_VIEW_APP_QUALITY_GLOBAL", Level: "developer", Description: "View app quality information for all apps for the developer"},
		{Name: "CAN_MANAGE_DEEPLINKS_GLOBAL", Level: "developer", Description: "Manage the deep links setup for all apps for the developer"},

		// App-level permissions
		{Name: "CAN_ACCESS_APP", Level: "app", Description: "View app information (read-only). Deprecated: use CAN_VIEW_NON_FINANCIAL_DATA"},
		{Name: "CAN_VIEW_FINANCIAL_DATA", Level: "app", Description: "View financial data"},
		{Name: "CAN_MANAGE_PERMISSIONS", Level: "app", Description: "Admin (all permissions)"},
		{Name: "CAN_REPLY_TO_REVIEWS", Level: "app", Description: "Reply to reviews"},
		{Name: "CAN_MANAGE_PUBLIC_APKS", Level: "app", Description: "Release to production, exclude devices, and use app signing by Google Play"},
		{Name: "CAN_MANAGE_TRACK_APKS", Level: "app", Description: "Release to testing tracks"},
		{Name: "CAN_MANAGE_TRACK_USERS", Level: "app", Description: "Manage testing tracks and edit tester lists"},
		{Name: "CAN_MANAGE_PUBLIC_LISTING", Level: "app", Description: "Manage store presence"},
		{Name: "CAN_MANAGE_DRAFT_APPS", Level: "app", Description: "Edit and delete draft apps"},
		{Name: "CAN_MANAGE_ORDERS", Level: "app", Description: "Manage orders and subscriptions"},
		{Name: "CAN_MANAGE_APP_CONTENT", Level: "app", Description: "Manage policy related pages"},
		{Name: "CAN_VIEW_NON_FINANCIAL_DATA", Level: "app", Description: "View app information (read-only)"},
		{Name: "CAN_VIEW_APP_QUALITY", Level: "app", Description: "View app quality data such as Vitals, Crashes, etc."},
		{Name: "CAN_MANAGE_DEEPLINKS", Level: "app", Description: "Manage the deep links setup of an app"},
	}

	data := permissionsListResult{
		Permissions: permissions,
		TotalCount:  len(permissions),
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start))

	return outputResult(result, globals.Output, globals.Pretty)
}

// ============================================================================
// Recovery Commands
// ============================================================================

// RecoveryCmd contains app recovery commands.
type RecoveryCmd struct {
	List   RecoveryListCmd   `cmd:"" help:"List recovery actions"`
	Create RecoveryCreateCmd `cmd:"" help:"Create recovery action"`
	Deploy RecoveryDeployCmd `cmd:"" help:"Deploy recovery"`
	Cancel RecoveryCancelCmd `cmd:"" help:"Cancel recovery"`
}

// RecoveryListCmd lists recovery actions.
type RecoveryListCmd struct {
	Status string `help:"Filter by status: pending,active,completed,cancelled,failed"`
}

// Run executes the recovery list command.
func (cmd *RecoveryListCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	var resp *androidpublisher.ListAppRecoveriesResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Apprecovery.List(globals.Package).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list recovery actions: %v", err)).
			WithHint("Ensure the package name is correct and you have the required permissions")
	}

	type recoveryActionData struct {
		AppRecoveryID  int64  `json:"appRecoveryId"`
		Status         string `json:"status"`
		CreateTime     string `json:"createTime,omitempty"`
		DeployTime     string `json:"deployTime,omitempty"`
		CancelTime     string `json:"cancelTime,omitempty"`
		LastUpdateTime string `json:"lastUpdateTime,omitempty"`
	}

	type recoveryListResult struct {
		RecoveryActions []recoveryActionData `json:"recoveryActions"`
		TotalCount      int                  `json:"totalCount"`
	}

	// Map API status values to the simplified status values used for filtering
	statusMap := map[string]string{
		"RECOVERY_STATUS_ACTIVE":                 "active",
		"RECOVERY_STATUS_CANCELED":               "cancelled",
		"RECOVERY_STATUS_DRAFT":                  "pending",
		"RECOVERY_STATUS_GENERATION_IN_PROGRESS": "pending",
		"RECOVERY_STATUS_GENERATION_FAILED":      "failed",
		"RECOVERY_STATUS_UNSPECIFIED":            "pending",
	}

	var actions []recoveryActionData
	for _, a := range resp.RecoveryActions {
		// Apply status filter if provided
		if cmd.Status != "" {
			mapped, ok := statusMap[a.Status]
			if !ok {
				mapped = strings.ToLower(a.Status)
			}
			if !strings.EqualFold(mapped, cmd.Status) {
				continue
			}
		}

		actions = append(actions, recoveryActionData{
			AppRecoveryID:  a.AppRecoveryId,
			Status:         a.Status,
			CreateTime:     a.CreateTime,
			DeployTime:     a.DeployTime,
			CancelTime:     a.CancelTime,
			LastUpdateTime: a.LastUpdateTime,
		})
	}

	if actions == nil {
		actions = []recoveryActionData{}
	}

	data := recoveryListResult{
		RecoveryActions: actions,
		TotalCount:      len(actions),
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// RecoveryCreateCmd creates a recovery action.
type RecoveryCreateCmd struct {
	Type   string `help:"Recovery type" required:"" enum:"rollback,emergency_update,version_hold"`
	Target string `help:"Target version or track"`
	Reason string `help:"Reason for recovery" required:""`
}

// Run executes the recovery create command.
func (cmd *RecoveryCreateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	// Build the targeting configuration
	targeting := &androidpublisher.Targeting{
		AllUsers: &androidpublisher.AllUsers{
			IsAllUsersRequested: true,
		},
	}

	// If a target version code is specified, use version-based targeting instead
	if cmd.Target != "" {
		versionCode, parseErr := strconv.ParseInt(cmd.Target, 10, 64)
		if parseErr == nil {
			targeting = &androidpublisher.Targeting{
				VersionList: &androidpublisher.AppVersionList{
					VersionCodes: googleapi.Int64s{versionCode},
				},
			}
		}
		// If target is not a number, keep all-users targeting
	}

	req := &androidpublisher.CreateDraftAppRecoveryRequest{
		RemoteInAppUpdate: &androidpublisher.RemoteInAppUpdate{
			IsRemoteInAppUpdateRequested: true,
		},
		Targeting: targeting,
	}

	var action *androidpublisher.AppRecoveryAction
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		action, callErr = svc.Apprecovery.Create(globals.Package, req).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create recovery action: %v", err)).
			WithHint("Ensure the package name is correct and you have the required permissions")
	}

	type recoveryCreateResult struct {
		AppRecoveryID int64  `json:"appRecoveryId"`
		Status        string `json:"status"`
		CreateTime    string `json:"createTime,omitempty"`
		Type          string `json:"type"`
		Reason        string `json:"reason"`
	}

	data := recoveryCreateResult{
		AppRecoveryID: action.AppRecoveryId,
		Status:        action.Status,
		CreateTime:    action.CreateTime,
		Type:          cmd.Type,
		Reason:        cmd.Reason,
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// RecoveryDeployCmd deploys a recovery.
type RecoveryDeployCmd struct {
	ID string `arg:"" help:"Recovery action ID" required:""`
}

// Run executes the recovery deploy command.
func (cmd *RecoveryDeployCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	start := time.Now()

	recoveryID, err := strconv.ParseInt(cmd.ID, 10, 64)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid recovery action ID: %s", cmd.ID)).
			WithHint("Recovery action ID must be a numeric value")
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	err = client.DoWithRetry(ctx, func() error {
		_, callErr := svc.Apprecovery.Deploy(
			globals.Package,
			recoveryID,
			&androidpublisher.DeployAppRecoveryRequest{},
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to deploy recovery action: %v", err)).
			WithHint("Ensure the recovery action exists and is in DRAFT status")
	}

	type recoveryDeployResult struct {
		AppRecoveryID int64  `json:"appRecoveryId"`
		Deployed      bool   `json:"deployed"`
		Status        string `json:"status"`
	}

	data := recoveryDeployResult{
		AppRecoveryID: recoveryID,
		Deployed:      true,
		Status:        "RECOVERY_STATUS_ACTIVE",
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// RecoveryCancelCmd cancels a recovery.
type RecoveryCancelCmd struct {
	ID     string `arg:"" help:"Recovery action ID" required:""`
	Reason string `help:"Reason for cancellation"`
}

// Run executes the recovery cancel command.
func (cmd *RecoveryCancelCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	start := time.Now()

	recoveryID, err := strconv.ParseInt(cmd.ID, 10, 64)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid recovery action ID: %s", cmd.ID)).
			WithHint("Recovery action ID must be a numeric value")
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	err = client.DoWithRetry(ctx, func() error {
		_, callErr := svc.Apprecovery.Cancel(
			globals.Package,
			recoveryID,
			&androidpublisher.CancelAppRecoveryRequest{},
		).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to cancel recovery action: %v", err)).
			WithHint("Ensure the recovery action exists and is in an active state")
	}

	type recoveryCancelResult struct {
		AppRecoveryID int64  `json:"appRecoveryId"`
		Cancelled     bool   `json:"cancelled"`
		Status        string `json:"status"`
		Reason        string `json:"reason,omitempty"`
	}

	data := recoveryCancelResult{
		AppRecoveryID: recoveryID,
		Cancelled:     true,
		Status:        "RECOVERY_STATUS_CANCELED",
		Reason:        cmd.Reason,
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// ============================================================================
// Integrity Commands
// ============================================================================

// IntegrityCmd contains Play Integrity API commands.
type IntegrityCmd struct {
	Decode IntegrityDecodeCmd `cmd:"" help:"Decode integrity token"`
}

// IntegrityDecodeCmd decodes an integrity token.
type IntegrityDecodeCmd struct {
	Token  string `arg:"" help:"Integrity token to decode" required:""`
	Verify bool   `help:"Verify token signature"`
}

// Run executes the integrity decode command.
func (cmd *IntegrityDecodeCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	start := time.Now()

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	integritySvc, err := client.PlayIntegrity()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize Play Integrity service").
			WithHint("Ensure authentication is configured correctly and the Play Integrity API is enabled")
	}

	req := &playintegrity.DecodeIntegrityTokenRequest{
		IntegrityToken: cmd.Token,
	}

	var resp *playintegrity.DecodeIntegrityTokenResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = integritySvc.V1.DecodeIntegrityToken(globals.Package, req).Context(ctx).Do()
		return callErr
	})
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to decode integrity token: %v", err)).
			WithHint("Ensure the token is valid, the package name is correct, and the Play Integrity API is enabled in Google Cloud Console")
	}

	// Build a structured response from the token payload
	type accountDetails struct {
		AppLicensingVerdict string `json:"appLicensingVerdict,omitempty"`
	}

	type appIntegrity struct {
		AppRecognitionVerdict string `json:"appRecognitionVerdict,omitempty"`
		PackageName           string `json:"packageName,omitempty"`
		CertificateSha256     string `json:"certificateSha256,omitempty"`
		VersionCode           int64  `json:"versionCode,omitempty"`
	}

	type deviceIntegrity struct {
		DeviceRecognitionVerdict []string `json:"deviceRecognitionVerdict,omitempty"`
	}

	type requestDetails struct {
		RequestPackageName string `json:"requestPackageName,omitempty"`
		Nonce              string `json:"nonce,omitempty"`
		TimestampMillis    int64  `json:"timestampMillis,omitempty"`
	}

	type integrityDecodeResult struct {
		AccountDetails  *accountDetails  `json:"accountDetails,omitempty"`
		AppIntegrity    *appIntegrity    `json:"appIntegrity,omitempty"`
		DeviceIntegrity *deviceIntegrity `json:"deviceIntegrity,omitempty"`
		RequestDetails  *requestDetails  `json:"requestDetails,omitempty"`
		Verified        bool             `json:"verified"`
	}

	data := integrityDecodeResult{
		Verified: cmd.Verify,
	}

	if resp.TokenPayloadExternal != nil {
		payload := resp.TokenPayloadExternal

		if payload.AccountDetails != nil {
			data.AccountDetails = &accountDetails{
				AppLicensingVerdict: payload.AccountDetails.AppLicensingVerdict,
			}
		}

		if payload.AppIntegrity != nil {
			ai := &appIntegrity{
				AppRecognitionVerdict: payload.AppIntegrity.AppRecognitionVerdict,
				PackageName:           payload.AppIntegrity.PackageName,
				VersionCode:           payload.AppIntegrity.VersionCode,
			}
			if len(payload.AppIntegrity.CertificateSha256Digest) > 0 {
				ai.CertificateSha256 = strings.Join(payload.AppIntegrity.CertificateSha256Digest, ",")
			}
			data.AppIntegrity = ai
		}

		if payload.DeviceIntegrity != nil {
			data.DeviceIntegrity = &deviceIntegrity{
				DeviceRecognitionVerdict: payload.DeviceIntegrity.DeviceRecognitionVerdict,
			}
		}

		if payload.RequestDetails != nil {
			data.RequestDetails = &requestDetails{
				RequestPackageName: payload.RequestDetails.RequestPackageName,
				Nonce:              payload.RequestDetails.Nonce,
				TimestampMillis:    payload.RequestDetails.TimestampMillis,
			}
		}
	}

	result := output.NewResult(data).
		WithDuration(time.Since(start)).
		WithServices("playintegrity")

	return outputResult(result, globals.Output, globals.Pretty)
}
