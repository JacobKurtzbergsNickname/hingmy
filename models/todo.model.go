package models

import (
	"time"
)

// Todo - represents a task with a title, description, and completion status
type Todo struct {
	ID          int       `db:"id"`
	Title       string    `db:"title"`
	Description string    `db:"description"`
	Completed   bool      `db:"completed"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	DeletedAt   time.Time `db:"deleted_at"`
	DueDate     time.Time `db:"due_date"`
}

// Tag - represents a label that can be associated with a todo item
type Tag struct {
	ID        int       `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

// TagEntity - represents the association between a todo item and a tag
type TagEntity struct {
	ID        int       `db:"id"`
	TodoID    int       `db:"todo_id"`
	TagID     int       `db:"tag_id"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}

type Note struct {
	ID        int       `db:"id"`
	TodoID    int       `db:"todo_id"`
	Content   string    `db:"content"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	DeletedAt time.Time `db:"deleted_at"`
}
