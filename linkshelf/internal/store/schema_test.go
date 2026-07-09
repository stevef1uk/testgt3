package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema returned error: %v", err)
	}

	// Verify the table exists by querying it.
	row := db.QueryRow("SELECT count(*) FROM links")
	var count int
	if err := row.Scan(&count); err != nil {
		t.Fatalf("could not query links table: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 rows, got %d", count)
	}
}

func TestInitSchema_idempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("first InitSchema call failed: %v", err)
	}
	// Second call must not error (idempotent).
	if err := InitSchema(db); err != nil {
		t.Fatalf("second InitSchema call failed (not idempotent): %v", err)
	}
}
