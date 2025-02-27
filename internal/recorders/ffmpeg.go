package recorders

import (
	"os/exec"
	"path"
	"regexp"
	"slices"
	"strconv"
	"vigilis/internal/config"
	"vigilis/internal/logger"
)

type RecordMode int

const (
	RecordModeDirect RecordMode = iota
	//RecordMode2
)

type (
	FfmpegConfig struct {
		Path string
	}
)

type cmdArgs = []string

const Filename = "%Y%m%d-%H%M%S.mkv"

var recordArgs = map[RecordMode]cmdArgs{
	// See https://medium.com/@tom.humph/saving-rtsp-camera-streams-with-ffmpeg-baab7e80d767
	RecordModeDirect: cmdArgs{
		"-hide_banner", "-y",
		"-loglevel", "error",
		"-rtsp_transport", "tcp",
		"-use_wallclock_as_timestamps", "1",
		"-vcodec", "copy",
		"-acodec", "copy",
		"-f", "segment",
		"-reset_timestamps", "1",
		"-segment_time", "" + strconv.Itoa(10*60), // in minutes
		"-segment_atclocktime", "1", // minute to start a new segment
		"-segment_format", "mkv",
		"-strftime", "1",
	},
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

func BuildCommand(r Recorder) (string, []string) {
	// TODO Add the record mode to the camera config
	// TODO Add custom args to the camera config
	args := recordArgs[RecordModeDirect]

	outputPath := path.Join(r.OutputDir, Filename)

	return Ffmpeg.Path,
		slices.Concat(
			[]string{"-i", r.Camera.StreamUrl},
			args,
			[]string{outputPath},
		)
}
