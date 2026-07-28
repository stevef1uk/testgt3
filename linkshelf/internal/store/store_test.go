package store

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupStore(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	DB = db
	return db
}

func closeStore(db *sql.DB) {
	db.Close()
	DB = nil
}

func TestListEmpty(t *testing.T) {
	db := setupStore(t)
	defer closeStore(db)

	ctx := context.Background()
	links, err := List(ctx)
	if err != nil {
		t.Fatalf("List on empty store: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected empty list, got %d items", len(links))
	}
}

func TestCreateAndList(t *testing.T) {
	db := setupStore(t)
	defer closeStore(db)

	ctx := context.Background()
	link, err := Create(ctx, "Test Title", "https://example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if link.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if link.Title != "Test Title" {
		t.Fatalf("expected Title 'Test Title', got %q", link.Title)
	}
	if link.URL != "https://example.com" {
		t.Fatalf("expected URL 'https://example.com', got %q", link.URL)
	}
	if link.CreatedAt == "" {
		t.Fatal("expected non-empty CreatedAt")
	}

	links, err := List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ID != link.ID {
		t.Fatalf("ID mismatch: %d vs %d", links[0].ID, link.ID)
	}
}

func TestCreateEmptyValidation(t *testing.T) {
	db := setupStore(t)
	defer closeStore(db)

	ctx := context.Background()
	_, err := Create(ctx, "", "https://example.com")
	if err == nil {
		t.Fatal("expected error for empty title")
	}
	_, err = Create(ctx, "title", "")
	if err == nil {
		t.Fatal("expected error for empty url")
	}
}

func TestDelete(t *testing.T) {
	db := setupStore(t)
	defer closeStore(db)

	ctx := context.Background()
	link, err := Create(ctx, "Delete Me", "https://delete.me")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := Delete(ctx, link.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	links, err := List(ctx)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}

func TestDeleteNonExistent(t *testing.T) {
	db := setupStore(t)
	defer closeStore(db)

	ctx := context.Background()
	if err := Delete(ctx, 999); err != nil {
		t.Fatalf("Delete on non-existent id: %v", err)
	}
}
