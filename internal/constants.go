package internal

import "time"

// ===== Script Configuration =====
const (
	// MaxScriptSize is the maximum allowed job script size (1MB)
	MaxScriptSize = 1_000_000
	// MaxScriptSizeReadable is the human-readable version for error messages
	MaxScriptSizeReadable = "1MB"
)

// ===== Timeout Configuration =====
const (
	// MinTimeoutSeconds is the minimum job execution timeout (1 second)
	MinTimeoutSeconds = 1
	// MaxTimeoutSeconds is the maximum job execution timeout (24 hours)
	MaxTimeoutSeconds = 86400 // 24 hours in seconds
	// DefaultTimeoutSeconds is the default job execution timeout (1 hour)
	DefaultTimeoutSeconds = 3600 // 1 hour in seconds
)

// ===== Retry Configuration =====
const (
	// MinRetryCount is the minimum number of retries allowed (0)
	MinRetryCount = 0
	// MaxRetryCount is the maximum number of retries allowed (10)
	MaxRetryCount = 10
	// MinRetryDelaySeconds is the minimum retry delay (0 seconds)
	MinRetryDelaySeconds = 0
	// MaxRetryDelaySeconds is the maximum retry delay (24 hours)
	MaxRetryDelaySeconds = 86400
)

// ===== Job Configuration =====
const (
	// MaxJobNameLength is the maximum length of a job name
	MaxJobNameLength = 255
	// DefaultWorkingDir is the default working directory for job execution
	DefaultWorkingDir = "/tmp"
	// DefaultTimeZone is the default timezone for job scheduling
	DefaultTimeZone = "UTC"
	// DefaultNotifyOn is the default notification trigger ("failure", "success", "always")
	DefaultNotifyOn = "failure"
)

// ===== Request Size Limits =====
const (
	// MaxRequestBodySize is the maximum allowed HTTP request body size (10MB)
	MaxRequestBodySize = 10 * 1024 * 1024
)

// ===== Log Streaming =====
const (
	// LogStreamBufferSize is the buffer size for log streaming (matches OS page size)
	LogStreamBufferSize = 4096 // 4KB page size
)

// ===== Channel Buffers =====
const (
	// WebSocketBroadcastChannelSize is the buffer size for WebSocket broadcast channel
	WebSocketBroadcastChannelSize = 100
	// JobQueueChannelSize is the buffer size for the job queue
	JobQueueChannelSize = 100
)

// ===== Pagination =====
const (
	// DefaultPageLimit is the default number of items per page
	DefaultPageLimit = 100
	// MaxPageLimit is the maximum number of items per page
	MaxPageLimit = 1000
)

// ===== Job Status Values =====
const (
	// JobStatusPending indicates a job is waiting to be executed
	JobStatusPending = "pending"
	// JobStatusRunning indicates a job is currently running
	JobStatusRunning = "running"
	// JobStatusSuccess indicates a job completed successfully
	JobStatusSuccess = "success"
	// JobStatusFailure indicates a job failed during execution
	JobStatusFailure = "failure"
	// JobStatusTimeout indicates a job exceeded its timeout
	JobStatusTimeout = "timeout"
	// JobStatusCancelled indicates a job was manually cancelled
	JobStatusCancelled = "cancelled"
)

// ===== User Roles =====
const (
	// RoleAdmin is the administrator role with full permissions
	RoleAdmin = "admin"
	// RoleUser is the standard user role with limited permissions
	RoleUser = "user"
)

// ===== Notification Triggers =====
const (
	// NotifyAlways sends notifications for all job executions
	NotifyAlways = "always"
	// NotifySuccess sends notifications only for successful executions
	NotifySuccess = "success"
	// NotifyFailure sends notifications only for failed executions
	NotifyFailure = "failure"
)

// ===== Trigger Types =====
const (
	// TriggerScheduled indicates a job was triggered by scheduler
	TriggerScheduled = "scheduled"
	// TriggerManual indicates a job was triggered manually
	TriggerManual = "manual"
)

// ===== Log Streams =====
const (
	// StreamStdout identifies standard output logs
	StreamStdout = "stdout"
	// StreamStderr identifies standard error logs
	StreamStderr = "stderr"
	// StreamSystem identifies system/internal logs
	StreamSystem = "system"
)

// ===== Database & Cleanup =====
const (
	// DefaultLogRetentionDays is the default number of days to retain job logs
	DefaultLogRetentionDays = 30
	// LogCleanupInterval is how often to run log cleanup
	LogCleanupInterval = 24 * time.Hour
	// SchedulerCheckInterval is how often the scheduler checks for jobs to run
	SchedulerCheckInterval = time.Minute
)

// ===== CORS =====
const (
	// CORSAllowAllOrigins represents wildcard CORS origin
	CORSAllowAllOrigins = "*"
)

// ===== Exit Codes =====
const (
	// ExitCodeSuccess is the standard success exit code
	ExitCodeSuccess = 0
	// ExitCodeTimeout is the exit code for timeout
	ExitCodeTimeout = 124 // Standard timeout exit code
)
