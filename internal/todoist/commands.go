package todoist

import (
	"crypto/rand"
	"fmt"
)

// Command type strings as defined by the Todoist Sync API specification.
// Use these constants rather than raw strings so typos are caught at compile time.
const (
	CommandTypeItemAdd      = "item_add"
	CommandTypeItemUpdate   = "item_update"
	CommandTypeItemComplete = "item_complete"
	CommandTypeItemDelete   = "item_delete"
)

// newUUID generates a random UUID v4 string using crypto/rand.
// The Sync API requires UUID v4 for both the per-command uuid and temp_id fields.
func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failure is unrecoverable on any sane platform.
		panic(fmt.Sprintf("todoist: generating UUID: %v", err))
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant bits
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%12x",
		b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// AddItemCommand builds a Command that creates a new Todoist task.
// description and due are optional — pass an empty string to omit them.
// due accepts a human string like "tomorrow" or "2026-05-20".
func AddItemCommand(content, description, due string) Command {
	args := map[string]interface{}{"content": content}
	if description != "" {
		args["description"] = description
	}
	if due != "" {
		args["due"] = map[string]string{"string": due}
	}
	return Command{
		Type:   CommandTypeItemAdd,
		TempID: newUUID(),
		UUID:   newUUID(),
		Args:   args,
	}
}

// UpdateItemCommand builds a Command that modifies an existing task by its server ID.
// Only non-empty fields are included in the update payload.
func UpdateItemCommand(id, content, description, due string) Command {
	args := map[string]interface{}{"id": id}
	if content != "" {
		args["content"] = content
	}
	if description != "" {
		args["description"] = description
	}
	if due != "" {
		args["due"] = map[string]string{"string": due}
	}
	return Command{
		Type:   CommandTypeItemUpdate,
		TempID: newUUID(),
		UUID:   newUUID(),
		Args:   args,
	}
}

// CompleteItemCommand builds a Command that marks a task as done by its server ID.
func CompleteItemCommand(id string) Command {
	return Command{
		Type:   CommandTypeItemComplete,
		TempID: newUUID(),
		UUID:   newUUID(),
		Args:   map[string]interface{}{"id": id},
	}
}

// DeleteItemCommand builds a Command that permanently deletes a task by its server ID.
func DeleteItemCommand(id string) Command {
	return Command{
		Type:   CommandTypeItemDelete,
		TempID: newUUID(),
		UUID:   newUUID(),
		Args:   map[string]interface{}{"id": id},
	}
}
