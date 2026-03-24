// Package auth provides the AuthService and TokenStore interfaces plus the
// StoredToken envelope used across all auth implementations.
package auth

import (
	"context"
	"time"
)

// AuthService is the single interface both PasswordAuth and OAuthAuth satisfy.
type AuthService interface {
	Login(ctx context.Context) (string, error)
	Logout() error
	Status() (string, error)
}

// TokenStore persists, loads, and deletes the local credential envelope.
type TokenStore interface {
	Save(token StoredToken) error
	Load() (StoredToken, error)
	Delete() error
}

// StoredToken is the full credential envelope stored on disk / in the keychain.
// Storing the full envelope (not just the access token string) enables
// transparent refresh without user intervention.
type StoredToken struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// IsExpired reports whether the token has expired or will expire within the
// given threshold.
func (t StoredToken) IsExpired(threshold time.Duration) bool {
	return time.Now().Add(threshold).After(t.ExpiresAt)
}

// MaskToken returns a redacted version of the access token for display purposes.
// Tokens shorter than 8 characters are fully masked.
func MaskToken(token string) string {
	const bullets = "••••••••"
	if len(token) < 8 {
		return bullets
	}
	return token[:4] + bullets + token[len(token)-4:]
}
