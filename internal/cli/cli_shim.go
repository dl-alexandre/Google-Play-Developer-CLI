// Package cli provides CLI functionality via Kong framework.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"google.golang.org/api/androidpublisher/v3"

	"github.com/dl-alexandre/gpd/internal/auth"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
	"github.com/dl-alexandre/gpd/internal/storage"
)

// CLI provides a shim for the old Cobra-based CLI to support Kong migration.
// This is a temporary compatibility layer until all methods are migrated.
type CLI struct {
	packageName  string
	outputFormat string
	quiet        bool
	verbose      bool
	timeout      time.Duration
	keyPath      string
	authMgr      *auth.Manager
	stdout       io.Writer
	stderr       io.Writer
}

// createCLI creates a new CLI instance from globals for method compatibility.
func createCLI(globals *Globals) *CLI {
	secureStorage := storage.New()
	authMgr := auth.NewManager(secureStorage)

	return &CLI{
		packageName:  globals.Package,
		outputFormat: globals.Output,
		quiet:        globals.Quiet,
		verbose:      globals.Verbose,
		timeout:      globals.Timeout,
		keyPath:      globals.KeyPath,
		authMgr:      authMgr,
		stdout:       os.Stdout,
		stderr:       os.Stderr,
	}
}

// requirePackage validates package name is set.
func (c *CLI) requirePackage() error {
	if c.packageName == "" {
		return errors.NewAPIError(errors.CodeValidationError, "package name is required").
			WithHint("Use --package flag or set GPD_PACKAGE environment variable")
	}
	return nil
}

// getPublisherService returns the Android Publisher service.
func (c *CLI) getPublisherService(ctx context.Context) (*androidpublisher.Service, error) {
	return nil, errors.NewAPIError(errors.CodeGeneralError, "publisher service not yet implemented in Kong migration")
}

// Output outputs a result.
func (c *CLI) Output(result *output.Result) error {
	_, _ = fmt.Fprintln(c.stdout, result)
	return nil
}

// OutputError outputs an error.
func (c *CLI) OutputError(err *errors.APIError) error {
	_, _ = fmt.Fprintf(c.stderr, "Error: %s\n", err.Message)
	return err
}

// ExitSuccess returns the success exit code.
func ExitSuccess() int {
	return errors.ExitSuccess
}

// ExitError returns the general error exit code.
func ExitError() int {
	return errors.ExitGeneralError
}

// Stub methods for Kong migration compatibility
// These return "not implemented" to allow compilation while work continues

func (c *CLI) publishUpload(ctx context.Context, file string, opts obbOptions, editID string, noAutoCommit, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish upload not yet implemented in Kong migration")
}

func (c *CLI) publishRelease(ctx context.Context, track, name, status string, versionCodes, retainVersionCodes []string, priority int, notesFile string, editID string, noAutoCommit, dryRun, wait bool, waitTimeout string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish release not yet implemented in Kong migration")
}

func (c *CLI) publishRollout(ctx context.Context, track string, percentage float64, editID string, noAutoCommit, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish rollout not yet implemented in Kong migration")
}

func (c *CLI) publishPromote(ctx context.Context, fromTrack, toTrack string, percentage float64, editID string, noAutoCommit, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish promote not yet implemented in Kong migration")
}

func (c *CLI) publishHalt(ctx context.Context, track string, editID string, noAutoCommit bool, force bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish halt not yet implemented in Kong migration")
}

func (c *CLI) publishRollback(ctx context.Context, track string, versionCode string, editID string, noAutoCommit bool, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish rollback not yet implemented in Kong migration")
}

func (c *CLI) publishStatus(ctx context.Context, track string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish status not yet implemented in Kong migration")
}

func (c *CLI) publishTracks(ctx context.Context) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish tracks not yet implemented in Kong migration")
}

func (c *CLI) publishCapabilities(ctx context.Context) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish capabilities not yet implemented in Kong migration")
}

func (c *CLI) publishListingUpdate(ctx context.Context, locale, title, shortDesc, fullDesc, editID string, noAutoCommit bool, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing update not yet implemented in Kong migration")
}

func (c *CLI) publishListingGet(ctx context.Context, locale string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing get not yet implemented in Kong migration")
}

func (c *CLI) publishListingDelete(ctx context.Context, locale string, editID string, noAutoCommit bool, dryRun bool, confirm bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing delete not yet implemented in Kong migration")
}

func (c *CLI) publishListingDeleteAll(ctx context.Context, editID string, noAutoCommit bool, dryRun bool, confirm bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish listing delete all not yet implemented in Kong migration")
}

func (c *CLI) publishDetailsGet(ctx context.Context) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish details get not yet implemented in Kong migration")
}

func (c *CLI) publishDetailsUpdate(ctx context.Context, email, phone, website, defaultLanguage string, editID string, noAutoCommit bool, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish details update not yet implemented in Kong migration")
}

func (c *CLI) publishDetailsPatch(ctx context.Context, email, phone, website, defaultLanguage, updateMask string, editID string, noAutoCommit bool, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish details patch not yet implemented in Kong migration")
}

func (c *CLI) publishImagesUpload(ctx context.Context, imageType, file, locale string, syncImages bool, editID string, noAutoCommit bool, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images upload not yet implemented in Kong migration")
}

func (c *CLI) publishImagesList(ctx context.Context, imageType, locale string, editID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images list not yet implemented in Kong migration")
}

func (c *CLI) publishImagesDelete(ctx context.Context, imageType, imageID, locale string, editID string, noAutoCommit bool, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images delete not yet implemented in Kong migration")
}

func (c *CLI) publishImagesDeleteAll(ctx context.Context, imageType, locale string, editID string, noAutoCommit bool, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish images delete all not yet implemented in Kong migration")
}

func (c *CLI) publishAssetsUpload(ctx context.Context, assetType, file string, editID string, noAutoCommit bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish assets upload not yet implemented in Kong migration")
}

func (c *CLI) publishAssetsSpec(ctx context.Context, assetType string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish assets spec not yet implemented in Kong migration")
}

func (c *CLI) publishDeobfuscationUpload(ctx context.Context, versionCode int64, file, mappingType string, force bool, editID string, noAutoCommit bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish deobfuscation upload not yet implemented in Kong migration")
}

func (c *CLI) publishTestersList(ctx context.Context, track string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers list not yet implemented in Kong migration")
}

func (c *CLI) publishTestersAdd(ctx context.Context, track string, emails []string, editID string, noAutoCommit bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers add not yet implemented in Kong migration")
}

func (c *CLI) publishTestersRemove(ctx context.Context, track string, emails []string, editID string, noAutoCommit bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish testers remove not yet implemented in Kong migration")
}

func (c *CLI) publishInternalShareUpload(ctx context.Context, file string, expiresIn time.Duration) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish internal share upload not yet implemented in Kong migration")
}

func (c *CLI) publishBuildsList(ctx context.Context, buildType string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds list not yet implemented in Kong migration")
}

func (c *CLI) publishBuildsGet(ctx context.Context, versionCode int64, buildType string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds get not yet implemented in Kong migration")
}

func (c *CLI) publishBuildsExpire(ctx context.Context, versionCode int64, buildType string, force bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds expire not yet implemented in Kong migration")
}

func (c *CLI) publishBuildsExpireAll(ctx context.Context, buildType string, force bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish builds expire all not yet implemented in Kong migration")
}

func (c *CLI) publishBetaGroupsAdd(ctx context.Context, groupName string, emails []string, editID string, noAutoCommit bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta groups add not yet implemented in Kong migration")
}

func (c *CLI) publishBetaGroupsRemove(ctx context.Context, groupName string, emails []string, editID string, noAutoCommit bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta groups remove not yet implemented in Kong migration")
}

func (c *CLI) publishBetaGroupsList(ctx context.Context, groupName string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "publish beta groups list not yet implemented in Kong migration")
}

// Vitals stubs
func (c *CLI) vitalsCrashes(ctx context.Context, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals crashes not yet implemented in Kong migration")
}

func (c *CLI) vitalsANRs(ctx context.Context, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals ANRs not yet implemented in Kong migration")
}

func (c *CLI) vitalsExcessiveWakeups(ctx context.Context, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals excessive wakeups not yet implemented in Kong migration")
}

func (c *CLI) vitalsLmkRate(ctx context.Context, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals LMK rate not yet implemented in Kong migration")
}

func (c *CLI) vitalsSlowRendering(ctx context.Context, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals slow rendering not yet implemented in Kong migration")
}

func (c *CLI) vitalsSlowStart(ctx context.Context, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals slow start not yet implemented in Kong migration")
}

func (c *CLI) vitalsStuckWakelocks(ctx context.Context, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals stuck wakelocks not yet implemented in Kong migration")
}

func (c *CLI) vitalsErrorsIssuesSearch(ctx context.Context, errorType, clusterID string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals errors issues search not yet implemented in Kong migration")
}

func (c *CLI) vitalsErrorsReportsSearch(ctx context.Context, errorType, clusterID, versionCode string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals errors reports search not yet implemented in Kong migration")
}

func (c *CLI) vitalsErrorsCountsGet(ctx context.Context, metricSet, metric string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals errors counts get not yet implemented in Kong migration")
}

func (c *CLI) vitalsErrorsCountsQuery(ctx context.Context, metricSet string, metrics []string, interval string, filters map[string]string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals errors counts query not yet implemented in Kong migration")
}

func (c *CLI) vitalsAnomaliesList(ctx context.Context, metric string, startTime, endTime time.Time) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals anomalies list not yet implemented in Kong migration")
}

func (c *CLI) vitalsQuery(ctx context.Context, metric, startDate, endDate string, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals query not yet implemented in Kong migration")
}

func (c *CLI) vitalsCapabilities(ctx context.Context) error {
	return errors.NewAPIError(errors.CodeGeneralError, "vitals capabilities not yet implemented in Kong migration")
}

// Apps stubs
func (c *CLI) appsList(ctx context.Context, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "apps list not yet implemented in Kong migration")
}

func (c *CLI) appsGet(ctx context.Context, packageName string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "apps get not yet implemented in Kong migration")
}

// Games stubs
func (c *CLI) gamesAchievementsReset(ctx context.Context, achievementID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games achievements reset not yet implemented in Kong migration")
}

func (c *CLI) gamesAchievementsResetAll(ctx context.Context, allPlayers bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games achievements reset all not yet implemented in Kong migration")
}

func (c *CLI) gamesAchievementsResetForAllPlayers(ctx context.Context, achievementID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games achievements reset for all players not yet implemented in Kong migration")
}

func (c *CLI) gamesAchievementsResetMultipleForAllPlayers(ctx context.Context, achievementIDs []string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games achievements reset multiple not yet implemented in Kong migration")
}

func (c *CLI) gamesScoresReset(ctx context.Context, leaderboardID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games scores reset not yet implemented in Kong migration")
}

func (c *CLI) gamesScoresResetAll(ctx context.Context, allPlayers bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games scores reset all not yet implemented in Kong migration")
}

func (c *CLI) gamesScoresResetForAllPlayers(ctx context.Context, leaderboardID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games scores reset for all players not yet implemented in Kong migration")
}

func (c *CLI) gamesScoresResetMultipleForAllPlayers(ctx context.Context, leaderboardIDs []string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games scores reset multiple not yet implemented in Kong migration")
}

func (c *CLI) gamesEventsReset(ctx context.Context, eventID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games events reset not yet implemented in Kong migration")
}

func (c *CLI) gamesEventsResetAll(ctx context.Context, allPlayers bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games events reset all not yet implemented in Kong migration")
}

func (c *CLI) gamesEventsResetForAllPlayers(ctx context.Context, eventID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games events reset for all players not yet implemented in Kong migration")
}

func (c *CLI) gamesEventsResetMultipleForAllPlayers(ctx context.Context, eventIDs []string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games events reset multiple not yet implemented in Kong migration")
}

func (c *CLI) gamesPlayersHide(ctx context.Context, applicationID, playerID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games players hide not yet implemented in Kong migration")
}

func (c *CLI) gamesPlayersUnhide(ctx context.Context, applicationID, playerID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games players unhide not yet implemented in Kong migration")
}

func (c *CLI) gamesCapabilities(ctx context.Context) error {
	return errors.NewAPIError(errors.CodeGeneralError, "games capabilities not yet implemented in Kong migration")
}

// Analytics stubs
func (c *CLI) analyticsQuery(ctx context.Context, startDate, endDate string, metrics, dimensions []string, format string, pageSize int64, pageToken string, all bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "analytics query not yet implemented in Kong migration")
}

func (c *CLI) analyticsCapabilities(ctx context.Context) error {
	return errors.NewAPIError(errors.CodeGeneralError, "analytics capabilities not yet implemented in Kong migration")
}

// Reviews stubs
func (c *CLI) reviewsList(ctx context.Context, params reviewsListParams) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews list not yet implemented in Kong migration")
}

func (c *CLI) reviewsGet(ctx context.Context, reviewID string, translate string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews get not yet implemented in Kong migration")
}

func (c *CLI) reviewsReply(ctx context.Context, reviewID, replyText, template string, rateLimit time.Duration, dryRun bool) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews reply not yet implemented in Kong migration")
}

func (c *CLI) reviewsResponseGet(ctx context.Context, reviewID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews response get not yet implemented in Kong migration")
}

func (c *CLI) reviewsResponseDelete(ctx context.Context, reviewID string) error {
	return errors.NewAPIError(errors.CodeGeneralError, "reviews response delete not yet implemented in Kong migration")
}

// Type definitions for compatibility
type obbOptions struct {
	mainPath              string
	patchPath             string
	mainReferenceVersion  int64
	patchReferenceVersion int64
}

type detailsPatchParams struct {
	email           string
	phone           string
	website         string
	defaultLanguage string
	updateMask      string
	editID          string
	noAutoCommit    bool
}

type reviewsListParams struct {
	MinRating int
	MaxRating int
	Language  string
	StartDate string
	EndDate   string
	PageSize  int64
	PageToken string
	All       bool
}
