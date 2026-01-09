package executor

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"sync"
	"time"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

// LogBroadcaster is a callback function for broadcasting logs via WebSocket
type LogBroadcaster func(runID string, stream string, content string, timestamp time.Time)

// StatusBroadcaster is a callback function for broadcasting status changes via WebSocket
type StatusBroadcaster func(runID string, status string)

// Executor handles job execution
type Executor struct {
	store             *store.Store
	logBroadcaster    LogBroadcaster
	statusBroadcaster StatusBroadcaster
}

// New creates a new executor
func New(st *store.Store) *Executor {
	return &Executor{store: st}
}

// SetLogBroadcaster sets the callback for broadcasting logs
func (e *Executor) SetLogBroadcaster(broadcaster LogBroadcaster) {
	e.logBroadcaster = broadcaster
}

// SetStatusBroadcaster sets the callback for broadcasting status changes
func (e *Executor) SetStatusBroadcaster(broadcaster StatusBroadcaster) {
	e.statusBroadcaster = broadcaster
}

// Execute runs a job and returns the run result
func (e *Executor) Execute(ctx context.Context, run *store.Run, job *store.Job) error {
	// Validate job script
	if job.Script == "" {
		run.Status = internal.JobStatusFailure
		msg := "Job script is empty"
		run.ErrorMsg = &msg
		e.store.UpdateRun(run)
		return fmt.Errorf("empty script")
	}

	if len(job.Script) > internal.MaxScriptSize {
		run.Status = internal.JobStatusFailure
		msg := fmt.Sprintf("Job script exceeds maximum size (%s)", internal.MaxScriptSizeReadable)
		run.ErrorMsg = &msg
		e.store.UpdateRun(run)
		return fmt.Errorf("script too large")
	}

	// Update run status to running
	run.Status = internal.JobStatusRunning
	now := time.Now()
	run.StartedAt = &now
	if err := e.store.UpdateRun(run); err != nil {
		log.Printf("Failed to update run status: %v\n", err)
	}
	// Broadcast status change via WebSocket
	if e.statusBroadcaster != nil {
		e.statusBroadcaster(run.ID, run.Status)
	}

	// Create timeout context
	timeoutDuration := time.Duration(job.TimeoutSeconds) * time.Second
	execCtx, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	// Create command - scripts executed as-is (admin only, by design)
	cmd := exec.CommandContext(execCtx, "bash", "-c", job.Script)
	cmd.Dir = job.WorkingDir

	// Set up pipes for stdout/stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		run.Status = internal.JobStatusFailure
		msg := fmt.Sprintf("Failed to create stdout pipe: %v", err)
		run.ErrorMsg = &msg
		e.store.UpdateRun(run)
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		run.Status = internal.JobStatusFailure
		msg := fmt.Sprintf("Failed to create stderr pipe: %v", err)
		run.ErrorMsg = &msg
		e.store.UpdateRun(run)
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		run.Status = internal.JobStatusFailure
		msg := fmt.Sprintf("Failed to start command: %v", err)
		run.ErrorMsg = &msg
		e.store.UpdateRun(run)
		return err
	}

	// Stream logs concurrently with synchronization
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		e.streamLogs(run.ID, stdout, "stdout")
	}()
	go func() {
		defer wg.Done()
		e.streamLogs(run.ID, stderr, "stderr")
	}()

	// Wait for command to complete or timeout
	err = cmd.Wait()
	// Ensure all logs are fully written before proceeding
	wg.Wait()

	// Determine final status and update run
	e.finalizeRun(run, job, err, execCtx)

	// Log final status
	finalMsg := fmt.Sprintf("Job %s with status: %s", run.ID, run.Status)
	e.store.AddLog(run.ID, internal.StreamSystem, finalMsg)
	// Broadcast final log
	if e.logBroadcaster != nil {
		e.logBroadcaster(run.ID, internal.StreamSystem, finalMsg, time.Now())
	}

	// Update run in database
	if err := e.store.UpdateRun(run); err != nil {
		log.Printf("Failed to update run: %v\n", err)
	}

	// Broadcast final status change via WebSocket
	if e.statusBroadcaster != nil {
		e.statusBroadcaster(run.ID, run.Status)
	}

	return nil
}

// streamLogs reads from a pipe and stores logs
func (e *Executor) streamLogs(runID string, pipe interface{}, stream string) {
	// Simple implementation - in production, would use bufio.Scanner
	// For now, just ensure pipe is read
	if r, ok := pipe.(interface{ Read(p []byte) (n int, err error) }); ok {
		buf := make([]byte, internal.LogStreamBufferSize)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				lines := strings.Split(string(buf[:n]), "\n")
				for _, line := range lines {
					if line != "" {
						timestamp := time.Now()
						if _, err := e.store.AddLog(runID, stream, line); err != nil {
							log.Printf("Failed to add log: %v\n", err)
						}
						// Broadcast log via WebSocket
						if e.logBroadcaster != nil {
							e.logBroadcaster(runID, stream, line, timestamp)
						}
					}
				}
			}
			if err != nil {
				break
			}
		}
	}
}

// CanExecute checks if a job can be executed (respecting concurrency limits)
func (e *Executor) CanExecute() bool {
	// In Phase 1, we only allow one concurrent job
	runs, err := e.store.ListRuns(nil, 1, 0)
	if err != nil || len(runs) == 0 {
		return true
	}

	status := runs[0].Status
	return status != internal.JobStatusRunning && status != internal.JobStatusPending
}

// GetRunningJob returns the currently running job, if any
func (e *Executor) GetRunningJob() *store.Run {
	runs, err := e.store.ListRuns(nil, 1, 0)
	if err != nil || len(runs) == 0 {
		return nil
	}

	if runs[0].Status == internal.JobStatusRunning {
		return runs[0]
	}
	return nil
}

// finalizeRun sets the final status, exit code, and error message for a completed run.
// Extracted from Execute() to reduce its complexity and improve maintainability.
func (e *Executor) finalizeRun(run *store.Run, job *store.Job, cmdErr error, execCtx context.Context) {
	finished := time.Now()
	run.FinishedAt = &finished
	duration := int64(finished.Sub(*run.StartedAt).Milliseconds())
	run.DurationMs = &duration

	if cmdErr != nil {
		if errors.Is(execCtx.Err(), context.DeadlineExceeded) {
			run.Status = internal.JobStatusTimeout
			msg := fmt.Sprintf("Job exceeded timeout of %d seconds", job.TimeoutSeconds)
			run.ErrorMsg = &msg
			code := internal.ExitCodeTimeout
			run.ExitCode = &code
		} else {
			run.Status = internal.JobStatusFailure
			msg := cmdErr.Error()
			run.ErrorMsg = &msg
			// Try to extract exit code from command execution error
			var exitErr *exec.ExitError
			if errors.As(cmdErr, &exitErr) {
				code := exitErr.ExitCode()
				run.ExitCode = &code
			}
		}
	} else {
		run.Status = internal.JobStatusSuccess
		code := internal.ExitCodeSuccess
		run.ExitCode = &code
	}
}
