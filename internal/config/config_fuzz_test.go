//go:build unit
// +build unit

package config

import (
	"encoding/json"
	"testing"
)

// FuzzConfigJSON fuzz tests config JSON parsing
func FuzzConfigJSON(f *testing.F) {
	// Seed corpus with valid configs
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"defaultPackage":"com.example.app"}`))
	f.Add([]byte(`{"outputFormat":"json"}`))
	f.Add([]byte(`{"timeoutSeconds":30}`))
	f.Add([]byte(`{"activeProfile":"production"}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var cfg Config
		// Try to parse - should not panic
		_ = json.Unmarshal(data, &cfg)

		// Validation should also not panic
		_ = cfg.Validate()
	})
}

// FuzzPackageName fuzz tests package name validation
func FuzzPackageName(f *testing.F) {
	// Seed with valid and invalid names
	f.Add("com.example.app")
	f.Add("invalid package")
	f.Add("123invalid")
	f.Add("com.example.UPPER")
	f.Add("a.b.c")
	f.Add("a")
	f.Add("")

	f.Fuzz(func(t *testing.T, name string) {
		// Should not panic
		_ = isValidPackageName(name)
	})
}
