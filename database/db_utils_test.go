package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestCheckForTables_Empty(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory DB: %v", err)
	}
	defer db.Close()

	result, err := CheckForTables(db, []string{"todos", "tags"})
	if err != nil {
		t.Fatalf("CheckForTables: %v", err)
	}
	for table, exists := range result {
		if exists {
			t.Errorf("expected table %q to not exist on empty DB", table)
		}
	}
}

func TestCheckForTables_WithTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory DB: %v", err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE todos (id INTEGER PRIMARY KEY)`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	result, err := CheckForTables(db, []string{"todos", "tags"})
	if err != nil {
		t.Fatalf("CheckForTables: %v", err)
	}
	if !result["todos"] {
		t.Error("expected 'todos' to exist")
	}
	if result["tags"] {
		t.Error("expected 'tags' to not exist")
	}
}
