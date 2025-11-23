package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"math/cmplx"
	"os"

	"os/exec"

	"github.com/go-audio/wav"
	"github.com/mjibson/go-dsp/fft"
	"github.com/mjibson/go-dsp/window"
)

const (
	WindowSize       = 2048
	Overlap          = 0.50 // 50% overlap
	TargetSampleRate = 11200
)

// ProcessedFragment represents the result of processing a single audio fragment
type ProcessedFragment struct {
	TimeOffset float64
	Spectrum   []complex128
	Magnitudes []float64
}

// Resample resamples the input audio from oldRate to newRate using ffmpeg.
// It uses pipes to avoid writing files to disk.
func Resample(input []float64, oldRate, newRate int) ([]float64, error) {
	if oldRate == newRate {
		return input, nil
	}

	inputBuf := new(bytes.Buffer)
	for _, s := range input {
		if err := binary.Write(inputBuf, binary.LittleEndian, float32(s)); err != nil {
			return nil, fmt.Errorf("binary write failed: %w", err)
		}
	}

	cmd := exec.Command(
		"ffmpeg",
		"-f", "f32le",
		"-ar", fmt.Sprint(oldRate),
		"-ac", "1",
		"-i", "pipe:0",
		"-ar", fmt.Sprint(newRate),
		"-ac", "1",
		"-f", "f32le",
		"pipe:1",
		"-loglevel", "error",
		"-nostats",
	)

	cmd.Stdin = inputBuf

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg failed: %v, stderr: %s", err, stderr.String())
	}

	outBytes := stdout.Bytes()
	if len(outBytes)%4 != 0 {
		return nil, fmt.Errorf("invalid ffmpeg output length: %d", len(outBytes))
	}

	out := make([]float64, len(outBytes)/4)
	for i := 0; i < len(out); i++ {
		bits := binary.LittleEndian.Uint32(outBytes[i*4 : (i+1)*4])
		out[i] = float64(math.Float32frombits(bits))
	}

	return out, nil
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
