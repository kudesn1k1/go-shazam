package fingerprint

import (
	"go-shazam/internal/audio"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractPeaks(t *testing.T) {
	// SampleRate 11200 with WindowSize 2048 → bin size ≈ 5.47 Hz.
	// Place a peak at ≈1001 Hz (bin 183). 1001 Hz falls into band 1 (400-1600).
	binIdx := 183

	magnitudes := make([]float64, audio.WindowSize)
	magnitudes[binIdx] = 100.0

	// Noise that should not become peaks: bin 10 (~55 Hz, below the 80 Hz cutoff
	// so getBandIndex returns -1), bin 250 (~1367 Hz, in band 1 but small mag).
	magnitudes[10] = 5.0
	magnitudes[250] = 2.0

	fragments := []audio.ProcessedFragment{
		{TimeOffset: 0.0, Magnitudes: magnitudes},
	}

	peaks := ExtractPeaks(fragments, 11200)

	found := false
	for _, p := range peaks {
		if p.BandIndex == 1 {
			found = true
			assert.InDelta(t, 1001.0, p.Frequency, 6.0)
			assert.Equal(t, 100.0, p.Magnitude)
		}
	}

	assert.True(t, found, "should find a peak in band 1 (400-1600 Hz)")
}

func TestGetBandIndex(t *testing.T) {
	// Bands (peaks.go): 80-400, 400-1600, 1600-3200, 3200-5600
	assert.Equal(t, 0, getBandIndex(100))  // first band (80-400)
	assert.Equal(t, 1, getBandIndex(800))  // second band (400-1600)
	assert.Equal(t, 2, getBandIndex(2000)) // third band (1600-3200)
	assert.Equal(t, 3, getBandIndex(4000)) // fourth band (3200-5600)

	// Out of range
	assert.Equal(t, -1, getBandIndex(50))   // below 80 Hz
	assert.Equal(t, -1, getBandIndex(6000)) // above 5600 Hz
}
