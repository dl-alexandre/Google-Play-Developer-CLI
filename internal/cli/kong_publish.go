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

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
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
	File               string `arg:"" help:"File to upload (APK or AAB)" type:"existingfile"`
	Track              string `help:"Target track" default:"internal" enum:"internal,alpha,beta,production"`
	EditID             string `help:"Explicit edit transaction ID"`
	ObbMain            string `help:"Main expansion file path"`
	ObbPatch           string `help:"Patch expansion file path"`
	ObbMainRefVersion  int64  `help:"Reference version code for main expansion file"`
	ObbPatchRefVersion int64  `help:"Reference version code for patch expansion file"`
	NoAutoCommit       bool   `help:"Keep edit open for manual commit"`
	DryRun             bool   `help:"Show intended actions without executing"`
}

const fileTypeAAB = "aab"

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
	if ext != ".apk" && ext != ".aab" {
		return nil, "", errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid file type: %s. Only .apk and .aab files are supported", ext))
	}

	fileType := "apk"
	if ext == ".aab" {
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
	Track               string   `help:"Release track" default:"internal" enum:"internal,alpha,beta,production"`
	Name                string   `help:"Release name"`
	Status              string   `help:"Release status" default:"draft" enum:"draft,completed,halted,inProgress"`
	VersionCodes        []string `help:"Version codes to include (repeatable)"`
	RetainVersionCodes  []string `help:"Version codes to retain (repeatable)"`
	InAppUpdatePriority int      `help:"In-app update priority (0-5)" default:"-1"`
	ReleaseNotesFile    string   `help:"JSON file with localized release notes" type:"existingfile"`
	EditID              string   `help:"Explicit edit transaction ID"`
	NoAutoCommit        bool     `help:"Keep edit open for manual commit"`
	DryRun              bool     `help:"Show intended actions without executing"`
	Wait                bool     `help:"Wait for release to complete"`
	WaitTimeout         string   `help:"Maximum time to wait" default:"30m"`
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish rollout not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish promote not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish halt not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish rollback not yet implemented")
}

// PublishStatusCmd gets track status.
type PublishStatusCmd struct {
	Track string `help:"Release track (leave empty for all tracks)"`
}

// Run executes the status command.
func (cmd *PublishStatusCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish status not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish capabilities not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing update not yet implemented")
}

// PublishListingGetCmd gets store listing.
type PublishListingGetCmd struct {
	Locale string `help:"Locale code (leave empty for all)"`
}

// Run executes the listing get command.
func (cmd *PublishListingGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing get not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing delete not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish details get not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish details update not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish details patch not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish images upload not yet implemented")
}

// PublishImagesListCmd lists images.
type PublishImagesListCmd struct {
	Type   string `arg:"" help:"Image type (icon, featureGraphic, phoneScreenshots, etc.)"`
	Locale string `help:"Locale code" default:"en-US"`
	EditID string `help:"Explicit edit transaction ID"`
}

// Run executes the images list command.
func (cmd *PublishImagesListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images list not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish images delete not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish images deleteall not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish assets upload not yet implemented")
}

// PublishAssetsSpecCmd outputs asset validation matrix.
type PublishAssetsSpecCmd struct{}

// Run executes the assets spec command.
func (cmd *PublishAssetsSpecCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish assets spec not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish deobfuscation upload not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers add not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers remove not yet implemented")
}

// PublishTestersListCmd lists tester groups.
type PublishTestersListCmd struct {
	Track string `help:"Track to list testers for (empty for all)"`
}

// Run executes the testers list command.
func (cmd *PublishTestersListCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers list not yet implemented")
}

// PublishTestersGetCmd gets tester groups for a track.
type PublishTestersGetCmd struct {
	Track string `help:"Track to get testers for (required)"`
}

// Run executes the testers get command.
func (cmd *PublishTestersGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers get not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds list not yet implemented")
}

// PublishBuildsGetCmd gets build details.
type PublishBuildsGetCmd struct {
	VersionCode int64  `arg:"" help:"Version code to get"`
	Type        string `help:"Build type (apk, bundle, all)" default:"all"`
	EditID      string `help:"Explicit edit transaction ID"`
}

// Run executes the builds get command.
func (cmd *PublishBuildsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds get not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds expire not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds expire-all not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups list not yet implemented")
}

// PublishBetaGroupsGetCmd gets beta group details.
type PublishBetaGroupsGetCmd struct {
	Group string `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
}

// Run executes the beta-groups get command.
func (cmd *PublishBetaGroupsGetCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups get not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups create not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups update not yet implemented")
}

// PublishBetaGroupsDeleteCmd deletes a beta group.
type PublishBetaGroupsDeleteCmd struct {
	Group string `arg:"" help:"Beta group (track) name: internal, alpha, or beta"`
}

// Run executes the beta-groups delete command.
func (cmd *PublishBetaGroupsDeleteCmd) Run(globals *Globals) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups delete not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups add-testers not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta-groups remove-testers not yet implemented")
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
	return errors.NewAPIError(errors.CodeGeneralError, "publish internal-share upload not yet implemented")
}
