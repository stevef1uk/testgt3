package store_test

import (
	"database/sql"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema_CreatesTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	// Verify the table exists by querying it
	rows, err := db.Query("SELECT id, title, url, created_at FROM links")
	if err != nil {
		t.Fatalf("table 'links' does not exist after InitSchema: %v", err)
	}
	rows.Close()
}

func TestInitSchema_Idempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("first InitSchema: %v", err)
	}
	// Running InitSchema again should not error (idempotent)
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("second InitSchema (idempotent): %v", err)
	}
}

func TestInitSchema_InsertAndRead(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	_, err = db.Exec("INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)",
		"Test", "https://test.com", "2025-01-01T00:00:00Z")
	if err != nil {
		t.Fatalf("insert failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM links").Scan(&count); err != nil {
		t.Fatalf("count query: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 row, got %d", count)
	}
}
