package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	// Open in‑memory SQLite database.
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Initialize schema.
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Verify that the links table exists.
	var cnt int
	row := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='links'`)
	if err := row.Scan(&cnt); err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if cnt != 1 {
		t.Fatalf("expected links table to exist")
	}
}
