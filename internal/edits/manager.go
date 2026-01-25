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
	"runtime"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/dl-alexandre/gpd/internal/config"
	"github.com/dl-alexandre/gpd/internal/errors"
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
	StateDraft      EditState = "draft"
	StateValidating EditState = "validating"
	StateCommitted  EditState = "committed"
	StateAborted    EditState = "aborted"
)

type Manager struct {
	editsDir   string
	cacheDir   string
	mu         sync.RWMutex
	lockFiles  map[string]*LockFile
	Idempotent *IdempotencyStore
}

func NewManager() *Manager {
	paths := config.GetPaths()
	return &Manager{
		editsDir:   filepath.Join(paths.ConfigDir, "edits"),
		cacheDir:   filepath.Join(paths.CacheDir, "artifacts"),
		lockFiles:  make(map[string]*LockFile),
		Idempotent: NewIdempotencyStore(),
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
	editTTL          = 7 * 24 * time.Hour
	editIdleTTL      = 1 * time.Hour
)

// AcquireLock acquires a lock for the given package.
func (m *Manager) AcquireLock(ctx context.Context, packageName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := os.MkdirAll(m.editsDir, 0700); err != nil {
		return err
	}

	lockPath := filepath.Join(m.editsDir, packageName+".lock")
	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

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
				if err := os.Remove(lockPath); err != nil {
					return false
				}
			} else {
				return false
			}
		}
	}

	// Try to create new lock atomically
	lockData, err := json.Marshal(newLock)
	if err != nil {
		return false
	}
	f, err := os.OpenFile(lockPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
	if err != nil {
		return false
	}
	_, err = f.Write(lockData)
	if err != nil {
		if closeErr := f.Close(); closeErr != nil {
			return false
		}
		return false
	}
	if err := f.Close(); err != nil {
		return false
	}
	return true
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

func (m *Manager) SaveEdit(edit *Edit) error {
	if err := os.MkdirAll(m.editsDir, 0700); err != nil {
		return err
	}

	path := m.editPath(edit.PackageName, edit.Handle)
	data, err := json.MarshalIndent(edit, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0600)
}

func (m *Manager) LoadEdit(packageName, handle string) (*Edit, error) {
	path := m.editPath(packageName, handle)
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

func (m *Manager) DeleteEdit(packageName, handle string) error {
	path := m.editPath(packageName, handle)
	return os.Remove(path)
}

func (m *Manager) ListEdits(packageName string) ([]*Edit, error) {
	if err := os.MkdirAll(m.editsDir, 0700); err != nil {
		return nil, err
	}

	pattern := filepath.Join(m.editsDir, m.editPrefix(packageName)+"*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var editsList []*Edit
	for _, path := range matches {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var edit Edit
		if json.Unmarshal(data, &edit) != nil {
			continue
		}
		editsList = append(editsList, &edit)
	}
	return editsList, nil
}

func (m *Manager) UpdateEditState(packageName, handle string, state EditState) (*Edit, error) {
	edit, err := m.LoadEdit(packageName, handle)
	if err != nil {
		return nil, err
	}
	if edit == nil {
		return nil, errors.NewAPIError(errors.CodeNotFound, "edit not found")
	}
	edit.State = state
	edit.LastUsedAt = time.Now()
	if err := m.SaveEdit(edit); err != nil {
		return nil, err
	}
	return edit, nil
}

func (m *Manager) TouchEdit(packageName, handle string) (*Edit, error) {
	edit, err := m.LoadEdit(packageName, handle)
	if err != nil {
		return nil, err
	}
	if edit == nil {
		return nil, errors.NewAPIError(errors.CodeNotFound, "edit not found")
	}
	edit.LastUsedAt = time.Now()
	if err := m.SaveEdit(edit); err != nil {
		return nil, err
	}
	return edit, nil
}

func (m *Manager) IsEditExpired(edit *Edit, now time.Time) bool {
	if edit == nil {
		return true
	}
	if now.Sub(edit.CreatedAt) > editTTL {
		return true
	}
	if now.Sub(edit.LastUsedAt) > editIdleTTL {
		return true
	}
	return false
}

func (m *Manager) editPath(packageName, handle string) string {
	return filepath.Join(m.editsDir, m.editPrefix(packageName)+m.sanitizeHandle(handle)+".json")
}

func (m *Manager) editPrefix(packageName string) string {
	return m.sanitizeHandle(packageName) + "_"
}

func (m *Manager) sanitizeHandle(handle string) string {
	if handle == "" {
		handle = "default"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_")
	return replacer.Replace(handle)
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

const (
	largeFileThreshold = 100 * 1024 * 1024
	hashBufferSize     = 64 * 1024
)

type ProgressCallback func(bytesProcessed, totalBytes int64)

func (m *Manager) GetCachedArtifactByHash(packageName, hash string) (*CacheEntry, error) {
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

	if time.Now().After(entry.ExpiresAt) {
		if err := os.Remove(cachePath); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return nil, nil
	}

	return &entry, nil
}

func (m *Manager) CacheArtifactWithHash(packageName, artifactPath, hash string, versionCode int64) error {
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

func (m *Manager) CleanExpiredCache() error {
	return m.CleanExpiredCacheWithContext(context.Background())
}

func (m *Manager) CleanExpiredCacheWithContext(ctx context.Context) error {
	var paths []string
	err := filepath.Walk(m.cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}
		paths = append(paths, path)
		return nil
	})
	if err != nil {
		return err
	}

	if len(paths) == 0 {
		return nil
	}

	workers := runtime.GOMAXPROCS(0)
	if workers > 4 {
		workers = 4
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(workers)

	for _, p := range paths {
		g.Go(func() error {
			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}

			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}

			var entry CacheEntry
			if json.Unmarshal(data, &entry) != nil {
				return nil
			}

			if time.Now().After(entry.ExpiresAt) {
				if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
			return nil
		})
	}

	return g.Wait()
}

func HashFile(path string) (string, error) {
	return HashFileWithProgress(path, nil)
}

func HashFileWithProgress(path string, progress ProgressCallback) (hash string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	info, err := f.Stat()
	if err != nil {
		return "", err
	}

	h := sha256.New()
	totalBytes := info.Size()

	if totalBytes <= largeFileThreshold && progress == nil {
		if _, err := io.Copy(h, f); err != nil {
			return "", err
		}
		return hex.EncodeToString(h.Sum(nil)), nil
	}

	buf := make([]byte, hashBufferSize)
	var bytesProcessed int64

	for {
		n, err := f.Read(buf)
		if n > 0 {
			if _, err := h.Write(buf[:n]); err != nil {
				return "", err
			}
			bytesProcessed += int64(n)
			if progress != nil {
				progress(bytesProcessed, totalBytes)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// IdempotencyKey generates an idempotency key for an operation.
func IdempotencyKey(operation string, args ...string) string {
	h := sha256.New()
	if _, err := h.Write([]byte(operation)); err != nil {
		return ""
	}
	for _, arg := range args {
		if _, err := h.Write([]byte(arg)); err != nil {
			return ""
		}
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

const idempotencyTTL = 24 * time.Hour

type IdempotencyEntry struct {
	Key         string      `json:"key"`
	Operation   string      `json:"operation"`
	PackageName string      `json:"packageName"`
	ContentHash string      `json:"contentHash,omitempty"`
	Data        interface{} `json:"data"`
	Timestamp   time.Time   `json:"timestamp"`
	ExpiresAt   time.Time   `json:"expiresAt"`
}

type IdempotencyStore struct {
	dir string
	ttl time.Duration
}

func NewIdempotencyStore() *IdempotencyStore {
	paths := config.GetPaths()
	return &IdempotencyStore{
		dir: filepath.Join(paths.CacheDir, "idempotency"),
		ttl: idempotencyTTL,
	}
}

func (s *IdempotencyStore) generateKey(operation, packageName, contentHash string) string {
	h := sha256.New()
	if _, err := h.Write([]byte(operation)); err != nil {
		return ""
	}
	if _, err := h.Write([]byte(packageName)); err != nil {
		return ""
	}
	if _, err := h.Write([]byte(contentHash)); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))[:32]
}

func (s *IdempotencyStore) Check(key string) (bool, error) {
	result, err := s.Get(key)
	if err != nil {
		return false, err
	}
	return result.Found, nil
}

func (s *IdempotencyStore) Record(key string, data interface{}) error {
	return s.RecordWithMeta(key, "", "", "", data)
}

func (s *IdempotencyStore) RecordWithMeta(key, operation, packageName, contentHash string, data interface{}) error {
	if err := os.MkdirAll(s.dir, 0700); err != nil {
		return err
	}

	entry := &IdempotencyEntry{
		Key:         key,
		Operation:   operation,
		PackageName: packageName,
		ContentHash: contentHash,
		Data:        data,
		Timestamp:   time.Now(),
		ExpiresAt:   time.Now().Add(s.ttl),
	}

	jsonData, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return err
	}

	path := filepath.Join(s.dir, key+".json")
	return os.WriteFile(path, jsonData, 0600)
}

func (s *IdempotencyStore) Clear(key string) error {
	path := filepath.Join(s.dir, key+".json")
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

type CheckResult struct {
	Found     bool        `json:"found"`
	Key       string      `json:"key"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp,omitempty"`
	Expired   bool        `json:"expired,omitempty"`
}

func (s *IdempotencyStore) Get(key string) (*CheckResult, error) {
	path := filepath.Join(s.dir, key+".json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return &CheckResult{Found: false, Key: key}, nil
	}
	if err != nil {
		return nil, err
	}

	var entry IdempotencyEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, err
	}

	if time.Now().After(entry.ExpiresAt) {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return nil, err
		}
		return &CheckResult{Found: false, Key: key, Expired: true}, nil
	}

	return &CheckResult{
		Found:     true,
		Key:       key,
		Data:      entry.Data,
		Timestamp: entry.Timestamp,
	}, nil
}

func (s *IdempotencyStore) CleanExpired() error {
	return s.CleanExpiredWithContext(context.Background())
}

func (s *IdempotencyStore) CleanExpiredWithContext(ctx context.Context) error {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var paths []string
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		paths = append(paths, filepath.Join(s.dir, entry.Name()))
	}

	if len(paths) == 0 {
		return nil
	}

	workers := runtime.GOMAXPROCS(0)
	if workers > 4 {
		workers = 4
	}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(workers)

	for _, p := range paths {
		g.Go(func() error {
			select {
			case <-gctx.Done():
				return gctx.Err()
			default:
			}

			data, err := os.ReadFile(p)
			if err != nil {
				return nil
			}

			var e IdempotencyEntry
			if json.Unmarshal(data, &e) != nil {
				return nil
			}

			if time.Now().After(e.ExpiresAt) {
				if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
					return err
				}
			}
			return nil
		})
	}

	return g.Wait()
}

type UploadResult struct {
	VersionCode int64  `json:"versionCode"`
	SHA256      string `json:"sha256"`
	Path        string `json:"path"`
	Size        int64  `json:"size"`
	Type        string `json:"type"`
	EditID      string `json:"editId"`
}

func (s *IdempotencyStore) CheckUploadByHash(packageName, hash string) (*CheckResult, string, error) {
	key := s.generateKey("upload", packageName, hash)
	result, err := s.Get(key)
	if err != nil {
		return nil, key, err
	}

	return result, key, nil
}

func (s *IdempotencyStore) RecordUpload(key, packageName, hash string, result *UploadResult) error {
	return s.RecordWithMeta(key, "upload", packageName, hash, result)
}

type CommitResult struct {
	EditID    string `json:"editId"`
	Committed bool   `json:"committed"`
}

func (s *IdempotencyStore) CheckCommit(packageName, editID, contentIdentifier string) (*CheckResult, string, error) {
	key := s.generateKey("commit", packageName, editID+":"+contentIdentifier)
	result, err := s.Get(key)
	if err != nil {
		return nil, key, err
	}
	return result, key, nil
}

func (s *IdempotencyStore) RecordCommit(key, packageName, editID string) error {
	result := &CommitResult{
		EditID:    editID,
		Committed: true,
	}
	return s.RecordWithMeta(key, "commit", packageName, editID, result)
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
