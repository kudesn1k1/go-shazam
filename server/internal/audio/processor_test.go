package audio_test

import (
	"go-shazam/internal/audio"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessAudio(t *testing.T) {
	sampleRate := 11200
	durationSeconds := 1.0
	freq := 1000.0 // 1000 Hz

	// Generate 1 second of 1000 Hz sine wave
	numSamples := int(float64(sampleRate) * durationSeconds)
	samples := make([]float64, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		samples[i] = math.Sin(2 * math.Pi * freq * t)
	}

	// We need at least WindowSize samples to process.
	// WindowSize is 4096.
	// With overlap 0.5 (2048 step), we should get:
	// Fragment 1: 0..4096
	// Fragment 2: 2048..6144
	// Fragment 3: 4096..8192
	// Fragment 4: 6144..10240
	// Fragment 5: 8192..12288 (out of bounds for 11200 samples)
	// So we expect 4 fragments.

	fragments, err := audio.ProcessAudio(samples, sampleRate)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(fragments), 1)

	frag := fragments[0]
	assert.Equal(t, audio.WindowSize, len(frag.Spectrum))
	assert.Equal(t, audio.WindowSize, len(frag.Magnitudes))

	maxMag := 0.0
	maxIndex := 0
	for i, mag := range frag.Magnitudes {
		// FFTReal usually returns full size but symmetric?
		// Let's check only up to WindowSize/2
		if i < audio.WindowSize/2 {
			if mag > maxMag {
				maxMag = mag
				maxIndex = i
			}
		}
	}

	// Calculate expected bin
	// Bin resolution = SampleRate / WindowSize = 11200 / 4096 = 2.734375 Hz
	// Expected bin = 1000 / 2.734375 = 365.7
	expectedBin := int(math.Round(freq * float64(audio.WindowSize) / float64(sampleRate)))

	assert.InDelta(t, expectedBin, maxIndex, 1, "Peak frequency bin should match input frequency")

	// Check that windowing was applied.
	// This is hard to test strictly without comparing to non-windowed version,
	// but we can assert that the peak is significant.
	assert.Greater(t, maxMag, 10.0, "Peak magnitude should be significant")
}
