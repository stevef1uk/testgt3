package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchemaCreatesLinksTable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
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

func TestLinkHasExpectedSchemaColumns(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema() error = %v", err)
	}

	rows, err := db.Query(`PRAGMA table_info(links)`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	want := []string{"id", "title", "url", "created_at"}
	var got []string
	for rows.Next() {
		var cid int
		var name, typ string
		var notNull, primaryKey int
		var defaultValue any
		if err := rows.Scan(&cid, &name, &typ, &notNull, &defaultValue, &primaryKey); err != nil {
			t.Fatal(err)
		}
		got = append(got, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("columns = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("columns = %v, want %v", got, want)
		}
	}
}
