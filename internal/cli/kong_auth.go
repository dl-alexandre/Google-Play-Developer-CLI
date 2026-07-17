package cli

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/auth"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/config"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/storage"
)

// AuthCmd contains authentication commands.
type AuthCmd struct {
	Status   AuthStatusCmd   `cmd:"" help:"Check authentication status"`
	Login    AuthLoginCmd    `cmd:"" help:"Authenticate with Google Play"`
	Init     AuthInitCmd     `cmd:"" help:"Initialize auth for a profile (alias of login)"`
	Logout   AuthLogoutCmd   `cmd:"" help:"Sign out and clear stored credentials for a profile"`
	Delete   AuthDeleteCmd   `cmd:"" help:"Delete a stored authentication profile"`
	List     AuthListCmd     `cmd:"" help:"List stored authentication profiles"`
	Switch   AuthSwitchCmd   `cmd:"" help:"Switch the active authentication profile"`
	Check    AuthCheckCmd    `cmd:"" help:"Validate package permissions for current credentials"`
	Doctor   AuthDoctorCmd   `cmd:"" help:"Diagnose authentication setup"`
	Diagnose AuthDiagnoseCmd `cmd:"" help:"Detailed auth diagnostics (alias of doctor)"`
}

// AuthStatusCmd checks authentication status.
type AuthStatusCmd struct{}

// Run executes the auth status command.
func (cmd *AuthStatusCmd) Run(globals *Globals) error {
	ctx := authContext(globals)
	authMgr := newAuthManager()

	// Try to authenticate with empty key path to load existing credentials
	// from storage, environment, or ADC
	_, _ = authMgr.Authenticate(ctx, globals.KeyPath)

	status, err := authMgr.GetStatus(ctx)
	if err != nil {
		return err
	}

	result := output.NewResult(status)
	return outputResult(result, globals.Output, globals.Pretty)
}

// AuthLoginCmd authenticates with Google Play.
type AuthLoginCmd struct {
	Profile string `arg:"" optional:"" help:"Profile name to store credentials under"`
	Key     string `help:"Path to service account key file" type:"existingfile"`
}

// Run executes the auth login command.
func (cmd *AuthLoginCmd) Run(globals *Globals) error {
	return runAuthLogin(globals, cmd.Profile, cmd.Key)
}

// AuthInitCmd is an ASC-parity alias for login with an optional profile.
type AuthInitCmd struct {
	Profile string `arg:"" optional:"" help:"Profile name to initialize"`
	Key     string `help:"Path to service account key file" type:"existingfile"`
}

// Run executes the auth init command.
func (cmd *AuthInitCmd) Run(globals *Globals) error {
	return runAuthLogin(globals, cmd.Profile, cmd.Key)
}

func runAuthLogin(globals *Globals, profileArg, keyPath string) error {
	ctx := authContext(globals)
	authMgr := newAuthManager()

	profile := strings.TrimSpace(profileArg)
	if profile == "" {
		profile = strings.TrimSpace(globals.Profile)
	}
	if profile == "" {
		profile = "default"
	}
	authMgr.SetActiveProfile(profile)

	if keyPath == "" {
		keyPath = globals.KeyPath
	}

	creds, err := authMgr.Authenticate(ctx, keyPath)
	if err != nil {
		return err
	}

	if err := config.SetActiveProfile(profile); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("authenticated but failed to persist active profile: %v", err)).
			WithHint("Credentials may still work for this process; check config write permissions")
	}
	globals.Profile = profile
	applyAuthGlobals(globals)

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"profile": profile,
		"origin":  creds.Origin.String(),
		"email":   creds.Email,
		"keyPath": creds.KeyPath,
	})
	return outputResult(result, globals.Output, globals.Pretty)
}

// AuthLogoutCmd signs out and clears stored credentials for a profile.
// Note: use --name (not --profile) so it does not collide with the global --profile flag.
type AuthLogoutCmd struct {
	Name string `name:"name" help:"Profile to sign out of (default: active profile)"`
	All  bool   `help:"Sign out of all profiles and clear all stored credentials"`
}

// Run executes the auth logout command.
func (cmd *AuthLogoutCmd) Run(globals *Globals) error {
	authMgr := newAuthManager()

	if cmd.All {
		if err := authMgr.ClearAllProfiles(); err != nil {
			return errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to clear all profiles: %v", err))
		}
		result := output.NewResult(map[string]interface{}{
			"success": true,
			"all":     true,
			"message": "Signed out of all profiles",
		})
		return outputResult(result, globals.Output, globals.Pretty)
	}

	profile := strings.TrimSpace(cmd.Name)
	if profile == "" {
		profile = authMgr.GetActiveProfile()
	}
	if profile == "" {
		profile = "default"
	}

	if err := authMgr.ClearProfile(profile); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to clear profile %q: %v", profile, err))
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"profile": profile,
		"message": "Signed out successfully",
	})
	return outputResult(result, globals.Output, globals.Pretty)
}

// AuthDeleteCmd removes a stored authentication profile.
type AuthDeleteCmd struct {
	Profile string `arg:"" help:"Profile name to delete" required:""`
	Force   bool   `help:"Allow deleting the currently active profile (switches to default)"`
}

// Run executes the auth delete command.
func (cmd *AuthDeleteCmd) Run(globals *Globals) error {
	profile := strings.TrimSpace(cmd.Profile)
	if profile == "" {
		return errors.NewAPIError(errors.CodeValidationError, "profile name is required")
	}

	authMgr := newAuthManager()
	active := resolveCLIActiveProfile(authMgr)

	if profile == active && !cmd.Force {
		return errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("cannot delete active profile %q without --force", profile)).
			WithHint("Switch to another profile first, or pass --force to delete and switch to default")
	}

	existed, err := authMgr.DeleteProfile(profile)
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to delete profile %q: %v", profile, err))
	}
	if !existed {
		return errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("profile %q not found", profile)).
			WithHint("Run `gpd auth list` to see stored profiles")
	}

	switchedTo := ""
	if profile == active {
		if err := config.SetActiveProfile("default"); err != nil {
			return errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("deleted profile but failed to switch to default: %v", err)).
				WithHint("Run `gpd auth switch default` manually")
		}
		globals.Profile = "default"
		applyAuthGlobals(globals)
		switchedTo = "default"
	}

	data := map[string]interface{}{
		"success": true,
		"profile": profile,
		"deleted": true,
		"message": fmt.Sprintf("Deleted profile %q", profile),
	}
	if switchedTo != "" {
		data["activeProfile"] = switchedTo
		data["message"] = fmt.Sprintf("Deleted active profile %q; switched to %q", profile, switchedTo)
	}

	result := output.NewResult(data)
	return outputResult(result, globals.Output, globals.Pretty)
}

// resolveCLIActiveProfile returns the effective active profile from manager + config.
func resolveCLIActiveProfile(authMgr *auth.Manager) string {
	active := "default"
	if authMgr != nil {
		active = authMgr.GetActiveProfile()
	}
	if cfg, loadErr := config.Load(); loadErr == nil && cfg != nil && cfg.ActiveProfile != "" {
		active = cfg.ActiveProfile
	}
	return active
}

// AuthListCmd lists stored authentication profiles.
type AuthListCmd struct{}

// Run executes the auth list command.
func (cmd *AuthListCmd) Run(globals *Globals) error {
	authMgr := newAuthManager()
	profiles, err := authMgr.ListProfiles()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to list profiles: %v", err))
	}

	active := authMgr.GetActiveProfile()
	if cfg, loadErr := config.Load(); loadErr == nil && cfg != nil && cfg.ActiveProfile != "" {
		active = cfg.ActiveProfile
	}

	type profileRow struct {
		Profile   string   `json:"profile"`
		Active    bool     `json:"active"`
		Origin    string   `json:"origin,omitempty"`
		Email     string   `json:"email,omitempty"`
		Scopes    []string `json:"scopes,omitempty"`
		UpdatedAt string   `json:"updatedAt,omitempty"`
		Expiry    string   `json:"tokenExpiry,omitempty"`
	}

	rows := make([]profileRow, 0, len(profiles))
	seen := map[string]bool{}
	for _, p := range profiles {
		rows = append(rows, profileRow{
			Profile:   p.Profile,
			Active:    p.Profile == active,
			Origin:    p.Origin,
			Email:     p.Email,
			Scopes:    p.Scopes,
			UpdatedAt: p.UpdatedAt,
			Expiry:    p.TokenExpiry,
		})
		seen[p.Profile] = true
	}
	// Always surface the active profile even when no token metadata exists yet.
	if !seen[active] {
		rows = append([]profileRow{{
			Profile: active,
			Active:  true,
		}}, rows...)
	}

	result := output.NewResult(map[string]interface{}{
		"activeProfile": active,
		"profiles":      rows,
		"count":         len(rows),
	})
	return outputResult(result, globals.Output, globals.Pretty)
}

// AuthSwitchCmd switches the active authentication profile.
type AuthSwitchCmd struct {
	Profile string `arg:"" help:"Profile name to activate" required:""`
}

// Run executes the auth switch command.
func (cmd *AuthSwitchCmd) Run(globals *Globals) error {
	profile := strings.TrimSpace(cmd.Profile)
	if profile == "" {
		return errors.NewAPIError(errors.CodeValidationError, "profile name is required")
	}

	authMgr := newAuthManager()
	profiles, err := authMgr.ListProfiles()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to list profiles: %v", err))
	}

	found := profile == "default"
	var meta *auth.TokenMetadata
	for i := range profiles {
		if profiles[i].Profile == profile {
			found = true
			meta = &profiles[i]
			break
		}
	}
	if !found {
		return errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("profile %q not found", profile)).
			WithHint("Run `gpd auth list` to see stored profiles, or `gpd auth login <profile> --key ...` to create one")
	}

	if err := config.SetActiveProfile(profile); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to set active profile: %v", err))
	}
	globals.Profile = profile
	applyAuthGlobals(globals)

	data := map[string]interface{}{
		"success": true,
		"profile": profile,
		"active":  true,
	}
	if meta != nil {
		data["origin"] = meta.Origin
		data["email"] = meta.Email
		data["updatedAt"] = meta.UpdatedAt
	}

	result := output.NewResult(data)
	return outputResult(result, globals.Output, globals.Pretty)
}

// AuthCheckCmd validates that credentials can access a package.
// Use the global --package flag.
type AuthCheckCmd struct{}

// Run executes the auth check command.
func (cmd *AuthCheckCmd) Run(globals *Globals) error {
	ctx := authContext(globals)
	pkg := strings.TrimSpace(globals.Package)
	if err := requirePackage(pkg); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, err.Error()).
			WithHint("Pass --package com.example.app")
	}

	authMgr := newAuthManager()
	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource,
		api.WithTimeout(globals.Timeout),
		api.WithVerboseLogging(globals.Verbose))
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return err
	}

	permissions := []*auth.PermissionCheck{}
	valid := true

	// Probe: create a disposable edit, then delete it.
	editCheck := &auth.PermissionCheck{
		Surface:  "androidpublisher.edits",
		TestCall: "edits.insert+edits.delete",
	}
	edit, editErr := svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
	if editErr != nil {
		valid = false
		editCheck.HasAccess = false
		editCheck.Error = editErr.Error()
	} else {
		editCheck.HasAccess = true
		if delErr := svc.Edits.Delete(pkg, edit.Id).Context(ctx).Do(); delErr != nil {
			// Access worked; cleanup failure is non-fatal but reported.
			editCheck.Error = fmt.Sprintf("cleanup failed: %v", delErr)
		}
	}
	permissions = append(permissions, editCheck)

	check := &auth.CheckResult{
		Valid:       valid,
		Origin:      creds.Origin.String(),
		Email:       creds.Email,
		Permissions: permissions,
	}

	result := output.NewResult(map[string]interface{}{
		"package":     pkg,
		"profile":     authMgr.GetActiveProfile(),
		"valid":       check.Valid,
		"origin":      check.Origin,
		"email":       check.Email,
		"permissions": check.Permissions,
	})

	if !valid {
		// Still emit structured output, then return a typed error for exit code.
		_ = outputResult(result, globals.Output, globals.Pretty)
		return errors.NewAPIError(errors.CodePermissionDenied,
			fmt.Sprintf("credentials cannot access package %s", pkg)).
			WithHint("Grant the service account Android Publisher access for this app in Play Console")
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// AuthDoctorCmd diagnoses authentication setup.
type AuthDoctorCmd struct {
	RefreshCheck bool `help:"Attempt token refresh / credential load" name:"refresh-check"`
	Network      bool `help:"Run a lightweight network permission probe (requires --package)"`
}

// Run executes the auth doctor command.
func (cmd *AuthDoctorCmd) Run(globals *Globals) error {
	return runAuthDoctor(globals, cmd.RefreshCheck, cmd.Network)
}

// AuthDiagnoseCmd is an alias of doctor with the same flags.
type AuthDiagnoseCmd struct {
	RefreshCheck bool `help:"Attempt token refresh / credential load" name:"refresh-check"`
	Network      bool `help:"Run a lightweight network permission probe (requires --package)"`
}

// Run executes the auth diagnose command.
func (cmd *AuthDiagnoseCmd) Run(globals *Globals) error {
	return runAuthDoctor(globals, cmd.RefreshCheck, cmd.Network)
}

type doctorCheck struct {
	Status         string `json:"status"` // ok, warn, fail, info
	Message        string `json:"message"`
	Recommendation string `json:"recommendation,omitempty"`
}

type doctorSection struct {
	Title  string        `json:"title"`
	Checks []doctorCheck `json:"checks"`
}

type doctorSummary struct {
	OK       int `json:"ok"`
	Info     int `json:"info"`
	Warnings int `json:"warnings"`
	Errors   int `json:"errors"`
}

func runAuthDoctor(globals *Globals, refreshCheck, network bool) error {
	ctx := authContext(globals)
	sections := []doctorSection{}
	recommendations := []string{}

	// Section: Environment
	envChecks := []doctorCheck{}
	if key := config.GetEnvServiceAccountKey(); key != "" {
		envChecks = append(envChecks, doctorCheck{Status: "ok", Message: "GPD_SERVICE_ACCOUNT_KEY is set"})
	} else {
		envChecks = append(envChecks, doctorCheck{Status: "info", Message: "GPD_SERVICE_ACCOUNT_KEY is not set"})
	}
	if gac := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); gac != "" {
		if _, err := os.Stat(gac); err == nil {
			envChecks = append(envChecks, doctorCheck{Status: "ok", Message: "GOOGLE_APPLICATION_CREDENTIALS points to an existing file"})
		} else {
			envChecks = append(envChecks, doctorCheck{
				Status:         "fail",
				Message:        "GOOGLE_APPLICATION_CREDENTIALS is set but file is missing",
				Recommendation: "Fix the path or unset GOOGLE_APPLICATION_CREDENTIALS",
			})
			recommendations = append(recommendations, "Fix or unset GOOGLE_APPLICATION_CREDENTIALS")
		}
	} else {
		envChecks = append(envChecks, doctorCheck{Status: "info", Message: "GOOGLE_APPLICATION_CREDENTIALS is not set"})
	}
	if p := config.GetEnvAuthProfile(); p != "" {
		envChecks = append(envChecks, doctorCheck{Status: "info", Message: fmt.Sprintf("GPD_AUTH_PROFILE=%s", p)})
	}
	sections = append(sections, doctorSection{Title: "Environment", Checks: envChecks})

	// Section: Storage
	store := storage.New()
	storeChecks := []doctorCheck{
		{
			Status:  boolStatus(store.Available(), "ok", "warn"),
			Message: fmt.Sprintf("Secure storage available=%v platform=%s", store.Available(), storage.Platform()),
		},
	}
	if !store.Available() {
		storeChecks[0].Recommendation = "Tokens will stay in memory only; set --store-tokens=never in CI or use service account --key"
		recommendations = append(recommendations, "Secure storage unavailable; prefer --key service-account JSON in CI")
	}
	paths := config.GetPaths()
	if _, err := os.Stat(paths.ConfigDir); err == nil {
		storeChecks = append(storeChecks, doctorCheck{Status: "ok", Message: "Config directory exists: " + paths.ConfigDir})
	} else {
		storeChecks = append(storeChecks, doctorCheck{
			Status:         "warn",
			Message:        "Config directory missing: " + paths.ConfigDir,
			Recommendation: "Run `gpd config init` or `gpd auth login --key ...`",
		})
	}
	sections = append(sections, doctorSection{Title: "Storage", Checks: storeChecks})

	// Section: Profiles
	authMgr := newAuthManager()
	active := authMgr.GetActiveProfile()
	if cfg, err := config.Load(); err == nil && cfg != nil && cfg.ActiveProfile != "" {
		active = cfg.ActiveProfile
	}
	profiles, listErr := authMgr.ListProfiles()
	profileChecks := []doctorCheck{
		{Status: "info", Message: fmt.Sprintf("Active profile: %s", active)},
	}
	if listErr != nil {
		profileChecks = append(profileChecks, doctorCheck{
			Status:  "fail",
			Message: fmt.Sprintf("Failed to list profiles: %v", listErr),
		})
	} else {
		profileChecks = append(profileChecks, doctorCheck{
			Status:  "ok",
			Message: fmt.Sprintf("Stored profiles: %d", len(profiles)),
		})
		for _, p := range profiles {
			profileChecks = append(profileChecks, doctorCheck{
				Status:  "info",
				Message: fmt.Sprintf("profile=%s origin=%s email=%s", p.Profile, p.Origin, p.Email),
			})
		}
	}
	sections = append(sections, doctorSection{Title: "Profiles", Checks: profileChecks})

	// Section: Credentials (optional refresh)
	credChecks := []doctorCheck{}
	if refreshCheck || globals.KeyPath != "" {
		creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
		if err != nil {
			credChecks = append(credChecks, doctorCheck{
				Status:         "fail",
				Message:        fmt.Sprintf("Credential load failed: %v", err),
				Recommendation: "Run `gpd auth login --key /path/to/service-account.json`",
			})
			recommendations = append(recommendations, "Re-authenticate with a valid service account key or ADC")
		} else {
			credChecks = append(credChecks, doctorCheck{
				Status:  "ok",
				Message: fmt.Sprintf("Credentials loaded origin=%s email=%s", creds.Origin.String(), creds.Email),
			})
			status, stErr := authMgr.GetStatus(ctx)
			if stErr != nil {
				credChecks = append(credChecks, doctorCheck{Status: "warn", Message: stErr.Error()})
			} else if status != nil {
				if status.TokenValid {
					credChecks = append(credChecks, doctorCheck{Status: "ok", Message: "Token is valid"})
				} else {
					credChecks = append(credChecks, doctorCheck{
						Status:         "fail",
						Message:        "Token is not valid",
						Recommendation: "Re-run `gpd auth login` or check clock skew",
					})
					recommendations = append(recommendations, "Re-authenticate; token is invalid or expired")
				}
			}
		}
	} else {
		credChecks = append(credChecks, doctorCheck{
			Status:         "info",
			Message:        "Skipped credential load (pass --refresh-check or --key-path)",
			Recommendation: "Use --refresh-check to validate token acquisition",
		})
	}
	sections = append(sections, doctorSection{Title: "Credentials", Checks: credChecks})

	// Section: Network probe
	if network {
		netChecks := []doctorCheck{}
		pkg := strings.TrimSpace(globals.Package)
		if pkg == "" {
			netChecks = append(netChecks, doctorCheck{
				Status:         "fail",
				Message:        "Network probe requested but --package is empty",
				Recommendation: "Pass --package com.example.app with --network",
			})
			recommendations = append(recommendations, "Provide --package for network permission probe")
		} else {
			if err := probePackageAccess(ctx, globals, pkg); err != nil {
				netChecks = append(netChecks, doctorCheck{
					Status:         "fail",
					Message:        fmt.Sprintf("Package access failed for %s: %v", pkg, err),
					Recommendation: "Grant Play Console access to the service account for this package",
				})
				recommendations = append(recommendations, "Fix Play Console user access for the service account")
			} else {
				netChecks = append(netChecks, doctorCheck{
					Status:  "ok",
					Message: fmt.Sprintf("Package access OK for %s", pkg),
				})
			}
		}
		sections = append(sections, doctorSection{Title: "Network", Checks: netChecks})
	}

	// Section: Runtime
	sections = append(sections, doctorSection{
		Title: "Runtime",
		Checks: []doctorCheck{
			{Status: "info", Message: fmt.Sprintf("goos=%s goarch=%s", runtime.GOOS, runtime.GOARCH)},
			{Status: "info", Message: fmt.Sprintf("timeout=%s storeTokens=%s", globals.Timeout, globals.StoreTokens)},
		},
	})

	summary := doctorSummary{}
	for _, sec := range sections {
		for _, c := range sec.Checks {
			switch c.Status {
			case "ok":
				summary.OK++
			case "info":
				summary.Info++
			case "warn":
				summary.Warnings++
			case "fail":
				summary.Errors++
			}
		}
	}

	report := map[string]interface{}{
		"sections":        sections,
		"summary":         summary,
		"recommendations": recommendations,
		"activeProfile":   active,
		"generatedAt":     time.Now().UTC().Format(time.RFC3339),
	}

	result := output.NewResult(report)
	if err := outputResult(result, globals.Output, globals.Pretty); err != nil {
		return err
	}
	if summary.Errors > 0 {
		return errors.NewAPIError(errors.CodeAuthFailure, "auth doctor found failures").
			WithHint("Review recommendations in the doctor report")
	}
	return nil
}

func probePackageAccess(ctx context.Context, globals *Globals, pkg string) error {
	authMgr := newAuthManager()
	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}
	client, err := api.NewClient(ctx, creds.TokenSource,
		api.WithTimeout(globals.Timeout),
		api.WithVerboseLogging(globals.Verbose))
	if err != nil {
		return err
	}
	svc, err := client.AndroidPublisher()
	if err != nil {
		return err
	}
	edit, err := svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return err
	}
	_ = svc.Edits.Delete(pkg, edit.Id).Context(ctx).Do()
	return nil
}

func authContext(globals *Globals) context.Context {
	if globals != nil && globals.Context != nil {
		return globals.Context
	}
	return context.Background()
}

func boolStatus(ok bool, whenTrue, whenFalse string) string {
	if ok {
		return whenTrue
	}
	return whenFalse
}
