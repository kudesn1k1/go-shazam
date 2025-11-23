package fingerprint

import (
	"go-shazam/internal/audio"
	"sort"
)

// Frequency bands configuration
var bands = []struct {
	MinFreq float64
	MaxFreq float64
}{
	{0, 300},
	{300, 2000},
	{2000, 5000},
	{5000, 5600}, // Nyquist
}

const (
	MADMultiplier = 3.0
)

func ExtractPeaks(fragments []audio.ProcessedFragment, sampleRate int) []Peak {
	var peaks []Peak
	binSize := float64(sampleRate) / float64(audio.WindowSize)

	for _, fragment := range fragments {
		bandMagnitudes := make([][]float64, len(bands))
		bandIndices := make([][]int, len(bands))

		for i, mag := range fragment.Magnitudes {
			// Only check up to Nyquist
			if i >= audio.WindowSize/2 {
				break
			}

			freq := float64(i) * binSize
			bandIdx := getBandIndex(freq)
			if bandIdx == -1 {
				continue
			}

			bandMagnitudes[bandIdx] = append(bandMagnitudes[bandIdx], mag)
			bandIndices[bandIdx] = append(bandIndices[bandIdx], i)
		}

		// Process each band
		for b := range bands {
			mags := bandMagnitudes[b]
			if len(mags) == 0 {
				continue
			}

			// Calculate dynamic threshold using Median
			sortedMags := make([]float64, len(mags))
			copy(sortedMags, mags)
			sort.Float64s(sortedMags)

			median := sortedMags[len(sortedMags)/2]

			// 2. Calculate MAD (Median Absolute Deviation)? Or just use Median * Multiplier?
			// Robust statistical threshold often uses Median + k * MAD
			// MAD = median(|x_i - median|)
			// Let's calculate MAD
			/*
				deviations := make([]float64, len(mags))
				for k, v := range mags {
					deviations[k] = math.Abs(v - median)
				}
				sort.Float64s(deviations)
				mad := deviations[len(deviations)/2]
				threshold := median + MADMultiplier*mad
			*/

			// Simplified approach first: if max peak in band is significantly higher than median
			threshold := median * 2.0
			// Ensure threshold isn't too low (noise floor)
			if threshold < 1.0 {
				threshold = 1.0
			}

			// Find local maxima in this band that exceed threshold
			// Currently we just take the MAX peak.
			// To improve robustness, we should just pick the strongest peak that satisfies criteria.

			var bestPeak *Peak
			maxMag := -1.0

			for k, mag := range mags {
				if mag > threshold {
					if mag > maxMag {
						maxMag = mag

						originalIdx := bandIndices[b][k]
						freq := float64(originalIdx) * binSize

						bestPeak = &Peak{
							Frequency: freq,
							Magnitude: mag,
							Time:      fragment.TimeOffset,
							BandIndex: b,
						}
					}
				}
			}

			if bestPeak != nil {
				peaks = append(peaks, *bestPeak)
			}
		}
	}

	return peaks
}

func getBandIndex(freq float64) int {
	for i, band := range bands {
		if freq >= band.MinFreq && freq < band.MaxFreq {
			return i
		}
	}
	return -1
}
