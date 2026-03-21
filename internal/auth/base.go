package auth

import (
	"fmt"
	"net/http"
	"time"
)

// baseAuth holds dependencies shared by PasswordAuth and OAuthAuth: the token
// store and an HTTP client. Embed it in concrete auth implementations.
type baseAuth struct {
	store      TokenStore
	httpClient *http.Client
}

// Logout revokes the token with the provider (best-effort) and deletes it from
// the local store. The caller should warn the user if revocation fails but
// should still remove the local token.
func (b *baseAuth) Logout() error {
	return b.store.Delete()
}

// Status loads the current token from the store and returns a human-readable
// summary, or an error if the user is not logged in.
func (b *baseAuth) Status() (string, error) {
	tok, err := b.store.Load()
	if err != nil {
		return "", fmt.Errorf("not logged in")
	}

	expiry := "never"
	if !tok.ExpiresAt.IsZero() {
		if tok.IsExpired(0) {
			expiry = "expired"
		} else {
			expiry = tok.ExpiresAt.Format(time.RFC3339)
		}
	}

	return fmt.Sprintf("token: %s  expires: %s", MaskToken(tok.AccessToken), expiry), nil
}

// defaultHTTPClient returns an HTTP client with a sensible timeout.
func defaultHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}
