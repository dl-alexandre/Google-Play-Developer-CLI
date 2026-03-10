// Package apidrift provides comprehensive API drift detection for the gpd CLI.
// It analyzes three key aspects:
// 1. What endpoints the CLI implements
// 2. What endpoints exist in the Google API Discovery document
// 3. What endpoints are missing from the Go client library (the actual drift)
package apidrift

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const unknownStatus = "unknown"

// ImplementationDriftReport shows what the CLI implements vs what's available
type ImplementationDriftReport struct {
	Timestamp         time.Time `json:"timestamp"`
	DiscoveryURL      string    `json:"discovery_url"`
	DiscoveryRevision string    `json:"discovery_revision"`
	GoModVersion      string    `json:"go_mod_version"`
	CLIName           string    `json:"cli_name"`

	// What the CLI implements
	CLIEndpoints []EndpointInfo `json:"cli_endpoints"`
	CLITotal     int            `json:"cli_total"`

	// Drift analysis
	DriftDetected bool `json:"drift_detected"`
	DriftScore    int  `json:"drift_score"`

	// Missing from Go client but CLI tries to use (THE ACTUAL DRIFT)
	MissingInClient []EndpointDrift `json:"missing_in_client"`

	// Available in discovery but CLI doesn't implement yet
	NotImplemented []EndpointInfo `json:"not_implemented"`

	// Working correctly
	WorkingEndpoints []EndpointInfo `json:"working_endpoints"`
}

type EndpointInfo struct {
	ID          string `json:"id"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	Description string `json:"description"`
	CLILocation string `json:"cli_location,omitempty"`
}

type EndpointDrift struct {
	EndpointInfo
	Issue       string `json:"issue"`
	TypeMissing string `json:"type_missing,omitempty"`
	Suggestion  string `json:"suggestion"`
}

// ImplementationDriftDetector analyzes CLI implementation drift
type ImplementationDriftDetector struct {
	DiscoveryURL string
	CLISourceDir string
	GoModPath    string
	HTTPClient   *http.Client
}

// NewImplementationDriftDetector creates a drift detector focused on CLI implementation
func NewImplementationDriftDetector(discoveryURL, cliSourceDir, goModPath string) *ImplementationDriftDetector {
	return &ImplementationDriftDetector{
		DiscoveryURL: discoveryURL,
		CLISourceDir: cliSourceDir,
		GoModPath:    goModPath,
		HTTPClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchDiscoveryDocument retrieves the discovery API document
//
//nolint:dupl // Similar to detector.go but kept separate for clarity
func (d *ImplementationDriftDetector) FetchDiscoveryDocument() (*DiscoveryDocument, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, d.DiscoveryURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := d.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("discovery API returned status %d: %s", resp.StatusCode, string(body))
	}

	var doc DiscoveryDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("failed to decode discovery document: %w", err)
	}

	return &doc, nil
}

// ExtractDiscoveryEndpoints gets all endpoints from the discovery document
func (d *ImplementationDriftDetector) ExtractDiscoveryEndpoints(doc *DiscoveryDocument) map[string]EndpointInfo {
	endpoints := make(map[string]EndpointInfo)
	d.extractResources("", doc.Resources, endpoints)
	return endpoints
}

func (d *ImplementationDriftDetector) extractResources(prefix string, resources map[string]Resource, endpoints map[string]EndpointInfo) {
	for name, resource := range resources {
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}

		for methodName := range resource.Methods {
			method := resource.Methods[methodName]
			endpointID := fullName + "." + methodName
			endpoints[endpointID] = EndpointInfo{
				ID:          endpointID,
				Path:        method.Path,
				Method:      method.HTTPMethod,
				Description: method.Description,
			}
		}

		if len(resource.Resources) > 0 {
			d.extractResources(fullName, resource.Resources, endpoints)
		}
	}
}

// ExtractCLIEndpoints analyzes CLI source code to find implemented endpoints
func (d *ImplementationDriftDetector) ExtractCLIEndpoints() (map[string]EndpointInfo, error) {
	endpoints := make(map[string]EndpointInfo)

	// Pattern to match API calls like: svc.Monetization.Onetimeproducts.List
	// or svc.Edits.Tracks.Get
	apiCallPattern := regexp.MustCompile(`svc\.([A-Za-z]+)((?:\.[A-Za-z]+)*)\.([A-Za-z]+)\(`)

	err := filepath.Walk(d.CLISourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Validate path is within expected directory
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(d.CLISourceDir)) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		matches := apiCallPattern.FindAllStringSubmatch(string(content), -1)
		for _, match := range matches {
			if len(match) < 4 {
				continue
			}
			service := match[1]
			subServices := match[2] // e.g., ".Subscriptions.BasePlans"
			method := match[3]

			// Map to discovery endpoint ID
			endpointID := d.mapServiceToEndpoint(service, subServices, method)
			if endpointID != "" {
				endpoints[endpointID] = EndpointInfo{
					ID:          endpointID,
					CLILocation: filepath.Base(path),
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze CLI code: %w", err)
	}

	return endpoints, nil
}

// mapServiceToEndpoint converts Go service calls to discovery endpoint IDs
func (d *ImplementationDriftDetector) mapServiceToEndpoint(service, subServices, method string) string {
	// Build the endpoint ID
	parts := []string{strings.ToLower(service)}

	// Parse sub-services
	if subServices != "" {
		subs := strings.Split(strings.Trim(subServices, "."), ".")
		for _, sub := range subs {
			parts = append(parts, strings.ToLower(sub))
		}
	}

	// Add the method
	parts = append(parts, strings.ToLower(method))

	return strings.Join(parts, ".")
}

// CheckGoClientAvailability checks if types exist in the Go client
func (d *ImplementationDriftDetector) CheckGoClientAvailability(endpointID string) (available bool, reason string) {
	// Parse endpoint to determine what types are needed
	parts := strings.Split(endpointID, ".")
	if len(parts) < 2 {
		return false, unknownStatus
	}

	// Extract potential type names from endpoint
	// e.g., "monetization.onetimeproducts.list" -> "OneTimeProduct", "ListOneTimeProductsResponse"

	// Check patterns for types we know are missing
	switch {
	case strings.Contains(endpointID, "onetimeproducts"):
		// These are NEW in discovery but not in Go client v0.270.0
		return false, "OneTimeProduct types not in Go client v0.270.0"
	case strings.Contains(endpointID, "apprecovery"):
		return false, "AppRecovery types not in Go client v0.270.0"
	case strings.Contains(endpointID, "externaltransactions"):
		return false, "ExternalTransaction types not in Go client v0.270.0"
	case strings.Contains(endpointID, "generatedapks"):
		return false, "GeneratedApk types not in Go client v0.270.0"
	case strings.Contains(endpointID, "device") && strings.Contains(endpointID, "tier"):
		return false, "DeviceTierConfig types not in Go client v0.270.0"
	case strings.Contains(endpointID, "datasafety"):
		return false, "DataSafety types not in Go client v0.270.0"
	case strings.Contains(endpointID, "grants"):
		return false, "Grant types not in Go client v0.270.0"
	}

	// Default: assume available (we'll verify at compile time)
	return true, ""
}

// Detect performs implementation drift detection
func (d *ImplementationDriftDetector) Detect() (*ImplementationDriftReport, error) {
	report := &ImplementationDriftReport{
		Timestamp:    time.Now().UTC(),
		DiscoveryURL: d.DiscoveryURL,
		CLIName:      "gpd",
	}

	// Fetch discovery document
	doc, err := d.FetchDiscoveryDocument()
	if err != nil {
		return nil, err
	}

	report.DiscoveryRevision = doc.Revision

	// Get go.mod version
	report.GoModVersion = d.extractGoModVersion()

	// Extract discovery endpoints
	discoveryEndpoints := d.ExtractDiscoveryEndpoints(doc)

	// Extract CLI endpoints
	cliEndpoints, err := d.ExtractCLIEndpoints()
	if err != nil {
		return nil, err
	}

	report.CLITotal = len(cliEndpoints)

	// Analyze drift
	for id, cliInfo := range cliEndpoints {
		cliInfo.ID = id
		report.CLIEndpoints = append(report.CLIEndpoints, cliInfo)

		// Check if in discovery
		discoveryInfo, inDiscovery := discoveryEndpoints[id]

		if !inDiscovery {
			// CLI implements something not in discovery (deprecated?)
			report.MissingInClient = append(report.MissingInClient, EndpointDrift{
				EndpointInfo: cliInfo,
				Issue:        "not_in_discovery",
				Suggestion:   "Endpoint may be deprecated or removed from API",
			})
			report.DriftDetected = true
			report.DriftScore++
			continue
		}

		// Update with discovery info
		cliInfo.Path = discoveryInfo.Path
		cliInfo.Method = discoveryInfo.Method
		cliInfo.Description = discoveryInfo.Description

		// Check if available in Go client
		available, reason := d.CheckGoClientAvailability(id)
		if !available {
			report.MissingInClient = append(report.MissingInClient, EndpointDrift{
				EndpointInfo: cliInfo,
				Issue:        "missing_from_go_client",
				TypeMissing:  reason,
				Suggestion:   "Update google.golang.org/api or implement workaround",
			})
			report.DriftScore += 10 // Higher weight for this critical drift
			report.DriftDetected = true
		} else {
			report.WorkingEndpoints = append(report.WorkingEndpoints, cliInfo)
		}
	}

	// Find endpoints in discovery not implemented by CLI
	for id, info := range discoveryEndpoints {
		if _, implemented := cliEndpoints[id]; !implemented {
			report.NotImplemented = append(report.NotImplemented, info)
		}
	}

	// Sort all lists
	sort.Slice(report.CLIEndpoints, func(i, j int) bool {
		return report.CLIEndpoints[i].ID < report.CLIEndpoints[j].ID
	})
	sort.Slice(report.MissingInClient, func(i, j int) bool {
		return report.MissingInClient[i].ID < report.MissingInClient[j].ID
	})
	sort.Slice(report.NotImplemented, func(i, j int) bool {
		return report.NotImplemented[i].ID < report.NotImplemented[j].ID
	})
	sort.Slice(report.WorkingEndpoints, func(i, j int) bool {
		return report.WorkingEndpoints[i].ID < report.WorkingEndpoints[j].ID
	})

	return report, nil
}

func (d *ImplementationDriftDetector) extractGoModVersion() string {
	content, err := os.ReadFile(d.GoModPath)
	if err != nil {
		return "unknown"
	}

	re := regexp.MustCompile(`google\.golang\.org/api\s+v([\d.]+)`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) >= 2 {
		return matches[1]
	}
	return "unknown"
}

// PrintReport prints a formatted report
func (r *ImplementationDriftReport) PrintReport() {
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Println("   CLI Implementation Drift Report")
	fmt.Println("=" + strings.Repeat("=", 70))
	fmt.Printf("\nTimestamp: %s\n", r.Timestamp.Format(time.RFC3339))
	fmt.Printf("CLI: %s\n", r.CLIName)
	fmt.Printf("Go Module Version: %s\n", r.GoModVersion)
	fmt.Printf("Discovery Revision: %s\n\n", r.DiscoveryRevision)

	fmt.Printf("📊 Summary:\n")
	fmt.Printf("   CLI Endpoints:     %d\n", r.CLITotal)
	fmt.Printf("   Working:           %d\n", len(r.WorkingEndpoints))
	fmt.Printf("   Missing in Client: %d (⚠️ DRIFT)\n", len(r.MissingInClient))
	fmt.Printf("   Not Implemented:   %d\n", len(r.NotImplemented))
	fmt.Printf("   Drift Score:       %d\n\n", r.DriftScore)

	if len(r.MissingInClient) > 0 {
		fmt.Println("🚨 CRITICAL: Endpoints CLI Uses But Missing from Go Client")
		fmt.Println(strings.Repeat("-", 70))
		for i := range r.MissingInClient {
			drift := r.MissingInClient[i]
			fmt.Printf("\n   %s\n", drift.ID)
			fmt.Printf("   Location: %s\n", drift.CLILocation)
			fmt.Printf("   Issue: %s\n", drift.Issue)
			if drift.TypeMissing != "" {
				fmt.Printf("   Missing Types: %s\n", drift.TypeMissing)
			}
			fmt.Printf("   Action: %s\n", drift.Suggestion)
		}
		fmt.Println()
	}

	if len(r.WorkingEndpoints) > 0 {
		fmt.Printf("✅ Working Endpoints (%d):\n", len(r.WorkingEndpoints))
		for _, ep := range r.WorkingEndpoints {
			fmt.Printf("   • %s (%s %s)\n", ep.ID, ep.Method, ep.Path)
		}
		fmt.Println()
	}

	if len(r.NotImplemented) > 0 {
		fmt.Printf("📋 Discovery Endpoints Not Yet in CLI (%d):\n", len(r.NotImplemented))
		count := 0
		for _, ep := range r.NotImplemented {
			if count >= 10 {
				fmt.Printf("   ... and %d more\n", len(r.NotImplemented)-10)
				break
			}
			fmt.Printf("   • %s\n", ep.ID)
			count++
		}
		fmt.Println()
	}
}

// SaveReport saves the report to a JSON file
func (r *ImplementationDriftReport) SaveReport(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
