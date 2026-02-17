package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/schollz/progressbar/v3"
	"golang.org/x/term"
	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/logging"
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

	hash, err := c.hashFileForUpload(filePath, info)
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

func (c *CLI) hashFileForUpload(filePath string, info os.FileInfo) (string, error) {
	if !c.shouldShowHashProgress(info) {
		return edits.HashFile(filePath)
	}

	bar := progressbar.NewOptions64(
		info.Size(),
		progressbar.OptionSetWriter(c.stderr),
		progressbar.OptionSetDescription("Hashing artifact"),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetWidth(24),
		progressbar.OptionThrottle(100*time.Millisecond),
	)

	var last int64
	hash, err := edits.HashFileWithProgress(filePath, func(processed, _ int64) {
		delta := processed - last
		if delta > 0 {
			_ = bar.Add64(delta)
			last = processed
		}
	})
	if err != nil {
		return "", err
	}
	_ = bar.Finish()
	return hash, nil
}

func (c *CLI) shouldShowHashProgress(info os.FileInfo) bool {
	if c.quiet || info == nil || info.Size() < 32*1024*1024 {
		return false
	}
	f, ok := c.stderr.(*os.File)
	if !ok {
		return false
	}
	// #nosec G115 -- File descriptor fits in int on all supported platforms
	return term.IsTerminal(int(f.Fd()))
}

func (c *CLI) checkIdempotentUpload(uc *uploadContext, editMgr *edits.Manager) *output.Result {
	idempotencyResult, idempotencyKey, err := editMgr.Idempotent.CheckUploadByHash(c.packageName, uc.hash)
	uc.idempotencyKey = idempotencyKey
	if err != nil {
		logging.Warn("failed to check idempotent upload", logging.String("package", c.packageName), logging.Err(err))
		return nil
	}

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

	if uc.ext == extAAB {
		bundle, uploadErr := publisher.Edits.Bundles.Upload(c.packageName, editID).Media(f).Context(ctx).Do()
		closeErr := f.Close()
		if uploadErr != nil {
			if closeErr != nil {
				return 0, fmt.Errorf("failed to upload bundle: %v; close error: %v", uploadErr, closeErr)
			}
			return 0, fmt.Errorf("failed to upload bundle: %w", uploadErr)
		}
		if closeErr != nil {
			return 0, fmt.Errorf("failed to close artifact: %w", closeErr)
		}
		return bundle.VersionCode, nil
	}

	apk, uploadErr := publisher.Edits.Apks.Upload(c.packageName, editID).Media(f).Context(ctx).Do()
	closeErr := f.Close()
	if uploadErr != nil {
		if closeErr != nil {
			return 0, fmt.Errorf("failed to upload APK: %v; close error: %v", uploadErr, closeErr)
		}
		return 0, fmt.Errorf("failed to upload APK: %w", uploadErr)
	}
	if closeErr != nil {
		return 0, fmt.Errorf("failed to close artifact: %w", closeErr)
	}
	return apk.VersionCode, nil
}

func (c *CLI) publishUpload(ctx context.Context, filePath string, obbOpts obbOptions, editID string, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	uc, err := c.validateUploadFile(filePath)
	if err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if apiErr := validateObbOptions(uc.ext, obbOpts); apiErr != nil {
		return c.OutputError(apiErr)
	}

	var obbMainInfo os.FileInfo
	var obbPatchInfo os.FileInfo
	if obbOpts.mainPath != "" {
		info, apiErr := validateObbFile(obbOpts.mainPath)
		if apiErr != nil {
			return c.OutputError(apiErr)
		}
		obbMainInfo = info
	}
	if obbOpts.patchPath != "" {
		info, apiErr := validateObbFile(obbOpts.patchPath)
		if apiErr != nil {
			return c.OutputError(apiErr)
		}
		obbPatchInfo = info
	}

	editMgr := edits.NewManager()
	hasObb := hasObbOptions(obbOpts)

	if result := c.checkIdempotentUpload(uc, editMgr); result != nil {
		if !hasObb {
			return c.Output(result)
		}
	}

	if result := c.checkCachedUpload(uc, editMgr); result != nil {
		if !hasObb {
			return c.Output(result)
		}
	}

	if dryRun {
		resultData := c.buildUploadDryRunResult(filePath, uc, obbMainInfo, obbPatchInfo, obbOpts)
		result := output.NewResult(resultData)
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

	versionCode, err := c.uploadArtifact(ctx, publisher, edit.ServerID, uc)
	if err != nil {
		if created {
			if cleanupErr := publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do(); cleanupErr != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.ServerID), logging.Err(cleanupErr))
			}
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}

	obbResults, apiErr := c.uploadObbFiles(ctx, publisher, edit.ServerID, versionCode, obbOpts)
	if apiErr != nil {
		if created {
			if cleanupErr := publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do(); cleanupErr != nil {
				logging.Warn("failed to delete edit", logging.String("package", c.packageName), logging.String("editId", edit.ServerID), logging.Err(cleanupErr))
			}
		}
		return c.OutputError(apiErr)
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	if err := editMgr.CacheArtifactWithHash(c.packageName, filePath, uc.hash, versionCode); err != nil {
		logging.Warn("failed to cache artifact", logging.String("package", c.packageName), logging.String("path", filePath), logging.Err(err))
	}

	uploadResult := &edits.UploadResult{
		VersionCode: versionCode,
		SHA256:      uc.hash,
		Path:        filePath,
		Size:        uc.info.Size(),
		Type:        uc.ext[1:],
		EditID:      edit.ServerID,
	}
	if err := editMgr.Idempotent.RecordUpload(uc.idempotencyKey, c.packageName, uc.hash, uploadResult); err != nil {
		logging.Warn("failed to record idempotent upload", logging.String("package", c.packageName), logging.String("editId", edit.ServerID), logging.Err(err))
	}

	resultData := map[string]interface{}{
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
	}
	if len(obbResults) > 0 {
		resultData["expansionFiles"] = obbResults
	}
	result := output.NewResult(resultData)
	return c.Output(result.WithServices("androidpublisher"))
}

func (c *CLI) buildUploadDryRunResult(filePath string, uc *uploadContext, obbMainInfo, obbPatchInfo os.FileInfo, obbOpts obbOptions) map[string]interface{} {
	resultData := map[string]interface{}{
		"dryRun":    true,
		"action":    "upload",
		"path":      filePath,
		"sha256":    uc.hash,
		"size":      uc.info.Size(),
		"sizeHuman": edits.FormatBytes(uc.info.Size()),
		"type":      uc.ext[1:],
		"package":   c.packageName,
	}
	obbData := map[string]interface{}{}
	if obbMainInfo != nil {
		obbData["main"] = map[string]interface{}{
			"path":      obbOpts.mainPath,
			"size":      obbMainInfo.Size(),
			"sizeHuman": edits.FormatBytes(obbMainInfo.Size()),
		}
	} else if obbOpts.mainReferenceVersion > 0 {
		obbData["main"] = map[string]interface{}{
			"referencesVersion": obbOpts.mainReferenceVersion,
		}
	}
	if obbPatchInfo != nil {
		obbData["patch"] = map[string]interface{}{
			"path":      obbOpts.patchPath,
			"size":      obbPatchInfo.Size(),
			"sizeHuman": edits.FormatBytes(obbPatchInfo.Size()),
		}
	} else if obbOpts.patchReferenceVersion > 0 {
		obbData["patch"] = map[string]interface{}{
			"referencesVersion": obbOpts.patchReferenceVersion,
		}
	}
	if len(obbData) > 0 {
		resultData["expansionFiles"] = obbData
	}
	return resultData
}

func hasObbOptions(opts obbOptions) bool {
	return opts.mainPath != "" || opts.patchPath != "" || opts.mainReferenceVersion > 0 || opts.patchReferenceVersion > 0
}

func (c *CLI) publishInternalShareUpload(ctx context.Context, filePath string, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		result := output.NewErrorResult(err.(*errors.APIError)).WithServices("androidpublisher")
		return c.Output(result)
	}
	info, err := os.Stat(filePath)
	if err != nil {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath))).WithServices("androidpublisher")
		return c.Output(result)
	}
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != extAPK && ext != extAAB {
		result := output.NewErrorResult(errors.NewAPIError(errors.CodeValidationError,
			"file must be an APK or AAB").WithHint("Supported formats: .apk, .aab")).WithServices("androidpublisher")
		return c.Output(result)
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
	var resp *androidpublisher.InternalAppSharingArtifact
	if ext == extAPK {
		resp, err = publisher.Internalappsharingartifacts.Uploadapk(c.packageName).Media(f).Context(ctx).Do()
	} else {
		resp, err = publisher.Internalappsharingartifacts.Uploadbundle(c.packageName).Media(f).Context(ctx).Do()
	}
	closeErr := f.Close()
	if err != nil {
		if closeErr != nil {
			logging.Warn("failed to close artifact", logging.String("path", filePath), logging.Err(closeErr))
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	if closeErr != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, closeErr.Error()))
	}
	result := output.NewResult(map[string]interface{}{
		"success":  true,
		"artifact": resp,
		"package":  c.packageName,
	})
	return c.Output(result.WithServices("androidpublisher"))
}
