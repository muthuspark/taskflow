package api

import (
	"fmt"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

// JobValidator validates job creation and update requests
type JobValidator struct{}

// NewJobValidator creates a new job validator
func NewJobValidator() *JobValidator {
	return &JobValidator{}
}

// JobRequest represents the common fields for create/update requests
type JobRequest struct {
	Name              string
	Description       string
	Script            string
	WorkingDir        string
	TimeoutSeconds    int
	RetryCount        int
	RetryDelaySeconds int
	NotifyEmails      string
	NotifyOn          string
	Timezone          string
}

// ValidationError represents a validation error with code
type ValidationError struct {
	Message string
	Code    string
}

// ValidateJobRequest validates all job fields consistently
// Returns nil if valid, otherwise returns a ValidationError
func (v *JobValidator) ValidateJobRequest(req *JobRequest) *ValidationError {
	// Validate required fields
	if req.Name == "" || req.Script == "" {
		return &ValidationError{
			Message: "Name and script are required",
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate name length
	if len(req.Name) > internal.MaxJobNameLength {
		return &ValidationError{
			Message: fmt.Sprintf("Job name too long (max %d characters)", internal.MaxJobNameLength),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate script length
	if len(req.Script) > internal.MaxScriptSize {
		return &ValidationError{
			Message: fmt.Sprintf("Script too long (max %s)", internal.MaxScriptSizeReadable),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate timeout
	if req.TimeoutSeconds < internal.MinTimeoutSeconds || req.TimeoutSeconds > internal.MaxTimeoutSeconds {
		return &ValidationError{
			Message: fmt.Sprintf("Timeout must be between %d and %d seconds", internal.MinTimeoutSeconds, internal.MaxTimeoutSeconds),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate retry count
	if req.RetryCount < internal.MinRetryCount || req.RetryCount > internal.MaxRetryCount {
		return &ValidationError{
			Message: fmt.Sprintf("Retry count must be between %d and %d", internal.MinRetryCount, internal.MaxRetryCount),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate retry delay
	if req.RetryDelaySeconds < internal.MinRetryDelaySeconds || req.RetryDelaySeconds > internal.MaxRetryDelaySeconds {
		return &ValidationError{
			Message: fmt.Sprintf("Retry delay must be between %d and %d seconds", internal.MinRetryDelaySeconds, internal.MaxRetryDelaySeconds),
			Code:    "VALIDATION_ERROR",
		}
	}

	// Validate notify_on enum
	if !v.isValidNotifyOn(req.NotifyOn) {
		return &ValidationError{
			Message: "Invalid notify_on value",
			Code:    "VALIDATION_ERROR",
		}
	}

	return nil
}

// isValidNotifyOn checks if the notify_on value is valid
func (v *JobValidator) isValidNotifyOn(notifyOn string) bool {
	if notifyOn == "" {
		return true // Empty is allowed (uses default)
	}
	return notifyOn == internal.NotifyAlways ||
		notifyOn == internal.NotifyFailure ||
		notifyOn == internal.NotifySuccess
}

// ApplyDefaults applies default values to job request fields
func (v *JobValidator) ApplyDefaults(req *JobRequest) {
	if req.WorkingDir == "" {
		req.WorkingDir = internal.DefaultWorkingDir
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = internal.DefaultTimeoutSeconds
	}
	if req.NotifyOn == "" {
		req.NotifyOn = internal.DefaultNotifyOn
	}
	if req.Timezone == "" {
		req.Timezone = internal.DefaultTimeZone
	}
}

// ToJobModel converts a validated request to a job model
func (v *JobValidator) ToJobModel(req *JobRequest, jobID *string) *store.Job {
	job := &store.Job{
		Name:              req.Name,
		Description:       req.Description,
		Script:            req.Script,
		WorkingDir:        req.WorkingDir,
		TimeoutSeconds:    req.TimeoutSeconds,
		RetryCount:        req.RetryCount,
		RetryDelaySeconds: req.RetryDelaySeconds,
		NotifyEmails:      req.NotifyEmails,
		NotifyOn:          req.NotifyOn,
		Timezone:          req.Timezone,
	}
	if jobID != nil {
		job.ID = *jobID
	}
	return job
}

// ValidateScheduleRequest validates schedule fields
type ScheduleRequest struct {
	Years    []int
	Months   []int
	Days     []int
	Weekdays []int
	Hours    []int
	Minutes  []int
}

// ValidateScheduleRequest validates all schedule fields
func (v *JobValidator) ValidateScheduleRequest(req *ScheduleRequest) *ValidationError {
	// Validate months
	for _, m := range req.Months {
		if m < 1 || m > 12 {
			return &ValidationError{
				Message: "Months must be between 1-12",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	// Validate days
	for _, d := range req.Days {
		if d < 1 || d > 31 {
			return &ValidationError{
				Message: "Days must be between 1-31",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	// Validate hours
	for _, h := range req.Hours {
		if h < 0 || h > 23 {
			return &ValidationError{
				Message: "Hours must be between 0-23",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	// Validate minutes
	for _, min := range req.Minutes {
		if min < 0 || min > 59 {
			return &ValidationError{
				Message: "Minutes must be between 0-59",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	// Validate weekdays
	for _, wd := range req.Weekdays {
		if wd < 0 || wd > 6 {
			return &ValidationError{
				Message: "Weekdays must be between 0-6",
				Code:    "VALIDATION_ERROR",
			}
		}
	}

	return nil
}
