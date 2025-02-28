package files

import (
	"io/fs"
	"os"
	"path/filepath"
	"time"
	"vigilis/internal/config"
	"vigilis/internal/logger"
)

type purger struct {
	limit time.Duration
	count int
}

func DeleteOldRecordings() {
	logger.Info("Deleting old recordings...")

	path := config.Vigilis.Storage.Path
	p := &purger{
		limit: config.Vigilis.Storage.RetentionDaysDuration(),
	}

	// Walk the recordings directory and try to purge files
	err := filepath.WalkDir(path, p.purge())
	if err != nil {
		logger.Error("Error deleting old recordings: %v", err)
		return
	}

	if p.count > 0 {
		logger.Info("Deleted %d recording(s)", p.count)
	} else {
		logger.Info("No recordings were deleted")
	}
}

func (p *purger) purge() func(path string, entry fs.DirEntry, _ error) error {
	return func(path string, entry fs.DirEntry, _ error) error {
		// Don't try to purge directories
		if entry.IsDir() {
			return nil
		}

		// Get the file info
		info, err := entry.Info()
		if err != nil {
			logger.Error("Error analysing recording for possible deletion: %v", err)
			return nil
		}

		// Check if the age of the file is past the retention limit
		age := time.Since(info.ModTime())
		if age > p.limit {
			err := os.Remove(path)
			if err != nil {
				logger.Warn("Error deleting recording %v: %v", path, err)
				return nil
			}

			p.count++
			logger.Trace("Recording %v deleted", path)
		}

		return nil
	}
}
