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

	"github.com/dl-alexandre/gpd/internal/config"
)

const (
	defaultAuthProfile  = "default"
	tokenKeySeparator   = "--"
	tokenMetadataSuffix = ".meta.json"
)

type StoredToken struct {
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
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
	byProfile := make(map[string]TokenMetadata)
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
			byProfile[meta.Profile] = *meta
			byProfileMod[meta.Profile] = info.ModTime()
		}
	}
	profiles := make([]TokenMetadata, 0, len(byProfile))
	for _, meta := range byProfile {
		profiles = append(profiles, meta)
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
