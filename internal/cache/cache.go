package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache provides simple file-based caching
type Cache struct {
	dir string
	ttl time.Duration
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Data      []byte    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
	TTL       int       `json:"ttl"`
}

// New creates a new cache instance
func New(dir string, ttl time.Duration) *Cache {
	return &Cache{
		dir: dir,
		ttl: ttl,
	}
}

// GenerateKey creates a cache key from parameters
func (c *Cache) GenerateKey(prefix string, params map[string]interface{}) string {
	// Create a deterministic string from params
	data, _ := json.Marshal(params)
	hash := sha256.Sum256(append([]byte(prefix), data...))
	return hex.EncodeToString(hash[:])
}

// Get retrieves an item from the cache
func (c *Cache) Get(key string) ([]byte, bool) {
	path := c.filePath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, false
	}

	// Check if expired
	if time.Since(entry.CreatedAt) > c.ttl {
		_ = os.Remove(path)
		return nil, false
	}

	return entry.Data, true
}

// Set stores an item in the cache
func (c *Cache) Set(key string, data []byte) error {
	path := c.filePath(key)

	// Ensure cache directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	entry := CacheEntry{
		Data:      data,
		CreatedAt: time.Now(),
		TTL:       int(c.ttl.Seconds()),
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to encode cache entry: %w", err)
	}

	if err := os.WriteFile(path, encoded, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Delete removes an item from the cache
func (c *Cache) Delete(key string) error {
	path := c.filePath(key)
	return os.Remove(path)
}

// Clear removes all cached items
func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}

// filePath returns the file path for a cache key
func (c *Cache) filePath(key string) string {
	// Use first 2 chars of key as subdirectory to avoid too many files in one dir
	return filepath.Join(c.dir, key[:2], key)
}

// Stats returns cache statistics
func (c *Cache) Stats() (total int, size int64, err error) {
	err = filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		if !info.IsDir() {
			total++
			size += info.Size()
		}
		return nil
	})
	return
}

// Cleanup removes expired cache entries
func (c *Cache) Cleanup() error {
	return filepath.Walk(c.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking
		}
		if info.IsDir() {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var entry CacheEntry
		if err := json.Unmarshal(data, &entry); err != nil {
			_ = os.Remove(path)
			return nil
		}

		ttl := time.Duration(entry.TTL) * time.Second
		if time.Since(entry.CreatedAt) > ttl {
			_ = os.Remove(path)
		}

		return nil
	})
}
