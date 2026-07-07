package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/yourorg/audio-match/internal/models"
)

type Repository interface {
	InsertSong(ctx context.Context, song *models.Song) (int64, error)
	GetSongs(ctx context.Context) ([]models.Song, error)
	GetSongByID(ctx context.Context, id int64) (*models.Song, error)
	DeleteSong(ctx context.Context, id int64) error
	BatchInsertFingerprints(ctx context.Context, fingerprints []models.Fingerprint) error
	FindMatchesByHashes(ctx context.Context, hashes []string) ([]models.Fingerprint, error)
}

type SQLiteRepo struct {
	db *sql.DB
}

func NewSQLiteRepo(db *sql.DB) *SQLiteRepo {
	return &SQLiteRepo{db: db}
}

func (r *SQLiteRepo) InsertSong(ctx context.Context, song *models.Song) (int64, error) {
	query := `INSERT INTO songs (title, artist, duration, file_path) VALUES (?, ?, ?, ?) RETURNING id`
	var id int64
	err := r.db.QueryRowContext(ctx, query, song.Title, song.Artist, song.Duration, song.FilePath).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert song: %w", err)
	}
	return id, nil
}

func (r *SQLiteRepo) GetSongs(ctx context.Context) ([]models.Song, error) {
	query := `SELECT id, title, artist, duration, file_path, created_at FROM songs ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var s models.Song
		if err := rows.Scan(&s.ID, &s.Title, &s.Artist, &s.Duration, &s.FilePath, &s.CreatedAt); err != nil {
			return nil, err
		}
		songs = append(songs, s)
	}
	return songs, nil
}

func (r *SQLiteRepo) GetSongByID(ctx context.Context, id int64) (*models.Song, error) {
	query := `SELECT id, title, artist, duration, file_path, created_at FROM songs WHERE id = ?`
	var s models.Song
	err := r.db.QueryRowContext(ctx, query, id).Scan(&s.ID, &s.Title, &s.Artist, &s.Duration, &s.FilePath, &s.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *SQLiteRepo) DeleteSong(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM songs WHERE id = ?", id)
	return err
}

// Keep YOUR exact function signature here (the variables might be named slightly differently)
func (r *SQLiteRepo) BatchInsertFingerprints(ctx context.Context, fingerprints []models.Fingerprint) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		log.Printf("🔥 DB Error (BeginTx): %v\n", err)
		return err
	}

	// 🚨 IMPORTANT: I changed 'offset' to 'anchor_time' here to match your struct!
	// If your database actually uses the word 'offset', change it back.
	stmt, err := tx.PrepareContext(ctx, "INSERT INTO fingerprints (song_id, hash, anchor_time) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("🔥 DB Error (Prepare): %v\n", err)
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, fp := range fingerprints {
		_, err = stmt.ExecContext(ctx, fp.SongID, fp.Hash, fp.AnchorTime)
		if err != nil {
			log.Printf("🔥 DB Error (Exec): %v\n", err)
			tx.Rollback()
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		log.Printf("🔥 DB Error (Commit): %v\n", err)
		return err
	}
	return nil
}

func (r *SQLiteRepo) FindMatchesByHashes(ctx context.Context, hashes []string) ([]models.Fingerprint, error) {
	if len(hashes) == 0 {
		return nil, nil
	}
	const chunkSize = 1000
	var results []models.Fingerprint

	for i := 0; i < len(hashes); i += chunkSize {
		end := i + chunkSize
		if end > len(hashes) {
			end = len(hashes)
		}
		chunk := hashes[i:end]
		placeholders := make([]string, len(chunk))
		args := make([]interface{}, len(chunk))
		for j, h := range chunk {
			placeholders[j] = "?"
			args[j] = h
		}
		query := fmt.Sprintf(`SELECT song_id, hash, anchor_time FROM fingerprints WHERE hash IN (%s)`, strings.Join(placeholders, ","))
		rows, err := r.db.QueryContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var fp models.Fingerprint
			if err := rows.Scan(&fp.SongID, &fp.Hash, &fp.AnchorTime); err == nil {
				results = append(results, fp)
			}
		}
		rows.Close()
	}
	return results, nil
}
