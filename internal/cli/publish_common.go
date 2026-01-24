package cli

import (
	"context"
	"fmt"
	"net/mail"
	"net/url"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
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

	contentID := fmt.Sprintf("%s:%d", edit.ServerID, edit.CreatedAt.Unix())
	idempotencyResult, idempotencyKey, _ := editMgr.Idempotent.CheckCommit(c.packageName, edit.ServerID, contentID)
	if idempotencyResult != nil && idempotencyResult.Found {
		_, _ = editMgr.UpdateEditState(c.packageName, edit.Handle, edits.StateCommitted)
		_ = editMgr.DeleteEdit(c.packageName, edit.Handle)
		return nil
	}

	_, err := publisher.Edits.Commit(c.packageName, edit.ServerID).Context(ctx).Do()
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to commit edit: %v", err))
	}

	_ = editMgr.Idempotent.RecordCommit(idempotencyKey, c.packageName, edit.ServerID)

	_, _ = editMgr.UpdateEditState(c.packageName, edit.Handle, edits.StateCommitted)
	_ = editMgr.DeleteEdit(c.packageName, edit.Handle)
	return nil
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
