package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/errors"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/output"
	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/workflow"
)

// WorkflowCmd contains declarative workflow commands.
type WorkflowCmd struct {
	Run      WorkflowRunCmd      `cmd:"" help:"Execute a workflow from a JSON file"`
	List     WorkflowListCmd     `cmd:"" help:"List available workflows and run history"`
	Show     WorkflowShowCmd     `cmd:"" help:"Show workflow definition and details"`
	Status   WorkflowStatusCmd   `cmd:"" help:"Show status of a workflow run"`
	Init     WorkflowInitCmd     `cmd:"" help:"Create a new workflow from template"`
	Logs     WorkflowLogsCmd     `cmd:"" help:"Show logs from a workflow run step"`
	Validate WorkflowValidateCmd `cmd:"" help:"Validate workflow file for errors"`
}

// WorkflowRunCmd executes a workflow from a JSON file.
type WorkflowRunCmd struct {
	File        string            `help:"Path to workflow JSON file" required:"" type:"path"`
	Resume      string            `help:"Resume a previous run by ID"`
	DryRun      bool              `help:"Validate workflow without executing"`
	Force       bool              `help:"Re-run completed steps even in resume mode"`
	Watch       bool              `help:"Watch workflow execution with real-time progress updates"`
	WatchFormat string            `help:"Watch mode output format (text, json, tui)" enum:"text,json,tui" default:"text"`
	Env         map[string]string `help:"Environment variables for workflow" placeholder:"KEY=VALUE"`
}

// Run executes a workflow.
func (cmd *WorkflowRunCmd) Run(globals *Globals) error {
	// Set up workflow state directory
	stateDir := globals.getWorkflowDir()
	stateManager := workflow.NewStateManager(stateDir)

	// Create run options
	opts := workflow.RunOptions{
		DryRun:      cmd.DryRun,
		ResumeRunID: cmd.Resume,
		Force:       cmd.Force,
		Verbose:     globals.Verbose,
		GlobalEnv:   cmd.Env,
		Watch:       cmd.Watch,
		WatchFormat: workflow.WatchFormat(cmd.WatchFormat),
	}

	// Create runner
	runner := workflow.NewRunner(stateManager, opts)

	// Execute workflow
	ctx := context.Background()
	state, err := runner.Run(ctx, cmd.File)

	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "workflow execution failed").
			WithHint(err.Error())
	}

	// Build result
	result := output.NewResult(map[string]interface{}{
		"runId":     state.RunID,
		"workflow":  state.Workflow.Name,
		"status":    state.Status,
		"steps":     len(state.Workflow.Steps),
		"completed": countCompletedSteps(state),
	}).WithServices("workflow")

	if state.Error != "" {
		result = output.NewResult(map[string]interface{}{
			"runId":       state.RunID,
			"workflow":    state.Workflow.Name,
			"status":      state.Status,
			"error":       state.Error,
			"currentStep": state.CurrentStep,
		}).WithServices("workflow")
	}

	return outputResult(result, globals.Output, globals.Pretty)
}

// WorkflowListCmd lists available workflows and run history.
type WorkflowListCmd struct {
	All bool `help:"Include run history"`
}

// Run lists workflows.
func (cmd *WorkflowListCmd) Run(globals *Globals) error {
	stateDir := globals.getWorkflowDir()
	stateManager := workflow.NewStateManager(stateDir)

	// Get workflow definitions
	workflows, err := listWorkflowDefinitions(stateDir)
	if err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to list workflows").
			WithHint(err.Error())
	}

	// Get run history if requested
	var runs []workflow.RunState
	if cmd.All {
		runs, err = stateManager.List()
		if err != nil {
			return errors.NewAPIError(errors.CodeGeneralError, "failed to list runs").
				WithHint(err.Error())
		}
	}

	// Build result
	data := map[string]interface{}{
		"workflows": workflows,
	}
	if cmd.All {
		data["runs"] = runs
	}

	result := output.NewResult(data).WithServices("workflow")
	return outputResult(result, globals.Output, globals.Pretty)
}

// WorkflowShowCmd shows workflow definition.
type WorkflowShowCmd struct {
	Workflow string `arg:"" help:"Workflow name or file path"`
}

// Run shows workflow details.
func (cmd *WorkflowShowCmd) Run(globals *Globals) error {
	// Try to parse as file path first
	var wf *workflow.Workflow
	var err error

	parser := workflow.NewParser()

	if _, statErr := os.Stat(cmd.Workflow); statErr == nil {
		// It's a file
		wf, err = parser.ParseFile(cmd.Workflow)
	} else {
		// Try to find in definitions directory
		stateDir := globals.getWorkflowDir()
		defPath := filepath.Join(stateDir, "definitions", cmd.Workflow+".json")
		wf, err = parser.ParseFile(defPath)
	}

	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, "workflow not found").
			WithHint(err.Error())
	}

	result := output.NewResult(map[string]interface{}{
		"name":        wf.Name,
		"description": wf.Description,
		"steps":       wf.Steps,
		"env":         wf.Env,
	}).WithServices("workflow")

	return outputResult(result, globals.Output, globals.Pretty)
}

// WorkflowStatusCmd shows workflow run status.
type WorkflowStatusCmd struct {
	RunID string `arg:"" help:"Run ID to check status for"`
}

// Run shows run status.
func (cmd *WorkflowStatusCmd) Run(globals *Globals) error {
	stateDir := globals.getWorkflowDir()
	stateManager := workflow.NewStateManager(stateDir)

	state, err := stateManager.Load(cmd.RunID)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, "run not found").
			WithHint(err.Error())
	}

	// Build step status list
	stepStatuses := make([]map[string]interface{}, 0, len(state.Workflow.Steps))
	for _, step := range state.Workflow.Steps {
		status := map[string]interface{}{
			"name":   step.Name,
			"status": "pending",
		}

		if result, ok := state.StepOutputs[step.Name]; ok {
			if result.ExitCode == 0 {
				status["status"] = "completed"
			} else {
				status["status"] = "failed"
				status["exitCode"] = result.ExitCode
				if result.Error != "" {
					status["error"] = result.Error
				}
			}
			status["duration"] = result.Duration.Milliseconds()
			if len(result.Data) > 0 {
				status["outputs"] = result.Data
			}
		}

		stepStatuses = append(stepStatuses, status)
	}

	result := output.NewResult(map[string]interface{}{
		"runId":       state.RunID,
		"workflow":    state.Workflow.Name,
		"status":      state.Status,
		"startedAt":   state.StartedAt,
		"finishedAt":  state.FinishedAt,
		"error":       state.Error,
		"steps":       stepStatuses,
		"currentStep": state.CurrentStep,
	}).WithServices("workflow")

	return outputResult(result, globals.Output, globals.Pretty)
}

// WorkflowInitCmd creates a new workflow from template.
type WorkflowInitCmd struct {
	Name       string `arg:"" help:"Workflow name"`
	Template   string `help:"Template to use" enum:"simple,release,cicd,custom" default:"simple"`
	File       string `help:"Output file path (defaults to <name>.json)" type:"path"`
	AppPackage string `help:"App package name for the workflow"`
}

// Run creates a new workflow file from template.
func (cmd *WorkflowInitCmd) Run(globals *Globals) error {
	// Determine output file
	outputFile := cmd.File
	if outputFile == "" {
		outputFile = cmd.Name + ".json"
	}

	// Check if file already exists
	if _, err := os.Stat(outputFile); err == nil {
		return errors.NewAPIError(errors.CodeValidationError, "workflow file already exists").
			WithHint(outputFile)
	}

	// Generate template
	var content string
	switch cmd.Template {
	case "release":
		content = generateReleaseTemplate(cmd.Name, cmd.AppPackage)
	case "cicd":
		content = generateCicdTemplate(cmd.Name, cmd.AppPackage)
	case "custom":
		content = generateCustomTemplate(cmd.Name)
	default: // simple
		content = generateSimpleTemplate(cmd.Name, cmd.AppPackage)
	}

	// Write file
	if err := os.WriteFile(outputFile, []byte(content), 0644); err != nil {
		return errors.NewAPIError(errors.CodeGeneralError, "failed to create workflow file").
			WithHint(err.Error())
	}

	result := output.NewResult(map[string]interface{}{
		"file":     outputFile,
		"template": cmd.Template,
		"name":     cmd.Name,
	}).WithNoOp(fmt.Sprintf("Created workflow: %s", outputFile))

	return outputResult(result, globals.Output, globals.Pretty)
}

// WorkflowLogsCmd shows logs from a workflow run step.
type WorkflowLogsCmd struct {
	RunID string `arg:"" help:"Run ID"`
	Step  string `arg:"" help:"Step name (omit for all steps)" optional:""`
}

// Run displays logs from a workflow run.
func (cmd *WorkflowLogsCmd) Run(globals *Globals) error {
	stateDir := globals.getWorkflowDir()
	stateManager := workflow.NewStateManager(stateDir)

	state, err := stateManager.Load(cmd.RunID)
	if err != nil {
		return errors.NewAPIError(errors.CodeValidationError, "run not found").
			WithHint(err.Error())
	}

	// If step specified, show only that step's logs
	if cmd.Step != "" {
		output, ok := state.StepOutputs[cmd.Step]
		if !ok {
			return errors.NewAPIError(errors.CodeValidationError, "step not found in run").
				WithHint(cmd.Step)
		}

		// Output directly to stdout for readability
		fmt.Printf("=== Step: %s ===\n", cmd.Step)
		fmt.Printf("Exit Code: %d\n", output.ExitCode)
		fmt.Printf("Duration: %v\n\n", output.Duration)

		if output.Stdout != "" {
			fmt.Println("--- STDOUT ---")
			fmt.Println(output.Stdout)
		}

		if output.Stderr != "" {
			fmt.Println("--- STDERR ---")
			fmt.Println(output.Stderr)
		}

		return nil
	}

	// Show all steps
	logs := make(map[string]interface{})
	for _, step := range state.Workflow.Steps {
		if output, ok := state.StepOutputs[step.Name]; ok {
			logs[step.Name] = map[string]interface{}{
				"exitCode": output.ExitCode,
				"stdout":   output.Stdout,
				"stderr":   output.Stderr,
				"duration": output.Duration.Milliseconds(),
			}
		}
	}

	result := output.NewResult(logs).WithServices("workflow")
	return outputResult(result, globals.Output, globals.Pretty)
}

// WorkflowValidateCmd validates a workflow file.
type WorkflowValidateCmd struct {
	File string `arg:"" help:"Path to workflow JSON file" required:"" type:"path"`
}

// ValidationIssue represents a single validation issue.
type ValidationIssue struct {
	Type    string `json:"type"`
	Field   string `json:"field,omitempty"`
	Step    string `json:"step,omitempty"`
	Message string `json:"message"`
}

// ValidationResult represents the complete validation result.
type ValidationResult struct {
	Valid   bool              `json:"valid"`
	Issues  []ValidationIssue `json:"issues,omitempty"`
	Summary map[string]int    `json:"summary"`
}

// Run validates a workflow file and returns detailed results.
func (cmd *WorkflowValidateCmd) Run(globals *Globals) error {
	parser := workflow.NewParser()
	result := &ValidationResult{
		Valid:  true,
		Issues: []ValidationIssue{},
		Summary: map[string]int{
			"total":    0,
			"errors":   0,
			"warnings": 0,
		},
	}

	// Parse and validate the workflow
	wf, err := parser.ParseFile(cmd.File)
	if err != nil {
		result.Valid = false
		result.Summary["total"]++
		result.Summary["errors"]++

		// Determine issue type from error
		issue := ValidationIssue{
			Type:    "error",
			Message: err.Error(),
		}

		// Try to extract more specific information
		if valErr, ok := err.(*workflow.ValidationError); ok {
			issue.Field = valErr.Field
		}

		result.Issues = append(result.Issues, issue)
	} else {
		// Perform additional validations on the parsed workflow
		cmd.validateStepNameUniqueness(wf, result)
		cmd.validateDependencies(wf, result)
		cmd.validateCircularDependencies(wf, result)
		cmd.validateCommandSyntax(wf, result)
		cmd.validateVariableInterpolation(wf, result)
	}

	// Prepare output
	outputData := map[string]interface{}{
		"file":     cmd.File,
		"valid":    result.Valid,
		"issues":   result.Issues,
		"summary":  result.Summary,
		"steps":    0,
		"workflow": "",
	}

	if wf != nil {
		outputData["steps"] = len(wf.Steps)
		outputData["workflow"] = wf.Name
	}

	resultOutput := output.NewResult(outputData).WithServices("workflow")

	// Add warnings if there are any
	if result.Summary["warnings"] > 0 {
		var warnings []string
		for _, issue := range result.Issues {
			if issue.Type == "warning" {
				warnings = append(warnings, issue.Message)
			}
		}
		resultOutput = resultOutput.WithWarnings(warnings...)
	}

	return outputResult(resultOutput, globals.Output, globals.Pretty)
}

// validateStepNameUniqueness checks for duplicate step names.
func (cmd *WorkflowValidateCmd) validateStepNameUniqueness(wf *workflow.Workflow, result *ValidationResult) {
	names := make(map[string]int) // name -> count
	for _, step := range wf.Steps {
		names[step.Name]++
	}

	for name, count := range names {
		if count > 1 {
			result.Valid = false
			result.Summary["total"]++
			result.Summary["errors"]++
			result.Issues = append(result.Issues, ValidationIssue{
				Type:    "error",
				Field:   "steps",
				Message: fmt.Sprintf("duplicate step name '%s' appears %d times", name, count),
			})
		}
	}
}

// validateDependencies checks that all dependencies reference existing steps.
func (cmd *WorkflowValidateCmd) validateDependencies(wf *workflow.Workflow, result *ValidationResult) {
	validStepNames := make(map[string]bool)
	for _, step := range wf.Steps {
		validStepNames[step.Name] = true
	}

	for _, step := range wf.Steps {
		for _, dep := range step.DependsOn {
			if !validStepNames[dep] {
				result.Valid = false
				result.Summary["total"]++
				result.Summary["errors"]++
				result.Issues = append(result.Issues, ValidationIssue{
					Type:    "error",
					Step:    step.Name,
					Field:   "dependsOn",
					Message: fmt.Sprintf("step '%s' has unknown dependency: '%s'", step.Name, dep),
				})
			}
		}
	}
}

// validateCircularDependencies detects cycles in step dependencies.
func (cmd *WorkflowValidateCmd) validateCircularDependencies(wf *workflow.Workflow, result *ValidationResult) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var check func(stepName string) error
	check = func(stepName string) error {
		visited[stepName] = true
		recStack[stepName] = true

		step, ok := wf.GetStep(stepName)
		if !ok {
			recStack[stepName] = false
			return nil
		}

		for _, dep := range step.DependsOn {
			if !visited[dep] {
				if err := check(dep); err != nil {
					return err
				}
			} else if recStack[dep] {
				return fmt.Errorf("circular dependency detected: %s -> %s", stepName, dep)
			}
		}

		recStack[stepName] = false
		return nil
	}

	for _, step := range wf.Steps {
		if !visited[step.Name] {
			if err := check(step.Name); err != nil {
				result.Valid = false
				result.Summary["total"]++
				result.Summary["errors"]++
				result.Issues = append(result.Issues, ValidationIssue{
					Type:    "error",
					Field:   "steps",
					Message: err.Error(),
				})
			}
		}
	}
}

// validateCommandSyntax checks command syntax including brace balancing.
func (cmd *WorkflowValidateCmd) validateCommandSyntax(wf *workflow.Workflow, result *ValidationResult) {
	for _, step := range wf.Steps {
		if step.Command == "" {
			result.Valid = false
			result.Summary["total"]++
			result.Summary["errors"]++
			result.Issues = append(result.Issues, ValidationIssue{
				Type:    "error",
				Step:    step.Name,
				Field:   "command",
				Message: fmt.Sprintf("step '%s' has empty command", step.Name),
			})
			continue
		}

		// Check for unbalanced braces
		openCount := 0
		for i := 0; i < len(step.Command); i++ {
			if i+1 < len(step.Command) && step.Command[i] == '$' && step.Command[i+1] == '{' {
				openCount++
			} else if step.Command[i] == '}' {
				openCount--
				if openCount < 0 {
					break
				}
			}
		}

		if openCount != 0 {
			result.Valid = false
			result.Summary["total"]++
			result.Summary["errors"]++
			result.Issues = append(result.Issues, ValidationIssue{
				Type:    "error",
				Step:    step.Name,
				Field:   "command",
				Message: fmt.Sprintf("step '%s' has unbalanced braces in command", step.Name),
			})
		}
	}
}

// validateVariableInterpolation checks if variable references can be resolved.
func (cmd *WorkflowValidateCmd) validateVariableInterpolation(wf *workflow.Workflow, result *ValidationResult) {
	validStepNames := make(map[string]bool)
	for _, step := range wf.Steps {
		validStepNames[step.Name] = true
	}

	// Collect environment variables from workflow
	workflowEnv := make(map[string]bool)
	if wf.Env != nil {
		for key := range wf.Env {
			workflowEnv[key] = true
		}
	}

	// Check each step's command and condition for variable references
	for _, step := range wf.Steps {
		// Check command
		cmd.validateVariableReferences(step.Name, "command", step.Command, validStepNames, workflowEnv, result)

		// Check condition if present
		if step.Condition != "" {
			cmd.validateVariableReferences(step.Name, "condition", step.Condition, validStepNames, workflowEnv, result)
		}

		// Check workingDir if present
		if step.WorkingDir != "" {
			cmd.validateVariableReferences(step.Name, "workingDir", step.WorkingDir, validStepNames, workflowEnv, result)
		}
	}
}

// validateVariableReferences validates ${...} references in a string value.
func (cmd *WorkflowValidateCmd) validateVariableReferences(stepName, field, value string, validStepNames, workflowEnv map[string]bool, result *ValidationResult) {
	// Extract all ${...} references
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(value, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		ref := match[1] // The content inside ${...}

		// Check for step output references: steps.<name>.<field>
		if strings.HasPrefix(ref, "steps.") {
			parts := strings.Split(ref, ".")
			if len(parts) >= 3 {
				depStepName := parts[1]
				if !validStepNames[depStepName] {
					result.Valid = false
					result.Summary["total"]++
					result.Summary["errors"]++
					result.Issues = append(result.Issues, ValidationIssue{
						Type:    "error",
						Step:    stepName,
						Field:   field,
						Message: fmt.Sprintf("step '%s' references unknown step '%s' in ${%s}", stepName, depStepName, ref),
					})
				}
			}
		}

		// Check for env references - these might be warnings since they could be set at runtime
		if strings.HasPrefix(ref, "env.") || (!strings.Contains(ref, ".") && !strings.HasPrefix(ref, "steps.")) {
			envKey := ref
			if strings.HasPrefix(ref, "env.") {
				envKey = strings.TrimPrefix(ref, "env.")
			}

			if !workflowEnv[envKey] {
				result.Summary["total"]++
				result.Summary["warnings"]++
				result.Issues = append(result.Issues, ValidationIssue{
					Type:    "warning",
					Step:    stepName,
					Field:   field,
					Message: fmt.Sprintf("step '%s' references environment variable '%s' which may not be set at runtime", stepName, envKey),
				})
			}
		}
	}
}

// getWorkflowDir returns the workflow state directory.
func (g *Globals) getWorkflowDir() string {
	if g.CacheDir != "" {
		return filepath.Join(g.CacheDir, "workflows")
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), ".gpd", "workflows")
	}
	return filepath.Join(homeDir, ".gpd", "workflows")
}

// countCompletedSteps counts successfully completed steps.
func countCompletedSteps(state *workflow.RunState) int {
	count := 0
	for _, result := range state.StepResults {
		if result.Output.ExitCode == 0 {
			count++
		}
	}
	return count
}

// listWorkflowDefinitions scans the definitions directory for workflows.
func listWorkflowDefinitions(stateDir string) ([]map[string]string, error) {
	defsDir := filepath.Join(stateDir, "definitions")

	entries, err := os.ReadDir(defsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []map[string]string{}, nil
		}
		return nil, err
	}

	var workflows []map[string]string
	parser := workflow.NewParser()

	for _, entry := range entries {
		if entry.IsDir() || !hasJSONExtension(entry.Name()) {
			continue
		}

		name := entry.Name()[:len(entry.Name())-5] // Remove .json
		path := filepath.Join(defsDir, entry.Name())

		// Parse to get metadata
		wf, err := parser.ParseFile(path)
		if err != nil {
			continue // Skip invalid files
		}

		workflows = append(workflows, map[string]string{
			"name":        name,
			"description": wf.Description,
			"steps":       fmt.Sprintf("%d", len(wf.Steps)),
			"path":        path,
		})
	}

	return workflows, nil
}

func hasJSONExtension(name string) bool {
	return len(name) > 5 && name[len(name)-5:] == ".json"
}

// Template generators
func generateSimpleTemplate(name, pkg string) string {
	if pkg == "" {
		pkg = "com.example.app"
	}
	return fmt.Sprintf(`{
  "name": "%s",
  "description": "Simple workflow: upload and release",
  "env": {
    "PACKAGE": "%s"
  },
  "steps": [
    {
      "name": "upload",
      "command": "gpd publish upload app.aab --package ${env.PACKAGE} --output json",
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "release",
      "command": "gpd publish release --package ${env.PACKAGE} --track internal --version-code ${steps.upload.versionCode}",
      "dependsOn": ["upload"]
    }
  ]
}
`, name, pkg)
}

func generateReleaseTemplate(name, pkg string) string {
	if pkg == "" {
		pkg = "com.example.app"
	}
	return fmt.Sprintf(`{
  "name": "%s",
  "description": "Production release with validation and rollout",
  "env": {
    "PACKAGE": "%s"
  },
  "steps": [
    {
      "name": "validate",
      "command": "gpd automation validate --package ${env.PACKAGE} --checks all --strict"
    },
    {
      "name": "upload",
      "command": "gpd publish upload app.aab --package ${env.PACKAGE} --output json",
      "dependsOn": ["validate"],
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "internal",
      "command": "gpd publish release --package ${env.PACKAGE} --track internal --version-code ${steps.upload.versionCode}",
      "dependsOn": ["upload"]
    }
  ]
}
`, name, pkg)
}

func generateCicdTemplate(name, pkg string) string {
	if pkg == "" {
		pkg = "com.example.app"
	}
	return fmt.Sprintf(`{
  "name": "%s",
  "description": "Complete CI/CD pipeline",
  "env": {
    "PACKAGE": "%s"
  },
  "steps": [
    {
      "name": "build",
      "type": "shell",
      "command": "./gradlew bundleRelease"
    },
    {
      "name": "upload",
      "command": "gpd publish upload app/build/outputs/bundle/release/app-release.aab --package ${env.PACKAGE} --output json",
      "dependsOn": ["build"],
      "captureOutputs": ["versionCode"]
    },
    {
      "name": "release",
      "command": "gpd publish release --package ${env.PACKAGE} --track internal --version-code ${steps.upload.versionCode}",
      "dependsOn": ["upload"]
    }
  ]
}
`, name, pkg)
}

func generateCustomTemplate(name string) string {
	return fmt.Sprintf(`{
  "name": "%s",
  "description": "Custom workflow - edit as needed",
  "env": {},
  "steps": [
    {
      "name": "step1",
      "command": "echo 'Step 1'",
      "captureOutputs": []
    },
    {
      "name": "step2",
      "command": "echo 'Step 2: ${steps.step1.stdout}'",
      "dependsOn": ["step1"]
    }
  ]
}
`, name)
}
