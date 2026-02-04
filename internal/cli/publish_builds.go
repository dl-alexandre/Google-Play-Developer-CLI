package cli

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/logging"
	"github.com/dl-alexandre/gpd/internal/output"
)

type buildType string

const (
	buildTypeAll    buildType = "all"
	buildTypeAPK    buildType = "apk"
	buildTypeBundle buildType = "bundle"
)

func (c *CLI) addPublishBuildsCommands(publishCmd *cobra.Command) {
	buildsCmd := &cobra.Command{
		Use:   "builds",
		Short: "Manage uploaded builds",
		Long:  "List and inspect APKs and App Bundles uploaded in an edit.",
	}

	var (
		buildKind string
		editID    string
	)

	buildsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List uploaded builds",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishBuildsList(cmd.Context(), buildKind, editID)
		},
	}
	buildsListCmd.Flags().StringVar(&buildKind, "type", "all", "Build type (apk, bundle, all)")
	buildsListCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")

	buildsGetCmd := &cobra.Command{
		Use:     "get <version-code>",
		Aliases: []string{"info"},
		Short:   "Get build details",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			versionCode, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || versionCode <= 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"version code must be a positive integer"))
			}
			return c.publishBuildsGet(cmd.Context(), versionCode, buildKind, editID)
		},
	}
	buildsGetCmd.Flags().StringVar(&buildKind, "type", "all", "Build type (apk, bundle, all)")
	buildsGetCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")

	var (
		expireConfirm  bool
		expireDryRun   bool
		expireNoCommit bool
	)

	buildsExpireCmd := &cobra.Command{
		Use:   "expire <version-code>",
		Short: "Expire a build from tracks",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			versionCode, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || versionCode <= 0 {
				return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
					"version code must be a positive integer"))
			}
			return c.publishBuildsExpire(cmd.Context(), versionCode, editID, expireNoCommit, expireDryRun, expireConfirm)
		},
	}
	buildsExpireCmd.Flags().BoolVar(&expireConfirm, "confirm", false, "Confirm destructive operation")
	buildsExpireCmd.Flags().BoolVar(&expireDryRun, "dry-run", false, "Show intended actions without executing")
	buildsExpireCmd.Flags().BoolVar(&expireNoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	buildsExpireCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")

	buildsExpireAllCmd := &cobra.Command{
		Use:   "expire-all",
		Short: "Expire all builds from tracks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishBuildsExpireAll(cmd.Context(), editID, expireNoCommit, expireDryRun, expireConfirm)
		},
	}
	buildsExpireAllCmd.Flags().BoolVar(&expireConfirm, "confirm", false, "Confirm destructive operation")
	buildsExpireAllCmd.Flags().BoolVar(&expireDryRun, "dry-run", false, "Show intended actions without executing")
	buildsExpireAllCmd.Flags().BoolVar(&expireNoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	buildsExpireAllCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")

	buildsCmd.AddCommand(buildsListCmd, buildsGetCmd, buildsExpireCmd, buildsExpireAllCmd)
	publishCmd.AddCommand(buildsCmd)
}

func (c *CLI) publishBuildsList(ctx context.Context, buildKind, editID string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	kind, apiErr := normalizeBuildType(buildKind)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	activeEditID := editID
	if activeEditID == "" {
		edit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to create edit: %v", err)))
		}
		activeEditID = edit.Id
		defer func() {
			if err := publisher.Edits.Delete(c.packageName, activeEditID).Context(ctx).Do(); err != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", activeEditID), logging.Err(err))
			}
		}()
	}

	result := map[string]interface{}{
		"package": c.packageName,
		"editId":  activeEditID,
	}

	if kind == buildTypeAll || kind == buildTypeBundle {
		bundles, err := publisher.Edits.Bundles.List(c.packageName, activeEditID).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		result["bundles"] = bundles.Bundles
	}
	if kind == buildTypeAll || kind == buildTypeAPK {
		apks, err := publisher.Edits.Apks.List(c.packageName, activeEditID).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		result["apks"] = apks.Apks
	}

	return c.Output(output.NewResult(result).WithServices("androidpublisher"))
}

func (c *CLI) publishBuildsGet(ctx context.Context, versionCode int64, buildKind, editID string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	kind, apiErr := normalizeBuildType(buildKind)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	activeEditID := editID
	if activeEditID == "" {
		edit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to create edit: %v", err)))
		}
		activeEditID = edit.Id
		defer func() {
			if err := publisher.Edits.Delete(c.packageName, activeEditID).Context(ctx).Do(); err != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", activeEditID), logging.Err(err))
			}
		}()
	}

	result := map[string]interface{}{
		"package":     c.packageName,
		"editId":      activeEditID,
		"versionCode": versionCode,
	}

	found := false
	if kind == buildTypeAll || kind == buildTypeBundle {
		bundles, err := publisher.Edits.Bundles.List(c.packageName, activeEditID).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		for _, bundle := range bundles.Bundles {
			if bundle.VersionCode == versionCode {
				result["bundle"] = bundle
				found = true
				break
			}
		}
	}
	if kind == buildTypeAll || kind == buildTypeAPK {
		apks, err := publisher.Edits.Apks.List(c.packageName, activeEditID).Context(ctx).Do()
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		for _, apk := range apks.Apks {
			if apk.VersionCode == versionCode {
				result["apk"] = apk
				found = true
				break
			}
		}
	}

	if !found {
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("build not found for version code %d", versionCode)))
	}

	return c.Output(output.NewResult(result).WithServices("androidpublisher"))
}

func (c *CLI) publishBuildsExpire(ctx context.Context, versionCode int64, editID string, noAutoCommit, dryRun, confirm bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if !confirm {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"--confirm flag required for destructive operations"))
	}
	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":      true,
			"action":      "builds_expire",
			"versionCode": versionCode,
			"package":     c.packageName,
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

	tracksResp, err := publisher.Edits.Tracks.List(c.packageName, edit.ServerID).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	updatedTracks := make([]string, 0)
	removedCount := 0
	for _, track := range tracksResp.Tracks {
		changed := false
		updatedReleases := make([]*androidpublisher.TrackRelease, 0, len(track.Releases))
		for _, release := range track.Releases {
			remaining := make([]int64, 0, len(release.VersionCodes))
			for _, vc := range release.VersionCodes {
				if vc == versionCode {
					removedCount++
					changed = true
					continue
				}
				remaining = append(remaining, vc)
			}
			if len(remaining) == 0 {
				if len(release.VersionCodes) > 0 {
					changed = true
				}
				continue
			}
			release.VersionCodes = remaining
			updatedReleases = append(updatedReleases, release)
		}
		if !changed {
			continue
		}
		track.Releases = updatedReleases
		if _, err := publisher.Edits.Tracks.Update(c.packageName, edit.ServerID, track.Track, track).Context(ctx).Do(); err != nil {
			if created {
				_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to update track %s: %v", track.Track, err)))
		}
		updatedTracks = append(updatedTracks, track.Track)
	}

	if removedCount == 0 {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("version code %d not found on any track", versionCode)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":      true,
		"versionCode":  versionCode,
		"tracks":       updatedTracks,
		"removedCount": removedCount,
		"package":      c.packageName,
		"editId":       edit.ServerID,
		"committed":    !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishBuildsExpireAll(ctx context.Context, editID string, noAutoCommit, dryRun, confirm bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	if !confirm {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"--confirm flag required for destructive operations"))
	}
	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":  true,
			"action":  "builds_expire_all",
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

	tracksResp, err := publisher.Edits.Tracks.List(c.packageName, edit.ServerID).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	updatedTracks := make([]string, 0)
	for _, track := range tracksResp.Tracks {
		if len(track.Releases) == 0 {
			continue
		}
		track.Releases = nil
		if _, err := publisher.Edits.Tracks.Update(c.packageName, edit.ServerID, track.Track, track).Context(ctx).Do(); err != nil {
			if created {
				_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to update track %s: %v", track.Track, err)))
		}
		updatedTracks = append(updatedTracks, track.Track)
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"tracks":    updatedTracks,
		"package":   c.packageName,
		"editId":    edit.ServerID,
		"committed": !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func normalizeBuildType(value string) (buildType, *errors.APIError) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "", "all":
		return buildTypeAll, nil
	case "apk", "apks":
		return buildTypeAPK, nil
	case "bundle", "bundles", "aab", "aabs":
		return buildTypeBundle, nil
	default:
		return "", errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid build type: %s", value)).
			WithHint("Use --type apk, bundle, or all")
	}
}
