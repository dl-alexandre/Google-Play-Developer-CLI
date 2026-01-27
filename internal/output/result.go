// Package output provides structured output formatting for gpd.
package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/dl-alexandre/gpd/internal/errors"
)

// Format represents the output format type.
type Format string

const (
	FormatJSON     Format = "json"
	FormatTable    Format = "table"
	FormatMarkdown Format = "markdown"
	FormatCSV      Format = "csv" // Only for analytics/vitals
)

// Metadata contains response metadata.
type Metadata struct {
	// Core fields (always present)
	NoOp       bool     `json:"noop"`
	DurationMs int64    `json:"durationMs"`
	Services   []string `json:"services"`

	// Optional pagination fields
	RequestID     string `json:"requestId,omitempty"`
	PageToken     string `json:"pageToken,omitempty"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	HasMorePages  *bool  `json:"hasMorePages,omitempty"`

	// Optional warning/info fields
	Warnings []string `json:"warnings,omitempty"`

	// Extended fields for specific operations
	Partial          bool       `json:"partial,omitempty"`
	ScannedCount     int        `json:"scannedCount,omitempty"`
	FilteredCount    int        `json:"filteredCount,omitempty"`
	TotalAvailable   int        `json:"totalAvailable,omitempty"`
	Retries          int        `json:"retries,omitempty"`
	DataFreshnessUTC *time.Time `json:"dataFreshnessUtc,omitempty"`
	NoOpReason       string     `json:"noopReason,omitempty"`
}

// Result represents the standard JSON envelope structure.
type Result struct {
	Data     interface{}      `json:"data"`
	Error    *errors.APIError `json:"error"`
	Meta     *Metadata        `json:"meta"`
	ExitCode int              `json:"-"` // Process state only, not in JSON
}

// NewResult creates a successful result with data.
func NewResult(data interface{}) *Result {
	return &Result{
		Data:     data,
		Error:    nil,
		Meta:     &Metadata{Services: []string{}},
		ExitCode: errors.ExitSuccess,
	}
}

// NewErrorResult creates an error result.
func NewErrorResult(err *errors.APIError) *Result {
	return &Result{
		Data:     nil,
		Error:    err,
		Meta:     &Metadata{Services: []string{}},
		ExitCode: err.ExitCode(),
	}
}

// NewEmptyResult creates a result with no data (for operations that don't return data).
func NewEmptyResult() *Result {
	return &Result{
		Data:     nil,
		Error:    nil,
		Meta:     &Metadata{Services: []string{}},
		ExitCode: errors.ExitSuccess,
	}
}

// WithDuration sets the duration metadata.
func (r *Result) WithDuration(d time.Duration) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.DurationMs = d.Milliseconds()
	return r
}

// WithServices sets the services metadata.
func (r *Result) WithServices(services ...string) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.Services = services
	return r
}

// WithNoOp marks the result as a no-op with a reason.
func (r *Result) WithNoOp(reason string) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.NoOp = true
	r.Meta.NoOpReason = reason
	return r
}

// WithPagination sets pagination metadata.
func (r *Result) WithPagination(pageToken, nextPageToken string) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.PageToken = pageToken
	r.Meta.NextPageToken = nextPageToken
	hasMore := nextPageToken != ""
	r.Meta.HasMorePages = &hasMore
	return r
}

// WithWarnings adds warnings to the result.
func (r *Result) WithWarnings(warnings ...string) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.Warnings = append(r.Meta.Warnings, warnings...)
	return r
}

// WithPartial marks the result as partial with scan metadata.
func (r *Result) WithPartial(scanned, filtered, total int) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.Partial = true
	r.Meta.ScannedCount = scanned
	r.Meta.FilteredCount = filtered
	r.Meta.TotalAvailable = total
	return r
}

// WithRetries sets the retry count metadata.
func (r *Result) WithRetries(count int) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.Retries = count
	return r
}

// WithRequestID sets the request ID metadata.
func (r *Result) WithRequestID(id string) *Result {
	if r.Meta == nil {
		r.Meta = &Metadata{}
	}
	r.Meta.RequestID = id
	return r
}

// Manager handles output formatting and writing.
type Manager struct {
	format Format
	pretty bool
	fields []string
	writer io.Writer
}

// NewManager creates a new output manager.
func NewManager(w io.Writer) *Manager {
	return &Manager{
		format: FormatJSON,
		pretty: false,
		writer: w,
	}
}

// SetFormat sets the output format.
func (m *Manager) SetFormat(f Format) *Manager {
	m.format = f
	return m
}

// SetPretty enables pretty printing for JSON.
func (m *Manager) SetPretty(pretty bool) *Manager {
	m.pretty = pretty
	return m
}

// SetFields sets field projection paths.
func (m *Manager) SetFields(fields []string) *Manager {
	m.fields = fields
	return m
}

// Write formats and writes the result.
func (m *Manager) Write(r *Result) error {
	switch m.format {
	case FormatJSON:
		return m.writeJSON(r)
	case FormatTable:
		return m.writeTable(r)
	case FormatMarkdown:
		return m.writeMarkdown(r)
	case FormatCSV:
		return m.writeCSV(r)
	default:
		return m.writeJSON(r)
	}
}

func (m *Manager) writeJSON(r *Result) error {
	var data []byte
	var err error

	// Apply field projection if specified
	output := m.applyFieldProjection(r)

	if m.pretty {
		data, err = json.MarshalIndent(output, "", "  ")
	} else {
		data, err = json.Marshal(output)
	}
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(m.writer, string(data))
	return err
}

func (m *Manager) writeTable(r *Result) error {
	if r.Error != nil {
		return m.writeJSON(r) // Errors always as JSON
	}

	// Convert data to table format
	data := r.Data
	if data == nil {
		return nil
	}

	// Handle slice data
	switch v := data.(type) {
	case []interface{}:
		return m.writeTableSlice(v)
	case map[string]interface{}:
		return m.writeTableMap(v)
	default:
		// Fall back to JSON for complex types
		return m.writeJSON(r)
	}
}

func (m *Manager) writeTableSlice(data []interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Get headers from first item
	first, ok := data[0].(map[string]interface{})
	if !ok {
		return m.writeJSON(&Result{Data: data})
	}

	var headers []string
	for k := range first {
		headers = append(headers, k)
	}

	// Write headers
	if _, err := fmt.Fprintln(m.writer, strings.Join(headers, "\t")); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(m.writer, strings.Repeat("-", len(strings.Join(headers, "\t")))); err != nil {
		return err
	}

	// Write rows
	for _, item := range data {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var values []string
		for _, h := range headers {
			values = append(values, fmt.Sprintf("%v", row[h]))
		}
		if _, err := fmt.Fprintln(m.writer, strings.Join(values, "\t")); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) writeTableMap(data map[string]interface{}) error {
	for k, v := range data {
		if _, err := fmt.Fprintf(m.writer, "%s:\t%v\n", k, v); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) writeMarkdown(r *Result) error {
	if r.Error != nil {
		if _, err := fmt.Fprintf(m.writer, "## Error\n\n**Code:** %s\n\n**Message:** %s\n", r.Error.Code, r.Error.Message); err != nil {
			return err
		}
		if r.Error.Hint != "" {
			if _, err := fmt.Fprintf(m.writer, "\n**Hint:** %s\n", r.Error.Hint); err != nil {
				return err
			}
		}
		return nil
	}

	data := r.Data
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case []interface{}:
		return m.writeMarkdownTable(v)
	case map[string]interface{}:
		return m.writeMarkdownMap(v)
	default:
		return m.writeJSON(r)
	}
}

func (m *Manager) writeMarkdownTable(data []interface{}) error {
	if len(data) == 0 {
		if _, err := fmt.Fprintln(m.writer, "*No data*"); err != nil {
			return err
		}
		return nil
	}

	first, ok := data[0].(map[string]interface{})
	if !ok {
		return m.writeJSON(&Result{Data: data})
	}

	var headers []string
	for k := range first {
		headers = append(headers, k)
	}

	// Write headers
	if _, err := fmt.Fprintf(m.writer, "| %s |\n", strings.Join(headers, " | ")); err != nil {
		return err
	}
	var sep []string
	for range headers {
		sep = append(sep, "---")
	}
	if _, err := fmt.Fprintf(m.writer, "| %s |\n", strings.Join(sep, " | ")); err != nil {
		return err
	}

	// Write rows
	for _, item := range data {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var values []string
		for _, h := range headers {
			values = append(values, fmt.Sprintf("%v", row[h]))
		}
		if _, err := fmt.Fprintf(m.writer, "| %s |\n", strings.Join(values, " | ")); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) writeMarkdownMap(data map[string]interface{}) error {
	for k, v := range data {
		if _, err := fmt.Fprintf(m.writer, "- **%s:** %v\n", k, v); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) writeCSV(r *Result) error {
	if r.Error != nil {
		return m.writeJSON(r)
	}

	data := r.Data
	if data == nil {
		return nil
	}

	slice, ok := data.([]interface{})
	if !ok {
		return m.writeJSON(r)
	}

	if len(slice) == 0 {
		return nil
	}

	first, ok := slice[0].(map[string]interface{})
	if !ok {
		return m.writeJSON(r)
	}

	var headers []string
	for k := range first {
		headers = append(headers, k)
	}

	// Write CSV header
	if _, err := fmt.Fprintln(m.writer, strings.Join(headers, ",")); err != nil {
		return err
	}

	// Write rows
	for _, item := range slice {
		row, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		var values []string
		for _, h := range headers {
			val := fmt.Sprintf("%v", row[h])
			// Escape CSV values
			if strings.Contains(val, ",") || strings.Contains(val, "\"") || strings.Contains(val, "\n") {
				val = "\"" + strings.ReplaceAll(val, "\"", "\"\"") + "\""
			}
			values = append(values, val)
		}
		if _, err := fmt.Fprintln(m.writer, strings.Join(values, ",")); err != nil {
			return err
		}
	}

	return nil
}

// applyFieldProjection applies --fields projection to the result.
func (m *Manager) applyFieldProjection(r *Result) interface{} {
	if len(m.fields) == 0 {
		return r
	}

	// For simplicity, return full result if no projection
	// Full implementation would extract specific fields using dotted paths
	return r
}

// ParseFormat parses a format string into a Format type.
func ParseFormat(s string) Format {
	switch strings.ToLower(s) {
	case "json":
		return FormatJSON
	case "table":
		return FormatTable
	case "markdown", "md":
		return FormatMarkdown
	case "csv":
		return FormatCSV
	default:
		return FormatJSON
	}
}
