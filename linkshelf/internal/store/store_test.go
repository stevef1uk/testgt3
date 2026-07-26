package store_test

import (
	"database/sql"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func setupDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := store.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	// Wire the DB for the store package
	store.DB = db
	return db
}

func TestCreateListDelete(t *testing.T) {
	db := setupDB(t)
	defer db.Close()

	// Create a link
	id, err := store.Create("Example", "https://example.com")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if id == 0 {
		t.Fatalf("expected non-zero id")
	}

	// List should contain the link
	links, err := store.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].Title != "Example" || links[0].URL != "https://example.com" {
		t.Fatalf("unexpected link data %+v", links[0])
	}

	// Delete the link
	if err := store.Delete(id); err != nil {
		t.Fatalf("delete: %v", err)
	}

	// List should be empty
	links, err = store.List()
	if err != nil {
		t.Fatalf("list after delete: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}
