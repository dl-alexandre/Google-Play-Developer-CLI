package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

func TestCreateCLI(t *testing.T) {
	t.Parallel()
	globals := &Globals{
		Package:     "com.example.test",
		Output:      "json",
		Quiet:       true,
		Verbose:     false,
		Timeout:     60 * time.Second,
		KeyPath:     "/path/to/key.json",
		StoreTokens: "secure",
	}

	cli := createCLI(globals)

	if cli == nil {
		t.Fatal("createCLI() returned nil")
	}
	if cli.packageName != "com.example.test" {
		t.Errorf("packageName = %q, want 'com.example.test'", cli.packageName)
	}
	if cli.outputFormat != "json" {
		t.Errorf("outputFormat = %q, want 'json'", cli.outputFormat)
	}
	if !cli.quiet {
		t.Error("quiet should be true")
	}
	if cli.verbose {
		t.Error("verbose should be false")
	}
	if cli.timeout != 60*time.Second {
		t.Errorf("timeout = %v, want 60s", cli.timeout)
	}
	if cli.keyPath != "/path/to/key.json" {
		t.Errorf("keyPath = %q, want '/path/to/key.json'", cli.keyPath)
	}
	if cli.authMgr == nil {
		t.Error("authMgr should not be nil")
	}
	if cli.stdout == nil {
		t.Error("stdout should not be nil")
	}
	if cli.stderr == nil {
		t.Error("stderr should not be nil")
	}
}

func TestCreateCLIWithMinimalGlobals(t *testing.T) {
	t.Parallel()
	globals := &Globals{}

	cli := createCLI(globals)

	if cli == nil {
		t.Fatal("createCLI() returned nil")
	}
	if cli.packageName != "" {
		t.Error("packageName should be empty")
	}
	if cli.outputFormat != "" {
		t.Error("outputFormat should be empty")
	}
	if cli.quiet {
		t.Error("quiet should be false")
	}
	if cli.verbose {
		t.Error("verbose should be false")
	}
}

func TestCLIRequirePackage(t *testing.T) {
	t.Parallel()
	cli := &CLI{packageName: "com.example.valid"}

	err := cli.requirePackage()
	if err != nil {
		t.Errorf("requirePackage() error = %v, want nil", err)
	}
}

func TestCLIRequirePackageEmpty(t *testing.T) {
	t.Parallel()
	cli := &CLI{packageName: ""}

	err := cli.requirePackage()
	if err == nil {
		t.Fatal("requirePackage() should return error for empty package")
	}

	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("error should be *APIError, got %T", err)
	}
	if apiErr.Code != errors.CodeValidationError {
		t.Errorf("error code = %v, want %v", apiErr.Code, errors.CodeValidationError)
	}
	if !strings.Contains(apiErr.Message, "package name is required") {
		t.Errorf("error message = %q, should contain 'package name is required'", apiErr.Message)
	}
	if !strings.Contains(apiErr.Hint, "--package") {
		t.Errorf("error hint = %q, should contain '--package'", apiErr.Hint)
	}
}

func TestCLIGetPublisherService(t *testing.T) {
	t.Parallel()
	cli := &CLI{}
	ctx := context.Background()

	svc, err := cli.getPublisherService(ctx)
	if svc != nil {
		t.Error("getPublisherService() should return nil service")
	}
	if err == nil {
		t.Fatal("getPublisherService() should return error")
	}

	apiErr, ok := err.(*errors.APIError)
	if !ok {
		t.Fatalf("error should be *APIError, got %T", err)
	}
	if apiErr.Code != errors.CodeGeneralError {
		t.Errorf("error code = %v, want %v", apiErr.Code, errors.CodeGeneralError)
	}
	if !strings.Contains(apiErr.Message, "publisher service not yet implemented") {
		t.Errorf("error message = %q, should contain 'publisher service not yet implemented'", apiErr.Message)
	}
}

func TestCLIOutput(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	cli := &CLI{stdout: &stdout}

	result := output.NewResult(map[string]interface{}{"test": "value"})
	err := cli.Output(result)

	if err != nil {
		t.Errorf("Output() error = %v, want nil", err)
	}

	output := stdout.String()
	if output == "" {
		t.Error("Output() should write to stdout")
	}
}

func TestCLIOutputNilResult(t *testing.T) {
	t.Parallel()
	var stdout bytes.Buffer
	cli := &CLI{stdout: &stdout}

	var result *output.Result
	err := cli.Output(result)

	if err != nil {
		t.Errorf("Output() error = %v, want nil", err)
	}
}

func TestCLIOutputError(t *testing.T) {
	t.Parallel()
	var stderr bytes.Buffer
	cli := &CLI{stderr: &stderr}

	apiErr := errors.NewAPIError(errors.CodeValidationError, "test error").
		WithHint("test hint")
	err := cli.OutputError(apiErr)

	if err == nil {
		t.Fatal("OutputError() should return the error")
	}
	if err != apiErr {
		t.Error("OutputError() should return the same error")
	}

	output := stderr.String()
	if !strings.Contains(output, "Error: test error") {
		t.Errorf("stderr = %q, should contain 'Error: test error'", output)
	}
}

func TestExitSuccess(t *testing.T) {
	t.Parallel()
	code := ExitSuccess()
	if code != 0 {
		t.Errorf("ExitSuccess() = %d, want 0", code)
	}
}

func TestExitError(t *testing.T) {
	t.Parallel()
	code := ExitError()
	if code != 1 {
		t.Errorf("ExitError() = %d, want 1", code)
	}
}

func TestCLIStructFields(t *testing.T) {
	t.Parallel()
	cli := &CLI{
		packageName:  "test",
		outputFormat: "json",
		quiet:        true,
		verbose:      false,
		timeout:      30 * time.Second,
		keyPath:      "/key.json",
	}

	if cli.packageName != "test" {
		t.Error("packageName field not set correctly")
	}
	if cli.outputFormat != "json" {
		t.Error("outputFormat field not set correctly")
	}
	if !cli.quiet {
		t.Error("quiet field not set correctly")
	}
	if cli.verbose {
		t.Error("verbose field not set correctly")
	}
	if cli.timeout != 30*time.Second {
		t.Error("timeout field not set correctly")
	}
	if cli.keyPath != "/key.json" {
		t.Error("keyPath field not set correctly")
	}
}

// Test all publish stub methods return proper errors
func TestPublishStubMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cli := &CLI{}

	stubMethods := []struct {
		name string
		fn   func() error
	}{
		{"publishUpload", func() error { return cli.publishUpload(ctx, "file.apk", obbOptions{}, "edit1", false, false) }},
		{"publishRelease", func() error { return cli.publishRelease(ctx, "internal", "v1.0", "draft", []string{"1"}, []string{}, 0, "", "edit1", false, false, false, "") }},
		{"publishRollout", func() error { return cli.publishRollout(ctx, "internal", 50.0, "edit1", false, false) }},
		{"publishPromote", func() error { return cli.publishPromote(ctx, "internal", "alpha", 100.0, "edit1", false, false) }},
		{"publishHalt", func() error { return cli.publishHalt(ctx, "internal", "edit1", false, false) }},
		{"publishRollback", func() error { return cli.publishRollback(ctx, "internal", "1", "edit1", false, false) }},
		{"publishStatus", func() error { return cli.publishStatus(ctx, "internal") }},
		{"publishTracks", func() error { return cli.publishTracks(ctx) }},
		{"publishCapabilities", func() error { return cli.publishCapabilities(ctx) }},
		{"publishListingUpdate", func() error { return cli.publishListingUpdate(ctx, "en", "Title", "Short", "Full", "edit1", false, false) }},
		{"publishListingGet", func() error { return cli.publishListingGet(ctx, "en") }},
		{"publishListingDelete", func() error { return cli.publishListingDelete(ctx, "en", "edit1", false, false, false) }},
		{"publishListingDeleteAll", func() error { return cli.publishListingDeleteAll(ctx, "edit1", false, false, false) }},
		{"publishDetailsGet", func() error { return cli.publishDetailsGet(ctx) }},
		{"publishDetailsUpdate", func() error { return cli.publishDetailsUpdate(ctx, "test@example.com", "123", "http://example.com", "en", "edit1", false, false) }},
		{"publishDetailsPatch", func() error { return cli.publishDetailsPatch(ctx, "test@example.com", "123", "http://example.com", "en", "email", "edit1", false, false) }},
		{"publishImagesUpload", func() error { return cli.publishImagesUpload(ctx, "icon", "file.png", "en", false, "edit1", false, false) }},
		{"publishImagesList", func() error { return cli.publishImagesList(ctx, "icon", "en", "edit1") }},
		{"publishImagesDelete", func() error { return cli.publishImagesDelete(ctx, "icon", "img1", "en", "edit1", false, false) }},
		{"publishImagesDeleteAll", func() error { return cli.publishImagesDeleteAll(ctx, "icon", "en", "edit1", false, false) }},
		{"publishAssetsUpload", func() error { return cli.publishAssetsUpload(ctx, "asset", "file", "edit1", false) }},
		{"publishAssetsSpec", func() error { return cli.publishAssetsSpec(ctx, "asset") }},
		{"publishDeobfuscationUpload", func() error { return cli.publishDeobfuscationUpload(ctx, 1, "mapping.txt", "proguard", false, "edit1", false) }},
		{"publishTestersList", func() error { return cli.publishTestersList(ctx, "internal") }},
		{"publishTestersAdd", func() error { return cli.publishTestersAdd(ctx, "internal", []string{"test@example.com"}, "edit1", false) }},
		{"publishTestersRemove", func() error { return cli.publishTestersRemove(ctx, "internal", []string{"test@example.com"}, "edit1", false) }},
		{"publishInternalShareUpload", func() error { return cli.publishInternalShareUpload(ctx, "file.apk", time.Hour) }},
		{"publishBuildsList", func() error { return cli.publishBuildsList(ctx, "apk", 10, "", false) }},
		{"publishBuildsGet", func() error { return cli.publishBuildsGet(ctx, 1, "apk") }},
		{"publishBuildsExpire", func() error { return cli.publishBuildsExpire(ctx, 1, "apk", false) }},
		{"publishBuildsExpireAll", func() error { return cli.publishBuildsExpireAll(ctx, "apk", false) }},
		{"publishBetaGroupsAdd", func() error { return cli.publishBetaGroupsAdd(ctx, "group1", []string{"test@example.com"}, "edit1", false) }},
		{"publishBetaGroupsRemove", func() error { return cli.publishBetaGroupsRemove(ctx, "group1", []string{"test@example.com"}, "edit1", false) }},
		{"publishBetaGroupsList", func() error { return cli.publishBetaGroupsList(ctx, "group1") }},
	}

	for _, tt := range stubMethods {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatalf("%s should return error", tt.name)
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("%s error should be *APIError, got %T", tt.name, err)
			}
			if apiErr.Code != errors.CodeGeneralError {
				t.Errorf("%s error code = %v, want %v", tt.name, apiErr.Code, errors.CodeGeneralError)
			}
			if !strings.Contains(apiErr.Message, "not yet implemented") {
				t.Errorf("%s error message = %q, should contain 'not yet implemented'", tt.name, apiErr.Message)
			}
		})
	}
}

// Test all vitals stub methods return proper errors
func TestVitalsStubMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cli := &CLI{}

	stubMethods := []struct {
		name string
		fn   func() error
	}{
		{"vitalsCrashes", func() error { return cli.vitalsCrashes(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsANRs", func() error { return cli.vitalsANRs(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsExcessiveWakeups", func() error { return cli.vitalsExcessiveWakeups(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsLmkRate", func() error { return cli.vitalsLmkRate(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsSlowRendering", func() error { return cli.vitalsSlowRendering(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsSlowStart", func() error { return cli.vitalsSlowStart(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsStuckWakelocks", func() error { return cli.vitalsStuckWakelocks(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsErrorsIssuesSearch", func() error { return cli.vitalsErrorsIssuesSearch(ctx, "crash", "cluster1", 10, "", false) }},
		{"vitalsErrorsReportsSearch", func() error { return cli.vitalsErrorsReportsSearch(ctx, "crash", "cluster1", "1", 10, "", false) }},
		{"vitalsErrorsCountsGet", func() error { return cli.vitalsErrorsCountsGet(ctx, "metrics", "count") }},
		{"vitalsErrorsCountsQuery", func() error { return cli.vitalsErrorsCountsQuery(ctx, "metrics", []string{"count"}, "day", map[string]string{}) }},
		{"vitalsAnomaliesList", func() error { return cli.vitalsAnomaliesList(ctx, "crash", time.Now().Add(-7*24*time.Hour), time.Now()) }},
		{"vitalsQuery", func() error { return cli.vitalsQuery(ctx, "crash", "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }},
		{"vitalsCapabilities", func() error { return cli.vitalsCapabilities(ctx) }},
	}

	for _, tt := range stubMethods {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatalf("%s should return error", tt.name)
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("%s error should be *APIError, got %T", tt.name, err)
			}
			if apiErr.Code != errors.CodeGeneralError {
				t.Errorf("%s error code = %v, want %v", tt.name, apiErr.Code, errors.CodeGeneralError)
			}
		})
	}
}

// Test all apps stub methods return proper errors
func TestAppsStubMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cli := &CLI{}

	stubMethods := []struct {
		name string
		fn   func() error
	}{
		{"appsList", func() error { return cli.appsList(ctx, 10, "", false) }},
		{"appsGet", func() error { return cli.appsGet(ctx, "com.example.app") }},
	}

	for _, tt := range stubMethods {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatalf("%s should return error", tt.name)
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("%s error should be *APIError, got %T", tt.name, err)
			}
			if apiErr.Code != errors.CodeGeneralError {
				t.Errorf("%s error code = %v, want %v", tt.name, apiErr.Code, errors.CodeGeneralError)
			}
		})
	}
}

// Test all games stub methods return proper errors
func TestGamesStubMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cli := &CLI{}

	stubMethods := []struct {
		name string
		fn   func() error
	}{
		{"gamesAchievementsReset", func() error { return cli.gamesAchievementsReset(ctx, "achievement1") }},
		{"gamesAchievementsResetAll", func() error { return cli.gamesAchievementsResetAll(ctx, true) }},
		{"gamesAchievementsResetForAllPlayers", func() error { return cli.gamesAchievementsResetForAllPlayers(ctx, "achievement1") }},
		{"gamesAchievementsResetMultipleForAllPlayers", func() error { return cli.gamesAchievementsResetMultipleForAllPlayers(ctx, []string{"a1", "a2"}) }},
		{"gamesScoresReset", func() error { return cli.gamesScoresReset(ctx, "leaderboard1") }},
		{"gamesScoresResetAll", func() error { return cli.gamesScoresResetAll(ctx, true) }},
		{"gamesScoresResetForAllPlayers", func() error { return cli.gamesScoresResetForAllPlayers(ctx, "leaderboard1") }},
		{"gamesScoresResetMultipleForAllPlayers", func() error { return cli.gamesScoresResetMultipleForAllPlayers(ctx, []string{"l1", "l2"}) }},
		{"gamesEventsReset", func() error { return cli.gamesEventsReset(ctx, "event1") }},
		{"gamesEventsResetAll", func() error { return cli.gamesEventsResetAll(ctx, true) }},
		{"gamesEventsResetForAllPlayers", func() error { return cli.gamesEventsResetForAllPlayers(ctx, "event1") }},
		{"gamesEventsResetMultipleForAllPlayers", func() error { return cli.gamesEventsResetMultipleForAllPlayers(ctx, []string{"e1", "e2"}) }},
		{"gamesPlayersHide", func() error { return cli.gamesPlayersHide(ctx, "app1", "player1") }},
		{"gamesPlayersUnhide", func() error { return cli.gamesPlayersUnhide(ctx, "app1", "player1") }},
		{"gamesCapabilities", func() error { return cli.gamesCapabilities(ctx) }},
	}

	for _, tt := range stubMethods {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatalf("%s should return error", tt.name)
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("%s error should be *APIError, got %T", tt.name, err)
			}
			if apiErr.Code != errors.CodeGeneralError {
				t.Errorf("%s error code = %v, want %v", tt.name, apiErr.Code, errors.CodeGeneralError)
			}
		})
	}
}

// Test all analytics stub methods return proper errors
func TestAnalyticsStubMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cli := &CLI{}

	stubMethods := []struct {
		name string
		fn   func() error
	}{
		{"analyticsQuery", func() error { return cli.analyticsQuery(ctx, "2024-01-01", "2024-01-31", []string{"metric"}, []string{}, "json", 10, "", false) }},
		{"analyticsCapabilities", func() error { return cli.analyticsCapabilities(ctx) }},
	}

	for _, tt := range stubMethods {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatalf("%s should return error", tt.name)
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("%s error should be *APIError, got %T", tt.name, err)
			}
			if apiErr.Code != errors.CodeGeneralError {
				t.Errorf("%s error code = %v, want %v", tt.name, apiErr.Code, errors.CodeGeneralError)
			}
		})
	}
}

// Test all reviews stub methods return proper errors
func TestReviewsStubMethods(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cli := &CLI{}

	stubMethods := []struct {
		name string
		fn   func() error
	}{
		{"reviewsList", func() error { return cli.reviewsList(ctx, reviewsListParams{}) }},
		{"reviewsGet", func() error { return cli.reviewsGet(ctx, "review1", "en") }},
		{"reviewsReply", func() error { return cli.reviewsReply(ctx, "review1", "Thank you", "", time.Second, false) }},
		{"reviewsResponseGet", func() error { return cli.reviewsResponseGet(ctx, "review1") }},
		{"reviewsResponseDelete", func() error { return cli.reviewsResponseDelete(ctx, "review1") }},
	}

	for _, tt := range stubMethods {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatalf("%s should return error", tt.name)
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("%s error should be *APIError, got %T", tt.name, err)
			}
			if apiErr.Code != errors.CodeGeneralError {
				t.Errorf("%s error code = %v, want %v", tt.name, apiErr.Code, errors.CodeGeneralError)
			}
		})
	}
}

// Test compatibility type definitions
func TestObbOptionsStruct(t *testing.T) {
	t.Parallel()
	opts := obbOptions{
		mainPath:              "/path/to/main.obb",
		patchPath:             "/path/to/patch.obb",
		mainReferenceVersion:  1,
		patchReferenceVersion: 2,
	}

	if opts.mainPath != "/path/to/main.obb" {
		t.Error("mainPath field not set correctly")
	}
	if opts.patchPath != "/path/to/patch.obb" {
		t.Error("patchPath field not set correctly")
	}
	if opts.mainReferenceVersion != 1 {
		t.Error("mainReferenceVersion field not set correctly")
	}
	if opts.patchReferenceVersion != 2 {
		t.Error("patchReferenceVersion field not set correctly")
	}
}

func TestDetailsPatchParamsStruct(t *testing.T) {
	t.Parallel()
	params := detailsPatchParams{
		email:           "test@example.com",
		phone:           "123-456-7890",
		website:         "https://example.com",
		defaultLanguage: "en",
		updateMask:      "email,phone",
		editID:          "edit1",
		noAutoCommit:    true,
	}

	if params.email != "test@example.com" {
		t.Error("email field not set correctly")
	}
	if params.phone != "123-456-7890" {
		t.Error("phone field not set correctly")
	}
	if params.website != "https://example.com" {
		t.Error("website field not set correctly")
	}
	if params.defaultLanguage != "en" {
		t.Error("defaultLanguage field not set correctly")
	}
	if params.updateMask != "email,phone" {
		t.Error("updateMask field not set correctly")
	}
	if params.editID != "edit1" {
		t.Error("editID field not set correctly")
	}
	if !params.noAutoCommit {
		t.Error("noAutoCommit field not set correctly")
	}
}

func TestReviewsListParamsStruct(t *testing.T) {
	t.Parallel()
	params := reviewsListParams{
		MinRating: 3,
		MaxRating: 5,
		Language:  "en",
		StartDate: "2024-01-01",
		EndDate:   "2024-01-31",
		PageSize:  10,
		PageToken: "token1",
		All:       true,
	}

	if params.MinRating != 3 {
		t.Error("MinRating field not set correctly")
	}
	if params.MaxRating != 5 {
		t.Error("MaxRating field not set correctly")
	}
	if params.Language != "en" {
		t.Error("Language field not set correctly")
	}
	if params.StartDate != "2024-01-01" {
		t.Error("StartDate field not set correctly")
	}
	if params.EndDate != "2024-01-31" {
		t.Error("EndDate field not set correctly")
	}
	if params.PageSize != 10 {
		t.Error("PageSize field not set correctly")
	}
	if params.PageToken != "token1" {
		t.Error("PageToken field not set correctly")
	}
	if !params.All {
		t.Error("All field not set correctly")
	}
}

// Test error message content for specific stub methods
func TestStubMethodErrorMessages(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	cli := &CLI{}

	tests := []struct {
		name            string
		fn              func() error
		expectedMessage string
	}{
		{"publishUpload", func() error { return cli.publishUpload(ctx, "file.apk", obbOptions{}, "edit1", false, false) }, "publish upload"},
		{"publishRelease", func() error { return cli.publishRelease(ctx, "internal", "v1.0", "draft", []string{"1"}, []string{}, 0, "", "edit1", false, false, false, "") }, "publish release"},
		{"vitalsCrashes", func() error { return cli.vitalsCrashes(ctx, "2024-01-01", "2024-01-31", []string{}, "json", 10, "", false) }, "vitals crashes"},
		{"gamesAchievementsReset", func() error { return cli.gamesAchievementsReset(ctx, "achievement1") }, "games achievements reset"},
		{"reviewsList", func() error { return cli.reviewsList(ctx, reviewsListParams{}) }, "reviews list"},
		{"analyticsQuery", func() error { return cli.analyticsQuery(ctx, "2024-01-01", "2024-01-31", []string{"metric"}, []string{}, "json", 10, "", false) }, "analytics query"},
		{"appsList", func() error { return cli.appsList(ctx, 10, "", false) }, "apps list"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.fn()
			if err == nil {
				t.Fatalf("%s should return error", tt.name)
			}
			apiErr, ok := err.(*errors.APIError)
			if !ok {
				t.Fatalf("%s error should be *APIError, got %T", tt.name, err)
			}
			if !strings.Contains(apiErr.Message, tt.expectedMessage) {
				t.Errorf("%s error message = %q, should contain %q", tt.name, apiErr.Message, tt.expectedMessage)
			}
		})
	}
}

// Test that CLI fields are properly initialized with nil globals fields
func TestCreateCLINilFields(t *testing.T) {
	t.Parallel()
	globals := &Globals{
		Package: "",
	}

	cli := createCLI(globals)

	if cli.packageName != "" {
		t.Error("empty package name should remain empty")
	}
}

// Test that CLI stdout and stderr are set correctly to custom writers
func TestCLICustomWriters(t *testing.T) {
	t.Parallel()
	var stdout, stderr bytes.Buffer

	cli := &CLI{
		stdout: &stdout,
		stderr: &stderr,
	}

	result := output.NewResult(map[string]interface{}{"key": "value"})
	cli.Output(result)

	if stdout.Len() == 0 {
		t.Error("stdout should have content after Output()")
	}

	apiErr := errors.NewAPIError(errors.CodeValidationError, "test")
	cli.OutputError(apiErr)

	if stderr.Len() == 0 {
		t.Error("stderr should have content after OutputError()")
	}
}
