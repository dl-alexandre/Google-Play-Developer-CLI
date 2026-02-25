//go:build unit
// +build unit

package fastlane

import (
	"testing"
)

// FuzzParseDirectory fuzz tests directory parsing (requires filesystem, mostly for checking no panic)
func FuzzParseDirectory(f *testing.F) {
	// Just test that ParseDirectory doesn't panic on various inputs
	f.Add("/tmp/nonexistent")
	f.Add("en-US")
	f.Add("de-DE")
	f.Add("ja-JP")

	f.Fuzz(func(t *testing.T, dir string) {
		// Should not panic (may return error, but shouldn't panic)
		_, _ = ParseDirectory(dir)
	})
}

// FuzzWriteDirectory fuzz tests directory writing
func FuzzWriteDirectory(f *testing.F) {
	// Test with various directory names
	f.Add("/tmp/test-output")
	f.Add("output")
	f.Add("")

	f.Fuzz(func(t *testing.T, dir string) {
		// Should not panic
		_ = WriteDirectory(dir, nil)
	})
}

// FuzzIsImageFile fuzz tests image file detection
func FuzzIsImageFile(f *testing.F) {
	f.Add("image.png")
	f.Add("image.jpg")
	f.Add("image.jpeg")
	f.Add("document.txt")
	f.Add("script.sh")
	f.Add("")
	f.Add("UPPER.PNG")

	f.Fuzz(func(t *testing.T, filename string) {
		// Should not panic
		_ = isImageFile(filename)
	})
}

// FuzzIsScreenshotDir fuzz tests screenshot directory detection
func FuzzIsScreenshotDir(f *testing.F) {
	f.Add("phoneScreenshots")
	f.Add("sevenInchScreenshots")
	f.Add("tenInchScreenshots")
	f.Add("invalidDir")
	f.Add("")

	f.Fuzz(func(t *testing.T, name string) {
		// Should not panic
		_ = isScreenshotDir(name)
	})
}
