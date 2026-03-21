package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"hingmy/internal/ui"
)

// PasswordAuth implements AuthService using username/password credential
// exchange against a backend API endpoint.
type PasswordAuth struct {
	baseAuth
	apiBaseURL string
	ui         ui.UI
}

// PasswordOption is a functional option for PasswordAuth.
type PasswordOption func(*PasswordAuth)

// WithPasswordUI sets the UI implementation.
func WithPasswordUI(u ui.UI) PasswordOption {
	return func(a *PasswordAuth) { a.ui = u }
}

// WithPasswordHTTPClient overrides the default HTTP client.
func WithPasswordHTTPClient(c *http.Client) PasswordOption {
	return func(a *PasswordAuth) { a.httpClient = c }
}

// NewPasswordAuth constructs a PasswordAuth with functional options.
func NewPasswordAuth(store TokenStore, apiBaseURL string, opts ...PasswordOption) *PasswordAuth {
	a := &PasswordAuth{
		baseAuth:   baseAuth{store: store, httpClient: defaultHTTPClient()},
		apiBaseURL: apiBaseURL,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// loginRequest is the JSON body sent to the auth endpoint.
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// loginResponse is the JSON body returned by a successful auth endpoint call.
type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // seconds
}

// Login prompts for email and password interactively, exchanges them with the
// backend API, and persists the returned token.
//
// SEC: credentials are never accepted via flags or environment variables —
// they appear in shell history and process listings.
func (p *PasswordAuth) Login(ctx context.Context) (string, error) {
	fmt.Print("Email: ")
	var email string
	if _, err := fmt.Scanln(&email); err != nil {
		return "", fmt.Errorf("reading email: %w", err)
	}

	password, err := readPassword("Password: ")
	if err != nil {
		return "", fmt.Errorf("reading password: %w", err)
	}

	var spinner ui.Spinner
	if p.ui != nil {
		spinner = p.ui.Spinner("Logging in...")
	}

	tok, err := p.exchangeCredentials(ctx, email, password)
	if err != nil {
		if spinner != nil {
			spinner.Stop() //nolint:errcheck
		}
		return "", err
	}
	if spinner != nil {
		spinner.Stop() //nolint:errcheck
	}

	if err := p.store.Save(tok); err != nil {
		return "", fmt.Errorf("saving token: %w", err)
	}

	return tok.AccessToken, nil
}

func (p *PasswordAuth) exchangeCredentials(ctx context.Context, email, password string) (StoredToken, error) {
	body, err := json.Marshal(loginRequest{Email: email, Password: password})
	if err != nil {
		return StoredToken{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiBaseURL+"/auth/login", bytes.NewReader(body))
	if err != nil {
		return StoredToken{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return StoredToken{}, fmt.Errorf("password.exchangeCredentials: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return StoredToken{}, fmt.Errorf("password.exchangeCredentials: status %d", resp.StatusCode)
	}

	var lr loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return StoredToken{}, fmt.Errorf("password.exchangeCredentials: decoding response: %w", err)
	}

	expiresAt := time.Time{}
	if lr.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(lr.ExpiresIn) * time.Second)
	}

	return StoredToken{
		AccessToken:  lr.AccessToken,
		RefreshToken: lr.RefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}
