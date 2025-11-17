package database

import (
	"database/sql"
	"fmt"
	"taedae/database/sqlc"
)

type Accessor struct {
	*sqlc.Queries
}

func NewAccessor() (*Accessor, error) {
	// Get database path from environment variable
	dbPath, err := GetDatabasePathFromEnv("DB_PATH")
	if err != nil {
		return nil, fmt.Errorf("failed to get database path from env: %w", err)
	}

	// Get complete database path
	dbPath, err = JoinToUserHomeDirectory(dbPath)
	if err != nil {
		return nil,
			fmt.Errorf("failed to join database path to user home directory: %w", err)
	}

	// Open the database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return &Accessor{Queries: sqlc.New(db)}, nil
}
