package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

const (
	extAPK = ".apk"
	extAAB = ".aab"
)

type uploadContext struct {
	filePath       string
	info           os.FileInfo
	ext            string
	hash           string
	idempotencyKey string
}

func (c *CLI) validateUploadFile(filePath string) (*uploadContext, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath))
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != extAAB && ext != extAPK {
		return nil, errors.NewAPIError(errors.CodeValidationError,
			"file must be an AAB or APK").WithHint("Supported formats: .aab, .apk")
	}

	hash, err := edits.HashFile(filePath)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to compute file hash: %v", err))
	}

	return &uploadContext{
		filePath: filePath,
		info:     info,
		ext:      ext,
		hash:     hash,
	}, nil
}

func (c *CLI) checkIdempotentUpload(uc *uploadContext, editMgr *edits.Manager) *output.Result {
	idempotencyResult, idempotencyKey, _ := editMgr.Idempotent.CheckUploadByHash(c.packageName, uc.hash)
	uc.idempotencyKey = idempotencyKey

	if idempotencyResult == nil || !idempotencyResult.Found {
		return nil
	}

	data, ok := idempotencyResult.Data.(map[string]interface{})
	if !ok {
		return nil
	}

	return output.NewResult(map[string]interface{}{
		"idempotent":  true,
		"versionCode": data["versionCode"],
		"sha256":      data["sha256"],
		"path":        uc.filePath,
		"size":        uc.info.Size(),
		"sizeHuman":   edits.FormatBytes(uc.info.Size()),
		"type":        uc.ext[1:],
		"package":     c.packageName,
		"editId":      data["editId"],
		"recordedAt":  idempotencyResult.Timestamp,
	}).WithNoOp("upload already completed").WithServices("androidpublisher")
}

func (c *CLI) checkCachedUpload(uc *uploadContext, editMgr *edits.Manager) *output.Result {
	cached, err := editMgr.GetCachedArtifactByHash(c.packageName, uc.hash)
	if err != nil || cached == nil {
		return nil
	}

	return output.NewResult(map[string]interface{}{
		"cached":    true,
		"sha256":    cached.SHA256,
		"path":      uc.filePath,
		"size":      uc.info.Size(),
		"sizeHuman": edits.FormatBytes(uc.info.Size()),
	}).WithNoOp("artifact already uploaded").WithServices("androidpublisher")
}

func (c *CLI) uploadArtifact(ctx context.Context, publisher *androidpublisher.Service, editID string, uc *uploadContext) (int64, error) {
	f, err := os.Open(uc.filePath)
	if err != nil {
		return 0, errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}
	defer f.Close()

	if uc.ext == extAAB {
		bundle, err := publisher.Edits.Bundles.Upload(c.packageName, editID).Media(f).Context(ctx).Do()
		if err != nil {
			return 0, fmt.Errorf("failed to upload bundle: %w", err)
		}
		return bundle.VersionCode, nil
	}

	apk, err := publisher.Edits.Apks.Upload(c.packageName, editID).Media(f).Context(ctx).Do()
	if err != nil {
		return 0, fmt.Errorf("failed to upload APK: %w", err)
	}
	return apk.VersionCode, nil
}

func (c *CLI) publishUpload(ctx context.Context, filePath, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	uc, err := c.validateUploadFile(filePath)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	editMgr := edits.NewManager()

	if result := c.checkIdempotentUpload(uc, editMgr); result != nil {
		return c.Output(result)
	}

	if result := c.checkCachedUpload(uc, editMgr); result != nil {
		return c.Output(result)
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":    true,
			"action":    "upload",
			"path":      filePath,
			"sha256":    uc.hash,
			"size":      uc.info.Size(),
			"sizeHuman": edits.FormatBytes(uc.info.Size()),
			"type":      uc.ext[1:],
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

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	versionCode, err := c.uploadArtifact(ctx, publisher, edit.ServerID, uc)
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	_ = editMgr.CacheArtifactWithHash(c.packageName, filePath, uc.hash, versionCode)

	uploadResult := &edits.UploadResult{
		VersionCode: versionCode,
		SHA256:      uc.hash,
		Path:        filePath,
		Size:        uc.info.Size(),
		Type:        uc.ext[1:],
		EditID:      edit.ServerID,
	}
	_ = editMgr.Idempotent.RecordUpload(uc.idempotencyKey, c.packageName, uc.hash, uploadResult)

	result := output.NewResult(map[string]interface{}{
		"success":     true,
		"versionCode": versionCode,
		"sha256":      uc.hash,
		"path":        filePath,
		"size":        uc.info.Size(),
		"sizeHuman":   edits.FormatBytes(uc.info.Size()),
		"type":        uc.ext[1:],
		"package":     c.packageName,
		"editId":      edit.ServerID,
		"committed":   !noAutoCommit,
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
	if ext != extAPK && ext != extAAB {
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
	if ext == extAPK {
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
