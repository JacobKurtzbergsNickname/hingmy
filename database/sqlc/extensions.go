package sqlc

import "fmt"

func (t *Todo) ToString() string {
	return fmt.Sprintf(
		`"Todo{ID: %d,
		Title: %s, 
		Description: %s, 
		DueDate: %v, 
		Completed: %v, 
		CreatedAt: %v, 
		UpdatedAt: %v, 
		DeletedAt: %v}"`,
		t.ID,
		t.Title,
		t.Description.String,
		t.DueDate.Time,
		t.Completed.Bool,
		t.CreatedAt.Time,
		t.UpdatedAt.Time,
		t.DeletedAt.Time,
	)
}
