package cli

import (
	"os"
	"os/exec"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/extensions"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/logging"
)

// tryRunExtension attempts to execute an extension if the command is an extension name.
// Returns true if an extension was found and executed, false otherwise.
func tryRunExtension(args []string) bool {
	if len(args) == 0 {
		return false
	}

	cmdName := args[0]

	// Skip if it's a known global flag or built-in command
	if isGlobalFlag(cmdName) || extensions.IsBuiltInCommand(cmdName) {
		return false
	}

	// Check if this is an installed extension
	if !extensions.IsInstalled(cmdName) {
		return false
	}

	// Get the extension executable path
	execPath, err := extensions.GetExecutablePath(cmdName)
	if err != nil {
		logging.Debug("Failed to get extension executable path",
			logging.String("extension", cmdName),
			logging.String("error", err.Error()),
		)
		return false
	}

	// Verify the executable exists
	if _, err := os.Stat(execPath); os.IsNotExist(err) {
		logging.Debug("Extension executable not found",
			logging.String("extension", cmdName),
			logging.String("path", execPath),
		)
		return false
	}

	// Execute the extension with remaining arguments
	// Use syscall.Exec on Unix systems for proper signal handling
	// Fall back to os/exec on Windows
	extArgs := args[1:]

	logging.Debug("Executing extension",
		logging.String("extension", cmdName),
		logging.String("executable", execPath),
		logging.Int("arg_count", len(extArgs)),
	)

	// Try execve on Unix for proper signal handling
	if syscallExecAvailable() {
		err := execExtension(execPath, extArgs)
		if err != nil {
			logging.Debug("syscall.Exec failed, falling back to os/exec",
				logging.String("error", err.Error()),
			)
			// Fall through to os/exec
		}
	}

	// Use os/exec for Windows or as fallback
	cmd := exec.Command(execPath, extArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Propagate the exit code
			if exitErr.ExitCode() != 0 {
				os.Exit(exitErr.ExitCode())
			}
		} else {
			logging.Debug("Extension execution failed",
				logging.String("extension", cmdName),
				logging.String("error", err.Error()),
			)
			os.Exit(1)
		}
	}

	os.Exit(0)
	return true // Should never reach here
}

// isGlobalFlag checks if the argument is a global flag.
func isGlobalFlag(arg string) bool {
	globalFlags := []string{
		"-h", "--help",
		"-v", "--version",
		"--verbose",
		"--package",
		"--output",
		"--pretty",
		"--timeout",
		"--key",
		"--profile",
	}

	for _, flag := range globalFlags {
		if arg == flag {
			return true
		}
	}
	return false
}
