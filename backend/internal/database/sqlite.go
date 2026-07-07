package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	// Safe speed boost
	db.Exec("PRAGMA journal_mode=WAL;")

	// --- THE FIX: AUTO-CREATE TABLES ---
	// If the database is ever empty or deleted, this instantly rebuilds the structure.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			artist TEXT NOT NULL,
			duration REAL NOT NULL,
			file_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS fingerprints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			song_id INTEGER NOT NULL,
			hash TEXT NOT NULL,
			anchor_time REAL NOT NULL,
			FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
		);

		CREATE INDEX IF NOT EXISTS idx_fingerprints_song_id ON fingerprints(song_id);
		CREATE INDEX IF NOT EXISTS idx_fingerprints_hash ON fingerprints(hash);
	`)

	if err != nil {
		log.Printf("🚨 FATAL TABLE CREATION ERROR: %v\n", err)
		return nil, err
	}

	log.Println("Database connection established and tables verified.")
	return db, nil
}
