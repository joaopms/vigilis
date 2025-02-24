package recorders

import (
	"os/exec"
	"regexp"
	"vigilis/internal/config"
	"vigilis/internal/logger"
)

type FfmpegConfig struct {
	Path string
}

var Ffmpeg FfmpegConfig

func CheckFfmpeg() {
	path := config.Vigilis.Recorder.FfmpegPath

	// Check if the path is valid
	fullPath, err := exec.LookPath(path)
	if err != nil {
		logger.Error("ffmpeg not found: %v", err)
		logger.Fatal("Make sure you have ffmpeg installed or provide a valid path in the config")
		return
	}

	Ffmpeg.Path = fullPath
	logger.Trace("ffmpeg found at %v", fullPath)

	// Print the ffmpeg version
	version := FfmpegVersion()
	logger.Info("Working with ffmpeg version %v", version)
}

func FfmpegVersion() string {
	// Run ffmpeg with the version flag
	cmd := exec.Command(Ffmpeg.Path, "-version")
	output, err := cmd.Output()
	if err != nil {
		logger.Fatal("error running ffmpeg: %v", err)
	}

	// Extract the version from the output
	verRegex := regexp.MustCompile(`^ffmpeg version (\d+.\d+)`)
	verResult := verRegex.FindStringSubmatch(string(output))
	if verResult == nil {
		logger.Fatal("error getting ffmpeg version, `ffmpeg -version` didn't return the expected output")
	}

	return verResult[1] // 0 is the match, 1 is the version group
}
