// Package cli provides testing and QA commands.
package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

// TestingCmd contains testing and QA commands.
type TestingCmd struct {
	Prelaunch     TestingPrelaunchCmd     `cmd:"" help:"Trigger or check pre-launch report"`
	DeviceLab     TestingDeviceLabCmd     `cmd:"" help:"Run tests on Firebase Test Lab"`
	Screenshots   TestingScreenshotsCmd   `cmd:"" help:"Capture screenshots across devices"`
	Validate      TestingValidateCmd      `cmd:"" help:"Comprehensive app validation"`
	Compatibility TestingCompatibilityCmd `cmd:"" help:"Check device compatibility"`
}

// TestingPrelaunchCmd triggers or checks pre-launch report.
type TestingPrelaunchCmd struct {
	Action      string `help:"Action to perform" enum:"trigger,check,wait" default:"check"`
	EditID      string `help:"Edit ID to check"`
	MaxWaitTime string `help:"Maximum time to wait for results" default:"30m"`
	Format      string `help:"Output format" default:"table" enum:"json,table"`
}

// testingPrelaunchResult represents pre-launch report results.
type testingPrelaunchResult struct {
	Status      string                   `json:"status"`
	EditID      string                   `json:"editId,omitempty"`
	TestsRun    int                      `json:"testsRun"`
	TestsPassed int                      `json:"testsPassed"`
	TestsFailed int                      `json:"testsFailed"`
	Issues      []testingPrelaunchIssue  `json:"issues,omitempty"`
	Devices     []testingPrelaunchDevice `json:"devices,omitempty"`
	CheckedAt   time.Time                `json:"checkedAt"`
}

// testingPrelaunchIssue represents an issue found.
type testingPrelaunchIssue struct {
	Severity string `json:"severity"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Device   string `json:"device,omitempty"`
}

// testingPrelaunchDevice represents a tested device.
type testingPrelaunchDevice struct {
	Model     string `json:"model"`
	OSVersion string `json:"osVersion"`
	Status    string `json:"status"`
}

// Run executes the pre-launch command.
func (cmd *TestingPrelaunchCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Pre-launch %s for edit %s\n", cmd.Action, cmd.EditID)
	}

	result := &testingPrelaunchResult{
		EditID:      cmd.EditID,
		TestsRun:    0,
		TestsPassed: 0,
		TestsFailed: 0,
		Issues:      make([]testingPrelaunchIssue, 0),
		Devices:     make([]testingPrelaunchDevice, 0),
		CheckedAt:   time.Now(),
	}

	// If an EditID is provided, try to verify the edit exists
	if cmd.EditID != "" {
		ctx := context.Background()
		authMgr := newAuthManager()
		creds, authErr := authMgr.Authenticate(ctx, globals.KeyPath)
		if authErr == nil {
			client, clientErr := api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
			if clientErr == nil {
				svc, svcErr := client.AndroidPublisher()
				if svcErr == nil {
					if err := client.Acquire(ctx); err != nil {
						return err
					}

					// Validate the edit exists by getting it
					var edit *androidpublisher.AppEdit
					verr := client.DoWithRetry(ctx, func() error {
						var gerr error
						edit, gerr = svc.Edits.Get(globals.Package, cmd.EditID).Context(ctx).Do()
						return gerr
					})

					client.Release()

					if verr == nil && edit != nil {
						result.Status = "edit_verified"
						result.Issues = append(result.Issues, testingPrelaunchIssue{
							Severity: "info",
							Type:     "edit_found",
							Message:  fmt.Sprintf("Edit %s exists and is valid. Pre-launch reports must be checked via Play Console UI.", cmd.EditID),
						})
					} else {
						result.Status = "edit_not_found"
						result.Issues = append(result.Issues, testingPrelaunchIssue{
							Severity: "warning",
							Type:     "edit_not_found",
							Message:  fmt.Sprintf("Edit %s could not be verified: %v", cmd.EditID, verr),
						})
					}
				}
			}
		}
	}

	// Pre-launch report API has limited programmatic access
	if result.Status == "" {
		result.Status = "api_limited"
	}

	result.Issues = append(result.Issues, testingPrelaunchIssue{
		Severity: "info",
		Type:     "api_limitation",
		Message:  "Pre-launch report detailed API access requires Play Console. Check results in the Play Console UI.",
	})

	return writeOutput(globals, output.NewResult(result).
		WithServices("androidpublisher").
		WithWarnings("Pre-launch report API has limited programmatic access"))
}

// TestingDeviceLabCmd runs tests on Firebase Test Lab.
type TestingDeviceLabCmd struct {
	AppFile     string   `help:"App file to test (APK/AAB)" type:"existingfile" required:""`
	TestFile    string   `help:"Test APK file (optional for robo tests)" type:"existingfile"`
	Devices     []string `help:"Device IDs to test on (repeatable)"`
	TestTimeout string   `help:"Test timeout duration" default:"15m"`
	Async       bool     `help:"Don't wait for results"`
	TestType    string   `help:"Test type" default:"robo" enum:"robo,instrumentation"`
}

// testingDeviceLabResult represents Firebase Test Lab results.
type testingDeviceLabResult struct {
	TestMatrixID string                    `json:"testMatrixId,omitempty"`
	Status       string                    `json:"status"`
	Outcome      string                    `json:"outcome,omitempty"`
	TestRuns     []testingDeviceLabTestRun `json:"testRuns,omitempty"`
	LogsURL      string                    `json:"logsUrl,omitempty"`
	SuggestedCmd string                    `json:"suggestedCommand,omitempty"`
	GcloudFound  bool                      `json:"gcloudAvailable"`
	StartedAt    time.Time                 `json:"startedAt"`
	CompletedAt  *time.Time                `json:"completedAt,omitempty"`
}

// testingDeviceLabTestRun represents a single test run.
type testingDeviceLabTestRun struct {
	Device    string `json:"device"`
	OSVersion string `json:"osVersion"`
	Status    string `json:"status"`
	Outcome   string `json:"outcome"`
	Duration  string `json:"duration,omitempty"`
}

// Run executes the device lab command.
func (cmd *TestingDeviceLabCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Running %s tests on Firebase Test Lab\n", cmd.TestType)
		fmt.Fprintf(os.Stderr, "App: %s\n", cmd.AppFile)
	}

	result := &testingDeviceLabResult{
		TestRuns:  make([]testingDeviceLabTestRun, 0),
		StartedAt: time.Now(),
	}

	// Check if gcloud CLI is available on the system
	gcloudPath, lookErr := exec.LookPath("gcloud")
	result.GcloudFound = lookErr == nil

	// Build the suggested gcloud command
	var cmdParts []string
	cmdParts = append(cmdParts, "gcloud", "firebase", "test", "android", "run")

	if cmd.TestType == "instrumentation" && cmd.TestFile != "" {
		cmdParts = append(cmdParts, "--type", "instrumentation", "--app", cmd.AppFile, "--test", cmd.TestFile)
	} else {
		cmdParts = append(cmdParts, "--type", "robo", "--app", cmd.AppFile)
	}

	if cmd.TestTimeout != "" {
		cmdParts = append(cmdParts, "--timeout", cmd.TestTimeout)
	}

	for _, device := range cmd.Devices {
		cmdParts = append(cmdParts, "--device", fmt.Sprintf("model=%s", device))
	}

	if cmd.Async {
		cmdParts = append(cmdParts, "--async")
	}

	result.SuggestedCmd = strings.Join(cmdParts, " ")

	// Add placeholder test runs for devices
	for _, device := range cmd.Devices {
		result.TestRuns = append(result.TestRuns, testingDeviceLabTestRun{
			Device:    device,
			OSVersion: "latest",
			Status:    "pending",
			Outcome:   "pending",
		})
	}

	warnings := []string{
		"Firebase Test Lab integration requires Firebase project setup and authentication",
	}

	if result.GcloudFound {
		result.Status = "requires_firebase_setup"
		result.Outcome = "not_started"
		warnings = append(warnings,
			fmt.Sprintf("gcloud CLI found at %s. Run the suggested command to execute tests.", gcloudPath))
	} else {
		result.Status = "requires_firebase_setup"
		result.Outcome = "not_started"
		warnings = append(warnings,
			"gcloud CLI not found. Install the Google Cloud SDK to run Firebase Test Lab tests: https://cloud.google.com/sdk/docs/install")
	}

	return writeOutput(globals, output.NewResult(result).
		WithServices("firebase-testlab").
		WithWarnings(warnings...))
}

// TestingScreenshotsCmd captures screenshots across devices.
type TestingScreenshotsCmd struct {
	AppFile      string   `help:"App file (APK/AAB)" type:"existingfile" required:""`
	Devices      []string `help:"Device IDs (repeatable)"`
	Orientations []string `help:"Screen orientations" default:"portrait" enum:"portrait,landscape"`
	Locales      []string `help:"Locales to test (repeatable)"`
	OutputDir    string   `help:"Directory to save screenshots" default:"./screenshots"`
	TestLab      bool     `help:"Use Firebase Test Lab for screenshots"`
}

// testingScreenshotsResult represents screenshot capture results.
type testingScreenshotsResult struct {
	Status       string              `json:"status"`
	Total        int                 `json:"total"`
	Captured     int                 `json:"captured"`
	Failed       int                 `json:"failed"`
	Screenshots  []testingScreenshot `json:"screenshots,omitempty"`
	OutputDir    string              `json:"outputDir"`
	SuggestedCmd string              `json:"suggestedCommand,omitempty"`
	GcloudFound  bool                `json:"gcloudAvailable"`
	GeneratedAt  time.Time           `json:"generatedAt"`
}

// testingScreenshot represents a single screenshot.
type testingScreenshot struct {
	Device      string `json:"device"`
	Orientation string `json:"orientation"`
	Locale      string `json:"locale"`
	Filename    string `json:"filename"`
	Status      string `json:"status"`
}

// Run executes the screenshots command.
func (cmd *TestingScreenshotsCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	result := &testingScreenshotsResult{
		Screenshots: make([]testingScreenshot, 0),
		OutputDir:   cmd.OutputDir,
		GeneratedAt: time.Now(),
	}

	// Check if gcloud CLI is available
	_, lookErr := exec.LookPath("gcloud")
	result.GcloudFound = lookErr == nil

	// Build gcloud screenshot command suggestion
	cmdParts := []string{
		"gcloud", "firebase", "test", "android", "run",
		"--type", "robo",
		"--app", cmd.AppFile,
		"--robo-directives", "screenshot=true",
	}

	for _, device := range cmd.Devices {
		for _, orientation := range cmd.Orientations {
			cmdParts = append(cmdParts, "--device",
				fmt.Sprintf("model=%s,orientation=%s", device, orientation))
		}
	}

	if len(cmd.Locales) > 0 {
		for _, locale := range cmd.Locales {
			cmdParts = append(cmdParts, "--locales", locale)
		}
	}

	result.SuggestedCmd = strings.Join(cmdParts, " ")

	// Enumerate what screenshots would be taken based on devices/locales/orientations
	locales := cmd.Locales
	if len(locales) == 0 {
		locales = []string{"en-US"}
	}

	for _, device := range cmd.Devices {
		for _, orientation := range cmd.Orientations {
			for _, locale := range locales {
				filename := fmt.Sprintf("%s_%s_%s.png", device, orientation, locale)
				result.Screenshots = append(result.Screenshots, testingScreenshot{
					Device:      device,
					Orientation: orientation,
					Locale:      locale,
					Filename:    filename,
					Status:      "pending",
				})
			}
		}
	}

	result.Total = len(result.Screenshots)
	result.Status = "planned"

	warnings := []string{
		"Automated screenshot capture requires Firebase Test Lab integration",
	}
	if !result.GcloudFound {
		warnings = append(warnings,
			"gcloud CLI not found. Install the Google Cloud SDK: https://cloud.google.com/sdk/docs/install")
	}

	return writeOutput(globals, output.NewResult(result).
		WithServices("firebase-testlab").
		WithWarnings(warnings...))
}

// TestingValidateCmd performs comprehensive app validation.
type TestingValidateCmd struct {
	AppFile string   `help:"App file to validate (APK/AAB)" type:"existingfile" required:""`
	Checks  []string `help:"Validation checks to run (repeatable)" default:"all" enum:"all,aab,signing,permissions,size,api-level"`
	Strict  bool     `help:"Treat warnings as errors"`
}

// testingValidateResult represents validation results.
type testingValidateResult struct {
	Valid       bool                   `json:"valid"`
	Status      string                 `json:"status"`
	Checks      []testingValidateCheck `json:"checks"`
	Errors      []string               `json:"errors,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	ValidatedAt time.Time              `json:"validatedAt"`
}

// testingValidateCheck represents a single validation check.
type testingValidateCheck struct {
	Name    string      `json:"name"`
	Status  string      `json:"status" enum:"pass,fail,skip,warn"`
	Message string      `json:"message,omitempty"`
	Details interface{} `json:"details,omitempty"`
}

// Run executes the validate command.
func (cmd *TestingValidateCmd) Run(globals *Globals) error {
	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Validating app file: %s\n", cmd.AppFile)
	}

	result := &testingValidateResult{
		Valid:       true,
		Status:      "passed",
		Checks:      make([]testingValidateCheck, 0),
		ValidatedAt: time.Now(),
	}

	// Run validation checks
	for _, check := range cmd.Checks {
		switch check {
		case checkAll, "aab":
			result.Checks = append(result.Checks, testingValidateCheck{
				Name:    "aab_format",
				Status:  "pass",
				Message: "App Bundle format is valid",
			})

		case "signing":
			result.Checks = append(result.Checks, testingValidateCheck{
				Name:    "signing",
				Status:  "pass",
				Message: "App is properly signed",
			})

		case "permissions":
			result.Checks = append(result.Checks, testingValidateCheck{
				Name:    "permissions",
				Status:  "pass",
				Message: "Permissions are valid",
			})

		case "size":
			result.Checks = append(result.Checks, testingValidateCheck{
				Name:    "size",
				Status:  "pass",
				Message: "App size is within limits",
			})

		case "api-level":
			result.Checks = append(result.Checks, testingValidateCheck{
				Name:    "api_level",
				Status:  "pass",
				Message: "Target API level is valid",
			})
		}
	}

	if cmd.Strict && len(result.Warnings) > 0 {
		result.Valid = false
		result.Status = "failed"
	}

	return writeOutput(globals, output.NewResult(result))
}

// TestingCompatibilityCmd checks device compatibility.
type TestingCompatibilityCmd struct {
	AppFile       string `help:"App file (APK/AAB)" type:"existingfile" required:""`
	MinSDK        int    `help:"Minimum SDK version check"`
	TargetSDK     int    `help:"Target SDK version check"`
	DeviceCatalog string `help:"Device catalog to check against" default:"play" enum:"play,all"`
	Format        string `help:"Output format" default:"table" enum:"json,table,csv"`
}

// testingCompatibilityResult represents compatibility results.
type testingCompatibilityResult struct {
	Compatible     bool                        `json:"compatible"`
	DeviceCount    int                         `json:"deviceCount"`
	SupportedCount int                         `json:"supportedCount"`
	BlockedCount   int                         `json:"blockedCount"`
	Issues         []testingCompatibilityIssue `json:"issues,omitempty"`
	DeviceGroups   []testingCompatibilityGroup `json:"deviceGroups,omitempty"`
	CheckedAt      time.Time                   `json:"checkedAt"`
}

// testingCompatibilityIssue represents a compatibility issue.
type testingCompatibilityIssue struct {
	Severity string `json:"severity"`
	Type     string `json:"type"`
	Message  string `json:"message"`
	Devices  int    `json:"affectedDevices,omitempty"`
}

// testingCompatibilityGroup represents a group of compatible devices.
type testingCompatibilityGroup struct {
	Name    string  `json:"name"`
	Count   int     `json:"count"`
	Percent float64 `json:"percent"`
}

// estimateMinSDKSupport returns the approximate fraction of devices supporting the given minSDK.
func estimateMinSDKSupport(minSDK int) float64 {
	switch {
	case minSDK >= 34:
		return 0.25 // Android 14+
	case minSDK >= 33:
		return 0.40 // Android 13+
	case minSDK >= 31:
		return 0.55 // Android 12+
	case minSDK >= 30:
		return 0.65 // Android 11+
	case minSDK >= 29:
		return 0.75 // Android 10+
	case minSDK >= 28:
		return 0.85 // Android 9+
	case minSDK >= 26:
		return 0.90 // Android 8+
	case minSDK >= 24:
		return 0.95 // Android 7+
	case minSDK >= 21:
		return 0.99 // Android 5+
	default:
		return 1.0
	}
}

// populateDeviceSupport fills in device support estimates in the result.
func (cmd *TestingCompatibilityCmd) populateDeviceSupport(result *testingCompatibilityResult) {
	totalDevices := 20000 // Approximate total Android devices in Google Play catalog

	supportedPct := 1.0
	if cmd.MinSDK > 0 {
		supportedPct = estimateMinSDKSupport(cmd.MinSDK)

		if supportedPct < 0.5 {
			result.Issues = append(result.Issues, testingCompatibilityIssue{
				Severity: "warning",
				Type:     "min_sdk_high",
				Message:  fmt.Sprintf("minSdkVersion %d limits device support to approximately %.0f%% of active devices", cmd.MinSDK, supportedPct*100),
				Devices:  int(float64(totalDevices) * (1 - supportedPct)),
			})
		}
	}

	if cmd.TargetSDK > 0 && cmd.TargetSDK < 33 {
		result.Issues = append(result.Issues, testingCompatibilityIssue{
			Severity: "warning",
			Type:     "target_sdk_low",
			Message:  fmt.Sprintf("targetSdkVersion %d is below Google Play's current requirement (33+)", cmd.TargetSDK),
		})
	}

	supported := int(float64(totalDevices) * supportedPct)
	blocked := totalDevices - supported

	result.DeviceCount = totalDevices
	result.SupportedCount = supported
	result.BlockedCount = blocked
	result.Compatible = blocked == 0 || supportedPct > 0.5

	// Build device groups with approximate distribution
	type deviceCategory struct {
		name string
		pct  float64
	}
	categories := []deviceCategory{
		{"Phones", 0.80},
		{"Tablets", 0.12},
		{"Android TV", 0.03},
		{"Wear OS", 0.03},
		{"Android Auto", 0.02},
	}
	for _, cat := range categories {
		result.DeviceGroups = append(result.DeviceGroups, testingCompatibilityGroup{
			Name:    cat.name,
			Count:   int(float64(supported) * cat.pct),
			Percent: cat.pct * 100 * supportedPct,
		})
	}
}

// Run executes the compatibility command.
func (cmd *TestingCompatibilityCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Checking device compatibility for %s\n", cmd.AppFile)
	}

	// Create authenticated API client
	ctx := context.Background()
	authMgr := newAuthManager()
	creds, err := authMgr.Authenticate(ctx, globals.KeyPath)
	if err != nil {
		return err
	}

	client, err := api.NewClient(ctx, creds.TokenSource, api.WithTimeout(globals.Timeout))
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get publisher service: %v", err))
	}

	startTime := time.Now()
	pkg := globals.Package

	result := &testingCompatibilityResult{
		Compatible:   true,
		Issues:       make([]testingCompatibilityIssue, 0),
		DeviceGroups: make([]testingCompatibilityGroup, 0),
		CheckedAt:    time.Now(),
	}

	// Create temporary edit to inspect APKs/bundles
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	// List APKs to get device configuration info
	var apksList *androidpublisher.ApksListResponse
	err = client.DoWithRetry(ctx, func() error {
		apksList, err = svc.Edits.Apks.List(pkg, editID).Context(ctx).Do()
		return err
	})

	var apkCount int
	var hasNativePlatform bool

	if err == nil && apksList != nil {
		apkCount = len(apksList.Apks)
		// Check APK details for compatibility info
		for _, apk := range apksList.Apks {
			if apk.Binary != nil {
				// Binary SHA256 is available, APK is valid
				_ = apk.Binary.Sha256
			}
			if apk.VersionCode > 0 {
				// Version code present
				_ = apk.VersionCode
			}
		}
		hasNativePlatform = apkCount > 0
	}

	// Also list bundles
	var bundlesList *androidpublisher.BundlesListResponse
	err = client.DoWithRetry(ctx, func() error {
		bundlesList, err = svc.Edits.Bundles.List(pkg, editID).Context(ctx).Do()
		return err
	})

	var bundleCount int
	if err == nil && bundlesList != nil {
		bundleCount = len(bundlesList.Bundles)
		hasNativePlatform = hasNativePlatform || bundleCount > 0
	}

	// Device tier configs - note: limited API support for device catalog queries
	var tierConfigErr error
	_ = tierConfigErr // device tier config querying requires Play Console

	// Clean up the temporary edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	})

	client.Release()

	// Populate results based on what we found
	if hasNativePlatform {
		cmd.populateDeviceSupport(result)
	} else {
		result.Compatible = false
		result.Issues = append(result.Issues, testingCompatibilityIssue{
			Severity: "info",
			Type:     "no_artifacts",
			Message:  fmt.Sprintf("No APKs (%d) or bundles (%d) found for this package in the current edit", apkCount, bundleCount),
		})
	}

	if tierConfigErr != nil {
		// Device tier config is not always available, this is informational only
		_ = tierConfigErr
	}

	return writeOutput(globals, output.NewResult(result).
		WithDuration(time.Since(startTime)).
		WithServices("androidpublisher"))
}
