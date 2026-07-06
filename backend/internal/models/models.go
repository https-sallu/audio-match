package models

import "time"

type Song struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Artist    string    `json:"artist"`
	Duration  float64   `json:"duration"`
	FilePath  string    `json:"file_path"`
	CreatedAt time.Time `json:"created_at"`
}

type Fingerprint struct {
	ID         int64   `json:"id"`
	SongID     int64   `json:"song_id"`
	Hash       string  `json:"hash"`
	AnchorTime float64 `json:"anchor_time"`
}

type MatchResult struct {
	SongID            int64   `json:"song_id"`
	Title             string  `json:"title"`
	Artist            string  `json:"artist"`
	ConfidenceScore   float64 `json:"confidence_score"`
	TotalMatches      int     `json:"total_matches"`
	OffsetConsistency int     `json:"offset_consistency"`
}