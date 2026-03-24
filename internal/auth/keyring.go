package auth

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

const keyringService = "mycli"
const keyringUser = "token"

// KeyringStore persists tokens in the OS keychain using platform-specific CLI
// tools (secret-tool on Linux, security on macOS). If the keychain tool is
// unavailable it returns ErrKeychainUnavailable so the caller can fall back to
// FileStore.
type KeyringStore struct{}

// ErrKeychainUnavailable is returned when the OS keychain tool cannot be found.
var ErrKeychainUnavailable = fmt.Errorf("keychain tool unavailable")

// NewKeyringStore creates a KeyringStore.
func NewKeyringStore() *KeyringStore { return &KeyringStore{} }

// Save encodes the token as JSON and stores it under the service/user key.
func (k *KeyringStore) Save(token StoredToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("keyring.Save: marshalling: %w", err)
	}
	return k.set(string(data))
}

// Load retrieves the token from the OS keychain.
func (k *KeyringStore) Load() (StoredToken, error) {
	raw, err := k.get()
	if err != nil {
		return StoredToken{}, err
	}

	var tok StoredToken
	if err := json.Unmarshal([]byte(raw), &tok); err != nil {
		return StoredToken{}, fmt.Errorf("keyring.Load: parsing token: %w", err)
	}
	return tok, nil
}

// Delete removes the keychain entry.
func (k *KeyringStore) Delete() error {
	return k.del()
}

// --- platform-specific helpers (Linux via secret-tool, macOS via security) ---

func (k *KeyringStore) set(value string) error {
	if path, err := exec.LookPath("secret-tool"); err == nil {
		cmd := exec.Command(path, "store",
			"--label="+keyringService+" token",
			"service", keyringService,
			"account", keyringUser)
		cmd.Stdin = strings.NewReader(value)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("keyring.set (secret-tool): %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	if path, err := exec.LookPath("security"); err == nil {
		cmd := exec.Command(path, "add-generic-password",
			"-s", keyringService,
			"-a", keyringUser,
			"-w", value,
			"-U")
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("keyring.set (security): %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	return ErrKeychainUnavailable
}

func (k *KeyringStore) get() (string, error) {
	if path, err := exec.LookPath("secret-tool"); err == nil {
		out, err := exec.Command(path, "lookup", "service", keyringService, "account", keyringUser).Output()
		if err != nil {
			return "", fmt.Errorf("keyring.get (secret-tool): not found or error: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	if path, err := exec.LookPath("security"); err == nil {
		out, err := exec.Command(path, "find-generic-password",
			"-s", keyringService,
			"-a", keyringUser,
			"-w").Output()
		if err != nil {
			return "", fmt.Errorf("keyring.get (security): not found or error: %w", err)
		}
		return strings.TrimSpace(string(out)), nil
	}

	return "", ErrKeychainUnavailable
}

func (k *KeyringStore) del() error {
	if path, err := exec.LookPath("secret-tool"); err == nil {
		cmd := exec.Command(path, "clear", "service", keyringService, "account", keyringUser)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("keyring.del (secret-tool): %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	if path, err := exec.LookPath("security"); err == nil {
		cmd := exec.Command(path, "delete-generic-password",
			"-s", keyringService,
			"-a", keyringUser)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("keyring.del (security): %s: %w", strings.TrimSpace(string(out)), err)
		}
		return nil
	}

	return ErrKeychainUnavailable
}
