package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Store handles all database operations
type Store struct {
	db *sql.DB
}

// New creates a new Store instance and initializes the database
func New(dbPath string) (*Store, error) {
	// DSN options apply to every pooled connection, including those opened later.
	separator := "?"
	if strings.Contains(dbPath, "?") {
		separator = "&"
	}
	db, err := sql.Open("sqlite3", dbPath+separator+"_busy_timeout=5000&_synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Enable WAL mode for concurrent read/write access.
	// Without WAL, writers hold exclusive locks that block all readers,
	// causing API queries to stall while the executor writes job logs.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return nil, fmt.Errorf("failed to enable WAL mode: %w", err)
	}

	// SQLite serializes writes; retain pooled connections for concurrent readers.
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)

	// Run migrations
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Store{db: db}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// DB returns the underlying database connection for advanced queries
func (s *Store) DB() *sql.DB {
	return s.db
}
