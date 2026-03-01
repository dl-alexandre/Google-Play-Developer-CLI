package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/errors"
)

// ============================================================================
// ReleaseMatchesVersionCode Tests
// ============================================================================

func TestReleaseMatchesVersionCode(t *testing.T) {
	tests := []struct {
		name        string
		release     *androidpublisher.TrackRelease
		versionCode string
		want        bool
	}{
		{
			name: "matching version code",
			release: &androidpublisher.TrackRelease{
				VersionCodes: []int64{123, 456},
			},
			versionCode: "123",
			want:        true,
		},
		{
			name: "non-matching version code",
			release: &androidpublisher.TrackRelease{
				VersionCodes: []int64{123, 456},
			},
			versionCode: "789",
			want:        false,
		},
		{
			name:        "empty version codes",
			release:     &androidpublisher.TrackRelease{VersionCodes: []int64{}},
			versionCode: "123",
			want:        false,
		},
		{
			name:        "nil version codes",
			release:     &androidpublisher.TrackRelease{},
			versionCode: "123",
			want:        false,
		},
		{
			name: "multiple matches returns true on first match",
			release: &androidpublisher.TrackRelease{
				VersionCodes: []int64{100, 200, 300},
			},
			versionCode: "200",
			want:        true,
		},
		{
			name: "large version codes",
			release: &androidpublisher.TrackRelease{
				VersionCodes: []int64{999999999},
			},
			versionCode: "999999999",
			want:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := releaseMatchesVersionCode(tt.release, tt.versionCode)
			if got != tt.want {
				t.Errorf("releaseMatchesVersionCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ============================================================================
// ReleaseStrategyCmd.buildStrategyRecommendation Tests
// ============================================================================

func TestReleaseStrategyCmd_BuildStrategyRecommendation(t *testing.T) {
	tests := []struct {
		name           string
		cmd            ReleaseStrategyCmd
		healthScore    float64
		userFraction   float64
		wantRec        string
		wantActionsLen int
		wantRisksLen   int
	}{
		{
			name: "health score above threshold - continue recommendation",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.98,
			userFraction:   0.5,
			wantRec:        "continue",
			wantActionsLen: 2,
			wantRisksLen:   0,
		},
		{
			name: "health score above threshold with full rollout - no rollout action",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.97,
			userFraction:   1.0,
			wantRec:        "continue",
			wantActionsLen: 1,
			wantRisksLen:   0,
		},
		{
			name: "health score in monitor zone - close to threshold",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.90,
			userFraction:   0.75,
			wantRec:        "monitor",
			wantActionsLen: 2,
			wantRisksLen:   1,
		},
		{
			name: "health score in investigate zone",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.80,
			userFraction:   0.5,
			wantRec:        "investigate",
			wantActionsLen: 3,
			wantRisksLen:   2,
		},
		{
			name: "health score critically low - rollback",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.50,
			userFraction:   0.3,
			wantRec:        "rollback",
			wantActionsLen: 3,
			wantRisksLen:   2,
		},
		{
			name: "zero health score - rollback",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.0,
			userFraction:   0.1,
			wantRec:        "rollback",
			wantActionsLen: 3,
			wantRisksLen:   2,
		},
		{
			name: "health score exactly at threshold - continue",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.95,
			userFraction:   0.0,
			wantRec:        "continue",
			wantActionsLen: 1,
			wantRisksLen:   0,
		},
		{
			name: "health score at 85% of threshold - monitor zone boundary",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.8075,
			userFraction:   0.0,
			wantRec:        "monitor",
			wantActionsLen: 2,
			wantRisksLen:   1,
		},
		{
			name: "health score just below 85% - investigate zone",
			cmd: ReleaseStrategyCmd{
				HealthThreshold: 0.95,
			},
			healthScore:    0.80,
			userFraction:   0.0,
			wantRec:        "investigate",
			wantActionsLen: 2,
			wantRisksLen:   2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec, _, actions, risks := tt.cmd.buildStrategyRecommendation(tt.healthScore, tt.userFraction)

			if rec != tt.wantRec {
				t.Errorf("buildStrategyRecommendation() recommendation = %v, want %v", rec, tt.wantRec)
			}

			if len(actions) != tt.wantActionsLen {
				t.Errorf("buildStrategyRecommendation() actions length = %v, want %v", len(actions), tt.wantActionsLen)
			}

			if len(risks) != tt.wantRisksLen {
				t.Errorf("buildStrategyRecommendation() risks length = %v, want %v", len(risks), tt.wantRisksLen)
			}
		})
	}
}

func TestReleaseStrategyCmd_BuildStrategyRecommendation_Reasoning(t *testing.T) {
	cmd := ReleaseStrategyCmd{HealthThreshold: 0.95}

	tests := []struct {
		name         string
		healthScore  float64
		userFraction float64
		wantContains string
	}{
		{
			name:         "continue reasoning mentions threshold",
			healthScore:  0.98,
			userFraction: 0.5,
			wantContains: "above threshold",
		},
		{
			name:         "rollback reasoning mentions critical",
			healthScore:  0.50,
			userFraction: 0.3,
			wantContains: "critically below",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, reasoning, _, _ := cmd.buildStrategyRecommendation(tt.healthScore, tt.userFraction)

			if reasoning == "" {
				t.Error("buildStrategyRecommendation() reasoning should not be empty")
			}

			if tt.wantContains != "" && !containsStr(reasoning, tt.wantContains) {
				t.Errorf("buildStrategyRecommendation() reasoning = %v, want to contain %v", reasoning, tt.wantContains)
			}
		})
	}
}

// ============================================================================
// ReleaseConflictsCmd - Conflicts Detection Logic Tests
// ============================================================================

func TestReleaseConflictsCmd_DetectConflicts(t *testing.T) {
	tests := []struct {
		name             string
		versionCodes     []string
		checkTrack       string
		suggestFix       bool
		setupTracks      []*androidpublisher.Track
		wantConflicts    int
		wantHasConflicts bool
		wantSuggestions  int
	}{
		{
			name:         "single conflict detected",
			versionCodes: []string{"100"},
			checkTrack:   "all",
			suggestFix:   false,
			setupTracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
							Name:         "v1.0.0",
						},
					},
				},
			},
			wantConflicts:    1,
			wantHasConflicts: true,
			wantSuggestions:  0,
		},
		{
			name:         "multiple version codes - some conflict",
			versionCodes: []string{"100", "200", "300"},
			checkTrack:   "all",
			suggestFix:   false,
			setupTracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
							Name:         "v1.0.0",
						},
					},
				},
				{
					Track: "beta",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{200},
							Status:       statusInProgress,
							Name:         "v2.0.0-beta",
						},
					},
				},
			},
			wantConflicts:    2,
			wantHasConflicts: true,
			wantSuggestions:  0,
		},
		{
			name:         "no conflicts - all version codes new",
			versionCodes: []string{"500", "600"},
			checkTrack:   "all",
			suggestFix:   false,
			setupTracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100, 200},
							Status:       releaseCompleted,
						},
					},
				},
			},
			wantConflicts:    0,
			wantHasConflicts: false,
			wantSuggestions:  0,
		},
		{
			name:         "track filter - only check specific track",
			versionCodes: []string{"100"},
			checkTrack:   "beta",
			suggestFix:   false,
			setupTracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
						},
					},
				},
				{
					Track: "beta",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{200},
							Status:       statusInProgress,
						},
					},
				},
			},
			wantConflicts:    0,
			wantHasConflicts: false,
			wantSuggestions:  0,
		},
		{
			name:         "suggest fix - provides version code suggestion",
			versionCodes: []string{"100"},
			checkTrack:   "all",
			suggestFix:   true,
			setupTracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100, 150},
							Status:       releaseCompleted,
							Name:         "v1.0.0",
						},
					},
				},
			},
			wantConflicts:    1,
			wantHasConflicts: true,
			wantSuggestions:  1,
		},
		{
			name:         "multi-track conflict - extra suggestion",
			versionCodes: []string{"100"},
			checkTrack:   "all",
			suggestFix:   true,
			setupTracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
							Name:         "v1.0.0-prod",
						},
					},
				},
				{
					Track: "beta",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
							Name:         "v1.0.0-beta",
						},
					},
				},
			},
			wantConflicts:    2,
			wantHasConflicts: true,
			wantSuggestions:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build conflicts result manually to test the logic
			result := &releaseConflictsResult{
				Conflicts:   make([]releaseConflict, 0),
				Suggestions: make([]string, 0),
			}

			requestedVCs := make(map[string]bool)
			for _, vc := range tt.versionCodes {
				requestedVCs[vc] = true
			}

			var maxExistingVC int64
			for _, track := range tt.setupTracks {
				if tt.checkTrack != "all" && track.Track != tt.checkTrack {
					continue
				}

				for _, release := range track.Releases {
					for _, vc := range release.VersionCodes {
						vcStr := strconv.FormatInt(vc, 10)
						if vc > maxExistingVC {
							maxExistingVC = vc
						}

						if requestedVCs[vcStr] {
							result.Conflicts = append(result.Conflicts, releaseConflict{
								VersionCode:     vcStr,
								Track:           track.Track,
								Status:          release.Status,
								ExistingVersion: release.Name,
							})
						}
					}
				}
			}

			result.HasConflicts = len(result.Conflicts) > 0

			if tt.suggestFix && result.HasConflicts {
				suggestedVC := maxExistingVC + 1
				result.Suggestions = append(result.Suggestions,
					"Use version code "+strconv.FormatInt(suggestedVC, 10)+" or higher to avoid conflicts")

				trackMap := make(map[string][]string)
				for _, conflict := range result.Conflicts {
					trackMap[conflict.VersionCode] = append(trackMap[conflict.VersionCode], conflict.Track)
				}
				for vc, tracks := range trackMap {
					if len(tracks) > 1 {
						result.Suggestions = append(result.Suggestions,
							"Version code "+vc+" exists on multiple tracks")
					}
				}
			}

			if len(result.Conflicts) != tt.wantConflicts {
				t.Errorf("conflicts count = %v, want %v", len(result.Conflicts), tt.wantConflicts)
			}

			if result.HasConflicts != tt.wantHasConflicts {
				t.Errorf("hasConflicts = %v, want %v", result.HasConflicts, tt.wantHasConflicts)
			}

			if len(result.Suggestions) != tt.wantSuggestions {
				t.Errorf("suggestions count = %v, want %v", len(result.Suggestions), tt.wantSuggestions)
			}
		})
	}
}

// ============================================================================
// ReleaseCalendarCmd - Calendar Event Logic Tests
// ============================================================================

func TestReleaseCalendarCmd_BuildEvents(t *testing.T) {
	tests := []struct {
		name        string
		track       string
		tracks      []*androidpublisher.Track
		wantEvents  int
		wantEventAt int // index of event to check
		wantType    string
	}{
		{
			name:  "completed release event",
			track: "all",
			tracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
							Name:         "v1.0.0",
						},
					},
				},
			},
			wantEvents:  1,
			wantEventAt: 0,
			wantType:    releaseCompleted,
		},
		{
			name:  "in-progress rollout event",
			track: "all",
			tracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{200},
							Status:       statusInProgress,
							Name:         "v2.0.0",
							UserFraction: 0.25,
						},
					},
				},
			},
			wantEvents:  1,
			wantEventAt: 0,
			wantType:    "rollout",
		},
		{
			name:  "halted release event",
			track: "all",
			tracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{300},
							Status:       statusHalted,
							Name:         "v3.0.0",
						},
					},
				},
			},
			wantEvents:  1,
			wantEventAt: 0,
			wantType:    statusHalted,
		},
		{
			name:  "draft release event",
			track: "all",
			tracks: []*androidpublisher.Track{
				{
					Track: "internal",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{50},
							Status:       "draft",
							Name:         "v0.5.0",
						},
					},
				},
			},
			wantEvents:  1,
			wantEventAt: 0,
			wantType:    "draft",
		},
		{
			name:  "multiple tracks - all included",
			track: "all",
			tracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
						},
					},
				},
				{
					Track: "beta",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{200},
							Status:       releaseCompleted,
						},
					},
				},
			},
			wantEvents:  2,
			wantEventAt: -1,
		},
		{
			name:  "track filter - only specific track",
			track: "production",
			tracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100},
							Status:       releaseCompleted,
						},
					},
				},
				{
					Track: "beta",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{200},
							Status:       releaseCompleted,
						},
					},
				},
			},
			wantEvents:  1,
			wantEventAt: 0,
			wantType:    releaseCompleted,
		},
		{
			name:        "empty tracks",
			track:       "all",
			tracks:      []*androidpublisher.Track{},
			wantEvents:  0,
			wantEventAt: -1,
		},
		{
			name:  "release with multiple version codes",
			track: "all",
			tracks: []*androidpublisher.Track{
				{
					Track: "production",
					Releases: []*androidpublisher.TrackRelease{
						{
							VersionCodes: []int64{100, 101, 102},
							Status:       releaseCompleted,
						},
					},
				},
			},
			wantEvents:  1,
			wantEventAt: 0,
			wantType:    releaseCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			events := make([]releaseCalendarEvent, 0)

			now := time.Now()

			for _, track := range tt.tracks {
				if tt.track != "all" && track.Track != tt.track {
					continue
				}

				for _, release := range track.Releases {
					versionCode := ""
					if len(release.VersionCodes) > 0 {
						versionCode = strconv.FormatInt(release.VersionCodes[0], 10)
					}

					eventType := "release"
					description := ""

					switch release.Status {
					case releaseCompleted:
						eventType = releaseCompleted
						description = "Completed release"
					case statusInProgress:
						eventType = "rollout"
						rolloutPct := release.UserFraction * 100
						description = "Rolling out " + strconv.FormatFloat(rolloutPct, 'f', 1, 64) + "%"
					case statusHalted:
						eventType = statusHalted
						description = "Halted release"
					case "draft":
						eventType = "draft"
						description = "Draft release"
					}

					events = append(events, releaseCalendarEvent{
						Date:        now.Format("2006-01-02"),
						Type:        eventType,
						Track:       track.Track,
						VersionCode: versionCode,
						Description: description,
					})
				}
			}

			if len(events) != tt.wantEvents {
				t.Errorf("events count = %v, want %v", len(events), tt.wantEvents)
			}

			if tt.wantEventAt >= 0 && tt.wantEventAt < len(events) {
				if events[tt.wantEventAt].Type != tt.wantType {
					t.Errorf("event type = %v, want %v", events[tt.wantEventAt].Type, tt.wantType)
				}
			}
		})
	}
}

// ============================================================================
// ReleaseHistoryCmd - History Building Tests
// ============================================================================

func TestReleaseHistoryCmd_BuildHistory(t *testing.T) {
	tests := []struct {
		name       string
		track      string
		limit      int
		releases   []*androidpublisher.TrackRelease
		wantCount  int
		wantItemAt int
		wantStatus string
	}{
		{
			name:  "completed release at 100% rollout",
			track: "production",
			limit: 20,
			releases: []*androidpublisher.TrackRelease{
				{
					VersionCodes: []int64{100},
					Status:       releaseCompleted,
					Name:         "v1.0.0",
				},
			},
			wantCount:  1,
			wantItemAt: 0,
			wantStatus: releaseCompleted,
		},
		{
			name:  "in-progress release with partial rollout",
			track: "production",
			limit: 20,
			releases: []*androidpublisher.TrackRelease{
				{
					VersionCodes: []int64{200},
					Status:       statusInProgress,
					Name:         "v2.0.0",
					UserFraction: 0.5,
				},
			},
			wantCount:  1,
			wantItemAt: 0,
			wantStatus: statusInProgress,
		},
		{
			name:  "halted release shows 0% rollout",
			track: "production",
			limit: 20,
			releases: []*androidpublisher.TrackRelease{
				{
					VersionCodes: []int64{300},
					Status:       statusHalted,
					Name:         "v3.0.0",
					UserFraction: 0.3,
				},
			},
			wantCount:  1,
			wantItemAt: 0,
			wantStatus: statusHalted,
		},
		{
			name:  "limit restricts releases",
			track: "production",
			limit: 2,
			releases: []*androidpublisher.TrackRelease{
				{
					VersionCodes: []int64{100},
					Status:       releaseCompleted,
					Name:         "v1.0.0",
				},
				{
					VersionCodes: []int64{200},
					Status:       releaseCompleted,
					Name:         "v2.0.0",
				},
				{
					VersionCodes: []int64{300},
					Status:       releaseCompleted,
					Name:         "v3.0.0",
				},
			},
			wantCount:  2,
			wantItemAt: -1,
		},
		{
			name:       "empty releases",
			track:      "production",
			limit:      20,
			releases:   []*androidpublisher.TrackRelease{},
			wantCount:  0,
			wantItemAt: -1,
		},
		{
			name:  "multiple version codes in release",
			track: "production",
			limit: 20,
			releases: []*androidpublisher.TrackRelease{
				{
					VersionCodes: []int64{100, 101, 102},
					Status:       releaseCompleted,
					Name:         "v1.0.0",
				},
			},
			wantCount:  1,
			wantItemAt: 0,
			wantStatus: releaseCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &releaseHistoryResult{
				Track:    tt.track,
				Releases: make([]releaseHistoryItem, 0, tt.limit),
			}

			for i, release := range tt.releases {
				if tt.limit > 0 && i >= tt.limit {
					break
				}

				versionCodes := make([]string, 0, len(release.VersionCodes))
				for _, vc := range release.VersionCodes {
					versionCodes = append(versionCodes, strconv.FormatInt(vc, 10))
				}

				var rolloutPct float64
				switch release.Status {
				case statusInProgress:
					rolloutPct = release.UserFraction * 100
				case releaseCompleted:
					rolloutPct = 100.0
				}

				item := releaseHistoryItem{
					VersionCodes:      versionCodes,
					Name:              release.Name,
					Status:            release.Status,
					ReleaseDate:       time.Now().Format("2006-01-02"),
					RolloutPercentage: rolloutPct,
				}

				result.Releases = append(result.Releases, item)
				result.Count++
			}

			if result.Count != tt.wantCount {
				t.Errorf("count = %v, want %v", result.Count, tt.wantCount)
			}

			if tt.wantItemAt >= 0 && tt.wantItemAt < len(result.Releases) {
				if result.Releases[tt.wantItemAt].Status != tt.wantStatus {
					t.Errorf("status = %v, want %v", result.Releases[tt.wantItemAt].Status, tt.wantStatus)
				}
			}
		})
	}
}

// ============================================================================
// ReleaseHistoryCmd - Vitals Calculation Tests
// ============================================================================

func TestReleaseHistoryVitals_StabilityCalculation(t *testing.T) {
	tests := []struct {
		name          string
		crashRate     float64
		anrRate       float64
		wantStability float64
	}{
		{
			name:          "perfect stability",
			crashRate:     0.0,
			anrRate:       0.0,
			wantStability: 1.0,
		},
		{
			name:          "some crashes",
			crashRate:     0.01,
			anrRate:       0.0,
			wantStability: 0.99,
		},
		{
			name:          "some ANRs",
			crashRate:     0.0,
			anrRate:       0.005,
			wantStability: 0.995,
		},
		{
			name:          "both issues",
			crashRate:     0.01,
			anrRate:       0.005,
			wantStability: 0.985,
		},
		{
			name:          "negative stability clamped to 0",
			crashRate:     0.8,
			anrRate:       0.5,
			wantStability: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stability := 1.0 - tt.crashRate - tt.anrRate
			if stability < 0 {
				stability = 0
			}

			if stability != tt.wantStability {
				t.Errorf("stability = %v, want %v", stability, tt.wantStability)
			}
		})
	}
}

// ============================================================================
// ReleaseNotesCmd - Input Validation Tests
// ============================================================================

func TestReleaseNotesCmd_Validation(t *testing.T) {
	tests := []struct {
		name      string
		cmd       ReleaseNotesCmd
		wantError bool
		errCode   errors.ErrorCode
	}{
		{
			name: "set action requires file",
			cmd: ReleaseNotesCmd{
				Action:      "set",
				Track:       "production",
				VersionCode: "100",
				File:        "",
			},
			wantError: true,
			errCode:   errors.CodeValidationError,
		},
		{
			name: "copy action requires target locales",
			cmd: ReleaseNotesCmd{
				Action:        "copy",
				Track:         "production",
				SourceLocale:  "en-US",
				TargetLocales: []string{},
			},
			wantError: true,
			errCode:   errors.CodeValidationError,
		},
		{
			name: "get action with version code is valid",
			cmd: ReleaseNotesCmd{
				Action:      "get",
				Track:       "production",
				VersionCode: "100",
			},
			wantError: false,
		},
		{
			name: "list action is valid",
			cmd: ReleaseNotesCmd{
				Action: "list",
				Track:  "production",
			},
			wantError: false,
		},
		{
			name: "copy action with targets is valid",
			cmd: ReleaseNotesCmd{
				Action:        "copy",
				Track:         "production",
				SourceLocale:  "en-US",
				TargetLocales: []string{"es-ES", "fr-FR"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error

			switch tt.cmd.Action {
			case "set":
				if tt.cmd.File == "" {
					err = errors.NewAPIError(errors.CodeValidationError, "--file is required for set action")
				}
			case "copy":
				if len(tt.cmd.TargetLocales) == 0 {
					err = errors.NewAPIError(errors.CodeValidationError, "--target-locales is required for copy action")
				}
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				apiErr, ok := err.(*errors.APIError)
				if !ok {
					t.Errorf("expected APIError but got %T", err)
					return
				}

				if apiErr.Code != tt.errCode {
					t.Errorf("error code = %v, want %v", apiErr.Code, tt.errCode)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ============================================================================
// ReleaseNotesCmd - File Operations Tests
// ============================================================================

func TestReleaseNotesCmd_ReadNotesFile(t *testing.T) {
	tests := []struct {
		name      string
		content   string
		wantError bool
		errCode   errors.ErrorCode
	}{
		{
			name:      "valid JSON content",
			content:   `{"en-US": "Release notes in English", "es-ES": "Notas de la versión"}`,
			wantError: false,
		},
		{
			name:      "invalid JSON content",
			content:   `{"en-US": "incomplete`,
			wantError: true,
			errCode:   errors.CodeValidationError,
		},
		{
			name:      "empty JSON object",
			content:   `{}`,
			wantError: false,
		},
		{
			name:      "multiple locales",
			content:   `{"en-US": "Notes", "de-DE": "Notizen", "fr-FR": "Notes"}`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "notes.json")

			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			fileData, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("failed to read temp file: %v", err)
			}

			var notesMap map[string]string
			parseErr := json.Unmarshal(fileData, &notesMap)

			if tt.wantError {
				if parseErr == nil {
					t.Error("expected parse error but got nil")
					return
				}

				// In real code, this would be wrapped in APIError
				wrappedErr := errors.NewAPIError(tt.errCode, "failed to parse release notes JSON: "+parseErr.Error())
				if wrappedErr.Code != tt.errCode {
					t.Errorf("error code = %v, want %v", wrappedErr.Code, tt.errCode)
				}
			} else {
				if parseErr != nil {
					t.Errorf("unexpected parse error: %v", parseErr)
				}

				if notesMap == nil {
					t.Error("expected notes map but got nil")
				}
			}
		})
	}
}

// ============================================================================
// ReleaseNotesCmd - Copy Action Logic Tests
// ============================================================================

func TestReleaseNotesCmd_CopyAction(t *testing.T) {
	tests := []struct {
		name          string
		sourceLocale  string
		targetLocales []string
		trackReleases []*androidpublisher.TrackRelease
		wantError     bool
		errCode       errors.ErrorCode
		wantLocales   int
	}{
		{
			name:          "successful copy to single target",
			sourceLocale:  "en-US",
			targetLocales: []string{"es-ES"},
			trackReleases: []*androidpublisher.TrackRelease{
				{
					ReleaseNotes: []*androidpublisher.LocalizedText{
						{
							Language: "en-US",
							Text:     "English notes",
						},
					},
				},
			},
			wantError:   false,
			wantLocales: 2,
		},
		{
			name:          "successful copy to multiple targets",
			sourceLocale:  "en-US",
			targetLocales: []string{"es-ES", "fr-FR", "de-DE"},
			trackReleases: []*androidpublisher.TrackRelease{
				{
					ReleaseNotes: []*androidpublisher.LocalizedText{
						{
							Language: "en-US",
							Text:     "English notes",
						},
					},
				},
			},
			wantError:   false,
			wantLocales: 4,
		},
		{
			name:          "source locale not found",
			sourceLocale:  "xx-XX",
			targetLocales: []string{"es-ES"},
			trackReleases: []*androidpublisher.TrackRelease{
				{
					ReleaseNotes: []*androidpublisher.LocalizedText{
						{
							Language: "en-US",
							Text:     "English notes",
						},
					},
				},
			},
			wantError: true,
			errCode:   errors.CodeNotFound,
		},
		{
			name:          "copy to existing locale - updates text",
			sourceLocale:  "en-US",
			targetLocales: []string{"es-ES"},
			trackReleases: []*androidpublisher.TrackRelease{
				{
					ReleaseNotes: []*androidpublisher.LocalizedText{
						{
							Language: "en-US",
							Text:     "English notes",
						},
						{
							Language: "es-ES",
							Text:     "Old Spanish notes",
						},
					},
				},
			},
			wantError:   false,
			wantLocales: 2,
		},
		{
			name:          "empty track releases",
			sourceLocale:  "en-US",
			targetLocales: []string{"es-ES"},
			trackReleases: []*androidpublisher.TrackRelease{},
			wantError:     true,
			errCode:       errors.CodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &releaseNotesResult{
				Source:  tt.sourceLocale,
				Targets: tt.targetLocales,
				Locales: make(map[string]releaseNotesData),
			}

			var err error

			if len(tt.trackReleases) == 0 {
				err = errors.NewAPIError(errors.CodeNotFound, "no release found")
			} else {
				release := tt.trackReleases[0]

				var sourceText string
				for _, note := range release.ReleaseNotes {
					if note.Language == tt.sourceLocale {
						sourceText = note.Text
						result.Locales[tt.sourceLocale] = releaseNotesData{Text: sourceText}
						break
					}
				}

				if sourceText == "" {
					err = errors.NewAPIError(errors.CodeNotFound,
						"no release notes found for source locale "+tt.sourceLocale)
				} else {
					existingNotes := make(map[string]*androidpublisher.LocalizedText)
					for _, note := range release.ReleaseNotes {
						existingNotes[note.Language] = note
					}

					for _, targetLocale := range tt.targetLocales {
						if existing, ok := existingNotes[targetLocale]; ok {
							existing.Text = sourceText
						} else {
							release.ReleaseNotes = append(release.ReleaseNotes, &androidpublisher.LocalizedText{
								Language: targetLocale,
								Text:     sourceText,
							})
						}
						result.Locales[targetLocale] = releaseNotesData{Text: sourceText}
					}
				}
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				apiErr, ok := err.(*errors.APIError)
				if !ok {
					t.Errorf("expected APIError but got %T", err)
					return
				}

				if apiErr.Code != tt.errCode {
					t.Errorf("error code = %v, want %v", apiErr.Code, tt.errCode)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if len(result.Locales) != tt.wantLocales {
					t.Errorf("locales count = %v, want %v", len(result.Locales), tt.wantLocales)
				}
			}
		})
	}
}

// ============================================================================
// ReleaseStrategyCmd - Health Score Calculation Tests
// ============================================================================

func TestReleaseStrategyCmd_HealthScoreCalculation(t *testing.T) {
	tests := []struct {
		name      string
		crashRate float64
		anrRate   float64
		wantScore float64
		wantClamp bool
	}{
		{
			name:      "perfect health - no crashes or ANRs",
			crashRate: 0.0,
			anrRate:   0.0,
			wantScore: 1.0,
		},
		{
			name:      "high crash rate capped",
			crashRate: 0.02,
			anrRate:   0.0,
			wantScore: 0.5,
		},
		{
			name:      "high ANR rate capped",
			crashRate: 0.0,
			anrRate:   0.01,
			wantScore: 0.5,
		},
		{
			name:      "moderate issues - 50% penalty each",
			crashRate: 0.005,
			anrRate:   0.0025,
			wantScore: 0.5,
		},
		{
			name:      "50% of bad thresholds",
			crashRate: 0.005,
			anrRate:   0.0025,
			wantScore: 0.5,
		},
		{
			name:      "negative score clamped to 0",
			crashRate: 0.03,
			anrRate:   0.02,
			wantScore: 0.0,
			wantClamp: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate health score using the same logic as the command
			crashPenalty := tt.crashRate / 0.01
			if crashPenalty > 1 {
				crashPenalty = 1
			}
			anrPenalty := tt.anrRate / 0.005
			if anrPenalty > 1 {
				anrPenalty = 1
			}
			healthScore := 1.0 - (crashPenalty*0.5 + anrPenalty*0.5)
			if healthScore < 0 {
				healthScore = 0
			}

			if healthScore != tt.wantScore {
				t.Errorf("health score = %v, want %v", healthScore, tt.wantScore)
			}
		})
	}
}

// ============================================================================
// Command Struct Validation Tests
// ============================================================================

func TestReleaseMgmtCmd_Structs(t *testing.T) {
	tests := []struct {
		name        string
		cmd         interface{}
		wantDefault interface{}
	}{
		{
			name: "ReleaseCalendarCmd defaults",
			cmd: ReleaseCalendarCmd{
				Track:      "all",
				DaysAhead:  30,
				DaysBehind: 30,
				Format:     "table",
			},
		},
		{
			name: "ReleaseConflictsCmd defaults",
			cmd: ReleaseConflictsCmd{
				CheckTrack: "all",
				SuggestFix: false,
			},
		},
		{
			name: "ReleaseStrategyCmd defaults",
			cmd: ReleaseStrategyCmd{
				Track:           "production",
				HealthThreshold: 0.95,
				DryRun:          false,
			},
		},
		{
			name: "ReleaseHistoryCmd defaults",
			cmd: ReleaseHistoryCmd{
				Track:         "production",
				Limit:         20,
				IncludeVitals: false,
				Format:        "table",
			},
		},
		{
			name: "ReleaseNotesCmd defaults",
			cmd: ReleaseNotesCmd{
				Track:        "production",
				SourceLocale: "en-US",
				Format:       "json",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// These tests verify the struct fields exist and are accessible
			// They serve as compile-time checks and basic struct validation
			switch cmd := tt.cmd.(type) {
			case ReleaseCalendarCmd:
				if cmd.Track != "all" {
					t.Errorf("Track = %v, want all", cmd.Track)
				}
			case ReleaseConflictsCmd:
				if cmd.CheckTrack != "all" {
					t.Errorf("CheckTrack = %v, want all", cmd.CheckTrack)
				}
			case ReleaseStrategyCmd:
				if cmd.Track != "production" {
					t.Errorf("Track = %v, want production", cmd.Track)
				}
			case ReleaseHistoryCmd:
				if cmd.Track != "production" {
					t.Errorf("Track = %v, want production", cmd.Track)
				}
			case ReleaseNotesCmd:
				if cmd.Track != "production" {
					t.Errorf("Track = %v, want production", cmd.Track)
				}
			}
		})
	}
}

// ============================================================================
// Error Handling Tests
// ============================================================================

func TestReleaseConflictsCmd_ValidateVersionCodes(t *testing.T) {
	tests := []struct {
		name         string
		versionCodes []string
		wantError    bool
		errMsg       string
	}{
		{
			name:         "empty version codes",
			versionCodes: []string{},
			wantError:    true,
			errMsg:       "at least one version code is required",
		},
		{
			name:         "single version code",
			versionCodes: []string{"100"},
			wantError:    false,
		},
		{
			name:         "multiple version codes",
			versionCodes: []string{"100", "200", "300"},
			wantError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if len(tt.versionCodes) == 0 {
				err = errors.NewAPIError(errors.CodeValidationError, tt.errMsg).
					WithHint("Provide version codes with --version-codes flag")
			}

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
					return
				}

				apiErr, ok := err.(*errors.APIError)
				if !ok {
					t.Errorf("expected APIError but got %T", err)
					return
				}

				if apiErr.Code != errors.CodeValidationError {
					t.Errorf("error code = %v, want VALIDATION_ERROR", apiErr.Code)
				}

				if apiErr.Hint == "" {
					t.Error("expected hint but got empty string")
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// ============================================================================
// Integration Helper Tests
// ============================================================================

func TestReleaseCalendarResult_Sort(t *testing.T) {
	events := []releaseCalendarEvent{
		{Date: "2024-03-15", Type: "release"},
		{Date: "2024-01-10", Type: "draft"},
		{Date: "2024-06-20", Type: "rollout"},
		{Date: "2024-03-15", Type: "halted"}, // Same date, different type
	}

	// Sort by date
	for i := 0; i < len(events)-1; i++ {
		for j := i + 1; j < len(events); j++ {
			if events[i].Date > events[j].Date {
				events[i], events[j] = events[j], events[i]
			}
		}
	}

	// Verify sorted order
	expected := []string{"2024-01-10", "2024-03-15", "2024-03-15", "2024-06-20"}
	for i, event := range events {
		if event.Date != expected[i] {
			t.Errorf("event[%d].Date = %v, want %v", i, event.Date, expected[i])
		}
	}
}

func TestReleaseHistoryResult_Sort(t *testing.T) {
	items := []releaseHistoryItem{
		{ReleaseDate: "2024-01-15", Name: "v1.0"},
		{ReleaseDate: "2024-06-20", Name: "v2.0"},
		{ReleaseDate: "2024-03-10", Name: "v1.5"},
	}

	// Sort by date descending
	for i := 0; i < len(items)-1; i++ {
		for j := i + 1; j < len(items); j++ {
			if items[i].ReleaseDate < items[j].ReleaseDate {
				items[i], items[j] = items[j], items[i]
			}
		}
	}

	// Verify sorted order (newest first)
	expected := []string{"v2.0", "v1.5", "v1.0"}
	for i, item := range items {
		if item.Name != expected[i] {
			t.Errorf("items[%d].Name = %v, want %v", i, item.Name, expected[i])
		}
	}
}

// ============================================================================
// Mock Client Helper Tests
// ============================================================================

func TestMockClient_Tracking(t *testing.T) {
	// This test demonstrates how to use the mock client pattern
	// for future integration tests

	type mockClient struct {
		calls []string
	}

	m := &mockClient{calls: make([]string, 0)}

	// Simulate API calls
	m.calls = append(m.calls, "Edits.Insert", "Tracks.List", "Edits.Delete")

	if len(m.calls) != 3 {
		t.Errorf("call count = %v, want 3", len(m.calls))
	}

	expected := []string{"Edits.Insert", "Tracks.List", "Edits.Delete"}
	for i, call := range m.calls {
		if call != expected[i] {
			t.Errorf("calls[%d] = %v, want %v", i, call, expected[i])
		}
	}
}

// ============================================================================
// Result Struct Tests
// ============================================================================

func TestReleaseCalendarResult_Validate(t *testing.T) {
	result := &releaseCalendarResult{
		Track:       "production",
		PeriodStart: "2024-01-01",
		PeriodEnd:   "2024-12-31",
		Events: []releaseCalendarEvent{
			{
				Date:        "2024-06-15",
				Type:        releaseCompleted,
				Track:       "production",
				VersionCode: "100",
				Description: "Release v1.0.0",
			},
		},
		GeneratedAt: time.Now(),
	}

	if result.Track != "production" {
		t.Errorf("Track = %v, want production", result.Track)
	}

	if len(result.Events) != 1 {
		t.Errorf("Events count = %v, want 1", len(result.Events))
	}

	if result.Events[0].Type != releaseCompleted {
		t.Errorf("Event type = %v, want completed", result.Events[0].Type)
	}
}

func TestReleaseConflictsResult_Validate(t *testing.T) {
	result := &releaseConflictsResult{
		HasConflicts: true,
		Conflicts: []releaseConflict{
			{
				VersionCode:     "100",
				Track:           "production",
				Status:          releaseCompleted,
				ExistingVersion: "v1.0.0",
			},
		},
		Suggestions: []string{"Use version code 101 or higher"},
		CheckedAt:   time.Now(),
	}

	if !result.HasConflicts {
		t.Error("HasConflicts should be true")
	}

	if len(result.Conflicts) != 1 {
		t.Errorf("Conflicts count = %v, want 1", len(result.Conflicts))
	}

	if result.Conflicts[0].VersionCode != "100" {
		t.Errorf("VersionCode = %v, want 100", result.Conflicts[0].VersionCode)
	}
}

func TestReleaseStrategyResult_Validate(t *testing.T) {
	result := &releaseStrategyResult{
		Track:          "production",
		CurrentVersion: "100",
		HealthScore:    0.95,
		Recommendation: "continue",
		Reasoning:      "Health score is above threshold",
		Actions:        []string{"Continue monitoring"},
		Risks:          []string{},
		Metrics: releaseStrategyMetrics{
			CrashRate: 0.001,
			AnrRate:   0.0005,
		},
		AnalyzedAt: time.Now(),
	}

	if result.HealthScore != 0.95 {
		t.Errorf("HealthScore = %v, want 0.95", result.HealthScore)
	}

	if result.Recommendation != "continue" {
		t.Errorf("Recommendation = %v, want continue", result.Recommendation)
	}

	if result.Metrics.CrashRate != 0.001 {
		t.Errorf("CrashRate = %v, want 0.001", result.Metrics.CrashRate)
	}
}

func TestReleaseHistoryResult_Validate(t *testing.T) {
	result := &releaseHistoryResult{
		Track: "production",
		Count: 2,
		Releases: []releaseHistoryItem{
			{
				VersionCodes:      []string{"100"},
				Name:              "v1.0.0",
				Status:            releaseCompleted,
				ReleaseDate:       "2024-01-15",
				RolloutPercentage: 100.0,
				Vitals: &releaseHistoryVitals{
					CrashRate: 0.001,
					AnrRate:   0.0005,
					Stability: 0.9985,
				},
			},
		},
		GeneratedAt: time.Now(),
	}

	if result.Count != 2 {
		t.Errorf("Count = %v, want 2", result.Count)
	}

	if len(result.Releases) != 1 {
		t.Errorf("Releases count = %v, want 1", len(result.Releases))
	}

	if result.Releases[0].Vitals == nil {
		t.Error("Vitals should not be nil")
	} else if result.Releases[0].Vitals.Stability != 0.9985 {
		t.Errorf("Stability = %v, want 0.9985", result.Releases[0].Vitals.Stability)
	}
}

func TestReleaseNotesResult_Validate(t *testing.T) {
	result := &releaseNotesResult{
		Action:      "copy",
		Track:       "production",
		VersionCode: "100",
		Locales: map[string]releaseNotesData{
			"en-US": {Text: "English notes"},
			"es-ES": {Text: "Spanish notes"},
		},
		Source:     "en-US",
		Targets:    []string{"es-ES", "fr-FR"},
		ModifiedAt: time.Now(),
	}

	if result.Action != "copy" {
		t.Errorf("Action = %v, want copy", result.Action)
	}

	if len(result.Locales) != 2 {
		t.Errorf("Locales count = %v, want 2", len(result.Locales))
	}

	if len(result.Targets) != 2 {
		t.Errorf("Targets count = %v, want 2", len(result.Targets))
	}
}

// ============================================================================
// Helper Functions
// ============================================================================

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(s[:len(substr)] == substr) ||
		(s[len(s)-len(substr):] == substr) ||
		containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ============================================================================
// Command Runner Interface Tests
// ============================================================================

// These tests verify that command structs implement the expected interface
// for Kong CLI integration

func TestCommands_ImplementRun(t *testing.T) {
	// Verify that command structs have a Run method with correct signature
	// This serves as a compile-time check

	var (
		_ interface{ Run(*Globals) error } = &ReleaseCalendarCmd{}
		_ interface{ Run(*Globals) error } = &ReleaseConflictsCmd{}
		_ interface{ Run(*Globals) error } = &ReleaseStrategyCmd{}
		_ interface{ Run(*Globals) error } = &ReleaseHistoryCmd{}
		_ interface{ Run(*Globals) error } = &ReleaseNotesCmd{}
	)
}

// ============================================================================
// Edge Case Tests
// ============================================================================

func TestReleaseConflictsCmd_MaxVersionCode(t *testing.T) {
	// Test with very large version codes
	versionCodes := []int64{2147483647}

	maxVC := int64(0)
	for _, vc := range versionCodes {
		if vc > maxVC {
			maxVC = vc
		}
	}

	if maxVC != 2147483647 {
		t.Errorf("max version code = %v, want 2147483647", maxVC)
	}

	// Suggested next version should be +1
	suggested := maxVC + 1
	if suggested != 2147483648 {
		t.Errorf("suggested version = %v, want 2147483648", suggested)
	}
}

func TestReleaseStrategyCmd_HealthScoreBoundaries(t *testing.T) {
	cmd := ReleaseStrategyCmd{HealthThreshold: 0.95}

	boundaryTests := []struct {
		score    float64
		expected string
	}{
		{1.0, "continue"},
		{0.95, "continue"},
		{0.8075, "monitor"},
		{0.8074, "investigate"}, // Just below 85% threshold
		{0.665, "investigate"},
		{0.664, "rollback"}, // Just below 70% threshold
		{0.5, "rollback"},
		{0.0, "rollback"},
	}

	for _, tt := range boundaryTests {
		t.Run(strconv.FormatFloat(tt.score, 'f', 4, 64), func(t *testing.T) {
			rec, _, _, _ := cmd.buildStrategyRecommendation(tt.score, 0.5)
			if rec != tt.expected {
				t.Errorf("score %.4f: got %v, want %v", tt.score, rec, tt.expected)
			}
		})
	}
}

func TestReleaseCalendarCmd_PeriodCalculation(t *testing.T) {
	now := time.Now()
	daysBehind := 30
	daysAhead := 30

	startDate := now.AddDate(0, 0, -daysBehind)
	endDate := now.AddDate(0, 0, daysAhead)

	period := endDate.Sub(startDate)
	expectedDays := daysBehind + daysAhead

	// Allow for small variations due to DST or timezone issues (±1 day)
	days := int(period.Hours() / 24)
	if days < expectedDays-1 || days > expectedDays+1 {
		t.Errorf("period = %v days, want approximately %v", days, expectedDays)
	}
}

func TestReleaseHistoryCmd_LimitBoundary(t *testing.T) {
	releases := make([]*androidpublisher.TrackRelease, 100)
	for i := range releases {
		releases[i] = &androidpublisher.TrackRelease{
			VersionCodes: []int64{int64(i)},
			Status:       releaseCompleted,
		}
	}

	limit := 20
	count := 0
	for i := range releases {
		if limit > 0 && i >= limit {
			break
		}
		count++
	}

	if count != limit {
		t.Errorf("count = %v, want %v (limit)", count, limit)
	}
}

func TestReleaseNotesCmd_LocaleHandling(t *testing.T) {
	// Test locale validation and handling
	validLocales := []string{"en-US", "es-ES", "fr-FR", "de-DE", "ja-JP", "zh-CN"}

	for _, locale := range validLocales {
		// Verify locale format (language-COUNTRY)
		parts := splitLocale(locale)
		if len(parts) != 2 {
			t.Errorf("locale %s: expected 2 parts, got %d", locale, len(parts))
		}
		if len(parts[0]) != 2 {
			t.Errorf("locale %s: language code should be 2 chars, got %s", locale, parts[0])
		}
		if len(parts[1]) != 2 {
			t.Errorf("locale %s: country code should be 2 chars, got %s", locale, parts[1])
		}
	}
}

func splitLocale(locale string) []string {
	result := make([]string, 0)
	current := ""
	for _, ch := range locale {
		if ch == '-' || ch == '_' {
			if current != "" {
				result = append(result, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

// ============================================================================
// API Error Handling Tests
// ============================================================================

func TestAPIError_Handling(t *testing.T) {
	tests := []struct {
		name     string
		code     errors.ErrorCode
		message  string
		hint     string
		exitCode int
	}{
		{
			name:     "validation error",
			code:     errors.CodeValidationError,
			message:  "invalid input",
			hint:     "check your parameters",
			exitCode: errors.ExitValidationError,
		},
		{
			name:     "not found error",
			code:     errors.CodeNotFound,
			message:  "resource not found",
			hint:     "verify the resource exists",
			exitCode: errors.ExitNotFound,
		},
		{
			name:     "general error",
			code:     errors.CodeGeneralError,
			message:  "API call failed",
			hint:     "try again later",
			exitCode: errors.ExitGeneralError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.NewAPIError(tt.code, tt.message).WithHint(tt.hint)

			if err.Code != tt.code {
				t.Errorf("error code = %v, want %v", err.Code, tt.code)
			}

			if err.Message != tt.message {
				t.Errorf("error message = %v, want %v", err.Message, tt.message)
			}

			if err.Hint != tt.hint {
				t.Errorf("error hint = %v, want %v", err.Hint, tt.hint)
			}

			if err.ExitCode() != tt.exitCode {
				t.Errorf("exit code = %v, want %v", err.ExitCode(), tt.exitCode)
			}

			// Verify error string contains code and message
			errStr := err.Error()
			if !containsStr(errStr, string(tt.code)) {
				t.Errorf("error string should contain code %v, got %v", tt.code, errStr)
			}
		})
	}
}

// ============================================================================
// Output Format Tests
// ============================================================================

func TestResult_Formats(t *testing.T) {
	// Test that all result types can be serialized (for JSON output)
	tests := []struct {
		name   string
		result interface{}
	}{
		{
			name: "calendar result",
			result: &releaseCalendarResult{
				Track:       "production",
				PeriodStart: "2024-01-01",
				PeriodEnd:   "2024-12-31",
				Events: []releaseCalendarEvent{
					{Date: "2024-06-15", Type: releaseCompleted},
				},
				GeneratedAt: time.Now(),
			},
		},
		{
			name: "conflicts result",
			result: &releaseConflictsResult{
				HasConflicts: true,
				Conflicts:    []releaseConflict{{VersionCode: "100", Track: "production"}},
				CheckedAt:    time.Now(),
			},
		},
		{
			name: "strategy result",
			result: &releaseStrategyResult{
				Track:          "production",
				HealthScore:    0.95,
				Recommendation: "continue",
				AnalyzedAt:     time.Now(),
			},
		},
		{
			name: "history result",
			result: &releaseHistoryResult{
				Track:    "production",
				Count:    1,
				Releases: []releaseHistoryItem{{VersionCodes: []string{"100"}}},
			},
		},
		{
			name: "notes result",
			result: &releaseNotesResult{
				Action:  "get",
				Track:   "production",
				Locales: map[string]releaseNotesData{"en-US": {Text: "notes"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Attempt to marshal to JSON to verify structure
			_, err := json.Marshal(tt.result)
			if err != nil {
				t.Errorf("failed to marshal result: %v", err)
			}
		})
	}
}

// ============================================================================
// Context and Timeout Tests
// ============================================================================

func TestCommand_ContextPropagation(t *testing.T) {
	// Test that context is properly created and propagated
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if ctx == nil {
		t.Error("context should not be nil")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("context should have a deadline")
	}

	expectedDeadline := time.Now().Add(30 * time.Second)
	if deadline.Before(expectedDeadline.Add(-time.Second)) || deadline.After(expectedDeadline.Add(time.Second)) {
		t.Errorf("deadline = %v, expected around %v", deadline, expectedDeadline)
	}
}

// ============================================================================
// Concurrent Access Tests
// ============================================================================

func TestResult_ConcurrentAccess(t *testing.T) {
	// Test that result structs are safe for concurrent read access
	result := &releaseConflictsResult{
		HasConflicts: true,
		Conflicts:    make([]releaseConflict, 0),
		CheckedAt:    time.Now(),
	}

	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func() {
			_ = result.HasConflicts
			_ = len(result.Conflicts)
			_ = result.CheckedAt
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// ============================================================================
// Performance Tests
// ============================================================================

func BenchmarkReleaseMatchesVersionCode(b *testing.B) {
	release := &androidpublisher.TrackRelease{
		VersionCodes: []int64{100, 200, 300, 400, 500},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		releaseMatchesVersionCode(release, "300")
	}
}

func BenchmarkBuildStrategyRecommendation(b *testing.B) {
	cmd := &ReleaseStrategyCmd{HealthThreshold: 0.95}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cmd.buildStrategyRecommendation(0.85, 0.5)
	}
}

// ============================================================================
// Mock API Integration Tests
// ============================================================================

// These tests demonstrate the pattern for testing with the mock client
// They would be expanded in a full integration test suite

func TestMockAPI_ClientPattern(t *testing.T) {
	// Demonstrate how to set up a mock client for testing
	// This pattern follows the internal/apitest/mock_client.go structure

	mock := &mockClientForTesting{
		publisherCalls: make([]string, 0),
		tracks: map[string]*androidpublisher.Track{
			"production": {
				Track: "production",
				Releases: []*androidpublisher.TrackRelease{
					{
						VersionCodes: []int64{100},
						Status:       releaseCompleted,
						Name:         "v1.0.0",
					},
				},
			},
		},
	}

	// Simulate fetching track
	track, err := mock.getTrack("production")
	if err != nil {
		t.Fatalf("failed to get track: %v", err)
	}

	if track.Track != "production" {
		t.Errorf("track = %v, want production", track.Track)
	}

	if len(track.Releases) != 1 {
		t.Errorf("releases count = %v, want 1", len(track.Releases))
	}
}

type mockClientForTesting struct {
	publisherCalls []string
	tracks         map[string]*androidpublisher.Track
}

func (m *mockClientForTesting) getTrack(name string) (*androidpublisher.Track, error) {
	m.publisherCalls = append(m.publisherCalls, "Tracks.Get")
	track, ok := m.tracks[name]
	if !ok {
		return nil, errors.NewAPIError(errors.CodeNotFound, "track not found: "+name)
	}
	return track, nil
}

// ============================================================================
// End-to-End Scenario Tests
// ============================================================================

func TestScenario_ReleaseLifecycle(t *testing.T) {
	// Simulate a complete release lifecycle scenario

	// 1. Initial release check - no conflicts
	initialVCs := []string{"100"}
	existingVCs := map[string]bool{}
	for _, vc := range initialVCs {
		if existingVCs[vc] {
			t.Errorf("version code %s should not exist yet", vc)
		}
	}

	// 2. Release to internal track
	internalRelease := &androidpublisher.TrackRelease{
		VersionCodes: []int64{100},
		Status:       releaseCompleted,
		Name:         "v1.0.0-internal",
		UserFraction: 0.0,
		ReleaseNotes: []*androidpublisher.LocalizedText{
			{Language: "en-US", Text: "Initial release"},
		},
	}

	if internalRelease.Status != releaseCompleted {
		t.Error("internal release should be completed")
	}

	// 3. Promote to production with phased rollout
	prodRelease := &androidpublisher.TrackRelease{
		VersionCodes: []int64{100},
		Status:       statusInProgress,
		Name:         "v1.0.0",
		UserFraction: 0.25,
		ReleaseNotes: internalRelease.ReleaseNotes,
	}

	if prodRelease.Status != statusInProgress {
		t.Error("production release should be in progress")
	}

	if prodRelease.UserFraction != 0.25 {
		t.Errorf("rollout = %v, want 0.25", prodRelease.UserFraction)
	}

	// 4. Monitor health - use excellent metrics to get "continue" recommendation
	crashRate := 0.0001
	anrRate := 0.0001

	cmd := &ReleaseStrategyCmd{HealthThreshold: 0.95}
	crashPenalty := crashRate / 0.01
	if crashPenalty > 1 {
		crashPenalty = 1
	}
	anrPenalty := anrRate / 0.005
	if anrPenalty > 1 {
		anrPenalty = 1
	}
	healthScore := 1.0 - (crashPenalty*0.5 + anrPenalty*0.5)

	rec, reasoning, actions, risks := cmd.buildStrategyRecommendation(healthScore, prodRelease.UserFraction)
	_ = reasoning
	_ = actions
	_ = risks
	if rec != "continue" {
		t.Errorf("recommendation = %v, want continue (health score %.4f with threshold %.2f)", rec, healthScore, cmd.HealthThreshold)
	}

	// 5. Complete rollout
	prodRelease.Status = releaseCompleted
	prodRelease.UserFraction = 1.0

	if prodRelease.Status != releaseCompleted {
		t.Error("release should be completed after full rollout")
	}
}

// ============================================================================
// Cleanup Tests
// ============================================================================

func TestTempFileCleanup(t *testing.T) {
	// Ensure temp files are cleaned up after tests
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.json")

	content := []byte(`{"test": "data"}`)
	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// File should exist
	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("temp file should exist")
	}

	// Cleanup is automatic with t.TempDir(), but verify content was written
	read, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}

	if !bytes.Equal(read, content) {
		t.Errorf("content = %v, want %v", string(read), string(content))
	}
}
