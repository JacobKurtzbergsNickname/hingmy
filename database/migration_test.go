package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func openMemDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory DB: %v", err)
	}
	return db
}

func TestCreateTodosTable(t *testing.T) {
	db := openMemDB(t)
	defer db.Close()

	if err := CreateTodosTable(db); err != nil {
		t.Fatalf("CreateTodosTable: %v", err)
	}
	// Should be idempotent
	if err := CreateTodosTable(db); err != nil {
		t.Fatalf("CreateTodosTable (second call): %v", err)
	}

	result, err := CheckForTables(db, []string{"todos"})
	if err != nil {
		t.Fatalf("CheckForTables: %v", err)
	}
	if !result["todos"] {
		t.Error("expected todos table to exist")
	}
}

func TestCreateTagsTable(t *testing.T) {
	db := openMemDB(t)
	defer db.Close()

	if err := CreateTagsTable(db); err != nil {
		t.Fatalf("CreateTagsTable: %v", err)
	}

	result, err := CheckForTables(db, []string{"tags"})
	if err != nil {
		t.Fatalf("CheckForTables: %v", err)
	}
	if !result["tags"] {
		t.Error("expected tags table to exist")
	}
}

func TestCreateTagEntitiesTable(t *testing.T) {
	db := openMemDB(t)
	defer db.Close()

	if err := CreateTagEntitiesTable(db); err != nil {
		t.Fatalf("CreateTagEntitiesTable: %v", err)
	}

	result, err := CheckForTables(db, []string{"tag_entities"})
	if err != nil {
		t.Fatalf("CheckForTables: %v", err)
	}
	if !result["tag_entities"] {
		t.Error("expected tag_entities table to exist")
	}
}

func TestCreateNotesTable(t *testing.T) {
	db := openMemDB(t)
	defer db.Close()

	if err := CreateNotesTable(db); err != nil {
		t.Fatalf("CreateNotesTable: %v", err)
	}

	result, err := CheckForTables(db, []string{"notes"})
	if err != nil {
		t.Fatalf("CheckForTables: %v", err)
	}
	if !result["notes"] {
		t.Error("expected notes table to exist")
	}
}
