package fingerprint

import (
	"go-shazam/internal/audio"
)

// Frequency bands configuration
// We use logarithmic bands to ensure we capture features across the spectrum
var bands = []struct {
	MinFreq float64
	MaxFreq float64
}{
	{30, 40},
	{40, 80},
	{80, 160},
	{160, 511},
	{511, 2200},
	{2200, 5000},
}

const (
	ThresholdMultiplier = 1.5 // Minimum magnitude multiplier above average to be considered a peak
	MinPeakMagnitude    = 100.0
)

func ExtractPeaks(fragments []audio.ProcessedFragment, sampleRate int) []Peak {
	var peaks []Peak
	binSize := float64(sampleRate) / float64(audio.WindowSize)

	for _, fragment := range fragments {
		// Calculate average magnitude per band for this fragment
		bandAverages := make([]float64, len(bands))
		bandCounts := make([]int, len(bands))
		maxPeakInBand := make([]*Peak, len(bands))

		// Calculate averages
		for i, mag := range fragment.Magnitudes {
			// Only check up to Nyquist
			if i >= audio.WindowSize/2 {
				break
			}

			// Skip weak signals to avoid noise fingerprints
			// Cause silent magnitudes affected the average and the peak detection
			if mag < MinPeakMagnitude {
				continue
			}

			freq := float64(i) * binSize
			bandIdx := getBandIndex(freq)
			if bandIdx == -1 {
				continue
			}

			bandAverages[bandIdx] += mag
			bandCounts[bandIdx]++
		}

		for i := range bandAverages {
			if bandCounts[i] > 0 {
				bandAverages[i] /= float64(bandCounts[i])
			}
		}

		// Find strongest peak in each band
		for i, mag := range fragment.Magnitudes {
			if i >= audio.WindowSize/2 {
				break
			}

			freq := float64(i) * binSize
			bandIdx := getBandIndex(freq)
			if bandIdx == -1 {
				continue
			}

			// Check if this magnitude is a candidate (above dynamic threshold)
			if mag > bandAverages[bandIdx]*ThresholdMultiplier {
				if maxPeakInBand[bandIdx] == nil || mag > maxPeakInBand[bandIdx].Magnitude {
					maxPeakInBand[bandIdx] = &Peak{
						Frequency: freq,
						Magnitude: mag,
						Time:      fragment.TimeOffset,
						BandIndex: bandIdx,
					}
				}
			}
		}

		for _, peak := range maxPeakInBand {
			if peak != nil {
				peaks = append(peaks, *peak)
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
