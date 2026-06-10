package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestSetupTestDB(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}
}

func TestList(t *testing.T) {
	// Test List function
}
