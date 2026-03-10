package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// SystemApksCmd contains system APKs management commands.
type SystemApksCmd struct {
	Variants SystemApksVariantsCmd `cmd:"" help:"System APK variants management"`
}

// SystemApksVariantsCmd contains system APK variant commands.
type SystemApksVariantsCmd struct {
	List     SystemApksVariantsListCmd     `cmd:"" help:"List system APK variants for a version code"`
	Get      SystemApksVariantsGetCmd      `cmd:"" help:"Get a specific system APK variant"`
	Create   SystemApksVariantsCreateCmd   `cmd:"" help:"Create a system APK variant for a device spec"`
	Download SystemApksVariantsDownloadCmd `cmd:"" help:"Download a system APK variant"`
}

// SystemApksVariantsListCmd lists system APK variants for a version code.
type SystemApksVariantsListCmd struct {
	VersionCode int64 `help:"Version code of the app bundle" required:""`
}

// systemApksVariantInfo represents a variant in the list.
type systemApksVariantInfo struct {
	VariantId        int64                 `json:"variantId,omitempty"`
	DeviceSpec       *deviceSpecInfo       `json:"deviceSpec,omitempty"`
	SystemApkOptions *systemApkOptionsInfo `json:"systemApkOptions,omitempty"`
}

// deviceSpecInfo represents device specifications.
type deviceSpecInfo struct {
	ScreenDensity    int64    `json:"screenDensity,omitempty"`
	SupportedAbis    []string `json:"supportedAbis,omitempty"`
	SupportedLocales []string `json:"supportedLocales,omitempty"`
}

// systemApkOptionsInfo represents system APK options.
type systemApkOptionsInfo struct {
	Rotated                     bool `json:"rotated,omitempty"`
	UncompressedDexFiles        bool `json:"uncompressedDexFiles,omitempty"`
	UncompressedNativeLibraries bool `json:"uncompressedNativeLibraries,omitempty"`
}

// systemApksVariantsListResult represents the list result.
type systemApksVariantsListResult struct {
	PackageName string                   `json:"packageName"`
	VersionCode int64                    `json:"versionCode"`
	Variants    []*systemApksVariantInfo `json:"variants,omitempty"`
}

// Run executes the list command.
func (cmd *SystemApksVariantsListCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
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

	pkg := globals.Package

	// List system APK variants
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var resp *androidpublisher.SystemApksListResponse
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Systemapks.Variants.List(pkg, cmd.VersionCode).Context(ctx).Do()
		return callErr
	})

	client.Release()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("system APK variants not found for version code %d", cmd.VersionCode)).
				WithHint("Ensure the bundle has been uploaded and system APKs have been generated")
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list system APK variants: %v", err))
	}

	// Convert to output format
	result := &systemApksVariantsListResult{
		PackageName: globals.Package,
		VersionCode: cmd.VersionCode,
	}

	for _, variant := range resp.Variants {
		variantInfo := convertVariantToInfo(variant)
		result.Variants = append(result.Variants, variantInfo)
	}

	res := output.NewResult(result).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(res, globals.Output, globals.Pretty)
}

// convertVariantToInfo converts an androidpublisher.Variant to systemApksVariantInfo.
func convertVariantToInfo(variant *androidpublisher.Variant) *systemApksVariantInfo {
	info := &systemApksVariantInfo{
		VariantId: variant.VariantId,
	}

	if variant.DeviceSpec != nil {
		info.DeviceSpec = &deviceSpecInfo{
			ScreenDensity:    variant.DeviceSpec.ScreenDensity,
			SupportedAbis:    variant.DeviceSpec.SupportedAbis,
			SupportedLocales: variant.DeviceSpec.SupportedLocales,
		}
	}

	if variant.Options != nil {
		info.SystemApkOptions = &systemApkOptionsInfo{
			Rotated:                     variant.Options.Rotated,
			UncompressedDexFiles:        variant.Options.UncompressedDexFiles,
			UncompressedNativeLibraries: variant.Options.UncompressedNativeLibraries,
		}
	}

	return info
}

// SystemApksVariantsGetCmd gets a specific system APK variant.
type SystemApksVariantsGetCmd struct {
	VersionCode int64 `help:"Version code of the app bundle" required:""`
	VariantId   int64 `help:"Variant ID of the system APK" required:""`
}

// systemApksVariantGetResult represents the get result.
type systemApksVariantGetResult struct {
	PackageName string                 `json:"packageName"`
	VersionCode int64                  `json:"versionCode"`
	Variant     *systemApksVariantInfo `json:"variant,omitempty"`
}

// Run executes the get command.
func (cmd *SystemApksVariantsGetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
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

	pkg := globals.Package

	// Get system APK variant
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var resp *androidpublisher.Variant
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		resp, callErr = svc.Systemapks.Variants.Get(pkg, cmd.VersionCode, cmd.VariantId).Context(ctx).Do()
		return callErr
	})

	client.Release()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("system APK variant %d not found for version code %d", cmd.VariantId, cmd.VersionCode)).
				WithHint("Use 'gpd system-apks variants list' to see available variants")
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get system APK variant: %v", err))
	}

	// Convert to output format
	result := &systemApksVariantGetResult{
		PackageName: globals.Package,
		VersionCode: cmd.VersionCode,
		Variant:     convertVariantToInfo(resp),
	}

	res := output.NewResult(result).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(res, globals.Output, globals.Pretty)
}

// SystemApksVariantsCreateCmd creates a system APK variant for a device spec.
type SystemApksVariantsCreateCmd struct {
	VersionCode  int64  `help:"Version code of the app bundle" required:""`
	DeviceSpec   string `help:"Device specification (JSON or comma-separated: density=480,abis=arm64-v8a,locales=en-US,es-ES)"`
	File         string `help:"Path to JSON file with device specification" type:"existingfile"`
	NoAutoCommit bool   `help:"Keep edit open after operation"`
	DryRun       bool   `help:"Show actions without executing"`
}

// systemApksVariantsCreateResult represents the create result.
type systemApksVariantsCreateResult struct {
	PackageName string                 `json:"packageName"`
	VersionCode int64                  `json:"versionCode"`
	Variant     *systemApksVariantInfo `json:"variant,omitempty"`
}

// Run executes the create command.
func (cmd *SystemApksVariantsCreateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	// Validate that at least one device spec source is provided
	if cmd.DeviceSpec == "" && cmd.File == "" {
		return errors.NewAPIError(errors.CodeValidationError, "device specification is required").
			WithHint("Provide --device-spec or --file flag with device specification")
	}

	// Parse device specification
	spec, err := cmd.parseDeviceSpec()
	if err != nil {
		return err
	}

	// Handle dry-run
	if cmd.DryRun {
		result := &systemApksVariantsCreateResult{
			PackageName: globals.Package,
			VersionCode: cmd.VersionCode,
			Variant: &systemApksVariantInfo{
				DeviceSpec: &deviceSpecInfo{
					ScreenDensity:    spec.ScreenDensity,
					SupportedAbis:    spec.SupportedAbis,
					SupportedLocales: spec.SupportedLocales,
				},
			},
		}
		res := output.NewResult(result).
			WithDuration(time.Since(start)).
			WithServices("androidpublisher")
		return outputResult(res, globals.Output, globals.Pretty)
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

	pkg := globals.Package

	// Create system APK variant
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var resp *androidpublisher.Variant
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		// Wrap DeviceSpec in Variant for the Create API
		variant := &androidpublisher.Variant{
			DeviceSpec: spec,
		}
		resp, callErr = svc.Systemapks.Variants.Create(pkg, cmd.VersionCode, variant).Context(ctx).Do()
		return callErr
	})

	client.Release()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("bundle with version code %d not found", cmd.VersionCode)).
				WithHint("Ensure the bundle has been uploaded to Google Play")
		}
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 400 {
			return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid device specification: %v", err)).
				WithHint("Check that all required fields are provided and valid")
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create system APK variant: %v", err))
	}

	// Convert to output format
	result := &systemApksVariantsCreateResult{
		PackageName: globals.Package,
		VersionCode: cmd.VersionCode,
		Variant:     convertVariantToInfo(resp),
	}

	res := output.NewResult(result).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(res, globals.Output, globals.Pretty)
}

// parseDeviceSpec parses the device specification from file or inline input.
func (cmd *SystemApksVariantsCreateCmd) parseDeviceSpec() (*androidpublisher.DeviceSpec, error) {
	spec := &androidpublisher.DeviceSpec{}

	// Parse from file if provided
	if cmd.File != "" {
		data, err := os.ReadFile(cmd.File)
		if err != nil {
			return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read device spec file: %v", err)).
				WithHint("Ensure the file exists and is readable")
		}
		if err := json.Unmarshal(data, spec); err != nil {
			return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse device spec JSON: %v", err)).
				WithHint("Ensure the file contains valid JSON with device specification fields")
		}
		return spec, nil
	}

	// Parse inline device spec
	if cmd.DeviceSpec != "" {
		pairs := strings.Split(cmd.DeviceSpec, ",")
		for _, pair := range pairs {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}

			eq := strings.Index(pair, "=")
			if eq == -1 {
				return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid device spec format: %s", pair)).
					WithHint("Use format: key=value,key2=value2 (e.g., density=480,abis=arm64-v8a,locales=en-US)")
			}

			key := strings.ToLower(strings.TrimSpace(pair[:eq]))
			value := strings.TrimSpace(pair[eq+1:])

			switch key {
			case "density", "screendensity":
				density, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid screen density: %s", value)).
						WithHint("Screen density must be an integer (e.g., 480)")
				}
				spec.ScreenDensity = density
			case "abis", "supportedabis":
				spec.SupportedAbis = strings.Split(value, "|")
			case "locales", "supportedlocales":
				spec.SupportedLocales = strings.Split(value, "|")
			default:
				return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("unknown device spec key: %s", key)).
					WithHint("Valid keys: density, abis, locales")
			}
		}
	}

	// Validate required fields
	if spec.ScreenDensity == 0 && len(spec.SupportedAbis) == 0 && len(spec.SupportedLocales) == 0 {
		return nil, errors.NewAPIError(errors.CodeValidationError, "at least one device specification field must be provided").
			WithHint("Provide density, abis, or locales (e.g., density=480 or abis=arm64-v8a)")
	}

	return spec, nil
}

// SystemApksVariantsDownloadCmd downloads a system APK variant.
type SystemApksVariantsDownloadCmd struct {
	VersionCode int64  `help:"Version code of the app bundle" required:""`
	VariantID   int64  `help:"Variant ID to download" required:""`
	OutputFile  string `help:"Output file path (default: systemapk-{variantId}.apk)" name:"output-file"`
}

// systemApksVariantsDownloadResult represents the download result.
type systemApksVariantsDownloadResult struct {
	PackageName string `json:"packageName"`
	VersionCode int64  `json:"versionCode"`
	VariantID   int64  `json:"variantId"`
	File        string `json:"file"`
	Size        int64  `json:"size,omitempty"`
}

// Run executes the download command.
func (cmd *SystemApksVariantsDownloadCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	// Determine output file
	outputFile := cmd.OutputFile
	if outputFile == "" {
		outputFile = fmt.Sprintf("systemapk-%d.apk", cmd.VariantID)
	}

	// Ensure directory exists
	dir := filepath.Dir(outputFile)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create directory: %v", err))
		}
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

	pkg := globals.Package

	// Download the APK - the Download method returns an *http.Response
	if err := client.AcquireForUpload(ctx); err != nil {
		return err
	}

	var resp *http.Response
	err = client.DoWithRetry(ctx, func() error {
		var callErr error
		//nolint:bodyclose // Response body is closed in the deferred close below
		resp, callErr = svc.Systemapks.Variants.Download(pkg, cmd.VersionCode, cmd.VariantID).Context(ctx).Download()
		return callErr
	})

	client.ReleaseForUpload()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("system APK variant %d not found for version code %d", cmd.VariantID, cmd.VersionCode)).
				WithHint("Use 'gpd system-apks variants list' to see available variants")
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to download APK: %v", err))
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			// Log close error but don't override original error
			_ = cerr
		}
	}()

	// Write to file
	file, err := os.Create(outputFile)
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create output file: %v", err))
	}
	defer func() {
		if cerr := file.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	size, err := io.Copy(file, resp.Body)
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to write APK file: %v", err))
	}

	result := &systemApksVariantsDownloadResult{
		PackageName: globals.Package,
		VersionCode: cmd.VersionCode,
		VariantID:   cmd.VariantID,
		File:        outputFile,
		Size:        size,
	}

	res := output.NewResult(result).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(res, globals.Output, globals.Pretty)
}
