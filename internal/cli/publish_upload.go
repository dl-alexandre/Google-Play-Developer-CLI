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

// File extension constants
const (
	extAPK = ".apk"
	extAAB = ".aab"
)

func (c *CLI) publishUpload(ctx context.Context, filePath, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	info, err := os.Stat(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath)))
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != extAAB && ext != extAPK {
		return c.OutputError(errors.NewAPIError(errors.CodeValidationError,
			"file must be an AAB or APK").WithHint("Supported formats: .aab, .apk"))
	}

	hash, err := edits.HashFile(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to compute file hash: %v", err)))
	}

	editMgr := edits.NewManager()

	idempotencyResult, idempotencyKey, _ := editMgr.Idempotent.CheckUploadByHash(c.packageName, hash)
	if idempotencyResult != nil && idempotencyResult.Found {
		if data, ok := idempotencyResult.Data.(map[string]interface{}); ok {
			result := output.NewResult(map[string]interface{}{
				"idempotent":  true,
				"versionCode": data["versionCode"],
				"sha256":      data["sha256"],
				"path":        filePath,
				"size":        info.Size(),
				"sizeHuman":   edits.FormatBytes(info.Size()),
				"type":        ext[1:],
				"package":     c.packageName,
				"editId":      data["editId"],
				"recordedAt":  idempotencyResult.Timestamp,
			})
			return c.Output(result.WithNoOp("upload already completed").WithServices("androidpublisher"))
		}
	}

	cached, err := editMgr.GetCachedArtifactByHash(c.packageName, hash)
	if err == nil && cached != nil {
		result := output.NewResult(map[string]interface{}{
			"cached":    true,
			"sha256":    cached.SHA256,
			"path":      filePath,
			"size":      info.Size(),
			"sizeHuman": edits.FormatBytes(info.Size()),
		})
		return c.Output(result.WithNoOp("artifact already uploaded").WithServices("androidpublisher"))
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":    true,
			"action":    "upload",
			"path":      filePath,
			"sha256":    hash,
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

	editMgr, edit, created, err := c.prepareEdit(ctx, publisher, editID)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer f.Close()

	var versionCode int64
	if ext == extAAB {
		bundle, err := publisher.Edits.Bundles.Upload(c.packageName, edit.ServerID).
			Media(f).Context(ctx).Do()
		if err != nil {
			if created {
				_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to upload bundle: %v", err)))
		}
		versionCode = bundle.VersionCode
	} else {
		apk, err := publisher.Edits.Apks.Upload(c.packageName, edit.ServerID).
			Media(f).Context(ctx).Do()
		if err != nil {
			if created {
				_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
			}
			return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
				fmt.Sprintf("failed to upload APK: %v", err)))
		}
		versionCode = apk.VersionCode
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	_ = editMgr.CacheArtifactWithHash(c.packageName, filePath, hash, versionCode)

	uploadResult := &edits.UploadResult{
		VersionCode: versionCode,
		SHA256:      hash,
		Path:        filePath,
		Size:        info.Size(),
		Type:        ext[1:],
		EditID:      edit.ServerID,
	}
	_ = editMgr.Idempotent.RecordUpload(idempotencyKey, c.packageName, hash, uploadResult)

	result := output.NewResult(map[string]interface{}{
		"success":     true,
		"versionCode": versionCode,
		"sha256":      hash,
		"path":        filePath,
		"size":        info.Size(),
		"sizeHuman":   edits.FormatBytes(info.Size()),
		"type":        ext[1:],
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
