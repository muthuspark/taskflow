package store

import (
	"database/sql"
	"encoding/json"
	"time"
)

// User represents a system user
type User struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      string    `json:"role"` // "admin" or "user"
	CreatedAt time.Time `json:"created_at"`
	LastLogin *time.Time `json:"last_login"`
	// PasswordHash is not exposed in JSON
	PasswordHash string `json:"-"`
}

// Job represents a scheduled job
type Job struct {
	ID                string         `json:"id"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	Script            string         `json:"script"`
	WorkingDir        string         `json:"working_dir"`
	TimeoutSeconds    int            `json:"timeout_seconds"`
	RetryCount        int            `json:"retry_count"`
	RetryDelaySeconds int            `json:"retry_delay_seconds"`
	Enabled           bool           `json:"enabled"`
	NotifyEmails      string         `json:"notify_emails"`
	NotifyOn          string         `json:"notify_on"` // "always", "failure", "success"
	Timezone          string         `json:"timezone"`
	CreatedBy         int            `json:"created_by"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
}

// Schedule represents cron-like scheduling
type Schedule struct {
	ID       int            `json:"id"`
	JobID    string         `json:"job_id"`
	Years    []int          `json:"years"`    // nil = any
	Months   []int          `json:"months"`   // 1-12
	Days     []int          `json:"days"`     // 1-31
	Weekdays []int          `json:"weekdays"` // 0-6 (Sun-Sat)
	Hours    []int          `json:"hours"`    // 0-23
	Minutes  []int          `json:"minutes"`  // 0-59
}

// Run represents a job execution
type Run struct {
	ID          string         `json:"id"`
	JobID       string         `json:"job_id"`
	Status      string         `json:"status"` // "pending", "running", "success", "failure", "timeout", "cancelled"
	ExitCode    *int           `json:"exit_code"`
	TriggerType string         `json:"trigger_type"` // "scheduled", "manual"
	StartedAt   *time.Time     `json:"started_at"`
	FinishedAt  *time.Time     `json:"finished_at"`
	DurationMs  *int64         `json:"duration_ms"`
	ErrorMsg    *string        `json:"error_message"`
}

// LogEntry represents a log line from job execution
type LogEntry struct {
	ID        int       `json:"id"`
	RunID     string    `json:"run_id"`
	Timestamp time.Time `json:"timestamp"`
	Stream    string    `json:"stream"` // "stdout", "stderr", "system"
	Content   string    `json:"content"`
}

// Metric represents resource usage at a point in time
type Metric struct {
	ID            int       `json:"id"`
	RunID         string    `json:"run_id"`
	Timestamp     time.Time `json:"timestamp"`
	CPUPercent    float64   `json:"cpu_percent"`
	MemoryBytes   int64     `json:"memory_bytes"`
	MemoryPercent float64   `json:"memory_percent"`
}

// MetricAggregate represents aggregated metrics over a period
type MetricAggregate struct {
	ID               int       `json:"id"`
	JobID            string    `json:"job_id"`
	PeriodType       string    `json:"period_type"` // "hourly", "daily"
	PeriodStart      time.Time `json:"period_start"`
	RunCount         int       `json:"run_count"`
	AvgDurationMs    int64     `json:"avg_duration_ms"`
	AvgCPUPercent    float64   `json:"avg_cpu_percent"`
	AvgMemoryBytes   int64     `json:"avg_memory_bytes"`
	MaxCPUPercent    float64   `json:"max_cpu_percent"`
	MaxMemoryBytes   int64     `json:"max_memory_bytes"`
	SuccessCount     int       `json:"success_count"`
	FailureCount     int       `json:"failure_count"`
}

// Setting represents a key-value setting
type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Scan helpers for JSON array fields
func (s *Schedule) ScanYears(val interface{}) error {
	return scanJSONArray(val, &s.Years)
}

func (s *Schedule) ScanMonths(val interface{}) error {
	return scanJSONArray(val, &s.Months)
}

func (s *Schedule) ScanDays(val interface{}) error {
	return scanJSONArray(val, &s.Days)
}

func (s *Schedule) ScanWeekdays(val interface{}) error {
	return scanJSONArray(val, &s.Weekdays)
}

func (s *Schedule) ScanHours(val interface{}) error {
	return scanJSONArray(val, &s.Hours)
}

func (s *Schedule) ScanMinutes(val interface{}) error {
	return scanJSONArray(val, &s.Minutes)
}

func scanJSONArray(val interface{}, target *[]int) error {
	if val == nil {
		*target = nil
		return nil
	}

	// Handle string representation from database
	if str, ok := val.(string); ok {
		return json.Unmarshal([]byte(str), target)
	}

	// Handle byte slice representation
	if bytes, ok := val.([]byte); ok {
		return json.Unmarshal(bytes, target)
	}

	// Initialize empty array for unexpected types
	*target = []int{}
	return nil
}

// NullInt64ToPointer converts sql.NullInt64 to *int64
func NullInt64ToPointer(n sql.NullInt64) *int64 {
	if n.Valid {
		return &n.Int64
	}
	return nil
}

// NullTimeToPointer converts sql.NullTime to *time.Time
func NullTimeToPointer(n sql.NullTime) *time.Time {
	if n.Valid {
		return &n.Time
	}
	return nil
}

// PointerToNullInt64 converts *int to sql.NullInt64
func PointerToNullInt64(p *int) sql.NullInt64 {
	if p != nil {
		return sql.NullInt64{Int64: int64(*p), Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

// PointerToNullInt64Ptr converts *int64 to sql.NullInt64
func PointerToNullInt64Ptr(p *int64) sql.NullInt64 {
	if p != nil {
		return sql.NullInt64{Int64: *p, Valid: true}
	}
	return sql.NullInt64{Valid: false}
}

// PointerToNullTime converts *time.Time to sql.NullTime
func PointerToNullTime(p *time.Time) sql.NullTime {
	if p != nil {
		return sql.NullTime{Time: *p, Valid: true}
	}
	return sql.NullTime{Valid: false}
}
