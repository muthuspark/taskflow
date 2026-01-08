package store

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateRun creates a new job run
func (s *Store) CreateRun(jobID, triggerType string) (*Run, error) {
	run := &Run{
		ID:          uuid.New().String(),
		JobID:       jobID,
		Status:      "pending",
		TriggerType: triggerType,
	}

	_, err := s.db.Exec(
		`INSERT INTO runs (id, job_id, status, trigger_type) VALUES (?, ?, ?, ?)`,
		run.ID, run.JobID, run.Status, run.TriggerType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create run: %w", err)
	}

	return run, nil
}

// GetRun retrieves a run by ID
func (s *Store) GetRun(id string) (*Run, error) {
	run := &Run{}
	var exitCode sql.NullInt64
	var startedAt, finishedAt sql.NullTime
	var durationMs sql.NullInt64
	var errorMsg sql.NullString

	err := s.db.QueryRow(
		`SELECT id, job_id, status, exit_code, trigger_type, started_at, finished_at, duration_ms, error_message
		 FROM runs WHERE id = ?`,
		id,
	).Scan(
		&run.ID, &run.JobID, &run.Status, &exitCode, &run.TriggerType,
		&startedAt, &finishedAt, &durationMs, &errorMsg,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("run not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get run: %w", err)
	}

	populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)
	return run, nil
}

// ListRuns retrieves runs with optional filtering
func (s *Store) ListRuns(jobID *string, limit int, offset int) ([]*Run, error) {
	// Validate and normalize pagination parameters
	const maxLimit = 1000
	if limit <= 0 || limit > maxLimit {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	baseQuery := `SELECT id, job_id, status, exit_code, trigger_type, started_at, finished_at, duration_ms, error_message
	 FROM runs`
	orderAndPagination := ` ORDER BY started_at DESC LIMIT ? OFFSET ?`

	var rows *sql.Rows
	var err error

	if jobID != nil {
		query := baseQuery + ` WHERE job_id = ?` + orderAndPagination
		rows, err = s.db.Query(query, *jobID, limit, offset)
	} else {
		query := baseQuery + orderAndPagination
		rows, err = s.db.Query(query, limit, offset)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list runs: %w", err)
	}
	defer rows.Close()

	runs := make([]*Run, 0)
	for rows.Next() {
		run := &Run{}
		var exitCode sql.NullInt64
		var startedAt, finishedAt sql.NullTime
		var durationMs sql.NullInt64
		var errorMsg sql.NullString

		if err := rows.Scan(
			&run.ID, &run.JobID, &run.Status, &exitCode, &run.TriggerType,
			&startedAt, &finishedAt, &durationMs, &errorMsg,
		); err != nil {
			return nil, fmt.Errorf("failed to scan run: %w", err)
		}

		populateRunPointers(run, exitCode, startedAt, finishedAt, durationMs, errorMsg)
		runs = append(runs, run)
	}

	return runs, rows.Err()
}

// UpdateRun updates a run's status and metadata
func (s *Store) UpdateRun(run *Run) error {
	_, err := s.db.Exec(
		`UPDATE runs SET status = ?, exit_code = ?, started_at = ?, finished_at = ?, duration_ms = ?, error_message = ?
		 WHERE id = ?`,
		run.Status,
		PointerToNullInt64(run.ExitCode),
		PointerToNullTime(run.StartedAt),
		PointerToNullTime(run.FinishedAt),
		PointerToNullInt64Ptr(run.DurationMs),
		run.ErrorMsg,
		run.ID,
	)
	return err
}

// DeleteRun deletes a run and associated logs/metrics
func (s *Store) DeleteRun(id string) error {
	// Delete associated logs
	if _, err := s.db.Exec(`DELETE FROM logs WHERE run_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete logs: %w", err)
	}

	// Delete associated metrics
	if _, err := s.db.Exec(`DELETE FROM metrics WHERE run_id = ?`, id); err != nil {
		return fmt.Errorf("failed to delete metrics: %w", err)
	}

	// Delete the run
	result, err := s.db.Exec(`DELETE FROM runs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete run: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("run not found")
	}

	return nil
}

// DeleteOldRuns deletes runs older than the specified number of days
func (s *Store) DeleteOldRuns(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)
	result, err := s.db.Exec(
		`DELETE FROM runs WHERE started_at < ?`,
		cutoff,
	)
	if err != nil {
		return fmt.Errorf("failed to delete old runs: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}
	fmt.Printf("Deleted %d old runs\n", rows)
	return nil
}

// populateRunPointers converts nullable database types to Run struct pointers.
// This eliminates duplicate null-checking code in GetRun and ListRuns.
// Follows DRY principle: null conversion logic in one place.
func populateRunPointers(run *Run, exitCode sql.NullInt64, startedAt, finishedAt sql.NullTime, durationMs sql.NullInt64, errorMsg sql.NullString) {
	if exitCode.Valid {
		code := int(exitCode.Int64)
		run.ExitCode = &code
	}
	if startedAt.Valid {
		run.StartedAt = &startedAt.Time
	}
	if finishedAt.Valid {
		run.FinishedAt = &finishedAt.Time
	}
	if durationMs.Valid {
		run.DurationMs = &durationMs.Int64
	}
	if errorMsg.Valid {
		run.ErrorMsg = &errorMsg.String
	}
}
