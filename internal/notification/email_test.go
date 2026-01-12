package notification

import (
	"errors"
	"testing"
	"time"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

// mockSettingsProvider implements SMTPSettingsProvider for testing
type mockSettingsProvider struct {
	settings *store.SMTPSettings
	err      error
}

func (m *mockSettingsProvider) GetSMTPSettings() (*store.SMTPSettings, error) {
	return m.settings, m.err
}

func TestShouldNotify(t *testing.T) {
	tests := []struct {
		name     string
		notifyOn string
		status   string
		want     bool
	}{
		// NotifyAlways cases
		{"always_on_success", internal.NotifyAlways, internal.JobStatusSuccess, true},
		{"always_on_failure", internal.NotifyAlways, internal.JobStatusFailure, true},
		{"always_on_timeout", internal.NotifyAlways, internal.JobStatusTimeout, true},
		{"always_on_running", internal.NotifyAlways, internal.JobStatusRunning, true},

		// NotifySuccess cases
		{"success_on_success", internal.NotifySuccess, internal.JobStatusSuccess, true},
		{"success_on_failure", internal.NotifySuccess, internal.JobStatusFailure, false},
		{"success_on_timeout", internal.NotifySuccess, internal.JobStatusTimeout, false},

		// NotifyFailure cases
		{"failure_on_success", internal.NotifyFailure, internal.JobStatusSuccess, false},
		{"failure_on_failure", internal.NotifyFailure, internal.JobStatusFailure, true},
		{"failure_on_timeout", internal.NotifyFailure, internal.JobStatusTimeout, true},

		// Default (empty) should use DefaultNotifyOn (failure)
		{"empty_on_success", "", internal.JobStatusSuccess, false},
		{"empty_on_failure", "", internal.JobStatusFailure, true},

		// Invalid notifyOn value
		{"invalid_on_success", "invalid", internal.JobStatusSuccess, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldNotify(tt.notifyOn, tt.status)
			if got != tt.want {
				t.Errorf("shouldNotify(%q, %q) = %v, want %v", tt.notifyOn, tt.status, got, tt.want)
			}
		})
	}
}

func TestParseEmails(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty_string", "", nil},
		{"single_email", "test@example.com", []string{"test@example.com"}},
		{"multiple_emails", "a@b.com, c@d.com", []string{"a@b.com", "c@d.com"}},
		{"with_whitespace", "  a@b.com  ,  c@d.com  ", []string{"a@b.com", "c@d.com"}},
		{"invalid_no_at", "invalid", nil},
		{"mixed_valid_invalid", "valid@test.com, invalid, another@test.com", []string{"valid@test.com", "another@test.com"}},
		{"empty_between_commas", "a@b.com,,c@d.com", []string{"a@b.com", "c@d.com"}},
		{"only_commas", ",,,", nil},
		{"whitespace_only", "   ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseEmails(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("parseEmails(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("parseEmails(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestIsConfigured(t *testing.T) {
	tests := []struct {
		name     string
		settings *store.SMTPSettings
		want     bool
	}{
		{"fully_configured", &store.SMTPSettings{Server: "smtp.test.com", Port: 587}, true},
		{"missing_server", &store.SMTPSettings{Server: "", Port: 587}, false},
		{"missing_port", &store.SMTPSettings{Server: "smtp.test.com", Port: 0}, false},
		{"both_missing", &store.SMTPSettings{Server: "", Port: 0}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isConfigured(tt.settings)
			if got != tt.want {
				t.Errorf("isConfigured() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name       string
		durationMs *int64
		want       string
	}{
		{"nil", nil, notAvailable},
		{"zero", ptr(int64(0)), "0.0 seconds"},
		{"seconds", ptr(int64(30000)), "30.0 seconds"},
		{"minutes", ptr(int64(120000)), "2.0 minutes"},
		{"hours", ptr(int64(7200000)), "2.0 hours"},
		{"sub_second", ptr(int64(500)), "0.5 seconds"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatDuration(tt.durationMs)
			if got != tt.want {
				t.Errorf("formatDuration() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name string
		time *time.Time
		want string
	}{
		{"nil", nil, notAvailable},
		{"valid", timePtr(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)), "2024-01-15 10:30:00 UTC"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTime(tt.time)
			if got != tt.want {
				t.Errorf("formatTime() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatExitCode(t *testing.T) {
	tests := []struct {
		name string
		code *int
		want string
	}{
		{"nil", nil, notAvailable},
		{"zero", intPtr(0), "0"},
		{"positive", intPtr(1), "1"},
		{"negative", intPtr(-1), "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatExitCode(tt.code)
			if got != tt.want {
				t.Errorf("formatExitCode() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatErrorSection(t *testing.T) {
	tests := []struct {
		name   string
		errMsg *string
		empty  bool
	}{
		{"nil", nil, true},
		{"empty_string", strPtr(""), true},
		{"with_message", strPtr("Something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatErrorSection(tt.errMsg)
			if tt.empty && got != "" {
				t.Errorf("formatErrorSection() = %q, want empty", got)
			}
			if !tt.empty && got == "" {
				t.Errorf("formatErrorSection() = empty, want non-empty")
			}
		})
	}
}

func TestGetStatusEmoji(t *testing.T) {
	tests := []struct {
		status string
		want   string
	}{
		{internal.JobStatusSuccess, "✅"},
		{internal.JobStatusFailure, "❌"},
		{internal.JobStatusTimeout, "⏰"},
		{internal.JobStatusRunning, "ℹ️"},
		{"unknown", "ℹ️"},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := getStatusEmoji(tt.status)
			if got != tt.want {
				t.Errorf("getStatusEmoji(%q) = %q, want %q", tt.status, got, tt.want)
			}
		})
	}
}

func TestSanitizeHeader(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"no_special_chars", "Normal Subject", "Normal Subject"},
		{"with_newline", "Subject\nInjected", "SubjectInjected"},
		{"with_carriage_return", "Subject\rInjected", "SubjectInjected"},
		{"with_both", "Subject\r\nInjected: header", "SubjectInjected: header"},
		{"multiple_newlines", "A\n\n\nB", "AB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeHeader(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeHeader(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildEmailContent(t *testing.T) {
	job := &store.Job{
		ID:          "job-123",
		Name:        "Test Job",
		Description: "A test job",
	}

	duration := int64(5000)
	exitCode := 0
	finishedAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	run := &store.Run{
		ID:          "run-456",
		Status:      internal.JobStatusSuccess,
		TriggerType: "manual",
		DurationMs:  &duration,
		ExitCode:    &exitCode,
		FinishedAt:  &finishedAt,
	}

	subject, body := buildEmailContent(job, run)

	// Check subject
	if subject == "" {
		t.Error("buildEmailContent() subject is empty")
	}
	if !containsAll(subject, "[TaskFlow]", "SUCCESS", "Test Job") {
		t.Errorf("buildEmailContent() subject = %q, missing expected parts", subject)
	}

	// Check body
	if body == "" {
		t.Error("buildEmailContent() body is empty")
	}
	if !containsAll(body, "Test Job", "A test job", "run-456", "manual", "5.0 seconds") {
		t.Errorf("buildEmailContent() body missing expected content")
	}
}

func TestBuildMessage(t *testing.T) {
	msg := buildMessage("TaskFlow", "noreply@test.com", []string{"user@test.com"}, "Test Subject", "Test body")

	expectedParts := []string{
		"From: TaskFlow <noreply@test.com>",
		"To: user@test.com",
		"Subject: Test Subject",
		"MIME-Version: 1.0",
		"Content-Type: text/plain",
		"Test body",
	}

	for _, part := range expectedParts {
		if !contains(msg, part) {
			t.Errorf("buildMessage() missing %q", part)
		}
	}
}

func TestSendJobNotification_NoNotifyNeeded(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: &store.SMTPSettings{Server: "smtp.test.com", Port: 587},
	}
	notifier := New(provider)

	job := &store.Job{NotifyOn: internal.NotifyFailure}
	run := &store.Run{Status: internal.JobStatusSuccess}

	err := notifier.SendJobNotification(job, run)
	if err != nil {
		t.Errorf("SendJobNotification() error = %v, want nil", err)
	}
}

func TestSendJobNotification_NoEmails(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: &store.SMTPSettings{Server: "smtp.test.com", Port: 587},
	}
	notifier := New(provider)

	job := &store.Job{NotifyOn: internal.NotifyAlways, NotifyEmails: ""}
	run := &store.Run{Status: internal.JobStatusSuccess}

	err := notifier.SendJobNotification(job, run)
	if err != nil {
		t.Errorf("SendJobNotification() error = %v, want nil", err)
	}
}

func TestSendJobNotification_SettingsError(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: nil,
		err:      errors.New("database error"),
	}
	notifier := New(provider)

	job := &store.Job{NotifyOn: internal.NotifyAlways, NotifyEmails: "test@test.com"}
	run := &store.Run{Status: internal.JobStatusSuccess}

	err := notifier.SendJobNotification(job, run)
	if err == nil {
		t.Error("SendJobNotification() expected error, got nil")
	}
}

func TestSendJobNotification_SMTPNotConfigured(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: &store.SMTPSettings{Server: "", Port: 0},
	}
	notifier := New(provider)

	job := &store.Job{NotifyOn: internal.NotifyAlways, NotifyEmails: "test@test.com"}
	run := &store.Run{Status: internal.JobStatusSuccess}

	// Should return nil (skip gracefully) when SMTP not configured
	err := notifier.SendJobNotification(job, run)
	if err != nil {
		t.Errorf("SendJobNotification() error = %v, want nil", err)
	}
}

// Helper functions
func ptr(i int64) *int64    { return &i }
func intPtr(i int) *int     { return &i }
func strPtr(s string) *string { return &s }
func timePtr(t time.Time) *time.Time { return &t }

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsAll(s string, substrs ...string) bool {
	for _, substr := range substrs {
		if !contains(s, substr) {
			return false
		}
	}
	return true
}
