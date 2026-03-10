package cli

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// GeneratedApksCmd contains generated APKs management commands.
type GeneratedApksCmd struct {
	List     GeneratedApksListCmd     `cmd:"" help:"List generated APK variants for a bundle"`
	Download GeneratedApksDownloadCmd `cmd:"" help:"Download a generated APK"`
}

// GeneratedApksListCmd lists generated APK variants for a bundle.
type GeneratedApksListCmd struct {
	VersionCode        int64  `help:"Version code of the app bundle" required:""`
	DeviceTierConfigID string `help:"Device tier config ID (optional)"`
}

// generatedApksListResult represents the list result.
type generatedApksListResult struct {
	PackageName   string                        `json:"packageName"`
	VersionCode   int64                         `json:"versionCode"`
	GeneratedApks []*generatedApksPerSigningKey `json:"generatedApks,omitempty"`
}

type generatedApksPerSigningKey struct {
	CertificateSha256Hash    string                     `json:"certificateSha256Hash,omitempty"`
	GeneratedSplitApks       []*generatedSplitApkInfo   `json:"generatedSplitApks,omitempty"`
	GeneratedStandaloneApks  []*generatedStandaloneApk  `json:"generatedStandaloneApks,omitempty"`
	GeneratedUniversalApk    *generatedUniversalApkInfo `json:"generatedUniversalApk,omitempty"`
	GeneratedAssetPackSlices []*generatedAssetPackSlice `json:"generatedAssetPackSlices,omitempty"`
}

type generatedSplitApkInfo struct {
	ModuleName string `json:"moduleName,omitempty"`
	VariantId  int64  `json:"variantId,omitempty"`
	DownloadId string `json:"downloadId,omitempty"`
	SplitId    string `json:"splitId,omitempty"`
}

type generatedStandaloneApk struct {
	VariantId  int64  `json:"variantId,omitempty"`
	DownloadId string `json:"downloadId,omitempty"`
}

type generatedUniversalApkInfo struct {
	DownloadId string `json:"downloadId,omitempty"`
}

type generatedAssetPackSlice struct {
	DownloadId string `json:"downloadId,omitempty"`
	ModuleName string `json:"moduleName,omitempty"`
	SliceId    string `json:"sliceId,omitempty"`
}

// Run executes the list command.
func (cmd *GeneratedApksListCmd) Run(globals *Globals) error {
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

	// List generated APKs
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var resp *androidpublisher.GeneratedApksListResponse
	err = client.DoWithRetry(ctx, func() error {
		var callOpts []googleapi.CallOption
		if cmd.DeviceTierConfigID != "" {
			callOpts = append(callOpts, googleapi.QueryParameter("deviceTierConfigId", cmd.DeviceTierConfigID))
		}
		resp, err = svc.Generatedapks.List(pkg, cmd.VersionCode).Context(ctx).Do(callOpts...)
		return err
	})

	client.Release()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("generated APKs not found for version code %d", cmd.VersionCode)).
				WithHint("Ensure the bundle has been uploaded and processed")
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list generated APKs: %v", err))
	}

	// Convert to output format
	result := &generatedApksListResult{
		PackageName: globals.Package,
		VersionCode: cmd.VersionCode,
	}

	for _, apk := range resp.GeneratedApks {
		signingKey := &generatedApksPerSigningKey{
			CertificateSha256Hash: apk.CertificateSha256Hash,
		}

		// Convert split APKs
		for _, splitApk := range apk.GeneratedSplitApks {
			signingKey.GeneratedSplitApks = append(signingKey.GeneratedSplitApks, &generatedSplitApkInfo{
				ModuleName: splitApk.ModuleName,
				VariantId:  splitApk.VariantId,
				DownloadId: splitApk.DownloadId,
				SplitId:    splitApk.SplitId,
			})
		}

		// Convert standalone APKs
		for _, standaloneApk := range apk.GeneratedStandaloneApks {
			signingKey.GeneratedStandaloneApks = append(signingKey.GeneratedStandaloneApks, &generatedStandaloneApk{
				VariantId:  standaloneApk.VariantId,
				DownloadId: standaloneApk.DownloadId,
			})
		}

		// Convert universal APK
		if apk.GeneratedUniversalApk != nil {
			signingKey.GeneratedUniversalApk = &generatedUniversalApkInfo{
				DownloadId: apk.GeneratedUniversalApk.DownloadId,
			}
		}

		// Convert asset pack slices
		for _, slice := range apk.GeneratedAssetPackSlices {
			signingKey.GeneratedAssetPackSlices = append(signingKey.GeneratedAssetPackSlices, &generatedAssetPackSlice{
				DownloadId: slice.DownloadId,
				ModuleName: slice.ModuleName,
				SliceId:    slice.SliceId,
			})
		}

		result.GeneratedApks = append(result.GeneratedApks, signingKey)
	}

	res := output.NewResult(result).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(res, globals.Output, globals.Pretty)
}

// GeneratedApksDownloadCmd downloads a generated APK.
type GeneratedApksDownloadCmd struct {
	VersionCode        int64  `help:"Version code of the app bundle" required:""`
	DownloadID         string `help:"Download ID from generatedapks.list" required:""`
	OutputFile         string `help:"Output file path (default: {downloadId}.apk in current directory)" name:"output-file"`
	DeviceTierConfigID string `help:"Device tier config ID (optional)"`
}

// generatedApksDownloadResult represents the download result.
type generatedApksDownloadResult struct {
	PackageName string `json:"packageName"`
	VersionCode int64  `json:"versionCode"`
	DownloadID  string `json:"downloadId"`
	File        string `json:"file"`
	Size        int64  `json:"size,omitempty"`
}

// Run executes the download command.
func (cmd *GeneratedApksDownloadCmd) Run(globals *Globals) error {
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
		outputFile = fmt.Sprintf("%s.apk", cmd.DownloadID)
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
		var callOpts []googleapi.CallOption
		if cmd.DeviceTierConfigID != "" {
			callOpts = append(callOpts, googleapi.QueryParameter("deviceTierConfigId", cmd.DeviceTierConfigID))
		}
		//nolint:bodyclose // Response body is closed in the deferred close below
		resp, err = svc.Generatedapks.Download(pkg, cmd.VersionCode, cmd.DownloadID).Context(ctx).Download(callOpts...)
		return err
	})

	client.ReleaseForUpload()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("APK not found for download ID: %s", cmd.DownloadID)).
				WithHint("Use 'gpd generated-apks list' to get valid download IDs")
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to download APK: %v", err))
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil && err == nil {
			err = cerr
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

	result := &generatedApksDownloadResult{
		PackageName: globals.Package,
		VersionCode: cmd.VersionCode,
		DownloadID:  cmd.DownloadID,
		File:        outputFile,
		Size:        size,
	}

	res := output.NewResult(result).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(res, globals.Output, globals.Pretty)
}
