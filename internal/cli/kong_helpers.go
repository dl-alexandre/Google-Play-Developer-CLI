package cli

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/auth"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/storage"
)

const (
	formatJSON     = "json"
	formatMarkdown = "markdown"
	formatTable    = "table"
	formatExcel    = "excel"
)

// Package-level auth defaults applied after globals are resolved in RunKongCLI.
// This keeps existing newAuthManager() call sites profile-aware without a
// wide refactor of every command package.
var (
	authDefaultsMu     sync.RWMutex
	authDefaultProfile string
	authDefaultStore   string
)

// applyAuthGlobals records profile and store-tokens defaults for newAuthManager.
func applyAuthGlobals(globals *Globals) {
	if globals == nil {
		return
	}
	authDefaultsMu.Lock()
	defer authDefaultsMu.Unlock()
	authDefaultProfile = strings.TrimSpace(globals.Profile)
	authDefaultStore = strings.TrimSpace(globals.StoreTokens)
}

// newAuthManager creates a new auth manager instance with resolved defaults.
func newAuthManager() *auth.Manager {
	secureStorage := storage.New()
	mgr := auth.NewManager(secureStorage)

	authDefaultsMu.RLock()
	profile := authDefaultProfile
	store := authDefaultStore
	authDefaultsMu.RUnlock()

	if store != "" {
		mgr.SetStoreTokens(store)
	}
	if profile != "" {
		mgr.SetActiveProfile(profile)
	}
	return mgr
}

// outputResult formats and outputs a result based on the format.
func outputResult(result *output.Result, format string, pretty bool) error {
	manager := output.NewManager(os.Stdout)
	manager.SetPretty(pretty)

	// Convert string format to output.Format type
	switch strings.ToLower(format) {
	case formatTable:
		manager.SetFormat(output.FormatTable)
	case formatExcel:
		manager.SetFormat(output.FormatExcel)
	default:
		manager.SetFormat(output.FormatJSON)
	}

	return manager.Write(result)
}

// requirePackage validates that a package name is provided.
func requirePackage(pkg string) error {
	if pkg == "" {
		return fmt.Errorf("package name is required")
	}
	return nil
}
