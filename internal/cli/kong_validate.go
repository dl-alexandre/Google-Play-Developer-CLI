package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
)

// ValidateCmd is a high-level readiness report (ASC `validate` analogue).
// It composes local checks and an optional dry-run plan for publish steps.
// Network calls are skipped by default for safe CI preflight; pass --network
// to run an opt-in package access probe (requires --package and credentials).
type ValidateCmd struct {
	Track   string `help:"Target track for the readiness plan" default:"internal" enum:"internal,alpha,beta,production"`
	File    string `help:"Optional APK/AAB path to validate locally" type:"existingfile"`
	Strict  bool   `help:"Treat warnings as failures"`
	DryRun  bool   `help:"Plan checks without network side effects (default true)" default:"true"`
	Network bool   `help:"Opt-in network probes: package access, track list, listing (requires --package and credentials)"`
}

// validatePackageAccessProbe is the network probe used by validate --network.
// Defaults to probePackageAccess; tests may override.
var validatePackageAccessProbe = probePackageAccess

// validateTrackProbe lists tracks for a package (injectable for tests).
var validateTrackProbe = defaultValidateTrackProbe

// validateListingProbe checks default listing presence (injectable for tests).
var validateListingProbe = defaultValidateListingProbe

type readinessCheck struct {
	Name           string `json:"name"`
	Status         string `json:"status"` // pass, fail, warn, skip
	Message        string `json:"message"`
	Recommendation string `json:"recommendation,omitempty"`
}

// Run executes the validate command.
func (cmd *ValidateCmd) Run(globals *Globals) error {
	pkg := strings.TrimSpace(globals.Package)
	checks := make([]readinessCheck, 0, 8)

	// Package required for a real readiness report.
	if pkg == "" {
		checks = append(checks, readinessCheck{
			Name:    "package",
			Status:  "fail",
			Message: "package name is required (--package)",
		})
	} else {
		checks = append(checks, readinessCheck{
			Name:    "package",
			Status:  "pass",
			Message: fmt.Sprintf("package=%s", pkg),
		})
	}

	// Track validity
	if !api.IsValidTrack(cmd.Track) {
		checks = append(checks, readinessCheck{
			Name:    "track",
			Status:  "fail",
			Message: fmt.Sprintf("invalid track %q", cmd.Track),
		})
	} else {
		checks = append(checks, readinessCheck{
			Name:    "track",
			Status:  "pass",
			Message: fmt.Sprintf("track=%s", cmd.Track),
		})
	}

	// Optional artifact checks
	if strings.TrimSpace(cmd.File) != "" {
		checks = append(checks, validateArtifactFile(cmd.File)...)
	} else {
		checks = append(checks, readinessCheck{
			Name:    "artifact",
			Status:  "skip",
			Message: "no --file provided; skipped local artifact checks",
		})
	}

	// Profile / output contract notes (local)
	profile := strings.TrimSpace(globals.Profile)
	if profile == "" {
		profile = "default"
	}
	checks = append(checks, readinessCheck{
		Name:    "profile",
		Status:  "pass",
		Message: fmt.Sprintf("auth profile=%s", profile),
	})
	checks = append(checks, readinessCheck{
		Name:    "output",
		Status:  "pass",
		Message: fmt.Sprintf("output=%s", globals.Output),
	})

	// Opt-in network probes. Default remains local-safe for CI without credentials.
	checks = append(checks, cmd.runNetworkPackageAccessCheck(globals, pkg))
	checks = append(checks, cmd.runNetworkTrackCheck(globals, pkg))
	checks = append(checks, cmd.runNetworkListingCheck(globals, pkg))

	// Planned publish steps (informational; no side effects from the plan itself)
	plan := []string{
		fmt.Sprintf("publish upload --package %s --track %s", pkgOrPlaceholder(pkg), cmd.Track),
		fmt.Sprintf("publish status --package %s --track %s", pkgOrPlaceholder(pkg), cmd.Track),
	}

	passed, failed, warnings, skipped := 0, 0, 0, 0
	for _, c := range checks {
		switch c.Status {
		case "pass":
			passed++
		case "fail":
			failed++
		case "warn":
			warnings++
		default:
			skipped++
		}
	}

	status := "ready"
	if failed > 0 {
		status = "not_ready"
	} else if warnings > 0 && cmd.Strict {
		status = "not_ready"
	} else if warnings > 0 {
		status = "ready_with_warnings"
	}

	data := map[string]interface{}{
		"status":   status,
		"package":  pkg,
		"track":    cmd.Track,
		"dryRun":   cmd.DryRun,
		"network":  cmd.Network,
		"strict":   cmd.Strict,
		"checks":   checks,
		"passed":   passed,
		"failed":   failed,
		"warnings": warnings,
		"skipped":  skipped,
		"plan":     plan,
		"next": []string{
			"gpd auth check --package <pkg>",
			"gpd validate --package <pkg> --network",
			"gpd publish play <app.aab> --package <pkg> --track <track> --dry-run",
			"gpd publish play <app.aab> --package <pkg> --track <track>",
		},
	}

	result := output.NewResult(data).WithServices("validate")
	// No-op only when no network side effects occurred.
	if cmd.DryRun && !cmd.Network {
		result = result.WithNoOp("dry-run mode")
	}

	if err := outputResult(result, globals.Output, globals.Pretty); err != nil {
		return err
	}

	if status == "not_ready" {
		return errors.NewAPIError(errors.CodeValidationError, "readiness validation failed").
			WithHint("Fix failing checks in the validate report").
			WithDetails(map[string]interface{}{"failed": failed, "warnings": warnings})
	}
	return nil
}

// runNetworkPackageAccessCheck returns the network.package_access readiness check.
// When --network is not set, the check is skipped (local-safe default).
func (cmd *ValidateCmd) runNetworkPackageAccessCheck(globals *Globals, pkg string) readinessCheck {
	const name = "network.package_access"

	if !cmd.Network {
		return readinessCheck{
			Name:    name,
			Status:  "skip",
			Message: "network probes skipped (pass --network to enable; requires --package and credentials)",
		}
	}

	if pkg == "" {
		return readinessCheck{
			Name:           name,
			Status:         "fail",
			Message:        "network probe requires --package",
			Recommendation: "Pass --package com.example.app with --network",
		}
	}

	ctx := authContext(globals)
	if ctx == nil {
		ctx = context.Background()
	}

	if err := validatePackageAccessProbe(ctx, globals, pkg); err != nil {
		return readinessCheck{
			Name:   name,
			Status: "fail",
			Message: fmt.Sprintf(
				"package access probe failed for %s: %v",
				pkg, err,
			),
			Recommendation: "Configure credentials (`gpd auth login --key /path/to/sa.json` or ADC) and grant the service account Android Publisher access for this app in Play Console",
		}
	}

	return readinessCheck{
		Name:    name,
		Status:  "pass",
		Message: fmt.Sprintf("package access OK for %s (edits.insert+edits.delete)", pkg),
	}
}

// runNetworkTrackCheck verifies the target track is visible via the Publisher API.
func (cmd *ValidateCmd) runNetworkTrackCheck(globals *Globals, pkg string) readinessCheck {
	const name = "network.track"
	if !cmd.Network {
		return readinessCheck{
			Name:    name,
			Status:  "skip",
			Message: "network probes skipped (pass --network to enable)",
		}
	}
	if pkg == "" {
		return readinessCheck{
			Name:           name,
			Status:         "fail",
			Message:        "track probe requires --package",
			Recommendation: "Pass --package with --network",
		}
	}
	ctx := authContext(globals)
	if ctx == nil {
		ctx = context.Background()
	}
	tracks, err := validateTrackProbe(ctx, globals, pkg)
	if err != nil {
		return readinessCheck{
			Name:           name,
			Status:         "fail",
			Message:        fmt.Sprintf("track list failed for %s: %v", pkg, err),
			Recommendation: "Ensure credentials can list tracks (Android Publisher + app access)",
		}
	}
	want := cmd.Track
	found := false
	for _, t := range tracks {
		if t == want {
			found = true
			break
		}
	}
	if !found {
		return readinessCheck{
			Name:           name,
			Status:         "fail",
			Message:        fmt.Sprintf("track %q not found among %v", want, tracks),
			Recommendation: "Use a valid track (internal/alpha/beta/production) or create the track in Play Console",
		}
	}
	return readinessCheck{
		Name:    name,
		Status:  "pass",
		Message: fmt.Sprintf("track %q present (listed %d tracks)", want, len(tracks)),
	}
}

// runNetworkListingCheck verifies a default-locale listing can be read (warn if empty).
func (cmd *ValidateCmd) runNetworkListingCheck(globals *Globals, pkg string) readinessCheck {
	const name = "network.listing"
	if !cmd.Network {
		return readinessCheck{
			Name:    name,
			Status:  "skip",
			Message: "network probes skipped (pass --network to enable)",
		}
	}
	if pkg == "" {
		return readinessCheck{
			Name:           name,
			Status:         "fail",
			Message:        "listing probe requires --package",
			Recommendation: "Pass --package with --network",
		}
	}
	ctx := authContext(globals)
	if ctx == nil {
		ctx = context.Background()
	}
	title, err := validateListingProbe(ctx, globals, pkg)
	if err != nil {
		return readinessCheck{
			Name:           name,
			Status:         "warn",
			Message:        fmt.Sprintf("listing probe failed for %s: %v", pkg, err),
			Recommendation: "Ensure store listing exists for at least one locale, or ignore if app is not listed yet",
		}
	}
	if strings.TrimSpace(title) == "" {
		return readinessCheck{
			Name:           name,
			Status:         "warn",
			Message:        "listing readable but title empty",
			Recommendation: "Set listing title via gpd publish listing update",
		}
	}
	return readinessCheck{
		Name:    name,
		Status:  "pass",
		Message: fmt.Sprintf("listing OK (title=%q)", title),
	}
}

func defaultValidateTrackProbe(ctx context.Context, globals *Globals, pkg string) ([]string, error) {
	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return nil, err
	}
	svc, err := client.AndroidPublisher()
	if err != nil {
		return nil, err
	}
	// Disposable edit to list tracks (Play requires an edit for track list).
	edit, err := svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = svc.Edits.Delete(pkg, edit.Id).Context(ctx).Do()
	}()
	resp, err := svc.Edits.Tracks.List(pkg, edit.Id).Context(ctx).Do()
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(resp.Tracks))
	for _, t := range resp.Tracks {
		if t != nil && t.Track != "" {
			out = append(out, t.Track)
		}
	}
	return out, nil
}

func defaultValidateListingProbe(ctx context.Context, globals *Globals, pkg string) (string, error) {
	client, err := createAPIClient(ctx, globals)
	if err != nil {
		return "", err
	}
	svc, err := client.AndroidPublisher()
	if err != nil {
		return "", err
	}
	edit, err := svc.Edits.Insert(pkg, &androidpublisher.AppEdit{}).Context(ctx).Do()
	if err != nil {
		return "", err
	}
	defer func() {
		_ = svc.Edits.Delete(pkg, edit.Id).Context(ctx).Do()
	}()
	// Prefer en-US; fall back to first listing if present.
	listing, err := svc.Edits.Listings.Get(pkg, edit.Id, "en-US").Context(ctx).Do()
	if err == nil && listing != nil {
		return listing.Title, nil
	}
	list, lerr := svc.Edits.Listings.List(pkg, edit.Id).Context(ctx).Do()
	if lerr != nil {
		return "", err
	}
	if list == nil || len(list.Listings) == 0 {
		return "", fmt.Errorf("no store listings found")
	}
	return list.Listings[0].Title, nil
}

func pkgOrPlaceholder(pkg string) string {
	if pkg == "" {
		return "<package>"
	}
	return pkg
}

func validateArtifactFile(path string) []readinessCheck {
	out := make([]readinessCheck, 0, 3)
	info, err := os.Stat(path)
	if err != nil {
		return []readinessCheck{{
			Name:    "artifact.exists",
			Status:  "fail",
			Message: fmt.Sprintf("cannot stat file: %v", err),
		}}
	}
	if info.IsDir() {
		return []readinessCheck{{
			Name:    "artifact.exists",
			Status:  "fail",
			Message: "path is a directory, expected APK/AAB file",
		}}
	}
	out = append(out, readinessCheck{
		Name:    "artifact.exists",
		Status:  "pass",
		Message: fmt.Sprintf("file=%s size=%d", path, info.Size()),
	})

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".aab", ".apk":
		out = append(out, readinessCheck{
			Name:    "artifact.type",
			Status:  "pass",
			Message: fmt.Sprintf("type=%s", strings.TrimPrefix(ext, ".")),
		})
	default:
		out = append(out, readinessCheck{
			Name:    "artifact.type",
			Status:  "fail",
			Message: fmt.Sprintf("unsupported extension %q (want .aab or .apk)", ext),
		})
	}

	if info.Size() == 0 {
		out = append(out, readinessCheck{
			Name:    "artifact.size",
			Status:  "fail",
			Message: "file is empty",
		})
	} else {
		out = append(out, readinessCheck{
			Name:    "artifact.size",
			Status:  "pass",
			Message: fmt.Sprintf("%d bytes", info.Size()),
		})
	}
	return out
}
