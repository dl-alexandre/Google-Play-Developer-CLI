package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/dl-alexandre/Google-Play-Developer-CLI/internal/config"
)

const (
	defaultAuthProfile  = "default"
	tokenKeySeparator   = "--"
	tokenMetadataSuffix = ".meta.json"
)

type StoredToken struct {
	AccessToken  string   `json:"access_token"`            // #nosec G117 -- OAuth token field, required
	RefreshToken string   `json:"refresh_token,omitempty"` // #nosec G117 -- OAuth token field, required
	TokenType    string   `json:"token_type,omitempty"`
	Expiry       string   `json:"expiry"`
	Scopes       []string `json:"scopes,omitempty"`
	Origin       string   `json:"origin,omitempty"`
	Email        string   `json:"email,omitempty"`
	KeyPath      string   `json:"keyPath,omitempty"`
	ClientID     string   `json:"clientId,omitempty"`
}

type TokenMetadata struct {
	Profile       string   `json:"profile"`
	ClientIDHash  string   `json:"clientIdHash"`
	ClientIDLast4 string   `json:"clientIdLast4,omitempty"`
	Origin        string   `json:"origin"`
	Email         string   `json:"email,omitempty"`
	Scopes        []string `json:"scopes,omitempty"`
	TokenExpiry   string   `json:"tokenExpiry,omitempty"`
	UpdatedAt     string   `json:"updatedAt"`
}

func (m *Manager) storageEnabled() bool {
	if m.storeTokensMode == "never" {
		return false
	}
	if m.storage == nil || !m.storage.Available() {
		return false
	}
	return true
}

func (m *Manager) TokenLocation() string {
	if m.storageEnabled() {
		return "secure-storage"
	}
	return "memory"
}

func (m *Manager) LoadTokenMetadata(profile string) (*TokenMetadata, error) {
	meta, _, err := findTokenMetadata(profile)
	return meta, err
}

func (m *Manager) ListProfiles() ([]TokenMetadata, error) {
	paths := config.GetPaths()
	dir := filepath.Join(paths.ConfigDir, "tokens")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []TokenMetadata{}, nil
		}
		return nil, err
	}
	byProfile := make(map[string]*TokenMetadata)
	byProfileMod := make(map[string]time.Time)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), tokenMetadataSuffix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		metaPath := filepath.Join(dir, entry.Name())
		meta, err := readTokenMetadata(metaPath)
		if err != nil || meta == nil || meta.Profile == "" {
			continue
		}
		if priorMod, ok := byProfileMod[meta.Profile]; !ok || info.ModTime().After(priorMod) {
			byProfile[meta.Profile] = meta
			byProfileMod[meta.Profile] = info.ModTime()
		}
	}
	profiles := make([]TokenMetadata, 0, len(byProfile))
	for _, meta := range byProfile {
		if meta == nil {
			continue
		}
		profiles = append(profiles, *meta)
	}
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Profile < profiles[j].Profile
	})
	return profiles, nil
}

func tokenStorageKey(profile, hash string) string {
	if hash == "" {
		return profile
	}
	return profile + tokenKeySeparator + hash
}

func clientIDHash(clientID string) (hash, last4 string) {
	if clientID == "" {
		return "", ""
	}
	hashBytes := sha256.Sum256([]byte(clientID))
	hexHash := hex.EncodeToString(hashBytes[:])
	last4 = clientID
	if len(clientID) > 4 {
		last4 = clientID[len(clientID)-4:]
	}
	return hexHash, last4
}

func tokenMetadataPath(key string) string {
	paths := config.GetPaths()
	return filepath.Join(paths.ConfigDir, "tokens", key+tokenMetadataSuffix)
}

func writeTokenMetadata(key string, metadata *TokenMetadata) error {
	paths := config.GetPaths()
	dir := filepath.Join(paths.ConfigDir, "tokens")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(tokenMetadataPath(key), data, 0600)
}

func readTokenMetadata(path string) (*TokenMetadata, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var metadata TokenMetadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		return nil, err
	}
	return &metadata, nil
}

func findTokenMetadata(profile string) (*TokenMetadata, string, error) {
	paths := config.GetPaths()
	dir := filepath.Join(paths.ConfigDir, "tokens")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", err
	}
	var selected *TokenMetadata
	var selectedKey string
	var selectedMod time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), tokenMetadataSuffix) {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			continue
		}
		metaPath := filepath.Join(dir, entry.Name())
		meta, err := readTokenMetadata(metaPath)
		if err != nil {
			continue
		}
		if meta.Profile != profile {
			continue
		}
		if selected == nil || info.ModTime().After(selectedMod) {
			selected = meta
			selectedKey = strings.TrimSuffix(entry.Name(), tokenMetadataSuffix)
			selectedMod = info.ModTime()
		}
	}
	return selected, selectedKey, nil
}

func (m *Manager) loadStoredToken(key string) (*oauth2.Token, error) {
	if !m.storageEnabled() {
		return nil, nil
	}
	data, err := m.storage.Retrieve(key)
	if err != nil {
		return nil, err
	}
	var stored StoredToken
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, err
	}
	token := &oauth2.Token{
		AccessToken:  stored.AccessToken,
		RefreshToken: stored.RefreshToken,
		TokenType:    stored.TokenType,
	}
	if stored.Expiry != "" {
		expiry, err := time.Parse(time.RFC3339, stored.Expiry)
		if err == nil {
			token.Expiry = expiry
		}
	}
	if !token.Valid() && token.RefreshToken == "" {
		return nil, nil
	}
	return token, nil
}

// normalizeProfile returns a non-empty profile name, defaulting to "default".
func normalizeProfile(profile string) string {
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return defaultAuthProfile
	}
	return profile
}

// collectProfileTokenKeys returns all storage keys that belong to profile,
// based on token metadata files under the config tokens directory.
func collectProfileTokenKeys(profile string) ([]string, error) {
	profile = normalizeProfile(profile)
	paths := config.GetPaths()
	dir := filepath.Join(paths.ConfigDir, "tokens")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	seen := make(map[string]struct{})
	var keys []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), tokenMetadataSuffix) {
			continue
		}
		metaPath := filepath.Join(dir, entry.Name())
		meta, err := readTokenMetadata(metaPath)
		if err != nil || meta == nil {
			continue
		}
		if meta.Profile != profile {
			continue
		}
		key := strings.TrimSuffix(entry.Name(), tokenMetadataSuffix)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys, nil
}

// collectAllTokenKeys returns every storage key that has a metadata file.
func collectAllTokenKeys() ([]string, error) {
	paths := config.GetPaths()
	dir := filepath.Join(paths.ConfigDir, "tokens")
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var keys []string
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), tokenMetadataSuffix) {
			continue
		}
		key := strings.TrimSuffix(entry.Name(), tokenMetadataSuffix)
		if key == "" {
			continue
		}
		keys = append(keys, key)
	}
	return keys, nil
}

// deleteTokenKey removes a token from secure storage (best-effort) and deletes
// its metadata file if present.
func (m *Manager) deleteTokenKey(key string) error {
	if key == "" {
		return nil
	}
	if m.storage != nil && m.storage.Available() {
		// Best-effort: missing keys are not fatal during cleanup.
		_ = m.storage.Delete(key)
	}
	metaPath := tokenMetadataPath(key)
	if err := os.Remove(metaPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// ClearProfile removes stored tokens and metadata for the given profile.
// If the profile is the manager's active profile, in-memory credentials are cleared.
// Succeeds even when no stored data exists for the profile.
func (m *Manager) ClearProfile(profile string) error {
	profile = normalizeProfile(profile)
	keys, err := collectProfileTokenKeys(profile)
	if err != nil {
		return err
	}
	// Always attempt the bare profile key (used when client ID hash is empty).
	bare := tokenStorageKey(profile, "")
	hasBare := false
	for _, k := range keys {
		if k == bare {
			hasBare = true
			break
		}
	}
	if !hasBare {
		keys = append(keys, bare)
	}

	for _, key := range keys {
		if err := m.deleteTokenKey(key); err != nil {
			return err
		}
	}

	if m.GetActiveProfile() == profile {
		m.Clear()
	}
	return nil
}

// ClearAllProfiles removes tokens and metadata for every stored profile and
// clears in-memory credentials.
func (m *Manager) ClearAllProfiles() error {
	keys, err := collectAllTokenKeys()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := m.deleteTokenKey(key); err != nil {
			return err
		}
	}
	m.Clear()
	return nil
}

// DeleteProfile removes a profile's stored credentials and metadata.
// Returns false for existed when no metadata was found for the profile.
// Unlike ClearProfile, callers use existed to report "not found".
func (m *Manager) DeleteProfile(profile string) (existed bool, err error) {
	profile = normalizeProfile(profile)
	keys, err := collectProfileTokenKeys(profile)
	if err != nil {
		return false, err
	}
	if len(keys) == 0 {
		// Also treat bare secure-storage entry without metadata as existing.
		if m.storage != nil && m.storage.Available() {
			if data, retrieveErr := m.storage.Retrieve(tokenStorageKey(profile, "")); retrieveErr == nil && len(data) > 0 {
				if clearErr := m.ClearProfile(profile); clearErr != nil {
					return true, clearErr
				}
				return true, nil
			}
		}
		return false, nil
	}
	if err := m.ClearProfile(profile); err != nil {
		return true, err
	}
	return true, nil
}

// ProfileExists reports whether any token metadata is stored for profile.
func (m *Manager) ProfileExists(profile string) (bool, error) {
	profile = normalizeProfile(profile)
	keys, err := collectProfileTokenKeys(profile)
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}
