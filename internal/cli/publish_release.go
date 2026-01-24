package cli

import (
	"context"
	"fmt"
	"strconv"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

// Release status constants
const (
	statusCompleted  = "completed"
	statusInProgress = "inProgress"
)

func (c *CLI) publishRelease(ctx context.Context, track, name, status string, versionCodes []string, _, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if !config.IsValidTrack(track) {
		return c.OutputError(errors.ErrTrackInvalid)
	}

	if !api.IsValidReleaseStatus(status) {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid status: %s", status)).
			WithHint("Valid statuses: draft, completed, halted, inProgress"))
	}

	var codes []int64
	for _, vc := range versionCodes {
		code, err := strconv.ParseInt(vc, 10, 64)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				fmt.Sprintf("invalid version code: %s", vc)))
		}
		codes = append(codes, code)
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":       true,
			"action":       "release",
			"track":        track,
			"name":         name,
			"status":       status,
			"versionCodes": codes,
			"package":      c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to get track: %v", err)))
	}

	release := &struct {
		Name         string  `json:"name,omitempty"`
		VersionCodes []int64 `json:"versionCodes"`
		Status       string  `json:"status"`
	}{
		Name:         name,
		VersionCodes: codes,
		Status:       status,
	}

	_ = trackInfo
	_ = release

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"track":        track,
		"name":         name,
		"status":       status,
		"versionCodes": codes,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishRollout(ctx context.Context, track string, percentage float64, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if percentage < 0.01 || percentage > 100 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"percentage must be between 0.01 and 100"))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":     true,
			"action":     "rollout",
			"track":      track,
			"percentage": percentage,
			"package":    c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("track not found: %s", track)))
	}

	var updatedRelease *androidpublisher.TrackRelease
	for i, release := range trackInfo.Releases {
		if release.Status == statusInProgress {
			userFraction := percentage / 100.0
			trackInfo.Releases[i].UserFraction = userFraction
			updatedRelease = trackInfo.Releases[i]
			break
		}
	}

	if updatedRelease == nil {
		_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"no in-progress release found on track").
			WithHint("Create a staged rollout release first with status 'inProgress'"))
	}

	_, err = publisher.Edits.Tracks.Update(c.packageName, edit.ServerID, track, trackInfo).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update track: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"track":        track,
		"percentage":   percentage,
		"userFraction": percentage / 100.0,
		"versionCodes": updatedRelease.VersionCodes,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// promoteParams holds parameters for the promote operation.
type promoteParams struct {
	fromTrack    string
	toTrack      string
	percentage   float64
	editID       string
	noAutoCommit bool
}

// validatePromoteInput validates the input parameters for promotion.
func validatePromoteInput(fromTrack, toTrack string) *errors.APIError {
	if !config.IsValidTrack(fromTrack) || !config.IsValidTrack(toTrack) {
		return errors.ErrTrackInvalid
	}
	if fromTrack == toTrack {
		return errors.NewAPIError(errors.CodeValidationError,
			"source and destination tracks must be different")
	}
	return nil
}

// findActiveRelease finds the first active (completed or in-progress) release from a track.
func findActiveRelease(sourceTrack *androidpublisher.Track) *androidpublisher.TrackRelease {
	for _, release := range sourceTrack.Releases {
		if release.Status == statusCompleted || release.Status == statusInProgress {
			return release
		}
	}
	return nil
}

// createPromotedRelease creates a new release for the destination track based on the source release.
func createPromotedRelease(sourceRelease *androidpublisher.TrackRelease, percentage float64) *androidpublisher.TrackRelease {
	newRelease := &androidpublisher.TrackRelease{
		Name:         sourceRelease.Name,
		VersionCodes: sourceRelease.VersionCodes,
		ReleaseNotes: sourceRelease.ReleaseNotes,
	}
	if percentage > 0 && percentage < 100 {
		newRelease.Status = statusInProgress
		newRelease.UserFraction = percentage / 100.0
	} else {
		newRelease.Status = statusCompleted
	}
	return newRelease
}

func (c *CLI) publishPromote(ctx context.Context, fromTrack, toTrack string, percentage float64, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if err := validatePromoteInput(fromTrack, toTrack); err != nil {
		return c.OutputError(err)
	}

	if dryRun {
		return c.outputPromoteDryRun(fromTrack, toTrack, percentage)
	}

	params := promoteParams{
		fromTrack:    fromTrack,
		toTrack:      toTrack,
		percentage:   percentage,
		editID:       editID,
		noAutoCommit: noAutoCommit,
	}

	return c.executePromote(ctx, params)
}

// outputPromoteDryRun outputs the dry run result for promotion.
func (c *CLI) outputPromoteDryRun(fromTrack, toTrack string, percentage float64) error {
	result := output.NewResult(map[string]interface{}{
		"dryRun":     true,
		"action":     "promote",
		"fromTrack":  fromTrack,
		"toTrack":    toTrack,
		"percentage": percentage,
		"package":    c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// executePromote performs the actual promotion operation.
func (c *CLI) executePromote(ctx context.Context, params promoteParams) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, params.editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	cleanupEdit := func() {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
	}

	sourceRelease, apiErr := c.getSourceRelease(ctx, publisher, edit.ServerID, params.fromTrack, cleanupEdit)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}

	newRelease := createPromotedRelease(sourceRelease, params.percentage)

	if apiErr := c.updateDestinationTrack(ctx, publisher, edit.ServerID, params.toTrack, newRelease, cleanupEdit); apiErr != nil {
		return c.OutputError(apiErr)
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !params.noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"fromTrack":    params.fromTrack,
		"toTrack":      params.toTrack,
		"versionCodes": sourceRelease.VersionCodes,
		"status":       newRelease.Status,
		"percentage":   params.percentage,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !params.noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// getSourceRelease retrieves the active release from the source track.
func (c *CLI) getSourceRelease(ctx context.Context, publisher *androidpublisher.Service, editID, fromTrack string, cleanup func()) (*androidpublisher.TrackRelease, *errors.APIError) {
	sourceTrack, err := publisher.Edits.Tracks.Get(c.packageName, editID, fromTrack).Context(ctx).Do()
	if err != nil {
		cleanup()
		return nil, errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("source track not found: %s", fromTrack))
	}

	sourceRelease := findActiveRelease(sourceTrack)
	if sourceRelease == nil {
		cleanup()
		return nil, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("no active release found on track: %s", fromTrack)).
			WithHint("Ensure the source track has a completed or in-progress release")
	}

	return sourceRelease, nil
}

// updateDestinationTrack updates the destination track with the new release.
func (c *CLI) updateDestinationTrack(ctx context.Context, publisher *androidpublisher.Service, editID, toTrack string, newRelease *androidpublisher.TrackRelease, cleanup func()) *errors.APIError {
	destTrack, err := publisher.Edits.Tracks.Get(c.packageName, editID, toTrack).Context(ctx).Do()
	if err != nil {
		destTrack = &androidpublisher.Track{
			Track: toTrack,
		}
	}

	destTrack.Releases = []*androidpublisher.TrackRelease{newRelease}

	_, err = publisher.Edits.Tracks.Update(c.packageName, editID, toTrack, destTrack).Context(ctx).Do()
	if err != nil {
		cleanup()
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update destination track: %v", err))
	}

	return nil
}

func (c *CLI) publishHalt(ctx context.Context, track, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":  true,
			"action":  "halt",
			"track":   track,
			"package": c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("track not found: %s", track)))
	}

	var haltedRelease *androidpublisher.TrackRelease
	for i, release := range trackInfo.Releases {
		if release.Status == statusInProgress {
			trackInfo.Releases[i].Status = "halted"
			haltedRelease = trackInfo.Releases[i]
			break
		}
	}

	if haltedRelease == nil {
		_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"no in-progress release found on track").
			WithHint("Only releases with status 'inProgress' can be halted"))
	}

	_, err = publisher.Edits.Tracks.Update(c.packageName, edit.ServerID, track, trackInfo).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update track: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"track":        track,
		"status":       "halted",
		"versionCodes": haltedRelease.VersionCodes,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// rollbackResult holds the result of finding a release to rollback to.
type rollbackResult struct {
	release      *androidpublisher.TrackRelease
	versionCodes []int64
}

// findReleaseByVersionCode searches for a release containing the specified version code.
func findReleaseByVersionCode(trackInfo *androidpublisher.Track, targetVersionCode int64) *rollbackResult {
	for _, release := range trackInfo.Releases {
		for _, vc := range release.VersionCodes {
			if vc == targetVersionCode {
				return &rollbackResult{
					release:      release,
					versionCodes: []int64{targetVersionCode},
				}
			}
		}
	}
	return nil
}

// findPreviousRelease finds the most recent completed release for rollback.
func findPreviousRelease(trackInfo *androidpublisher.Track) *rollbackResult {
	for _, release := range trackInfo.Releases {
		if release.Status == statusCompleted {
			return &rollbackResult{
				release:      release,
				versionCodes: release.VersionCodes,
			}
		}
	}
	return nil
}

func (c *CLI) publishRollback(ctx context.Context, track, versionCode, editID string, noAutoCommit, dryRun bool) error {
	if err := c.validateRollbackInput(track); err != nil {
		return c.OutputError(err)
	}

	targetVersionCode, err := c.parseVersionCodeOpt(versionCode)
	if err != nil {
		return c.OutputError(err)
	}

	if dryRun {
		return c.outputRollbackDryRun(track, versionCode)
	}

	return c.executeRollback(ctx, track, targetVersionCode, editID, noAutoCommit)
}

// validateRollbackInput validates the input parameters for rollback.
func (c *CLI) validateRollbackInput(track string) *errors.APIError {
	if err := c.requirePackage(); err != nil {
		return err.(*errors.APIError)
	}
	if !config.IsValidTrack(track) {
		return errors.ErrTrackInvalid
	}
	return nil
}

// parseVersionCodeOpt parses an optional version code string.
func (c *CLI) parseVersionCodeOpt(versionCode string) (int64, *errors.APIError) {
	if versionCode == "" {
		return 0, nil
	}
	parsed, err := strconv.ParseInt(versionCode, 10, 64)
	if err != nil {
		return 0, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid version code: %s", versionCode))
	}
	return parsed, nil
}

// outputRollbackDryRun outputs the dry run result for rollback.
func (c *CLI) outputRollbackDryRun(track, versionCode string) error {
	result := output.NewResult(map[string]interface{}{
		"dryRun":      true,
		"action":      "rollback",
		"track":       track,
		"versionCode": versionCode,
		"package":     c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// executeRollback performs the actual rollback operation.
func (c *CLI) executeRollback(ctx context.Context, track string, targetVersionCode int64, editID string, noAutoCommit bool) error {
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	cleanupEdit := func() {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
	}

	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		cleanupEdit()
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("track not found: %s", track)))
	}

	rollback, apiErr := c.findRollbackTarget(trackInfo, targetVersionCode, track)
	if apiErr != nil {
		cleanupEdit()
		return c.OutputError(apiErr)
	}

	if apiErr := c.updateTrackWithRollback(ctx, publisher, edit.ServerID, track, trackInfo, rollback.versionCodes); apiErr != nil {
		cleanupEdit()
		return c.OutputError(apiErr)
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"track":        track,
		"versionCodes": rollback.versionCodes,
		"releaseName":  rollback.release.Name,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// findRollbackTarget finds the appropriate release to rollback to.
func (c *CLI) findRollbackTarget(trackInfo *androidpublisher.Track, targetVersionCode int64, track string) (*rollbackResult, *errors.APIError) {
	if targetVersionCode > 0 {
		result := findReleaseByVersionCode(trackInfo, targetVersionCode)
		if result == nil {
			return nil, errors.NewAPIError(errors.CodeNotFound,
				fmt.Sprintf("version code %d not found in track history", targetVersionCode)).
				WithHint("Check available versions with 'gpd publish status --track " + track + "'")
		}
		return result, nil
	}

	result := findPreviousRelease(trackInfo)
	if result == nil {
		return nil, errors.NewAPIError(errors.CodeValidationError,
			"no previous release found to rollback to").
			WithHint("Specify a version code with --version-code flag")
	}
	return result, nil
}

// updateTrackWithRollback updates the track with the rollback release.
func (c *CLI) updateTrackWithRollback(ctx context.Context, publisher *androidpublisher.Service, editID, track string, trackInfo *androidpublisher.Track, versionCodes []int64) *errors.APIError {
	newRelease := &androidpublisher.TrackRelease{
		VersionCodes: versionCodes,
		Status:       statusCompleted,
	}
	trackInfo.Releases = []*androidpublisher.TrackRelease{newRelease}

	_, err := publisher.Edits.Tracks.Update(c.packageName, editID, track, trackInfo).Context(ctx).Do()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update track: %v", err))
	}
	return nil
}

func (c *CLI) publishStatus(ctx context.Context, track string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	edit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer func() { _ = publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do() }()

	if track != "" {
		trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.Id, track).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
				fmt.Sprintf("track not found: %s", track)))
		}
		result := output.NewResult(trackInfo)
		return c.Output(result.WithServices("androidpublisher"))
	}

	tracks, err := publisher.Edits.Tracks.List(c.packageName, edit.Id).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(tracks.Tracks)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishTracks(ctx context.Context) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	return c.publishStatus(ctx, "")
}

func (c *CLI) publishCapabilities(_ context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"supportedTracks": config.ValidTracks(),
		"supportedStatuses": []string{
			string(api.StatusDraft),
			string(api.StatusCompleted),
			string(api.StatusHalted),
			string(api.StatusInProgress),
		},
		"supportedArtifacts": []string{"aab", "apk"},
		"rolloutRange": map[string]interface{}{
			"min": 0.01,
			"max": 100.0,
		},
		"maxInternalTesters": 200,
	})
	return c.Output(result.WithServices("androidpublisher"))
}
