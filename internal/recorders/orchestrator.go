package recorders

import (
	"errors"
	"fmt"
	"os"
	"path"
	"slices"
	"strings"
	"time"
	"vigilis/internal/config"
	"vigilis/internal/logger"
)

const OutputDirPerms = 0700 // only owner has permission

var orchestrator = Orchestrator{
	recorders:      make([]*Recorder, 0),
	lastRecorderId: 0,
}

type Orchestrator struct {
	recorders      []*Recorder
	lastRecorderId int // TODO Theoretically this can overflow and break binary search

	startRecorder chan int // Index of the recorder to be (re)started
}

func (o *Orchestrator) initializeRecorders(cameras []*config.Camera) {
	basePath := config.Vigilis.Storage.Path
	o.startRecorder = make(chan int, len(cameras))

	for _, camera := range cameras {
		o.lastRecorderId += 1
		o.recorders = append(o.recorders, &Recorder{
			Camera:    camera,
			OutputDir: path.Join(basePath, camera.Id),
			index:     o.lastRecorderId,
		})

		logger.Trace("Recorder for camera %v initialized", camera.Id)
	}
}

func (o *Orchestrator) startRecorders() {
	logger.Info("Starting recorders")
	for _, recorder := range o.recorders {
		go recorder.Record()
	}
}

func (o *Orchestrator) stopRecorders() {
	logger.Info("Stopping recorders")
	for _, recorder := range o.recorders {
		recorder.StopRecording()
	}
}

// waitForRecorders waits for every recorder to exit before returning
func (o *Orchestrator) waitForRecorders() {
	for {
		// Check if any recorder is not stopped; if so, we need to wait
		shouldWait := slices.ContainsFunc(o.recorders, func(recorder *Recorder) bool {
			return recorder.State != StateStopped
		})
		if !shouldWait {
			logger.Info("All recorders are stopped")
			break
		}

		time.Sleep(100 * time.Millisecond)
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

// recreateRecorder forks a Recorder and adds it to the list
func (o *Orchestrator) recreateRecorder(index int) *Recorder {
	// Get the Recorder
	r, rSliceIndex := o.recorderByIndex(index)
	if r == nil {
		panic("recorder not found by index")
	}

	// Create the new Recorder
	newRecorder := r.copy()
	o.lastRecorderId += 1
	newRecorder.index = o.lastRecorderId

	// Add it to the list
	recorders := append(o.recorders, newRecorder)

	// Purge the old recorder
	recorders = slices.Delete(recorders, rSliceIndex, rSliceIndex+1)

	o.recorders = recorders

	return newRecorder
}

func (o *Orchestrator) recorderByIndex(recordIndex int) (*Recorder, int) {
	// We can use binary search because we're appending new recorders with incrementing IDs to the end of the slice
	sliceIndex, found := slices.BinarySearchFunc(o.recorders, recordIndex, func(rec *Recorder, id int) int {
		if rec.index == id {
			return 0
		}

		if rec.index > id {
			return 1
		} else {
			return -1
		}
	})

	if found {
		return o.recorders[sliceIndex], sliceIndex
	}

	return nil, -1
}

// TODO Only do this in linux
func (o *Orchestrator) checkRecorderStates() {
	// Check the state of Recorder processes
	for _, recorder := range o.recorders {
		// Skip stopped recorders or recorders that are not running yet
		if recorder.State == StateStopped || recorder.State == StateReady {
			continue
		}

		pid := recorder.process.Pid

		// Get the process stat file
		data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil {
			logger.Error("Error reading state of process with PID %d", pid)
			continue
		}

		// Extract the fields
		fields := strings.Fields(string(data))
		if len(fields) < 3 {
			logger.Error("Unexpected format found when reading state of process with PID %d", pid)
			continue
		}

		linuxState := fields[2]
		recorderState, err := linuxToRecorderState(linuxState)
		if err != nil {
			logger.Error("Invalid process state for PID %d: %s", pid, linuxState)
			continue
		}

		if recorder.State != recorderState {
			logger.Info("%s recorder > Process with PID %d changed from %s to %s", recorder.Camera.Id, pid, recorder.State.String(), recorderState.String())
			recorder.State = recorderState
		}
	}
}

// linuxToRecorderState translates a Linux process state to a RecorderState
func linuxToRecorderState(linuxState string) (RecorderState, error) {
	switch linuxState {
	case "D": // Uninterruptible sleep
		return StatePaused, nil
	case "S": // Interruptible sleep
		fallthrough
	case "R": // Running
		return StateRunning, nil
	case "Z": // Zombie
		fallthrough
	case "T": // Stopped
		return StateStopped, nil
	}

	return StateStopped, errors.New("unknown process state: " + linuxState)
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
	select {
	// Re-start Recorder (by recreation) when one goes down
	case i := <-orchestrator.startRecorder:
		rec := orchestrator.recreateRecorder(i)
		go rec.Record()
	default:
		orchestrator.checkRecorderStates()
	}
}

func Stop() {
	orchestrator.stopRecorders()
	orchestrator.waitForRecorders()
}
