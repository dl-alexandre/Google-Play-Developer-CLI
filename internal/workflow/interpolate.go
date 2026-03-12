package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Interpolator handles variable substitution in workflow commands.
type Interpolator struct {
	stepOutputs map[string]StepOutput
	env         map[string]string
}

// NewInterpolator creates a new interpolator with the given context.
func NewInterpolator(stepOutputs map[string]StepOutput, env map[string]string) *Interpolator {
	if env == nil {
		env = make(map[string]string)
	}
	return &Interpolator{
		stepOutputs: stepOutputs,
		env:         env,
	}
}

// Interpolate replaces all variable references in a string.
// Supported formats:
//   - ${steps.<name>.<field>} - Step output field
//   - ${env.<name>} - Environment variable
//   - ${<name>} - Environment variable (shorthand)
func (i *Interpolator) Interpolate(input string) (string, error) {
	result := input

	// Find all variable references: ${...}
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	matches := re.FindAllStringSubmatch(input, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		fullMatch := match[0]
		varName := match[1]

		value, err := i.resolveVariable(varName)
		if err != nil {
			return "", fmt.Errorf("failed to resolve %s: %w", fullMatch, err)
		}

		result = strings.Replace(result, fullMatch, value, 1)
	}

	return result, nil
}

// InterpolateSlice applies interpolation to all strings in a slice.
func (i *Interpolator) InterpolateSlice(inputs []string) ([]string, error) {
	result := make([]string, len(inputs))
	for j, input := range inputs {
		interpolated, err := i.Interpolate(input)
		if err != nil {
			return nil, err
		}
		result[j] = interpolated
	}
	return result, nil
}

// InterpolateMap applies interpolation to all values in a map.
func (i *Interpolator) InterpolateMap(inputs map[string]string) (map[string]string, error) {
	result := make(map[string]string, len(inputs))
	for k, v := range inputs {
		interpolated, err := i.Interpolate(v)
		if err != nil {
			return nil, err
		}
		result[k] = interpolated
	}
	return result, nil
}

func (i *Interpolator) resolveVariable(varName string) (string, error) {
	parts := strings.Split(varName, ".")

	if len(parts) == 0 {
		return "", fmt.Errorf("empty variable name")
	}

	switch parts[0] {
	case "steps":
		if len(parts) < 3 {
			return "", fmt.Errorf("step variable requires format: steps.<name>.<field>")
		}
		stepName := parts[1]
		fieldPath := parts[2:]
		return i.resolveStepOutput(stepName, fieldPath)

	case "env":
		if len(parts) < 2 {
			return "", fmt.Errorf("env variable requires format: env.<name>")
		}
		return i.resolveEnv(strings.Join(parts[1:], "."))

	default:
		// Try as environment variable
		return i.resolveEnv(varName)
	}
}

func (i *Interpolator) resolveStepOutput(stepName string, fieldPath []string) (string, error) {
	output, ok := i.stepOutputs[stepName]
	if !ok {
		return "", fmt.Errorf("no output found for step: %s", stepName)
	}

	// First, try to find in captured Data fields
	if len(fieldPath) == 1 {
		field := fieldPath[0]
		if value, ok := output.Data[field]; ok {
			return formatValue(value), nil
		}
	}

	// Navigate through nested data structure
	value, err := getNestedValue(output.Data, fieldPath)
	if err == nil {
		return formatValue(value), nil
	}

	// Try accessing other StepOutput fields directly
	if len(fieldPath) == 1 {
		switch fieldPath[0] {
		case "exitCode":
			return strconv.Itoa(output.ExitCode), nil
		case "stdout":
			return output.Stdout, nil
		case "stderr":
			return output.Stderr, nil
		}
	}

	return "", fmt.Errorf("field %s not found in step %s output", strings.Join(fieldPath, "."), stepName)
}

func (i *Interpolator) resolveEnv(name string) (string, error) {
	// First check workflow env
	if value, ok := i.env[name]; ok {
		return value, nil
	}

	// Then check system env
	if value, ok := os.LookupEnv(name); ok {
		return value, nil
	}

	return "", fmt.Errorf("environment variable not found: %s", name)
}

func getNestedValue(data map[string]interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return data, nil
	}

	current, ok := data[path[0]]
	if !ok {
		return nil, fmt.Errorf("field not found: %s", path[0])
	}

	for i := 1; i < len(path); i++ {
		switch v := current.(type) {
		case map[string]interface{}:
			next, ok := v[path[i]]
			if !ok {
				return nil, fmt.Errorf("field not found: %s", path[i])
			}
			current = next
		case []interface{}:
			index, err := strconv.Atoi(path[i])
			if err != nil || index < 0 || index >= len(v) {
				return nil, fmt.Errorf("invalid array index: %s", path[i])
			}
			current = v[index]
		default:
			return nil, fmt.Errorf("cannot traverse into %T", current)
		}
	}

	return current, nil
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	case nil:
		return ""
	default:
		// Try to marshal as JSON
		data, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprintf("%v", v)
		}
		return string(data)
	}
}

// ParseAndExtractJSON attempts to parse stdout as JSON and extract specified fields.
// This is used to capture outputs from gpd commands that output JSON.
func ParseAndExtractJSON(stdout string, fields []string) (map[string]interface{}, error) {
	if len(fields) == 0 {
		return nil, nil
	}

	// Try to find JSON in stdout (it might have other text around it)
	jsonStart := strings.Index(stdout, "{")
	jsonEnd := strings.LastIndex(stdout, "}")
	if jsonStart == -1 || jsonEnd == -1 || jsonStart >= jsonEnd {
		// Try to parse entire stdout
		var result map[string]interface{}
		if err := json.Unmarshal([]byte(stdout), &result); err != nil {
			return nil, fmt.Errorf("no valid JSON found in output")
		}
		return extractFields(result, fields)
	}

	jsonStr := stdout[jsonStart : jsonEnd+1]
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return extractFields(result, fields)
}

func extractFields(data map[string]interface{}, fields []string) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, field := range fields {
		parts := strings.Split(field, ".")
		value, err := getNestedValue(data, parts)
		if err != nil {
			// Field not found, skip it
			continue
		}
		result[field] = value
	}

	return result, nil
}
