// Package cli provides bulk operations commands for batch processing.
package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

const (
	extAAB = ".aab"
	extAPK = ".apk"
)

// BulkCmd contains batch operations commands.
type BulkCmd struct {
	Upload   BulkUploadCmd   `cmd:"" help:"Upload multiple APKs/AABs in parallel"`
	Listings BulkListingsCmd `cmd:"" help:"Update store listings across multiple locales"`
	Images   BulkImagesCmd   `cmd:"" help:"Batch upload images for multiple types"`
	Tracks   BulkTracksCmd   `cmd:"" help:"Update multiple tracks at once"`
}

// BulkUploadCmd uploads multiple APK/AAB files in parallel.
type BulkUploadCmd struct {
	Files                     []string `arg:"" help:"APK/AAB files to upload" type:"existingfile"`
	Track                     string   `help:"Target track" default:"internal" enum:"internal,alpha,beta,production"`
	EditID                    string   `help:"Explicit edit transaction ID"`
	NoAutoCommit              bool     `help:"Keep edit open for manual commit"`
	InProgressReviewBehaviour string   `help:"Behavior when committing while review in progress: THROW_ERROR_IF_IN_PROGRESS, CANCEL_IN_PROGRESS_AND_SUBMIT, or IN_PROGRESS_REVIEW_BEHAVIOUR_UNSPECIFIED" enum:"THROW_ERROR_IF_IN_PROGRESS,CANCEL_IN_PROGRESS_AND_SUBMIT,IN_PROGRESS_REVIEW_BEHAVIOUR_UNSPECIFIED," default:""`
	DryRun                    bool     `help:"Show intended actions without executing"`
	MaxParallel               int      `help:"Maximum parallel uploads" default:"3"`
}

// bulkUploadResult represents the result of a bulk upload operation.
type bulkUploadResult struct {
	SuccessCount   int                    `json:"successCount"`
	FailureCount   int                    `json:"failureCount"`
	SkippedCount   int                    `json:"skippedCount"`
	Uploads        []bulkUploadItemResult `json:"uploads"`
	EditID         string                 `json:"editId,omitempty"`
	Committed      bool                   `json:"committed"`
	ProcessingTime string                 `json:"processingTime"`
}

// bulkUploadItemResult represents the result of a single upload.
type bulkUploadItemResult struct {
	File        string `json:"file"`
	VersionCode int64  `json:"versionCode,omitempty"`
	Status      string `json:"status"`
	Error       string `json:"error,omitempty"`
	SHA1        string `json:"sha1,omitempty"`
}

// Run executes the bulk upload command.
func (cmd *BulkUploadCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if len(cmd.Files) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "at least one file is required").
			WithHint("Provide APK or AAB files to upload")
	}

	start := time.Now()

	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Bulk upload: %d file(s) to track %s\n", len(cmd.Files), cmd.Track)
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"files":       cmd.Files,
			"track":       cmd.Track,
			"dryRun":      true,
			"wouldUpload": len(cmd.Files),
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - no files uploaded")
		return writeOutput(globals, result)
	}

	// Create authenticated API client
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
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

	// Create edit if not specified
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}

		svc, err := client.AndroidPublisher()
		if err != nil {
			client.Release()
			return err
		}

		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})

		client.Release()

		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, "failed to create edit").
				WithDetails(map[string]interface{}{"error": err.Error()})
		}
		editID = edit.Id
		if globals.Verbose {
			fmt.Fprintf(os.Stderr, "Created edit: %s\n", editID)
		}
	}

	// Process uploads in parallel with controlled concurrency
	result := &bulkUploadResult{
		Uploads: make([]bulkUploadItemResult, 0, len(cmd.Files)),
		EditID:  editID,
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, cmd.MaxParallel)
	var mu sync.Mutex

	for _, file := range cmd.Files {
		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			item := cmd.uploadFile(ctx, client, globals.Package, editID, f, globals.Verbose)

			mu.Lock()
			result.Uploads = append(result.Uploads, item)
			switch item.Status {
			case "success":
				result.SuccessCount++
			case "skipped":
				result.SkippedCount++
			default:
				result.FailureCount++
			}
			mu.Unlock()
		}(file)
	}

	wg.Wait()

	// Commit edit if auto-commit enabled
	if !cmd.NoAutoCommit && result.FailureCount == 0 {
		if err := cmd.commitEdit(ctx, client, globals.Package, editID); err != nil {
			if globals.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to commit edit: %v\n", err)
			}
		} else {
			result.Committed = true
		}
	}

	result.ProcessingTime = time.Since(start).String()

	outputResult := output.NewResult(result).
		WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	if result.FailureCount > 0 {
		outputResult = outputResult.WithWarnings(fmt.Sprintf("%d uploads failed", result.FailureCount))
	}

	return writeOutput(globals, outputResult)
}

func (cmd *BulkUploadCmd) uploadFile(ctx context.Context, client *api.Client, pkg, editID, file string, verbose bool) bulkUploadItemResult {
	// Detect file type
	ext := strings.ToLower(filepath.Ext(file))

	f, err := os.Open(file)
	if err != nil {
		return bulkUploadItemResult{File: file, Status: "failed", Error: err.Error()}
	}
	defer func() { _ = f.Close() }()

	svc, err := client.AndroidPublisher()
	if err != nil {
		return bulkUploadItemResult{File: file, Status: "failed", Error: err.Error()}
	}

	if err := client.AcquireForUpload(ctx); err != nil {
		return bulkUploadItemResult{File: file, Status: "failed", Error: err.Error()}
	}
	defer client.ReleaseForUpload()

	if verbose {
		fmt.Fprintf(os.Stderr, "Uploading %s (%s)\n", filepath.Base(file), ext)
	}

	switch ext {
	case extAAB:
		var bundle *androidpublisher.Bundle
		err = client.DoWithRetry(ctx, func() error {
			var uerr error
			bundle, uerr = svc.Edits.Bundles.Upload(pkg, editID).Media(f).Context(ctx).Do()
			return uerr
		})
		if err != nil {
			return bulkUploadItemResult{File: file, Status: "failed", Error: err.Error()}
		}
		return bulkUploadItemResult{File: file, VersionCode: bundle.VersionCode, Status: "success", SHA1: bundle.Sha1}
	case extAPK:
		var apk *androidpublisher.Apk
		err = client.DoWithRetry(ctx, func() error {
			var uerr error
			apk, uerr = svc.Edits.Apks.Upload(pkg, editID).Media(f).Context(ctx).Do()
			return uerr
		})
		if err != nil {
			return bulkUploadItemResult{File: file, Status: "failed", Error: err.Error()}
		}
		return bulkUploadItemResult{File: file, VersionCode: apk.VersionCode, Status: "success"}
	default:
		return bulkUploadItemResult{File: file, Status: "failed", Error: "unsupported file type: " + ext}
	}
}

func (cmd *BulkUploadCmd) commitEdit(ctx context.Context, client *api.Client, pkg, editID string) error {
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	defer client.Release()

	svc, err := client.AndroidPublisher()
	if err != nil {
		return err
	}

	return client.DoWithRetry(ctx, func() error {
		if cmd.InProgressReviewBehaviour != "" {
			_, err := svc.Edits.Commit(pkg, editID).Context(ctx).Do(googleapi.QueryParameter("inProgressReviewBehaviour", cmd.InProgressReviewBehaviour))
			return err
		}
		_, err := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
		return err
	})
}

// BulkListingsCmd updates store listings across multiple locales.
type BulkListingsCmd struct {
	DataFile string `help:"JSON file with locale->listing mappings" type:"existingfile" required:""`
	EditID   string `help:"Explicit edit transaction ID"`
	DryRun   bool   `help:"Show intended actions without executing"`
}

// bulkListingData represents the structure of the listings data file.
type bulkListingData map[string]struct {
	Title            string `json:"title"`
	ShortDescription string `json:"shortDescription"`
	FullDescription  string `json:"fullDescription"`
	Video            string `json:"video,omitempty"`
}

// bulkListingsResult represents the result of bulk listings update.
type bulkListingsResult struct {
	SuccessCount int                     `json:"successCount"`
	FailureCount int                     `json:"failureCount"`
	Locales      []bulkListingItemResult `json:"locales"`
	EditID       string                  `json:"editId,omitempty"`
}

// bulkListingItemResult represents the result for a single locale.
type bulkListingItemResult struct {
	Locale string `json:"locale"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Run executes the bulk listings update command.
func (cmd *BulkListingsCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	// Read and parse the listings data file
	data, err := os.ReadFile(cmd.DataFile)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, "failed to read listings file").
			WithHint("Ensure the file exists and is readable").
			WithDetails(map[string]interface{}{"file": cmd.DataFile, "error": err.Error()})
	}

	var listings bulkListingData
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.NewAPIError(errors.CodeValidationError, "failed to parse listings JSON").
			WithHint("Ensure the file contains valid JSON with locale keys").
			WithDetails(map[string]interface{}{"error": err.Error()})
	}

	if len(listings) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "no listings found in data file").
			WithHint("Provide at least one locale with listing data")
	}

	if cmd.DryRun {
		locales := make([]string, 0, len(listings))
		for locale := range listings {
			locales = append(locales, locale)
		}

		result := output.NewResult(map[string]interface{}{
			"locales":     locales,
			"count":       len(listings),
			"dryRun":      true,
			"wouldUpdate": len(listings),
		}).WithNoOp("dry run - no listings updated")
		return writeOutput(globals, result)
	}

	// Create authenticated API client
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
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

	// Create edit if not specified
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}

		svc, svcErr := client.AndroidPublisher()
		if svcErr != nil {
			client.Release()
			return svcErr
		}

		var edit *androidpublisher.AppEdit
		createErr := client.DoWithRetry(ctx, func() error {
			var e error
			edit, e = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return e
		})

		client.Release()

		if createErr != nil {
			return errors.NewAPIError(errors.CodeGeneralError, "failed to create edit").
				WithDetails(map[string]interface{}{"error": createErr.Error()})
		}
		editID = edit.Id
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return err
	}

	result := &bulkListingsResult{
		Locales: make([]bulkListingItemResult, 0, len(listings)),
		EditID:  editID,
	}

	for locale, listing := range listings {
		cmd.updateListing(ctx, client, svc, globals.Package, editID, locale, listing, result)
	}

	// Commit edit if no failures
	if result.FailureCount == 0 {
		if commitErr := client.Acquire(ctx); commitErr == nil {
			commitErr = client.DoWithRetry(ctx, func() error {
				_, e := svc.Edits.Commit(globals.Package, editID).Context(ctx).Do()
				return e
			})
			client.Release()
			if commitErr != nil && globals.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to commit edit: %v\n", commitErr)
			}
		}
	}

	outputResult := output.NewResult(result).
		WithServices("androidpublisher")

	if result.FailureCount > 0 {
		outputResult = outputResult.WithWarnings(fmt.Sprintf("%d locale updates failed", result.FailureCount))
	}

	return writeOutput(globals, outputResult)
}

func (cmd *BulkListingsCmd) updateListing(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg, editID, locale string, listing struct {
	Title            string `json:"title"`
	ShortDescription string `json:"shortDescription"`
	FullDescription  string `json:"fullDescription"`
	Video            string `json:"video,omitempty"`
}, result *bulkListingsResult) {
	if err := client.Acquire(ctx); err != nil {
		result.FailureCount++
		result.Locales = append(result.Locales, bulkListingItemResult{
			Locale: locale, Status: "failed", Error: err.Error(),
		})
		return
	}

	updateErr := client.DoWithRetry(ctx, func() error {
		_, e := svc.Edits.Listings.Update(pkg, editID, locale, &androidpublisher.Listing{
			Title:            listing.Title,
			ShortDescription: listing.ShortDescription,
			FullDescription:  listing.FullDescription,
			Video:            listing.Video,
		}).Context(ctx).Do()
		return e
	})

	client.Release()

	if updateErr != nil {
		result.FailureCount++
		result.Locales = append(result.Locales, bulkListingItemResult{
			Locale: locale, Status: "failed", Error: updateErr.Error(),
		})
	} else {
		result.SuccessCount++
		result.Locales = append(result.Locales, bulkListingItemResult{
			Locale: locale, Status: "success",
		})
	}
}

// BulkImagesCmd batch uploads images for multiple types.
type BulkImagesCmd struct {
	ImageDir    string `help:"Directory with images organized by type/locale" type:"existingdir" required:""`
	Locale      string `help:"Target locale (overrides directory structure)" default:"en-US"`
	EditID      string `help:"Explicit edit transaction ID"`
	DryRun      bool   `help:"Show intended actions without executing"`
	MaxParallel int    `help:"Maximum parallel uploads" default:"3"`
}

// bulkImagesResult represents the result of bulk image upload.
type bulkImagesResult struct {
	SuccessCount int                   `json:"successCount"`
	FailureCount int                   `json:"failureCount"`
	Images       []bulkImageItemResult `json:"images"`
	EditID       string                `json:"editId,omitempty"`
}

// bulkImageItemResult represents the result for a single image.
type bulkImageItemResult struct {
	Type     string `json:"type"`
	Locale   string `json:"locale"`
	Filename string `json:"filename"`
	Status   string `json:"status"`
	Error    string `json:"error,omitempty"`
}

// scanImageDirectory walks the image directory and returns discovered images.
func (cmd *BulkImagesCmd) scanImageDirectory() ([]bulkImageItemResult, error) {
	var images []bulkImageItemResult
	err := filepath.Walk(cmd.ImageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Determine image type from directory structure
		relPath, _ := filepath.Rel(cmd.ImageDir, path)
		dir := filepath.Dir(relPath)

		// Directory structure expected: <image-type>/<locale>/<filename>
		// or <image-type>/<filename> for default locale
		parts := strings.Split(dir, string(filepath.Separator))
		imageType := ""
		locale := cmd.Locale

		if len(parts) > 0 && parts[0] != "." {
			imageType = parts[0]
		}
		if len(parts) > 1 {
			locale = parts[1]
		}

		if imageType != "" {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" {
				images = append(images, bulkImageItemResult{
					Type:     imageType,
					Locale:   locale,
					Filename: path,
					Status:   "pending",
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, "failed to scan image directory").
			WithDetails(map[string]interface{}{"error": err.Error()})
	}

	if len(images) == 0 {
		return nil, errors.NewAPIError(errors.CodeValidationError, "no images found in directory").
			WithHint("Ensure images are organized in subdirectories by type (e.g., phoneScreenshots/, featureGraphic/)")
	}

	return images, nil
}

// uploadImagesParallel uploads images in parallel with controlled concurrency and collects results.
func (cmd *BulkImagesCmd) uploadImagesParallel(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg, editID string, images []bulkImageItemResult) *bulkImagesResult {
	result := &bulkImagesResult{
		Images: make([]bulkImageItemResult, 0, len(images)),
		EditID: editID,
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, cmd.MaxParallel)
	var mu sync.Mutex

	for _, img := range images {
		wg.Add(1)
		go func(item bulkImageItemResult) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			imgResult := cmd.uploadImage(ctx, client, svc, pkg, editID, &item)

			mu.Lock()
			if imgResult.Status == "success" {
				result.SuccessCount++
			} else {
				result.FailureCount++
			}
			result.Images = append(result.Images, imgResult)
			mu.Unlock()
		}(img)
	}

	wg.Wait()
	return result
}

// Run executes the bulk images upload command.
func (cmd *BulkImagesCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	images, err := cmd.scanImageDirectory()
	if err != nil {
		return err
	}

	if cmd.DryRun {
		return writeOutput(globals, output.NewResult(map[string]interface{}{
			"images":      images,
			"count":       len(images),
			"dryRun":      true,
			"wouldUpload": len(images),
		}).WithNoOp("dry run - no images uploaded"))
	}

	// Create authenticated API client
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
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

	// Create edit if not specified
	editID := cmd.EditID
	if editID == "" {
		if acquireErr := client.Acquire(ctx); acquireErr != nil {
			return acquireErr
		}

		svc, svcErr := client.AndroidPublisher()
		if svcErr != nil {
			client.Release()
			return svcErr
		}

		var edit *androidpublisher.AppEdit
		createErr := client.DoWithRetry(ctx, func() error {
			var e error
			edit, e = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return e
		})

		client.Release()

		if createErr != nil {
			return errors.NewAPIError(errors.CodeGeneralError, "failed to create edit").
				WithDetails(map[string]interface{}{"error": createErr.Error()})
		}
		editID = edit.Id
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return err
	}

	result := cmd.uploadImagesParallel(ctx, client, svc, globals.Package, editID, images)

	// Commit edit if no failures
	if result.FailureCount == 0 {
		if acquireErr := client.Acquire(ctx); acquireErr == nil {
			commitErr := client.DoWithRetry(ctx, func() error {
				_, e := svc.Edits.Commit(globals.Package, editID).Context(ctx).Do()
				return e
			})
			client.Release()
			if commitErr != nil && globals.Verbose {
				fmt.Fprintf(os.Stderr, "Warning: failed to commit edit: %v\n", commitErr)
			}
		}
	}

	outputResult := output.NewResult(result).
		WithServices("androidpublisher")

	if result.FailureCount > 0 {
		outputResult = outputResult.WithWarnings(fmt.Sprintf("%d image uploads failed", result.FailureCount))
	}

	return writeOutput(globals, outputResult)
}

func (cmd *BulkImagesCmd) uploadImage(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg, editID string, item *bulkImageItemResult) bulkImageItemResult {
	f, openErr := os.Open(item.Filename)
	if openErr != nil {
		return bulkImageItemResult{
			Type: item.Type, Locale: item.Locale, Filename: item.Filename,
			Status: "failed", Error: openErr.Error(),
		}
	}
	defer func() { _ = f.Close() }()

	if acquireErr := client.AcquireForUpload(ctx); acquireErr != nil {
		return bulkImageItemResult{
			Type: item.Type, Locale: item.Locale, Filename: item.Filename,
			Status: "failed", Error: acquireErr.Error(),
		}
	}

	uploadErr := client.DoWithRetry(ctx, func() error {
		_, e := svc.Edits.Images.Upload(pkg, editID, item.Locale, item.Type).Media(f).Context(ctx).Do()
		return e
	})

	client.ReleaseForUpload()

	if uploadErr != nil {
		return bulkImageItemResult{
			Type: item.Type, Locale: item.Locale, Filename: item.Filename,
			Status: "failed", Error: uploadErr.Error(),
		}
	}
	return bulkImageItemResult{
		Type: item.Type, Locale: item.Locale, Filename: item.Filename,
		Status: "success",
	}
}

// BulkTracksCmd updates multiple tracks at once.
type BulkTracksCmd struct {
	Tracks       []string `help:"Tracks to update (repeatable)" enum:"internal,alpha,beta,production" required:""`
	VersionCodes []string `help:"Version codes to include (repeatable)" required:""`
	Status       string   `help:"Release status" default:"draft" enum:"draft,completed,halted,inProgress"`
	Name         string   `help:"Release name"`
	EditID       string   `help:"Explicit edit transaction ID"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// bulkTracksResult represents the result of bulk track update.
type bulkTracksResult struct {
	SuccessCount int                   `json:"successCount"`
	FailureCount int                   `json:"failureCount"`
	Tracks       []bulkTrackItemResult `json:"tracks"`
	EditID       string                `json:"editId,omitempty"`
	Committed    bool                  `json:"committed"`
}

// bulkTrackItemResult represents the result for a single track.
type bulkTrackItemResult struct {
	Track        string   `json:"track"`
	Status       string   `json:"status"`
	VersionCodes []string `json:"versionCodes"`
	Error        string   `json:"error,omitempty"`
}

// Run executes the bulk tracks update command.
func (cmd *BulkTracksCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if len(cmd.Tracks) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "at least one track is required")
	}
	if len(cmd.VersionCodes) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "at least one version code is required")
	}

	if cmd.DryRun {
		return writeOutput(globals, output.NewResult(map[string]interface{}{
			"tracks":       cmd.Tracks,
			"versionCodes": cmd.VersionCodes,
			"status":       cmd.Status,
			"dryRun":       true,
			"wouldUpdate":  len(cmd.Tracks),
		}).WithNoOp("dry run - no tracks updated"))
	}

	// Parse version codes to int64
	versionCodes := make([]int64, 0, len(cmd.VersionCodes))
	for _, vc := range cmd.VersionCodes {
		code, parseErr := strconv.ParseInt(vc, 10, 64)
		if parseErr != nil {
			return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid version code %q: %v", vc, parseErr)).
				WithHint("Version codes must be positive integers")
		}
		versionCodes = append(versionCodes, code)
	}

	// Create authenticated API client
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
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

	// Create edit if not specified
	editID := cmd.EditID
	if editID == "" {
		if acquireErr := client.Acquire(ctx); acquireErr != nil {
			return acquireErr
		}

		svc, svcErr := client.AndroidPublisher()
		if svcErr != nil {
			client.Release()
			return svcErr
		}

		var edit *androidpublisher.AppEdit
		createErr := client.DoWithRetry(ctx, func() error {
			var e error
			edit, e = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return e
		})

		client.Release()

		if createErr != nil {
			return errors.NewAPIError(errors.CodeGeneralError, "failed to create edit").
				WithDetails(map[string]interface{}{"error": createErr.Error()})
		}
		editID = edit.Id
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return err
	}

	result := &bulkTracksResult{
		Tracks: make([]bulkTrackItemResult, 0, len(cmd.Tracks)),
		EditID: editID,
	}

	for _, track := range cmd.Tracks {
		cmd.updateTrack(ctx, client, svc, globals.Package, editID, track, versionCodes, result)
	}

	// Commit edit if no failures
	if result.FailureCount == 0 {
		if acquireErr := client.Acquire(ctx); acquireErr == nil {
			commitErr := client.DoWithRetry(ctx, func() error {
				_, e := svc.Edits.Commit(globals.Package, editID).Context(ctx).Do()
				return e
			})
			client.Release()
			if commitErr != nil {
				if globals.Verbose {
					fmt.Fprintf(os.Stderr, "Warning: failed to commit edit: %v\n", commitErr)
				}
			} else {
				result.Committed = true
			}
		}
	}

	outputResult := output.NewResult(result).
		WithServices("androidpublisher")

	if result.FailureCount > 0 {
		outputResult = outputResult.WithWarnings(fmt.Sprintf("%d track updates failed", result.FailureCount))
	}

	return writeOutput(globals, outputResult)
}

func (cmd *BulkTracksCmd) updateTrack(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg, editID, track string, versionCodes []int64, result *bulkTracksResult) {
	if acquireErr := client.Acquire(ctx); acquireErr != nil {
		result.FailureCount++
		result.Tracks = append(result.Tracks, bulkTrackItemResult{
			Track: track, Status: "failed", VersionCodes: cmd.VersionCodes, Error: acquireErr.Error(),
		})
		return
	}

	release := &androidpublisher.TrackRelease{
		VersionCodes: versionCodes,
		Status:       cmd.Status,
	}
	if cmd.Name != "" {
		release.Name = cmd.Name
	}

	updateErr := client.DoWithRetry(ctx, func() error {
		_, e := svc.Edits.Tracks.Update(pkg, editID, track, &androidpublisher.Track{
			Track:    track,
			Releases: []*androidpublisher.TrackRelease{release},
		}).Context(ctx).Do()
		return e
	})

	client.Release()

	if updateErr != nil {
		result.FailureCount++
		result.Tracks = append(result.Tracks, bulkTrackItemResult{
			Track: track, Status: "failed", VersionCodes: cmd.VersionCodes, Error: updateErr.Error(),
		})
	} else {
		result.SuccessCount++
		result.Tracks = append(result.Tracks, bulkTrackItemResult{
			Track: track, Status: "success", VersionCodes: cmd.VersionCodes,
		})
	}
}
