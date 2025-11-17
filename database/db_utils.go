package database

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

var (
	envLoadOnce sync.Once
	envLoadErr  error
)

// loadEnvOnce ensures environment variables are loaded only once
func loadEnvOnce() {
	envLoadOnce.Do(func() {
		envLoadErr = godotenv.Load()
	})
}

func GetDatabasePathFromEnv(envVar string) (string, error) {
	loadEnvOnce()
	if envLoadErr != nil {
		return "", envLoadErr
	}

	return os.Getenv(envVar), nil
}

func JoinToUserHomeDirectory(relativePath string) (string, error) {
	userHomeDir, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s%s",
			userHomeDir,
			string(os.PathSeparator),
			relativePath,
		),
		nil
}

func CheckForTables(db *sql.DB, tableNames []string) (map[string]bool, error) {
	results := make(map[string]bool)
	for _, tableName := range tableNames {
		var count int
		query := `SELECT COUNT(name) FROM sqlite_master WHERE type='table' AND name=?;`
		err := db.QueryRow(query, tableName).Scan(&count)
		if err != nil {
			return nil, err
		}
		results[tableName] = (count > 0)
	}
	return results, nil
}
