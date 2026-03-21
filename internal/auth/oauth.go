package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hingmy/internal/ui"

	"github.com/pkg/browser"
)

// OAuthConfig holds the provider-specific URLs and client credentials needed
// for the OAuth 2.0 Device Authorization Flow.
type OAuthConfig struct {
	ClientID           string
	DeviceAuthEndpoint string
	TokenEndpoint      string
	RevocationEndpoint string
	Scopes             []string
}

// OAuthAuth implements AuthService using the OAuth 2.0 Device Authorization
// Flow (RFC 8628). It works in headless and SSH environments without requiring
// a local HTTP server.
type OAuthAuth struct {
	baseAuth
	config  OAuthConfig
	ui      ui.UI
	timeout time.Duration
}

// Option is a functional option for OAuthAuth.
type Option func(*OAuthAuth)

// WithUI sets the UI implementation.
func WithUI(u ui.UI) Option { return func(o *OAuthAuth) { o.ui = u } }

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(c *http.Client) Option { return func(o *OAuthAuth) { o.httpClient = c } }

// WithTimeout sets the maximum time to wait for the user to authorise the
// device code. Defaults to 5 minutes.
func WithTimeout(d time.Duration) Option { return func(o *OAuthAuth) { o.timeout = d } }

// NewOAuthAuth constructs an OAuthAuth with functional options.
func NewOAuthAuth(store TokenStore, cfg OAuthConfig, opts ...Option) *OAuthAuth {
	a := &OAuthAuth{
		baseAuth: baseAuth{store: store, httpClient: defaultHTTPClient()},
		config:   cfg,
		timeout:  5 * time.Minute,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// deviceAuthResponse is the JSON response from the device authorization endpoint.
type deviceAuthResponse struct {
	DeviceCode              string `json:"device_code"`
	UserCode                string `json:"user_code"`
	VerificationURI         string `json:"verification_uri"`
	VerificationURIComplete string `json:"verification_uri_complete"`
	ExpiresIn               int    `json:"expires_in"`
	Interval                int    `json:"interval"`
}

// tokenResponse is the JSON response from the token endpoint.
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
	Error        string `json:"error"`
}

// Login orchestrates the Device Flow. It reads like a table of contents:
// request the device code, show it to the user, try to open a browser, then
// poll until approved or expired.
func (o *OAuthAuth) Login(ctx context.Context) (string, error) {
	deviceAuth, err := o.requestDeviceCode(ctx)
	if err != nil {
		return "", err
	}

	o.promptUser(deviceAuth)
	o.tryOpenBrowser(deviceAuth.VerificationURIComplete)

	ctx, cancel := context.WithTimeout(ctx, o.timeout)
	defer cancel()

	tok, err := o.pollForToken(ctx, deviceAuth)
	if err != nil {
		return "", err
	}

	if err := o.store.Save(tok); err != nil {
		return "", fmt.Errorf("oauth.Login: saving token: %w", err)
	}

	return tok.AccessToken, nil
}

// Logout revokes the token with the provider (best-effort) then deletes locally.
func (o *OAuthAuth) Logout() error {
	tok, err := o.store.Load()
	if err == nil && o.config.RevocationEndpoint != "" {
		_ = o.revokeToken(tok.AccessToken) // best-effort; warn caller on failure
	}
	return o.store.Delete()
}

func (o *OAuthAuth) requestDeviceCode(ctx context.Context) (deviceAuthResponse, error) {
	data := url.Values{}
	data.Set("client_id", o.config.ClientID)
	if len(o.config.Scopes) > 0 {
		data.Set("scope", strings.Join(o.config.Scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.config.DeviceAuthEndpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return deviceAuthResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return deviceAuthResponse{}, fmt.Errorf("oauth.requestDeviceCode: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return deviceAuthResponse{}, fmt.Errorf("oauth.requestDeviceCode: status %d", resp.StatusCode)
	}

	var dar deviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&dar); err != nil {
		return deviceAuthResponse{}, fmt.Errorf("oauth.requestDeviceCode: decoding response: %w", err)
	}
	if dar.Interval == 0 {
		dar.Interval = 5 // RFC 8628 default
	}
	return dar, nil
}

// promptUser displays the verification URL and user code in a persistent box.
// The box is intentionally persistent — unlike spinners it does not self-clear,
// so the user never loses sight of the code while the polling spinner runs below.
func (o *OAuthAuth) promptUser(dar deviceAuthResponse) {
	verifyURL := dar.VerificationURI
	if dar.VerificationURIComplete != "" {
		verifyURL = dar.VerificationURIComplete
	}

	body := fmt.Sprintf("Open this URL in your browser:\n  %s\n\nThen enter the code: %s",
		verifyURL, dar.UserCode)

	if o.ui != nil {
		o.ui.Box("Authorise Device", body)
	} else {
		fmt.Println(body)
	}
}

// tryOpenBrowser attempts to open the verification URL in the default browser.
// Failures are silently ignored — the user already has the URL from promptUser.
func (o *OAuthAuth) tryOpenBrowser(verificationURI string) {
	_ = browser.OpenURL(verificationURI)
}

// pollForToken polls the token endpoint with exponential backoff until the user
// approves the device code or the context is cancelled / deadline exceeded.
func (o *OAuthAuth) pollForToken(ctx context.Context, dar deviceAuthResponse) (StoredToken, error) {
	interval := time.Duration(dar.Interval) * time.Second

	var spinner ui.Spinner
	if o.ui != nil {
		spinner = o.ui.Spinner("Waiting for authorisation...")
	}

	for {
		select {
		case <-ctx.Done():
			if spinner != nil {
				spinner.Stop() //nolint:errcheck
			}
			return StoredToken{}, fmt.Errorf("oauth.pollForToken: timed out waiting for authorisation")
		case <-time.After(interval):
		}

		tok, err := o.tryExchangeDeviceCode(ctx, dar.DeviceCode)
		if err == nil {
			if spinner != nil {
				spinner.Stop() //nolint:errcheck
			}
			return tok, nil
		}

		switch err.Error() {
		case "authorization_pending":
			// normal — keep polling
		case "slow_down":
			interval += 5 * time.Second
		default:
			if spinner != nil {
				spinner.Stop() //nolint:errcheck
			}
			return StoredToken{}, fmt.Errorf("oauth.pollForToken: %w", err)
		}
	}
}

func (o *OAuthAuth) tryExchangeDeviceCode(ctx context.Context, deviceCode string) (StoredToken, error) {
	data := url.Values{}
	data.Set("client_id", o.config.ClientID)
	data.Set("device_code", deviceCode)
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.config.TokenEndpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return StoredToken{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return StoredToken{}, err
	}
	defer resp.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return StoredToken{}, fmt.Errorf("decoding token response: %w", err)
	}

	if tr.Error != "" {
		return StoredToken{}, fmt.Errorf("%s", tr.Error)
	}

	expiresAt := time.Time{}
	if tr.ExpiresIn > 0 {
		expiresAt = time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	}

	return StoredToken{
		AccessToken:  tr.AccessToken,
		RefreshToken: tr.RefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

func (o *OAuthAuth) revokeToken(accessToken string) error {
	data := url.Values{}
	data.Set("token", accessToken)

	req, err := http.NewRequest(http.MethodPost, o.config.RevocationEndpoint,
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := o.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("oauth.revokeToken: status %d", resp.StatusCode)
	}
	return nil
}
