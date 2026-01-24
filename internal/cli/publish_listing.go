package cli

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) publishListingUpdate(ctx context.Context, locale, title, shortDesc, fullDesc, editID string, noAutoCommit, dryRun bool) error {
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

	updatedListing, err := publisher.Edits.Listings.Update(c.packageName, edit.ServerID, locale, listing).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
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

	if locale != "" {
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
	defer func() { _ = publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do() }()

	details, err := publisher.Edits.Details.Get(c.packageName, edit.Id).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	return c.Output(output.NewResult(details).WithServices("androidpublisher"))
}

func (c *CLI) publishDetailsUpdate(ctx context.Context, email, phone, website, defaultLanguage, editID string, noAutoCommit, dryRun bool) error {
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
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	details := &androidpublisher.AppDetails{
		ContactEmail:    email,
		ContactPhone:    phone,
		ContactWebsite:  website,
		DefaultLanguage: defaultLanguage,
	}

	updated, err := publisher.Edits.Details.Update(c.packageName, edit.ServerID, details).Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"details":   updated,
		"package":   c.packageName,
		"editId":    edit.ServerID,
		"committed": !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

// detailsPatchParams holds parameters for patching app details.
type detailsPatchParams struct {
	email           string
	phone           string
	website         string
	defaultLanguage string
	updateMask      string
	editID          string
	noAutoCommit    bool
}

// validateDetailsPatchInput validates the input parameters for details patch.
func validateDetailsPatchInput(email, phone, website, defaultLanguage string) *errors.APIError {
	if email == "" && phone == "" && website == "" && defaultLanguage == "" {
		return errors.NewAPIError(errors.CodeValidationError, "at least one field is required")
	}
	if email != "" && !isValidEmail(email) {
		return errors.NewAPIError(errors.CodeValidationError, "invalid contact email")
	}
	if website != "" && !isValidURL(website) {
		return errors.NewAPIError(errors.CodeValidationError, "invalid contact website")
	}
	return nil
}

// buildUpdateMask constructs the update mask from non-empty fields.
func buildUpdateMask(email, phone, website, defaultLanguage, existingMask string) string {
	if existingMask != "" {
		return existingMask
	}
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
	return strings.Join(fields, ",")
}

func (c *CLI) publishDetailsPatch(ctx context.Context, email, phone, website, defaultLanguage, updateMask, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if err := validateDetailsPatchInput(email, phone, website, defaultLanguage); err != nil {
		return c.OutputError(err)
	}

	if dryRun {
		return c.outputDetailsPatchDryRun(email, phone, website, defaultLanguage, updateMask)
	}

	params := &detailsPatchParams{
		email:           email,
		phone:           phone,
		website:         website,
		defaultLanguage: defaultLanguage,
		updateMask:      buildUpdateMask(email, phone, website, defaultLanguage, updateMask),
		editID:          editID,
		noAutoCommit:    noAutoCommit,
	}

	return c.executeDetailsPatch(ctx, params)
}

// outputDetailsPatchDryRun outputs the dry run result for details patch.
func (c *CLI) outputDetailsPatchDryRun(email, phone, website, defaultLanguage, updateMask string) error {
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

// executeDetailsPatch performs the actual patch operation.
func (c *CLI) executeDetailsPatch(ctx context.Context, params *detailsPatchParams) error {
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

	details := &androidpublisher.AppDetails{
		ContactEmail:    params.email,
		ContactPhone:    params.phone,
		ContactWebsite:  params.website,
		DefaultLanguage: params.defaultLanguage,
	}

	call := publisher.Edits.Details.Patch(c.packageName, edit.ServerID, details)
	_ = params.updateMask // Note: API does not support field mask for Details.Patch
	updated, err := call.Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !params.noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":   true,
		"details":   updated,
		"package":   c.packageName,
		"editId":    edit.ServerID,
		"committed": !params.noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}
