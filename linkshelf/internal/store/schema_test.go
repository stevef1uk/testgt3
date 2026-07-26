package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	var count int
	row := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='links'`)
	if err := row.Scan(&count); err != nil {
		t.Fatalf("query sqlite_master: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected links table to exist, got count=%d", count)
	}
}
