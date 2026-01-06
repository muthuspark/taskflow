package store

import (
	"database/sql"
	"fmt"
)

// Migrations contains all database schema migrations
var migrations = []struct {
	name  string
	query string
}{
	{
		name: "001_create_users",
		query: `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    email TEXT,
    role TEXT DEFAULT 'user',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    last_login DATETIME
);
`,
	},
	{
		name: "002_create_jobs",
		query: `
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    script TEXT NOT NULL,
    working_dir TEXT DEFAULT '/tmp',
    timeout_seconds INTEGER DEFAULT 3600,
    retry_count INTEGER DEFAULT 0,
    retry_delay_seconds INTEGER DEFAULT 60,
    enabled BOOLEAN DEFAULT 1,
    notify_emails TEXT,
    notify_on TEXT DEFAULT 'failure',
    timezone TEXT DEFAULT 'UTC',
    created_by INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`,
	},
	{
		name: "003_create_schedules",
		query: `
CREATE TABLE IF NOT EXISTS schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT REFERENCES jobs(id) ON DELETE CASCADE,
    years TEXT,
    months TEXT,
    days TEXT,
    weekdays TEXT,
    hours TEXT,
    minutes TEXT,
    UNIQUE(job_id)
);
`,
	},
	{
		name: "004_create_runs",
		query: `
CREATE TABLE IF NOT EXISTS runs (
    id TEXT PRIMARY KEY,
    job_id TEXT REFERENCES jobs(id) ON DELETE CASCADE,
    status TEXT NOT NULL,
    exit_code INTEGER,
    trigger_type TEXT,
    started_at DATETIME,
    finished_at DATETIME,
    duration_ms INTEGER,
    error_message TEXT
);
CREATE INDEX IF NOT EXISTS idx_runs_job_id ON runs(job_id);
CREATE INDEX IF NOT EXISTS idx_runs_started_at ON runs(started_at);
`,
	},
	{
		name: "005_create_logs",
		query: `
CREATE TABLE IF NOT EXISTS logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT REFERENCES runs(id) ON DELETE CASCADE,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    stream TEXT,
    content TEXT
);
CREATE INDEX IF NOT EXISTS idx_logs_run_id ON logs(run_id);
`,
	},
	{
		name: "006_create_metrics",
		query: `
CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    run_id TEXT REFERENCES runs(id) ON DELETE CASCADE,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    cpu_percent REAL,
    memory_bytes INTEGER,
    memory_percent REAL
);
CREATE INDEX IF NOT EXISTS idx_metrics_run_id ON metrics(run_id);
`,
	},
	{
		name: "007_create_metrics_aggregate",
		query: `
CREATE TABLE IF NOT EXISTS metrics_aggregate (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    job_id TEXT REFERENCES jobs(id) ON DELETE CASCADE,
    period_type TEXT,
    period_start DATETIME,
    run_count INTEGER,
    avg_duration_ms INTEGER,
    avg_cpu_percent REAL,
    avg_memory_bytes INTEGER,
    max_cpu_percent REAL,
    max_memory_bytes INTEGER,
    success_count INTEGER,
    failure_count INTEGER
);
CREATE INDEX IF NOT EXISTS idx_metrics_aggregate_job_period
    ON metrics_aggregate(job_id, period_type, period_start);
`,
	},
	{
		name: "008_create_settings",
		query: `
CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`,
	},
}

// RunMigrations executes all pending migrations
func RunMigrations(db *sql.DB) error {
	// Create migrations table if not exists
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL,
    executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	for _, m := range migrations {
		// Check if migration already ran
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE name = ?)", m.name).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check migration %s: %w", m.name, err)
		}

		if exists {
			continue
		}

		// Execute migration
		if _, err := db.Exec(m.query); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", m.name, err)
		}

		// Record migration as executed
		if _, err := db.Exec("INSERT INTO schema_migrations (name) VALUES (?)", m.name); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", m.name, err)
		}
	}

	return nil
}
