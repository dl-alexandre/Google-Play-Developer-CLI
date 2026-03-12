// Package extensions provides the gpd extension system for installable subcommands.
package extensions

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// InstallOptions contains options for installing an extension.
type InstallOptions struct {
	Source    string        // GitHub repo (owner/repo) or local path
	Pin       bool          // Pin to specific ref
	PinnedRef string        // Tag or commit to pin to
	Force     bool          // Overwrite existing
	Timeout   time.Duration // HTTP timeout
}

// InstallResult contains information about the installed extension.
type InstallResult struct {
	Extension *Extension
	Installed bool // Whether this was a new install or update
}

// Install installs an extension from a GitHub repository or local path.
func Install(ctx context.Context, opts InstallOptions) (*InstallResult, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}

	// Determine install source type
	if isLocalPath(opts.Source) {
		return installLocal(ctx, opts)
	}

	return installFromGitHub(ctx, opts)
}

// isLocalPath checks if the source is a local filesystem path.
func isLocalPath(source string) bool {
	// Check for relative path indicators
	if strings.HasPrefix(source, ".") || strings.HasPrefix(source, "/") || strings.HasPrefix(source, "~") {
		return true
	}

	// Check if it's a valid GitHub repo format (owner/repo)
	// Pattern: alphanumeric + hyphens/underscores, single slash, no dots (no domain names)
	githubPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+$`)
	return !githubPattern.MatchString(source)
}

// installLocal installs an extension from a local directory.
func installLocal(ctx context.Context, opts InstallOptions) (*InstallResult, error) {
	path := expandPath(opts.Source)

	// Validate the local extension
	manifest, err := loadLocalManifest(path)
	if err != nil {
		return nil, fmt.Errorf("loading local extension manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return nil, fmt.Errorf("invalid extension manifest: %w", err)
	}

	// Check for conflicts with built-in commands
	if IsBuiltInCommand(manifest.Name) {
		return nil, fmt.Errorf("extension name %q conflicts with built-in command", manifest.Name)
	}

	// Check if already installed
	extDir := filepath.Join(GetExtensionsDir(), manifest.Name)
	alreadyInstalled := false
	if _, err := os.Stat(extDir); err == nil {
		if !opts.Force {
			return nil, fmt.Errorf("extension %q already installed (use --force to overwrite)", manifest.Name)
		}
		alreadyInstalled = true
		// Remove existing
		if err := os.RemoveAll(extDir); err != nil {
			return nil, fmt.Errorf("removing existing extension: %w", err)
		}
	}

	// Copy extension files
	if err := copyExtensionFiles(path, extDir, manifest); err != nil {
		return nil, fmt.Errorf("copying extension files: %w", err)
	}

	// Create metadata file
	ext := &Extension{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Bin:         manifest.DefaultBinName(),
		Source:      path,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Pinned:      opts.Pin,
		PinnedRef:   opts.PinnedRef,
		Type:        detectExtensionType(extDir, manifest),
	}

	if err := saveExtensionMetadata(ext); err != nil {
		return nil, fmt.Errorf("saving extension metadata: %w", err)
	}

	return &InstallResult{
		Extension: ext,
		Installed: !alreadyInstalled,
	}, nil
}

// installFromGitHub installs an extension from a GitHub repository.
func installFromGitHub(ctx context.Context, opts InstallOptions) (*InstallResult, error) {
	// Parse owner/repo
	parts := strings.Split(opts.Source, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid GitHub repository format: %s (expected: owner/repo)", opts.Source)
	}
	owner, repo := parts[0], parts[1]

	// Determine extension name from repo (gpd-<name> or just <name>)
	extName := repo
	if strings.HasPrefix(repo, "gpd-") {
		extName = strings.TrimPrefix(repo, "gpd-")
	}

	// Check for conflicts with built-in commands
	if IsBuiltInCommand(extName) {
		return nil, fmt.Errorf("extension name %q conflicts with built-in command", extName)
	}

	// Check if already installed
	extDir := filepath.Join(GetExtensionsDir(), extName)
	alreadyInstalled := false
	if _, err := os.Stat(extDir); err == nil {
		if !opts.Force {
			return nil, fmt.Errorf("extension %q already installed (use --force to overwrite)", extName)
		}
		alreadyInstalled = true
		if err := os.RemoveAll(extDir); err != nil {
			return nil, fmt.Errorf("removing existing extension: %w", err)
		}
	}

	// Try to install from GitHub Release first
	releaseURL, err := findReleaseAsset(ctx, opts)
	if err == nil && releaseURL != "" {
		if err := installFromRelease(ctx, opts, releaseURL, extDir, extName); err == nil {
			return &InstallResult{
				Extension: loadInstalledExtension(extName),
				Installed: !alreadyInstalled,
			}, nil
		}
		// Fall through to script install
	}

	// Fall back to cloning the repo
	if err := installFromRepoClone(ctx, opts, owner, repo, extDir, extName); err != nil {
		return nil, fmt.Errorf("installing from repository: %w", err)
	}

	return &InstallResult{
		Extension: loadInstalledExtension(extName),
		Installed: !alreadyInstalled,
	}, nil
}

// findReleaseAsset finds a matching release asset for the current platform.
func findReleaseAsset(ctx context.Context, opts InstallOptions) (string, error) {
	// Parse owner/repo from source
	parts := strings.Split(opts.Source, "/")
	if len(parts) != 2 {
		return "", fmt.Errorf("invalid GitHub repository format: %s", opts.Source)
	}
	owner, repo := parts[0], parts[1]

	// Determine API URL based on whether we're pinned to a specific ref
	var apiURL string
	if opts.Pin && opts.PinnedRef != "" {
		// Use specific tag
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, opts.PinnedRef)
	} else {
		// Use latest release
		apiURL = fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: opts.Timeout}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}

	// Set headers for GitHub API
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "gpd-cli-extension-installer")

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("querying GitHub API: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	// Check status code
	if resp.StatusCode == http.StatusNotFound {
		return "", fmt.Errorf("no releases found for %s/%s", owner, repo)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GitHub API returned HTTP %d", resp.StatusCode)
	}

	// Parse the release response
	var release struct {
		TagName string `json:"tag_name"`
		Assets  []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
			Size               int64  `json:"size"`
		} `json:"assets"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return "", fmt.Errorf("parsing release response: %w", err)
	}

	// Determine extension name from repo
	extName := repo
	if strings.HasPrefix(repo, "gpd-") {
		extName = strings.TrimPrefix(repo, "gpd-")
	}

	// Get current platform
	goos := runtime.GOOS
	goarch := runtime.GOARCH

	// Map Go arch names to common release naming conventions
	archAliases := map[string][]string{
		"amd64": {"amd64", "x86_64", "x64"},
		"arm64": {"arm64", "aarch64"},
		"386":   {"386", "i386", "x86"},
	}

	// Build list of possible platform identifiers
	var platformPatterns []string
	archs := []string{goarch}
	if aliases, ok := archAliases[goarch]; ok {
		archs = append(archs, aliases...)
	}

	for _, arch := range archs {
		// Pattern: gpd-{name}_{GOOS}_{GOARCH}.tar.gz
		platformPatterns = append(platformPatterns, fmt.Sprintf("%s_%s_%s", extName, goos, arch))
		// Pattern: {name}_{GOOS}_{GOARCH}.tar.gz
		platformPatterns = append(platformPatterns, fmt.Sprintf("_%s_%s", goos, arch))
		// Pattern: gpd-{name}-{GOOS}-{GOARCH}.tar.gz
		platformPatterns = append(platformPatterns, fmt.Sprintf("%s-%s-%s", extName, goos, arch))
	}

	// Also handle Windows-specific naming
	if goos == "windows" {
		for _, arch := range archs {
			platformPatterns = append(platformPatterns, fmt.Sprintf("%s_windows_%s", extName, arch))
			platformPatterns = append(platformPatterns, fmt.Sprintf("_%s_%s", "windows", arch))
		}
	}

	// Find matching asset
	for _, asset := range release.Assets {
		name := asset.Name
		// Only consider .tar.gz archives
		if !strings.HasSuffix(name, ".tar.gz") && !strings.HasSuffix(name, ".tgz") {
			continue
		}

		// Check if asset name matches any platform pattern
		for _, pattern := range platformPatterns {
			if strings.Contains(name, pattern) {
				return asset.BrowserDownloadURL, nil
			}
		}

		// Also try matching generic patterns
		lowerName := strings.ToLower(name)
		lowerOS := strings.ToLower(goos)
		for _, arch := range archs {
			lowerArch := strings.ToLower(arch)
			// Match patterns like: {name}_darwin_arm64.tar.gz, {name}-darwin-arm64.tar.gz
			if (strings.Contains(lowerName, lowerOS) && strings.Contains(lowerName, lowerArch)) ||
				(strings.Contains(lowerName, goos) && strings.Contains(lowerName, arch)) {
				return asset.BrowserDownloadURL, nil
			}
		}
	}

	return "", fmt.Errorf("no release asset found for platform %s/%s (extension: %s)", goos, goarch, extName)
}

// installFromRelease installs from a GitHub Release asset.
func installFromRelease(ctx context.Context, opts InstallOptions, releaseURL, extDir, extName string) error {
	client := &http.Client{Timeout: opts.Timeout}

	req, err := http.NewRequestWithContext(ctx, "GET", releaseURL, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}

	// Set headers for GitHub download
	req.Header.Set("User-Agent", "gpd-cli-extension-installer")
	req.Header.Set("Accept", "application/octet-stream")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading release: HTTP %d", resp.StatusCode)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "gpd-extension-*.tar.gz")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	// Download with progress tracking
	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return fmt.Errorf("downloading release: %w", err)
	}
	_ = tmpFile.Close()

	// Create extension directory
	if err := os.MkdirAll(extDir, 0755); err != nil {
		return fmt.Errorf("creating extension directory: %w", err)
	}

	// Extract archive
	if err := extractArchive(tmpFile.Name(), extDir); err != nil {
		return fmt.Errorf("extracting release: %w", err)
	}

	// Verify the manifest exists
	manifest, err := loadLocalManifest(extDir)
	if err != nil {
		return fmt.Errorf("loading extension manifest after extraction: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid extension manifest: %w", err)
	}

	// Verify executable exists
	binName := manifest.DefaultBinName()
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(extDir, binName)

	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("extension executable not found after extraction: %s", binPath)
	}

	// Ensure executable has proper permissions
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binPath, 0755); err != nil {
			return fmt.Errorf("setting executable permissions: %w", err)
		}
	}

	// Create extension metadata
	ext := &Extension{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Bin:         manifest.DefaultBinName(),
		Source:      opts.Source,
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Pinned:      opts.Pin,
		PinnedRef:   opts.PinnedRef,
		Type:        detectExtensionType(extDir, manifest),
	}

	if err := saveExtensionMetadata(ext); err != nil {
		return fmt.Errorf("saving extension metadata: %w", err)
	}

	return nil
}

// extractArchive extracts a tar.gz archive to the destination directory.
func extractArchive(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() { _ = gzr.Close() }()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		path := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(path, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				_ = outFile.Close()
				return err
			}
			_ = outFile.Close()
		}
	}

	return nil
}

// installFromRepoClone clones a repo and installs from it.
func installFromRepoClone(ctx context.Context, opts InstallOptions, owner, repo, extDir, extName string) error {
	// Construct repo URL
	repoURL := fmt.Sprintf("https://github.com/%s/%s.git", owner, repo)

	// Create temp directory for cloning
	tmpDir, err := os.MkdirTemp("", "gpd-extension-clone-*")
	if err != nil {
		return fmt.Errorf("creating temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	cloneDir := filepath.Join(tmpDir, "repo")

	// Clone the repository with shallow depth for speed
	cloneArgs := []string{"clone", "--depth", "1", repoURL, cloneDir}
	cmd := exec.CommandContext(ctx, "git", cloneArgs...)
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cloning repository: %w", err)
	}

	// If a specific ref is pinned, fetch and checkout that ref
	if opts.PinnedRef != "" {
		// Fetch the specific ref
		fetchCmd := exec.CommandContext(ctx, "git", "-C", cloneDir, "fetch", "--depth", "1", "origin", opts.PinnedRef)
		fetchCmd.Stderr = os.Stderr
		if err := fetchCmd.Run(); err != nil {
			return fmt.Errorf("fetching pinned ref %q: %w", opts.PinnedRef, err)
		}

		// Checkout the ref
		checkoutCmd := exec.CommandContext(ctx, "git", "-C", cloneDir, "checkout", opts.PinnedRef)
		checkoutCmd.Stderr = os.Stderr
		if err := checkoutCmd.Run(); err != nil {
			return fmt.Errorf("checking out pinned ref %q: %w", opts.PinnedRef, err)
		}
	}

	// Load and validate the manifest
	manifest, err := loadLocalManifest(cloneDir)
	if err != nil {
		return fmt.Errorf("loading extension manifest: %w", err)
	}

	if err := manifest.Validate(); err != nil {
		return fmt.Errorf("invalid extension manifest: %w", err)
	}

	// Create extension directory
	if err := os.MkdirAll(extDir, 0755); err != nil {
		return fmt.Errorf("creating extension directory: %w", err)
	}

	// Determine binary name
	binName := manifest.DefaultBinName()
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	srcBin := filepath.Join(cloneDir, binName)
	dstBin := filepath.Join(extDir, binName)

	// Check if executable exists
	if _, err := os.Stat(srcBin); os.IsNotExist(err) {
		return fmt.Errorf("extension executable not found: %s", srcBin)
	}

	// Copy executable with proper permissions
	if err := copyFile(srcBin, dstBin, 0755); err != nil {
		return fmt.Errorf("copying executable: %w", err)
	}

	// Copy manifest
	manifestSrc := filepath.Join(cloneDir, ".gpd-extension")
	manifestDst := filepath.Join(extDir, ".gpd-extension")
	if err := copyFile(manifestSrc, manifestDst, 0644); err != nil {
		return fmt.Errorf("copying manifest: %w", err)
	}

	// Create extension metadata
	ext := &Extension{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Bin:         manifest.DefaultBinName(),
		Source:      fmt.Sprintf("%s/%s", owner, repo),
		InstalledAt: time.Now(),
		UpdatedAt:   time.Now(),
		Pinned:      opts.Pin,
		PinnedRef:   opts.PinnedRef,
		Type:        detectExtensionType(extDir, manifest),
	}

	if err := saveExtensionMetadata(ext); err != nil {
		return fmt.Errorf("saving extension metadata: %w", err)
	}

	return nil
}

// loadLocalManifest loads the manifest from a local extension directory.
func loadLocalManifest(path string) (*Manifest, error) {
	manifestPath := filepath.Join(path, ".gpd-extension")

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("reading manifest: %w", err)
	}

	var manifest Manifest

	// Try JSON first
	if err := json.Unmarshal(data, &manifest); err == nil {
		return &manifest, nil
	}

	// Fall back to YAML
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("parsing manifest: %w", err)
	}

	return &manifest, nil
}

// copyExtensionFiles copies extension files from source to destination.
func copyExtensionFiles(srcDir, dstDir string, manifest *Manifest) error {
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	binName := manifest.DefaultBinName()
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	srcBin := filepath.Join(srcDir, binName)
	dstBin := filepath.Join(dstDir, binName)

	// Check if executable exists
	if _, err := os.Stat(srcBin); os.IsNotExist(err) {
		return fmt.Errorf("extension executable not found: %s", srcBin)
	}

	// Copy executable
	if err := copyFile(srcBin, dstBin, 0755); err != nil {
		return fmt.Errorf("copying executable: %w", err)
	}

	// Copy manifest
	manifestSrc := filepath.Join(srcDir, ".gpd-extension")
	manifestDst := filepath.Join(dstDir, ".gpd-extension")
	if err := copyFile(manifestSrc, manifestDst, 0644); err != nil {
		return fmt.Errorf("copying manifest: %w", err)
	}

	return nil
}

// copyFile copies a file from src to dst with the specified permissions.
func copyFile(src, dst string, perm os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }()

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return nil
}

// saveExtensionMetadata saves the extension metadata to disk.
func saveExtensionMetadata(ext *Extension) error {
	dir := filepath.Join(GetExtensionsDir(), ext.Name)
	metaPath := filepath.Join(dir, ".gpd-extension")

	data, err := yaml.Marshal(ext)
	if err != nil {
		return err
	}

	return os.WriteFile(metaPath, data, 0644)
}

// loadInstalledExtension loads an installed extension by name.
func loadInstalledExtension(name string) *Extension {
	ext, _ := LoadExtension(name)
	return ext
}

// detectExtensionType determines if an extension is binary or script-based.
func detectExtensionType(extDir string, manifest *Manifest) string {
	binName := manifest.DefaultBinName()
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	binPath := filepath.Join(extDir, binName)

	// Check if it's a script (shebang)
	file, err := os.Open(binPath)
	if err != nil {
		return "unknown"
	}
	defer func() { _ = file.Close() }()

	buf := make([]byte, 2)
	if _, err := file.Read(buf); err != nil {
		return "binary"
	}

	if string(buf) == "#!" {
		return "script"
	}

	return "binary"
}

// expandPath expands ~ to home directory.
func expandPath(path string) string {
	if strings.HasPrefix(path, "~") {
		home := getHomeDir()
		path = filepath.Join(home, path[1:])
	}
	return path
}

// IsBuiltInCommand checks if a command name is reserved for built-in commands.
func IsBuiltInCommand(name string) bool {
	builtins := []string{
		"auth", "config", "publish", "reviews", "vitals", "monitor",
		"analytics", "purchases", "monetization", "permissions", "recovery",
		"apps", "games", "integrity", "migrate", "customapp", "custom-app",
		"generatedapks", "generated-apks", "systemapks", "system-apks",
		"grouping", "version", "check-update", "completion", "maintenance",
		"bulk", "compare", "release-mgmt", "release-management", "testing",
		"automation", "extension", "help",
	}

	for _, cmd := range builtins {
		if cmd == name {
			return true
		}
	}
	return false
}
