package recorders

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"time"
	"vigilis/internal/config"
	"vigilis/internal/logger"
	"vigilis/internal/util"
)

const ExitTimeout = 5 * time.Second

const (
	ExitReasonStop = "stop requested"
)

type Recorder struct {
	Camera    *config.Camera
	OutputDir string
	index     int

	// Process related data
	process *os.Process
	stdout  bytes.Buffer
	stderr  bytes.Buffer
}

// StartRecording starts a new recording
func (r Recorder) StartRecording() {
	camId := r.Camera.Id

	// Prepare the command
	path, args := BuildCommand(r)

	cmd := exec.Command(path, args...)
	cmd.Stdout = &r.stdout
	cmd.Stderr = &r.stderr

	// Run the command
	err := cmd.Start()
	if err != nil {
		logger.Error("%v recorder > Error spawning %v process: %v", camId, cmd.Args[0], err)
		// TODO Try again but not forever
		return
	}

	r.process = cmd.Process

	pid := cmd.Process.Pid
	logger.Info("%v recorder > Process spawned with PID %d", camId, pid)

	// Wait for the command to exit
	cmdErr := cmd.Wait()

	// Start the new process as soon as this one exits to avoid loosing footage
	r.restart()

	// Log errors, exclude interruptions
	if cmdErr != nil && cmdErr.Error() != "signal: interrupt" {
		logger.Error("%v recorder > Process %d exited with error: %v", camId, pid, cmdErr)

		util.LogBuffer(r.stderr, "stderr", logger.Info, camId+" recorder")
		return
	}

	logger.Info("%v recorder > Recording stopped", camId)
}

// StopRecording stops the recording by exiting the process
func (r Recorder) StopRecording() {
	r.exit(ExitReasonStop)
}

// exit tries to gracefully exit the process, forcing it after a while if needed
func (r Recorder) exit(reason string) {
	camId := r.Camera.Id
	pid := r.process.Pid

	logger.Trace("%v recorder > Gracefully stopping recorder (PID: %d): %v", camId, pid, reason)

	// Try to gracefully exit the process
	err := r.process.Signal(os.Interrupt)
	if err != nil {
		logger.Warn("%v recorder > Error sending interrupt to process with PID %d: %v", camId, pid, err)
	}

	// Check the process after a while
	time.AfterFunc(ExitTimeout, func() {
		// Try to kill the process
		err = r.process.Kill()
		if err != nil {
			if errors.Is(err, os.ErrProcessDone) { // Process is already finished
				logger.Trace("%v recorder > Recording stopped gracefully before timeout (PID %d)", camId, pid)
			} else {
				logger.Error("%v recorder > Error killing process with PID %d: %v", camId, pid, err)
			}

			return
		}

		logger.Warn("%v recorder > Recording forcefully stopped after timeout (PID %d)", camId, pid)
	})
}

// restart signals the orchestrator to (re)start the process
func (r Recorder) restart() {
	// TODO Increase channel count?
	orchestrator.startRecorder <- r.index
}
