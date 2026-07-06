package dsp

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/yourorg/audio-match/internal/models"
)

const (
	TargetZoneTimeDelay = 3
	TargetZoneTimeWidth = 20
	FanValue            = 5
)

func GenerateFingerprints(peaks []Peak, songID int64) []models.Fingerprint {
	var fingerprints []models.Fingerprint
	frameToSeconds := float64(StepSize) / float64(ExpectedSampleRate)

	for i := 0; i < len(peaks); i++ {
		anchor := peaks[i]
		pairsCreated := 0

		for j := i + 1; j < len(peaks) && pairsCreated < FanValue; j++ {
			target := peaks[j]
			timeDelta := target.TimeFrame - anchor.TimeFrame

			if timeDelta >= TargetZoneTimeDelay && timeDelta <= (TargetZoneTimeDelay+TargetZoneTimeWidth) {
				rawHash := fmt.Sprintf("%d|%d|%d", anchor.FreqBin, target.FreqBin, timeDelta)
				hash := sha1.Sum([]byte(rawHash))
				hashStr := hex.EncodeToString(hash[:])
				anchorTimeSecs := float64(anchor.TimeFrame) * frameToSeconds

				fingerprints = append(fingerprints, models.Fingerprint{
					SongID:     songID,
					Hash:       hashStr,
					AnchorTime: anchorTimeSecs,
				})
				pairsCreated++
			}
			if timeDelta > (TargetZoneTimeDelay + TargetZoneTimeWidth) {
				break
			}
		}
	}
	return fingerprints
}