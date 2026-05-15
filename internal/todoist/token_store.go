package todoist

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// fullSyncToken is the sentinel value that instructs the Sync API to return a
// complete snapshot rather than an incremental diff.
const fullSyncToken = "*"

// SyncTokenStore persists the Todoist incremental sync token between runs.
// The token lets the Sync API return only what changed since the last call,
// keeping responses small after the initial full sync.
type SyncTokenStore interface {
	// Load returns the stored token, or fullSyncToken ("*") if none exists yet.
	Load() (string, error)
	// Save persists token for the next run.
	Save(token string) error
	// Reset deletes the stored token, forcing a full sync on the next call.
	Reset() error
}

// FileSyncTokenStore stores the sync token as a plain text file.
// It is safe for concurrent use within a single process.
type FileSyncTokenStore struct {
	path string
	mu   sync.Mutex
}

// NewFileSyncTokenStore creates a FileSyncTokenStore at path.
// If path is empty it defaults to ~/.local/share/hingmy/todoist_sync_token.
func NewFileSyncTokenStore(path string) (*FileSyncTokenStore, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("todoist.FileSyncTokenStore: resolving home directory: %w", err)
		}
		path = filepath.Join(home, ".local", "share", "hingmy", "todoist_sync_token")
	}
	return &FileSyncTokenStore{path: path}, nil
}

// Save writes token to disk atomically using a temp file + rename, with 0600
// permissions so only the current user can read it.
func (s *FileSyncTokenStore) Save(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("todoist.FileSyncTokenStore.Save: creating directory: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(token), 0o600); err != nil {
		return fmt.Errorf("todoist.FileSyncTokenStore.Save: writing temp file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp) //nolint:errcheck
		return fmt.Errorf("todoist.FileSyncTokenStore.Save: renaming temp file: %w", err)
	}
	return nil
}

// Load returns the stored sync token. If the file does not exist it returns
// fullSyncToken ("*"), triggering a full sync on the next API call.
func (s *FileSyncTokenStore) Load() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fullSyncToken, nil
		}
		return fullSyncToken, fmt.Errorf("todoist.FileSyncTokenStore.Load: %w", err)
	}
	token := string(data)
	if token == "" {
		return fullSyncToken, nil
	}
	return token, nil
}

// Reset deletes the stored token file so the next sync is a full sync.
func (s *FileSyncTokenStore) Reset() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("todoist.FileSyncTokenStore.Reset: %w", err)
	}
	return nil
}
