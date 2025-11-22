package audio

import (
	"fmt"
	"math/cmplx"
	"os"

	"github.com/go-audio/wav"
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

const (
	WindowSize       = 4096
	Overlap          = 0.5 // 50% overlap
	TargetSampleRate = 11200
)

// ProcessedFragment represents the result of processing a single audio fragment
type ProcessedFragment struct {
	TimeOffset float64
	Spectrum   []complex128
	Magnitudes []float64
}

// Resample resamples the input audio from oldRate to newRate using linear interpolation.
func Resample(input []float64, oldRate, newRate int) []float64 {
	if oldRate == newRate {
		return input
	}

	ratio := float64(oldRate) / float64(newRate)
	newLength := int(float64(len(input)) / ratio)
	output := make([]float64, newLength)

	for i := 0; i < newLength; i++ {
		pos := float64(i) * ratio
		index := int(pos)
		frac := pos - float64(index)

		if index+1 < len(input) {
			output[i] = input[index]*(1-frac) + input[index+1]*frac
		} else {
			output[i] = input[index]
		}
	}

	return output
}

// Returns the audio samples and sample rate
func LoadWav(path string) ([]float64, int, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	decoder := wav.NewDecoder(f)
	if !decoder.IsValidFile() {
		return nil, 0, fmt.Errorf("invalid wav file")
	}

	buf, err := decoder.FullPCMBuffer() //consider PCMBuffer not to load the entire file into memory
	if err != nil {
		return nil, 0, fmt.Errorf("failed to decode wav: %w", err)
	}

	if buf.Format.NumChannels != 1 {
		return nil, 0, fmt.Errorf("only mono audio is supported, got %d channels", buf.Format.NumChannels)
	}

	// Convert int buffer to float64
	floats := make([]float64, len(buf.Data))
	// We need to know the bit depth to normalize, but for FFT relative values matter.
	// Usually we normalize to [-1, 1].
	// buf.SourceBitDepth
	factor := 1.0
	switch buf.SourceBitDepth {
	case 8:
		factor = 128.0
	case 16:
		factor = 32768.0
	case 24:
		factor = 8388608.0
	case 32:
		factor = 2147483648.0
	}

	for i, sample := range buf.Data {
		floats[i] = float64(sample) / factor
	}

	return floats, buf.Format.SampleRate, nil
}

// ProcessAudio processes the audio samples: chunks, applies Hamming window, and performs FFT.
func ProcessAudio(samples []float64, sampleRate int) ([]ProcessedFragment, error) {
	step := int(WindowSize * (1 - Overlap))
	if step == 0 {
		step = 1 // Avoid infinite loop if WindowSize is small
	}

	var fragments []ProcessedFragment

	win := window.Hamming(WindowSize)

	for i := 0; i <= len(samples)-WindowSize; i += step {
		chunk := samples[i : i+WindowSize]

		// Apply window
		windowedChunk := make([]float64, WindowSize)
		for j := 0; j < WindowSize; j++ {
			windowedChunk[j] = chunk[j] * win[j]
		}

		spectrum := fft.FFTReal(windowedChunk)

		magnitudes := make([]float64, len(spectrum))
		for j, val := range spectrum {
			magnitudes[j] = cmplx.Abs(val)
		}

		// Store result
		fragments = append(fragments, ProcessedFragment{
			TimeOffset: float64(i) / float64(sampleRate),
			Spectrum:   spectrum,
			Magnitudes: magnitudes,
		})
	}

	return fragments, nil
}
