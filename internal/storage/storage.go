// Package storage provides platform-specific secure credential storage.
package storage

import (
	"runtime"

	"github.com/99designs/keyring"
)

const (
	serviceName = "gpd"
)

// SecureStorage provides secure credential storage using the system keychain.
type SecureStorage struct {
	ring      keyring.Keyring
	available bool
}

// New creates a new SecureStorage instance.
func New() *SecureStorage {
	ring, err := keyring.Open(keyring.Config{
		ServiceName: serviceName,
		// macOS specific
		KeychainName:                   "gpd",
		KeychainTrustApplication:       true,
		KeychainSynchronizable:         false,
		KeychainAccessibleWhenUnlocked: true,
		// Linux specific - prefer Secret Service
		LibSecretCollectionName: "gpd",
		// Windows specific
		WinCredPrefix: "gpd",
		// Disable file-based fallback for security
		FileDir:          "",
		FilePasswordFunc: nil,
	})

	if err != nil {
		return &SecureStorage{available: false}
	}

	return &SecureStorage{
		ring:      ring,
		available: true,
	}
}

// Store stores a value securely.
func (s *SecureStorage) Store(key string, value []byte) error {
	if !s.available {
		return ErrStorageUnavailable
	}
	return s.ring.Set(keyring.Item{
		Key:  key,
		Data: value,
	})
}

// Retrieve retrieves a value from secure storage.
func (s *SecureStorage) Retrieve(key string) ([]byte, error) {
	if !s.available {
		return nil, ErrStorageUnavailable
	}
	item, err := s.ring.Get(key)
	if err != nil {
		if err == keyring.ErrKeyNotFound {
			return nil, ErrKeyNotFound
		}
		return nil, err
	}
	return item.Data, nil
}

// Delete removes a value from secure storage.
func (s *SecureStorage) Delete(key string) error {
	if !s.available {
		return ErrStorageUnavailable
	}
	return s.ring.Remove(key)
}

// Available returns whether secure storage is available.
func (s *SecureStorage) Available() bool {
	return s.available
}

// Platform returns the current platform name.
func Platform() string {
	return runtime.GOOS
}

// Errors for storage operations.
type StorageError struct {
	message string
}

func (e *StorageError) Error() string {
	return e.message
}

var (
	ErrStorageUnavailable = &StorageError{"secure storage not available on this platform"}
	ErrKeyNotFound        = &StorageError{"key not found in secure storage"}
)
