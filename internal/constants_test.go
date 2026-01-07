package internal

import (
	"testing"
)

// TestScriptSizeConstant verifies the script size limit constant
func TestScriptSizeConstant(t *testing.T) {
	tests := []struct {
		name     string
		size     int
		expected bool
	}{
		{
			name:     "script at limit",
			size:     MaxScriptSize,
			expected: true,
		},
		{
			name:     "script just under limit",
			size:     MaxScriptSize - 1,
			expected: true,
		},
		{
			name:     "script over limit",
			size:     MaxScriptSize + 1,
			expected: false,
		},
		{
			name:     "empty script",
			size:     0,
			expected: true,
		},
		{
			name:     "1MB exactly",
			size:     1_000_000,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed := tt.size <= MaxScriptSize
			if allowed != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, allowed)
			}
		})
	}
}

// TestTimeoutConstants verifies timeout configuration constants
func TestTimeoutConstants(t *testing.T) {
	tests := []struct {
		name       string
		timeout    int
		shouldPass bool
	}{
		{
			name:       "minimum timeout",
			timeout:    MinTimeoutSeconds,
			shouldPass: true,
		},
		{
			name:       "below minimum timeout",
			timeout:    0,
			shouldPass: false,
		},
		{
			name:       "maximum timeout (24 hours)",
			timeout:    MaxTimeoutSeconds,
			shouldPass: true,
		},
		{
			name:       "above maximum timeout",
			timeout:    MaxTimeoutSeconds + 1,
			shouldPass: false,
		},
		{
			name:       "default timeout (1 hour)",
			timeout:    DefaultTimeoutSeconds,
			shouldPass: true,
		},
		{
			name:       "common 5-minute timeout",
			timeout:    300,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.timeout >= MinTimeoutSeconds && tt.timeout <= MaxTimeoutSeconds
			if valid != tt.shouldPass {
				t.Errorf("timeout %d: expected %v, got %v", tt.timeout, tt.shouldPass, valid)
			}
		})
	}
}

// TestRetryConstants verifies retry configuration constants
func TestRetryConstants(t *testing.T) {
	tests := []struct {
		name       string
		retries    int
		shouldPass bool
	}{
		{
			name:       "minimum retries (0)",
			retries:    MinRetryCount,
			shouldPass: true,
		},
		{
			name:       "negative retries",
			retries:    -1,
			shouldPass: false,
		},
		{
			name:       "maximum retries (10)",
			retries:    MaxRetryCount,
			shouldPass: true,
		},
		{
			name:       "above maximum retries",
			retries:    MaxRetryCount + 1,
			shouldPass: false,
		},
		{
			name:       "typical retry count",
			retries:    3,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.retries >= MinRetryCount && tt.retries <= MaxRetryCount
			if valid != tt.shouldPass {
				t.Errorf("retries %d: expected %v, got %v", tt.retries, tt.shouldPass, valid)
			}
		})
	}
}

// TestJobNameConstant verifies job name length limit
func TestJobNameConstant(t *testing.T) {
	tests := []struct {
		name       string
		nameLength int
		shouldPass bool
	}{
		{
			name:       "empty name",
			nameLength: 0,
			shouldPass: true,
		},
		{
			name:       "short name",
			nameLength: 10,
			shouldPass: true,
		},
		{
			name:       "at limit",
			nameLength: MaxJobNameLength,
			shouldPass: true,
		},
		{
			name:       "over limit",
			nameLength: MaxJobNameLength + 1,
			shouldPass: false,
		},
		{
			name:       "255 characters",
			nameLength: 255,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := tt.nameLength <= MaxJobNameLength
			if valid != tt.shouldPass {
				t.Errorf("name length %d: expected %v, got %v", tt.nameLength, tt.shouldPass, valid)
			}
		})
	}
}

// TestJobStatusConstants verifies all job status values are valid
func TestJobStatusConstants(t *testing.T) {
	validStatuses := map[string]bool{
		JobStatusPending:   true,
		JobStatusRunning:   true,
		JobStatusSuccess:   true,
		JobStatusFailure:   true,
		JobStatusTimeout:   true,
		JobStatusCancelled: true,
	}

	for status := range validStatuses {
		t.Run("status_"+status, func(t *testing.T) {
			if len(status) == 0 {
				t.Error("status should not be empty")
			}
		})
	}
}

// TestRoleConstants verifies role values
func TestRoleConstants(t *testing.T) {
	tests := []struct {
		role    string
		isAdmin bool
	}{
		{role: RoleAdmin, isAdmin: true},
		{role: RoleUser, isAdmin: false},
	}

	for _, tt := range tests {
		t.Run(tt.role, func(t *testing.T) {
			if tt.isAdmin && tt.role != RoleAdmin {
				t.Error("expected admin role")
			}
			if !tt.isAdmin && tt.role != RoleUser {
				t.Error("expected user role")
			}
		})
	}
}

// TestNotificationTriggerConstants verifies notification trigger values
func TestNotificationTriggerConstants(t *testing.T) {
	validTriggers := map[string]bool{
		NotifyAlways:  true,
		NotifySuccess: true,
		NotifyFailure: true,
	}

	if len(validTriggers) != 3 {
		t.Errorf("expected 3 notification triggers, got %d", len(validTriggers))
	}
}

// TestLogStreamConstants verifies log stream type values
func TestLogStreamConstants(t *testing.T) {
	tests := []struct {
		name     string
		stream   string
		expected string
	}{
		{name: "stdout", stream: StreamStdout, expected: "stdout"},
		{name: "stderr", stream: StreamStderr, expected: "stderr"},
		{name: "system", stream: StreamSystem, expected: "system"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.stream != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, tt.stream)
			}
		})
	}
}

// TestPaginationConstants verifies pagination limits
func TestPaginationConstants(t *testing.T) {
	if DefaultPageLimit <= 0 {
		t.Error("default page limit should be positive")
	}

	if MaxPageLimit <= 0 {
		t.Error("max page limit should be positive")
	}

	if DefaultPageLimit > MaxPageLimit {
		t.Error("default page limit should not exceed max page limit")
	}

	t.Run("valid_default_limit", func(t *testing.T) {
		if DefaultPageLimit != 100 {
			t.Errorf("expected default page limit 100, got %d", DefaultPageLimit)
		}
	})

	t.Run("valid_max_limit", func(t *testing.T) {
		if MaxPageLimit != 1000 {
			t.Errorf("expected max page limit 1000, got %d", MaxPageLimit)
		}
	})
}

// TestRequestBodySizeLimit verifies request body size limit
func TestRequestBodySizeLimit(t *testing.T) {
	if MaxRequestBodySize <= 0 {
		t.Error("request body size limit should be positive")
	}

	expectedSize := int64(10 * 1024 * 1024) // 10MB
	if MaxRequestBodySize != expectedSize {
		t.Errorf("expected request body size %d, got %d", expectedSize, MaxRequestBodySize)
	}
}

// TestExitCodeConstants verifies exit code values
func TestExitCodeConstants(t *testing.T) {
	if ExitCodeSuccess != 0 {
		t.Errorf("success exit code should be 0, got %d", ExitCodeSuccess)
	}

	if ExitCodeTimeout != 124 {
		t.Errorf("timeout exit code should be 124, got %d", ExitCodeTimeout)
	}
}

// TestConstantConsistency verifies constants are consistent across the package
func TestConstantConsistency(t *testing.T) {
	t.Run("default_working_dir", func(t *testing.T) {
		if DefaultWorkingDir != "/tmp" {
			t.Errorf("expected default working dir /tmp, got %s", DefaultWorkingDir)
		}
	})

	t.Run("default_timezone", func(t *testing.T) {
		if DefaultTimeZone != "UTC" {
			t.Errorf("expected default timezone UTC, got %s", DefaultTimeZone)
		}
	})

	t.Run("default_notify_on", func(t *testing.T) {
		if DefaultNotifyOn != NotifyFailure {
			t.Errorf("expected default notify on %s, got %s", NotifyFailure, DefaultNotifyOn)
		}
	})

	t.Run("default_timeout_equals_1_hour", func(t *testing.T) {
		if DefaultTimeoutSeconds != 3600 {
			t.Errorf("expected default timeout 3600 seconds (1 hour), got %d", DefaultTimeoutSeconds)
		}
	})
}

// BenchmarkConstantLookup benchmarks accessing constants
func BenchmarkConstantLookup(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = MaxScriptSize
		_ = MaxTimeoutSeconds
		_ = RoleAdmin
		_ = JobStatusSuccess
	}
}
