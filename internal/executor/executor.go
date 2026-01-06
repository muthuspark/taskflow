package executor

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/taskflow/taskflow/internal/store"
)

// Executor handles job execution
type Executor struct {
	store *store.Store
}

// New creates a new executor
func New(st *store.Store) *Executor {
	return &Executor{store: st}
}

// Execute runs a job and returns the run result
func (e *Executor) Execute(ctx context.Context, run *store.Run, job *store.Job) error {
	// Validate job script
	if job.Script == "" {
		run.Status = "failure"
		run.ErrorMsg = "Job script is empty"
		e.store.UpdateRun(run)
		return fmt.Errorf("empty script")
	}

	if len(job.Script) > 1000000 { // 1MB limit
		run.Status = "failure"
		run.ErrorMsg = "Job script exceeds maximum size (1MB)"
		e.store.UpdateRun(run)
		return fmt.Errorf("script too large")
	}

	// Update run status to running
	run.Status = "running"
	now := time.Now()
	run.StartedAt = &now
	if err := e.store.UpdateRun(run); err != nil {
		log.Printf("Failed to update run status: %v\n", err)
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
		run.Status = "failure"
		run.ErrorMsg = fmt.Sprintf("Failed to create stdout pipe: %v", err)
		e.store.UpdateRun(run)
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		run.Status = "failure"
		run.ErrorMsg = fmt.Sprintf("Failed to create stderr pipe: %v", err)
		e.store.UpdateRun(run)
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		run.Status = "failure"
		run.ErrorMsg = fmt.Sprintf("Failed to start command: %v", err)
		e.store.UpdateRun(run)
		return err
	}

	// Stream logs concurrently
	go e.streamLogs(run.ID, stdout, "stdout")
	go e.streamLogs(run.ID, stderr, "stderr")

	// Wait for command to complete or timeout
	err = cmd.Wait()

	// Determine final status
	finished := time.Now()
	run.FinishedAt = &finished
	duration := int64(finished.Sub(*run.StartedAt).Milliseconds())
	run.DurationMs = &duration

	if err != nil {
		if execCtx.Err() == context.DeadlineExceeded {
			run.Status = "timeout"
			run.ErrorMsg = fmt.Sprintf("Job exceeded timeout of %d seconds", job.TimeoutSeconds)
			code := 124
			run.ExitCode = &code
		} else {
			run.Status = "failure"
			run.ErrorMsg = err.Error()
			// Try to get exit code
			if exitErr, ok := err.(*exec.ExitError); ok {
				code := exitErr.ExitCode()
				run.ExitCode = &code
			}
		}
	} else {
		run.Status = "success"
		code := 0
		run.ExitCode = &code
	}

	// Log final status
	e.store.AddLog(run.ID, "system", fmt.Sprintf("Job %s with status: %s", run.ID, run.Status))

	// Update run in database
	if err := e.store.UpdateRun(run); err != nil {
		log.Printf("Failed to update run: %v\n", err)
	}

	return nil
}

// streamLogs reads from a pipe and stores logs
func (e *Executor) streamLogs(runID string, pipe interface{}, stream string) {
	// Simple implementation - in production, would use bufio.Scanner
	// For now, just ensure pipe is read
	if r, ok := pipe.(interface{ Read(p []byte) (n int, err error) }); ok {
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if n > 0 {
				lines := strings.Split(string(buf[:n]), "\n")
				for _, line := range lines {
					if line != "" {
						if _, err := e.store.AddLog(runID, stream, line); err != nil {
							log.Printf("Failed to add log: %v\n", err)
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
	// Check if any runs are currently executing
	runs, err := e.store.ListRuns(nil, 1, 0)
	if err != nil || len(runs) == 0 {
		return true
	}

	run := runs[0]
	return run.Status != "running" && run.Status != "pending"
}

// GetRunningJob returns the currently running job, if any
func (e *Executor) GetRunningJob() *store.Run {
	runs, err := e.store.ListRuns(nil, 1, 0)
	if err != nil || len(runs) == 0 {
		return nil
	}

	run := runs[0]
	if run.Status == "running" {
		return run
	}
	return nil
}
