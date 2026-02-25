// Package testutil provides utilities for Go testing including golden file management.
package testutil

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// UpdateGoldenFiles is set by the -update flag to regenerate golden files
var UpdateGoldenFiles = flag.Bool("update", false, "Update golden files")

// GoldenFile represents a golden file for snapshot testing
type GoldenFile struct {
	t      *testing.T
	name   string
	dir    string
	update bool
}

// NewGolden creates a new golden file handler
func NewGolden(t *testing.T, name string) *GoldenFile {
	t.Helper()
	return &GoldenFile{
		t:      t,
		name:   name,
		dir:    filepath.Join("testdata", "golden"),
		update: *UpdateGoldenFiles,
	}
}

// WithDir sets a custom directory for the golden file
func (g *GoldenFile) WithDir(dir string) *GoldenFile {
	t := g.t
	t.Helper()
	return &GoldenFile{
		t:      t,
		name:   g.name,
		dir:    dir,
		update: g.update,
	}
}

// Path returns the full path to the golden file
func (g *GoldenFile) Path() string {
	return filepath.Join(g.dir, g.name)
}

// Compare compares the actual output against the golden file
func (g *GoldenFile) Compare(actual []byte) error {
	t := g.t
	t.Helper()

	// Ensure directory exists
	if err := os.MkdirAll(g.dir, 0755); err != nil {
		return fmt.Errorf("creating golden directory: %w", err)
	}

	path := g.Path()

	// Update mode: write the actual output
	if g.update {
		if err := os.WriteFile(path, actual, 0644); err != nil {
			return fmt.Errorf("writing golden file: %w", err)
		}
		t.Logf("Updated golden file: %s", path)
		return nil
	}

	// Compare mode: read expected and compare
	expected, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("golden file not found: %s (run with -update to create)", path)
		}
		return fmt.Errorf("reading golden file: %w", err)
	}

	if !bytes.Equal(expected, actual) {
		return &GoldenMismatch{
			GoldenFile: path,
			Expected:   string(expected),
			Actual:     string(actual),
		}
	}

	return nil
}

// CompareString compares string output against golden file
func (g *GoldenFile) CompareString(actual string) error {
	return g.Compare([]byte(actual))
}

// GoldenMismatch represents a mismatch between expected and actual output
type GoldenMismatch struct {
	GoldenFile string
	Expected   string
	Actual     string
}

func (m *GoldenMismatch) Error() string {
	return fmt.Sprintf("golden file mismatch: %s\nExpected:\n%s\nActual:\n%s",
		m.GoldenFile, m.Expected, m.Actual)
}

// Diff returns a simple diff of expected vs actual
func (m *GoldenMismatch) Diff() string {
	expectedLines := strings.Split(m.Expected, "\n")
	actualLines := strings.Split(m.Actual, "\n")

	var diff strings.Builder
	diff.WriteString(fmt.Sprintf("--- %s (expected)\n", m.GoldenFile))
	diff.WriteString("+++ actual\n\n")

	maxLen := len(expectedLines)
	if len(actualLines) > maxLen {
		maxLen = len(actualLines)
	}

	for i := 0; i < maxLen; i++ {
		var exp, act string
		if i < len(expectedLines) {
			exp = expectedLines[i]
		}
		if i < len(actualLines) {
			act = actualLines[i]
		}

		if exp != act {
			diff.WriteString(fmt.Sprintf("Line %d:\n", i+1))
			diff.WriteString(fmt.Sprintf("- %s\n", exp))
			diff.WriteString(fmt.Sprintf("+ %s\n", act))
		}
	}

	return diff.String()
}

// Assert compares and fails the test if mismatch
func (g *GoldenFile) Assert(actual []byte) {
	t := g.t
	t.Helper()

	if err := g.Compare(actual); err != nil {
		if mismatch, ok := err.(*GoldenMismatch); ok {
			t.Errorf("Golden file mismatch:\n%s", mismatch.Diff())
		} else {
			t.Errorf("Golden file error: %v", err)
		}
	}
}

// AssertString compares string and fails the test if mismatch
func (g *GoldenFile) AssertString(actual string) {
	g.Assert([]byte(actual))
}

// Exists checks if the golden file exists
func (g *GoldenFile) Exists() bool {
	_, err := os.Stat(g.Path())
	return err == nil
}

// Read reads the golden file content
func (g *GoldenFile) Read() ([]byte, error) {
	return os.ReadFile(g.Path())
}

// ReadString reads the golden file as string
func (g *GoldenFile) ReadString() (string, error) {
	b, err := g.Read()
	return string(b), err
}
