package database

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// ##################################################################
// ######################## Manual migration ########################
// ##################################################################

// CreateNotesTable creates the notes table in the database
// CREATE TABLE notes (
//
//	id INTEGER PRIMARY KEY AUTOINCREMENT,
//	todo_id INTEGER NOT NULL,
//	content TEXT NOT NULL,
//	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	deleted_at DATETIME
//
// );
func CreateNotesTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo_id INTEGER NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME
	);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

// CreateTagsTable creates the tags table in the database
// CREATE TABLE tags (
//
//	id INTEGER PRIMARY KEY AUTOINCREMENT,
//	name TEXT NOT NULL UNIQUE,
//	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	deleted_at DATETIME
//
// );
func CreateTagsTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tags (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME
	);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

// CreateTagEntitiesTable creates the tag_entities table in the database
// CREATE TABLE tag_entities (
//
//	id INTEGER PRIMARY KEY AUTOINCREMENT,
//	todo_id INTEGER NOT NULL,
//	tag_id INTEGER NOT NULL,
//	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	deleted_at DATETIME
//
// );
func CreateTagEntitiesTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS tag_entities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		todo_id INTEGER NOT NULL,
		tag_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME
	);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

// CreateTodosTable creates the todos table in the database
// CREATE TABLE todos (
//
//	id INTEGER PRIMARY KEY AUTOINCREMENT,
//	title TEXT NOT NULL,
//	description TEXT,
//	completed BOOLEAN NOT NULL DEFAULT 0,
//	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	deleted_at DATETIME,
//	due_date DATETIME
//
// );
func CreateTodosTable(db *sql.DB) error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description TEXT,
		completed BOOLEAN NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		deleted_at DATETIME,
		due_date DATETIME
	);
	`
	_, err := db.Exec(createTableSQL)
	return err
}

func RunManualMigrations() (bool, error) {
	// Error tracking variable
	var errors string

	// State tracking variable
	var ranMigrations bool

	// Get database path from environment
	dbPath, err := GetDatabasePathFromEnv("DB_PATH")
	if err != nil {
		return false, fmt.Errorf("failed to get database path from environment: %w", err)
	}

	// Join to user home directory
	dbPath, err = JoinToUserHomeDirectory(dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to join database path to user home directory: %w", err)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// TODO: Check current schema version and apply only necessary migrations
	tablesExist, err := CheckForTables(db, []string{
		"todos",
		"tags",
		"tag_entities",
		"notes",
	})
	if err != nil {
		return false, fmt.Errorf("failed to check for existing tables: %w", err)
	}

	// Create tables
	// -> Todos
	if !tablesExist["todos"] {
		ranMigrations = true
		err = CreateTodosTable(db)
		if err != nil {
			errors += fmt.Sprintf("failed to create todos table: %v\n", err.Error())
		}
	}

	// -> Tags
	if !tablesExist["tags"] {
		ranMigrations = true
		err = CreateTagsTable(db)
		if err != nil {
			errors += fmt.Sprintf("failed to create tags table: %v\n", err.Error())
		}
	}

	// -> TagEntities
	if !tablesExist["tag_entities"] {
		ranMigrations = true
		err = CreateTagEntitiesTable(db)
		if err != nil {
			errors += fmt.Sprintf("failed to create tag_entities table: %v\n", err.Error())
		}
	}

	// -> Notes
	if !tablesExist["notes"] {
		ranMigrations = true
		err = CreateNotesTable(db)
		if err != nil {
			errors += fmt.Sprintf("failed to create notes table: %v\n", err.Error())
		}
	}

	if errors != "" {
		return false, fmt.Errorf("manual migration failed with errors:\n%s", errors)
	}

	return ranMigrations, nil
}

// GetWindowsMigrationPath returns the path to the migration files
// It expects the migrations to be in a "migrations" folder relative to the database
func GetWindowsMigrationPath(dbPath string) string {
	dbDir := filepath.Dir(dbPath)
	return filepath.Join(dbDir, "migrations")
}

func GetMigrationCompatiblePath(dbPath string) string {
	// Backslashes need to be converted to forward slashes for migration tool compatibility
	return filepath.ToSlash(GetWindowsMigrationPath(dbPath))
}

// IsDatabaseUpToDate checks if the database is at the latest migration version
func IsDatabaseUpToDate(dbPath string) (bool, error) {
	// Check if database exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return false, fmt.Errorf("database does not exist at path: %s", dbPath)
	}

	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return false, fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create migrate instance
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return false, fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrationPath := GetMigrationCompatiblePath(dbPath)
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file:///%s", migrationPath),
		"sqlite3",
		driver,
	)
	if err != nil {
		return false, fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Get current version
	currentVersion, dirty, err := m.Version()
	if err != nil {
		// If no migration has been applied yet, consider it up to date if no migrations exist
		if err == migrate.ErrNilVersion {
			// Check if migration files exist
			if _, err := os.Stat(migrationPath); os.IsNotExist(err) {
				return true, nil // No migrations exist, so it's "up to date"
			}
			return false, nil // Migrations exist but haven't been applied
		}
		return false, fmt.Errorf("failed to get migration version: %w", err)
	}

	if dirty {
		return false, fmt.Errorf("database is in a dirty state at version %d", currentVersion)
	}

	// Check if there are pending migrations
	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			return true, nil // Already up to date
		}
		return false, fmt.Errorf("failed to check for pending migrations: %w", err)
	}

	return false, nil // There were migrations to apply
}

// RunMigrations applies all pending migrations to the database
func RunMigrations(dbPath string) error {
	// Open database connection
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	// Create migrate instance
	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	// migrationPath := GetMigrationCompatiblePath(dbPath)
	// migrationPath := "migrations" // It's fucked right now, for some reason...
	m, err := migrate.NewWithDatabaseInstance(
		"migrations",
		"sqlite3",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	// Apply migrations
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// UpdateIfNotAtLatest checks if the database is up to date and runs migrations if needed
func UpdateIfNotAtLatest(dbPath string) error {
	upToDate, err := IsDatabaseUpToDate(dbPath)
	if err != nil {
		return fmt.Errorf("failed to check if database is up to date: %w", err)
	}

	if !upToDate {
		if err := RunMigrations(dbPath); err != nil {
			return fmt.Errorf("failed to update database: %w", err)
		}
	}

	return nil
}

// EnsureMigrationsDirectoryExists checks if migrations folder exists at the DB path location
// If not, it creates the folder and copies migration files from the project's migrations folder
func EnsureMigrationsDirectoryExists(dbPath string) error {
	// Get the target migrations path (where DB is located)
	targetMigrationPath := GetWindowsMigrationPath(dbPath)

	// Check if migrations folder already exists
	if _, err := os.Stat(targetMigrationPath); err == nil {
		// Migrations folder exists, we're done
		return nil
	}

	// Create the migrations directory
	if err := os.MkdirAll(targetMigrationPath, 0755); err != nil {
		return fmt.Errorf("failed to create migrations directory: %w", err)
	}

	// Find the source migrations folder (in project root)
	sourceMigrationPath, err := findProjectMigrationsFolder()
	if err != nil {
		return fmt.Errorf("failed to find project migrations folder: %w", err)
	}

	// Copy migration files from source to target
	if err := copyMigrationFiles(sourceMigrationPath, targetMigrationPath); err != nil {
		return fmt.Errorf("failed to copy migration files: %w", err)
	}

	return nil
}

// findProjectMigrationsFolder locates the migrations folder in the project
func findProjectMigrationsFolder() (string, error) {
	// Get current working directory
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Look for migrations folder in current directory and parent directories
	currentDir := wd
	for {
		migrationPath := filepath.Join(currentDir, "migrations")
		if _, err := os.Stat(migrationPath); err == nil {
			return migrationPath, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			// Reached root directory
			break
		}
		currentDir = parentDir
	}

	return "", fmt.Errorf("migrations folder not found in project")
}

// copyMigrationFiles copies all migration files from source to destination
func copyMigrationFiles(sourcePath, destPath string) error {
	// Read all files in source directory
	files, err := os.ReadDir(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source migrations directory: %w", err)
	}

	// Copy each migration file
	for _, file := range files {
		if file.IsDir() {
			continue // Skip directories
		}

		// Only copy .sql files
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		sourceFile := filepath.Join(sourcePath, file.Name())
		destFile := filepath.Join(destPath, file.Name())

		if err := copyFile(sourceFile, destFile); err != nil {
			return fmt.Errorf("failed to copy migration file %s: %w", file.Name(), err)
		}
	}

	return nil
}

// copyFile copies a single file from source to destination
func copyFile(sourcePath, destPath string) error {
	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy file contents
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}
