package store_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"

	"linkshelf/internal/store"
)

func TestInitSchemaCreatesLinksTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	var tableName string
	if err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'links'`).Scan(&tableName); err != nil {
		t.Fatalf("links table was not created: %v", err)
	}
	if tableName != "links" {
		t.Fatalf("table name = %q, want links", tableName)
	}
}

func TestInitSchemaIsIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("first init schema: %v", err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("second init schema: %v", err)
	}
}

func TestLinkHasSchemaFields(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	rows, err := db.Query(`PRAGMA table_info(links)`)
	if err != nil {
		t.Fatalf("inspect links schema: %v", err)
	}
	defer rows.Close()

	want := []string{"id", "title", "url", "created_at"}
	var got []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, primaryKey int
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatalf("scan schema row: %v", err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate schema: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("columns = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("column %d = %q, want %q", i, got[i], want[i])
		}
	}
}
