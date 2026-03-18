package database

import (
	"database/sql"
	"testing"

	"taedae/database/sqlc"

	_ "github.com/mattn/go-sqlite3"
)

func newTestAccessor(t *testing.T) (*Accessor, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory DB: %v", err)
	}
	if err := CreateTodosTable(db); err != nil {
		t.Fatalf("create todos table: %v", err)
	}
	return &Accessor{Queries: sqlc.New(db)}, db
}

func TestCreateTodo(t *testing.T) {
	accessor, db := newTestAccessor(t)
	defer db.Close()

	todo, err := accessor.CreateTodo("Buy haggis", "From the local market", "2026-03-25")
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if todo.Title != "Buy haggis" {
		t.Errorf("Title = %q, want %q", todo.Title, "Buy haggis")
	}
	if todo.Description.String != "From the local market" {
		t.Errorf("Description = %q, want %q", todo.Description.String, "From the local market")
	}
	if todo.Completed.Bool {
		t.Error("new todo should not be completed")
	}
	if todo.ID == 0 {
		t.Error("expected non-zero ID after insert")
	}
}

func TestCreateTodo_EmptyDescriptionAndDate(t *testing.T) {
	accessor, db := newTestAccessor(t)
	defer db.Close()

	todo, err := accessor.CreateTodo("Minimal todo", "", "")
	if err != nil {
		t.Fatalf("CreateTodo: %v", err)
	}
	if todo.Description.Valid {
		t.Errorf("expected null description, got %q", todo.Description.String)
	}
	if todo.DueDate.Valid {
		t.Error("expected null due date")
	}
}

func TestGetAllTodos(t *testing.T) {
	accessor, db := newTestAccessor(t)
	defer db.Close()

	for _, title := range []string{"First", "Second", "Third"} {
		if _, err := accessor.CreateTodo(title, "", ""); err != nil {
			t.Fatalf("CreateTodo %q: %v", title, err)
		}
	}

	todos, err := accessor.GetAllTodos()
	if err != nil {
		t.Fatalf("GetAllTodos: %v", err)
	}
	if len(todos) != 3 {
		t.Errorf("expected 3 todos, got %d", len(todos))
	}
}

func TestGetAllTodos_Empty(t *testing.T) {
	accessor, db := newTestAccessor(t)
	defer db.Close()

	todos, err := accessor.GetAllTodos()
	if err != nil {
		t.Fatalf("GetAllTodos on empty DB: %v", err)
	}
	if len(todos) != 0 {
		t.Errorf("expected 0 todos, got %d", len(todos))
	}
}
