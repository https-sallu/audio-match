package repository

import (
	"context"
	"database/sql"
	"fmt"
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
	_, err := r.db.ExecContext(ctx, `DELETE FROM songs WHERE id = ?`, id)
	return err
}

func (r *SQLiteRepo) BatchInsertFingerprints(ctx context.Context, fingerprints []models.Fingerprint) error {
	if len(fingerprints) == 0 {
		return nil
	}
	const chunkSize = 10000
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for i := 0; i < len(fingerprints); i += chunkSize {
		end := i + chunkSize
		if end > len(fingerprints) {
			end = len(fingerprints)
		}
		chunk := fingerprints[i:end]
		valueStrings := make([]string, 0, len(chunk))
		valueArgs := make([]interface{}, 0, len(chunk)*3)

		for _, fp := range chunk {
			valueStrings = append(valueStrings, "(?, ?, ?)")
			valueArgs = append(valueArgs, fp.SongID, fp.Hash, fp.AnchorTime)
		}
		query := fmt.Sprintf("INSERT INTO fingerprints (song_id, hash, anchor_time) VALUES %s", strings.Join(valueStrings, ","))
		if _, err := tx.ExecContext(ctx, query, valueArgs...); err != nil {
			return err
		}
	}
	return tx.Commit()
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