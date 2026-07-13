package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Verify the table exists by querying its schema
	var tableName string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='links'").Scan(&tableName)
	if err != nil {
		t.Fatalf("expected 'links' table to exist: %v", err)
	}
	if tableName != "links" {
		t.Fatalf("expected table name 'links', got %q", tableName)
	}
}
