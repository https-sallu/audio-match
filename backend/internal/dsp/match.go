package dsp

import (
	"fmt"
	"math"

	"github.com/yourorg/audio-match/internal/models"
)

type MatchEngine struct{}

func NewMatchEngine() *MatchEngine {
	return &MatchEngine{}
}

func (me *MatchEngine) FindBestMatch(queryFPs []models.Fingerprint, dbMatches []models.Fingerprint) *models.MatchResult {
	if len(dbMatches) == 0 {
		return nil
	}

	queryHashMap := make(map[string][]float64)
	for _, fp := range queryFPs {
		queryHashMap[fp.Hash] = append(queryHashMap[fp.Hash], fp.AnchorTime)
	}

	type songHistogram map[float64]int
	scores := make(map[int64]songHistogram)
	totalMatches := make(map[int64]int)

	for _, dbMatch := range dbMatches {
		qTimes, exists := queryHashMap[dbMatch.Hash]
		if !exists {
			continue
		}
		songID := dbMatch.SongID
		if scores[songID] == nil {
			scores[songID] = make(songHistogram)
		}
		for _, qTime := range qTimes {
			offset := dbMatch.AnchorTime - qTime
			binnedOffset := math.Round(offset*10.0) / 10.0
			scores[songID][binnedOffset]++
			totalMatches[songID]++
		}
	}

	var bestSongID int64
	var maxScore int
	for songID, histogram := range scores {
		for _, count := range histogram {
			if count > maxScore {
				maxScore = count
				bestSongID = songID
			}
		}
	}
	fmt.Printf("DEBUG: Best Song ID: %d | Max Score: %d | Total Hashes: %d\n", bestSongID, maxScore, len(queryFPs))
	if maxScore < 5 {
		return nil
	}

	confidence := (float64(maxScore) / float64(totalMatches[bestSongID])) * 100.0
	return &models.MatchResult{
		SongID:            bestSongID,
		ConfidenceScore:   math.Round(confidence*100) / 100,
		TotalMatches:      totalMatches[bestSongID],
		OffsetConsistency: maxScore,
	}
}
