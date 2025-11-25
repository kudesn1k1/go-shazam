package fingerprint

import (
	"go-shazam/internal/audio"
	"math"
	"sort"
)

// Frequency bands configuration
var bands = []struct {
	MinFreq float64
	MaxFreq float64
}{
	{80, 400},
	{400, 1600},
	{1600, 3200},
	{3200, 5600},
}

const (
	MADMultiplier   = 3.0
	MaxPeaksPerBand = 1
	MinMagnitude    = 1.0
)

type peakCandidate struct {
	frequency float64
	magnitude float64
	binIndex  int
}

func ExtractPeaks(fragments []audio.ProcessedFragment, sampleRate int) []Peak {
	var peaks []Peak
	binSize := float64(sampleRate) / float64(audio.WindowSize)

	for _, fragment := range fragments {
		bandCandidates := make([][]peakCandidate, len(bands))

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

			bandCandidates[bandIdx] = append(bandCandidates[bandIdx], peakCandidate{
				frequency: freq,
				magnitude: mag,
				binIndex:  i,
			})
		}

		// Process each band
		for b := range bands {
			candidates := bandCandidates[b]
			if len(candidates) == 0 {
				continue
			}

			threshold := calculateMADThreshold(candidates)

			localMaxima := findLocalMaxima(candidates, fragment.Magnitudes, threshold)

			if len(localMaxima) == 0 {
				continue
			}

			// Sort by magnitude and take only the strongest peak
			sort.Slice(localMaxima, func(i, j int) bool {
				return localMaxima[i].magnitude > localMaxima[j].magnitude
			})

			count := min(len(localMaxima), MaxPeaksPerBand)
			for i := 0; i < count; i++ {
				peaks = append(peaks, Peak{
					Frequency: localMaxima[i].frequency,
					Magnitude: localMaxima[i].magnitude,
					Time:      fragment.TimeOffset,
					BandIndex: b,
				})
			}
		}
	}

	return peaks
}

// calculateMADThreshold computes threshold using Median Absolute Deviation
func calculateMADThreshold(candidates []peakCandidate) float64 {
	if len(candidates) == 0 {
		return MinMagnitude
	}

	mags := make([]float64, len(candidates))
	for i, c := range candidates {
		mags[i] = c.magnitude
	}

	sorted := make([]float64, len(mags))
	copy(sorted, mags)
	sort.Float64s(sorted)
	median := sorted[len(sorted)/2]

	deviations := make([]float64, len(mags))
	for i, m := range mags {
		deviations[i] = math.Abs(m - median)
	}
	sort.Float64s(deviations)
	mad := deviations[len(deviations)/2]

	threshold := median + MADMultiplier*mad

	if threshold < MinMagnitude {
		threshold = MinMagnitude
	}

	return threshold
}

// findLocalMaxima finds peaks that are local maxima in the spectrum
func findLocalMaxima(candidates []peakCandidate, magnitudes []float64, threshold float64) []peakCandidate {
	var maxima []peakCandidate

	for _, c := range candidates {
		if c.magnitude < threshold {
			continue
		}

		isMaximum := true
		idx := c.binIndex

		if idx > 0 && magnitudes[idx-1] >= c.magnitude {
			isMaximum = false
		}

		if idx < len(magnitudes)-1 && magnitudes[idx+1] >= c.magnitude {
			isMaximum = false
		}

		if isMaximum {
			maxima = append(maxima, c)
		}
	}

	return maxima
}

func getBandIndex(freq float64) int {
	for i, band := range bands {
		if freq >= band.MinFreq && freq < band.MaxFreq {
			return i
		}
	}
	return -1
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
