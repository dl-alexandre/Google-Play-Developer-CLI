package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// Parser handles workflow definition parsing and validation.
type Parser struct {
	// Add any parser configuration here
}

// NewParser creates a new workflow parser.
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile reads and parses a workflow from a JSON file.
func (p *Parser) ParseFile(path string) (*Workflow, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("workflow file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read workflow file: %w", err)
	}

	return p.Parse(data)
}

// Parse parses workflow definition from JSON bytes.
func (p *Parser) Parse(data []byte) (*Workflow, error) {
	var workflow Workflow
	if err := json.Unmarshal(data, &workflow); err != nil {
		return nil, fmt.Errorf("failed to parse workflow JSON: %w", err)
	}

	// Set default step types if not specified
	for i := range workflow.Steps {
		if workflow.Steps[i].Type == "" {
			// Try to infer type from command
			if strings.HasPrefix(workflow.Steps[i].Command, "gpd ") {
				workflow.Steps[i].Type = StepTypeGPD
			} else {
				workflow.Steps[i].Type = StepTypeShell
			}
		}
	}

	if err := workflow.Validate(); err != nil {
		return nil, err
	}

	// Check for circular dependencies
	if err := p.checkCircularDependencies(&workflow); err != nil {
		return nil, err
	}

	return &workflow, nil
}

// checkCircularDependencies detects cycles in step dependencies.
func (p *Parser) checkCircularDependencies(workflow *Workflow) error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var check func(stepName string) error
	check = func(stepName string) error {
		visited[stepName] = true
		recStack[stepName] = true

		step, ok := workflow.GetStep(stepName)
		if !ok {
			return nil // Step doesn't exist (should be caught by validation)
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

	for _, step := range workflow.Steps {
		if !visited[step.Name] {
			if err := check(step.Name); err != nil {
				return err
			}
		}
	}

	return nil
}

// TopologicalSort returns steps in dependency order.
func (p *Parser) TopologicalSort(workflow *Workflow) ([]Step, error) {
	// Build adjacency list and in-degree count
	inDegree := make(map[string]int)
	dependents := make(map[string][]string)

	// Initialize
	for _, step := range workflow.Steps {
		inDegree[step.Name] = 0
	}

	// Calculate in-degrees and dependents
	for _, step := range workflow.Steps {
		for _, dep := range step.DependsOn {
			inDegree[step.Name]++
			dependents[dep] = append(dependents[dep], step.Name)
		}
	}

	// Find all steps with no dependencies
	var queue []string
	for _, step := range workflow.Steps {
		if inDegree[step.Name] == 0 {
			queue = append(queue, step.Name)
		}
	}

	// Process steps
	var sorted []Step
	stepMap := make(map[string]Step)
	for _, step := range workflow.Steps {
		stepMap[step.Name] = step
	}

	for len(queue) > 0 {
		// Dequeue
		name := queue[0]
		queue = queue[1:]

		step, ok := stepMap[name]
		if !ok {
			continue
		}
		sorted = append(sorted, step)

		// Update in-degrees of dependents
		for _, dependent := range dependents[name] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// Check if all steps were processed
	if len(sorted) != len(workflow.Steps) {
		return nil, fmt.Errorf("dependency resolution failed - possible circular dependency")
	}

	return sorted, nil
}

// ValidateCommandSyntax performs basic validation of a command string.
func ValidateCommandSyntax(command string) error {
	if command == "" {
		return fmt.Errorf("command cannot be empty")
	}

	// Check for unbalanced braces in interpolation syntax
	openCount := strings.Count(command, "${")
	closeCount := strings.Count(command, "}")

	if openCount != closeCount {
		return fmt.Errorf("unbalanced braces in command: %s", command)
	}

	return nil
}

// MergeWorkflows merges two workflows, with the second taking precedence.
func MergeWorkflows(base, override *Workflow) *Workflow {
	if override == nil {
		return base
	}
	if base == nil {
		return override
	}

	merged := &Workflow{
		Name:        override.Name,
		Description: override.Description,
		Env:         make(map[string]string),
		Steps:       override.Steps,
	}

	// Merge env vars (override takes precedence)
	for k, v := range base.Env {
		merged.Env[k] = v
	}
	for k, v := range override.Env {
		merged.Env[k] = v
	}

	return merged
}

// ExpandTilde expands ~ to home directory in paths.
func ExpandTilde(path string) (string, error) {
	if !strings.HasPrefix(path, "~/") {
		return path, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}

	return strings.Replace(path, "~/", home+"/", 1), nil
}
