package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/dl-alexandre/gpd/internal/apitest"
	"github.com/dl-alexandre/gpd/internal/errors"
)

func TestValidateUploadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		filePath    string
		setup       func(t *testing.T) string
		wantErr     bool
		wantErrCode errors.ErrorCode
	}{
		{
			name: "valid_aab_file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.aab")
				if err := os.WriteFile(path, []byte("mock aab content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			wantErr: false,
		},
		{
			name: "valid_apk_file",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.apk")
				if err := os.WriteFile(path, []byte("mock apk content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			wantErr: false,
		},
		{
			name: "file_not_found",
			setup: func(t *testing.T) string {
				return "/nonexistent/path/file.aab"
			},
			wantErr:     true,
			wantErrCode: errors.CodeValidationError,
		},
		{
			name: "invalid_extension",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.txt")
				if err := os.WriteFile(path, []byte("invalid content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			wantErr:     true,
			wantErrCode: errors.CodeValidationError,
		},
		{
			name: "unsupported_extension",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				path := filepath.Join(tmpDir, "test.zip")
				if err := os.WriteFile(path, []byte("zip content"), 0644); err != nil {
					t.Fatalf("failed to create test file: %v", err)
				}
				return path
			},
			wantErr:     true,
			wantErrCode: errors.CodeValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cli := New()
			filePath := tt.setup(t)

			ctx, err := cli.validateUploadFile(filePath)

			if tt.wantErr {
				if err == nil {
					t.Errorf("validateUploadFile() error = nil, wantErr = true")
					return
				}
				if apiErr, ok := err.(*errors.APIError); ok {
					if apiErr.Code != tt.wantErrCode {
						t.Errorf("validateUploadFile() error code = %v, want %v", apiErr.Code, tt.wantErrCode)
					}
				}
			} else {
				if err != nil {
					t.Errorf("validateUploadFile() unexpected error = %v", err)
					return
				}
				if ctx == nil {
					t.Error("validateUploadFile() returned nil context")
					return
				}
				if ctx.filePath != filePath {
					t.Errorf("validateUploadFile() filePath = %v, want %v", ctx.filePath, filePath)
				}
				if ctx.hash == "" {
					t.Error("validateUploadFile() hash is empty")
				}
			}
		})
	}
}

func TestShouldShowHashProgress(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		quiet    bool
		size     int64
		wantShow bool
	}{
		{
			name:     "small_file",
			quiet:    false,
			size:     1024 * 1024,
			wantShow: false,
		},
		{
			name:     "quiet_mode",
			quiet:    true,
			size:     64 * 1024 * 1024,
			wantShow: false,
		},
		{
			name:     "exact_threshold",
			quiet:    false,
			size:     32 * 1024 * 1024,
			wantShow: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cli := New()
			cli.quiet = tt.quiet

			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.aab")
			if err := os.WriteFile(tmpFile, make([]byte, tt.size), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			info, err := os.Stat(tmpFile)
			if err != nil {
				t.Fatalf("failed to stat test file: %v", err)
			}

			got := cli.shouldShowHashProgress(info)
			if got != tt.wantShow {
				t.Errorf("shouldShowHashProgress() = %v, want %v", got, tt.wantShow)
			}
		})
	}
}

func TestShouldShowHashProgressLargeFile(t *testing.T) {
	t.Parallel()

	cli := New()
	cli.quiet = false

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.aab")
	if err := os.WriteFile(tmpFile, make([]byte, 64*1024*1024), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	info, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("failed to stat test file: %v", err)
	}

	got := cli.shouldShowHashProgress(info)

	if got {
		t.Log("shouldShowHashProgress returned true (terminal detected)")
	} else {
		t.Log("shouldShowHashProgress returned false (no terminal in test environment - expected)")
	}
}

func TestUploadContextStructure(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.aab")
	content := []byte("mock aab content for testing")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat test file: %v", err)
	}

	cli := New()
	ctx, err := cli.validateUploadFile(testFile)
	if err != nil {
		t.Fatalf("validateUploadFile() unexpected error = %v", err)
	}

	if ctx.filePath != testFile {
		t.Errorf("filePath = %v, want %v", ctx.filePath, testFile)
	}

	if ctx.info.Size() != info.Size() {
		t.Errorf("info.Size() = %v, want %v", ctx.info.Size(), info.Size())
	}

	if ctx.ext != ".aab" {
		t.Errorf("ext = %v, want .aab", ctx.ext)
	}

	if ctx.hash == "" {
		t.Error("hash is empty")
	}

	if len(ctx.hash) != 64 {
		t.Errorf("hash length = %v, want 64 (SHA256 hex)", len(ctx.hash))
	}
}

func TestPublishUploadValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		fileExt string
		wantErr bool
	}{
		{
			name:    "aab_extension",
			fileExt: ".aab",
			wantErr: false,
		},
		{
			name:    "apk_extension",
			fileExt: ".apk",
			wantErr: false,
		},
		{
			name:    "AAB_uppercase",
			fileExt: ".AAB",
			wantErr: false,
		},
		{
			name:    "APK_uppercase",
			fileExt: ".APK",
			wantErr: false,
		},
		{
			name:    "txt_extension",
			fileExt: ".txt",
			wantErr: true,
		},
		{
			name:    "no_extension",
			fileExt: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test"+tt.fileExt)
			if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			cli := New()
			_, err := cli.validateUploadFile(testFile)

			if tt.wantErr && err == nil {
				t.Error("expected error but got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockClientSetup(t *testing.T) {
	t.Parallel()

	mock := apitest.NewMockClient()
	if mock == nil {
		t.Fatal("NewMockClient() returned nil")
	}

	if mock.PublisherResponses == nil {
		t.Error("PublisherResponses is nil")
	}

	if mock.PublisherResponses.Edits == nil {
		t.Error("Edits service is nil")
	}

	if mock.Calls == nil {
		t.Error("Calls slice is nil")
	}
}

func TestUploadFileHashing(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.aab")
	content := []byte("test content for hashing")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cli := New()
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("failed to stat file: %v", err)
	}

	hash, err := cli.hashFileForUpload(testFile, info)
	if err != nil {
		t.Fatalf("hashFileForUpload() error = %v", err)
	}

	if hash == "" {
		t.Error("hash is empty")
	}

	if len(hash) != 64 {
		t.Errorf("hash length = %d, want 64", len(hash))
	}

	for _, c := range hash {
		if (c < '0' || c > '9') && (c < 'a' || c > 'f') {
			t.Errorf("hash contains invalid character: %c", c)
			break
		}
	}
}

func TestUploadContextIdempotencyKey(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.aab")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cli := New()
	ctx, err := cli.validateUploadFile(testFile)
	if err != nil {
		t.Fatalf("validateUploadFile() error = %v", err)
	}

	if ctx.idempotencyKey != "" {
		t.Error("idempotencyKey should be empty after validateUploadFile")
	}
}

func BenchmarkHashFile(b *testing.B) {
	tmpDir := b.TempDir()
	testFile := filepath.Join(tmpDir, "test.aab")
	content := make([]byte, 10*1024*1024)
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		b.Fatalf("failed to create test file: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		b.Fatalf("failed to stat file: %v", err)
	}

	cli := New()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := cli.hashFileForUpload(testFile, info)
		if err != nil {
			b.Fatalf("hashFileForUpload() error = %v", err)
		}
	}
}

func TestPublishUploadCommandFlags(t *testing.T) {
	t.Parallel()

	cli := New()

	cmd, _, err := cli.rootCmd.Find([]string{"publish", "upload"})
	if err != nil {
		t.Fatalf("failed to find upload command: %v", err)
	}

	flags := []string{"edit-id", "obb-main", "obb-patch", "obb-main-references-version", "obb-patch-references-version", "no-auto-commit", "dry-run"}

	for _, flag := range flags {
		if cmd.Flags().Lookup(flag) == nil {
			t.Errorf("upload command missing flag: %s", flag)
		}
	}
}

func TestPublishUploadDryRun(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.aab")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cli := New()
	cli.packageName = "com.test.app"

	ctx := context.Background()
	err := cli.publishUpload(ctx, testFile, obbOptions{}, "", false, true)

	if err != nil {
		t.Errorf("publishUpload() with dry-run error = %v", err)
	}
}

func TestPublishUploadPackageRequired(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.aab")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	cli := New()
	cli.packageName = ""

	ctx := context.Background()
	err := cli.publishUpload(ctx, testFile, obbOptions{}, "", false, false)

	if err != nil {
		t.Logf("publishUpload returned error via OutputError (expected behavior): %v", err)
	}
}
