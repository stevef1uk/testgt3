package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3" // Use sqlite3 for testing
)

func TestInitSchema(t *testing.T) {
	// Create an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Test initializing the schema
	err = InitSchema(db)
	if err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Verify that the 'links' table exists
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='links';")
	if err != nil {
		t.Fatalf("failed to query for links table: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Errorf("links table was not created")
	}

	// Test calling InitSchema again to ensure it's idempotent
	err = InitSchema(db)
	if err != nil {
		t.Fatalf("InitSchema failed on second call: %v", err)
	}
}
