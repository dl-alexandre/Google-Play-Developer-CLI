// Package extensions provides the gpd extension system for installable subcommands.
package extensions

import (
	"fmt"
	"os"
	"path/filepath"
)

// Remove removes an installed extension.
func Remove(name string) error {
	// Check if extension exists
	if !IsInstalled(name) {
		return fmt.Errorf("extension %q is not installed", name)
	}

	extDir := filepath.Join(GetExtensionsDir(), name)

	// Remove the extension directory
	if err := os.RemoveAll(extDir); err != nil {
		return fmt.Errorf("removing extension: %w", err)
	}

	return nil
}
