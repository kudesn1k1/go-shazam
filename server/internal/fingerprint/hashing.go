package fingerprint

import (
	"fmt"
	"hash/fnv"
	"sort"

	"github.com/google/uuid"
)

const (
	TargetZoneStartOffset = 5  // Number of peaks ahead to start target zone
	TargetZoneSize        = 50 // Size of target zone (in number of peaks to check, or time)
	FanOut                = 5  // Number of nearest neighbors to pair with
)

func CreateHashes(peaks []Peak, songID uuid.UUID) []Hash {
	var hashes []Hash

	sort.Slice(peaks, func(i, j int) bool {
		return peaks[i].Time < peaks[j].Time
	})

	for i, anchor := range peaks {
		// Target Zone: [anchor.Time + 0.5s, anchor.Time + 5.0s]

		targetCount := 0
		for j := i + 1; j < len(peaks); j++ {
			target := peaks[j]
			timeDelta := target.Time - anchor.Time

			if timeDelta < 0.1 { // Too close
				continue
			}
			if timeDelta > 2.0 { // Too far
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

func generateHash(f1, f2, dt float64) uint32 {
	// Quantize frequencies and time delta to reduce sensitivity to noise
	// Frequencies are already bin-centered, but time delta needs quantization.

	// Let's bin frequencies to integer (they are float from bin center)
	// and time delta to integer (milliseconds / 10).

	freq1 := int(f1)
	freq2 := int(f2)
	timeDelta := int(dt * 100) // 10ms units

	// Create a unique string key: "f1:f2:dt"
	key := fmt.Sprintf("%d:%d:%d", freq1, freq2, timeDelta)

	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}
