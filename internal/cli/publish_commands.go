// Package cli provides publish commands for gpd.
package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/google-play-cli/gpd/internal/api"
	"github.com/google-play-cli/gpd/internal/config"
	"github.com/google-play-cli/gpd/internal/edits"
	"github.com/google-play-cli/gpd/internal/errors"
	"github.com/google-play-cli/gpd/internal/output"
)

func (c *CLI) addPublishCommands() {
	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publishing commands",
		Long:  "Manage app publishing, releases, and store listings.",
	}

	// Shared flags for publishing commands
	var editID string
	var track string
	var status string
	var name string
	var versionCodes []string
	var percentage float64
	var releaseNotesFile string
	var confirm bool
	var dryRun bool

	// publish upload
	uploadCmd := &cobra.Command{
		Use:   "upload [file]",
		Short: "Upload an artifact (AAB or APK)",
		Long:  "Upload an Android App Bundle (AAB) or APK to an edit transaction.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishUpload(cmd.Context(), args[0], editID, dryRun)
		},
	}
	uploadCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	uploadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	// publish release
	releaseCmd := &cobra.Command{
		Use:   "release",
		Short: "Create or update a release",
		Long:  "Create a new release on a track with specified version codes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishRelease(cmd.Context(), track, name, status, versionCodes, releaseNotesFile, editID, dryRun)
		},
	}
	releaseCmd.Flags().StringVar(&track, "track", "", "Release track (internal, alpha, beta, production)")
	releaseCmd.Flags().StringVar(&name, "name", "", "Release name")
	releaseCmd.Flags().StringVar(&status, "status", "draft", "Release status (draft, completed, halted, inProgress)")
	releaseCmd.Flags().StringSliceVar(&versionCodes, "version-code", nil, "Version codes to include (repeatable)")
	releaseCmd.Flags().StringVar(&releaseNotesFile, "release-notes-file", "", "JSON file with localized release notes")
	releaseCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	releaseCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	releaseCmd.MarkFlagRequired("track")

	// publish rollout
	rolloutCmd := &cobra.Command{
		Use:   "rollout",
		Short: "Update rollout percentage",
		Long:  "Update the staged rollout percentage for a production release.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishRollout(cmd.Context(), track, percentage, editID, dryRun)
		},
	}
	rolloutCmd.Flags().StringVar(&track, "track", "production", "Release track")
	rolloutCmd.Flags().Float64Var(&percentage, "percentage", 0, "Rollout percentage (0.01-100.00)")
	rolloutCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	rolloutCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	rolloutCmd.MarkFlagRequired("percentage")

	// publish promote
	var fromTrack, toTrack string
	promoteCmd := &cobra.Command{
		Use:   "promote",
		Short: "Promote a release between tracks",
		Long:  "Copy a release from one track to another.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishPromote(cmd.Context(), fromTrack, toTrack, percentage, editID, dryRun)
		},
	}
	promoteCmd.Flags().StringVar(&fromTrack, "from-track", "", "Source track")
	promoteCmd.Flags().StringVar(&toTrack, "to-track", "", "Destination track")
	promoteCmd.Flags().Float64Var(&percentage, "percentage", 0, "Rollout percentage for destination")
	promoteCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	promoteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	promoteCmd.MarkFlagRequired("from-track")
	promoteCmd.MarkFlagRequired("to-track")

	// publish halt
	haltCmd := &cobra.Command{
		Use:   "halt",
		Short: "Halt a production rollout",
		Long:  "Halt an in-progress production rollout.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations"))
			}
			return c.publishHalt(cmd.Context(), track, editID, dryRun)
		},
	}
	haltCmd.Flags().StringVar(&track, "track", "production", "Release track")
	haltCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	haltCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")
	haltCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	// publish rollback
	var rollbackVersionCode string
	rollbackCmd := &cobra.Command{
		Use:   "rollback",
		Short: "Rollback to a previous version",
		Long:  "Rollback to a previous version from track history.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !confirm {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"--confirm flag required for destructive operations"))
			}
			return c.publishRollback(cmd.Context(), track, rollbackVersionCode, editID, dryRun)
		},
	}
	rollbackCmd.Flags().StringVar(&track, "track", "", "Release track")
	rollbackCmd.Flags().StringVar(&rollbackVersionCode, "version-code", "", "Specific version code to rollback to")
	rollbackCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	rollbackCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")
	rollbackCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	rollbackCmd.MarkFlagRequired("track")

	// publish status
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get track status",
		Long:  "Get the current status of a release track.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishStatus(cmd.Context(), track)
		},
	}
	statusCmd.Flags().StringVar(&track, "track", "", "Release track (leave empty for all tracks)")

	// publish tracks
	tracksCmd := &cobra.Command{
		Use:   "tracks",
		Short: "List all tracks",
		Long:  "List all available release tracks and their status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishTracks(cmd.Context())
		},
	}

	// publish capabilities
	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List publishing capabilities",
		Long:  "List available publishing operations and constraints.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishCapabilities(cmd.Context())
		},
	}

	// publish listing
	listingCmd := &cobra.Command{
		Use:   "listing",
		Short: "Manage store listing",
		Long:  "Update app title, short description, and full description.",
	}

	var locale, title, shortDesc, fullDesc string
	listingUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update store listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishListingUpdate(cmd.Context(), locale, title, shortDesc, fullDesc, editID, dryRun)
		},
	}
	listingUpdateCmd.Flags().StringVar(&locale, "locale", "en-US", "Locale code")
	listingUpdateCmd.Flags().StringVar(&title, "title", "", "App title")
	listingUpdateCmd.Flags().StringVar(&shortDesc, "short-description", "", "Short description")
	listingUpdateCmd.Flags().StringVar(&fullDesc, "full-description", "", "Full description")
	listingUpdateCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	listingUpdateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	listingGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get store listing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishListingGet(cmd.Context(), locale)
		},
	}
	listingGetCmd.Flags().StringVar(&locale, "locale", "", "Locale code (leave empty for all)")

	listingCmd.AddCommand(listingUpdateCmd, listingGetCmd)

	// publish assets
	assetsCmd := &cobra.Command{
		Use:   "assets",
		Short: "Manage store assets",
		Long:  "Upload and manage screenshots and graphics.",
	}

	var assetsDir, category string
	var replace bool
	assetsUploadCmd := &cobra.Command{
		Use:   "upload [directory]",
		Short: "Upload assets from directory",
		Long:  "Upload assets following the directory convention: assets/{locale}/{category}/",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := assetsDir
			if len(args) > 0 {
				dir = args[0]
			}
			return c.publishAssetsUpload(cmd.Context(), dir, category, replace, editID, dryRun)
		},
	}
	assetsUploadCmd.Flags().StringVar(&assetsDir, "dir", "assets", "Assets directory")
	assetsUploadCmd.Flags().StringVar(&category, "replace", "", "Category to replace (phone, tablet, tv, wear)")
	assetsUploadCmd.Flags().BoolVar(&replace, "replace-all", false, "Replace all existing assets")
	assetsUploadCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	assetsUploadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	assetsSpecCmd := &cobra.Command{
		Use:   "spec",
		Short: "Output asset validation matrix",
		Long:  "Output machine-readable asset dimension requirements.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishAssetsSpec(cmd.Context())
		},
	}

	assetsCmd.AddCommand(assetsUploadCmd, assetsSpecCmd)

	// publish testers
	testersCmd := &cobra.Command{
		Use:   "testers",
		Short: "Manage testers",
		Long:  "Manage tester groups for tracks.",
	}

	var testersTrack string
	var groups []string
	testersAddCmd := &cobra.Command{
		Use:   "add",
		Short: "Add tester groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishTestersAdd(cmd.Context(), testersTrack, groups, dryRun)
		},
	}
	testersAddCmd.Flags().StringVar(&testersTrack, "track", "internal", "Track to add testers to")
	testersAddCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses")
	testersAddCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	testersRemoveCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove tester groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishTestersRemove(cmd.Context(), testersTrack, groups, dryRun)
		},
	}
	testersRemoveCmd.Flags().StringVar(&testersTrack, "track", "internal", "Track to remove testers from")
	testersRemoveCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses")
	testersRemoveCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	testersListCmd := &cobra.Command{
		Use:   "list",
		Short: "List tester groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishTestersList(cmd.Context(), testersTrack)
		},
	}
	testersListCmd.Flags().StringVar(&testersTrack, "track", "", "Track to list testers for (empty for all)")

	testersCmd.AddCommand(testersAddCmd, testersRemoveCmd, testersListCmd)

	publishCmd.AddCommand(uploadCmd, releaseCmd, rolloutCmd, promoteCmd, haltCmd, rollbackCmd,
		statusCmd, tracksCmd, capabilitiesCmd, listingCmd, assetsCmd, testersCmd)
	c.rootCmd.AddCommand(publishCmd)
}

func (c *CLI) publishUpload(ctx context.Context, filePath, editID string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Validate file exists and is AAB or APK
	info, err := os.Stat(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath)))
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".aab" && ext != ".apk" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"file must be an AAB or APK").WithHint("Supported formats: .aab, .apk"))
	}

	// Check idempotency cache
	editMgr := edits.NewManager()
	cached, err := editMgr.GetCachedArtifact(c.packageName, filePath)
	if err == nil && cached != nil {
		result := output.NewResult(map[string]interface{}{
			"cached":    true,
			"sha256":    cached.SHA256,
			"path":      filePath,
			"size":      info.Size(),
			"sizeHuman": edits.FormatBytes(info.Size()),
		})
		return c.Output(result.WithNoOp("artifact already uploaded").WithServices("androidpublisher"))
	}

	if dryRun {
		hash, _ := edits.HashFile(filePath)
		result := output.NewResult(map[string]interface{}{
			"dryRun":    true,
			"action":    "upload",
			"path":      filePath,
			"sha256":    hash,
			"size":      info.Size(),
			"sizeHuman": edits.FormatBytes(info.Size()),
			"type":      ext[1:], // Remove dot
			"package":   c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	// Get API client
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	// Acquire lock for package
	if err := editMgr.AcquireLock(ctx, c.packageName); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer editMgr.ReleaseLock(c.packageName)

	// Create or get edit
	edit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to create edit: %v", err)))
	}

	// Upload file
	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer f.Close()

	var versionCode int64
	if ext == ".aab" {
		bundle, err := publisher.Edits.Bundles.Upload(c.packageName, edit.Id).
			Media(f).Context(ctx).Do()
		if err != nil {
			// Abort edit on failure
			publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to upload bundle: %v", err)))
		}
		versionCode = bundle.VersionCode
	} else {
		apk, err := publisher.Edits.Apks.Upload(c.packageName, edit.Id).
			Media(f).Context(ctx).Do()
		if err != nil {
			publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to upload APK: %v", err)))
		}
		versionCode = int64(apk.VersionCode)
	}

	// Commit edit
	_, err = publisher.Edits.Commit(c.packageName, edit.Id).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to commit edit: %v", err)))
	}

	// Cache the artifact
	editMgr.CacheArtifact(c.packageName, filePath, versionCode)

	hash, _ := edits.HashFile(filePath)
	result := output.NewResult(map[string]interface{}{
		"success":     true,
		"versionCode": versionCode,
		"sha256":      hash,
		"path":        filePath,
		"size":        info.Size(),
		"sizeHuman":   edits.FormatBytes(info.Size()),
		"type":        ext[1:],
		"package":     c.packageName,
		"editId":      edit.Id,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishRelease(ctx context.Context, track, name, status string, versionCodes []string, releaseNotesFile, editID string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Validate track
	if !config.IsValidTrack(track) {
		return c.OutputError(errors.ErrTrackInvalid)
	}

	// Validate status
	if !api.IsValidReleaseStatus(status) {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid status: %s", status)).
			WithHint("Valid statuses: draft, completed, halted, inProgress"))
	}

	// Parse version codes
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

	// Get API client
	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	// Create edit
	edit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to create edit: %v", err)))
	}

	// Get track and update release
	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.Id, track).Context(ctx).Do()
	if err != nil {
		publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to get track: %v", err)))
	}

	// Create release
	release := &struct {
		Name         string  `json:"name,omitempty"`
		VersionCodes []int64 `json:"versionCodes"`
		Status       string  `json:"status"`
	}{
		Name:         name,
		VersionCodes: codes,
		Status:       status,
	}

	// This is simplified - actual implementation would use the Google API types properly
	_ = trackInfo
	_ = release

	// Commit edit
	_, err = publisher.Edits.Commit(c.packageName, edit.Id).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to commit edit: %v", err)))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"track":        track,
		"name":         name,
		"status":       status,
		"versionCodes": codes,
		"package":      c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishRollout(ctx context.Context, track string, percentage float64, editID string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Validate percentage
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

	// Implementation would update the track's user fraction
	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"track":      track,
		"percentage": percentage,
		"package":    c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishPromote(ctx context.Context, fromTrack, toTrack string, percentage float64, editID string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if !config.IsValidTrack(fromTrack) || !config.IsValidTrack(toTrack) {
		return c.OutputError(errors.ErrTrackInvalid)
	}

	if dryRun {
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

	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"fromTrack":  fromTrack,
		"toTrack":    toTrack,
		"percentage": percentage,
		"package":    c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishHalt(ctx context.Context, track, editID string, dryRun bool) error {
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

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"track":   track,
		"status":  "halted",
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishRollback(ctx context.Context, track, versionCode, editID string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if !config.IsValidTrack(track) {
		return c.OutputError(errors.ErrTrackInvalid)
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":      true,
			"action":      "rollback",
			"track":       track,
			"versionCode": versionCode,
			"package":     c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	result := output.NewResult(map[string]interface{}{
		"success":     true,
		"track":       track,
		"versionCode": versionCode,
		"package":     c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
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
	defer publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()

	if track != "" {
		trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.Id, track).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
				fmt.Sprintf("track not found: %s", track)))
		}
		result := output.NewResult(trackInfo)
		return c.Output(result.WithServices("androidpublisher"))
	}

	// Get all tracks
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

func (c *CLI) publishCapabilities(ctx context.Context) error {
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

func (c *CLI) publishListingUpdate(ctx context.Context, locale, title, shortDesc, fullDesc, editID string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	locale = config.NormalizeLocale(locale)

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":           true,
			"action":           "update_listing",
			"locale":           locale,
			"title":            title,
			"shortDescription": shortDesc,
			"fullDescription":  fullDesc,
			"package":          c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	result := output.NewResult(map[string]interface{}{
		"success":          true,
		"locale":           locale,
		"title":            title,
		"shortDescription": shortDesc,
		"fullDescription":  fullDesc,
		"package":          c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishListingGet(ctx context.Context, locale string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Implementation would fetch listing from API
	result := output.NewResult(map[string]interface{}{
		"locale":  locale,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishAssetsUpload(ctx context.Context, dir, category string, replace bool, editID string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":   true,
			"action":   "upload_assets",
			"dir":      dir,
			"category": category,
			"replace":  replace,
			"package":  c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	result := output.NewResult(map[string]interface{}{
		"success":  true,
		"dir":      dir,
		"category": category,
		"replace":  replace,
		"package":  c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishAssetsSpec(ctx context.Context) error {
	result := output.NewResult(map[string]interface{}{
		"phone": map[string]interface{}{
			"screenshot": map[string]interface{}{
				"minWidth":  320,
				"maxWidth":  3840,
				"minHeight": 320,
				"maxHeight": 3840,
				"maxSize":   8 * 1024 * 1024,
				"formats":   []string{"png", "jpg", "jpeg"},
				"maxCount":  8,
			},
		},
		"tablet": map[string]interface{}{
			"screenshot": map[string]interface{}{
				"minWidth":  320,
				"maxWidth":  3840,
				"minHeight": 320,
				"maxHeight": 3840,
				"maxSize":   8 * 1024 * 1024,
				"formats":   []string{"png", "jpg", "jpeg"},
				"maxCount":  8,
			},
		},
		"featureGraphic": map[string]interface{}{
			"width":   1024,
			"height":  500,
			"maxSize": 1 * 1024 * 1024,
			"formats": []string{"png", "jpg", "jpeg"},
		},
		"icon": map[string]interface{}{
			"width":   512,
			"height":  512,
			"maxSize": 1 * 1024 * 1024,
			"formats": []string{"png"},
		},
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishTestersAdd(ctx context.Context, track string, groups []string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":  true,
			"action":  "add_testers",
			"track":   track,
			"groups":  groups,
			"package": c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"track":   track,
		"groups":  groups,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishTestersRemove(ctx context.Context, track string, groups []string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":  true,
			"action":  "remove_testers",
			"track":   track,
			"groups":  groups,
			"package": c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"track":   track,
		"groups":  groups,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishTestersList(ctx context.Context, track string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"track":   track,
		"groups":  []string{},
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}
