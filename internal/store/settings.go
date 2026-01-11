package store

import (
	"database/sql"
	"time"
)

// GetSetting retrieves a setting by key
func (s *Store) GetSetting(key string) (*Setting, error) {
	var setting Setting
	err := s.db.QueryRow(
		`SELECT key, value, updated_at FROM settings WHERE key = ?`,
		key,
	).Scan(&setting.Key, &setting.Value, &setting.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &setting, nil
}

// SetSetting creates or updates a setting
func (s *Store) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, ?)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at`,
		key, value, time.Now(),
	)
	return err
}

// GetSettings retrieves multiple settings by key prefix
func (s *Store) GetSettings(keyPrefix string) ([]Setting, error) {
	rows, err := s.db.Query(
		`SELECT key, value, updated_at FROM settings WHERE key LIKE ?`,
		keyPrefix+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var settings []Setting
	for rows.Next() {
		var setting Setting
		if err := rows.Scan(&setting.Key, &setting.Value, &setting.UpdatedAt); err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	return settings, rows.Err()
}

// DeleteSetting removes a setting by key
func (s *Store) DeleteSetting(key string) error {
	_, err := s.db.Exec(`DELETE FROM settings WHERE key = ?`, key)
	return err
}

// SMTP Settings helpers

// SMTPSettings holds SMTP configuration
type SMTPSettings struct {
	Server   string `json:"server"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	FromName string `json:"from_name"`
	FromEmail string `json:"from_email"`
}

// GetSMTPSettings retrieves all SMTP settings
func (s *Store) GetSMTPSettings() (*SMTPSettings, error) {
	settings := &SMTPSettings{}

	// Get each setting
	if setting, err := s.GetSetting("smtp_server"); err != nil {
		return nil, err
	} else if setting != nil {
		settings.Server = setting.Value
	}

	if setting, err := s.GetSetting("smtp_port"); err != nil {
		return nil, err
	} else if setting != nil {
		var port int
		if _, err := parseIntSafe(setting.Value, &port); err == nil {
			settings.Port = port
		}
	}

	if setting, err := s.GetSetting("smtp_username"); err != nil {
		return nil, err
	} else if setting != nil {
		settings.Username = setting.Value
	}

	if setting, err := s.GetSetting("smtp_password"); err != nil {
		return nil, err
	} else if setting != nil {
		settings.Password = setting.Value
	}

	if setting, err := s.GetSetting("smtp_from_name"); err != nil {
		return nil, err
	} else if setting != nil {
		settings.FromName = setting.Value
	}

	if setting, err := s.GetSetting("smtp_from_email"); err != nil {
		return nil, err
	} else if setting != nil {
		settings.FromEmail = setting.Value
	}

	return settings, nil
}

// SetSMTPSettings saves all SMTP settings
func (s *Store) SetSMTPSettings(settings *SMTPSettings) error {
	if err := s.SetSetting("smtp_server", settings.Server); err != nil {
		return err
	}
	if err := s.SetSetting("smtp_port", intToString(settings.Port)); err != nil {
		return err
	}
	if err := s.SetSetting("smtp_username", settings.Username); err != nil {
		return err
	}
	if err := s.SetSetting("smtp_password", settings.Password); err != nil {
		return err
	}
	if err := s.SetSetting("smtp_from_name", settings.FromName); err != nil {
		return err
	}
	if err := s.SetSetting("smtp_from_email", settings.FromEmail); err != nil {
		return err
	}
	return nil
}

// Helper functions
func parseIntSafe(s string, result *int) (bool, error) {
	if s == "" {
		return false, nil
	}
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + int(c-'0')
	}
	*result = n
	return true, nil
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}
