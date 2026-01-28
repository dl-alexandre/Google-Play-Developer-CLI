package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dl-alexandre/gpd/internal/errors"
)

func TestResolveTokenInput(t *testing.T) {
	t.Helper()

	tempDir := t.TempDir()
	emptyFile := filepath.Join(tempDir, "empty.txt")
	if err := os.WriteFile(emptyFile, []byte(" \n "), 0o644); err != nil {
		t.Fatalf("write empty file: %v", err)
	}

	tokenFile := filepath.Join(tempDir, "token.txt")
	if err := os.WriteFile(tokenFile, []byte("  token-value \n"), 0o644); err != nil {
		t.Fatalf("write token file: %v", err)
	}

	tests := []struct {
		name      string
		token     string
		tokenFile string
		wantErr   bool
		wantCode  errors.ErrorCode
		wantMsg   string
		wantValue string
	}{
		{
			name:      "both token and file",
			token:     "value",
			tokenFile: tokenFile,
			wantErr:   true,
			wantCode:  errors.CodeValidationError,
			wantMsg:   "provide --token or --token-file, not both",
		},
		{
			name:     "whitespace token",
			token:    "   ",
			wantErr:  true,
			wantCode: errors.CodeValidationError,
			wantMsg:  "integrity token is required",
		},
		{
			name:      "token trimmed",
			token:     "  abc \n",
			wantValue: "abc",
		},
		{
			name:      "missing token file",
			tokenFile: filepath.Join(tempDir, "missing.txt"),
			wantErr:   true,
			wantCode:  errors.CodeValidationError,
			wantMsg:   "failed to read token file",
		},
		{
			name:      "empty token file",
			tokenFile: emptyFile,
			wantErr:   true,
			wantCode:  errors.CodeValidationError,
			wantMsg:   "token file is empty",
		},
		{
			name:      "token file trimmed",
			tokenFile: tokenFile,
			wantValue: "token-value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			value, err := resolveTokenInput(tc.token, tc.tokenFile)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				if err.Code != tc.wantCode {
					t.Fatalf("expected code %s, got %s", tc.wantCode, err.Code)
				}
				if err.Message != tc.wantMsg {
					t.Fatalf("expected message %q, got %q", tc.wantMsg, err.Message)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if value != tc.wantValue {
				t.Fatalf("expected value %q, got %q", tc.wantValue, value)
			}
		})
	}
}
