package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// CreateJob creates a new job
func (s *Store) CreateJob(job *Job) (*Job, error) {
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.UpdatedAt.IsZero() {
		job.UpdatedAt = time.Now()
	}

	_, err := s.db.Exec(
		`INSERT INTO jobs (id, name, description, script, working_dir, timeout_seconds,
		 retry_count, retry_delay_seconds, enabled, notify_emails, notify_on, timezone,
		 created_by, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		job.ID, job.Name, job.Description, job.Script, job.WorkingDir, job.TimeoutSeconds,
		job.RetryCount, job.RetryDelaySeconds, job.Enabled, job.NotifyEmails, job.NotifyOn,
		job.Timezone, job.CreatedBy, job.CreatedAt, job.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	return job, nil
}

// GetJob retrieves a job by ID
func (s *Store) GetJob(id string) (*Job, error) {
	job := &Job{}

	err := s.db.QueryRow(
		`SELECT id, name, description, script, working_dir, timeout_seconds,
		 retry_count, retry_delay_seconds, enabled, notify_emails, notify_on, timezone,
		 created_by, created_at, updated_at FROM jobs WHERE id = ?`,
		id,
	).Scan(
		&job.ID, &job.Name, &job.Description, &job.Script, &job.WorkingDir,
		&job.TimeoutSeconds, &job.RetryCount, &job.RetryDelaySeconds, &job.Enabled,
		&job.NotifyEmails, &job.NotifyOn, &job.Timezone, &job.CreatedBy,
		&job.CreatedAt, &job.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("job not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	return job, nil
}

// ListJobs retrieves all jobs, optionally filtered by creator
func (s *Store) ListJobs(createdBy *int) ([]*Job, error) {
	query := `SELECT id, name, description, script, working_dir, timeout_seconds,
	 retry_count, retry_delay_seconds, enabled, notify_emails, notify_on, timezone,
	 created_by, created_at, updated_at FROM jobs`

	var rows *sql.Rows
	var err error

	if createdBy != nil {
		query += ` WHERE created_by = ?`
		rows, err = s.db.Query(query + ` ORDER BY created_at DESC`, *createdBy)
	} else {
		rows, err = s.db.Query(query + ` ORDER BY created_at DESC`)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	var jobs []*Job
	for rows.Next() {
		job := &Job{}
		if err := rows.Scan(
			&job.ID, &job.Name, &job.Description, &job.Script, &job.WorkingDir,
			&job.TimeoutSeconds, &job.RetryCount, &job.RetryDelaySeconds, &job.Enabled,
			&job.NotifyEmails, &job.NotifyOn, &job.Timezone, &job.CreatedBy,
			&job.CreatedAt, &job.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}
		jobs = append(jobs, job)
	}

	return jobs, rows.Err()
}

// UpdateJob updates a job
func (s *Store) UpdateJob(job *Job) error {
	job.UpdatedAt = time.Now()

	result, err := s.db.Exec(
		`UPDATE jobs SET name = ?, description = ?, script = ?, working_dir = ?,
		 timeout_seconds = ?, retry_count = ?, retry_delay_seconds = ?, enabled = ?,
		 notify_emails = ?, notify_on = ?, timezone = ?, updated_at = ?
		 WHERE id = ?`,
		job.Name, job.Description, job.Script, job.WorkingDir, job.TimeoutSeconds,
		job.RetryCount, job.RetryDelaySeconds, job.Enabled, job.NotifyEmails,
		job.NotifyOn, job.Timezone, job.UpdatedAt, job.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("job not found")
	}

	return nil
}

// DeleteJob deletes a job
func (s *Store) DeleteJob(id string) error {
	result, err := s.db.Exec(`DELETE FROM jobs WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rows == 0 {
		return errors.New("job not found")
	}

	return nil
}

// SetJobSchedule saves or updates a job's schedule
func (s *Store) SetJobSchedule(jobID string, schedule *Schedule) error {
	yearsJSON, _ := json.Marshal(schedule.Years)
	monthsJSON, _ := json.Marshal(schedule.Months)
	daysJSON, _ := json.Marshal(schedule.Days)
	weekdaysJSON, _ := json.Marshal(schedule.Weekdays)
	hoursJSON, _ := json.Marshal(schedule.Hours)
	minutesJSON, _ := json.Marshal(schedule.Minutes)

	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO schedules (job_id, years, months, days, weekdays, hours, minutes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		jobID, string(yearsJSON), string(monthsJSON), string(daysJSON),
		string(weekdaysJSON), string(hoursJSON), string(minutesJSON),
	)
	return err
}

// GetJobSchedule retrieves a job's schedule
func (s *Store) GetJobSchedule(jobID string) (*Schedule, error) {
	schedule := &Schedule{JobID: jobID}
	var yearsJSON, monthsJSON, daysJSON, weekdaysJSON, hoursJSON, minutesJSON sql.NullString

	err := s.db.QueryRow(
		`SELECT id, years, months, days, weekdays, hours, minutes FROM schedules WHERE job_id = ?`,
		jobID,
	).Scan(&schedule.ID, &yearsJSON, &monthsJSON, &daysJSON, &weekdaysJSON, &hoursJSON, &minutesJSON)

	if err == sql.ErrNoRows {
		// Return empty schedule if none exists
		return &Schedule{JobID: jobID}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get schedule: %w", err)
	}

	// Unmarshal JSON arrays with error handling
	if yearsJSON.Valid {
		if err := json.Unmarshal([]byte(yearsJSON.String), &schedule.Years); err != nil {
			return nil, fmt.Errorf("failed to unmarshal years: %w", err)
		}
	}
	if monthsJSON.Valid {
		if err := json.Unmarshal([]byte(monthsJSON.String), &schedule.Months); err != nil {
			return nil, fmt.Errorf("failed to unmarshal months: %w", err)
		}
	}
	if daysJSON.Valid {
		if err := json.Unmarshal([]byte(daysJSON.String), &schedule.Days); err != nil {
			return nil, fmt.Errorf("failed to unmarshal days: %w", err)
		}
	}
	if weekdaysJSON.Valid {
		if err := json.Unmarshal([]byte(weekdaysJSON.String), &schedule.Weekdays); err != nil {
			return nil, fmt.Errorf("failed to unmarshal weekdays: %w", err)
		}
	}
	if hoursJSON.Valid {
		if err := json.Unmarshal([]byte(hoursJSON.String), &schedule.Hours); err != nil {
			return nil, fmt.Errorf("failed to unmarshal hours: %w", err)
		}
	}
	if minutesJSON.Valid {
		if err := json.Unmarshal([]byte(minutesJSON.String), &schedule.Minutes); err != nil {
			return nil, fmt.Errorf("failed to unmarshal minutes: %w", err)
		}
	}

	return schedule, nil
}
