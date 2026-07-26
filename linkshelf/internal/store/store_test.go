package store_test

import (
	"context"
	"database/sql"
	"testing"

	"linkshelf/internal/store"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := store.InitSchema(db); err != nil {
		t.Fatalf("InitSchema: %v", err)
	}

	// Set package-level DB for the test
	store.DB = db
	return db
}

func TestList_Empty(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	links, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List returned error on empty db: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected empty list, got %d links", len(links))
	}
}

func TestCreateAndList(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	created, err := store.Create(ctx, "Example", "https://example.com")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Title != "Example" {
		t.Errorf("expected title 'Example', got %q", created.Title)
	}
	if created.URL != "https://example.com" {
		t.Errorf("expected URL 'https://example.com', got %q", created.URL)
	}
	if created.ID == 0 {
		t.Errorf("expected non-zero ID")
	}
	if created.CreatedAt == "" {
		t.Errorf("expected non-empty CreatedAt")
	}

	links, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(links))
	}
	if links[0].ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, links[0].ID)
	}
}

func TestDelete(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	created, err := store.Create(ctx, "To Delete", "https://delete.me")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if err := store.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	links, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(links) != 0 {
		t.Fatalf("expected 0 links after delete, got %d", len(links))
	}
}

func TestDelete_NotFound(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	err := store.Delete(ctx, 999)
	if err == nil {
		t.Fatal("expected error for deleting non-existent link, got nil")
	}
}

func TestCreate_MultipleOrder(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()

	a, _ := store.Create(ctx, "A", "https://a.com")
	b, _ := store.Create(ctx, "B", "https://b.com")
	c, _ := store.Create(ctx, "C", "https://c.com")

	links, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(links) != 3 {
		t.Fatalf("expected 3 links, got %d", len(links))
	}
	// Expect C, B, A (ORDER BY id DESC)
	if links[0].ID != c.ID {
		t.Errorf("expected first link ID %d (C), got %d", c.ID, links[0].ID)
	}
	if links[1].ID != b.ID {
		t.Errorf("expected second link ID %d (B), got %d", b.ID, links[1].ID)
	}
	if links[2].ID != a.ID {
		t.Errorf("expected third link ID %d (A), got %d", a.ID, links[2].ID)
	}
}
