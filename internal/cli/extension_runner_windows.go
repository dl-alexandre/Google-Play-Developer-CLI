//go:build windows
// +build windows

package cli

// syscallExecAvailable returns false on Windows since syscall.Exec is not available.
func syscallExecAvailable() bool {
	// Windows does not support syscall.Exec
	// We always use os/exec.Command on Windows
	return false
}

// execExtension is a no-op on Windows since we always use os/exec.
// This function exists for API compatibility but should never be called on Windows.
func execExtension(execPath string, args []string) error {
	// On Windows, we use exec.Command instead
	// This function should not be called on Windows
	panic("execExtension should not be called on Windows - use exec.Command instead")
}
