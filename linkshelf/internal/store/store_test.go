package store

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestStoreCRUD(t *testing.T) {
	// Open in‑memory SQLite database.
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer db.Close()

	// Initialize schema.
	if err := InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}

	// Wire package‑level DB.
	DB = db

	// Create a new link.
	link := Link{
		Title:     "Example",
		URL:       "https://example.com",
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	id, err := Create(link)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non‑zero ID")
	}

	// List should contain the created link.
	links, err := List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ID != id {
		t.Fatalf("expected ID %d, got %d", id, links[0].ID)
	}

	// Delete the link.
	if err := Delete(id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// List should now be empty.
	links, err = List()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}
