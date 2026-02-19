package auth

import (
	"runtime"
	"testing"
)

func TestValidatePathExtended(t *testing.T) {
	// Use platform-appropriate paths
	var absPath1, absPath2 string
	if runtime.GOOS == "windows" {
		absPath1 = `C:\Program Files\gpd`
		absPath2 = `C:\Users\user\AppData\gpd`
	} else {
		absPath1 = "/usr/local/bin"
		absPath2 = "/home/user/.config/gpd"
	}

	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid absolute path",
			path:    absPath1,
			wantErr: false,
		},
		{
			name:    "valid absolute path with subdirs",
			path:    absPath2,
			wantErr: false,
		},
		{
			name:    "relative path",
			path:    "./config",
			wantErr: true,
		},
		{
			name:    "relative path with dots",
			path:    "../config",
			wantErr: true,
		},
		{
			name:    "empty path",
			path:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()

	if err := validatePath(tmpDir); err != nil {
		t.Errorf("validatePath(%q) should not error, got: %v", tmpDir, err)
	}

	if err := validatePath("./relative"); err == nil {
		t.Error("validatePath(\"./relative\") should error for relative path")
	}
}
