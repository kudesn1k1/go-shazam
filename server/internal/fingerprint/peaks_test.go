package fingerprint

import (
	"go-shazam/internal/audio"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPeaks(t *testing.T) {
	// Simulate a fragment with a peak at specific frequency
	// Sample rate 11200, Window 4096 -> Bin size ~2.73 Hz

	// Let's put a peak at ~1000 Hz.
	// 1000 / 2.73 = 366 (approx bin index)
	binIdx := 366

	magnitudes := make([]float64, audio.WindowSize)
	magnitudes[binIdx] = 100.0 // Significant magnitude

	// Add some noise
	magnitudes[10] = 5.0
	magnitudes[500] = 2.0

	fragments := []audio.ProcessedFragment{
		{
			TimeOffset: 0.0,
			Magnitudes: magnitudes,
		},
	}

	peaks := ExtractPeaks(fragments, 11200)

	// We expect at least one peak in the band covering 1000 Hz
	// Bands: 30-40, 40-80, 80-160, 160-511, 511-2200, 2200-5000
	// 1000 Hz falls into 511-2200 (Index 4)

	found := false
	for _, p := range peaks {
		if p.BandIndex == 4 {
			found = true
			assert.InDelta(t, 1000.0, p.Frequency, 5.0) // Allow some resolution error
			assert.Equal(t, 100.0, p.Magnitude)
		}
	}

	assert.True(t, found, "Should find a peak in the 1000Hz band")
}

func TestGetBandIndex(t *testing.T) {
	// 30-40
	assert.Equal(t, 0, getBandIndex(35))
	// 40-80
	assert.Equal(t, 1, getBandIndex(50))
	// Out of range
	assert.Equal(t, -1, getBandIndex(10))
	assert.Equal(t, -1, getBandIndex(6000))
}
