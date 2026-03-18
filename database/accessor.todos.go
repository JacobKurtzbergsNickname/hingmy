package database

import (
	"context"
	"database/sql"
	"hingmy/database/sqlc"
	"time"
)

const RFC3339DateLayout = "2006-01-02"

func (a *Accessor) CreateTodo(
	title string,
	description string,
	dueDate string,
) (*sqlc.Todo, error) {
	// Parse dueDate string to time.Time
	dueDateAsTime, err := time.Parse(RFC3339DateLayout, dueDate)
	if err != nil {
		dueDateAsTime = time.Time{}
	}

	// Prepare parameters for CreateTodo
	params := sqlc.CreateTodoParams{
		Title: title,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
		DueDate: sql.NullTime{
			Time:  dueDateAsTime,
			Valid: !dueDateAsTime.IsZero(),
		},
		CreatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		UpdatedAt: sql.NullTime{Time: time.Now(), Valid: true},
		DeletedAt: sql.NullTime{Time: time.Time{}, Valid: false},
		Completed: sql.NullBool{Bool: false, Valid: true},
	}

	// Call the generated CreateTodo method
	todo, err := a.Queries.CreateTodo(context.Background(), params)
	if err != nil {
		return nil, err
	}
	return &todo, nil
}

func (a *Accessor) GetAllTodos() ([]sqlc.Todo, error) {
	todos, err := a.Queries.ListTodos(context.Background())
	if err != nil {
		return nil, err
	}
	return todos, nil
}

func (a *Accessor) UpdateTodo(id int64, title string, description string, dueDate string, completed bool) error {
	dueDateAsTime, err := time.Parse(RFC3339DateLayout, dueDate)
	if err != nil {
		dueDateAsTime = time.Time{}
	}

	return a.Queries.UpdateTodo(context.Background(), sqlc.UpdateTodoParams{
		ID:    id,
		Title: title,
		Description: sql.NullString{
			String: description,
			Valid:  description != "",
		},
		DueDate: sql.NullTime{
			Time:  dueDateAsTime,
			Valid: !dueDateAsTime.IsZero(),
		},
		Completed: sql.NullBool{Bool: completed, Valid: true},
	})
}

func (a *Accessor) SoftDeleteTodo(id int64) error {
	return a.Queries.SoftDeleteTodo(context.Background(), id)
}
