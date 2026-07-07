package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yourorg/audio-match/internal/dsp"
	"github.com/yourorg/audio-match/internal/models"
	"github.com/yourorg/audio-match/internal/repository"
)

type API struct {
	repo *repository.SQLiteRepo
}

func NewAPI(repo *repository.SQLiteRepo) *API {
	return &API{repo: repo}
}

func (api *API) HandleListSongs(w http.ResponseWriter, r *http.Request) {
	songs, err := api.repo.GetSongs(r.Context())
	if err != nil {
		// Log removed from here, it was the wrong endpoint!
		http.Error(w, "Failed to fetch songs", http.StatusInternalServerError)
		return
	}
	if songs == nil {
		songs = []models.Song{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(songs)
}

func (api *API) HandleGetSong(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}
	song, err := api.repo.GetSongByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	if song == nil {
		http.Error(w, "Song not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(song)
}

func (api *API) HandleDeleteSong(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid song ID", http.StatusBadRequest)
		return
	}
	song, _ := api.repo.GetSongByID(r.Context(), id)
	if song != nil {
		os.Remove(song.FilePath)
	}
	if err := api.repo.DeleteSong(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete song", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (api *API) HandleImport(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	title := r.FormValue("title")
	artist := r.FormValue("artist")
	if title == "" || artist == "" {
		http.Error(w, "Title and artist are required", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Audio file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	datasetDir := "./dataset"
	os.MkdirAll(datasetDir, os.ModePerm)
	filename := fmt.Sprintf("%d_%s", time.Now().Unix(), header.Filename)
	filePath := filepath.Join(datasetDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	audioData, err := dsp.ReadMonoWAV(filePath)
	if err != nil {
		os.Remove(filePath)
		http.Error(w, "Invalid audio format: "+err.Error(), http.StatusBadRequest)
		return
	}

	duration := float64(len(audioData)) / float64(dsp.ExpectedSampleRate)
	song := &models.Song{Title: title, Artist: artist, Duration: duration, FilePath: filePath}

	id, err := api.repo.InsertSong(r.Context(), song)
	if err != nil {
		// X-RAY LOG 1: Checks if inserting the basic song data fails
		log.Printf("🚨 UPLOAD CRASH REASON (InsertSong): %v\n", err)
		http.Error(w, "Failed to insert song", http.StatusInternalServerError)
		return
	}
	song.ID = id

	spectrogram := dsp.GenerateSpectrogram(audioData)
	peaks := dsp.ExtractPeaks(spectrogram)
	fingerprints := dsp.GenerateFingerprints(peaks, song.ID)

	if err := api.repo.BatchInsertFingerprints(r.Context(), fingerprints); err != nil {
		// X-RAY LOG 2: Checks if the mass fingerprint insertion fails
		log.Printf("🚨 UPLOAD CRASH REASON (BatchInsert): %v\n", err)
		api.repo.DeleteSong(r.Context(), song.ID)
		http.Error(w, "Failed to save fingerprints", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(song)
}

func (api *API) HandleRecognize(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Audio file required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	tempFile := filepath.Join(os.TempDir(), fmt.Sprintf("query_%d.wav", time.Now().UnixNano()))
	dst, err := os.Create(tempFile)
	if err != nil {
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}
	io.Copy(dst, file)
	dst.Close()
	defer os.Remove(tempFile)

	audioData, err := dsp.ReadMonoWAV(tempFile)
	if err != nil {
		http.Error(w, "Invalid audio format", http.StatusBadRequest)
		return
	}

	spectrogram := dsp.GenerateSpectrogram(audioData)
	peaks := dsp.ExtractPeaks(spectrogram)
	queryFPs := dsp.GenerateFingerprints(peaks, 0)

	hashStrings := make([]string, len(queryFPs))
	for i, fp := range queryFPs {
		hashStrings[i] = fp.Hash
	}

	dbMatches, err := api.repo.FindMatchesByHashes(r.Context(), hashStrings)
	if err != nil {
		http.Error(w, "Database query failed", http.StatusInternalServerError)
		return
	}

	engine := dsp.NewMatchEngine()
	result := engine.FindBestMatch(queryFPs, dbMatches)

	if result == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "No match found"})
		return
	}

	song, err := api.repo.GetSongByID(r.Context(), result.SongID)
	if err == nil && song != nil {
		result.Title = song.Title
		result.Artist = song.Artist
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}
