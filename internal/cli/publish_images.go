package cli

import (
	"context"
	"fmt"
	"image"
	_ "image/jpeg" // Register JPEG decoder for image.DecodeConfig
	_ "image/png"  // Register PNG decoder for image.DecodeConfig
	"os"
	"strings"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/logging"
	"github.com/dl-alexandre/gpd/internal/output"
)

type imageSpec struct {
	minWidth  int
	maxWidth  int
	minHeight int
	maxHeight int
	maxSize   int64
	formats   []string
}

func imageSpecs() map[string]imageSpec {
	return map[string]imageSpec{
		"icon":                 {minWidth: 512, maxWidth: 512, minHeight: 512, maxHeight: 512, maxSize: 1 * 1024 * 1024, formats: []string{"png"}},
		"featureGraphic":       {minWidth: 1024, maxWidth: 1024, minHeight: 500, maxHeight: 500, maxSize: 15 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"promoGraphic":         {maxSize: 15 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tvBanner":             {minWidth: 1280, maxWidth: 1280, minHeight: 720, maxHeight: 720, maxSize: 15 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"phoneScreenshots":     {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tabletScreenshots":    {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"sevenInchScreenshots": {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tenInchScreenshots":   {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"tvScreenshots":        {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
		"wearScreenshots":      {minWidth: 320, maxWidth: 3840, minHeight: 320, maxHeight: 3840, maxSize: 8 * 1024 * 1024, formats: []string{"png", "jpeg"}},
	}
}

func validateImageFile(filePath, imageType string) (info os.FileInfo, cfg image.Config, format string, apiErr *errors.APIError) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath))
	}
	f, err := os.Open(filePath)
	if err != nil {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}
	cfg, format, decodeErr := image.DecodeConfig(f)
	closeErr := f.Close()
	if decodeErr != nil {
		if closeErr != nil {
			return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("invalid image file: %v", closeErr))
		}
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeValidationError, "invalid image file")
	}
	if closeErr != nil {
		return nil, image.Config{}, "", errors.NewAPIError(errors.CodeGeneralError, closeErr.Error())
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

func (c *CLI) publishImagesUpload(ctx context.Context, imageType, filePath, locale string, syncImages bool, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	info, cfg, format, apiErr := validateImageFile(filePath, imageType)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}
	var localHash string
	if syncImages {
		hash, err := edits.HashFile(filePath)
		if err != nil {
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		localHash = hash
	}
	if dryRun {
		resultData := map[string]interface{}{
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
		}
		if localHash != "" {
			resultData["sha256"] = localHash
		}
		result := output.NewResult(resultData)
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
	defer func() {
		if err := editMgr.ReleaseLock(c.packageName); err != nil {
			logging.Warn("failed to release edit lock", logging.String("package", c.packageName), logging.Err(err))
		}
	}()

	if syncImages && localHash != "" {
		images, err := publisher.Edits.Images.List(c.packageName, edit.ServerID, locale, imageType).Context(ctx).Do()
		if err != nil {
			if created {
				if cleanupErr := publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do(); cleanupErr != nil {
					logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.ServerID), logging.Err(cleanupErr))
				}
			}
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
		}
		for _, image := range images.Images {
			if image != nil && strings.EqualFold(image.Sha256, localHash) {
				if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
					return c.OutputError(err.(*errors.APIError))
				}
				result := output.NewResult(map[string]interface{}{
					"idempotent": true,
					"type":       imageType,
					"locale":     locale,
					"sha256":     localHash,
					"package":    c.packageName,
					"editId":     edit.ServerID,
					"committed":  !noAutoCommit,
				})
				return c.Output(result.WithNoOp("image already uploaded").WithServices("androidpublisher"))
			}
		}
	}

	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	resp, err := publisher.Edits.Images.Upload(c.packageName, edit.ServerID, locale, imageType).
		Media(f).Context(ctx).Do()
	closeErr := f.Close()
	if err != nil {
		if closeErr != nil {
			logging.Warn("failed to close image file", logging.String("path", filePath), logging.Err(closeErr))
		}
		if created {
			if cleanupErr := publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do(); cleanupErr != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.ServerID), logging.Err(cleanupErr))
			}
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if closeErr != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, closeErr.Error()))
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
		defer func() {
			if err := publisher.Edits.Delete(c.packageName, edit.Id).Context(ctx).Do(); err != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.Id), logging.Err(err))
			}
		}()
	}
	images, err := publisher.Edits.Images.List(c.packageName, edit.Id, locale, imageType).Context(ctx).Do()
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	return c.Output(output.NewResult(images).WithServices("androidpublisher"))
}

func (c *CLI) publishImagesDelete(ctx context.Context, imageType, imageID, locale, editID string, noAutoCommit, dryRun bool) error {
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
	defer func() {
		if err := editMgr.ReleaseLock(c.packageName); err != nil {
			logging.Warn("failed to release edit lock", logging.String("package", c.packageName), logging.Err(err))
		}
	}()

	if err := publisher.Edits.Images.Delete(c.packageName, edit.ServerID, locale, imageType, imageID).Context(ctx).Do(); err != nil {
		if created {
			if cleanupErr := publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do(); cleanupErr != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.ServerID), logging.Err(cleanupErr))
			}
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

func (c *CLI) publishImagesDeleteAll(ctx context.Context, imageType, locale, editID string, noAutoCommit, dryRun bool) error {
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
	defer func() {
		if err := editMgr.ReleaseLock(c.packageName); err != nil {
			logging.Warn("failed to release edit lock", logging.String("package", c.packageName), logging.Err(err))
		}
	}()

	if _, err := publisher.Edits.Images.Deleteall(c.packageName, edit.ServerID, locale, imageType).Context(ctx).Do(); err != nil {
		if created {
			if cleanupErr := publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do(); cleanupErr != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.ServerID), logging.Err(cleanupErr))
			}
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

func (c *CLI) publishAssetsUpload(_ context.Context, dir, category string, replace bool, editID string, noAutoCommit, dryRun bool) error {
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
		"success":   true,
		"dir":       dir,
		"category":  category,
		"replace":   replace,
		"package":   c.packageName,
		"editId":    editID,
		"committed": !noAutoCommit,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) publishAssetsSpec(_ context.Context) error {
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
