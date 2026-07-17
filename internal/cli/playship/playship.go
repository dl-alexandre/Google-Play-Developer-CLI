// Package playship holds pure helpers for high-level Play publish planning
// (gpd publish play). Kong adapters live in package cli.
package playship

import (
	"fmt"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/api"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
)

// ResolveReleaseParams maps percentage + status into API release fields.
// When percentage > 0, status is forced to inProgress and userFraction is percentage/100.
// When percentage is 0, status is the requested status and userFraction is 0 (full track).
func ResolveReleaseParams(status string, percentage float64) (releaseStatus string, userFraction float64, err error) {
	if percentage < 0 || percentage > 100 {
		return "", 0, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid percentage: %g", percentage)).
			WithHint("Percentage must be between 0 and 100")
	}
	if percentage > 0 {
		if percentage < 0.01 {
			return "", 0, errors.NewAPIError(errors.CodeValidationError,
				fmt.Sprintf("invalid staged percentage: %g", percentage)).
				WithHint("Staged rollout percentage must be at least 0.01")
		}
		return string(api.StatusInProgress), percentage / 100.0, nil
	}
	if status == "" {
		status = string(api.StatusCompleted)
	}
	if !api.IsValidReleaseStatus(status) {
		return "", 0, errors.NewAPIError(errors.CodeValidationError,
			fmt.Sprintf("invalid release status: %s", status)).
			WithHint("Valid statuses are: draft, completed, halted, inProgress")
	}
	return status, 0, nil
}

// BuildTrackRelease builds the Tracks.Update payload for publish play.
func BuildTrackRelease(trackName, releaseStatus string, versionCodes []int64, userFraction float64) *androidpublisher.Track {
	release := &androidpublisher.TrackRelease{
		Status:       releaseStatus,
		VersionCodes: versionCodes,
	}
	if userFraction > 0 {
		release.UserFraction = userFraction
	}
	return &androidpublisher.Track{
		Track:    trackName,
		Releases: []*androidpublisher.TrackRelease{release},
	}
}

// BuildPlan returns the operator-visible plan for dry-run and docs.
func BuildPlan(pkg, file, track, status string, percentage, userFraction float64) map[string]interface{} {
	releaseStatus, frac, _ := ResolveReleaseParams(status, percentage)
	if releaseStatus == "" {
		releaseStatus = status
	}
	if percentage > 0 {
		userFraction = frac
	}

	steps := []map[string]interface{}{
		{
			"step":    1,
			"action":  "validate",
			"command": fmt.Sprintf("gpd validate --package %s --track %s --file %s --dry-run", pkg, track, file),
		},
		{
			"step":    2,
			"action":  "upload",
			"command": fmt.Sprintf("gpd publish upload %s --package %s --track %s --no-auto-commit", file, pkg, track),
		},
		{
			"step":                    3,
			"action":                  "release",
			"command":                 fmt.Sprintf("gpd publish release --package %s --track %s --status %s --version-codes <uploaded>", pkg, track, releaseStatus),
			"status":                  releaseStatus,
			"userFraction":            userFraction,
			"track":                   track,
			"assignsArtifactToTrack":  true,
		},
		{
			"step":    4,
			"action":  "status",
			"command": fmt.Sprintf("gpd publish status --package %s --track %s", pkg, track),
		},
	}

	return map[string]interface{}{
		"package":       pkg,
		"track":         track,
		"file":          file,
		"percentage":    percentage,
		"status":        status,
		"releaseStatus": releaseStatus,
		"userFraction":  userFraction,
		"steps":         steps,
	}
}
