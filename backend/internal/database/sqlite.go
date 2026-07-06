package database

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dsn string, migrationPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dsn+"?_fk=1&_timeout=120000")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	migrationBytes, err := os.ReadFile(migrationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read migration file: %w", err)
	}
	if _, err := db.Exec(string(migrationBytes)); err != nil {
		return nil, fmt.Errorf("failed to execute migrations: %w", err)
	}
	return db, nil
}