package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// PublishCmd contains publishing commands.
type PublishCmd struct {
	Upload        PublishUploadCmd        `cmd:"" help:"Upload APK or AAB"`
	Release       PublishReleaseCmd       `cmd:"" help:"Create or update a release"`
	Rollout       PublishRolloutCmd       `cmd:"" help:"Update rollout percentage"`
	Promote       PublishPromoteCmd       `cmd:"" help:"Promote a release between tracks"`
	Halt          PublishHaltCmd          `cmd:"" help:"Halt a production rollout"`
	Rollback      PublishRollbackCmd      `cmd:"" help:"Rollback to a previous version"`
	Status        PublishStatusCmd        `cmd:"" help:"Get track status"`
	Tracks        PublishTracksCmd        `cmd:"" help:"List all tracks"`
	Capabilities  PublishCapabilitiesCmd  `cmd:"" help:"List publishing capabilities"`
	Listing       PublishListingCmd       `cmd:"" help:"Manage store listing"`
	Details       PublishDetailsCmd       `cmd:"" help:"Manage app details"`
	Images        PublishImagesCmd        `cmd:"" help:"Manage store images"`
	Assets        PublishAssetsCmd        `cmd:"" help:"Manage store assets"`
	Deobfuscation PublishDeobfuscationCmd `cmd:"" help:"Manage deobfuscation files"`
	Testers       PublishTestersCmd       `cmd:"" help:"Manage testers"`
	Builds        PublishBuildsCmd        `cmd:"" help:"Manage uploaded builds"`
	BetaGroups    PublishBetaGroupsCmd    `cmd:"" help:"Beta group management (ASC compatibility)"`
	InternalShare PublishInternalShareCmd `cmd:"" help:"Upload artifacts for internal sharing"`
}

// PublishUploadCmd uploads APK or AAB.
type PublishUploadCmd struct {
	File                      string `arg:"" help:"File to upload (APK or AAB)" type:"existingfile"`
	Track                     string `help:"Target track" default:"internal" enum:"internal,alpha,beta,production"`
	EditID                    string `help:"Explicit edit transaction ID"`
	ObbMain                   string `help:"Main expansion file path"`
	ObbPatch                  string `help:"Patch expansion file path"`
	ObbMainRefVersion         int64  `help:"Reference version code for main expansion file"`
	ObbPatchRefVersion        int64  `help:"Reference version code for patch expansion file"`
	NoAutoCommit              bool   `help:"Keep edit open for manual commit"`
	InProgressReviewBehaviour string `help:"Behavior when committing while review in progress: THROW_ERROR_IF_IN_PROGRESS, CANCEL_IN_PROGRESS_AND_SUBMIT, or IN_PROGRESS_REVIEW_BEHAVIOUR_UNSPECIFIED" enum:"THROW_ERROR_IF_IN_PROGRESS,CANCEL_IN_PROGRESS_AND_SUBMIT,IN_PROGRESS_REVIEW_BEHAVIOUR_UNSPECIFIED," default:""`
	DryRun                    bool   `help:"Show intended actions without executing"`
}

const (
	fileTypeAAB        = "aab"
	fileTypeAPK        = "apk"
	releaseCompleted   = "completed"
	releaseStatusDraft = "draft"
)

type uploadResult struct {
	VersionCode int64  `json:"versionCode"`
	SHA1        string `json:"sha1,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
	Type        string `json:"type"`
	EditID      string `json:"editId"`
	Committed   bool   `json:"committed"`
	File        string `json:"file"`
	Size        int64  `json:"size"`
}

// Run executes the upload command.
func (cmd *PublishUploadCmd) Run(globals *Globals) error {
	ctx := context.Background()
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	fileInfo, fileType, err := cmd.validateUploadFile()
	if err != nil {
		return err
	}

	if cmd.DryRun {
		return cmd.handleDryRunUpload(start, fileType, globals)
	}

	client, svc, err := cmd.createUploadClient(ctx, globals)
	if err != nil {
		return err
	}

	editID, err := cmd.getOrCreateEditID(ctx, client, svc, globals.Package)
	if err != nil {
		return err
	}

	versionCode, sha1, sha256, err := cmd.uploadBinary(ctx, client, svc, globals.Package, editID, fileType)
	if err != nil {
		return err
	}

	if err := cmd.uploadExpansionFiles(ctx, client, svc, globals.Package, editID, versionCode); err != nil {
		return err
	}

	committed, err := cmd.commitUploadEdit(ctx, client, svc, globals.Package, editID)
	if err != nil {
		return err
	}

	return cmd.buildUploadResult(start, fileInfo, fileType, editID, versionCode, sha1, sha256, committed, globals)
}

// validateUploadFile validates the file exists and is APK/AAB.
//
//nolint:gocritic // Named results would shadow local variables
func (cmd *PublishUploadCmd) validateUploadFile() (os.FileInfo, string, error) {
	if cmd.File == "" {
		return nil, "", errors.NewAPIError(errors.CodeValidationError, "file is required").
			WithHint("Provide an APK or AAB file to upload")
	}

	fileInfo, err := os.Stat(cmd.File)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("file not found: %s", cmd.File))
		}
		return nil, "", errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to access file: %v", err))
	}

	ext := strings.ToLower(filepath.Ext(cmd.File))
	if ext != extAPK && ext != extAAB {
		return nil, "", errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid file type: %s. Only .apk and .aab files are supported", ext))
	}

	fileType := fileTypeAPK
	if ext == extAAB {
		fileType = fileTypeAAB
	}

	return fileInfo, fileType, nil
}

// handleDryRunUpload handles dry-run mode for upload command.
func (cmd *PublishUploadCmd) handleDryRunUpload(start time.Time, fileType string, globals *Globals) error {
	result := output.NewResult(map[string]interface{}{
		"file":   cmd.File,
		"type":   fileType,
		"track":  cmd.Track,
		"dryRun": true,
	}).WithDuration(time.Since(start)).
		WithNoOp("dry run - no file uploaded")
	return outputResult(result, globals.Output, globals.Pretty)
}

// createUploadClient creates API client and service for upload.
func (cmd *PublishUploadCmd) createUploadClient(ctx context.Context, globals *Globals) (*api.Client, *androidpublisher.Service, error) {
	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return nil, nil, err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return nil, nil, errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	return client, svc, nil
}

// getOrCreateEditID gets existing or creates new edit ID.
func (cmd *PublishUploadCmd) getOrCreateEditID(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName string) (string, error) {
	editID := cmd.EditID
	if editID != "" {
		return editID, nil
	}

	if err := client.Acquire(ctx); err != nil {
		return "", err
	}

	var edit *androidpublisher.AppEdit
	var err error
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(packageName, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})

	client.Release()

	if err != nil {
		return "", errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	return edit.Id, nil
}

// uploadBinary uploads APK or AAB and returns version code and hashes.
//
//nolint:gocritic // Named results would shadow local variables
func (cmd *PublishUploadCmd) uploadBinary(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID, fileType string) (int64, string, string, error) {
	var versionCode int64
	var sha1, sha256 string

	if err := client.AcquireForUpload(ctx); err != nil {
		return 0, "", "", err
	}

	var err error
	if fileType == fileTypeAAB {
		err = cmd.uploadBundle(ctx, client, svc, packageName, editID, &versionCode, &sha1, &sha256)
	} else {
		err = cmd.uploadAPK(ctx, client, svc, packageName, editID, &versionCode, &sha1, &sha256)
	}

	client.ReleaseForUpload()

	if err != nil {
		return 0, "", "", errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to upload %s: %v", fileType, err))
	}

	return versionCode, sha1, sha256, nil
}

// uploadBundle uploads an AAB bundle.
func (cmd *PublishUploadCmd) uploadBundle(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID string, versionCode *int64, sha1, sha256 *string) error {
	return client.DoWithRetry(ctx, func() error {
		file, err := os.Open(cmd.File)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				_ = cerr
			}
		}()

		bundle, err := svc.Edits.Bundles.Upload(packageName, editID).Media(file).Context(ctx).Do()
		if err != nil {
			return err
		}
		if bundle != nil {
			*versionCode = bundle.VersionCode
			*sha1 = bundle.Sha1
			*sha256 = bundle.Sha256
		}
		return nil
	})
}

// uploadAPK uploads an APK file.
func (cmd *PublishUploadCmd) uploadAPK(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID string, versionCode *int64, sha1, sha256 *string) error {
	return client.DoWithRetry(ctx, func() error {
		file, err := os.Open(cmd.File)
		if err != nil {
			return err
		}
		defer func() {
			if cerr := file.Close(); cerr != nil {
				_ = cerr
			}
		}()

		apk, err := svc.Edits.Apks.Upload(packageName, editID).Media(file).Context(ctx).Do()
		if err != nil {
			return err
		}
		if apk != nil && apk.Binary != nil {
			*versionCode = apk.VersionCode
			*sha1 = apk.Binary.Sha1
			*sha256 = apk.Binary.Sha256
		}
		return nil
	})
}

// uploadExpansionFiles uploads expansion files if specified.
func (cmd *PublishUploadCmd) uploadExpansionFiles(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID string, versionCode int64) error {
	if cmd.ObbMain != "" {
		if err := cmd.uploadExpansionFile(ctx, client, svc, packageName, editID, versionCode, cmd.ObbMain, cmd.ObbMainRefVersion, "main"); err != nil {
			return err
		}
	}

	if cmd.ObbPatch != "" {
		if err := cmd.uploadExpansionFile(ctx, client, svc, packageName, editID, versionCode, cmd.ObbPatch, cmd.ObbPatchRefVersion, "patch"); err != nil {
			return err
		}
	}

	return nil
}

// commitUploadEdit commits the edit if auto-commit is enabled.
func (cmd *PublishUploadCmd) commitUploadEdit(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID string) (bool, error) {
	if cmd.NoAutoCommit {
		return false, nil
	}

	if err := client.Acquire(ctx); err != nil {
		return false, err
	}

	err := client.DoWithRetry(ctx, func() error {
		if cmd.InProgressReviewBehaviour != "" {
			_, cerr := svc.Edits.Commit(packageName, editID).Context(ctx).Do(googleapi.QueryParameter("inProgressReviewBehaviour", cmd.InProgressReviewBehaviour))
			return cerr
		}
		_, cerr := svc.Edits.Commit(packageName, editID).Context(ctx).Do()
		return cerr
	})

	client.Release()

	if err != nil {
		return false, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
			WithHint("The upload succeeded but the edit could not be committed. You may need to commit manually.")
	}

	return true, nil
}

// buildUploadResult builds and outputs the upload result.
func (cmd *PublishUploadCmd) buildUploadResult(start time.Time, fileInfo os.FileInfo, fileType, editID string, versionCode int64, sha1, sha256 string, committed bool, globals *Globals) error {
	result := output.NewResult(uploadResult{
		VersionCode: versionCode,
		SHA1:        sha1,
		SHA256:      sha256,
		Type:        fileType,
		EditID:      editID,
		Committed:   committed,
		File:        cmd.File,
		Size:        fileInfo.Size(),
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	if !committed && !cmd.NoAutoCommit {
		result = result.WithWarnings("Edit not committed due to error")
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

func (cmd *PublishUploadCmd) uploadExpansionFile(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID string, versionCode int64, filePath string, refVersion int64, fileType string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to open expansion file: %v", err))
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			_ = cerr
		}
	}()

	expansionFileType := "main"
	if fileType == "patch" {
		expansionFileType = "patch"
	}

	expansionFile := &androidpublisher.ExpansionFile{
		ReferencesVersion: refVersion,
	}

	return client.DoWithRetry(ctx, func() error {
		_, err := svc.Edits.Expansionfiles.Update(packageName, editID, versionCode, expansionFileType, expansionFile).Context(ctx).Do()
		return err
	})
}

// PublishReleaseCmd creates or updates a release.
type PublishReleaseCmd struct {
	Track                     string   `help:"Release track" default:"internal" enum:"internal,alpha,beta,production"`
	Name                      string   `help:"Release name"`
	Status                    string   `help:"Release status" default:"draft" enum:"draft,completed,halted,inProgress"`
	VersionCodes              []string `help:"Version codes to include (repeatable)"`
	RetainVersionCodes        []string `help:"Version codes to retain (repeatable)"`
	InAppUpdatePriority       int      `help:"In-app update priority (0-5)" default:"-1"`
	ReleaseNotesFile          string   `help:"JSON file with localized release notes" type:"existingfile"`
	EditID                    string   `help:"Explicit edit transaction ID"`
	NoAutoCommit              bool     `help:"Keep edit open for manual commit"`
	InProgressReviewBehaviour string   `help:"Behavior when committing while review in progress: THROW_ERROR_IF_IN_PROGRESS, CANCEL_IN_PROGRESS_AND_SUBMIT, or IN_PROGRESS_REVIEW_BEHAVIOUR_UNSPECIFIED" enum:"THROW_ERROR_IF_IN_PROGRESS,CANCEL_IN_PROGRESS_AND_SUBMIT,IN_PROGRESS_REVIEW_BEHAVIOUR_UNSPECIFIED," default:""`
	DryRun                    bool     `help:"Show intended actions without executing"`
	Wait                      bool     `help:"Wait for release to complete"`
	WaitTimeout               string   `help:"Maximum time to wait" default:"30m"`
}

// releaseResult represents the result of a release operation.
type releaseResult struct {
	Track        string  `json:"track"`
	Name         string  `json:"name,omitempty"`
	Status       string  `json:"status"`
	VersionCodes []int64 `json:"versionCodes"`
	EditID       string  `json:"editId"`
	Committed    bool    `json:"committed"`
}

// Run executes the release command.
func (cmd *PublishReleaseCmd) Run(globals *Globals) error {
	ctx := context.Background()
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	// Validate track and status
	if !api.IsValidTrack(cmd.Track) {
		return errors.ErrTrackInvalid
	}

	if !api.IsValidReleaseStatus(cmd.Status) {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid release status: %s", cmd.Status)).
			WithHint("Valid statuses are: draft, completed, halted, inProgress")
	}

	versionCodes, releaseNotes, err := cmd.parseReleaseInputs()
	if err != nil {
		return err
	}

	if cmd.DryRun {
		return cmd.handleDryRunRelease(start, versionCodes, globals)
	}

	client, svc, err := cmd.createReleaseClient(ctx, globals)
	if err != nil {
		return err
	}

	editID, err := cmd.getOrCreateReleaseEditID(ctx, client, svc, globals.Package)
	if err != nil {
		return err
	}

	track := cmd.buildTrack(versionCodes, releaseNotes)

	if err := cmd.updateReleaseTrack(ctx, client, svc, globals.Package, editID, track); err != nil {
		return err
	}

	committed, err := cmd.commitReleaseEdit(ctx, client, svc, globals.Package, editID)
	if err != nil {
		return err
	}

	return cmd.buildReleaseResult(start, editID, versionCodes, committed, globals)
}

// parseReleaseInputs parses version codes and loads release notes.
//
//nolint:gocritic // Named results would shadow local variables
func (cmd *PublishReleaseCmd) parseReleaseInputs() ([]int64, map[string]string, error) {
	versionCodes, err := cmd.parseVersionCodes()
	if err != nil {
		return nil, nil, err
	}

	releaseNotes, err := cmd.loadReleaseNotes()
	if err != nil {
		return nil, nil, err
	}

	return versionCodes, releaseNotes, nil
}

// parseVersionCodes parses version codes from command arguments.
func (cmd *PublishReleaseCmd) parseVersionCodes() ([]int64, error) {
	var versionCodes []int64
	for _, vc := range cmd.VersionCodes {
		code, err := strconv.ParseInt(vc, 10, 64)
		if err != nil {
			return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid version code: %s", vc))
		}
		versionCodes = append(versionCodes, code)
	}

	if len(versionCodes) == 0 {
		return nil, errors.NewAPIError(errors.CodeValidationError, "at least one version code is required").
			WithHint("Upload artifacts first, then specify their version codes with --version-codes")
	}

	return versionCodes, nil
}

// loadReleaseNotes loads release notes from file if specified.
func (cmd *PublishReleaseCmd) loadReleaseNotes() (map[string]string, error) {
	releaseNotes := make(map[string]string)
	if cmd.ReleaseNotesFile == "" {
		return releaseNotes, nil
	}

	data, err := os.ReadFile(cmd.ReleaseNotesFile)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to read release notes file: %v", err))
	}

	if err := json.Unmarshal(data, &releaseNotes); err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to parse release notes JSON: %v", err))
	}

	return releaseNotes, nil
}

// handleDryRunRelease handles dry-run mode for release command.
func (cmd *PublishReleaseCmd) handleDryRunRelease(start time.Time, versionCodes []int64, globals *Globals) error {
	result := output.NewResult(map[string]interface{}{
		"track":        cmd.Track,
		"status":       cmd.Status,
		"versionCodes": versionCodes,
		"dryRun":       true,
	}).WithDuration(time.Since(start)).
		WithNoOp("dry run - no release created")
	return outputResult(result, globals.Output, globals.Pretty)
}

// createReleaseClient creates API client and service for release.
func (cmd *PublishReleaseCmd) createReleaseClient(ctx context.Context, globals *Globals) (*api.Client, *androidpublisher.Service, error) {
	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return nil, nil, err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return nil, nil, errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	return client, svc, nil
}

// getOrCreateReleaseEditID gets existing or creates new edit ID for release.
func (cmd *PublishReleaseCmd) getOrCreateReleaseEditID(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName string) (string, error) {
	editID := cmd.EditID
	if editID != "" {
		return editID, nil
	}

	if err := client.Acquire(ctx); err != nil {
		return "", err
	}

	var edit *androidpublisher.AppEdit
	var err error
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(packageName, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})

	client.Release()

	if err != nil {
		return "", errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	return edit.Id, nil
}

// buildTrack builds the track object for release.
func (cmd *PublishReleaseCmd) buildTrack(versionCodes []int64, releaseNotes map[string]string) *androidpublisher.Track {
	release := &androidpublisher.TrackRelease{
		Name:         cmd.Name,
		Status:       cmd.Status,
		VersionCodes: versionCodes,
	}

	if cmd.InAppUpdatePriority >= 0 && cmd.InAppUpdatePriority <= 5 {
		release.InAppUpdatePriority = int64(cmd.InAppUpdatePriority)
	}

	if len(releaseNotes) > 0 {
		release.ReleaseNotes = cmd.buildLocalizedReleaseNotes(releaseNotes)
	}

	return &androidpublisher.Track{
		Track:    cmd.Track,
		Releases: []*androidpublisher.TrackRelease{release},
	}
}

// buildLocalizedReleaseNotes builds localized release notes array.
func (cmd *PublishReleaseCmd) buildLocalizedReleaseNotes(releaseNotes map[string]string) []*androidpublisher.LocalizedText {
	var localizedTexts []*androidpublisher.LocalizedText
	for locale, text := range releaseNotes {
		localizedTexts = append(localizedTexts, &androidpublisher.LocalizedText{
			Language: locale,
			Text:     text,
		})
	}
	return localizedTexts
}

// updateReleaseTrack updates the track with the release.
func (cmd *PublishReleaseCmd) updateReleaseTrack(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID string, track *androidpublisher.Track) error {
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	err := client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Tracks.Update(packageName, editID, cmd.Track, track).Context(ctx).Do()
		return uerr
	})

	client.Release()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("track not found: %s", cmd.Track))
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update track: %v", err))
	}

	return nil
}

// commitReleaseEdit commits the edit if auto-commit is enabled.
func (cmd *PublishReleaseCmd) commitReleaseEdit(ctx context.Context, client *api.Client, svc *androidpublisher.Service, packageName, editID string) (bool, error) {
	if cmd.NoAutoCommit {
		return false, nil
	}

	if err := client.Acquire(ctx); err != nil {
		return false, err
	}

	err := client.DoWithRetry(ctx, func() error {
		if cmd.InProgressReviewBehaviour != "" {
			_, cerr := svc.Edits.Commit(packageName, editID).Context(ctx).Do(googleapi.QueryParameter("inProgressReviewBehaviour", cmd.InProgressReviewBehaviour))
			return cerr
		}
		_, cerr := svc.Edits.Commit(packageName, editID).Context(ctx).Do()
		return cerr
	})

	client.Release()

	if err != nil {
		return false, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
			WithHint("The release was created but the edit could not be committed. You may need to commit manually.")
	}

	return true, nil
}

// buildReleaseResult builds and outputs the release result.
func (cmd *PublishReleaseCmd) buildReleaseResult(start time.Time, editID string, versionCodes []int64, committed bool, globals *Globals) error {
	result := output.NewResult(releaseResult{
		Track:        cmd.Track,
		Name:         cmd.Name,
		Status:       cmd.Status,
		VersionCodes: versionCodes,
		EditID:       editID,
		Committed:    committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishRolloutCmd updates rollout percentage.
type PublishRolloutCmd struct {
	Track        string  `help:"Release track" default:"production" enum:"internal,alpha,beta,production"`
	Percentage   float64 `help:"Rollout percentage (0.01-100.00)"`
	EditID       string  `help:"Explicit edit transaction ID"`
	NoAutoCommit bool    `help:"Keep edit open for manual commit"`
	DryRun       bool    `help:"Show intended actions without executing"`
}

// Run executes the rollout command.
func (cmd *PublishRolloutCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.Percentage < 0.01 || cmd.Percentage > 100.0 {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid rollout percentage: %.2f", cmd.Percentage)).
			WithHint("Percentage must be between 0.01 and 100.00")
	}

	userFraction := cmd.Percentage / 100.0

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"track":        cmd.Track,
			"percentage":   cmd.Percentage,
			"userFraction": userFraction,
			"dryRun":       true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - rollout not updated")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get current track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track %s: %v", cmd.Track, err))
	}

	// Find inProgress release and update userFraction
	found := false
	for _, release := range track.Releases {
		if release.Status == statusInProgress {
			release.UserFraction = userFraction
			found = true
			break
		}
	}
	if !found {
		return errors.NewAPIError(errors.CodeNotFound, "no in-progress release found on track").
			WithHint("Only releases with status 'inProgress' can have their rollout percentage updated")
	}

	// Update track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	err = client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Tracks.Update(pkg, editID, cmd.Track, track).Context(ctx).Do()
		return uerr
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update rollout: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The rollout was updated but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"track":        cmd.Track,
		"percentage":   cmd.Percentage,
		"userFraction": userFraction,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishPromoteCmd promotes a release between tracks.
type PublishPromoteCmd struct {
	FromTrack    string  `help:"Source track"`
	ToTrack      string  `help:"Destination track"`
	Percentage   float64 `help:"Rollout percentage for destination" default:"0"`
	EditID       string  `help:"Explicit edit transaction ID"`
	NoAutoCommit bool    `help:"Keep edit open for manual commit"`
	DryRun       bool    `help:"Show intended actions without executing"`
}

// Run executes the promote command.
func (cmd *PublishPromoteCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.FromTrack == "" || cmd.ToTrack == "" {
		return errors.NewAPIError(errors.CodeValidationError, "both --from-track and --to-track are required").
			WithHint("Specify source and destination tracks, e.g., --from-track=beta --to-track=production")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"fromTrack":  cmd.FromTrack,
			"toTrack":    cmd.ToTrack,
			"percentage": cmd.Percentage,
			"dryRun":     true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - release not promoted")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get source track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var sourceTrack *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		sourceTrack, err = svc.Edits.Tracks.Get(pkg, editID, cmd.FromTrack).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get source track %s: %v", cmd.FromTrack, err))
	}

	if len(sourceTrack.Releases) == 0 {
		return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("no releases found on source track %s", cmd.FromTrack))
	}

	// Get the latest release from the source track
	latestRelease := sourceTrack.Releases[0]
	for _, r := range sourceTrack.Releases {
		if r.Status == releaseCompleted || r.Status == statusInProgress {
			latestRelease = r
			break
		}
	}

	// Build the target release
	targetRelease := &androidpublisher.TrackRelease{
		Name:         latestRelease.Name,
		VersionCodes: latestRelease.VersionCodes,
		ReleaseNotes: latestRelease.ReleaseNotes,
	}

	if cmd.Percentage > 0 && cmd.Percentage < 100 {
		targetRelease.Status = statusInProgress
		targetRelease.UserFraction = cmd.Percentage / 100.0
	} else {
		targetRelease.Status = releaseCompleted
	}

	targetTrack := &androidpublisher.Track{
		Track:    cmd.ToTrack,
		Releases: []*androidpublisher.TrackRelease{targetRelease},
	}

	// Update target track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	err = client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Tracks.Update(pkg, editID, cmd.ToTrack, targetTrack).Context(ctx).Do()
		return uerr
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update target track %s: %v", cmd.ToTrack, err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The promotion was configured but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"fromTrack":    cmd.FromTrack,
		"toTrack":      cmd.ToTrack,
		"versionCodes": latestRelease.VersionCodes,
		"status":       targetRelease.Status,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishHaltCmd halts a production rollout.
type PublishHaltCmd struct {
	Track        string `help:"Release track" default:"production" enum:"internal,alpha,beta,production"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the halt command.
func (cmd *PublishHaltCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "halt requires confirmation").
			WithHint("Use --confirm to confirm this destructive operation")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"track":  cmd.Track,
			"action": "halt",
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - rollout not halted")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get current track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track %s: %v", cmd.Track, err))
	}

	// Find inProgress release and set to halted
	found := false
	var haltedVersionCodes []int64
	for _, release := range track.Releases {
		if release.Status == statusInProgress {
			release.Status = statusHalted
			haltedVersionCodes = release.VersionCodes
			found = true
			break
		}
	}
	if !found {
		return errors.NewAPIError(errors.CodeNotFound, "no in-progress release found on track").
			WithHint("Only releases with status 'inProgress' can be halted")
	}

	// Update track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	err = client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Tracks.Update(pkg, editID, cmd.Track, track).Context(ctx).Do()
		return uerr
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to halt rollout: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The halt was applied but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"track":        cmd.Track,
		"action":       statusHalted,
		"versionCodes": haltedVersionCodes,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishRollbackCmd rolls back to a previous version.
type PublishRollbackCmd struct {
	Track        string `help:"Release track"`
	VersionCode  string `help:"Specific version code to rollback to"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the rollback command.
func (cmd *PublishRollbackCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.Track == "" {
		return errors.NewAPIError(errors.CodeValidationError, "track is required for rollback").
			WithHint("Specify the track with --track, e.g., --track=production")
	}

	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "rollback requires confirmation").
			WithHint("Use --confirm to confirm this destructive operation")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"track":  cmd.Track,
			"action": "rollback",
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - rollback not executed")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get current track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var track *androidpublisher.Track
	err = client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track %s: %v", cmd.Track, err))
	}

	// Find the inProgress release and halt it to trigger rollback
	found := false
	var rolledBackVersionCodes []int64
	for _, release := range track.Releases {
		if release.Status == statusInProgress {
			release.Status = statusHalted
			rolledBackVersionCodes = release.VersionCodes
			found = true
			break
		}
	}
	if !found {
		return errors.NewAPIError(errors.CodeNotFound, "no in-progress release found on track to rollback").
			WithHint("Rollback halts the current in-progress release so the previous version resumes serving")
	}

	// Update track
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	err = client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Tracks.Update(pkg, editID, cmd.Track, track).Context(ctx).Do()
		return uerr
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to rollback: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The rollback was applied but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"track":                  cmd.Track,
		"action":                 "rollback",
		"rolledBackVersionCodes": rolledBackVersionCodes,
		"editId":                 editID,
		"committed":              committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishStatusCmd gets track status.
type PublishStatusCmd struct {
	Track string `help:"Release track (leave empty for all tracks)"`
}

// Run executes the status command.
func (cmd *PublishStatusCmd) Run(globals *Globals) error {
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

	// Create a temporary edit
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

	type releaseInfo struct {
		Name         string  `json:"name,omitempty"`
		Status       string  `json:"status"`
		VersionCodes []int64 `json:"versionCodes,omitempty"`
		UserFraction float64 `json:"userFraction,omitempty"`
	}

	type trackStatus struct {
		Track    string        `json:"track"`
		Releases []releaseInfo `json:"releases,omitempty"`
	}

	var data interface{}

	if cmd.Track != "" {
		// Get specific track
		var track *androidpublisher.Track
		err = client.DoWithRetry(ctx, func() error {
			track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Track).Context(ctx).Do()
			return err
		})

		// Clean up edit
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()

		if err != nil {
			if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
				return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("track not found: %s", cmd.Track))
			}
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get track: %v", err))
		}

		ts := trackStatus{Track: track.Track}
		for _, r := range track.Releases {
			ri := releaseInfo{
				Name:         r.Name,
				Status:       r.Status,
				VersionCodes: r.VersionCodes,
				UserFraction: r.UserFraction,
			}
			ts.Releases = append(ts.Releases, ri)
		}
		data = ts
	} else {
		// List all tracks
		var tracksList *androidpublisher.TracksListResponse
		err = client.DoWithRetry(ctx, func() error {
			tracksList, err = svc.Edits.Tracks.List(pkg, editID).Context(ctx).Do()
			return err
		})

		// Clean up edit
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()

		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
		}

		var statuses []trackStatus
		for _, t := range tracksList.Tracks {
			ts := trackStatus{Track: t.Track}
			for _, r := range t.Releases {
				ri := releaseInfo{
					Name:         r.Name,
					Status:       r.Status,
					VersionCodes: r.VersionCodes,
					UserFraction: r.UserFraction,
				}
				ts.Releases = append(ts.Releases, ri)
			}
			statuses = append(statuses, ts)
		}
		data = map[string]interface{}{
			"tracks": statuses,
			"count":  len(statuses),
		}
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishTracksCmd lists all tracks.
type PublishTracksCmd struct{}

// trackInfo represents simplified track information.
type trackInfo struct {
	Track    string   `json:"track"`
	Releases []string `json:"releases,omitempty"`
}

// tracksListResult represents the result of listing tracks.
type tracksListResult struct {
	Tracks []trackInfo `json:"tracks"`
	Count  int         `json:"count"`
}

// Run executes the tracks command.
func (cmd *PublishTracksCmd) Run(globals *Globals) error {
	ctx := context.Background()
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	// Create API client
	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}

	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service").
			WithHint("Ensure authentication is configured correctly")
	}

	// Create a temporary edit to fetch tracks
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(globals.Package, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})

	if err != nil {
		client.Release()
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}

	editID := edit.Id

	// List tracks
	var tracksList *androidpublisher.TracksListResponse
	err = client.DoWithRetry(ctx, func() error {
		tracksList, err = svc.Edits.Tracks.List(globals.Package, editID).Context(ctx).Do()
		return err
	})

	// Clean up - delete the temporary edit
	// Note: Edits are automatically cleaned up if not committed, but we'll be explicit
	_ = client.DoWithRetry(ctx, func() error {
		derr := svc.Edits.Delete(globals.Package, editID).Context(ctx).Do()
		return derr
	})

	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
	}

	// Convert to output format
	var tracks []trackInfo
	for _, t := range tracksList.Tracks {
		info := trackInfo{
			Track: t.Track,
		}

		// Extract release names
		for _, release := range t.Releases {
			if release.Name != "" {
				info.Releases = append(info.Releases, release.Name)
			} else if release.Status != "" {
				info.Releases = append(info.Releases, fmt.Sprintf("%s (status: %s)", release.Name, release.Status))
			}
		}

		tracks = append(tracks, info)
	}

	result := output.NewResult(tracksListResult{
		Tracks: tracks,
		Count:  len(tracks),
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishCapabilitiesCmd lists publishing capabilities.
type PublishCapabilitiesCmd struct{}

// Run executes the capabilities command.
func (cmd *PublishCapabilitiesCmd) Run(globals *Globals) error {
	start := time.Now()

	capabilities := map[string]interface{}{
		"tracks": []string{"internal", "alpha", "beta", "production"},
		"releaseStatuses": []string{
			releaseStatusDraft, releaseCompleted, statusHalted, statusInProgress,
		},
		"imageTypes": []string{
			"icon", "featureGraphic", "promoGraphic",
			"phoneScreenshots", "sevenInchScreenshots",
			"tenInchScreenshots", "tvScreenshots",
			"tvBanner", "wearScreenshots",
		},
		"uploadFormats": []string{"apk", "aab"},
		"deobfuscationTypes": []string{
			"proguard", "nativeCode",
		},
		"expansionFileTypes":    []string{"main", "patch"},
		"maxScreenshotsPerType": 8,
		"maxImageSizes": map[string]string{
			"icon":                 "512x512 PNG (32-bit with alpha)",
			"featureGraphic":       "1024x500 JPEG or 24-bit PNG (no alpha)",
			"promoGraphic":         "180x120 JPEG or 24-bit PNG (no alpha)",
			"phoneScreenshots":     "min 320px, max 3840px, aspect ratio 16:9 or 9:16",
			"sevenInchScreenshots": "min 320px, max 3840px",
			"tenInchScreenshots":   "min 320px, max 3840px",
			"tvScreenshots":        "1280x720 or 1920x1080",
			"tvBanner":             "1280x720",
			"wearScreenshots":      "min 384px, max 3840px",
		},
		"maxExpansionFileSize": "2GB",
		"maxApkSize":           "150MB",
		"maxBundleSize":        "150MB",
		"inAppUpdatePriority":  "0-5",
	}

	result := output.NewResult(capabilities).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishListingCmd manages store listing.
type PublishListingCmd struct {
	Update PublishListingUpdateCmd `cmd:"" help:"Update store listing"`
	Get    PublishListingGetCmd    `cmd:"" help:"Get store listing"`
	Delete PublishListingDeleteCmd `cmd:"" help:"Delete store listing"`
}

// PublishListingUpdateCmd updates store listing.
type PublishListingUpdateCmd struct {
	Locale       string `help:"Locale code" default:"en-US"`
	Title        string `help:"App title"`
	ShortDesc    string `help:"Short description"`
	FullDesc     string `help:"Full description"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the listing update command.
func (cmd *PublishListingUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"locale":    cmd.Locale,
			"title":     cmd.Title,
			"shortDesc": cmd.ShortDesc,
			"fullDesc":  cmd.FullDesc,
			"dryRun":    true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - listing not updated")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	listing := &androidpublisher.Listing{
		Title:            cmd.Title,
		ShortDescription: cmd.ShortDesc,
		FullDescription:  cmd.FullDesc,
	}

	// Update listing
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updatedListing *androidpublisher.Listing
	err = client.DoWithRetry(ctx, func() error {
		updatedListing, err = svc.Edits.Listings.Update(pkg, editID, cmd.Locale, listing).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update listing: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The listing was updated but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"locale":           cmd.Locale,
		"title":            updatedListing.Title,
		"shortDescription": updatedListing.ShortDescription,
		"fullDescription":  updatedListing.FullDescription,
		"language":         updatedListing.Language,
		"editId":           editID,
		"committed":        committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishListingGetCmd gets store listing.
type PublishListingGetCmd struct {
	Locale string `help:"Locale code (leave empty for all)"`
}

// Run executes the listing get command.
func (cmd *PublishListingGetCmd) Run(globals *Globals) error {
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

	// Create a temporary edit
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
	var data interface{}

	if cmd.Locale != "" {
		// Get specific locale listing
		var listing *androidpublisher.Listing
		err = client.DoWithRetry(ctx, func() error {
			listing, err = svc.Edits.Listings.Get(pkg, editID, cmd.Locale).Context(ctx).Do()
			return err
		})

		// Clean up edit
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()

		if err != nil {
			if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
				return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("listing not found for locale: %s", cmd.Locale))
			}
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get listing: %v", err))
		}

		data = map[string]interface{}{
			"language":         listing.Language,
			"title":            listing.Title,
			"shortDescription": listing.ShortDescription,
			"fullDescription":  listing.FullDescription,
		}
	} else {
		// List all listings
		var listingsResp *androidpublisher.ListingsListResponse
		err = client.DoWithRetry(ctx, func() error {
			listingsResp, err = svc.Edits.Listings.List(pkg, editID).Context(ctx).Do()
			return err
		})

		// Clean up edit
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
		client.Release()

		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list listings: %v", err))
		}

		var listings []map[string]interface{}
		for _, l := range listingsResp.Listings {
			listings = append(listings, map[string]interface{}{
				"language":         l.Language,
				"title":            l.Title,
				"shortDescription": l.ShortDescription,
				"fullDescription":  l.FullDescription,
			})
		}
		data = map[string]interface{}{
			"listings": listings,
			"count":    len(listings),
		}
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishListingDeleteCmd deletes store listing.
type PublishListingDeleteCmd struct {
	Locale       string `help:"Locale code (required)"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
	All          bool   `help:"Delete all store listings"`
}

// Run executes the listing delete command.
func (cmd *PublishListingDeleteCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if !cmd.All && cmd.Locale == "" {
		return errors.NewAPIError(errors.CodeValidationError, "locale is required").
			WithHint("Specify --locale for the listing to delete, or use --all to delete all listings")
	}

	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "delete requires confirmation").
			WithHint("Use --confirm to confirm this destructive operation")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"locale": cmd.Locale,
			"all":    cmd.All,
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - listing not deleted")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Delete listing(s)
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	if cmd.All {
		err = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Listings.Deleteall(pkg, editID).Context(ctx).Do()
		})
	} else {
		err = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Listings.Delete(pkg, editID, cmd.Locale).Context(ctx).Do()
		})
	}
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete listing: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The listing was deleted but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"locale":    cmd.Locale,
		"all":       cmd.All,
		"deleted":   true,
		"editId":    editID,
		"committed": committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishDetailsCmd manages app details.
type PublishDetailsCmd struct {
	Get    PublishDetailsGetCmd    `cmd:"" help:"Get app details"`
	Update PublishDetailsUpdateCmd `cmd:"" help:"Update app details"`
	Patch  PublishDetailsPatchCmd  `cmd:"" help:"Patch app details"`
}

// PublishDetailsGetCmd gets app details.
type PublishDetailsGetCmd struct{}

// Run executes the details get command.
func (cmd *PublishDetailsGetCmd) Run(globals *Globals) error {
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

	// Create a temporary edit
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

	// Get details
	var details *androidpublisher.AppDetails
	err = client.DoWithRetry(ctx, func() error {
		details, err = svc.Edits.Details.Get(pkg, editID).Context(ctx).Do()
		return err
	})

	// Clean up edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	})
	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get app details: %v", err))
	}

	data := map[string]interface{}{
		"contactEmail":    details.ContactEmail,
		"contactPhone":    details.ContactPhone,
		"contactWebsite":  details.ContactWebsite,
		"defaultLanguage": details.DefaultLanguage,
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishDetailsUpdateCmd updates app details.
type PublishDetailsUpdateCmd struct {
	ContactEmail    string `help:"Contact email"`
	ContactPhone    string `help:"Contact phone"`
	ContactWebsite  string `help:"Contact website"`
	DefaultLanguage string `help:"Default language"`
	EditID          string `help:"Explicit edit transaction ID"`
	NoAutoCommit    bool   `help:"Keep edit open for manual commit"`
	DryRun          bool   `help:"Show intended actions without executing"`
}

// Run executes the details update command.
func (cmd *PublishDetailsUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"contactEmail":    cmd.ContactEmail,
			"contactPhone":    cmd.ContactPhone,
			"contactWebsite":  cmd.ContactWebsite,
			"defaultLanguage": cmd.DefaultLanguage,
			"dryRun":          true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - details not updated")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	details := &androidpublisher.AppDetails{
		ContactEmail:    cmd.ContactEmail,
		ContactPhone:    cmd.ContactPhone,
		ContactWebsite:  cmd.ContactWebsite,
		DefaultLanguage: cmd.DefaultLanguage,
	}

	// Update details
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updatedDetails *androidpublisher.AppDetails
	err = client.DoWithRetry(ctx, func() error {
		updatedDetails, err = svc.Edits.Details.Update(pkg, editID, details).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update details: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The details were updated but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"contactEmail":    updatedDetails.ContactEmail,
		"contactPhone":    updatedDetails.ContactPhone,
		"contactWebsite":  updatedDetails.ContactWebsite,
		"defaultLanguage": updatedDetails.DefaultLanguage,
		"editId":          editID,
		"committed":       committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishDetailsPatchCmd patches app details.
type PublishDetailsPatchCmd struct {
	ContactEmail    string `help:"Contact email"`
	ContactPhone    string `help:"Contact phone"`
	ContactWebsite  string `help:"Contact website"`
	DefaultLanguage string `help:"Default language"`
	UpdateMask      string `help:"Fields to update (comma-separated)"`
	EditID          string `help:"Explicit edit transaction ID"`
	NoAutoCommit    bool   `help:"Keep edit open for manual commit"`
	DryRun          bool   `help:"Show intended actions without executing"`
}

// Run executes the details patch command.
func (cmd *PublishDetailsPatchCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"contactEmail":    cmd.ContactEmail,
			"contactPhone":    cmd.ContactPhone,
			"contactWebsite":  cmd.ContactWebsite,
			"defaultLanguage": cmd.DefaultLanguage,
			"updateMask":      cmd.UpdateMask,
			"dryRun":          true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - details not patched")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	details := &androidpublisher.AppDetails{
		ContactEmail:    cmd.ContactEmail,
		ContactPhone:    cmd.ContactPhone,
		ContactWebsite:  cmd.ContactWebsite,
		DefaultLanguage: cmd.DefaultLanguage,
	}

	// Patch details
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var patchedDetails *androidpublisher.AppDetails
	err = client.DoWithRetry(ctx, func() error {
		patchedDetails, err = svc.Edits.Details.Patch(pkg, editID, details).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to patch details: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The details were patched but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"contactEmail":    patchedDetails.ContactEmail,
		"contactPhone":    patchedDetails.ContactPhone,
		"contactWebsite":  patchedDetails.ContactWebsite,
		"defaultLanguage": patchedDetails.DefaultLanguage,
		"editId":          editID,
		"committed":       committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishImagesCmd manages store images.
type PublishImagesCmd struct {
	Upload    PublishImagesUploadCmd    `cmd:"" help:"Upload an image"`
	List      PublishImagesListCmd      `cmd:"" help:"List images"`
	Delete    PublishImagesDeleteCmd    `cmd:"" help:"Delete an image"`
	DeleteAll PublishImagesDeleteAllCmd `cmd:"" help:"Delete all images for type"`
}

// PublishImagesUploadCmd uploads an image.
type PublishImagesUploadCmd struct {
	Type         string `arg:"" help:"Image type (icon, featureGraphic, phoneScreenshots, etc.)"`
	File         string `arg:"" help:"Image file path" type:"existingfile"`
	Locale       string `help:"Locale code" default:"en-US"`
	SyncImages   bool   `help:"Skip upload if identical image already exists"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the images upload command.
func (cmd *PublishImagesUploadCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.File == "" {
		return errors.NewAPIError(errors.CodeValidationError, "file is required").
			WithHint("Provide an image file to upload")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"type":   cmd.Type,
			"file":   cmd.File,
			"locale": cmd.Locale,
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - image not uploaded")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Open file
	file, err := os.Open(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to open image file: %v", err))
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			_ = cerr
		}
	}()

	// Upload image
	if err := client.AcquireForUpload(ctx); err != nil {
		return err
	}
	var uploadResp *androidpublisher.ImagesUploadResponse
	err = client.DoWithRetry(ctx, func() error {
		uploadResp, err = svc.Edits.Images.Upload(pkg, editID, cmd.Locale, cmd.Type).Media(file).Context(ctx).Do()
		return err
	})
	client.ReleaseForUpload()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to upload image: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The image was uploaded but the edit could not be committed")
		}
		committed = true
	}

	data := map[string]interface{}{
		"type":      cmd.Type,
		"locale":    cmd.Locale,
		"file":      cmd.File,
		"editId":    editID,
		"committed": committed,
	}
	if uploadResp != nil && uploadResp.Image != nil {
		data["imageId"] = uploadResp.Image.Id
		data["sha1"] = uploadResp.Image.Sha1
		data["sha256"] = uploadResp.Image.Sha256
		data["url"] = uploadResp.Image.Url
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishImagesListCmd lists images.
type PublishImagesListCmd struct {
	Type   string `arg:"" help:"Image type (icon, featureGraphic, phoneScreenshots, etc.)"`
	Locale string `help:"Locale code" default:"en-US"`
	EditID string `help:"Explicit edit transaction ID"`
}

// Run executes the images list command.
func (cmd *PublishImagesListCmd) Run(globals *Globals) error {
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

	// Create or reuse a temporary edit
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	editID := cmd.EditID
	createdEdit := false
	if editID == "" {
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		if err != nil {
			client.Release()
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
		createdEdit = true
	}

	// List images
	var imagesResp *androidpublisher.ImagesListResponse
	err = client.DoWithRetry(ctx, func() error {
		imagesResp, err = svc.Edits.Images.List(pkg, editID, cmd.Locale, cmd.Type).Context(ctx).Do()
		return err
	})

	// Clean up temporary edit
	if createdEdit {
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
	}
	client.Release()

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list images: %v", err))
	}

	var images []map[string]interface{}
	if imagesResp != nil {
		for _, img := range imagesResp.Images {
			images = append(images, map[string]interface{}{
				"id":     img.Id,
				"url":    img.Url,
				"sha1":   img.Sha1,
				"sha256": img.Sha256,
			})
		}
	}

	data := map[string]interface{}{
		"type":   cmd.Type,
		"locale": cmd.Locale,
		"images": images,
		"count":  len(images),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishImagesDeleteCmd deletes an image.
type PublishImagesDeleteCmd struct {
	Type         string `arg:"" help:"Image type"`
	ID           string `arg:"" help:"Image ID to delete"`
	Locale       string `help:"Locale code" default:"en-US"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the images delete command.
func (cmd *PublishImagesDeleteCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"type":    cmd.Type,
			"imageId": cmd.ID,
			"locale":  cmd.Locale,
			"dryRun":  true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - image not deleted")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Delete image
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	err = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Images.Delete(pkg, editID, cmd.Locale, cmd.Type, cmd.ID).Context(ctx).Do()
	})
	client.Release()
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("image not found: %s", cmd.ID))
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete image: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The image was deleted but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"type":      cmd.Type,
		"imageId":   cmd.ID,
		"locale":    cmd.Locale,
		"deleted":   true,
		"editId":    editID,
		"committed": committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishImagesDeleteAllCmd deletes all images for type.
type PublishImagesDeleteAllCmd struct {
	Type         string `arg:"" help:"Image type"`
	Locale       string `help:"Locale code" default:"en-US"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the images deleteall command.
func (cmd *PublishImagesDeleteAllCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"type":   cmd.Type,
			"locale": cmd.Locale,
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - images not deleted")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Delete all images for type
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var deleteResp *androidpublisher.ImagesDeleteAllResponse
	err = client.DoWithRetry(ctx, func() error {
		deleteResp, err = svc.Edits.Images.Deleteall(pkg, editID, cmd.Locale, cmd.Type).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete all images: %v", err))
	}

	deletedCount := 0
	if deleteResp != nil && deleteResp.Deleted != nil {
		deletedCount = len(deleteResp.Deleted)
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("Images were deleted but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"type":         cmd.Type,
		"locale":       cmd.Locale,
		"deletedCount": deletedCount,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishAssetsCmd manages store assets.
type PublishAssetsCmd struct {
	Upload PublishAssetsUploadCmd `cmd:"" help:"Upload assets from directory"`
	Spec   PublishAssetsSpecCmd   `cmd:"" help:"Output asset validation matrix"`
}

// PublishAssetsUploadCmd uploads assets from directory.
type PublishAssetsUploadCmd struct {
	Dir          string `arg:"" help:"Assets directory" default:"assets"`
	Category     string `help:"Category to replace (phone, tablet, tv, wear)"`
	ReplaceAll   bool   `help:"Replace all existing assets"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the assets upload command.
func (cmd *PublishAssetsUploadCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	// Map category to image types
	categoryImageTypes := map[string][]string{
		"phone":  {"phoneScreenshots", "icon", "featureGraphic", "promoGraphic"},
		"tablet": {"sevenInchScreenshots", "tenInchScreenshots"},
		"tv":     {"tvScreenshots", "tvBanner"},
		"wear":   {"wearScreenshots"},
	}

	if cmd.Category != "" {
		if _, ok := categoryImageTypes[cmd.Category]; !ok {
			return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid category: %s", cmd.Category)).
				WithHint("Valid categories are: phone, tablet, tv, wear")
		}
	}

	// Check directory exists
	dirInfo, err := os.Stat(cmd.Dir)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to access directory: %v", err))
	}
	if !dirInfo.IsDir() {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("not a directory: %s", cmd.Dir))
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"dir":        cmd.Dir,
			"category":   cmd.Category,
			"replaceAll": cmd.ReplaceAll,
			"dryRun":     true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - assets not uploaded")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Scan for image files in directory and upload them
	uploadedCount, uploadErrors, err := cmd.scanAndUploadImages(ctx, client, svc, pkg, editID)
	if err != nil {
		return err
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit && uploadedCount > 0 {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("Assets were uploaded but the edit could not be committed")
		}
		committed = true
	}

	resultData := map[string]interface{}{
		"dir":           cmd.Dir,
		"uploadedCount": uploadedCount,
		"editId":        editID,
		"committed":     committed,
	}

	result := output.NewResult(resultData).WithDuration(time.Since(start)).WithServices("androidpublisher")
	if len(uploadErrors) > 0 {
		result = result.WithWarnings(uploadErrors...)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// scanAndUploadImages walks the asset directory and uploads image files.
//
//nolint:gocritic // Named results would shadow local variables
func (cmd *PublishAssetsUploadCmd) scanAndUploadImages(ctx context.Context, client *api.Client, svc *androidpublisher.Service, pkg, editID string) (int, []string, error) {
	// Expected structure: {dir}/{imageType}/{locale}/*.png or {dir}/{locale}/{imageType}/*.png
	uploadedCount := 0
	var uploadErrors []string

	entries, err := os.ReadDir(cmd.Dir)
	if err != nil {
		return 0, nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to read directory: %v", err))
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		imageType := entry.Name()

		subEntries, subErr := os.ReadDir(filepath.Join(cmd.Dir, imageType))
		if subErr != nil {
			continue
		}

		for _, subEntry := range subEntries {
			filePath := filepath.Join(cmd.Dir, imageType, subEntry.Name())
			if subEntry.IsDir() {
				continue
			}

			ext := strings.ToLower(filepath.Ext(subEntry.Name()))
			if ext != ".png" && ext != ".jpg" && ext != ".jpeg" {
				continue
			}

			imgFile, oerr := os.Open(filePath)
			if oerr != nil {
				uploadErrors = append(uploadErrors, fmt.Sprintf("failed to open %s: %v", filePath, oerr))
				continue
			}

			if uerr := client.AcquireForUpload(ctx); uerr != nil {
				_ = imgFile.Close()
				uploadErrors = append(uploadErrors, fmt.Sprintf("failed to acquire upload lock: %v", uerr))
				continue
			}

			locale := "en-US" // default locale
			uerr := client.DoWithRetry(ctx, func() error {
				_, ierr := svc.Edits.Images.Upload(pkg, editID, locale, imageType).Media(imgFile).Context(ctx).Do()
				return ierr
			})
			client.ReleaseForUpload()
			_ = imgFile.Close()

			if uerr != nil {
				uploadErrors = append(uploadErrors, fmt.Sprintf("failed to upload %s: %v", filePath, uerr))
			} else {
				uploadedCount++
			}
		}
	}

	return uploadedCount, uploadErrors, nil
}

// PublishAssetsSpecCmd outputs asset validation matrix.
type PublishAssetsSpecCmd struct{}

// Run executes the assets spec command.
func (cmd *PublishAssetsSpecCmd) Run(globals *Globals) error {
	start := time.Now()

	spec := map[string]interface{}{
		"imageTypes": map[string]interface{}{
			"icon": map[string]interface{}{
				"dimensions": "512x512",
				"format":     "32-bit PNG with alpha",
				"maxSize":    "1MB",
				"maxCount":   1,
			},
			"featureGraphic": map[string]interface{}{
				"dimensions": "1024x500",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "1MB",
				"maxCount":   1,
			},
			"promoGraphic": map[string]interface{}{
				"dimensions": "180x120",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "1MB",
				"maxCount":   1,
			},
			"phoneScreenshots": map[string]interface{}{
				"dimensions": "min 320px, max 3840px per side; 16:9 or 9:16 aspect ratio",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "8MB per image",
				"maxCount":   8,
			},
			"sevenInchScreenshots": map[string]interface{}{
				"dimensions": "min 320px, max 3840px per side",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "8MB per image",
				"maxCount":   8,
			},
			"tenInchScreenshots": map[string]interface{}{
				"dimensions": "min 320px, max 3840px per side",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "8MB per image",
				"maxCount":   8,
			},
			"tvScreenshots": map[string]interface{}{
				"dimensions": "1280x720 or 1920x1080",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "8MB per image",
				"maxCount":   8,
			},
			"tvBanner": map[string]interface{}{
				"dimensions": "1280x720",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "1MB",
				"maxCount":   1,
			},
			"wearScreenshots": map[string]interface{}{
				"dimensions": "min 384px, max 3840px per side",
				"format":     "JPEG or 24-bit PNG (no alpha)",
				"maxSize":    "8MB per image",
				"maxCount":   8,
			},
		},
		"expansionFiles": map[string]interface{}{
			"main": map[string]interface{}{
				"maxSize": "2GB",
				"format":  "OBB (opaque binary blob)",
			},
			"patch": map[string]interface{}{
				"maxSize": "2GB",
				"format":  "OBB (opaque binary blob)",
			},
		},
		"artifacts": map[string]interface{}{
			"apk": map[string]interface{}{
				"maxSize": "150MB",
			},
			"aab": map[string]interface{}{
				"maxSize": "150MB",
			},
		},
	}

	result := output.NewResult(spec).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishDeobfuscationCmd manages deobfuscation files.
type PublishDeobfuscationCmd struct {
	Upload PublishDeobfuscationUploadCmd `cmd:"" help:"Upload deobfuscation file"`
}

// PublishDeobfuscationUploadCmd uploads deobfuscation file.
type PublishDeobfuscationUploadCmd struct {
	File         string `arg:"" help:"File to upload" type:"existingfile"`
	Type         string `help:"Deobfuscation file type: proguard or nativeCode" required:"" enum:"proguard,nativeCode"`
	VersionCode  int64  `help:"Version code to associate"`
	EditID       string `help:"Explicit edit transaction ID"`
	ChunkSize    int64  `help:"Upload chunk size in bytes" default:"10485760"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the deobfuscation upload command.
func (cmd *PublishDeobfuscationUploadCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.VersionCode <= 0 {
		return errors.NewAPIError(errors.CodeValidationError, "version code is required").
			WithHint("Specify the version code with --version-code")
	}

	if cmd.File == "" {
		return errors.NewAPIError(errors.CodeValidationError, "file is required").
			WithHint("Provide a deobfuscation file (mapping.txt or native debug symbols)")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"file":        cmd.File,
			"type":        cmd.Type,
			"versionCode": cmd.VersionCode,
			"dryRun":      true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - deobfuscation file not uploaded")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Open file
	file, err := os.Open(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to open deobfuscation file: %v", err))
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			_ = cerr
		}
	}()

	// Upload deobfuscation file
	if err := client.AcquireForUpload(ctx); err != nil {
		return err
	}
	var uploadResp *androidpublisher.DeobfuscationFilesUploadResponse
	err = client.DoWithRetry(ctx, func() error {
		uploadResp, err = svc.Edits.Deobfuscationfiles.Upload(pkg, editID, cmd.VersionCode, cmd.Type).Media(file).Context(ctx).Do()
		return err
	})
	client.ReleaseForUpload()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to upload deobfuscation file: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The deobfuscation file was uploaded but the edit could not be committed")
		}
		committed = true
	}

	data := map[string]interface{}{
		"file":        cmd.File,
		"type":        cmd.Type,
		"versionCode": cmd.VersionCode,
		"editId":      editID,
		"committed":   committed,
	}
	if uploadResp != nil && uploadResp.DeobfuscationFile != nil {
		data["symbolType"] = uploadResp.DeobfuscationFile.SymbolType
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishTestersCmd manages testers.
type PublishTestersCmd struct {
	Add    PublishTestersAddCmd    `cmd:"" help:"Add tester groups"`
	Remove PublishTestersRemoveCmd `cmd:"" help:"Remove tester groups"`
	List   PublishTestersListCmd   `cmd:"" help:"List tester groups"`
	Get    PublishTestersGetCmd    `cmd:"" help:"Get tester groups for a track"`
}

// PublishTestersAddCmd adds tester groups.
type PublishTestersAddCmd struct {
	Track        string   `help:"Track to add testers to" default:"internal"`
	Groups       []string `help:"Google Group email addresses"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the testers add command.
func (cmd *PublishTestersAddCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if len(cmd.Groups) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "at least one group is required").
			WithHint("Specify Google Group email addresses with --groups")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"track":  cmd.Track,
			"groups": cmd.Groups,
			"action": "add",
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - testers not added")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get current testers
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var testers *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		testers, err = svc.Edits.Testers.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		// If 404, start with empty testers
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			testers = &androidpublisher.Testers{}
		} else {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get current testers: %v", err))
		}
	}

	// Append new groups (avoid duplicates)
	existingGroups := make(map[string]bool)
	for _, g := range testers.GoogleGroups {
		existingGroups[g] = true
	}
	for _, g := range cmd.Groups {
		if !existingGroups[g] {
			testers.GoogleGroups = append(testers.GoogleGroups, g)
		}
	}

	// Update testers
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updatedTesters *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		updatedTesters, err = svc.Edits.Testers.Update(pkg, editID, cmd.Track, testers).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update testers: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("Testers were added but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"track":        cmd.Track,
		"addedGroups":  cmd.Groups,
		"googleGroups": updatedTesters.GoogleGroups,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishTestersRemoveCmd removes tester groups.
type PublishTestersRemoveCmd struct {
	Track        string   `help:"Track to remove testers from" default:"internal"`
	Groups       []string `help:"Google Group email addresses"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the testers remove command.
func (cmd *PublishTestersRemoveCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if len(cmd.Groups) == 0 {
		return errors.NewAPIError(errors.CodeValidationError, "at least one group is required").
			WithHint("Specify Google Group email addresses with --groups")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"track":  cmd.Track,
			"groups": cmd.Groups,
			"action": "remove",
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - testers not removed")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get current testers
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var testers *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		testers, err = svc.Edits.Testers.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get current testers: %v", err))
	}

	// Remove specified groups
	removeSet := make(map[string]bool)
	for _, g := range cmd.Groups {
		removeSet[g] = true
	}
	var filteredGroups []string
	for _, g := range testers.GoogleGroups {
		if !removeSet[g] {
			filteredGroups = append(filteredGroups, g)
		}
	}
	testers.GoogleGroups = filteredGroups

	// Update testers
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updatedTesters *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		updatedTesters, err = svc.Edits.Testers.Update(pkg, editID, cmd.Track, testers).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update testers: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("Testers were removed but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"track":         cmd.Track,
		"removedGroups": cmd.Groups,
		"googleGroups":  updatedTesters.GoogleGroups,
		"editId":        editID,
		"committed":     committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishTestersListCmd lists tester groups.
type PublishTestersListCmd struct {
	Track string `help:"Track to list testers for (empty for all)"`
}

// Run executes the testers list command.
func (cmd *PublishTestersListCmd) Run(globals *Globals) error {
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

	// Create a temporary edit
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

	tracks := []string{cmd.Track}
	if cmd.Track == "" {
		tracks = []string{"internal", "alpha", "beta", "production"}
	}

	type trackTesters struct {
		Track        string   `json:"track"`
		GoogleGroups []string `json:"googleGroups"`
	}

	var allTesters []trackTesters
	for _, track := range tracks {
		var testers *androidpublisher.Testers
		err = client.DoWithRetry(ctx, func() error {
			testers, err = svc.Edits.Testers.Get(pkg, editID, track).Context(ctx).Do()
			return err
		})
		if err != nil {
			// Skip tracks with errors (e.g., no testers configured)
			if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
				continue
			}
			continue
		}
		allTesters = append(allTesters, trackTesters{
			Track:        track,
			GoogleGroups: testers.GoogleGroups,
		})
	}

	// Clean up edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	})
	client.Release()

	data := map[string]interface{}{
		"testers": allTesters,
		"count":   len(allTesters),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishTestersGetCmd gets tester groups for a track.
type PublishTestersGetCmd struct {
	Track string `help:"Track to get testers for (required)"`
}

// Run executes the testers get command.
func (cmd *PublishTestersGetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.Track == "" {
		return errors.NewAPIError(errors.CodeValidationError, "track is required").
			WithHint("Specify a track with --track, e.g., --track=internal")
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

	// Create a temporary edit
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

	// Get testers
	var testers *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		testers, err = svc.Edits.Testers.Get(pkg, editID, cmd.Track).Context(ctx).Do()
		return err
	})

	// Clean up edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	})
	client.Release()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("no testers found for track: %s", cmd.Track))
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get testers: %v", err))
	}

	data := map[string]interface{}{
		"track":        cmd.Track,
		"googleGroups": testers.GoogleGroups,
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBuildsCmd manages uploaded builds.
type PublishBuildsCmd struct {
	List      PublishBuildsListCmd      `cmd:"" help:"List uploaded builds"`
	Get       PublishBuildsGetCmd       `cmd:"" help:"Get build details"`
	Expire    PublishBuildsExpireCmd    `cmd:"" help:"Expire a build from tracks"`
	ExpireAll PublishBuildsExpireAllCmd `cmd:"" help:"Expire all builds from tracks"`
}

// PublishBuildsListCmd lists uploaded builds.
type PublishBuildsListCmd struct {
	Type   string `help:"Build type (apk, bundle, all)" default:"all"`
	EditID string `help:"Explicit edit transaction ID"`
}

// Run executes the builds list command.
func (cmd *PublishBuildsListCmd) Run(globals *Globals) error {
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

	// Create or reuse a temporary edit
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	editID := cmd.EditID
	createdEdit := false
	if editID == "" {
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		if err != nil {
			client.Release()
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
		createdEdit = true
	}

	type buildInfo struct {
		VersionCode int64  `json:"versionCode"`
		Type        string `json:"type"`
		SHA1        string `json:"sha1,omitempty"`
		SHA256      string `json:"sha256,omitempty"`
	}

	var builds []buildInfo

	// List APKs
	if cmd.Type == checkAll || cmd.Type == "apk" {
		var apkList *androidpublisher.ApksListResponse
		err = client.DoWithRetry(ctx, func() error {
			apkList, err = svc.Edits.Apks.List(pkg, editID).Context(ctx).Do()
			return err
		})
		if err == nil && apkList != nil {
			for _, apk := range apkList.Apks {
				b := buildInfo{
					VersionCode: apk.VersionCode,
					Type:        fileTypeAPK,
				}
				if apk.Binary != nil {
					b.SHA1 = apk.Binary.Sha1
					b.SHA256 = apk.Binary.Sha256
				}
				builds = append(builds, b)
			}
		}
	}

	// List bundles
	if cmd.Type == checkAll || cmd.Type == "bundle" {
		var bundleList *androidpublisher.BundlesListResponse
		err = client.DoWithRetry(ctx, func() error {
			bundleList, err = svc.Edits.Bundles.List(pkg, editID).Context(ctx).Do()
			return err
		})
		if err == nil && bundleList != nil {
			for _, bundle := range bundleList.Bundles {
				builds = append(builds, buildInfo{
					VersionCode: bundle.VersionCode,
					Type:        fileTypeAAB,
					SHA1:        bundle.Sha1,
					SHA256:      bundle.Sha256,
				})
			}
		}
	}

	// Clean up temporary edit
	if createdEdit {
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
	}
	client.Release()

	data := map[string]interface{}{
		"builds": builds,
		"count":  len(builds),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBuildsGetCmd gets build details.
type PublishBuildsGetCmd struct {
	VersionCode int64  `arg:"" help:"Version code to get"`
	Type        string `help:"Build type (apk, bundle, all)" default:"all"`
	EditID      string `help:"Explicit edit transaction ID"`
}

// Run executes the builds get command.
func (cmd *PublishBuildsGetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.VersionCode <= 0 {
		return errors.NewAPIError(errors.CodeValidationError, "version code is required")
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

	// Create or reuse a temporary edit
	if err := client.Acquire(ctx); err != nil {
		return err
	}

	editID := cmd.EditID
	createdEdit := false
	if editID == "" {
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		if err != nil {
			client.Release()
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
		createdEdit = true
	}

	var found bool
	var buildData map[string]interface{}

	// Search APKs
	if cmd.Type == checkAll || cmd.Type == "apk" {
		var apkList *androidpublisher.ApksListResponse
		err = client.DoWithRetry(ctx, func() error {
			apkList, err = svc.Edits.Apks.List(pkg, editID).Context(ctx).Do()
			return err
		})
		if err == nil && apkList != nil {
			for _, apk := range apkList.Apks {
				if apk.VersionCode == cmd.VersionCode {
					buildData = map[string]interface{}{
						"versionCode": apk.VersionCode,
						"type":        fileTypeAPK,
					}
					if apk.Binary != nil {
						buildData["sha1"] = apk.Binary.Sha1
						buildData["sha256"] = apk.Binary.Sha256
					}
					found = true
					break
				}
			}
		}
	}

	// Search bundles if not found yet
	if !found && (cmd.Type == checkAll || cmd.Type == "bundle") {
		var bundleList *androidpublisher.BundlesListResponse
		err = client.DoWithRetry(ctx, func() error {
			bundleList, err = svc.Edits.Bundles.List(pkg, editID).Context(ctx).Do()
			return err
		})
		if err == nil && bundleList != nil {
			for _, bundle := range bundleList.Bundles {
				if bundle.VersionCode == cmd.VersionCode {
					buildData = map[string]interface{}{
						"versionCode": bundle.VersionCode,
						"type":        fileTypeAAB,
						"sha1":        bundle.Sha1,
						"sha256":      bundle.Sha256,
					}
					found = true
					break
				}
			}
		}
	}

	// Clean up temporary edit
	if createdEdit {
		_ = client.DoWithRetry(ctx, func() error {
			return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
		})
	}
	client.Release()

	if !found {
		return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("build not found for version code: %d", cmd.VersionCode))
	}

	result := output.NewResult(buildData).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBuildsExpireCmd expires a build from tracks.
type PublishBuildsExpireCmd struct {
	VersionCode  int64  `arg:"" help:"Version code to expire"`
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the builds expire command.
func (cmd *PublishBuildsExpireCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.VersionCode <= 0 {
		return errors.NewAPIError(errors.CodeValidationError, "version code is required")
	}

	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "expire requires confirmation").
			WithHint("Use --confirm to confirm this destructive operation")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"versionCode": cmd.VersionCode,
			"action":      "expire",
			"dryRun":      true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - build not expired")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// List all tracks and remove the version code from any releases
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var tracksList *androidpublisher.TracksListResponse
	err = client.DoWithRetry(ctx, func() error {
		tracksList, err = svc.Edits.Tracks.List(pkg, editID).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
	}

	var modifiedTracks []string
	for _, track := range tracksList.Tracks {
		modified := false
		for _, release := range track.Releases {
			var filteredCodes []int64
			for _, vc := range release.VersionCodes {
				if vc != cmd.VersionCode {
					filteredCodes = append(filteredCodes, vc)
				} else {
					modified = true
				}
			}
			release.VersionCodes = filteredCodes
		}

		if modified {
			if err := client.Acquire(ctx); err != nil {
				return err
			}
			err = client.DoWithRetry(ctx, func() error {
				_, uerr := svc.Edits.Tracks.Update(pkg, editID, track.Track, track).Context(ctx).Do()
				return uerr
			})
			client.Release()
			if err != nil {
				return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update track %s: %v", track.Track, err))
			}
			modifiedTracks = append(modifiedTracks, track.Track)
		}
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit && len(modifiedTracks) > 0 {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The version code was removed from tracks but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"versionCode":    cmd.VersionCode,
		"modifiedTracks": modifiedTracks,
		"editId":         editID,
		"committed":      committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBuildsExpireAllCmd expires all builds from tracks.
type PublishBuildsExpireAllCmd struct {
	EditID       string `help:"Explicit edit transaction ID"`
	NoAutoCommit bool   `help:"Keep edit open for manual commit"`
	Confirm      bool   `help:"Confirm destructive operation"`
	DryRun       bool   `help:"Show intended actions without executing"`
}

// Run executes the builds expire-all command.
func (cmd *PublishBuildsExpireAllCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if !cmd.Confirm {
		return errors.NewAPIError(errors.CodeValidationError, "expire-all requires confirmation").
			WithHint("Use --confirm to confirm this destructive operation")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"action": "expire-all",
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - no builds expired")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// List all tracks and clear version codes from draft releases
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var tracksList *androidpublisher.TracksListResponse
	err = client.DoWithRetry(ctx, func() error {
		tracksList, err = svc.Edits.Tracks.List(pkg, editID).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
	}

	var modifiedTracks []string
	var expiredCount int
	for _, track := range tracksList.Tracks {
		modified := false
		for _, release := range track.Releases {
			if release.Status == releaseStatusDraft && len(release.VersionCodes) > 0 {
				expiredCount += len(release.VersionCodes)
				release.VersionCodes = nil
				modified = true
			}
		}

		if modified {
			if err := client.Acquire(ctx); err != nil {
				return err
			}
			err = client.DoWithRetry(ctx, func() error {
				_, uerr := svc.Edits.Tracks.Update(pkg, editID, track.Track, track).Context(ctx).Do()
				return uerr
			})
			client.Release()
			if err != nil {
				return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update track %s: %v", track.Track, err))
			}
			modifiedTracks = append(modifiedTracks, track.Track)
		}
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit && len(modifiedTracks) > 0 {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("Builds were expired but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"expiredCount":   expiredCount,
		"modifiedTracks": modifiedTracks,
		"editId":         editID,
		"committed":      committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBetaGroupsCmd manages beta groups (ASC compatibility).
type PublishBetaGroupsCmd struct {
	List          PublishBetaGroupsListCmd          `cmd:"" help:"List beta groups"`
	Get           PublishBetaGroupsGetCmd           `cmd:"" help:"Get beta group details"`
	Create        PublishBetaGroupsCreateCmd        `cmd:"" help:"Create beta group"`
	Update        PublishBetaGroupsUpdateCmd        `cmd:"" help:"Update beta group testers"`
	Delete        PublishBetaGroupsDeleteCmd        `cmd:"" help:"Delete beta group"`
	AddTesters    PublishBetaGroupsAddTestersCmd    `cmd:"" help:"Add tester Google Groups to a beta group"`
	RemoveTesters PublishBetaGroupsRemoveTestersCmd `cmd:"" help:"Remove tester Google Groups from a beta group"`
}

// PublishBetaGroupsListCmd lists beta groups.
type PublishBetaGroupsListCmd struct {
	Track string `help:"Track to list (internal, alpha, beta). Empty lists all supported tracks"`
}

// Run executes the beta-groups list command.
func (cmd *PublishBetaGroupsListCmd) Run(globals *Globals) error {
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

	// Create a temporary edit
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

	// Beta groups map to testing tracks
	testingTracks := []string{"internal", "alpha", "beta"}
	if cmd.Track != "" {
		testingTracks = []string{cmd.Track}
	}

	type betaGroup struct {
		Name         string   `json:"name"`
		Track        string   `json:"track"`
		GoogleGroups []string `json:"googleGroups"`
	}

	var groups []betaGroup
	for _, track := range testingTracks {
		var testers *androidpublisher.Testers
		err = client.DoWithRetry(ctx, func() error {
			testers, err = svc.Edits.Testers.Get(pkg, editID, track).Context(ctx).Do()
			return err
		})
		if err != nil {
			continue
		}
		groups = append(groups, betaGroup{
			Name:         track,
			Track:        track,
			GoogleGroups: testers.GoogleGroups,
		})
	}

	// Clean up edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	})
	client.Release()

	data := map[string]interface{}{
		"betaGroups": groups,
		"count":      len(groups),
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBetaGroupsGetCmd gets beta group details.
type PublishBetaGroupsGetCmd struct {
	Group string `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
}

// Run executes the beta-groups get command.
func (cmd *PublishBetaGroupsGetCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.Group == "" {
		return errors.NewAPIError(errors.CodeValidationError, "group name is required").
			WithHint("Specify a beta group (track) name: internal, alpha, or beta")
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

	// Create a temporary edit
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

	// Get testers for the track
	var testers *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		testers, err = svc.Edits.Testers.Get(pkg, editID, cmd.Group).Context(ctx).Do()
		return err
	})

	// Get track info
	var track *androidpublisher.Track
	trackErr := client.DoWithRetry(ctx, func() error {
		track, err = svc.Edits.Tracks.Get(pkg, editID, cmd.Group).Context(ctx).Do()
		return err
	})

	// Clean up edit
	_ = client.DoWithRetry(ctx, func() error {
		return svc.Edits.Delete(pkg, editID).Context(ctx).Do()
	})
	client.Release()

	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return errors.NewAPIError(errors.CodeNotFound, fmt.Sprintf("beta group not found: %s", cmd.Group))
		}
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get beta group: %v", err))
	}

	data := map[string]interface{}{
		"name":         cmd.Group,
		"track":        cmd.Group,
		"googleGroups": testers.GoogleGroups,
	}

	if trackErr == nil && track != nil && len(track.Releases) > 0 {
		var releases []map[string]interface{}
		for _, r := range track.Releases {
			releases = append(releases, map[string]interface{}{
				"status":       r.Status,
				"versionCodes": r.VersionCodes,
				"name":         r.Name,
			})
		}
		data["releases"] = releases
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBetaGroupsCreateCmd creates a beta group.
type PublishBetaGroupsCreateCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups create command.
func (cmd *PublishBetaGroupsCreateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.Group == "" {
		return errors.NewAPIError(errors.CodeValidationError, "group name is required").
			WithHint("Specify a beta group (track) name: internal, alpha, or beta")
	}

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"group":  cmd.Group,
			"groups": cmd.Groups,
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - beta group not created")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Create or reuse edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Set up testers for the track
	testers := &androidpublisher.Testers{
		GoogleGroups: cmd.Groups,
	}

	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updatedTesters *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		updatedTesters, err = svc.Edits.Testers.Update(pkg, editID, cmd.Group, testers).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create beta group: %v", err))
	}

	// Commit
	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err)).
				WithHint("The beta group was created but the edit could not be committed")
		}
		committed = true
	}

	result := output.NewResult(map[string]interface{}{
		"name":         cmd.Group,
		"track":        cmd.Group,
		"googleGroups": updatedTesters.GoogleGroups,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).
		WithServices("androidpublisher")

	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBetaGroupsUpdateCmd updates beta group testers.
type PublishBetaGroupsUpdateCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses to add"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups update command.
func (cmd *PublishBetaGroupsUpdateCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()
	pkg := globals.Package

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"group":        cmd.Group,
			"googleGroups": cmd.Groups,
			"dryRun":       true,
		}).WithNoOp("dry run - no beta group updated")
		return outputResult(result, globals.Output, globals.Pretty)
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}
	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service")
	}

	// Create edit
	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Update testers for the track
	testers := &androidpublisher.Testers{
		GoogleGroups: cmd.Groups,
	}
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updatedTesters *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		updatedTesters, err = svc.Edits.Testers.Update(pkg, editID, cmd.Group, testers).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update beta group: %v", err))
	}

	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err == nil {
			committed = true
		}
	}

	result := output.NewResult(map[string]interface{}{
		"name":         cmd.Group,
		"googleGroups": updatedTesters.GoogleGroups,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBetaGroupsDeleteCmd deletes a beta group.
type PublishBetaGroupsDeleteCmd struct {
	Group string `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
}

// Run executes the beta-groups delete command.
func (cmd *PublishBetaGroupsDeleteCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()
	pkg := globals.Package

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}
	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service")
	}

	// Create edit
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var edit *androidpublisher.AppEdit
	err = client.DoWithRetry(ctx, func() error {
		edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
	}
	editID := edit.Id

	// Clear testers for the track (effectively deleting the beta group)
	testers := &androidpublisher.Testers{
		GoogleGroups: []string{},
	}
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	err = client.DoWithRetry(ctx, func() error {
		_, uerr := svc.Edits.Testers.Update(pkg, editID, cmd.Group, testers).Context(ctx).Do()
		return uerr
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete beta group: %v", err))
	}

	// Commit
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	err = client.DoWithRetry(ctx, func() error {
		_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
		return cerr
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to commit edit: %v", err))
	}

	result := output.NewResult(map[string]interface{}{
		"name":    cmd.Group,
		"deleted": true,
		"editId":  editID,
	}).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBetaGroupsAddTestersCmd adds tester Google Groups to a beta group.
type PublishBetaGroupsAddTestersCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses to add"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups add-testers command.
func (cmd *PublishBetaGroupsAddTestersCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()
	pkg := globals.Package

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"group":  cmd.Group,
			"adding": cmd.Groups,
			"dryRun": true,
		}).WithNoOp("dry run - no testers added")
		return outputResult(result, globals.Output, globals.Pretty)
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}
	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service")
	}

	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get current testers
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var current *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		current, err = svc.Edits.Testers.Get(pkg, editID, cmd.Group).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get current testers: %v", err))
	}

	// Merge new groups with existing
	existing := make(map[string]bool)
	for _, g := range current.GoogleGroups {
		existing[g] = true
	}
	merged := current.GoogleGroups
	var added []string
	for _, g := range cmd.Groups {
		if !existing[g] {
			merged = append(merged, g)
			added = append(added, g)
		}
	}

	// Update
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updated *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		updated, err = svc.Edits.Testers.Update(pkg, editID, cmd.Group, &androidpublisher.Testers{GoogleGroups: merged}).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to add testers: %v", err))
	}

	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err == nil {
			committed = true
		}
	}

	result := output.NewResult(map[string]interface{}{
		"group":        cmd.Group,
		"added":        added,
		"totalGroups":  len(updated.GoogleGroups),
		"googleGroups": updated.GoogleGroups,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishBetaGroupsRemoveTestersCmd removes tester Google Groups from a beta group.
type PublishBetaGroupsRemoveTestersCmd struct {
	Group        string   `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
	Groups       []string `help:"Google Group email addresses to remove"`
	EditID       string   `help:"Explicit edit transaction ID"`
	NoAutoCommit bool     `help:"Keep edit open for manual commit"`
	DryRun       bool     `help:"Show intended actions without executing"`
}

// Run executes the beta-groups remove-testers command.
func (cmd *PublishBetaGroupsRemoveTestersCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	if globals.Package == "" {
		return errors.ErrPackageRequired
	}
	start := time.Now()
	pkg := globals.Package

	if cmd.DryRun {
		result := output.NewResult(map[string]interface{}{
			"group":    cmd.Group,
			"removing": cmd.Groups,
			"dryRun":   true,
		}).WithNoOp("dry run - no testers removed")
		return outputResult(result, globals.Output, globals.Pretty)
	}

	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return err
	}
	svc, err := client.AndroidPublisher()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to initialize publisher service")
	}

	editID := cmd.EditID
	if editID == "" {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		var edit *androidpublisher.AppEdit
		err = client.DoWithRetry(ctx, func() error {
			edit, err = svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
			return err
		})
		client.Release()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err))
		}
		editID = edit.Id
	}

	// Get current testers
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var current *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		current, err = svc.Edits.Testers.Get(pkg, editID, cmd.Group).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to get current testers: %v", err))
	}

	// Remove specified groups
	toRemove := make(map[string]bool)
	for _, g := range cmd.Groups {
		toRemove[g] = true
	}
	var remaining []string
	var removed []string
	for _, g := range current.GoogleGroups {
		if toRemove[g] {
			removed = append(removed, g)
		} else {
			remaining = append(remaining, g)
		}
	}

	// Update
	if err := client.Acquire(ctx); err != nil {
		return err
	}
	var updated *androidpublisher.Testers
	err = client.DoWithRetry(ctx, func() error {
		updated, err = svc.Edits.Testers.Update(pkg, editID, cmd.Group, &androidpublisher.Testers{GoogleGroups: remaining}).Context(ctx).Do()
		return err
	})
	client.Release()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to remove testers: %v", err))
	}

	committed := false
	if !cmd.NoAutoCommit {
		if err := client.Acquire(ctx); err != nil {
			return err
		}
		err = client.DoWithRetry(ctx, func() error {
			_, cerr := svc.Edits.Commit(pkg, editID).Context(ctx).Do()
			return cerr
		})
		client.Release()
		if err == nil {
			committed = true
		}
	}

	result := output.NewResult(map[string]interface{}{
		"group":        cmd.Group,
		"removed":      removed,
		"totalGroups":  len(updated.GoogleGroups),
		"googleGroups": updated.GoogleGroups,
		"editId":       editID,
		"committed":    committed,
	}).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}

// PublishInternalShareCmd uploads artifacts for internal sharing.
type PublishInternalShareCmd struct {
	Upload PublishInternalShareUploadCmd `cmd:"" help:"Upload artifact for internal sharing"`
}

// PublishInternalShareUploadCmd uploads artifact for internal sharing.
type PublishInternalShareUploadCmd struct {
	File   string `arg:"" help:"File to upload (APK or AAB)" type:"existingfile"`
	DryRun bool   `help:"Show intended actions without executing"`
}

// Run executes the internal-share upload command.
func (cmd *PublishInternalShareUploadCmd) Run(globals *Globals) error {
	ctx := globals.Context
	if ctx == nil {
		ctx = context.Background()
	}
	start := time.Now()

	if globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if cmd.File == "" {
		return errors.NewAPIError(errors.CodeValidationError, "file is required").
			WithHint("Provide an APK or AAB file to upload for internal sharing")
	}

	ext := strings.ToLower(filepath.Ext(cmd.File))
	if ext != extAPK && ext != extAAB {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid file type: %s. Only .apk and .aab files are supported", ext))
	}

	if cmd.DryRun {
		fileType := fileTypeAPK
		if ext == extAAB {
			fileType = fileTypeAAB
		}
		result := output.NewResult(map[string]interface{}{
			"file":   cmd.File,
			"type":   fileType,
			"dryRun": true,
		}).WithDuration(time.Since(start)).
			WithNoOp("dry run - artifact not uploaded for internal sharing")
		return outputResult(result, globals.Output, globals.Pretty)
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

	// Open file
	file, err := os.Open(cmd.File)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("failed to open file: %v", err))
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			_ = cerr
		}
	}()

	// Upload for internal sharing (no edit needed)
	if err := client.AcquireForUpload(ctx); err != nil {
		return err
	}

	var data map[string]interface{}
	if ext == extAAB {
		var resp *androidpublisher.InternalAppSharingArtifact
		err = client.DoWithRetry(ctx, func() error {
			resp, err = svc.Internalappsharingartifacts.Uploadbundle(pkg).Media(file).Context(ctx).Do()
			return err
		})
		client.ReleaseForUpload()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to upload bundle for internal sharing: %v", err))
		}
		data = map[string]interface{}{
			"type":        fileTypeAAB,
			"file":        cmd.File,
			"downloadUrl": resp.DownloadUrl,
			"sha256":      resp.Sha256,
		}
		if resp.CertificateFingerprint != "" {
			data["certificateFingerprint"] = resp.CertificateFingerprint
		}
	} else {
		var resp *androidpublisher.InternalAppSharingArtifact
		err = client.DoWithRetry(ctx, func() error {
			resp, err = svc.Internalappsharingartifacts.Uploadapk(pkg).Media(file).Context(ctx).Do()
			return err
		})
		client.ReleaseForUpload()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to upload APK for internal sharing: %v", err))
		}
		data = map[string]interface{}{
			"type":        fileTypeAPK,
			"file":        cmd.File,
			"downloadUrl": resp.DownloadUrl,
			"sha256":      resp.Sha256,
		}
		if resp.CertificateFingerprint != "" {
			data["certificateFingerprint"] = resp.CertificateFingerprint
		}
	}

	result := output.NewResult(data).WithDuration(time.Since(start)).WithServices("androidpublisher")
	return outputResult(result, globals.Output, globals.Pretty)
}
