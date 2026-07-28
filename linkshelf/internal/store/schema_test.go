package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	// Open an in‑memory SQLite database.
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite db: %v", err)
	}
	defer db.Close()

	// Initialize the schema.
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	// Verify the table exists.
	var count int
	row := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='links'`)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected links table to exist, got count=%d", count)
	}
}
