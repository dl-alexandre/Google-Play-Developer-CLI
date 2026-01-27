package auth

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type memoryStorage struct {
	mu        sync.Mutex
	data      map[string][]byte
	available bool
}

func (m *memoryStorage) Store(key string, value []byte) error {
	if !m.available {
		return errors.New("storage unavailable")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.data == nil {
		m.data = make(map[string][]byte)
	}
	m.data[key] = append([]byte(nil), value...)
	return nil
}

func (m *memoryStorage) Retrieve(key string) ([]byte, error) {
	if !m.available {
		return nil, errors.New("storage unavailable")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	val, ok := m.data[key]
	if !ok {
		return nil, errors.New("not found")
	}
	return append([]byte(nil), val...), nil
}

func (m *memoryStorage) Delete(key string) error {
	if !m.available {
		return errors.New("storage unavailable")
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *memoryStorage) Available() bool {
	return m.available
}

func filepathWithTempFile(t *testing.T, data []byte) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "key.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write temp file error: %v", err)
	}
	return path
}
