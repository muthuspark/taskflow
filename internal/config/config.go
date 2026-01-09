package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port              int
	DBPath            string
	JWTSecret         string
	LogLevel          string
	SMTPServer        string
	SMTPPort          int
	SMTPUsername      string
	SMTPPassword      string
	AllowedOrigins    string
	LogRetentionDays  int
	APIBasePath       string
}

func Load() *Config {
	cfg := &Config{
		Port:             8080,
		DBPath:           "taskflow.db",
		LogLevel:         "info",
		LogRetentionDays: 30,
		APIBasePath:      "/api",
	}

	if port := os.Getenv("PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Port = p
		}
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.DBPath = dbPath
	}

	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		cfg.JWTSecret = secret
	}

	if level := os.Getenv("LOG_LEVEL"); level != "" {
		cfg.LogLevel = level
	}

	if server := os.Getenv("SMTP_SERVER"); server != "" {
		cfg.SMTPServer = server
	}

	if port := os.Getenv("SMTP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.SMTPPort = p
		}
	}

	if user := os.Getenv("SMTP_USERNAME"); user != "" {
		cfg.SMTPUsername = user
	}

	if pass := os.Getenv("SMTP_PASSWORD"); pass != "" {
		cfg.SMTPPassword = pass
	}

	if origins := os.Getenv("ALLOWED_ORIGINS"); origins != "" {
		cfg.AllowedOrigins = origins
	} else {
		cfg.AllowedOrigins = "*"
	}

	if days := os.Getenv("LOG_RETENTION_DAYS"); days != "" {
		if d, err := strconv.Atoi(days); err == nil {
			cfg.LogRetentionDays = d
		}
	}

	if basePath := os.Getenv("API_BASE_PATH"); basePath != "" {
		// Ensure base path starts with / and doesn't end with /
		if !strings.HasPrefix(basePath, "/") {
			basePath = "/" + basePath
		}
		basePath = strings.TrimSuffix(basePath, "/")
		cfg.APIBasePath = basePath
	}

	return cfg
}
