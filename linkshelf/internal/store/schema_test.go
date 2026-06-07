package store

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestInitSchemaCreatesTable(t *testing.T) {
	// Open an in‑memory SQLite database.
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	// Initialize the schema.
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Verify that the table exists with the correct columns.
	const query = `PRAGMA table_info(links);`
	rows, err := db.Query(query)
	if err != nil {
		t.Fatalf("PRAGMA query failed: %v", err)
	}
	defer rows.Close()

	type colInfo struct {
		CID          int
		Name         string
		Type         string
		NotNull      int
		DefaultValue sql.NullString
		PK           int
	}
	cols := make(map[string]colInfo)
	for rows.Next() {
		var ci colInfo
		if err := rows.Scan(&ci.CID, &ci.Name, &ci.Type, &ci.NotNull, &ci.DefaultValue, &ci.PK); err != nil {
			t.Fatalf("row scan failed: %v", err)
		}
		cols[ci.Name] = ci
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("rows error: %v", err)
	}

	// Expected columns.
	expected := map[string]struct {
		typ     string
		notnull int
		pk      int
	}{
		"id":         {"INTEGER", 0, 1},
		"title":      {"TEXT", 1, 0},
		"url":        {"TEXT", 1, 0},
		"created_at": {"TEXT", 1, 0},
	}
	for name, exp := range expected {
		ci, ok := cols[name]
		if !ok {
			t.Fatalf("column %s not found", name)
		}
		if ci.Type != exp.typ {
			t.Fatalf("column %s type = %s; want %s", name, ci.Type, exp.typ)
		}
		if ci.NotNull != exp.notnull {
			t.Fatalf("column %s NOT NULL = %d; want %d", name, ci.NotNull, exp.notnull)
		}
		if ci.PK != exp.pk {
			t.Fatalf("column %s primary key = %d; want %d", name, ci.PK, exp.pk)
		}
	}
}

// Test that the CHECK constraints for title length and URL prefix are enforced.
func TestInitSchemaConstraints(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()
	if err := InitSchema(db); err != nil {
		t.Fatalf("InitSchema failed: %v", err)
	}

	// Title length > 200 should fail.
	longTitle := make([]byte, 201)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	_, err = db.Exec(`INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`,
		string(longTitle), "https://example.com", time.Now().Format(time.RFC3339))
	if err == nil {
		t.Fatalf("expected title length constraint error, got nil")
	}

	// Invalid URL should fail.
	_, err = db.Exec(`INSERT INTO links (title, url, created_at) VALUES (?, ?, ?)`,
		"Valid Title", "ftp://example.com", time.Now().Format(time.RFC3339))
	if err == nil {
		t.Fatalf("expected URL prefix constraint error, got nil")
	}
}
