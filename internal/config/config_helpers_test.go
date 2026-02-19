package config

import "testing"

func TestIsValidPackageNameExtended(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{
			name:  "valid lowercase",
			value: "com.example.app",
			want:  true,
		},
		{
			name:  "valid with numbers",
			value: "app123.version456",
			want:  true,
		},
		{
			name:  "valid with underscore",
			value: "com_example_app",
			want:  true,
		},
		{
			name:  "single letter",
			value: "a",
			want:  true,
		},
		{
			name:  "empty string",
			value: "",
			want:  false,
		},
		{
			name:  "uppercase first letter",
			value: "Com.example.app",
			want:  false,
		},
		{
			name:  "uppercase in middle",
			value: "com.Example.app",
			want:  false,
		},
		{
			name:  "invalid character hyphen",
			value: "com-example-app",
			want:  false,
		},
		{
			name:  "invalid character space",
			value: "com example app",
			want:  false,
		},
		{
			name:  "invalid character special",
			value: "com@example",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidPackageName(tt.value); got != tt.want {
				t.Errorf("isValidPackageName(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
