package cli

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) publishTestersAdd(ctx context.Context, track string, groups []string, editID string, noAutoCommit, dryRun bool) error {
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

	testers, err := publisher.Edits.Testers.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		testers = &androidpublisher.Testers{}
	}

	existingGroups := make(map[string]bool)
	for _, g := range testers.GoogleGroups {
		existingGroups[g] = true
	}
	for _, g := range groups {
		if !existingGroups[g] {
			testers.GoogleGroups = append(testers.GoogleGroups, g)
		}
	}

	_, err = publisher.Edits.Testers.Update(c.packageName, edit.ServerID, track, testers).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
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

func (c *CLI) publishTestersRemove(ctx context.Context, track string, groups []string, editID string, noAutoCommit, dryRun bool) error {
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

	testers, err := publisher.Edits.Testers.Get(c.packageName, edit.ServerID, track).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeNotFound,
			fmt.Sprintf("no testers found for track: %s", track)))
	}

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

	_, err = publisher.Edits.Testers.Update(c.packageName, edit.ServerID, track, testers).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
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
	defer func() { _ = publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do() }()

	if track != "" {
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

	tracks := []string{"internal", "alpha", "beta", "production"}
	testersData := make(map[string]interface{})
	var mu sync.Mutex

	g, gctx := errgroup.WithContext(ctx)
	for _, t := range tracks {
		t := t
		g.Go(func() error {
			if err := client.Acquire(gctx); err != nil {
				return err
			}
			defer client.Release()

			testers, err := publisher.Edits.Testers.Get(c.packageName, edit.Id, t).Context(gctx).Do()
			if err != nil {
				if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
					return nil
				}
				return err
			}
			if len(testers.GoogleGroups) > 0 {
				mu.Lock()
				testersData[t] = map[string]interface{}{
					"googleGroups": testers.GoogleGroups,
				}
				mu.Unlock()
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	result := output.NewResult(map[string]interface{}{
		"testers": testersData,
		"package": c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}
