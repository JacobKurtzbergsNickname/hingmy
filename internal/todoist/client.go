package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const syncEndpoint = "https://api.todoist.com/sync/v9/sync"

// Client handles HTTP communication with the Todoist Sync API.
// Construct one with NewClient; all methods are safe for concurrent use.
type Client struct {
	token      string
	httpClient *http.Client
}

// NewClient creates a Client authenticated with apiToken.
// Pass nil for httpClient to use a default with a 30 s timeout.
func NewClient(apiToken string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{token: apiToken, httpClient: httpClient}
}

// Sync POSTs req to the Todoist Sync API and returns the decoded response.
// All network and protocol errors are wrapped with "todoist.Client.Sync".
func (c *Client) Sync(ctx context.Context, req SyncRequest) (*SyncResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("todoist.Client.Sync: marshalling request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, syncEndpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("todoist.Client.Sync: building request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("todoist.Client.Sync: sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("todoist.Client.Sync: unexpected HTTP status %d", resp.StatusCode)
	}

	var syncResp SyncResponse
	if err := json.NewDecoder(resp.Body).Decode(&syncResp); err != nil {
		return nil, fmt.Errorf("todoist.Client.Sync: decoding response: %w", err)
	}
	return &syncResp, nil
}
