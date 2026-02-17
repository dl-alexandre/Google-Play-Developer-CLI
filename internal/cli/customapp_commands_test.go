package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dl-alexandre/gpd/internal/errors"
)

func requireAPIError(t *testing.T, err *errors.APIError) *errors.APIError {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error")
		return &errors.APIError{}
	}
	return err
}

func TestValidateCustomAppAPK(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	validAPK := filepath.Join(tempDir, "app.apk")
	if err := os.WriteFile(validAPK, []byte("apk"), 0o644); err != nil {
		t.Fatalf("write apk: %v", err)
	}

	invalidExt := filepath.Join(tempDir, "app.aab")
	if err := os.WriteFile(invalidExt, []byte("aab"), 0o644); err != nil {
		t.Fatalf("write aab: %v", err)
	}

	tests := []struct {
		name     string
		path     string
		wantErr  bool
		wantCode errors.ErrorCode
		wantMsg  string
	}{
		{
			name:     "missing file",
			path:     filepath.Join(tempDir, "missing.apk"),
			wantErr:  true,
			wantCode: errors.CodeValidationError,
			wantMsg:  "file not found",
		},
		{
			name:     "directory path",
			path:     tempDir,
			wantErr:  true,
			wantCode: errors.CodeValidationError,
			wantMsg:  "APK path is a directory",
		},
		{
			name:     "invalid extension",
			path:     invalidExt,
			wantErr:  true,
			wantCode: errors.CodeValidationError,
			wantMsg:  "custom app upload must be an APK",
		},
		{
			name:    "valid apk",
			path:    validAPK,
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotErr := validateCustomAppAPK(tc.path)
			if !tc.wantErr {
				if gotErr != nil {
					t.Fatalf("unexpected error: %v", gotErr)
				}
				return
			}

			apiErr := requireAPIError(t, gotErr)
			if apiErr.Code != tc.wantCode {
				t.Fatalf("expected code %s, got %s", tc.wantCode, apiErr.Code)
			}
			if apiErr.Message != tc.wantMsg {
				t.Fatalf("expected message %q, got %q", tc.wantMsg, apiErr.Message)
			}
		})
	}
}
