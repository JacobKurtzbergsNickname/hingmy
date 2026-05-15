package todoist

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pkg/browser"
)

const (
	todoistAuthBaseURL = "https://todoist.com/oauth/authorize"
	todoistTokenURL    = "https://todoist.com/oauth/access_token"
	oauthCallbackPort  = "8888"
	oauthScopes        = "data:read_write,data:delete"
)

// OAuthConfig holds the Todoist app credentials needed for the OAuth 2.0 flow.
// Obtain these by registering an app at https://developer.todoist.com.
type OAuthConfig struct {
	ClientID     string // TODOIST_CLIENT_ID env var
	ClientSecret string // TODOIST_CLIENT_SECRET env var
}

// AccessTokenStore persists the Todoist OAuth access token to a file.
// This is separate from SyncTokenStore — it stores auth credentials, not sync state.
type AccessTokenStore struct {
	path string
	mu   sync.Mutex
}

// NewAccessTokenStore creates an AccessTokenStore at path.
// If path is empty it defaults to ~/.local/share/hingmy/todoist_access_token.
func NewAccessTokenStore(path string) (*AccessTokenStore, error) {
	if path == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("todoist.AccessTokenStore: resolving home directory: %w", err)
		}
		path = filepath.Join(home, ".local", "share", "hingmy", "todoist_access_token")
	}
	return &AccessTokenStore{path: path}, nil
}

// Load returns the stored access token, or an empty string if none is saved.
func (s *AccessTokenStore) Load() (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("todoist.AccessTokenStore.Load: %w", err)
	}
	return strings.TrimSpace(string(data)), nil
}

// Save writes token to disk atomically with 0600 permissions.
func (s *AccessTokenStore) Save(token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return fmt.Errorf("todoist.AccessTokenStore.Save: creating directory: %w", err)
	}

	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(token), 0o600); err != nil {
		return fmt.Errorf("todoist.AccessTokenStore.Save: writing temp file: %w", err)
	}
	if err := os.Rename(tmp, s.path); err != nil {
		os.Remove(tmp) //nolint:errcheck
		return fmt.Errorf("todoist.AccessTokenStore.Save: renaming temp file: %w", err)
	}
	return nil
}

// Delete removes the stored access token file.
func (s *AccessTokenStore) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("todoist.AccessTokenStore.Delete: %w", err)
	}
	return nil
}

// RunOAuthFlow executes the Todoist Authorization Code OAuth 2.0 flow:
//  1. Opens the Todoist consent page in the user's browser.
//  2. Starts a local HTTP server on localhost:8888 to receive the callback.
//  3. Exchanges the authorization code for an access token.
//  4. Saves the token via store.
//
// The flow times out after 2 minutes if the user doesn't complete authorization.
func RunOAuthFlow(cfg OAuthConfig, store *AccessTokenStore) error {
	if cfg.ClientID == "" || cfg.ClientSecret == "" {
		return fmt.Errorf("todoist.RunOAuthFlow: TODOIST_CLIENT_ID and TODOIST_CLIENT_SECRET must be set in .env")
	}

	state, err := randomState()
	if err != nil {
		return fmt.Errorf("todoist.RunOAuthFlow: generating state: %w", err)
	}

	redirectURI := "http://localhost:" + oauthCallbackPort + "/callback"

	authURL := buildAuthURL(cfg.ClientID, state, redirectURI)
	fmt.Printf("Opening yer browser tae authorise Todoist...\n%s\n\n", authURL)
	_ = browser.OpenURL(authURL)

	code, returnedState, err := waitForCallback(oauthCallbackPort, 2*time.Minute)
	if err != nil {
		return fmt.Errorf("todoist.RunOAuthFlow: waiting for callback: %w", err)
	}
	if returnedState != state {
		return fmt.Errorf("todoist.RunOAuthFlow: state mismatch — possible CSRF attempt")
	}

	token, err := exchangeCodeForToken(cfg, code)
	if err != nil {
		return fmt.Errorf("todoist.RunOAuthFlow: exchanging code: %w", err)
	}

	if err := store.Save(token); err != nil {
		return fmt.Errorf("todoist.RunOAuthFlow: saving token: %w", err)
	}
	return nil
}

// buildAuthURL constructs the Todoist authorization URL with all required parameters.
func buildAuthURL(clientID, state, redirectURI string) string {
	params := url.Values{}
	params.Set("client_id", clientID)
	params.Set("scope", oauthScopes)
	params.Set("state", state)
	params.Set("redirect_uri", redirectURI)
	return todoistAuthBaseURL + "?" + params.Encode()
}

// oauthCallbackResult carries the result of the local callback HTTP handler.
type oauthCallbackResult struct {
	code  string
	state string
	err   error
}

// waitForCallback starts a local HTTP server that captures the OAuth callback,
// then shuts down and returns the authorization code and state.
func waitForCallback(port string, timeout time.Duration) (code, state string, err error) {
	resultCh := make(chan oauthCallbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if errMsg := q.Get("error"); errMsg != "" {
			fmt.Fprintf(w, "<html><body><h2>Authorization denied</h2><p>%s</p><p>Ye can close this tab.</p></body></html>", errMsg)
			resultCh <- oauthCallbackResult{err: fmt.Errorf("todoist: authorization denied: %s", errMsg)}
			return
		}
		fmt.Fprint(w, "<html><body><h2>Aye, ye're in!</h2><p>Authorization complete — ye can close this tab.</p></body></html>")
		resultCh <- oauthCallbackResult{code: q.Get("code"), state: q.Get("state")}
	})

	listener, err := net.Listen("tcp", "localhost:"+port)
	if err != nil {
		return "", "", fmt.Errorf("starting local callback server on port %s: %w", port, err)
	}

	srv := &http.Server{
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() { _ = srv.Serve(listener) }()
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	select {
	case result := <-resultCh:
		return result.code, result.state, result.err
	case <-time.After(timeout):
		return "", "", fmt.Errorf("timed out waiting for browser authorization after %s", timeout)
	}
}

// todoistTokenResponse is the JSON body returned by the Todoist token endpoint.
type todoistTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Error       string `json:"error"`
}

// exchangeCodeForToken POSTs the authorization code to the Todoist token endpoint
// and returns the resulting access token string.
func exchangeCodeForToken(cfg OAuthConfig, code string) (string, error) {
	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("code", code)

	resp, err := http.Post( //nolint:gosec
		todoistTokenURL,
		"application/x-www-form-urlencoded",
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		return "", fmt.Errorf("posting to token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("token endpoint returned HTTP %d", resp.StatusCode)
	}

	var tr todoistTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", fmt.Errorf("decoding token response: %w", err)
	}
	if tr.Error != "" {
		return "", fmt.Errorf("token endpoint error: %s", tr.Error)
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("token endpoint returned empty access token")
	}
	return tr.AccessToken, nil
}

// randomState generates a cryptographically random hex string for the OAuth
// state parameter, protecting against CSRF attacks.
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
