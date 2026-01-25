package cli

import (
	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/errors"
)

func (c *CLI) addPublishCommands() {
	publishCmd := &cobra.Command{
		Use:   "publish",
		Short: "Publishing commands",
		Long:  "Manage app publishing, releases, and store listings.",
	}

	c.addPublishUploadCommands(publishCmd)
	c.addPublishReleaseCommands(publishCmd)
	c.addPublishStatusCommands(publishCmd)
	c.addPublishListingCommands(publishCmd)
	c.addPublishDetailsCommands(publishCmd)
	c.addPublishAssetsCommands(publishCmd)
	c.addPublishImagesCommands(publishCmd)
	c.addPublishDeobfuscationCommands(publishCmd)
	c.addPublishInternalShareCommands(publishCmd)
	c.addPublishTestersCommands(publishCmd)
	c.addEditCommands(publishCmd)

	c.rootCmd.AddCommand(publishCmd)
}

func (c *CLI) addPublishUploadCommands(publishCmd *cobra.Command) {
	var (
		editID       string
		dryRun       bool
		noAutoCommit bool
	)

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

	publishCmd.AddCommand(uploadCmd)
}

func (c *CLI) addPublishReleaseCommands(publishCmd *cobra.Command) {
	var (
		editID              string
		track               string
		status              string
		name                string
		versionCodes        []string
		percentage          float64
		releaseNotesFile    string
		confirm             bool
		dryRun              bool
		noAutoCommit        bool
		fromTrack           string
		toTrack             string
		rollbackVersionCode string
		wait                bool
		waitTimeout         string
	)

	releaseCmd := &cobra.Command{
		Use:   "release",
		Short: "Create or update a release",
		Long:  "Create a new release on a track with specified version codes.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishRelease(cmd.Context(), track, name, status, versionCodes, releaseNotesFile, editID, noAutoCommit, dryRun, wait, waitTimeout)
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
	releaseCmd.Flags().BoolVar(&wait, "wait", false, "Wait for release to complete")
	releaseCmd.Flags().StringVar(&waitTimeout, "wait-timeout", "30m", "Maximum time to wait (e.g., 30m, 1h)")
	_ = releaseCmd.MarkFlagRequired("track")

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

	publishCmd.AddCommand(releaseCmd, rolloutCmd, promoteCmd, haltCmd, rollbackCmd)
}

func (c *CLI) addPublishStatusCommands(publishCmd *cobra.Command) {
	var track string

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Get track status",
		Long:  "Get the current status of a release track.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishStatus(cmd.Context(), track)
		},
	}
	statusCmd.Flags().StringVar(&track, "track", "", "Release track (leave empty for all tracks)")

	tracksCmd := &cobra.Command{
		Use:   "tracks",
		Short: "List all tracks",
		Long:  "List all available release tracks and their status.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishTracks(cmd.Context())
		},
	}

	capabilitiesCmd := &cobra.Command{
		Use:   "capabilities",
		Short: "List publishing capabilities",
		Long:  "List available publishing operations and constraints.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishCapabilities(cmd.Context())
		},
	}

	publishCmd.AddCommand(statusCmd, tracksCmd, capabilitiesCmd)
}

func (c *CLI) addPublishListingCommands(publishCmd *cobra.Command) {
	listingCmd := &cobra.Command{
		Use:   "listing",
		Short: "Manage store listing",
		Long:  "Update app title, short description, and full description.",
	}

	var (
		locale       string
		title        string
		shortDesc    string
		fullDesc     string
		editID       string
		noAutoCommit bool
		dryRun       bool
	)

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
	publishCmd.AddCommand(listingCmd)
}

func (c *CLI) addPublishDetailsCommands(publishCmd *cobra.Command) {
	detailsCmd := &cobra.Command{
		Use:   "details",
		Short: "Manage app details",
		Long:  "Get and update app contact information and settings.",
	}

	var (
		contactEmail    string
		contactPhone    string
		contactWebsite  string
		defaultLanguage string
		updateMask      string
		editID          string
		noAutoCommit    bool
		dryRun          bool
	)

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
	publishCmd.AddCommand(detailsCmd)
}

func (c *CLI) addPublishAssetsCommands(publishCmd *cobra.Command) {
	assetsCmd := &cobra.Command{
		Use:   "assets",
		Short: "Manage store assets",
		Long:  "Upload and manage screenshots and graphics.",
	}

	var (
		assetsDir    string
		category     string
		replace      bool
		editID       string
		noAutoCommit bool
		dryRun       bool
	)

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
	publishCmd.AddCommand(assetsCmd)
}

func (c *CLI) addPublishImagesCommands(publishCmd *cobra.Command) {
	imagesCmd := &cobra.Command{
		Use:   "images",
		Short: "Manage store images",
		Long:  "Upload, list, and delete store images using edits.images.",
	}

	var (
		imageLocale  string
		editID       string
		noAutoCommit bool
		dryRun       bool
	)

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
	publishCmd.AddCommand(imagesCmd)
}

func (c *CLI) addPublishDeobfuscationCommands(publishCmd *cobra.Command) {
	deobfuscationCmd := &cobra.Command{
		Use:   "deobfuscation",
		Short: "Manage deobfuscation files",
		Long:  "Upload ProGuard/R8 mappings and native debug symbols.",
	}

	var (
		deobfuscationType        string
		deobfuscationVersionCode int64
		deobfuscationChunkSize   int64
		editID                   string
		noAutoCommit             bool
		dryRun                   bool
	)

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
	publishCmd.AddCommand(deobfuscationCmd)
}

func (c *CLI) addPublishInternalShareCommands(publishCmd *cobra.Command) {
	internalShareCmd := &cobra.Command{
		Use:     "internal-share",
		Aliases: []string{"internal", "share"},
		Short:   "Upload artifacts for internal sharing",
		Long:    "Upload APK/AAB for internal testing without edit workflow.",
	}

	var dryRun bool

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
	publishCmd.AddCommand(internalShareCmd)
}

func (c *CLI) addPublishTestersCommands(publishCmd *cobra.Command) {
	testersCmd := &cobra.Command{
		Use:   "testers",
		Short: "Manage testers",
		Long:  "Manage tester groups for tracks.",
	}

	var (
		testersTrack string
		groups       []string
		editID       string
		noAutoCommit bool
		dryRun       bool
	)

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
	publishCmd.AddCommand(testersCmd)
}
