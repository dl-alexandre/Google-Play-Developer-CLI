// Package edits provides edit transaction management for gpd.
package edits

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google-play-cli/gpd/internal/config"
	"github.com/google-play-cli/gpd/internal/errors"
)

// Edit represents an edit transaction.
type Edit struct {
	Handle      string    `json:"handle"`   // Local name from --edit-id
	ServerID    string    `json:"serverId"` // Actual edit ID from Google
	PackageName string    `json:"packageName"`
	CreatedAt   time.Time `json:"createdAt"`
	LastUsedAt  time.Time `json:"lastUsedAt"`
	State       EditState `json:"state"`
}

// EditState represents the state of an edit.
type EditState string

const (
	StateOpen      EditState = "open"
	StateCommitted EditState = "committed"
	StateAborted   EditState = "aborted"
)

// Manager handles edit transactions.
type Manager struct {
	editsDir  string
	cacheDir  string
	mu        sync.Mutex
	lockFiles map[string]*LockFile
}

// NewManager creates a new edit manager.
func NewManager() *Manager {
	paths := config.GetPaths()
	return &Manager{
		editsDir:  filepath.Join(paths.ConfigDir, "edits"),
		cacheDir:  filepath.Join(paths.CacheDir, "artifacts"),
		lockFiles: make(map[string]*LockFile),
	}
}

// LockFile represents a file lock for concurrent access protection.
type LockFile struct {
	PID       int       `json:"pid"`
	Hostname  string    `json:"hostname"`
	CreatedAt time.Time `json:"createdAt"`
	Command   string    `json:"command"`
	Heartbeat time.Time `json:"heartbeat,omitempty"`
}

const (
	lockTimeout      = 30 * time.Second
	staleLockAge     = 4 * time.Hour
	lockPollInterval = 100 * time.Millisecond
)

// AcquireLock acquires a lock for the given package.
func (m *Manager) AcquireLock(ctx context.Context, packageName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.editsDir, 0700); err != nil {
		return err
	}

	lockPath := filepath.Join(m.editsDir, packageName+".lock")
	hostname, _ := os.Hostname()

	lockData := &LockFile{
		PID:       os.Getpid(),
		Hostname:  hostname,
		CreatedAt: time.Now(),
		Command:   os.Args[0],
	}

	deadline := time.Now().Add(lockTimeout)
	for time.Now().Before(deadline) {
		// Try to create lock file atomically
		if m.tryAcquireLock(lockPath, lockData, hostname) {
			m.lockFiles[packageName] = lockData
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(lockPollInterval):
			continue
		}
	}

	return errors.ErrFileLockTimeout
}

func (m *Manager) tryAcquireLock(lockPath string, newLock *LockFile, hostname string) bool {
	// Check existing lock
	data, err := os.ReadFile(lockPath)
	if err == nil {
		var existing LockFile
		if json.Unmarshal(data, &existing) == nil {
			// Check if lock is stale
			if m.isLockStale(&existing, hostname) {
				// Remove stale lock
				os.Remove(lockPath)
			} else {
				return false
			}
		}
	}

	// Try to create new lock atomically
	lockData, _ := json.Marshal(newLock)
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return false
	}
	defer f.Close()

	_, err = f.Write(lockData)
	return err == nil
}

func (m *Manager) isLockStale(lock *LockFile, currentHostname string) bool {
	// Different hostname AND old enough
	if lock.Hostname != currentHostname && time.Since(lock.CreatedAt) > staleLockAge {
		return true
	}

	// Same hostname but process is dead
	if lock.Hostname == currentHostname && !isProcessAlive(lock.PID) {
		return true
	}

	return false
}

// isProcessAlive checks if a process is still running (best effort).
func isProcessAlive(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Unix, FindProcess always succeeds, so we need to send signal 0
	// On Windows, FindProcess fails if process doesn't exist
	err = process.Signal(os.Signal(nil))
	return err == nil
}

// ReleaseLock releases the lock for the given package.
func (m *Manager) ReleaseLock(packageName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	lockPath := filepath.Join(m.editsDir, packageName+".lock")
	delete(m.lockFiles, packageName)
	return os.Remove(lockPath)
}

// SaveEdit persists an edit mapping to disk.
func (m *Manager) SaveEdit(edit *Edit) error {
	if err := os.MkdirAll(m.editsDir, 0700); err != nil {
		return err
	}

	path := filepath.Join(m.editsDir, edit.PackageName+".json")
	data, err := json.MarshalIndent(edit, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

// LoadEdit loads an edit mapping from disk.
func (m *Manager) LoadEdit(packageName string) (*Edit, error) {
	path := filepath.Join(m.editsDir, packageName+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var edit Edit
	if err := json.Unmarshal(data, &edit); err != nil {
		return nil, err
	}

	return &edit, nil
}

// DeleteEdit removes an edit mapping from disk.
func (m *Manager) DeleteEdit(packageName string) error {
	path := filepath.Join(m.editsDir, packageName+".json")
	return os.Remove(path)
}

// CacheEntry represents a cached artifact entry.
type CacheEntry struct {
	SHA256    string    `json:"sha256"`
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	CachedAt  time.Time `json:"cachedAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

const cacheTTL = 24 * time.Hour

// GetCachedArtifact checks if an artifact is cached.
func (m *Manager) GetCachedArtifact(packageName, artifactPath string) (*CacheEntry, error) {
	hash, err := m.hashFile(artifactPath)
	if err != nil {
		return nil, err
	}

	cachePath := filepath.Join(m.cacheDir, packageName, hash+".json")
	data, err := os.ReadFile(cachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		os.Remove(cachePath)
		return nil, nil
	}

	return &entry, nil
}

// CacheArtifact caches an artifact.
func (m *Manager) CacheArtifact(packageName, artifactPath string, versionCode int64) error {
	hash, err := m.hashFile(artifactPath)
	if err != nil {
		return err
	}

	cacheDir := filepath.Join(m.cacheDir, packageName)
	if err := os.MkdirAll(cacheDir, 0700); err != nil {
		return err
	}

	info, err := os.Stat(artifactPath)
	if err != nil {
		return err
	}

	entry := &CacheEntry{
		SHA256:    hash,
		Path:      artifactPath,
		Size:      info.Size(),
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(cacheTTL),
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	cachePath := filepath.Join(cacheDir, hash+".json")
	return os.WriteFile(cachePath, data, 0600)
}

func (m *Manager) hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CleanExpiredCache removes expired cache entries.
func (m *Manager) CleanExpiredCache() error {
	return filepath.Walk(m.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var entry CacheEntry
		if json.Unmarshal(data, &entry) != nil {
			return nil
		}

		if time.Now().After(entry.ExpiresAt) {
			os.Remove(path)
		}

		return nil
	})
}

// HashFile calculates SHA256 hash of a file.
func HashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// IdempotencyKey generates an idempotency key for an operation.
func IdempotencyKey(operation string, args ...string) string {
	h := sha256.New()
	h.Write([]byte(operation))
	for _, arg := range args {
		h.Write([]byte(arg))
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// CheckIdempotency represents an idempotency check result.
type CheckIdempotency struct {
	Operation   string      `json:"operation"`
	Key         string      `json:"key"`
	LocalState  interface{} `json:"localState"`
	RemoteState interface{} `json:"remoteState"`
	Matches     bool        `json:"matches"`
	Action      string      `json:"action"` // "skip", "update", "create"
}

// IdempotencyStore stores idempotency keys.
type IdempotencyStore struct {
	dir string
}

// NewIdempotencyStore creates a new idempotency store.
func NewIdempotencyStore() *IdempotencyStore {
	paths := config.GetPaths()
	return &IdempotencyStore{
		dir: filepath.Join(paths.CacheDir, "idempotency"),
	}
}

// Check checks if an operation was already performed.
func (s *IdempotencyStore) Check(key string) (bool, error) {
	path := filepath.Join(s.dir, key+".json")
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Record records that an operation was performed.
func (s *IdempotencyStore) Record(key string, data interface{}) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}

	entry := struct {
		Key       string      `json:"key"`
		Data      interface{} `json:"data"`
		Timestamp time.Time   `json:"timestamp"`
	}{
		Key:       key,
		Data:      data,
		Timestamp: time.Now(),
	}

	jsonData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, key+".json")
	return os.WriteFile(path, jsonData, 0600)
}

// Clear removes an idempotency record.
func (s *IdempotencyStore) Clear(key string) error {
	path := filepath.Join(s.dir, key+".json")
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// CheckResult holds the result of an idempotency check.
type CheckResult struct {
	Found     bool        `json:"found"`
	Key       string      `json:"key"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
}

// Get retrieves an idempotency record.
func (s *IdempotencyStore) Get(key string) (*CheckResult, error) {
	path := filepath.Join(s.dir, key+".json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &CheckResult{Found: false, Key: key}, nil
	}
	if err != nil {
		return nil, err
	}

	var entry struct {
		Key       string      `json:"key"`
		Data      interface{} `json:"data"`
		Timestamp time.Time   `json:"timestamp"`
	}
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	return &CheckResult{
		Found:     true,
		Key:       key,
		Data:      entry.Data,
		Timestamp: entry.Timestamp,
	}, nil
}

// FormatBytes formats bytes as a human-readable string.
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
