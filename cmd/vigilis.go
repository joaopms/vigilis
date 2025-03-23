package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"
	"vigilis/internal/config"
	"vigilis/internal/files"
	"vigilis/internal/logger"
	"vigilis/internal/recorders"
)

var (
	configFile string = "config.yaml"

	debug      bool
	dumpConfig bool
)

var version = "0.0.0-development" // Version is automatically set when building

func init() {
	const configUsage = "path to the config file"
	flag.StringVar(&configFile, "c", configFile, configUsage)
	flag.StringVar(&configFile, "config", configFile, configUsage)

	flag.BoolVar(&debug, "debug", false, "print debug messages to stdout")
	flag.BoolVar(&dumpConfig, "dump-config", false, "dump the parsed config to stdout")

	flag.BoolFunc("version", "prints the version and exits", printVersion)
}

func main() {
	flag.Parse()

	// Setup the logger
	logger.Setup(debug)
	defer logger.Stop()

	logger.Info("Starting Vigilis v%s", version)

	if dumpConfig && !debug {
		logger.Warn("dump-config is enable but debug is not enabled, skipping config dump")
	}

	// Load the config
	config.ReadFromFile(configFile)
	if dumpConfig && debug {
		prettyConfig, err := json.MarshalIndent(config.Vigilis, "", "  ")
		if err != nil {
			logger.Error("Unable to dump config: %v", err)
		} else {
			logger.Trace("Config dump in JSON format:\n%v", string(prettyConfig))
		}
	}

	// Check for dependencies
	recorders.CheckFfmpeg()

	// Initialize the camera recorders
	recorders.Init(config.Vigilis.Cameras)

	// Delete old recordings
	go files.DeleteOldRecordings()

	run()
}

// Main application loop
func run() {
	tick := time.Tick(time.Second * 1)
	recordingTick := time.Tick(recorders.RecordingLengthMinutes * time.Minute)

	for {
		select {
		// TODO Channel to capture SIGINT/SIGTERM on vigilis
		case <-recordingTick:
			// Periodically delete old recordings
			go files.DeleteOldRecordings()
		case <-tick:
			recorders.Loop()
		}
	}
}

func printVersion(_ string) error {
	fmt.Printf("v%s\n", version)
	os.Exit(0)
	return nil
}
