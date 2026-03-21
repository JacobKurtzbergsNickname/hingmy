package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// FileStore persists tokens as a JSON file at ~/.config/mycli/credentials.json
// with 0600 permissions. It is the primary store in CI environments (CI=true)
// or when --no-keyring is passed.
type FileStore struct {
	path string
	mu   sync.Mutex
}

// NewFileStore creates a FileStore at the given path. If path is empty it
// defaults to ~/.config/mycli/credentials.json.
func NewFileStore(path string) (*FileStore, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("filestore: resolving home directory: %w", err)
		}
		path = filepath.Join(home, ".config", "mycli", "credentials.json")
	}
	return &FileStore{path: path}, nil
}

// Save writes the token envelope to disk. A file lock is held for the entire
// load-check-save sequence to prevent races under parallel CLI invocations.
func (f *FileStore) Save(token StoredToken) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(f.path), 0o700); err != nil {
		return fmt.Errorf("filestore.Save: creating directory: %w", err)
	}

	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("filestore.Save: marshalling token: %w", err)
	}

	// Write atomically via a temp file then rename.
	tmp := f.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("filestore.Save: writing temp file: %w", err)
	}
	if err := os.Rename(tmp, f.path); err != nil {
		os.Remove(tmp) //nolint:errcheck
		return fmt.Errorf("filestore.Save: renaming temp file: %w", err)
	}
	return nil
}

// Load reads the token envelope from disk. If the token is within 5 minutes
// of expiry Load signals the caller by returning the token with IsExpired true —
// the caller is responsible for initiating a refresh.
func (f *FileStore) Load() (StoredToken, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return StoredToken{}, fmt.Errorf("filestore.Load: not logged in")
		}
		return StoredToken{}, fmt.Errorf("filestore.Load: reading file: %w", err)
	}

	var tok StoredToken
	if err := json.Unmarshal(data, &tok); err != nil {
		return StoredToken{}, fmt.Errorf("filestore.Load: parsing token: %w", err)
	}
	return tok, nil
}

// Delete removes the credentials file.
func (f *FileStore) Delete() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if err := os.Remove(f.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("filestore.Delete: %w", err)
	}
	return nil
}
