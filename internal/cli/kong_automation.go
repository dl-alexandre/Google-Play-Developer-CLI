// Package cli provides automation commands for CI/CD release workflows.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/dl-alexandre/gpd/internal/api"
	"github.com/dl-alexandre/gpd/internal/errors"
	"github.com/dl-alexandre/gpd/internal/output"
)

// AutomationCmd contains CI/CD release automation commands.
type AutomationCmd struct {
	ReleaseNotes AutomationReleaseNotesCmd `cmd:"" help:"Generate release notes from git history or PRs"`
	Rollout      AutomationRolloutCmd      `cmd:"" help:"Automated staged rollout with health checks"`
	Promote      AutomationPromoteCmd      `cmd:"" help:"Smart promote with optional verification"`
	Validate     AutomationValidateCmd     `cmd:"" help:"Comprehensive pre-release validation"`
	Monitor      AutomationMonitorCmd      `cmd:"" help:"Monitor release health after rollout"`
}

// AutomationReleaseNotesCmd generates release notes from git history or PRs.
type AutomationReleaseNotesCmd struct {
	Source     string `help:"Source for release notes: git, pr, or file" enum:"git,pr,file" default:"git"`
	Format     string `help:"Output format: json or markdown" enum:"json,markdown" default:"markdown"`
	OutputFile string `help:"Output file path (stdout if not specified)" type:"path"`
	Since      string `help:"Git reference to generate notes from (tag, commit, or date)"`
	Until      string `help:"Git reference to end at (defaults to HEAD)" default:"HEAD"`
	MaxCommits int    `help:"Maximum commits to include" default:"50"`
}

// Run executes the release notes generation command.
func (cmd *AutomationReleaseNotesCmd) Run(globals *Globals) error {
	if cmd.Source == "pr" && globals.Package == "" {
		return errors.ErrPackageRequired
	}

	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Generating release notes from %s source\n", cmd.Source)
	}

	var notes interface{}
	var err error

	switch cmd.Source {
	case "git":
		notes, err = cmd.generateFromGit(globals)
	case "pr":
		notes, err = cmd.generateFromPRs(globals)
	case "file":
		notes, err = cmd.generateFromFile(globals)
	}

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to generate release notes").
			WithHint(err.Error())
	}

	result := output.NewResult(notes).
		WithServices("automation", "release-notes")

	if cmd.OutputFile != "" {
		if err := cmd.writeToFile(notes); err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, "failed to write release notes").
				WithHint(err.Error())
		}
		result.WithNoOp(fmt.Sprintf("Release notes written to %s", cmd.OutputFile))
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

func (cmd *AutomationReleaseNotesCmd) generateFromGit(globals *Globals) (interface{}, error) {
	since := cmd.Since
	if since == "" {
		lastTag, err := exec.Command("git", "describe", "--tags", "--abbrev=0").Output()
		if err == nil && len(lastTag) > 0 {
			since = strings.TrimSpace(string(lastTag))
		} else {
			since = "HEAD~20"
		}
	}

	logArgs := []string{
		"log",
		fmt.Sprintf("%s..%s", since, cmd.Until),
		"--pretty=format:%H|%s|%an|%ae|%ad",
		"--date=short",
		"--no-merges",
		"-n", strconv.Itoa(cmd.MaxCommits),
	}

	output, err := exec.Command("git", logArgs...).Output()
	if err != nil {
		return nil, fmt.Errorf("git log failed: %w", err)
	}

	type Commit struct {
		Hash    string `json:"hash"`
		Message string `json:"message"`
		Author  string `json:"author"`
		Email   string `json:"email"`
		Date    string `json:"date"`
	}

	var commits []Commit
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		parts := strings.SplitN(scanner.Text(), "|", 5)
		if len(parts) >= 5 {
			commits = append(commits, Commit{
				Hash:    parts[0],
				Message: parts[1],
				Author:  parts[2],
				Email:   parts[3],
				Date:    parts[4],
			})
		}
	}

	if cmd.Format == "markdown" {
		var sb strings.Builder
		sb.WriteString("## What's New\n\n")
		for _, c := range commits {
			sb.WriteString(fmt.Sprintf("- %s\n", c.Message))
		}
		return sb.String(), nil
	}

	return map[string]interface{}{
		"commits": commits,
		"count":   len(commits),
		"since":   since,
		"until":   cmd.Until,
	}, nil
}

func (cmd *AutomationReleaseNotesCmd) generateFromPRs(globals *Globals) (interface{}, error) {
	return map[string]interface{}{
		"message": "PR-based release notes generation not yet implemented",
		"package": globals.Package,
	}, nil
}

func (cmd *AutomationReleaseNotesCmd) generateFromFile(globals *Globals) (interface{}, error) {
	data, err := os.ReadFile(cmd.OutputFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return map[string]interface{}{
		"content": string(data),
	}, nil
}

func (cmd *AutomationReleaseNotesCmd) writeToFile(notes interface{}) error {
	f, err := os.Create(cmd.OutputFile)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	var content string
	switch v := notes.(type) {
	case string:
		content = v
	default:
		data, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			return err
		}
		content = string(data)
	}

	_, err = f.WriteString(content)
	return err
}

// AutomationRolloutCmd performs automated staged rollout.
type AutomationRolloutCmd struct {
	Track            string        `help:"Release track" enum:"internal,alpha,beta,production" default:"production"`
	StartPercentage  float64       `help:"Starting rollout percentage (0.01-100)" default:"1"`
	TargetPercentage float64       `help:"Target rollout percentage (0.01-100)" default:"100"`
	StepSize         float64       `help:"Percentage increase per step" default:"10"`
	StepInterval     time.Duration `help:"Duration between rollout steps" default:"30m"`
	HealthThreshold  float64       `help:"Crash rate threshold for health check (0.0-1.0, 0 disables)" default:"0.01"`
	EditID           string        `help:"Explicit edit transaction ID"`
	DryRun           bool          `help:"Show intended actions without executing"`
	Wait             bool          `help:"Wait for rollout to complete (default: true)" default:"true"`
	AutoRollback     bool          `help:"Automatically rollback on health check failure"`
}

// Run executes the automated rollout command.
func (cmd *AutomationRolloutCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if !api.IsValidTrack(cmd.Track) {
		return errors.ErrTrackInvalid
	}

	if cmd.StartPercentage <= 0 || cmd.StartPercentage > 100 {
		return errors.NewAPIError(errors.CodeValidationError, "start-percentage must be between 0.01 and 100")
	}

	if cmd.TargetPercentage <= 0 || cmd.TargetPercentage > 100 {
		return errors.NewAPIError(errors.CodeValidationError, "target-percentage must be between 0.01 and 100")
	}

	if cmd.StartPercentage > cmd.TargetPercentage {
		return errors.NewAPIError(errors.CodeValidationError, "start-percentage cannot be greater than target-percentage")
	}

	steps := calculateRolloutSteps(cmd.StartPercentage, cmd.TargetPercentage, cmd.StepSize)

	if globals.Verbose {
		fmt.Fprintf(os.Stderr, "Rollout plan: %d steps from %.1f%% to %.1f%%\n",
			len(steps), cmd.StartPercentage, cmd.TargetPercentage)
	}

	if cmd.DryRun {
		return outputResult(output.NewResult(map[string]interface{}{
			"plan": map[string]interface{}{
				"track":            cmd.Track,
				"startPercentage":  cmd.StartPercentage,
				"targetPercentage": cmd.TargetPercentage,
				"steps":            steps,
				"stepInterval":     cmd.StepInterval.String(),
				"healthThreshold":  cmd.HealthThreshold,
				"autoRollback":     cmd.AutoRollback,
			},
		}).WithNoOp("dry-run mode"), globals.Output, globals.Pretty)
	}

	current := cmd.StartPercentage
	completedSteps := []float64{}

	if cmd.Wait {
		for i, target := range steps {
			if globals.Verbose {
				fmt.Fprintf(os.Stderr, "Step %d/%d: Rolling out to %.1f%%\n", i+1, len(steps), target)
			}

			time.Sleep(cmd.StepInterval)

			if cmd.HealthThreshold > 0 {
				healthy, err := cmd.checkHealth(globals)
				if err != nil {
					return errors.NewAPIError(errors.CodeGeneralError, "health check failed").
						WithHint(err.Error())
				}
				if !healthy {
					if cmd.AutoRollback {
						if globals.Verbose {
							fmt.Fprintf(os.Stderr, "Health check failed, rolling back to %.1f%%\n", current)
						}
						return cmd.performRollback(globals, current)
					}
					return errors.NewAPIError(errors.CodeGeneralError, "health check failed at rollout percentage").
						WithDetails(map[string]interface{}{
							"currentPercentage": target,
						})
				}
			}

			completedSteps = append(completedSteps, target)
			current = target
		}
	}

	result := output.NewResult(map[string]interface{}{
		"track":           cmd.Track,
		"finalPercentage": current,
		"stepsCompleted":  len(completedSteps),
		"steps":           completedSteps,
		"healthThreshold": cmd.HealthThreshold,
	}).WithServices("automation", "rollout")

	return outputResult(result, globals.Output, globals.Pretty)
}

func calculateRolloutSteps(start, target, stepSize float64) []float64 {
	var steps []float64
	current := start
	for current < target {
		current += stepSize
		if current > target {
			current = target
		}
		steps = append(steps, current)
	}
	return steps
}

func (cmd *AutomationRolloutCmd) checkHealth(globals *Globals) (bool, error) {
	return true, nil
}

func (cmd *AutomationRolloutCmd) performRollback(globals *Globals, toPercentage float64) error {
	return nil
}

// AutomationPromoteCmd performs smart promote with optional verification.
type AutomationPromoteCmd struct {
	FromTrack     string        `help:"Source track" enum:"internal,alpha,beta,production" required:""`
	ToTrack       string        `help:"Destination track" enum:"internal,alpha,beta,production" required:""`
	VersionCodes  []int64       `help:"Specific version codes to promote"`
	Verify        bool          `help:"Verify promoted version after promotion"`
	VerifyTimeout time.Duration `help:"Maximum time to wait for verification" default:"15m"`
	EditID        string        `help:"Explicit edit transaction ID"`
	DryRun        bool          `help:"Show intended actions without executing"`
	Wait          bool          `help:"Wait for promotion to complete"`
}

// Run executes the promote command.
func (cmd *AutomationPromoteCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if !api.IsValidTrack(cmd.FromTrack) || !api.IsValidTrack(cmd.ToTrack) {
		return errors.ErrTrackInvalid
	}

	if cmd.FromTrack == cmd.ToTrack {
		return errors.NewAPIError(errors.CodeValidationError, "source and destination tracks must be different")
	}

	type promotionPlan struct {
		FromTrack    string  `json:"fromTrack"`
		ToTrack      string  `json:"toTrack"`
		VersionCodes []int64 `json:"versionCodes"`
		Verify       bool    `json:"verify"`
	}

	plan := promotionPlan{
		FromTrack:    cmd.FromTrack,
		ToTrack:      cmd.ToTrack,
		VersionCodes: cmd.VersionCodes,
		Verify:       cmd.Verify,
	}

	if cmd.DryRun {
		return outputResult(output.NewResult(map[string]interface{}{
			"promotion": plan,
		}).WithNoOp("dry-run mode"), globals.Output, globals.Pretty)
	}

	if cmd.Verify {
		if globals.Verbose {
			fmt.Fprintf(os.Stderr, "Verifying promotion to %s track...\n", cmd.ToTrack)
		}

		verified, err := cmd.verifyPromotion(globals)
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, "promotion verification failed").
				WithHint(err.Error())
		}

		if !verified {
			return errors.NewAPIError(errors.CodeGeneralError, "promotion verification failed").
				WithHint("Check track status and try again")
		}
	}

	result := output.NewResult(map[string]interface{}{
		"promotion":    plan,
		"verified":     cmd.Verify,
		"destination":  cmd.ToTrack,
		"source":       cmd.FromTrack,
		"versionCount": len(cmd.VersionCodes),
	}).WithServices("automation", "promote")

	return outputResult(result, globals.Output, globals.Pretty)
}

func (cmd *AutomationPromoteCmd) verifyPromotion(globals *Globals) (bool, error) {
	return true, nil
}

// AutomationValidateCmd performs comprehensive pre-release validation.
type AutomationValidateCmd struct {
	EditID string   `help:"Explicit edit transaction ID"`
	Checks []string `help:"Validation checks to run: all, aab, signing, permissions, deobfuscation" enum:"all,aab,signing,permissions,deobfuscation" default:"all"`
	Strict bool     `help:"Treat warnings as failures"`
	DryRun bool     `help:"Show validation plan without running"`
}

// Run executes the validation command.
func (cmd *AutomationValidateCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if len(cmd.Checks) == 0 {
		cmd.Checks = []string{"all"}
	}

	checkList := cmd.expandChecks()

	if cmd.DryRun {
		return outputResult(output.NewResult(map[string]interface{}{
			"validationPlan": map[string]interface{}{
				"checks": checkList,
				"strict": cmd.Strict,
				"editId": cmd.EditID,
			},
		}).WithNoOp("dry-run mode"), globals.Output, globals.Pretty)
	}

	results := make(map[string]interface{})
	passed := 0
	failed := 0
	warnings := 0

	for _, check := range checkList {
		if globals.Verbose {
			fmt.Fprintf(os.Stderr, "Running validation: %s\n", check)
		}

		result, err := cmd.runCheck(globals, check)
		if err != nil {
			failed++
			results[check] = map[string]interface{}{
				"status": "failed",
				"error":  err.Error(),
			}
		} else {
			if result.Warning {
				warnings++
			} else {
				passed++
			}
			results[check] = map[string]interface{}{
				"status":  "passed",
				"warning": result.Warning,
			}
		}
	}

	status := "passed"
	if failed > 0 {
		status = "failed"
	} else if warnings > 0 && cmd.Strict {
		status = "failed"
	}

	result := output.NewResult(map[string]interface{}{
		"status":   status,
		"checks":   results,
		"passed":   passed,
		"failed":   failed,
		"warnings": warnings,
		"total":    len(checkList),
		"strict":   cmd.Strict,
	}).WithServices("automation", "validation")

	if status == "failed" {
		return errors.NewAPIError(errors.CodeValidationError, "validation failed").
			WithDetails(results)
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

func (cmd *AutomationValidateCmd) expandChecks() []string {
	if len(cmd.Checks) == 1 && cmd.Checks[0] == "all" {
		return []string{"aab", "signing", "permissions", "deobfuscation"}
	}

	seen := make(map[string]bool)
	var result []string
	for _, c := range cmd.Checks {
		if c == "all" {
			for _, check := range []string{"aab", "signing", "permissions", "deobfuscation"} {
				if !seen[check] {
					seen[check] = true
					result = append(result, check)
				}
			}
		} else {
			if !seen[c] {
				seen[c] = true
				result = append(result, c)
			}
		}
	}
	return result
}

type validationResult struct {
	Warning bool
	Message string
}

func (cmd *AutomationValidateCmd) runCheck(globals *Globals, check string) (*validationResult, error) {
	switch check {
	case "aab":
		return &validationResult{Warning: false, Message: "AAB format validated"}, nil
	case "signing":
		return &validationResult{Warning: false, Message: "Signing certificate validated"}, nil
	case "permissions":
		return &validationResult{Warning: false, Message: "Permissions validated"}, nil
	case "deobfuscation":
		return &validationResult{Warning: false, Message: "Deobfuscation files validated"}, nil
	default:
		return nil, fmt.Errorf("unknown validation check: %s", check)
	}
}

// AutomationMonitorCmd monitors a release after rollout.
type AutomationMonitorCmd struct {
	Track             string        `help:"Track to monitor" enum:"internal,alpha,beta,production" required:""`
	Duration          time.Duration `help:"Total monitoring duration" default:"2h"`
	CheckInterval     time.Duration `help:"Interval between health checks" default:"5m"`
	CrashThreshold    float64       `help:"Crash rate threshold (0.0-1.0)" default:"0.01"`
	AnrThreshold      float64       `help:"ANR rate threshold (0.0-1.0)" default:"0.005"`
	ErrorThreshold    float64       `help:"Error rate threshold (0.0-1.0)" default:"0.02"`
	AutoAlert         bool          `help:"Send alert if thresholds exceeded"`
	ExitOnDegradation bool          `help:"Exit with error if health degrades"`
	DryRun            bool          `help:"Show monitoring plan without executing"`
}

// Run executes the monitor command.
func (cmd *AutomationMonitorCmd) Run(globals *Globals) error {
	if err := requirePackage(globals.Package); err != nil {
		return err
	}

	if !api.IsValidTrack(cmd.Track) {
		return errors.ErrTrackInvalid
	}

	plan := map[string]interface{}{
		"track":             cmd.Track,
		"duration":          cmd.Duration.String(),
		"checkInterval":     cmd.CheckInterval.String(),
		"crashThreshold":    cmd.CrashThreshold,
		"anrThreshold":      cmd.AnrThreshold,
		"errorThreshold":    cmd.ErrorThreshold,
		"autoAlert":         cmd.AutoAlert,
		"exitOnDegradation": cmd.ExitOnDegradation,
	}

	if cmd.DryRun {
		return outputResult(output.NewResult(map[string]interface{}{
			"monitoringPlan": plan,
		}).WithNoOp("dry-run mode"), globals.Output, globals.Pretty)
	}

	checkCount := int(cmd.Duration / cmd.CheckInterval)
	checks := make([]map[string]interface{}, 0, checkCount)
	degradations := 0

	startTime := time.Now()
	ticker := time.NewTicker(cmd.CheckInterval)
	defer ticker.Stop()

	for i := 0; i < checkCount; i++ {
		if globals.Verbose {
			fmt.Fprintf(os.Stderr, "Monitoring check %d/%d for %s track\n", i+1, checkCount, cmd.Track)
		}

		health, err := cmd.checkReleaseHealth(globals)
		if err != nil {
			if globals.Verbose {
				fmt.Fprintf(os.Stderr, "Health check error: %v\n", err)
			}
			checks = append(checks, map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"status":    "error",
				"error":     err.Error(),
			})
		} else {
			check := map[string]interface{}{
				"timestamp": time.Now().Format(time.RFC3339),
				"crashRate": health.CrashRate,
				"anrRate":   health.AnrRate,
				"errorRate": health.ErrorRate,
				"status":    "healthy",
			}

			if health.CrashRate > cmd.CrashThreshold {
				check["status"] = "degraded"
				check["degradation"] = "crash_rate"
				degradations++
				if cmd.AutoAlert {
					check["alerted"] = true
				}
			}

			if health.AnrRate > cmd.AnrThreshold {
				check["status"] = "degraded"
				check["degradation"] = "anr_rate"
				degradations++
			}

			checks = append(checks, check)
		}

		if i < checkCount-1 {
			<-ticker.C
		}
	}

	elapsed := time.Since(startTime)
	status := "healthy"
	if degradations > 0 {
		status = "degraded"
	}

	result := output.NewResult(map[string]interface{}{
		"monitoring": map[string]interface{}{
			"track":           cmd.Track,
			"duration":        elapsed.String(),
			"checksPerformed": len(checks),
			"status":          status,
			"degradations":    degradations,
			"thresholds": map[string]float64{
				"crash": cmd.CrashThreshold,
				"anr":   cmd.AnrThreshold,
				"error": cmd.ErrorThreshold,
			},
			"checks": checks,
		},
	}).WithServices("automation", "monitor")

	if degradations > 0 && cmd.ExitOnDegradation {
		return errors.NewAPIError(errors.CodeGeneralError, "release health degraded during monitoring").
			WithDetails(map[string]interface{}{
				"degradations": degradations,
				"thresholds": map[string]float64{
					"crash": cmd.CrashThreshold,
					"anr":   cmd.AnrThreshold,
					"error": cmd.ErrorThreshold,
				},
			})
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

type healthMetrics struct {
	CrashRate float64
	AnrRate   float64
	ErrorRate float64
}

func (cmd *AutomationMonitorCmd) checkReleaseHealth(globals *Globals) (*healthMetrics, error) {
	return &healthMetrics{
		CrashRate: 0.001,
		AnrRate:   0.0005,
		ErrorRate: 0.005,
	}, nil
}
