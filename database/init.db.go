package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// DatabaseBExists checks if a database file exists at the specified path
// The path is constructed using the user's home directory and an environment variable
func DatabaseBExists(envVarName string) (bool, string, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return false, "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	relativePath := os.Getenv(envVarName)
	if relativePath == "" {
		return false, "", fmt.Errorf("environment variable %s is not set", envVarName)
	}

	dbPath := filepath.Join(userHomeDir, relativePath)

	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		return false, dbPath, nil
	}
	if err != nil {
		return false, dbPath, fmt.Errorf("error checking database file: %w", err)
	}

	return true, dbPath, nil
}

// CreateDatabase creates a new SQLite database at the specified path
func CreateDatabase(dbPath string) error {
	// Ensure the directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	// Create the database file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}
	defer db.Close()

	// Test the connection
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	return nil
}

// CreateIfNotExists checks if the database exists and creates it if it doesn't
func CreateIfNotExists(envVarName string) (bool, error) {
	// Database state tracking variable
	var databaseCreated bool = false

	// Check if database exists
	exists, dbPath, err := DatabaseBExists(envVarName)
	if err != nil {
		return false, err
	}

	if !exists {
		if err := CreateDatabase(dbPath); err != nil {
			return false, fmt.Errorf("failed to create database at %s: %w", dbPath, err)
		}
		databaseCreated = true
	}

	return databaseCreated, nil
}
