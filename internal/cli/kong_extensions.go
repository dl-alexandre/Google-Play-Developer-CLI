package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/extensions"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// ExtensionCmd manages CLI extensions.
type ExtensionCmd struct {
	Install ExtensionInstallCmd `cmd:"" help:"Install an extension"`
	List    ExtensionListCmd    `cmd:"" help:"List installed extensions"`
	Remove  ExtensionRemoveCmd  `cmd:"" help:"Remove an extension"`
	Upgrade ExtensionUpgradeCmd `cmd:"" help:"Upgrade an extension"`
	Exec    ExtensionExecCmd    `cmd:"" help:"Execute an extension explicitly"`
}

// ExtensionInstallCmd installs an extension from GitHub or local path.
type ExtensionInstallCmd struct {
	Source string `arg:"" help:"Extension source (GitHub: owner/repo or local: path)"`
	Pin    bool   `help:"Pin to current version (disable auto-upgrade)"`
	Force  bool   `help:"Overwrite existing extension"`
}

func (c *ExtensionInstallCmd) Run(globals *Globals) error {
	if c.Source == "" {
		return fmt.Errorf("extension source is required")
	}

	opts := extensions.InstallOptions{
		Source: c.Source,
		Pin:    c.Pin,
		Force:  c.Force,
	}

	result, err := extensions.Install(globals.Context, opts)
	if err != nil {
		return fmt.Errorf("installing extension: %w", err)
	}

	action := "Installed"
	if !result.Installed {
		action = "Updated"
	}

	if globals.Output == "json" {
		res := output.NewResult(map[string]interface{}{
			"action":      action,
			"name":        result.Extension.Name,
			"version":     result.Extension.Version,
			"description": result.Extension.Description,
			"source":      result.Extension.Source,
		})
		return outputResult(res, globals.Output, globals.Pretty)
	}

	fmt.Printf("%s extension %s@%s\n", action, result.Extension.Name, result.Extension.Version)
	if result.Extension.Description != "" {
		fmt.Printf("  %s\n", result.Extension.Description)
	}

	return nil
}

// ExtensionListCmd lists installed extensions.
type ExtensionListCmd struct {
	Format string `help:"Output format: table, json" default:"table" enum:"table,json"`
}

func (c *ExtensionListCmd) Run(globals *Globals) error {
	extList, err := extensions.List()
	if err != nil {
		return fmt.Errorf("listing extensions: %w", err)
	}

	if len(extList) == 0 {
		if globals.Output == "json" {
			res := output.NewResult([]extensions.Extension{})
			return outputResult(res, "json", false)
		}
		fmt.Println("No extensions installed.")
		fmt.Println("Install extensions with: gpd extension install owner/gpd-<name>")
		return nil
	}

	if globals.Output == "json" {
		res := output.NewResult(extList)
		return outputResult(res, globals.Output, globals.Pretty)
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tVERSION\tDESCRIPTION\tSOURCE")
	for _, ext := range extList {
		desc := ext.Description
		if len(desc) > 40 {
			desc = desc[:37] + "..."
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", ext.Name, ext.Version, desc, ext.Source)
	}
	_ = w.Flush()

	return nil
}

// ExtensionRemoveCmd removes an installed extension.
type ExtensionRemoveCmd struct {
	Name string `arg:"" help:"Extension name to remove"`
}

func (c *ExtensionRemoveCmd) Run(globals *Globals) error {
	if c.Name == "" {
		return fmt.Errorf("extension name is required")
	}

	if err := extensions.Remove(c.Name); err != nil {
		return err
	}

	if globals.Output == "json" {
		res := output.NewResult(map[string]string{
			"removed": c.Name,
		})
		return outputResult(res, "json", false)
	}

	fmt.Printf("Removed extension %s\n", c.Name)
	return nil
}

// ExtensionUpgradeCmd upgrades an installed extension.
type ExtensionUpgradeCmd struct {
	Name  string `arg:"" help:"Extension name to upgrade (or --all)"`
	All   bool   `help:"Upgrade all extensions"`
	Force bool   `help:"Force upgrade even if extension is pinned"`
	Pin   bool   `help:"Pin to current version after upgrade"`
}

type upgradeResult struct {
	Name       string `json:"name"`
	OldVersion string `json:"oldVersion"`
	NewVersion string `json:"newVersion"`
	Status     string `json:"status"`
	Error      string `json:"error,omitempty"`
}

func (c *ExtensionUpgradeCmd) Run(globals *Globals) error {
	if c.Name == "" && !c.All {
		return fmt.Errorf("extension name or --all flag is required")
	}

	var extensionsToUpgrade []string
	var warnings []string

	if c.All {
		// Get all installed extensions
		extList, err := extensions.List()
		if err != nil {
			return fmt.Errorf("listing extensions: %w", err)
		}
		if len(extList) == 0 {
			if globals.Output == "json" {
				res := output.NewResult([]upgradeResult{})
				return outputResult(res, globals.Output, globals.Pretty)
			}
			fmt.Println("No extensions installed.")
			return nil
		}
		for _, ext := range extList {
			extensionsToUpgrade = append(extensionsToUpgrade, ext.Name)
		}
	} else {
		// Single extension specified
		if !extensions.IsInstalled(c.Name) {
			return fmt.Errorf("extension %q is not installed", c.Name)
		}
		extensionsToUpgrade = []string{c.Name}
	}

	// Track upgrade results
	var results []upgradeResult

	for _, extName := range extensionsToUpgrade {
		// Load existing extension metadata
		extInfo, err := extensions.LoadExtension(extName)
		if err != nil {
			result := upgradeResult{
				Name:   extName,
				Status: "failed",
				Error:  fmt.Sprintf("loading extension metadata: %v", err),
			}
			results = append(results, result)
			if !c.All {
				return fmt.Errorf("loading extension %q metadata: %w", extName, err)
			}
			warnings = append(warnings, fmt.Sprintf("Failed to load %s: %v", extName, err))
			continue
		}

		// Check if pinned
		if extInfo.Pinned && !c.Force {
			result := upgradeResult{
				Name:       extName,
				OldVersion: extInfo.Version,
				NewVersion: extInfo.Version,
				Status:     "skipped",
				Error:      "pinned (use --force to upgrade)",
			}
			results = append(results, result)
			continue
		}

		// Store old version for comparison
		oldVersion := extInfo.Version

		// Perform upgrade using Install with Force=true
		installOpts := extensions.InstallOptions{
			Source:    extInfo.Source,
			Force:     true,
			Pin:       c.Pin,
			PinnedRef: extInfo.PinnedRef,
		}

		installResult, err := extensions.Install(globals.Context, installOpts)
		if err != nil {
			result := upgradeResult{
				Name:       extName,
				OldVersion: oldVersion,
				Status:     "failed",
				Error:      err.Error(),
			}
			results = append(results, result)
			if !c.All {
				return fmt.Errorf("upgrading extension %q: %w", extName, err)
			}
			warnings = append(warnings, fmt.Sprintf("Failed to upgrade %s: %v", extName, err))
			continue
		}

		// Determine status
		status := "upgraded"
		if installResult.Extension.Version == oldVersion {
			status = "up-to-date"
		}

		result := upgradeResult{
			Name:       extName,
			OldVersion: oldVersion,
			NewVersion: installResult.Extension.Version,
			Status:     status,
		}
		results = append(results, result)
	}

	// Output results
	if globals.Output == "json" {
		res := output.NewResult(results)
		if len(warnings) > 0 {
			res = res.WithWarnings(warnings...)
		}
		return outputResult(res, globals.Output, globals.Pretty)
	}

	// Table output
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tOLD VERSION\tNEW VERSION\tSTATUS")
	for _, result := range results {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", result.Name, result.OldVersion, result.NewVersion, result.Status)
	}
	_ = w.Flush()

	if len(warnings) > 0 {
		_, _ = fmt.Fprintln(os.Stderr, "\nWarnings:")
		for _, warning := range warnings {
			_, _ = fmt.Fprintf(os.Stderr, "  - %s\n", warning)
		}
	}

	return nil
}

// ExtensionExecCmd executes an extension explicitly.
type ExtensionExecCmd struct {
	Name string   `arg:"" help:"Extension name to execute"`
	Args []string `arg:"" optional:"" help:"Arguments to pass to extension"`
}

func (c *ExtensionExecCmd) Run(globals *Globals) error {
	if c.Name == "" {
		return fmt.Errorf("extension name is required")
	}

	// Check if extension exists
	if !extensions.IsInstalled(c.Name) {
		return fmt.Errorf("extension %q is not installed", c.Name)
	}

	// Execute the extension with forwarded arguments
	// This will replace the current process on Unix or run as subprocess
	args := append([]string{c.Name}, c.Args...)
	if tryRunExtension(args) {
		// Extension was executed - this should not return as the process exits
		return nil
	}

	return fmt.Errorf("failed to execute extension %q", c.Name)
}
