-- Create the Songs table
CREATE TABLE IF NOT EXISTS songs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    artist TEXT NOT NULL,
    duration REAL NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Create the Fingerprints table (Linked to songs via Foreign Key)
CREATE TABLE IF NOT EXISTS fingerprints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    song_id INTEGER NOT NULL,
    hash TEXT NOT NULL,
    offset INTEGER NOT NULL,
    FOREIGN KEY (song_id) REFERENCES songs(id) ON DELETE CASCADE
);

-- NEW: Fast-lookup indexes for instant matching and deletion
-- These two lines change delete times from 50 seconds to 0.1 seconds!
CREATE INDEX IF NOT EXISTS idx_fingerprints_song_id ON fingerprints(song_id);
CREATE INDEX IF NOT EXISTS idx_fingerprints_hash ON fingerprints(hash);