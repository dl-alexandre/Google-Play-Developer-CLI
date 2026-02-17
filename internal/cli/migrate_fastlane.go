package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/logging"
	"github.com/dl-alexandre/gpd/internal/migrate"
	"github.com/dl-alexandre/gpd/internal/migrate/fastlane"
	"github.com/dl-alexandre/gpd/internal/output"
)

func validateURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("only https URLs are allowed")
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	host := u.Hostname()
	ip := net.ParseIP(host)
	if ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() {
			return fmt.Errorf("internal/private IP addresses are not allowed")
		}
	}
	return nil
}

var fastlaneImageTypes = []string{
	fastlaneImageIcon,
	fastlaneImageFeatureGraphic,
	fastlaneImagePromoGraphic,
	fastlaneImageTvBanner,
	fastlaneImagePhoneScreenshots,
	fastlaneImageTabletScreenshots,
	fastlaneImageSevenInchScreenshots,
	fastlaneImageTenInchScreenshots,
	fastlaneImageTvScreenshots,
	fastlaneImageWearScreenshots,
}

var fastlaneSingleImageTypes = map[string]bool{
	fastlaneImageIcon:           true,
	fastlaneImageFeatureGraphic: true,
	fastlaneImagePromoGraphic:   true,
	fastlaneImageTvBanner:       true,
}

var fastlaneImageTypeSet = map[string]bool{
	fastlaneImageIcon:                 true,
	fastlaneImageFeatureGraphic:       true,
	fastlaneImagePromoGraphic:         true,
	fastlaneImageTvBanner:             true,
	fastlaneImagePhoneScreenshots:     true,
	fastlaneImageTabletScreenshots:    true,
	fastlaneImageSevenInchScreenshots: true,
	fastlaneImageTenInchScreenshots:   true,
	fastlaneImageTvScreenshots:        true,
	fastlaneImageWearScreenshots:      true,
}

const (
	fastlaneMetadataDir               = "fastlane/metadata/android"
	fastlaneImageIcon                 = "icon"
	fastlaneImageFeatureGraphic       = "featureGraphic"
	fastlaneImagePromoGraphic         = "promoGraphic"
	fastlaneImageTvBanner             = "tvBanner"
	fastlaneImagePhoneScreenshots     = "phoneScreenshots"
	fastlaneImageTabletScreenshots    = "tabletScreenshots"
	fastlaneImageSevenInchScreenshots = "sevenInchScreenshots"
	fastlaneImageTenInchScreenshots   = "tenInchScreenshots"
	fastlaneImageTvScreenshots        = "tvScreenshots"
	fastlaneImageWearScreenshots      = "wearScreenshots"
)

type changelogSet struct {
	defaultText string
	hasDefault  bool
	byVersion   map[int64]string
}

func (c *CLI) addMigrateFastlaneCommands(migrateCmd *cobra.Command) {
	fastlaneCmd := &cobra.Command{
		Use:   "fastlane",
		Short: "Fastlane supply format migration",
		Long:  "Migrate metadata between fastlane supply format and Google Play.",
	}

	var validateDir string
	validateCmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate fastlane metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.migrateFastlaneValidate(cmd.Context(), validateDir)
		},
	}
	validateCmd.Flags().StringVar(&validateDir, "dir", fastlaneMetadataDir, "Fastlane metadata directory")

	var (
		exportDir     string
		includeImages bool
		locales       []string
	)
	exportCmd := &cobra.Command{
		Use:   "export",
		Short: "Export metadata to fastlane format",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.migrateFastlaneExport(cmd.Context(), exportDir, includeImages, locales)
		},
	}
	exportCmd.Flags().StringVar(&exportDir, "output", fastlaneMetadataDir, "Output directory")
	exportCmd.Flags().BoolVar(&includeImages, "include-images", false, "Download and write images")
	exportCmd.Flags().StringSliceVar(&locales, "locales", nil, "Locales to export (comma-separated)")

	var (
		importDir     string
		skipImages    bool
		replaceImages bool
		syncImages    bool
		editID        string
		noAutoCommit  bool
		dryRun        bool
	)
	importCmd := &cobra.Command{
		Use:   "import",
		Short: "Import metadata from fastlane format",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.migrateFastlaneImport(cmd.Context(), importDir, skipImages, replaceImages, syncImages, editID, noAutoCommit, dryRun)
		},
	}
	importCmd.Flags().StringVar(&importDir, "dir", fastlaneMetadataDir, "Fastlane metadata directory")
	importCmd.Flags().BoolVar(&skipImages, "skip-images", false, "Skip uploading images")
	importCmd.Flags().BoolVar(&replaceImages, "replace-images", false, "Delete existing images before upload")
	importCmd.Flags().BoolVar(&syncImages, "sync-images", false, "Skip uploading images that already exist")
	importCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	importCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	importCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	fastlaneCmd.AddCommand(validateCmd, exportCmd, importCmd)
	migrateCmd.AddCommand(fastlaneCmd)
}

func (c *CLI) migrateFastlaneValidate(_ context.Context, dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "dir is required").
			WithHint("Provide --dir or use the default fastlane/metadata/android structure"))
		return c.Output(result.WithServices("migrate"))
	}

	metadata, err := fastlane.ParseDirectory(dir)
	if err != nil {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, err.Error()).
			WithHint("Ensure the fastlane metadata directory exists (fastlane/metadata/android) or use --dir"))
		return c.Output(result.WithServices("migrate"))
	}
	if len(metadata) == 0 {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError, "no locale directories found").
			WithHint("Fastlane metadata should include locale folders like fastlane/metadata/android/en-US"))
		return c.Output(result.WithServices("migrate"))
	}

	var allErrors []migrate.ValidationError
	locales := make([]string, 0, len(metadata))
	for i := range metadata {
		meta := &metadata[i]
		locales = append(locales, meta.Locale)
		errs := fastlane.ValidateLocale(meta)
		if len(errs) > 0 {
			allErrors = append(allErrors, errs...)
		}
	}
	sort.Strings(locales)

	result := output.NewResult(map[string]interface{}{
		"dir":         dir,
		"valid":       len(allErrors) == 0,
		"errors":      allErrors,
		"locales":     locales,
		"localeCount": len(locales),
	})
	return c.Output(result.WithServices("migrate"))
}

func (c *CLI) migrateFastlaneExport(ctx context.Context, outputDir string, includeImages bool, locales []string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	outputDir = strings.TrimSpace(outputDir)
	if outputDir == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "output directory is required"))
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
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to create edit: %v", err)))
	}
	defer func() {
		if err := publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do(); err != nil {
			logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.Id), logging.Err(err))
		}
	}()

	listings, err := publisher.Edits.Listings.List(c.packageName, edit.Id).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	filter := normalizeLocaleFilter(locales)
	var metadata []fastlane.LocaleMetadata
	for _, listing := range listings.Listings {
		locale := config.NormalizeLocale(listing.Language)
		if filter != nil && !filter[locale] {
			continue
		}
		meta := fastlane.LocaleMetadata{
			Locale:              locale,
			Title:               listing.Title,
			TitleSet:            true,
			ShortDescription:    listing.ShortDescription,
			ShortDescriptionSet: true,
			FullDescription:     listing.FullDescription,
			FullDescriptionSet:  true,
		}
		if listing.Video != "" {
			meta.Video = listing.Video
			meta.VideoSet = true
		}
		metadata = append(metadata, meta)
	}
	if len(metadata) == 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "no listings found for export"))
	}

	localeSet := map[string]bool{}
	for i := range metadata {
		localeSet[metadata[i].Locale] = true
	}
	changelogs, apiErr := c.exportFastlaneChangelogs(ctx, publisher, edit.Id, localeSet)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}
	if len(changelogs) > 0 {
		for i := range metadata {
			if notes, ok := changelogs[metadata[i].Locale]; ok && len(notes) > 0 {
				metadata[i].Changelogs = notes
			}
		}
	}

	if err := fastlane.WriteDirectory(outputDir, metadata); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	warnings := []string{}
	imagesOutput := map[string]map[string]int{}
	if includeImages {
		httpClient := &http.Client{Timeout: c.timeout}
		for i := range metadata {
			locale := metadata[i].Locale
			counts, warns := c.exportFastlaneImages(ctx, publisher, edit.Id, outputDir, locale, httpClient)
			if len(warns) > 0 {
				warnings = append(warnings, warns...)
			}
			if len(counts) > 0 {
				imagesOutput[locale] = counts
			}
		}
	}

	localesList := make([]string, 0, len(metadata))
	for i := range metadata {
		localesList = append(localesList, metadata[i].Locale)
	}
	sort.Strings(localesList)

	resultData := map[string]interface{}{
		"success":     true,
		"output":      outputDir,
		"package":     c.packageName,
		"locales":     localesList,
		"localeCount": len(localesList),
	}
	if includeImages {
		resultData["images"] = imagesOutput
	}

	result := output.NewResult(resultData).WithServices("androidpublisher")
	if len(warnings) > 0 {
		result.WithWarnings(warnings...)
	}
	return c.Output(result)
}

func (c *CLI) exportFastlaneImages(ctx context.Context, publisher *androidpublisher.Service, editID, outputDir, locale string, httpClient *http.Client) (counts map[string]int, warnings []string) {
	normalizedLocale := config.NormalizeLocale(locale)
	imagesDir := filepath.Join(outputDir, locale, "images")
	if err := os.MkdirAll(imagesDir, 0o755); err != nil {
		return nil, []string{fmt.Sprintf("failed to create images dir for %s: %v", locale, err)}
	}

	counts = map[string]int{}
	warnings = []string{}
	for _, imageType := range fastlaneImageTypes {
		images, err := publisher.Edits.Images.List(c.packageName, editID, normalizedLocale, imageType).Context(ctx).Do()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to list images for %s/%s: %v", locale, imageType, err))
			continue
		}
		if images == nil || len(images.Images) == 0 {
			continue
		}

		if fastlaneSingleImageTypes[imageType] {
			image := images.Images[0]
			if image.Url == "" {
				warnings = append(warnings, fmt.Sprintf("missing image url for %s/%s", locale, imageType))
				continue
			}
			destBase := filepath.Join(imagesDir, imageType)
			if err := downloadImage(ctx, httpClient, image.Url, destBase); err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to download %s/%s: %v", locale, imageType, err))
				continue
			}
			counts[imageType] = 1
			if len(images.Images) > 1 {
				warnings = append(warnings, fmt.Sprintf("multiple images found for %s/%s, exported first only", locale, imageType))
			}
			continue
		}

		typeDir := filepath.Join(imagesDir, imageType)
		if err := os.MkdirAll(typeDir, 0o755); err != nil {
			warnings = append(warnings, fmt.Sprintf("failed to create %s/%s dir: %v", locale, imageType, err))
			continue
		}
		for i, image := range images.Images {
			if image.Url == "" {
				warnings = append(warnings, fmt.Sprintf("missing image url for %s/%s index %d", locale, imageType, i+1))
				continue
			}
			destBase := filepath.Join(typeDir, fmt.Sprintf("%02d", i+1))
			if err := downloadImage(ctx, httpClient, image.Url, destBase); err != nil {
				warnings = append(warnings, fmt.Sprintf("failed to download %s/%s index %d: %v", locale, imageType, i+1, err))
				continue
			}
			counts[imageType]++
		}
	}
	return counts, warnings
}

func (c *CLI) migrateFastlaneImport(ctx context.Context, dir string, skipImages, replaceImages, syncImages bool, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	dir = strings.TrimSpace(dir)
	if dir == "" {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "dir is required"))
	}

	metadata, err := fastlane.ParseDirectory(dir)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, err.Error()))
	}
	if len(metadata) == 0 {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError, "no locale directories found"))
	}

	changelogSets := buildChangelogSets(metadata)

	if dryRun {
		resultData := c.buildFastlaneImportDryRun(dir, metadata, skipImages)
		return c.Output(output.NewResult(resultData))
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
	defer func() {
		if err := editMgr.ReleaseLock(c.packageName); err != nil {
			logging.Warn("failed to release edit lock", logging.String("package", c.packageName), logging.Err(err))
		}
	}()

	updatedListings, imageOutput, skippedImageOutput, apiErr := c.applyFastlaneImportMetadata(ctx, publisher, edit.ServerID, metadata, skipImages, replaceImages, syncImages)
	if apiErr != nil {
		c.cleanupEditOnError(ctx, publisher, edit.ServerID, created)
		return c.OutputError(apiErr)
	}

	updatedTracks := 0
	updatedReleases := 0
	if len(changelogSets) > 0 {
		trackUpdates, releaseUpdates, apiErr := c.importFastlaneReleaseNotes(ctx, publisher, edit.ServerID, changelogSets)
		if apiErr != nil {
			c.cleanupEditOnError(ctx, publisher, edit.ServerID, created)
			return c.OutputError(apiErr)
		}
		updatedTracks = trackUpdates
		updatedReleases = releaseUpdates
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	localesList := make([]string, 0, len(metadata))
	for i := range metadata {
		localesList = append(localesList, metadata[i].Locale)
	}
	sort.Strings(localesList)

	resultData := map[string]interface{}{
		"success":         true,
		"dir":             dir,
		"package":         c.packageName,
		"locales":         localesList,
		"localeCount":     len(localesList),
		"updatedListings": updatedListings,
		"updatedTracks":   updatedTracks,
		"updatedReleases": updatedReleases,
		"editId":          edit.ServerID,
		"committed":       !noAutoCommit,
	}
	if !skipImages {
		resultData["images"] = imageOutput
	}
	if len(skippedImageOutput) > 0 {
		resultData["skippedImages"] = skippedImageOutput
	}

	result := output.NewResult(resultData).WithServices("androidpublisher")
	return c.Output(result)
}

func (c *CLI) buildFastlaneImportDryRun(dir string, metadata []fastlane.LocaleMetadata, skipImages bool) map[string]interface{} {
	resultData := map[string]interface{}{
		"dryRun":      true,
		"action":      "migrate_fastlane_import",
		"dir":         dir,
		"package":     c.packageName,
		"localeCount": len(metadata),
	}
	localesList := make([]string, 0, len(metadata))
	imageCounts := map[string]map[string]int{}
	for i := range metadata {
		meta := &metadata[i]
		localesList = append(localesList, meta.Locale)
		if skipImages {
			continue
		}
		if len(meta.Images) == 0 {
			continue
		}
		counts := map[string]int{}
		for imageType, paths := range meta.Images {
			counts[imageType] = len(paths)
		}
		imageCounts[meta.Locale] = counts
	}
	sort.Strings(localesList)
	resultData["locales"] = localesList
	if !skipImages {
		resultData["images"] = imageCounts
	}
	return resultData
}

func (c *CLI) applyFastlaneImportMetadata(ctx context.Context, publisher *androidpublisher.Service, editID string, metadata []fastlane.LocaleMetadata, skipImages, replaceImages, syncImages bool) (updatedListings int, imageOutput, skippedImageOutput map[string]map[string]int, apiErr *errors.APIError) {
	imageOutput = map[string]map[string]int{}
	skippedImageOutput = map[string]map[string]int{}
	for i := range metadata {
		meta := &metadata[i]
		locale := config.NormalizeLocale(meta.Locale)
		listing := &androidpublisher.Listing{
			Language: locale,
		}
		if meta.TitleSet {
			listing.Title = meta.Title
		}
		if meta.ShortDescriptionSet {
			listing.ShortDescription = meta.ShortDescription
		}
		if meta.FullDescriptionSet {
			listing.FullDescription = meta.FullDescription
		}
		if meta.VideoSet {
			listing.Video = meta.Video
		}

		_, err := publisher.Edits.Listings.Update(c.packageName, editID, locale, listing).Context(ctx).Do()
		if err != nil {
			return updatedListings, imageOutput, skippedImageOutput, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update listing for %s: %v", locale, err))
		}
		updatedListings++

		if skipImages || len(meta.Images) == 0 {
			continue
		}
		counts, skipped, apiErr := c.importFastlaneImages(ctx, publisher, editID, locale, meta.Images, replaceImages, syncImages)
		if apiErr != nil {
			return updatedListings, imageOutput, skippedImageOutput, apiErr
		}
		if len(counts) > 0 {
			imageOutput[meta.Locale] = counts
		}
		if len(skipped) > 0 {
			skippedImageOutput[meta.Locale] = skipped
		}
	}

	return updatedListings, imageOutput, skippedImageOutput, nil
}

func (c *CLI) importFastlaneImages(ctx context.Context, publisher *androidpublisher.Service, editID, locale string, images map[string][]string, replaceImages, syncImages bool) (counts, skipped map[string]int, apiErr *errors.APIError) {
	counts = map[string]int{}
	skipped = map[string]int{}
	hashCache := map[string]string{}

	for imageType, paths := range images {
		if len(paths) == 0 {
			continue
		}
		if !fastlaneImageTypeSet[imageType] {
			return nil, nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("unsupported image type: %s", imageType))
		}
		if replaceImages {
			if _, err := publisher.Edits.Images.Deleteall(c.packageName, editID, locale, imageType).Context(ctx).Do(); err != nil {
				return nil, nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to delete images for %s/%s: %v", locale, imageType, err))
			}
		}
		existingHashes := map[string]bool{}
		if syncImages && !replaceImages {
			list, err := publisher.Edits.Images.List(c.packageName, editID, locale, imageType).Context(ctx).Do()
			if err != nil {
				return nil, nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list images for %s/%s: %v", locale, imageType, err))
			}
			for _, image := range list.Images {
				if image == nil || image.Sha256 == "" {
					continue
				}
				existingHashes[strings.ToLower(image.Sha256)] = true
			}
		}
		for _, filePath := range paths {
			if syncImages && !replaceImages && len(existingHashes) > 0 {
				hash, ok := hashCache[filePath]
				if !ok {
					var err error
					hash, err = edits.HashFile(filePath)
					if err != nil {
						return nil, nil, errors.NewAPIError(errors.CodeGeneralError, err.Error())
					}
					hashCache[filePath] = hash
				}
				if existingHashes[strings.ToLower(hash)] {
					skipped[imageType]++
					continue
				}
			}
			if _, _, _, apiErr := validateImageFile(filePath, imageType); apiErr != nil {
				return nil, nil, apiErr
			}
			f, err := os.Open(filePath)
			if err != nil {
				return nil, nil, errors.NewAPIError(errors.CodeGeneralError, err.Error())
			}
			_, uploadErr := publisher.Edits.Images.Upload(c.packageName, editID, locale, imageType).Media(f).Context(ctx).Do()
			closeErr := f.Close()
			if uploadErr != nil {
				if closeErr != nil {
					return nil, nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("%v; close error: %v", uploadErr, closeErr))
				}
				return nil, nil, errors.NewAPIError(errors.CodeGeneralError, uploadErr.Error())
			}
			if closeErr != nil {
				return nil, nil, errors.NewAPIError(errors.CodeGeneralError, closeErr.Error())
			}
			counts[imageType]++
		}
	}

	if len(skipped) == 0 {
		skipped = nil
	}
	return counts, skipped, nil
}

func (c *CLI) exportFastlaneChangelogs(ctx context.Context, publisher *androidpublisher.Service, editID string, locales map[string]bool) (map[string]map[string]string, *errors.APIError) {
	tracks, err := publisher.Edits.Tracks.List(c.packageName, editID).Context(ctx).Do()
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
	}
	if tracks == nil || len(tracks.Tracks) == 0 {
		return nil, nil
	}
	sortTracksByPriority(tracks.Tracks)

	notesByLocale := map[string]map[string]string{}
	for _, track := range tracks.Tracks {
		if track == nil {
			continue
		}
		for _, release := range track.Releases {
			if release == nil || len(release.ReleaseNotes) == 0 || len(release.VersionCodes) == 0 {
				continue
			}
			for _, note := range release.ReleaseNotes {
				if note == nil || note.Text == "" {
					continue
				}
				locale := config.NormalizeLocale(note.Language)
				if locales != nil && !locales[locale] {
					continue
				}
				for _, code := range release.VersionCodes {
					key := strconv.FormatInt(code, 10)
					if notesByLocale[locale] == nil {
						notesByLocale[locale] = map[string]string{}
					}
					if _, exists := notesByLocale[locale][key]; exists {
						continue
					}
					notesByLocale[locale][key] = note.Text
				}
			}
		}
	}
	return notesByLocale, nil
}

const defaultChangelogKey = "default"

func buildChangelogSets(metadata []fastlane.LocaleMetadata) map[string]changelogSet {
	sets := map[string]changelogSet{}
	for i := range metadata {
		meta := &metadata[i]
		if len(meta.Changelogs) == 0 {
			continue
		}
		set := changelogSet{byVersion: map[int64]string{}}
		for key, text := range meta.Changelogs {
			if key == defaultChangelogKey {
				set.defaultText = text
				set.hasDefault = true
				continue
			}
			versionCode, err := strconv.ParseInt(key, 10, 64)
			if err != nil {
				continue
			}
			set.byVersion[versionCode] = text
		}
		if len(set.byVersion) == 0 && !set.hasDefault {
			continue
		}
		locale := config.NormalizeLocale(meta.Locale)
		sets[locale] = set
	}
	return sets
}

func (c *CLI) importFastlaneReleaseNotes(ctx context.Context, publisher *androidpublisher.Service, editID string, changelogSets map[string]changelogSet) (updatedTracks, updatedReleases int, apiErr *errors.APIError) {
	tracks, err := publisher.Edits.Tracks.List(c.packageName, editID).Context(ctx).Do()
	if err != nil {
		return 0, 0, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to list tracks: %v", err))
	}
	if tracks == nil || len(tracks.Tracks) == 0 {
		return 0, 0, nil
	}
	for _, track := range tracks.Tracks {
		if track == nil || len(track.Releases) == 0 {
			continue
		}
		trackChanged := false
		for _, release := range track.Releases {
			if release == nil {
				continue
			}
			if applyReleaseNotes(release, changelogSets) {
				trackChanged = true
				updatedReleases++
			}
		}
		if !trackChanged {
			continue
		}
		_, err := publisher.Edits.Tracks.Update(c.packageName, editID, track.Track, track).Context(ctx).Do()
		if err != nil {
			return updatedTracks, updatedReleases, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("failed to update track %s: %v", track.Track, err))
		}
		updatedTracks++
	}
	return updatedTracks, updatedReleases, nil
}

func applyReleaseNotes(release *androidpublisher.TrackRelease, changelogSets map[string]changelogSet) bool {
	if release == nil || len(changelogSets) == 0 {
		return false
	}
	existing := map[string]string{}
	for _, note := range release.ReleaseNotes {
		if note == nil {
			continue
		}
		locale := config.NormalizeLocale(note.Language)
		existing[locale] = note.Text
	}
	changed := false
	for locale, set := range changelogSets {
		text, ok := selectChangelogText(set, release.VersionCodes)
		if !ok {
			continue
		}
		if current, ok := existing[locale]; ok && current == text {
			continue
		}
		existing[locale] = text
		changed = true
	}
	if !changed {
		return false
	}
	locales := make([]string, 0, len(existing))
	for locale := range existing {
		locales = append(locales, locale)
	}
	sort.Strings(locales)
	release.ReleaseNotes = make([]*androidpublisher.LocalizedText, 0, len(locales))
	for _, locale := range locales {
		release.ReleaseNotes = append(release.ReleaseNotes, &androidpublisher.LocalizedText{
			Language: locale,
			Text:     existing[locale],
		})
	}
	return true
}

func selectChangelogText(set changelogSet, versionCodes []int64) (string, bool) {
	var selected string
	var selectedVersion int64
	found := false
	for _, code := range versionCodes {
		if text, ok := set.byVersion[code]; ok {
			if !found || code > selectedVersion {
				selected = text
				selectedVersion = code
				found = true
			}
		}
	}
	if found {
		return selected, true
	}
	if set.hasDefault {
		return set.defaultText, true
	}
	return "", false
}

func sortTracksByPriority(tracks []*androidpublisher.Track) {
	sort.SliceStable(tracks, func(i, j int) bool {
		if tracks[i] == nil {
			return false
		}
		if tracks[j] == nil {
			return true
		}
		pi := trackPriority(tracks[i].Track)
		pj := trackPriority(tracks[j].Track)
		if pi != pj {
			return pi < pj
		}
		return tracks[i].Track < tracks[j].Track
	})
}

func trackPriority(track string) int {
	switch track {
	case "production":
		return 0
	case "beta":
		return 1
	case "alpha":
		return 2
	case "internal":
		return 3
	default:
		return 10
	}
}

func normalizeLocaleFilter(locales []string) map[string]bool {
	if len(locales) == 0 {
		return nil
	}
	filter := map[string]bool{}
	for _, locale := range locales {
		normalized := strings.TrimSpace(locale)
		if normalized == "" {
			continue
		}
		filter[config.NormalizeLocale(normalized)] = true
	}
	if len(filter) == 0 {
		return nil
	}
	return filter
}

func downloadImage(ctx context.Context, httpClient *http.Client, rawURL, destBase string) (err error) {
	if err := validateURL(rawURL); err != nil {
		return fmt.Errorf("invalid image URL: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req) // #nosec G704 -- URL validated above
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	buf := make([]byte, 512)
	n, err := io.ReadFull(resp.Body, buf)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return err
	}

	contentType := http.DetectContentType(buf[:n])
	ext := extensionForContentType(contentType)
	if ext == "" {
		ext = extensionFromURL(rawURL)
	}
	if ext == "" {
		ext = ".img"
	}

	destPath := destBase + ext
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}

	tempPath := destPath + ".tmp"
	file, err := os.Create(tempPath)
	if err != nil {
		return err
	}

	cleanupTemp := func(copyErr error) error {
		closeErr := file.Close()
		removeErr := os.Remove(tempPath)
		if closeErr != nil && removeErr != nil {
			return fmt.Errorf("%v; close error: %v; remove error: %v", copyErr, closeErr, removeErr)
		}
		if closeErr != nil {
			return fmt.Errorf("%v; close error: %v", copyErr, closeErr)
		}
		if removeErr != nil {
			return fmt.Errorf("%v; remove error: %v", copyErr, removeErr)
		}
		return copyErr
	}

	if n > 0 {
		if _, err := io.Copy(file, bytes.NewReader(buf[:n])); err != nil {
			return cleanupTemp(err)
		}
	}
	if _, err := io.Copy(file, resp.Body); err != nil {
		return cleanupTemp(err)
	}
	if err := file.Close(); err != nil {
		removeErr := os.Remove(tempPath)
		if removeErr != nil {
			return fmt.Errorf("%v; remove error: %v", err, removeErr)
		}
		return err
	}
	if err := os.Rename(tempPath, destPath); err != nil {
		removeErr := os.Remove(tempPath)
		if removeErr != nil {
			return fmt.Errorf("%v; remove error: %v", err, removeErr)
		}
		return err
	}
	return nil
}

func extensionForContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/jpeg":
		return ".jpg"
	default:
		return ""
	}
}

func extensionFromURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	ext := path.Ext(parsed.Path)
	if ext == "" {
		return ""
	}
	switch strings.ToLower(ext) {
	case ".png", ".jpg", ".jpeg":
		return ext
	default:
		return ""
	}
}
