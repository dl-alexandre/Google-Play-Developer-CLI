// Package cli provides beta group compatibility commands for gpd.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

var supportedBetaGroupTracks = map[string]bool{
	"internal": true,
	"alpha":    true,
	"beta":     true,
}

func (c *CLI) addPublishBetaGroupsCommands(publishCmd *cobra.Command) {
	betaGroupsCmd := &cobra.Command{
		Use:   "beta-groups",
		Short: "ASC-style beta group compatibility",
		Long:  "Compatibility layer mapping ASC beta group workflows to Google Play track testers.",
	}

	var (
		groupName    string
		groups       []string
		track        string
		editID       string
		noAutoCommit bool
		dryRun       bool
	)

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List beta groups",
		Long:  "List beta groups. In Google Play, beta groups map to tester settings on internal/alpha/beta tracks.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishTestersList(cmd.Context(), track)
		},
	}
	listCmd.Flags().StringVar(&track, "track", "", "Track to list (internal, alpha, beta). Empty lists all supported tracks")

	getCmd := &cobra.Command{
		Use:   "get <group>",
		Short: "Get beta group details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName = args[0]
			if err := validateBetaGroupTrack(groupName); err != nil {
				result := output.NewErrorResult(err).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.publishTestersList(cmd.Context(), groupName)
		},
	}

	addTestersCmd := &cobra.Command{
		Use:   "add-testers <group>",
		Short: "Add tester Google Groups to a beta group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName = args[0]
			if err := validateBetaGroupTrack(groupName); err != nil {
				result := output.NewErrorResult(err).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.publishTestersAdd(cmd.Context(), groupName, groups, editID, noAutoCommit, dryRun)
		},
	}
	addTestersCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses to add")
	addTestersCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	addTestersCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	addTestersCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	removeTestersCmd := &cobra.Command{
		Use:   "remove-testers <group>",
		Short: "Remove tester Google Groups from a beta group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName = args[0]
			if err := validateBetaGroupTrack(groupName); err != nil {
				result := output.NewErrorResult(err).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.publishTestersRemove(cmd.Context(), groupName, groups, editID, noAutoCommit, dryRun)
		},
	}
	removeTestersCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses to remove")
	removeTestersCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	removeTestersCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	removeTestersCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	createCmd := &cobra.Command{
		Use:   "create <group>",
		Short: "Create beta group (compatibility)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName = args[0]
			if err := validateBetaGroupTrack(groupName); err != nil {
				result := output.NewErrorResult(err).WithServices("androidpublisher")
				return c.Output(result)
			}
			if len(groups) == 0 {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"at least one --group is required for create").
					WithHint("Use --group to seed tester Google Groups for the selected track beta group"),
				).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.publishTestersAdd(cmd.Context(), groupName, groups, editID, noAutoCommit, dryRun)
		},
	}
	createCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses")
	createCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	createCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	createCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	updateCmd := &cobra.Command{
		Use:   "update <group>",
		Short: "Update beta group testers (compatibility)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName = args[0]
			if err := validateBetaGroupTrack(groupName); err != nil {
				result := output.NewErrorResult(err).WithServices("androidpublisher")
				return c.Output(result)
			}
			if len(groups) == 0 {
				result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
					"at least one --group is required for update").
					WithHint("Use add-testers/remove-testers for incremental changes"),
				).WithServices("androidpublisher")
				return c.Output(result)
			}
			return c.publishTestersAdd(cmd.Context(), groupName, groups, editID, noAutoCommit, dryRun)
		},
	}
	updateCmd.Flags().StringSliceVar(&groups, "group", nil, "Google Group email addresses to add")
	updateCmd.Flags().StringVar(&editID, "edit-id", "", "Explicit edit transaction ID")
	updateCmd.Flags().BoolVar(&noAutoCommit, "no-auto-commit", false, "Keep edit open for manual commit")
	updateCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show intended actions without executing")

	deleteCmd := &cobra.Command{
		Use:   "delete <group>",
		Short: "Delete beta group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			groupName = args[0]
			if err := validateBetaGroupTrack(groupName); err != nil {
				result := output.NewErrorResult(err).WithServices("androidpublisher")
				return c.Output(result)
			}
			result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
				"beta group delete is not a direct Google Play operation").
				WithHint("Use 'gpd publish beta-groups remove-testers <group> --group ...' to remove group memberships from the mapped track"),
			).WithServices("androidpublisher")
			return c.Output(result)
		},
	}

	betaGroupsCmd.AddCommand(listCmd, getCmd, createCmd, updateCmd, deleteCmd, addTestersCmd, removeTestersCmd)
	publishCmd.AddCommand(betaGroupsCmd)
}

func validateBetaGroupTrack(track string) *errors.APIError {
	if supportedBetaGroupTracks[track] {
		return nil
	}
	return errors.NewAPIError(errors.CodeValidationError,
		fmt.Sprintf("unsupported beta group: %s", track)).
		WithHint("Use one of: internal, alpha, beta. Google Play maps tester groups to tracks, not standalone beta group objects")
}
