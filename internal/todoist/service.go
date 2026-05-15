package todoist

import (
	"context"
	"fmt"
)

// Service is the high-level API for Todoist CRUD operations.
// It is the only type that cmd/todoist.go imports; callers never touch
// SyncRequest, SyncResponse, or Command directly.
type Service struct {
	client     *Client
	tokenStore SyncTokenStore
}

// NewService constructs a Service. Both arguments are required.
func NewService(client *Client, tokenStore SyncTokenStore) *Service {
	return &Service{client: client, tokenStore: tokenStore}
}

// ListItems performs an incremental sync (or a full sync on the first run)
// and returns all incomplete, non-deleted items from the response.
// The updated sync token is persisted before returning.
func (s *Service) ListItems(ctx context.Context) ([]Item, error) {
	token, err := s.tokenStore.Load()
	if err != nil {
		return nil, fmt.Errorf("todoist.Service.ListItems: loading sync token: %w", err)
	}

	resp, err := s.client.Sync(ctx, SyncRequest{
		SyncToken:     token,
		ResourceTypes: []string{"items"},
	})
	if err != nil {
		return nil, fmt.Errorf("todoist.Service.ListItems: %w", err)
	}

	// Persist the new token before filtering — even if we return 0 items the
	// token should advance so the next call doesn't re-fetch the same delta.
	_ = s.tokenStore.Save(resp.SyncToken)

	var active []Item
	for _, item := range resp.Items {
		if !item.IsDeleted && !item.Checked {
			active = append(active, item)
		}
	}
	return active, nil
}

// CreateItem sends an item_add command and returns the server-assigned task ID.
// content is required; description and due are optional (pass "" to omit).
// due accepts a human string, e.g. "tomorrow" or "2026-05-20".
func (s *Service) CreateItem(ctx context.Context, content, description, due string) (string, error) {
	cmd := AddItemCommand(content, description, due)

	token, err := s.tokenStore.Load()
	if err != nil {
		return "", fmt.Errorf("todoist.Service.CreateItem: loading sync token: %w", err)
	}

	resp, err := s.client.Sync(ctx, SyncRequest{
		SyncToken:     token,
		ResourceTypes: []string{"items"},
		Commands:      []Command{cmd},
	})
	if err != nil {
		return "", fmt.Errorf("todoist.Service.CreateItem: %w", err)
	}

	_ = s.tokenStore.Save(resp.SyncToken)

	realID, ok := resp.TempIDMapping[cmd.TempID]
	if !ok {
		return "", fmt.Errorf("todoist.Service.CreateItem: server did not return an ID for temp_id %s", cmd.TempID)
	}
	return realID, nil
}

// UpdateItem modifies an existing task by its server-assigned ID.
// Pass "" for any field you do not want to change.
func (s *Service) UpdateItem(ctx context.Context, id, content, description, due string) error {
	return s.sendCommand(ctx, UpdateItemCommand(id, content, description, due), "todoist.Service.UpdateItem")
}

// CompleteItem marks a task as done by its server-assigned ID.
func (s *Service) CompleteItem(ctx context.Context, id string) error {
	return s.sendCommand(ctx, CompleteItemCommand(id), "todoist.Service.CompleteItem")
}

// DeleteItem permanently removes a task by its server-assigned ID.
func (s *Service) DeleteItem(ctx context.Context, id string) error {
	return s.sendCommand(ctx, DeleteItemCommand(id), "todoist.Service.DeleteItem")
}

// sendCommand is the shared helper for write-only operations.
// It loads the current sync token, sends the command, saves the returned token,
// and checks the per-command sync_status for rejection errors.
func (s *Service) sendCommand(ctx context.Context, cmd Command, caller string) error {
	token, err := s.tokenStore.Load()
	if err != nil {
		return fmt.Errorf("%s: loading sync token: %w", caller, err)
	}

	resp, err := s.client.Sync(ctx, SyncRequest{
		SyncToken:     token,
		ResourceTypes: []string{"items"},
		Commands:      []Command{cmd},
	})
	if err != nil {
		return fmt.Errorf("%s: %w", caller, err)
	}

	_ = s.tokenStore.Save(resp.SyncToken)

	// The Sync API returns "ok" for successful commands, or an error object.
	if status, exists := resp.SyncStatus[cmd.UUID]; exists {
		if s, ok := status.(string); ok && s == "ok" {
			return nil
		}
		// Anything other than the string "ok" is treated as an error.
		return fmt.Errorf("%s: command rejected by server: %v", caller, status)
	}

	return nil
}
