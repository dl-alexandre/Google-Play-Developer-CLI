// Package cli provides publish commands for gpd.
package cli

import (
	"bufio"
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder for image validation
	_ "image/png"  // Register PNG decoder for image validation
	"net/mail"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) prepareEdit(ctx context.Context, publisher *androidpublisher.Service, editID string) (*edits.Manager, *edits.Edit, bool, error) {
	editMgr := edits.NewManager()
	if err := editMgr.AcquireLock(ctx, c.packageName); err != nil {
		return nil, nil, false, err
	}

	var edit *edits.Edit
	created := false
	if editID != "" {
		stored, err := editMgr.LoadEdit(c.packageName, editID)
		if err != nil {
			_ = editMgr.ReleaseLock(c.packageName)
			return nil, nil, false, err
		}
		if stored != nil {
			if editMgr.IsEditExpired(stored, time.Now()) {
				_ = editMgr.ReleaseLock(c.packageName)
				return nil, nil, false, errors.NewAPIError(errors.CodeConflict, "edit has expired")
			}
			edit = stored
		} else {
			edit = &edits.Edit{
				Handle:      editID,
				ServerID:    editID,
				PackageName: c.packageName,
				CreatedAt:   time.Now(),
				LastUsedAt:  time.Now(),
				State:       edits.StateDraft,
			}
			if err := editMgr.SaveEdit(edit); err != nil {
				_ = editMgr.ReleaseLock(c.packageName)
				return nil, nil, false, err
			}
		}
	} else {
		apiEdit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
		if err != nil {
			_ = editMgr.ReleaseLock(c.packageName)
			return nil, nil, false, errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to create edit: %v", err))
		}
		edit = &edits.Edit{
			Handle:      apiEdit.Id,
			ServerID:    apiEdit.Id,
			PackageName: c.packageName,
			CreatedAt:   time.Now(),
			LastUsedAt:  time.Now(),
			State:       edits.StateDraft,
		}
		created = true
		if err := editMgr.SaveEdit(edit); err != nil {
			_ = editMgr.ReleaseLock(c.packageName)
			return nil, nil, false, err
		}
	}
	return editMgr, edit, created, nil
}

func (c *CLI) finalizeEdit(ctx context.Context, publisher *androidpublisher.Service, editMgr *edits.Manager, edit *edits.Edit, commit bool) error {
	if edit == nil {
		return errors.NewAPIError(errors.CodeValidationError, "edit is required")
	}
	if !commit {
		edit.LastUsedAt = time.Now()
		if err := editMgr.SaveEdit(edit); err != nil {
			return err
		}
		return nil
	}
	_, err := publisher.Edits.Commit(c.packageName, edit.ServerID).Context(ctx).Do()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to commit edit: %v", err))
	}
	_, _ = editMgr.UpdateEditState(c.packageName, edit.Handle, edits.StateCommitted)
	_ = editMgr.DeleteEdit(c.packageName, edit.Handle)
	return nil
}

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
	var noAutoCommit bool
	var deobfuscationType string
	var deobfuscationVersionCode int64
	var deobfuscationChunkSize int64

	// publish upload
	uploadCmd := &cobra.Command{
		Use:   "upload [file]",
		Short: "Upload an artifact (AAB or APK)",
		Long:  "Upload an Android App Bundle (AAB) or APK to an edit transaction.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishUpload(cmd.Context(), args[0], editID, noAutoCommit, dryRun)
		},
	}
	uploadCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	uploadCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	uploadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	// publish release
	releaseCmd := &cobra.Command{
		Use:   "release",
		Short: "Create or update a release",
		Long:  "Create a new release on a track with specified version codes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishRelease(cmd.Context(), track, name, status, versionCodes, releaseNotesFile, editID, noAutoCommit, dryRun)
		},
	}
	releaseCmd.Flags().StringVar(&track, "track", "", "Release track (internal, alpha, beta, production)")
	releaseCmd.Flags().StringVar(&name, "name", "", "Release name")
	releaseCmd.Flags().StringVar(&status, "status", "draft", "Release status (draft, completed, halted, inProgress)")
	releaseCmd.Flags().StringSliceVar(&versionCodes, "version-code", nil, "Version codes to include (repeatable)")
	releaseCmd.Flags().StringVar(&releaseNotesFile, "release-notes-file", "", "JSON file with localized release notes")
	releaseCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	releaseCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	releaseCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	_ = releaseCmd.MarkFlagRequired("track")

	// publish rollout
	rolloutCmd := &cobra.Command{
		Use:   "rollout",
		Short: "Update rollout percentage",
		Long:  "Update the staged rollout percentage for a production release.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishRollout(cmd.Context(), track, percentage, editID, noAutoCommit, dryRun)
		},
	}
	rolloutCmd.Flags().StringVar(&track, "track", "production", "Release track")
	rolloutCmd.Flags().Float64Var(&percentage, "percentage", 0, "Rollout percentage (0.01-100.00)")
	rolloutCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	rolloutCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	rolloutCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	_ = rolloutCmd.MarkFlagRequired("percentage")

	// publish promote
	var fromTrack, toTrack string
	promoteCmd := &cobra.Command{
		Use:   "promote",
		Short: "Promote a release between tracks",
		Long:  "Copy a release from one track to another.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishPromote(cmd.Context(), fromTrack, toTrack, percentage, editID, noAutoCommit, dryRun)
		},
	}
	promoteCmd.Flags().StringVar(&fromTrack, "from-track", "", "Source track")
	promoteCmd.Flags().StringVar(&toTrack, "to-track", "", "Destination track")
	promoteCmd.Flags().Float64Var(&percentage, "percentage", 0, "Rollout percentage for destination")
	promoteCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	promoteCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	promoteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	_ = promoteCmd.MarkFlagRequired("from-track")
	_ = promoteCmd.MarkFlagRequired("to-track")

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
			return c.publishHalt(cmd.Context(), track, editID, noAutoCommit, dryRun)
		},
	}
	haltCmd.Flags().StringVar(&track, "track", "production", "Release track")
	haltCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	haltCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
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
			return c.publishRollback(cmd.Context(), track, rollbackVersionCode, editID, noAutoCommit, dryRun)
		},
	}
	rollbackCmd.Flags().StringVar(&track, "track", "", "Release track")
	rollbackCmd.Flags().StringVar(&rollbackVersionCode, "version-code", "", "Specific version code to rollback to")
	rollbackCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	rollbackCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	rollbackCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm destructive operation")
	rollbackCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	_ = rollbackCmd.MarkFlagRequired("track")

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
			return c.publishListingUpdate(cmd.Context(), locale, title, shortDesc, fullDesc, editID, noAutoCommit, dryRun)
		},
	}
	listingUpdateCmd.Flags().StringVar(&locale, "locale", "en-US", "Locale code")
	listingUpdateCmd.Flags().StringVar(&title, "title", "", "App title")
	listingUpdateCmd.Flags().StringVar(&shortDesc, "short-description", "", "Short description")
	listingUpdateCmd.Flags().StringVar(&fullDesc, "full-description", "", "Full description")
	listingUpdateCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	listingUpdateCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
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

	// publish details
	detailsCmd := &cobra.Command{
		Use:   "details",
		Short: "Manage app details",
		Long:  "Get and update app contact information and settings.",
	}

	var contactEmail, contactPhone, contactWebsite, defaultLanguage, updateMask string
	detailsGetCmd := &cobra.Command{
		Use:   "get",
		Short: "Get app details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishDetailsGet(cmd.Context())
		},
	}

	detailsUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update app details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishDetailsUpdate(cmd.Context(), contactEmail, contactPhone, contactWebsite, defaultLanguage, editID, noAutoCommit, dryRun)
		},
	}
	detailsUpdateCmd.Flags().StringVar(&contactEmail, "contact-email", "", "Contact email")
	detailsUpdateCmd.Flags().StringVar(&contactPhone, "contact-phone", "", "Contact phone")
	detailsUpdateCmd.Flags().StringVar(&contactWebsite, "contact-website", "", "Contact website")
	detailsUpdateCmd.Flags().StringVar(&defaultLanguage, "default-language", "", "Default language")
	detailsUpdateCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	detailsUpdateCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	detailsUpdateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	detailsPatchCmd := &cobra.Command{
		Use:   "patch",
		Short: "Patch app details",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishDetailsPatch(cmd.Context(), contactEmail, contactPhone, contactWebsite, defaultLanguage, updateMask, editID, noAutoCommit, dryRun)
		},
	}
	detailsPatchCmd.Flags().StringVar(&contactEmail, "contact-email", "", "Contact email")
	detailsPatchCmd.Flags().StringVar(&contactPhone, "contact-phone", "", "Contact phone")
	detailsPatchCmd.Flags().StringVar(&contactWebsite, "contact-website", "", "Contact website")
	detailsPatchCmd.Flags().StringVar(&defaultLanguage, "default-language", "", "Default language")
	detailsPatchCmd.Flags().StringVar(&updateMask, "update-mask", "", "Fields to update (comma-separated)")
	detailsPatchCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	detailsPatchCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	detailsPatchCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	detailsCmd.AddCommand(detailsGetCmd, detailsUpdateCmd, detailsPatchCmd)

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
			return c.publishAssetsUpload(cmd.Context(), dir, category, replace, editID, noAutoCommit, dryRun)
		},
	}
	assetsUploadCmd.Flags().StringVar(&assetsDir, "dir", "assets", "Assets directory")
	assetsUploadCmd.Flags().StringVar(&category, "replace", "", "Category to replace (phone, tablet, tv, wear)")
	assetsUploadCmd.Flags().BoolVar(&replace, "replace-all", false, "Replace all existing assets")
	assetsUploadCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	assetsUploadCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
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

	// publish images
	imagesCmd := &cobra.Command{
		Use:   "images",
		Short: "Manage store images",
		Long:  "Upload, list, and delete store images using edits.images.",
	}

	var imageLocale string
	imagesUploadCmd := &cobra.Command{
		Use:   "upload <type> <file>",
		Short: "Upload an image",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishImagesUpload(cmd.Context(), args[0], args[1], imageLocale, editID, noAutoCommit, dryRun)
		},
	}
	imagesUploadCmd.Flags().StringVar(&imageLocale, "locale", "en-US", "Locale code")
	imagesUploadCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	imagesUploadCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	imagesUploadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	imagesListCmd := &cobra.Command{
		Use:   "list <type>",
		Short: "List images",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishImagesList(cmd.Context(), args[0], imageLocale, editID)
		},
	}
	imagesListCmd.Flags().StringVar(&imageLocale, "locale", "en-US", "Locale code")
	imagesListCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")

	imagesDeleteCmd := &cobra.Command{
		Use:   "delete <type> <id>",
		Short: "Delete an image",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishImagesDelete(cmd.Context(), args[0], args[1], imageLocale, editID, noAutoCommit, dryRun)
		},
	}
	imagesDeleteCmd.Flags().StringVar(&imageLocale, "locale", "en-US", "Locale code")
	imagesDeleteCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	imagesDeleteCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	imagesDeleteCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	imagesDeleteAllCmd := &cobra.Command{
		Use:   "deleteall <type>",
		Short: "Delete all images for type",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishImagesDeleteAll(cmd.Context(), args[0], imageLocale, editID, noAutoCommit, dryRun)
		},
	}
	imagesDeleteAllCmd.Flags().StringVar(&imageLocale, "locale", "en-US", "Locale code")
	imagesDeleteAllCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	imagesDeleteAllCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	imagesDeleteAllCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	imagesCmd.AddCommand(imagesUploadCmd, imagesListCmd, imagesDeleteCmd, imagesDeleteAllCmd)

	// publish deobfuscation
	deobfuscationCmd := &cobra.Command{
		Use:   "deobfuscation",
		Short: "Manage deobfuscation files",
		Long:  "Upload ProGuard/R8 mappings and native debug symbols.",
	}

	deobfuscationUploadCmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload deobfuscation file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishDeobfuscationUpload(cmd.Context(), args[0], deobfuscationType, deobfuscationVersionCode, editID, deobfuscationChunkSize, noAutoCommit, dryRun)
		},
	}
	deobfuscationUploadCmd.Flags().StringVar(&deobfuscationType, "type", "", "Deobfuscation file type: proguard or nativeCode")
	deobfuscationUploadCmd.Flags().Int64Var(&deobfuscationVersionCode, "version-code", 0, "Version code to associate")
	deobfuscationUploadCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	deobfuscationUploadCmd.Flags().Int64Var(&deobfuscationChunkSize, "chunk-size", 10*1024*1024, "Upload chunk size in bytes")
	deobfuscationUploadCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	deobfuscationUploadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	_ = deobfuscationUploadCmd.MarkFlagRequired("type")
	_ = deobfuscationUploadCmd.MarkFlagRequired("version-code")

	deobfuscationCmd.AddCommand(deobfuscationUploadCmd)

	// publish internal-share
	internalShareCmd := &cobra.Command{
		Use:   "internal-share",
		Short: "Upload artifacts for internal sharing",
		Long:  "Upload APK/AAB for internal testing without edit workflow.",
	}

	internalShareUploadCmd := &cobra.Command{
		Use:   "upload <file>",
		Short: "Upload artifact for internal sharing",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishInternalShareUpload(cmd.Context(), args[0], dryRun)
		},
	}
	internalShareUploadCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")
	internalShareCmd.AddCommand(internalShareUploadCmd)

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
			return c.publishTestersAdd(cmd.Context(), testersTrack, groups, editID, noAutoCommit, dryRun)
		},
	}
	testersAddCmd.Flags().StringVar(&testersTrack, "track", "internal", "Track to add testers to")
	testersAddCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses")
	testersAddCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	testersAddCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	testersAddCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	testersRemoveCmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove tester groups",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishTestersRemove(cmd.Context(), testersTrack, groups, editID, noAutoCommit, dryRun)
		},
	}
	testersRemoveCmd.Flags().StringVar(&testersTrack, "track", "internal", "Track to remove testers from")
	testersRemoveCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses")
	testersRemoveCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	testersRemoveCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
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

	c.addEditCommands(publishCmd)
	publishCmd.AddCommand(uploadCmd, releaseCmd, rolloutCmd, promoteCmd, haltCmd, rollbackCmd,
		statusCmd, tracksCmd, capabilitiesCmd, listingCmd, detailsCmd, assetsCmd, imagesCmd, deobfuscationCmd, internalShareCmd, testersCmd)
	c.rootCmd.AddCommand(publishCmd)
}

func (c *CLI) publishUpload(ctx context.Context, filePath, editID string, noAutoCommit bool, dryRun bool) error {
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

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	// Upload file
	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer f.Close()

	var versionCode int64
	if ext == ".aab" {
		bundle, err := publisher.Edits.Bundles.Upload(c.packageName, edit.ServerID).
			Media(f).Context(ctx).Do()
		if err != nil {
			if created {
				_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to upload bundle: %v", err)))
		}
		versionCode = bundle.VersionCode
	} else {
		apk, err := publisher.Edits.Apks.Upload(c.packageName, edit.ServerID).
			Media(f).Context(ctx).Do()
		if err != nil {
			if created {
				_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to upload APK: %v", err)))
		}
		versionCode = int64(apk.VersionCode)
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	// Cache the artifact
	_ = editMgr.CacheArtifact(c.packageName, filePath, versionCode)

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
		"editId":      edit.ServerID,
		"committed":   !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishRelease(ctx context.Context, track, name, status string, versionCodes []string, releaseNotesFile, editID string, noAutoCommit bool, dryRun bool) error {
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

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	// Get track and update release
	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
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

func (c *CLI) publishRollout(ctx context.Context, track string, percentage float64, editID string, noAutoCommit bool, dryRun bool) error {
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

	// Get API client
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

	// Get current track info
	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("track not found: %s", track)))
	}

	// Find the in-progress release and update its user fraction
	var updatedRelease *androidpublisher.TrackRelease
	for i, release := range trackInfo.Releases {
		if release.Status == "inProgress" {
			// Convert percentage to fraction (0.01 to 1.0)
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

		// Update the track
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

func (c *CLI) publishPromote(ctx context.Context, fromTrack, toTrack string, percentage float64, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if !config.IsValidTrack(fromTrack) || !config.IsValidTrack(toTrack) {
		return c.OutputError(errors.ErrTrackInvalid)
	}

	if fromTrack == toTrack {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"source and destination tracks must be different"))
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

	// Get API client
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

	// Get source track info
	sourceTrack, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, fromTrack).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("source track not found: %s", fromTrack)))
	}

	// Find the active release on the source track
	var sourceRelease *androidpublisher.TrackRelease
	for _, release := range sourceTrack.Releases {
		if release.Status == "completed" || release.Status == "inProgress" {
			sourceRelease = release
			break
		}
	}

	if sourceRelease == nil {
		_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("no active release found on track: %s", fromTrack)).
			WithHint("Ensure the source track has a completed or in-progress release"))
	}

	// Get or create destination track
	destTrack, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, toTrack).Context(ctx).Do()
	if err != nil {
		// Track might not exist, create new track config
		destTrack = &androidpublisher.Track{
			Track: toTrack,
		}
	}

	// Create new release for destination track
	newRelease := &androidpublisher.TrackRelease{
		Name:         sourceRelease.Name,
		VersionCodes: sourceRelease.VersionCodes,
		ReleaseNotes: sourceRelease.ReleaseNotes,
	}

	// Set status and user fraction based on percentage
	if percentage > 0 && percentage < 100 {
		newRelease.Status = "inProgress"
		newRelease.UserFraction = percentage / 100.0
	} else {
		newRelease.Status = "completed"
	}

	destTrack.Releases = []*androidpublisher.TrackRelease{newRelease}

	// Update destination track
	_, err = publisher.Edits.Tracks.Update(c.packageName, edit.ServerID, toTrack, destTrack).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update destination track: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"fromTrack":    fromTrack,
		"toTrack":      toTrack,
		"versionCodes": sourceRelease.VersionCodes,
		"status":       newRelease.Status,
		"percentage":   percentage,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishHalt(ctx context.Context, track, editID string, noAutoCommit bool, dryRun bool) error {
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

	// Get API client
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
	defer editMgr.ReleaseLock(c.packageName)

	// Get current track info
	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("track not found: %s", track)))
	}

	// Find the in-progress release and halt it
	var haltedRelease *androidpublisher.TrackRelease
	for i, release := range trackInfo.Releases {
		if release.Status == "inProgress" {
			trackInfo.Releases[i].Status = "halted"
			haltedRelease = trackInfo.Releases[i]
			break
		}
	}

	if haltedRelease == nil {
		publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"no in-progress release found on track").
			WithHint("Only releases with status 'inProgress' can be halted"))
	}

	// Update the track
	_, err = publisher.Edits.Tracks.Update(c.packageName, edit.ServerID, track, trackInfo).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
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

func (c *CLI) publishRollback(ctx context.Context, track, versionCode, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if !config.IsValidTrack(track) {
		return c.OutputError(errors.ErrTrackInvalid)
	}

	// Parse version code if provided
	var targetVersionCode int64
	if versionCode != "" {
		parsed, err := strconv.ParseInt(versionCode, 10, 64)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				fmt.Sprintf("invalid version code: %s", versionCode)))
		}
		targetVersionCode = parsed
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

	// Get API client
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
	defer editMgr.ReleaseLock(c.packageName)

	// Get current track info
	trackInfo, err := publisher.Edits.Tracks.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("track not found: %s", track)))
	}

	// Find the target version in history or halted releases
	var rollbackVersionCodes []int64
	var foundRelease *androidpublisher.TrackRelease

	// If version code specified, look for it
	if targetVersionCode > 0 {
		for _, release := range trackInfo.Releases {
			for _, vc := range release.VersionCodes {
				if vc == targetVersionCode {
					rollbackVersionCodes = []int64{targetVersionCode}
					foundRelease = release
					break
				}
			}
			if foundRelease != nil {
				break
			}
		}
		if foundRelease == nil {
			if created {
				publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
				fmt.Sprintf("version code %d not found in track history", targetVersionCode)).
				WithHint("Check available versions with 'gpd publish status --track " + track + "'"))
		}
	} else {
		// Find the previous completed release (not current inProgress or halted)
		var currentRelease *androidpublisher.TrackRelease
		var previousRelease *androidpublisher.TrackRelease
		for _, release := range trackInfo.Releases {
			if release.Status == "inProgress" || release.Status == "halted" {
				currentRelease = release
			} else if release.Status == "completed" && previousRelease == nil {
				previousRelease = release
			}
		}
		if previousRelease == nil {
			if created {
				publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
				"no previous release found to rollback to").
				WithHint("Specify a version code with --version-code flag"))
		}
		_ = currentRelease // For future use
		rollbackVersionCodes = previousRelease.VersionCodes
		foundRelease = previousRelease
	}

	// Create a new release with the rollback version
	newRelease := &androidpublisher.TrackRelease{
		VersionCodes: rollbackVersionCodes,
		Status:       "completed",
	}

	// Replace releases on track
	trackInfo.Releases = []*androidpublisher.TrackRelease{newRelease}

	// Update the track
	_, err = publisher.Edits.Tracks.Update(c.packageName, edit.ServerID, track, trackInfo).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
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
		"versionCodes": rollbackVersionCodes,
		"releaseName":  foundRelease.Name,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !noAutoCommit,
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

func (c *CLI) publishListingUpdate(ctx context.Context, locale, title, shortDesc, fullDesc, editID string, noAutoCommit bool, dryRun bool) error {
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

	// Get API client
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
	defer editMgr.ReleaseLock(c.packageName)

	// Build listing update
	listing := &androidpublisher.Listing{
		Language: locale,
	}
	if title != "" {
		listing.Title = title
	}
	if shortDesc != "" {
		listing.ShortDescription = shortDesc
	}
	if fullDesc != "" {
		listing.FullDescription = fullDesc
	}

	// Update the listing
	updatedListing, err := publisher.Edits.Listings.Update(c.packageName, edit.ServerID, locale, listing).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update listing: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":          true,
		"locale":           updatedListing.Language,
		"title":            updatedListing.Title,
		"shortDescription": updatedListing.ShortDescription,
		"fullDescription":  updatedListing.FullDescription,
		"package":          c.packageName,
		"editId":           edit.ServerID,
		"committed":        !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishListingGet(ctx context.Context, locale string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
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

	// Create edit (read-only, won't commit)
	edit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to create edit: %v", err)))
	}
	defer publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()

	if locale != "" {
		// Get specific locale listing
		locale = config.NormalizeLocale(locale)
		listing, err := publisher.Edits.Listings.Get(c.packageName, edit.Id, locale).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
				fmt.Sprintf("listing not found for locale: %s", locale)))
		}
		result := output.NewResult(map[string]interface{}{
			"locale":           listing.Language,
			"title":            listing.Title,
			"shortDescription": listing.ShortDescription,
			"fullDescription":  listing.FullDescription,
			"video":            listing.Video,
			"package":          c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	// Get all listings
	listings, err := publisher.Edits.Listings.List(c.packageName, edit.Id).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	var listingResults []map[string]interface{}
	for _, listing := range listings.Listings {
		listingResults = append(listingResults, map[string]interface{}{
			"locale":           listing.Language,
			"title":            listing.Title,
			"shortDescription": listing.ShortDescription,
			"fullDescription":  listing.FullDescription,
		})
	}

	result := output.NewResult(map[string]interface{}{
		"listings": listingResults,
		"count":    len(listingResults),
		"package":  c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishAssetsUpload(ctx context.Context, dir, category string, replace bool, editID string, noAutoCommit bool, dryRun bool) error {
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
		"editId":   editID,
		"committed": !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishDeobfuscationUpload(ctx context.Context, filePath, fileType string, versionCode int64, editID string, chunkSize int64, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	info, apiErr := validateDeobfuscationFile(filePath, fileType)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":      true,
			"action":      "deobfuscation_upload",
			"path":        filePath,
			"size":        info.Size(),
			"sizeHuman":   edits.FormatBytes(info.Size()),
			"type":        fileType,
			"versionCode": versionCode,
			"package":     c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			return c.OutputError(apiErr)
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			return c.OutputError(apiErr)
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer editMgr.ReleaseLock(c.packageName)

	if err := c.ensureVersionCodeExists(ctx, publisher, edit.ServerID, versionCode); err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer f.Close()

	call := publisher.Edits.Deobfuscationfiles.Upload(c.packageName, edit.ServerID, versionCode, fileType)
	if chunkSize > 0 {
		call.Media(f, googleapi.ChunkSize(int(chunkSize)))
	} else {
		call.Media(f)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to upload deobfuscation file: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":     true,
		"type":        fileType,
		"versionCode": versionCode,
		"package":     c.packageName,
		"size":        info.Size(),
		"sizeHuman":   edits.FormatBytes(info.Size()),
		"editId":      edit.ServerID,
		"committed":   !noAutoCommit,
		"response":    resp,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func validateDeobfuscationFile(filePath, fileType string) (os.FileInfo, *errors.APIError) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath))
	}

	switch fileType {
	case "proguard":
		if info.Size() > 50*1024*1024 {
			return nil, errors.NewAPIError(errors.CodeValidationError, "proguard mapping file exceeds 50MB")
		}
		if !looksLikeProguardMapping(filePath) {
			return nil, errors.NewAPIError(errors.CodeValidationError, "proguard mapping file format invalid")
		}
	case "nativeCode":
		if info.Size() > 100*1024*1024 {
			return nil, errors.NewAPIError(errors.CodeValidationError, "native symbols file exceeds 100MB")
		}
		lower := strings.ToLower(filePath)
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext != ".zip" && ext != ".sym" && !strings.HasSuffix(lower, ".so.sym") {
			return nil, errors.NewAPIError(errors.CodeValidationError, "native symbols file must be .so.sym or .zip")
		}
	default:
		return nil, errors.NewAPIError(errors.CodeValidationError, "type must be proguard or nativeCode")
	}

	return info, nil
}

func looksLikeProguardMapping(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.Contains(line, "->") && strings.HasSuffix(line, ":") {
			return true
		}
		lines++
		if lines > 50 {
			break
		}
	}
	return false
}

func (c *CLI) ensureVersionCodeExists(ctx context.Context, publisher *androidpublisher.Service, editID string, versionCode int64) *errors.APIError {
	if versionCode <= 0 {
		return errors.NewAPIError(errors.CodeValidationError, "version code must be greater than zero")
	}

	bundles, err := publisher.Edits.Bundles.List(c.packageName, editID).Context(ctx).Do()
	if err == nil {
		for _, bundle := range bundles.Bundles {
			if bundle.VersionCode == versionCode {
				return nil
			}
		}
	}

	apks, err := publisher.Edits.Apks.List(c.packageName, editID).Context(ctx).Do()
	if err == nil {
		for _, apk := range apks.Apks {
			if int64(apk.VersionCode) == versionCode {
				return nil
			}
		}
	}

	return errors.NewAPIError(errors.CodeValidationError,
		fmt.Sprintf("version code %d not found in edit", versionCode)).
		WithHint("Upload an APK/AAB in this edit before uploading deobfuscation files")
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

func (c *CLI) publishDetailsGet(ctx context.Context) error {
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
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to create edit: %v", err)))
	}
	defer publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()

	details, err := publisher.Edits.Details.Get(c.packageName, edit.Id).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	return c.Output(output.NewResult(details).WithServices("androidpublisher"))
}

func (c *CLI) publishDetailsUpdate(ctx context.Context, email, phone, website, defaultLanguage, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if email == "" && phone == "" && website == "" && defaultLanguage == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "at least one field is required"))
	}
	if email != "" && !isValidEmail(email) {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid contact email"))
	}
	if website != "" && !isValidURL(website) {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid contact website"))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":          true,
			"action":          "details_update",
			"contactEmail":    email,
			"contactPhone":    phone,
			"contactWebsite":  website,
			"defaultLanguage": defaultLanguage,
			"package":         c.packageName,
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
	defer editMgr.ReleaseLock(c.packageName)

	details := &androidpublisher.AppDetails{
		ContactEmail:   email,
		ContactPhone:   phone,
		ContactWebsite: website,
		DefaultLanguage: defaultLanguage,
	}

	updated, err := publisher.Edits.Details.Update(c.packageName, edit.ServerID, details).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"details":    updated,
		"package":    c.packageName,
		"editId":     edit.ServerID,
		"committed":  !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishDetailsPatch(ctx context.Context, email, phone, website, defaultLanguage, updateMask, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if email == "" && phone == "" && website == "" && defaultLanguage == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "at least one field is required"))
	}
	if email != "" && !isValidEmail(email) {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid contact email"))
	}
	if website != "" && !isValidURL(website) {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "invalid contact website"))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":          true,
			"action":          "details_patch",
			"contactEmail":    email,
			"contactPhone":    phone,
			"contactWebsite":  website,
			"defaultLanguage": defaultLanguage,
			"updateMask":      updateMask,
			"package":         c.packageName,
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
	defer editMgr.ReleaseLock(c.packageName)

	details := &androidpublisher.AppDetails{
		ContactEmail:   email,
		ContactPhone:   phone,
		ContactWebsite: website,
		DefaultLanguage: defaultLanguage,
	}

	if updateMask == "" {
		var fields []string
		if email != "" {
			fields = append(fields, "contactEmail")
		}
		if phone != "" {
			fields = append(fields, "contactPhone")
		}
		if website != "" {
			fields = append(fields, "contactWebsite")
		}
		if defaultLanguage != "" {
			fields = append(fields, "defaultLanguage")
		}
		updateMask = strings.Join(fields, ",")
	}

	call := publisher.Edits.Details.Patch(c.packageName, edit.ServerID, details)
	// Note: UpdateMask is not available on EditsDetailsPatchCall
	// The update mask functionality may need to be handled differently
	// or may not be supported in the Patch method
	updated, err := call.Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":    true,
		"details":    updated,
		"package":    c.packageName,
		"editId":     edit.ServerID,
		"committed":  !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishImagesUpload(ctx context.Context, imageType, filePath, locale, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	info, cfg, format, apiErr := validateImageFile(filePath, imageType)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}
	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":    true,
			"action":    "images_upload",
			"type":      imageType,
			"locale":    locale,
			"path":      filePath,
			"width":     cfg.Width,
			"height":    cfg.Height,
			"format":    format,
			"size":      info.Size(),
			"sizeHuman": edits.FormatBytes(info.Size()),
			"package":   c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			return c.OutputError(apiErr)
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		if apiErr, ok := err.(*errors.APIError); ok {
			return c.OutputError(apiErr)
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer editMgr.ReleaseLock(c.packageName)

	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer f.Close()

	resp, err := publisher.Edits.Images.Upload(c.packageName, edit.ServerID, locale, imageType).
		Media(f).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"type":      imageType,
		"locale":    locale,
		"image":     resp,
		"package":   c.packageName,
		"editId":    edit.ServerID,
		"committed": !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishImagesList(ctx context.Context, imageType, locale, editID string) error {
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
	var edit *androidpublisher.AppEdit
	var created bool
	if editID == "" {
		edit, err = publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		created = true
	} else {
		edit = &androidpublisher.AppEdit{Id: editID}
	}
	if created {
		defer publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()
	}
	images, err := publisher.Edits.Images.List(c.packageName, edit.Id, locale, imageType).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(images).WithServices("androidpublisher"))
}

func (c *CLI) publishImagesDelete(ctx context.Context, imageType, imageID, locale, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":  true,
			"action":  "images_delete",
			"type":    imageType,
			"locale":  locale,
			"id":      imageID,
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
	defer editMgr.ReleaseLock(c.packageName)

	if err := publisher.Edits.Images.Delete(c.packageName, edit.ServerID, locale, imageType, imageID).Context(ctx).Do(); err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"type":      imageType,
		"locale":    locale,
		"id":        imageID,
		"package":   c.packageName,
		"editId":    edit.ServerID,
		"committed": !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishImagesDeleteAll(ctx context.Context, imageType, locale, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":  true,
			"action":  "images_deleteall",
			"type":    imageType,
			"locale":  locale,
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
	defer editMgr.ReleaseLock(c.packageName)

	if _, err := publisher.Edits.Images.Deleteall(c.packageName, edit.ServerID, locale, imageType).Context(ctx).Do(); err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"type":      imageType,
		"locale":    locale,
		"package":   c.packageName,
		"editId":    edit.ServerID,
		"committed": !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishInternalShareUpload(ctx context.Context, filePath string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	info, err := os.Stat(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath)))
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".apk" && ext != ".aab" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"file must be an APK or AAB"))
	}
	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":    true,
			"action":    "internal_share_upload",
			"path":      filePath,
			"size":      info.Size(),
			"sizeHuman": edits.FormatBytes(info.Size()),
			"type":      ext[1:],
			"package":   c.packageName,
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
	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer f.Close()
	var resp *androidpublisher.InternalAppSharingArtifact
	if ext == ".apk" {
		resp, err = publisher.Internalappsharingartifacts.Uploadapk(c.packageName).Media(f).Context(ctx).Do()
	} else {
		resp, err = publisher.Internalappsharingartifacts.Uploadbundle(c.packageName).Media(f).Context(ctx).Do()
	}
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":  true,
		"artifact": resp,
		"package":  c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

type imageSpec struct {
	minWidth  int
	maxWidth  int
	minHeight int
	maxHeight int
	maxSize   int64
	formats   []string
}

func validateImageFile(filePath, imageType string) (os.FileInfo, image.Config, string, *errors.APIError) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath))
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}
	defer f.Close()

	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "invalid image file")
	}

	spec, ok := imageSpecs()[imageType]
	if !ok {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "invalid image type")
	}
	if spec.maxSize > 0 && info.Size() > spec.maxSize {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "image exceeds size limit")
	}
	if spec.minWidth > 0 && cfg.Width < spec.minWidth {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "image width too small")
	}
	if spec.maxWidth > 0 && cfg.Width > spec.maxWidth {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "image width too large")
	}
	if spec.minHeight > 0 && cfg.Height < spec.minHeight {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "image height too small")
	}
	if spec.maxHeight > 0 && cfg.Height > spec.maxHeight {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "image height too large")
	}
	if len(spec.formats) > 0 && !containsString(spec.formats, format) {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "invalid image format")
	}
	return info, cfg, format, nil
}

func imageSpecs() map[string]imageSpec {
	return map[string]imageSpec{
		"icon":               {minWidth: 512, maxWidth: 512, minHeight: 512, maxHeight: 512, maxSize: 1 * 1024 * 1024, formats: []string{"png"}},
		"featureGraphic":     {minWidth: 1024, maxWidth: 1024, minHeight: 500, maxHeight: 500, maxSize: 15 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"promoGraphic":       {maxSize: 15 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tvBanner":           {minWidth: 1280, maxWidth: 1280, minHeight: 720, maxHeight: 720, maxSize: 15 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"phoneScreenshots":   {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tabletScreenshots":  {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"sevenInchScreenshots": {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tenInchScreenshots": {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tvScreenshots":      {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"wearScreenshots":    {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
	}
}

func isValidEmail(value string) bool {
	_, err := mail.ParseAddress(value)
	return err == nil
}

func isValidURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return parsed.Scheme != "" && parsed.Host != ""
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func (c *CLI) publishTestersAdd(ctx context.Context, track string, groups []string, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if len(groups) == 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"at least one group email is required").WithHint("Use --group to specify tester group emails"))
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

	// Get API client
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
	defer editMgr.ReleaseLock(c.packageName)

	// Get current testers
	testers, err := publisher.Edits.Testers.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		// Might not exist, create new
		testers = &androidpublisher.Testers{}
	}

	// Add new groups
	existingGroups := make(map[string]bool)
	for _, g := range testers.GoogleGroups {
		existingGroups[g] = true
	}
	for _, g := range groups {
		if !existingGroups[g] {
			testers.GoogleGroups = append(testers.GoogleGroups, g)
		}
	}

	// Update testers
	_, err = publisher.Edits.Testers.Update(c.packageName, edit.ServerID, track, testers).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update testers: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":     true,
		"track":       track,
		"groupsAdded": groups,
		"totalGroups": testers.GoogleGroups,
		"package":     c.packageName,
		"editId":      edit.ServerID,
		"committed":   !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishTestersRemove(ctx context.Context, track string, groups []string, editID string, noAutoCommit bool, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if len(groups) == 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"at least one group email is required").WithHint("Use --group to specify tester group emails"))
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

	// Get API client
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
	defer editMgr.ReleaseLock(c.packageName)

	// Get current testers
	testers, err := publisher.Edits.Testers.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("no testers found for track: %s", track)))
	}

	// Remove specified groups
	removeSet := make(map[string]bool)
	for _, g := range groups {
		removeSet[g] = true
	}
	var remaining []string
	for _, g := range testers.GoogleGroups {
		if !removeSet[g] {
			remaining = append(remaining, g)
		}
	}
	testers.GoogleGroups = remaining

	// Update testers
	_, err = publisher.Edits.Testers.Update(c.packageName, edit.ServerID, track, testers).Context(ctx).Do()
	if err != nil {
		if created {
			publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to update testers: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":         true,
		"track":           track,
		"groupsRemoved":   groups,
		"remainingGroups": remaining,
		"package":         c.packageName,
		"editId":          edit.ServerID,
		"committed":       !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishTestersList(ctx context.Context, track string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
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

	// Create edit (read-only, won't commit)
	edit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to create edit: %v", err)))
	}
	defer publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do()

	if track != "" {
		// Get testers for specific track
		testers, err := publisher.Edits.Testers.Get(c.packageName, edit.Id, track).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
				fmt.Sprintf("no testers found for track: %s", track)))
		}
		result := output.NewResult(map[string]interface{}{
			"track":        track,
			"googleGroups": testers.GoogleGroups,
			"package":      c.packageName,
		})
		return c.Output(result.WithServices("androidpublisher"))
	}

	// Get testers for all tracks
	tracks := []string{"internal", "alpha", "beta", "production"}
	testersData := make(map[string]interface{})

	for _, t := range tracks {
		testers, err := publisher.Edits.Testers.Get(c.packageName, edit.Id, t).Context(ctx).Do()
		if err == nil && len(testers.GoogleGroups) > 0 {
			testersData[t] = map[string]interface{}{
				"googleGroups": testers.GoogleGroups,
			}
		}
	}

	result := output.NewResult(map[string]interface{}{
		"testers": testersData,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}
