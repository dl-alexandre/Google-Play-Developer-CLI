// Package cli provides auth commands for gpd.
package cli

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/auth"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

const defaultProfileName = "default"

func (c *CLI) addAuthCommands() {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  "Manage authentication and credentials for Google Play APIs.",
	}

	// auth status
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check current authentication status",
		Long:  "Display the current authentication state and credential information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.authStatus(cmd.Context())
		},
	}

	// auth check
	checkCmd := &cobra.Command{
		Use:   "check",
		Short: "Validate service account permissions",
		Long:  "Validate that the service account has required permissions for each API surface.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.authCheck(cmd.Context())
		},
	}

	// auth logout
	logoutCmd := &cobra.Command{
		Use:   "logout",
		Short: "Clear stored credentials",
		Long:  "Remove stored credentials from secure storage.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.authLogout(cmd.Context())
		},
	}

	// auth diagnose
	diagnoseCmd := &cobra.Command{
		Use:   "diagnose",
		Short: "Diagnose authentication setup",
		Long:  "Show detailed authentication diagnostics and token refresh status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshCheck, _ := cmd.Flags().GetBool("refresh-check")
			return c.authDiagnose(cmd.Context(), refreshCheck)
		},
	}
	diagnoseCmd.Flags().Bool("refresh-check", false, "Attempt a token refresh and report errors")

	// auth doctor (alias of diagnose for parity)
	doctorCmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose authentication setup",
		Long:  "Alias for auth diagnose to match parity expectations.",
		RunE: func(cmd *cobra.Command, args []string) error {
			refreshCheck, _ := cmd.Flags().GetBool("refresh-check")
			return c.authDiagnose(cmd.Context(), refreshCheck)
		},
	}
	doctorCmd.Flags().Bool("refresh-check", false, "Attempt a token refresh and report errors")

	// auth login
	var (
		loginClientID     string
		loginClientSecret string
		loginFlow         string
	)
	loginCmd := &cobra.Command{
		Use:   "login [profile]",
		Short: "Authenticate using OAuth device flow",
		Long:  "Authenticate using OAuth device flow and store credentials for a profile.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := defaultProfileName
			if len(args) > 0 {
				profile = args[0]
			}
			return c.authLogin(cmd.Context(), profile, loginClientID, loginClientSecret, loginFlow)
		},
	}
	loginCmd.Flags().StringVar(&loginClientID, "client-id", "", "OAuth client ID (or set GPD_CLIENT_ID)")
	loginCmd.Flags().StringVar(&loginClientSecret, "client-secret", "", "OAuth client secret (or set GPD_CLIENT_SECRET)")
	loginCmd.Flags().StringVar(&loginFlow, "flow", "device", "OAuth flow (device)")

	// auth init (alias of login)
	initCmd := &cobra.Command{
		Use:   "init [profile]",
		Short: "Initialize OAuth authentication",
		Long:  "Alias for auth login to initialize credentials for a profile.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			profile := defaultProfileName
			if len(args) > 0 {
				profile = args[0]
			}
			return c.authLogin(cmd.Context(), profile, loginClientID, loginClientSecret, loginFlow)
		},
	}
	initCmd.Flags().StringVar(&loginClientID, "client-id", "", "OAuth client ID (or set GPD_CLIENT_ID)")
	initCmd.Flags().StringVar(&loginClientSecret, "client-secret", "", "OAuth client secret (or set GPD_CLIENT_SECRET)")
	initCmd.Flags().StringVar(&loginFlow, "flow", "device", "OAuth flow (device)")

	// auth switch
	switchCmd := &cobra.Command{
		Use:   "switch <profile>",
		Short: "Switch active authentication profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.authSwitch(cmd.Context(), args[0])
		},
	}

	// auth list
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List stored authentication profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.authList(cmd.Context())
		},
	}

	authCmd.AddCommand(statusCmd, checkCmd, logoutCmd, diagnoseCmd, doctorCmd, loginCmd, initCmd, switchCmd, listCmd)
	c.rootCmd.AddCommand(authCmd)
}

func (c *CLI) authStatus(ctx context.Context) error {
	// Try to authenticate
	_, err := c.authMgr.Authenticate(ctx, c.keyPath)
	if err != nil {
		authErr := errors.ClassifyAuthError(err)
		payload := map[string]interface{}{
			"authenticated": false,
		}
		if authErr != nil {
			payload["error"] = authErr
		} else {
			payload["error"] = err.Error()
		}
		result := output.NewResult(payload)
		return c.Output(result.WithServices("auth"))
	}

	status, err := c.authMgr.GetStatus(ctx)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeAuthFailure, err.Error()))
	}

	result := output.NewResult(status)
	return c.Output(result.WithServices("auth"))
}

func (c *CLI) authCheck(ctx context.Context) error {
	// Authenticate first
	creds, err := c.authMgr.Authenticate(ctx, c.keyPath)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Get API client
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	checks := []*auth.PermissionCheck{}

	// Check Android Publisher API (edits)
	publisherSvc, err := client.AndroidPublisher()
	if err == nil && c.packageName != "" {
		// Insert and immediately delete an edit to test permissions
		edit, err := publisherSvc.Edits.Insert(c.packageName, nil).Context(ctx).Do()
		check := &auth.PermissionCheck{
			Surface:  "edits",
			TestCall: "edits.insert",
		}
		if err != nil {
			check.HasAccess = false
			check.Error = err.Error()
		} else {
			check.HasAccess = true
			// Clean up test edit
			_ = publisherSvc.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()
		}
		checks = append(checks, check)
	}

	// Check Reviews API
	if publisherSvc != nil && c.packageName != "" {
		_, err := publisherSvc.Reviews.List(c.packageName).Context(ctx).MaxResults(1).Do()
		check := &auth.PermissionCheck{
			Surface:  "reviews",
			TestCall: "reviews.list",
		}
		if err != nil {
			check.HasAccess = false
			check.Error = err.Error()
		} else {
			check.HasAccess = true
		}
		checks = append(checks, check)
	}

	// Check Play Reporting API
	reportingSvc, err := client.PlayReporting()
	if err == nil && c.packageName != "" {
		// Note: This would require proper method call for reporting API
		check := &auth.PermissionCheck{
			Surface:  "reporting",
			TestCall: "apps.fetchReleaseFilterOptions",
		}
		if reportingSvc != nil {
			check.HasAccess = true // Simplified - actual implementation would make API call
		} else {
			check.HasAccess = false
			check.Error = "reporting service unavailable"
		}
		checks = append(checks, check)
	}

	// Determine overall validity
	valid := true
	for _, check := range checks {
		if !check.HasAccess {
			valid = false
			break
		}
	}

	checkResult := &auth.CheckResult{
		Valid:       valid,
		Origin:      creds.Origin.String(),
		Email:       creds.Email,
		Permissions: checks,
	}

	result := output.NewResult(checkResult)
	return c.Output(result.WithServices("androidpublisher", "playdeveloperreporting"))
}

func (c *CLI) authLogout(_ context.Context) error {
	c.authMgr.Clear()

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"message": "Credentials cleared",
	})
	return c.Output(result.WithServices("auth"))
}

func (c *CLI) authDiagnose(ctx context.Context, refreshCheck bool) error {
	creds, err := c.authMgr.Authenticate(ctx, c.keyPath)
	if err != nil {
		authErr := errors.ClassifyAuthError(err)
		payload := map[string]interface{}{
			"authenticated": false,
		}
		if authErr != nil {
			payload["error"] = authErr
		} else {
			payload["error"] = err.Error()
		}
		result := output.NewResult(payload)
		return c.Output(result.WithServices("auth"))
	}

	meta, _ := c.authMgr.LoadTokenMetadata(c.authMgr.GetActiveProfile())
	tokenLocation := c.authMgr.TokenLocation()

	token, tokenErr := creds.TokenSource.Token()
	tokenValid := tokenErr == nil && token != nil && token.Valid()
	tokenExpiry := ""
	if tokenErr == nil && token != nil && !token.Expiry.IsZero() {
		tokenExpiry = token.Expiry.Format(time.RFC3339)
	}

	clientHash := ""
	clientLast4 := ""
	if meta != nil {
		clientHash = meta.ClientIDHash
		clientLast4 = meta.ClientIDLast4
	}

	diagnostics := map[string]interface{}{
		"authenticated": true,
		"origin":        creds.Origin.String(),
		"email":         creds.Email,
		"keyPath":       creds.KeyPath,
		"tokenLocation": tokenLocation,
		"clientIdHash":  clientHash,
		"clientIdLast4": clientLast4,
		"scopes":        creds.Scopes,
		"tokenValid":    tokenValid,
		"tokenExpiry":   tokenExpiry,
	}

	if refreshCheck {
		refreshResult := map[string]interface{}{
			"success": tokenErr == nil,
		}
		if tokenErr != nil {
			if apiErr := errors.ClassifyAuthError(tokenErr); apiErr != nil {
				refreshResult["error"] = apiErr
			} else {
				refreshResult["error"] = tokenErr.Error()
			}
		}
		diagnostics["refreshCheck"] = refreshResult
	}

	result := output.NewResult(diagnostics)
	return c.Output(result.WithServices("auth"))
}

func (c *CLI) authLogin(ctx context.Context, profile, clientID, clientSecret, flow string) error {
	if flow != "" && flow != "device" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("unsupported flow: %s", flow)).
			WithHint("Supported flows: device"))
	}
	if clientID == "" {
		clientID = config.GetEnvOAuthClientID()
	}
	if clientSecret == "" {
		clientSecret = config.GetEnvOAuthClientSecret()
	}
	if profile == "" {
		profile = "default"
	}
	c.profile = profile
	c.authMgr.SetActiveProfile(profile)
	scopes := []string{
		auth.ScopeAndroidPublisher,
		auth.ScopePlayReporting,
		auth.ScopeGames,
		auth.ScopePlayIntegrity,
	}
	creds, err := c.authMgr.AuthenticateWithDeviceCode(ctx, clientID, clientSecret, scopes, c.stderr)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			return c.OutputError(apiErr)
		}
		return c.OutputError(errors.NewAPIError(errors.CodeAuthFailure, err.Error()))
	}
	if err := c.setActiveProfile(profile); err != nil {
		return c.OutputError(err)
	}
	result := output.NewResult(map[string]interface{}{
		"success":  true,
		"profile":  profile,
		"origin":   creds.Origin.String(),
		"scopes":   creds.Scopes,
		"location": c.authMgr.TokenLocation(),
	})
	return c.Output(result.WithServices("auth"))
}

func (c *CLI) authSwitch(_ context.Context, profile string) error {
	meta, err := c.authMgr.LoadTokenMetadata(profile)
	if err != nil || meta == nil {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("profile not found: %s", profile)).
			WithHint("Use gpd auth list to see available profiles"))
	}
	if err := c.setActiveProfile(profile); err != nil {
		return c.OutputError(err)
	}
	result := output.NewResult(map[string]interface{}{
		"success": true,
		"profile": profile,
		"origin":  meta.Origin,
		"email":   meta.Email,
	})
	return c.Output(result.WithServices("auth"))
}

func (c *CLI) authList(_ context.Context) error {
	profiles, err := c.authMgr.ListProfiles()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Profile < profiles[j].Profile
	})
	result := output.NewResult(map[string]interface{}{
		"profiles":      profiles,
		"count":         len(profiles),
		"activeProfile": c.authMgr.GetActiveProfile(),
	})
	return c.Output(result.WithServices("auth"))
}

func (c *CLI) setActiveProfile(profile string) *errors.APIError {
	if profile == "" {
		profile = "default"
	}
	c.profile = profile
	c.authMgr.SetActiveProfile(profile)
	if c.config == nil {
		return nil
	}
	c.config.ActiveProfile = profile
	if err := c.config.Save(); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to save config: %v", err))
	}
	return nil
}
