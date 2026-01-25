package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
)

type obbOptions struct {
	mainPath              string
	patchPath             string
	mainReferenceVersion  int64
	patchReferenceVersion int64
}

func validateObbOptions(ext string, opts obbOptions) *errors.APIError {
	hasObb := opts.mainPath != "" || opts.patchPath != "" || opts.mainReferenceVersion > 0 || opts.patchReferenceVersion > 0
	if hasObb && ext != extAPK {
		return errors.NewAPIError(errors.CodeValidationError, "expansion files are only supported for APK uploads")
	}
	if opts.mainPath != "" && opts.mainReferenceVersion > 0 {
		return errors.NewAPIError(errors.CodeValidationError, "cannot set both --obb-main and --obb-main-references-version")
	}
	if opts.patchPath != "" && opts.patchReferenceVersion > 0 {
		return errors.NewAPIError(errors.CodeValidationError, "cannot set both --obb-patch and --obb-patch-references-version")
	}
	if opts.mainReferenceVersion < 0 || opts.patchReferenceVersion < 0 {
		return errors.NewAPIError(errors.CodeValidationError, "expansion file reference version must be positive")
	}
	return nil
}

func validateObbFile(path string) (os.FileInfo, *errors.APIError) {
	if strings.TrimSpace(path) == "" {
		return nil, errors.NewAPIError(errors.CodeValidationError, "expansion file path is required")
	}
	info, err := os.Stat(path)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("file not found: %s", path))
	}
	if info.IsDir() {
		return nil, errors.NewAPIError(errors.CodeValidationError, fmt.Sprintf("file is a directory: %s", path))
	}
	return info, nil
}

func (c *CLI) uploadObbFiles(ctx context.Context, publisher *androidpublisher.Service, editID string, versionCode int64, opts obbOptions) (map[string]*androidpublisher.ExpansionFile, *errors.APIError) {
	results := map[string]*androidpublisher.ExpansionFile{}
	if opts.mainPath != "" {
		resp, err := uploadObbFile(ctx, publisher, c.packageName, editID, versionCode, "main", opts.mainPath)
		if err != nil {
			return nil, err
		}
		results["main"] = resp
	} else if opts.mainReferenceVersion > 0 {
		resp, err := updateObbReference(ctx, publisher, c.packageName, editID, versionCode, "main", opts.mainReferenceVersion)
		if err != nil {
			return nil, err
		}
		results["main"] = resp
	}

	if opts.patchPath != "" {
		resp, err := uploadObbFile(ctx, publisher, c.packageName, editID, versionCode, "patch", opts.patchPath)
		if err != nil {
			return nil, err
		}
		results["patch"] = resp
	} else if opts.patchReferenceVersion > 0 {
		resp, err := updateObbReference(ctx, publisher, c.packageName, editID, versionCode, "patch", opts.patchReferenceVersion)
		if err != nil {
			return nil, err
		}
		results["patch"] = resp
	}

	if len(results) == 0 {
		return nil, nil
	}
	return results, nil
}

func uploadObbFile(ctx context.Context, publisher *androidpublisher.Service, packageName, editID string, versionCode int64, obbType, filePath string) (*androidpublisher.ExpansionFile, *errors.APIError) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}

	resp, err := publisher.Edits.Expansionfiles.Upload(packageName, editID, versionCode, obbType).Media(f).Context(ctx).Do()
	closeErr := f.Close()
	if err != nil {
		if closeErr != nil {
			return nil, errors.NewAPIError(errors.CodeGeneralError, fmt.Sprintf("%v; close error: %v", err, closeErr))
		}
		return nil, errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}
	if closeErr != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, closeErr.Error())
	}
	if resp == nil {
		return nil, nil
	}
	return resp.ExpansionFile, nil
}

func updateObbReference(ctx context.Context, publisher *androidpublisher.Service, packageName, editID string, versionCode int64, obbType string, referenceVersion int64) (*androidpublisher.ExpansionFile, *errors.APIError) {
	expansion := &androidpublisher.ExpansionFile{ReferencesVersion: referenceVersion}
	resp, err := publisher.Edits.Expansionfiles.Update(packageName, editID, versionCode, obbType, expansion).Context(ctx).Do()
	if err != nil {
		return nil, errors.NewAPIError(errors.CodeGeneralError, err.Error())
	}
	return resp, nil
}
