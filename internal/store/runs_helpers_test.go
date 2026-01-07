package store

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestPopulateRunPointersAllValid tests when all nullable fields are valid
func TestPopulateRunPointersAllValid(t *testing.T) {
	run := &Run{ID: "test-run"}

	exitCode := sql.NullInt64{Int64: 0, Valid: true}
	startedAt := sql.NullTime{Time: time.Date(2026, 1, 7, 14, 0, 0, 0, time.UTC), Valid: true}
	finishedAt := sql.NullTime{Time: time.Date(2026, 1, 7, 14, 5, 0, 0, time.UTC), Valid: true}
	durationMs := sql.NullInt64{Int64: 300000, Valid: true}
	errorMsg := sql.NullString{String: "test error", Valid: true}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	assert.NotNil(t, run.ExitCode)
	assert.Equal(t, 0, *run.ExitCode)
	assert.NotNil(t, run.StartedAt)
	assert.Equal(t, time.Date(2026, 1, 7, 14, 0, 0, 0, time.UTC), *run.StartedAt)
	assert.NotNil(t, run.FinishedAt)
	assert.NotNil(t, run.DurationMs)
	assert.Equal(t, int64(300000), *run.DurationMs)
	assert.NotNil(t, run.ErrorMsg)
	assert.Equal(t, "test error", *run.ErrorMsg)
}

// TestPopulateRunPointersAllNull tests when all nullable fields are NULL
func TestPopulateRunPointersAllNull(t *testing.T) {
	run := &Run{ID: "test-run"}

	exitCode := sql.NullInt64{Valid: false}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{Valid: false}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	// All pointers should remain nil
	assert.Nil(t, run.ExitCode)
	assert.Nil(t, run.StartedAt)
	assert.Nil(t, run.FinishedAt)
	assert.Nil(t, run.DurationMs)
	assert.Nil(t, run.ErrorMsg)
}

// TestPopulateRunPointersMixedValid tests when some fields are valid, some are NULL
func TestPopulateRunPointersMixedValid(t *testing.T) {
	run := &Run{ID: "test-run"}

	exitCode := sql.NullInt64{Int64: 1, Valid: true}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Time: time.Now(), Valid: true}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{String: "error", Valid: true}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	// Valid fields should be populated
	assert.NotNil(t, run.ExitCode)
	assert.Equal(t, 1, *run.ExitCode)
	assert.NotNil(t, run.FinishedAt)
	assert.NotNil(t, run.ErrorMsg)

	// NULL fields should remain nil
	assert.Nil(t, run.StartedAt)
	assert.Nil(t, run.DurationMs)
}

// TestPopulateRunPointersExitCodeZero tests that exit code 0 (success) is properly handled
func TestPopulateRunPointersExitCodeZero(t *testing.T) {
	run := &Run{ID: "test-run"}

	// Exit code 0 is valid and should create a pointer
	exitCode := sql.NullInt64{Int64: 0, Valid: true}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{Valid: false}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	assert.NotNil(t, run.ExitCode)
	assert.Equal(t, 0, *run.ExitCode)
}

// TestPopulateRunPointersNegativeExitCode tests handling of negative exit codes
func TestPopulateRunPointersNegativeExitCode(t *testing.T) {
	run := &Run{ID: "test-run"}

	// Negative exit codes can occur with timeout/signal termination
	exitCode := sql.NullInt64{Int64: -1, Valid: true}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{Valid: false}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	assert.NotNil(t, run.ExitCode)
	assert.Equal(t, -1, *run.ExitCode)
}

// TestPopulateRunPointersLargeExitCode tests handling of large exit codes
func TestPopulateRunPointersLargeExitCode(t *testing.T) {
	run := &Run{ID: "test-run"}

	exitCode := sql.NullInt64{Int64: 255, Valid: true}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{Valid: false}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	assert.NotNil(t, run.ExitCode)
	assert.Equal(t, 255, *run.ExitCode)
}

// TestPopulateRunPointersLargeDuration tests handling of large duration values
func TestPopulateRunPointersLargeDuration(t *testing.T) {
	run := &Run{ID: "test-run"}

	// 24 hours in milliseconds
	largeMs := int64(86400000)

	exitCode := sql.NullInt64{Valid: false}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Int64: largeMs, Valid: true}
	errorMsg := sql.NullString{Valid: false}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	assert.NotNil(t, run.DurationMs)
	assert.Equal(t, largeMs, *run.DurationMs)
}

// TestPopulateRunPointersEmptyErrorMessage tests handling of empty error messages
func TestPopulateRunPointersEmptyErrorMessage(t *testing.T) {
	run := &Run{ID: "test-run"}

	exitCode := sql.NullInt64{Valid: false}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{String: "", Valid: true}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	// Empty string is still a valid value
	assert.NotNil(t, run.ErrorMsg)
	assert.Equal(t, "", *run.ErrorMsg)
}

// TestPopulateRunPointersLongErrorMessage tests handling of very long error messages
func TestPopulateRunPointersLongErrorMessage(t *testing.T) {
	run := &Run{ID: "test-run"}

	longMsg := "This is a very long error message " +
		"that contains detailed information about what went wrong " +
		"and includes stack traces and system diagnostics " +
		"that can be quite lengthy"

	exitCode := sql.NullInt64{Valid: false}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{String: longMsg, Valid: true}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	assert.NotNil(t, run.ErrorMsg)
	assert.Equal(t, longMsg, *run.ErrorMsg)
}

// TestPopulateRunPointersDoesNotModifyExistingFields tests that function doesn't overwrite fields unexpectedly
func TestPopulateRunPointersDoesNotModifyExistingFields(t *testing.T) {
	originalID := "test-run-id"
	originalStatus := "success"
	originalJobID := "job-123"
	originalTriggerType := "manual"

	run := &Run{
		ID:          originalID,
		Status:      originalStatus,
		JobID:       originalJobID,
		TriggerType: originalTriggerType,
	}

	exitCode := sql.NullInt64{Valid: false}
	startedAt := sql.NullTime{Valid: false}
	finishedAt := sql.NullTime{Valid: false}
	durationMs := sql.NullInt64{Valid: false}
	errorMsg := sql.NullString{Valid: false}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

	// Original fields should be unchanged
	assert.Equal(t, originalID, run.ID)
	assert.Equal(t, originalStatus, run.Status)
	assert.Equal(t, originalJobID, run.JobID)
	assert.Equal(t, originalTriggerType, run.TriggerType)
}

// TestPopulateRunPointersTimeHandling tests proper handling of various time values
func TestPopulateRunPointersTimeHandling(t *testing.T) {
	tests := []struct {
		name string
		time time.Time
	}{
		{"epoch", time.Unix(0, 0).UTC()},
		{"far_past", time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)},
		{"current", time.Now()},
		{"far_future", time.Date(2099, 12, 31, 23, 59, 59, 0, time.UTC)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			run := &Run{ID: "test-run"}

			exitCode := sql.NullInt64{Valid: false}
			startedAt := sql.NullTime{Time: tt.time, Valid: true}
			finishedAt := sql.NullTime{Valid: false}
			durationMs := sql.NullInt64{Valid: false}
			errorMsg := sql.NullString{Valid: false}

			populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)

			assert.NotNil(t, run.StartedAt)
			assert.Equal(t, tt.time, *run.StartedAt)
		})
	}
}
