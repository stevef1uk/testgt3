package store_test

import (
	"database/sql"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchemaCreatesLinksTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	var tableName string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'links'`).Scan(&tableName)
	if err != nil {
		t.Fatalf("links table was not created: %v", err)
	}
	if tableName != "links" {
		t.Fatalf("table name = %q, want links", tableName)
	}
}

func TestInitSchemaIsIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatal(err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("second InitSchema() error = %v", err)
	}
}
