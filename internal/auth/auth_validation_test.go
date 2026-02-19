package auth

import (
	"testing"
)

func TestValidatePathExtended(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid absolute path",
			path:    "/usr/local/bin",
			wantErr: false,
		},
		{
			name:    "valid absolute path with subdirs",
			path:    "/home/user/.config/gpd",
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
