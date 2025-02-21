package recorders

import (
	"time"
	"vigilis/internal/config"
	"vigilis/internal/logger"
)

var orchestrator = Orchestrator{
	recorders:     make([]Recorder, 0),
	startRecorder: make(chan int),
}

type Orchestrator struct {
	recorders []Recorder

	startRecorder chan int // Index of the recorder to be (re)started
}

// Init starts all recorders
func Init(cameras []config.Camera) {
	recorders := &orchestrator.recorders

	// Initialize the recorders
	for i, camera := range cameras {
		orchestrator.recorders = append(*recorders, Recorder{
			Camera: camera,
			index:  i,
		})

		logger.Trace("Recorder for camera %v initialized", camera.Id)
	}

	// Start the recorders
	for _, recorder := range *recorders {
		go recorder.StartRecording()
	}
}

// Run is the main loop that takes care of re-starting recorders
func Run() {
	recorders := orchestrator.recorders

	for {
		select {
		case i := <-orchestrator.startRecorder:
			recorder := recorders[i]
			go recorder.StartRecording()
		default:
			// TODO Channel to capture SIGINT/SIGTERM on vigilis
			time.Sleep(time.Second)
		}
	}
}
