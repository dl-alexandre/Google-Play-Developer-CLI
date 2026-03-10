// Package apidrift provides tools for detecting API drift between the Google Play
// Developer API discovery document and the Go client library.
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

// DiscoveryDocument represents the Google API Discovery document structure
type DiscoveryDocument struct {
	Kind              string               `json:"kind"`
	ID                string               `json:"id"`
	Name              string               `json:"name"`
	Version           string               `json:"version"`
	Revision          string               `json:"revision"`
	Title             string               `json:"title"`
	Description       string               `json:"description"`
	DiscoveryVersion  string               `json:"discoveryVersion"`
	BaseURL           string               `json:"baseUrl"`
	BasePath          string               `json:"basePath"`
	RootURL           string               `json:"rootUrl"`
	ServicePath       string               `json:"servicePath"`
	BatchPath         string               `json:"batchPath"`
	DocumentationLink string               `json:"documentationLink"`
	Protocol          string               `json:"protocol"`
	Schemas           map[string]Schema    `json:"schemas"`
	Resources         map[string]Resource  `json:"resources"`
	Auth              Auth                 `json:"auth"`
	Parameters        map[string]Parameter `json:"parameters"`
}

// Schema represents an API schema definition
type Schema struct {
	ID          string              `json:"id"`
	Type        string              `json:"type"`
	Description string              `json:"description"`
	Properties  map[string]Property `json:"properties"`
	Required    []string            `json:"required"`
}

// Property represents a schema property
type Property struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Ref         string    `json:"$ref"`
	Format      string    `json:"format"`
	Items       *Property `json:"items"`
}

// Resource represents an API resource with methods
type Resource struct {
	Methods   map[string]Method   `json:"methods"`
	Resources map[string]Resource `json:"resources"`
}

// Method represents an API method
type Method struct {
	ID             string               `json:"id"`
	Path           string               `json:"path"`
	HTTPMethod     string               `json:"httpMethod"`
	Description    string               `json:"description"`
	Scopes         []string             `json:"scopes"`
	Parameters     map[string]Parameter `json:"parameters"`
	Request        *TypeRef             `json:"request,omitempty"`
	Response       *TypeRef             `json:"response,omitempty"`
	ParameterOrder []string             `json:"parameterOrder"`
	FlatPath       string               `json:"flatPath"`
}

// TypeRef references a type
type TypeRef struct {
	Ref string `json:"$ref"`
}

// Parameter represents a method parameter
type Parameter struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Required    bool     `json:"required,omitempty"`
	Location    string   `json:"location"`
	Default     string   `json:"default,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

// Auth represents API authentication info
type Auth struct {
	OAuth2 OAuth2 `json:"oauth2"`
}

// OAuth2 represents OAuth2 scopes
type OAuth2 struct {
	Scopes map[string]Scope `json:"scopes"`
}

// Scope represents an OAuth scope
type Scope struct {
	Description string `json:"description"`
}

// DriftReport contains the results of drift detection
type DriftReport struct {
	Timestamp         time.Time        `json:"timestamp"`
	DiscoveryURL      string           `json:"discovery_url"`
	DiscoveryRevision string           `json:"discovery_revision"`
	DiscoveryDate     time.Time        `json:"discovery_date"`
	GoModVersion      string           `json:"go_mod_version"`
	GoModDate         time.Time        `json:"go_mod_date"`
	Endpoints         EndpointAnalysis `json:"endpoints"`
	Schemas           SchemaAnalysis   `json:"schemas"`
	DriftDetected     bool             `json:"drift_detected"`
	DriftScore        int              `json:"drift_score"`
	Summary           string           `json:"summary"`
}

// EndpointAnalysis tracks endpoint comparison
type EndpointAnalysis struct {
	DiscoveryTotal   int      `json:"discovery_total"`
	ImplementedTotal int      `json:"implemented_total"`
	MissingInClient  []string `json:"missing_in_client"`
	NewInDiscovery   []string `json:"new_in_discovery"`
	Deprecated       []string `json:"deprecated"`
}

// SchemaAnalysis tracks schema comparison
type SchemaAnalysis struct {
	DiscoveryTotal   int          `json:"discovery_total"`
	ImplementedTotal int          `json:"implemented_total"`
	MissingFields    []FieldDrift `json:"missing_fields"`
}

// FieldDrift represents a missing or changed field
type FieldDrift struct {
	Schema string `json:"schema"`
	Field  string `json:"field"`
	Issue  string `json:"issue"`
}

// Detector is the main drift detection engine
type Detector struct {
	DiscoveryURL    string
	GoModPath       string
	ClientSourceDir string
	HTTPClient      *http.Client
}

// NewDetector creates a new drift detector
func NewDetector(discoveryURL, goModPath, clientSourceDir string) *Detector {
	return &Detector{
		DiscoveryURL:    discoveryURL,
		GoModPath:       goModPath,
		ClientSourceDir: clientSourceDir,
		HTTPClient:      &http.Client{Timeout: 30 * time.Second},
	}
}

// FetchDiscoveryDocument retrieves and parses the discovery API document
//
//nolint:dupl // Similar to implementation.go but kept separate for clarity
func (d *Detector) FetchDiscoveryDocument() (*DiscoveryDocument, error) {
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

// ParseGoModVersion extracts the google.golang.org/api version from go.mod
func (d *Detector) ParseGoModVersion() (string, error) {
	content, err := os.ReadFile(d.GoModPath)
	if err != nil {
		return "", fmt.Errorf("failed to read go.mod: %w", err)
	}

	// Look for google.golang.org/api version
	re := regexp.MustCompile(`google\.golang\.org/api\s+v([\d.]+)`)
	matches := re.FindStringSubmatch(string(content))
	if len(matches) < 2 {
		return "", fmt.Errorf("could not find google.golang.org/api version in go.mod")
	}

	return matches[1], nil
}

// ExtractEndpointsFromDiscovery extracts all endpoints from the discovery document
func (d *Detector) ExtractEndpointsFromDiscovery(doc *DiscoveryDocument) map[string]Method {
	endpoints := make(map[string]Method)
	d.extractResources("", doc.Resources, endpoints)
	return endpoints
}

func (d *Detector) extractResources(prefix string, resources map[string]Resource, endpoints map[string]Method) {
	for name := range resources {
		resource := resources[name]
		fullName := name
		if prefix != "" {
			fullName = prefix + "." + name
		}

		// Extract methods at this level
		for methodName := range resource.Methods {
			method := resource.Methods[methodName]
			endpointID := fullName + "." + methodName
			endpoints[endpointID] = method
		}

		// Recurse into nested resources
		if len(resource.Resources) > 0 {
			d.extractResources(fullName, resource.Resources, endpoints)
		}
	}
}

// AnalyzeClientCode scans the Go client library source to find implemented methods
func (d *Detector) AnalyzeClientCode() (map[string]bool, error) {
	implemented := make(map[string]bool)

	// Walk the client source directory
	err := filepath.Walk(d.ClientSourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Validate path is within expected directory
		if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(d.ClientSourceDir)) {
			return nil
		}

		content, err := os.ReadFile(path) //nolint:gosec // Path validated above
		if err != nil {
			return err
		}

		// Look for method definitions
		// Pattern: func (r *XXX) YYY(...)
		methodPattern := regexp.MustCompile(`func \(r \*([A-Za-z]+)\) ([A-Za-z]+)\(`)
		matches := methodPattern.FindAllStringSubmatch(string(content), -1)

		for _, match := range matches {
			if len(match) >= 3 {
				receiver := match[1]
				methodName := match[2]

				// Map receiver names to API resource names
				endpointID := d.mapReceiverToEndpoint(receiver, methodName)
				if endpointID != "" {
					implemented[endpointID] = true
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to analyze client code: %w", err)
	}

	return implemented, nil
}

// mapReceiverToEndpoint maps Go receiver types to API endpoint IDs
func (d *Detector) mapReceiverToEndpoint(receiver, method string) string {
	// Common mappings for androidpublisher API
	mappings := map[string]string{
		"EditsService":                    "edits",
		"EditsApksService":                "edits.apks",
		"EditsBundlesService":             "edits.bundles",
		"EditsDeobfuscationFilesService":  "edits.deobfuscationfiles",
		"EditsDetailsService":             "edits.details",
		"EditsExpansionFilesService":      "edits.expansionfiles",
		"EditsImagesService":              "edits.images",
		"EditsListingsService":            "edits.listings",
		"Edits testersService":            "edits.testers",
		"EditsTracksService":              "edits.tracks",
		"InappproductsService":            "inappproducts",
		"MonetizationService":             "monetization",
		"OrdersService":                   "orders",
		"PurchasesService":                "purchases",
		"PurchasesSubscriptionsService":   "purchases.subscriptions",
		"PurchasesVoidedpurchasesService": "purchases.voidedpurchases",
		"ReviewsService":                  "reviews",
		"SubscriptionsService":            "subscriptions",
		"UsersService":                    "users",
		"ApplicationsService":             "applications",
		"ApprecoveryService":              "apprecovery",
		"DeviceTierConfigsService":        "applications.deviceTierConfigs",
		"ExternaltransactionsService":     "externaltransactions",
		"GeneratedapksService":            "generatedapks",
		"GrantService":                    "grants",
		"OneTimeProductsService":          "monetization.onetimeproducts",
		"SafetyLabelsService":             "applications.dataSafety",
	}

	base, ok := mappings[receiver]
	if !ok {
		return ""
	}

	return base + "." + strings.ToLower(method)
}

// Detect performs drift detection and generates a report
func (d *Detector) Detect() (*DriftReport, error) {
	report := &DriftReport{
		Timestamp:    time.Now().UTC(),
		DiscoveryURL: d.DiscoveryURL,
	}

	// Fetch discovery document
	doc, err := d.FetchDiscoveryDocument()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch discovery document: %w", err)
	}

	report.DiscoveryRevision = doc.Revision
	report.DiscoveryDate = parseRevisionDate(doc.Revision)

	// Parse go.mod version
	goModVersion, err := d.ParseGoModVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to parse go.mod: %w", err)
	}
	report.GoModVersion = goModVersion

	// Extract endpoints from discovery
	discoveryEndpoints := d.ExtractEndpointsFromDiscovery(doc)
	report.Endpoints.DiscoveryTotal = len(discoveryEndpoints)

	// Analyze client code
	implementedEndpoints, err := d.AnalyzeClientCode()
	if err != nil {
		return nil, fmt.Errorf("failed to analyze client code: %w", err)
	}
	report.Endpoints.ImplementedTotal = len(implementedEndpoints)

	// Compare endpoints
	for endpointID := range discoveryEndpoints {
		if !implementedEndpoints[endpointID] {
			report.Endpoints.MissingInClient = append(report.Endpoints.MissingInClient, endpointID)
		}
	}

	for endpointID := range implementedEndpoints {
		if _, ok := discoveryEndpoints[endpointID]; !ok {
			report.Endpoints.Deprecated = append(report.Endpoints.Deprecated, endpointID)
		}
	}

	// Calculate drift
	report.DriftDetected = len(report.Endpoints.MissingInClient) > 0 || len(report.Endpoints.Deprecated) > 0
	report.DriftScore = len(report.Endpoints.MissingInClient) + len(report.Endpoints.Deprecated)

	// Generate summary
	report.Summary = d.generateSummary(report)

	return report, nil
}

// parseRevisionDate attempts to parse the revision date string
func parseRevisionDate(revision string) time.Time {
	// Try common formats
	formats := []string{
		"20060102",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, revision); err == nil {
			return t
		}
	}

	return time.Time{}
}

// generateSummary creates a human-readable summary of the drift
func (d *Detector) generateSummary(report *DriftReport) string {
	var parts []string

	if !report.DriftDetected {
		return fmt.Sprintf("✅ No API drift detected. All %d endpoints are implemented.", report.Endpoints.ImplementedTotal)
	}

	if len(report.Endpoints.MissingInClient) > 0 {
		parts = append(parts, fmt.Sprintf("⚠️ %d endpoints missing from Go client:", len(report.Endpoints.MissingInClient)))
		for _, ep := range report.Endpoints.MissingInClient[:minInt(5, len(report.Endpoints.MissingInClient))] {
			parts = append(parts, "  - "+ep)
		}
		if len(report.Endpoints.MissingInClient) > 5 {
			parts = append(parts, fmt.Sprintf("  ... and %d more", len(report.Endpoints.MissingInClient)-5))
		}
	}

	if len(report.Endpoints.Deprecated) > 0 {
		parts = append(parts, fmt.Sprintf("⚠️ %d deprecated endpoints in client:", len(report.Endpoints.Deprecated)))
		for _, ep := range report.Endpoints.Deprecated[:minInt(5, len(report.Endpoints.Deprecated))] {
			parts = append(parts, "  - "+ep)
		}
		if len(report.Endpoints.Deprecated) > 5 {
			parts = append(parts, fmt.Sprintf("  ... and %d more", len(report.Endpoints.Deprecated)-5))
		}
	}

	parts = append(parts, "", fmt.Sprintf("Discovery revision: %s", report.DiscoveryRevision), fmt.Sprintf("Go module version: %s", report.GoModVersion))

	return strings.Join(parts, "\n")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// SaveReport writes the report to a file
func (r *DriftReport) SaveReport(path string) error {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// PrintReport prints a formatted report to stdout
func (r *DriftReport) PrintReport() {
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Println("Google Play Developer API Drift Detection Report")
	fmt.Println("=" + strings.Repeat("=", 60))
	fmt.Printf("Timestamp: %s\n\n", r.Timestamp.Format(time.RFC3339))

	fmt.Printf("Discovery URL:      %s\n", r.DiscoveryURL)
	fmt.Printf("Discovery Revision: %s\n", r.DiscoveryRevision)
	fmt.Printf("Go Module Version:  %s\n\n", r.GoModVersion)

	fmt.Printf("Endpoints:\n")
	fmt.Printf("  Discovery Total:   %d\n", r.Endpoints.DiscoveryTotal)
	fmt.Printf("  Implemented Total: %d\n", r.Endpoints.ImplementedTotal)
	fmt.Printf("  Missing in Client: %d\n", len(r.Endpoints.MissingInClient))
	fmt.Printf("  Deprecated:        %d\n\n", len(r.Endpoints.Deprecated))

	if len(r.Endpoints.MissingInClient) > 0 {
		fmt.Println("Missing Endpoints (Discovery → Go Client):")
		sort.Strings(r.Endpoints.MissingInClient)
		for _, ep := range r.Endpoints.MissingInClient {
			fmt.Printf("  - %s\n", ep)
		}
		fmt.Println()
	}

	if len(r.Endpoints.Deprecated) > 0 {
		fmt.Println("Deprecated Endpoints (Go Client → Discovery):")
		sort.Strings(r.Endpoints.Deprecated)
		for _, ep := range r.Endpoints.Deprecated {
			fmt.Printf("  - %s\n", ep)
		}
		fmt.Println()
	}

	fmt.Printf("Drift Score: %d\n", r.DriftScore)
	if r.DriftDetected {
		fmt.Printf("Status: ⚠️ DRIFT DETECTED\n")
	} else {
		fmt.Printf("Status: ✅ No drift detected\n")
	}
	fmt.Println()
}
