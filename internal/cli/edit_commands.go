package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) addEditCommands(publishCmd *cobra.Command) {
	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Manage edit transactions",
		Long:  "Create, inspect, validate, commit, and delete edit transactions.",
	}

	editCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new edit",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishEditCreate(cmd.Context())
		},
	}

	editListCmd := &cobra.Command{
		Use:   "list",
		Short: "List cached edits",
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishEditList(cmd.Context())
		},
	}

	editGetCmd := &cobra.Command{
		Use:   "get <edit-id>",
		Short: "Get edit details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishEditGet(cmd.Context(), args[0])
		},
	}

	editCommitCmd := &cobra.Command{
		Use:   "commit <edit-id>",
		Short: "Commit an edit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishEditCommit(cmd.Context(), args[0])
		},
	}

	editValidateCmd := &cobra.Command{
		Use:   "validate <edit-id>",
		Short: "Validate an edit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishEditValidate(cmd.Context(), args[0])
		},
	}

	editDeleteCmd := &cobra.Command{
		Use:   "delete <edit-id>",
		Short: "Delete an edit",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return c.publishEditDelete(cmd.Context(), args[0])
		},
	}

	editCmd.AddCommand(editCreateCmd, editListCmd, editGetCmd, editCommitCmd, editValidateCmd, editDeleteCmd)
	publishCmd.AddCommand(editCmd)
}

func (c *CLI) publishEditCreate(ctx context.Context) error {
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

	editMgr := edits.NewManager()
	if err := editMgr.AcquireLock(ctx, c.packageName); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	apiEdit, err := publisher.Edits.Insert(c.packageName, nil).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to create edit: %v", err)))
	}

	edit := &edits.Edit{
		Handle:      apiEdit.Id,
		ServerID:    apiEdit.Id,
		PackageName: c.packageName,
		CreatedAt:   time.Now(),
		LastUsedAt:  time.Now(),
		State:       edits.StateDraft,
	}
	if err := editMgr.SaveEdit(edit); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"editId":     edit.ServerID,
		"package":    edit.PackageName,
		"createdAt":  edit.CreatedAt,
		"lastUsedAt": edit.LastUsedAt,
		"state":      edit.State,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishEditList(_ context.Context) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	editMgr := edits.NewManager()
	editsList, err := editMgr.ListEdits(c.packageName)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	now := time.Now()
	var results []map[string]interface{}
	for _, edit := range editsList {
		results = append(results, map[string]interface{}{
			"editId":     edit.ServerID,
			"handle":     edit.Handle,
			"package":    edit.PackageName,
			"createdAt":  edit.CreatedAt,
			"lastUsedAt": edit.LastUsedAt,
			"state":      edit.State,
			"expired":    editMgr.IsEditExpired(edit, now),
		})
	}

	result := output.NewResult(map[string]interface{}{
		"edits":   results,
		"count":   len(results),
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishEditGet(ctx context.Context, editID string) error {
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

	editMgr := edits.NewManager()
	serverID, local, err := c.resolveEdit(editMgr, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	remote, err := publisher.Edits.Get(c.packageName, serverID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to get edit: %v", err)))
	}

	result := output.NewResult(map[string]interface{}{
		"editId":  serverID,
		"local":   local,
		"remote":  remote,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishEditCommit(ctx context.Context, editID string) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	editMgr := edits.NewManager()
	serverID, local, err := c.resolveEdit(editMgr, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	contentID := serverID
	if local != nil {
		contentID = fmt.Sprintf("%s:%d", serverID, local.CreatedAt.Unix())
	}
	idempotencyResult, idempotencyKey, _ := editMgr.Idempotent.CheckCommit(c.packageName, serverID, contentID)
	if idempotencyResult != nil && idempotencyResult.Found {
		result := output.NewResult(map[string]interface{}{
			"idempotent": true,
			"editId":     serverID,
			"package":    c.packageName,
			"recordedAt": idempotencyResult.Timestamp,
		})
		return c.Output(result.WithNoOp("commit already completed").WithServices("androidpublisher"))
	}

	client, err := c.getAPIClient(ctx)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	publisher, err := client.AndroidPublisher()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	_, err = publisher.Edits.Commit(c.packageName, serverID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to commit edit: %v", err)))
	}

	_ = editMgr.Idempotent.RecordCommit(idempotencyKey, c.packageName, serverID)

	if local != nil {
		_, _ = editMgr.UpdateEditState(c.packageName, local.Handle, edits.StateCommitted)
		_ = editMgr.DeleteEdit(c.packageName, local.Handle)
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"editId":  serverID,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishEditValidate(ctx context.Context, editID string) error {
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

	editMgr := edits.NewManager()
	serverID, local, err := c.resolveEdit(editMgr, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if local != nil {
		_, _ = editMgr.UpdateEditState(c.packageName, local.Handle, edits.StateValidating)
	}

	_, err = publisher.Edits.Validate(c.packageName, serverID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to validate edit: %v", err)))
	}

	if local != nil {
		_, _ = editMgr.UpdateEditState(c.packageName, local.Handle, edits.StateDraft)
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"editId":  serverID,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishEditDelete(ctx context.Context, editID string) error {
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

	editMgr := edits.NewManager()
	serverID, local, err := c.resolveEdit(editMgr, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	err = publisher.Edits.Delete(c.packageName, serverID).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to delete edit: %v", err)))
	}

	if local != nil {
		_, _ = editMgr.UpdateEditState(c.packageName, local.Handle, edits.StateAborted)
		_ = editMgr.DeleteEdit(c.packageName, local.Handle)
	}

	result := output.NewResult(map[string]interface{}{
		"success": true,
		"editId":  serverID,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) resolveEdit(editMgr *edits.Manager, editID string) (string, *edits.Edit, error) {
	local, err := editMgr.LoadEdit(c.packageName, editID)
	if err != nil {
		return "", nil, err
	}
	if local != nil {
		if editMgr.IsEditExpired(local, time.Now()) {
			return "", nil, errors.NewAPIError(errors.CodeConflict, "edit has expired")
		}
		if local.ServerID != "" {
			return local.ServerID, local, nil
		}
	}
	return editID, local, nil
}
