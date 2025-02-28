package recorders

import (
	"os"
	"path"
	"vigilis/internal/config"
	"vigilis/internal/logger"
)

const OutputDirPerms = 0700 // only owner has permission

var orchestrator = Orchestrator{
	recorders:     make([]Recorder, 0),
	startRecorder: make(chan int),
}

type Orchestrator struct {
	recorders []Recorder

	startRecorder chan int // Index of the recorder to be (re)started
}

func (o *Orchestrator) initializeRecorders(cameras []*config.Camera) {
	basePath := config.Vigilis.Storage.Path

	for i, camera := range cameras {
		o.recorders = append(o.recorders, Recorder{
			Camera:    camera,
			OutputDir: path.Join(basePath, camera.Id),
			index:     i,
		})

		logger.Trace("Recorder for camera %v initialized", camera.Id)
	}
}

func (o *Orchestrator) startRecorders() {
	for _, recorder := range o.recorders {
		go recorder.StartRecording()
	}
}

func (o *Orchestrator) ensureRecordingDirectories() {
	for _, recorder := range o.recorders {
		cam := recorder.Camera

		err := os.MkdirAll(recorder.OutputDir, OutputDirPerms)
		if err != nil {
			logger.Fatal("Error creating directory for camera %v: %v", cam.Id, err)
		}
	}
}

// Init starts all recorders
func Init(cameras []*config.Camera) {
	// Initialize the recorders
	orchestrator.initializeRecorders(cameras)

	// Create directories if needed
	orchestrator.ensureRecordingDirectories()

	// Start the recorders
	orchestrator.startRecorders()
}

// Loop takes care of re-starting recorders
func Loop() {
	recorders := orchestrator.recorders

	select {
	// Re-start recorder when one goes down
	case i := <-orchestrator.startRecorder:
		recorder := recorders[i]
		go recorder.StartRecording()
	default:
	}
}
