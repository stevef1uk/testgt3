package store

import (
	"database/sql"
	"os"
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
		t.Fatalf("InitSchema: %v", err)
	}

	// Verify table exists
	var name string
	err = db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='links'").Scan(&name)
	if err != nil {
		t.Fatalf("table links not found: %v", err)
	}
	if name != "links" {
		t.Fatalf("expected table name 'links', got %q", name)
	}

	// Verify columns
	rows, err := db.Query("PRAGMA table_info(links)")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	cols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		cols[name] = true
	}
	for _, want := range []string{"id", "title", "url", "created_at"} {
		if !cols[want] {
			t.Errorf("missing column %q", want)
		}
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
