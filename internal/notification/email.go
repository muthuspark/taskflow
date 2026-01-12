package notification

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/smtp"
	"strings"
	"time"

	internal "github.com/taskflow/taskflow/internal"
	"github.com/taskflow/taskflow/internal/store"
)

const (
	smtpPortTLS       = 465
	defaultFromName   = "TaskFlow"
	notAvailable      = "N/A"
	emailSubjectPrefix = "[TaskFlow]"
)

// SMTPSettingsProvider abstracts SMTP settings retrieval for testability
type SMTPSettingsProvider interface {
	GetSMTPSettings() (*store.SMTPSettings, error)
}

// Notifier handles sending notifications
type Notifier struct {
	settingsProvider SMTPSettingsProvider
}

// New creates a new Notifier
func New(provider SMTPSettingsProvider) *Notifier {
	return &Notifier{settingsProvider: provider}
}

// SendJobNotification sends email notification for a completed job run
func (n *Notifier) SendJobNotification(job *store.Job, run *store.Run) error {
	if !shouldNotify(job.NotifyOn, run.Status) {
		return nil
	}

	emails := parseEmails(job.NotifyEmails)
	if len(emails) == 0 {
		return nil
	}

	settings, err := n.settingsProvider.GetSMTPSettings()
	if err != nil {
		return fmt.Errorf("failed to get SMTP settings: %w", err)
	}

	if !isConfigured(settings) {
		log.Printf("SMTP not configured, skipping notification for job %s", job.ID)
		return nil
	}

	subject, body := buildEmailContent(job, run)
	return sendEmail(settings, emails, subject, body)
}

// shouldNotify determines if a notification should be sent
func shouldNotify(notifyOn, status string) bool {
	if notifyOn == "" {
		notifyOn = internal.DefaultNotifyOn
	}

	switch notifyOn {
	case internal.NotifyAlways:
		return true
	case internal.NotifySuccess:
		return status == internal.JobStatusSuccess
	case internal.NotifyFailure:
		return status == internal.JobStatusFailure || status == internal.JobStatusTimeout
	default:
		return false
	}
}

// isConfigured checks if SMTP settings are properly configured
func isConfigured(settings *store.SMTPSettings) bool {
	return settings.Server != "" && settings.Port != 0
}

// parseEmails splits a comma-separated list of emails
func parseEmails(emailsStr string) []string {
	if emailsStr == "" {
		return nil
	}

	parts := strings.Split(emailsStr, ",")
	emails := make([]string, 0, len(parts))
	for _, email := range parts {
		email = strings.TrimSpace(email)
		if email != "" && strings.Contains(email, "@") {
			emails = append(emails, email)
		}
	}
	return emails
}

// buildEmailContent creates the subject and body for the notification email
func buildEmailContent(job *store.Job, run *store.Run) (subject, body string) {
	statusEmoji := getStatusEmoji(run.Status)
	statusText := strings.ToUpper(run.Status)

	subject = fmt.Sprintf("%s %s Job %s: %s", emailSubjectPrefix, statusEmoji, statusText, job.Name)

	body = fmt.Sprintf(`TaskFlow Job Notification
=========================

Job: %s
Description: %s
Status: %s %s
Run ID: %s
Trigger: %s

Execution Details:
------------------
Duration: %s
Exit Code: %s
Finished: %s
%s
---
This is an automated notification from TaskFlow.
`,
		job.Name,
		job.Description,
		statusEmoji,
		statusText,
		run.ID,
		run.TriggerType,
		formatDuration(run.DurationMs),
		formatExitCode(run.ExitCode),
		formatTime(run.FinishedAt),
		formatErrorSection(run.ErrorMsg),
	)

	return subject, body
}

// formatDuration formats milliseconds into a human-readable string
func formatDuration(durationMs *int64) string {
	if durationMs == nil {
		return notAvailable
	}

	duration := time.Duration(*durationMs) * time.Millisecond
	switch {
	case duration >= time.Hour:
		return fmt.Sprintf("%.1f hours", duration.Hours())
	case duration >= time.Minute:
		return fmt.Sprintf("%.1f minutes", duration.Minutes())
	default:
		return fmt.Sprintf("%.1f seconds", duration.Seconds())
	}
}

// formatTime formats a timestamp for display
func formatTime(t *time.Time) string {
	if t == nil {
		return notAvailable
	}
	return t.Format("2006-01-02 15:04:05 MST")
}

// formatExitCode formats an exit code for display
func formatExitCode(code *int) string {
	if code == nil {
		return notAvailable
	}
	return fmt.Sprintf("%d", *code)
}

// formatErrorSection formats the error message section
func formatErrorSection(errMsg *string) string {
	if errMsg == nil || *errMsg == "" {
		return ""
	}
	return fmt.Sprintf(`
Error Message:
--------------
%s
`, *errMsg)
}

// getStatusEmoji returns an emoji for the job status
func getStatusEmoji(status string) string {
	switch status {
	case internal.JobStatusSuccess:
		return "✅"
	case internal.JobStatusFailure:
		return "❌"
	case internal.JobStatusTimeout:
		return "⏰"
	default:
		return "ℹ️"
	}
}

// sanitizeHeader removes newlines to prevent header injection
func sanitizeHeader(s string) string {
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}

// sendEmail sends an email via SMTP
func sendEmail(settings *store.SMTPSettings, to []string, subject, body string) error {
	from := settings.FromEmail
	if from == "" {
		from = settings.Username
	}
	if from == "" {
		return fmt.Errorf("no from email address configured")
	}

	fromName := settings.FromName
	if fromName == "" {
		fromName = defaultFromName
	}

	msg := buildMessage(sanitizeHeader(fromName), sanitizeHeader(from), to, sanitizeHeader(subject), body)
	addr := fmt.Sprintf("%s:%d", settings.Server, settings.Port)

	if settings.Port == smtpPortTLS {
		return sendWithTLS(settings, addr, from, to, msg)
	}
	return sendWithSTARTTLS(settings, addr, from, to, msg)
}

// buildMessage constructs the email message with headers
func buildMessage(fromName, from string, to []string, subject, body string) string {
	var msg strings.Builder
	msg.WriteString(fmt.Sprintf("From: %s <%s>\r\n", fromName, from))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msg.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(body)
	return msg.String()
}

// sendWithTLS sends email using implicit TLS (port 465)
func sendWithTLS(settings *store.SMTPSettings, addr, from string, to []string, msg string) error {
	tlsConfig := &tls.Config{ServerName: settings.Server}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, settings.Server)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	return sendWithClient(client, settings, from, to, msg)
}

// sendWithSTARTTLS sends email using STARTTLS (ports 25, 587)
func sendWithSTARTTLS(settings *store.SMTPSettings, addr, from string, to []string, msg string) error {
	client, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	if err := client.Hello("localhost"); err != nil {
		return fmt.Errorf("SMTP HELLO failed: %w", err)
	}

	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{ServerName: settings.Server}
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("SMTP STARTTLS failed: %w", err)
		}
	}

	return sendWithClient(client, settings, from, to, msg)
}

// sendWithClient handles authentication and message delivery (DRY extracted)
func sendWithClient(client *smtp.Client, settings *store.SMTPSettings, from string, to []string, msg string) error {
	if settings.Username != "" && settings.Password != "" {
		auth := smtp.PlainAuth("", settings.Username, settings.Password, settings.Server)
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}

	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP MAIL command failed: %w", err)
	}

	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("SMTP RCPT command failed for %s: %w", recipient, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("SMTP DATA command failed: %w", err)
	}

	if _, err = w.Write([]byte(msg)); err != nil {
		return fmt.Errorf("failed to write email body: %w", err)
	}

	if err = w.Close(); err != nil {
		return fmt.Errorf("failed to close email writer: %w", err)
	}

	return client.Quit()
}
