package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// NewTestStore creates an in-memory SQLite database for testing
func NewTestStore(t *testing.T) *Store {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to create test database: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping test database: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	return &Store{db: db}
}
