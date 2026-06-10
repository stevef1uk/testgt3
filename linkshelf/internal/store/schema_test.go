package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory db: %v", err)
	}
	defer db.Close()
	if err := InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	var exists bool
	err = db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='links'`).Scan(&exists)
	if err != nil {
		t.Fatalf("check table: %v", err)
	}
	if !exists {
		t.Error("links table was not created")
	}
}
