package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/androidpublisher/v3"
	"google.golang.org/api/googleapi"

	"github.com/dl-alexandre/gpd/internal/edits"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func (c *CLI) publishDeobfuscationUpload(ctx context.Context, filePath, fileType string, versionCode int64, editID string, chunkSize int64, noAutoCommit, dryRun bool) error {
	if err := c.requirePackage(); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	info, apiErr := validateDeobfuscationFile(filePath, fileType)
	if apiErr != nil {
		return c.OutputError(apiErr)
	}

	if dryRun {
		result := output.NewResult(map[string]interface{}{
			"dryRun":      true,
			"action":      "deobfuscation_upload",
			"path":        filePath,
			"size":        info.Size(),
			"sizeHuman":   edits.FormatBytes(info.Size()),
			"type":        fileType,
			"versionCode": versionCode,
			"package":     c.packageName,
		})
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
	defer func() { _ = editMgr.ReleaseLock(c.packageName) }()

	if err := c.ensureVersionCodeExists(ctx, publisher, edit.ServerID, versionCode); err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(err)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError, err.Error()))
	}
	defer f.Close()

	call := publisher.Edits.Deobfuscationfiles.Upload(c.packageName, edit.ServerID, versionCode, fileType)
	if chunkSize > 0 {
		call.Media(f, googleapi.ChunkSize(int(chunkSize)))
	} else {
		call.Media(f)
	}

	resp, err := call.Context(ctx).Do()
	if err != nil {
		if created {
			_ = publisher.Edits.Delete(c.packageName, edit.ServerID).Context(ctx).Do()
		}
		return c.OutputError(errors.NewAPIError(errors.CodeGeneralError,
			fmt.Sprintf("failed to upload deobfuscation file: %v", err)))
	}

	if err := c.finalizeEdit(ctx, publisher, editMgr, edit, !noAutoCommit); err != nil {
		return c.OutputError(err.(*errors.APIError))
	}

	result := output.NewResult(map[string]interface{}{
		"success":     true,
		"type":        fileType,
		"versionCode": versionCode,
		"package":     c.packageName,
		"size":        info.Size(),
		"sizeHuman":   edits.FormatBytes(info.Size()),
		"editId":      edit.ServerID,
		"committed":   !noAutoCommit,
		"response":    resp,
	})
	return c.Output(result.WithServices("androidpublisher"))
}

func validateDeobfuscationFile(filePath, fileType string) (os.FileInfo, *errors.APIError) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("file not found: %s", filePath))
	}

	switch fileType {
	case "proguard":
		if info.Size() > 50*1024*1024 {
			return nil, errors.NewAPIError(errors.CodeValidationError, "proguard mapping file exceeds 50MB")
		}
		if !looksLikeProguardMapping(filePath) {
			return nil, errors.NewAPIError(errors.CodeValidationError, "proguard mapping file format invalid")
		}
	case "nativeCode":
		if info.Size() > 100*1024*1024 {
			return nil, errors.NewAPIError(errors.CodeValidationError, "native symbols file exceeds 100MB")
		}
		lower := strings.ToLower(filePath)
		ext := strings.ToLower(filepath.Ext(filePath))
		if ext != ".zip" && ext != ".sym" && !strings.HasSuffix(lower, ".so.sym") {
			return nil, errors.NewAPIError(errors.CodeValidationError, "native symbols file must be .so.sym or .zip")
		}
	default:
		return nil, errors.NewAPIError(errors.CodeValidationError, "type must be proguard or nativeCode")
	}

	return info, nil
}

func looksLikeProguardMapping(filePath string) bool {
	f, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lines := 0
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if strings.Contains(line, "->") && strings.HasSuffix(line, ":") {
			return true
		}
		lines++
		if lines > 50 {
			break
		}
	}
	return false
}

func (c *CLI) ensureVersionCodeExists(ctx context.Context, publisher *androidpublisher.Service, editID string, versionCode int64) *errors.APIError {
	if versionCode <= 0 {
		return errors.NewAPIError(errors.CodeValidationError, "version code must be greater than zero")
	}

	bundles, err := publisher.Edits.Bundles.List(c.packageName, editID).Context(ctx).Do()
	if err == nil {
		for _, bundle := range bundles.Bundles {
			if bundle.VersionCode == versionCode {
				return nil
			}
		}
	}

	apks, err := publisher.Edits.Apks.List(c.packageName, editID).Context(ctx).Do()
	if err == nil {
		for _, apk := range apks.Apks {
			if apk.VersionCode == versionCode {
				return nil
			}
		}
	}

	return errors.NewAPIError(errors.CodeValidationError,
		fmt.Sprintf("version code %d not found in edit", versionCode)).
		WithHint("Upload an APK/AAB in this edit before uploading deobfuscation files")
}
