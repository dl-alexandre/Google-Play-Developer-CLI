//go:build unit
// +build unit

package cli

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/cli/playship"
)

func TestValidateCmd_DryRun_NoPackage(t *testing.T) {
	cmd := &ValidateCmd{Track: "internal", DryRun: true}
	globals := &Globals{Output: "json", Pretty: false}
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("expected validation error without package")
	}
}

func TestValidateCmd_DryRun_WithPackage(t *testing.T) {
	cmd := &ValidateCmd{Track: "internal", DryRun: true}
	globals := &Globals{Output: "json", Package: "com.example.app", Profile: "default"}
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCmd_DryRun_WithArtifact(t *testing.T) {
	dir := t.TempDir()
	aab := filepath.Join(dir, "app.aab")
	if err := os.WriteFile(aab, []byte("fake-aab-bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := &ValidateCmd{Track: "production", File: aab, DryRun: true}
	globals := &Globals{Output: "json", Package: "com.example.app"}
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateCmd_DryRun_BadArtifactExt(t *testing.T) {
	dir := t.TempDir()
	bad := filepath.Join(dir, "app.txt")
	if err := os.WriteFile(bad, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := &ValidateCmd{Track: "internal", File: bad, DryRun: true}
	globals := &Globals{Output: "json", Package: "com.example.app"}
	err := cmd.Run(globals)
	if err == nil {
		t.Fatal("expected failure for bad artifact extension")
	}
	if !strings.Contains(err.Error(), "validation") && !strings.Contains(err.Error(), "readiness") {
		// APIError message should mention validation/readiness
		t.Logf("error: %v", err)
	}
}

func TestValidateArtifactFile_Helpers(t *testing.T) {
	dir := t.TempDir()
	apk := filepath.Join(dir, "x.apk")
	if err := os.WriteFile(apk, []byte("apk"), 0600); err != nil {
		t.Fatal(err)
	}
	checks := validateArtifactFile(apk)
	if len(checks) < 2 {
		t.Fatalf("expected checks, got %d", len(checks))
	}
	for _, c := range checks {
		if c.Status == "fail" {
			t.Fatalf("unexpected fail: %+v", c)
		}
	}
}

// TestValidateCmd_NetworkFlagOnStruct ensures --network is wired on ValidateCmd.
func TestValidateCmd_NetworkFlagOnStruct(t *testing.T) {
	typ := reflect.TypeOf(ValidateCmd{})
	field, ok := typ.FieldByName("Network")
	if !ok {
		t.Fatal("ValidateCmd missing Network field for --network flag")
	}
	if field.Type.Kind() != reflect.Bool {
		t.Fatalf("Network field type = %v, want bool", field.Type)
	}
	tag := field.Tag.Get("help")
	if tag == "" {
		t.Fatal("Network field should have a help tag for Kong CLI")
	}
	if !strings.Contains(strings.ToLower(tag), "network") {
		t.Fatalf("Network help tag should mention network: %q", tag)
	}
}

// TestValidateCmd_DefaultSkipsNetwork verifies local-safe default: without --network,
// network.package_access is skipped and offline validation still succeeds.
func TestValidateCmd_DefaultSkipsNetwork(t *testing.T) {
	body := captureValidateStdout(t, func() error {
		cmd := &ValidateCmd{Track: "internal", DryRun: true, Network: false}
		globals := &Globals{Output: "json", Pretty: false, Package: "com.example.app"}
		return cmd.Run(globals)
	})

	report := parseValidateJSON(t, body)
	if network, _ := report["network"].(bool); network {
		t.Fatalf("network flag in report = true, want false by default")
	}
	check := findValidateCheck(t, report, "network.package_access")
	if check["status"] != "skip" {
		t.Fatalf("network.package_access status = %v, want skip without --network", check["status"])
	}
	if report["status"] != "ready" {
		t.Fatalf("status = %v, want ready for offline default", report["status"])
	}
}

// TestValidateCmd_NetworkRequiresPackage fails clearly when --network lacks --package.
func TestValidateCmd_NetworkRequiresPackage(t *testing.T) {
	var ranProbe bool
	prev := validatePackageAccessProbe
	validatePackageAccessProbe = func(ctx context.Context, globals *Globals, pkg string) error {
		ranProbe = true
		return nil
	}
	t.Cleanup(func() { validatePackageAccessProbe = prev })

	body := captureValidateStdout(t, func() error {
		cmd := &ValidateCmd{Track: "internal", DryRun: true, Network: true}
		globals := &Globals{Output: "json", Pretty: false}
		return cmd.Run(globals)
	})
	if ranProbe {
		t.Fatal("package access probe must not run when --package is empty")
	}

	report := parseValidateJSON(t, body)
	check := findValidateCheck(t, report, "network.package_access")
	if check["status"] != "fail" {
		t.Fatalf("status = %v, want fail", check["status"])
	}
	msg, _ := check["message"].(string)
	if !strings.Contains(msg, "--package") {
		t.Fatalf("message should mention --package: %q", msg)
	}
	rec, _ := check["recommendation"].(string)
	if rec == "" {
		t.Fatal("expected recommendation when network probe lacks package")
	}
	if report["status"] != "not_ready" {
		t.Fatalf("report status = %v, want not_ready", report["status"])
	}
}

// TestValidateCmd_NetworkProbePass uses an injected probe success path.
func TestValidateCmd_NetworkProbePass(t *testing.T) {
	var sawPkg string
	prevA, prevT, prevL := validatePackageAccessProbe, validateTrackProbe, validateListingProbe
	validatePackageAccessProbe = func(ctx context.Context, globals *Globals, pkg string) error {
		sawPkg = pkg
		return nil
	}
	validateTrackProbe = func(ctx context.Context, g *Globals, pkg string) ([]string, error) {
		return []string{"internal", "production"}, nil
	}
	validateListingProbe = func(ctx context.Context, g *Globals, pkg string) (string, error) {
		return "Example", nil
	}
	t.Cleanup(func() {
		validatePackageAccessProbe = prevA
		validateTrackProbe = prevT
		validateListingProbe = prevL
	})

	body := captureValidateStdout(t, func() error {
		cmd := &ValidateCmd{Track: "internal", DryRun: true, Network: true}
		globals := &Globals{Output: "json", Pretty: false, Package: "com.example.app"}
		return cmd.Run(globals)
	})
	if sawPkg != "com.example.app" {
		t.Fatalf("probe package = %q, want com.example.app", sawPkg)
	}

	report := parseValidateJSON(t, body)
	if network, _ := report["network"].(bool); !network {
		t.Fatal("report.network should be true when --network is set")
	}
	check := findValidateCheck(t, report, "network.package_access")
	if check["status"] != "pass" {
		t.Fatalf("status = %v, want pass", check["status"])
	}
	if report["status"] != "ready" {
		t.Fatalf("report status = %v, want ready", report["status"])
	}
}

// TestValidateCmd_NetworkProbeFailNoCredentials fails with a clear recommendation
// when the probe reports credential/access failure (no panic).
func TestValidateCmd_NetworkProbeFailNoCredentials(t *testing.T) {
	prevA, prevT, prevL := validatePackageAccessProbe, validateTrackProbe, validateListingProbe
	validatePackageAccessProbe = func(ctx context.Context, globals *Globals, pkg string) error {
		return errors.New("no credentials configured")
	}
	validateTrackProbe = func(ctx context.Context, g *Globals, pkg string) ([]string, error) {
		return nil, errors.New("no credentials configured")
	}
	validateListingProbe = func(ctx context.Context, g *Globals, pkg string) (string, error) {
		return "", errors.New("no credentials configured")
	}
	t.Cleanup(func() {
		validatePackageAccessProbe = prevA
		validateTrackProbe = prevT
		validateListingProbe = prevL
	})

	body := captureValidateStdout(t, func() error {
		cmd := &ValidateCmd{Track: "internal", DryRun: true, Network: true}
		globals := &Globals{Output: "json", Pretty: false, Package: "com.example.app"}
		return cmd.Run(globals)
	})

	report := parseValidateJSON(t, body)
	check := findValidateCheck(t, report, "network.package_access")
	if check["status"] != "fail" {
		t.Fatalf("status = %v, want fail", check["status"])
	}
	msg, _ := check["message"].(string)
	if !strings.Contains(msg, "no credentials") && !strings.Contains(msg, "probe failed") {
		t.Fatalf("message should describe probe failure: %q", msg)
	}
	rec, _ := check["recommendation"].(string)
	if rec == "" || !strings.Contains(strings.ToLower(rec), "credential") {
		t.Fatalf("recommendation should mention credentials: %q", rec)
	}
	if report["status"] != "not_ready" {
		t.Fatalf("report status = %v, want not_ready", report["status"])
	}
}

// TestValidateCmd_NetworkNotInvokedWithoutFlag ensures the probe is never called
// unless --network is set (offline CI safety).
func TestValidateCmd_NetworkNotInvokedWithoutFlag(t *testing.T) {
	var called bool
	prevA, prevT, prevL := validatePackageAccessProbe, validateTrackProbe, validateListingProbe
	validatePackageAccessProbe = func(ctx context.Context, globals *Globals, pkg string) error {
		called = true
		return errors.New("should not be called")
	}
	validateTrackProbe = func(ctx context.Context, g *Globals, pkg string) ([]string, error) {
		called = true
		return nil, errors.New("should not be called")
	}
	validateListingProbe = func(ctx context.Context, g *Globals, pkg string) (string, error) {
		called = true
		return "", errors.New("should not be called")
	}
	t.Cleanup(func() {
		validatePackageAccessProbe = prevA
		validateTrackProbe = prevT
		validateListingProbe = prevL
	})

	cmd := &ValidateCmd{Track: "internal", DryRun: true, Network: false}
	globals := &Globals{Output: "json", Pretty: false, Package: "com.example.app"}
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("offline validate should succeed: %v", err)
	}
	if called {
		t.Fatal("package access probe must not run without --network")
	}
}

func captureValidateStdout(t *testing.T, fn func() error) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	runErr := fn()
	_ = w.Close()
	os.Stdout = old
	out, _ := io.ReadAll(r)
	// Network fail / missing package intentionally returns readiness error after printing.
	_ = runErr
	return string(out)
}

func parseValidateJSON(t *testing.T, body string) map[string]interface{} {
	t.Helper()
	// output.Result may wrap under "data"; accept either flat or nested.
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(body), &root); err != nil {
		t.Fatalf("json unmarshal: %v\nbody: %s", err, body)
	}
	if data, ok := root["data"].(map[string]interface{}); ok {
		return data
	}
	return root
}

func findValidateCheck(t *testing.T, report map[string]interface{}, name string) map[string]interface{} {
	t.Helper()
	raw, ok := report["checks"]
	if !ok {
		t.Fatalf("report missing checks: %#v", report)
	}
	list, ok := raw.([]interface{})
	if !ok {
		t.Fatalf("checks type %T", raw)
	}
	for _, item := range list {
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if m["name"] == name {
			return m
		}
	}
	t.Fatalf("check %q not found in %#v", name, list)
	return nil
}

func TestPublishPlayCmd_DryRun(t *testing.T) {
	dir := t.TempDir()
	aab := filepath.Join(dir, "app.aab")
	if err := os.WriteFile(aab, []byte("fake"), 0600); err != nil {
		t.Fatal(err)
	}
	cmd := &PublishPlayCmd{File: aab, Track: "internal", Status: "completed", DryRun: true}
	globals := &Globals{Output: "json", Package: "com.example.app"}
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPublishPlayCmd_RequiresPackage(t *testing.T) {
	dir := t.TempDir()
	aab := filepath.Join(dir, "app.aab")
	_ = os.WriteFile(aab, []byte("fake"), 0600)
	cmd := &PublishPlayCmd{File: aab, Track: "internal", DryRun: true}
	globals := &Globals{Output: "json"}
	if err := cmd.Run(globals); err == nil {
		t.Fatal("expected package required error")
	}
}

func TestResolvePlayReleaseParams(t *testing.T) {
	tests := []struct {
		name       string
		status     string
		pct        float64
		wantStatus string
		wantFrac   float64
		wantErr    bool
	}{
		{name: "full completed", status: "completed", pct: 0, wantStatus: "completed", wantFrac: 0},
		{name: "full draft", status: "draft", pct: 0, wantStatus: "draft", wantFrac: 0},
		{name: "empty status defaults completed", status: "", pct: 0, wantStatus: "completed", wantFrac: 0},
		{name: "staged 10 percent", status: "completed", pct: 10, wantStatus: "inProgress", wantFrac: 0.10},
		{name: "staged 100 percent", status: "draft", pct: 100, wantStatus: "inProgress", wantFrac: 1.0},
		{name: "invalid status", status: "nope", pct: 0, wantErr: true},
		{name: "pct too high", status: "completed", pct: 101, wantErr: true},
		{name: "pct negative", status: "completed", pct: -1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStatus, gotFrac, err := playship.ResolveReleaseParams(tt.status, tt.pct)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if gotStatus != tt.wantStatus {
				t.Fatalf("status = %q, want %q", gotStatus, tt.wantStatus)
			}
			if gotFrac != tt.wantFrac {
				t.Fatalf("userFraction = %v, want %v", gotFrac, tt.wantFrac)
			}
		})
	}
}

func TestBuildPlayTrackRelease_AppliesStatusAndFraction(t *testing.T) {
	// Full release: completed, no fraction
	track := playship.BuildTrackRelease("production", "completed", []int64{42}, 0)
	if track.Track != "production" {
		t.Fatalf("track = %q", track.Track)
	}
	if len(track.Releases) != 1 {
		t.Fatalf("releases len = %d", len(track.Releases))
	}
	rel := track.Releases[0]
	if rel.Status != "completed" {
		t.Fatalf("status = %q, want completed (Status flag must apply)", rel.Status)
	}
	if rel.UserFraction != 0 {
		t.Fatalf("userFraction = %v, want 0 for full release", rel.UserFraction)
	}
	if len(rel.VersionCodes) != 1 || rel.VersionCodes[0] != 42 {
		t.Fatalf("versionCodes = %v, want [42]", rel.VersionCodes)
	}

	// Staged: inProgress + fraction
	staged := playship.BuildTrackRelease("production", "inProgress", []int64{99}, 0.25)
	srel := staged.Releases[0]
	if srel.Status != "inProgress" {
		t.Fatalf("staged status = %q, want inProgress", srel.Status)
	}
	if srel.UserFraction != 0.25 {
		t.Fatalf("staged userFraction = %v, want 0.25", srel.UserFraction)
	}
}

func TestBuildPublishPlayPlan_IncludesReleaseWithStatus(t *testing.T) {
	plan := playship.BuildPlan("com.example.app", "app.aab", "internal", "draft", 0, 0)
	if plan["releaseStatus"] != "draft" {
		t.Fatalf("releaseStatus = %v, want draft from --status", plan["releaseStatus"])
	}
	steps, ok := plan["steps"].([]map[string]interface{})
	if !ok {
		t.Fatalf("steps type %T", plan["steps"])
	}
	var releaseStep map[string]interface{}
	for _, s := range steps {
		if s["action"] == "release" {
			releaseStep = s
			break
		}
	}
	if releaseStep == nil {
		t.Fatal("plan missing release step that assigns artifact to track")
	}
	if releaseStep["status"] != "draft" {
		t.Fatalf("release step status = %v, want draft", releaseStep["status"])
	}
	if releaseStep["assignsArtifactToTrack"] != true {
		t.Fatal("release step must assign artifact to track")
	}

	// Staged plan forces inProgress
	stagedPlan := playship.BuildPlan("com.example.app", "app.aab", "production", "completed", 10, 0)
	if stagedPlan["releaseStatus"] != "inProgress" {
		t.Fatalf("staged releaseStatus = %v, want inProgress", stagedPlan["releaseStatus"])
	}
	if stagedPlan["userFraction"] != 0.10 {
		t.Fatalf("staged userFraction = %v, want 0.10", stagedPlan["userFraction"])
	}
}

func TestPublishPlayCmd_DryRunPlanReflectsStatus(t *testing.T) {
	dir := t.TempDir()
	aab := filepath.Join(dir, "app.aab")
	if err := os.WriteFile(aab, []byte("fake"), 0600); err != nil {
		t.Fatal(err)
	}

	// Capture stdout of dry-run to assert status appears in shipped output path.
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	cmd := &PublishPlayCmd{File: aab, Track: "beta", Status: "halted", DryRun: true}
	globals := &Globals{Output: "json", Package: "com.example.app"}
	runErr := cmd.Run(globals)
	_ = w.Close()
	os.Stdout = old
	if runErr != nil {
		t.Fatalf("run: %v", runErr)
	}
	out, _ := io.ReadAll(r)
	body := string(out)
	if !strings.Contains(body, `"status":"halted"`) && !strings.Contains(body, `"releaseStatus":"halted"`) {
		t.Fatalf("dry-run output missing halted status: %s", body)
	}
	if !strings.Contains(body, `"action":"release"`) {
		t.Fatalf("dry-run output missing release action: %s", body)
	}
}

func TestValidateCmd_NetworkTrackAndListingInjected(t *testing.T) {
	origAccess := validatePackageAccessProbe
	origTrack := validateTrackProbe
	origList := validateListingProbe
	t.Cleanup(func() {
		validatePackageAccessProbe = origAccess
		validateTrackProbe = origTrack
		validateListingProbe = origList
	})
	validatePackageAccessProbe = func(ctx context.Context, g *Globals, pkg string) error { return nil }
	validateTrackProbe = func(ctx context.Context, g *Globals, pkg string) ([]string, error) {
		return []string{"internal", "production"}, nil
	}
	validateListingProbe = func(ctx context.Context, g *Globals, pkg string) (string, error) {
		return "My App", nil
	}
	cmd := &ValidateCmd{Track: "internal", DryRun: true, Network: true}
	globals := &Globals{Output: "json", Package: "com.example.app"}
	if err := cmd.Run(globals); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidateCmd_NetworkTrackMissing(t *testing.T) {
	origAccess := validatePackageAccessProbe
	origTrack := validateTrackProbe
	origList := validateListingProbe
	t.Cleanup(func() {
		validatePackageAccessProbe = origAccess
		validateTrackProbe = origTrack
		validateListingProbe = origList
	})
	validatePackageAccessProbe = func(ctx context.Context, g *Globals, pkg string) error { return nil }
	validateTrackProbe = func(ctx context.Context, g *Globals, pkg string) ([]string, error) {
		return []string{"alpha"}, nil
	}
	validateListingProbe = func(ctx context.Context, g *Globals, pkg string) (string, error) {
		return "T", nil
	}
	cmd := &ValidateCmd{Track: "production", DryRun: true, Network: true}
	globals := &Globals{Output: "json", Package: "com.example.app"}
	if err := cmd.Run(globals); err == nil {
		t.Fatal("expected not_ready when track missing")
	}
}
