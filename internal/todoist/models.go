package todoist

// Due represents a task's due date as returned by the Todoist Sync API.
type Due struct {
	Date        string `json:"date"`   // "YYYY-MM-DD" or RFC3339 datetime
	String      string `json:"string"` // human-readable, e.g. "every day at 10am"
	Lang        string `json:"lang"`
	IsRecurring bool   `json:"is_recurring"`
}

// Item represents a Todoist task as returned by the Sync API.
// Field names match the Sync API JSON keys exactly so the mapping is obvious.
type Item struct {
	ID          string `json:"id"`
	UserID      string `json:"user_id"`
	ProjectID   string `json:"project_id"`
	Content     string `json:"content"`     // the task title
	Description string `json:"description"` // optional longer text
	Due         *Due   `json:"due"`
	Checked     bool   `json:"checked"`    // true = completed
	IsDeleted   bool   `json:"is_deleted"` // true = soft-deleted on Todoist side
}

// Command is a single write operation sent inside the Sync API's `commands` array.
// Use the builder functions in commands.go rather than constructing this directly.
type Command struct {
	Type   string                 `json:"type"`
	TempID string                 `json:"temp_id"`
	UUID   string                 `json:"uuid"`
	Args   map[string]interface{} `json:"args"`
}

// SyncRequest is the full POST body for the Todoist Sync API endpoint.
type SyncRequest struct {
	SyncToken     string    `json:"sync_token"`
	ResourceTypes []string  `json:"resource_types"`
	Commands      []Command `json:"commands,omitempty"`
}

// SyncResponse is the decoded JSON response from the Todoist Sync API.
type SyncResponse struct {
	SyncToken     string                 `json:"sync_token"`
	FullSync      bool                   `json:"full_sync"`
	Items         []Item                 `json:"items"`
	TempIDMapping map[string]string      `json:"temp_id_mapping"` // temp_id → real server ID
	SyncStatus    map[string]interface{} `json:"sync_status"`     // uuid → "ok" or error object
}
