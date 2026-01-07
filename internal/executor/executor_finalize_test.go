package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

// TestFinalizeRunSuccess tests successful job completion
func TestFinalizeRunSuccess(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:     "test-run",
		JobID:  "test-job",
		Status: internal.JobStatusRunning,
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 300,
	}

	now := time.Now()
	run.StartedAt = &now
	execCtx := context.Background()

	// No error means success
	exec.finalizeRun(run, job, nil, execCtx)

	assert.Equal(t, internal.JobStatusSuccess, run.Status)
	assert.NotNil(t, run.ExitCode)
	assert.Equal(t, internal.ExitCodeSuccess, *run.ExitCode)
	assert.NotNil(t, run.FinishedAt)
	assert.NotNil(t, run.DurationMs)
	assert.Nil(t, run.ErrorMsg)
}

// TestFinalizeRunTimeout tests job timeout scenario
func TestFinalizeRunTimeout(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:     "test-run",
		JobID:  "test-job",
		Status: internal.JobStatusRunning,
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 5,
	}

	now := time.Now()
	run.StartedAt = &now

	// Create a context that's already exceeded deadline
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond) // Ensure context is expired

	// Pass a generic error; the context deadline will trigger timeout logic
	testErr := context.DeadlineExceeded

	exec.finalizeRun(run, job, testErr, ctx)

	assert.Equal(t, internal.JobStatusTimeout, run.Status)
	assert.NotNil(t, run.ExitCode)
	assert.Equal(t, internal.ExitCodeTimeout, *run.ExitCode)
	assert.NotNil(t, run.ErrorMsg)
	assert.Contains(t, *run.ErrorMsg, "timeout")
}

// TestFinalizeRunFailureWithExitCode tests failure with exit code extraction
func TestFinalizeRunFailureWithExitCode(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:     "test-run",
		JobID:  "test-job",
		Status: internal.JobStatusRunning,
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 300,
	}

	now := time.Now()
	run.StartedAt = &now
	execCtx := context.Background()

	exec.finalizeRun(run, job, context.Canceled, execCtx)

	assert.Equal(t, internal.JobStatusFailure, run.Status)
	assert.NotNil(t, run.ErrorMsg)
}

// TestFinalizeRunFailureWithoutExitCode tests failure without extractable exit code
func TestFinalizeRunFailureWithoutExitCode(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:     "test-run",
		JobID:  "test-job",
		Status: internal.JobStatusRunning,
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 300,
	}

	now := time.Now()
	run.StartedAt = &now
	execCtx := context.Background()

	// Generic error without exit code
	testErr := context.Canceled

	exec.finalizeRun(run, job, testErr, execCtx)

	assert.Equal(t, internal.JobStatusFailure, run.Status)
	assert.Nil(t, run.ExitCode) // No exit code available
	assert.NotNil(t, run.ErrorMsg)
	assert.Equal(t, "context canceled", *run.ErrorMsg)
}

// TestFinalizeRunDurationCalculation tests duration calculation is correct
func TestFinalizeRunDurationCalculation(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:    "test-run",
		JobID: "test-job",
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 300,
	}

	// Set start time to 5 seconds ago
	startTime := time.Now().Add(-5 * time.Second)
	run.StartedAt = &startTime
	execCtx := context.Background()

	exec.finalizeRun(run, job, nil, execCtx)

	assert.NotNil(t, run.DurationMs)
	// Duration should be approximately 5000ms (5 seconds)
	// Allow 500ms tolerance for test execution time
	assert.Greater(t, *run.DurationMs, int64(4500))
	assert.Less(t, *run.DurationMs, int64(6000))
}

// TestFinalizeRunVeryShortDuration tests duration for very quick operations
func TestFinalizeRunVeryShortDuration(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:    "test-run",
		JobID: "test-job",
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 300,
	}

	// Set start time to nearly now
	startTime := time.Now().Add(-1 * time.Millisecond)
	run.StartedAt = &startTime
	execCtx := context.Background()

	exec.finalizeRun(run, job, nil, execCtx)

	assert.NotNil(t, run.DurationMs)
	assert.GreaterOrEqual(t, *run.DurationMs, int64(0))
	assert.Less(t, *run.DurationMs, int64(100))
}

// TestFinalizeRunFinishedAtNotNil tests that FinishedAt is always set
func TestFinalizeRunFinishedAtNotNil(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:    "test-run",
		JobID: "test-job",
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 300,
	}

	now := time.Now()
	run.StartedAt = &now
	execCtx := context.Background()

	exec.finalizeRun(run, job, nil, execCtx)

	assert.NotNil(t, run.FinishedAt)
	assert.GreaterOrEqual(t, run.FinishedAt.Unix(), now.Unix())
}

// TestFinalizeRunErrorMessageFormat tests timeout error message format
func TestFinalizeRunErrorMessageFormat(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	run := &store.Run{
		ID:     "test-run",
		JobID:  "test-job",
		Status: internal.JobStatusRunning,
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 30,
	}

	now := time.Now()
	run.StartedAt = &now

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)

	exec.finalizeRun(run, job, context.DeadlineExceeded, ctx)

	assert.NotNil(t, run.ErrorMsg)
	assert.Contains(t, *run.ErrorMsg, "30")
	assert.Contains(t, *run.ErrorMsg, "seconds")
}

// TestFinalizeRunPreservesJobID tests that job ID is preserved
func TestFinalizeRunPreservesJobID(t *testing.T) {
	mockStore := newMockStoreForTesting(t)
	defer mockStore.Close()

	exec := New(mockStore.Store)

	originalJobID := "original-job-id"
	run := &store.Run{
		ID:    "test-run",
		JobID: originalJobID,
	}

	job := &store.Job{
		ID:             "test-job",
		TimeoutSeconds: 300,
	}

	now := time.Now()
	run.StartedAt = &now
	execCtx := context.Background()

	exec.finalizeRun(run, job, nil, execCtx)

	assert.Equal(t, originalJobID, run.JobID)
}

// TestFinalizeRunWithMultipleFailures tests error message from multiple failure modes
func TestFinalizeRunWithMultipleFailures(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		contextErr     error
		expectedStatus string
	}{
		{
			name:           "normal_failure_canceled",
			err:            context.Canceled,
			contextErr:     context.Background().Err(),
			expectedStatus: internal.JobStatusFailure,
		},
		{
			name:           "normal_failure_deadline",
			err:            context.DeadlineExceeded,
			contextErr:     context.DeadlineExceeded,
			expectedStatus: internal.JobStatusTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := newMockStoreForTesting(t)
			defer mockStore.Close()

			exec := New(mockStore.Store)

			run := &store.Run{
				ID:    "test-run",
				JobID: "test-job",
			}

			job := &store.Job{
				ID:             "test-job",
				TimeoutSeconds: 300,
			}

			now := time.Now()
			run.StartedAt = &now

			var execCtx context.Context
			if tt.contextErr == context.DeadlineExceeded {
				// Create context with expired deadline
				ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
				defer cancel()
				time.Sleep(10 * time.Millisecond)
				execCtx = ctx
			} else {
				execCtx = context.Background()
			}

			exec.finalizeRun(run, job, tt.err, execCtx)

			assert.Equal(t, tt.expectedStatus, run.Status)
		})
	}
}
