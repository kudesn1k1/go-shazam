package converter

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

func ConvertToWav(inputPath string, sampleRate int) (string, error) {
	outputPath := strings.TrimSuffix(inputPath, filepath.Ext(inputPath)) + ".wav"

	// ffmpeg -i input.mp3 -ar 11205 -ac 1 output.wav
	// -ar sets the sample rate
	// -ac 1 sets mono channel
	// -y overwrites output file
	cmd := exec.Command("ffmpeg", "-y", "-i", inputPath, "-ar", fmt.Sprintf("%d", sampleRate), "-ac", "1", outputPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffmpeg conversion failed: %s, output: %s", err, string(output))
	}

	return outputPath, nil
}
