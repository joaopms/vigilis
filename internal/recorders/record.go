package recorders

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"time"
	"vigilis/internal/config"
	"vigilis/internal/logger"
)

// TODO Set this on the config
const RecordingDuration = 10 * time.Second

const ExitTimeout = 5 * time.Second

const (
	ExitReasonDuration = "reached desired duration"
	ExitReasonStop     = "stop requested"
)

type Recorder struct {
	Camera config.Camera
	index  int

	// Process related data
	process *os.Process
	stdout  bytes.Buffer
	stderr  bytes.Buffer
}

// StartRecording starts a new recording
func (r *Recorder) StartRecording() {
	camId := r.Camera.Id

	// Build the command
	cmd := exec.Command("sleep", "10")
	cmd.Stdout = &r.stdout
	cmd.Stderr = &r.stderr

	// Run the command
	err := cmd.Start()
	if err != nil {
		logger.Error("%v recorder > Error spawning %v process: %v", camId, cmd.Args[0], err)
		// TODO Try again but not forever
		return
	}

	// Automatically exit the process after a while
	exitTimer := time.AfterFunc(RecordingDuration, func() {
		r.exit(ExitReasonDuration)
	})

	r.process = cmd.Process

	pid := cmd.Process.Pid
	logger.Info("%v recorder > Process spawned with PID %d", camId, pid)

	// --------------------------------

	// TODO Properly dump logs

	for r.stdout.Len() > 0 {
		output, err := r.stdout.ReadBytes('\n')
		if err != nil {
			logger.Warn("Error reading stdout: %v", err)
			break
		}

		// Remove the last character, \n
		output = output[:len(output)-1]

		logger.Trace("stdout output: %v", string(output))
	}

	for r.stderr.Len() > 0 {
		output, err := r.stderr.ReadBytes('\n')
		if err != nil {
			logger.Warn("Error reading stderr: %v", err)
			break
		}

		// Remove the last character, \n
		output = output[:len(output)-1]

		logger.Warn("stderr output: %v", string(output))
	}

	// --------------------------------

	// Wait for the command to exit
	cmdErr := cmd.Wait()

	// Start the new process as soon as this one exits to avoid loosing footage
	r.restart()

	// Stop the restarting of the current process
	exitTimer.Stop()

	// Log errors, exclude interruptions
	if cmdErr != nil && cmdErr.Error() != "signal: interrupt" {
		logger.Error("%v recorder > Process %d exited with error: %v", camId, pid, cmdErr)
		return
	}

	logger.Info("%v recorder > Recording stopped", camId)
}

// StopRecording stops the recording by exiting the process
func (r *Recorder) StopRecording() {
	r.exit(ExitReasonStop)
}

// exit tries to gracefully exit the process, forcing it after a while if needed
func (r *Recorder) exit(reason string) {
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
func (r *Recorder) restart() {
	// TODO Increase channel count?
	orchestrator.startRecorder <- r.index
}
