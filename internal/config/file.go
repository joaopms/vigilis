package config

import (
	"os"
	"path/filepath"
	"vigilis/internal/logger"
)

func ReadFromFile(path string) {
	logger.Info("Provided config file path: %v", path)

	// Get the current path
	pwd, err := os.Getwd()
	if err != nil {
		logger.Fatal("Unable to get current path: %v", err)
	}

	// Compute the full path
	fullPath := filepath.Join(pwd, path)
	logger.Info("Full path to config: %v", fullPath)

	// Read the file
	data, err := os.ReadFile(fullPath)
	if err != nil {
		logger.Fatal("Unable to read config file: %v", err)
	}

	err = Parse(data)
	if err != nil {
		logger.Fatal("Error parsing the config.\n%v", err)
	}
}
