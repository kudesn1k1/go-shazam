package fingerprint

import (
	"go-shazam/internal/audio"
	"sort"

	"github.com/google/uuid"
)

const (
	FanOut       = 5
	MinTimeDelta = 0.1
	MaxTimeDelta = 3.0
	TimeDeltaRes = 100
)

func CreateHashes(peaks []Peak, songID uuid.UUID) []Hash {
	var hashes []Hash

	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].Time < peaks[j].Time
	})

	for i, anchor := range peaks {
		targetCount := 0

		for j := i + 1; j < len(peaks); j++ {
			target := peaks[j]
			timeDelta := target.Time - anchor.Time

			// Skip peaks that are too close
			if timeDelta < MinTimeDelta {
				continue
			}

			// Stop looking if we are too far ahead
			if timeDelta > MaxTimeDelta {
				break
			}

			hashes = append(hashes, Hash{
				HashValue:  generateHash(anchor.Frequency, target.Frequency, timeDelta),
				SongID:     songID,
				TimeOffset: anchor.Time,
			})

			targetCount++
			if targetCount >= FanOut {
				break
			}
		}
	}

	return hashes
}

// generateHash creates a 64-bit hash using bit packing
// Layout: [freq1: 10 bits][freq2: 10 bits][timeDelta: 14 bits]
func generateHash(f1, f2, dt float64) int64 {
	// At 11200 Hz sample rate with 2048 window, bin size = 11200/2048 ≈ 5.47 Hz
	const binSize = float64(audio.TargetSampleRate) / float64(audio.WindowSize)

	freq1Bin := int(f1/binSize) & 0x3FF
	freq2Bin := int(f2/binSize) & 0x3FF
	timeDeltaBin := int(dt*TimeDeltaRes) & 0x3FFF

	return int64(freq1Bin)<<24 | int64(freq2Bin)<<14 | int64(timeDeltaBin)
}
