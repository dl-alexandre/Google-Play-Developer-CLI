//go:build !windows
// +build !windows

package cli

import "syscall"

// syscallExecAvailable returns true on Unix systems where syscall.Exec is available.
func syscallExecAvailable() bool {
	// syscall.Exec is always available on Unix systems (Linux, macOS, BSD, etc.)
	return true
}

// execExtension executes the extension using syscall.Exec on Unix systems.
func execExtension(execPath string, args []string) error {
	return syscall.Exec(execPath, append([]string{execPath}, args...), syscall.Environ())
}
